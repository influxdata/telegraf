package mesos

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type Role string

const (
	MASTER Role = "master"
	SLAVE       = "slave"
)

type Mesos struct {
	Timeout    int
	Masters    []string
	MasterCols []string `toml:"master_collections"`
	Slaves     []string
	SlaveCols  []string `toml:"slave_collections"`
	//SlaveTasks bool
	SSL      bool `toml:"ssl"`
	Insecure bool `toml:"insecure"`
}

var allMetrics = map[Role][]string{
	MASTER: []string{"resources", "master", "system", "agents", "frameworks", "tasks", "messages", "evqueue", "registrar"},
	SLAVE:  []string{"resources", "agent", "system", "executors", "tasks", "messages"},
}

var sampleConfig = `
  ## Timeout, in ms.
  timeout = 100
  ## A list of Mesos masters.
  masters = ["localhost:5050"]
  ## Master metrics groups to be collected, by default, all enabled.
  master_collections = [
    "resources",
    "master",
    "system",
    "agents",
    "frameworks",
    "tasks",
    "messages",
    "evqueue",
    "registrar",
  ]
  ## A list of Mesos slaves, default is []
  # slaves = []
  ## Slave metrics groups to be collected, by default, all enabled.
  # slave_collections = [
  #   "resources",
  #   "agent",
  #   "system",
  #   "executors",
  #   "tasks",
  #   "messages",
  # ]
  # enable SSL
  # ssl = false
`

// SampleConfig returns a sample configuration block
func (m *Mesos) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Mesos plugin
func (m *Mesos) Description() string {
	return "Telegraf plugin for gathering metrics from N Mesos masters"
}

func (m *Mesos) SetDefaults() {
	if len(m.MasterCols) == 0 {
		m.MasterCols = allMetrics[MASTER]
	}

	if len(m.SlaveCols) == 0 {
		m.SlaveCols = allMetrics[SLAVE]
	}

	if m.Timeout == 0 {
		log.Println("I! [mesos] Missing timeout value, setting default value (100ms)")
		m.Timeout = 100
	}
}

// Gather() metrics from given list of Mesos Masters
func (m *Mesos) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	m.SetDefaults()

	for _, v := range m.Masters {
		wg.Add(1)
		go func(c string) {
			acc.AddError(m.gatherMainMetrics(c, ":5050", MASTER, acc))
			wg.Done()
			return
		}(v)
	}

	for _, v := range m.Slaves {
		wg.Add(1)
		go func(c string) {
			acc.AddError(m.gatherMainMetrics(c, ":5051", SLAVE, acc))
			wg.Done()
			return
		}(v)

		// if !m.SlaveTasks {
		// 	continue
		// }

		// wg.Add(1)
		// go func(c string) {
		// 	acc.AddError(m.gatherSlaveTaskMetrics(c, ":5051", acc))
		// 	wg.Done()
		// 	return
		// }(v)
	}

	wg.Wait()

	return nil
}

// metricsDiff() returns set names for removal
func metricsDiff(role Role, w []string) []string {
	b := []string{}
	s := make(map[string]bool)

	if len(w) == 0 {
		return b
	}

	for _, v := range w {
		s[v] = true
	}

	for _, d := range allMetrics[role] {
		if _, ok := s[d]; !ok {
			b = append(b, d)
		}
	}

	return b
}

// masterBlocks serves as kind of metrics registry groupping them in sets
func getMetrics(role Role, group string) []string {
	var m map[string][]string

	m = make(map[string][]string)

	if role == MASTER {
		m["resources"] = []string{
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
		}

		m["master"] = []string{
			"master/elected",
			"master/uptime_secs",
		}

		m["system"] = []string{
			"system/cpus_total",
			"system/load_15min",
			"system/load_5min",
			"system/load_1min",
			"system/mem_free_bytes",
			"system/mem_total_bytes",
		}

		m["agents"] = []string{
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
		}

		m["frameworks"] = []string{
			"master/frameworks_active",
			"master/frameworks_connected",
			"master/frameworks_disconnected",
			"master/frameworks_inactive",
			"master/outstanding_offers",
		}

		m["tasks"] = []string{
			"master/tasks_error",
			"master/tasks_failed",
			"master/tasks_finished",
			"master/tasks_killed",
			"master/tasks_lost",
			"master/tasks_running",
			"master/tasks_staging",
			"master/tasks_starting",
		}

		m["messages"] = []string{
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
		}

		m["evqueue"] = []string{
			"master/event_queue_dispatches",
			"master/event_queue_http_requests",
			"master/event_queue_messages",
		}

		m["registrar"] = []string{
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
		}
	} else if role == SLAVE {
		m["resources"] = []string{
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
		}

		m["agent"] = []string{
			"slave/registered",
			"slave/uptime_secs",
		}

		m["system"] = []string{
			"system/cpus_total",
			"system/load_15min",
			"system/load_5min",
			"system/load_1min",
			"system/mem_free_bytes",
			"system/mem_total_bytes",
		}

		m["executors"] = []string{
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
		}

		m["tasks"] = []string{
			"slave/tasks_failed",
			"slave/tasks_finished",
			"slave/tasks_killed",
			"slave/tasks_lost",
			"slave/tasks_running",
			"slave/tasks_staging",
			"slave/tasks_starting",
		}

		m["messages"] = []string{
			"slave/invalid_framework_messages",
			"slave/invalid_status_updates",
			"slave/valid_framework_messages",
			"slave/valid_status_updates",
		}
	}

	ret, ok := m[group]

	if !ok {
		log.Printf("I! [mesos] Unkown %s metrics group: %s\n", role, group)
		return []string{}
	}

	return ret
}

