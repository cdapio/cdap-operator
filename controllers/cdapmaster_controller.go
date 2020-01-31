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
		Complete(NewReconciler(mgr))
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
// Intentionally leave a blank line, otherwise controller-gen won't generate RBAC

func NewReconciler(mgr manager.Manager) *gr.Reconciler {
	return gr.
		WithManager(mgr).
		For(&alpha1.CDAPMaster{}, alpha1.GroupVersion).
		Using(&ConfigMapHandler{}).
		Using(&ServiceHandler{}).
		WithErrorHandler(HandleError).
		WithDefaulter(ApplyDefaults).
		Build()
}

func HandleError(resource interface{}, err error, kind string) {
	cm := resource.(*alpha1.CDAPMaster)
	if err != nil {
		cm.Status.SetError("ErrorSeen", err.Error())
	} else {
		cm.Status.ClearError()
	}
}

func ApplyDefaults(resource interface{}) {
	cm := resource.(*alpha1.CDAPMaster)
	cm.ApplyDefaults()
}

// Handling reconciling ConfigMapHandler objects
type ConfigMapHandler struct{}

func (h *ConfigMapHandler) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&corev1.ConfigMapList{}).
		Get()
}

func (h *ConfigMapHandler) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	var expected []reconciler.Object
	m := rsrc.(*alpha1.CDAPMaster)

	configs := map[string][]string{
		cconf: {"cdap-site.xml", "logback.xml", "logback-container.xml"},
		hconf: {"core-site.xml"},
	}

	templateData := struct {
		Master *alpha1.CDAPMaster
	}{
		Master: m,
	}

	fillTemplate := func(templateFile string) (string, error) {
		template, err := template.New(templateFile).ParseFiles(templateDir + templateFile)
		if err != nil {
			return "", err
		}
		var output strings.Builder
		if err := template.Execute(&output, templateData); err != nil {
			return "", err
		}
		return output.String(), nil
	}

	for key, templateFiles := range configs {
		spec := NewConfigMapSpec(getObjName(m.Name, key), m.Namespace, mergeLabels(m.Labels, rsrclabels))
		for _, file := range templateFiles {
			data, err := fillTemplate(file)
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
type ServiceHandler struct{}

func (h *ServiceHandler) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&appsv1.DeploymentList{}).
		For(&corev1.ServiceList{}).
		For(&appsv1.StatefulSetList{}).
		Get()
}

