package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"cdap.io/cdap-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

// For ConfigMap
type ConfigMapSpec struct {
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Data      map[string]string `json:"configMap,omitempty"`
}

func newConfigMapSpec(master *v1alpha1.CDAPMaster, name string, labels map[string]string) *ConfigMapSpec {
	s := new(ConfigMapSpec)
	s.Name = name
	s.Namespace = master.Namespace
	s.Labels = labels
	s.Data = make(map[string]string)
	return s
}

func (s *ConfigMapSpec) AddData(key, val string) *ConfigMapSpec {
	s.Data[key] = val
	return s
}

// For containers in either StatefulSet or Deployment
type ContainerSpec struct {
	Name             string                        `json:"name,omitempty"`
	Image            string                        `json:"image,omitempty"`
	ImagePullPolicy  corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	WorkingDir       string                        `json:"workingDir,omitempty"`
	Command          []string                      `json:"command,omitempty"`
	Args             []string                      `json:"args,omitempty"`
	Env              []corev1.EnvVar               `json:"env,omitempty"`
	ResourceRequests map[string]*resource.Quantity `json:"resourceRequests,omitempty"`
	ResourceLimits   map[string]*resource.Quantity `json:"resourceLimits,omitempty"`
	DataDir          string                        `json:"dataDir,omitempty"`
	Lifecycle        *corev1.Lifecycle             `json:"lifecycle,omitempty"`
	Ports            []corev1.ContainerPort        `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort" protobuf:"bytes,6,rep,name=ports"`
	LivenessProbe    *corev1.Probe                 `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
	ReadinessProbe   *corev1.Probe                 `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
}

func containerSpecFromContainer(container *corev1.Container, dataDir string) *ContainerSpec {
	additionalContainer := new(ContainerSpec)
	additionalContainer.Name = strings.ToLower(container.Name)
	additionalContainer.Image = container.Image
	additionalContainer.ImagePullPolicy = container.ImagePullPolicy
	additionalContainer.WorkingDir = container.WorkingDir
	additionalContainer.Args = container.Args
	additionalContainer.Env = container.Env
	additionalContainer.DataDir = dataDir
	additionalContainer.LivenessProbe = container.LivenessProbe
	additionalContainer.ReadinessProbe = container.ReadinessProbe
	additionalContainer.Ports = container.Ports

	return additionalContainer
}

func newContainerSpec(master *v1alpha1.CDAPMaster, name, dataDir string) *ContainerSpec {
	c := new(ContainerSpec)
	c.Name = strings.ToLower(name)
	c.Image = master.Status.ImageToUse
	c.ImagePullPolicy = master.Spec.ImagePullPolicy
	c.WorkingDir = ""
	c.Args = []string{"io.cdap.cdap.master.environment.k8s." + name + "ServiceMain", "--env=k8s"}
	c.Env = []corev1.EnvVar{}
	c.DataDir = dataDir
	return c
}
func (s *ContainerSpec) setImage(image string) *ContainerSpec {
	s.Image = image
	return s
}
func (s *ContainerSpec) setWorkingDir(workingDir string) *ContainerSpec {
	s.WorkingDir = workingDir
	return s
}
func (s *ContainerSpec) setCommand(command ...string) *ContainerSpec {
	s.Command = []string{}
	for _, c := range command {
		s.Command = append(s.Command, c)
	}
	return s
}

func (s *ContainerSpec) setArgs(arg ...string) *ContainerSpec {
	s.Args = []string{}
	for _, a := range arg {
		s.Args = append(s.Args, a)
	}
	return s
}

func (s *ContainerSpec) addEnv(name, value string) *ContainerSpec {
	envVar := corev1.EnvVar{
		Name:      name,
		Value:     value,
		ValueFrom: nil,
	}
	s.Env = append(s.Env, envVar)
	return s
}

func (s *ContainerSpec) setEnv(envVar []corev1.EnvVar) *ContainerSpec {
	s.Env = envVar
	return s
}

func (s *ContainerSpec) setResources(resources *corev1.ResourceRequirements) *ContainerSpec {
	if resources == nil {
		return s
	}
	if len(resources.Requests) > 0 {
		s.ResourceRequests = make(map[string]*resource.Quantity)
		for name, quantity := range resources.Requests {
			q := new(resource.Quantity)
			*q = quantity.DeepCopy()
			s.ResourceRequests[string(name)] = q
		}
	}
	if len(resources.Limits) > 0 {
		s.ResourceLimits = make(map[string]*resource.Quantity)
		for name, quantity := range resources.Limits {
			q := new(resource.Quantity)
			*q = quantity.DeepCopy()
			s.ResourceLimits[string(name)] = q
		}
	}
	return s
}

