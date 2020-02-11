package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller Suite", func() {
	Describe("Reflection to get CDAPServiceSpec", func() {
		It("Successfully get pointer to CDAPServiceSpec", func() {
			master := &v1alpha1.CDAPMaster{}
			serviceToSpec := map[string]*v1alpha1.CDAPServiceSpec{
				serviceLogs: &master.Spec.Logs.CDAPServiceSpec,
				serviceAppFabric: &master.Spec.AppFabric.CDAPServiceSpec,
				serviceMetrics: &master.Spec.Metrics.CDAPServiceSpec,
				serviceRouter: &master.Spec.Router.CDAPServiceSpec,
				serviceMessaging: &master.Spec.Messaging.CDAPServiceSpec,
				serviceMetadata: &master.Spec.Metadata.CDAPServiceSpec,
				servicePreview: &master.Spec.Preview.CDAPServiceSpec,
				serviceUserInterface: &master.Spec.UserInterface.CDAPServiceSpec,
			}
			for service, expectedSpec := range serviceToSpec {
				spec, err := getCDAPServiceSpec(master, service)
				Expect(err).To(BeNil())
				Expect(spec).To(Equal(expectedSpec))
			}
		})
		It("Successfully get pointer to CDAPServiceSpec", func() {
			master := &v1alpha1.CDAPMaster{}
			invalidService := "InvalidService"
			spec, err := getCDAPServiceSpec(master, invalidService)
			Expect(err).NotTo(BeNil())
			Expect(spec).To(BeNil())
		})
	})
})
