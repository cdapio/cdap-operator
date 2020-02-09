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

var (
	versionUpdateStatus *VersionUpdateStatus
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	versionUpdateStatus = new(VersionUpdateStatus)
	versionUpdateStatus.init()
}

// Main function to handle image version upgrade/downgrade
func updateVersion(master *v1alpha1.CDAPMaster, labels map[string]string, observed []reconciler.Object) ([]reconciler.Object, error) {
	// Let the current update complete if there is any
	if isConditionTrue(master, versionUpdateStatus.Inprogress) {
		log.Printf("Version update ingress. Continue... ")
		return doUpgrade(master, labels, observed)
	}

	// Update UI image version
	curUIVersion, err := getCurrentUserInterfaceVersion(master)
	if err != nil {
		return nil, err
	}
	newUIVersion, err := getNewUserInterfaceVersion(master)
	if err != nil {
		return nil, err
	}
	if curUIVersion.empty || compareVersion(curUIVersion, newUIVersion) != 0 {
		setUserInterfaceVersionToUse(master)
		log.Printf("Version update: for UserInterface %s->%s", curUIVersion.rawString, newUIVersion.rawString)
		return []reconciler.Object{}, nil
	}

	// Update backend service image version
	curVersion, err := getCurrentVersion(master)
	if err != nil {
		return nil, err
	}
	newVersion, err := getNewVersion(master)
	if err != nil {
		return nil, err
	}
	if curVersion.empty {
		setVersionToUse(master)
		return []reconciler.Object{}, nil
	}

	switch compareVersion(curVersion, newVersion) {
	case -1:
		versionUpdateStatus.clearAllConditions(master)
		setCondition(master, versionUpdateStatus.Inprogress)
		master.Status.UpgradeStartTimeMillis = getCurrentTimeMs()
		log.Printf("Version update: start upgrading %s -> %s ", curVersion.rawString, newVersion.rawString)
		return doUpgrade(master, labels, observed)
	case 0:
		break
	case 1:
		versionUpdateStatus.clearAllConditions(master)
		setCondition(master, versionUpdateStatus.Inprogress)
		master.Status.DowngradeStartTimeMillis = getCurrentTimeMs()
		log.Printf("Version update: start downgrading %s -> %s ", curVersion.rawString, newVersion.rawString)
		return doDowngrade(master)

	}
	return []reconciler.Object{}, nil
}

func doDowngrade(master *v1alpha1.CDAPMaster) ([]reconciler.Object, error) {
	setVersionToUse(master)
	setCondition(master, versionUpdateStatus.DowngradeSucceeded)
	clearCondition(master, versionUpdateStatus.Inprogress)
	log.Printf("Version update: downgrade completed")
	return []reconciler.Object{}, nil
}

