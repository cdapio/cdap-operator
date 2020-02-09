package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	"encoding/json"
	"fmt"
	"github.com/nsf/jsondiff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler/manager/k8s"
)

func fromJson(filename string, obj interface{}) error {
	file, _ := ioutil.ReadFile(filename)
	if err := json.Unmarshal([]byte(file), obj); err != nil {
		return fmt.Errorf("failed to read expected spec json file %v", err)
	}
	return nil
}

var _ = Describe("Controller Suite", func() {
	Describe("NumPods == 0", func() {
		It("k8s objects", func() {
			emptyLabels := make(map[string]string)

			master := &v1alpha1.CDAPMaster{}
			err := fromJson("testdata/cdap_master_cr_num_pods_0.json", master)
			Expect(err).To(BeNil())

			spec, err := buildCDAPDeploymentSpec(master, emptyLabels)
			Expect(err).To(BeNil())

			objs, err := buildObjects(spec)
			Expect(err).To(BeNil())

			readExpectedJson := func(fileName string) []byte {
				json, err := ioutil.ReadFile("testdata/" + fileName)
				Expect(err).To(BeNil())
				return json
			}

			diffJson := func(expected, actual []byte) {
				opts := jsondiff.DefaultConsoleOptions()
				diff, _ := jsondiff.Compare(expected, actual, &opts)
				Expect(diff.String()).To(Equal(jsondiff.SupersetMatch.String()))
			}

			var strategyHandler DeploymentStrategy
			strategyHandler.Init()
			serviceGroupMap, err := strategyHandler.getStrategy(0)

			for _, obj := range objs {
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*appsv1.StatefulSet); ok {
					for k, _ := range serviceGroupMap.stateful {
						if o.Name == getObjectName(master.Name, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + ".json")
							diffJson(expected, actual)
						}
					}
				}
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*appsv1.Deployment); ok {
					for k, _ := range serviceGroupMap.deployment {
						if o.Name == getObjectName(master.Name, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + ".json")
							diffJson(expected, actual)
						}
					}
				}
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
					for k, _ := range serviceGroupMap.networkService {
						if o.Name == getObjectName(master.Name, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + "_service.json")
							diffJson(expected, actual)
						}
					}
				}
			}
		})
	})
})
