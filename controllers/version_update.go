package controllers

import (
	v1alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"reflect"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	"sigs.k8s.io/controller-reconciler/pkg/status"
	"strconv"
	"strings"
	"time"
)

type VersionUpdated = bool

var (
	updateStatus *VersionUpdateStatus
)

func init() {
	updateStatus = new(VersionUpdateStatus)
	updateStatus.init()
}

/////////////////////////////////////////////////////////////
///// Main functions for handle image upgarde/downgrade /////
/////////////////////////////////////////////////////////////

func handleVersionUpdate(master *v1alpha1.CDAPMaster, labels map[string]string, observed []reconciler.Object) ([]reconciler.Object, error) {
	// Let the current update complete if there is any
	if isConditionTrue(master, updateStatus.Inprogress) {
		log.Printf("Version update ingress. Continue... ")
		return upgradeForBackend(master, labels, observed)
	}

	if objs, versionUpdated, err := updateForUserInterface(master); err != nil {
		return nil, err
	} else if versionUpdated {
		return objs, nil
	}

	// Update backend service image version
	curVersion, err := getCurrentImageVersion(master)
	if err != nil {
		return nil, err
	}
	newVersion, err := getNewImageVersion(master)
	if err != nil {
		return nil, err
	}
	if len(curVersion.rawString) == 0 {
		setImageToUse(master)
		return []reconciler.Object{}, nil
	}

	switch compareVersion(curVersion, newVersion) {
	case -1:
		// Upgrade case

		// Don't retry upgrade if it failed.
		if isConditionTrue(master, updateStatus.UpgradeFailed) {
			return []reconciler.Object{}, nil
		}

		// Clear all conditions in preparation for a fresh upgrade
		updateStatus.clearAllConditions(master)

		setCondition(master, updateStatus.Inprogress)
		master.Status.UpgradeStartTimeMillis = getCurrentTimeMs()
		log.Printf("Version update: start upgrading %s -> %s ", curVersion.rawString, newVersion.rawString)
		return upgradeForBackend(master, labels, observed)
	case 0:
		// Reset all condition so that failed upgraded/downgrade can be retried later if needed.
		// This is needed when last upgrade failed and user has reset the version in spec.
		updateStatus.clearAllConditions(master)
		break
	case 1:
		// Downgrade

		// At the moment, downgrade never fails, so no need to check if isConditionTrue(downgrade failed)
		updateStatus.clearAllConditions(master)
		setCondition(master, updateStatus.Inprogress)
		master.Status.DowngradeStartTimeMillis = getCurrentTimeMs()
		log.Printf("Version update: start downgrading %s -> %s ", curVersion.rawString, newVersion.rawString)
		return downgradeForBackend(master)

	}
	return []reconciler.Object{}, nil
}

func updateForUserInterface(master *v1alpha1.CDAPMaster) ([]reconciler.Object, VersionUpdated, error) {
	// Update UI image version
	curUIVersion, err := getCurrentUserInterfaceVersion(master)
	if err != nil {
		return nil, false, err
	}
	newUIVersion, err := getNewUserInterfaceVersion(master)
	if err != nil {
		return nil, false, err
	}
	if len(curUIVersion.rawString) == 0 || compareVersion(curUIVersion, newUIVersion) != 0 {
		setUserInterfaceVersionToUse(master)
		log.Printf("Version update: for UserInterface %s->%s", curUIVersion.rawString, newUIVersion.rawString)
		return []reconciler.Object{}, true, nil
	}
	return []reconciler.Object{}, false, nil
}

func downgradeForBackend(master *v1alpha1.CDAPMaster) ([]reconciler.Object, error) {
	// Directly set the image to use. No pre- post- downgrade to run at the moment.
	setImageToUse(master)
	setCondition(master, updateStatus.DowngradeSucceeded)
	clearCondition(master, updateStatus.Inprogress)
	log.Printf("Version update: downgrade completed")
	return []reconciler.Object{}, nil
}

