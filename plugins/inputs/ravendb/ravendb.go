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

const defaultTimeout = 5

// RavenDB defines the configuration necessary for gathering metrics,
// see the sample config for further details
type RavenDB struct {
	URL  string `toml:"url"`
	Name string `toml:"name"`

	Timeout         internal.Duration `toml:"timeout"`

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
	requestUrlServer string
	requestUrlDatabases string
	requestUrlIndexes string
	requestUrlCollection string
}

var sampleConfig = `
  ## Node URL and port that RavenDB is listening on.
  url = "https://localhost:8080"

  ## RavenDB X509 client certificate setup
  # tls_cert = "/etc/telegraf/raven.crt"
  # tls_key = "/etc/telegraf/raven.key"

  ## Optional request timeout
  ##
  ## Timeout, specifies the amount of time to wait
  ## for a server's response headers after fully writing the request and 
  ## time limit for requests made by this client.
  # timeout = "5s"

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

	if r.GatherServerStats {
		wg.Add(1)

		go func () {
			defer wg.Done()
			r.gatherServer(acc)
		}()
	}

	if r.GatherDbStats {
		wg.Add(1)

		go func () {
			defer wg.Done()
			r.gatherDatabases(acc)
		}()
	}

	if r.GatherIndexStats {
		wg.Add(1)

		go func () {
			defer wg.Done()
			r.gatherIndexes(acc)
		}()
	}

	if r.GatherCollectionStats {
		wg.Add(1)

		go func () {
			defer wg.Done()
			r.gatherCollections(acc)
		}()
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
		ResponseHeaderTimeout: r.Timeout.Duration,
		TLSClientConfig:       tlsCfg,
	}
	r.client = &http.Client{
		Transport: tr,
		Timeout:   r.Timeout.Duration,
	}

	return nil
}

func (r *RavenDB) requestJSON(u string, target interface{}) error {
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

func (r *RavenDB) gatherServer(acc telegraf.Accumulator) {
	serverResponse := &serverMetricsResponse{}

	err := r.requestJSON(r.requestUrlServer, &serverResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	tags := map[string]string{
		"cluster_id": serverResponse.Cluster.Id,
		"node_tag":   serverResponse.Cluster.NodeTag,
		"url":        r.URL,
	}

	if serverResponse.Config.PublicServerUrl != nil {
		tags["public_server_url"] = *serverResponse.Config.PublicServerUrl
	}

	fields := map[string]interface{}{
		"backup_current_number_of_running_backups":                      serverResponse.Backup.CurrentNumberOfRunningBackups,
		"backup_max_number_of_concurrent_backups":                       serverResponse.Backup.MaxNumberOfConcurrentBackups,
		"certificate_server_certificate_expiration_left_in_sec":         serverResponse.Certificate.ServerCertificateExpirationLeftInSec,
		"cluster_current_term":                                          serverResponse.Cluster.CurrentTerm,
		"cluster_index":                                                 serverResponse.Cluster.Index,
		"cluster_node_state":                                            serverResponse.Cluster.NodeState,
		"config_server_urls":                                            strings.Join(serverResponse.Config.ServerUrls, ";"),
		"cpu_assigned_processor_count":                                  serverResponse.Cpu.AssignedProcessorCount,
		"cpu_machine_io_wait":                                           serverResponse.Cpu.MachineIoWait,
		"cpu_machine_usage":                                             serverResponse.Cpu.MachineUsage,
		"cpu_process_usage":                                             serverResponse.Cpu.ProcessUsage,
		"cpu_processor_count":                                           serverResponse.Cpu.ProcessorCount,
		"cpu_thread_pool_available_worker_threads":                      serverResponse.Cpu.ThreadPoolAvailableWorkerThreads,
		"cpu_thread_pool_available_completion_port_threads":             serverResponse.Cpu.ThreadPoolAvailableCompletionPortThreads,
		"databases_loaded_count":                                        serverResponse.Databases.LoadedCount,
		"databases_total_count":                                         serverResponse.Databases.TotalCount,
		"disk_remaining_storage_space_percentage":                       serverResponse.Disk.RemainingStorageSpacePercentage,
		"disk_system_store_used_data_file_size_in_mb":                   serverResponse.Disk.SystemStoreUsedDataFileSizeInMb,
		"disk_system_store_total_data_file_size_in_mb":                  serverResponse.Disk.SystemStoreTotalDataFileSizeInMb,
		"disk_total_free_space_in_mb":                                   serverResponse.Disk.TotalFreeSpaceInMb,
		"license_expiration_left_in_sec":                                serverResponse.License.ExpirationLeftInSec,
		"license_max_cores":                                             serverResponse.License.MaxCores,
		"license_type":                                                  serverResponse.License.Type,
		"license_utilized_cpu_cores":                                    serverResponse.License.UtilizedCpuCores,
		"memory_allocated_in_mb":                                        serverResponse.Memory.AllocatedMemoryInMb,
		"memory_installed_in_mb":                                        serverResponse.Memory.InstalledMemoryInMb,
		"memory_low_memory_severity":                                    serverResponse.Memory.LowMemorySeverity,
		"memory_physical_in_mb":                                         serverResponse.Memory.PhysicalMemoryInMb,
		"memory_total_dirty_in_mb":                                      serverResponse.Memory.TotalDirtyInMb,
		"memory_total_swap_size_in_mb":                                  serverResponse.Memory.TotalSwapSizeInMb,
		"memory_total_swap_usage_in_mb":                                 serverResponse.Memory.TotalSwapUsageInMb,
		"memory_working_set_swap_usage_in_mb":                           serverResponse.Memory.WorkingSetSwapUsageInMb,
		"network_concurrent_requests_count":                             serverResponse.Network.ConcurrentRequestsCount,
		"network_last_authorized_non_cluster_admin_request_time_in_sec": serverResponse.Network.LastAuthorizedNonClusterAdminRequestTimeInSec,
		"network_last_request_time_in_sec":                              serverResponse.Network.LastRequestTimeInSec,
		"network_requests_per_sec":                                      serverResponse.Network.RequestsPerSec,
		"network_tcp_active_connections":                                serverResponse.Network.TcpActiveConnections,
		"network_total_requests":                                        serverResponse.Network.TotalRequests,
		"server_full_version":                                           serverResponse.ServerFullVersion,
		"server_process_id":                                             serverResponse.ServerProcessId,
		"server_version":                                                serverResponse.ServerVersion,
		"uptime_in_sec":                                                 serverResponse.UpTimeInSec,
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

func (r *RavenDB) gatherDatabases(acc telegraf.Accumulator) {
	databasesResponse := &databasesMetricResponse{}

	err := r.requestJSON(r.requestUrlDatabases, &databasesResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, dbResponse := range databasesResponse.Results {
		tags := map[string]string{
			"database_id":   dbResponse.DatabaseId,
			"database_name": dbResponse.DatabaseName,
			"node_tag":      databasesResponse.NodeTag,
			"url":           r.URL,
		}

		if databasesResponse.PublicServerUrl != nil {
			tags["public_server_url"] = *databasesResponse.PublicServerUrl
		}

		fields := map[string]interface{}{
			"counts_alerts":                               dbResponse.Counts.Alerts,
			"counts_attachments":                          dbResponse.Counts.Attachments,
			"counts_documents":                            dbResponse.Counts.Documents,
			"counts_performance_hints":                    dbResponse.Counts.PerformanceHints,
			"counts_rehabs":                               dbResponse.Counts.Rehabs,
			"counts_replication_factor":                   dbResponse.Counts.ReplicationFactor,
			"counts_revisions":                            dbResponse.Counts.Revisions,
			"counts_unique_attachments":                   dbResponse.Counts.UniqueAttachments,
			"indexes_auto_count":                          dbResponse.Indexes.AutoCount,
			"indexes_count":                               dbResponse.Indexes.Count,
			"indexes_errored_count":                       dbResponse.Indexes.ErroredCount,
			"indexes_errors_count":                        dbResponse.Indexes.ErrorsCount,
			"indexes_disabled_count":                      dbResponse.Indexes.DisabledCount,
			"indexes_idle_count":                          dbResponse.Indexes.IdleCount,
			"indexes_stale_count":                         dbResponse.Indexes.StaleCount,
			"indexes_static_count":                        dbResponse.Indexes.StaticCount,
			"statistics_doc_puts_per_sec":                 dbResponse.Statistics.DocPutsPerSec,
			"statistics_map_index_indexes_per_sec":        dbResponse.Statistics.MapIndexIndexesPerSec,
			"statistics_map_reduce_index_mapped_per_sec":  dbResponse.Statistics.MapReduceIndexMappedPerSec,
			"statistics_map_reduce_index_reduced_per_sec": dbResponse.Statistics.MapReduceIndexReducedPerSec,
			"statistics_request_average_duration_in_ms":   dbResponse.Statistics.RequestAverageDurationInMs,
			"statistics_requests_count":                   dbResponse.Statistics.RequestsCount,
			"statistics_requests_per_sec":                 dbResponse.Statistics.RequestsPerSec,
			"storage_documents_allocated_data_file_in_mb": dbResponse.Storage.DocumentsAllocatedDataFileInMb,
			"storage_documents_used_data_file_in_mb":      dbResponse.Storage.DocumentsUsedDataFileInMb,
			"storage_indexes_allocated_data_file_in_mb":   dbResponse.Storage.IndexesAllocatedDataFileInMb,
			"storage_indexes_used_data_file_in_mb":        dbResponse.Storage.IndexesUsedDataFileInMb,
			"storage_total_allocated_storage_file_in_mb":  dbResponse.Storage.TotalAllocatedStorageFileInMb,
			"storage_total_free_space_in_mb":              dbResponse.Storage.TotalFreeSpaceInMb,
			"time_since_last_backup_in_sec":               dbResponse.TimeSinceLastBackupInSec,
			"uptime_in_sec":                               dbResponse.UptimeInSec,
		}

		acc.AddFields("ravendb_databases", fields, tags)
	}
}

func (r *RavenDB) gatherIndexes(acc telegraf.Accumulator) {
	indexesResponse := &indexesMetricResponse{}

	err := r.requestJSON(r.requestUrlIndexes, &indexesResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, perDbIndexResponse := range indexesResponse.Results {
		for _, indexResponse := range perDbIndexResponse.Indexes {
			tags := map[string]string{
				"database_name": perDbIndexResponse.DatabaseName,
				"index_name":    indexResponse.IndexName,
				"node_tag":      indexesResponse.NodeTag,
				"url":           r.URL,
			}

			if indexesResponse.PublicServerUrl != nil {
				tags["public_server_url"] = *indexesResponse.PublicServerUrl
			}

			fields := map[string]interface{}{
				"errors":                          indexResponse.Errors,
				"is_invalid":                      indexResponse.IsInvalid,
				"lock_mode":                       indexResponse.LockMode,
				"mapped_per_sec":                  indexResponse.MappedPerSec,
				"priority":                        indexResponse.Priority,
				"reduced_per_sec":                 indexResponse.ReducedPerSec,
				"state":                           indexResponse.State,
				"status":                          indexResponse.Status,
				"time_since_last_indexing_in_sec": indexResponse.TimeSinceLastIndexingInSec,
				"time_since_last_query_in_sec":    indexResponse.TimeSinceLastQueryInSec,
				"type":                            indexResponse.Type,
			}

			acc.AddFields("ravendb_indexes", fields, tags)
		}
	}
}

func (r *RavenDB) gatherCollections(acc telegraf.Accumulator) {
	collectionsResponse := &collectionsMetricResponse{}

	err := r.requestJSON(r.requestUrlCollection, &collectionsResponse)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, perDbCollectionMetrics := range collectionsResponse.Results {
		for _, collectionMetrics := range perDbCollectionMetrics.Collections {
			tags := map[string]string{
				"collection_name": collectionMetrics.CollectionName,
				"database_name":   perDbCollectionMetrics.DatabaseName,
				"node_tag":        collectionsResponse.NodeTag,
				"url":             r.URL,
			}

			if collectionsResponse.PublicServerUrl != nil {
				tags["public_server_url"] = *collectionsResponse.PublicServerUrl
			}

			fields := map[string]interface{}{
				"documents_count":          collectionMetrics.DocumentsCount,
				"documents_size_in_bytes":  collectionMetrics.DocumentsSizeInBytes,
				"revisions_size_in_bytes":  collectionMetrics.RevisionsSizeInBytes,
				"tombstones_size_in_bytes": collectionMetrics.TombstonesSizeInBytes,
				"total_size_in_bytes":      collectionMetrics.TotalSizeInBytes,
			}

			acc.AddFields("ravendb_collections", fields, tags)
		}
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

func (r *RavenDB) Init() error {
	if r.URL == "" {
		r.URL = defaultURL
	}

	r.requestUrlServer = fmt.Sprintf("%s%s", r.URL, "/admin/monitoring/v1/server")
	r.requestUrlDatabases = fmt.Sprintf("%s%s", r.URL, "/admin/monitoring/v1/databases"+prepareDbNamesUrlPart(r.DbStatsDbs))
	r.requestUrlIndexes = fmt.Sprintf("%s%s", r.URL, "/admin/monitoring/v1/indexes"+prepareDbNamesUrlPart(r.IndexStatsDbs))
	r.requestUrlCollection = fmt.Sprintf("%s%s", r.URL, "/admin/monitoring/v1/collections"+prepareDbNamesUrlPart(r.IndexStatsDbs))

	return nil
}

func init() {
	inputs.Add("ravendb", func() telegraf.Input {
		return &RavenDB{
			Timeout:         internal.Duration{Duration: defaultTimeout * time.Second},
			GatherServerStats:     true,
			GatherDbStats:         true,
			GatherIndexStats:      true,
			GatherCollectionStats: true,
		}
	})
}
