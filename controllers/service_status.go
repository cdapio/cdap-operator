package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cdap.io/cdap-operator/api/v1alpha1"
	"sigs.k8s.io/controller-reconciler/pkg/status"
)

type CDAPService = string

const (
	appFabricSvc        CDAPService = "appfabric"
	datasetExecutorSvc  CDAPService = "dataset.executor"
	logSaverSvc         CDAPService = "log.saver"
	messagingSvc        CDAPService = "messaging.service"
	metadataSvc         CDAPService = "metadata.service"
	metricsSvc          CDAPService = "metrics"
	metricsProcessorSvc CDAPService = "metrics.processor"
	runtimeSvc          CDAPService = "runtime"
)

var (
	serviceStatus          *CDAPServiceStatus
	cdapServices           []CDAPService
	cdapServiceQuery       string
	cdapSystemServiceQuery string
)

func init() {
	serviceStatus = new(CDAPServiceStatus)
	serviceStatus.init()
	cdapServices = []CDAPService{appFabricSvc, datasetExecutorSvc, logSaverSvc, messagingSvc, metadataSvc, metricsSvc, metricsProcessorSvc, runtimeSvc}
}

type CDAPServiceStatus struct {
	Ready     status.Condition
	Unhealthy status.Condition
	Unknown   status.Condition
}

type SystemServiceStatusResponse struct {
	Status      string      `json:"status"`
	StatusCode  json.Number `json:"statusCode"`
	AppId       string      `json:"appId"`
	ProgramType string      `json:"programType"`
	ProgramId   string      `json:"programId"`
}

func (s *CDAPServiceStatus) init() {
	s.Ready = status.Condition{
		Type:    "CDAPStatusReady",
		Reason:  "Start",
		Message: "CDAP services are ready",
	}
	s.Unhealthy = status.Condition{
		Type:    "CDAPStatusUnhealthy",
		Reason:  "Start",
		Message: "One or more CDAP services are unavailable",
	}
	s.Unknown = status.Condition{
		Type:    "CDAPStatusUnknown",
		Reason:  "Start",
		Message: "CDAP service status is unknown",
	}
}

func setServiceStatusUnknown(master *v1alpha1.CDAPMaster) {
	setCondition(master, serviceStatus.Unknown)
	clearCondition(master, serviceStatus.Ready)
	clearCondition(master, serviceStatus.Unhealthy)
}

func setServiceStatusUnhealthy(master *v1alpha1.CDAPMaster) {
	setCondition(master, serviceStatus.Unhealthy)
	clearCondition(master, serviceStatus.Ready)
	clearCondition(master, serviceStatus.Unknown)
}

func setServiceStatusReady(master *v1alpha1.CDAPMaster) {
	setCondition(master, serviceStatus.Ready)
	clearCondition(master, serviceStatus.Unhealthy)
	clearCondition(master, serviceStatus.Unknown)
}

func setCdapQueryUrls(master *v1alpha1.CDAPMaster) {
	cdapServiceQuery = fmt.Sprintf("http://%s:11015/v3/system/services/status", getObjName(master, serviceRouter))
	cdapSystemServiceQuery = fmt.Sprintf("http://%s:11015/v3/namespaces/system/status", getObjName(master, serviceRouter))
}

func updateServiceStatus(master *v1alpha1.CDAPMaster) {
	setCdapQueryUrls(master)
	updateServiceStatusHelper(master)
}

func updateServiceStatusHelper(master *v1alpha1.CDAPMaster) {
	// Check if CDAP services are up.
	resp, err := http.Get(cdapServiceQuery)
	if err == nil {
		d := json.NewDecoder(resp.Body)
		var svcStatus map[string]string
		err = d.Decode(&svcStatus)
		if err == nil {
			for _, s := range cdapServices {
				if status, ok := svcStatus[s]; !ok || status != "OK" {
					log.Printf("Service %q is not up, /v3/system/services returned: %s", s, svcStatus)
					setServiceStatusUnhealthy(master)
					return
				}
			}
		} else {
			log.Printf("Failed to decode CDAP service status response: %q", err)
			setServiceStatusUnknown(master)
			return
		}
	} else {
		log.Printf("Failed to get CDAP service status: %q", err)
		setServiceStatusUnknown(master)
		return
	}

	// Check if system services are running.
	var jsonStr = []byte(`[{"appId": "dataprep", "programType": "Service", "programId": "service"}, {"appId": "pipeline", "programType": "Service", "programId": "studio"}]
	`)
	resp, err = http.Post(cdapSystemServiceQuery, "application/json", bytes.NewBuffer(jsonStr))
	if err == nil {
		d := json.NewDecoder(resp.Body)
		var serviceStatus []SystemServiceStatusResponse
		err = d.Decode(&serviceStatus)
		if err == nil {
			if len(serviceStatus) != 2 {
				log.Printf("Expected 2 service status entries, got: %q", serviceStatus)
				setServiceStatusUnhealthy(master)
				return
			}
			for i := 0; i < len(serviceStatus); i++ {
				if sc, _ := serviceStatus[i].StatusCode.Int64(); sc != http.StatusOK {
					log.Printf("%s status: %s", serviceStatus[i].AppId, serviceStatus[i].StatusCode)
					setServiceStatusUnhealthy(master)
					return
				}
			}
		} else {
			log.Printf("Failed to decode system service status response: %q", err)
			setServiceStatusUnknown(master)
			return
		}
	} else {
		log.Printf("Failed to get system service status: %q", err)
		setServiceStatusUnknown(master)
		return
	}
	setServiceStatusReady(master)
}
