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

package v1alpha1

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kubesdk/pkg/component"
	"sigs.k8s.io/kubesdk/pkg/resource"
	"sigs.k8s.io/kubesdk/pkg/resource/manager/k8s"
)

// ApplyDefaults will default missing values from the CDAPMaster
func (r *CDAPMaster) ApplyDefaults() {
	r.Status.EnsureStandardConditions()
	r.Status.ResetComponentList()

	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[instanceLabel] = r.Name

	spec := &r.Spec
	if spec.Image == "" {
		spec.Image = defaultImage
	}
	if spec.UserInterfaceImage == "" {
		spec.UserInterfaceImage = defaultUserInterfaceImage
	}

	if spec.Router.ServicePort == nil {
		spec.Router.ServicePort = int32Ptr(defaultRouterPort)
	}
	if spec.UserInterface.ServicePort == nil {
		spec.UserInterface.ServicePort = int32Ptr(defaultUserInterfacePort)
	}

	if spec.Config == nil {
		spec.Config = make(map[string]string)
	}
	// Set the local data directory
	spec.Config[confLocalDataDirKey] = localDataDir

	// Set the cconf entry for the router and UI service and ports
	spec.Config[confRouterServerAddress] = fmt.Sprintf("cdap-%s-%s", r.Name, strings.ToLower(string(Router)))
	spec.Config[confRouterBindPort] = strconv.Itoa(int(*spec.Router.ServicePort))
	spec.Config[confUserInterfaceBindPort] = strconv.Itoa(int(*spec.UserInterface.ServicePort))

	// Disable explore
	spec.Config[confExploreEnabled] = "false"
}

// HandleError records status or error in status
func (r *CDAPMaster) HandleError(err error) {
	if err != nil {
		r.Status.SetError("ErrorSeen", err.Error())
	} else {
		r.Status.ClearError()
	}
}

