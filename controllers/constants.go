package controllers

// ServiceName is the name identifying various CDAP services
type ServiceName = string

const (
	// serviceAppFabric defines the service type for app-fabric
	serviceAppFabric ServiceName = "AppFabric"

	// serviceLogs defines the service type for log processing and serving service
	serviceLogs ServiceName = "Logs"

	// serviceMessaging defines the service type for TMS
	serviceMessaging ServiceName = "Messaging"

	// serviceMetadata defines the service type for metadata service
	serviceMetadata ServiceName = "Metadata"

	// serviceMetrics defines the service type for metrics process and serving
	serviceMetrics ServiceName = "Metrics"

	// servicePreview defines the service type for preview service
	servicePreview ServiceName = "Preview"

	// serviceRouter defines the service type for the router
	serviceRouter ServiceName = "Router"

	// serviceUserInterface defines the service type for user interface
	serviceUserInterface ServiceName = "UserInterface"
)

const (
	confExploreEnabled        = "explore.enabled"
	confLocalDataDirKey       = "local.data.dir"
	confLocalDataDirVal       = "/data"
	confRouterServerAddress   = "router.server.address"
	confRouterBindPort        = "router.bind.port"
	confUserInterfaceBindPort = "dashboard.bind.port"

	defaultImage             = "gcr.io/cdapio/cdap:latest"
	defaultRouterPort        = 11015
	defaultUserInterfacePort = 11011

	// kubernetes labels
	labelInstanceKey        = "cdap.instance"
	labelContainerKeyPrefix = "cdap.container."

	// kubernetes object name related
	objectNamePrefix = "cdap-"
	configMapCConf   = "cconf"
	configMapHConf   = "hconf"

	// yaml template
	templateDir         = "templates/"
	templateStatefulSet = "cdap-sts.yaml"
	templateDeployment  = "cdap-deployment.yaml"
	templateService     = "cdap-service.yaml"

	// CDAP services
	containerStorageMain = "io.cdap.cdap.master.environment.k8s.StorageMain"
)

const (
	containerLabel = "cdap.container"
	// Heap memory related constants
	javaMinHeapRatio     = float64(0.6)
	javaReservedNonHeap  = int64(768 * 1024 * 1024)
	uiDeploymentTemplate = "cdap-ui-deployment.yaml"
	upgradeJobTemplate   = "upgrade-job.yaml"
	//UpgradeFailed             = "upgrade-failed"
	//postUpgradeFailed         = "post-upgrade-failed"
	//postUpgradeFinished       = "post-upgrade-finished"
	//upgradeStartMessage       = "Upgrade started, received updated CR."
	//upgradeFailedInitMessage  = "Failed to create job, upgrade failed."
	//upgradeJobFailedMessage   = "Upgrade job failed."
	//upgradeJobFinishedMessage = "Upgrade job finished."
	//upgradeJobSkippedMessage  = "Upgrade job skipped."
	//upgradeResetMessage       = "Upgrade spec reset."
	versionUpgradeFailureLimit = 4
	latestVersion              = "latest"
)
