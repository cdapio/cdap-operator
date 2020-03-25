package controllers

// ServiceName is the name identifying various CDAP services
type ServiceName = string

// ServiceNames must match the field name defined in CDAPMasterSpec as they are used by reflection to
// find the field value.
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

	// serviceRuntime defines the service type for runtime service.
	// This is an optional serivce which may or may not be deployed depending on the setting of customer resource.
	serviceRuntime ServiceName = "Runtime"

	// serviceRouter defines the service type for the router
	serviceRouter ServiceName = "Router"

	// serviceUserInterface defines the service type for user interface
	serviceUserInterface ServiceName = "UserInterface"
)

const (
	// CRD field name (used by reflection)
	fieldNameCDAPServiceSpec         = "CDAPServiceSpec"
	fieldNameCDAPStatefulServiceSpec = "CDAPStatefulServiceSpec"
	fieldNameCDAPExternalServiceSpec = "CDAPExternalServiceSpec"

	// cconf and hconf
	confExploreEnabled        = "explore.enabled"
	confLocalDataDirKey       = "local.data.dir"
	confLocalDataDirVal       = "/data"
	confRouterServerAddress   = "router.server.address"
	confRouterBindPort        = "router.bind.port"
	confUserInterfaceBindPort = "dashboard.bind.port"

	// default values
	defaultImage             = "gcr.io/cdapio/cdap:latest"
	defaultRouterPort        = 11015
	defaultUserInterfacePort = 11011
	defaultStorageSize       = "200Gi"

	// kubernetes labels
	labelInstanceKey        = "cdap.instance"
	labelContainerKeyPrefix = "cdap.container."

	// kubernetes object name related
	objectNamePrefix    = "cdap-"
	configMapCConf      = "cconf"
	configMapHConf      = "hconf"
	configMapSysAppConf = "sysappconf"

	// yaml template
	templateDir         = "templates/"
	templateStatefulSet = "cdap-sts.yaml"
	templateDeployment  = "cdap-deployment.yaml"
	templateService     = "cdap-service.yaml"
	templateUpgradeJob  = "upgrade-job.yaml"

	// Image version upgrade/downgrade
	imageVersionLatest = "latest"
	// Have a high number of retry count to increase the chance of pre-/post- upgrade job succeeding,
	// because for instance it may take a while for CDAP services to be fully up after pods being restarted
	// following image version update, thus post-upgrade job may have to be retried a number of times before
	// it can actually communicate with CDAP services.
	imageVersionUpgradeJobMaxRetryCount = 10

	// CDAP services
	containerStorageMain = "io.cdap.cdap.master.environment.k8s.StorageMain"

	// Java heap size
	javaMinHeapRatio          = float64(0.6)
	javaReservedNonHeap       = int64(768 * 1024 * 1024)
	javaMaxHeapSizeEnvVarName = "JAVA_HEAPMAX"

	Bytes     = int64(1)
	kiloBytes = int64(1024)
	megaBytes = int64(1024 * 1024)
	gigaBytes = int64(1024 * 1024 * 1024)
)
