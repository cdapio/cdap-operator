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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CDAPMasterSpec defines the desired state of CDAPMaster
type CDAPMasterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image           string                  `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy       `json:"imagePullPolicy,omitempty"`
	Services        []CDAPMasterService     `json:"services"`
	SecuritySecret  *corev1.SecretReference `json:"securitySecret,omitempty"`
	LocationURI     string                  `json:"locationURI"`
	Config          map[string]string       `json:"config,omitempty"`
}

// CDAPMasterService defines specification for one CDAP master service
type CDAPMasterService struct {
	Type         ServiceType                       `json:"type"`
	Instances    *int32                            `json:"instances"`
	Resources    *corev1.ResourceRequirements      `json:"resources,omitempty"`
	VolumeSpec   *corev1.PersistentVolumeClaimSpec `json:"volumeSpec,omitempty"`
	NodeSelector map[string]string                 `json:"nodeSelector,omitempty"`
}

// CDAPMasterStatus defines the observed state of CDAPMaster
type CDAPMasterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

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