// Components returns components for this resource
func (r *CDAPMaster) Components() []component.Component {
	components := []component.Component{}

	// Add the master spec as a component
	components = append(components, component.Component{
		Handle:   &r.Spec,
		Name:     r.Name,
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	// Create components for each of the CDAP services
	// There is no requirement on the start order, but try to put more essential one first
	components = append(components, component.Component{
		Handle:   &r.Spec.Messaging,
		Name:     string(Messaging),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.AppFabric,
		Name:     string(AppFabric),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.Metrics,
		Name:     string(Metrics),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.Logs,
		Name:     string(Logs),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.Metadata,
		Name:     string(Metadata),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.Preview,
		Name:     string(Preview),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.Router,
		Name:     string(Router),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})
	components = append(components, component.Component{
		Handle:   &r.Spec.UserInterface,
		Name:     string(UserInterface),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	})

	return components
}

// OwnerRef returns owner ref object with the component's resource as owner
func (r *CDAPMaster) OwnerRef() *metav1.OwnerReference {
	return metav1.NewControllerRef(r, schema.GroupVersionKind{
		Group:   SchemeGroupVersion.Group,
		Version: SchemeGroupVersion.Version,
		Kind:    "CDAPMaster",
	})
}

// ExpectedResources - returns resources for the CDAP master
func (s *CDAPMasterSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	master := rsrc.(*CDAPMaster)

	labels := make(component.KVMap)
	labels.Merge(master.Labels, rsrclabels)

	// Add the cdap and hadoop ConfigMap
	configs := map[string][]string{
		"cconf": []string{"cdap-site.xml", "logback.xml", "logback-container.xml"},
		"hconf": []string{"core-site.xml"},
	}

	for k, v := range configs {
		_, err := master.addConfigMapItem(master.getConfigName(k), labels, v, resources)
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

// ExpectedResources for AppFabric service
func (s *AppFabricSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getServiceResources(rsrc, rsrclabels, AppFabric)
}

// ExpectedResources for Logs service
func (s *LogsSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getStatefulServiceResources(rsrc, rsrclabels, Logs)
}

// ExpectedResources for Messaging service
func (s *MessagingSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getStatefulServiceResources(rsrc, rsrclabels, Messaging)
}

// ExpectedResources for Metadata service
func (s *MetadataSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getServiceResources(rsrc, rsrclabels, Metadata)
}

// ExpectedResources for Metrics service
func (s *MetricsSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getStatefulServiceResources(rsrc, rsrclabels, Metrics)
}

// ExpectedResources for Preview service
func (s *PreviewSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getStatefulServiceResources(rsrc, rsrclabels, Preview)
}

// ExpectedResources for Router service
func (s *RouterSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getExternalServiceResources(rsrc, rsrclabels, Router, deploymentTemplate)
}

// ExpectedResources for UserInterface service
func (s *UserInterfaceSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	return s.getExternalServiceResources(rsrc, rsrclabels, UserInterface, uiDeploymentTemplate)
}

// Mutate for Router service. This is needed to fix the nodePort
func (s *RouterSpec) Mutate(rsrc interface{}, rsrclabels map[string]string, expected, dependent, observed *resource.Bag) (*resource.Bag, error) {
	s.setNodePort(expected, observed)
	return expected, nil
}

// Mutate for UserInterface service. This is needed to fix the nodePort
func (s *UserInterfaceSpec) Mutate(rsrc interface{}, rsrclabels map[string]string, expected, dependent, observed *resource.Bag) (*resource.Bag, error) {
	s.setNodePort(expected, observed)
	return expected, nil
}

func (s *CDAPExternalServiceSpec) setNodePort(expected, observed *resource.Bag) {
	// Get the service from the expected list.
	var expectedService *corev1.Service
	for _, item := range expected.ByType(k8s.Type) {
		if service, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
			expectedService = service
			break
		}
	}
	// Find the service being observed. Extract nodePort from the service spec and set it to expected
	for _, item := range observed.ByType(k8s.Type) {
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

// Struct containing data for templatization
type templateBaseValue struct {
	Name               string
	Labels             map[string]string
	Master             *CDAPMaster
	Replicas           *int32
	ServiceAccountName string
	ServiceType        ServiceType
	DataDir            string
	CConfName          string
	HConfName          string
}

// Struct containing data for templatization using CDAPServiceSpec
type serviceValue struct {
	templateBaseValue
	Service *CDAPServiceSpec
}

// Struct containing data for templatization using CDAPStatefulServiceSpec
type statefulServiceValue struct {
	templateBaseValue
	Service *CDAPStatefulServiceSpec
}

// Struct containing data for templatization using CDAPExternalServiceSpec
type externalServiceValue struct {
	templateBaseValue
	Service *CDAPExternalServiceSpec
}

// Returns the config map name for the given configuration type
func (r *CDAPMaster) getConfigName(confType string) string {
	return fmt.Sprintf("cdap-%s-%s", r.Name, confType)
}

// Sets values to the given templateBaseValue based on the resources provided
func (v *templateBaseValue) setTemplateValue(rsrc interface{}, rsrclabels map[string]string, serviceType ServiceType, serviceLabels map[string]string, serviceAccount string) {
	master := rsrc.(*CDAPMaster)
	labels := make(component.KVMap)
	labels.Merge(master.Labels, serviceLabels, rsrclabels)

	// Set the cdap.container label. It is for service selector to route correctly
	name := fmt.Sprintf("cdap-%s-%s", master.Name, strings.ToLower(string(serviceType)))
	labels[containerLabel] = name

	ServiceAccountName := master.Spec.ServiceAccountName
	if serviceAccount != "" {
		ServiceAccountName = serviceAccount
	}

	v.Name = name
	v.Labels = labels
	v.Master = master
	v.ServiceAccountName = ServiceAccountName
	v.ServiceType = serviceType
	v.DataDir = localDataDir
	v.CConfName = master.getConfigName("cconf")
	v.HConfName = master.getConfigName("hconf")
}

// Gets the set of resources for the given service represented by the CDAPServiceSpec.
// It consists of a Deployment for the given serviceType
func (s *CDAPServiceSpec) getServiceResources(rsrc interface{}, rsrclabels map[string]string, serviceType ServiceType) (*resource.Bag, error) {
	ngdata := &serviceValue{
		Service: s,
	}
	ngdata.setTemplateValue(rsrc, rsrclabels, serviceType, s.Labels, s.ServiceAccountName)
	return s.addResourceItem(deploymentTemplate, ngdata, &appsv1.DeploymentList{}, new(resource.Bag))
}

// Gets the set of resources for the given service represented by the CDAPStatefulServiceSpec
// It consists of a StatefulSet for the given serviceType
func (s *CDAPStatefulServiceSpec) getStatefulServiceResources(rsrc interface{}, rsrclabels map[string]string, serviceType ServiceType) (*resource.Bag, error) {
	ngdata := &statefulServiceValue{
		Service: s,
	}
	ngdata.setTemplateValue(rsrc, rsrclabels, serviceType, s.Labels, s.ServiceAccountName)
	return s.addResourceItem(statefulSetTemplate, ngdata, &appsv1.StatefulSetList{}, new(resource.Bag))
}

// Gets the set of resources for the given service represented by the CDAPExternalServiceSpec
// It consists of a Deployment and a NodePort Service for the given serviceType
func (s *CDAPExternalServiceSpec) getExternalServiceResources(rsrc interface{}, rsrclabels map[string]string, serviceType ServiceType, template string) (*resource.Bag, error) {
	ngdata := &externalServiceValue{
		Service: s,
	}
	ngdata.Replicas = s.Replicas
	ngdata.setTemplateValue(rsrc, rsrclabels, serviceType, s.Labels, s.ServiceAccountName)

	resources, err := s.addResourceItem(template, ngdata, &appsv1.DeploymentList{}, new(resource.Bag))
	if err != nil {
		return nil, err
	}
	resources, err = s.addResourceItem(serviceTemplate, ngdata, &corev1.ServiceList{}, resources)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// Adds a resource.Item to the given resource.Bag by executing the given template.
func (s *CDAPServiceSpec) addResourceItem(template string, v interface{}, listType metav1.ListInterface, resources *resource.Bag) (*resource.Bag, error) {
	rinfo, err := k8s.ItemFromFile(templateDir+template, v, listType)
	if err != nil {
		return nil, err
	}
	// Set the resource for the first container if the object has container
	setResources(rinfo.Obj.(*k8s.Object).Obj, s.Resources)
	resources.Add(*rinfo)
	return resources, nil
}

// Creates a int32 pointer for the given value
func int32Ptr(value int32) *int32 {
	return &value
}

// Adds a resource.Item of ConfigMap type to the given resource.Bag
func (r *CDAPMaster) addConfigMapItem(name string, labels map[string]string, templates []string, resources *resource.Bag) (*resource.Bag, error) {
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

	resources.Add(resource.Item{
		Type:      k8s.Type,
		Lifecycle: resource.LifecycleManaged,
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