func (h *ServiceHandler) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	var expected, objs []reconciler.Object
	var err error

	m := rsrc.(*alpha1.CDAPMaster)

	cconf := getObjName(m.Name, cconf)
	hconf := getObjName(m.Name, hconf)
	labels := mergeLabels(m.Labels, rsrclabels)
	dataDir := alpha1.LocalDataDir

	// Build a single StatefulSet
	buildStatefulset := func(name string, service alpha1.ServiceName, serviceSpec *alpha1.CDAPStatefulServiceSpec) *StatefulSpec {
		initContainer := NewContainerSpec("create-storage", "StorageMain", m, nil, dataDir)
		container := NewContainerSpec(strings.ToLower(string(service)), getServiceMain(service), m, serviceSpec.Resources, dataDir)
		return NewStatefulSpec(name, 1, labels, &serviceSpec.CDAPServiceSpec, m, cconf, hconf).
			AddLabel(containerLabel, name).
			WithInitContainer(initContainer).
			WithContainer(container).
			WithStorage(serviceSpec.StorageClassName, serviceSpec.StorageSize)
	}

	// Build a single Deployment
	buildDeployment := func(name string, service alpha1.ServiceName, serviceSpec *alpha1.CDAPServiceSpec) *DeploymentSpec {
		container := NewContainerSpec(strings.ToLower(string(service)), getServiceMain(service), m, serviceSpec.Resources, dataDir)
		return NewDeploymentSpec(name, 1, labels, serviceSpec, m, cconf, hconf).
			AddLabel(containerLabel, name).
			WithContainer(container)
	}

	// Build user interface Deployment
	// TODO: consider making deployemnt shared by both UI and services
	buildUserInterface := func(name string, serviceName alpha1.ServiceName, serviceSpec *alpha1.CDAPScalableServiceSpec) *UserInterfaceSpec {
		container := NewContainerSpec(strings.ToLower(string(serviceName)), "", m, serviceSpec.Resources, dataDir).
			SetImage(m.Spec.UserInterfaceImage)
		return NewUserInterfaceSpec(name, getReplicas(serviceSpec.Replicas), labels, &serviceSpec.CDAPServiceSpec, m, cconf, hconf).
			AddLabel(containerLabel, name).
			WithContainer(container)
	}

	// Build a single NodePort service
	buildNetworkService := func(name string, serviceSpec *alpha1.CDAPExternalServiceSpec) *NetworkServiceSpec {
		return NewNetworkServiceSpec(name, labels, serviceSpec.ServiceType, serviceSpec.ServicePort, m).
			AddLabel(containerLabel, name)
	}

	spec := NewCDAPDeploymentSpec().
		WithStateful(buildStatefulset(getObjName(m.Name, "logs"), alpha1.ServiceLogs, &m.Spec.Logs.CDAPStatefulServiceSpec)).
		WithStateful(buildStatefulset(getObjName(m.Name, "messaging"), alpha1.ServiceMessaging, &m.Spec.Messaging.CDAPStatefulServiceSpec)).
		WithStateful(buildStatefulset(getObjName(m.Name, "metrics"), alpha1.ServiceMetrics, &m.Spec.Preview.CDAPStatefulServiceSpec)).
		WithStateful(buildStatefulset(getObjName(m.Name, "preview"), alpha1.ServicePreview, &m.Spec.Preview.CDAPStatefulServiceSpec)).
		WithDeployment(buildDeployment(getObjName(m.Name, "appfab"), alpha1.ServiceAppFabric, &m.Spec.AppFabric.CDAPServiceSpec)).
		WithDeployment(buildDeployment(getObjName(m.Name, "metadata"), alpha1.ServiceMetadata, &m.Spec.Metadata.CDAPServiceSpec)).
		WithDeployment(buildDeployment(getObjName(m.Name, "router"), alpha1.ServiceRouter, &m.Spec.Router.CDAPServiceSpec)).
		WithUserInterface(buildUserInterface(getObjName(m.Name, "userinterface"), alpha1.ServiceUserInterface, &m.Spec.UserInterface.CDAPExternalServiceSpec.CDAPScalableServiceSpec)).
		WithNetworkService(buildNetworkService(getObjName(m.Name, "router"), &m.Spec.Router.CDAPExternalServiceSpec)).
		WithNetworkService(buildNetworkService(getObjName(m.Name, "userinterface"), &m.Spec.UserInterface.CDAPExternalServiceSpec))

	objs, err = buildObjects(spec)
	if err != nil {
		return []reconciler.Object{}, err
	}
	expected = append(expected, objs...)

	// Copy NodePort from obseved to expected
	copyNodePort(expected, observed)

	return expected, err
}

func buildObjects(spec *CDAPDeploymentSpec) ([]reconciler.Object, error) {
	var objs []reconciler.Object
	for _, s := range spec.Stateful {
		obj, err := buildStatefulObject(s)
		if err != nil {
			return nil, err
		}
		objs = append(objs, *obj)
	}
	for _, s := range spec.Deployment {
		obj, err := buildDeploymentObject(s)
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
	obj, err := buildUIDeploymentObject(spec.UserInterface)
	if err != nil {
		return nil, err
	}
	objs = append(objs, *obj)

	return objs, nil
}

func buildStatefulObject(spec *StatefulSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+statefulSetTemplate, spec, &appsv1.StatefulSetList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func buildDeploymentObject(spec *DeploymentSpec) (*reconciler.Object, error) {
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

func buildUIDeploymentObject(spec *UserInterfaceSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+uiDeploymentTemplate, spec, &appsv1.DeploymentList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Copy the nodePort from observed to the expected to ensure the nodePort is unchanged
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

func getObjName(masterName, name string) string {
	return fmt.Sprintf("%s%s-%s", objectNamePrefix, masterName, strings.ToLower(name))
}

func getServiceMain(name alpha1.ServiceName) string {
	return fmt.Sprintf("%sServiceMain", name)
}

func getReplicas(replicas *int32) int32 {
	var r int32 = 1
	if replicas != nil {
		r = *replicas
	}
	return r
}

func mergeLabels(current, added map[string]string) map[string]string {
	labels := make(reconciler.KVMap)
	labels.Merge(current, added)
	return labels
}
