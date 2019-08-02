package mesos

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var masterMetrics map[string]interface{}
var masterTestServer *httptest.Server
var slaveMetrics map[string]interface{}

// var slaveTaskMetrics map[string]interface{}
var slaveTestServer *httptest.Server

func randUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func generateMetrics() {
	masterMetrics = make(map[string]interface{})

	metricNames := []string{
		// resources
		"master/cpus_percent",
		"master/cpus_used",
		"master/cpus_total",
		"master/cpus_revocable_percent",
		"master/cpus_revocable_total",
		"master/cpus_revocable_used",
		"master/disk_percent",
		"master/disk_used",
		"master/disk_total",
		"master/disk_revocable_percent",
		"master/disk_revocable_total",
		"master/disk_revocable_used",
		"master/gpus_percent",
		"master/gpus_used",
		"master/gpus_total",
		"master/gpus_revocable_percent",
		"master/gpus_revocable_total",
		"master/gpus_revocable_used",
		"master/mem_percent",
		"master/mem_used",
		"master/mem_total",
		"master/mem_revocable_percent",
		"master/mem_revocable_total",
		"master/mem_revocable_used",
		// master
		"master/elected",
		"master/uptime_secs",
		// system
		"system/cpus_total",
		"system/load_15min",
		"system/load_5min",
		"system/load_1min",
		"system/mem_free_bytes",
		"system/mem_total_bytes",
		// agents
		"master/slave_registrations",
		"master/slave_removals",
		"master/slave_reregistrations",
		"master/slave_shutdowns_scheduled",
		"master/slave_shutdowns_canceled",
		"master/slave_shutdowns_completed",
		"master/slaves_active",
		"master/slaves_connected",
		"master/slaves_disconnected",
		"master/slaves_inactive",
		"master/slave_unreachable_canceled",
		"master/slave_unreachable_completed",
		"master/slave_unreachable_scheduled",
		"master/slaves_unreachable",
		// frameworks
		"master/frameworks_active",
		"master/frameworks_connected",
		"master/frameworks_disconnected",
		"master/frameworks_inactive",
		"master/outstanding_offers",
		// framework offers
		"master/frameworks/marathon/abc-123/calls",
		"master/frameworks/marathon/abc-123/calls/accept",
		"master/frameworks/marathon/abc-123/events",
		"master/frameworks/marathon/abc-123/events/error",
		"master/frameworks/marathon/abc-123/offers/sent",
		"master/frameworks/marathon/abc-123/operations",
		"master/frameworks/marathon/abc-123/operations/create",
		"master/frameworks/marathon/abc-123/roles/*/suppressed",
		"master/frameworks/marathon/abc-123/subscribed",
		"master/frameworks/marathon/abc-123/tasks/active/task_killing",
		"master/frameworks/marathon/abc-123/tasks/active/task_dropped",
		"master/frameworks/marathon/abc-123/tasks/terminal/task_dropped",
		"master/frameworks/marathon/abc-123/unknown/unknown", // test case for unknown metric type
		// tasks
		"master/tasks_error",
		"master/tasks_failed",
		"master/tasks_finished",
		"master/tasks_killed",
		"master/tasks_lost",
		"master/tasks_running",
		"master/tasks_staging",
		"master/tasks_starting",
		"master/tasks_dropped",
		"master/tasks_gone",
		"master/tasks_gone_by_operator",
		"master/tasks_killing",
		"master/tasks_unreachable",
		// messages
		"master/invalid_executor_to_framework_messages",
		"master/invalid_framework_to_executor_messages",
		"master/invalid_status_update_acknowledgements",
		"master/invalid_status_updates",
		"master/dropped_messages",
		"master/messages_authenticate",
		"master/messages_deactivate_framework",
		"master/messages_decline_offers",
		"master/messages_executor_to_framework",
		"master/messages_exited_executor",
		"master/messages_framework_to_executor",
		"master/messages_kill_task",
		"master/messages_launch_tasks",
		"master/messages_reconcile_tasks",
		"master/messages_register_framework",
		"master/messages_register_slave",
		"master/messages_reregister_framework",
		"master/messages_reregister_slave",
		"master/messages_resource_request",
		"master/messages_revive_offers",
		"master/messages_status_update",
		"master/messages_status_update_acknowledgement",
		"master/messages_unregister_framework",
		"master/messages_unregister_slave",
		"master/messages_update_slave",
		"master/recovery_slave_removals",
		"master/slave_removals/reason_registered",
		"master/slave_removals/reason_unhealthy",
		"master/slave_removals/reason_unregistered",
		"master/valid_framework_to_executor_messages",
		"master/valid_status_update_acknowledgements",
		"master/valid_status_updates",
		"master/task_lost/source_master/reason_invalid_offers",
		"master/task_lost/source_master/reason_slave_removed",
		"master/task_lost/source_slave/reason_executor_terminated",
		"master/valid_executor_to_framework_messages",
		"master/invalid_operation_status_update_acknowledgements",
		"master/messages_operation_status_update_acknowledgement",
		"master/messages_reconcile_operations",
		"master/messages_suppress_offers",
		"master/valid_operation_status_update_acknowledgements",
		// evgqueue
		"master/event_queue_dispatches",
		"master/event_queue_http_requests",
		"master/event_queue_messages",
		"master/operator_event_stream_subscribers",
		// registrar
		"registrar/log/ensemble_size",
		"registrar/log/recovered",
		"registrar/queued_operations",
		"registrar/registry_size_bytes",
		"registrar/state_fetch_ms",
		"registrar/state_store_ms",
		"registrar/state_store_ms/max",
		"registrar/state_store_ms/min",
		"registrar/state_store_ms/p50",
		"registrar/state_store_ms/p90",
		"registrar/state_store_ms/p95",
		"registrar/state_store_ms/p99",
		"registrar/state_store_ms/p999",
		"registrar/state_store_ms/p9999",
		"registrar/state_store_ms/count",
		// allocator
		"allocator/mesos/allocation_run_ms",
		"allocator/mesos/allocation_run_ms/count",
		"allocator/mesos/allocation_run_ms/max",
		"allocator/mesos/allocation_run_ms/min",
		"allocator/mesos/allocation_run_ms/p50",
		"allocator/mesos/allocation_run_ms/p90",
		"allocator/mesos/allocation_run_ms/p95",
		"allocator/mesos/allocation_run_ms/p99",
		"allocator/mesos/allocation_run_ms/p999",
		"allocator/mesos/allocation_run_ms/p9999",
		"allocator/mesos/allocation_runs",
		"allocator/mesos/allocation_run_latency_ms",
		"allocator/mesos/allocation_run_latency_ms/count",
		"allocator/mesos/allocation_run_latency_ms/max",
		"allocator/mesos/allocation_run_latency_ms/min",
		"allocator/mesos/allocation_run_latency_ms/p50",
		"allocator/mesos/allocation_run_latency_ms/p90",
		"allocator/mesos/allocation_run_latency_ms/p95",
		"allocator/mesos/allocation_run_latency_ms/p99",
		"allocator/mesos/allocation_run_latency_ms/p999",
		"allocator/mesos/allocation_run_latency_ms/p9999",
		"allocator/mesos/roles/*/shares/dominant",
		// test case against hash collisions in TaggedFields
		// e.g. framework_name=marathon and role_name=marathon
		"allocator/mesos/roles/marathon/shares/dominant",
		"allocator/mesos/event_queue_dispatches",
		"allocator/mesos/offer_filters/roles/*/active",
		"allocator/mesos/quota/roles/*/resources/disk/offered_or_allocated",
		"allocator/mesos/quota/roles/*/resources/mem/guarantee",
		"allocator/mesos/quota/roles/*/resources/disk/guarantee",
		"allocator/mesos/resources/cpus/offered_or_allocated",
		"allocator/mesos/resources/cpus/total",
		"allocator/mesos/resources/disk/offered_or_allocated",
		"allocator/mesos/resources/disk/total",
		"allocator/mesos/resources/mem/offered_or_allocated",
		"allocator/mesos/resources/mem/total",
		"allocator/mesos/unknown/unknown/unknown/unknown", // test case for unknown metric type
	}

	for _, k := range metricNames {
		masterMetrics[k] = rand.Float64()
	}

	slaveMetrics = make(map[string]interface{})

	metricNames = []string{
		// resources
		"slave/cpus_percent",
		"slave/cpus_used",
		"slave/cpus_total",
		"slave/cpus_revocable_percent",
		"slave/cpus_revocable_total",
		"slave/cpus_revocable_used",
		"slave/disk_percent",
		"slave/disk_used",
		"slave/disk_total",
		"slave/disk_revocable_percent",
		"slave/disk_revocable_total",
		"slave/disk_revocable_used",
		"slave/gpus_percent",
		"slave/gpus_used",
		"slave/gpus_total",
		"slave/gpus_revocable_percent",
		"slave/gpus_revocable_total",
		"slave/gpus_revocable_used",
		"slave/mem_percent",
		"slave/mem_used",
		"slave/mem_total",
		"slave/mem_revocable_percent",
		"slave/mem_revocable_total",
		"slave/mem_revocable_used",
		// agent
		"slave/registered",
		"slave/uptime_secs",
		// system
		"system/cpus_total",
		"system/load_15min",
		"system/load_5min",
		"system/load_1min",
		"system/mem_free_bytes",
		"system/mem_total_bytes",
		// executors
		"containerizer/mesos/container_destroy_errors",
		"slave/container_launch_errors",
		"slave/executors_preempted",
		"slave/frameworks_active",
		"slave/executor_directory_max_allowed_age_secs",
		"slave/executors_registering",
		"slave/executors_running",
		"slave/executors_terminated",
		"slave/executors_terminating",
		"slave/recovery_errors",
		// tasks
		"slave/tasks_failed",
		"slave/tasks_finished",
		"slave/tasks_killed",
		"slave/tasks_lost",
		"slave/tasks_running",
		"slave/tasks_staging",
		"slave/tasks_starting",
		// messages
		"slave/invalid_framework_messages",
		"slave/invalid_status_updates",
		"slave/valid_framework_messages",
		"slave/valid_status_updates",
	}

	for _, k := range metricNames {
		slaveMetrics[k] = rand.Float64()
	}

	// slaveTaskMetrics = map[string]interface{}{
	// 	"executor_id":   fmt.Sprintf("task_name.%s", randUUID()),
	// 	"executor_name": "Some task description",
	// 	"framework_id":  randUUID(),
	// 	"source":        fmt.Sprintf("task_source.%s", randUUID()),
	// 	"statistics": map[string]interface{}{
	// 		"cpus_limit":                    rand.Float64(),
	// 		"cpus_system_time_secs":         rand.Float64(),
	// 		"cpus_user_time_secs":           rand.Float64(),
	// 		"mem_anon_bytes":                float64(rand.Int63()),
	// 		"mem_cache_bytes":               float64(rand.Int63()),
	// 		"mem_critical_pressure_counter": float64(rand.Int63()),
	// 		"mem_file_bytes":                float64(rand.Int63()),
	// 		"mem_limit_bytes":               float64(rand.Int63()),
	// 		"mem_low_pressure_counter":      float64(rand.Int63()),
	// 		"mem_mapped_file_bytes":         float64(rand.Int63()),
	// 		"mem_medium_pressure_counter":   float64(rand.Int63()),
	// 		"mem_rss_bytes":                 float64(rand.Int63()),
	// 		"mem_swap_bytes":                float64(rand.Int63()),
	// 		"mem_total_bytes":               float64(rand.Int63()),
	// 		"mem_total_memsw_bytes":         float64(rand.Int63()),
	// 		"mem_unevictable_bytes":         float64(rand.Int63()),
	// 		"timestamp":                     rand.Float64(),
	// 	},
	// }
}

