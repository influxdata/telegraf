package ravendb

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// Test against fully filled data
func TestRavenDBGeneratesMetricsFull(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonFilePath string

		switch r.URL.Path {
		case "/admin/monitoring/v1/databases":
			jsonFilePath = "testdata/databases_full.json"
		case "/admin/monitoring/v1/server":
			jsonFilePath = "testdata/server_full.json"
		case "/admin/monitoring/v1/indexes":
			jsonFilePath = "testdata/indexes_full.json"
		case "/admin/monitoring/v1/collections":
			jsonFilePath = "testdata/collections_full.json"

		default:
			require.Failf(t, "Cannot handle request for uri %s", r.URL.Path)
		}

		data, err := os.ReadFile(jsonFilePath)
		require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)

		_, err = w.Write(data)
		require.NoError(t, err)
	}))
	defer ts.Close()

	r := &RavenDB{
		URL:          ts.URL,
		StatsInclude: []string{"server", "databases", "indexes", "collections"},
		Log:          testutil.Logger{},
	}

	require.NoError(t, r.Init())

	acc := &testutil.Accumulator{}

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)

	serverFields := map[string]interface{}{
		"server_version":                                                "5.1",
		"server_full_version":                                           "5.1.1-custom-51",
		"uptime_in_sec":                                                 int64(30),
		"server_process_id":                                             26360,
		"config_server_urls":                                            "http://127.0.0.1:8080;http://192.168.0.1:8080",
		"config_tcp_server_urls":                                        "tcp://127.0.0.1:3888;tcp://192.168.0.1:3888",
		"config_public_tcp_server_urls":                                 "tcp://2.3.4.5:3888;tcp://6.7.8.9:3888",
		"backup_max_number_of_concurrent_backups":                       4,
		"backup_current_number_of_running_backups":                      2,
		"cpu_process_usage":                                             6.28,
		"cpu_machine_usage":                                             41.05,
		"cpu_machine_io_wait":                                           2.55,
		"cpu_processor_count":                                           8,
		"cpu_assigned_processor_count":                                  7,
		"cpu_thread_pool_available_worker_threads":                      32766,
		"cpu_thread_pool_available_completion_port_threads":             1000,
		"memory_allocated_in_mb":                                        235,
		"memory_installed_in_mb":                                        16384,
		"memory_physical_in_mb":                                         16250,
		"memory_low_memory_severity":                                    "None",
		"memory_total_swap_size_in_mb":                                  1024,
		"memory_total_swap_usage_in_mb":                                 456,
		"memory_working_set_swap_usage_in_mb":                           89,
		"memory_total_dirty_in_mb":                                      1,
		"disk_system_store_used_data_file_size_in_mb":                   28,
		"disk_system_store_total_data_file_size_in_mb":                  32,
		"disk_total_free_space_in_mb":                                   52078,
		"disk_remaining_storage_space_percentage":                       22,
		"license_type":                                                  "Enterprise",
		"license_expiration_left_in_sec":                                25466947.5,
		"license_utilized_cpu_cores":                                    8,
		"license_max_cores":                                             256,
		"network_tcp_active_connections":                                84,
		"network_concurrent_requests_count":                             1,
		"network_total_requests":                                        3,
		"network_requests_per_sec":                                      0.03322,
		"network_last_request_time_in_sec":                              0.0264977,
		"network_last_authorized_non_cluster_admin_request_time_in_sec": 0.04,
		"certificate_server_certificate_expiration_left_in_sec":         float64(104),
		"certificate_well_known_admin_certificates":                     "a909502dd82ae41433e6f83886b00d4277a32a7b;4444444444444444444444444444444444444444",
		"cluster_node_state":                                            "Leader",
		"cluster_current_term":                                          28,
		"cluster_index":                                                 104,
		"databases_total_count":                                         25,
		"databases_loaded_count":                                        2,
	}

	serverTags := map[string]string{
		"url":               ts.URL,
		"node_tag":          "A",
		"cluster_id":        "6b535a18-558f-4e53-a479-a514efc16aab",
		"public_server_url": "http://raven1:8080",
	}

	defaultTime := time.Unix(0, 0)

	dbFields := map[string]interface{}{
		"uptime_in_sec":                               float64(1396),
		"time_since_last_backup_in_sec":               104.3,
		"counts_documents":                            425189,
		"counts_revisions":                            429605,
		"counts_attachments":                          17,
		"counts_unique_attachments":                   16,
		"counts_alerts":                               2,
		"counts_rehabs":                               3,
		"counts_performance_hints":                    5,
		"counts_replication_factor":                   2,
		"statistics_doc_puts_per_sec":                 23.4,
		"statistics_map_index_indexes_per_sec":        82.5,
		"statistics_map_reduce_index_mapped_per_sec":  50.3,
		"statistics_map_reduce_index_reduced_per_sec": 85.2,
		"statistics_requests_per_sec":                 22.5,
		"statistics_requests_count":                   809,
		"statistics_request_average_duration_in_ms":   0.55,
		"indexes_count":                               7,
		"indexes_stale_count":                         1,
		"indexes_errors_count":                        2,
		"indexes_static_count":                        7,
		"indexes_auto_count":                          3,
		"indexes_idle_count":                          4,
		"indexes_disabled_count":                      5,
		"indexes_errored_count":                       6,
		"storage_documents_allocated_data_file_in_mb": 1024,
		"storage_documents_used_data_file_in_mb":      942,
		"storage_indexes_allocated_data_file_in_mb":   464,
		"storage_indexes_used_data_file_in_mb":        278,
		"storage_total_allocated_storage_file_in_mb":  1496,
		"storage_total_free_space_in_mb":              52074,
	}

	dbTags := map[string]string{
		"url":               ts.URL,
		"node_tag":          "A",
		"database_name":     "db2",
		"database_id":       "06eefe8b-d720-4a8d-a809-2c5af9a4abb5",
		"public_server_url": "http://myhost:8080",
	}

	indexFields := map[string]interface{}{
		"priority":                        "Normal",
		"state":                           "Normal",
		"errors":                          0,
		"time_since_last_query_in_sec":    3.4712567,
		"time_since_last_indexing_in_sec": 3.4642612,
		"lock_mode":                       "Unlock",
		"is_invalid":                      true,
		"status":                          "Running",
		"mapped_per_sec":                  102.34,
		"reduced_per_sec":                 593.23,
		"type":                            "MapReduce",
	}

	indexTags := map[string]string{
		"url":               ts.URL,
		"node_tag":          "A",
		"public_server_url": "http://localhost:8080",
		"database_name":     "db1",
		"index_name":        "Product/Rating",
	}

	collectionFields := map[string]interface{}{
		"documents_count":          830,
		"total_size_in_bytes":      2744320,
		"documents_size_in_bytes":  868352,
		"tombstones_size_in_bytes": 122880,
		"revisions_size_in_bytes":  1753088,
	}

	collectionTags := map[string]string{
		"url":               ts.URL,
		"node_tag":          "A",
		"database_name":     "db1",
		"collection_name":   "Orders",
		"public_server_url": "http://localhost:8080",
	}

	serverExpected := testutil.MustMetric("ravendb_server", serverTags, serverFields, defaultTime)
	dbExpected := testutil.MustMetric("ravendb_databases", dbTags, dbFields, defaultTime)
	indexExpected := testutil.MustMetric("ravendb_indexes", indexTags, indexFields, defaultTime)
	collectionsExpected := testutil.MustMetric("ravendb_collections", collectionTags, collectionFields, defaultTime)

	for _, metric := range acc.GetTelegrafMetrics() {
		switch metric.Name() {
		case "ravendb_server":
			testutil.RequireMetricEqual(t, serverExpected, metric, testutil.IgnoreTime())
		case "ravendb_databases":
			testutil.RequireMetricEqual(t, dbExpected, metric, testutil.IgnoreTime())
		case "ravendb_indexes":
			testutil.RequireMetricEqual(t, indexExpected, metric, testutil.IgnoreTime())
		case "ravendb_collections":
			testutil.RequireMetricEqual(t, collectionsExpected, metric, testutil.IgnoreTime())
		}
	}
}

