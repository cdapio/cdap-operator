package v1alpha1

// ServiceName is the name identifying various CDAP master services
type ServiceName string

const (
	// ServiceAppFabric defines the service type for app-fabric
	ServiceAppFabric ServiceName = "AppFabric"

	// ServiceLogs defines the service type for log processing and serving service
	ServiceLogs ServiceName = "Logs"

	// ServiceMessaging defines the service type for TMS
	ServiceMessaging ServiceName = "Messaging"

	// ServiceMetadata defines the service type for metadata service
	ServiceMetadata ServiceName = "Metadata"

	// ServiceMetrics defines the service type for metrics process and serving
	ServiceMetrics ServiceName = "Metrics"

	// ServicePreview defines the service type for preview service
	ServicePreview ServiceName = "Preview"

	// ServiceRouter defines the service type for the router
	ServiceRouter ServiceName = "Router"

	// ServiceUserInterface defines the service type for user interface
	ServiceUserInterface ServiceName = "UserInterface"
)

const (
	confExploreEnabled = "explore.enabled"
	// Property key in cdap-site.xml for configuring local data directory
	confLocalDataDirKey       = "local.data.dir"
	confRouterServerAddress   = "router.server.address"
	confRouterBindPort        = "router.bind.port"
	confUserInterfaceBindPort = "dashboard.bind.port"

	// Value for the local data directory
	defaultImage             = "gcr.io/cdapio/cdap:latest"
	defaultRouterPort        = 11015
	defaultUserInterfacePort = 11011
)

// Exported constants
const (
	LocalDataDir  = "/data"
	InstanceLabel = "cdap.instance"
)
