package elasticsearch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const statsPath = "/_nodes/stats"
const statsPathLocal = "/_nodes/_local/stats"
const healthPath = "/_cluster/health"

type node struct {
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

const sampleConfig = `
  ## specify a list of one or more Elasticsearch servers
  servers = ["http://localhost:9200"]

  ## set local to false when you want to read the indices stats from all nodes
  ## within the cluster
  local = true

  ## set cluster_health to true when you want to also obtain cluster level stats
  cluster_health = false
`

// Elasticsearch is a plugin to read stats from one or many Elasticsearch
// servers.
type Elasticsearch struct {
	Local         bool
	Servers       []string
	ClusterHealth bool
	client        *http.Client
}

// NewElasticsearch return a new instance of Elasticsearch
func NewElasticsearch() *Elasticsearch {
	tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(4 * time.Second),
	}
	return &Elasticsearch{client: client}
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
			if err := e.gatherNodeStats(url, acc); err != nil {
				errChan.C <- err
				return
			}
			if e.ClusterHealth {
				e.gatherClusterStats(fmt.Sprintf("%s/_cluster/health?level=indices", s), acc)
			}
		}(serv, acc)
	}

	wg.Wait()
	return errChan.Error()
}

func (e *Elasticsearch) gatherNodeStats(url string, acc telegraf.Accumulator) error {
	nodeStats := &struct {
		ClusterName string           `json:"cluster_name"`
		Nodes       map[string]*node `json:"nodes"`
	}{}
	if err := e.gatherData(url, nodeStats); err != nil {
		return err
	}
	for id, n := range nodeStats.Nodes {
		tags := map[string]string{
			"node_id":      id,
			"node_host":    n.Host,
			"node_name":    n.Name,
			"cluster_name": nodeStats.ClusterName,
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
			err := f.FlattenJSON("", s)
			if err != nil {
				return err
			}
			acc.AddFields("elasticsearch_"+p, f.Fields, tags, now)
		}
	}
	return nil
}

func (e *Elasticsearch) gatherClusterStats(url string, acc telegraf.Accumulator) error {
	clusterStats := &clusterHealth{}
	if err := e.gatherData(url, clusterStats); err != nil {
		return err
	}
	measurementTime := time.Now()
	clusterFields := map[string]interface{}{
		"status":                clusterStats.Status,
		"timed_out":             clusterStats.TimedOut,
		"number_of_nodes":       clusterStats.NumberOfNodes,
		"number_of_data_nodes":  clusterStats.NumberOfDataNodes,
		"active_primary_shards": clusterStats.ActivePrimaryShards,
		"active_shards":         clusterStats.ActiveShards,
		"relocating_shards":     clusterStats.RelocatingShards,
		"initializing_shards":   clusterStats.InitializingShards,
		"unassigned_shards":     clusterStats.UnassignedShards,
	}
	acc.AddFields(
		"elasticsearch_cluster_health",
		clusterFields,
		map[string]string{"name": clusterStats.ClusterName},
		measurementTime,
	)

	for name, health := range clusterStats.Indices {
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
			"elasticsearch_indices",
			indexFields,
			map[string]string{"index": name},
			measurementTime,
		)
	}
	return nil
}

func (e *Elasticsearch) gatherData(url string, v interface{}) error {
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
