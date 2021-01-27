package ravendb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
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
			panic(fmt.Sprintf("Cannot handle request for uri %s", r.URL.Path))
		}

		data, err := ioutil.ReadFile(jsonFilePath)

		if err != nil {
			panic(fmt.Sprintf("could not read from data file %s", jsonFilePath))
		}

		w.Write(data)
	}))
	defer ts.Close()

	r := &RavenDB{
		URL:                   ts.URL,
		GatherDbStats:         true,
		GatherIndexStats:      true,
		GatherServerStats:     true,
		GatherCollectionStats: true,
	}

	acc := &testutil.Accumulator{}

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)

	serverFields := map[string]interface{}{
		"server_version":                                                "5.1",
		"server_full_version":                                           "5.1.1-custom-51",
		"uptime_in_sec":                                                 30,
		"server_process_id":                                             26360,
		"config_server_urls":                                            "http://127.0.0.1:8080;http://192.168.0.1:8080",
		"config_tcp_server_urls":                                        "tcp://127.0.0.1:3888;tcp://192.168.0.1:3888",
		"config_public_tcp_server_urls":                                 "tcp://2.3.4.5:3888;tcp://6.7.8.9:3888",
		"backup_max_number_of_concurrent_backups":                       4,
		"backup_current_number_of_running_backups":                      2,
		"cpu_process_usage":                                             6.28442,
		"cpu_machine_usage":                                             41.0779,
		"cpu_machine_io_wait":                                           2.55,
		"cpu_processor_count":                                           8,
		"cpu_assigned_processor_count":                                  7,
		"cpu_thread_pool_available_worker_threads":                      32766,
		"cpu_thread_pool_available_completion_port_threads":             1000,
		"memory_allocated_in_mb":                                        235,
		"memory_installed_in_mb":                                        16384,
		"memory_physical_in_mb":                                         16250,
		"memory_low_memory_severity":                                    0,
		"memory_total_swap_size_in_mb":                                  1024,
		"memory_total_swap_usage_in_mb":                                 456,
		"memory_working_set_swap_usage_in_mb":                           89,
		"memory_total_dirty_in_mb":                                      1,
		"disk_system_store_used_data_file_size_in_mb":                   28,
		"disk_system_store_total_data_file_size_in_mb":                  32,
		"disk_total_free_space_in_mb":                                   52078,
		"disk_remaining_storage_space_percentage":                       22,
		"license_type":                                                  "Enterprise",
		"license_expiration_left_in_sec":                                25466947.94,
		"license_utilized_cpu_cores":                                    8,
		"license_max_cores":                                             256,
		"network_tcp_active_connections":                                84,
		"network_concurrent_requests_count":                             1,
		"network_total_requests":                                        3,
		"network_requests_per_sec":                                      0.03322,
		"network_last_request_time_in_sec":                              0.0265,
		"network_last_authorized_non_cluster_admin_request_time_in_sec": 0.04,
		"certificate_server_certificate_expiration_left_in_sec":         104,
		"certificate_well_known_admin_certificates":                     "a909502dd82ae41433e6f83886b00d4277a32a7b;4444444444444444444444444444444444444444",
		"cluster_node_state":                                            4,
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

	compareFields(t, serverFields, acc, "ravendb_server")
	compareTags(t, serverTags, acc, "ravendb_server")

	dbFields := map[string]interface{}{
		"uptime_in_sec":                               1396,
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
		"statistics_request_average_duration":         0.55,
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

	compareFields(t, dbFields, acc, "ravendb_databases")
	compareTags(t, dbTags, acc, "ravendb_databases")

	indexFields := map[string]interface{}{
		"priority":                        "Normal",
		"state":                           "Normal",
		"errors":                          0,
		"time_since_last_query_in_sec":    3.4712567,
		"time_since_last_indexing_in_sec": 3.4642612,
		"lock_mode":                       "Unlock",
		"is_invalid":                      1,
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

	compareFields(t, indexFields, acc, "ravendb_indexes")
	compareTags(t, indexTags, acc, "ravendb_indexes")

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

	compareFields(t, collectionFields, acc, "ravendb_collections")
	compareTags(t, collectionTags, acc, "ravendb_collections")
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
			panic(fmt.Sprintf("Cannot handle request for uri %s", r.URL.Path))
		}

		data, err := ioutil.ReadFile(jsonFilePath)

		if err != nil {
			panic(fmt.Sprintf("could not read from data file %s", jsonFilePath))
		}

		w.Write(data)
	}))
	defer ts.Close()

	r := &RavenDB{
		URL:                   ts.URL,
		GatherServerStats:     true,
		GatherDbStats:         true,
		GatherIndexStats:      true,
		GatherCollectionStats: true,
	}

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
		"cpu_process_usage":                                 6.28442,
		"cpu_machine_usage":                                 41.0779,
		"cpu_processor_count":                               8,
		"cpu_assigned_processor_count":                      7,
		"cpu_thread_pool_available_worker_threads":          32766,
		"cpu_thread_pool_available_completion_port_threads": 1000,
		"memory_allocated_in_mb":                            235,
		"memory_installed_in_mb":                            16384,
		"memory_physical_in_mb":                             16250,
		"memory_low_memory_severity":                        2,
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
		"cluster_node_state":                                4,
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

	compareFields(t, serverFields, acc, "ravendb_server")
	compareTags(t, serverTags, acc, "ravendb_server")

	dbFields := map[string]interface{}{
		"uptime_in_sec":                               1396,
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
		"statistics_request_average_duration":         0.55,
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

	compareFields(t, dbFields, acc, "ravendb_databases")
	compareTags(t, dbTags, acc, "ravendb_databases")

	indexFields := map[string]interface{}{
		"priority":        "Normal",
		"state":           "Normal",
		"errors":          0,
		"lock_mode":       "Unlock",
		"is_invalid":      0,
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

	compareFields(t, indexFields, acc, "ravendb_indexes")
	compareTags(t, indexTags, acc, "ravendb_indexes")

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

	compareFields(t, collectionFields, acc, "ravendb_collections")
	compareTags(t, collectionTags, acc, "ravendb_collections")
}

func compareFields(t *testing.T, expectedFields map[string]interface{},
	accumulator *testutil.Accumulator, measurementKey string) {

	measurement, exist := accumulator.Get(measurementKey)

	assert.True(t, exist, "There is measurement %s", measurementKey)
	assert.Equal(t, len(expectedFields), len(measurement.Fields))

	for metricName, metricValue := range expectedFields {
		actualMetricValue := measurement.Fields[metricName]

		if accumulator.HasStringField(measurementKey, metricName) {
			assert.Equal(t, metricValue, actualMetricValue,
				"Metric name: %s", metricName)
		} else {
			floatValue, err := getFloat(actualMetricValue)
			if err != nil {
				assert.Error(t, err)
			}
			assert.InDelta(t, metricValue, floatValue, 2,
				"Metric name: %s", metricName)
		}
	}
}

func compareTags(t *testing.T, expectedTags map[string]string,
	accumulator *testutil.Accumulator, measurementKey string) {

	measurement, exist := accumulator.Get(measurementKey)

	assert.True(t, exist, "There is measurement %s", measurementKey)
	assert.Equal(t, len(expectedTags), len(measurement.Tags))

	for tagName, tagValue := range expectedTags {
		actualTagValue := measurement.Tags[tagName]

		if accumulator.HasTag(measurementKey, tagName) {
			assert.Equal(t, tagValue, actualTagValue,
				"Tag name: %s", tagName)
		} else {
			assert.Fail(t, "Missing tag: %s", tagName)
		}
	}
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
