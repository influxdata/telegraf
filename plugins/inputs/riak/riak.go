//go:generate ../../../tools/readme_config_includer/generator
package riak

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Type Riak gathers statistics from one or more Riak instances
type Riak struct {
	// Servers is a slice of servers as http addresses (ex. http://127.0.0.1:8098)
	Servers []string

	client *http.Client
}

// NewRiak return a new instance of Riak with a default http client
func NewRiak() *Riak {
	tr := &http.Transport{ResponseHeaderTimeout: 3 * time.Second}
	client := &http.Client{
		Transport: tr,
		Timeout:   4 * time.Second,
	}
	return &Riak{client: client}
}

// Type riakStats represents the data that is received from Riak
type riakStats struct {
	ClusterAaeFsmActive        int64  `json:"clusteraae_fsm_active"`
	ClusterAaeFsmCreate        int64  `json:"clusteraae_fsm_create"`
	ClusterAaeFsmCreateError   int64  `json:"clusteraae_fsm_create_error"`
	CPUAvg1                    int64  `json:"cpu_avg1"`
	CPUAvg15                   int64  `json:"cpu_avg15"`
	CPUAvg5                    int64  `json:"cpu_avg5"`
	CPUNprocs                  int64  `json:"cpu_nprocs"`
	ConsistentGetObjsize100    int64  `json:"consistent_get_objsize_100"`
	ConsistentGetObjsize95     int64  `json:"consistent_get_objsize_95"`
	ConsistentGetObjsize99     int64  `json:"consistent_get_objsize_99"`
	ConsistentGetObjsizeMean   int64  `json:"consistent_get_objsize_mean"`
	ConsistentGetObjsizeMedian int64  `json:"consistent_get_objsize_median"`
	ConsistentGetTime100       int64  `json:"consistent_get_time_100"`
	ConsistentGetTime95        int64  `json:"consistent_get_time_95"`
	ConsistentGetTime99        int64  `json:"consistent_get_time_99"`
	ConsistentGetTimeMean      int64  `json:"consistent_get_time_mean"`
	ConsistentGetTimeMedian    int64  `json:"consistent_get_time_median"`
	ConsistentGets             int64  `json:"consistent_gets"`
	ConsistentGetsTotal        int64  `json:"consistent_gets_total"`
	ConsistentPutObjsize100    int64  `json:"consistent_put_objsize_100"`
	ConsistentPutObjsize95     int64  `json:"consistent_put_objsize_95"`
	ConsistentPutObjsize99     int64  `json:"consistent_put_objsize_99"`
	ConsistentPutObjsizeMean   int64  `json:"consistent_put_objsize_mean"`
	ConsistentPutObjsizeMedian int64  `json:"consistent_put_objsize_median"`
	ConsistentPutTime100       int64  `json:"consistent_put_time_100"`
	ConsistentPutTime95        int64  `json:"consistent_put_time_95"`
	ConsistentPutTime99        int64  `json:"consistent_put_time_99"`
	ConsistentPutTimeMean      int64  `json:"consistent_put_time_mean"`
	ConsistentPutTimeMedian    int64  `json:"consistent_put_time_median"`
	ConsistentPuts             int64  `json:"consistent_puts"`
	ConsistentPutsTotal        int64  `json:"consistent_puts_total"`
	ConvergeDelayLast          int64  `json:"converge_delay_last"`
	ConvergeDelayMax           int64  `json:"converge_delay_max"`
	ConvergeDelayMean          int64  `json:"converge_delay_mean"`
	ConvergeDelayMin           int64  `json:"converge_delay_min"`
	CoordLocalSoftLoadedTotal  int64  `json:"coord_local_soft_loaded_total"`
	CoordLocalUnloadedTotal    int64  `json:"coord_local_unloaded_total"`
	CoordRedirLeastLoadedTotal int64  `json:"coord_redir_least_loaded_total"`
	CoordRedirLoadedLocalTotal int64  `json:"coord_redir_loaded_local_total"`
	CoordRedirUnloadedTotal    int64  `json:"coord_redir_unloaded_total"`
	CounterActorCounts100      int64  `json:"counter_actor_counts_100"`
	CounterActorCounts95       int64  `json:"counter_actor_counts_95"`
	CounterActorCounts99       int64  `json:"counter_actor_counts_99"`
	CounterActorCountsMean     int64  `json:"counter_actor_counts_mean"`
	CounterActorCountsMedian   int64  `json:"counter_actor_counts_median"`
	DroppedVnodeRequestsTotals int64  `json:"dropped_vnode_requests_totals"`
	ExecutingMappers           int64  `json:"executing_mappers"`
	GossipReceived             int64  `json:"gossip_received"`
	HandoffTimeouts            int64  `json:"handoff_timeouts"`
	HllBytes                   int64  `json:"hll_bytes"`
	HllBytes100                int64  `json:"hll_bytes_100"`
	HllBytes95                 int64  `json:"hll_bytes_95"`
	HllBytes99                 int64  `json:"hll_bytes_99"`
	HllBytesMean               int64  `json:"hll_bytes_mean"`
	HllBytesMedian             int64  `json:"hll_bytes_median"`
	HllBytesTotal              int64  `json:"hll_bytes_total"`
	MemoryCode                 int64  `json:"memory_code"`
	MemoryEts                  int64  `json:"memory_ets"`
	MemoryProcesses            int64  `json:"memory_processes"`
	MemorySystem               int64  `json:"memory_system"`
	MemoryTotal                int64  `json:"memory_total"`
	NodeGetFsmObjsize100       int64  `json:"node_get_fsm_objsize_100"`
	NodeGetFsmObjsize95        int64  `json:"node_get_fsm_objsize_95"`
	NodeGetFsmObjsize99        int64  `json:"node_get_fsm_objsize_99"`
	NodeGetFsmObjsizeMean      int64  `json:"node_get_fsm_objsize_mean"`
	NodeGetFsmObjsizeMedian    int64  `json:"node_get_fsm_objsize_median"`
	NodeGetFsmSiblings100      int64  `json:"node_get_fsm_siblings_100"`
	NodeGetFsmSiblings95       int64  `json:"node_get_fsm_siblings_95"`
	NodeGetFsmSiblings99       int64  `json:"node_get_fsm_siblings_99"`
	NodeGetFsmSiblingsMean     int64  `json:"node_get_fsm_siblings_mean"`
	NodeGetFsmSiblingsMedian   int64  `json:"node_get_fsm_siblings_median"`
	NodeGetFsmTime100          int64  `json:"node_get_fsm_time_100"`
	NodeGetFsmTime95           int64  `json:"node_get_fsm_time_95"`
	NodeGetFsmTime99           int64  `json:"node_get_fsm_time_99"`
	NodeGetFsmTimeMean         int64  `json:"node_get_fsm_time_mean"`
	NodeGetFsmTimeMedian       int64  `json:"node_get_fsm_time_median"`
	NodeGets                   int64  `json:"node_gets"`
	NodeGetsTotal              int64  `json:"node_gets_total"`
	Nodename                   string `json:"nodename"`
	NodePutFsmTime100          int64  `json:"node_put_fsm_time_100"`
	NodePutFsmTime95           int64  `json:"node_put_fsm_time_95"`
	NodePutFsmTime99           int64  `json:"node_put_fsm_time_99"`
	NodePutFsmTimeMean         int64  `json:"node_put_fsm_time_mean"`
	NodePutFsmTimeMedian       int64  `json:"node_put_fsm_time_median"`
	NodePuts                   int64  `json:"node_puts"`
	NodePutsTotal              int64  `json:"node_puts_total"`
	PbcActive                  int64  `json:"pbc_active"`
	PbcConnects                int64  `json:"pbc_connects"`
	PbcConnectsTotal           int64  `json:"pbc_connects_total"`
	VnodeGets                  int64  `json:"vnode_gets"`
	VnodeGetsTotal             int64  `json:"vnode_gets_total"`
	VnodeIndexReads            int64  `json:"vnode_index_reads"`
	VnodeIndexReadsTotal       int64  `json:"vnode_index_reads_total"`
	VnodeIndexWrites           int64  `json:"vnode_index_writes"`
	VnodeIndexWritesTotal      int64  `json:"vnode_index_writes_total"`
	VnodePuts                  int64  `json:"vnode_puts"`
	VnodePutsTotal             int64  `json:"vnode_puts_total"`
	ReadRepairs                int64  `json:"read_repairs"`
	ReadRepairsTotal           int64  `json:"read_repairs_total"`
}