func doUpgrade(master *v1alpha1.CDAPMaster, labels map[string]string, observed []reconciler.Object) ([]reconciler.Object, error) {
	findJob := func(jobName string) *batchv1.Job {
		var job *batchv1.Job = nil
		objName := getObjectName(master.Name, jobName)
		item := k8s.GetItem(observed, &batchv1.Job{}, objName, master.Namespace)
		if item != nil {
			job = item.(*batchv1.Job)
		}
		return job
	}

	createJob := func(jobSpec *UpgradeJobSpec) (*reconciler.Object, error) {
		jobObject, err := buildUpgradeJobObject(jobSpec)
		if err != nil {
			return nil, err
		}
		return jobObject, nil
	}

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

	if !isConditionTrue(master, versionUpdateStatus.PreUpgradeJobDone) {
		log.Printf("Version update: pre-upgrade job not completed")
		preJobName := getPreUpgradeJobName(master.Status.UpgradeStartTimeMillis)
		preJobSpec := buildPreUpgradeJobSpec(master, labels)
		job := findJob(preJobName)
		if job == nil {
			obj, err := createJob(preJobSpec)
			if err != nil {
				return nil, err
			}
			log.Printf("Version update: creating pre-upgrade job")
			return []reconciler.Object{*obj}, nil
		} else if job.Status.Succeeded > 0 {
			log.Printf("Version update: pre-upgrade job succeeded")
			setCondition(master, versionUpdateStatus.PreUpgradeJobDone)
			// Return empty to delete preUpgrade jobObj
			return []reconciler.Object{}, nil
		} else if job.Status.Failed >= versionUpgradeFailureLimit {
			// Reach terminal state
			log.Printf("Version update: pre-upgrade job failed, exceeded max retries.")
			setCondition(master, versionUpdateStatus.UpgradeFailed)
			clearCondition(master, versionUpdateStatus.Inprogress)
			return []reconciler.Object{}, nil
		} else {
			log.Printf("Version update: pre-upgrade job inprogress.")
			return []reconciler.Object{*buildObject(job)}, nil
		}
	}

	// Set the image to use
	if !isConditionTrue(master, versionUpdateStatus.VersionUpdated) {
		setVersionToUse(master)
		log.Printf("Version update: set new version.")
		setCondition(master, versionUpdateStatus.VersionUpdated)
		return []reconciler.Object{}, nil
	}

	if !isConditionTrue(master, versionUpdateStatus.PostUpgradeJobDone) {
		log.Printf("Version update: post-upgrade job not completed")
		postJobName := getPostUpgradeJobName(master.Status.UpgradeStartTimeMillis)
		postJobSpec := buildPostUpgradeJobSpec(master, labels)
		job := findJob(postJobName)
		if job == nil {
			obj, err := createJob(postJobSpec)
			if err != nil {
				return nil, err
			}
			log.Printf("Version update: creating post-upgrade job")
			return []reconciler.Object{*obj}, nil
		} else if job.Status.Succeeded > 0 {
			log.Printf("Version update: post-upgrade job succeeded")
			setCondition(master, versionUpdateStatus.PostUpgradeJobDone)
			// Return empty to delete postUpgrade jobObj
			return []reconciler.Object{}, nil
		} else if job.Status.Failed >= versionUpgradeFailureLimit {
			// Reach terminal state
			log.Printf("Version update: post-upgrade job failed, exceeded max retries.")
			setCondition(master, versionUpdateStatus.UpgradeFailed)
			clearCondition(master, versionUpdateStatus.Inprogress)
			return []reconciler.Object{*buildObject(job)}, nil
		} else {
			log.Printf("Version update: post-upgrade job inprogress.")
			return []reconciler.Object{*buildObject(job)}, nil
		}
	}
	log.Printf("Version update: upgrade succeeded.")
	setCondition(master, versionUpdateStatus.UpgradeSucceeded)
	clearCondition(master, versionUpdateStatus.Inprogress)
	return []reconciler.Object{}, nil
}

type VersionUpdateStatus struct {
	// common states
	Inprogress     status.Condition
	VersionUpdated status.Condition

	// states for upgrade
	PreUpgradeJobDone  status.Condition
	PostUpgradeJobDone status.Condition
	UpgradeSucceeded   status.Condition
	UpgradeFailed      status.Condition

	// states for downgrade
	DowngradeSucceeded status.Condition
}

func (s *VersionUpdateStatus) init() {
	log.Println("version update status init")
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
	s.PreUpgradeJobDone = status.Condition{
		Type:    "VersionUpdatePreUpgradeJobDone",
		Reason:  "Start",
		Message: "Version update pre-upgrade job is done",
	}
	s.PostUpgradeJobDone = status.Condition{
		Type:    "VersionUpdatePostUpgradeJobDone",
		Reason:  "Start",
		Message: "Version update post-upgrade job done",
	}
	s.UpgradeSucceeded = status.Condition{
		Type:    "VersionUpdateUpgradeSucceeded",
		Reason:  "Start",
		Message: "Version upgrade has completed successfully",
	}
	s.UpgradeFailed = status.Condition{
		Type:    "VersionUpdateUpgradeFailed",
		Reason:  "Start",
		Message: "Version upgrade has failed",
	}

	// States for downgrade
	s.DowngradeSucceeded = status.Condition{
		Type:    "VersionUpdateDowngradeSucceeded",
		Reason:  "Start",
		Message: "Version downgrade has succeeded",
	}

}

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

