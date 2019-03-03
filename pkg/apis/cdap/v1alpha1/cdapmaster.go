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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/kubesdk/pkg/component"
	"sigs.k8s.io/kubesdk/pkg/resource"
	"sigs.k8s.io/kubesdk/pkg/resource/manager/k8s"
)

var logger = logf.Log.WithName("cdap.controller")

const (
	// Property key in cdap-site.xml for configuring local data directory
	localDataDirKey = "local.data.dir"
	// Value for the local data directory
	localDataDir              = "/data"
	instanceLabel             = "cdap.instance"
	containerLabel            = "cdap.container"
	templateLabel             = ".cdap.template"
	templateDir               = "templates/"
	deploymentTemplate        = "cdap-deployment.yaml"
	uiDeploymentTemplate      = "cdap-ui-deployment.yaml"
	statefulSetTemplate       = "cdap-sts.yaml"
	defaultImage              = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.0.0-SNAPSHOT"
	defaultUserInterfaceImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion-ui:6.0.0-SNAPSHOT"
)

// ApplyDefaults will default missing values from the CDAPMaster
func (r *CDAPMaster) ApplyDefaults() {
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[instanceLabel] = r.Name

	spec := r.Spec
	if spec.Image == "" {
		spec.Image = defaultImage
	}
	if spec.UserInterfaceImage == "" {
		spec.UserInterfaceImage = defaultUserInterfaceImage
	}
}

// HandleError records status or error in status
func (r *CDAPMaster) HandleError(err error) {
	logger.Error(err, "Error: "+err.Error())
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
	components = append(components, r.serviceComponent(&r.Spec.AppFabric, AppFabric, 1, false, deploymentTemplate))
	// TODO: uncomment log when it is ready
	// components = append(components, r.serviceComponent(&r.Spec.Log, Log, 1, true, statefulSetTemplate))
	components = append(components, r.serviceComponent(&r.Spec.Messaging, Messaging, 1, true, statefulSetTemplate))
	components = append(components, r.serviceComponent(&r.Spec.Metadata, Metadata, 4, false, deploymentTemplate))
	components = append(components, r.serviceComponent(&r.Spec.Metrics, Metrics, 1, true, statefulSetTemplate))
	components = append(components, r.serviceComponent(&r.Spec.Preview, Preview, 1, true, statefulSetTemplate))
	components = append(components, r.serviceComponent(&r.Spec.Router, Router, 10, false, deploymentTemplate))
	components = append(components, r.serviceComponent(&r.Spec.UserInterface, UserInterface, 10, false, uiDeploymentTemplate))

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

func (s *CDAPMasterServiceSpec) getServiceName(r *CDAPMaster) string {
	return fmt.Sprintf("cdap-%s-%s", r.Name, s.Name)
}

func (r *CDAPMaster) getConfigName(confType string) string {
	return fmt.Sprintf("cdap-%s-%s", r.Name, confType)
}

// templateData carries value for templating
type templateData struct {
	Name        string
	Labels      map[string]string
	Master      *CDAPMaster
	Service     *CDAPMasterServiceSpec
	ServiceType string
	DataDir     string
	CConfName   string
	HConfName   string
}

// ExpectedResources - returns resources for the CDAP master
func (s *CDAPMasterSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	master := rsrc.(*CDAPMaster)

	// Add the cdap and hadoop ConfigMap
	configs := map[string][]string{
		"cconf": []string{"cdap-site.xml", "logback.xml", "logback-container.xml"},
		"hconf": []string{"core-site.xml"},
	}

	for k, v := range configs {
		rinfo, err := master.createConfigMapItem(master.getConfigName(k), v)
		if err != nil {
			return nil, err
		}
		resources.Add(*rinfo)
	}

	return resources, nil
}

// ExpectedResources - returns resources for a cdap master service
func (s *CDAPMasterServiceSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	master := rsrc.(*CDAPMaster)

	// Get the template name and remove it from the spec
	template := s.Labels[templateLabel]
	delete(s.Labels, templateLabel)

	labels := make(map[string]string)
	for k, v := range master.Labels {
		labels[k] = v
	}
	for k, v := range s.Labels {
		labels[k] = v
	}

	// Set the cdap.container label. It is for service selector to route correctly
	name := s.getServiceName(master)
	labels[containerLabel] = name

	ngdata := templateData{
		Name:        name,
		Labels:      labels,
		Master:      master,
		Service:     s,
		ServiceType: rsrclabels[component.LabelComponent],
		DataDir:     localDataDir,
		CConfName:   master.getConfigName("cconf"),
		HConfName:   master.getConfigName("hconf"),
	}

	rinfo, err := s.createServiceItem(&ngdata, template)
	if err != nil {
		return nil, err
	}
	resources.Add(*rinfo)
	return resources, nil
}

func int32Ptr(value int32) *int32 {
	return &value
}

// Creates a component.Component representing the given CDAP master service
func (r *CDAPMaster) serviceComponent(s *CDAPMasterServiceSpec, serviceType ServiceType, maxReplicas int32, hasStorage bool, template string) component.Component {
	s.Name = strings.ToLower(string(serviceType))
	if s.Replicas == nil {
		s.Replicas = int32Ptr(1)
	}
	if *s.Replicas > maxReplicas {
		s.Replicas = &maxReplicas
	}
	if hasStorage && s.StorageSize == "" {
		s.StorageSize = "50Gi"
	}
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	// Set the template to use via label. It will be removed in the ExpectedResources method
	s.Labels[templateLabel] = template

	return component.Component{
		Handle:   s,
		Name:     string(serviceType),
		CR:       r,
		OwnerRef: r.OwnerRef(),
	}
}

func getListType(tmpl string) (metav1.ListInterface, error) {
	switch t := tmpl; t {
	case deploymentTemplate:
		return &appsv1.DeploymentList{}, nil
	case uiDeploymentTemplate:
		return &appsv1.DeploymentList{}, nil
	case statefulSetTemplate:
		return &appsv1.StatefulSetList{}, nil
	default:
		return nil, errors.New("Unsupported template type " + tmpl)
	}
}

func (r *CDAPMaster) createConfigMapItem(name string, templates []string) (*resource.Item, error) {
	// Creates the configMap object
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Data: make(map[string]string),
	}

	ngdata := templateData{
		Master: r,
	}

	// Load template files that goes into config map
	for _, tmplFile := range templates {
		content, err := logbackFromFile(tmplFile, &ngdata)
		if err != nil {
			return nil, err
		}
		configMap.Data[tmplFile] = content
	}

	return &resource.Item{
		Type:      k8s.Type,
		Lifecycle: resource.LifecycleManaged,
		Obj: &k8s.Object{
			Obj:     configMap.DeepCopyObject().(metav1.Object),
			ObjList: &corev1.ConfigMapList{},
		},
	}, nil
}

// Creates a resource.Item based on the given CDAPMasterServiceSpec and template
func (s *CDAPMasterServiceSpec) createServiceItem(v interface{}, template string) (*resource.Item, error) {
	listType, err := getListType(template)
	if err != nil {
		return nil, err
	}

	rinfo, err := k8s.ItemFromFile(templateDir+template, v, listType)
	if err != nil {
		return nil, err
	}
	// Set resource to the first container if needed
	if s.Resources != nil {
		setResources(rinfo.Obj.(*k8s.Object).Obj, s.Resources)
	}
	return rinfo, err
}

// Sets resources to the given object. It uses reflection to find and set the
// field `Spec.Template.Spec.Containers[0].Resources`
func setResources(obj interface{}, resources *corev1.ResourceRequirements) {
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
