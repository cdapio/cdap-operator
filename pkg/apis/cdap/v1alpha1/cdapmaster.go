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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kubesdk/pkg/component"
	"sigs.k8s.io/kubesdk/pkg/resource"
	"sigs.k8s.io/kubesdk/pkg/resource/manager/k8s"
)

const (
	// TemplatePath is the directory for storing the templates
	TemplatePath = "templates/"
)

// HandleError records status or error in status
func (r *CDAPMaster) HandleError(err error) {
}

// Components returns components for this resource
func (r *CDAPMaster) Components() []component.Component {
	return []component.Component{
		{
			Handle:   &r.Spec,
			Name:     "cdapmaster",
			CR:       r,
			OwnerRef: r.OwnerRef(),
		},
	}
}

// OwnerRef returns owner ref object with the component's resource as owner
func (r *CDAPMaster) OwnerRef() *metav1.OwnerReference {
	return metav1.NewControllerRef(r, schema.GroupVersionKind{
		Group:   SchemeGroupVersion.Group,
		Version: SchemeGroupVersion.Version,
		Kind:    "CDAPMaster",
	})
}

// TemplateValues contains values for template
type TemplateValues struct {
	Name      string
	Namespace string
	Master    *CDAPMaster
}

func (s *CDAPMasterSpec) sts(v interface{}) (*resource.Item, error) {
	return k8s.ItemFromFile(TemplatePath+"cdap-sts.yaml", v, &appsv1.StatefulSetList{})
}

// ExpectedResources - returns resources
func (s *CDAPMasterSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	r := rsrc.(*CDAPMaster)

	var ngdata = TemplateValues{
		Name:      r.Name,
		Namespace: r.Namespace,
		Master:    r,
	}

	for _, fn := range []resource.GetItemFn{s.sts} {
		rinfo, err := fn(&ngdata)
		if err != nil {
			return nil, err
		}
		resources.Add(*rinfo)
	}

	return resources, nil
}
