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

// Type riakStats represents the data that is received from Riak.
//
// For a complete list of all metrics supported by Riak, see the "[riak_kv_stat]"
// and "[riak_kv_stat_bc]" modules.
//
// [riak_kv_stat]: https://github.com/basho/riak_kv/blob/develop-3.0/src/riak_kv_stat.erl#L179
// [riak_kv_stat_bc]: https://github.com/basho/riak_kv/blob/develop-3.0/src/riak_kv_stat_bc.erl
type riakStats struct {
	ClusterAaeFsmActive               int64    `json:"clusteraae_fsm_active"`
	ClusterAaeFsmCreate               int64    `json:"clusteraae_fsm_create"`
	ClusterAaeFsmCreateError          int64    `json:"clusteraae_fsm_create_error"`
	ConnectedNodes                    []string `json:"connected_nodes"`
	ConsistentGetObjsize100           int64    `json:"consistent_get_objsize_100"`
	ConsistentGetObjsize95            int64    `json:"consistent_get_objsize_95"`
	ConsistentGetObjsize99            int64    `json:"consistent_get_objsize_99"`
	ConsistentGetObjsizeMean          int64    `json:"consistent_get_objsize_mean"`
	ConsistentGetObjsizeMedian        int64    `json:"consistent_get_objsize_median"`
	ConsistentGetTime100              int64    `json:"consistent_get_time_100"`
	ConsistentGetTime95               int64    `json:"consistent_get_time_95"`
	ConsistentGetTime99               int64    `json:"consistent_get_time_99"`
	ConsistentGetTimeMean             int64    `json:"consistent_get_time_mean"`
	ConsistentGetTimeMedian           int64    `json:"consistent_get_time_median"`
	ConsistentGets                    int64    `json:"consistent_gets"`
	ConsistentGetsTotal               int64    `json:"consistent_gets_total"`
	ConsistentPutObjsize100           int64    `json:"consistent_put_objsize_100"`
	ConsistentPutObjsize95            int64    `json:"consistent_put_objsize_95"`
	ConsistentPutObjsize99            int64    `json:"consistent_put_objsize_99"`
	ConsistentPutObjsizeMean          int64    `json:"consistent_put_objsize_mean"`
	ConsistentPutObjsizeMedian        int64    `json:"consistent_put_objsize_median"`
	ConsistentPutTime100              int64    `json:"consistent_put_time_100"`
	ConsistentPutTime95               int64    `json:"consistent_put_time_95"`
	ConsistentPutTime99               int64    `json:"consistent_put_time_99"`
	ConsistentPutTimeMean             int64    `json:"consistent_put_time_mean"`
	ConsistentPutTimeMedian           int64    `json:"consistent_put_time_median"`
	ConsistentPuts                    int64    `json:"consistent_puts"`
	ConsistentPutsTotal               int64    `json:"consistent_puts_total"`
	ConvergeDelayLast                 int64    `json:"converge_delay_last"`
	ConvergeDelayMax                  int64    `json:"converge_delay_max"`
	ConvergeDelayMean                 int64    `json:"converge_delay_mean"`
	ConvergeDelayMin                  int64    `json:"converge_delay_min"`
	CoordLocalSoftLoadedTotal         int64    `json:"coord_local_soft_loaded_total"`
	CoordLocalUnloadedTotal           int64    `json:"coord_local_unloaded_total"`
	CoordRedirLeastLoadedTotal        int64    `json:"coord_redir_least_loaded_total"`
	CoordRedirLoadedLocalTotal        int64    `json:"coord_redir_loaded_local_total"`
	CoordRedirUnloadedTotal           int64    `json:"coord_redir_unloaded_total"`
	CoordRedirsTotal                  int64    `json:"coord_redirs_total"`
	CounterActorCounts100             int64    `json:"counter_actor_counts_100"`
	CounterActorCounts95              int64    `json:"counter_actor_counts_95"`
	CounterActorCounts99              int64    `json:"counter_actor_counts_99"`
	CounterActorCountsMean            int64    `json:"counter_actor_counts_mean"`
	CounterActorCountsMedian          int64    `json:"counter_actor_counts_median"`
	CPUAvg1                           int64    `json:"cpu_avg1"`
	CPUAvg15                          int64    `json:"cpu_avg15"`
	CPUAvg5                           int64    `json:"cpu_avg5"`
	CPUNprocs                         int64    `json:"cpu_nprocs"`
	DroppedVnodeRequestsTotals        int64    `json:"dropped_vnode_requests_totals"`
	ExecutingMappers                  int64    `json:"executing_mappers"`
	GossipReceived                    int64    `json:"gossip_received"`
	HandoffTimeouts                   int64    `json:"handoff_timeouts"`
	HllBytes                          int64    `json:"hll_bytes"`
	HllBytes100                       int64    `json:"hll_bytes_100"`
	HllBytes95                        int64    `json:"hll_bytes_95"`
	HllBytes99                        int64    `json:"hll_bytes_99"`
	HllBytesMean                      int64    `json:"hll_bytes_mean"`
	HllBytesMedian                    int64    `json:"hll_bytes_median"`
	HllBytesTotal                     int64    `json:"hll_bytes_total"`
	IgnoredGossipTotal                int64    `json:"ignored_gossip_total"`
	IndexFsmActive                    int64    `json:"index_fsm_active"`
	IndexFsmComplete                  int64    `json:"index_fsm_complete"`
	IndexFsmCreate                    int64    `json:"index_fsm_create"`
	IndexFsmCreateError               int64    `json:"index_fsm_create_error"`
	IndexFsmResults100                int64    `json:"index_fsm_results_100"`
	IndexFsmResults95                 int64    `json:"index_fsm_results_95"`
	IndexFsmResults99                 int64    `json:"index_fsm_results_99"`
	IndexFsmResultsMean               int64    `json:"index_fsm_results_mean"`
	IndexFsmResultsMedian             int64    `json:"index_fsm_results_median"`
	IndexFsmTime100                   int64    `json:"index_fsm_time_100"`
	IndexFsmTime95                    int64    `json:"index_fsm_time_95"`
	IndexFsmTime99                    int64    `json:"index_fsm_time_99"`
	IndexFsmTimeMean                  int64    `json:"index_fsm_time_mean"`
	IndexFsmTimeMedian                int64    `json:"index_fsm_time_median"`
	LatePutFsmCoordinatorAck          int64    `json:"late_put_fsm_coordinator_ack"`
	LeveldbReadBlockError             int64    `json:"leveldb_read_block_error"`
	ListFsmActive                     int64    `json:"list_fsm_active"`
	ListFsmCreate                     int64    `json:"list_fsm_create"`
	ListFsmCreateError                int64    `json:"list_fsm_create_error"`
	ListFsmCreateErrorTotal           int64    `json:"list_fsm_create_error_total"`
	ListFsmCreateTotal                int64    `json:"list_fsm_create_total"`
	MapActorCounts100                 int64    `json:"map_actor_counts_100"`
	MapActorCounts95                  int64    `json:"map_actor_counts_95"`
	MapActorCounts99                  int64    `json:"map_actor_counts_99"`
	MapActorCountsMean                int64    `json:"map_actor_counts_mean"`
	MapActorCountsMedian              int64    `json:"map_actor_counts_median"`
	MemAllocated                      int64    `json:"mem_allocated"`
	MemTotal                          int64    `json:"mem_total"`
	MemoryAtom                        int64    `json:"memory_atom"`
	MemoryAtomUsed                    int64    `json:"memory_atom_used"`
	MemoryBinary                      int64    `json:"memory_binary"`
	MemoryCode                        int64    `json:"memory_code"`
	MemoryEts                         int64    `json:"memory_ets"`
	MemoryProcesses                   int64    `json:"memory_processes"`
	MemoryProcessesUsed               int64    `json:"memory_processes_used"`
	MemorySystem                      int64    `json:"memory_system"`
	MemoryTotal                       int64    `json:"memory_total"`
	NgrfetchNofetch                   int64    `json:"ngrfetch_nofetch"`
	NgrfetchNofetchTotal              int64    `json:"ngrfetch_nofetch_total"`
	NgrfetchPrefetch                  int64    `json:"ngrfetch_prefetch"`
	NgrfetchPrefetchTotal             int64    `json:"ngrfetch_prefetch_total"`
	NgrfetchTofetch                   int64    `json:"ngrfetch_tofetch"`
	NgrfetchTofetchTotal              int64    `json:"ngrfetch_tofetch_total"`
	NgrreplEmpty                      int64    `json:"ngrrepl_empty"`
	NgrreplEmptyTotal                 int64    `json:"ngrrepl_empty_total"`
	NgrreplError                      int64    `json:"ngrrepl_error"`
	NgrreplErrorTotal                 int64    `json:"ngrrepl_error_total"`
	NgrreplObject                     int64    `json:"ngrrepl_object"`
	NgrreplObjectTotal                int64    `json:"ngrrepl_object_total"`
	NgrreplSrcdiscard                 int64    `json:"ngrrepl_srcdiscard"`
	NgrreplSrcdiscardTotal            int64    `json:"ngrrepl_srcdiscard_total"`
	NodeGetFsmActive                  int64    `json:"node_get_fsm_active"`
	NodeGetFsmActive60s               int64    `json:"node_get_fsm_active_60s"`
	NodeGetFsmCounterObjsize100       int64    `json:"node_get_fsm_counter_objsize_100"`
	NodeGetFsmCounterObjsize95        int64    `json:"node_get_fsm_counter_objsize_95"`
	NodeGetFsmCounterObjsize99        int64    `json:"node_get_fsm_counter_objsize_99"`
	NodeGetFsmCounterObjsizeMean      int64    `json:"node_get_fsm_counter_objsize_mean"`
	NodeGetFsmCounterObjsizeMedian    int64    `json:"node_get_fsm_counter_objsize_median"`
	NodeGetFsmCounterSiblings100      int64    `json:"node_get_fsm_counter_siblings_100"`
	NodeGetFsmCounterSiblings95       int64    `json:"node_get_fsm_counter_siblings_95"`
	NodeGetFsmCounterSiblings99       int64    `json:"node_get_fsm_counter_siblings_99"`
	NodeGetFsmCounterSiblingsMean     int64    `json:"node_get_fsm_counter_siblings_mean"`
	NodeGetFsmCounterSiblingsMedian   int64    `json:"node_get_fsm_counter_siblings_median"`
	NodeGetFsmCounterTime100          int64    `json:"node_get_fsm_counter_time_100"`
	NodeGetFsmCounterTime95           int64    `json:"node_get_fsm_counter_time_95"`
	NodeGetFsmCounterTime99           int64    `json:"node_get_fsm_counter_time_99"`
	NodeGetFsmCounterTimeMean         int64    `json:"node_get_fsm_counter_time_mean"`
	NodeGetFsmCounterTimeMedian       int64    `json:"node_get_fsm_counter_time_median"`
	NodeGetFsmErrors                  int64    `json:"node_get_fsm_errors"`
	NodeGetFsmErrorsTotal             int64    `json:"node_get_fsm_errors_total"`
	NodeGetFsmHllObjsize100           int64    `json:"node_get_fsm_hll_objsize_100"`
	NodeGetFsmHllObjsize95            int64    `json:"node_get_fsm_hll_objsize_95"`
	NodeGetFsmHllObjsize99            int64    `json:"node_get_fsm_hll_objsize_99"`
	NodeGetFsmHllObjsizeMean          int64    `json:"node_get_fsm_hll_objsize_mean"`
	NodeGetFsmHllObjsizeMedian        int64    `json:"node_get_fsm_hll_objsize_median"`
	NodeGetFsmHllSiblings100          int64    `json:"node_get_fsm_hll_siblings_100"`
	NodeGetFsmHllSiblings95           int64    `json:"node_get_fsm_hll_siblings_95"`
	NodeGetFsmHllSiblings99           int64    `json:"node_get_fsm_hll_siblings_99"`
	NodeGetFsmHllSiblingsMean         int64    `json:"node_get_fsm_hll_siblings_mean"`
	NodeGetFsmHllSiblingsMedian       int64    `json:"node_get_fsm_hll_siblings_median"`
	NodeGetFsmHllTime100              int64    `json:"node_get_fsm_hll_time_100"`
	NodeGetFsmHllTime95               int64    `json:"node_get_fsm_hll_time_95"`
	NodeGetFsmHllTime99               int64    `json:"node_get_fsm_hll_time_99"`
	NodeGetFsmHllTimeMean             int64    `json:"node_get_fsm_hll_time_mean"`
	NodeGetFsmHllTimeMedian           int64    `json:"node_get_fsm_hll_time_median"`
	NodeGetFsmInRate                  int64    `json:"node_get_fsm_in_rate"`
	NodeGetFsmMapObjsize100           int64    `json:"node_get_fsm_map_objsize_100"`
	NodeGetFsmMapObjsize95            int64    `json:"node_get_fsm_map_objsize_95"`
	NodeGetFsmMapObjsize99            int64    `json:"node_get_fsm_map_objsize_99"`
	NodeGetFsmMapObjsizeMean          int64    `json:"node_get_fsm_map_objsize_mean"`
	NodeGetFsmMapObjsizeMedian        int64    `json:"node_get_fsm_map_objsize_median"`
	NodeGetFsmMapSiblings100          int64    `json:"node_get_fsm_map_siblings_100"`
	NodeGetFsmMapSiblings95           int64    `json:"node_get_fsm_map_siblings_95"`
	NodeGetFsmMapSiblings99           int64    `json:"node_get_fsm_map_siblings_99"`
	NodeGetFsmMapSiblingsMean         int64    `json:"node_get_fsm_map_siblings_mean"`
	NodeGetFsmMapSiblingsMedian       int64    `json:"node_get_fsm_map_siblings_median"`
	NodeGetFsmMapTime100              int64    `json:"node_get_fsm_map_time_100"`
	NodeGetFsmMapTime95               int64    `json:"node_get_fsm_map_time_95"`
	NodeGetFsmMapTime99               int64    `json:"node_get_fsm_map_time_99"`
	NodeGetFsmMapTimeMean             int64    `json:"node_get_fsm_map_time_mean"`
	NodeGetFsmMapTimeMedian           int64    `json:"node_get_fsm_map_time_median"`
	NodeGetFsmObjsize100              int64    `json:"node_get_fsm_objsize_100"`
	NodeGetFsmObjsize95               int64    `json:"node_get_fsm_objsize_95"`
	NodeGetFsmObjsize99               int64    `json:"node_get_fsm_objsize_99"`
	NodeGetFsmObjsizeMean             int64    `json:"node_get_fsm_objsize_mean"`
	NodeGetFsmObjsizeMedian           int64    `json:"node_get_fsm_objsize_median"`
	NodeGetFsmOutRate                 int64    `json:"node_get_fsm_out_rate"`
	NodeGetFsmRejected                int64    `json:"node_get_fsm_rejected"`
	NodeGetFsmRejected60s             int64    `json:"node_get_fsm_rejected_60s"`
	NodeGetFsmRejectedTotal           int64    `json:"node_get_fsm_rejected_total"`
	NodeGetFsmSetObjsize100           int64    `json:"node_get_fsm_set_objsize_100"`
	NodeGetFsmSetObjsize95            int64    `json:"node_get_fsm_set_objsize_95"`
	NodeGetFsmSetObjsize99            int64    `json:"node_get_fsm_set_objsize_99"`
	NodeGetFsmSetObjsizeMean          int64    `json:"node_get_fsm_set_objsize_mean"`
	NodeGetFsmSetObjsizeMedian        int64    `json:"node_get_fsm_set_objsize_median"`
	NodeGetFsmSetSiblings100          int64    `json:"node_get_fsm_set_siblings_100"`
	NodeGetFsmSetSiblings95           int64    `json:"node_get_fsm_set_siblings_95"`
	NodeGetFsmSetSiblings99           int64    `json:"node_get_fsm_set_siblings_99"`
	NodeGetFsmSetSiblingsMean         int64    `json:"node_get_fsm_set_siblings_mean"`
	NodeGetFsmSetSiblingsMedian       int64    `json:"node_get_fsm_set_siblings_median"`
	NodeGetFsmSetTime100              int64    `json:"node_get_fsm_set_time_100"`
	NodeGetFsmSetTime95               int64    `json:"node_get_fsm_set_time_95"`
	NodeGetFsmSetTime99               int64    `json:"node_get_fsm_set_time_99"`
	NodeGetFsmSetTimeMean             int64    `json:"node_get_fsm_set_time_mean"`
	NodeGetFsmSetTimeMedian           int64    `json:"node_get_fsm_set_time_median"`
	NodeGetFsmSiblings100             int64    `json:"node_get_fsm_siblings_100"`
	NodeGetFsmSiblings95              int64    `json:"node_get_fsm_siblings_95"`
	NodeGetFsmSiblings99              int64    `json:"node_get_fsm_siblings_99"`
	NodeGetFsmSiblingsMean            int64    `json:"node_get_fsm_siblings_mean"`
	NodeGetFsmSiblingsMedian          int64    `json:"node_get_fsm_siblings_median"`
	NodeGetFsmTime100                 int64    `json:"node_get_fsm_time_100"`
	NodeGetFsmTime95                  int64    `json:"node_get_fsm_time_95"`
	NodeGetFsmTime99                  int64    `json:"node_get_fsm_time_99"`
	NodeGetFsmTimeMean                int64    `json:"node_get_fsm_time_mean"`
	NodeGetFsmTimeMedian              int64    `json:"node_get_fsm_time_median"`
	NodeGets                          int64    `json:"node_gets"`
	NodeGetsCounter                   int64    `json:"node_gets_counter"`
	NodeGetsCounterTotal              int64    `json:"node_gets_counter_total"`
	NodeGetsHll                       int64    `json:"node_gets_hll"`
	NodeGetsHllTotal                  int64    `json:"node_gets_hll_total"`
	NodeGetsMap                       int64    `json:"node_gets_map"`
	NodeGetsMapTotal                  int64    `json:"node_gets_map_total"`
	NodeGetsSet                       int64    `json:"node_gets_set"`
	NodeGetsSetTotal                  int64    `json:"node_gets_set_total"`
	NodeGetsTotal                     int64    `json:"node_gets_total"`
	NodePutFsmActive                  int64    `json:"node_put_fsm_active"`
	NodePutFsmActive60s               int64    `json:"node_put_fsm_active_60s"`
	NodePutFsmCounterTime100          int64    `json:"node_put_fsm_counter_time_100"`
	NodePutFsmCounterTime95           int64    `json:"node_put_fsm_counter_time_95"`
	NodePutFsmCounterTime99           int64    `json:"node_put_fsm_counter_time_99"`
	NodePutFsmCounterTimeMean         int64    `json:"node_put_fsm_counter_time_mean"`
	NodePutFsmCounterTimeMedian       int64    `json:"node_put_fsm_counter_time_median"`
	NodePutFsmHllTime100              int64    `json:"node_put_fsm_hll_time_100"`
	NodePutFsmHllTime95               int64    `json:"node_put_fsm_hll_time_95"`
	NodePutFsmHllTime99               int64    `json:"node_put_fsm_hll_time_99"`
	NodePutFsmHllTimeMean             int64    `json:"node_put_fsm_hll_time_mean"`
	NodePutFsmHllTimeMedian           int64    `json:"node_put_fsm_hll_time_median"`
	NodePutFsmInRate                  int64    `json:"node_put_fsm_in_rate"`
	NodePutFsmMapTime100              int64    `json:"node_put_fsm_map_time_100"`
	NodePutFsmMapTime95               int64    `json:"node_put_fsm_map_time_95"`
	NodePutFsmMapTime99               int64    `json:"node_put_fsm_map_time_99"`
	NodePutFsmMapTimeMean             int64    `json:"node_put_fsm_map_time_mean"`
	NodePutFsmMapTimeMedian           int64    `json:"node_put_fsm_map_time_median"`
	NodePutFsmOutRate                 int64    `json:"node_put_fsm_out_rate"`
	NodePutFsmRejected                int64    `json:"node_put_fsm_rejected"`
	NodePutFsmRejected60s             int64    `json:"node_put_fsm_rejected_60s"`
	NodePutFsmRejectedTotal           int64    `json:"node_put_fsm_rejected_total"`
	NodePutFsmSetTime100              int64    `json:"node_put_fsm_set_time_100"`
	NodePutFsmSetTime95               int64    `json:"node_put_fsm_set_time_95"`
	NodePutFsmSetTime99               int64    `json:"node_put_fsm_set_time_99"`
	NodePutFsmSetTimeMean             int64    `json:"node_put_fsm_set_time_mean"`
	NodePutFsmSetTimeMedian           int64    `json:"node_put_fsm_set_time_median"`
	NodePutFsmTime100                 int64    `json:"node_put_fsm_time_100"`
	NodePutFsmTime95                  int64    `json:"node_put_fsm_time_95"`
	NodePutFsmTime99                  int64    `json:"node_put_fsm_time_99"`
	NodePutFsmTimeMean                int64    `json:"node_put_fsm_time_mean"`
	NodePutFsmTimeMedian              int64    `json:"node_put_fsm_time_median"`
	NodePuts                          int64    `json:"node_puts"`
	NodePutsCounter                   int64    `json:"node_puts_counter"`
	NodePutsCounterTotal              int64    `json:"node_puts_counter_total"`
	NodePutsHll                       int64    `json:"node_puts_hll"`
	NodePutsHllTotal                  int64    `json:"node_puts_hll_total"`
	NodePutsMap                       int64    `json:"node_puts_map"`
	NodePutsMapTotal                  int64    `json:"node_puts_map_total"`
	NodePutsSet                       int64    `json:"node_puts_set"`
	NodePutsSetTotal                  int64    `json:"node_puts_set_total"`
	NodePutsTotal                     int64    `json:"node_puts_total"`
	Nodename                          string   `json:"nodename"`
	ObjectCounterMerge                int64    `json:"object_counter_merge"`
	ObjectCounterMergeTime100         int64    `json:"object_counter_merge_time_100"`
	ObjectCounterMergeTime95          int64    `json:"object_counter_merge_time_95"`
	ObjectCounterMergeTime99          int64    `json:"object_counter_merge_time_99"`
	ObjectCounterMergeTimeMean        int64    `json:"object_counter_merge_time_mean"`
	ObjectCounterMergeTimeMedian      int64    `json:"object_counter_merge_time_median"`
	ObjectCounterMergeTotal           int64    `json:"object_counter_merge_total"`
	ObjectHllMerge                    int64    `json:"object_hll_merge"`
	ObjectHllMergeTime100             int64    `json:"object_hll_merge_time_100"`
	ObjectHllMergeTime95              int64    `json:"object_hll_merge_time_95"`
	ObjectHllMergeTime99              int64    `json:"object_hll_merge_time_99"`
	ObjectHllMergeTimeMean            int64    `json:"object_hll_merge_time_mean"`
	ObjectHllMergeTimeMedian          int64    `json:"object_hll_merge_time_median"`
	ObjectHllMergeTotal               int64    `json:"object_hll_merge_total"`
	ObjectMapMerge                    int64    `json:"object_map_merge"`
	ObjectMapMergeTime100             int64    `json:"object_map_merge_time_100"`
	ObjectMapMergeTime95              int64    `json:"object_map_merge_time_95"`
	ObjectMapMergeTime99              int64    `json:"object_map_merge_time_99"`
	ObjectMapMergeTimeMean            int64    `json:"object_map_merge_time_mean"`
	ObjectMapMergeTimeMedian          int64    `json:"object_map_merge_time_median"`
	ObjectMapMergeTotal               int64    `json:"object_map_merge_total"`
	ObjectMerge                       int64    `json:"object_merge"`
	ObjectMergeTime100                int64    `json:"object_merge_time_100"`
	ObjectMergeTime95                 int64    `json:"object_merge_time_95"`
	ObjectMergeTime99                 int64    `json:"object_merge_time_99"`
	ObjectMergeTimeMean               int64    `json:"object_merge_time_mean"`
	ObjectMergeTimeMedian             int64    `json:"object_merge_time_median"`
	ObjectMergeTotal                  int64    `json:"object_merge_total"`
	ObjectSetMerge                    int64    `json:"object_set_merge"`
	ObjectSetMergeTime100             int64    `json:"object_set_merge_time_100"`
	ObjectSetMergeTime95              int64    `json:"object_set_merge_time_95"`
	ObjectSetMergeTime99              int64    `json:"object_set_merge_time_99"`
	ObjectSetMergeTimeMean            int64    `json:"object_set_merge_time_mean"`
	ObjectSetMergeTimeMedian          int64    `json:"object_set_merge_time_median"`
	ObjectSetMergeTotal               int64    `json:"object_set_merge_total"`
	PbcActive                         int64    `json:"pbc_active"`
	PbcConnects                       int64    `json:"pbc_connects"`
	PbcConnectsTotal                  int64    `json:"pbc_connects_total"`
	PipelineActive                    int64    `json:"pipeline_active"`
	PipelineCreateCount               int64    `json:"pipeline_create_count"`
	PipelineCreateErrorCount          int64    `json:"pipeline_create_error_count"`
	PipelineCreateErrorOne            int64    `json:"pipeline_create_error_one"`
	PipelineCreateOne                 int64    `json:"pipeline_create_one"`
	PostcommitFail                    int64    `json:"postcommit_fail"`
	PrecommitFail                     int64    `json:"precommit_fail"`
	ReadRepairs                       int64    `json:"read_repairs"`
	ReadRepairsCounter                int64    `json:"read_repairs_counter"`
	ReadRepairsCounterTotal           int64    `json:"read_repairs_counter_total"`
	ReadRepairsFallbackNotfoundCount  int64    `json:"read_repairs_fallback_notfound_count"`
	ReadRepairsFallbackNotfoundOne    int64    `json:"read_repairs_fallback_notfound_one"`
	ReadRepairsFallbackOutofdateCount int64    `json:"read_repairs_fallback_outofdate_count"`
	ReadRepairsFallbackOutofdateOne   int64    `json:"read_repairs_fallback_outofdate_one"`
	ReadRepairsHll                    int64    `json:"read_repairs_hll"`
	ReadRepairsHllTotal               int64    `json:"read_repairs_hll_total"`
	ReadRepairsMap                    int64    `json:"read_repairs_map"`
	ReadRepairsMapTotal               int64    `json:"read_repairs_map_total"`
	ReadRepairsPrimaryNotfoundCount   int64    `json:"read_repairs_primary_notfound_count"`
	ReadRepairsPrimaryNotfoundOne     int64    `json:"read_repairs_primary_notfound_one"`
	ReadRepairsPrimaryOutofdateCount  int64    `json:"read_repairs_primary_outofdate_count"`
	ReadRepairsPrimaryOutofdateOne    int64    `json:"read_repairs_primary_outofdate_one"`
	ReadRepairsSet                    int64    `json:"read_repairs_set"`
	ReadRepairsSetTotal               int64    `json:"read_repairs_set_total"`
	ReadRepairsTotal                  int64    `json:"read_repairs_total"`
	RebalanceDelayLast                int64    `json:"rebalance_delay_last"`
	RebalanceDelayMax                 int64    `json:"rebalance_delay_max"`
	RebalanceDelayMean                int64    `json:"rebalance_delay_mean"`
	RebalanceDelayMin                 int64    `json:"rebalance_delay_min"`
	RejectedHandoffs                  int64    `json:"rejected_handoffs"`
	RiakKvVnodeqMax                   int64    `json:"riak_kv_vnodeq_max"`
	RiakKvVnodeqMean                  int64    `json:"riak_kv_vnodeq_mean"`
	RiakKvVnodeqMedian                int64    `json:"riak_kv_vnodeq_median"`
	RiakKvVnodeqMin                   int64    `json:"riak_kv_vnodeq_min"`
	RiakKvVnodeqTotal                 int64    `json:"riak_kv_vnodeq_total"`
	RiakKvVnodesRunning               int64    `json:"riak_kv_vnodes_running"`
	RiakPipeVnodeqMax                 int64    `json:"riak_pipe_vnodeq_max"`
	RiakPipeVnodeqMean                int64    `json:"riak_pipe_vnodeq_mean"`
	RiakPipeVnodeqMedian              int64    `json:"riak_pipe_vnodeq_median"`
	RiakPipeVnodeqMin                 int64    `json:"riak_pipe_vnodeq_min"`
	RiakPipeVnodeqTotal               int64    `json:"riak_pipe_vnodeq_total"`
	RiakPipeVnodesRunning             int64    `json:"riak_pipe_vnodes_running"`
	RingCreationSize                  int64    `json:"ring_creation_size"`
	RingMembers                       []string `json:"ring_members"`
	RingNumPartitions                 int64    `json:"ring_num_partitions"`
	RingOwnership                     string   `json:"ring_ownership"`
	RingsReconciled                   int64    `json:"rings_reconciled"`
	RingsReconciledTotal              int64    `json:"rings_reconciled_total"`
	SetActorCounts100                 int64    `json:"set_actor_counts_100"`
	SetActorCounts95                  int64    `json:"set_actor_counts_95"`
	SetActorCounts99                  int64    `json:"set_actor_counts_99"`
	SetActorCountsMean                int64    `json:"set_actor_counts_mean"`
	SetActorCountsMedian              int64    `json:"set_actor_counts_median"`
	SkippedReadRepairs                int64    `json:"skipped_read_repairs"`
	SkippedReadRepairsTotal           int64    `json:"skipped_read_repairs_total"`
	SoftLoadedVnodeMboxTotal          int64    `json:"soft_loaded_vnode_mbox_total"`
	StorageBackend                    string   `json:"storage_backend"`
	SysDriverVersion                  string   `json:"sys_driver_version"`
	SysGlobalHeapsSize                string   `json:"sys_global_heaps_size"`
	SysHeapType                       string   `json:"sys_heap_type"`
	SysLogicalProcessors              int64    `json:"sys_logical_processors"`
	SysMonitorCount                   int64    `json:"sys_monitor_count"`
	SysOtpRelease                     string   `json:"sys_otp_release"`
	SysPortCount                      int64    `json:"sys_port_count"`
	SysProcessCount                   int64    `json:"sys_process_count"`
	SysSmpSupport                     bool     `json:"sys_smp_support"`
	SysSystemArchitecture             string   `json:"sys_system_architecture"`
	SysSystemVersion                  string   `json:"sys_system_version"`
	SysThreadPoolSize                 int64    `json:"sys_thread_pool_size"`
	SysThreadsEnabled                 bool     `json:"sys_threads_enabled"`
	SysWordsize                       int64    `json:"sys_wordsize"`
	TictacaaeBranchCompare            int64    `json:"tictacaae_branch_compare"`
	TictacaaeBranchCompareTotal       int64    `json:"tictacaae_branch_compare_total"`
	TictacaaeBucket                   int64    `json:"tictacaae_bucket"`
	TictacaaeBucketTotal              int64    `json:"tictacaae_bucket_total"`
	TictacaaeClockCompare             int64    `json:"tictacaae_clock_compare"`
	TictacaaeClockCompareTotal        int64    `json:"tictacaae_clock_compare_total"`
	TictacaaeError                    int64    `json:"tictacaae_error"`
	TictacaaeErrorTotal               int64    `json:"tictacaae_error_total"`
	TictacaaeExchange                 int64    `json:"tictacaae_exchange"`
	TictacaaeExchangeTotal            int64    `json:"tictacaae_exchange_total"`
	TictacaaeModtime                  int64    `json:"tictacaae_modtime"`
	TictacaaeModtimeTotal             int64    `json:"tictacaae_modtime_total"`
	TictacaaeNotSupported             int64    `json:"tictacaae_not_supported"`
	TictacaaeNotSupportedTotal        int64    `json:"tictacaae_not_supported_total"`
	TictacaaeQueueMicrosecMax         int64    `json:"tictacaae_queue_microsec__max"`
	TictacaaeQueueMicrosecMean        int64    `json:"tictacaae_queue_microsec_mean"`
	TictacaaeRootCompare              int64    `json:"tictacaae_root_compare"`
	TictacaaeRootCompareTotal         int64    `json:"tictacaae_root_compare_total"`
	TictacaaeTimeout                  int64    `json:"tictacaae_timeout"`
	TictacaaeTimeoutTotal             int64    `json:"tictacaae_timeout_total"`
	TtaaefsAllcheckTotal              int64    `json:"ttaaefs_allcheck_total"`
	TtaaefsDaycheckTotal              int64    `json:"ttaaefs_daycheck_total"`
	TtaaefsFailTime100                int64    `json:"ttaaefs_fail_time_100"`
	TtaaefsFailTotal                  int64    `json:"ttaaefs_fail_total"`
	TtaaefsHourcheckTotal             int64    `json:"ttaaefs_hourcheck_total"`
	TtaaefsNosyncTime100              int64    `json:"ttaaefs_nosync_time_100"`
	TtaaefsNosyncTotal                int64    `json:"ttaaefs_nosync_total"`
	TtaaefsRangecheckTotal            int64    `json:"ttaaefs_rangecheck_total"`
	TtaaefsSnkAheadTotal              int64    `json:"ttaaefs_snk_ahead_total"`
	TtaaefsSrcAheadTotal              int64    `json:"ttaaefs_src_ahead_total"`
	TtaaefsSyncTime100                int64    `json:"ttaaefs_sync_time_100"`
	TtaaefsSyncTotal                  int64    `json:"ttaaefs_sync_total"`
	VnodeCounterUpdate                int64    `json:"vnode_counter_update"`
	VnodeCounterUpdateTime100         int64    `json:"vnode_counter_update_time_100"`
	VnodeCounterUpdateTime95          int64    `json:"vnode_counter_update_time_95"`
	VnodeCounterUpdateTime99          int64    `json:"vnode_counter_update_time_99"`
	VnodeCounterUpdateTimeMean        int64    `json:"vnode_counter_update_time_mean"`
	VnodeCounterUpdateTimeMedian      int64    `json:"vnode_counter_update_time_median"`
	VnodeCounterUpdateTotal           int64    `json:"vnode_counter_update_total"`
	VnodeGetFsmTime100                int64    `json:"vnode_get_fsm_time_100"`
	VnodeGetFsmTime95                 int64    `json:"vnode_get_fsm_time_95"`
	VnodeGetFsmTime99                 int64    `json:"vnode_get_fsm_time_99"`
	VnodeGetFsmTimeMean               int64    `json:"vnode_get_fsm_time_mean"`
	VnodeGetFsmTimeMedian             int64    `json:"vnode_get_fsm_time_median"`
	VnodeGets                         int64    `json:"vnode_gets"`
	VnodeGetsTotal                    int64    `json:"vnode_gets_total"`
	VnodeHeadFsmTime100               int64    `json:"vnode_head_fsm_time_100"`
	VnodeHeadFsmTime95                int64    `json:"vnode_head_fsm_time_95"`
	VnodeHeadFsmTime99                int64    `json:"vnode_head_fsm_time_99"`
	VnodeHeadFsmTimeMean              int64    `json:"vnode_head_fsm_time_mean"`
	VnodeHeadFsmTimeMedian            int64    `json:"vnode_head_fsm_time_median"`
	VnodeHeads                        int64    `json:"vnode_heads"`
	VnodeHeadsTotal                   int64    `json:"vnode_heads_total"`
	VnodeHllUpdate                    int64    `json:"vnode_hll_update"`
	VnodeHllUpdateTime100             int64    `json:"vnode_hll_update_time_100"`
	VnodeHllUpdateTime95              int64    `json:"vnode_hll_update_time_95"`
	VnodeHllUpdateTime99              int64    `json:"vnode_hll_update_time_99"`
	VnodeHllUpdateTimeMean            int64    `json:"vnode_hll_update_time_mean"`
	VnodeHllUpdateTimeMedian          int64    `json:"vnode_hll_update_time_median"`
	VnodeHllUpdateTotal               int64    `json:"vnode_hll_update_total"`
	VnodeIndexDeletes                 int64    `json:"vnode_index_deletes"`
	VnodeIndexDeletesPostings         int64    `json:"vnode_index_deletes_postings"`
	VnodeIndexDeletesPostingsTotal    int64    `json:"vnode_index_deletes_postings_total"`
	VnodeIndexDeletesTotal            int64    `json:"vnode_index_deletes_total"`
	VnodeIndexReads                   int64    `json:"vnode_index_reads"`
	VnodeIndexReadsTotal              int64    `json:"vnode_index_reads_total"`
	VnodeIndexRefreshes               int64    `json:"vnode_index_refreshes"`
	VnodeIndexRefreshesTotal          int64    `json:"vnode_index_refreshes_total"`
	VnodeIndexWrites                  int64    `json:"vnode_index_writes"`
	VnodeIndexWritesPostings          int64    `json:"vnode_index_writes_postings"`
	VnodeIndexWritesPostingsTotal     int64    `json:"vnode_index_writes_postings_total"`
	VnodeIndexWritesTotal             int64    `json:"vnode_index_writes_total"`
	VnodeMapUpdate                    int64    `json:"vnode_map_update"`
	VnodeMapUpdateTime100             int64    `json:"vnode_map_update_time_100"`
	VnodeMapUpdateTime95              int64    `json:"vnode_map_update_time_95"`
	VnodeMapUpdateTime99              int64    `json:"vnode_map_update_time_99"`
	VnodeMapUpdateTimeMean            int64    `json:"vnode_map_update_time_mean"`
	VnodeMapUpdateTimeMedian          int64    `json:"vnode_map_update_time_median"`
	VnodeMapUpdateTotal               int64    `json:"vnode_map_update_total"`
	VnodeMboxCheckTimeoutTotal        int64    `json:"vnode_mbox_check_timeout_total"`
	VnodePutFsmTime100                int64    `json:"vnode_put_fsm_time_100"`
	VnodePutFsmTime95                 int64    `json:"vnode_put_fsm_time_95"`
	VnodePutFsmTime99                 int64    `json:"vnode_put_fsm_time_99"`
	VnodePutFsmTimeMean               int64    `json:"vnode_put_fsm_time_mean"`
	VnodePutFsmTimeMedian             int64    `json:"vnode_put_fsm_time_median"`
	VnodePuts                         int64    `json:"vnode_puts"`
	VnodePutsTotal                    int64    `json:"vnode_puts_total"`
	VnodeSetUpdate                    int64    `json:"vnode_set_update"`
	VnodeSetUpdateTime100             int64    `json:"vnode_set_update_time_100"`
	VnodeSetUpdateTime95              int64    `json:"vnode_set_update_time_95"`
	VnodeSetUpdateTime99              int64    `json:"vnode_set_update_time_99"`
	VnodeSetUpdateTimeMean            int64    `json:"vnode_set_update_time_mean"`
	VnodeSetUpdateTimeMedian          int64    `json:"vnode_set_update_time_median"`
	VnodeSetUpdateTotal               int64    `json:"vnode_set_update_total"`
	WorkerAf1PoolQueuetime100         int64    `json:"worker_af1_pool_queuetime_100"`
	WorkerAf1PoolQueuetimeMean        int64    `json:"worker_af1_pool_queuetime_mean"`
	WorkerAf1PoolTotal                int64    `json:"worker_af1_pool_total"`
	WorkerAf1PoolWorktime100          int64    `json:"worker_af1_pool_worktime_100"`
	WorkerAf1PoolWorktimeMean         int64    `json:"worker_af1_pool_worktime_mean"`
	WorkerAf2PoolQueuetime100         int64    `json:"worker_af2_pool_queuetime_100"`
	WorkerAf2PoolQueuetimeMean        int64    `json:"worker_af2_pool_queuetime_mean"`
	WorkerAf2PoolTotal                int64    `json:"worker_af2_pool_total"`
	WorkerAf2PoolWorktime100          int64    `json:"worker_af2_pool_worktime_100"`
	WorkerAf2PoolWorktimeMean         int64    `json:"worker_af2_pool_worktime_mean"`
	WorkerAf3PoolQueuetime100         int64    `json:"worker_af3_pool_queuetime_100"`
	WorkerAf3PoolQueuetimeMean        int64    `json:"worker_af3_pool_queuetime_mean"`
	WorkerAf3PoolTotal                int64    `json:"worker_af3_pool_total"`
	WorkerAf3PoolWorktime100          int64    `json:"worker_af3_pool_worktime_100"`
	WorkerAf3PoolWorktimeMean         int64    `json:"worker_af3_pool_worktime_mean"`
	WorkerAf4PoolQueuetime100         int64    `json:"worker_af4_pool_queuetime_100"`
	WorkerAf4PoolQueuetimeMean        int64    `json:"worker_af4_pool_queuetime_mean"`
	WorkerAf4PoolTotal                int64    `json:"worker_af4_pool_total"`
	WorkerAf4PoolWorktime100          int64    `json:"worker_af4_pool_worktime_100"`
	WorkerAf4PoolWorktimeMean         int64    `json:"worker_af4_pool_worktime_mean"`
	WorkerBePoolQueuetime100          int64    `json:"worker_be_pool_queuetime_100"`
	WorkerBePoolQueuetimeMean         int64    `json:"worker_be_pool_queuetime_mean"`
	WorkerBePoolTotal                 int64    `json:"worker_be_pool_total"`
	WorkerBePoolWorktime100           int64    `json:"worker_be_pool_worktime_100"`
	WorkerBePoolWorktimeMean          int64    `json:"worker_be_pool_worktime_mean"`
	WorkerNodeWorkerPoolQueuetime100  int64    `json:"worker_node_worker_pool_queuetime_100"`
	WorkerNodeWorkerPoolQueuetimeMean int64    `json:"worker_node_worker_pool_queuetime_mean"`
	WorkerNodeWorkerPoolTotal         int64    `json:"worker_node_worker_pool_total"`
	WorkerNodeWorkerPoolWorktime100   int64    `json:"worker_node_worker_pool_worktime_100"`
	WorkerNodeWorkerPoolWorktimeMean  int64    `json:"worker_node_worker_pool_worktime_mean"`
	WorkerUnregisteredQueuetime100    int64    `json:"worker_unregistered_queuetime_100"`
	WorkerUnregisteredQueuetimeMean   int64    `json:"worker_unregistered_queuetime_mean"`
	WorkerUnregisteredTotal           int64    `json:"worker_unregistered_total"`
	WorkerUnregisteredWorktime100     int64    `json:"worker_unregistered_worktime_100"`
	WorkerUnregisteredWorktimeMean    int64    `json:"worker_unregistered_worktime_mean"`
	WorkerVnodePoolQueuetime100       int64    `json:"worker_vnode_pool_queuetime_100"`
	WorkerVnodePoolQueuetimeMean      int64    `json:"worker_vnode_pool_queuetime_mean"`
	WorkerVnodePoolTotal              int64    `json:"worker_vnode_pool_total"`
	WorkerVnodePoolWorktime100        int64    `json:"worker_vnode_pool_worktime_100"`
	WorkerVnodePoolWorktimeMean       int64    `json:"worker_vnode_pool_worktime_mean"`
	WriteOnceMerge                    int64    `json:"write_once_merge"`
	WriteOncePutObjsize100            int64    `json:"write_once_put_objsize_100"`
	WriteOncePutObjsize95             int64    `json:"write_once_put_objsize_95"`
	WriteOncePutObjsize99             int64    `json:"write_once_put_objsize_99"`
	WriteOncePutObjsizeMean           int64    `json:"write_once_put_objsize_mean"`
	WriteOncePutObjsizeMedian         int64    `json:"write_once_put_objsize_median"`
	WriteOncePutTime100               int64    `json:"write_once_put_time_100"`
	WriteOncePutTime95                int64    `json:"write_once_put_time_95"`
	WriteOncePutTime99                int64    `json:"write_once_put_time_99"`
	WriteOncePutTimeMean              int64    `json:"write_once_put_time_mean"`
	WriteOncePutTimeMedian            int64    `json:"write_once_put_time_median"`
	WriteOncePuts                     int64    `json:"write_once_puts"`
	WriteOncePutsTotal                int64    `json:"write_once_puts_total"`
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
		"clusteraae_fsm_active":                  stats.ClusterAaeFsmActive,
		"clusteraae_fsm_create":                  stats.ClusterAaeFsmCreate,
		"clusteraae_fsm_create_error":            stats.ClusterAaeFsmCreateError,
		"connected_nodes":                        stats.ConnectedNodes,
		"consistent_get_objsize_100":             stats.ConsistentGetObjsize100,
		"consistent_get_objsize_95":              stats.ConsistentGetObjsize95,
		"consistent_get_objsize_99":              stats.ConsistentGetObjsize99,
		"consistent_get_objsize_mean":            stats.ConsistentGetObjsizeMean,
		"consistent_get_objsize_median":          stats.ConsistentGetObjsizeMedian,
		"consistent_get_time_100":                stats.ConsistentGetTime100,
		"consistent_get_time_95":                 stats.ConsistentGetTime95,
		"consistent_get_time_99":                 stats.ConsistentGetTime99,
		"consistent_get_time_mean":               stats.ConsistentGetTimeMean,
		"consistent_get_time_median":             stats.ConsistentGetTimeMedian,
		"consistent_gets":                        stats.ConsistentGets,
		"consistent_gets_total":                  stats.ConsistentGetsTotal,
		"consistent_put_objsize_100":             stats.ConsistentPutObjsize100,
		"consistent_put_objsize_95":              stats.ConsistentPutObjsize95,
		"consistent_put_objsize_99":              stats.ConsistentPutObjsize99,
		"consistent_put_objsize_mean":            stats.ConsistentPutObjsizeMean,
		"consistent_put_objsize_median":          stats.ConsistentPutObjsizeMedian,
		"consistent_put_time_100":                stats.ConsistentPutTime100,
		"consistent_put_time_95":                 stats.ConsistentPutTime95,
		"consistent_put_time_99":                 stats.ConsistentPutTime99,
		"consistent_put_time_mean":               stats.ConsistentPutTimeMean,
		"consistent_put_time_median":             stats.ConsistentPutTimeMedian,
		"consistent_puts":                        stats.ConsistentPuts,
		"consistent_puts_total":                  stats.ConsistentPutsTotal,
		"converge_delay_last":                    stats.ConvergeDelayLast,
		"converge_delay_max":                     stats.ConvergeDelayMax,
		"converge_delay_mean":                    stats.ConvergeDelayMean,
		"converge_delay_min":                     stats.ConvergeDelayMin,
		"coord_local_soft_loaded_total":          stats.CoordLocalSoftLoadedTotal,
		"coord_local_unloaded_total":             stats.CoordLocalUnloadedTotal,
		"coord_redir_least_loaded_total":         stats.CoordRedirLeastLoadedTotal,
		"coord_redir_loaded_local_total":         stats.CoordRedirLoadedLocalTotal,
		"coord_redir_unloaded_total":             stats.CoordRedirUnloadedTotal,
		"coord_redirs_total":                     stats.CoordRedirsTotal,
		"counter_actor_counts_100":               stats.CounterActorCounts100,
		"counter_actor_counts_95":                stats.CounterActorCounts95,
		"counter_actor_counts_99":                stats.CounterActorCounts99,
		"counter_actor_counts_mean":              stats.CounterActorCountsMean,
		"counter_actor_counts_median":            stats.CounterActorCountsMedian,
		"cpu_avg1":                               stats.CPUAvg1,
		"cpu_avg15":                              stats.CPUAvg15,
		"cpu_avg5":                               stats.CPUAvg5,
		"cpu_nprocs":                             stats.CPUNprocs,
		"dropped_vnode_requests_totals":          stats.DroppedVnodeRequestsTotals,
		"executing_mappers":                      stats.ExecutingMappers,
		"gossip_received":                        stats.GossipReceived,
		"handoff_timeouts":                       stats.HandoffTimeouts,
		"hll_bytes":                              stats.HllBytes,
		"hll_bytes_100":                          stats.HllBytes100,
		"hll_bytes_95":                           stats.HllBytes95,
		"hll_bytes_99":                           stats.HllBytes99,
		"hll_bytes_mean":                         stats.HllBytesMean,
		"hll_bytes_median":                       stats.HllBytesMedian,
		"hll_bytes_total":                        stats.HllBytesTotal,
		"ignored_gossip_total":                   stats.IgnoredGossipTotal,
		"index_fsm_active":                       stats.IndexFsmActive,
		"index_fsm_complete":                     stats.IndexFsmComplete,
		"index_fsm_create":                       stats.IndexFsmCreate,
		"index_fsm_create_error":                 stats.IndexFsmCreateError,
		"index_fsm_results_100":                  stats.IndexFsmResults100,
		"index_fsm_results_95":                   stats.IndexFsmResults95,
		"index_fsm_results_99":                   stats.IndexFsmResults99,
		"index_fsm_results_mean":                 stats.IndexFsmResultsMean,
		"index_fsm_results_median":               stats.IndexFsmResultsMedian,
		"index_fsm_time_100":                     stats.IndexFsmTime100,
		"index_fsm_time_95":                      stats.IndexFsmTime95,
		"index_fsm_time_99":                      stats.IndexFsmTime99,
		"index_fsm_time_mean":                    stats.IndexFsmTimeMean,
		"index_fsm_time_median":                  stats.IndexFsmTimeMedian,
		"late_put_fsm_coordinator_ack":           stats.LatePutFsmCoordinatorAck,
		"leveldb_read_block_error":               stats.LeveldbReadBlockError,
		"list_fsm_active":                        stats.ListFsmActive,
		"list_fsm_create":                        stats.ListFsmCreate,
		"list_fsm_create_error":                  stats.ListFsmCreateError,
		"list_fsm_create_error_total":            stats.ListFsmCreateErrorTotal,
		"list_fsm_create_total":                  stats.ListFsmCreateTotal,
		"map_actor_counts_100":                   stats.MapActorCounts100,
		"map_actor_counts_95":                    stats.MapActorCounts95,
		"map_actor_counts_99":                    stats.MapActorCounts99,
		"map_actor_counts_mean":                  stats.MapActorCountsMean,
		"map_actor_counts_median":                stats.MapActorCountsMedian,
		"mem_allocated":                          stats.MemAllocated,
		"mem_total":                              stats.MemTotal,
		"memory_atom":                            stats.MemoryAtom,
		"memory_atom_used":                       stats.MemoryAtomUsed,
		"memory_binary":                          stats.MemoryBinary,
		"memory_code":                            stats.MemoryCode,
		"memory_ets":                             stats.MemoryEts,
		"memory_processes":                       stats.MemoryProcesses,
		"memory_processes_used":                  stats.MemoryProcessesUsed,
		"memory_system":                          stats.MemorySystem,
		"memory_total":                           stats.MemoryTotal,
		"ngrfetch_nofetch":                       stats.NgrfetchNofetch,
		"ngrfetch_nofetch_total":                 stats.NgrfetchNofetchTotal,
		"ngrfetch_prefetch":                      stats.NgrfetchPrefetch,
		"ngrfetch_prefetch_total":                stats.NgrfetchPrefetchTotal,
		"ngrfetch_tofetch":                       stats.NgrfetchTofetch,
		"ngrfetch_tofetch_total":                 stats.NgrfetchTofetchTotal,
		"ngrrepl_empty":                          stats.NgrreplEmpty,
		"ngrrepl_empty_total":                    stats.NgrreplEmptyTotal,
		"ngrrepl_error":                          stats.NgrreplError,
		"ngrrepl_error_total":                    stats.NgrreplErrorTotal,
		"ngrrepl_object":                         stats.NgrreplObject,
		"ngrrepl_object_total":                   stats.NgrreplObjectTotal,
		"ngrrepl_srcdiscard":                     stats.NgrreplSrcdiscard,
		"ngrrepl_srcdiscard_total":               stats.NgrreplSrcdiscardTotal,
		"node_get_fsm_active":                    stats.NodeGetFsmActive,
		"node_get_fsm_active_60s":                stats.NodeGetFsmActive60s,
		"node_get_fsm_counter_objsize_100":       stats.NodeGetFsmCounterObjsize100,
		"node_get_fsm_counter_objsize_95":        stats.NodeGetFsmCounterObjsize95,
		"node_get_fsm_counter_objsize_99":        stats.NodeGetFsmCounterObjsize99,
		"node_get_fsm_counter_objsize_mean":      stats.NodeGetFsmCounterObjsizeMean,
		"node_get_fsm_counter_objsize_median":    stats.NodeGetFsmCounterObjsizeMedian,
		"node_get_fsm_counter_siblings_100":      stats.NodeGetFsmCounterSiblings100,
		"node_get_fsm_counter_siblings_95":       stats.NodeGetFsmCounterSiblings95,
		"node_get_fsm_counter_siblings_99":       stats.NodeGetFsmCounterSiblings99,
		"node_get_fsm_counter_siblings_mean":     stats.NodeGetFsmCounterSiblingsMean,
		"node_get_fsm_counter_siblings_median":   stats.NodeGetFsmCounterSiblingsMedian,
		"node_get_fsm_counter_time_100":          stats.NodeGetFsmCounterTime100,
		"node_get_fsm_counter_time_95":           stats.NodeGetFsmCounterTime95,
		"node_get_fsm_counter_time_99":           stats.NodeGetFsmCounterTime99,
		"node_get_fsm_counter_time_mean":         stats.NodeGetFsmCounterTimeMean,
		"node_get_fsm_counter_time_median":       stats.NodeGetFsmCounterTimeMedian,
		"node_get_fsm_errors":                    stats.NodeGetFsmErrors,
		"node_get_fsm_errors_total":              stats.NodeGetFsmErrorsTotal,
		"node_get_fsm_hll_objsize_100":           stats.NodeGetFsmHllObjsize100,
		"node_get_fsm_hll_objsize_95":            stats.NodeGetFsmHllObjsize95,
		"node_get_fsm_hll_objsize_99":            stats.NodeGetFsmHllObjsize99,
		"node_get_fsm_hll_objsize_mean":          stats.NodeGetFsmHllObjsizeMean,
		"node_get_fsm_hll_objsize_median":        stats.NodeGetFsmHllObjsizeMedian,
		"node_get_fsm_hll_siblings_100":          stats.NodeGetFsmHllSiblings100,
		"node_get_fsm_hll_siblings_95":           stats.NodeGetFsmHllSiblings95,
		"node_get_fsm_hll_siblings_99":           stats.NodeGetFsmHllSiblings99,
		"node_get_fsm_hll_siblings_mean":         stats.NodeGetFsmHllSiblingsMean,
		"node_get_fsm_hll_siblings_median":       stats.NodeGetFsmHllSiblingsMedian,
		"node_get_fsm_hll_time_100":              stats.NodeGetFsmHllTime100,
		"node_get_fsm_hll_time_95":               stats.NodeGetFsmHllTime95,
		"node_get_fsm_hll_time_99":               stats.NodeGetFsmHllTime99,
		"node_get_fsm_hll_time_mean":             stats.NodeGetFsmHllTimeMean,
		"node_get_fsm_hll_time_median":           stats.NodeGetFsmHllTimeMedian,
		"node_get_fsm_in_rate":                   stats.NodeGetFsmInRate,
		"node_get_fsm_map_objsize_100":           stats.NodeGetFsmMapObjsize100,
		"node_get_fsm_map_objsize_95":            stats.NodeGetFsmMapObjsize95,
		"node_get_fsm_map_objsize_99":            stats.NodeGetFsmMapObjsize99,
		"node_get_fsm_map_objsize_mean":          stats.NodeGetFsmMapObjsizeMean,
		"node_get_fsm_map_objsize_median":        stats.NodeGetFsmMapObjsizeMedian,
		"node_get_fsm_map_siblings_100":          stats.NodeGetFsmMapSiblings100,
		"node_get_fsm_map_siblings_95":           stats.NodeGetFsmMapSiblings95,
		"node_get_fsm_map_siblings_99":           stats.NodeGetFsmMapSiblings99,
		"node_get_fsm_map_siblings_mean":         stats.NodeGetFsmMapSiblingsMean,
		"node_get_fsm_map_siblings_median":       stats.NodeGetFsmMapSiblingsMedian,
		"node_get_fsm_map_time_100":              stats.NodeGetFsmMapTime100,
		"node_get_fsm_map_time_95":               stats.NodeGetFsmMapTime95,
		"node_get_fsm_map_time_99":               stats.NodeGetFsmMapTime99,
		"node_get_fsm_map_time_mean":             stats.NodeGetFsmMapTimeMean,
		"node_get_fsm_map_time_median":           stats.NodeGetFsmMapTimeMedian,
		"node_get_fsm_objsize_100":               stats.NodeGetFsmObjsize100,
		"node_get_fsm_objsize_95":                stats.NodeGetFsmObjsize95,
		"node_get_fsm_objsize_99":                stats.NodeGetFsmObjsize99,
		"node_get_fsm_objsize_mean":              stats.NodeGetFsmObjsizeMean,
		"node_get_fsm_objsize_median":            stats.NodeGetFsmObjsizeMedian,
		"node_get_fsm_out_rate":                  stats.NodeGetFsmOutRate,
		"node_get_fsm_rejected":                  stats.NodeGetFsmRejected,
		"node_get_fsm_rejected_60s":              stats.NodeGetFsmRejected60s,
		"node_get_fsm_rejected_total":            stats.NodeGetFsmRejectedTotal,
		"node_get_fsm_set_objsize_100":           stats.NodeGetFsmSetObjsize100,
		"node_get_fsm_set_objsize_95":            stats.NodeGetFsmSetObjsize95,
		"node_get_fsm_set_objsize_99":            stats.NodeGetFsmSetObjsize99,
		"node_get_fsm_set_objsize_mean":          stats.NodeGetFsmSetObjsizeMean,
		"node_get_fsm_set_objsize_median":        stats.NodeGetFsmSetObjsizeMedian,
		"node_get_fsm_set_siblings_100":          stats.NodeGetFsmSetSiblings100,
		"node_get_fsm_set_siblings_95":           stats.NodeGetFsmSetSiblings95,
		"node_get_fsm_set_siblings_99":           stats.NodeGetFsmSetSiblings99,
		"node_get_fsm_set_siblings_mean":         stats.NodeGetFsmSetSiblingsMean,
		"node_get_fsm_set_siblings_median":       stats.NodeGetFsmSetSiblingsMedian,
		"node_get_fsm_set_time_100":              stats.NodeGetFsmSetTime100,
		"node_get_fsm_set_time_95":               stats.NodeGetFsmSetTime95,
		"node_get_fsm_set_time_99":               stats.NodeGetFsmSetTime99,
		"node_get_fsm_set_time_mean":             stats.NodeGetFsmSetTimeMean,
		"node_get_fsm_set_time_median":           stats.NodeGetFsmSetTimeMedian,
		"node_get_fsm_siblings_100":              stats.NodeGetFsmSiblings100,
		"node_get_fsm_siblings_95":               stats.NodeGetFsmSiblings95,
		"node_get_fsm_siblings_99":               stats.NodeGetFsmSiblings99,
		"node_get_fsm_siblings_mean":             stats.NodeGetFsmSiblingsMean,
		"node_get_fsm_siblings_median":           stats.NodeGetFsmSiblingsMedian,
		"node_get_fsm_time_100":                  stats.NodeGetFsmTime100,
		"node_get_fsm_time_95":                   stats.NodeGetFsmTime95,
		"node_get_fsm_time_99":                   stats.NodeGetFsmTime99,
		"node_get_fsm_time_mean":                 stats.NodeGetFsmTimeMean,
		"node_get_fsm_time_median":               stats.NodeGetFsmTimeMedian,
		"node_gets":                              stats.NodeGets,
		"node_gets_counter":                      stats.NodeGetsCounter,
		"node_gets_counter_total":                stats.NodeGetsCounterTotal,
		"node_gets_hll":                          stats.NodeGetsHll,
		"node_gets_hll_total":                    stats.NodeGetsHllTotal,
		"node_gets_map":                          stats.NodeGetsMap,
		"node_gets_map_total":                    stats.NodeGetsMapTotal,
		"node_gets_set":                          stats.NodeGetsSet,
		"node_gets_set_total":                    stats.NodeGetsSetTotal,
		"node_gets_total":                        stats.NodeGetsTotal,
		"node_put_fsm_active":                    stats.NodePutFsmActive,
		"node_put_fsm_active_60s":                stats.NodePutFsmActive60s,
		"node_put_fsm_counter_time_100":          stats.NodePutFsmCounterTime100,
		"node_put_fsm_counter_time_95":           stats.NodePutFsmCounterTime95,
		"node_put_fsm_counter_time_99":           stats.NodePutFsmCounterTime99,
		"node_put_fsm_counter_time_mean":         stats.NodePutFsmCounterTimeMean,
		"node_put_fsm_counter_time_median":       stats.NodePutFsmCounterTimeMedian,
		"node_put_fsm_hll_time_100":              stats.NodePutFsmHllTime100,
		"node_put_fsm_hll_time_95":               stats.NodePutFsmHllTime95,
		"node_put_fsm_hll_time_99":               stats.NodePutFsmHllTime99,
		"node_put_fsm_hll_time_mean":             stats.NodePutFsmHllTimeMean,
		"node_put_fsm_hll_time_median":           stats.NodePutFsmHllTimeMedian,
		"node_put_fsm_in_rate":                   stats.NodePutFsmInRate,
		"node_put_fsm_map_time_100":              stats.NodePutFsmMapTime100,
		"node_put_fsm_map_time_95":               stats.NodePutFsmMapTime95,
		"node_put_fsm_map_time_99":               stats.NodePutFsmMapTime99,
		"node_put_fsm_map_time_mean":             stats.NodePutFsmMapTimeMean,
		"node_put_fsm_map_time_median":           stats.NodePutFsmMapTimeMedian,
		"node_put_fsm_out_rate":                  stats.NodePutFsmOutRate,
		"node_put_fsm_rejected":                  stats.NodePutFsmRejected,
		"node_put_fsm_rejected_60s":              stats.NodePutFsmRejected60s,
		"node_put_fsm_rejected_total":            stats.NodePutFsmRejectedTotal,
		"node_put_fsm_set_time_100":              stats.NodePutFsmSetTime100,
		"node_put_fsm_set_time_95":               stats.NodePutFsmSetTime95,
		"node_put_fsm_set_time_99":               stats.NodePutFsmSetTime99,
		"node_put_fsm_set_time_mean":             stats.NodePutFsmSetTimeMean,
		"node_put_fsm_set_time_median":           stats.NodePutFsmSetTimeMedian,
		"node_put_fsm_time_100":                  stats.NodePutFsmTime100,
		"node_put_fsm_time_95":                   stats.NodePutFsmTime95,
		"node_put_fsm_time_99":                   stats.NodePutFsmTime99,
		"node_put_fsm_time_mean":                 stats.NodePutFsmTimeMean,
		"node_put_fsm_time_median":               stats.NodePutFsmTimeMedian,
		"node_puts":                              stats.NodePuts,
		"node_puts_counter":                      stats.NodePutsCounter,
		"node_puts_counter_total":                stats.NodePutsCounterTotal,
		"node_puts_hll":                          stats.NodePutsHll,
		"node_puts_hll_total":                    stats.NodePutsHllTotal,
		"node_puts_map":                          stats.NodePutsMap,
		"node_puts_map_total":                    stats.NodePutsMapTotal,
		"node_puts_set":                          stats.NodePutsSet,
		"node_puts_set_total":                    stats.NodePutsSetTotal,
		"node_puts_total":                        stats.NodePutsTotal,
		"nodename":                               stats.Nodename,
		"object_counter_merge":                   stats.ObjectCounterMerge,
		"object_counter_merge_time_100":          stats.ObjectCounterMergeTime100,
		"object_counter_merge_time_95":           stats.ObjectCounterMergeTime95,
		"object_counter_merge_time_99":           stats.ObjectCounterMergeTime99,
		"object_counter_merge_time_mean":         stats.ObjectCounterMergeTimeMean,
		"object_counter_merge_time_median":       stats.ObjectCounterMergeTimeMedian,
		"object_counter_merge_total":             stats.ObjectCounterMergeTotal,
		"object_hll_merge":                       stats.ObjectHllMerge,
		"object_hll_merge_time_100":              stats.ObjectHllMergeTime100,
		"object_hll_merge_time_95":               stats.ObjectHllMergeTime95,
		"object_hll_merge_time_99":               stats.ObjectHllMergeTime99,
		"object_hll_merge_time_mean":             stats.ObjectHllMergeTimeMean,
		"object_hll_merge_time_median":           stats.ObjectHllMergeTimeMedian,
		"object_hll_merge_total":                 stats.ObjectHllMergeTotal,
		"object_map_merge":                       stats.ObjectMapMerge,
		"object_map_merge_time_100":              stats.ObjectMapMergeTime100,
		"object_map_merge_time_95":               stats.ObjectMapMergeTime95,
		"object_map_merge_time_99":               stats.ObjectMapMergeTime99,
		"object_map_merge_time_mean":             stats.ObjectMapMergeTimeMean,
		"object_map_merge_time_median":           stats.ObjectMapMergeTimeMedian,
		"object_map_merge_total":                 stats.ObjectMapMergeTotal,
		"object_merge":                           stats.ObjectMerge,
		"object_merge_time_100":                  stats.ObjectMergeTime100,
		"object_merge_time_95":                   stats.ObjectMergeTime95,
		"object_merge_time_99":                   stats.ObjectMergeTime99,
		"object_merge_time_mean":                 stats.ObjectMergeTimeMean,
		"object_merge_time_median":               stats.ObjectMergeTimeMedian,
		"object_merge_total":                     stats.ObjectMergeTotal,
		"object_set_merge":                       stats.ObjectSetMerge,
		"object_set_merge_time_100":              stats.ObjectSetMergeTime100,
		"object_set_merge_time_95":               stats.ObjectSetMergeTime95,
		"object_set_merge_time_99":               stats.ObjectSetMergeTime99,
		"object_set_merge_time_mean":             stats.ObjectSetMergeTimeMean,
		"object_set_merge_time_median":           stats.ObjectSetMergeTimeMedian,
		"object_set_merge_total":                 stats.ObjectSetMergeTotal,
		"pbc_active":                             stats.PbcActive,
		"pbc_connects":                           stats.PbcConnects,
		"pbc_connects_total":                     stats.PbcConnectsTotal,
		"pipeline_active":                        stats.PipelineActive,
		"pipeline_create_count":                  stats.PipelineCreateCount,
		"pipeline_create_error_count":            stats.PipelineCreateErrorCount,
		"pipeline_create_error_one":              stats.PipelineCreateErrorOne,
		"pipeline_create_one":                    stats.PipelineCreateOne,
		"postcommit_fail":                        stats.PostcommitFail,
		"precommit_fail":                         stats.PrecommitFail,
		"read_repairs":                           stats.ReadRepairs,
		"read_repairs_counter":                   stats.ReadRepairsCounter,
		"read_repairs_counter_total":             stats.ReadRepairsCounterTotal,
		"read_repairs_fallback_notfound_count":   stats.ReadRepairsFallbackNotfoundCount,
		"read_repairs_fallback_notfound_one":     stats.ReadRepairsFallbackNotfoundOne,
		"read_repairs_fallback_outofdate_count":  stats.ReadRepairsFallbackOutofdateCount,
		"read_repairs_fallback_outofdate_one":    stats.ReadRepairsFallbackOutofdateOne,
		"read_repairs_hll":                       stats.ReadRepairsHll,
		"read_repairs_hll_total":                 stats.ReadRepairsHllTotal,
		"read_repairs_map":                       stats.ReadRepairsMap,
		"read_repairs_map_total":                 stats.ReadRepairsMapTotal,
		"read_repairs_primary_notfound_count":    stats.ReadRepairsPrimaryNotfoundCount,
		"read_repairs_primary_notfound_one":      stats.ReadRepairsPrimaryNotfoundOne,
		"read_repairs_primary_outofdate_count":   stats.ReadRepairsPrimaryOutofdateCount,
		"read_repairs_primary_outofdate_one":     stats.ReadRepairsPrimaryOutofdateOne,
		"read_repairs_set":                       stats.ReadRepairsSet,
		"read_repairs_set_total":                 stats.ReadRepairsSetTotal,
		"read_repairs_total":                     stats.ReadRepairsTotal,
		"rebalance_delay_last":                   stats.RebalanceDelayLast,
		"rebalance_delay_max":                    stats.RebalanceDelayMax,
		"rebalance_delay_mean":                   stats.RebalanceDelayMean,
		"rebalance_delay_min":                    stats.RebalanceDelayMin,
		"rejected_handoffs":                      stats.RejectedHandoffs,
		"riak_kv_vnodeq_max":                     stats.RiakKvVnodeqMax,
		"riak_kv_vnodeq_mean":                    stats.RiakKvVnodeqMean,
		"riak_kv_vnodeq_median":                  stats.RiakKvVnodeqMedian,
		"riak_kv_vnodeq_min":                     stats.RiakKvVnodeqMin,
		"riak_kv_vnodeq_total":                   stats.RiakKvVnodeqTotal,
		"riak_kv_vnodes_running":                 stats.RiakKvVnodesRunning,
		"riak_pipe_vnodeq_max":                   stats.RiakPipeVnodeqMax,
		"riak_pipe_vnodeq_mean":                  stats.RiakPipeVnodeqMean,
		"riak_pipe_vnodeq_median":                stats.RiakPipeVnodeqMedian,
		"riak_pipe_vnodeq_min":                   stats.RiakPipeVnodeqMin,
		"riak_pipe_vnodeq_total":                 stats.RiakPipeVnodeqTotal,
		"riak_pipe_vnodes_running":               stats.RiakPipeVnodesRunning,
		"ring_creation_size":                     stats.RingCreationSize,
		"ring_members":                           stats.RingMembers,
		"ring_num_partitions":                    stats.RingNumPartitions,
		"ring_ownership":                         stats.RingOwnership,
		"rings_reconciled":                       stats.RingsReconciled,
		"rings_reconciled_total":                 stats.RingsReconciledTotal,
		"set_actor_counts_100":                   stats.SetActorCounts100,
		"set_actor_counts_95":                    stats.SetActorCounts95,
		"set_actor_counts_99":                    stats.SetActorCounts99,
		"set_actor_counts_mean":                  stats.SetActorCountsMean,
		"set_actor_counts_median":                stats.SetActorCountsMedian,
		"skipped_read_repairs":                   stats.SkippedReadRepairs,
		"skipped_read_repairs_total":             stats.SkippedReadRepairsTotal,
		"soft_loaded_vnode_mbox_total":           stats.SoftLoadedVnodeMboxTotal,
		"storage_backend":                        stats.StorageBackend,
		"sys_driver_version":                     stats.SysDriverVersion,
		"sys_global_heaps_size":                  stats.SysGlobalHeapsSize,
		"sys_heap_type":                          stats.SysHeapType,
		"sys_logical_processors":                 stats.SysLogicalProcessors,
		"sys_monitor_count":                      stats.SysMonitorCount,
		"sys_otp_release":                        stats.SysOtpRelease,
		"sys_port_count":                         stats.SysPortCount,
		"sys_process_count":                      stats.SysProcessCount,
		"sys_smp_support":                        stats.SysSmpSupport,
		"sys_system_architecture":                stats.SysSystemArchitecture,
		"sys_system_version":                     stats.SysSystemVersion,
		"sys_thread_pool_size":                   stats.SysThreadPoolSize,
		"sys_threads_enabled":                    stats.SysThreadsEnabled,
		"sys_wordsize":                           stats.SysWordsize,
		"tictacaae_branch_compare":               stats.TictacaaeBranchCompare,
		"tictacaae_branch_compare_total":         stats.TictacaaeBranchCompareTotal,
		"tictacaae_bucket":                       stats.TictacaaeBucket,
		"tictacaae_bucket_total":                 stats.TictacaaeBucketTotal,
		"tictacaae_clock_compare":                stats.TictacaaeClockCompare,
		"tictacaae_clock_compare_total":          stats.TictacaaeClockCompareTotal,
		"tictacaae_error":                        stats.TictacaaeError,
		"tictacaae_error_total":                  stats.TictacaaeErrorTotal,
		"tictacaae_exchange":                     stats.TictacaaeExchange,
		"tictacaae_exchange_total":               stats.TictacaaeExchangeTotal,
		"tictacaae_modtime":                      stats.TictacaaeModtime,
		"tictacaae_modtime_total":                stats.TictacaaeModtimeTotal,
		"tictacaae_not_supported":                stats.TictacaaeNotSupported,
		"tictacaae_not_supported_total":          stats.TictacaaeNotSupportedTotal,
		"tictacaae_queue_microsec__max":          stats.TictacaaeQueueMicrosecMax,
		"tictacaae_queue_microsec_mean":          stats.TictacaaeQueueMicrosecMean,
		"tictacaae_root_compare":                 stats.TictacaaeRootCompare,
		"tictacaae_root_compare_total":           stats.TictacaaeRootCompareTotal,
		"tictacaae_timeout":                      stats.TictacaaeTimeout,
		"tictacaae_timeout_total":                stats.TictacaaeTimeoutTotal,
		"ttaaefs_allcheck_total":                 stats.TtaaefsAllcheckTotal,
		"ttaaefs_daycheck_total":                 stats.TtaaefsDaycheckTotal,
		"ttaaefs_fail_time_100":                  stats.TtaaefsFailTime100,
		"ttaaefs_fail_total":                     stats.TtaaefsFailTotal,
		"ttaaefs_hourcheck_total":                stats.TtaaefsHourcheckTotal,
		"ttaaefs_nosync_time_100":                stats.TtaaefsNosyncTime100,
		"ttaaefs_nosync_total":                   stats.TtaaefsNosyncTotal,
		"ttaaefs_rangecheck_total":               stats.TtaaefsRangecheckTotal,
		"ttaaefs_snk_ahead_total":                stats.TtaaefsSnkAheadTotal,
		"ttaaefs_src_ahead_total":                stats.TtaaefsSrcAheadTotal,
		"ttaaefs_sync_time_100":                  stats.TtaaefsSyncTime100,
		"ttaaefs_sync_total":                     stats.TtaaefsSyncTotal,
		"vnode_counter_update":                   stats.VnodeCounterUpdate,
		"vnode_counter_update_time_100":          stats.VnodeCounterUpdateTime100,
		"vnode_counter_update_time_95":           stats.VnodeCounterUpdateTime95,
		"vnode_counter_update_time_99":           stats.VnodeCounterUpdateTime99,
		"vnode_counter_update_time_mean":         stats.VnodeCounterUpdateTimeMean,
		"vnode_counter_update_time_median":       stats.VnodeCounterUpdateTimeMedian,
		"vnode_counter_update_total":             stats.VnodeCounterUpdateTotal,
		"vnode_get_fsm_time_100":                 stats.VnodeGetFsmTime100,
		"vnode_get_fsm_time_95":                  stats.VnodeGetFsmTime95,
		"vnode_get_fsm_time_99":                  stats.VnodeGetFsmTime99,
		"vnode_get_fsm_time_mean":                stats.VnodeGetFsmTimeMean,
		"vnode_get_fsm_time_median":              stats.VnodeGetFsmTimeMedian,
		"vnode_gets":                             stats.VnodeGets,
		"vnode_gets_total":                       stats.VnodeGetsTotal,
		"vnode_head_fsm_time_100":                stats.VnodeHeadFsmTime100,
		"vnode_head_fsm_time_95":                 stats.VnodeHeadFsmTime95,
		"vnode_head_fsm_time_99":                 stats.VnodeHeadFsmTime99,
		"vnode_head_fsm_time_mean":               stats.VnodeHeadFsmTimeMean,
		"vnode_head_fsm_time_median":             stats.VnodeHeadFsmTimeMedian,
		"vnode_heads":                            stats.VnodeHeads,
		"vnode_heads_total":                      stats.VnodeHeadsTotal,
		"vnode_hll_update":                       stats.VnodeHllUpdate,
		"vnode_hll_update_time_100":              stats.VnodeHllUpdateTime100,
		"vnode_hll_update_time_95":               stats.VnodeHllUpdateTime95,
		"vnode_hll_update_time_99":               stats.VnodeHllUpdateTime99,
		"vnode_hll_update_time_mean":             stats.VnodeHllUpdateTimeMean,
		"vnode_hll_update_time_median":           stats.VnodeHllUpdateTimeMedian,
		"vnode_hll_update_total":                 stats.VnodeHllUpdateTotal,
		"vnode_index_deletes":                    stats.VnodeIndexDeletes,
		"vnode_index_deletes_postings":           stats.VnodeIndexDeletesPostings,
		"vnode_index_deletes_postings_total":     stats.VnodeIndexDeletesPostingsTotal,
		"vnode_index_deletes_total":              stats.VnodeIndexDeletesTotal,
		"vnode_index_reads":                      stats.VnodeIndexReads,
		"vnode_index_reads_total":                stats.VnodeIndexReadsTotal,
		"vnode_index_refreshes":                  stats.VnodeIndexRefreshes,
		"vnode_index_refreshes_total":            stats.VnodeIndexRefreshesTotal,
		"vnode_index_writes":                     stats.VnodeIndexWrites,
		"vnode_index_writes_postings":            stats.VnodeIndexWritesPostings,
		"vnode_index_writes_postings_total":      stats.VnodeIndexWritesPostingsTotal,
		"vnode_index_writes_total":               stats.VnodeIndexWritesTotal,
		"vnode_map_update":                       stats.VnodeMapUpdate,
		"vnode_map_update_time_100":              stats.VnodeMapUpdateTime100,
		"vnode_map_update_time_95":               stats.VnodeMapUpdateTime95,
		"vnode_map_update_time_99":               stats.VnodeMapUpdateTime99,
		"vnode_map_update_time_mean":             stats.VnodeMapUpdateTimeMean,
		"vnode_map_update_time_median":           stats.VnodeMapUpdateTimeMedian,
		"vnode_map_update_total":                 stats.VnodeMapUpdateTotal,
		"vnode_mbox_check_timeout_total":         stats.VnodeMboxCheckTimeoutTotal,
		"vnode_put_fsm_time_100":                 stats.VnodePutFsmTime100,
		"vnode_put_fsm_time_95":                  stats.VnodePutFsmTime95,
		"vnode_put_fsm_time_99":                  stats.VnodePutFsmTime99,
		"vnode_put_fsm_time_mean":                stats.VnodePutFsmTimeMean,
		"vnode_put_fsm_time_median":              stats.VnodePutFsmTimeMedian,
		"vnode_puts":                             stats.VnodePuts,
		"vnode_puts_total":                       stats.VnodePutsTotal,
		"vnode_set_update":                       stats.VnodeSetUpdate,
		"vnode_set_update_time_100":              stats.VnodeSetUpdateTime100,
		"vnode_set_update_time_95":               stats.VnodeSetUpdateTime95,
		"vnode_set_update_time_99":               stats.VnodeSetUpdateTime99,
		"vnode_set_update_time_mean":             stats.VnodeSetUpdateTimeMean,
		"vnode_set_update_time_median":           stats.VnodeSetUpdateTimeMedian,
		"vnode_set_update_total":                 stats.VnodeSetUpdateTotal,
		"worker_af1_pool_queuetime_100":          stats.WorkerAf1PoolQueuetime100,
		"worker_af1_pool_queuetime_mean":         stats.WorkerAf1PoolQueuetimeMean,
		"worker_af1_pool_total":                  stats.WorkerAf1PoolTotal,
		"worker_af1_pool_worktime_100":           stats.WorkerAf1PoolWorktime100,
		"worker_af1_pool_worktime_mean":          stats.WorkerAf1PoolWorktimeMean,
		"worker_af2_pool_queuetime_100":          stats.WorkerAf2PoolQueuetime100,
		"worker_af2_pool_queuetime_mean":         stats.WorkerAf2PoolQueuetimeMean,
		"worker_af2_pool_total":                  stats.WorkerAf2PoolTotal,
		"worker_af2_pool_worktime_100":           stats.WorkerAf2PoolWorktime100,
		"worker_af2_pool_worktime_mean":          stats.WorkerAf2PoolWorktimeMean,
		"worker_af3_pool_queuetime_100":          stats.WorkerAf3PoolQueuetime100,
		"worker_af3_pool_queuetime_mean":         stats.WorkerAf3PoolQueuetimeMean,
		"worker_af3_pool_total":                  stats.WorkerAf3PoolTotal,
		"worker_af3_pool_worktime_100":           stats.WorkerAf3PoolWorktime100,
		"worker_af3_pool_worktime_mean":          stats.WorkerAf3PoolWorktimeMean,
		"worker_af4_pool_queuetime_100":          stats.WorkerAf4PoolQueuetime100,
		"worker_af4_pool_queuetime_mean":         stats.WorkerAf4PoolQueuetimeMean,
		"worker_af4_pool_total":                  stats.WorkerAf4PoolTotal,
		"worker_af4_pool_worktime_100":           stats.WorkerAf4PoolWorktime100,
		"worker_af4_pool_worktime_mean":          stats.WorkerAf4PoolWorktimeMean,
		"worker_be_pool_queuetime_100":           stats.WorkerBePoolQueuetime100,
		"worker_be_pool_queuetime_mean":          stats.WorkerBePoolQueuetimeMean,
		"worker_be_pool_total":                   stats.WorkerBePoolTotal,
		"worker_be_pool_worktime_100":            stats.WorkerBePoolWorktime100,
		"worker_be_pool_worktime_mean":           stats.WorkerBePoolWorktimeMean,
		"worker_node_worker_pool_queuetime_100":  stats.WorkerNodeWorkerPoolQueuetime100,
		"worker_node_worker_pool_queuetime_mean": stats.WorkerNodeWorkerPoolQueuetimeMean,
		"worker_node_worker_pool_total":          stats.WorkerNodeWorkerPoolTotal,
		"worker_node_worker_pool_worktime_100":   stats.WorkerNodeWorkerPoolWorktime100,
		"worker_node_worker_pool_worktime_mean":  stats.WorkerNodeWorkerPoolWorktimeMean,
		"worker_unregistered_queuetime_100":      stats.WorkerUnregisteredQueuetime100,
		"worker_unregistered_queuetime_mean":     stats.WorkerUnregisteredQueuetimeMean,
		"worker_unregistered_total":              stats.WorkerUnregisteredTotal,
		"worker_unregistered_worktime_100":       stats.WorkerUnregisteredWorktime100,
		"worker_unregistered_worktime_mean":      stats.WorkerUnregisteredWorktimeMean,
		"worker_vnode_pool_queuetime_100":        stats.WorkerVnodePoolQueuetime100,
		"worker_vnode_pool_queuetime_mean":       stats.WorkerVnodePoolQueuetimeMean,
		"worker_vnode_pool_total":                stats.WorkerVnodePoolTotal,
		"worker_vnode_pool_worktime_100":         stats.WorkerVnodePoolWorktime100,
		"worker_vnode_pool_worktime_mean":        stats.WorkerVnodePoolWorktimeMean,
		"write_once_merge":                       stats.WriteOnceMerge,
		"write_once_put_objsize_100":             stats.WriteOncePutObjsize100,
		"write_once_put_objsize_95":              stats.WriteOncePutObjsize95,
		"write_once_put_objsize_99":              stats.WriteOncePutObjsize99,
		"write_once_put_objsize_mean":            stats.WriteOncePutObjsizeMean,
		"write_once_put_objsize_median":          stats.WriteOncePutObjsizeMedian,
		"write_once_put_time_100":                stats.WriteOncePutTime100,
		"write_once_put_time_95":                 stats.WriteOncePutTime95,
		"write_once_put_time_99":                 stats.WriteOncePutTime99,
		"write_once_put_time_mean":               stats.WriteOncePutTimeMean,
		"write_once_put_time_median":             stats.WriteOncePutTimeMedian,
		"write_once_puts":                        stats.WriteOncePuts,
		"write_once_puts_total":                  stats.WriteOncePutsTotal,
	}

	if err != nil {
		return fmt.Errorf("unable to build map of field values: %s", err)
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
