package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller Suite", func() {
	Describe("Reflection to get CDAPServiceSpec", func() {
		var (
			master                *v1alpha1.CDAPMaster
			serviceToSpec         map[string]*v1alpha1.CDAPServiceSpec
			serviceToScalableSpec map[string]*v1alpha1.CDAPScalableServiceSpec
			serviceToStatefulSpec map[string]*v1alpha1.CDAPStatefulServiceSpec
			serviceToExternalSpec map[string]*v1alpha1.CDAPExternalServiceSpec
		)
		BeforeEach(func() {
			master = &v1alpha1.CDAPMaster{}
			serviceToSpec = map[string]*v1alpha1.CDAPServiceSpec{
				serviceLogs:          &master.Spec.Logs.CDAPServiceSpec,
				serviceAppFabric:     &master.Spec.AppFabric.CDAPServiceSpec,
				serviceMetrics:       &master.Spec.Metrics.CDAPServiceSpec,
				serviceRouter:        &master.Spec.Router.CDAPServiceSpec,
				serviceMessaging:     &master.Spec.Messaging.CDAPServiceSpec,
				serviceMetadata:      &master.Spec.Metadata.CDAPServiceSpec,
				servicePreview:       &master.Spec.Preview.CDAPServiceSpec,
				serviceUserInterface: &master.Spec.UserInterface.CDAPServiceSpec,
			}
			serviceToScalableSpec = map[string]*v1alpha1.CDAPScalableServiceSpec{
				serviceLogs:          nil,
				serviceAppFabric:     &master.Spec.AppFabric.CDAPScalableServiceSpec,
				serviceMetrics:       nil,
				serviceRouter:        &master.Spec.Router.CDAPScalableServiceSpec,
				serviceMessaging:     nil,
				serviceMetadata:      &master.Spec.Metadata.CDAPScalableServiceSpec,
				servicePreview:       nil,
				serviceUserInterface: &master.Spec.UserInterface.CDAPScalableServiceSpec,
			}
			serviceToStatefulSpec = map[string]*v1alpha1.CDAPStatefulServiceSpec{
				serviceLogs:          &master.Spec.Logs.CDAPStatefulServiceSpec,
				serviceAppFabric:     nil,
				serviceMetrics:       &master.Spec.Metrics.CDAPStatefulServiceSpec,
				serviceRouter:        nil,
				serviceMessaging:     &master.Spec.Messaging.CDAPStatefulServiceSpec,
				serviceMetadata:      nil,
				servicePreview:       &master.Spec.Preview.CDAPStatefulServiceSpec,
				serviceUserInterface: nil,
			}
			serviceToExternalSpec = map[string]*v1alpha1.CDAPExternalServiceSpec{
				serviceLogs:          nil,
				serviceAppFabric:     nil,
				serviceMetrics:       nil,
				serviceRouter:        &master.Spec.Router.CDAPExternalServiceSpec,
				serviceMessaging:     nil,
				serviceMetadata:      nil,
				servicePreview:       nil,
				serviceUserInterface: &master.Spec.UserInterface.CDAPExternalServiceSpec,
			}
		})
		It("Successfully get pointer to CDAPServiceSpec", func() {
			for service, expectedSpec := range serviceToSpec {
				spec, err := getCDAPServiceSpec(master, service)
				Expect(err).To(BeNil())
				Expect(spec).To(Equal(expectedSpec))
			}
		})
		It("Failed get pointer to CDAPServiceSpec due to invalid field name", func() {
			invalidService := "InvalidService"
			spec, err := getCDAPServiceSpec(master, invalidService)
			Expect(err).NotTo(BeNil())
			Expect(spec).To(BeNil())
		})
		It("Successfully get pointer to CDAPScalableServiceSpec", func() {
			for service, expectedSpec := range serviceToScalableSpec {
				spec, err := getCDAPScalableServiceSpec(master, service)
				Expect(err).To(BeNil())
				Expect(spec).To(Equal(expectedSpec))
			}
		})
		It("Failed get pointer to CDAPScalableServiceSpec due to invalid field name", func() {
			invalidService := "InvalidService"
			spec, err := getCDAPScalableServiceSpec(master, invalidService)
			Expect(err).NotTo(BeNil())
			Expect(spec).To(BeNil())
		})
		It("Successfully get pointer to CDAPStatefulServiceSpec", func() {
			for service, expectedSpec := range serviceToStatefulSpec {
				spec, err := getCDAPStatefulServiceSpec(master, service)
				Expect(err).To(BeNil())
				Expect(spec).To(Equal(expectedSpec))
			}
		})
		It("Failed get pointer to CDAPStatefulServiceSpec", func() {
			invalidService := "InvalidService"
			spec, err := getCDAPStatefulServiceSpec(master, invalidService)
			Expect(err).NotTo(BeNil())
			Expect(spec).To(BeNil())
		})
		It("Successfully get pointer to CDAPExternalServiceSpec", func() {
			for service, expectedSpec := range serviceToExternalSpec {
				spec, err := getCDAPExternalServiceSpec(master, service)
				Expect(err).To(BeNil())
				Expect(spec).To(Equal(expectedSpec))
			}
		})
		It("Failed get pointer to CDAPExternalServiceSpec", func() {
			invalidService := "InvalidService"
			spec, err := getCDAPExternalServiceSpec(master, invalidService)
			Expect(err).NotTo(BeNil())
			Expect(spec).To(BeNil())
		})
	})
})
