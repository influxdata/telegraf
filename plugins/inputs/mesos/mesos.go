package mesos

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type Mesos struct {
	Timeout    int
	Masters    []string
	MasterCols []string `toml:"master_collections"`
}

var defaultMetrics = []string{
	"resources", "master", "system", "slaves", "frameworks",
	"tasks", "messages", "evqueue", "messages", "registrar",
}

var sampleConfig = `
  # Timeout, in ms.
  timeout = 100
  # A list of Mesos masters, default value is localhost:5050.
  masters = ["localhost:5050"]
  # Metrics groups to be collected, by default, all enabled.
  master_collections = ["resources","master","system","slaves","frameworks","messages","evqueue","registrar"]
`

// SampleConfig returns a sample configuration block
func (m *Mesos) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Mesos plugin
func (m *Mesos) Description() string {
	return "Telegraf plugin for gathering metrics from N Mesos masters"
}

// Gather() metrics from given list of Mesos Masters
func (m *Mesos) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	var errorChannel chan error

	if len(m.Masters) == 0 {
		m.Masters = []string{"localhost:5050"}
	}

	errorChannel = make(chan error, len(m.Masters)*2)

	for _, v := range m.Masters {
		wg.Add(1)
		go func(c string) {
			errorChannel <- m.gatherMetrics(c, acc)
			wg.Done()
			return
		}(v)
	}

	wg.Wait()
	close(errorChannel)
	errorStrings := []string{}

	// Gather all errors for returning them at once
	for err := range errorChannel {
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) > 0 {
		return errors.New(strings.Join(errorStrings, "\n"))
	}
	return nil
}

// metricsDiff() returns set names for removal
func metricsDiff(w []string) []string {
	b := []string{}
	s := make(map[string]bool)

	if len(w) == 0 {
		return b
	}

	for _, v := range w {
		s[v] = true
	}

	for _, d := range defaultMetrics {
		if _, ok := s[d]; !ok {
			b = append(b, d)
		}
	}

	return b
}

// masterBlocks serves as kind of metrics registry groupping them in sets
func masterBlocks(g string) []string {
	var m map[string][]string

	m = make(map[string][]string)

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

	m["slaves"] = []string{
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

	ret, ok := m[g]

	if !ok {
		log.Println("[mesos] Unkown metrics group: ", g)
		return []string{}
	}

	return ret
}

// removeGroup(), remove unwanted sets
func (m *Mesos) removeGroup(j *map[string]interface{}) {
	var ok bool

	b := metricsDiff(m.MasterCols)

	for _, k := range b {
		for _, v := range masterBlocks(k) {
			if _, ok = (*j)[v]; ok {
				delete((*j), v)
			}
		}
	}
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

// This should not belong to the object
func (m *Mesos) gatherMetrics(a string, acc telegraf.Accumulator) error {
	var jsonOut map[string]interface{}

	host, _, err := net.SplitHostPort(a)
	if err != nil {
		host = a
		a = a + ":5050"
	}

	tags := map[string]string{
		"server": host,
	}

	if m.Timeout == 0 {
		log.Println("[mesos] Missing timeout value, setting default value (100ms)")
		m.Timeout = 100
	}

	ts := strconv.Itoa(m.Timeout) + "ms"

	resp, err := client.Get("http://" + a + "/metrics/snapshot?timeout=" + ts)

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

	m.removeGroup(&jsonOut)

	jf := jsonparser.JSONFlattener{}

	err = jf.FlattenJSON("", jsonOut)

	if err != nil {
		return err
	}

	acc.AddFields("mesos", jf.Fields, tags)

	return nil
}

func init() {
	inputs.Add("mesos", func() telegraf.Input {
		return &Mesos{}
	})
}
