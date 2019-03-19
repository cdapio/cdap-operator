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
	"time"

	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
)

// UpdateComponentStatus for updating status for CDAP master
func (s *CDAPMasterSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	var period time.Duration
	master := rsrci.(*CDAPMaster)
	stts := &master.Status
	ready := stts.ComponentMeta.UpdateStatus(reconciler.ObjectsByType(reconciled, k8s.Type))
	stts.Meta.UpdateStatus(&ready, err)
	return period
}

// UpdateComponentStatus for updating status for AppFabric service
func (s *AppFabricSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status Logs service
func (s *LogsSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status Messaging service
func (s *MessagingSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status Metadata service
func (s *MetadataSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status Metrics service
func (s *MetricsSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status Preview service
func (s *PreviewSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status Router service
func (s *RouterSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// UpdateComponentStatus for updating status UserInterface service
func (s *UserInterfaceSpec) UpdateComponentStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateComponentStatus(reconciled, err)
}

// Updates the component status
func (r *CDAPMaster) updateComponentStatus(reconciled []reconciler.Object, err error) time.Duration {
	var period time.Duration
	stts := &r.Status
	ready := stts.ComponentMeta.UpdateStatus(reconciler.ObjectsByType(reconciled, k8s.Type))
	stts.Meta.UpdateStatus(&ready, err)
	return period
}
