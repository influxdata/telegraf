package elasticsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

// mask for masking username/password from error messages
var mask = regexp.MustCompile(`https?:\/\/\S+:\S+@`)

// Nodestats are always generated, so simply define a constant for these endpoints
const statsPath = "/_nodes/stats"
const statsPathLocal = "/_nodes/_local/stats"

type nodeStat struct {
	Host       string            `json:"host"`
	Name       string            `json:"name"`
	Roles      []string          `json:"roles"`
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
	ActivePrimaryShards         int                    `json:"active_primary_shards"`
	ActiveShards                int                    `json:"active_shards"`
	ActiveShardsPercentAsNumber float64                `json:"active_shards_percent_as_number"`
	ClusterName                 string                 `json:"cluster_name"`
	DelayedUnassignedShards     int                    `json:"delayed_unassigned_shards"`
	InitializingShards          int                    `json:"initializing_shards"`
	NumberOfDataNodes           int                    `json:"number_of_data_nodes"`
	NumberOfInFlightFetch       int                    `json:"number_of_in_flight_fetch"`
	NumberOfNodes               int                    `json:"number_of_nodes"`
	NumberOfPendingTasks        int                    `json:"number_of_pending_tasks"`
	RelocatingShards            int                    `json:"relocating_shards"`
	Status                      string                 `json:"status"`
	TaskMaxWaitingInQueueMillis int                    `json:"task_max_waiting_in_queue_millis"`
	TimedOut                    bool                   `json:"timed_out"`
	UnassignedShards            int                    `json:"unassigned_shards"`
	Indices                     map[string]indexHealth `json:"indices"`
}

type indexHealth struct {
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	NumberOfReplicas    int    `json:"number_of_replicas"`
	NumberOfShards      int    `json:"number_of_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	Status              string `json:"status"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

type clusterStats struct {
	NodeName    string      `json:"node_name"`
	ClusterName string      `json:"cluster_name"`
	Status      string      `json:"status"`
	Indices     interface{} `json:"indices"`
	Nodes       interface{} `json:"nodes"`
}

type indexStat struct {
	Primaries interface{}              `json:"primaries"`
	Total     interface{}              `json:"total"`
	Shards    map[string][]interface{} `json:"shards"`
}

// Elasticsearch is a plugin to read stats from one or many Elasticsearch
// servers.
type Elasticsearch struct {
	Local                      bool            `toml:"local"`
	Servers                    []string        `toml:"servers"`
	HTTPTimeout                config.Duration `toml:"http_timeout"`
	ClusterHealth              bool            `toml:"cluster_health"`
	ClusterHealthLevel         string          `toml:"cluster_health_level"`
	ClusterStats               bool            `toml:"cluster_stats"`
	ClusterStatsOnlyFromMaster bool            `toml:"cluster_stats_only_from_master"`
	IndicesInclude             []string        `toml:"indices_include"`
	IndicesLevel               string          `toml:"indices_level"`
	NodeStats                  []string        `toml:"node_stats"`
	Username                   string          `toml:"username"`
	Password                   string          `toml:"password"`
	NumMostRecentIndices       int             `toml:"num_most_recent_indices"`

	tls.ClientConfig

	client          *http.Client
	serverInfo      map[string]serverInfo
	serverInfoMutex sync.Mutex
	indexMatchers   map[string]filter.Filter
}
type serverInfo struct {
	nodeID   string
	masterID string
}

func (i serverInfo) isMaster() bool {
	return i.nodeID == i.masterID
}

// NewElasticsearch return a new instance of Elasticsearch
func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		HTTPTimeout:                config.Duration(time.Second * 5),
		ClusterStatsOnlyFromMaster: true,
		ClusterHealthLevel:         "indices",
	}
}

// perform status mapping
func mapHealthStatusToCode(s string) int {
	switch strings.ToLower(s) {
	case "green":
		return 1
	case "yellow":
		return 2
	case "red":
		return 3
	}
	return 0
}

// perform shard status mapping
func mapShardStatusToCode(s string) int {
	switch strings.ToUpper(s) {
	case "UNASSIGNED":
		return 1
	case "INITIALIZING":
		return 2
	case "STARTED":
		return 3
	case "RELOCATING":
		return 4
	}
	return 0
}

