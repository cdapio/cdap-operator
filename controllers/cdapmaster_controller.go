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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	gr "sigs.k8s.io/controller-reconciler/pkg/genericreconciler"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	alpha1 "cdap.io/cdap-operator/api/v1alpha1"
)

const (
	containerLabel = "cdap.container"
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
