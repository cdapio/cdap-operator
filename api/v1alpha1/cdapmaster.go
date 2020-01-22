package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/controller-reconciler/pkg/finalizer"
)

// Creates a int32 pointer for the given value
func int32Ptr(value int32) *int32 {
	return &value
}

// ApplyDefaults will default missing values from the CdapMaster
func (r *CdapMaster) ApplyDefaults() {
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[InstanceLabel] = r.Name

	spec := &r.Spec
	if spec.Image == "" {
		spec.Image = defaultImage
	}
	if spec.UserInterfaceImage == "" {
		spec.UserInterfaceImage = spec.Image
	}

	if spec.Router.ServicePort == nil {
		spec.Router.ServicePort = int32Ptr(defaultRouterPort)
	}
	if spec.UserInterface.ServicePort == nil {
		spec.UserInterface.ServicePort = int32Ptr(defaultUserInterfacePort)
	}

	if spec.Config == nil {
		spec.Config = make(map[string]string)
	}
	// Set the local data directory
	spec.Config[confLocalDataDirKey] = LocalDataDir

	// Set the cconf entry for the router and UI service and ports
	spec.Config[confRouterServerAddress] = fmt.Sprintf("cdap-%s-%s", r.Name, strings.ToLower(string(ServiceRouter)))
	spec.Config[confRouterBindPort] = strconv.Itoa(int(*spec.Router.ServicePort))
	spec.Config[confUserInterfaceBindPort] = strconv.Itoa(int(*spec.UserInterface.ServicePort))

	// Disable explore
	spec.Config[confExploreEnabled] = "false"

	r.Status.ResetComponentList()
	r.Status.EnsureStandardConditions()
	finalizer.EnsureStandard(r)
}
