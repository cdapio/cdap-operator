// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppFabricSpec) DeepCopyInto(out *AppFabricSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppFabricSpec.
func (in *AppFabricSpec) DeepCopy() *AppFabricSpec {
	if in == nil {
		return nil
	}
	out := new(AppFabricSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPExternalServiceSpec) DeepCopyInto(out *CDAPExternalServiceSpec) {
	*out = *in
	in.CDAPScalableServiceSpec.DeepCopyInto(&out.CDAPScalableServiceSpec)
	if in.ServiceType != nil {
		in, out := &in.ServiceType, &out.ServiceType
		*out = new(string)
		**out = **in
	}
	if in.ServicePort != nil {
		in, out := &in.ServicePort, &out.ServicePort
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPExternalServiceSpec.
func (in *CDAPExternalServiceSpec) DeepCopy() *CDAPExternalServiceSpec {
	if in == nil {
		return nil
	}
	out := new(CDAPExternalServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPMaster) DeepCopyInto(out *CDAPMaster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPMaster.
func (in *CDAPMaster) DeepCopy() *CDAPMaster {
	if in == nil {
		return nil
	}
	out := new(CDAPMaster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CDAPMaster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPMasterList) DeepCopyInto(out *CDAPMasterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CDAPMaster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPMasterList.
func (in *CDAPMasterList) DeepCopy() *CDAPMasterList {
	if in == nil {
		return nil
	}
	out := new(CDAPMasterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CDAPMasterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPMasterSpec) DeepCopyInto(out *CDAPMasterSpec) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.SystemAppConfigs != nil {
		in, out := &in.SystemAppConfigs, &out.SystemAppConfigs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.LogLevels != nil {
		in, out := &in.LogLevels, &out.LogLevels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.AppFabric.DeepCopyInto(&out.AppFabric)
	in.Logs.DeepCopyInto(&out.Logs)
	in.Messaging.DeepCopyInto(&out.Messaging)
	in.Metadata.DeepCopyInto(&out.Metadata)
	in.Metrics.DeepCopyInto(&out.Metrics)
	in.Preview.DeepCopyInto(&out.Preview)
	in.Router.DeepCopyInto(&out.Router)
	in.UserInterface.DeepCopyInto(&out.UserInterface)
	if in.Runtime != nil {
		in, out := &in.Runtime, &out.Runtime
		*out = new(RuntimeSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPMasterSpec.
func (in *CDAPMasterSpec) DeepCopy() *CDAPMasterSpec {
	if in == nil {
		return nil
	}
	out := new(CDAPMasterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPMasterStatus) DeepCopyInto(out *CDAPMasterStatus) {
	*out = *in
	in.Meta.DeepCopyInto(&out.Meta)
	in.ComponentMeta.DeepCopyInto(&out.ComponentMeta)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPMasterStatus.
func (in *CDAPMasterStatus) DeepCopy() *CDAPMasterStatus {
	if in == nil {
		return nil
	}
	out := new(CDAPMasterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPScalableServiceSpec) DeepCopyInto(out *CDAPScalableServiceSpec) {
	*out = *in
	in.CDAPServiceSpec.DeepCopyInto(&out.CDAPServiceSpec)
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPScalableServiceSpec.
func (in *CDAPScalableServiceSpec) DeepCopy() *CDAPScalableServiceSpec {
	if in == nil {
		return nil
	}
	out := new(CDAPScalableServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPServiceSpec) DeepCopyInto(out *CDAPServiceSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RuntimeClassName != nil {
		in, out := &in.RuntimeClassName, &out.RuntimeClassName
		*out = new(string)
		**out = **in
	}
	if in.PriorityClassName != nil {
		in, out := &in.PriorityClassName, &out.PriorityClassName
		*out = new(string)
		**out = **in
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPServiceSpec.
func (in *CDAPServiceSpec) DeepCopy() *CDAPServiceSpec {
	if in == nil {
		return nil
	}
	out := new(CDAPServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CDAPStatefulServiceSpec) DeepCopyInto(out *CDAPStatefulServiceSpec) {
	*out = *in
	in.CDAPServiceSpec.DeepCopyInto(&out.CDAPServiceSpec)
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CDAPStatefulServiceSpec.
func (in *CDAPStatefulServiceSpec) DeepCopy() *CDAPStatefulServiceSpec {
	if in == nil {
		return nil
	}
	out := new(CDAPStatefulServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogsSpec) DeepCopyInto(out *LogsSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogsSpec.
func (in *LogsSpec) DeepCopy() *LogsSpec {
	if in == nil {
		return nil
	}
	out := new(LogsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MessagingSpec) DeepCopyInto(out *MessagingSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MessagingSpec.
func (in *MessagingSpec) DeepCopy() *MessagingSpec {
	if in == nil {
		return nil
	}
	out := new(MessagingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetadataSpec) DeepCopyInto(out *MetadataSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetadataSpec.
func (in *MetadataSpec) DeepCopy() *MetadataSpec {
	if in == nil {
		return nil
	}
	out := new(MetadataSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsSpec) DeepCopyInto(out *MetricsSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsSpec.
func (in *MetricsSpec) DeepCopy() *MetricsSpec {
	if in == nil {
		return nil
	}
	out := new(MetricsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PreviewSpec) DeepCopyInto(out *PreviewSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PreviewSpec.
func (in *PreviewSpec) DeepCopy() *PreviewSpec {
	if in == nil {
		return nil
	}
	out := new(PreviewSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouterSpec) DeepCopyInto(out *RouterSpec) {
	*out = *in
	in.CDAPExternalServiceSpec.DeepCopyInto(&out.CDAPExternalServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouterSpec.
func (in *RouterSpec) DeepCopy() *RouterSpec {
	if in == nil {
		return nil
	}
	out := new(RouterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RuntimeSpec) DeepCopyInto(out *RuntimeSpec) {
	*out = *in
	in.CDAPStatefulServiceSpec.DeepCopyInto(&out.CDAPStatefulServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RuntimeSpec.
func (in *RuntimeSpec) DeepCopy() *RuntimeSpec {
	if in == nil {
		return nil
	}
	out := new(RuntimeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UserInterfaceSpec) DeepCopyInto(out *UserInterfaceSpec) {
	*out = *in
	in.CDAPExternalServiceSpec.DeepCopyInto(&out.CDAPExternalServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UserInterfaceSpec.
func (in *UserInterfaceSpec) DeepCopy() *UserInterfaceSpec {
	if in == nil {
		return nil
	}
	out := new(UserInterfaceSpec)
	in.DeepCopyInto(out)
	return out
}