func (m *Mesos) filterMetrics(role Role, metrics *map[string]interface{}) {
	var ok bool
	var selectedMetrics []string

	if role == MASTER {
		selectedMetrics = m.MasterCols
	} else if role == SLAVE {
		selectedMetrics = m.SlaveCols
	}

	for _, k := range metricsDiff(role, selectedMetrics) {
		for _, v := range getMetrics(role, k) {
			if _, ok = (*metrics)[v]; ok {
				delete((*metrics), v)
			}
		}
	}
}

var client = &http.Client{
	Timeout: time.Duration(4 * time.Second),
}

// TaskStats struct for JSON API output /monitor/statistics
type TaskStats struct {
	ExecutorID  string                 `json:"executor_id"`
	FrameworkID string                 `json:"framework_id"`
	Statistics  map[string]interface{} `json:"statistics"`
}

func (m *Mesos) getHTTPPrefix() string {
	if m.SSL {
		return "https://"
	}
	return "http://"
}

func (m *Mesos) gatherSlaveTaskMetrics(address string, defaultPort string, acc telegraf.Accumulator) error {
	var metrics []TaskStats

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		address = address + defaultPort
	}

	tags := map[string]string{
		"server": host,
	}

	ts := strconv.Itoa(m.Timeout) + "ms"

	// client is global so need to set the insecure flag here if necessary
	client.Transport = &http.Transport{
		ResponseHeaderTimeout: time.Duration(3 * time.Second),
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: m.Insecure},
	}
	resp, err := client.Get(m.getHTTPPrefix() + address + "/monitor/statistics?timeout=" + ts)

	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(data), &metrics); err != nil {
		return errors.New("Error decoding JSON response")
	}

	for _, task := range metrics {
		tags["framework_id"] = task.FrameworkID

		jf := jsonparser.JSONFlattener{}
		err = jf.FlattenJSON("", task.Statistics)

		if err != nil {
			return err
		}

		timestamp := time.Unix(int64(jf.Fields["timestamp"].(float64)), 0)
		jf.Fields["executor_id"] = task.ExecutorID

		acc.AddFields("mesos_tasks", jf.Fields, tags, timestamp)
	}

	return nil
}

// This should not belong to the object
func (m *Mesos) gatherMainMetrics(a string, defaultPort string, role Role, acc telegraf.Accumulator) error {
	var jsonOut map[string]interface{}

	host, _, err := net.SplitHostPort(a)
	if err != nil {
		host = a
		a = a + defaultPort
	}

	tags := map[string]string{
		"server": host,
		"role":   string(role),
	}

	ts := strconv.Itoa(m.Timeout) + "ms"

	client.Transport = &http.Transport{
		ResponseHeaderTimeout: time.Duration(3 * time.Second),
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: m.Insecure},
	}
	resp, err := client.Get(m.getHTTPPrefix() + a + "/metrics/snapshot?timeout=" + ts)

	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(data), &jsonOut); err != nil {
		return errors.New("Error decoding JSON response")
	}

	m.filterMetrics(role, &jsonOut)

	jf := jsonparser.JSONFlattener{}

	err = jf.FlattenJSON("", jsonOut)

	if err != nil {
		return err
	}

	if role == MASTER {
		if jf.Fields["master/elected"] != 0.0 {
			tags["state"] = "leader"
		} else {
			tags["state"] = "standby"
		}
	}

	acc.AddFields("mesos", jf.Fields, tags)

	return nil
}

func init() {
	inputs.Add("mesos", func() telegraf.Input {
		return &Mesos{}
	})
}
