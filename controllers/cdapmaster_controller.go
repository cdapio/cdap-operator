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

package controllers

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	gr "sigs.k8s.io/controller-reconciler/pkg/genericreconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	objectNamePrefix = "cdap-"
	cconf            = "cconf"
	hconf            = "hconf"
	containerLabel   = "cdap.container"
	// Heap memory related constants
	javaMinHeapRatio     = float64(0.6)
	javaReservedNonHeap  = int64(768 * 1024 * 1024)
	templateDir          = "templates/"
	deploymentTemplate   = "cdap-deployment.yaml"
	uiDeploymentTemplate = "cdap-ui-deployment.yaml"
	statefulSetTemplate  = "cdap-sts.yaml"
	serviceTemplate      = "cdap-service.yaml"
	upgradeJobTemplate   = "upgrade-job.yaml"

	upgradeFailed             = "upgrade-failed"
	postUpgradeFailed         = "post-upgrade-failed"
	postUpgradeFinished       = "post-upgrade-finished"
	upgradeStartMessage       = "Upgrade started, received updated CR."
	upgradeFailedInitMessage  = "Failed to create job, upgrade failed."
	upgradeJobFailedMessage   = "Upgrade job failed."
	upgradeJobFinishedMessage = "Upgrade job finished."
	upgradeJobSkippedMessage  = "Upgrade job skipped."
	upgradeResetMessage       = "Upgrade spec reset."
	upgradeFailureLimit       = 4

	latestVersion = "latest"
)

// CDAPMasterReconciler reconciles a CDAPMaster object
type CDAPMasterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *CDAPMasterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alpha1.CDAPMaster{}).
		Complete(newReconciler(mgr))
}

// TBD kubebuilder:rbac:groups=app.k8s.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cdap.cdap.io,resources=cdapmasters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cdap.cdap.io,resources=cdapmasters/status,verbs=get;update;patch

func newReconciler(mgr manager.Manager) *gr.Reconciler {
	return gr.
		WithManager(mgr).
		For(&alpha1.CDAPMaster{}, alpha1.GroupVersion).
		Using(&ConfigMap{}).
		Using(&ServiceSet{}).
		WithErrorHandler(handleError).
		WithDefaulter(applyDefaults).
		Build()
}

func handleError(resource interface{}, err error, kind string) {
	cm := resource.(*alpha1.CDAPMaster)
	if err != nil {
		cm.Status.SetError("ErrorSeen", err.Error())
	} else {
		cm.Status.ClearError()
	}
}

func applyDefaults(resource interface{}) {
	cm := resource.(*alpha1.CDAPMaster)
	cm.ApplyDefaults()
}

// Handling reconciling ConfigMap objects
type ConfigMap struct{}

func (b *ConfigMap) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&corev1.ConfigMapList{}).
		Get()
}

func (b *ConfigMap) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	var expected []reconciler.Object
	master := rsrc.(*alpha1.CDAPMaster)

	configs := map[string][]string{
		cconf: {"cdap-site.xml", "logback.xml", "logback-container.xml"},
		hconf: {"core-site.xml"},
	}

	type templateBaseValue struct {
		Master *alpha1.CDAPMaster
	}
	ngdata := &templateBaseValue{
		Master: master,
	}

	fillTemplate := func(templateFile string, ngdata *templateBaseValue) (string, error) {
		template, err := template.New(templateFile).ParseFiles(templateDir + templateFile)
		if err != nil {
			return "", nil
		}
		var output strings.Builder
		if err := template.Execute(&output, ngdata); err != nil {
			return "", nil
		}
		return output.String(), nil
	}

	for key, templateFiles := range configs {
		spec := NewConfigMapSpec(getObjectName(master.Name, key), master.Namespace, mergeLabels(master.Labels, rsrclabels))
		for _, file := range templateFiles {
			data, err := fillTemplate(file, ngdata)
			if err != nil {
				return nil, err
			}
			spec = spec.WithData(file, data)
		}
		obj := buildConfigMapObject(spec)
		expected = append(expected, obj)
	}
	return expected, nil
}

func buildConfigMapObject(spec *ConfigMapSpec) reconciler.Object {
	// Creates the configMap object
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
			Labels:    spec.Labels,
		},
		Data: spec.Data,
	}

	obj := reconciler.Object{
		Type:      k8s.Type,
		Lifecycle: reconciler.LifecycleManaged,
		Obj: &k8s.Object{
			Obj:     configMap.DeepCopyObject().(metav1.Object),
			ObjList: &corev1.ConfigMapList{},
		},
	}
	return obj
}

// Handling reconciling deployment of all services
type ServiceSet struct{}

func (s *ServiceSet) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		For(&appsv1.StatefulSetList{}).
		Get()
}

