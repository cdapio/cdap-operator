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
	"strconv"
	"strings"
	"text/template"

	"cdap.io/cdap-operator/controllers/cdapmaster"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-reconciler/pkg/finalizer"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	gr "sigs.k8s.io/controller-reconciler/pkg/genericreconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CDAPMasterReconciler reconciles a CDAPMaster object
type CDAPMasterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *CDAPMasterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CDAPMaster{}).
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
	// Registering cdapmaster.* handlers (from old version tags/v1.0) in order to support backward compatibility
	// Essentially those handler will delete CDAP services and configures created by previous version of operator
	// and let the handlers in the new operator to re-deploy CDAP.
	return gr.
		WithManager(mgr).
		For(&v1alpha1.CDAPMaster{}, v1alpha1.GroupVersion).
		Using(&cdapmaster.Base{}).
		Using(&cdapmaster.Messaging{}).
		Using(&cdapmaster.AppFabric{}).
		Using(&cdapmaster.Metrics{}).
		Using(&cdapmaster.Logs{}).
		Using(&cdapmaster.Metadata{}).
		Using(&cdapmaster.Preview{}).
		Using(&cdapmaster.Router{}).
		Using(&cdapmaster.UserInterface{}).
		Using(&VersionUpdateHandler{}).
		Using(&ConfigMapHandler{}).
		Using(&ServiceHandler{}).
		WithErrorHandler(HandleError).
		WithDefaulter(ApplyDefaults).
		Build()
}

func HandleError(resource interface{}, err error, kind string) {
	cm := resource.(*v1alpha1.CDAPMaster)
	if err != nil {
		cm.Status.SetError("ErrorSeen", err.Error())
	} else {
		cm.Status.ClearError()
	}
}

func ApplyDefaults(resource interface{}) {
	r := resource.(*v1alpha1.CDAPMaster)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[labelInstanceKey] = r.Name

	spec := &r.Spec
	if spec.Image == "" {
		spec.Image = defaultImage
	}
	if spec.UserInterfaceImage == "" {
		spec.UserInterfaceImage = spec.Image
	}

	if spec.Router.ServicePort == nil {
		spec.Router.ServicePort = int32Ptr(defaultRouterPort)
	}
	if spec.UserInterface.ServicePort == nil {
		spec.UserInterface.ServicePort = int32Ptr(defaultUserInterfacePort)
	}

	if spec.Config == nil {
		spec.Config = make(map[string]string)
	}

	// Set the configMapCConf entry for the router and UI service and ports
	spec.Config[confRouterServerAddress] = fmt.Sprintf("cdap-%s-%s", r.Name, strings.ToLower(string(serviceRouter)))
	spec.Config[confRouterBindPort] = strconv.Itoa(int(*spec.Router.ServicePort))
	spec.Config[confUserInterfaceBindPort] = strconv.Itoa(int(*spec.UserInterface.ServicePort))

	// Set the default local data directory if it is not set in cdap-cr.
	if _, ok := spec.Config[confLocalDataDirKey]; !ok {
		spec.Config[confLocalDataDirKey] = confLocalDataDirVal
	}

	// Disable explore
	spec.Config[confExploreEnabled] = "false"

	r.Status.ResetComponentList()
	r.Status.EnsureStandardConditions()
	finalizer.EnsureStandard(r)
}

/////////////////////////////////////////////////////////
///// Handling reconciling ConfigMapHandler objects /////
/////////////////////////////////////////////////////////
type ConfigMapHandler struct{}

func (h *ConfigMapHandler) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&corev1.ConfigMapList{}).
		Get()
}

