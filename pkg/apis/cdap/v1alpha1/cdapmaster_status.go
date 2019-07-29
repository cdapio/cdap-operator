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

// UpdateStatus for updating status for CDAP master
func (s *CDAPMasterSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	var period time.Duration
	master := rsrci.(*CDAPMaster)
	stts := &master.Status
	ready := stts.ComponentMeta.UpdateStatus(reconciler.ObjectsByType(reconciled, k8s.Type))
	stts.Meta.UpdateStatus(&ready, err)
	return period
}

// UpdateStatus for updating status for AppFabric service
func (s *AppFabricSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status Logs service
func (s *LogsSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status Messaging service
func (s *MessagingSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status Metadata service
func (s *MetadataSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status Metrics service
func (s *MetricsSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status Preview service
func (s *PreviewSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status Router service
func (s *RouterSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// UpdateStatus for updating status UserInterface service
func (s *UserInterfaceSpec) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return rsrci.(*CDAPMaster).updateStatus(reconciled, err)
}

// updateStatus for updating component status
func (r *CDAPMaster) updateStatus(reconciled []reconciler.Object, err error) time.Duration {
	var period time.Duration
	stts := &r.Status
	ready := stts.ComponentMeta.UpdateStatus(reconciler.ObjectsByType(reconciled, k8s.Type))
	stts.Meta.UpdateStatus(&ready, err)
	return period
}
