package controllers

import (
	"fmt"
	"reflect"
	"strings"

	"cdap.io/cdap-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
)

var deploymentPlanner *DeploymentPlan

func init() {
	deploymentPlanner = &DeploymentPlan{}
	deploymentPlanner.Init()
}

type DeploymentPlan struct {
	// Map from number of Pods to a list of statefuls,  deployments and k8s services where each statefulset/deployment
	// is a multi-container Pod with a list of services in it.
	planMap map[int32]ServiceGroups
}
type ServiceGroups struct {
	// Map from statefulset object name to a list services in this statefulset.
	stateful map[ServiceGroupName]ServiceGroup
	// Map from deployment object name to a list services in this deployment.
	deployment map[ServiceGroupName]ServiceGroup
	// Map from k8s service object name to a CDAP service to be exposed.
	networkService map[NetworkServiceName]ServiceName
}
type ServiceGroupName = string
type ServiceGroup = []ServiceName
type NetworkServiceName = string

func (d *DeploymentPlan) Init() {
	d.planMap = make(map[int32]ServiceGroups)

	// Default: each service group runs in its own Pod and each service in it's own container.
	// If there are mltiple services in a pod,
	// the first one (index 0) will be considered as the
	// main container and subsequent ones as sidecar containers.
	d.planMap[0] = ServiceGroups{
		stateful: map[ServiceGroupName]ServiceGroup{
			"logs":      {serviceLogs, serviceSystemMetricsExporter},
			"messaging": {serviceMessaging, serviceSystemMetricsExporter},
			"metrics":   {serviceMetrics, serviceSystemMetricsExporter},
			"preview":   {servicePreview, serviceSystemMetricsExporter},
			"appfabric": {serviceAppFabric, serviceSystemMetricsExporter},
			"runtime":   {serviceRuntime, serviceSystemMetricsExporter},
		},
		deployment: map[ServiceGroupName]ServiceGroup{
			"authentication": {serviceAuthentication},
			"metadata":       {serviceMetadata},
			"router":         {serviceRouter},
			"userinterface":  {serviceUserInterface},
		},
		networkService: map[NetworkServiceName]ServiceName{
			"router":        serviceRouter,
			"userinterface": serviceUserInterface,
		},
	}
}

// Given desired number of pods, return a list of service groups where each group contains services colocated in the
// same pod.
func (d *DeploymentPlan) getPlan(numPods int32) (*ServiceGroups, error) {
	s, ok := d.planMap[numPods]
	if !ok {
		return nil, fmt.Errorf("unsupported deployment plan for NumPods %d", numPods)
	}
	return &s, nil
}

// Build deployment plan (e.g. a list of statefulsets, deployments and NodePort services)
func buildDeploymentPlanSpec(master *v1alpha1.CDAPMaster, labels map[string]string) (*DeploymentPlanSpec, error) {
	// Wait for version update handler to set the image version to use in the status field
	if master.Status.ImageToUse == "" || master.Status.UserInterfaceImageToUse == "" {
		return &DeploymentPlanSpec{}, nil
	}

	// Get the deployment plan. At the moment, it only supports numPods = 0 (i.e. one service per pod).
	// But can be easily extended to colocate services together in a single pod to form a multi-container pod
	var numPods int32 = 0
	serviceGroups, err := deploymentPlanner.getPlan(numPods)
	if err != nil {
		return nil, err
	}

	cconf := getObjName(master, configMapCConf)
	hconf := getObjName(master, configMapHConf)
	sysappconf := getObjName(master, configMapSysAppConf)
	dataDir := master.Spec.Config[confLocalDataDirKey]

	spec := newDeploymentPlanSpec()
	// Build statefulsets
	for k, v := range serviceGroups.stateful {
		name := k
		services := v
		stateful, err := buildStatefulSets(master, name, services, labels, cconf, hconf, sysappconf, dataDir)
		if err != nil {
			return nil, err
		}
		// stateful could be nil when the list of services are optional and they are disabled in CR
		// (i.e. service spec is set to nil)
		if stateful == nil {
			continue
		}
		spec = spec.withStateful(stateful)
	}

	// Build deployment
	for k, v := range serviceGroups.deployment {
		name := k
		services := v
		deploymentSpec, err := buildDeployment(master, name, services, labels, cconf, hconf, sysappconf, dataDir)
		if err != nil {
			return nil, err
		}
		// deploymentSpec could be nil when the list of services are optional and they are disabled in CR
		// (i.e. service spec is set to nil)
		if deploymentSpec == nil {
			continue
		}
		spec = spec.withDeployment(deploymentSpec)
	}
	// Build NodePort service
	for name, targetService := range serviceGroups.networkService {
		networkService, err := buildNetworkService(master, name, targetService, labels)
		if err != nil {
			return nil, err
		}
		spec = spec.withNetworkService(networkService)
	}
	return spec, nil
}

