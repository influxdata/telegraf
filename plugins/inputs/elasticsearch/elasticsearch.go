package elasticsearch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
	"io/ioutil"
	"strings"
)

const statsPath = "/_nodes/stats"
const statsPathLocal = "/_nodes/_local/stats"
const clusterHealthPath = "/_cluster/health"
const clusterStatsPath = "/_cluster/stats"
const catMasterPath = "/_cat/master"

type nodeStat struct {
	Host       string            `json:"host"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
	Indices    interface{}       `json:"indices"`
	OS         interface{}       `json:"os"`
	Process    interface{}       `json:"process"`
	JVM        interface{}       `json:"jvm"`
	ThreadPool interface{}       `json:"thread_pool"`
	FS         interface{}       `json:"fs"`
	Transport  interface{}       `json:"transport"`
	HTTP       interface{}       `json:"http"`
	Breakers   interface{}       `json:"breakers"`
}

type clusterHealth struct {
	ClusterName         string                 `json:"cluster_name"`
	Status              string                 `json:"status"`
	TimedOut            bool                   `json:"timed_out"`
	NumberOfNodes       int                    `json:"number_of_nodes"`
	NumberOfDataNodes   int                    `json:"number_of_data_nodes"`
	ActivePrimaryShards int                    `json:"active_primary_shards"`
	ActiveShards        int                    `json:"active_shards"`
	RelocatingShards    int                    `json:"relocating_shards"`
	InitializingShards  int                    `json:"initializing_shards"`
	UnassignedShards    int                    `json:"unassigned_shards"`
	Indices             map[string]indexHealth `json:"indices"`
}

type indexHealth struct {
	Status              string `json:"status"`
	NumberOfShards      int    `json:"number_of_shards"`
	NumberOfReplicas    int    `json:"number_of_replicas"`
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

type clusterStats struct {
	NodeName    string            `json:"node_name"`
	ClusterName string            `json:"cluster_name"`
	Status      string            `json:"status"`
	Indices     interface{}       `json:"indices"`
	Nodes       interface{}       `json:"nodes"`
}

type catMaster struct {
	NodeID         string     `json:"id"`
	NodeIP         string     `json:"ip"`
	NodeName       string     `json:"node"`
}

const sampleConfig = `
  ## specify a list of one or more Elasticsearch servers
  servers = ["http://localhost:9200"]

  ## Timeout for HTTP requests to the elastic search server(s)
  http_timeout = "5s"

  ## set local to false when you want to read the indices stats from all nodes
  ## within the cluster
  local = true

  ## set cluster_health to true when you want to also obtain cluster health stats
  cluster_health = false

  ## set cluster_stats to true when you want to also obtain cluster stats from
  ## Master nodes. Currently only implemented when local=true
  cluster_stats = false

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

// Elasticsearch is a plugin to read stats from one or many Elasticsearch
// servers.
type Elasticsearch struct {
	Local              bool
	Servers            []string
	HttpTimeout        internal.Duration
	ClusterHealth      bool
	ClusterStats       bool
	SSLCA              string `toml:"ssl_ca"`   // Path to CA file
	SSLCert            string `toml:"ssl_cert"` // Path to host cert file
	SSLKey             string `toml:"ssl_key"`  // Path to cert key file
	InsecureSkipVerify bool   // Use SSL but skip chain & host verification
	client             *http.Client
	catMasterResponse  string
	isMaster           bool
}

// NewElasticsearch return a new instance of Elasticsearch
func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		HttpTimeout: internal.Duration{Duration: time.Second * 5},
	}
}

// SampleConfig returns sample configuration for this plugin.
func (e *Elasticsearch) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (e *Elasticsearch) Description() string {
	return "Read stats from one or more Elasticsearch servers or clusters"
}

// Gather reads the stats from Elasticsearch and writes it to the
// Accumulator.
func (e *Elasticsearch) Gather(acc telegraf.Accumulator) error {
	if e.client == nil {
		client, err := e.createHttpClient()

		if err != nil {
			return err
		}
		e.client = client
	}

	errChan := errchan.New(len(e.Servers))
	var wg sync.WaitGroup
	wg.Add(len(e.Servers))

	for _, serv := range e.Servers {
		go func(s string, acc telegraf.Accumulator) {
			defer wg.Done()
			var url string
			if e.Local {
				url = s + statsPathLocal
			} else {
				url = s + statsPath
			}
			e.isMaster = false

			if e.ClusterStats {
				// get cat/master information here so NodeStats can determine
				// whether this entrance is on the Master
				e.setCatMaster(s+ catMasterPath)
			}

			// Always gather node states
			if err := e.gatherNodeStats(url, acc); err != nil {
				errChan.C <- err
				return
			}

			if e.ClusterHealth {
				url = s + clusterHealthPath + "?level=indices"
				e.gatherClusterHealth(url, acc)
			}

			if e.ClusterStats && e.isMaster {
				e.gatherClusterStats(s + clusterStatsPath, acc)
			}
		}(serv, acc)
	}

	wg.Wait()
	return errChan.Error()
}

