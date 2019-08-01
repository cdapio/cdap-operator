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

package cdapmaster

import (
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"time"
)

// UpdateStatus for updating status for CDAP master
func (s *Base) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status Messaging service
func (s *Messaging) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status for AppFabric service
func (s *AppFabric) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status Logs service
func (s *Logs) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status Metadata service
func (s *Metadata) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status Metrics service
func (s *Metrics) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status Preview service
func (s *Preview) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status Router service
func (s *Router) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}

// UpdateStatus for updating status UserInterface service
func (s *UserInterface) UpdateStatus(rsrci interface{}, reconciled []reconciler.Object, err error) time.Duration {
	return updateStatus(rsrci, reconciled, err)
}
