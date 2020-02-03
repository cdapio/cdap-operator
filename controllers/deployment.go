package controllers

import (
	alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
)

var strategyHandler DeploymentStrategy

func init() {
	strategyHandler.Init()
}

type DeploymentStrategy struct {
	// Map from number of Pods to a list of statefuls,  deployments and k8s services where each statefulset/deployment
	// is a multi-container Pod with a list of services in it.
	strategyMap map[int32]ServiceGroupMap
}
type ServiceGroupMap struct {
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

func (d *DeploymentStrategy) Init() {
	d.strategyMap = make(map[int32]ServiceGroupMap)

	// Default: each service runs in its own Pod
	d.strategyMap[0] = ServiceGroupMap{
		stateful: map[ServiceGroupName]ServiceGroup{
			"logs":      {serviceLogs},
			"messaging": {serviceMessaging},
			"metrics":   {serviceMetrics},
			"preview":   {servicePreview},
		},
		deployment: map[ServiceGroupName]ServiceGroup{
			"appfabric":     {serviceAppFabric},
			"metadata":      {serviceMetadata},
			"router":        {serviceRouter},
			"userinterface": {serviceUserInterface},
		},
		networkService: map[NetworkServiceName]ServiceName{
			"router":        serviceRouter,
			"userinterface": serviceUserInterface,
		},
	}

	// 1 Pod: all services run in a single Pod
	d.strategyMap[1] = ServiceGroupMap{
		stateful: map[ServiceGroupName]ServiceGroup{
			"standalone": {
				serviceLogs, serviceMessaging, serviceMetrics, servicePreview, serviceAppFabric, serviceMetadata,
				serviceRouter, serviceUserInterface},
		},
		networkService: map[NetworkServiceName]ServiceName{
			"router":        serviceRouter,
			"userinterface": serviceUserInterface,
		},
	}

	// 2 Pod: UserInterface in its own Pod. All other services in "Backend" Pod
	d.strategyMap[2] = ServiceGroupMap{
		stateful: map[ServiceGroupName]ServiceGroup{
			"backend": {
				serviceLogs, serviceMessaging, serviceMetrics, servicePreview, serviceAppFabric, serviceMetadata,
				serviceRouter},
		},
		deployment: map[ServiceGroupName]ServiceGroup{
			"userinterface": {serviceUserInterface},
		},
		networkService: map[NetworkServiceName]ServiceName{
			"router":        serviceRouter,
			"userinterface": serviceUserInterface,
		},
	}
}

func (d *DeploymentStrategy) getStrategy(numPods int32) (*ServiceGroupMap, error) {
	s, ok := d.strategyMap[numPods]
	if !ok {
		return nil, fmt.Errorf("unsupported deployment strategy for NumPods %d", numPods)
	}
	return &s, nil
}

func buildCDAPDeploymentSpec(master *alpha1.CDAPMaster, labels map[string]string) (*CDAPDeploymentSpec, error) {
	m := master
	cconf := getObjectName(m.Name, configMapCConf)
	hconf := getObjectName(m.Name, configMapHConf)
	dataDir := confLocalDataDirVal

	var numPods int32 = 0
	if m.Spec.NumPods != nil {
		numPods = *m.Spec.NumPods
	}
	serviceGroupMap, err := strategyHandler.getStrategy(numPods)
	if err != nil {
		return nil, err
	}

	updateSpecForUserInterface := func(c *ContainerSpec) *ContainerSpec {
		return c.
			setWorkingDir("/opt/cdap/ui").
			setCommand("bin/node").
			setArgs("index.js", "start").
			addEnv("NODE_ENV", "production")
	}

	// Build a single StatefulSet
	buildStateful := func(name string, serviceGroup ServiceGroup) (*StatefulSpec, error) {
		objectName := getObjectName(m.Name, name)
		serviceAccount := m.Spec.ServiceAccountName // TODO: scan all service specs to find override
		nodeSelector := make(map[string]string)     // TODO: scan all service specs to do a merge
		runtimeClass := ""                          // TODO: scan all service specs to find override
		priorityClass := ""                         // TODO: scan all service spaces to find override
		spec := newStatefulSpec(objectName, labels, cconf, hconf, master).
			setServiceAccountName(serviceAccount).
			setNodeSelector(nodeSelector).
			setRuntimeClassName(runtimeClass).
			setPriorityClassName(priorityClass)

		spec = spec.withInitContainer(
			newContainerSpec("StorageInit", dataDir, master).setArgs(containerStorageMain))

		for _, service := range serviceGroup {
			c := newContainerSpec(service, dataDir, m).setResources(getCDAPServiceSpec(service, m).Resources)
			if service == serviceUserInterface {
				c = updateSpecForUserInterface(c)
			}
			spec = spec.withContainer(c)
			// Adding a label to allow k8s service selector to easily find the pod
			spec = spec.addLabel(labelContainerKeyPrefix+service, m.Name)
		}

		var storageClassName *string
		var totalStorageSize uint64 = 0
		for _, service := range serviceGroup {
			serviceStatefulSpec := getCDAPStatefulServiceSpec(service, m)
			if serviceStatefulSpec == nil {
				continue
			}
			if serviceStatefulSpec.StorageSize != "" {
				size, err := FromHumanReadableBytes(serviceStatefulSpec.StorageSize)
				if err != nil {
					return nil, fmt.Errorf("unable to parse stroage size")
				}
				totalStorageSize += size
			}
			if serviceStatefulSpec.StorageClassName != nil {
				if storageClassName == nil {
					storageClassName = serviceStatefulSpec.StorageClassName
				} else if *storageClassName != *serviceStatefulSpec.StorageClassName {
					return nil, fmt.Errorf("StorageClassName inconsistent across services in a group")
				}
			}
		}
		storageSize := ""
		if totalStorageSize > 0 {
			storageSize = ToHumanReadableBytes(totalStorageSize)
		}
		spec = spec.withStorage(storageClassName, storageSize)
		return spec, nil
	}

	buildDeployment := func(name string, serviceGroup ServiceGroup) *DeploymentSpec {
		objectName := getObjectName(m.Name, name)
		serviceAccount := m.Spec.ServiceAccountName // TODO: scan all service specs to find override
		nodeSelector := make(map[string]string)     // TODO: scan all service specs to do a merge
		runtimeClass := ""                          // TODO: scan all service specs to find override
		priorityClass := ""                         // TODO: scan all service spaces to find override
		spec := newDeploymentSpec(objectName, labels, cconf, hconf, master).
			setServiceAccountName(serviceAccount).
			setNodeSelector(nodeSelector).
			setRuntimeClassName(runtimeClass).
			setPriorityClassName(priorityClass)
		for _, service := range serviceGroup {
			c := newContainerSpec(service, dataDir, m).setResources(getCDAPServiceSpec(service, m).Resources)
			if service == serviceUserInterface {
				c = updateSpecForUserInterface(c)
			}
			spec = spec.withContainer(c)

			// Adding a label to allow k8s service selector to easily find the pod
			spec = spec.addLabel(labelContainerKeyPrefix+service, m.Name)
		}
		return spec
	}

	buildNetworkService := func(name NetworkServiceName, target ServiceName, networkType *string, networkPort *int32) *NetworkServiceSpec {
		objectName := getObjectName(m.Name, name)
		spec := newNetworkServiceSpec(objectName, labels, networkType, networkPort, master)
		spec = spec.addSelector(labelContainerKeyPrefix+target, m.Name)
		return spec
	}

	spec := newCDAPDeploymentSpec()
	for k, v := range serviceGroupMap.stateful {
		statefulsetSpec, err := buildStateful(k, v)
		if err != nil {
			return nil, err
		}
		spec = spec.withStateful(statefulsetSpec)
	}
	for k, v := range serviceGroupMap.deployment {
		spec = spec.withDeployment(buildDeployment(k, v))
	}
	for name, targetService := range serviceGroupMap.networkService {
		s := getCDAPExternalService(targetService, master)
		spec = spec.withNetworkService(buildNetworkService(name, targetService, s.ServiceType, s.ServicePort))
	}
	return spec, nil
}

func buildObjects(spec *CDAPDeploymentSpec) ([]reconciler.Object, error) {
	var objs []reconciler.Object
	for _, s := range spec.Stateful {
		obj, err := buildStatefulObject(s)
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

func buildStatefulObject(spec *StatefulSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateStatefulSet, spec, &appsv1.StatefulSetList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func buildDeploymentObject(spec *DeploymentSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateDeployment, spec, &appsv1.DeploymentList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func buildNetworkServiceObject(spec *NetworkServiceSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateService, spec, &corev1.ServiceList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
