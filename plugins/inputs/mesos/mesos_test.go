package mesos

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var mesosMetrics map[string]interface{}
var ts *httptest.Server

func generateMetrics() {
	mesosMetrics = make(map[string]interface{})

	metricNames := []string{"master/cpus_percent", "master/cpus_used", "master/cpus_total",
		"master/cpus_revocable_percent", "master/cpus_revocable_total", "master/cpus_revocable_used",
		"master/disk_percent", "master/disk_used", "master/disk_total", "master/disk_revocable_percent",
		"master/disk_revocable_total", "master/disk_revocable_used", "master/mem_percent",
		"master/mem_used", "master/mem_total", "master/mem_revocable_percent", "master/mem_revocable_total",
		"master/mem_revocable_used", "master/elected", "master/uptime_secs", "system/cpus_total",
		"system/load_15min", "system/load_5min", "system/load_1min", "system/mem_free_bytes",
		"system/mem_total_bytes", "master/slave_registrations", "master/slave_removals",
		"master/slave_reregistrations", "master/slave_shutdowns_scheduled", "master/slave_shutdowns_canceled",
		"master/slave_shutdowns_completed", "master/slaves_active", "master/slaves_connected",
		"master/slaves_disconnected", "master/slaves_inactive", "master/frameworks_active",
		"master/frameworks_connected", "master/frameworks_disconnected", "master/frameworks_inactive",
		"master/outstanding_offers", "master/tasks_error", "master/tasks_failed", "master/tasks_finished",
		"master/tasks_killed", "master/tasks_lost", "master/tasks_running", "master/tasks_staging",
		"master/tasks_starting", "master/invalid_executor_to_framework_messages", "master/invalid_framework_to_executor_messages",
		"master/invalid_status_update_acknowledgements", "master/invalid_status_updates",
		"master/dropped_messages", "master/messages_authenticate", "master/messages_deactivate_framework",
		"master/messages_decline_offers", "master/messages_executor_to_framework", "master/messages_exited_executor",
		"master/messages_framework_to_executor", "master/messages_kill_task", "master/messages_launch_tasks",
		"master/messages_reconcile_tasks", "master/messages_register_framework", "master/messages_register_slave",
		"master/messages_reregister_framework", "master/messages_reregister_slave", "master/messages_resource_request",
		"master/messages_revive_offers", "master/messages_status_update", "master/messages_status_update_acknowledgement",
		"master/messages_unregister_framework", "master/messages_unregister_slave", "master/messages_update_slave",
		"master/recovery_slave_removals", "master/slave_removals/reason_registered", "master/slave_removals/reason_unhealthy",
		"master/slave_removals/reason_unregistered", "master/valid_framework_to_executor_messages", "master/valid_status_update_acknowledgements",
		"master/valid_status_updates", "master/task_lost/source_master/reason_invalid_offers",
		"master/task_lost/source_master/reason_slave_removed", "master/task_lost/source_slave/reason_executor_terminated",
		"master/valid_executor_to_framework_messages", "master/event_queue_dispatches",
		"master/event_queue_http_requests", "master/event_queue_messages", "registrar/state_fetch_ms",
		"registrar/state_store_ms", "registrar/state_store_ms/max", "registrar/state_store_ms/min",
		"registrar/state_store_ms/p50", "registrar/state_store_ms/p90", "registrar/state_store_ms/p95",
		"registrar/state_store_ms/p99", "registrar/state_store_ms/p999", "registrar/state_store_ms/p9999"}

	for _, k := range metricNames {
		mesosMetrics[k] = rand.Float64()
	}
}

func TestMain(m *testing.M) {
	generateMetrics()
	r := http.NewServeMux()
	r.HandleFunc("/metrics/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mesosMetrics)
	})
	ts = httptest.NewServer(r)
	rc := m.Run()
	ts.Close()
	os.Exit(rc)
}

func TestMesosMaster(t *testing.T) {
	var acc testutil.Accumulator

	m := Mesos{
		Masters: []string{ts.Listener.Addr().String()},
		Timeout: 10,
	}

	err := m.Gather(&acc)

	if err != nil {
		t.Errorf(err.Error())
	}

	acc.AssertContainsFields(t, "mesos", mesosMetrics)
}

func TestRemoveGroup(t *testing.T) {
	generateMetrics()

	m := Mesos{
		MasterCols: []string{
			"resources", "master", "registrar",
		},
	}
	b := []string{
		"system", "slaves", "frameworks",
		"messages", "evqueue",
	}

	m.removeGroup(&mesosMetrics)

	for _, v := range b {
		for _, x := range masterBlocks(v) {
			if _, ok := mesosMetrics[x]; ok {
				t.Errorf("Found key %s, it should be gone.", x)
			}
		}
	}
	for _, v := range m.MasterCols {
		for _, x := range masterBlocks(v) {
			if _, ok := mesosMetrics[x]; !ok {
				t.Errorf("Didn't find key %s, it should present.", x)
			}
		}
	}
}
