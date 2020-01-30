package controllers

import (
	v1alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type ContainerSpec struct {
	Name            string                       `json:"name,omitempty"`
	Image           string                       `json:"image,omitempty"`
	ServiceName     v1alpha1.ServiceType         `json:"servcieName,omitempty"`
	Env             []corev1.EnvVar              `json:"env,omitempty"`
	ImagePullPolicy corev1.PullPolicy            `json:"imagePullPolicy,omitempty"`
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
	SecuritySecret  string                       `json:"securitySecret,omitempty"`
}

func NewContainerSpec(serviceName v1alpha1.ServiceType, master *v1alpha1.CDAPMaster, serviceSpec *v1alpha1.CDAPServiceSpec) *ContainerSpec {
	c := new(ContainerSpec)
	c.Name = string(serviceName)
	c.Image = master.Spec.Image
	c.ServiceName = serviceName
	c.Env = []corev1.EnvVar{} // TODO(wyzhang):
	c.ImagePullPolicy = master.Spec.ImagePullPolicy
	c.Resources = serviceSpec.Resources
	c.SecuritySecret = master.Spec.SecuritySecret
	return c
}

type BaseSpec struct {
	Name               string             `json:"name,omitempty"`
	Labels             map[string]string  `json:"labels,omitempty"`
	Replicas           *int32             `json:"replicas,omitempty"`
	ServiceAccountName string             `json:"serviceAccountName,omitempty"`
	NodeSelector       *map[string]string `json:"nodeSelector,omitempty"`
	RuntimeClassName   string             `json:"runtimeClassName,omitempty"`
	PriorityClassName  string             `json:"priorityClassName,omitempty"`
	CdapConf           string             `json:"cdapConf,omitempty"`
	HadoopConf         string             `json:"hadoopConf,omitempty"`
}

func NewBaseSpec(name string, replicas *int32, nodeSelector *map[string]string, runtimeClass string, priorityClassName string, master *v1alpha1.CDAPMaster, cconf, hconf string) *BaseSpec {
	s := new(BaseSpec)
	s.Name = name
	s.Labels = master.Labels
	s.Replicas = replicas
	s.ServiceAccountName = master.Spec.ServiceAccountName
	s.NodeSelector = nodeSelector
	s.RuntimeClassName = runtimeClass
	s.PriorityClassName = priorityClassName
	s.CdapConf = cconf
	s.HadoopConf = hconf
	return s
}

type StatelessSpec struct {
	BaseSpec   *BaseSpec        `json:"BaseSpec,inline"`
	Containers []*ContainerSpec `json:"containers,omitempty"`
}

func NewStatelessSpec(name string, replicas *int32, nodeSelector *map[string]string, runtimeClass string, priorityClassName string, master *v1alpha1.CDAPMaster, cconf, hconf string) *StatelessSpec {
	s := new(StatelessSpec)
	s.BaseSpec = NewBaseSpec(name, replicas, nodeSelector, runtimeClass, priorityClassName, master, cconf, hconf)
	return s
}
func (s *StatelessSpec) WithContainer(containerSpec *ContainerSpec) *StatelessSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}

type StorageSpec struct {
	Name             string `json:"name,omitempty"`
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageSize      int32  `json:"storageSize,omitempty"`
}

func NewStorageSpec(name, storageClassName string, storageSize int32) *StorageSpec {
	s := new(StorageSpec)
	s.Name = name
	s.StorageClassName = storageClassName
	s.StorageSize = storageSize
	return s
}

type StatefulSpec struct {
	Base          *BaseSpec        `json:"BaseSpec,inline"`
	InitContainer []*ContainerSpec `json:"initContainer,omitempty"`
	Containers    []*ContainerSpec `json:"containers,omitempty"`
	Storage       *StorageSpec     `json:"storage,omitempty"`
}

func NewStateful(name string, replicas *int32, nodeSelector *map[string]string, runtimeClass string, priorityClassName string, master *v1alpha1.CDAPMaster, cconf, hconf string) *StatefulSpec {
	s := new(StatefulSpec)
	s.Base = NewBaseSpec(name, replicas, nodeSelector, runtimeClass, priorityClassName, master, cconf, hconf)
	return s
}
func (s *StatefulSpec) WithInitContainer(containerSpec *ContainerSpec) *StatefulSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}
func (s *StatefulSpec) WithContainer(containerSpec *ContainerSpec) *StatefulSpec {
	s.Containers = append(s.Containers, containerSpec)
	return s
}
func (s *StatefulSpec) WithStorage(name, storageClassName string, storageSize int32) {
	s.Storage = NewStorageSpec(name, storageClassName, storageSize)
}

type ExternalServiceSpec struct {
	Name        string            `json:"name,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	ServiceType *string           `json:"serviceType,omitempty"`
	ServicePort int32             `json:"servicePort,omitempty"`
}

func NewExternalService(name string, annotation, labels map[string]string, serviceType *string, port int32) *ExternalServiceSpec {
	s := new(ExternalServiceSpec)
	s.Name = name
	s.Annotations = annotation
	s.Labels = labels
	s.ServiceType = serviceType
	s.ServicePort = port
	return s
}
