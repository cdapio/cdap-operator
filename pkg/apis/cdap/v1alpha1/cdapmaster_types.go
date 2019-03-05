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

// CDAPMasterSpec defines the desired state of CDAPMaster
type CDAPMasterSpec struct {
	// Docker image name for the CDAP backend.
	Image string `json:"image,omitempty"`
	// Docker image name for the CDAP UI.
	UserInterfaceImage string `json:"userInterfaceImage,omitempty"`
	// Policy for pulling docker images on Pod creation.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Secret that contains security related configurations for CDAP.
	SecuritySecret string `json:"securitySecret,omitempty"`
	// The service account for all the service pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// An URI specifying an object storage for CDAP.
	LocationURI string `json:"locationURI"`
	// A set of configurations that goes into cdap-site.xml.
	Config map[string]string `json:"config,omitempty"`
	// A set of logger name to log level settings
	LogLevels map[string]string `json:"logLevels,omitempty"`
	// Specification for the CDAP app-fabric service
	AppFabric AppFabricSpec `json:"appFabric,omitempty"`
	// Specification for the CDAP logging service
	Logs LogsSpec `json:"logs,omitempty"`
	// Specification for the CDAP messaging service
	Messaging MessagingSpec `json:"messaging,omitempty"`
	// Specification for the CDAP metadata service
	Metadata MetadataSpec `json:"metadata,omitempty"`
	// Specification for the CDAP metrics service
	Metrics MetricsSpec `json:"metrics,omitempty"`
	// Specification for the CDAP preview service
	Preview PreviewSpec `json:"preview,omitempty"`
	// Specification for the CDAP router service
	Router RouterSpec `json:"router,omitempty"`
	// Specification for the CDAP UI service
	UserInterface UserInterfaceSpec `json:"userInterface,omitempty"`
}

// CDAPServiceSpec defines the base set of specifications applicable to all master services
type CDAPServiceSpec struct {
	// Metadata for the service.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Overrides the service account for the service pods
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Compute Resources required by the service.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// A selector which must be true for the pod to fit on a node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// CDAPScalableServiceSpec defines the base specification for master services that can have more than one instance
type CDAPScalableServiceSpec struct {
	CDAPServiceSpec `json:",inline"`
	// Number of replicas for the service.
	Replicas *int32 `json:"replicas,omitempty"`
}

// CDAPExternalServiceSpec defines the base specification for master services that expose to outside of the cluster
type CDAPExternalServiceSpec struct {
	CDAPScalableServiceSpec `json:",inline"`
	// The port number for the service.
	ServicePort *int32 `json:"servicePort,omitempty"`
}

// CDAPStatefulServiceSpec defines the base specification for stateful master services
type CDAPStatefulServiceSpec struct {
	CDAPServiceSpec `json:",inline"`
	// Specification for the persistent volumn size used by the service.
	StorageSize string `json:"storageSize,omitempty"`
}

// AppFabricSpec defines the specification for the AppFabric service
type AppFabricSpec struct {
	CDAPServiceSpec `json:",inline"`
}

// LogsSpec defines the specification for the Logs service
type LogsSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// MessagingSpec defines the specification for the TMS service
type MessagingSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// MetadataSpec defines the specification for the Metadata service
type MetadataSpec struct {
	CDAPServiceSpec `json:",inline"`
}

// MetricsSpec defines the specification for the Metrics service
type MetricsSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// PreviewSpec defines the specification for the Preview service
type PreviewSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// RouterSpec defines the specification for the Router service
type RouterSpec struct {
	CDAPExternalServiceSpec `json:",inline"`
}

// UserInterfaceSpec defines the specification for the UI service
type UserInterfaceSpec struct {
	CDAPExternalServiceSpec `json:",inline"`
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