func (s *ContainerSpec) setLifecycle(lifecycle *corev1.Lifecycle) *ContainerSpec {
	s.Lifecycle = lifecycle
	return s
}

// BaseSpec contains command fields for both StatefulSet and Deployment
type BaseSpec struct {
	Name                   string                    `json:"name,omitempty"`
	Namespace              string                    `json:"namespace,omitempty"`
	Labels                 map[string]string         `json:"labels,omitempty"`
	ServiceAccountName     string                    `json:"serviceAccountName,omitempty"`
	Replicas               int32                     `json:"replicas,omitempty"`
	NodeSelector           map[string]string         `json:"nodeSelector,omitempty"`
	RuntimeClassName       string                    `json:"runtimeClassName,omitempty"`
	PriorityClassName      string                    `json:"priorityClassName,omitempty"`
	SecuritySecret         string                    `json:"securitySecret,omitempty"`
	SecuritySecretPath     string                    `json:"securitySecretPath,omitempty"`
	CConf                  string                    `json:"cdapConf,omitempty"`
	HConf                  string                    `json:"hadoopConf,omitempty"`
	SysAppConf             string                    `json:"sysAppConf,omitempty"`
	ConfigMapVolumes       map[string]string         `json:"configMapVolumes,omitempty"`
	SecretVolumes          map[string]string         `json:"secretVolumes,omitempty"`
	AdditionalVolumes      []corev1.Volume           `json:"additionalVolumes,omitempty"`
	AdditionalVolumeMounts []corev1.VolumeMount      `json:"additionalVolumeMounts,omitempty"`
	SecurityContext        *v1alpha1.SecurityContext `json:"securityContext,omitempty"`
}

func newBaseSpec(master *v1alpha1.CDAPMaster, name string, labels map[string]string, cconf, hconf, sysappconf string) *BaseSpec {
	s := new(BaseSpec)
	s.Name = name
	s.Namespace = master.Namespace
	s.Labels = cloneMap(labels)
	s.ServiceAccountName = ""
	s.Replicas = 1
	s.RuntimeClassName = ""
	s.PriorityClassName = ""
	s.SecuritySecret = master.Spec.SecuritySecret
	s.SecuritySecretPath = defaultSecuritySecretPath
	s.CConf = cconf
	s.HConf = hconf
	s.SysAppConf = sysappconf
	s.ConfigMapVolumes = cloneMap(master.Spec.ConfigMapVolumes)
	s.SecretVolumes = cloneMap(master.Spec.SecretVolumes)
	s.AdditionalVolumes = master.Spec.AdditionalVolumes
	s.AdditionalVolumeMounts = master.Spec.AdditionalVolumeMounts
	return s
}

func (s *BaseSpec) setServiceAccountName(name string) *BaseSpec {
	s.ServiceAccountName = name
	return s
}

func (s *BaseSpec) setReplicas(replicas int32) *BaseSpec {
	s.Replicas = replicas
	return s
}

func (s *BaseSpec) setNodeSelector(nodeSelector map[string]string) *BaseSpec {
	if len(nodeSelector) == 0 {
		return s
	}
	s.NodeSelector = make(map[string]string)
	for k, v := range nodeSelector {
		s.NodeSelector[k] = v
	}
	return s
}

func (s *BaseSpec) setRuntimeClassName(name string) *BaseSpec {
	s.RuntimeClassName = name
	return s
}

func (s *BaseSpec) setPriorityClassName(name string) *BaseSpec {
	s.PriorityClassName = name
	return s
}

func (s *BaseSpec) addConfigMapVolumes(volumes map[string]string) (*BaseSpec, error) {
	if err := addVolumes(s.ConfigMapVolumes, volumes, "ConfigMap"); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *BaseSpec) addSecretVolumes(volumes map[string]string) (*BaseSpec, error) {
	if err := addVolumes(s.SecretVolumes, volumes, "Secret"); err != nil {
		return nil, err
	}
	return s, nil
}

func addVolumes(volumes, newVolumes map[string]string, typeName string) error {
	for k, v := range newVolumes {
		if val, exists := volumes[k]; exists {
			if val != v {
				return fmt.Errorf("failed to mount %s volume %v to %v due to already mounted to %v", typeName, k, v, val)
			}
		} else {
			volumes[k] = v
		}
	}
	return nil
}

