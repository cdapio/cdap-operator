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
//
// Important notes:
// * The field name of each service MUST match the constant values of ServiceName in constants.go as reflection
//   is used to find field value.
// * For services that are optional (i.e. may or may not be required for CDAP to be operational), their service
//   specification fields are pointers. By default, these optional services are disabled. Set to non-nil to enable them.
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
	// Env is a list of environment variables for the all service containers.
	Env []corev1.EnvVar `json:"env,omitempty"`
	// LocationURI is an URI specifying an object storage for CDAP.
	LocationURI string `json:"locationURI"`
	// Config is a set of configurations that goes into cdap-site.xml.
	Config map[string]string `json:"config,omitempty"`
	// ConfigMapVolumes defines a map from ConfigMap names to volume mount path.
	// Key is the configmap object name. Value is the mount path.
	// This adds ConfigMap data to the directory specified by the volume mount path.
	ConfigMapVolumes map[string]string `json:"configMapVolumes,omitempty"`
	// SecretVolumes defines a map from Secret names to volume mount path.
	// Key is the secret object name. Value is the mount path.
	// This adds Secret data to the directory specified by the volume mount path.
	SecretVolumes map[string]string `json:"secretVolumes,omitempty"`
	// SystemAppConfigs specifies configs used by CDAP to run system apps
	// dynamically. Each entry is of format <filename, json app config> which will
	// create a separate system config file with entry value as file content.
	SystemAppConfigs map[string]string `json:"systemappconfigs,omitempty"`
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
	// SupportBundle is specification for the CDAP support-bundle service.
	// This is an optional service and may not be required for CDAP to be operational.
	// To disable this service: either omit or set the field to nil
	// To enable this service: set it to a pointer to a SupportBundleSpec struct (can be an empty struct)
	SupportBundle *SupportBundleSpec `json:"supportBundle,omitempty"`
	// TetheringAgent is specification for the CDAP Tethering Agent service.
	// This is an optional service and may not be required for CDAP to be operational.
	// To disable this service: either omit or set the field to nil
	// To enable this service: set it to a pointer to a TetheringAgentSpec struct (can be an empty struct)
	TetheringAgent *TetheringAgentSpec `json:"tetheringAgent,omitempty"`
	// ArtifactCache is specification for the CDAP Artifact Cache service.
	// This is an optional service and may not be required for CDAP to be operational.
	// To disable this service: either omit or set the field to nil
	// To enable this service: set it to a pointer to a ArtifactCacheSpec struct (can be an empty struct)
	ArtifactCache *ArtifactCacheSpec `json:"artifactCache,omitempty"`
	// Runtime is specification for the CDAP runtime service.
	// This is an optional service and may not be required for CDAP to be operational.
	// To disable this service: either omit or set the field to nil
	// To enable this service: set it to a pointer to a RuntimeSpec struct (can be an empty struct)
	Runtime *RuntimeSpec `json:"runtime,omitempty"`
	// Auth is specification for the CDAP Auth service.
	// This is an optional service and may not be required for CDAP to be operational.
	// To disable this service: either omit or set the field to nil
	// To enable this service: set it to a pointer to a AuthenticationSpec struct (can be an empty struct)
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// SystemMetricsExporter is specification for the CDAP SystemMetricsExporter service.
	// This is an optional service and may not be required for CDAP to be operational.
	// To disable this service: either omit or set the field to nil
	// To enable this service: set it to a pointer to a SystemMetricsExporterSpec struct (can be an empty struct).
	// CDAPServiceSpec.EnableSystemMetrics field also needs to be set to true for stateful services which require
	// collection of system metrics. Services which have CDAPServiceSpec.EnableSystemMetrics as nil, missing or set to false,
	// will have metrics sidecar container disabled.
	SystemMetricsExporter *SystemMetricExporterSpec `json:"systemMetricsExporter,omitempty"`
	// SecurityContext defines the security context for all pods for all services.
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	// AdditionalVolumes defines a list of additional volumes for all services.
	// For information on supported volume types, see https://kubernetes.io/docs/concepts/storage/volumes/.
	AdditionalVolumes []corev1.Volume `json:"additionalVolumes,omitempty"`
	// AdditionalVolumeMounts defines a list of additional volume mounts for all services.
	// For information on suported volume mount types, see https://kubernetes.io/docs/concepts/storage/volumes/.
	AdditionalVolumeMounts []corev1.VolumeMount `json:"additionalVolumeMounts,omitempty"`
}