func (e *Elasticsearch) createHttpClient() (*http.Client, error) {
	tlsCfg, err := internal.GetTLSConfig(e.SSLCert, e.SSLKey, e.SSLCA, e.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: e.HttpTimeout.Duration,
		TLSClientConfig:       tlsCfg,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   e.HttpTimeout.Duration,
	}

	return client, nil
}

func (e *Elasticsearch) gatherNodeStats(url string, acc telegraf.Accumulator) error {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := e.gatherJsonData(url, nodeStats); err != nil {
		return err
	}

	for id, n := range nodeStats.Nodes {
		tags := map[string]string{
			"node_id":      id,
			"node_host":    n.Host,
			"node_name":    n.Name,
			"cluster_name": nodeStats.ClusterName,
		}

		if (e.ClusterStats && e.Local) {
			// check for master
			tokens := strings.Split(e.catMasterResponse, " ")
			masterNode := tokens[0] // get the node ID and compare it
			e.isMaster = (id == masterNode)
		}

		for k, v := range n.Attributes {
			tags["node_attribute_"+k] = v
		}

		stats := map[string]interface{}{
			"indices":     n.Indices,
			"os":          n.OS,
			"process":     n.Process,
			"jvm":         n.JVM,
			"thread_pool": n.ThreadPool,
			"fs":          n.FS,
			"transport":   n.Transport,
			"http":        n.HTTP,
			"breakers":    n.Breakers,
		}

		now := time.Now()
		for p, s := range stats {
			f := jsonparser.JSONFlattener{}
			// parse Json, ignoring strings and bools
			err := f.FlattenJSON("", s)
			if err != nil {
				return err
			}
			acc.AddFields("elasticsearch_"+p, f.Fields, tags, now)
		}
	}
	return nil
}

func (e *Elasticsearch) gatherClusterHealth(url string, acc telegraf.Accumulator) error {
	healthStats := &clusterHealth{}
	if err := e.gatherJsonData(url, healthStats); err != nil {
		return err
	}
	measurementTime := time.Now()
	clusterFields := map[string]interface{}{
		"status":                healthStats.Status,
		"timed_out":             healthStats.TimedOut,
		"number_of_nodes":       healthStats.NumberOfNodes,
		"number_of_data_nodes":  healthStats.NumberOfDataNodes,
		"active_primary_shards": healthStats.ActivePrimaryShards,
		"active_shards":         healthStats.ActiveShards,
		"relocating_shards":     healthStats.RelocatingShards,
		"initializing_shards":   healthStats.InitializingShards,
		"unassigned_shards":     healthStats.UnassignedShards,
	}
	acc.AddFields(
		"elasticsearch_cluster_health",
		clusterFields,
		map[string]string{"name": healthStats.ClusterName},
		measurementTime,
	)

	for name, health := range healthStats.Indices {
		indexFields := map[string]interface{}{
			"status":                health.Status,
			"number_of_shards":      health.NumberOfShards,
			"number_of_replicas":    health.NumberOfReplicas,
			"active_primary_shards": health.ActivePrimaryShards,
			"active_shards":         health.ActiveShards,
			"relocating_shards":     health.RelocatingShards,
			"initializing_shards":   health.InitializingShards,
			"unassigned_shards":     health.UnassignedShards,
		}
		acc.AddFields(
			"elasticsearch_cluster_health_indices",
			indexFields,
			map[string]string{"index": name},
			measurementTime,
		)
	}
	return nil
}

func (e *Elasticsearch) gatherClusterStats(url string, acc telegraf.Accumulator) error {
	clusterStats := &clusterStats{}
	if err := e.gatherJsonData(url, clusterStats); err != nil {
		return err
	}
	now := time.Now()
	tags := map[string]string{
		"node_name":     clusterStats.NodeName,
		"cluster_name":  clusterStats.ClusterName,
		"status":        clusterStats.Status,
	}

	stats := map[string]interface{}{
		"nodes":     clusterStats.Nodes,
		"indices":   clusterStats.Indices,
	}

	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse json, including bools and strings
		err := f.FullFlattenJSON("", s, true, true)
		if err != nil {
			return err
		}
		acc.AddFields("elasticsearch_clusterstats_"+p, f.Fields, tags, now)
	}

	return nil
}

func (e *Elasticsearch) setCatMaster(url string) error {
	response, err := e.gatherStringData(url)

	if  err != nil {
		return err
	}

	e.catMasterResponse = response
	return nil
}


func (e *Elasticsearch) gatherStringData(url string) (string, error) {
	r, err := e.client.Get(url)
	if err != nil {
		return "elasticsearch: API responded with error", err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return "elasticsearch: API responded with error", fmt.Errorf("status-code %d, expected %d",
			r.StatusCode, http.StatusOK)
	}
	htmlData, err := ioutil.ReadAll(r.Body)
	return string(htmlData), err
}

func (e *Elasticsearch) gatherJsonData(url string, v interface{}) error {
	r, err := e.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return fmt.Errorf("elasticsearch: API responded with status-code %d, expected %d",
			r.StatusCode, http.StatusOK)
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("elasticsearch", func() telegraf.Input {
		return NewElasticsearch()
	})
}
