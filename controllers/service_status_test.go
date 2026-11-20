package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"cdap.io/cdap-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller Suite", func() {
	Describe("CDAP service status", func() {
		It("CDAP is unavailable", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer ts.Close()

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatus(master)
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(true))
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(false))
		})
		It("Some CDAP service are not running", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"appfabric": "OK"}`)
			}))
			defer ts.Close()
			cdapServiceQuery = ts.URL

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatusHelper(master)
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(true))
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(false))
		})
		It("Some CDAP services are unhealthy", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"appfabric": "NOTOK", "metrics": "OK", "metrics.processor": "OK", "log.saver": "OK", "dataset.executor": "OK", "runtime": "OK", "messaging.service": "OK", "metadata.service": "OK"}`)
			}))
			defer ts.Close()
			cdapServiceQuery = ts.URL

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatusHelper(master)
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(true))
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(false))
		})
		It("CDAP system service status is not available", func() {
			ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"appfabric": "OK", "metrics": "OK", "metrics.processor": "OK", "log.saver": "OK", "dataset.executor": "OK", "runtime": "OK", "messaging.service": "OK", "metadata.service": "OK"}`)
			}))
			defer ts1.Close()
			cdapServiceQuery = ts1.URL

			ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer ts2.Close()
			cdapSystemServiceQuery = ts2.URL

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatusHelper(master)
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(true))
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(false))
		})
		It("Dataprep service is not running", func() {
			ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"appfabric": "OK", "metrics": "OK", "metrics.processor": "OK", "log.saver": "OK", "dataset.executor": "OK", "runtime": "OK", "messaging.service": "OK", "metadata.service": "OK"}`)
			}))
			defer ts1.Close()
			cdapServiceQuery = ts1.URL

			ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `[{"status": "Running", "statusCode": 200, "appId": "pipeline", "programType": "Service", "programId": "studio"}]`)
			}))
			defer ts2.Close()
			cdapSystemServiceQuery = ts2.URL

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatusHelper(master)
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(true))
		})

		It("Dataprep service is not healthy", func() {
			ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"appfabric": "OK", "metrics": "OK", "metrics.processor": "OK", "log.saver": "OK", "dataset.executor": "OK", "runtime": "OK", "messaging.service": "OK", "metadata.service": "OK"}`)
			}))
			defer ts1.Close()
			cdapServiceQuery = ts1.URL

			ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `[{"error": "Dataprep not found", "statusCode": 404, "appId": "dataprep", "programType": "service", "programId": "service"}, {"status": "Running", "statusCode": 200, "appId": "pipeline", "programType": "Service", "programId": "studio"}]`)
			}))
			defer ts2.Close()
			cdapSystemServiceQuery = ts2.URL

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatusHelper(master)
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(true))
		})
		It("CDAP is ready", func() {
			ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"appfabric": "OK", "metrics": "OK", "metrics.processor": "OK", "log.saver": "OK", "dataset.executor": "OK", "runtime": "OK", "messaging.service": "OK", "metadata.service": "OK"}`)
			}))
			defer ts1.Close()
			cdapServiceQuery = ts1.URL

			ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `[{"status": "RUNNING", "statusCode": 200, "appId": "dataprep", "programType": "service", "programId": "service"}, {"status": "Running", "statusCode": 200, "appId": "pipeline", "programType": "Service", "programId": "studio"}]`)
			}))
			defer ts2.Close()
			cdapSystemServiceQuery = ts2.URL

			master := &v1alpha1.CDAPMaster{}
			updateServiceStatusHelper(master)
			Expect(isConditionTrue(master, serviceStatus.Ready)).To(Equal(true))
			Expect(isConditionTrue(master, serviceStatus.Unknown)).To(Equal(false))
			Expect(isConditionTrue(master, serviceStatus.Unhealthy)).To(Equal(false))
		})
	})
})
