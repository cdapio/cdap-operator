package controllers

import (
	v1alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

// For ConfigMap
type ConfigMapSpec struct {
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Data      map[string]string `json:"configMap,omitempty"`
}

func newConfigMapSpec(name string, labels map[string]string, master *v1alpha1.CDAPMaster) *ConfigMapSpec {
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
}

func newContainerSpec(name, dataDir string, master *v1alpha1.CDAPMaster) *ContainerSpec {
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

// BaseSpec contains command fields for both StatefulSet and Deployment
type BaseSpec struct {
	Name               string            `json:"name,omitempty"`
	Namespace          string            `json:"namespace,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	Replicas           int32             `json:"replicas,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	RuntimeClassName   string            `json:"runtimeClassName,omitempty"`
	PriorityClassName  string            `json:"priorityClassName,omitempty"`
	SecuritySecret     string            `json:"securitySecret,omitempty"`
	CConf              string            `json:"cdapConf,omitempty"`
	HConf              string            `json:"hadoopConf,omitempty"`
}

func newBaseSpec(name string, labels map[string]string, cconf, hconf string, master *v1alpha1.CDAPMaster) *BaseSpec {
	s := new(BaseSpec)
	s.Name = name
	s.Namespace = master.Namespace
	s.Labels = labels
	s.ServiceAccountName = ""
	s.Replicas = 1
	s.RuntimeClassName = ""
	s.PriorityClassName = ""
	s.SecuritySecret = master.Spec.SecuritySecret
	s.CConf = cconf
	s.HConf = hconf
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

// For Deployment
type DeploymentSpec struct {
	Base       *BaseSpec        `json:"base,inline"`
	Containers []*ContainerSpec `json:"containers,omitempty"`
}

func newDeploymentSpec(name string, labels map[string]string, cconf, hconf string, master *v1alpha1.CDAPMaster) *DeploymentSpec {
	s := new(DeploymentSpec)
	s.Base = newBaseSpec(name, labels, cconf, hconf, master)
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

// For VolumnClaimTemplate in Statefulset
type StorageSpec struct {
	StorageClassName *string `json:"storageClassName,omitempty"`
	StorageSize      string  `json:"storageSize,omitempty"`
}

func newStorageSpec(storageClassName *string, storageSize string) *StorageSpec {
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

func newStatefulSpec(name string, labels map[string]string, cconf, hconf string, master *v1alpha1.CDAPMaster) *StatefulSpec {
	s := new(StatefulSpec)
	s.Base = newBaseSpec(name, labels, cconf, hconf, master)
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

func (s *StatefulSpec) withStorage(storageClassName *string, storageSize string) *StatefulSpec {
	s.Storage = newStorageSpec(storageClassName, storageSize)
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
type CDAPDeploymentSpec struct {
	Stateful        []*StatefulSpec       `json:"stateful,omitempty"`
	Deployment      []*DeploymentSpec     `json:"stateless,omitempty"`
	NetworkServices []*NetworkServiceSpec `json:"networkService,omitempty"`
}

func newCDAPDeploymentSpec() *CDAPDeploymentSpec {
	c := new(CDAPDeploymentSpec)
	return c
}
func (s *CDAPDeploymentSpec) withStateful(stateful *StatefulSpec) *CDAPDeploymentSpec {
	s.Stateful = append(s.Stateful, stateful)
	return s
}
func (s *CDAPDeploymentSpec) withDeployment(stateless *DeploymentSpec) *CDAPDeploymentSpec {
	s.Deployment = append(s.Deployment, stateless)
	return s
}
func (s *CDAPDeploymentSpec) withNetworkService(networkService *NetworkServiceSpec) *CDAPDeploymentSpec {
	s.NetworkServices = append(s.NetworkServices, networkService)
	return s
}

func (s *CDAPDeploymentSpec) toString() (string, error) {
	data, err := json.Marshal(*s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type UpgradeJobSpec struct {
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

func newUpgradeJobSpec(name string, labels map[string]string, startTimeMs int64, cconf, hconf string, master *v1alpha1.CDAPMaster) *UpgradeJobSpec {
	s := new(UpgradeJobSpec)
	s.Image = master.Spec.Image
	s.JobName = name
	s.Labels = labels
	s.HostName = getObjectName(master.Name, serviceRouter)
	s.BackoffLimit = imageVersionUpgradeFailureLimit
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

func (s *UpgradeJobSpec) SetPreUpgrade(isPreUpgrade bool) *UpgradeJobSpec {
	s.PreUpgrade = isPreUpgrade
	return s
}

func (s *UpgradeJobSpec) SetPostUpgrade(isPostUpgrade bool) *UpgradeJobSpec {
	s.PostUpgrade = isPostUpgrade
	return s
}