// Return a single single-/multi- container StatefulSets containing a list of supplied services
func buildStatefulSets(master *v1alpha1.CDAPMaster, name string, services ServiceGroup, labels map[string]string, cconf, hconf, sysappconf, dataDir string) (*StatefulSpec, error) {
	objName := getObjName(master, name)
	serviceAccount, err := getServiceAccount(master, services)
	if err != nil {
		return nil, err
	}
	runtimeClass, err := getRuntimeClass(master, services)
	if err != nil {
		return nil, err
	}
	priorityClass, err := getPriorityClass(master, services)
	if err != nil {
		return nil, err
	}
	nodeSelector, err := getNodeSelector(master, services)
	if err != nil {
		return nil, err
	}
	securityContext, err := getSecurityContext(master, services)
	if err != nil {
		return nil, err
	}

	spec := newStatefulSpec(master, objName, labels, cconf, hconf, sysappconf).
		setServiceAccountName(serviceAccount).
		setNodeSelector(nodeSelector).
		setRuntimeClassName(runtimeClass).
		setPriorityClassName(priorityClass).
		setSecurityContext(securityContext)

	// Add init container
	spec = spec.withInitContainer(
		newContainerSpec(master, "StorageInit", dataDir).setArgs(containerStorageMain))

	metricsSidecarInGroup, _ := findInStringArray(services, serviceSystemMetricsExporter)
	enableSystemMetricsExporter, jmxServerPort := jmxServerPort(&master.Spec)

	// Add each service as a container
	for idx, s := range services {
		ss, err := getCDAPServiceSpec(master, s)
		if err != nil {
			return nil, err
		}
		// This happens when the service is optional and disabled in CR
		// (i.e. service spec is set to nil)
		if ss == nil {
			continue
		}
		env := addJavaMaxHeapEnvIfNotPresent(ss.Env, ss.Resources)
		c := newContainerSpec(master, s, dataDir).
			setResources(ss.Resources).
			setEnv(env)
		isSidecar := (idx > 0)
		// Only main container run a jmx server if enabed
		if !isSidecar && metricsSidecarInGroup && enableSystemMetricsExporter {
			c = c.addEnv(javaOptsEnvVarName, fmt.Sprintf(runJMXServerJavaOptFormat, jmxServerPort))
		}

		if s == serviceUserInterface {
			c = updateSpecForUserInterface(master, c)
		}
		spec = spec.withContainer(c)
		// Adding a label to allow NodePort service selector to find the pod
		// Adding pod labels for multiple services causes re-conciliation errors when
		// a new service is added/removed as only `template` and `updateStrategy`
		// props of spec can be modified
		if !isSidecar {
			spec = spec.addLabel(labelContainerKeyPrefix+s, master.Name)
		}

		// Mount extra volumes from ConfigMap and Secret
		if _, err := spec.addConfigMapVolumes(ss.ConfigMapVolumes); err != nil {
			return nil, err
		}
		if _, err := spec.addSecretVolumes(ss.SecretVolumes); err != nil {
			return nil, err
		}
	}

	// All services are optional services and are disabled in CR.
	// Return nil to indicate no statefulset is built.
	if len(spec.Containers) == 0 {
		return nil, nil
	}

	// Get storage class and calculates total disk size required
	storageClass, err := getStorageClass(master, services)
	if err != nil {
		return nil, err
	}
	storageSize, err := aggregateStorageSize(master, services)
	if err != nil {
		return nil, err
	}
	spec = spec.withStorage(storageClass, storageSize)
	return spec, nil
}