// Test against minimum filled data
func TestRavenDBGeneratesMetricsMin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonFilePath string

		switch r.URL.Path {
		case "/admin/monitoring/v1/databases":
			jsonFilePath = "testdata/databases_min.json"
		case "/admin/monitoring/v1/server":
			jsonFilePath = "testdata/server_min.json"
		case "/admin/monitoring/v1/indexes":
			jsonFilePath = "testdata/indexes_min.json"
		case "/admin/monitoring/v1/collections":
			jsonFilePath = "testdata/collections_min.json"
		default:
			require.Failf(t, "Cannot handle request for uri %s", r.URL.Path)
		}

		data, err := os.ReadFile(jsonFilePath)
		require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)

		_, err = w.Write(data)
		require.NoError(t, err)
	}))
	defer ts.Close()

	r := &RavenDB{
		URL:          ts.URL,
		StatsInclude: []string{"server", "databases", "indexes", "collections"},
		Log:          testutil.Logger{},
	}

	require.NoError(t, r.Init())

	acc := &testutil.Accumulator{}

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)

	serverFields := map[string]interface{}{
		"server_version":                                    "5.1",
		"server_full_version":                               "5.1.1-custom-51",
		"uptime_in_sec":                                     30,
		"server_process_id":                                 26360,
		"config_server_urls":                                "http://127.0.0.1:8080",
		"backup_max_number_of_concurrent_backups":           4,
		"backup_current_number_of_running_backups":          2,
		"cpu_process_usage":                                 6.28,
		"cpu_machine_usage":                                 41.07,
		"cpu_processor_count":                               8,
		"cpu_assigned_processor_count":                      7,
		"cpu_thread_pool_available_worker_threads":          32766,
		"cpu_thread_pool_available_completion_port_threads": 1000,
		"memory_allocated_in_mb":                            235,
		"memory_installed_in_mb":                            16384,
		"memory_physical_in_mb":                             16250,
		"memory_low_memory_severity":                        "Low",
		"memory_total_swap_size_in_mb":                      1024,
		"memory_total_swap_usage_in_mb":                     456,
		"memory_working_set_swap_usage_in_mb":               89,
		"memory_total_dirty_in_mb":                          1,
		"disk_system_store_used_data_file_size_in_mb":       28,
		"disk_system_store_total_data_file_size_in_mb":      32,
		"disk_total_free_space_in_mb":                       52078,
		"disk_remaining_storage_space_percentage":           22,
		"license_type":                                      "Enterprise",
		"license_utilized_cpu_cores":                        8,
		"license_max_cores":                                 256,
		"network_tcp_active_connections":                    84,
		"network_concurrent_requests_count":                 1,
		"network_total_requests":                            3,
		"network_requests_per_sec":                          0.03322,
		"cluster_node_state":                                "Leader",
		"cluster_current_term":                              28,
		"cluster_index":                                     104,
		"databases_total_count":                             25,
		"databases_loaded_count":                            2,
	}

	serverTags := map[string]string{
		"url":        ts.URL,
		"node_tag":   "A",
		"cluster_id": "6b535a18-558f-4e53-a479-a514efc16aab",
	}

	dbFields := map[string]interface{}{
		"uptime_in_sec":                               float64(1396),
		"counts_documents":                            425189,
		"counts_revisions":                            429605,
		"counts_attachments":                          17,
		"counts_unique_attachments":                   16,
		"counts_alerts":                               2,
		"counts_rehabs":                               3,
		"counts_performance_hints":                    5,
		"counts_replication_factor":                   2,
		"statistics_doc_puts_per_sec":                 23.4,
		"statistics_map_index_indexes_per_sec":        82.5,
		"statistics_map_reduce_index_mapped_per_sec":  50.3,
		"statistics_map_reduce_index_reduced_per_sec": 85.2,
		"statistics_requests_per_sec":                 22.5,
		"statistics_requests_count":                   809,
		"statistics_request_average_duration_in_ms":   0.55,
		"indexes_count":                               7,
		"indexes_stale_count":                         1,
		"indexes_errors_count":                        2,
		"indexes_static_count":                        7,
		"indexes_auto_count":                          3,
		"indexes_idle_count":                          4,
		"indexes_disabled_count":                      5,
		"indexes_errored_count":                       6,
		"storage_documents_allocated_data_file_in_mb": 1024,
		"storage_documents_used_data_file_in_mb":      942,
		"storage_indexes_allocated_data_file_in_mb":   464,
		"storage_indexes_used_data_file_in_mb":        278,
		"storage_total_allocated_storage_file_in_mb":  1496,
		"storage_total_free_space_in_mb":              52074,
	}

	dbTags := map[string]string{
		"url":           ts.URL,
		"node_tag":      "A",
		"database_name": "db2",
		"database_id":   "06eefe8b-d720-4a8d-a809-2c5af9a4abb5",
	}

	indexFields := map[string]interface{}{
		"priority":        "Normal",
		"state":           "Normal",
		"errors":          0,
		"lock_mode":       "Unlock",
		"is_invalid":      false,
		"status":          "Running",
		"mapped_per_sec":  102.34,
		"reduced_per_sec": 593.23,
		"type":            "MapReduce",
	}

	indexTags := map[string]string{
		"url":           ts.URL,
		"node_tag":      "A",
		"database_name": "db1",
		"index_name":    "Product/Rating",
	}

	collectionFields := map[string]interface{}{
		"documents_count":          830,
		"total_size_in_bytes":      2744320,
		"documents_size_in_bytes":  868352,
		"tombstones_size_in_bytes": 122880,
		"revisions_size_in_bytes":  1753088,
	}

	collectionTags := map[string]string{
		"url":             ts.URL,
		"node_tag":        "A",
		"database_name":   "db1",
		"collection_name": "Orders",
	}

	defaultTime := time.Unix(0, 0)

	serverExpected := testutil.MustMetric("ravendb_server", serverTags, serverFields, defaultTime)
	dbExpected := testutil.MustMetric("ravendb_databases", dbTags, dbFields, defaultTime)
	indexExpected := testutil.MustMetric("ravendb_indexes", indexTags, indexFields, defaultTime)
	collectionsExpected := testutil.MustMetric("ravendb_collections", collectionTags, collectionFields, defaultTime)

	for _, metric := range acc.GetTelegrafMetrics() {
		switch metric.Name() {
		case "ravendb_server":
			testutil.RequireMetricEqual(t, serverExpected, metric, testutil.IgnoreTime())
		case "ravendb_databases":
			testutil.RequireMetricEqual(t, dbExpected, metric, testutil.IgnoreTime())
		case "ravendb_indexes":
			testutil.RequireMetricEqual(t, indexExpected, metric, testutil.IgnoreTime())
		case "ravendb_collections":
			testutil.RequireMetricEqual(t, collectionsExpected, metric, testutil.IgnoreTime())
		}
	}
}
