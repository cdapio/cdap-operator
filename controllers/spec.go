package controllers

import (
	v1alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// For ConfigMap
type ConfigMapSpec struct {
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Data      map[string]string `json:"configMap,omitempty"`
}

func NewConfigMapSpec(name, namespace string, labels map[string]string) *ConfigMapSpec {
	s := new(ConfigMapSpec)
	s.Name = name
	s.Namespace = namespace
	s.Labels = labels
	s.Data = make(map[string]string)
	return s
}
func (s *ConfigMapSpec) WithData(key, val string) *ConfigMapSpec {
	s.Data[key] = val
	return s
}

// For containers in either StatefulSet or Deployment
type ContainerSpec struct {
	Name            string                       `json:"name,omitempty"`
	Image           string                       `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy            `json:"imagePullPolicy,omitempty"`
	ServiceMain     string                       `json:"serviceMain,omitempty"`
	Env             []corev1.EnvVar              `json:"env,omitempty"`
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
	DataDir         string                       `json:"string,omitempty"`
}

func NewContainerSpec(name, serviceMain string, master *v1alpha1.CDAPMaster, resources *corev1.ResourceRequirements, dataDir string) *ContainerSpec {
	c := new(ContainerSpec)
	c.Name = name
	c.Image = master.Spec.Image
	c.ImagePullPolicy = master.Spec.ImagePullPolicy
	c.ServiceMain = serviceMain
	c.Env = []corev1.EnvVar{} // TODO: to be populated
	c.Resources = resources
	c.DataDir = dataDir
	return c
}
func (s *ContainerSpec) SetImage(image string) *ContainerSpec {
	s.Image = image
	return s
}

// BaseSpec is common to both StatefulSet and Deployment
type BaseSpec struct {
	Name               string            `json:"name,omitempty"`
	Namespace          string            `json:"namespace,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	Replicas           int32             `json:"replicas,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	RuntimeClassName   *string           `json:"runtimeClassName,omitempty"`
	PriorityClassName  *string           `json:"priorityClassName,omitempty"`
	SecuritySecret     string            `json:"securitySecret,omitempty"`
	CConf              string            `json:"cdapConf,omitempty"`
	HConf              string            `json:"hadoopConf,omitempty"`
}

func NewBaseSpec(name string, replicas int32, labels map[string]string, serviceSpec *v1alpha1.CDAPServiceSpec, master *v1alpha1.CDAPMaster, cconf, hconf string) *BaseSpec {
	s := new(BaseSpec)
	s.Name = name
	s.Namespace = master.Namespace
	s.Labels = labels
	s.ServiceAccountName = master.Spec.ServiceAccountName
	s.Replicas = replicas
	s.NodeSelector = serviceSpec.NodeSelector
	s.RuntimeClassName = serviceSpec.RuntimeClassName
	s.PriorityClassName = serviceSpec.PriorityClassName
	s.SecuritySecret = master.Spec.SecuritySecret
	s.CConf = cconf
	s.HConf = hconf
	return s
}

// For Deployment
type DeploymentSpec struct {
	Base       *BaseSpec        `json:"base,inline"`
	Containers []*ContainerSpec `json:"containers,omitempty"`
}

func NewDeploymentSpec(name string, replicas int32, labels map[string]string, serviceSpec *v1alpha1.CDAPServiceSpec, master *v1alpha1.CDAPMaster, cconf, hconf string) *DeploymentSpec {
	s := new(DeploymentSpec)
	s.Base = NewBaseSpec(name, replicas, labels, serviceSpec, master, cconf, hconf)
	return s
}
func (s *DeploymentSpec) AddLabel(key, val string) *DeploymentSpec {
	s.Base.Labels = mergeLabels(s.Base.Labels, map[string]string{key: val})
	return s
}
func (s *DeploymentSpec) WithContainer(containerSpec *ContainerSpec) *DeploymentSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}

// For VolumnClaimTemplate in Statefulset
type StorageSpec struct {
	StorageClassName *string `json:"storageClassName,omitempty"`
	StorageSize      string  `json:"storageSize,omitempty"`
}