func (*Riak) SampleConfig() string {
	return sampleConfig
}

// Reads stats from all configured servers.
func (r *Riak) Gather(acc telegraf.Accumulator) error {
	// Default to a single server at localhost (default port) if none specified
	if len(r.Servers) == 0 {
		r.Servers = []string{"http://127.0.0.1:8098"}
	}

	// Range over all servers, gathering stats. Returns early in case of any error.
	for _, s := range r.Servers {
		acc.AddError(r.gatherServer(s, acc))
	}

	return nil
}

// Gathers stats from a single server, adding them to the accumulator
func (r *Riak) gatherServer(s string, acc telegraf.Accumulator) error {
	// Parse the given URL to extract the server tag
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("riak unable to parse given server url %s: %s", s, err)
	}

	// Perform the GET request to the riak /stats endpoint
	resp, err := r.client.Get(s + "/stats")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Successful responses will always return status code 200
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("riak responded with unexpected status code %d", resp.StatusCode)
	}

	// Decode the response JSON into a new stats struct
	stats := &riakStats{}
	if err := json.NewDecoder(resp.Body).Decode(stats); err != nil {
		return fmt.Errorf("unable to decode riak response: %s", err)
	}

	// Build a map of tags
	tags := map[string]string{
		"nodename": stats.Nodename,
		"server":   u.Host,
	}

	// Build a map of field values
	fields := map[string]interface{}{
		"cluster_fms_active":             stats.ClusterAaeFsmActive,
		"cluster_fms_create":             stats.ClusterAaeFsmCreate,
		"cluster_fms_create_error":       stats.ClusterAaeFsmCreateError,
		"cpu_avg1":                       stats.CPUAvg1,
		"cpu_avg15":                      stats.CPUAvg15,
		"cpu_avg5":                       stats.CPUAvg5,
		"cpu_nprocs":                     stats.CPUNprocs,
		"consistent_get_objsize_100":     stats.ConsistentGetObjsize100,
		"consistent_get_objsize_95":      stats.ConsistentGetObjsize95,
		"consistent_get_objsize_99":      stats.ConsistentGetObjsize99,
		"consistent_get_objsize_mean":    stats.ConsistentGetObjsizeMean,
		"consistent_get_objsize_median":  stats.ConsistentGetObjsizeMedian,
		"consistent_get_time_100":        stats.ConsistentGetTime100,
		"consistent_get_time_95":         stats.ConsistentGetTime95,
		"consistent_get_time_99":         stats.ConsistentGetTime99,
		"consistent_get_time_mean":       stats.ConsistentGetTimeMean,
		"consistent_get_time_median":     stats.ConsistentGetTimeMedian,
		"consistent_gets":                stats.ConsistentGets,
		"consistent_gets_total":          stats.ConsistentGetsTotal,
		"consistent_put_objsize_100":     stats.ConsistentPutObjsize100,
		"consistent_put_objsize_95":      stats.ConsistentPutObjsize95,
		"consistent_put_objsize_99":      stats.ConsistentPutObjsize99,
		"consistent_put_objsize_mean":    stats.ConsistentPutObjsizeMean,
		"consistent_put_objsize_median":  stats.ConsistentPutObjsizeMedian,
		"consistent_put_time_100":        stats.ConsistentPutTime100,
		"consistent_put_time_95":         stats.ConsistentPutTime95,
		"consistent_put_time_99":         stats.ConsistentPutTime99,
		"consistent_put_time_mean":       stats.ConsistentPutTimeMean,
		"consistent_put_time_median":     stats.ConsistentPutTimeMedian,
		"consistent_puts":                stats.ConsistentPuts,
		"consistent_puts_total":          stats.ConsistentPutsTotal,
		"converge_delay_last":            stats.ConvergeDelayLast,
		"converge_delay_max":             stats.ConvergeDelayMax,
		"converge_delay_mean":            stats.ConvergeDelayMean,
		"converge_delay_min":             stats.ConvergeDelayMin,
		"coord_local_soft_loaded_total":  stats.CoordLocalSoftLoadedTotal,
		"coord_local_unloaded_total":     stats.CoordLocalUnloadedTotal,
		"coord_redir_least_loaded_total": stats.CoordRedirLeastLoadedTotal,
		"coord_redir_loaded_local_total": stats.CoordRedirLoadedLocalTotal,
		"coord_redir_unloaded_total":     stats.CoordRedirUnloadedTotal,
		"counter_actor_counts_100":       stats.CounterActorCounts100,
		"counter_actor_counts_95":        stats.CounterActorCounts95,
		"counter_actor_counts_99":        stats.CounterActorCounts99,
		"counter_actor_counts_mean":      stats.CounterActorCountsMean,
		"counter_actor_counts_median":    stats.CounterActorCountsMedian,
		"dropped_vnode_requests_totals":  stats.DroppedVnodeRequestsTotals,
		"executing_mappers":              stats.ExecutingMappers,
		"gossip_received":                stats.GossipReceived,
		"handoff_timeouts":               stats.HandoffTimeouts,
		"hll_bytes":                      stats.HllBytes,
		"hll_bytes_100":                  stats.HllBytes100,
		"hll_bytes_95":                   stats.HllBytes95,
		"hll_bytes_99":                   stats.HllBytes99,
		"hll_bytes_mean":                 stats.HllBytesMean,
		"hll_bytes_median":               stats.HllBytesMedian,
		"hll_bytes_total":                stats.HllBytesTotal,
		"memory_code":                    stats.MemoryCode,
		"memory_ets":                     stats.MemoryEts,
		"memory_processes":               stats.MemoryProcesses,
		"memory_system":                  stats.MemorySystem,
		"memory_total":                   stats.MemoryTotal,
		"node_get_fsm_objsize_100":       stats.NodeGetFsmObjsize100,
		"node_get_fsm_objsize_95":        stats.NodeGetFsmObjsize95,
		"node_get_fsm_objsize_99":        stats.NodeGetFsmObjsize99,
		"node_get_fsm_objsize_mean":      stats.NodeGetFsmObjsizeMean,
		"node_get_fsm_objsize_median":    stats.NodeGetFsmObjsizeMedian,
		"node_get_fsm_siblings_100":      stats.NodeGetFsmSiblings100,
		"node_get_fsm_siblings_95":       stats.NodeGetFsmSiblings95,
		"node_get_fsm_siblings_99":       stats.NodeGetFsmSiblings99,
		"node_get_fsm_siblings_mean":     stats.NodeGetFsmSiblingsMean,
		"node_get_fsm_siblings_median":   stats.NodeGetFsmSiblingsMedian,
		"node_get_fsm_time_100":          stats.NodeGetFsmTime100,
		"node_get_fsm_time_95":           stats.NodeGetFsmTime95,
		"node_get_fsm_time_99":           stats.NodeGetFsmTime99,
		"node_get_fsm_time_mean":         stats.NodeGetFsmTimeMean,
		"node_get_fsm_time_median":       stats.NodeGetFsmTimeMedian,
		"node_gets":                      stats.NodeGets,
		"node_gets_total":                stats.NodeGetsTotal,
		"node_put_fsm_time_100":          stats.NodePutFsmTime100,
		"node_put_fsm_time_95":           stats.NodePutFsmTime95,
		"node_put_fsm_time_99":           stats.NodePutFsmTime99,
		"node_put_fsm_time_mean":         stats.NodePutFsmTimeMean,
		"node_put_fsm_time_median":       stats.NodePutFsmTimeMedian,
		"node_puts":                      stats.NodePuts,
		"node_puts_total":                stats.NodePutsTotal,
		"pbc_active":                     stats.PbcActive,
		"pbc_connects":                   stats.PbcConnects,
		"pbc_connects_total":             stats.PbcConnectsTotal,
		"vnode_gets":                     stats.VnodeGets,
		"vnode_gets_total":               stats.VnodeGetsTotal,
		"vnode_index_reads":              stats.VnodeIndexReads,
		"vnode_index_reads_total":        stats.VnodeIndexReadsTotal,
		"vnode_index_writes":             stats.VnodeIndexWrites,
		"vnode_index_writes_total":       stats.VnodeIndexWritesTotal,
		"vnode_puts":                     stats.VnodePuts,
		"vnode_puts_total":               stats.VnodePutsTotal,
		"read_repairs":                   stats.ReadRepairs,
		"read_repairs_total":             stats.ReadRepairsTotal,
	}

	// Accumulate the tags and values
	acc.AddFields("riak", fields, tags)

	return nil
}

func init() {
	inputs.Add("riak", func() telegraf.Input {
		return NewRiak()
	})
}