// Return a single single-/multi- container deployment containing a list of supplied services
func buildDeployment(master *v1alpha1.CDAPMaster, name string, services ServiceGroup, labels map[string]string, cconf, hconf, sysappconf, dataDir string) (*DeploymentSpec, error) {
	objName := getObjName(master, name)
	serviceAccount, err := getServiceAccount(master, services)
	if err != nil {
		return nil, err
	}
	runtimeClass, err := getRuntimeClass(master, services)
	if err != nil {
		return nil, err
	}
	priorityClass, err := getPriorityClass(master, services)
	if err != nil {
		return nil, err
	}
	nodeSelector, err := getNodeSelector(master, services)
	if err != nil {
		return nil, err
	}
	replicas, err := getReplicas(master, services)
	if err != nil {
		return nil, err
	}
	securityContext, err := getSecurityContext(master, services)
	if err != nil {
		return nil, err
	}

	spec := newDeploymentSpec(master, objName, labels, cconf, hconf, sysappconf).
		setServiceAccountName(serviceAccount).
		setNodeSelector(nodeSelector).
		setRuntimeClassName(runtimeClass).
		setPriorityClassName(priorityClass).
		setReplicas(replicas).
		setSecurityContext(securityContext)

	// Add each service as a container
	for _, s := range services {
		ss, err := getCDAPServiceSpec(master, s)
		if err != nil {
			return nil, err
		}
		// This happens when the service is optional and disabled in CR
		// (i.e. service spec is set to nil)
		if ss == nil {
			continue
		}
		env := addJavaMaxHeapEnvIfNotPresent(ss.Env, ss.Resources)
		c := newContainerSpec(master, s, dataDir).setResources(ss.Resources).setEnv(env)
		if s == serviceUserInterface {
			c = updateSpecForUserInterface(master, c)
		}
		spec = spec.withContainer(c)

		// Adding a label to allow k8s service selector to easily find the pod
		spec = spec.addLabel(labelContainerKeyPrefix+s, master.Name)

		// Mount extra volumes from ConfigMap and Secret
		if _, err := spec.addConfigMapVolumes(ss.ConfigMapVolumes); err != nil {
			return nil, err
		}
		if _, err := spec.addSecretVolumes(ss.SecretVolumes); err != nil {
			return nil, err
		}
	}
	// All services are optional services and are disabled in CR.
	// Return nil to indicate no deployment is built.
	if len(spec.Containers) == 0 {
		return nil, nil
	}
	return spec, nil
}

// Return a list of reconciler objects (e.g. statefulsets, deployment, NodePort service) for the given deployment plan
func buildObjectsForDeploymentPlan(spec *DeploymentPlanSpec) ([]reconciler.Object, error) {
	var objs []reconciler.Object
	for _, s := range spec.Stateful {
		obj, err := buildStatefulSetsObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)
	}
	for _, s := range spec.Deployment {
		obj, err := buildDeploymentObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)
	}
	for _, s := range spec.NetworkServices {
		obj, err := buildNetworkServiceObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)

	}
	return objs, nil
}