func NewStorageSpec(storageClassName *string, storageSize string) *StorageSpec {
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

func NewStatefulSpec(name string, replicas int32, labels map[string]string, serviceSpec *v1alpha1.CDAPServiceSpec, master *v1alpha1.CDAPMaster, cconf, hconf string) *StatefulSpec {
	s := new(StatefulSpec)
	s.Base = NewBaseSpec(name, replicas, labels, serviceSpec, master, cconf, hconf)
	return s
}

func (s *StatefulSpec) AddLabel(key, val string) *StatefulSpec {
	s.Base.Labels = mergeLabels(s.Base.Labels, map[string]string{key: val})
	return s
}

func (s *StatefulSpec) WithInitContainer(containerSpec *ContainerSpec) *StatefulSpec {
	s.InitContainers = append(s.InitContainers, containerSpec)
	return s
}

func (s *StatefulSpec) WithContainer(containerSpec *ContainerSpec) *StatefulSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}

func (s *StatefulSpec) WithStorage(storageClassName *string, storageSize string) *StatefulSpec {
	s.Storage = NewStorageSpec(storageClassName, storageSize)
	return s
}

// For k8s Service.
// Name it "NetworkService" to avoid confusion with CDAP service
type NetworkServiceSpec struct {
	Name        string             `json:"name,omitempty"`
	Namespace   string             `json:"namespace,omitempty"`
	Annotations *map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	ServiceType *string            `json:"serviceType,omitempty"`
	ServicePort *int32             `json:"servicePort,omitempty"`
}

func NewNetworkServiceSpec(name string, labels map[string]string, serviceType *string, port *int32, master *v1alpha1.CDAPMaster) *NetworkServiceSpec {
	s := new(NetworkServiceSpec)
	s.Name = name
	s.Namespace = master.Namespace
	// TODO: are annotations needed?
	s.Annotations = nil
	s.Labels = labels
	s.ServiceType = serviceType
	s.ServicePort = port
	return s
}
func (s *NetworkServiceSpec) AddLabel(key, val string) *NetworkServiceSpec {
	s.Labels = mergeLabels(s.Labels, map[string]string{key: val})
	return s
}

// For CDAP user interface Deployment
type UserInterfaceSpec struct {
	Base       *BaseSpec        `json:"base,inline"`
	Containers []*ContainerSpec `json:"containers,omitempty"`
}

func NewUserInterfaceSpec(name string, replicas int32, labels map[string]string, serviceSpec *v1alpha1.CDAPServiceSpec, master *v1alpha1.CDAPMaster, cconf, hconf string) *UserInterfaceSpec {
	s := new(UserInterfaceSpec)
	s.Base = NewBaseSpec(name, replicas, labels, serviceSpec, master, cconf, hconf)
	return s
}

func (s *UserInterfaceSpec) AddLabel(key, val string) *UserInterfaceSpec {
	s.Base.Labels = mergeLabels(s.Base.Labels, map[string]string{key: val})
	return s
}

func (s *UserInterfaceSpec) WithContainer(containerSpec *ContainerSpec) *UserInterfaceSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}

// Top level CDAP service deployment configuration
type CDAPDeploymentSpec struct {
	Stateful        []*StatefulSpec       `json:"stateful,omitempty"`
	Deployment      []*DeploymentSpec     `json:"stateless,omitempty"`
	NetworkServices []*NetworkServiceSpec `json:"networkService,omitempty"`
	UserInterface   *UserInterfaceSpec    `json:"uispec,omitempty"`
}

func NewCDAPDeploymentSpec() *CDAPDeploymentSpec {
	c := new(CDAPDeploymentSpec)
	return c
}
func (s *CDAPDeploymentSpec) WithStateful(stateful *StatefulSpec) *CDAPDeploymentSpec {
	s.Stateful = append(s.Stateful, stateful)
	return s
}
func (s *CDAPDeploymentSpec) WithDeployment(stateless *DeploymentSpec) *CDAPDeploymentSpec {
	s.Deployment = append(s.Deployment, stateless)
	return s
}
func (s *CDAPDeploymentSpec) WithNetworkService(networkService *NetworkServiceSpec) *CDAPDeploymentSpec {
	s.NetworkServices = append(s.NetworkServices, networkService)
	return s
}
func (s *CDAPDeploymentSpec) WithUserInterface(uiSpec *UserInterfaceSpec) *CDAPDeploymentSpec {
	s.UserInterface = uiSpec
	return s
}
