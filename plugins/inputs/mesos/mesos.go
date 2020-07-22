package mesos

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type Role string

const (
	MASTER Role = "main"
	SLAVE       = "subordinate"
)

type Mesos struct {
	Timeout    int
	Mains    []string
	MainCols []string `toml:"main_collections"`
	Subordinates     []string
	SubordinateCols  []string `toml:"subordinate_collections"`
	tls.ClientConfig

	Log telegraf.Logger

	initialized bool
	client      *http.Client
	mainURLs  []*url.URL
	subordinateURLs   []*url.URL
}

var allMetrics = map[Role][]string{
	MASTER: {"resources", "main", "system", "agents", "frameworks", "framework_offers", "tasks", "messages", "evqueue", "registrar", "allocator"},
	SLAVE:  {"resources", "agent", "system", "executors", "tasks", "messages"},
}

var sampleConfig = `
  ## Timeout, in ms.
  timeout = 100

  ## A list of Mesos mains.
  mains = ["http://localhost:5050"]

  ## Main metrics groups to be collected, by default, all enabled.
  main_collections = [
    "resources",
    "main",
    "system",
    "agents",
    "frameworks",
    "framework_offers",
    "tasks",
    "messages",
    "evqueue",
    "registrar",
    "allocator",
  ]

  ## A list of Mesos subordinates, default is []
  # subordinates = []

  ## Subordinate metrics groups to be collected, by default, all enabled.
  # subordinate_collections = [
  #   "resources",
  #   "agent",
  #   "system",
  #   "executors",
  #   "tasks",
  #   "messages",
  # ]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

// SampleConfig returns a sample configuration block
func (m *Mesos) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Mesos plugin
func (m *Mesos) Description() string {
	return "Telegraf plugin for gathering metrics from N Mesos mains"
}

func parseURL(s string, role Role) (*url.URL, error) {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		host, port, err := net.SplitHostPort(s)
		// no port specified
		if err != nil {
			host = s
			switch role {
			case MASTER:
				port = "5050"
			case SLAVE:
				port = "5051"
			}
		}

		s = "http://" + host + ":" + port
		log.Printf("W! [inputs.mesos] using %q as connection URL; please update your configuration to use an URL", s)
	}

	return url.Parse(s)
}

func (m *Mesos) initialize() error {
	if len(m.MainCols) == 0 {
		m.MainCols = allMetrics[MASTER]
	}

	if len(m.SubordinateCols) == 0 {
		m.SubordinateCols = allMetrics[SLAVE]
	}

	if m.Timeout == 0 {
		m.Log.Info("Missing timeout value, setting default value (100ms)")
		m.Timeout = 100
	}

	rawQuery := "timeout=" + strconv.Itoa(m.Timeout) + "ms"

	m.mainURLs = make([]*url.URL, 0, len(m.Mains))
	for _, main := range m.Mains {
		u, err := parseURL(main, MASTER)
		if err != nil {
			return err
		}

		u.RawQuery = rawQuery
		m.mainURLs = append(m.mainURLs, u)
	}

	m.subordinateURLs = make([]*url.URL, 0, len(m.Subordinates))
	for _, subordinate := range m.Subordinates {
		u, err := parseURL(subordinate, SLAVE)
		if err != nil {
			return err
		}

		u.RawQuery = rawQuery
		m.subordinateURLs = append(m.subordinateURLs, u)
	}

	client, err := m.createHttpClient()
	if err != nil {
		return err
	}
	m.client = client

	return nil
}

// Gather() metrics from given list of Mesos Mains
func (m *Mesos) Gather(acc telegraf.Accumulator) error {
	if !m.initialized {
		err := m.initialize()
		if err != nil {
			return err
		}
		m.initialized = true
	}

	var wg sync.WaitGroup

	for _, main := range m.mainURLs {
		wg.Add(1)
		go func(main *url.URL) {
			acc.AddError(m.gatherMainMetrics(main, MASTER, acc))
			wg.Done()
			return
		}(main)
	}

	for _, subordinate := range m.subordinateURLs {
		wg.Add(1)
		go func(subordinate *url.URL) {
			acc.AddError(m.gatherMainMetrics(subordinate, SLAVE, acc))
			wg.Done()
			return
		}(subordinate)
	}

	wg.Wait()

	return nil
}

func (m *Mesos) createHttpClient() (*http.Client, error) {
	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsCfg,
		},
		Timeout: 4 * time.Second,
	}

	return client, nil
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

// mainBlocks serves as kind of metrics registry grouping them in sets
func getMetrics(role Role, group string) []string {
	var m map[string][]string

	m = make(map[string][]string)

	if role == MASTER {
		m["resources"] = []string{
			"main/cpus_percent",
			"main/cpus_used",
			"main/cpus_total",
			"main/cpus_revocable_percent",
			"main/cpus_revocable_total",
			"main/cpus_revocable_used",
			"main/disk_percent",
			"main/disk_used",
			"main/disk_total",
			"main/disk_revocable_percent",
			"main/disk_revocable_total",
			"main/disk_revocable_used",
			"main/gpus_percent",
			"main/gpus_used",
			"main/gpus_total",
			"main/gpus_revocable_percent",
			"main/gpus_revocable_total",
			"main/gpus_revocable_used",
			"main/mem_percent",
			"main/mem_used",
			"main/mem_total",
			"main/mem_revocable_percent",
			"main/mem_revocable_total",
			"main/mem_revocable_used",
		}

		m["main"] = []string{
			"main/elected",
			"main/uptime_secs",
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
			"main/subordinate_registrations",
			"main/subordinate_removals",
			"main/subordinate_reregistrations",
			"main/subordinate_shutdowns_scheduled",
			"main/subordinate_shutdowns_canceled",
			"main/subordinate_shutdowns_completed",
			"main/subordinates_active",
			"main/subordinates_connected",
			"main/subordinates_disconnected",
			"main/subordinates_inactive",
			"main/subordinate_unreachable_canceled",
			"main/subordinate_unreachable_completed",
			"main/subordinate_unreachable_scheduled",
			"main/subordinates_unreachable",
		}

		m["frameworks"] = []string{
			"main/frameworks_active",
			"main/frameworks_connected",
			"main/frameworks_disconnected",
			"main/frameworks_inactive",
			"main/outstanding_offers",
		}

		// framework_offers and allocator metrics have unpredictable names, so they can't be listed here.
		// These empty groups are included to prevent the "unknown metrics group" info log below.
		// filterMetrics() filters these metrics by looking for names with the corresponding prefix.
		m["framework_offers"] = []string{}
		m["allocator"] = []string{}

		m["tasks"] = []string{
			"main/tasks_error",
			"main/tasks_failed",
			"main/tasks_finished",
			"main/tasks_killed",
			"main/tasks_lost",
			"main/tasks_running",
			"main/tasks_staging",
			"main/tasks_starting",
			"main/tasks_dropped",
			"main/tasks_gone",
			"main/tasks_gone_by_operator",
			"main/tasks_killing",
			"main/tasks_unreachable",
		}

		m["messages"] = []string{
			"main/invalid_executor_to_framework_messages",
			"main/invalid_framework_to_executor_messages",
			"main/invalid_status_update_acknowledgements",
			"main/invalid_status_updates",
			"main/dropped_messages",
			"main/messages_authenticate",
			"main/messages_deactivate_framework",
			"main/messages_decline_offers",
			"main/messages_executor_to_framework",
			"main/messages_exited_executor",
			"main/messages_framework_to_executor",
			"main/messages_kill_task",
			"main/messages_launch_tasks",
			"main/messages_reconcile_tasks",
			"main/messages_register_framework",
			"main/messages_register_subordinate",
			"main/messages_reregister_framework",
			"main/messages_reregister_subordinate",
			"main/messages_resource_request",
			"main/messages_revive_offers",
			"main/messages_status_update",
			"main/messages_status_update_acknowledgement",
			"main/messages_unregister_framework",
			"main/messages_unregister_subordinate",
			"main/messages_update_subordinate",
			"main/recovery_subordinate_removals",
			"main/subordinate_removals/reason_registered",
			"main/subordinate_removals/reason_unhealthy",
			"main/subordinate_removals/reason_unregistered",
			"main/valid_framework_to_executor_messages",
			"main/valid_status_update_acknowledgements",
			"main/valid_status_updates",
			"main/task_lost/source_main/reason_invalid_offers",
			"main/task_lost/source_main/reason_subordinate_removed",
			"main/task_lost/source_subordinate/reason_executor_terminated",
			"main/valid_executor_to_framework_messages",
			"main/invalid_operation_status_update_acknowledgements",
			"main/messages_operation_status_update_acknowledgement",
			"main/messages_reconcile_operations",
			"main/messages_suppress_offers",
			"main/valid_operation_status_update_acknowledgements",
		}

		m["evqueue"] = []string{
			"main/event_queue_dispatches",
			"main/event_queue_http_requests",
			"main/event_queue_messages",
			"main/operator_event_stream_subscribers",
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
			"registrar/log/ensemble_size",
			"registrar/log/recovered",
			"registrar/queued_operations",
			"registrar/registry_size_bytes",
			"registrar/state_store_ms/count",
		}
	} else if role == SLAVE {
		m["resources"] = []string{
			"subordinate/cpus_percent",
			"subordinate/cpus_used",
			"subordinate/cpus_total",
			"subordinate/cpus_revocable_percent",
			"subordinate/cpus_revocable_total",
			"subordinate/cpus_revocable_used",
			"subordinate/disk_percent",
			"subordinate/disk_used",
			"subordinate/disk_total",
			"subordinate/disk_revocable_percent",
			"subordinate/disk_revocable_total",
			"subordinate/disk_revocable_used",
			"subordinate/gpus_percent",
			"subordinate/gpus_used",
			"subordinate/gpus_total",
			"subordinate/gpus_revocable_percent",
			"subordinate/gpus_revocable_total",
			"subordinate/gpus_revocable_used",
			"subordinate/mem_percent",
			"subordinate/mem_used",
			"subordinate/mem_total",
			"subordinate/mem_revocable_percent",
			"subordinate/mem_revocable_total",
			"subordinate/mem_revocable_used",
		}

		m["agent"] = []string{
			"subordinate/registered",
			"subordinate/uptime_secs",
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
			"subordinate/container_launch_errors",
			"subordinate/executors_preempted",
			"subordinate/frameworks_active",
			"subordinate/executor_directory_max_allowed_age_secs",
			"subordinate/executors_registering",
			"subordinate/executors_running",
			"subordinate/executors_terminated",
			"subordinate/executors_terminating",
			"subordinate/recovery_errors",
		}

		m["tasks"] = []string{
			"subordinate/tasks_failed",
			"subordinate/tasks_finished",
			"subordinate/tasks_killed",
			"subordinate/tasks_lost",
			"subordinate/tasks_running",
			"subordinate/tasks_staging",
			"subordinate/tasks_starting",
		}

		m["messages"] = []string{
			"subordinate/invalid_framework_messages",
			"subordinate/invalid_status_updates",
			"subordinate/valid_framework_messages",
			"subordinate/valid_status_updates",
		}
	}

	ret, ok := m[group]

	if !ok {
		log.Printf("I! [inputs.mesos] unknown role %q metrics group: %s", role, group)
		return []string{}
	}

	return ret
}

func (m *Mesos) filterMetrics(role Role, metrics *map[string]interface{}) {
	var ok bool
	var selectedMetrics []string

	if role == MASTER {
		selectedMetrics = m.MainCols
	} else if role == SLAVE {
		selectedMetrics = m.SubordinateCols
	}

	for _, k := range metricsDiff(role, selectedMetrics) {
		switch k {
		// allocator and framework_offers metrics have unpredictable names, so we have to identify them by name prefix.
		case "allocator":
			for m := range *metrics {
				if strings.HasPrefix(m, "allocator/") {
					delete((*metrics), m)
				}
			}
		case "framework_offers":
			for m := range *metrics {
				if strings.HasPrefix(m, "main/frameworks/") || strings.HasPrefix(m, "frameworks/") {
					delete((*metrics), m)
				}
			}

		// All other metrics have predictable names. We can use getMetrics() to retrieve them.
		default:
			for _, v := range getMetrics(role, k) {
				if _, ok = (*metrics)[v]; ok {
					delete((*metrics), v)
				}
			}
		}
	}
}

// TaskStats struct for JSON API output /monitor/statistics
type TaskStats struct {
	ExecutorID  string                 `json:"executor_id"`
	FrameworkID string                 `json:"framework_id"`
	Statistics  map[string]interface{} `json:"statistics"`
}

func (m *Mesos) gatherSubordinateTaskMetrics(u *url.URL, acc telegraf.Accumulator) error {
	var metrics []TaskStats

	tags := map[string]string{
		"server": u.Hostname(),
		"url":    urlTag(u),
	}

	resp, err := m.client.Get(withPath(u, "/monitor/statistics").String())

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

func withPath(u *url.URL, path string) *url.URL {
	c := *u
	c.Path = path
	return &c
}

func urlTag(u *url.URL) string {
	c := *u
	c.Path = ""
	c.User = nil
	c.RawQuery = ""
	return c.String()
}

// This should not belong to the object
func (m *Mesos) gatherMainMetrics(u *url.URL, role Role, acc telegraf.Accumulator) error {
	var jsonOut map[string]interface{}

	tags := map[string]string{
		"server": u.Hostname(),
		"url":    urlTag(u),
		"role":   string(role),
	}

	resp, err := m.client.Get(withPath(u, "/metrics/snapshot").String())

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
		if jf.Fields["main/elected"] != 0.0 {
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