// Return a reconciler statefulset object for the given statefulsets spec
func buildStatefulSetsObject(spec *StatefulSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateStatefulSet, spec, &appsv1.StatefulSetList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Return a reconciler deployment object for the given deployment spec
func buildDeploymentObject(spec *DeploymentSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateDeployment, spec, &appsv1.DeploymentList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Return a NodePort service to expose the supplied target service
func buildNetworkService(master *v1alpha1.CDAPMaster, name NetworkServiceName, target ServiceName, labels map[string]string) (*NetworkServiceSpec, error) {
	s, err := getCDAPExternalServiceSpec(master, target)
	if err != nil {
		return nil, err
	}
	objName := getObjName(master, name)
	return newNetworkServiceSpec(objName, labels, s.ServiceType, s.ServicePort, master).
		addSelector(labelContainerKeyPrefix+target, master.Name), nil
}

// Return a reconciler NodePort service object for the given network service spec
func buildNetworkServiceObject(spec *NetworkServiceSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateService, spec, &corev1.ServiceList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Return the StorageClassName for the supplied list of services. Ignore services that are stateless. May return empty.
// Fail if found  multiple conflicting settings, this is to ensure consistent settings for the supplied services to
// be colocated in the same pod
func getStorageClass(master *v1alpha1.CDAPMaster, services ServiceGroup) (string, error) {
	vals := make(map[string]bool)
	storageClass := ""
	for _, s := range services {
		ss, err := getCDAPStatefulServiceSpec(master, s)
		if err != nil {
			return "", err
		}
		if ss == nil {
			// the service in the supplied list might be a stateless.
			// Depending on deployment plan, we may colocate a stateful and a stateless service in the same pod.
			continue
		}
		if ss.StorageClassName == nil {
			continue
		}

		if len(*ss.StorageClassName) == 0 {
			continue
		}
		vals[*ss.StorageClassName] = true
		storageClass = *ss.StorageClassName
	}
	if len(vals) > 1 {
		return "", fmt.Errorf("conflicting StorageClassNames across services (%s)", strings.Join(services, ","))
	}
	return storageClass, nil
}

// Return the aggregated storage size across the supplied list of services. Ignore the services that are stateless
// and doesn't have storage settings. If no storage size found, return default storage size.
// Fail if unable to parse storage size string
func aggregateStorageSize(master *v1alpha1.CDAPMaster, services ServiceGroup) (string, error) {
	total := resource.NewQuantity(0, resource.BinarySI)
	for _, s := range services {
		ss, err := getCDAPStatefulServiceSpec(master, s)
		if err != nil {
			return "", err
		}
		if ss == nil {
			// the service in the supplied list might be a stateless.
			// Depending on deployment plan, we may colocate a stateful and a stateless service in the same pod.
			continue
		}
		if len(ss.StorageSize) == 0 {
			continue
		}
		size, err := resource.ParseQuantity(ss.StorageSize)
		if err != nil {
			return "", err
		}
		total.Add(size)
	}
	if total.IsZero() {
		return defaultStorageSize, nil
	}
	return total.String(), nil
}

// Return merged nodeSelector map across all supplied services
func getNodeSelector(master *v1alpha1.CDAPMaster, services ServiceGroup) (map[string]string, error) {
	nodeSelector := make(map[string]string)
	for _, service := range services {
		spec, err := getCDAPServiceSpec(master, service)
		if err != nil {
			return nil, err
		}
		// This happens when the service is optional and disabled in CR
		// (i.e. service spec is set to nil)
		if spec == nil {
			continue
		}
		nodeSelector = mergeMaps(nodeSelector, spec.NodeSelector)
	}
	return nodeSelector, nil
}

// Return PriorityClassName if all supplied services have the same setting, otherwise return an error
func getPriorityClass(master *v1alpha1.CDAPMaster, services ServiceGroup) (string, error) {
	priorityClass := ""
	if err := getStringFieldValue(master, services, "PriorityClassName", &priorityClass); err != nil {
		return "", err
	}
	return priorityClass, nil
}

// Return RuntimeClassName if all supplied services have the same setting, otherwise return an error
func getRuntimeClass(master *v1alpha1.CDAPMaster, services ServiceGroup) (string, error) {
	runtimeClass := ""
	if err := getStringFieldValue(master, services, "RuntimeClassName", &runtimeClass); err != nil {
		return "", err
	}
	return runtimeClass, nil
}

// Return default ServiceAccountName in CDAPMaster.Spec or the overridden setting if all supplied services have the
// same overridden setting, otherwise return an error
func getServiceAccount(master *v1alpha1.CDAPMaster, services ServiceGroup) (string, error) {
	serviceAccount := master.Spec.ServiceAccountName
	if err := getStringFieldValue(master, services, "ServiceAccountName", &serviceAccount); err != nil {
		return "", err
	}
	return serviceAccount, nil
}

// Return default SecurityContext in CDAPMaster.Spec or the overridden security context.
func getSecurityContext(master *v1alpha1.CDAPMaster, services ServiceGroup) (*v1alpha1.SecurityContext, error) {
	securityContext := master.Spec.SecurityContext
	val, err := getFieldValueIfUnique(master, services, "SecurityContext")
	if err != nil {
		return nil, err
	}
	if overriddenSecurityContext, ok := val.(v1alpha1.SecurityContext); ok {
		return &overriddenSecurityContext, nil
	}
	return securityContext, nil
}

// getReplicas returns the Replicas if all supplied services have the same setting, otherwise return an error
func getReplicas(master *v1alpha1.CDAPMaster, services ServiceGroup) (int32, error) {
	replicas := int32(0)
	for _, service := range services {
		if spec, err := getCDAPScalableServiceSpec(master, service); err != nil {
			return 0, nil
		} else if spec != nil {
			if spec.Replicas == nil {
				continue
			}
			if replicas == 0 {
				replicas = *spec.Replicas
			} else if replicas != *spec.Replicas {
				return 0, fmt.Errorf("value of field Replicas not the same across (%s)", strings.Join(services, ","))
			}
		}
	}

	if replicas == 0 {
		return 1, nil
	}
	return replicas, nil
}

// Use reflection to extract the string value of supplied field name across all service specs. Return the value only
// if it is the same across all services. Otherwise return an error.
// This is used to ensure  all services in the same pod not having conflicting Pod-level settings
func getStringFieldValue(master *v1alpha1.CDAPMaster, services ServiceGroup, fieldName string, value *string) error {
	if newVal, err := getFieldValueIfUnique(master, services, fieldName); err != nil {
		return err
	} else if newVal != nil {
		// The field is set.
		if newVal, ok := newVal.(string); !ok {
			return fmt.Errorf("unable to cast value of field %v to string ", fieldName)
		} else {
			*value = newVal
		}
	}
	return nil
}

// Use reflection to extract the value of supplied field name across all service specs. Return the value only
// if it is the same across all services. Otherwise return an error.
// This is used to ensure all services in the same pod not having conflicting Pod-level settings
func getFieldValueIfUnique(master *v1alpha1.CDAPMaster, services ServiceGroup, fieldName string) (interface{}, error) {
	values := make([]interface{}, 0)
	// Get field value from CDAPServiceSpec of each service
	for _, service := range services {
		specVal := reflect.ValueOf(master.Spec).FieldByName(service)
		if !specVal.IsValid() {
			return nil, fmt.Errorf("filed %s not valid", service)
		}
		// For optional service, its service field should be a pointer to spec
		if specVal.Kind() == reflect.Ptr {
			if specVal.IsNil() {
				continue
			}
			specVal = specVal.Elem()
		}
		fieldVal := reflect.ValueOf(specVal.Interface()).FieldByName(fieldName)
		if !fieldVal.IsValid() {
			return nil, fmt.Errorf("filed %s not valid", fieldName)
		}
		if fieldVal.Kind() == reflect.Ptr {
			// Skip if nil pointer
			if fieldVal.IsNil() {
				continue
			}
			fieldVal = fieldVal.Elem()
		}
		// Skip if empty
		if fieldVal.Kind() == reflect.String && fieldVal.Len() == 0 {
			continue
		}
		values = append(values, fieldVal.Interface())
	}
	// Return the value only if they are all the same.
	// May return nil indicating value is either empty or not set.
	var returnVal interface{} = nil
	for i := 0; i < len(values); i++ {
		if returnVal == nil {
			returnVal = values[i]
		} else if !reflect.DeepEqual(returnVal, values[i]) {
			return nil, fmt.Errorf("value of field %v not the same across (%s)", fieldName, strings.Join(services, ","))
		}
	}
	return returnVal, nil
}

// Update settings in container spec for userinterface service
func updateSpecForUserInterface(master *v1alpha1.CDAPMaster, spec *ContainerSpec) *ContainerSpec {
	return spec.
		setImage(master.Status.UserInterfaceImageToUse).
		setWorkingDir("/opt/cdap/ui").
		setCommand("bin/node").
		setArgs("index.js", "start").
		addEnv("NODE_ENV", "production")
}

// Derive from memory resource requirements and add java max heap size to the supplied env var array if not present
func addJavaMaxHeapEnvIfNotPresent(env []corev1.EnvVar, resources *corev1.ResourceRequirements) []corev1.EnvVar {
	if resources == nil {
		return env
	}

	// Nothing to set if already present
	hasMaxHeap := false
	for _, e := range env {
		if e.Name == javaMaxHeapSizeEnvVarName {
			hasMaxHeap = true
		}
	}
	if hasMaxHeap {
		return env
	}

	// Derive from memory resource requirement
	memory := max(resources.Requests.Memory().Value(), resources.Limits.Memory().Value())
	if memory > 0 {
		xmx := max(memory-javaReservedNonHeap, int64(float64(memory)*javaMinHeapRatio))
		env = append(env, corev1.EnvVar{
			Name:  javaMaxHeapSizeEnvVarName,
			Value: fmt.Sprintf("-Xmx%v", xmx),
		})
	}
	return env
}