func (s *BaseSpec) addAdditionalVolumes(additionalVolumes []corev1.Volume) (*BaseSpec, error) {
	for _, additionalVolume := range additionalVolumes {
		for _, specVolume := range s.AdditionalVolumes {
			if specVolume.Name == additionalVolume.Name {
				return nil, fmt.Errorf("failed to add custom volume %q due to already mounted as %+v", additionalVolume.Name, specVolume)
			}
		}
		s.AdditionalVolumes = append(s.AdditionalVolumes, additionalVolume)
	}
	return s, nil
}

func (s *BaseSpec) addAdditionalVolumeMounts(additionalVolumeMounts []corev1.VolumeMount) (*BaseSpec, error) {
	for _, additionalVolumeMount := range additionalVolumeMounts {
		for _, specVolumeMount := range s.AdditionalVolumeMounts {
			if specVolumeMount.Name == additionalVolumeMount.Name {
				return nil, fmt.Errorf("failed to mount custom volume %q at path %q due to already mounted as %+v", additionalVolumeMount.Name, additionalVolumeMount.MountPath, specVolumeMount)
			}
		}
		s.AdditionalVolumeMounts = append(s.AdditionalVolumeMounts, additionalVolumeMount)
	}
	return s, nil
}

func (s *BaseSpec) setSecurityContext(securityContext *v1alpha1.SecurityContext) *BaseSpec {
	s.SecurityContext = securityContext
	if securityContext != nil {
		// For non-boolean pointers, it is reasonable to set this value only if it is non-zero. Thus, we
		// do not need to set a default for the int64 pointers.
		// For boolean pointers, we always set the value and ensure there is a default set for nil values.
		// All defaults are least-restrictive and pulled from corev1.SecurityContext preserve backwards compatibility.
		if s.SecurityContext.RunAsNonRoot == nil {
			s.SecurityContext.RunAsNonRoot = new(bool)
			*s.SecurityContext.RunAsNonRoot = false
		}
		if s.SecurityContext.Privileged == nil {
			s.SecurityContext.Privileged = new(bool)
			*s.SecurityContext.Privileged = false
		}
		if s.SecurityContext.AllowPrivilegeEscalation == nil {
			s.SecurityContext.AllowPrivilegeEscalation = new(bool)
			*s.SecurityContext.AllowPrivilegeEscalation = true
		}
		if s.SecurityContext.ReadOnlyRootFilesystem == nil {
			s.SecurityContext.ReadOnlyRootFilesystem = new(bool)
			*s.SecurityContext.ReadOnlyRootFilesystem = false
		}
	}
	return s
}

// For Deployment
type DeploymentSpec struct {
	Base       *BaseSpec        `json:"base,inline"`
	Containers []*ContainerSpec `json:"containers,omitempty"`
}

func newDeploymentSpec(master *v1alpha1.CDAPMaster, name string, labels map[string]string, cconf, hconf, sysappconf string) *DeploymentSpec {
	s := new(DeploymentSpec)
	s.Base = newBaseSpec(master, name, labels, cconf, hconf, sysappconf)
	return s
}

func (s *DeploymentSpec) setServiceAccountName(name string) *DeploymentSpec {
	s.Base.setServiceAccountName(name)
	return s
}

func (s *DeploymentSpec) setReplicas(replicas int32) *DeploymentSpec {
	s.Base.setReplicas(replicas)
	return s
}

func (s *DeploymentSpec) setNodeSelector(nodeSelector map[string]string) *DeploymentSpec {
	s.Base.setNodeSelector(nodeSelector)
	return s
}

func (s *DeploymentSpec) setRuntimeClassName(name string) *DeploymentSpec {
	s.Base.setRuntimeClassName(name)
	return s
}

func (s *DeploymentSpec) setPriorityClassName(name string) *DeploymentSpec {
	s.Base.setPriorityClassName(name)
	return s
}

func (s *DeploymentSpec) addLabel(key, val string) *DeploymentSpec {
	s.Base.Labels = mergeMaps(s.Base.Labels, map[string]string{key: val})
	return s
}

func (s *DeploymentSpec) withContainer(containerSpec *ContainerSpec) *DeploymentSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}

