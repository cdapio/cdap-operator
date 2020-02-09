package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	"encoding/json"
	"fmt"
	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	"sigs.k8s.io/controller-reconciler/pkg/status"
	"testing"
)


func TestParsingImageString(t *testing.T) {
	assert := assert.New(t)

	{
		image := ""
		version, err := parseImageString(image)
		assert.Nil(err)
		assert.Equal(image, version.rawString)
		assert.False(version.latest)
	}
	{
		image := "gcr.io/cdapio/cdap:latest"
		version, err := parseImageString(image)
		assert.Nil(err)
		assert.Equal(image, version.rawString)
		assert.True(version.latest)
	}
	{
		image := "gcr.io/cdapio/cdap:6.0.0.0"
		version, err := parseImageString(image)
		assert.Nil(err)
		assert.Equal(image, version.rawString)
		assert.False(version.latest)
		assert.Equal([]int{6, 0, 0, 0}, version.components)
	}
	{
		imagePairs := []Pair{
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:latest"},
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.0.0.1"},
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.1.0"},
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:7"},
		}
		for _, imagePair := range imagePairs {
			low, err := parseImageString(imagePair.first.(string))
			assert.Nil(err)
			high, err := parseImageString(imagePair.second.(string))
			assert.Nil(err)
			assert.Equal(-1, compareVersion(low, high), fmt.Sprintf("%s < %s failed", low.rawString, high.rawString))
			assert.Equal(1, compareVersion(high, low), fmt.Sprintf("%s > %s failed", high.rawString, low.rawString))
		}
	}
	{
		imagePairs := []Pair{
			Pair{"gcr.io/cdapio/cdap:latest", "gcr.io/cdapio/cdap:latest"},
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.0.0"},
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.0"},
			Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6"},
		}
		for _, imagePair := range imagePairs {
			first, err := parseImageString(imagePair.first.(string))
			assert.Nil(err)
			second, err := parseImageString(imagePair.second.(string))
			assert.Nil(err)
			assert.Equal(0, compareVersion(first, second), fmt.Sprintf("%s == %s failed", first.rawString, second.rawString))
		}
	}
	{
		invalidImage := "gcr.io/cdapio/cdap"
		_, err := parseImageString(invalidImage)
		assert.NotNil(err)
	}
	{
		invalidImage := "gcr.io/cdapio/cdap:6:5:4"
		_, err := parseImageString(invalidImage)
		assert.NotNil(err)
	}
	{
		invalidImage := "gcr.io/cdapio/v6.1.0-rc2"
		_, err := parseImageString(invalidImage)
		assert.NotNil(err)
	}
	{
		invalidImage := "gcr.io/cdapio/6.2.0-SNAPSHOT"
		_, err := parseImageString(invalidImage)
		assert.NotNil(err)
	}

}

func TestGetVersionFromMasterSpec(t *testing.T) {
	newImage := "gcr.io/cdapio/cdap:6.2.0.0"
	newUIImage := "gcr.io/cdapio/cdap:6.2.0.0"

	curImage := "gcr.io/cdapio/cdap:6.1.0.0"
	curUIImage := "gcr.io/cdapio/cdap:6.1.0.0"

	assert := assert.New(t)

	master := &v1alpha1.CDAPMaster{
		Spec:       v1alpha1.CDAPMasterSpec{
			Image:              newImage,
			UserInterfaceImage: newUIImage,
		},
		Status:     v1alpha1.CDAPMasterStatus{
			ImageToUse:              curImage,
			UserInterfaceImageToUse: curUIImage,
		},
	}

	v, err := getNewImageVersion(master)
	assert.Nil(err)
	assert.Equal(newImage, v.rawString)

	v, err = getNewUserInterfaceVersion(master)
	assert.Nil(err)
	assert.Equal(newUIImage, v.rawString)

	v, err = getCurrentImageVersion(master)
	assert.Nil(err)
	assert.Equal(curImage, v.rawString)

	v, err = getCurrentUserInterfaceVersion(master)
	assert.Nil(err)
	assert.Equal(curUIImage, v.rawString)
}