type Version struct {
	rawString string

	empty bool
	// Boolean if the Version is "latest"
	latest bool
	// List of Version integers, starting from major
	components []int
}

func extractVersion(imageString string) (*Version, error) {
	if len(imageString) == 0 {
		return &Version{
			rawString: imageString,
			empty:     true,
		}, nil
	}
	splits := strings.Split(imageString, ":")
	if len(splits) != 2 {
		return nil, fmt.Errorf("failed to parse image string %s, not in expected format of xxx:xxx", imageString)
	}

	if splits[1] == latestVersion {
		return &Version{latest: true}, nil
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

func getCurrentVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	curImage := master.Status.ImageToUse
	curVersion, err := extractVersion(curImage)
	if err != nil {
		return nil, err
	}
	return curVersion, nil
}

func getNewVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	newImage := master.Spec.Image
	newVersion, err := extractVersion(newImage)
	if err != nil {
		return nil, err
	}
	return newVersion, nil
}

func setVersionToUse(master *v1alpha1.CDAPMaster) {
	master.Status.ImageToUse = master.Spec.Image
}

func getCurrentUserInterfaceVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	curImage := master.Status.UserInterfaceImageToUse
	curVersion, err := extractVersion(curImage)
	if err != nil {
		return nil, err
	}
	return curVersion, nil
}

func getNewUserInterfaceVersion(master *v1alpha1.CDAPMaster) (*Version, error) {
	newImage := master.Spec.UserInterfaceImage
	newVersion, err := extractVersion(newImage)
	if err != nil {
		return nil, err
	}
	return newVersion, nil
}

func setUserInterfaceVersionToUse(master *v1alpha1.CDAPMaster) {
	master.Status.UserInterfaceImageToUse = master.Spec.UserInterfaceImage
}

func getCurrentTimeMs() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func getPreUpgradeJobName(startTimeMs int64) string {
	return fmt.Sprintf("pre-upgrade-job-%d", startTimeMs)
}

func getPostUpgradeJobName(startTimeMs int64) string {
	return fmt.Sprintf("post-upgrade-job-%d", startTimeMs)
}

func buildPreUpgradeJobSpec(master *v1alpha1.CDAPMaster, labels map[string]string) *UpgradeJobSpec {
	startTimeMs := master.Status.UpgradeStartTimeMillis
	cconf := getObjectName(master.Name, configMapCConf)
	hconf := getObjectName(master.Name, configMapHConf)
	name := getObjectName(master.Name, getPreUpgradeJobName(startTimeMs))
	return newUpgradeJobSpec(name, labels, startTimeMs, cconf, hconf, master).SetPreUpgrade(true)
}

func buildPostUpgradeJobSpec(master *v1alpha1.CDAPMaster, labels map[string]string) *UpgradeJobSpec {
	startTimeMs := master.Status.UpgradeStartTimeMillis
	cconf := getObjectName(master.Name, configMapCConf)
	hconf := getObjectName(master.Name, configMapHConf)
	name := getObjectName(master.Name, getPostUpgradeJobName(startTimeMs))
	return newUpgradeJobSpec(name, labels, startTimeMs, cconf, hconf, master).SetPostUpgrade(true)
}

func buildUpgradeJobObject(spec *UpgradeJobSpec) (*reconciler.Object, error) {
	obj, err := k8s.ObjectFromFile(templateDir+upgradeJobTemplate, spec, &batchv1.JobList{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
