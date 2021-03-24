package ravendb

type serverMetricsResponse struct {
	ServerVersion     string               `json:"ServerVersion"`
	ServerFullVersion string               `json:"ServerFullVersion"`
	UpTimeInSec       int32                `json:"UpTimeInSec"`
	ServerProcessID   int32                `json:"ServerProcessId"`
	Backup            backupMetrics        `json:"Backup"`
	Config            configurationMetrics `json:"Config"`
	CPU               cpuMetrics           `json:"Cpu"`
	Memory            memoryMetrics        `json:"Memory"`
	Disk              diskMetrics          `json:"Disk"`
	License           licenseMetrics       `json:"License"`
	Network           networkMetrics       `json:"Network"`
	Certificate       certificateMetrics   `json:"Certificate"`
	Cluster           clusterMetrics       `json:"Cluster"`
	Databases         allDatabasesMetrics  `json:"Databases"`
}

type backupMetrics struct {
	CurrentNumberOfRunningBackups int32 `json:"CurrentNumberOfRunningBackups"`
	MaxNumberOfConcurrentBackups  int32 `json:"MaxNumberOfConcurrentBackups"`
}

type configurationMetrics struct {
	ServerUrls          []string `json:"ServerUrls"`
	PublicServerURL     *string  `json:"PublicServerUrl"`
	TCPServerURLs       []string `json:"TcpServerUrls"`
	PublicTCPServerURLs []string `json:"PublicTcpServerUrls"`
}

type cpuMetrics struct {
	ProcessUsage                             float64  `json:"ProcessUsage"`
	MachineUsage                             float64  `json:"MachineUsage"`
	MachineIoWait                            *float64 `json:"MachineIoWait"`
	ProcessorCount                           int32    `json:"ProcessorCount"`
	AssignedProcessorCount                   int32    `json:"AssignedProcessorCount"`
	ThreadPoolAvailableWorkerThreads         int32    `json:"ThreadPoolAvailableWorkerThreads"`
	ThreadPoolAvailableCompletionPortThreads int32    `json:"ThreadPoolAvailableCompletionPortThreads"`
}

type memoryMetrics struct {
	AllocatedMemoryInMb     int64  `json:"AllocatedMemoryInMb"`
	PhysicalMemoryInMb      int64  `json:"PhysicalMemoryInMb"`
	InstalledMemoryInMb     int64  `json:"InstalledMemoryInMb"`
	LowMemorySeverity       string `json:"LowMemorySeverity"`
	TotalSwapSizeInMb       int64  `json:"TotalSwapSizeInMb"`
	TotalSwapUsageInMb      int64  `json:"TotalSwapUsageInMb"`
	WorkingSetSwapUsageInMb int64  `json:"WorkingSetSwapUsageInMb"`
	TotalDirtyInMb          int64  `json:"TotalDirtyInMb"`
}

type diskMetrics struct {
	SystemStoreUsedDataFileSizeInMb  int64 `json:"SystemStoreUsedDataFileSizeInMb"`
	SystemStoreTotalDataFileSizeInMb int64 `json:"SystemStoreTotalDataFileSizeInMb"`
	TotalFreeSpaceInMb               int64 `json:"TotalFreeSpaceInMb"`
	RemainingStorageSpacePercentage  int64 `json:"RemainingStorageSpacePercentage"`
}

type licenseMetrics struct {
	Type                string   `json:"Type"`
	ExpirationLeftInSec *float64 `json:"ExpirationLeftInSec"`
	UtilizedCPUCores    int32    `json:"UtilizedCpuCores"`
	MaxCores            int32    `json:"MaxCores"`
}

type networkMetrics struct {
	TCPActiveConnections                          int64    `json:"TcpActiveConnections"`
	ConcurrentRequestsCount                       int64    `json:"ConcurrentRequestsCount"`
	TotalRequests                                 int64    `json:"TotalRequests"`
	RequestsPerSec                                float64  `json:"RequestsPerSec"`
	LastRequestTimeInSec                          *float64 `json:"LastRequestTimeInSec"`
	LastAuthorizedNonClusterAdminRequestTimeInSec *float64 `json:"LastAuthorizedNonClusterAdminRequestTimeInSec"`
}

type certificateMetrics struct {
	ServerCertificateExpirationLeftInSec *float64 `json:"ServerCertificateExpirationLeftInSec"`
	WellKnownAdminCertificates           []string `json:"WellKnownAdminCertificates"`
}

type clusterMetrics struct {
	NodeTag     string `json:"NodeTag"`
	NodeState   string `json:"NodeState"`
	CurrentTerm int64  `json:"CurrentTerm"`
	Index       int64  `json:"Index"`
	ID          string `json:"Id"`
}

type allDatabasesMetrics struct {
	TotalCount  int32 `json:"TotalCount"`
	LoadedCount int32 `json:"LoadedCount"`
}

type databasesMetricResponse struct {
	Results         []*databaseMetrics `json:"Results"`
	PublicServerURL *string            `json:"PublicServerUrl"`
	NodeTag         string             `json:"NodeTag"`
}

