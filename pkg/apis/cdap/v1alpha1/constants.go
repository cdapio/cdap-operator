/*
Copyright 2019 The CDAP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

// ServiceType is the name identifying various CDAP master services
type ServiceType string

const (
	// ServiceAppFabric defines the service type for app-fabric
	ServiceAppFabric ServiceType = "AppFabric"

	// ServiceLogs defines the service type for log processing and serving service
	ServiceLogs ServiceType = "Logs"

	// ServiceMessaging defines the service type for TMS
	ServiceMessaging ServiceType = "Messaging"

	// ServiceMetadata defines the service type for metadata service
	ServiceMetadata ServiceType = "Metadata"

	// ServiceMetrics defines the service type for metrics process and serving
	ServiceMetrics ServiceType = "Metrics"

	// ServicePreview defines the service type for preview service
	ServicePreview ServiceType = "Preview"

	// ServiceRouter defines the service type for the router
	ServiceRouter ServiceType = "Router"

	// ServiceUserInterface defines the service type for user interface
	ServiceUserInterface ServiceType = "UserInterface"
)

const (
	confExploreEnabled = "explore.enabled"
	// Property key in cdap-site.xml for configuring local data directory
	confLocalDataDirKey       = "local.data.dir"
	confRouterServerAddress   = "router.server.address"
	confRouterBindPort        = "router.bind.port"
	confUserInterfaceBindPort = "dashboard.bind.port"

	// Value for the local data directory
	defaultImage              = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.0.0-SNAPSHOT"
	defaultUserInterfaceImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion-ui:6.0.0-SNAPSHOT"
	defaultRouterPort         = 11015
	defaultUserInterfacePort  = 11011
)

// Exported constants
const (
	LocalDataDir  = "/data"
	InstanceLabel = "cdap.instance"
)