func (h *ConfigMapHandler) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	var expected []reconciler.Object
	m := rsrc.(*v1alpha1.CDAPMaster)

	configs := map[string][]string{
		configMapCConf: {"cdap-site.xml", "logback.xml", "logback-container.xml"},
		configMapHConf: {"core-site.xml"},
	}

	templateData := struct {
		Master *v1alpha1.CDAPMaster
	}{
		Master: m,
	}

	fillTemplate := func(templateFile string) (string, error) {
		tmpl, err := template.New(templateFile).Funcs(template.FuncMap{
			"hasPrefix": func(str, prefix string) bool {
				return strings.HasPrefix(str, prefix)
			},
			"trimPrefix": func(str, prefix string) string {
				return strings.TrimPrefix(str, prefix)
			},
		}).ParseFiles(templateDir + templateFile)
		if err != nil {
			return "", err
		}
		var output strings.Builder
		if err := tmpl.Execute(&output, templateData); err != nil {
			return "", err
		}
		return output.String(), nil
	}

	mergedLabelmap := mergeMaps(m.Labels, rsrclabels)
	for key, templateFiles := range configs {
		spec := newConfigMapSpec(m, getObjName(m, key), mergedLabelmap)
		for _, file := range templateFiles {
			data, err := fillTemplate(file)
			if err != nil {
				return nil, err
			}
			spec = spec.AddData(file, data)
		}
		obj := buildConfigMapObject(spec)
		expected = append(expected, obj)
	}

	// Creates system app config object. Creates one data object per system app config file.
	sysAppConfigSpec := newConfigMapSpec(m, getObjName(m, configMapSysAppConf), mergedLabelmap)
	for filename, sysAppConfig := range m.Spec.SystemAppConfigs {
		sysAppConfigSpec = sysAppConfigSpec.AddData(filename, sysAppConfig)
	}
	expected = append(expected, buildConfigMapObject(sysAppConfigSpec))
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

///////////////////////////////////////////////////////////
///// Handling reconciling deployment of all services /////
///////////////////////////////////////////////////////////
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

	m := rsrc.(*v1alpha1.CDAPMaster)
	// Merge in labels (e.g. "using: <handler method name>") added by underlying reconciler-controller library
	labels := mergeMaps(m.Labels, rsrclabels)

	// Build deployment specification that defines the statefulset, deployment, node port services to be created.
	spec, err := buildDeploymentPlanSpec(m, labels)
	if err != nil {
		return []reconciler.Object{}, err
	}
	objs, err = buildObjectsForDeploymentPlan(spec)
	if err != nil {
		return []reconciler.Object{}, err
	}
	expected = append(expected, objs...)

	// Copy NodePort from observed to ensure k8s services' nodePorts stay the same across reconciling iterators
	CopyNodePortIfAny(expected, observed)

	return expected, nil
}

// Copy the nodePort from observed to the expected to ensure the nodePort remains unchanged
func CopyNodePortIfAny(expected, observed []reconciler.Object) {
	// Map from CDAP service's namespaced name to a map from NodePort's name to port
	serviceToNodePorts := make(map[string]map[string]int32)

	// Return namespaced name
	getNName := func(s *corev1.Service) string {
		return s.Namespace + "-" + s.Name
	}

	for _, item := range reconciler.ObjectsByType(observed, k8s.Type) {
		service, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service)
		if !ok {
			continue
		}
		nodePorts := make(map[string]int32)
		for _, port := range service.Spec.Ports {
			nodePorts[port.Name] = port.NodePort
		}
		serviceToNodePorts[getNName(service)] = nodePorts
	}

	for _, item := range reconciler.ObjectsByType(expected, k8s.Type) {
		service, ok := item.Obj.(*k8s.Object).Obj.(*corev1.Service)
		if !ok {
			continue
		}
		oldNodePorts, ok := serviceToNodePorts[getNName(service)]
		if !ok {
			continue
		}
		for i := range service.Spec.Ports {
			newNodePort := &service.Spec.Ports[i]
			oldPort, ok := oldNodePorts[newNodePort.Name]
			if !ok {
				continue
			}
			newNodePort.NodePort = oldPort
		}
	}
}

///////////////////////////////////////////////////////
///// Handler for image version upgrade/downgrade /////
///////////////////////////////////////////////////////
type VersionUpdateHandler struct{}

func (h *VersionUpdateHandler) Observables(rsrc interface{}, labels map[string]string, dependent []reconciler.Object) []reconciler.Observable {
	return k8s.NewObservables().
		WithLabels(labels).
		For(&batchv1.JobList{}).
		Get()
}

func (h *VersionUpdateHandler) Objects(rsrc interface{}, rsrclabels map[string]string, observed, dependent, aggregated []reconciler.Object) ([]reconciler.Object, error) {
	m := rsrc.(*v1alpha1.CDAPMaster)
	labels := mergeMaps(m.Labels, rsrclabels)
	return handleVersionUpdate(m, labels, observed)
}