func TestMain(m *testing.M) {
	generateMetrics()

	masterRouter := http.NewServeMux()
	masterRouter.HandleFunc("/metrics/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(masterMetrics)
	})
	masterTestServer = httptest.NewServer(masterRouter)

	slaveRouter := http.NewServeMux()
	slaveRouter.HandleFunc("/metrics/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(slaveMetrics)
	})
	// slaveRouter.HandleFunc("/monitor/statistics", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode([]map[string]interface{}{slaveTaskMetrics})
	// })
	slaveTestServer = httptest.NewServer(slaveRouter)

	rc := m.Run()

	masterTestServer.Close()
	slaveTestServer.Close()
	os.Exit(rc)
}

func TestMesosMaster(t *testing.T) {
	var acc testutil.Accumulator

	m := Mesos{
		Masters: []string{masterTestServer.Listener.Addr().String()},
		Timeout: 10,
	}

	err := acc.GatherError(m.Gather)

	if err != nil {
		t.Errorf(err.Error())
	}

	expectedUntaggedMetrics := map[string]interface{}{}
	for k, v := range masterMetrics {
		parts := strings.Split(k, "/")
		if !strings.HasPrefix(k, "master/frameworks/") && (!strings.HasPrefix(k, "allocator/") || len(parts) <= 5) {
			expectedUntaggedMetrics[k] = v
		}
	}
	// for unknown allocator metric type test case, expect no additional tags
	expectedUntaggedMetrics["allocator/mesos/unknown/unknown/unknown/unknown"] = masterMetrics["allocator/mesos/unknown/unknown/unknown/unknown"]

	acc.AssertContainsFields(t, "mesos", expectedUntaggedMetrics)

	frameworkFields := []map[string]interface{}{
		// framework offers
		{
			// "unknown" metric type should still contain framework_name tag
			"master/frameworks/unknown/unknown":  masterMetrics["master/frameworks/marathon/abc-123/unknown/unknown"],
			"master/frameworks/calls_total":      masterMetrics["master/frameworks/marathon/abc-123/calls"],
			"master/frameworks/events_total":     masterMetrics["master/frameworks/marathon/abc-123/events"],
			"master/frameworks/operations_total": masterMetrics["master/frameworks/marathon/abc-123/operations"],
			"master/frameworks/subscribed_total": masterMetrics["master/frameworks/marathon/abc-123/subscribed"],
			"master/frameworks/offers/sent":      masterMetrics["master/frameworks/marathon/abc-123/offers/sent"],
		},
		{
			"master/frameworks/tasks/active": masterMetrics["master/frameworks/marathon/abc-123/tasks/active/task_killing"],
		},
		{
			"master/frameworks/tasks/active":   masterMetrics["master/frameworks/marathon/abc-123/tasks/active/task_dropped"],
			"master/frameworks/tasks/terminal": masterMetrics["master/frameworks/marathon/abc-123/tasks/terminal/task_dropped"],
		},
		{
			"master/frameworks/roles/suppressed": masterMetrics["master/frameworks/marathon/abc-123/roles/*/suppressed"],
		},
		{
			"master/frameworks/calls": masterMetrics["master/frameworks/marathon/abc-123/calls/accept"],
		},
		{
			"master/frameworks/events": masterMetrics["master/frameworks/marathon/abc-123/events/error"],
		},
		{
			"master/frameworks/operations": masterMetrics["master/frameworks/marathon/abc-123/operations/create"],
		},
		// allocator
		{
			"allocator/roles/shares/dominant":      masterMetrics["allocator/mesos/roles/*/shares/dominant"],
			"allocator/offer_filters/roles/active": masterMetrics["allocator/mesos/offer_filters/roles/*/active"],
		},
		{
			"allocator/roles/shares/dominant": masterMetrics["allocator/mesos/roles/marathon/shares/dominant"],
		},
		{
			"allocator/quota/roles/resources/offered_or_allocated": masterMetrics["allocator/mesos/quota/roles/*/resources/disk/offered_or_allocated"],
			"allocator/quota/roles/resources/guarantee":            masterMetrics["allocator/mesos/quota/roles/*/resources/disk/guarantee"],
		},
		{
			"allocator/quota/roles/resources/guarantee": masterMetrics["allocator/mesos/quota/roles/*/resources/mem/guarantee"],
		},
	}

	frameworkTags := []map[string]string{
		// framework offers
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
		},
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
			"task_state":     "task_killing",
		},
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
			"task_state":     "task_dropped",
		},
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
			"role_name":      "*",
		},
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
			"call_type":      "accept",
		},
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
			"event_type":     "error",
		},
		{
			"server":         m.masterURLs[0].Hostname(),
			"url":            masterTestServer.URL,
			"role":           "master",
			"state":          "leader",
			"framework_name": "marathon",
			"operation_type": "create",
		},
		// allocator
		{
			"server":    m.masterURLs[0].Hostname(),
			"url":       masterTestServer.URL,
			"role":      "master",
			"state":     "leader",
			"role_name": "*",
		},
		{
			"server":    m.masterURLs[0].Hostname(),
			"url":       masterTestServer.URL,
			"role":      "master",
			"state":     "leader",
			"role_name": "marathon",
		},
		{
			"server":    m.masterURLs[0].Hostname(),
			"url":       masterTestServer.URL,
			"role":      "master",
			"state":     "leader",
			"role_name": "*",
			"resource":  "disk",
		},
		{
			"server":    m.masterURLs[0].Hostname(),
			"url":       masterTestServer.URL,
			"role":      "master",
			"state":     "leader",
			"role_name": "*",
			"resource":  "mem",
		},
	}

	for i := 0; i < len(frameworkFields); i++ {
		acc.AssertContainsTaggedFields(t, "mesos", frameworkFields[i], frameworkTags[i])
		// Test that none of the other metrics share the same tags, which
		// tests against potential hash collisions in TaggedFields.
		for j := 0; j < len(frameworkFields); j++ {
			if j == i {
				continue
			}
			acc.AssertDoesNotContainsTaggedFields(t, "mesos", frameworkFields[j], frameworkTags[i])
		}
	}
}

