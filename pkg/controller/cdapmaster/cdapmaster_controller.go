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
	alpha1 "cdap.io/cdap-operator/pkg/apis/cdap/v1alpha1"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	gr "sigs.k8s.io/controller-reconciler/pkg/genericreconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
	"text/template"
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
)

// Add creates a new ESCluster Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
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
func setNodePort(s *alpha1.CDAPExternalServiceSpec, expected, observed []reconciler.Object) {
	// Get the service from the expected list.
	var expectedService *corev1.Service
	for _, item := range reconciler.ObjectsByType(expected, k8s.Type) {
		if service, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
			expectedService = service
			break
		}
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
	Env                map[string]string
}

// Sets values to the given templateBaseValue based on the resources provided
func (v *templateBaseValue) setTemplateValue(rsrc interface{}, rsrclabels map[string]string, serviceType alpha1.ServiceType, serviceLabels map[string]string, serviceAccount string, resources *corev1.ResourceRequirements) {
	master := rsrc.(*alpha1.CDAPMaster)
	labels := make(reconciler.KVMap)
	labels.Merge(master.Labels, serviceLabels, rsrclabels)

	// Set the cdap.container label. It is for service selector to route correctly
	name := fmt.Sprintf("cdap-%s-%s", master.Name, strings.ToLower(string(serviceType)))
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
			v.Env = make(map[string]string)
			v.Env["JAVA_HEAPMAX"] = fmt.Sprintf("-Xmx%v", xmx)
		}
	}

	v.Name = name
	v.Labels = labels
	v.Master = master
	v.ServiceAccountName = ServiceAccountName
	v.ServiceType = serviceType
	v.DataDir = alpha1.LocalDataDir
	v.CConfName = getConfigName(master, "cconf")
	v.HConfName = getConfigName(master, "hconf")
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

// Returns the config map name for the given configuration type
func getConfigName(r *alpha1.CDAPMaster, confType string) string {
	return fmt.Sprintf("cdap-%s-%s", r.Name, confType)
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
func addResourceItem(s *alpha1.CDAPServiceSpec, template string, v interface{}, listType metav1.ListInterface, resources []reconciler.Object) ([]reconciler.Object, error) {
	rinfo, err := k8s.ObjectFromFile(templateDir+template, v, listType)
	if err != nil {
		return nil, err
	}
	// Set the resource for the first container if the object has container
	setResources(rinfo.Obj.(*k8s.Object).Obj, s.Resources)
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
		resources, err = addConfigMapItem(master, getConfigName(master, k), labels, v, resources)
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

// Observables for base
func (b *Base) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&corev1.ConfigMapList{}).
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
	setNodePort(&r.Spec.Router.CDAPExternalServiceSpec, expected, observed)
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
	setNodePort(&r.Spec.UserInterface.CDAPExternalServiceSpec, expected, observed)
	return expected, err
}

// Observables for userinterface
func (s *UserInterface) Observables(rsrc interface{}, labels map[string]string) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		Get()
}