func (s *ServiceSet) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	var expected, objs []reconciler.Object
	var err error

	m := rsrc.(*alpha1.CDAPMaster)

	cconf := getObjectName(m.Name, cconf)
	hconf := getObjectName(m.Name, hconf)
	labels := mergeLabels(m.Labels, rsrclabels)
	dataDir := alpha1.LocalDataDir

	buildStatefulSpec := func(name string, serviceName alpha1.ServiceName, serviceSpec *alpha1.CDAPStatefulServiceSpec) *StatefulSpec {
		spec := NewStateful(name, 1, labels, &serviceSpec.CDAPServiceSpec, m, cconf, hconf).
			WithInitContainer(NewContainerSpec("create-storage", "StorageMain", m, nil, dataDir)).
			WithContainer(NewContainerSpec(strings.ToLower(string(serviceName)), getServiceMain(serviceName), m, serviceSpec.Resources, dataDir)).
			WithStorage(serviceSpec.StorageClassName, serviceSpec.StorageSize)
		return spec
	}

	buildStatelessSpec := func(name string, serviceName alpha1.ServiceName, serviceSpec *alpha1.CDAPServiceSpec) *StatelessSpec {
		spec := NewStatelessSpec(name, 1, labels, serviceSpec, m, cconf, hconf).
			WithContainer(NewContainerSpec(strings.ToLower(string(serviceName)), getServiceMain(serviceName), m, serviceSpec.Resources, dataDir))
		return spec
	}

	buildNetworkServiceSpec := func(name string, serviceSpec *alpha1.CDAPExternalServiceSpec) *NetworkServiceSpec {
		spec := NewNetworkServiceo_(name, labels, serviceSpec.ServiceType, serviceSpec.ServicePort, m)
		return spec
	}

	spec := NewDeploymentSpec().
		WithStateful(buildStatefulSpec(getObjectName(m.Name, "logs"), alpha1.ServiceLogs, &m.Spec.Logs.CDAPStatefulServiceSpec)).
		WithStateful(buildStatefulSpec(getObjectName(m.Name, "messaging"), alpha1.ServiceMessaging, &m.Spec.Messaging.CDAPStatefulServiceSpec)).
		WithStateful(buildStatefulSpec(getObjectName(m.Name, "metrics"), alpha1.ServiceMetrics, &m.Spec.Preview.CDAPStatefulServiceSpec)).
		WithStateful(buildStatefulSpec(getObjectName(m.Name, "preview"), alpha1.ServicePreview, &m.Spec.Preview.CDAPStatefulServiceSpec)).
		WithStateless(buildStatelessSpec(getObjectName(m.Name, "appfab"), alpha1.ServiceAppFabric, &m.Spec.AppFabric.CDAPServiceSpec)).
		WithStateless(buildStatelessSpec(getObjectName(m.Name, "metadata"), alpha1.ServiceMetadata, &m.Spec.Metadata.CDAPServiceSpec)).
		WithStateless(buildStatelessSpec(getObjectName(m.Name, "router"), alpha1.ServiceRouter, &m.Spec.Metadata.CDAPServiceSpec)).
		WithNetworkService(buildNetworkServiceSpec(getObjectName(m.Name, "router"), &m.Spec.Router.CDAPExternalServiceSpec))

	objs, err = buildObjects(spec)
	if err != nil {
		return []reconciler.Object{}, err
	}
	expected = append(expected, objs...)

	copyNodePort(expected, observed)

	return expected, err
}

func buildObjects(spec *DeploymentSpec) ([]reconciler.Object, error) {
	var objs []reconciler.Object
	for _, s := range spec.Stateful {
		obj, err := buildStatefulObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)
	}
	for _, s := range spec.Stateless {
		obj, err := buildStatelessObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)
	}
	for _, s := range spec.NetworkServices {
		obj, err := buildNetworkServiceObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)

	}
	return objs, nil
}

func buildStatefulObject(spec *StatefulSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+statefulSetTemplate, spec, &appsv1.StatefulSetList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func buildStatelessObject(spec *StatelessSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+deploymentTemplate, spec, &appsv1.DeploymentList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func buildNetworkServiceObject(spec *NetworkServiceSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+serviceTemplate, spec, &corev1.ServiceList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Set the nodePort in the expected service based on the observed service
func copyNodePort(expected, observed []reconciler.Object) {
	// Get the service from the expected list.
	var expectedService *corev1.Service
	for _, item := range reconciler.ObjectsByType(expected, k8s.Type) {
		if service, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
			expectedService = service
			break
		}
	}
	if expectedService == nil {
		return
	}
	// Find the service being observed. Extract nodePort from the service spec and set it to expected
	for _, item := range reconciler.ObjectsByType(observed, k8s.Type) {
		if observedService, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
			if observedService.Namespace == expectedService.Namespace && observedService.Name == expectedService.Name {
				var nodePorts = make(map[string]int32)
				for _, p := range observedService.Spec.Ports {
					nodePorts[p.Name] = p.NodePort
				}

				// Assigning existing node ports to the expected service
				for i := range expectedService.Spec.Ports {
					p := &expectedService.Spec.Ports[i]
					if nodePort, ok := nodePorts[p.Name]; ok {
						p.NodePort = nodePort
					}
				}
			}
		}
	}
}

func getObjectName(masterName, name string) string {
	return fmt.Sprintf("%s%s-%s", objectNamePrefix, masterName, strings.ToLower(name))
}

func getServiceMain(name alpha1.ServiceName) string {
	return fmt.Sprintf("%sServiceMain", name)
}

func mergeLabels(current, added map[string]string) map[string]string {
	labels := make(reconciler.KVMap)
	labels.Merge(current, added)
	return labels
}
