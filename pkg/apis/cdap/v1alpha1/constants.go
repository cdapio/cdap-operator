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
	// AppFabric defines the service type for app-fabric
	AppFabric ServiceType = "AppFabric"

	// Log defines the service type for log processing and serving service
	Log ServiceType = "Log"

	// Messaging defines the service type for TMS
	Messaging ServiceType = "Messaging"

	// Metadata defines the service type for metadata service
	Metadata ServiceType = "Metadata"

	// Metrics defines the service type for metrics process and serving
	Metrics ServiceType = "Metrics"

	// Preview defines the service type for preview service
	Preview ServiceType = "Preview"

	// Router defines the service type for the router
	Router ServiceType = "Router"

	// UserInterface defines the service type for user interface
	UserInterface ServiceType = "UserInterface"
)

const (
	// Property key in cdap-site.xml for configuring local data directory
	localDataDirKey = "local.data.dir"
	// Value for the local data directory
	localDataDir              = "/data"
	instanceLabel             = "cdap.instance"
	containerLabel            = "cdap.container"
	templateLabel             = ".cdap.template"
	templateDir               = "templates/"
	deploymentTemplate        = "cdap-deployment.yaml"
	uiDeploymentTemplate      = "cdap-ui-deployment.yaml"
	statefulSetTemplate       = "cdap-sts.yaml"
	serviceTemplate           = "cdap-service.yaml"
	confRouterServerAddress   = "router.server.address"
	confRouterBindPort        = "router.bind.port"
	confUserInterfaceBindPort = "dashboard.bind.port"
	defaultImage              = "gcr.io/cloud-data-fusion-images/cloud-data-fusion:6.0.0-SNAPSHOT"
	defaultUserInterfaceImage = "gcr.io/cloud-data-fusion-images/cloud-data-fusion-ui:6.0.0-SNAPSHOT"
	defaultRouterPort         = 11015
	defaultUserInterfacePort  = 11011
)
