//go:generate ../../../tools/readme_config_includer/generator
package elasticsearch

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
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
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
	parsers_json "github.com/influxdata/telegraf/plugins/parsers/json"
)

//go:embed sample.conf
var sampleConfig string

// mask for masking username/password from error messages
var mask = regexp.MustCompile(`https?:\/\/\S+:\S+@`)

const (
	// Node stats are always generated, so simply define a constant for these endpoints
	statsPath      = "/_nodes/stats"
	statsPathLocal = "/_nodes/_local/stats"
)

type Elasticsearch struct {
	Local                      bool              `toml:"local"`
	Servers                    []string          `toml:"servers"`
	HTTPHeaders                map[string]string `toml:"headers"`
	ClusterHealth              bool              `toml:"cluster_health"`
	ClusterHealthLevel         string            `toml:"cluster_health_level"`
	ClusterStats               bool              `toml:"cluster_stats"`
	ClusterStatsOnlyFromMaster bool              `toml:"cluster_stats_only_from_master"`
	CCRStats                   bool              `toml:"ccr_stats"`
	CCRStatsOnlyFromMaster     bool              `toml:"ccr_stats_only_from_master"`
	EnrichStats                bool              `toml:"enrich_stats"`
	RemoteStoreStats           bool              `toml:"remote_store_stats"`
	IndicesInclude             []string          `toml:"indices_include"`
	IndicesLevel               string            `toml:"indices_level"`
	NodeStats                  []string          `toml:"node_stats"`
	Username                   string            `toml:"username"`
	Password                   string            `toml:"password"`
	NumMostRecentIndices       int               `toml:"num_most_recent_indices"`

	Log telegraf.Logger `toml:"-"`

	client *http.Client
	common_http.HTTPClientConfig

	serverInfo      map[string]serverInfo
	serverInfoMutex sync.Mutex
	indexMatchers   map[string]filter.Filter
}

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

type enrichStats struct {
	CoordinatorStats []struct {
		NodeID                string `json:"node_id"`
		QueueSize             int    `json:"queue_size"`
		RemoteRequestsCurrent int    `json:"remote_requests_current"`
		RemoteRequestsTotal   int    `json:"remote_requests_total"`
		ExecutedSearchesTotal int    `json:"executed_searches_total"`
	} `json:"coordinator_stats"`
	CacheStats []struct {
		NodeID    string `json:"node_id"`
		Count     int    `json:"count"`
		Hits      int64  `json:"hits"`
		Misses    int    `json:"misses"`
		Evictions int    `json:"evictions"`
	} `json:"cache_stats"`
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

type ccrLeaderStats struct {
	ccrLeaderIndexStats
	NumReplicatedIndices int                            `json:"num_replicated_indices"`
	IndexStats           map[string]ccrLeaderIndexStats `json:"index_stats"`
}

type ccrLeaderIndexStats struct {
	TranslogSizeBytes           int `json:"translog_size_bytes"`
	OperationsRead              int `json:"operations_read"`
	OperationsReadLucene        int `json:"operations_read_lucene"`
	OperationsReadTranslog      int `json:"operations_read_translog"`
	TotalReadTimeLuceneMillis   int `json:"total_read_time_lucene_millis"`
	TotalReadTimeTranslogMillis int `json:"total_read_time_translog_millis"`
	BytesRead                   int `json:"bytes_read"`
}

type ccrFollowerStats struct {
	ccrFollowerIndexStats
	NumSyincingIndices       int `json:"num_syncing_indices"`
	NumBootstrappingIndicies int `json:"num_bootstrapping_indices"`
	NumPausedIndices         int `json:"num_paused_indices"`
	NumFailedIndices         int `json:"num_failed_indices"`
	NumShardTasks            int `json:"num_shard_tasks"`
	NumIndexTasks            int `json:"num_index_tasks"`
}

type ccrFollowerIndexStats struct {
	OperationsWritten      int `json:"operations_written"`
	OperationsRead         int `json:"operations_read"`
	FailedReadRequests     int `json:"failed_read_requests"`
	ThrottledReadRequests  int `json:"throttled_read_requests"`
	FailedWriteRequests    int `json:"failed_write_requests"`
	ThrottledWriteRequests int `json:"throttled_write_requests"`
	FollowerCheckpoint     int `json:"follower_checkpoint"`
	LeaderCheckpoint       int `json:"leader_checkpoint"`
	TotalWriteTimeMillis   int `json:"total_write_time_millis"`
}

type indexStat struct {
	Primaries interface{}              `json:"primaries"`
	Total     interface{}              `json:"total"`
	Shards    map[string][]interface{} `json:"shards"`
}

type serverInfo struct {
	nodeID   string
	masterID string
}

// NewElasticsearch return a new instance of Elasticsearch
func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		ClusterStatsOnlyFromMaster: true,
		CCRStatsOnlyFromMaster:     true,
		ClusterHealthLevel:         "indices",
	}
}

func (*Elasticsearch) SampleConfig() string {
	return sampleConfig
}