// CDAPServiceSpec defines the base set of specifications applicable to all master services.
//
// If the name of structure needs to be changed, update the code where it uses reflect to find this field.
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
	// ConfigMapVolumes defines a map from ConfigMap names to volume mount path for this service
	// Key is the configmap object name. Value is the mount path.
	// This adds ConfigMap data to the directory specified by the volume mount path.
	ConfigMapVolumes map[string]string `json:"configMapVolumes,omitempty"`
	// SecretVolumes defines a map from Secret names to volume mount path for this service
	// Key is the secret object name. Value is the mount path.
	// This adds Secret data to the directory specified by the volume mount path.
	SecretVolumes map[string]string `json:"secretVolumes,omitempty"`
	// AdditionalVolumes defines a list of additional volumes to mount to the service.
	// For information on supported volume types, see https://kubernetes.io/docs/concepts/storage/volumes/.
	AdditionalVolumes []corev1.Volume `json:"additionalVolumes,omitempty"`
	// AdditionalVolumeMounts defines a list of additional volume mounts for the service.
	// For information on suported volume mount types, see https://kubernetes.io/docs/concepts/storage/volumes/.
	AdditionalVolumeMounts []corev1.VolumeMount `json:"additionalVolumeMounts,omitempty"`
	// SecurityContext overrides the security context for the service pods.
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	// EnableSystemMetrics is an optional field that is considered along with CDAPMasterSpec.SystemMetricsExporter
	// to start a metrics collection container in statefulsets. SystemMetricsExporter is a global setting in CDAPMasterSpec.
	// When SystemMetricsExporter is absent, it disables metrics collection for all stateful services.
	// When SystemMetricsExporter is present, this value should also be set to true for services which require system metrics
	// collection.
	EnableSystemMetrics *bool `json:"enableSystemMetrics,omitempty"`
	// Lifecycle is to specify Container Lifecycle hooks provided by Kubernetes for containers.
	// This will not be applied to the init containers as init containers do not support lifecycle.
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty"`
}

// CDAPScalableServiceSpec defines the base specification for master services that can have more than one instance.
type CDAPScalableServiceSpec struct {
	CDAPServiceSpec `json:",inline"`
	// Replicas is number of replicas for the service.
	Replicas *int32 `json:"replicas,omitempty"`
}

// CDAPExternalServiceSpec defines the base specification for master services that expose to outside of the cluster.
//
// If the name of structure needs to be changed, update the code where it uses reflect to find this field.
type CDAPExternalServiceSpec struct {
	CDAPScalableServiceSpec `json:",inline"`
	// ServiceType is the service type in kubernetes, default is NodePort.
	ServiceType *string `json:"serviceType,omitempty"`
	// ServicePort is the port number for the service.
	ServicePort *int32 `json:"servicePort,omitempty"`
}

// CDAPStatefulServiceSpec defines the base specification for stateful master services.
//
// If the name of structure needs to be changed, update the code where it uses reflect to find this field.
type CDAPStatefulServiceSpec struct {
	CDAPServiceSpec `json:",inline"`
	// StorageSize is specification for the persistent volume size used by the service.
	StorageSize string `json:"storageSize,omitempty"`
	// StorageClassName is the name of the StorageClass for the persistent volume used by the service.
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// CDAPScalableStatefulServiceSpec defines the base specification for stateful master services that can scale to more than one instance.
//
// If the name of structure needs to be changed, update the code where it uses reflect to find this field.
type CDAPScalableStatefulServiceSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
	// Replicas is number of replicas for the service.
	Replicas *int32 `json:"replicas,omitempty"`
}

