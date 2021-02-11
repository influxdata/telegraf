package ravendb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// defaultURL will set a default value that corresponds to the default value
// used by RavenDB
const defaultURL = "http://localhost:8080"

// Default http timeouts
const defaultResponseHeaderTimeout = 3
const defaultClientTimeout = 4

type gatherFunc func(r *RavenDB, acc telegraf.Accumulator)

var gatherFunctions = []gatherFunc{gatherServer, gatherDatabases, gatherIndexes, gatherCollections}

// RavenDB defines the configuration necessary for gathering metrics,
// see the sample config for further details
type RavenDB struct {
	URL  string `toml:"url"`
	Name string `toml:"name"`

	ResponseHeaderTimeout internal.Duration `toml:"header_timeout"`
	ClientTimeout         internal.Duration `toml:"client_timeout"`

	GatherServerStats     bool     `toml:"gather_server_stats"`
	GatherDbStats         bool     `toml:"gather_db_stats"`
	GatherIndexStats      bool     `toml:"gather_index_stats"`
	GatherCollectionStats bool     `toml:"gather_collection_stats"`
	DbStatsDbs            []string `toml:"db_stats_dbs"`
	IndexStatsDbs         []string `toml:"index_stats_dbs"`
	CollectionStatsDbs    []string `toml:"collection_stats_dbs"`

	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	client *http.Client
}

type serverMetricsResponse struct {
	ServerVersion     string               `json:"ServerVersion"`
	ServerFullVersion string               `json:"ServerFullVersion"`
	UpTimeInSec       int32                `json:"UpTimeInSec"`
	ServerProcessId   int32                `json:"ServerProcessId"`
	Backup            backupMetrics        `json:"Backup"`
	Config            configurationMetrics `json:"Config"`
	Cpu               cpuMetrics           `json:"Cpu"`
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
	PublicServerUrl     *string  `json:"PublicServerUrl"`
	TcpServerUrls       []string `json:"TcpServerUrls"`
	PublicTcpServerUrls []string `json:"PublicTcpServerUrls"`
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
	UtilizedCpuCores    int32    `json:"UtilizedCpuCores"`
	MaxCores            int32    `json:"MaxCores"`
}

type networkMetrics struct {
	TcpActiveConnections                          int64    `json:"TcpActiveConnections"`
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
	Id          string `json:"Id"`
}

type allDatabasesMetrics struct {
	TotalCount  int32 `json:"TotalCount"`
	LoadedCount int32 `json:"LoadedCount"`
}

type databasesMetricResponse struct {
	Results         []*databaseMetrics `json:"Results"`
	PublicServerUrl *string            `json:"PublicServerUrl"`
	NodeTag         string             `json:"NodeTag"`
}

type databaseMetrics struct {
	DatabaseName             string   `json:"DatabaseName"`
	DatabaseId               string   `json:"DatabaseId"`
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
	Results         []*indexMetrics `json:"Results"`
	PublicServerUrl *string         `json:"PublicServerUrl"`
	NodeTag         string          `json:"NodeTag"`
}

type indexMetrics struct {
	DatabaseName               string   `json:"DatabaseName"`
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
	Results         []*collectionMetrics `json:"Results"`
	PublicServerUrl *string              `json:"PublicServerUrl"`
	NodeTag         string               `json:"NodeTag"`
}

type collectionMetrics struct {
	DatabaseName          string `json:"DatabaseName"`
	CollectionName        string `json:"CollectionName"`
	DocumentsCount        int64  `json:"DocumentsCount"`
	TotalSizeInBytes      int64  `json:"TotalSizeInBytes"`
	DocumentsSizeInBytes  int64  `json:"DocumentsSizeInBytes"`
	TombstonesSizeInBytes int64  `json:"TombstonesSizeInBytes"`
	RevisionsSizeInBytes  int64  `json:"RevisionsSizeInBytes"`
}