func TestMasterFilter(t *testing.T) {
	m := Mesos{
		MasterCols: []string{
			"resources", "master", "registrar",
		},
	}
	b := []string{
		"system", "agents", "frameworks",
		"messages", "evqueue", "tasks",
	}

	m.filterMetrics(MASTER, &masterMetrics)

	for _, v := range b {
		for _, x := range getMetrics(MASTER, v) {
			if _, ok := masterMetrics[x]; ok {
				t.Errorf("Found key %s, it should be gone.", x)
			}
		}
	}
	for _, v := range m.MasterCols {
		for _, x := range getMetrics(MASTER, v) {
			if _, ok := masterMetrics[x]; !ok {
				t.Errorf("Didn't find key %s, it should present.", x)
			}
		}
	}
}

func TestMesosSlave(t *testing.T) {
	var acc testutil.Accumulator

	m := Mesos{
		Masters: []string{},
		Slaves:  []string{slaveTestServer.Listener.Addr().String()},
		// SlaveTasks: true,
		Timeout: 10,
	}

	err := acc.GatherError(m.Gather)

	if err != nil {
		t.Errorf(err.Error())
	}

	acc.AssertContainsFields(t, "mesos", slaveMetrics)

	// expectedFields := make(map[string]interface{}, len(slaveTaskMetrics["statistics"].(map[string]interface{}))+1)
	// for k, v := range slaveTaskMetrics["statistics"].(map[string]interface{}) {
	// 	expectedFields[k] = v
	// }
	// expectedFields["executor_id"] = slaveTaskMetrics["executor_id"]

	// acc.AssertContainsTaggedFields(
	// 	t,
	// 	"mesos_tasks",
	// 	expectedFields,
	// 	map[string]string{"server": "127.0.0.1", "framework_id": slaveTaskMetrics["framework_id"].(string)})
}

