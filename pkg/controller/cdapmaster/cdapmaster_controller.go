/*
Copyright 2019 The CDAP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cdapmaster

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	alpha1 "cdap.io/cdap-operator/pkg/apis/cdap/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gr "sigs.k8s.io/controller-reconciler/pkg/genericreconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	"sigs.k8s.io/controller-reconciler/pkg/status"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	containerLabel = "cdap.container"
	// Heap memory related constants
	javaMinHeapRatio     = float64(0.6)
	javaReservedNonHeap  = int64(768 * 1024 * 1024)
	templateDir          = "templates/"
	deploymentTemplate   = "cdap-deployment.yaml"
	uiDeploymentTemplate = "cdap-ui-deployment.yaml"
	statefulSetTemplate  = "cdap-sts.yaml"
	serviceTemplate      = "cdap-service.yaml"
	upgradeJobTemplate   = "upgrade-job.yaml"

	upgradeFailed             = "upgrade-failed"
	postUpgradeFailed         = "post-upgrade-failed"
	postUpgradeFinished       = "post-upgrade-finished"
	upgradeStartMessage       = "Upgrade started, received updated CR."
	upgradeFailedInitMessage  = "Failed to create job, upgrade failed."
	upgradeJobFailedMessage   = "Upgrade job failed."
	upgradeJobFinishedMessage = "Upgrade job finished."
	upgradeJobSkippedMessage  = "Upgrade job skipped."
	upgradeResetMessage       = "Upgrade spec reset."
	upgradeFailureLimit       = 4

	latestVersion = "latest"
)

// Add the manager to the controller.
func Add(mgr manager.Manager) error {
	r := newReconciler(mgr)
	return r.Controller(nil)
}

// TBD kubebuilder:rbac:groups=app.k8s.io,resources=applications,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cdap.cdap.io,resources=cdapmasters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cdap.cdap.io,resources=cdapmasters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) *gr.Reconciler {
	return gr.
		WithManager(mgr).
		For(&alpha1.CDAPMaster{}, alpha1.SchemeGroupVersion).
		Using(&Base{}).
		Using(&Messaging{}).
		Using(&AppFabric{}).
		Using(&Metrics{}).
		Using(&Logs{}).
		Using(&Metadata{}).
		Using(&Preview{}).
		Using(&Router{}).
		Using(&UserInterface{}).
		WithErrorHandler(handleError).
		WithDefaulter(applyDefaults).
		Build()
}

func handleError(resource interface{}, err error, kind string) {
	cm := resource.(*alpha1.CDAPMaster)
	if err != nil {
		cm.Status.SetError("ErrorSeen", err.Error())
	} else {
		cm.Status.ClearError()
	}
}

func applyDefaults(resource interface{}) {
	cm := resource.(*alpha1.CDAPMaster)
	cm.ApplyDefaults()
}

// Base - interface to handle cdapmaster
type Base struct{}

// Messaging - interface to handle cdapmaster
type Messaging struct{}

// AppFabric - interface to handle cdapmaster
type AppFabric struct{}

// Metrics - interface to handle cdapmaster
type Metrics struct{}

// Logs - interface to handle cdapmaster
type Logs struct{}

// Metadata - interface to handle cdapmaster
type Metadata struct{}

// Preview - interface to handle cdapmaster
type Preview struct{}

// Router - interface to handle cdapmaster
type Router struct{}

// UserInterface - interface to handle cdapmaster
type UserInterface struct{}

// ------------------------------ Common -------------------------------------

// Set the nodePort in the expected service based on the observed service
func setNodePort(expected, observed []reconciler.Object) {
	// Get the service from the expected list.
	var expectedService *corev1.Service
	for _, item := range reconciler.ObjectsByType(expected, k8s.Type) {
		if service, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
			expectedService = service
			break
		}
	}
	if expectedService == nil {
		return
	}
	// Find the service being observed. Extract nodePort from the service spec and set it to expected
	for _, item := range reconciler.ObjectsByType(observed, k8s.Type) {
		if observedService, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
			var nodePorts = make(map[string]int32)
			for _, p := range observedService.Spec.Ports {
				nodePorts[p.Name] = p.NodePort
			}

			// Assigning existing node ports to the expected service
			for i := range expectedService.Spec.Ports {
				p := &expectedService.Spec.Ports[i]
				if nodePort, ok := nodePorts[p.Name]; ok {
					p.NodePort = nodePort
				}
			}
		}
	}
}

type templateBaseValue struct {
	Name               string
	Labels             map[string]string
	Master             *alpha1.CDAPMaster
	Replicas           *int32
	ServiceAccountName string
	ServiceType        alpha1.ServiceType
	DataDir            string
	CConfName          string
	HConfName          string
	JavaMaxHeap        *int64
}

// Sets values to the given templateBaseValue based on the resources provided
func (v *templateBaseValue) setTemplateValue(rsrc interface{}, rsrclabels map[string]string, serviceType alpha1.ServiceType, serviceLabels map[string]string, serviceAccount string, resources *corev1.ResourceRequirements) {
	master := rsrc.(*alpha1.CDAPMaster)
	labels := make(reconciler.KVMap)
	labels.Merge(master.Labels, serviceLabels, rsrclabels)

	// Set the cdap.container label. It is for service selector to route correctly
	name := getResourceName(master, string(serviceType))
	labels[containerLabel] = name

	ServiceAccountName := master.Spec.ServiceAccountName
	if serviceAccount != "" {
		ServiceAccountName = serviceAccount
	}

	// Set the JAVA_HEAPMAX environment
	if resources != nil {
		memory := resources.Requests.Memory().Value()
		if resources.Limits.Memory().Value() > memory {
			memory = resources.Limits.Memory().Value()
		}
		// If resources memory is provided, set the Xmx.
		// The Xmx value is subtracted with 768M, with min heap ratio of 0.6
		// These values are from cdap default
		if memory > 0 {
			xmx := memory - javaReservedNonHeap
			minHeapSize := int64(float64(memory) * javaMinHeapRatio)
			if xmx < 0 || xmx < minHeapSize {
				xmx = minHeapSize
			}
			v.JavaMaxHeap = &xmx
		}
	}

	v.Name = name
	v.Labels = labels
	v.Master = master
	v.ServiceAccountName = ServiceAccountName
	v.ServiceType = serviceType
	v.DataDir = alpha1.LocalDataDir
	v.CConfName = getResourceName(master, "cconf")
	v.HConfName = getResourceName(master, "hconf")
}

// Gets the set of resources for the given service represented by the CDAPStatefulServiceSpec
// It consists of a StatefulSet for the given serviceType
func getStatefulServiceResources(s *alpha1.CDAPStatefulServiceSpec, rsrc interface{}, rsrclabels map[string]string, serviceType alpha1.ServiceType) ([]reconciler.Object, error) {
	ngdata := &statefulServiceValue{
		Service: s,
	}
	ngdata.setTemplateValue(rsrc, rsrclabels, serviceType, s.Labels, s.ServiceAccountName, s.Resources)
	return addResourceItem(&s.CDAPServiceSpec, statefulSetTemplate, ngdata, &appsv1.StatefulSetList{}, []reconciler.Object{})
}

// Struct containing data for templatization using CDAPServiceSpec
type serviceValue struct {
	templateBaseValue
	Service *alpha1.CDAPServiceSpec
}

// Struct containing data for templatization using CDAPStatefulServiceSpec
type statefulServiceValue struct {
	templateBaseValue
	Service *alpha1.CDAPStatefulServiceSpec
}

// Struct containing data for templatization using CDAPExternalServiceSpec
type externalServiceValue struct {
	templateBaseValue
	Service *alpha1.CDAPExternalServiceSpec
}

// Struct containing data for templatization using UpgradeJobSpec
type upgradeValue struct {
	templateBaseValue
	Job *upgradeJobSpec
}

// UpgradeJobSpec defines the specification for the upgrade job
type upgradeJobSpec struct {
	Image              string    `json:"image,omitempty"`
	JobName            string    `json:"jobName,omitempty"`
	HostName           string    `json:"hostName,omitempty"`
	BackoffLimit       int32     `json:"backoffLimit,omitempty"`
	ReferentName       string    `json:"referentName,omitempty"`
	ReferentKind       string    `json:"referentKind,omitempty"`
	ReferentApiVersion string    `json:"referentApiVersion,omitempty"`
	ReferentUID        types.UID `json:"referentUID,omitempty"`
	SecuritySecret     string    `json:"securitySecret,omitempty"`
	StartTimeMillis    int64     `json:"startTimeMillis,omitempty"`
	Namespace          string    `json:"namespace,omitempty"`
}

// Returns the resource name for the given resource
func getResourceName(r *alpha1.CDAPMaster, resource string) string {
	return fmt.Sprintf("cdap-%s-%s", r.Name, strings.ToLower(resource))
}

// Gets the name of the upgrade job based on resource version. Name can be no more than 63 chars.
func getUpgradeJobName(r *alpha1.CDAPMaster) string {
	return fmt.Sprintf("cdap-%s-uj-%d", r.Name, r.Status.UpgradeStartTimeMillis)
}

// Gets the name of the post upgrade job based on resource version. Name can be no more than 63 chars.
func getPostUpgradeJobName(r *alpha1.CDAPMaster) string {
	return fmt.Sprintf("cdap-%s-puj-%d", r.Name, r.Status.UpgradeStartTimeMillis)
}

// Gets the set of resources for the given service represented by the CDAPServiceSpec.
// It consists of a Deployment for the given serviceType
func getServiceResources(s *alpha1.CDAPServiceSpec, rsrc interface{}, rsrclabels map[string]string, serviceType alpha1.ServiceType) ([]reconciler.Object, error) {
	ngdata := &serviceValue{
		Service: s,
	}
	ngdata.setTemplateValue(rsrc, rsrclabels, serviceType, s.Labels, s.ServiceAccountName, s.Resources)
	return addResourceItem(s, deploymentTemplate, ngdata, &appsv1.DeploymentList{}, []reconciler.Object{})
}

// Gets the set of resources for the given service represented by the CDAPExternalServiceSpec
// It consists of a Deployment and a NodePort Service for the given serviceType
func getExternalServiceResources(s *alpha1.CDAPExternalServiceSpec, rsrc interface{}, rsrclabels map[string]string, serviceType alpha1.ServiceType, template string) ([]reconciler.Object, error) {
	ngdata := &externalServiceValue{
		Service: s,
	}
	ngdata.Replicas = s.Replicas
	ngdata.setTemplateValue(rsrc, rsrclabels, serviceType, s.Labels, s.ServiceAccountName, s.Resources)

	resources, err := addResourceItem(&s.CDAPScalableServiceSpec.CDAPServiceSpec, template, ngdata, &appsv1.DeploymentList{}, []reconciler.Object{})
	if err != nil {
		return nil, err
	}
	resources, err = addResourceItem(&s.CDAPScalableServiceSpec.CDAPServiceSpec, serviceTemplate, ngdata, &corev1.ServiceList{}, resources)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// Adds a resource.Item to the given resource.Bag by executing the given template.
func addResourceItem(s *alpha1.CDAPServiceSpec, template string, data interface{}, listType metav1.ListInterface, resources []reconciler.Object) ([]reconciler.Object, error) {
	rinfo, err := k8s.ObjectFromFile(templateDir+template, data, listType)
	if err != nil {
		return nil, err
	}
	// Set the resource for the first container if the object has container
	setResources(rinfo.Obj.(*k8s.Object).Obj, s.Resources)
	// Set the environment for the first container if the object has container
	setEnv(rinfo.Obj.(*k8s.Object).Obj, data, s.Env)
	resources = append(resources, *rinfo)
	return resources, nil
}

// Adds a resource.Item of ConfigMap type to the given resource.Bag
func addConfigMapItem(r *alpha1.CDAPMaster, name string, labels map[string]string, templates []string, resources []reconciler.Object) ([]reconciler.Object, error) {
	// Creates the configMap object
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Data: make(map[string]string),
	}

	ngdata := &templateBaseValue{
		Master: r,
	}

	// Load template files that goes into config map
	for _, tmplFile := range templates {
		content, err := logbackFromFile(tmplFile, ngdata)
		if err != nil {
			return nil, err
		}
		configMap.Data[tmplFile] = content
	}

	resources = append(resources, reconciler.Object{
		Type:      k8s.Type,
		Lifecycle: reconciler.LifecycleManaged,
		Obj: &k8s.Object{
			Obj:     configMap.DeepCopyObject().(metav1.Object),
			ObjList: &corev1.ConfigMapList{},
		},
	})

	return resources, nil
}

// Sets resources to the given object. It uses reflection to find and set the
// field `Spec.Template.Spec.Containers[0].Resources`
func setResources(obj interface{}, resources *corev1.ResourceRequirements) {
	if resources == nil {
		return
	}
	value := reflect.ValueOf(obj).Elem()

	for _, fieldName := range []string{"Spec", "Template", "Spec", "Containers"} {
		value = value.FieldByName(fieldName)
		if !value.IsValid() {
			return
		}
	}
	resourcesValue := value.Index(0).FieldByName("Resources")
	resourcesValue.Set(reflect.ValueOf(*resources))
}

// Set the environment for the first container object. It uses reflection to find and set the field
// `Spec.Template.Spec.Containers[0].Env`
func setEnv(obj interface{}, data interface{}, env []corev1.EnvVar) {
	if len(env) == 0 {
		return
	}

	value := reflect.ValueOf(obj).Elem()

	for _, fieldName := range []string{"Spec", "Template", "Spec", "Containers"} {
		value = value.FieldByName(fieldName)
		if !value.IsValid() {
			return
		}
	}
	envValue := value.Index(0).FieldByName("Env")
	// Add all environment variables
	envSlice := reflect.AppendSlice(envValue, reflect.ValueOf(env))

	// Add the JAVA_HEAPMAX env
	maxHeapValue := reflect.ValueOf(data).Elem().FieldByName("JavaMaxHeap")
	if maxHeapValue.IsValid() && !maxHeapValue.IsNil() {
		envSlice = reflect.Append(envSlice, reflect.ValueOf(corev1.EnvVar{
			Name:  "JAVA_HEAPMAX",
			Value: fmt.Sprintf("-Xmx%v", maxHeapValue.Elem().Int()),
		}))
	}

	envValue.Set(envSlice)
}

// itemFromReader reads the logback xml with template substitution
func logbackFromFile(t string, data interface{}) (string, error) {
	tmpl, err := template.New(t).ParseFiles(templateDir + t)
	if err != nil {
		return "", err
	}

	var output strings.Builder
	err = tmpl.Execute(&output, data)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

func resetUpgradeConditions(master *alpha1.CDAPMaster, message string) {
	master.Status.ClearCondition(upgradeFailed, message, message)
	master.Status.ClearCondition(postUpgradeFailed, message, message)
	master.Status.ClearCondition(postUpgradeFinished, message, message)
}

// Adds a component list object in state in progress. This will cause UpdateStatus to update the
// ready type to state "false".
func addUpgradeComponentNotReady(master *alpha1.CDAPMaster) {
	os := status.ObjectStatus{}
	os.Name = "upgrade"
	os.Status = status.StatusInProgress
	master.Status.ComponentMeta.ComponentList.Objects = append(master.Status.ComponentMeta.ComponentList.Objects, os)
}

func prepareNewPreUpgradeJobResources(master *alpha1.CDAPMaster, rsrclabels map[string]string) ([]reconciler.Object, error) {
	resetUpgradeConditions(master, upgradeStartMessage)

	upgradeResources, err := getPreUpgradeJobResources(master, rsrclabels)
	if err != nil {
		master.Status.Meta.SetCondition(upgradeFailed, upgradeFailedInitMessage, upgradeFailedInitMessage)
		return nil, err
	}
	return upgradeResources, nil
}

// Gets pre upgrade job resources
func getPreUpgradeJobResources(master *alpha1.CDAPMaster, rsrclabels map[string]string) ([]reconciler.Object, error) {
	return getUpgradeJobResources(master, rsrclabels, 0, getUpgradeJobName(master))
}

// Gets post upgrade job resources
func getPostUpgradeJobResources(master *alpha1.CDAPMaster, rsrclabels map[string]string) ([]reconciler.Object, error) {
	return getUpgradeJobResources(master, rsrclabels, master.Status.UpgradeStartTimeMillis, getPostUpgradeJobName(master))
}

func getVersion(a string) *version {
	av := strings.Split(a, ":")
	if len(av) == 1 || av[1] == latestVersion {
		return &version{isLatest: true}
	}
	return &version{isLatest: false, versionList: strings.Split(av[1], ".")}
}

// Represents an image version
type version struct {
	// Boolean if the version is "latest"
	isLatest bool
	// List of version integers, starting from major
	versionList []string
}

func compareVersion(versiona, versionb *version) int {
	if versiona.isLatest && versionb.isLatest {
		return 0
	} else if versiona.isLatest {
		return -1
	} else if versionb.isLatest {
		return 1
	}

	loopMax := len(versionb.versionList)
	if len(versiona.versionList) > len(versionb.versionList) {
		loopMax = len(versiona.versionList)
	}
	for i := 0; i < loopMax; i++ {
		var x, y string
		if len(versiona.versionList) > i {
			x = versiona.versionList[i]
		}
		if len(versionb.versionList) > i {
			y = versionb.versionList[i]
		}
		xi, _ := strconv.Atoi(x)
		yi, _ := strconv.Atoi(y)
		if xi > yi {
			return -1
		} else if xi < yi {
			return 1
		}
	}
	return 0
}

// Gets upgrade job resources
func getUpgradeJobResources(master *alpha1.CDAPMaster, rsrclabels map[string]string, startTimeMillis int64, name string) ([]reconciler.Object, error) {
	var spec = &upgradeJobSpec{
		Image:              master.Spec.Image,
		JobName:            name,
		HostName:           getResourceName(master, string(alpha1.ServiceRouter)),
		BackoffLimit:       upgradeFailureLimit,
		ReferentName:       master.Name,
		ReferentKind:       master.Kind,
		ReferentApiVersion: master.APIVersion,
		ReferentUID:        master.UID,
		SecuritySecret:     master.Spec.SecuritySecret,
		Namespace:          master.Namespace,
	}
	if startTimeMillis != 0 {
		spec.StartTimeMillis = startTimeMillis
	}
	ngdata := &upgradeValue{Job: spec}
	ngdata.Labels = rsrclabels
	ngdata.CConfName = getResourceName(master, "cconf")
	ngdata.HConfName = getResourceName(master, "hconf")
	item, err := k8s.ObjectFromFile(templateDir+upgradeJobTemplate, ngdata, &batchv1.JobList{})
	if err != nil {
		return nil, err
	}
	return append([]reconciler.Object{}, *item), nil
}

// Updates the component status
func updateStatus(rsrc interface{}, reconciled []reconciler.Object, err error) time.Duration {
	var period time.Duration
	stts := &rsrc.(*alpha1.CDAPMaster).Status
	ready := stts.ComponentMeta.UpdateStatus(reconciler.ObjectsByType(reconciled, k8s.Type))
	stts.Meta.UpdateStatus(&ready, err)
	return period
}

// ------------------------------ Handlers -------------------------------------

// Objects - handler Objects
func (b *Base) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	var resources []reconciler.Object
	master := rsrc.(*alpha1.CDAPMaster)

	labels := make(reconciler.KVMap)
	labels.Merge(master.Labels, rsrclabels)

	// Add the cdap and hadoop ConfigMap
	configs := map[string][]string{
		"cconf": {"cdap-site.xml", "logback.xml", "logback-container.xml"},
		"hconf": {"core-site.xml"},
	}
	var err error
	for k, v := range configs {
		resources, err = addConfigMapItem(master, getResourceName(master, k), labels, v, resources)
		if err != nil {
			return nil, err
		}
	}

	// If upgrade has failed, do not set imageToUse, just return
	if master.Status.Meta.IsConditionTrue(upgradeFailed) {
		// User has reset the CR after failure. Remove upgrade failed status.
		if master.Spec.Image == master.Status.ImageToUse {
			resetUpgradeConditions(master, upgradeResetMessage)
		}
		return resources, nil
	}

	// Downgrade case, we want to set the imageToUse to spec while not performing pre or post upgrade jobs
	if compareVersion(getVersion(master.Status.ImageToUse), getVersion(master.Spec.Image)) == -1 {
		master.Status.ImageToUse = master.Spec.Image
		master.Status.UserInterfaceImageToUse = master.Spec.UserInterfaceImage
		master.Status.SetCondition(postUpgradeFinished, upgradeJobSkippedMessage, upgradeJobSkippedMessage)
		return resources, nil
	}

	// Post upgrade case, in which the image has updated successfully and all pods are back up
	if master.Status.UpgradeStartTimeMillis != 0 && master.Status.IsReady() {
		var postUpgradeJob *batchv1.Job
		item := k8s.GetItem(observed, &batchv1.Job{}, getPostUpgradeJobName(master), master.Namespace)
		if item != nil {
			postUpgradeJob = item.(*batchv1.Job)
		}

		postUpgradeResources, err := getPostUpgradeJobResources(master, rsrclabels)
		if err != nil {
			return nil, err
		}
		resources = append(resources, postUpgradeResources...)

		if postUpgradeJob == nil {
			return resources, nil
		}

		if postUpgradeJob.Status.Succeeded > 0 {
			// Post upgrade job has succeeded.
			master.Status.UpgradeStartTimeMillis = 0
			master.Status.Meta.SetCondition(postUpgradeFinished, upgradeJobFinishedMessage, upgradeJobFinishedMessage)
		} else if postUpgradeJob.Status.Failed >= upgradeFailureLimit {
			// Post upgrade job has failed
			master.Status.UpgradeStartTimeMillis = 0
			master.Status.Meta.SetCondition(postUpgradeFailed, upgradeJobFailedMessage, upgradeJobFailedMessage)
		}
		return resources, nil
	}

	// Early return if image does not need updating. Update imageToUse since we are in a non
	// failure case.
	if master.Status.ImageToUse == "" || master.Status.ImageToUse == master.Spec.Image {
		// Currently the backend and UI are upgraded together, but this may not be the case in the future
		master.Status.ImageToUse = master.Spec.Image
		master.Status.UserInterfaceImageToUse = master.Spec.UserInterfaceImage
		return resources, nil
	}

	var upgradeResources []reconciler.Object
	var upgradeJob *batchv1.Job
	if master.Status.UpgradeStartTimeMillis != 0 {
		item := k8s.GetItem(observed, &batchv1.Job{}, getUpgradeJobName(master), master.Namespace)
		if item != nil {
			upgradeJob = item.(*batchv1.Job)
		}
	}

	if upgradeJob == nil {
		// Add a new job and set ready status to false
		master.Status.UpgradeStartTimeMillis =
			time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
		var err error
		upgradeResources, err = prepareNewPreUpgradeJobResources(master, rsrclabels)
		if err != nil {
			return nil, err
		}
		addUpgradeComponentNotReady(master)
		return append(resources, upgradeResources...), nil
	}

	if upgradeJob.Status.Succeeded > 0 {
		// Pre upgrade job has succeeded.
		master.Status.ImageToUse = master.Spec.Image
		master.Status.UserInterfaceImageToUse = master.Spec.UserInterfaceImage
		addUpgradeComponentNotReady(master)
	} else if upgradeJob.Status.Failed >= upgradeFailureLimit {
		// Pre upgrade job has failed
		master.Status.UpgradeStartTimeMillis = 0
		master.Status.Meta.SetCondition(upgradeFailed, upgradeJobFailedMessage, upgradeJobFailedMessage)
	} else {
		// Pre upgrade job has already been initialized, but has not finished
		var err error
		upgradeResources, err = getPreUpgradeJobResources(master, rsrclabels)
		if err != nil {
			return nil, err
		}
		addUpgradeComponentNotReady(master)
		resources = append(resources, upgradeResources...)
	}

	return resources, nil
}

// Observables for base
func (b *Base) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&corev1.ConfigMapList{}).
		For(&batchv1.JobList{}).
		Get()
}

// Objects for Messaging service
func (s *Messaging) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	return getStatefulServiceResources(&r.Spec.Messaging.CDAPStatefulServiceSpec, rsrc, rsrclabels, alpha1.ServiceMessaging)
}

// Observables for messaging
func (s *Messaging) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for AppFabric service
func (s *AppFabric) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	return getServiceResources(&r.Spec.AppFabric.CDAPServiceSpec, rsrc, rsrclabels, alpha1.ServiceAppFabric)
}

// Observables for appfabric
func (s *AppFabric) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		Get()
}

// Objects for Logs service
func (s *Logs) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	return getStatefulServiceResources(&r.Spec.Logs.CDAPStatefulServiceSpec, rsrc, rsrclabels, alpha1.ServiceLogs)
}

// Observables for logs
func (s *Logs) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for Metadata service
func (s *Metadata) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	return getServiceResources(&r.Spec.Metadata.CDAPServiceSpec, rsrc, rsrclabels, alpha1.ServiceMetadata)
}

// Observables for metadata
func (s *Metadata) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		Get()
}

// Objects for Metrics service
func (s *Metrics) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	return getStatefulServiceResources(&r.Spec.Metrics.CDAPStatefulServiceSpec, rsrc, rsrclabels, alpha1.ServiceMetrics)
}

// Observables for metrics
func (s *Metrics) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for Preview service
func (s *Preview) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	return getStatefulServiceResources(&r.Spec.Preview.CDAPStatefulServiceSpec, rsrc, rsrclabels, alpha1.ServicePreview)
}

// Observables for preview
func (s *Preview) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for Router service
func (s *Router) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	expected, err := getExternalServiceResources(&r.Spec.Router.CDAPExternalServiceSpec, rsrc, rsrclabels, alpha1.ServiceRouter, deploymentTemplate)
	setNodePort(expected, observed)
	return expected, err
}

// Observables for router
func (s *Router) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		Get()
}

// Objects for UserInterface service
func (s *UserInterface) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	r := rsrc.(*alpha1.CDAPMaster)
	expected, err := getExternalServiceResources(&r.Spec.UserInterface.CDAPExternalServiceSpec, rsrc, rsrclabels, alpha1.ServiceUserInterface, uiDeploymentTemplate)
	setNodePort(expected, observed)
	return expected, err
}

// Observables for UserInterface
func (s *UserInterface) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		Get()
}