var sampleConfig = `
  ## Node URL and port that RavenDB is listening on.
  url = "https://localhost:8080"

  ## RavenDB X509 client certificate setup
  # tls_cert = "/etc/telegraf/raven.crt"
  # tls_key = "/etc/telegraf/raven.key"

  ## Optional request timeouts
  ##
  ## ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait
  ## for a server's response headers after fully writing the request.
  # header_timeout = "3s"
  ##
  ## client_timeout specifies a time limit for requests made by this client.
  ## Includes connection time, any redirects, and reading the response body.
  # client_timeout = "4s"

  ## When true, collect server stats
  # gather_server_stats = true

  ## When true, collect per database stats
  # gather_db_stats = true

  ## When true, collect per index stats
  # gather_index_stats = true

  ## When true, collect per collection stats
  # gather_collection_stats = true

  ## List of db where database stats are collected
  ## If empty, all db are concerned
  # db_stats_dbs = []

  ## List of db where index status are collected
  ## If empty, all indexes from all db are concerned
  # index_stats_dbs = []

  ## List of db where collection status are collected
  ## If empty, all collections from all db are concerned
  # collection_stats_dbs = []
`

// SampleConfig ...
func (r *RavenDB) SampleConfig() string {
	return sampleConfig
}

// Description ...
func (r *RavenDB) Description() string {
	return "Reads metrics from RavenDB servers via the Monitoring Endpoints"
}

// Gather ...
func (r *RavenDB) Gather(acc telegraf.Accumulator) error {
	err := r.ensureClient()
	if nil != err {
		r.Log.Errorf("Error with Client %s", err)
		return err
	}

	var wg sync.WaitGroup
	for _, f := range gatherFunctions {
		wg.Add(1)
		go func(gf gatherFunc) {
			defer wg.Done()
			gf(r, acc)
		}(f)
	}
	wg.Wait()

	return nil
}

func (r *RavenDB) ensureClient() error {
	if r.client != nil {
		return nil
	}

	tlsCfg, err := r.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: r.ResponseHeaderTimeout.Duration,
		TLSClientConfig:       tlsCfg,
	}
	r.client = &http.Client{
		Transport: tr,
		Timeout:   r.ClientTimeout.Duration,
	}

	return nil
}