func (s *DeploymentSpec) addConfigMapVolumes(volumes map[string]string) (*DeploymentSpec, error) {
	if _, err := s.Base.addConfigMapVolumes(volumes); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *DeploymentSpec) addSecretVolumes(volumes map[string]string) (*DeploymentSpec, error) {
	if _, err := s.Base.addSecretVolumes(volumes); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *DeploymentSpec) addAdditionalVolumes(additionalVolumes []corev1.Volume) (*DeploymentSpec, error) {
	if _, err := s.Base.addAdditionalVolumes(additionalVolumes); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *DeploymentSpec) addAdditionalVolumeMounts(additionalVolumeMounts []corev1.VolumeMount) (*DeploymentSpec, error) {
	if _, err := s.Base.addAdditionalVolumeMounts(additionalVolumeMounts); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *DeploymentSpec) setSecurityContext(securityContext *v1alpha1.SecurityContext) *DeploymentSpec {
	s.Base.setSecurityContext(securityContext)
	return s
}

// For VolumnClaimTemplate in Statefulset
type StorageSpec struct {
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageSize      string `json:"storageSize,omitempty"`
}

func newStorageSpec(storageClassName string, storageSize string) *StorageSpec {
	s := new(StorageSpec)
	s.StorageClassName = storageClassName
	s.StorageSize = storageSize
	return s
}

// For StatefulSet
type StatefulSpec struct {
	Base           *BaseSpec        `json:"Base,inline"`
	InitContainers []*ContainerSpec `json:"initContainer,omitempty"`
	Containers     []*ContainerSpec `json:"containers,omitempty"`
	Storage        *StorageSpec     `json:"storage,omitempty"`
}

func newStatefulSpec(master *v1alpha1.CDAPMaster, name string, labels map[string]string, cconf, hconf, sysappconf string) *StatefulSpec {
	s := new(StatefulSpec)
	s.Base = newBaseSpec(master, name, labels, cconf, hconf, sysappconf)
	return s
}

func (s *StatefulSpec) addLabel(key, val string) *StatefulSpec {
	s.Base.Labels = mergeMaps(s.Base.Labels, map[string]string{key: val})
	return s
}

func (s *StatefulSpec) setServiceAccountName(name string) *StatefulSpec {
	s.Base.setServiceAccountName(name)
	return s
}

func (s *StatefulSpec) setReplicas(replicas int32) *StatefulSpec {
	s.Base.setReplicas(replicas)
	return s
}

func (s *StatefulSpec) setNodeSelector(nodeSelector map[string]string) *StatefulSpec {
	s.Base.setNodeSelector(nodeSelector)
	return s
}

func (s *StatefulSpec) setRuntimeClassName(name string) *StatefulSpec {
	s.Base.setRuntimeClassName(name)
	return s
}

func (s *StatefulSpec) setPriorityClassName(name string) *StatefulSpec {
	s.Base.setPriorityClassName(name)
	return s
}

func (s *StatefulSpec) withInitContainer(containerSpec *ContainerSpec) *StatefulSpec {
	s.InitContainers = append(s.InitContainers, containerSpec)
	return s
}

func (s *StatefulSpec) withContainer(containerSpec *ContainerSpec) *StatefulSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}

func (s *StatefulSpec) withStorage(storageClassName string, storageSize string) *StatefulSpec {
	s.Storage = newStorageSpec(storageClassName, storageSize)
	return s
}

func (s *StatefulSpec) addConfigMapVolumes(volumes map[string]string) (*StatefulSpec, error) {
	if _, err := s.Base.addConfigMapVolumes(volumes); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StatefulSpec) addSecretVolumes(volumes map[string]string) (*StatefulSpec, error) {
	if _, err := s.Base.addSecretVolumes(volumes); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StatefulSpec) addAdditionalVolumes(additionalVolumes []corev1.Volume) (*StatefulSpec, error) {
	if _, err := s.Base.addAdditionalVolumes(additionalVolumes); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StatefulSpec) addAdditionalVolumeMounts(additionalVolumeMounts []corev1.VolumeMount) (*StatefulSpec, error) {
	if _, err := s.Base.addAdditionalVolumeMounts(additionalVolumeMounts); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StatefulSpec) setSecurityContext(securityContext *v1alpha1.SecurityContext) *StatefulSpec {
	s.Base.setSecurityContext(securityContext)
	return s
}

type NetworkServiceSpec struct {
	Name        string             `json:"name,omitempty"`
	Namespace   string             `json:"namespace,omitempty"`
	Annotations *map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	Selectors   map[string]string  `json:"selectors,omitempty"`
	ServiceType *string            `json:"serviceType,omitempty"`
	ServicePort *int32             `json:"servicePort,omitempty"`
}