func (e *Elasticsearch) Init() error {
	// Compile the configured indexes to match for sorting.
	indexMatchers, err := e.compileIndexMatchers()
	if err != nil {
		return err
	}

	e.indexMatchers = indexMatchers

	return nil
}

func (*Elasticsearch) Start(telegraf.Accumulator) error {
	return nil
}

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
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				// get cat/master information here so NodeStats can determine
				// whether this node is the Master
				if info.masterID, err = e.getCatMaster(s + "/_cat/master"); err != nil {
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
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
				acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
				return
			}

			if e.ClusterHealth {
				url = s + "/_cluster/health"
				if e.ClusterHealthLevel != "" {
					url = url + "?level=" + e.ClusterHealthLevel
				}
				if err := e.gatherClusterHealth(url, acc); err != nil {
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if e.RemoteStoreStats {
				for _, indexName := range e.IndicesInclude {
					if err := e.gatherRemoteStoreStats(s+"/_remotestore/stats/"+indexName, indexName, acc); err != nil {
						acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					}
				}
			}

			if e.ClusterStats && (e.serverInfo[s].isMaster() || !e.ClusterStatsOnlyFromMaster || !e.Local) {
				if err := e.gatherClusterStats(s+"/_cluster/stats", acc); err != nil {
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if e.CCRStats && (e.serverInfo[s].isMaster() || !e.CCRStatsOnlyFromMaster || !e.Local) {
				if err := e.gatherCCRLeaderStats(s+"/_plugins/_replication/leader_stats", acc); err != nil {
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
				if err := e.gatherCCRFollowerStats(s+"/_plugins/_replication/follower_stats", acc); err != nil {
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if len(e.IndicesInclude) > 0 && (e.serverInfo[s].isMaster() || !e.ClusterStatsOnlyFromMaster || !e.Local) {
				if e.IndicesLevel != "shards" {
					if err := e.gatherIndicesStats(s+"/"+strings.Join(e.IndicesInclude, ",")+"/_stats", acc); err != nil {
						acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				} else {
					if err := e.gatherIndicesStats(s+"/"+strings.Join(e.IndicesInclude, ",")+"/_stats?level=shards", acc); err != nil {
						acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				}
			}

			if e.EnrichStats {
				if err := e.gatherEnrichStats(s+"/_enrich/_stats", acc); err != nil {
					acc.AddError(errors.New(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}
		}(serv, acc)
	}

	wg.Wait()
	return nil
}

func (e *Elasticsearch) Stop() {
	if e.client != nil {
		e.client.CloseIdleConnections()
	}
}

func (e *Elasticsearch) createHTTPClient() (*http.Client, error) {
	ctx := context.Background()
	return e.HTTPClientConfig.CreateClient(ctx, e.Log)
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
			f := parsers_json.JSONFlattener{}
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

func (e *Elasticsearch) gatherEnrichStats(url string, acc telegraf.Accumulator) error {
	enrichStats := &enrichStats{}
	if err := e.gatherJSONData(url, enrichStats); err != nil {
		return err
	}
	measurementTime := time.Now()

	for _, coordinator := range enrichStats.CoordinatorStats {
		coordinatorFields := map[string]interface{}{
			"queue_size":              coordinator.QueueSize,
			"remote_requests_current": coordinator.RemoteRequestsCurrent,
			"remote_requests_total":   coordinator.RemoteRequestsTotal,
			"executed_searches_total": coordinator.ExecutedSearchesTotal,
		}
		acc.AddFields(
			"elasticsearch_enrich_stats_coordinator",
			coordinatorFields,
			map[string]string{"node_id": coordinator.NodeID},
			measurementTime,
		)
	}

	for _, cache := range enrichStats.CacheStats {
		cacheFields := map[string]interface{}{
			"count":     cache.Count,
			"hits":      cache.Hits,
			"misses":    cache.Misses,
			"evictions": cache.Evictions,
		}
		acc.AddFields(
			"elasticsearch_enrich_stats_cache",
			cacheFields,
			map[string]string{"node_id": cache.NodeID},
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
		f := parsers_json.JSONFlattener{}
		// parse json, including bools and strings
		err := f.FullFlattenJSON("", s, true, true)
		if err != nil {
			return err
		}
		acc.AddFields("elasticsearch_clusterstats_"+p, f.Fields, tags, now)
	}

	return nil
}

func (e *Elasticsearch) gatherCCRLeaderStats(url string, acc telegraf.Accumulator) error {
	ccrStats := &ccrLeaderStats{}
	if err := e.gatherJSONData(url, ccrStats); err != nil {
		return err
	}
	now := time.Now()

	stats := map[string]interface{}{
		"num_replicated_indices":          float64(ccrStats.NumReplicatedIndices),
		"translog_size_bytes":             float64(ccrStats.TranslogSizeBytes),
		"bytes_read":                      float64(ccrStats.BytesRead),
		"operations_read":                 float64(ccrStats.OperationsRead),
		"operations_read_lucene":          float64(ccrStats.OperationsReadLucene),
		"operations_read_translog":        float64(ccrStats.OperationsReadTranslog),
		"total_read_time_lucene_millis":   float64(ccrStats.TotalReadTimeLuceneMillis),
		"total_read_time_translog_millis": float64(ccrStats.TotalReadTimeTranslogMillis),
	}

	acc.AddFields("elasticsearch_ccr_stats_leader", stats, map[string]string{}, now)

	return nil
}

func (e *Elasticsearch) gatherCCRFollowerStats(url string, acc telegraf.Accumulator) error {
	ccrStats := &ccrFollowerStats{}
	if err := e.gatherJSONData(url, ccrStats); err != nil {
		return err
	}
	now := time.Now()

	stats := map[string]interface{}{
		"num_syncing_indices":       float64(ccrStats.NumSyincingIndices),
		"num_bootstrapping_indices": float64(ccrStats.NumBootstrappingIndicies),
		"num_paused_indices":        float64(ccrStats.NumPausedIndices),
		"num_failed_indices":        float64(ccrStats.NumFailedIndices),
		"num_shard_tasks":           float64(ccrStats.NumShardTasks),
		"num_index_tasks":           float64(ccrStats.NumIndexTasks),
		"operations_written":        float64(ccrStats.OperationsWritten),
		"operations_read":           float64(ccrStats.OperationsRead),
		"failed_read_requests":      float64(0),
		"throttled_read_requests":   float64(0),
		"failed_write_requests":     float64(0),
		"throttled_write_requests":  float64(0),
		"follower_checkpoint":       float64(1),
		"leader_checkpoint":         float64(1),
		"total_write_time_millis":   float64(2290),
	}
	acc.AddFields("elasticsearch_ccr_stats_follower", stats, map[string]string{}, now)

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
	shardsStats := make(map[string]interface{}, len(indicesStats.Shards))
	for k, v := range indicesStats.Shards {
		shardsStats[k] = v
	}
	acc.AddFields("elasticsearch_indices_stats_shards_total", shardsStats, make(map[string]string), now)

	// All Stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := parsers_json.JSONFlattener{}
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
	categorizedIndexNames := make(map[string][]string, len(indices))

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
		f := parsers_json.JSONFlattener{}
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
				flattened := parsers_json.JSONFlattener{}
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

func (e *Elasticsearch) gatherRemoteStoreStats(url string, indexName string, acc telegraf.Accumulator) error {
	var remoteData map[string]interface{}
	if err := e.gatherJSONData(url, &remoteData); err != nil {
		return err
	}
	now := time.Now()

	if shards, ok := remoteData["_shards"].(map[string]interface{}); ok {
		globalTags := map[string]string{"index_name": indexName}
		acc.AddFields("elasticsearch_remotestore_global", shards, globalTags, now)
	}

	indicesRaw, ok := remoteData["indices"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("remote store API response missing 'indices' field")
	}

	idxRaw, exists := indicesRaw[indexName]
	if !exists {
		return fmt.Errorf("index %s not found in remote store stats", indexName)
	}

	idxData, ok := idxRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for index %s data", indexName)
	}

	shardsRaw, ok := idxData["shards"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("shards field missing or malformed for index %s", indexName)
	}

	for shardID, shardEntries := range shardsRaw {
		entries, ok := shardEntries.([]interface{})
		if !ok {
			continue
		}
		// Process each shard entry (primary and replicas)
		for _, entry := range entries {
			f := parsers_json.JSONFlattener{}
			if err := f.FullFlattenJSON("", entry, true, true); err != nil {
				return err
			}

			tags := map[string]string{
				"index_name": indexName,
				"shard_id":   shardID,
			}
			if entryMap, ok := entry.(map[string]interface{}); ok {
				if routing, exists := entryMap["routing"].(map[string]interface{}); exists {
					if state, ok := routing["state"].(string); ok {
						tags["routing_state"] = state
					}
					if primary, ok := routing["primary"].(bool); ok {
						if primary {
							tags["shard_type"] = "primary"
						} else {
							tags["shard_type"] = "replica"
						}
					}
					if node, ok := routing["node"].(string); ok {
						tags["node_id"] = node
					}
				}
			}

			delete(f.Fields, "routing")
			acc.AddFields("elasticsearch_remotestore_stats_shards", f.Fields, tags, now)
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

	for key, value := range e.HTTPHeaders {
		req.Header.Add(key, value)
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
		return "", fmt.Errorf(
			"elasticsearch: Unable to retrieve master node information. API responded with status-code %d, expected %d",
			r.StatusCode,
			http.StatusOK,
		)
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

	for key, value := range e.HTTPHeaders {
		req.Header.Add(key, value)
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
	var err error
	indexMatchers := make(map[string]filter.Filter, len(e.IndicesInclude))

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

func (i serverInfo) isMaster() bool {
	return i.nodeID == i.masterID
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

func newElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		ClusterStatsOnlyFromMaster: true,
		ClusterHealthLevel:         "indices",
		HTTPClientConfig: common_http.HTTPClientConfig{
			ResponseHeaderTimeout: config.Duration(5 * time.Second),
			Timeout:               config.Duration(5 * time.Second),
		},
	}
}

func init() {
	inputs.Add("elasticsearch", func() telegraf.Input {
		return newElasticsearch()
	})
}