func (r *RavenDB) requestJSON(u string, target interface{}) error {
	if r.URL == "" {
		r.URL = defaultURL
	}
	u = fmt.Sprintf("%s%s", r.URL, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	r.Log.Debugf("%s: %s", u, resp.Status)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("invalid response code to request '%s': %d - %s", r.URL, resp.StatusCode, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func gatherServer(r *RavenDB, acc telegraf.Accumulator) {
	if !r.GatherServerStats {
		return
	}

	serverResponse := &serverMetricsResponse{}

	err := r.requestJSON("/admin/monitoring/v1/server", &serverResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"url":        r.URL,
		"node_tag":   serverResponse.Cluster.NodeTag,
		"cluster_id": serverResponse.Cluster.Id,
	}

	if serverResponse.Config.PublicServerUrl != nil {
		tags["public_server_url"] = *serverResponse.Config.PublicServerUrl
	}

	fields := map[string]interface{}{
		"server_version":                                                serverResponse.ServerVersion,
		"server_full_version":                                           serverResponse.ServerFullVersion,
		"uptime_in_sec":                                                 serverResponse.UpTimeInSec,
		"server_process_id":                                             serverResponse.ServerProcessId,
		"config_server_urls":                                            strings.Join(serverResponse.Config.ServerUrls, ";"),
		"backup_max_number_of_concurrent_backups":                       serverResponse.Backup.MaxNumberOfConcurrentBackups,
		"backup_current_number_of_running_backups":                      serverResponse.Backup.CurrentNumberOfRunningBackups,
		"cpu_process_usage":                                             serverResponse.Cpu.ProcessUsage,
		"cpu_machine_usage":                                             serverResponse.Cpu.MachineUsage,
		"cpu_processor_count":                                           serverResponse.Cpu.ProcessorCount,
		"cpu_assigned_processor_count":                                  serverResponse.Cpu.AssignedProcessorCount,
		"cpu_thread_pool_available_worker_threads":                      serverResponse.Cpu.ThreadPoolAvailableWorkerThreads,
		"cpu_thread_pool_available_completion_port_threads":             serverResponse.Cpu.ThreadPoolAvailableCompletionPortThreads,
		"memory_allocated_in_mb":                                        serverResponse.Memory.AllocatedMemoryInMb,
		"memory_installed_in_mb":                                        serverResponse.Memory.InstalledMemoryInMb,
		"memory_physical_in_mb":                                         serverResponse.Memory.PhysicalMemoryInMb,
		"memory_low_memory_severity":                                    mapMemorySeverity(serverResponse.Memory.LowMemorySeverity),
		"memory_total_swap_size_in_mb":                                  serverResponse.Memory.TotalSwapSizeInMb,
		"memory_total_swap_usage_in_mb":                                 serverResponse.Memory.TotalSwapUsageInMb,
		"memory_working_set_swap_usage_in_mb":                           serverResponse.Memory.WorkingSetSwapUsageInMb,
		"memory_total_dirty_in_mb":                                      serverResponse.Memory.TotalDirtyInMb,
		"disk_system_store_used_data_file_size_in_mb":                   serverResponse.Disk.SystemStoreUsedDataFileSizeInMb,
		"disk_system_store_total_data_file_size_in_mb":                  serverResponse.Disk.SystemStoreTotalDataFileSizeInMb,
		"disk_total_free_space_in_mb":                                   serverResponse.Disk.TotalFreeSpaceInMb,
		"disk_remaining_storage_space_percentage":                       serverResponse.Disk.RemainingStorageSpacePercentage,
		"license_type":                                                  serverResponse.License.Type,
		"license_utilized_cpu_cores":                                    serverResponse.License.UtilizedCpuCores,
		"license_max_cores":                                             serverResponse.License.MaxCores,
		"network_tcp_active_connections":                                serverResponse.Network.TcpActiveConnections,
		"network_concurrent_requests_count":                             serverResponse.Network.ConcurrentRequestsCount,
		"network_total_requests":                                        serverResponse.Network.TotalRequests,
		"network_requests_per_sec":                                      serverResponse.Network.RequestsPerSec,
		"cluster_node_state":                                            mapNodeState(serverResponse.Cluster.NodeState),
		"cluster_current_term":                                          serverResponse.Cluster.CurrentTerm,
		"cluster_index":                                                 serverResponse.Cluster.Index,
		"databases_total_count":                                         serverResponse.Databases.TotalCount,
		"databases_loaded_count":                                        serverResponse.Databases.LoadedCount,
		"cpu_machine_io_wait":                                           serverResponse.Cpu.MachineIoWait,
		"license_expiration_left_in_sec":                                serverResponse.License.ExpirationLeftInSec,
		"network_last_request_time_in_sec":                              serverResponse.Network.LastRequestTimeInSec,
		"network_last_authorized_non_cluster_admin_request_time_in_sec": serverResponse.Network.LastAuthorizedNonClusterAdminRequestTimeInSec,
		"certificate_server_certificate_expiration_left_in_sec":         serverResponse.Certificate.ServerCertificateExpirationLeftInSec,
	}

	if serverResponse.Config.TcpServerUrls != nil {
		fields["config_tcp_server_urls"] = strings.Join(serverResponse.Config.TcpServerUrls, ";")
	}

	if serverResponse.Config.PublicTcpServerUrls != nil {
		fields["config_public_tcp_server_urls"] = strings.Join(serverResponse.Config.PublicTcpServerUrls, ";")
	}

	if serverResponse.Certificate.WellKnownAdminCertificates != nil {
		fields["certificate_well_known_admin_certificates"] = strings.Join(serverResponse.Certificate.WellKnownAdminCertificates, ";")
	}

	acc.AddFields("ravendb_server", fields, tags)
}

func mapMemorySeverity(severity string) int8 {
	switch severity {
	case "None":
		return 0
	case "Low":
		return 1
	case "ExtremelyLow":
		return 2
	default:
		return -1
	}
}

func mapNodeState(state string) int8 {
	switch state {
	case "Passive":
		return 0
	case "Candidate":
		return 1
	case "Follower":
		return 2
	case "LeaderElect":
		return 3
	case "Leader":
		return 4
	default:
		return -1
	}
}

func gatherDatabases(r *RavenDB, acc telegraf.Accumulator) {
	if !r.GatherDbStats {
		return
	}

	databasesResponse := &databasesMetricResponse{}

	err := r.requestJSON("/admin/monitoring/v1/databases"+prepareDbNamesUrlPart(r.DbStatsDbs), &databasesResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, dbResponse := range databasesResponse.Results {
		tags := map[string]string{
			"url":           r.URL,
			"database_name": dbResponse.DatabaseName,
			"database_id":   dbResponse.DatabaseId,
			"node_tag":      databasesResponse.NodeTag,
		}

		if databasesResponse.PublicServerUrl != nil {
			tags["public_server_url"] = *databasesResponse.PublicServerUrl
		}

		fields := map[string]interface{}{
			"uptime_in_sec":                               dbResponse.UptimeInSec,
			"counts_documents":                            dbResponse.Counts.Documents,
			"counts_revisions":                            dbResponse.Counts.Revisions,
			"counts_attachments":                          dbResponse.Counts.Attachments,
			"counts_unique_attachments":                   dbResponse.Counts.UniqueAttachments,
			"counts_alerts":                               dbResponse.Counts.Alerts,
			"counts_rehabs":                               dbResponse.Counts.Rehabs,
			"counts_performance_hints":                    dbResponse.Counts.PerformanceHints,
			"counts_replication_factor":                   dbResponse.Counts.ReplicationFactor,
			"statistics_doc_puts_per_sec":                 dbResponse.Statistics.DocPutsPerSec,
			"statistics_map_index_indexes_per_sec":        dbResponse.Statistics.MapIndexIndexesPerSec,
			"statistics_map_reduce_index_mapped_per_sec":  dbResponse.Statistics.MapReduceIndexMappedPerSec,
			"statistics_map_reduce_index_reduced_per_sec": dbResponse.Statistics.MapReduceIndexReducedPerSec,
			"statistics_requests_per_sec":                 dbResponse.Statistics.RequestsPerSec,
			"statistics_requests_count":                   dbResponse.Statistics.RequestsCount,
			"statistics_request_average_duration_in_ms":   dbResponse.Statistics.RequestAverageDurationInMs,
			"indexes_count":                               dbResponse.Indexes.Count,
			"indexes_stale_count":                         dbResponse.Indexes.StaleCount,
			"indexes_errors_count":                        dbResponse.Indexes.ErrorsCount,
			"indexes_static_count":                        dbResponse.Indexes.StaticCount,
			"indexes_auto_count":                          dbResponse.Indexes.AutoCount,
			"indexes_idle_count":                          dbResponse.Indexes.IdleCount,
			"indexes_disabled_count":                      dbResponse.Indexes.DisabledCount,
			"indexes_errored_count":                       dbResponse.Indexes.ErroredCount,
			"storage_documents_allocated_data_file_in_mb": dbResponse.Storage.DocumentsAllocatedDataFileInMb,
			"storage_documents_used_data_file_in_mb":      dbResponse.Storage.DocumentsUsedDataFileInMb,
			"storage_indexes_allocated_data_file_in_mb":   dbResponse.Storage.IndexesAllocatedDataFileInMb,
			"storage_indexes_used_data_file_in_mb":        dbResponse.Storage.IndexesUsedDataFileInMb,
			"storage_total_allocated_storage_file_in_mb":  dbResponse.Storage.TotalAllocatedStorageFileInMb,
			"storage_total_free_space_in_mb":              dbResponse.Storage.TotalFreeSpaceInMb,
			"time_since_last_backup_in_sec":               dbResponse.TimeSinceLastBackupInSec,
		}

		acc.AddFields("ravendb_databases", fields, tags)
	}
}

func gatherIndexes(r *RavenDB, acc telegraf.Accumulator) {
	if !r.GatherIndexStats {
		return
	}

	indexesResponse := &indexesMetricResponse{}

	err := r.requestJSON("/admin/monitoring/v1/indexes"+prepareDbNamesUrlPart(r.IndexStatsDbs), &indexesResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, indexResponse := range indexesResponse.Results {
		tags := map[string]string{
			"url":           r.URL,
			"database_name": indexResponse.DatabaseName,
			"index_name":    indexResponse.IndexName,
			"node_tag":      indexesResponse.NodeTag,
		}

		if indexesResponse.PublicServerUrl != nil {
			tags["public_server_url"] = *indexesResponse.PublicServerUrl
		}

		fields := map[string]interface{}{
			"priority":                        indexResponse.Priority,
			"state":                           indexResponse.State,
			"errors":                          indexResponse.Errors,
			"lock_mode":                       indexResponse.LockMode,
			"is_invalid":                      indexResponse.IsInvalid,
			"status":                          indexResponse.Status,
			"mapped_per_sec":                  indexResponse.MappedPerSec,
			"reduced_per_sec":                 indexResponse.ReducedPerSec,
			"type":                            indexResponse.Type,
			"time_since_last_query_in_sec":    indexResponse.TimeSinceLastQueryInSec,
			"time_since_last_indexing_in_sec": indexResponse.TimeSinceLastIndexingInSec,
		}

		acc.AddFields("ravendb_indexes", fields, tags)
	}
}

func gatherCollections(r *RavenDB, acc telegraf.Accumulator) {
	if !r.GatherCollectionStats {
		return
	}

	collectionsResponse := &collectionsMetricResponse{}

	err := r.requestJSON("/admin/monitoring/v1/collections"+prepareDbNamesUrlPart(r.IndexStatsDbs), &collectionsResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, collectionMetrics := range collectionsResponse.Results {
		tags := map[string]string{
			"url":             r.URL,
			"node_tag":        collectionsResponse.NodeTag,
			"database_name":   collectionMetrics.DatabaseName,
			"collection_name": collectionMetrics.CollectionName,
		}

		if collectionsResponse.PublicServerUrl != nil {
			tags["public_server_url"] = *collectionsResponse.PublicServerUrl
		}

		fields := map[string]interface{}{
			"documents_count":          collectionMetrics.DocumentsCount,
			"total_size_in_bytes":      collectionMetrics.TotalSizeInBytes,
			"documents_size_in_bytes":  collectionMetrics.DocumentsSizeInBytes,
			"tombstones_size_in_bytes": collectionMetrics.TombstonesSizeInBytes,
			"revisions_size_in_bytes":  collectionMetrics.RevisionsSizeInBytes,
		}

		acc.AddFields("ravendb_collections", fields, tags)
	}
}

func prepareDbNamesUrlPart(dbNames []string) string {
	if len(dbNames) == 0 {
		return ""
	}
	result := "?" + dbNames[0]
	for _, db := range dbNames[1:] {
		result += "&name=" + url.QueryEscape(db)
	}

	return result
}

func init() {
	inputs.Add("ravendb", func() telegraf.Input {
		return &RavenDB{
			ResponseHeaderTimeout: internal.Duration{Duration: defaultResponseHeaderTimeout * time.Second},
			ClientTimeout:         internal.Duration{Duration: defaultClientTimeout * time.Second},
			GatherServerStats:     true,
			GatherDbStats:         true,
			GatherIndexStats:      true,
			GatherCollectionStats: true,
		}
	})
}
