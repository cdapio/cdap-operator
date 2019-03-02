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
	"strings"

	appsv1 "k8s.io/api/apps/v1"

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
	templatePath              = "templates/"
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

func (r *CDAPMaster) serviceComponent(s *CDAPMasterServiceSpec, serviceType ServiceType, maxReplicas int32, hasStorage bool) component.Component {
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
	// Remove the cdap.container label, as it is set through the template via service Name
	delete(s.Labels, containerLabel)
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
	components = append(components, r.serviceComponent(&r.Spec.Messaging, Messaging, 1, true))
	components = append(components, r.serviceComponent(&r.Spec.Metrics, Metrics, 1, true))
	components = append(components, r.serviceComponent(&r.Spec.Preview, Preview, 1, true))

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

type serviceData struct {
	Name        string
	Master      *CDAPMaster
	Service     *CDAPMasterServiceSpec
	ServiceType string
	DataDir     string
	CConfName   string
	HConfName   string
}

func (s *CDAPMasterServiceSpec) sts(v interface{}) (*resource.Item, error) {
	return k8s.ItemFromFile(templatePath+"cdap-sts.yaml", v, &appsv1.StatefulSetList{})
}

// ExpectedResources - returns resources for a cdap master service
func (s *CDAPMasterServiceSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	master := rsrc.(*CDAPMaster)
	ngdata := serviceData{
		Name:        s.getServiceName(master),
		Master:      master,
		Service:     s,
		ServiceType: rsrclabels[component.LabelComponent],
		DataDir:     localDataDir,
		CConfName:   "cdap-conf",
		HConfName:   "hadoop-conf",
	}

	for _, fn := range []resource.GetItemFn{s.sts} {
		rinfo, err := fn(&ngdata)
		if err != nil {
			return nil, err
		}
		sts := rinfo.Obj.(*k8s.Object).Obj.(*appsv1.StatefulSet)
		if s.Resources != nil {
			sts.Spec.Template.Spec.Containers[0].Resources = *s.Resources
		}

		resources.Add(*rinfo)
	}

	return resources, nil
}