func TestVersionUpdateStatusConditions(t *testing.T) {
	assert := assert.New(t)

	s := new(VersionUpdateStatus)
	s.init()

	master := &v1alpha1.CDAPMaster{}

	conditions := reflect.ValueOf(*s)
	for i := 0; i < conditions.NumField(); i++ {
		condition, ok := conditions.Field(i).Interface().(status.Condition)
		assert.True(ok)
		assert.False(isConditionTrue(master, condition))
		setCondition(master, condition)
		assert.True(isConditionTrue(master, condition))
		clearCondition(master, condition)
		assert.False(isConditionTrue(master, condition))
		setCondition(master, condition)
	}
	s.clearAllConditions(master)
	for i := 0; i < conditions.NumField(); i++ {
		condition, ok := conditions.Field(i).Interface().(status.Condition)
		assert.True(ok)
		assert.False(isConditionTrue(master, condition))
	}
}

func TestUpdateImageForUserInterface(t *testing.T) {
	assert := assert.New(t)
	newUIImage := "gcr.io/cdapio/cdap:6.0.0.0"
	curUIImage := "gcr.io/cdapio/cdap:6.1.0.0"

	master := &v1alpha1.CDAPMaster{
		Spec: v1alpha1.CDAPMasterSpec{
			UserInterfaceImage: newUIImage,
		},
		Status: v1alpha1.CDAPMasterStatus{
			UserInterfaceImageToUse: curUIImage,
		},
	}

	objs, updated, err := updateForUserInterface(master)
	assert.Equal([]reconciler.Object{}, objs)
	assert.True(updated)
	assert.Nil(err)
	assert.Equal(newUIImage, master.Status.UserInterfaceImageToUse)

	objs, updated, err = updateForUserInterface(master)
	assert.Equal([]reconciler.Object{}, objs)
	assert.False(updated)
	assert.Nil(err)
	assert.Equal(newUIImage, master.Status.UserInterfaceImageToUse)
}


func TestPostUpgradeJob(t *testing.T) {
	const upgradeTimeMs int64 = 1581238669880
	const newUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.5"
	const curUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.0"
	const name = "test"
	emptyLabels := make(map[string]string)
	assert := assert.New(t)

	master := &v1alpha1.CDAPMaster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.CDAPMasterSpec{
			Image: newUIImage,
		},
		Status: v1alpha1.CDAPMasterStatus{
			UpgradeStartTimeMillis: upgradeTimeMs,
			ImageToUse: curUIImage,
		},
	}
	postJobSpec := buildUpgradeJobSpec(getPostUpgradeJobName(master.Status.UpgradeStartTimeMillis), master, emptyLabels)
	object, err:= buildUpgradeJobObject(postJobSpec)
	assert.Nil(err)

	json, _ := json.Marshal(object.Obj.(*k8s.Object).Obj.(*batchv1.Job))
	expectedJson, err := ioutil.ReadFile("testdata/post_upgrade_job.json")
	opts := jsondiff.DefaultConsoleOptions()
	diff, text := jsondiff.Compare(expectedJson, json, &opts)
	assert.Equal(jsondiff.SupersetMatch, diff, text)
}

func TestPreUpgradeJob(t *testing.T) {
	const upgradeTimeMs int64 = 1581238669880
	const newUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.5"
	const curUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.0"
	const name = "test"
	emptyLabels := make(map[string]string)
	assert := assert.New(t)

	master := &v1alpha1.CDAPMaster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.CDAPMasterSpec{
			Image: newUIImage,
		},
		Status: v1alpha1.CDAPMasterStatus{
			UpgradeStartTimeMillis: upgradeTimeMs,
			ImageToUse: curUIImage,
		},
	}
	postJobSpec := buildUpgradeJobSpec(getPreUpgradeJobName(master.Status.UpgradeStartTimeMillis), master, emptyLabels)
	object, err:= buildUpgradeJobObject(postJobSpec)
	assert.Nil(err)

	json, _ := json.Marshal(object.Obj.(*k8s.Object).Obj.(*batchv1.Job))
	expectedJson, err := ioutil.ReadFile("testdata/pre_upgrade_job.json")
	opts := jsondiff.DefaultConsoleOptions()
	diff, text := jsondiff.Compare(expectedJson, json, &opts)
	assert.Equal(jsondiff.SupersetMatch, diff, text)
}





