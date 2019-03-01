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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceType is the name identifying various CDAP master services
type ServiceType string

const (
	// AppFabric defines the service type for app-fabric
	AppFabric ServiceType = "AppFabric"

	// Log defines the service type for log processing and serving service
	Log ServiceType = "Log"

	// Messaging defines the service type for TMS
	Messaging ServiceType = "Messaging"

	// Metadata defines the service type for metadata service
	Metadata ServiceType = "Metadata"

	// Metrics defines the service type for metrics process and serving
	Metrics ServiceType = "Metrics"

	// Router defines the service type for the router
	Router ServiceType = "Router"

	// UserInterface defines the service type for user interface
	UserInterface ServiceType = "UserInterface"
)

// CDAPMasterSpec defines the desired state of CDAPMaster
type CDAPMasterSpec struct {
	// Docker image name for the CDAP backend.
	Image string `json:"image,omitempty"`
	// Docker image name for the CDAP UI.
	UserInterfaceImage string `json:"userInterfaceImage,omitempty"`
	// Policy for pulling docker images on Pod creation.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// List of specifications for customizing each individual CDAP master service.
	Services []CDAPMasterServiceSpec `json:"services,omitempty"`
	// Secret that contains security related configurations for CDAP.
	SecuritySecret *corev1.SecretReference `json:"securitySecret,omitempty"`
	// An URI specifying an object storage for CDAP.
	LocationURI string `json:"locationURI"`
	// A set of configurations that goes into cdap-site.xml.
	Config map[string]string `json:"config,omitempty"`
}

// CDAPMasterServiceSpec defines specification for one CDAP master service
type CDAPMasterServiceSpec struct {
	// Metadata for the service.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The ServiceType that this specification is applying to
	Type ServiceType `json:"type"`
	// Number of replicas for the service.
	// The value will be ignored for services that doesn't support more than one replica
	Replicas *int32 `json:"replicas,omitempty"`
	// Compute Resources required by the service.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Specification for the persistent volumn used by the service.
	// The spec will be ignored for services that doesn't use persistent volume
	VolumeSpec *corev1.PersistentVolumeClaimSpec `json:"volumeSpec,omitempty"`
	// A selector which must be true for the pod to fit on a node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// CDAPMasterStatus defines the observed state of CDAPMaster
type CDAPMasterStatus struct {
	ObservedGeneration   int64  `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	RouterService        string `json:"routerService,omitempty"`
	UserInterfaceService string `json:"userInterfaceService,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CDAPMaster is the Schema for the cdapmasters API
// +k8s:openapi-gen=true
type CDAPMaster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CDAPMasterSpec   `json:"spec,omitempty"`
	Status CDAPMasterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CDAPMasterList contains a list of CDAPMaster
type CDAPMasterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CDAPMaster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CDAPMaster{}, &CDAPMasterList{})
}