func upgradeForBackend(master *v1alpha1.CDAPMaster, labels map[string]string, observed []reconciler.Object) ([]reconciler.Object, error) {
	// Find either pre- or post- upgrade job
	findJob := func(jobName string) *batchv1.Job {
		var job *batchv1.Job = nil
		objName := getObjName(master, jobName)
		item := k8s.GetItem(observed, &batchv1.Job{}, objName, master.Namespace)
		if item != nil {
			job = item.(*batchv1.Job)
		}
		return job
	}

	// Create either pre- or post- upgrade job object based on the supplied job spec
	createJob := func(jobSpec *VersionUpgradeJobSpec) (*reconciler.Object, error) {
		jobObject, err := buildUpgradeJobObject(jobSpec)
		if err != nil {
			return nil, err
		}
		return jobObject, nil
	}

	// Build reconciler object based on the given job
	buildObject := func(job *batchv1.Job) *reconciler.Object {
		jobObj := &reconciler.Object{
			Type:      k8s.Type,
			Lifecycle: reconciler.LifecycleManaged,
			Obj: &k8s.Object{
				Obj:     job.DeepCopyObject().(metav1.Object),
				ObjList: &batchv1.JobList{},
			},
		}
		return jobObj
	}

	// First, run pre-upgrade job
	//
	// Note that pre-upgrade job doesn't have an "activeDeadlineSeconds" set it on, so it will
	// try as many as imageVersionUpgradeJobMaxRetryCount times before giving up. If we ever
	// needed to set an overall deadline for the pre-upgrade job, the logic below needs to check
	// deadline exceeded condition on job's status
	if !isConditionTrue(master, updateStatus.PreUpgradeSucceeded) {
		log.Printf("Version update: pre-upgrade job not completed")
		preJobName := getPreUpgradeJobName(master.Status.UpgradeStartTimeMillis)
		preJobSpec := buildPreUpgradeJobSpec(getPreUpgradeJobName(master.Status.UpgradeStartTimeMillis), master, labels)
		job := findJob(preJobName)
		if job == nil {
			obj, err := createJob(preJobSpec)
			if err != nil {
				return nil, err
			}
			log.Printf("Version update: creating pre-upgrade job")
			return []reconciler.Object{*obj}, nil
		} else if job.Status.Succeeded > 0 {
			setCondition(master, updateStatus.PreUpgradeSucceeded)
			log.Printf("Version update: pre-upgrade job succeeded")
			// Return empty to delete preUpgrade jobObj
			return []reconciler.Object{}, nil
		} else if job.Status.Failed > imageVersionUpgradeJobMaxRetryCount {
			setCondition(master, updateStatus.PreUpgradeFailed)
			setCondition(master, updateStatus.UpgradeFailed)
			clearCondition(master, updateStatus.Inprogress)
			log.Printf("Version update: pre-upgrade job failed, exceeded max retries.")
			return []reconciler.Object{}, nil
		} else {
			log.Printf("Version update: pre-upgrade job inprogress.")
			return []reconciler.Object{*buildObject(job)}, nil
		}
	}

	// Then, actually update the image version
	if !isConditionTrue(master, updateStatus.VersionUpdated) {
		setImageToUse(master)
		setCondition(master, updateStatus.VersionUpdated)
		log.Printf("Version update: set new version.")
		return []reconciler.Object{}, nil
	}

	// At last, run post-upgrade job
	//
	// Note that post-upgrade job doesn't have an "activeDeadlineSeconds" set it on, so it will
	// try as many as imageVersionUpgradeJobMaxRetryCount times before giving up. If we ever
	// needed to set an overall deadline for the post-upgrade job, the logic below needs to check
	// deadline exceeded condition on job's status
	if !isConditionTrue(master, updateStatus.PostUpgradeSucceeded) {
		log.Printf("Version update: post-upgrade job not completed")
		postJobName := getPostUpgradeJobName(master.Status.UpgradeStartTimeMillis)
		postJobSpec := buildPostUpgradeJobSpec(getPostUpgradeJobName(master.Status.UpgradeStartTimeMillis), master, labels)
		job := findJob(postJobName)
		if job == nil {
			obj, err := createJob(postJobSpec)
			if err != nil {
				return nil, err
			}
			log.Printf("Version update: creating post-upgrade job")
			return []reconciler.Object{*obj}, nil
		} else if job.Status.Succeeded > 0 {
			setCondition(master, updateStatus.PostUpgradeSucceeded)
			log.Printf("Version update: post-upgrade job succeeded")
			// Return empty to delete postUpgrade job
			return []reconciler.Object{}, nil
		} else if job.Status.Failed > imageVersionUpgradeJobMaxRetryCount {
			setCondition(master, updateStatus.PostUpgradeFailed)
			setCondition(master, updateStatus.UpgradeFailed)
			clearCondition(master, updateStatus.Inprogress)
			log.Printf("Version update: post-upgrade job failed, exceeded max retries.")
			return []reconciler.Object{*buildObject(job)}, nil
		} else {
			log.Printf("Version update: post-upgrade job inprogress.")
			return []reconciler.Object{*buildObject(job)}, nil
		}
	}
	setCondition(master, updateStatus.UpgradeSucceeded)
	clearCondition(master, updateStatus.Inprogress)
	log.Printf("Version update: upgrade succeeded.")
	return []reconciler.Object{}, nil
}

///////////////////////////////////////////////////////////////////////
////// Struct and util functions to tack version upgrade progress /////
///////////////////////////////////////////////////////////////////////