// AppFabricSpec defines the specification for the AppFabric service.
type AppFabricSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
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
	CDAPScalableServiceSpec `json:",inline"`
}

// MetricsSpec defines the specification for the Metrics service.
type MetricsSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// PreviewSpec defines the specification for the Preview service.
type PreviewSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// RuntimeSpec defines the specification for the Runtime service.
type RuntimeSpec struct {
	CDAPScalableStatefulServiceSpec `json:",inline"`
}

// AuthenticationSpec defines the specification for the Authentication service.
type AuthenticationSpec struct {
	CDAPScalableServiceSpec `json:",inline"`
}

// RouterSpec defines the specification for the Router service.
type RouterSpec struct {
	CDAPExternalServiceSpec `json:",inline"`
}

// UserInterfaceSpec defines the specification for the UI service.
type UserInterfaceSpec struct {
	CDAPExternalServiceSpec `json:",inline"`
}

// SupportBundleSpec defines the specification for the SupportBundle service.
type SupportBundleSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// TetheringAgentSpec defines the specification for the TetheringAgent service.
type TetheringAgentSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

// ArtifactCacheSpec defines the specification for the ArtifactCache service.
type ArtifactCacheSpec struct {
	CDAPStatefulServiceSpec `json:",inline"`
}

//  SystemMetricExporterSpec defines the specification for the SystemMetricsExporter service.
type SystemMetricExporterSpec struct {
	CDAPServiceSpec `json:",inline"`
}

// CDAPMasterStatus defines the observed state of CDAPMaster
type CDAPMasterStatus struct {
	status.Meta          `json:",inline"`
	status.ComponentMeta `json:",inline"`
	// ImageToUse is the Docker image of CDAP backend the operator uses to deploy.
	ImageToUse string `json:"imageToUse,omitempty"`
	// UserInterfaceImageToUse is the Docker image of CDAP UI the operator uses to deploy.
	UserInterfaceImageToUse string `json:"userInterfaceImageToUse,omitempty"`
	// UpgradeStartTimeMillis is the start time in milliseconds of the upgrade process
	UpgradeStartTimeMillis int64 `json:"upgradeStartTimeMillis,omitempty"`
	// DowngradeStartTimeMillis is the start time in milliseconds of the downgrade process
	DowngradeStartTimeMillis int64 `json:"downgradeStartTimeMillis,omitempty"`
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

// SecurityContext defines fields for setting corev1.SecurityContext for containers and
// corev1.PodSecurityContext for pods.
// For additional information, see https://kubernetes.io/docs/tasks/configure-pod-container/security-context/.
type SecurityContext struct {
	// RunAsUser runs the pod as the specified user ID. It is applied at the pod level.
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	// RunAsGroup runs the pod as the specified group ID. It is applied at the pod level.
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
	// FSGroup mounts volumes as the specified group ID and gives the primary user access
	// to that group. It is applied at the pod level.
	FSGroup *int64 `json:"fsGroup,omitempty"`
	// AllowPrivilegeEscalation prevents the container process from running SUID binaries.
	// It is applied at the container level.
	AllowPrivilegeEscalation *bool `json:"allowPrivilegeEscalation,omitempty"`
	// RunAsNonRoot indicates that the container must run as a non-root user.
	// If true, the Kubelet will validate the image at runtime to ensure that it
	// does not run as UID 0 (root) and fail to start the container if it does.
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`
	// Privileged runs container in privileged mode. It is applied at the container level.
	// Processes in privileged containers are essentially equivalent to root on the host.
	Privileged *bool `json:"privileged,omitempty"`
	// ReadOnlyRootFilesystem specifies whether the container's root filesystem is read-only.
	// It is applied at the container level.
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`
}

func init() {
	SchemeBuilder.Register(&CDAPMaster{}, &CDAPMasterList{})
}
