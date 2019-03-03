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
	"strings"

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
	spec := r.Spec
	if spec.Image == "" {
		spec.Image = defaultImage
	}
	if spec.UserInterfaceImage == "" {
		spec.UserInterfaceImage = defaultUserInterfaceImage
	}
}

func int32Ptr(value int32) *int32 {
	return &value
}

// HandleError records status or error in status
func (r *CDAPMaster) HandleError(err error) {
	logger.Error(err, "Error")
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

// Components returns components for this resource
func (r *CDAPMaster) Components() []component.Component {
	components := []component.Component{}

	// Add the master spect as a component
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

// serviceData carries value for templating
type serviceData struct {
	Name        string
	Master      *CDAPMaster
	Service     *CDAPMasterServiceSpec
	ServiceType string
	DataDir     string
	CConfName   string
	HConfName   string
}

// Creates a resource.Item based on the given CDAPMasterServiceSpec and template
func (s *CDAPMasterServiceSpec) createResourceItem(v interface{}, template string) (*resource.Item, error) {
	rinfo, err := k8s.ItemFromFile(templateDir+template, v, &appsv1.StatefulSetList{})
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

// ExpectedResources - returns resources for the CDAP master
func (s *CDAPMasterSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	return resources, nil
}

// ExpectedResources - returns resources for a cdap master service
func (s *CDAPMasterServiceSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	master := rsrc.(*CDAPMaster)

	// Get the template name and remove it from the spec
	template := s.Labels[templateLabel]
	delete(s.Labels, templateLabel)

	// Set the cdap.container label. It is for service selector to route correctly
	name := s.getServiceName(master)
	s.Labels[containerLabel] = name

	ngdata := serviceData{
		Name:        name,
		Master:      master,
		Service:     s,
		ServiceType: rsrclabels[component.LabelComponent],
		DataDir:     localDataDir,
		CConfName:   "cdap-conf",
		HConfName:   "hadoop-conf",
	}

	rinfo, err := s.createResourceItem(&ngdata, template)
	if err != nil {
		return nil, err
	}
	resources.Add(*rinfo)
	return resources, nil
}