func TestSlaveFilter(t *testing.T) {
	m := Mesos{
		SlaveCols: []string{
			"resources", "agent", "tasks",
		},
	}
	b := []string{
		"system", "executors", "messages",
	}

	m.filterMetrics(SLAVE, &slaveMetrics)

	for _, v := range b {
		for _, x := range getMetrics(SLAVE, v) {
			if _, ok := slaveMetrics[x]; ok {
				t.Errorf("Found key %s, it should be gone.", x)
			}
		}
	}
	for _, v := range m.MasterCols {
		for _, x := range getMetrics(SLAVE, v) {
			if _, ok := slaveMetrics[x]; !ok {
				t.Errorf("Didn't find key %s, it should present.", x)
			}
		}
	}
}

func TestWithPathDoesNotModify(t *testing.T) {
	u, err := url.Parse("http://localhost:5051")
	require.NoError(t, err)
	v := withPath(u, "/xyzzy")
	require.Equal(t, u.String(), "http://localhost:5051")
	require.Equal(t, v.String(), "http://localhost:5051/xyzzy")
}

func TestURLTagDoesNotModify(t *testing.T) {
	u, err := url.Parse("http://a:b@localhost:5051?timeout=1ms")
	require.NoError(t, err)
	v := urlTag(u)
	require.Equal(t, u.String(), "http://a:b@localhost:5051?timeout=1ms")
	require.Equal(t, v, "http://localhost:5051")
}