// For upgrade:
// - When succeeded:
//   * PreUpgradeSucceeded, PostUpgradeSucceeded and UpgradeSucceeded are set
//   * Status.ImageToUse (new image) == Spec.Image (new image)
// - When failed, two cases
//   1) Preupgrade failed
//      * PreUpgradeFailed and UpgradeFailed are set
//      * Status.ImageToUse (new image) != Spec.Image (current image)
//   2) Postupgrade failed
//      * PostUpgradeFailed and UpgradeFailed are set
//      * Status.ImageToUse (new image) == Spec.Image (new image)
// For downgrade:
// - When succeeded:
//   * DowngradeSucceeded is set
//   * Status.ImageToUse (new image) == Spec.Image (new image)
// - When failed (currently not possible, as we just set the new version directly)
type VersionUpdateStatus struct {
	// common states
	Inprogress     status.Condition
	VersionUpdated status.Condition

	// states specifically upgrade
	PreUpgradeSucceeded  status.Condition
	PreUpgradeFailed     status.Condition
	PostUpgradeSucceeded status.Condition
	PostUpgradeFailed    status.Condition
	UpgradeSucceeded     status.Condition
	UpgradeFailed        status.Condition

	// states specifically downgrade
	DowngradeSucceeded status.Condition
}

func (s *VersionUpdateStatus) init() {
	// common states
	s.Inprogress = status.Condition{
		Type:    "VersionUpdateInprogress",
		Reason:  "Start",
		Message: "Version update is inprogress",
	}
	s.VersionUpdated = status.Condition{
		Type:    "VersionUpdated",
		Reason:  "Start",
		Message: "Version to be used has been updated ",
	}

	// States for upgrade
	s.PreUpgradeSucceeded = status.Condition{
		Type:    "VersionPreUpgradeJobSucceeded",
		Reason:  "Start",
		Message: "Version pre-upgrade job is succeeded",
	}
	s.PreUpgradeFailed = status.Condition{
		Type:    "VersionPreUpgradeJobFailed",
		Reason:  "Start",
		Message: "Version pre-upgrade job is failed",
	}
	s.PostUpgradeSucceeded = status.Condition{
		Type:    "VersionPostUpgradeJobSucceeded",
		Reason:  "Start",
		Message: "Version post-upgrade job succeeded",
	}
	s.PostUpgradeFailed = status.Condition{
		Type:    "VersionPostUpgradeJobFailed",
		Reason:  "Start",
		Message: "Version post-upgrade job failed",
	}
	s.UpgradeSucceeded = status.Condition{
		Type:    "VersionUpgradeSucceeded",
		Reason:  "Start",
		Message: "Version upgrade has completed successfully",
	}
	s.UpgradeFailed = status.Condition{
		Type:    "VersionUpgradeFailed",
		Reason:  "Start",
		Message: "Version upgrade has failed",
	}

	// States for downgrade
	s.DowngradeSucceeded = status.Condition{
		Type:    "VersionDowngradeSucceeded",
		Reason:  "Start",
		Message: "Version downgrade has succeeded",
	}

}

// Clear all conditions used for track version update progress.
// Used at the beginning of version upgrade/downgrade process
func (s *VersionUpdateStatus) clearAllConditions(master *v1alpha1.CDAPMaster) error {
	conditions := reflect.ValueOf(*s)
	for i := 0; i < conditions.NumField(); i++ {
		condition, ok := conditions.Field(i).Interface().(status.Condition)
		if !ok {
			return fmt.Errorf("failed to convert field %s to condition type", conditions.Field(i).Type().Name())
		}
		clearCondition(master, condition)
	}
	return nil
}

func isConditionTrue(master *v1alpha1.CDAPMaster, condition status.Condition) bool {
	return master.Status.IsConditionTrue(condition.Type)
}

func setCondition(master *v1alpha1.CDAPMaster, condition status.Condition) {
	master.Status.SetCondition(condition.Type, condition.Reason, condition.Message)
}

func clearCondition(master *v1alpha1.CDAPMaster, condition status.Condition) {
	master.Status.ClearCondition(condition.Type, condition.Reason, condition.Message)
}

////////////////////////////////////////////////////////////////////////////////
///// struct and functions for parsing and processing image version string /////
////////////////////////////////////////////////////////////////////////////////

type Version struct {
	// store the raw string that was parsed
	rawString string
	// Boolean if the Version is "latest"
	latest bool
	// List of Version integers, starting from major
	components []int
}