func newNetworkServiceSpec(name string, labels map[string]string, serviceType *string, port *int32, master *v1alpha1.CDAPMaster) *NetworkServiceSpec {
	s := new(NetworkServiceSpec)
	s.Name = name
	s.Namespace = master.Namespace
	// TODO: are annotations needed?
	s.Annotations = nil
	s.Labels = mergeMaps(labels, map[string]string{})
	s.Selectors = mergeMaps(labels, map[string]string{})
	s.ServiceType = serviceType
	s.ServicePort = port
	return s
}
func (s *NetworkServiceSpec) addLabel(key, val string) *NetworkServiceSpec {
	s.Labels = mergeMaps(s.Labels, map[string]string{key: val})
	return s
}
func (s *NetworkServiceSpec) addSelector(key, val string) *NetworkServiceSpec {
	s.Selectors = mergeMaps(s.Selectors, map[string]string{key: val})
	return s
}

// Top level CDAP service deployment configuration
type DeploymentPlanSpec struct {
	Stateful        []*StatefulSpec       `json:"stateful,omitempty"`
	Deployment      []*DeploymentSpec     `json:"stateless,omitempty"`
	NetworkServices []*NetworkServiceSpec `json:"networkService,omitempty"`
}

func newDeploymentPlanSpec() *DeploymentPlanSpec {
	c := new(DeploymentPlanSpec)
	return c
}
func (s *DeploymentPlanSpec) withStateful(stateful *StatefulSpec) *DeploymentPlanSpec {
	s.Stateful = append(s.Stateful, stateful)
	return s
}
func (s *DeploymentPlanSpec) withDeployment(stateless *DeploymentSpec) *DeploymentPlanSpec {
	s.Deployment = append(s.Deployment, stateless)
	return s
}
func (s *DeploymentPlanSpec) withNetworkService(networkService *NetworkServiceSpec) *DeploymentPlanSpec {
	s.NetworkServices = append(s.NetworkServices, networkService)
	return s
}

func (s *DeploymentPlanSpec) toString() (string, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type VersionUpgradeJobSpec struct {
	Image              string            `json:"image,omitempty"`
	JobName            string            `json:"jobName,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
	HostName           string            `json:"hostName,omitempty"`
	BackoffLimit       int32             `json:"backoffLimit,omitempty"`
	ReferentName       string            `json:"referentName,omitempty"`
	ReferentKind       string            `json:"referentKind,omitempty"`
	ReferentApiVersion string            `json:"referentApiVersion,omitempty"`
	ReferentUID        types.UID         `json:"referentUID,omitempty"`
	SecuritySecret     string            `json:"securitySecret,omitempty"`
	StartTimeMs        int64             `json:"startTimeMs,omitempty"`
	Namespace          string            `json:"namespace,omitempty"`
	CConf              string            `json:"cdapConf,omitempty"`
	HConf              string            `json:"hadoopConf,omitempty"`
	PreUpgrade         bool              `json:"preUpgrade,omitempty"`
	PostUpgrade        bool              `json:"postUpgrade,omitempty"`
}

func newUpgradeJobSpec(master *v1alpha1.CDAPMaster, name string, labels map[string]string, startTimeMs int64, cconf, hconf string) *VersionUpgradeJobSpec {
	s := new(VersionUpgradeJobSpec)
	s.Image = master.Spec.Image
	s.JobName = name
	s.Labels = labels
	s.HostName = getObjName(master, serviceRouter)
	s.BackoffLimit = imageVersionUpgradeJobMaxRetryCount
	s.ReferentName = master.Name
	s.ReferentKind = master.Kind
	s.ReferentApiVersion = master.APIVersion
	s.ReferentUID = master.UID
	s.SecuritySecret = master.Spec.SecuritySecret
	s.Namespace = master.Namespace
	s.StartTimeMs = startTimeMs
	s.CConf = cconf
	s.HConf = hconf
	return s
}

func (s *VersionUpgradeJobSpec) SetPreUpgrade(isPreUpgrade bool) *VersionUpgradeJobSpec {
	s.PreUpgrade = isPreUpgrade
	return s
}

func (s *VersionUpgradeJobSpec) SetPostUpgrade(isPostUpgrade bool) *VersionUpgradeJobSpec {
	s.PostUpgrade = isPostUpgrade
	return s
}
