/*
Copyright 2020 The CDAP Authors.
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
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////
// IMPORTANT NOTE:
// This file contains skeletons of reconciler handlers from previous version of CDAP operator "tags/v1.0".
// They are needed for backward compatibility (i.e. allow operator to be properly upgraded from v1.0 to current version)
//
// The handlers have been updated with returning empty []reconciler.Object{} in Observables, indicating deleting
// all k8s objects that were created by corresponding handlers. These handlers are also registered with the controller
// so they are part of reconciling loop just for cleaning up k8s objects created by previous version of operator.
//
// Details:
// The library controller-reconciler used by this operator internally adds a label in the form of
// <package_name>.<handler_struct_name> to k8s objects, thus identifying k8s objects managed by this operator.
// The new version of operator has different package and handler struct names, therefore without this file upgrading
// operator from old version to the new one in a kubernetes cluster with CDAP service deployed would cause the new
// operator unable to identify the existing CDAP services deployed by previous version of operator.
//
// We can remove this file once upgrading from tags/v1.0 to current version is no longer supported.

// Base - interface to handle cdapmaster
type Base struct{}

// Messaging - interface to handle cdapmaster
type Messaging struct{}

// AppFabric - interface to handle cdapmaster
type AppFabric struct{}

// Metrics - interface to handle cdapmaster
type Metrics struct{}

// Logs - interface to handle cdapmaster
type Logs struct{}

// Metadata - interface to handle cdapmaster
type Metadata struct{}

// Preview - interface to handle cdapmaster
type Preview struct{}

// Router - interface to handle cdapmaster
type Router struct{}

// UserInterface - interface to handle cdapmaster
type UserInterface struct{}

// SupportBundle - interface to handle cdapmaster
type SupportBundle struct{}

// TetheringAgent - interface to handle cdapmaster
type TetheringAgent struct{}

// ArtifactCache - interface to handle cdapmaster
type ArtifactCache struct{}

// Objects - handler Objects
func (b *Base) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for base
func (b *Base) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&corev1.ConfigMapList{}).
		For(&batchv1.JobList{}).
		Get()
}

// Objects for Messaging service
func (s *Messaging) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for messaging
func (s *Messaging) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for AppFabric service
func (s *AppFabric) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for appfabric
func (s *AppFabric) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		Get()
}

// Objects for Logs service
func (s *Logs) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for logs
func (s *Logs) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for Metadata service
func (s *Metadata) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for metadata
func (s *Metadata) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		Get()
}

// Objects for Metrics service
func (s *Metrics) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for metrics
func (s *Metrics) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for Preview service
func (s *Preview) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for preview
func (s *Preview) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for Router service
func (s *Router) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for router
func (s *Router) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		Get()
}

// Objects for UserInterface service
func (s *UserInterface) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for UserInterface
func (s *UserInterface) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		Get()
}

// Objects for SupportBundle service
func (s *SupportBundle) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for support-bundle
func (s *SupportBundle) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for TetheringAgent service
func (s *TetheringAgent) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for TetheringAgent
func (s *TetheringAgent) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}

// Objects for ArtifactCache service
func (s *ArtifactCache) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	return []reconciler.Object{}, nil
}

// Observables for ArtifactCache
func (s *ArtifactCache) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.StatefulSetList{}).
		Get()
}