// Parse image string to extract components of the version.
func parseImageString(imageString string) (*Version, error) {
	if len(imageString) == 0 {
		return &Version{}, nil
	}
	splits := strings.Split(imageString, ":")
	if len(splits) != 2 {
		return nil, fmt.Errorf("failed to parse image string %s, not in expected format of xxx:xxx", imageString)
	}

	if splits[1] == imageVersionLatest {
		return &Version{
			rawString: imageString,
			latest:    true,
		}, nil
	}

	versionString := splits[1]
	splits = strings.Split(versionString, ".")

	var components []int
	for _, s := range splits {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image string %s, unable to convert string to int", imageString)
		}
		components = append(components, i)
	}
	return &Version{
		rawString:  imageString,
		latest:     false,
		components: components,
	}, nil
}

// compare two parsed versions
// -1: left < right
//  0: left = right
//  1: left > right
func compareVersion(l, r *Version) int {
	if l.latest && r.latest {
		return 0
	} else if l.latest {
		return 1
	} else if r.latest {
		return -1
	}

	i := 0
	j := 0
	for i < len(l.components) && j < len(r.components) {
		if l.components[i] > r.components[j] {
			return 1
		} else if l.components[i] < r.components[j] {
			return -1
		}
		i++
		j++
	}
	for i < len(l.components) {
		if l.components[i] > 0 {
			return 1
		}
		i++
	}
	for j < len(r.components) {
		if r.components[j] > 0 {
			return 1
		}
		j++
	}
	return 0
}

//////////////////////////////////
///// Various util functions /////
//////////////////////////////////

func getCurrentImageVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	// current image version in use is stored in Status.ImageToUse
	curImage := master.Status.ImageToUse
	curVersion, err := parseImageString(curImage)
	if err != nil {
		return nil, err
	}
	return curVersion, nil
}

func getNewImageVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	// new image version to be deployed is stored in Spec.Image
	newImage := master.Spec.Image
	newVersion, err := parseImageString(newImage)
	if err != nil {
		return nil, err
	}
	return newVersion, nil
}

func setImageToUse(master *v1alpha1.CDAPMaster) {
	// This trigger actual image update as reconciler logic uses Status.ImageToUse to build expected state.
	master.Status.ImageToUse = master.Spec.Image
}

func getCurrentUserInterfaceVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	// current image version in use is stored in Status.UserInterfaceImageToUse
	curImage := master.Status.UserInterfaceImageToUse
	curVersion, err := parseImageString(curImage)
	if err != nil {
		return nil, err
	}
	return curVersion, nil
}

func getNewUserInterfaceVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	// new image version to be deployed is stored in Spec.UserInterfaceImage
	newImage := master.Spec.UserInterfaceImage
	newVersion, err := parseImageString(newImage)
	if err != nil {
		return nil, err
	}
	return newVersion, nil
}

func setUserInterfaceVersionToUse(master *v1alpha1.CDAPMaster) {
	// This trigger actual image update as reconciler logic uses Status.UserInterfaceImageToUse to build expected state.
	master.Status.UserInterfaceImageToUse = master.Spec.UserInterfaceImage
}

func getCurrentTimeMs() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// The returned name is just the suffix of actual k8s object name, as we prepend it with const string + CR name
func getPreUpgradeJobName(startTimeMs int64) string {
	return fmt.Sprintf("pre-upgrade-job-%d", startTimeMs)
}

// The returned name is just the suffix of actual k8s object name, as we prepend it with const string + CR name
func getPostUpgradeJobName(startTimeMs int64) string {
	return fmt.Sprintf("post-upgrade-job-%d", startTimeMs)
}

// Return pre-upgrade job spec
func buildPreUpgradeJobSpec(jobName string, master *v1alpha1.CDAPMaster, labels map[string]string) *VersionUpgradeJobSpec {
	startTimeMs := master.Status.UpgradeStartTimeMillis
	cconf := getObjName(master, configMapCConf)
	hconf := getObjName(master, configMapHConf)
	name := getObjName(master, jobName)
	return newUpgradeJobSpec(master, name, labels, startTimeMs, cconf, hconf).SetPreUpgrade(true)
}

// Return post-upgrade job spec
func buildPostUpgradeJobSpec(jobName string, master *v1alpha1.CDAPMaster, labels map[string]string) *VersionUpgradeJobSpec {
	startTimeMs := master.Status.UpgradeStartTimeMillis
	cconf := getObjName(master, configMapCConf)
	hconf := getObjName(master, configMapHConf)
	name := getObjName(master, jobName)
	return newUpgradeJobSpec(master, name, labels, startTimeMs, cconf, hconf).SetPostUpgrade(true)
}

// Given an upgrade job spec, return a reconciler object as expected state
func buildUpgradeJobObject(spec *VersionUpgradeJobSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+templateUpgradeJob, spec, &batchv1.JobList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
