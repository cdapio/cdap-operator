package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"cdap.io/cdap-operator/api/v1alpha1"
	"github.com/nsf/jsondiff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	Describe("Fully distributed one service per pod", func() {
		var (
			master *v1alpha1.CDAPMaster
		)
		BeforeEach(func() {
			master = &v1alpha1.CDAPMaster{}
			err := fromJson("testdata/cdap_master_cr.json", master)
			Expect(err).To(BeNil())
		})
		readExpectedJson := func(fileName string) []byte {
			json, err := ioutil.ReadFile("testdata/" + fileName)
			Expect(err).To(BeNil())
			return json
		}
		diffJson := func(expected, actual []byte) {
			opts := jsondiff.DefaultConsoleOptions()
			diff, text := jsondiff.Compare(expected, actual, &opts)
			Expect(diff.String()).To(Equal(jsondiff.SupersetMatch.String()), text)
		}
		It("k8s objs for all services including essential and optional", func() {
			emptyLabels := make(map[string]string)
			spec, err := buildDeploymentPlanSpec(master, emptyLabels)
			Expect(err).To(BeNil())
			objs, err := buildObjectsForDeploymentPlan(spec)
			Expect(err).To(BeNil())

			var strategyHandler DeploymentPlan
			strategyHandler.Init()
			serviceGroupMap, err := strategyHandler.getPlan(0)

			totalServiceCount := len(serviceGroupMap.stateful) + len(serviceGroupMap.deployment) + len(serviceGroupMap.networkService)
			expectedCount := totalServiceCount
			actualCount := 0
			for _, obj := range objs {
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*appsv1.StatefulSet); ok {
					for k, _ := range serviceGroupMap.stateful {
						if o.Name == getObjName(master, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + ".json")
							diffJson(expected, actual)
							actualCount++
						}
					}
				}
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*appsv1.Deployment); ok {
					for k, _ := range serviceGroupMap.deployment {
						if o.Name == getObjName(master, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + ".json")
							diffJson(expected, actual)
							actualCount++
						}
					}
				}
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
					for k, _ := range serviceGroupMap.networkService {
						if o.Name == getObjName(master, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + "_service.json")
							diffJson(expected, actual)
							actualCount++
						}
					}
				}
			}
			Expect(expectedCount).To(Equal(actualCount))
		})
		It("k8s objs for just essential services", func() {
			master.Spec.Runtime = nil
			numOptionalServices := 1

			emptyLabels := make(map[string]string)
			spec, err := buildDeploymentPlanSpec(master, emptyLabels)
			Expect(err).To(BeNil())
			objs, err := buildObjectsForDeploymentPlan(spec)
			Expect(err).To(BeNil())

			var strategyHandler DeploymentPlan
			strategyHandler.Init()
			serviceGroupMap, err := strategyHandler.getPlan(0)

			totalServiceCount := len(serviceGroupMap.stateful) + len(serviceGroupMap.deployment) + len(serviceGroupMap.networkService)
			expectedCount := totalServiceCount - numOptionalServices

			actualCount := 0
			for _, obj := range objs {
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*appsv1.StatefulSet); ok {
					for k, _ := range serviceGroupMap.stateful {
						if o.Name == getObjName(master, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + ".json")
							diffJson(expected, actual)
							actualCount++
						}
					}
				}
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*appsv1.Deployment); ok {
					for k, _ := range serviceGroupMap.deployment {
						if o.Name == getObjName(master, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + ".json")
							diffJson(expected, actual)
							actualCount++
						}
					}
				}
				if o, ok := obj.Obj.(*k8s.Object).Obj.(*corev1.Service); ok {
					for k, _ := range serviceGroupMap.networkService {
						if o.Name == getObjName(master, k) {
							actual, err := json.Marshal(o)
							Expect(err).To(BeNil())
							expected := readExpectedJson(k + "_service.json")
							diffJson(expected, actual)
							actualCount++
						}
					}
				}
			}
			Expect(expectedCount).To(Equal(actualCount))
		})
	})

	Describe("Set java max heap size env var", func() {
		var (
			envVar    []corev1.EnvVar
			resources *corev1.ResourceRequirements
		)
		BeforeEach(func() {
			envVar = []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "some_env_var_name",
					Value: "some_env_var_value",
				},
			}
			resources = &corev1.ResourceRequirements{
				Limits:   make(map[corev1.ResourceName]resource.Quantity),
				Requests: make(map[corev1.ResourceName]resource.Quantity),
			}
			resources.Requests.Memory().Add(*resource.NewQuantity(4*gigaBytes, resource.BinarySI))
			resources.Limits.Memory().Add(*resource.NewQuantity(8*gigaBytes, resource.BinarySI))
		})
		It("java max heap size already set", func() {
			envOld := append(envVar, corev1.EnvVar{
				Name:  javaMaxHeapSizeEnvVarName,
				Value: "-Xmx1024m",
			})
			envNew := addJavaMaxHeapEnvIfNotPresent(envOld, resources)
			Expect(envNew).To(Equal(envOld))
		})
		It("java max heap size added", func() {
			envNew := addJavaMaxHeapEnvIfNotPresent(envVar, resources)
			Expect(envNew).To(Equal(envVar))
		})
	})
	Describe("Extract field value from CDAPServiceSpec", func() {
		var (
			// Make the value the same as field name to simplifying the tests below
			serviceAccountName = "ServiceAccountName"
			runtimeClassName   = "RuntimeClassName"
			priorityClassName  = "PriorityClassName"
			invalidFiledValue  = "some_invalid_field_value"
			master             *v1alpha1.CDAPMaster
			emptyMaster        *v1alpha1.CDAPMaster
			invalidMaster      *v1alpha1.CDAPMaster
			services           ServiceGroup
		)
		BeforeEach(func() {
			emptyMaster = &v1alpha1.CDAPMaster{}
			master = &v1alpha1.CDAPMaster{
				Spec: v1alpha1.CDAPMasterSpec{
					ServiceAccountName: "service_account_name",
					Logs: v1alpha1.LogsSpec{
						CDAPStatefulServiceSpec: v1alpha1.CDAPStatefulServiceSpec{
							CDAPServiceSpec: v1alpha1.CDAPServiceSpec{
								ServiceAccountName: serviceAccountName,
								NodeSelector:       nil,
								RuntimeClassName:   &runtimeClassName,
								PriorityClassName:  &priorityClassName,
							},
						},
					},
					Metrics: v1alpha1.MetricsSpec{
						CDAPStatefulServiceSpec: v1alpha1.CDAPStatefulServiceSpec{
							CDAPServiceSpec: v1alpha1.CDAPServiceSpec{
								ServiceAccountName: serviceAccountName,
								NodeSelector:       nil,
								RuntimeClassName:   &runtimeClassName,
								PriorityClassName:  &priorityClassName,
							},
						},
					},
				},
			}
			invalidMaster = &v1alpha1.CDAPMaster{}
			*invalidMaster = *master
			invalidMaster.Spec.Logs.CDAPServiceSpec.ServiceAccountName = invalidFiledValue
			invalidMaster.Spec.Logs.CDAPServiceSpec.RuntimeClassName = &invalidFiledValue
			invalidMaster.Spec.Logs.CDAPServiceSpec.PriorityClassName = &invalidFiledValue
			// Adding AppFabric intentionally to test the case where fields are unset for one of the service
			services = ServiceGroup{serviceLogs, serviceMetrics, serviceAppFabric, serviceSupportBundle}
		})
		It("Extract empty field value", func() {
			for _, field := range []string{"ServiceAccountName", "RuntimeClassName", "PriorityClassName"} {
				val, err := getFieldValueIfUnique(emptyMaster, services, field)
				Expect(err).To(BeNil())
				Expect(val).To(BeNil())
			}
		})
		It("Extract valid non-empty field value", func() {
			for _, field := range []string{"ServiceAccountName", "RuntimeClassName", "PriorityClassName"} {
				val, err := getFieldValueIfUnique(master, services, field)
				Expect(err).To(BeNil())
				val, ok := val.(string)
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal(field))
			}
		})
		It("Extract invalid non-empty field value", func() {
			for _, field := range []string{"ServiceAccountName", "RuntimeClassName", "PriorityClassName"} {
				_, err := getFieldValueIfUnique(invalidMaster, services, field)
				Expect(err).NotTo(BeNil())
			}
		})
	})
})
