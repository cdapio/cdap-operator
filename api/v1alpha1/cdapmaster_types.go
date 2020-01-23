/*

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
	"sigs.k8s.io/controller-reconciler/pkg/status"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CDAPMasterSpec defines the desired state of CDAPMaster
type CDAPMasterSpec struct {
	// Image is the docker image name for the CDAP backend.
	Image string `json:"image,omitempty"`
	// UserInterfaceImage is the docker image name for the CDAP UI.
	UserInterfaceImage string `json:"userInterfaceImage,omitempty"`
	// ImagePullPolicy is the policy for pulling docker images on Pod creation.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// SecuritySecret is secret that contains security related configurations for CDAP.
	SecuritySecret string `json:"securitySecret,omitempty"`
	// ServiceAccountName is the service account for all the service pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// LocationURI is an URI specifying an object storage for CDAP.
	LocationURI string `json:"locationURI"`
	// Config is a set of configurations that goes into cdap-site.xml.
	Config map[string]string `json:"config,omitempty"`
	// LogLevels is a set of logger name to log level settings.
	LogLevels map[string]string `json:"logLevels,omitempty"`
	// AppFabric is specification for the CDAP app-fabric service.
	AppFabric AppFabricSpec `json:"appFabric,omitempty"`
	// Logs is specification for the CDAP logging service.
	Logs LogsSpec `json:"logs,omitempty"`
	// Messaging is specification for the CDAP messaging service.
	Messaging MessagingSpec `json:"messaging,omitempty"`
	// Metadata is specification for the CDAP metadata service.
	Metadata MetadataSpec `json:"metadata,omitempty"`
	// Metrics is specification for the CDAP metrics service.
	Metrics MetricsSpec `json:"metrics,omitempty"`
	// Preview is specification for the CDAP preview service.
	Preview PreviewSpec `json:"preview,omitempty"`
	// Router is specification for the CDAP router service.
	Router RouterSpec `json:"router,omitempty"`
	// UserInterface is specification for the CDAP UI service.
	UserInterface UserInterfaceSpec `json:"userInterface,omitempty"`
}

// CDAPServiceSpec defines the base set of specifications applicable to all master services.
type CDAPServiceSpec struct {
	// Metadata for the service.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// ServiceAccountName overrides the service account for the service pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Resources are Compute resources required by the service.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
	// to run pods for this service. If no RuntimeClass resource matches the named class, pods will not be running.
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`
	// PriorityClassName is to specify the priority of the pods for this service.
	PriorityClassName *string `json:"priorityClassName,omitempty"`
	// Env is a list of environment variables for the master service container.
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// CDAPScalableServiceSpec defines the base specification for master services that can have more than one instance.
type CDAPScalableServiceSpec struct {
	CDAPServiceSpec `json:",inline"`
	// Replicas is number of replicas for the service.
	Replicas *int32 `json:"replicas,omitempty"`
}

// CDAPExternalServiceSpec defines the base specification for master services that expose to outside of the cluster.
type CDAPExternalServiceSpec struct {
	CDAPScalableServiceSpec `json:",inline"`
	// ServiceType is the service type in kubernetes, default is NodePort.
	ServiceType *string `json:"serviceType,omitempty"`
	// ServicePort is the port number for the service.
	ServicePort *int32 `json:"servicePort,omitempty"`
}

// CDAPStatefulServiceSpec defines the base specification for stateful master services.
type CDAPStatefulServiceSpec struct {
	CDAPServiceSpec `json:",inline"`
	// StorageSize is specification for the persistent volume size used by the service.
	StorageSize string `json:"storageSize,omitempty"`
	// StorageClassName is the name of the StorageClass for the persistent volume used by the service.
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// AppFabricSpec defines the specification for the AppFabric service.
type AppFabricSpec struct {
	CDAPServiceSpec `json:",inline"`
}

// LogsSpec defines the specification for the Logs service.
type LogsSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// MessagingSpec defines the specification for the TMS service.
type MessagingSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// MetadataSpec defines the specification for the Metadata service
type MetadataSpec struct {
	CDAPServiceSpec `json:",inline"`
}

// MetricsSpec defines the specification for the Metrics service.
type MetricsSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// PreviewSpec defines the specification for the Preview service.
type PreviewSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// RouterSpec defines the specification for the Router service.
type RouterSpec struct {
	CDAPExternalServiceSpec `json:",inline"`
}

// UserInterfaceSpec defines the specification for the UI service.
type UserInterfaceSpec struct {
	CDAPExternalServiceSpec `json:",inline"`
}

// CDAPMasterStatus defines the observed state of CDAPMaster
type CDAPMasterStatus struct {
	status.Meta          `json:",inline"`
	status.ComponentMeta `json:",inline"`
	// ImageToUse is the Docker image of CDAP backend the operator uses to deploy.
	ImageToUse string `json:"imageToUse,omitempty"`
	// UserInterfaceImageToUse is the Docker image of CDAP UI the operator uses to deploy.
	UserInterfaceImageToUse string `json:"userInterfaceImageToUse,omitempty"`
	// UpgradeStartTimeMillis is the start time in millis of the upgrade process
	UpgradeStartTimeMillis int64 `json:"upgradeStartTimeMillis,omitempty"`
}

// +kubebuilder:object:root=true

// CDAPMaster is the Schema for the cdapmasters API
type CDAPMaster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CDAPMasterSpec   `json:"spec,omitempty"`
	Status CDAPMasterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CDAPMasterList contains a list of CDAPMaster
type CDAPMasterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CDAPMaster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CDAPMaster{}, &CDAPMasterList{})
}
