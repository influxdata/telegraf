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

// Define a constant for the endpoints that are always hit. Optional endpoints are
// defined where they are performed.
const statsPath = "/_nodes/stats"
const statsPathLocal = "/_nodes/_local/stats"

const sampleConfig = `
  ## specify a list of one or more Elasticsearch servers
  ## you can add username and password to your url to use basic authentication:
  ## servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

  ## Timeout for HTTP requests to the elastic search server(s)
  http_timeout = "5s"

  ## When local is true (the default), the node will read only its own stats.
  ## Set local to false when you want to read the node stats from all nodes
  ## of the cluster.
  local = true

  ## Set cluster_health to true when you want to also obtain cluster health stats
  cluster_health = false

  ## Set cluster_stats to true when you want to obtain cluster stats from the Master node.
  cluster_stats = false

  ## Set indices_stats to true when you want to obtain indices stats from the Master node.
  indices_stats = false

  ## Set indices_shards_stats to true when you want to obtain shard stats from the Master node.
  ## If set, then indices_stats is considered true as they are also provided with shard stats.
  indices_shards_stats = false

  ## Multiplier of the elasticsearch interval to be used to reduce the frequency of
  ## indices_stats and indices_shards_stats reports. The default interval is 10 seconds, and this
  ## multiplier is defaulted to 6 to cause these metrics to be taken once per minute by default.
  indices_interval_multiplier = 6

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
	Local                     bool
	Servers                   []string
	HttpTimeout               internal.Duration
	ClusterHealth             bool
	ClusterStats              bool
	IndicesStats              bool
	ShardsStats               bool
	IndicesIntervalMultiplier int16
	SSLCA                     string `toml:"ssl_ca"`   // Path to CA file
	SSLCert                   string `toml:"ssl_cert"` // Path to host cert file
	SSLKey                    string `toml:"ssl_key"`  // Path to cert key file
	InsecureSkipVerify        bool   // Use SSL but skip chain & host verification
	client                    *http.Client
	masterNodeId              string
	localNodeIsMaster         bool
	indicesCallCntr           int16
}

// NewElasticsearch return a new instance of Elasticsearch
func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		Local: true,
		IndicesIntervalMultiplier: 6,
		indicesCallCntr:           0,
		masterNodeId:              "",
		HttpTimeout:               internal.Duration{Duration: time.Second * 5},
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

	needCatMaster := e.ClusterStats || e.IndicesStats || e.ShardsStats

	for _, serv := range e.Servers {
		go func(s string, acc telegraf.Accumulator) {
			defer wg.Done()

			// Gather cat master when
			e.localNodeIsMaster = false
			if needCatMaster {
				if err := e.gatherCatMaster(s+"/_cat/master", acc); err != nil {
					errChan.C <- err
					return
				}
			}

			// Always gather node stats
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

			// Optional stats

			if e.ClusterHealth {
				e.gatherClusterHealth(s+"/_cluster/health?level=indices", acc)
			}

			if e.ClusterStats && e.localNodeIsMaster {
				e.gatherClusterStats(s+"/_cluster/stats", acc)
			}

			if e.IndicesStats && !e.ShardsStats && e.localNodeIsMaster {
				e.gatherIndicesStats(s+"_all/_stats", acc, false)
			}

			if e.ShardsStats && e.localNodeIsMaster {
				e.gatherIndicesStats(s+"_all/_stats?level=shards", acc, true)
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

		// check for master
		e.localNodeIsMaster = (id == e.masterNodeId)

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

	healthStats := &clusterHealth{}
	if err := e.gatherJsonData(url, healthStats); err != nil {
		return err
	}
	measurementTime := time.Now()
	clusterFields := map[string]interface{}{
		"cluster_name":          healthStats.ClusterName,
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
			"index_name":            name,
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
			map[string]string{"name": name},
			measurementTime,
		)
	}
	return nil
}

func (e *Elasticsearch) gatherClusterStats(url string, acc telegraf.Accumulator) error {
	type ClusterStats struct {
		ClusterName string      `json:"cluster_name"`
		Status      string      `json:"status"`
		Indices     interface{} `json:"indices"`
		Nodes       interface{} `json:"nodes"`
	}

	clusterStats := &ClusterStats{}
	if err := e.gatherJsonData(url, clusterStats); err != nil {
		return err
	}
	now := time.Now()

	clusterFields := map[string]interface{}{
		"cluster_name": clusterStats.ClusterName,
		"status":       clusterStats.Status,
	}

	acc.AddFields(
		"elasticsearch_clusterstats",
		clusterFields,
		map[string]string{"name": ""},
		now,
	)

	stats := map[string]interface{}{
		"nodes":   clusterStats.Nodes,
		"indices": clusterStats.Indices,
	}

	for name, stat := range stats {
		f := jsonparser.JSONFlattener{}
		// parse json, including bools and strings
		err := f.FullFlattenJSON("", stat, true, true)
		if err != nil {
			return err
		}
		acc.AddFields("elasticsearch_clusterstats_"+name, f.Fields, map[string]string{"name": ""}, now)
	}

	return nil
}

func (e *Elasticsearch) gatherIndicesStats(url string, acc telegraf.Accumulator, doShards bool) error {
	// only gather the indices stats at the correct interval
	e.indicesCallCntr = e.indicesCallCntr + 1
	if e.indicesCallCntr >= e.IndicesIntervalMultiplier {
		e.indicesCallCntr = 0
	} else {
		return nil
	}

	type IndexStat struct {
		Primaries interface{} `json:"primaries"`
		Total     interface{} `json:"total"`
		Shards    interface{} `json:"shards"`
	}

	indicesStats := &struct {
		ShardTotals map[string]interface{} `json:"_shards"`
		All         map[string]interface{} `json:"_all"`
		Indices     map[string]*IndexStat  `json:"indices"`
	}{}

	if err := e.gatherJsonData(url, indicesStats); err != nil {
		return err
	}
	now := time.Now()

	// pull out total shard stats
	_shardsStats := map[string]interface{}{}
	for k, v := range indicesStats.ShardTotals {
		_shardsStats[k] = v
	}
	acc.AddFields("elasticsearch_indicesstats_shards", _shardsStats, map[string]string{"name": ""}, now)

	// pull out all/total stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := jsonparser.JSONFlattener{}
		err := jsonParser.FullFlattenJSON("_", s, true, true)
		if err != nil {
			return err
		}
		acc.AddFields("elasticsearch_indicesstats_"+m, jsonParser.Fields, map[string]string{"index_name": "all"}, now)
	}

	// pull out each indices stats
	for id, index := range indicesStats.Indices {
		indexTag := map[string]string{"index_name": id}
		stats := map[string]interface{}{
			"primaries": index.Primaries,
			"total":     index.Total,
		}
		for m, s := range stats {
			f := jsonparser.JSONFlattener{}
			// parse Json, getting strings and bools
			err := f.FullFlattenJSON("", s, true, true)
			if err != nil {
				return err
			}
			acc.AddFields("elasticsearch_indicesstats_"+m, f.Fields, indexTag, now)
		}

		if doShards {
			// pull out all shard stats
			flattened := jsonparser.JSONFlattener{}
			err1 := flattened.FullFlattenJSON("", index.Shards, true, true)
			if err1 != nil {
				return err1
			}
			// Split the shards to tag them by their index number, and put primary/replica into
			// different measurements to make it easier to distinguish on Grafana
			//
			// Loop thru the flattened map. The map is unsorted/random per Go, so assumptions
			// about the ordering cannot be made.
			//
			// For each entry, parser the two leading integers to determine the shard name and
			// which shard type it is (primary/replica).
			//
			// Then add to primary/replica maps accordingly. Note that there is a map of these,
			// based on shard index. And a separate measurement is used for each shard, allowing
			// Grafana to use wild characters on the measurement name (and get all shards)
			primary := map[string]jsonparser.JSONFlattener{}
			replica := map[string]jsonparser.JSONFlattener{}
			primary = make(map[string]jsonparser.JSONFlattener)
			replica = make(map[string]jsonparser.JSONFlattener)

			for field, value := range flattened.Fields {
				// determine shard tag and primary/replica designation
				splitFieldName := strings.SplitAfterN(field, "_", 3)
				fieldName := splitFieldName[2]
				isPrimary := flattened.Fields[splitFieldName[0]+splitFieldName[1]+"routing_primary"]
				shardNumber := strings.TrimSuffix(splitFieldName[0], "_")
				// add to the appropriate measurement
				if isPrimary == true {
					jp, ok := primary[shardNumber]
					if !ok {
						jp = jsonparser.JSONFlattener{}
						jp.Fields = make(map[string]interface{})
						primary[shardNumber] = jp
					}
					jp.Fields[fieldName] = value
				} else {
					jp, ok := replica[shardNumber]
					if !ok {
						jp = jsonparser.JSONFlattener{}
						jp.Fields = make(map[string]interface{})
						replica[shardNumber] = jp
					}
					jp.Fields[fieldName] = value
				}
			}
			for shard := range primary {
				shardTags := map[string]string{
					"index_name": id,
					"shard_name": shard,
				}
				acc.AddFields("elasticsearch_indicesstats_shards_primary",
					primary[shard].Fields,
					shardTags,
					now)
			}
			for shard := range replica {
				shardTags := map[string]string{
					"index_name": id,
					"shard_name": shard,
				}
				acc.AddFields("elasticsearch_indicesstats_shards_replica",
					replica[shard].Fields,
					shardTags,
					now)
			}
		}

	}

	return nil
}

func (e *Elasticsearch) gatherCatMaster(url string, acc telegraf.Accumulator) error {
	r, err := e.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return fmt.Errorf("status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}
	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return err
	}

	// split the string to get the tokens.
	tokens := strings.Split(string(response), " ")

	if len(tokens) < 4 {
		return err
	}

	// the first token is the NodeId
	e.masterNodeId = tokens[0] // get the node ID and compare it

	stats := map[string]interface{}{
		"master_node_id":   tokens[0],
		"master_host":      tokens[1],
		"master_node_name": tokens[3],
	}

	now := time.Now()

	catMasterStats := map[string]interface{}{}
	for k, v := range stats {
		catMasterStats[k] = v
	}
	acc.AddFields("elasticsearch_catmaster", catMasterStats, map[string]string{"name": ""}, now)

	return nil
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