type databaseMetrics struct {
	DatabaseName             string   `json:"DatabaseName"`
	DatabaseID               string   `json:"DatabaseId"`
	UptimeInSec              float64  `json:"UptimeInSec"`
	TimeSinceLastBackupInSec *float64 `json:"TimeSinceLastBackupInSec"`

	Counts     databaseCounts     `json:"Counts"`
	Statistics databaseStatistics `json:"Statistics"`

	Indexes databaseIndexesMetrics `json:"Indexes"`
	Storage databaseStorageMetrics `json:"Storage"`
}

type databaseCounts struct {
	Documents         int64 `json:"Documents"`
	Revisions         int64 `json:"Revisions"`
	Attachments       int64 `json:"Attachments"`
	UniqueAttachments int64 `json:"UniqueAttachments"`
	Alerts            int64 `json:"Alerts"`
	Rehabs            int32 `json:"Rehabs"`
	PerformanceHints  int64 `json:"PerformanceHints"`
	ReplicationFactor int32 `json:"ReplicationFactor"`
}

type databaseStatistics struct {
	DocPutsPerSec               float64 `json:"DocPutsPerSec"`
	MapIndexIndexesPerSec       float64 `json:"MapIndexIndexesPerSec"`
	MapReduceIndexMappedPerSec  float64 `json:"MapReduceIndexMappedPerSec"`
	MapReduceIndexReducedPerSec float64 `json:"MapReduceIndexReducedPerSec"`
	RequestsPerSec              float64 `json:"RequestsPerSec"`
	RequestsCount               int32   `json:"RequestsCount"`
	RequestAverageDurationInMs  float64 `json:"RequestAverageDurationInMs"`
}

type databaseIndexesMetrics struct {
	Count         int64 `json:"Count"`
	StaleCount    int32 `json:"StaleCount"`
	ErrorsCount   int64 `json:"ErrorsCount"`
	StaticCount   int32 `json:"StaticCount"`
	AutoCount     int32 `json:"AutoCount"`
	IdleCount     int32 `json:"IdleCount"`
	DisabledCount int32 `json:"DisabledCount"`
	ErroredCount  int32 `json:"ErroredCount"`
}

type databaseStorageMetrics struct {
	DocumentsAllocatedDataFileInMb int64 `json:"DocumentsAllocatedDataFileInMb"`
	DocumentsUsedDataFileInMb      int64 `json:"DocumentsUsedDataFileInMb"`
	IndexesAllocatedDataFileInMb   int64 `json:"IndexesAllocatedDataFileInMb"`
	IndexesUsedDataFileInMb        int64 `json:"IndexesUsedDataFileInMb"`
	TotalAllocatedStorageFileInMb  int64 `json:"TotalAllocatedStorageFileInMb"`
	TotalFreeSpaceInMb             int64 `json:"TotalFreeSpaceInMb"`
}

type indexesMetricResponse struct {
	Results         []*perDatabaseIndexMetrics `json:"Results"`
	PublicServerURL *string                    `json:"PublicServerUrl"`
	NodeTag         string                     `json:"NodeTag"`
}

type perDatabaseIndexMetrics struct {
	DatabaseName string          `json:"DatabaseName"`
	Indexes      []*indexMetrics `json:"Indexes"`
}

type indexMetrics struct {
	IndexName                  string   `json:"IndexName"`
	Priority                   string   `json:"Priority"`
	State                      string   `json:"State"`
	Errors                     int32    `json:"Errors"`
	TimeSinceLastQueryInSec    *float64 `json:"TimeSinceLastQueryInSec"`
	TimeSinceLastIndexingInSec *float64 `json:"TimeSinceLastIndexingInSec"`
	LockMode                   string   `json:"LockMode"`
	IsInvalid                  bool     `json:"IsInvalid"`
	Status                     string   `json:"Status"`
	MappedPerSec               float64  `json:"MappedPerSec"`
	ReducedPerSec              float64  `json:"ReducedPerSec"`
	Type                       string   `json:"Type"`
	EntriesCount               int32    `json:"EntriesCount"`
}

type collectionsMetricResponse struct {
	Results         []*perDatabaseCollectionMetrics `json:"Results"`
	PublicServerURL *string                         `json:"PublicServerUrl"`
	NodeTag         string                          `json:"NodeTag"`
}

type perDatabaseCollectionMetrics struct {
	DatabaseName string               `json:"DatabaseName"`
	Collections  []*collectionMetrics `json:"Collections"`
}

type collectionMetrics struct {
	CollectionName        string `json:"CollectionName"`
	DocumentsCount        int64  `json:"DocumentsCount"`
	TotalSizeInBytes      int64  `json:"TotalSizeInBytes"`
	DocumentsSizeInBytes  int64  `json:"DocumentsSizeInBytes"`
	TombstonesSizeInBytes int64  `json:"TombstonesSizeInBytes"`
	RevisionsSizeInBytes  int64  `json:"RevisionsSizeInBytes"`
}