func TestTaggedFieldHash(t *testing.T) {
	assert := assert.New(t)
	tf := TaggedField{
		FrameworkName: "marathon",
		CallType:      "accept",
		EventType:     "error",
		OperationType: "create",
		TaskState:     "active",
		RoleName:      "marathon",
		FieldName:     "field/name",
		Resource:      "mem",
		Value:         1.0,
	}
	assert.Equal("tf_fn:marathon_ct:accept_et:error_ot:create_ts:active_rn:marathon_r:mem", tf.hash())

	// Test against hash collisions
	tf1 := TaggedField{
		FrameworkName: "marathon",
		FieldName:     "field/name/1",
		Value:         1.0,
	}
	tf2 := TaggedField{
		RoleName:  "marathon",
		FieldName: "field/name/2",
		Value:     1.0,
	}
	assert.NotEqual(tf1.hash(), tf2.hash())
}

func TestTaggedFieldTags(t *testing.T) {
	assert := assert.New(t)
	tf := TaggedField{
		FrameworkName: "marathon",
		CallType:      "accept",
		EventType:     "error",
		OperationType: "create",
		TaskState:     "active",
		RoleName:      "marathon",
		FieldName:     "field/name",
		Resource:      "mem",
		Value:         1.0,
	}

	expectedTags := fieldTags{
		"framework_name": "marathon",
		"call_type":      "accept",
		"event_type":     "error",
		"operation_type": "create",
		"task_state":     "active",
		"role_name":      "marathon",
		"resource":       "mem",
	}

	assert.Equal(expectedTags, tf.tags())
}
