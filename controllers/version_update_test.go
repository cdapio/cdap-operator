package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	"encoding/json"
	"github.com/nsf/jsondiff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
	"sigs.k8s.io/controller-reconciler/pkg/status"
)

var _ = Describe("Controller Suite", func() {
	Describe("Parsing image string", func() {
		It("Parse empty image string", func() {
			image := ""
			version, err := parseImageString(image)
			Expect(err).To(BeNil())
			Expect(image).To(Equal(version.rawString))
			Expect(version.latest).To(BeFalse())
		})
		It("Parse latest image string", func() {
			image := "gcr.io/cdapio/cdap:latest"
			version, err := parseImageString(image)
			Expect(err).To(BeNil())
			Expect(image).To(Equal(version.rawString))
			Expect(version.latest).To(BeTrue())
		})
		It("Parse normal image string", func() {
			image := "gcr.io/cdapio/cdap:6.0.0.0"
			version, err := parseImageString(image)
			Expect(err).To(BeNil())
			Expect(image).To(Equal(version.rawString))
			Expect(version.latest).To(BeFalse())
			Expect(version.components).To(Equal([]int{6, 0, 0, 0}))
		})
		It("Compare image versions", func() {
			imagePairs := []Pair{
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:latest"},
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.0.0.1"},
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.1.0"},
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:7"},
			}
			for _, imagePair := range imagePairs {
				low, err := parseImageString(imagePair.first.(string))
				Expect(err).To(BeNil())
				high, err := parseImageString(imagePair.second.(string))
				Expect(err).To(BeNil())
				Expect(compareVersion(low, high)).To(Equal(-1))
				Expect(compareVersion(high, low)).To(Equal(1))
			}
		})
		It("Compare same image versions", func() {
			imagePairs := []Pair{
				Pair{"gcr.io/cdapio/cdap:latest", "gcr.io/cdapio/cdap:latest"},
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.0.0"},
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6.0"},
				Pair{"gcr.io/cdapio/cdap:6.0.0.0", "gcr.io/cdapio/cdap:6"},
			}
			for _, imagePair := range imagePairs {
				first, err := parseImageString(imagePair.first.(string))
				Expect(err).To(BeNil())
				second, err := parseImageString(imagePair.second.(string))
				Expect(err).To(BeNil())
				Expect(compareVersion(first, second)).To(Equal(0))
			}
		})
		It("Fail to parse invalid image string", func() {
			invalidImage := "gcr.io/cdapio/cdap"
			_, err := parseImageString(invalidImage)
			Expect(err).NotTo(BeNil())
		})
		It("Fail to parse invalid image string", func() {
			invalidImage := "gcr.io/cdapio/cdap:6:5:4"
			_, err := parseImageString(invalidImage)
			Expect(err).NotTo(BeNil())

		})
		It("Fail to parse invalid image string", func() {
			invalidImage := "gcr.io/cdapio/v6.1.0-rc2"
			_, err := parseImageString(invalidImage)
			Expect(err).NotTo(BeNil())
		})
		It("Fail to parse invalid image string", func() {
			invalidImage := "gcr.io/cdapio/6.2.0-SNAPSHOT"
			_, err := parseImageString(invalidImage)
			Expect(err).NotTo(BeNil())
		})
	})
	Describe("Get versions from master spec", func() {
		It("Get versions", func() {
			newImage := "gcr.io/cdapio/cdap:6.2.0.0"
			newUIImage := "gcr.io/cdapio/cdap:6.2.0.0"

			curImage := "gcr.io/cdapio/cdap:6.1.0.0"
			curUIImage := "gcr.io/cdapio/cdap:6.1.0.0"
			master := &v1alpha1.CDAPMaster{
				Spec: v1alpha1.CDAPMasterSpec{
					Image:              newImage,
					UserInterfaceImage: newUIImage,
				},
				Status: v1alpha1.CDAPMasterStatus{
					ImageToUse:              curImage,
					UserInterfaceImageToUse: curUIImage,
				},
			}

			v, err := getNewImageVersion(master)
			Expect(err).To(BeNil())
			Expect(newImage).To(Equal(v.rawString))

			v, err = getNewUserInterfaceVersion(master)
			Expect(err).To(BeNil())
			Expect(newUIImage).To(Equal(v.rawString))

			v, err = getCurrentImageVersion(master)
			Expect(err).To(BeNil())
			Expect(curImage).To(Equal(v.rawString))

			v, err = getCurrentUserInterfaceVersion(master)
			Expect(err).To(BeNil())
			Expect(curUIImage).To(Equal(v.rawString))
		})
	})
	Describe("Version update status conditions", func() {
		It("", func() {
			s := new(VersionUpdateStatus)
			s.init()

			master := &v1alpha1.CDAPMaster{}

			conditions := reflect.ValueOf(*s)
			for i := 0; i < conditions.NumField(); i++ {
				condition, ok := conditions.Field(i).Interface().(status.Condition)
				Expect(ok).To(BeTrue())
				Expect(isConditionTrue(master, condition)).To(BeFalse())
				setCondition(master, condition)
				Expect(isConditionTrue(master, condition)).To(BeTrue())
				clearCondition(master, condition)
				Expect(isConditionTrue(master, condition)).To(BeFalse())
				setCondition(master, condition)
			}
			s.clearAllConditions(master)
			for i := 0; i < conditions.NumField(); i++ {
				condition, ok := conditions.Field(i).Interface().(status.Condition)
				Expect(ok).To(BeTrue())
				Expect(isConditionTrue(master, condition)).To(BeFalse())
			}
		})
	})
	Describe("Update image for user interface", func() {
		It("", func() {
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
			Expect(objs).To(Equal([]reconciler.Object{}))
			Expect(updated).To(BeTrue())
			Expect(err).To(BeNil())
			Expect(master.Status.UserInterfaceImageToUse).To(Equal(newUIImage))

			objs, updated, err = updateForUserInterface(master)
			Expect(objs).To(Equal([]reconciler.Object{}))
			Expect(updated).To(BeFalse())
			Expect(err).To(BeNil())
			Expect(master.Status.UserInterfaceImageToUse).To(Equal(newUIImage))
		})
	})
	Describe("Pre upgrade job", func() {
		It("k8s object", func() {
			const upgradeTimeMs int64 = 1581238669880
			const newUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.5"
			const curUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.0"
			const name = "test"
			emptyLabels := make(map[string]string)
			master := &v1alpha1.CDAPMaster{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: v1alpha1.CDAPMasterSpec{
					Image: newUIImage,
				},
				Status: v1alpha1.CDAPMasterStatus{
					UpgradeStartTimeMillis: upgradeTimeMs,
					ImageToUse:             curUIImage,
				},
			}
			postJobSpec := buildPreUpgradeJobSpec(getPreUpgradeJobName(master.Status.UpgradeStartTimeMillis), master, emptyLabels)
			object, err := buildUpgradeJobObject(postJobSpec)
			Expect(err).To(BeNil())

			json, _ := json.Marshal(object.Obj.(*k8s.Object).Obj.(*batchv1.Job))
			expectedJson, err := ioutil.ReadFile("testdata/pre_upgrade_job.json")
			opts := jsondiff.DefaultConsoleOptions()
			diff, text := jsondiff.Compare(expectedJson, json, &opts)
			Expect(diff.String()).To(Equal(jsondiff.SupersetMatch.String()), text)
		})
	})
	Describe("Post upgrade job", func() {
		It("k8s object", func() {
			const upgradeTimeMs int64 = 1581238669880
			const newUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.5"
			const curUIImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.1.0.0"
			const name = "test"
			emptyLabels := make(map[string]string)

			master := &v1alpha1.CDAPMaster{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: v1alpha1.CDAPMasterSpec{
					Image: newUIImage,
				},
				Status: v1alpha1.CDAPMasterStatus{
					UpgradeStartTimeMillis: upgradeTimeMs,
					ImageToUse:             curUIImage,
				},
			}
			postJobSpec := buildPostUpgradeJobSpec(getPostUpgradeJobName(master.Status.UpgradeStartTimeMillis), master, emptyLabels)
			object, err := buildUpgradeJobObject(postJobSpec)
			Expect(err).To(BeNil())

			json, _ := json.Marshal(object.Obj.(*k8s.Object).Obj.(*batchv1.Job))
			expectedJson, err := ioutil.ReadFile("testdata/post_upgrade_job.json")
			opts := jsondiff.DefaultConsoleOptions()
			diff, text := jsondiff.Compare(expectedJson, json, &opts)
			Expect(diff.String()).To(Equal(jsondiff.SupersetMatch.String()), text)
		})
	})
})