// Init the plugin.
func (e *Elasticsearch) Init() error {
	// Compile the configured indexes to match for sorting.
	indexMatchers, err := e.compileIndexMatchers()
	if err != nil {
		return err
	}

	e.indexMatchers = indexMatchers

	return nil
}

// Gather reads the stats from Elasticsearch and writes it to the
// Accumulator.
func (e *Elasticsearch) Gather(acc telegraf.Accumulator) error {
	if e.client == nil {
		client, err := e.createHTTPClient()

		if err != nil {
			return err
		}
		e.client = client
	}

	if e.ClusterStats || len(e.IndicesInclude) > 0 || len(e.IndicesLevel) > 0 {
		var wgC sync.WaitGroup
		wgC.Add(len(e.Servers))

		e.serverInfo = make(map[string]serverInfo)
		for _, serv := range e.Servers {
			go func(s string, acc telegraf.Accumulator) {
				defer wgC.Done()
				info := serverInfo{}

				var err error

				// Gather node ID
				if info.nodeID, err = e.gatherNodeID(s + "/_nodes/_local/name"); err != nil {
					acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				// get cat/master information here so NodeStats can determine
				// whether this node is the Master
				if info.masterID, err = e.getCatMaster(s + "/_cat/master"); err != nil {
					acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				e.serverInfoMutex.Lock()
				e.serverInfo[s] = info
				e.serverInfoMutex.Unlock()
			}(serv, acc)
		}
		wgC.Wait()
	}

	var wg sync.WaitGroup
	wg.Add(len(e.Servers))

	for _, serv := range e.Servers {
		go func(s string, acc telegraf.Accumulator) {
			defer wg.Done()
			url := e.nodeStatsURL(s)

			// Always gather node stats
			if err := e.gatherNodeStats(url, acc); err != nil {
				acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
				return
			}

			if e.ClusterHealth {
				url = s + "/_cluster/health"
				if e.ClusterHealthLevel != "" {
					url = url + "?level=" + e.ClusterHealthLevel
				}
				if err := e.gatherClusterHealth(url, acc); err != nil {
					acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if e.ClusterStats && (e.serverInfo[s].isMaster() || !e.ClusterStatsOnlyFromMaster || !e.Local) {
				if err := e.gatherClusterStats(s+"/_cluster/stats", acc); err != nil {
					acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if len(e.IndicesInclude) > 0 && (e.serverInfo[s].isMaster() || !e.ClusterStatsOnlyFromMaster || !e.Local) {
				if e.IndicesLevel != "shards" {
					if err := e.gatherIndicesStats(s+"/"+strings.Join(e.IndicesInclude, ",")+"/_stats", acc); err != nil {
						acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				} else {
					if err := e.gatherIndicesStats(s+"/"+strings.Join(e.IndicesInclude, ",")+"/_stats?level=shards", acc); err != nil {
						acc.AddError(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				}
			}
		}(serv, acc)
	}

	wg.Wait()
	return nil
}

func (e *Elasticsearch) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := e.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(e.HTTPTimeout),
		TLSClientConfig:       tlsCfg,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(e.HTTPTimeout),
	}

	return client, nil
}

func (e *Elasticsearch) nodeStatsURL(baseURL string) string {
	var url string

	if e.Local {
		url = baseURL + statsPathLocal
	} else {
		url = baseURL + statsPath
	}

	if len(e.NodeStats) == 0 {
		return url
	}

	return fmt.Sprintf("%s/%s", url, strings.Join(e.NodeStats, ","))
}

func (e *Elasticsearch) gatherNodeID(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := e.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	// Only 1 should be returned
	for id := range nodeStats.Nodes {
		return id, nil
	}
	return "", nil
}

func (e *Elasticsearch) gatherNodeStats(url string, acc telegraf.Accumulator) error {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := e.gatherJSONData(url, nodeStats); err != nil {
		return err
	}

	for id, n := range nodeStats.Nodes {
		sort.Strings(n.Roles)
		tags := map[string]string{
			"node_id":      id,
			"node_host":    n.Host,
			"node_name":    n.Name,
			"cluster_name": nodeStats.ClusterName,
			"node_roles":   strings.Join(n.Roles, ","),
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
			// if one of the individual node stats is not even in the
			// original result
			if s == nil {
				continue
			}
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
	if err := e.gatherJSONData(url, healthStats); err != nil {
		return err
	}
	measurementTime := time.Now()
	clusterFields := map[string]interface{}{
		"active_primary_shards":            healthStats.ActivePrimaryShards,
		"active_shards":                    healthStats.ActiveShards,
		"active_shards_percent_as_number":  healthStats.ActiveShardsPercentAsNumber,
		"delayed_unassigned_shards":        healthStats.DelayedUnassignedShards,
		"initializing_shards":              healthStats.InitializingShards,
		"number_of_data_nodes":             healthStats.NumberOfDataNodes,
		"number_of_in_flight_fetch":        healthStats.NumberOfInFlightFetch,
		"number_of_nodes":                  healthStats.NumberOfNodes,
		"number_of_pending_tasks":          healthStats.NumberOfPendingTasks,
		"relocating_shards":                healthStats.RelocatingShards,
		"status":                           healthStats.Status,
		"status_code":                      mapHealthStatusToCode(healthStats.Status),
		"task_max_waiting_in_queue_millis": healthStats.TaskMaxWaitingInQueueMillis,
		"timed_out":                        healthStats.TimedOut,
		"unassigned_shards":                healthStats.UnassignedShards,
	}
	acc.AddFields(
		"elasticsearch_cluster_health",
		clusterFields,
		map[string]string{"name": healthStats.ClusterName},
		measurementTime,
	)

	for name, health := range healthStats.Indices {
		indexFields := map[string]interface{}{
			"active_primary_shards": health.ActivePrimaryShards,
			"active_shards":         health.ActiveShards,
			"initializing_shards":   health.InitializingShards,
			"number_of_replicas":    health.NumberOfReplicas,
			"number_of_shards":      health.NumberOfShards,
			"relocating_shards":     health.RelocatingShards,
			"status":                health.Status,
			"status_code":           mapHealthStatusToCode(health.Status),
			"unassigned_shards":     health.UnassignedShards,
		}
		acc.AddFields(
			"elasticsearch_cluster_health_indices",
			indexFields,
			map[string]string{"index": name, "name": healthStats.ClusterName},
			measurementTime,
		)
	}
	return nil
}

func (e *Elasticsearch) gatherClusterStats(url string, acc telegraf.Accumulator) error {
	clusterStats := &clusterStats{}
	if err := e.gatherJSONData(url, clusterStats); err != nil {
		return err
	}
	now := time.Now()
	tags := map[string]string{
		"node_name":    clusterStats.NodeName,
		"cluster_name": clusterStats.ClusterName,
		"status":       clusterStats.Status,
	}

	stats := map[string]interface{}{
		"nodes":   clusterStats.Nodes,
		"indices": clusterStats.Indices,
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

func (e *Elasticsearch) gatherIndicesStats(url string, acc telegraf.Accumulator) error {
	indicesStats := &struct {
		Shards  map[string]interface{} `json:"_shards"`
		All     map[string]interface{} `json:"_all"`
		Indices map[string]indexStat   `json:"indices"`
	}{}

	if err := e.gatherJSONData(url, indicesStats); err != nil {
		return err
	}
	now := time.Now()

	// Total Shards Stats
	shardsStats := map[string]interface{}{}
	for k, v := range indicesStats.Shards {
		shardsStats[k] = v
	}
	acc.AddFields("elasticsearch_indices_stats_shards_total", shardsStats, map[string]string{}, now)

	// All Stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := jsonparser.JSONFlattener{}
		err := jsonParser.FullFlattenJSON("_", s, true, true)
		if err != nil {
			return err
		}
		acc.AddFields("elasticsearch_indices_stats_"+m, jsonParser.Fields, map[string]string{"index_name": "_all"}, now)
	}

	// Gather stats for each index.
	err := e.gatherIndividualIndicesStats(indicesStats.Indices, now, acc)

	return err
}

// gatherSortedIndicesStats gathers stats for all indices in no particular order.
func (e *Elasticsearch) gatherIndividualIndicesStats(indices map[string]indexStat, now time.Time, acc telegraf.Accumulator) error {
	// Sort indices into buckets based on their configured prefix, if any matches.
	categorizedIndexNames := e.categorizeIndices(indices)
	for _, matchingIndices := range categorizedIndexNames {
		// Establish the number of each category of indices to use. User can configure to use only the latest 'X' amount.
		indicesCount := len(matchingIndices)
		indicesToTrackCount := indicesCount

		// Sort the indices if configured to do so.
		if e.NumMostRecentIndices > 0 {
			if e.NumMostRecentIndices < indicesToTrackCount {
				indicesToTrackCount = e.NumMostRecentIndices
			}
			sort.Strings(matchingIndices)
		}

		// Gather only the number of indexes that have been configured, in descending order (most recent, if date-stamped).
		for i := indicesCount - 1; i >= indicesCount-indicesToTrackCount; i-- {
			indexName := matchingIndices[i]

			err := e.gatherSingleIndexStats(indexName, indices[indexName], now, acc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Elasticsearch) categorizeIndices(indices map[string]indexStat) map[string][]string {
	categorizedIndexNames := map[string][]string{}

	// If all indices are configured to be gathered, bucket them all together.
	if len(e.IndicesInclude) == 0 || e.IndicesInclude[0] == "_all" {
		for indexName := range indices {
			categorizedIndexNames["_all"] = append(categorizedIndexNames["_all"], indexName)
		}

		return categorizedIndexNames
	}

	// Bucket each returned index with its associated configured index (if any match).
	for indexName := range indices {
		match := indexName
		for name, matcher := range e.indexMatchers {
			// If a configured index matches one of the returned indexes, mark it as a match.
			if matcher.Match(match) {
				match = name
				break
			}
		}

		// Bucket all matching indices together for sorting.
		categorizedIndexNames[match] = append(categorizedIndexNames[match], indexName)
	}

	return categorizedIndexNames
}

func (e *Elasticsearch) gatherSingleIndexStats(name string, index indexStat, now time.Time, acc telegraf.Accumulator) error {
	indexTag := map[string]string{"index_name": name}
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
		acc.AddFields("elasticsearch_indices_stats_"+m, f.Fields, indexTag, now)
	}

	if e.IndicesLevel == "shards" {
		for shardNumber, shards := range index.Shards {
			for _, shard := range shards {
				// Get Shard Stats
				flattened := jsonparser.JSONFlattener{}
				err := flattened.FullFlattenJSON("", shard, true, true)
				if err != nil {
					return err
				}

				// determine shard tag and primary/replica designation
				shardType := "replica"
				routingPrimary, _ := flattened.Fields["routing_primary"].(bool)
				if routingPrimary {
					shardType = "primary"
				}
				delete(flattened.Fields, "routing_primary")

				routingState, ok := flattened.Fields["routing_state"].(string)
				if ok {
					flattened.Fields["routing_state"] = mapShardStatusToCode(routingState)
				}

				routingNode, _ := flattened.Fields["routing_node"].(string)
				shardTags := map[string]string{
					"index_name": name,
					"node_id":    routingNode,
					"shard_name": shardNumber,
					"type":       shardType,
				}

				for key, field := range flattened.Fields {
					switch field.(type) {
					case string, bool:
						delete(flattened.Fields, key)
					}
				}

				acc.AddFields("elasticsearch_indices_stats_shards",
					flattened.Fields,
					shardTags,
					now)
			}
		}
	}

	return nil
}

func (e *Elasticsearch) getCatMaster(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if e.Username != "" || e.Password != "" {
		req.SetBasicAuth(e.Username, e.Password)
	}

	r, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return "", fmt.Errorf("elasticsearch: Unable to retrieve master node information. API responded with status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}
	response, err := io.ReadAll(r.Body)

	if err != nil {
		return "", err
	}

	masterID := strings.Split(string(response), " ")[0]

	return masterID, nil
}

func (e *Elasticsearch) gatherJSONData(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if e.Username != "" || e.Password != "" {
		req.SetBasicAuth(e.Username, e.Password)
	}

	r, err := e.client.Do(req)
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

	return json.NewDecoder(r.Body).Decode(v)
}

func (e *Elasticsearch) compileIndexMatchers() (map[string]filter.Filter, error) {
	indexMatchers := map[string]filter.Filter{}
	var err error

	// Compile each configured index into a glob matcher.
	for _, configuredIndex := range e.IndicesInclude {
		if _, exists := indexMatchers[configuredIndex]; !exists {
			indexMatchers[configuredIndex], err = filter.Compile([]string{configuredIndex})
			if err != nil {
				return nil, err
			}
		}
	}

	return indexMatchers, nil
}

func init() {
	inputs.Add("elasticsearch", func() telegraf.Input {
		return NewElasticsearch()
	})
}
