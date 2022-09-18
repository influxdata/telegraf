//go:generate ../../../tools/readme_config_includer/generator
package riak

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
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
	ClusterAaeFsmActive      int64 `json:"clusteraae_fsm_active"`
	ClusterAaeFsmCreate      int64 `json:"clusteraae_fsm_create"`
	ClusterAaeFsmCreateError int64 `json:"clusteraae_fsm_create_error"`
	// ConnectedNodes                    []string `json:"connected_nodes"`
	ConsistentGetObjsize100           int64  `json:"consistent_get_objsize_100"`
	ConsistentGetObjsize95            int64  `json:"consistent_get_objsize_95"`
	ConsistentGetObjsize99            int64  `json:"consistent_get_objsize_99"`
	ConsistentGetObjsizeMean          int64  `json:"consistent_get_objsize_mean"`
	ConsistentGetObjsizeMedian        int64  `json:"consistent_get_objsize_median"`
	ConsistentGetTime100              int64  `json:"consistent_get_time_100"`
	ConsistentGetTime95               int64  `json:"consistent_get_time_95"`
	ConsistentGetTime99               int64  `json:"consistent_get_time_99"`
	ConsistentGetTimeMean             int64  `json:"consistent_get_time_mean"`
	ConsistentGetTimeMedian           int64  `json:"consistent_get_time_median"`
	ConsistentGets                    int64  `json:"consistent_gets"`
	ConsistentGetsTotal               int64  `json:"consistent_gets_total"`
	ConsistentPutObjsize100           int64  `json:"consistent_put_objsize_100"`
	ConsistentPutObjsize95            int64  `json:"consistent_put_objsize_95"`
	ConsistentPutObjsize99            int64  `json:"consistent_put_objsize_99"`
	ConsistentPutObjsizeMean          int64  `json:"consistent_put_objsize_mean"`
	ConsistentPutObjsizeMedian        int64  `json:"consistent_put_objsize_median"`
	ConsistentPutTime100              int64  `json:"consistent_put_time_100"`
	ConsistentPutTime95               int64  `json:"consistent_put_time_95"`
	ConsistentPutTime99               int64  `json:"consistent_put_time_99"`
	ConsistentPutTimeMean             int64  `json:"consistent_put_time_mean"`
	ConsistentPutTimeMedian           int64  `json:"consistent_put_time_median"`
	ConsistentPuts                    int64  `json:"consistent_puts"`
	ConsistentPutsTotal               int64  `json:"consistent_puts_total"`
	ConvergeDelayLast                 int64  `json:"converge_delay_last"`
	ConvergeDelayMax                  int64  `json:"converge_delay_max"`
	ConvergeDelayMean                 int64  `json:"converge_delay_mean"`
	ConvergeDelayMin                  int64  `json:"converge_delay_min"`
	CoordLocalSoftLoadedTotal         int64  `json:"coord_local_soft_loaded_total"`
	CoordLocalUnloadedTotal           int64  `json:"coord_local_unloaded_total"`
	CoordRedirLeastLoadedTotal        int64  `json:"coord_redir_least_loaded_total"`
	CoordRedirLoadedLocalTotal        int64  `json:"coord_redir_loaded_local_total"`
	CoordRedirUnloadedTotal           int64  `json:"coord_redir_unloaded_total"`
	CoordRedirsTotal                  int64  `json:"coord_redirs_total"`
	CounterActorCounts100             int64  `json:"counter_actor_counts_100"`
	CounterActorCounts95              int64  `json:"counter_actor_counts_95"`
	CounterActorCounts99              int64  `json:"counter_actor_counts_99"`
	CounterActorCountsMean            int64  `json:"counter_actor_counts_mean"`
	CounterActorCountsMedian          int64  `json:"counter_actor_counts_median"`
	CPUAvg1                           int64  `json:"cpu_avg1"`
	CPUAvg15                          int64  `json:"cpu_avg15"`
	CPUAvg5                           int64  `json:"cpu_avg5"`
	CPUNprocs                         int64  `json:"cpu_nprocs"`
	DroppedVnodeRequestsTotals        int64  `json:"dropped_vnode_requests_totals"`
	ExecutingMappers                  int64  `json:"executing_mappers"`
	GossipReceived                    int64  `json:"gossip_received"`
	HandoffTimeouts                   int64  `json:"handoff_timeouts"`
	HllBytes                          int64  `json:"hll_bytes"`
	HllBytes100                       int64  `json:"hll_bytes_100"`
	HllBytes95                        int64  `json:"hll_bytes_95"`
	HllBytes99                        int64  `json:"hll_bytes_99"`
	HllBytesMean                      int64  `json:"hll_bytes_mean"`
	HllBytesMedian                    int64  `json:"hll_bytes_median"`
	HllBytesTotal                     int64  `json:"hll_bytes_total"`
	IgnoredGossipTotal                int64  `json:"ignored_gossip_total"`
	IndexFsmActive                    int64  `json:"index_fsm_active"`
	IndexFsmComplete                  int64  `json:"index_fsm_complete"`
	IndexFsmCreate                    int64  `json:"index_fsm_create"`
	IndexFsmCreateError               int64  `json:"index_fsm_create_error"`
	IndexFsmResults100                int64  `json:"index_fsm_results_100"`
	IndexFsmResults95                 int64  `json:"index_fsm_results_95"`
	IndexFsmResults99                 int64  `json:"index_fsm_results_99"`
	IndexFsmResultsMean               int64  `json:"index_fsm_results_mean"`
	IndexFsmResultsMedian             int64  `json:"index_fsm_results_median"`
	IndexFsmTime100                   int64  `json:"index_fsm_time_100"`
	IndexFsmTime95                    int64  `json:"index_fsm_time_95"`
	IndexFsmTime99                    int64  `json:"index_fsm_time_99"`
	IndexFsmTimeMean                  int64  `json:"index_fsm_time_mean"`
	IndexFsmTimeMedian                int64  `json:"index_fsm_time_median"`
	LatePutFsmCoordinatorAck          int64  `json:"late_put_fsm_coordinator_ack"`
	LeveldbReadBlockError             int64  `json:"leveldb_read_block_error"`
	ListFsmActive                     int64  `json:"list_fsm_active"`
	ListFsmCreate                     int64  `json:"list_fsm_create"`
	ListFsmCreateError                int64  `json:"list_fsm_create_error"`
	ListFsmCreateErrorTotal           int64  `json:"list_fsm_create_error_total"`
	ListFsmCreateTotal                int64  `json:"list_fsm_create_total"`
	MapActorCounts100                 int64  `json:"map_actor_counts_100"`
	MapActorCounts95                  int64  `json:"map_actor_counts_95"`
	MapActorCounts99                  int64  `json:"map_actor_counts_99"`
	MapActorCountsMean                int64  `json:"map_actor_counts_mean"`
	MapActorCountsMedian              int64  `json:"map_actor_counts_median"`
	MemAllocated                      int64  `json:"mem_allocated"`
	MemTotal                          int64  `json:"mem_total"`
	MemoryAtom                        int64  `json:"memory_atom"`
	MemoryAtomUsed                    int64  `json:"memory_atom_used"`
	MemoryBinary                      int64  `json:"memory_binary"`
	MemoryCode                        int64  `json:"memory_code"`
	MemoryEts                         int64  `json:"memory_ets"`
	MemoryProcesses                   int64  `json:"memory_processes"`
	MemoryProcessesUsed               int64  `json:"memory_processes_used"`
	MemorySystem                      int64  `json:"memory_system"`
	MemoryTotal                       int64  `json:"memory_total"`
	NgrfetchNofetch                   int64  `json:"ngrfetch_nofetch"`
	NgrfetchNofetchTotal              int64  `json:"ngrfetch_nofetch_total"`
	NgrfetchPrefetch                  int64  `json:"ngrfetch_prefetch"`
	NgrfetchPrefetchTotal             int64  `json:"ngrfetch_prefetch_total"`
	NgrfetchTofetch                   int64  `json:"ngrfetch_tofetch"`
	NgrfetchTofetchTotal              int64  `json:"ngrfetch_tofetch_total"`
	NgrreplEmpty                      int64  `json:"ngrrepl_empty"`
	NgrreplEmptyTotal                 int64  `json:"ngrrepl_empty_total"`
	NgrreplError                      int64  `json:"ngrrepl_error"`
	NgrreplErrorTotal                 int64  `json:"ngrrepl_error_total"`
	NgrreplObject                     int64  `json:"ngrrepl_object"`
	NgrreplObjectTotal                int64  `json:"ngrrepl_object_total"`
	NgrreplSrcdiscard                 int64  `json:"ngrrepl_srcdiscard"`
	NgrreplSrcdiscardTotal            int64  `json:"ngrrepl_srcdiscard_total"`
	NodeGetFsmActive                  int64  `json:"node_get_fsm_active"`
	NodeGetFsmActive60s               int64  `json:"node_get_fsm_active_60s"`
	NodeGetFsmCounterObjsize100       int64  `json:"node_get_fsm_counter_objsize_100"`
	NodeGetFsmCounterObjsize95        int64  `json:"node_get_fsm_counter_objsize_95"`
	NodeGetFsmCounterObjsize99        int64  `json:"node_get_fsm_counter_objsize_99"`
	NodeGetFsmCounterObjsizeMean      int64  `json:"node_get_fsm_counter_objsize_mean"`
	NodeGetFsmCounterObjsizeMedian    int64  `json:"node_get_fsm_counter_objsize_median"`
	NodeGetFsmCounterSiblings100      int64  `json:"node_get_fsm_counter_siblings_100"`
	NodeGetFsmCounterSiblings95       int64  `json:"node_get_fsm_counter_siblings_95"`
	NodeGetFsmCounterSiblings99       int64  `json:"node_get_fsm_counter_siblings_99"`
	NodeGetFsmCounterSiblingsMean     int64  `json:"node_get_fsm_counter_siblings_mean"`
	NodeGetFsmCounterSiblingsMedian   int64  `json:"node_get_fsm_counter_siblings_median"`
	NodeGetFsmCounterTime100          int64  `json:"node_get_fsm_counter_time_100"`
	NodeGetFsmCounterTime95           int64  `json:"node_get_fsm_counter_time_95"`
	NodeGetFsmCounterTime99           int64  `json:"node_get_fsm_counter_time_99"`
	NodeGetFsmCounterTimeMean         int64  `json:"node_get_fsm_counter_time_mean"`
	NodeGetFsmCounterTimeMedian       int64  `json:"node_get_fsm_counter_time_median"`
	NodeGetFsmErrors                  int64  `json:"node_get_fsm_errors"`
	NodeGetFsmErrorsTotal             int64  `json:"node_get_fsm_errors_total"`
	NodeGetFsmHllObjsize100           int64  `json:"node_get_fsm_hll_objsize_100"`
	NodeGetFsmHllObjsize95            int64  `json:"node_get_fsm_hll_objsize_95"`
	NodeGetFsmHllObjsize99            int64  `json:"node_get_fsm_hll_objsize_99"`
	NodeGetFsmHllObjsizeMean          int64  `json:"node_get_fsm_hll_objsize_mean"`
	NodeGetFsmHllObjsizeMedian        int64  `json:"node_get_fsm_hll_objsize_median"`
	NodeGetFsmHllSiblings100          int64  `json:"node_get_fsm_hll_siblings_100"`
	NodeGetFsmHllSiblings95           int64  `json:"node_get_fsm_hll_siblings_95"`
	NodeGetFsmHllSiblings99           int64  `json:"node_get_fsm_hll_siblings_99"`
	NodeGetFsmHllSiblingsMean         int64  `json:"node_get_fsm_hll_siblings_mean"`
	NodeGetFsmHllSiblingsMedian       int64  `json:"node_get_fsm_hll_siblings_median"`
	NodeGetFsmHllTime100              int64  `json:"node_get_fsm_hll_time_100"`
	NodeGetFsmHllTime95               int64  `json:"node_get_fsm_hll_time_95"`
	NodeGetFsmHllTime99               int64  `json:"node_get_fsm_hll_time_99"`
	NodeGetFsmHllTimeMean             int64  `json:"node_get_fsm_hll_time_mean"`
	NodeGetFsmHllTimeMedian           int64  `json:"node_get_fsm_hll_time_median"`
	NodeGetFsmInRate                  int64  `json:"node_get_fsm_in_rate"`
	NodeGetFsmMapObjsize100           int64  `json:"node_get_fsm_map_objsize_100"`
	NodeGetFsmMapObjsize95            int64  `json:"node_get_fsm_map_objsize_95"`
	NodeGetFsmMapObjsize99            int64  `json:"node_get_fsm_map_objsize_99"`
	NodeGetFsmMapObjsizeMean          int64  `json:"node_get_fsm_map_objsize_mean"`
	NodeGetFsmMapObjsizeMedian        int64  `json:"node_get_fsm_map_objsize_median"`
	NodeGetFsmMapSiblings100          int64  `json:"node_get_fsm_map_siblings_100"`
	NodeGetFsmMapSiblings95           int64  `json:"node_get_fsm_map_siblings_95"`
	NodeGetFsmMapSiblings99           int64  `json:"node_get_fsm_map_siblings_99"`
	NodeGetFsmMapSiblingsMean         int64  `json:"node_get_fsm_map_siblings_mean"`
	NodeGetFsmMapSiblingsMedian       int64  `json:"node_get_fsm_map_siblings_median"`
	NodeGetFsmMapTime100              int64  `json:"node_get_fsm_map_time_100"`
	NodeGetFsmMapTime95               int64  `json:"node_get_fsm_map_time_95"`
	NodeGetFsmMapTime99               int64  `json:"node_get_fsm_map_time_99"`
	NodeGetFsmMapTimeMean             int64  `json:"node_get_fsm_map_time_mean"`
	NodeGetFsmMapTimeMedian           int64  `json:"node_get_fsm_map_time_median"`
	NodeGetFsmObjsize100              int64  `json:"node_get_fsm_objsize_100"`
	NodeGetFsmObjsize95               int64  `json:"node_get_fsm_objsize_95"`
	NodeGetFsmObjsize99               int64  `json:"node_get_fsm_objsize_99"`
	NodeGetFsmObjsizeMean             int64  `json:"node_get_fsm_objsize_mean"`
	NodeGetFsmObjsizeMedian           int64  `json:"node_get_fsm_objsize_median"`
	NodeGetFsmOutRate                 int64  `json:"node_get_fsm_out_rate"`
	NodeGetFsmRejected                int64  `json:"node_get_fsm_rejected"`
	NodeGetFsmRejected60s             int64  `json:"node_get_fsm_rejected_60s"`
	NodeGetFsmRejectedTotal           int64  `json:"node_get_fsm_rejected_total"`
	NodeGetFsmSetObjsize100           int64  `json:"node_get_fsm_set_objsize_100"`
	NodeGetFsmSetObjsize95            int64  `json:"node_get_fsm_set_objsize_95"`
	NodeGetFsmSetObjsize99            int64  `json:"node_get_fsm_set_objsize_99"`
	NodeGetFsmSetObjsizeMean          int64  `json:"node_get_fsm_set_objsize_mean"`
	NodeGetFsmSetObjsizeMedian        int64  `json:"node_get_fsm_set_objsize_median"`
	NodeGetFsmSetSiblings100          int64  `json:"node_get_fsm_set_siblings_100"`
	NodeGetFsmSetSiblings95           int64  `json:"node_get_fsm_set_siblings_95"`
	NodeGetFsmSetSiblings99           int64  `json:"node_get_fsm_set_siblings_99"`
	NodeGetFsmSetSiblingsMean         int64  `json:"node_get_fsm_set_siblings_mean"`
	NodeGetFsmSetSiblingsMedian       int64  `json:"node_get_fsm_set_siblings_median"`
	NodeGetFsmSetTime100              int64  `json:"node_get_fsm_set_time_100"`
	NodeGetFsmSetTime95               int64  `json:"node_get_fsm_set_time_95"`
	NodeGetFsmSetTime99               int64  `json:"node_get_fsm_set_time_99"`
	NodeGetFsmSetTimeMean             int64  `json:"node_get_fsm_set_time_mean"`
	NodeGetFsmSetTimeMedian           int64  `json:"node_get_fsm_set_time_median"`
	NodeGetFsmSiblings100             int64  `json:"node_get_fsm_siblings_100"`
	NodeGetFsmSiblings95              int64  `json:"node_get_fsm_siblings_95"`
	NodeGetFsmSiblings99              int64  `json:"node_get_fsm_siblings_99"`
	NodeGetFsmSiblingsMean            int64  `json:"node_get_fsm_siblings_mean"`
	NodeGetFsmSiblingsMedian          int64  `json:"node_get_fsm_siblings_median"`
	NodeGetFsmTime100                 int64  `json:"node_get_fsm_time_100"`
	NodeGetFsmTime95                  int64  `json:"node_get_fsm_time_95"`
	NodeGetFsmTime99                  int64  `json:"node_get_fsm_time_99"`
	NodeGetFsmTimeMean                int64  `json:"node_get_fsm_time_mean"`
	NodeGetFsmTimeMedian              int64  `json:"node_get_fsm_time_median"`
	NodeGets                          int64  `json:"node_gets"`
	NodeGetsCounter                   int64  `json:"node_gets_counter"`
	NodeGetsCounterTotal              int64  `json:"node_gets_counter_total"`
	NodeGetsHll                       int64  `json:"node_gets_hll"`
	NodeGetsHllTotal                  int64  `json:"node_gets_hll_total"`
	NodeGetsMap                       int64  `json:"node_gets_map"`
	NodeGetsMapTotal                  int64  `json:"node_gets_map_total"`
	NodeGetsSet                       int64  `json:"node_gets_set"`
	NodeGetsSetTotal                  int64  `json:"node_gets_set_total"`
	NodeGetsTotal                     int64  `json:"node_gets_total"`
	NodePutFsmActive                  int64  `json:"node_put_fsm_active"`
	NodePutFsmActive60s               int64  `json:"node_put_fsm_active_60s"`
	NodePutFsmCounterTime100          int64  `json:"node_put_fsm_counter_time_100"`
	NodePutFsmCounterTime95           int64  `json:"node_put_fsm_counter_time_95"`
	NodePutFsmCounterTime99           int64  `json:"node_put_fsm_counter_time_99"`
	NodePutFsmCounterTimeMean         int64  `json:"node_put_fsm_counter_time_mean"`
	NodePutFsmCounterTimeMedian       int64  `json:"node_put_fsm_counter_time_median"`
	NodePutFsmHllTime100              int64  `json:"node_put_fsm_hll_time_100"`
	NodePutFsmHllTime95               int64  `json:"node_put_fsm_hll_time_95"`
	NodePutFsmHllTime99               int64  `json:"node_put_fsm_hll_time_99"`
	NodePutFsmHllTimeMean             int64  `json:"node_put_fsm_hll_time_mean"`
	NodePutFsmHllTimeMedian           int64  `json:"node_put_fsm_hll_time_median"`
	NodePutFsmInRate                  int64  `json:"node_put_fsm_in_rate"`
	NodePutFsmMapTime100              int64  `json:"node_put_fsm_map_time_100"`
	NodePutFsmMapTime95               int64  `json:"node_put_fsm_map_time_95"`
	NodePutFsmMapTime99               int64  `json:"node_put_fsm_map_time_99"`
	NodePutFsmMapTimeMean             int64  `json:"node_put_fsm_map_time_mean"`
	NodePutFsmMapTimeMedian           int64  `json:"node_put_fsm_map_time_median"`
	NodePutFsmOutRate                 int64  `json:"node_put_fsm_out_rate"`
	NodePutFsmRejected                int64  `json:"node_put_fsm_rejected"`
	NodePutFsmRejected60s             int64  `json:"node_put_fsm_rejected_60s"`
	NodePutFsmRejectedTotal           int64  `json:"node_put_fsm_rejected_total"`
	NodePutFsmSetTime100              int64  `json:"node_put_fsm_set_time_100"`
	NodePutFsmSetTime95               int64  `json:"node_put_fsm_set_time_95"`
	NodePutFsmSetTime99               int64  `json:"node_put_fsm_set_time_99"`
	NodePutFsmSetTimeMean             int64  `json:"node_put_fsm_set_time_mean"`
	NodePutFsmSetTimeMedian           int64  `json:"node_put_fsm_set_time_median"`
	NodePutFsmTime100                 int64  `json:"node_put_fsm_time_100"`
	NodePutFsmTime95                  int64  `json:"node_put_fsm_time_95"`
	NodePutFsmTime99                  int64  `json:"node_put_fsm_time_99"`
	NodePutFsmTimeMean                int64  `json:"node_put_fsm_time_mean"`
	NodePutFsmTimeMedian              int64  `json:"node_put_fsm_time_median"`
	NodePuts                          int64  `json:"node_puts"`
	NodePutsCounter                   int64  `json:"node_puts_counter"`
	NodePutsCounterTotal              int64  `json:"node_puts_counter_total"`
	NodePutsHll                       int64  `json:"node_puts_hll"`
	NodePutsHllTotal                  int64  `json:"node_puts_hll_total"`
	NodePutsMap                       int64  `json:"node_puts_map"`
	NodePutsMapTotal                  int64  `json:"node_puts_map_total"`
	NodePutsSet                       int64  `json:"node_puts_set"`
	NodePutsSetTotal                  int64  `json:"node_puts_set_total"`
	NodePutsTotal                     int64  `json:"node_puts_total"`
	Nodename                          string `json:"nodename"`
	ObjectCounterMerge                int64  `json:"object_counter_merge"`
	ObjectCounterMergeTime100         int64  `json:"object_counter_merge_time_100"`
	ObjectCounterMergeTime95          int64  `json:"object_counter_merge_time_95"`
	ObjectCounterMergeTime99          int64  `json:"object_counter_merge_time_99"`
	ObjectCounterMergeTimeMean        int64  `json:"object_counter_merge_time_mean"`
	ObjectCounterMergeTimeMedian      int64  `json:"object_counter_merge_time_median"`
	ObjectCounterMergeTotal           int64  `json:"object_counter_merge_total"`
	ObjectHllMerge                    int64  `json:"object_hll_merge"`
	ObjectHllMergeTime100             int64  `json:"object_hll_merge_time_100"`
	ObjectHllMergeTime95              int64  `json:"object_hll_merge_time_95"`
	ObjectHllMergeTime99              int64  `json:"object_hll_merge_time_99"`
	ObjectHllMergeTimeMean            int64  `json:"object_hll_merge_time_mean"`
	ObjectHllMergeTimeMedian          int64  `json:"object_hll_merge_time_median"`
	ObjectHllMergeTotal               int64  `json:"object_hll_merge_total"`
	ObjectMapMerge                    int64  `json:"object_map_merge"`
	ObjectMapMergeTime100             int64  `json:"object_map_merge_time_100"`
	ObjectMapMergeTime95              int64  `json:"object_map_merge_time_95"`
	ObjectMapMergeTime99              int64  `json:"object_map_merge_time_99"`
	ObjectMapMergeTimeMean            int64  `json:"object_map_merge_time_mean"`
	ObjectMapMergeTimeMedian          int64  `json:"object_map_merge_time_median"`
	ObjectMapMergeTotal               int64  `json:"object_map_merge_total"`
	ObjectMerge                       int64  `json:"object_merge"`
	ObjectMergeTime100                int64  `json:"object_merge_time_100"`
	ObjectMergeTime95                 int64  `json:"object_merge_time_95"`
	ObjectMergeTime99                 int64  `json:"object_merge_time_99"`
	ObjectMergeTimeMean               int64  `json:"object_merge_time_mean"`
	ObjectMergeTimeMedian             int64  `json:"object_merge_time_median"`
	ObjectMergeTotal                  int64  `json:"object_merge_total"`
	ObjectSetMerge                    int64  `json:"object_set_merge"`
	ObjectSetMergeTime100             int64  `json:"object_set_merge_time_100"`
	ObjectSetMergeTime95              int64  `json:"object_set_merge_time_95"`
	ObjectSetMergeTime99              int64  `json:"object_set_merge_time_99"`
	ObjectSetMergeTimeMean            int64  `json:"object_set_merge_time_mean"`
	ObjectSetMergeTimeMedian          int64  `json:"object_set_merge_time_median"`
	ObjectSetMergeTotal               int64  `json:"object_set_merge_total"`
	PbcActive                         int64  `json:"pbc_active"`
	PbcConnects                       int64  `json:"pbc_connects"`
	PbcConnectsTotal                  int64  `json:"pbc_connects_total"`
	PipelineActive                    int64  `json:"pipeline_active"`
	PipelineCreateCount               int64  `json:"pipeline_create_count"`
	PipelineCreateErrorCount          int64  `json:"pipeline_create_error_count"`
	PipelineCreateErrorOne            int64  `json:"pipeline_create_error_one"`
	PipelineCreateOne                 int64  `json:"pipeline_create_one"`
	PostcommitFail                    int64  `json:"postcommit_fail"`
	PrecommitFail                     int64  `json:"precommit_fail"`
	ReadRepairs                       int64  `json:"read_repairs"`
	ReadRepairsCounter                int64  `json:"read_repairs_counter"`
	ReadRepairsCounterTotal           int64  `json:"read_repairs_counter_total"`
	ReadRepairsFallbackNotfoundCount  int64  `json:"read_repairs_fallback_notfound_count"`
	ReadRepairsFallbackNotfoundOne    int64  `json:"read_repairs_fallback_notfound_one"`
	ReadRepairsFallbackOutofdateCount int64  `json:"read_repairs_fallback_outofdate_count"`
	ReadRepairsFallbackOutofdateOne   int64  `json:"read_repairs_fallback_outofdate_one"`
	ReadRepairsHll                    int64  `json:"read_repairs_hll"`
	ReadRepairsHllTotal               int64  `json:"read_repairs_hll_total"`
	ReadRepairsMap                    int64  `json:"read_repairs_map"`
	ReadRepairsMapTotal               int64  `json:"read_repairs_map_total"`
	ReadRepairsPrimaryNotfoundCount   int64  `json:"read_repairs_primary_notfound_count"`
	ReadRepairsPrimaryNotfoundOne     int64  `json:"read_repairs_primary_notfound_one"`
	ReadRepairsPrimaryOutofdateCount  int64  `json:"read_repairs_primary_outofdate_count"`
	ReadRepairsPrimaryOutofdateOne    int64  `json:"read_repairs_primary_outofdate_one"`
	ReadRepairsSet                    int64  `json:"read_repairs_set"`
	ReadRepairsSetTotal               int64  `json:"read_repairs_set_total"`
	ReadRepairsTotal                  int64  `json:"read_repairs_total"`
	RebalanceDelayLast                int64  `json:"rebalance_delay_last"`
	RebalanceDelayMax                 int64  `json:"rebalance_delay_max"`
	RebalanceDelayMean                int64  `json:"rebalance_delay_mean"`
	RebalanceDelayMin                 int64  `json:"rebalance_delay_min"`
	RejectedHandoffs                  int64  `json:"rejected_handoffs"`
	RiakKvVnodeqMax                   int64  `json:"riak_kv_vnodeq_max"`
	RiakKvVnodeqMean                  int64  `json:"riak_kv_vnodeq_mean"`
	RiakKvVnodeqMedian                int64  `json:"riak_kv_vnodeq_median"`
	RiakKvVnodeqMin                   int64  `json:"riak_kv_vnodeq_min"`
	RiakKvVnodeqTotal                 int64  `json:"riak_kv_vnodeq_total"`
	RiakKvVnodesRunning               int64  `json:"riak_kv_vnodes_running"`
	RiakPipeVnodeqMax                 int64  `json:"riak_pipe_vnodeq_max"`
	RiakPipeVnodeqMean                int64  `json:"riak_pipe_vnodeq_mean"`
	RiakPipeVnodeqMedian              int64  `json:"riak_pipe_vnodeq_median"`
	RiakPipeVnodeqMin                 int64  `json:"riak_pipe_vnodeq_min"`
	RiakPipeVnodeqTotal               int64  `json:"riak_pipe_vnodeq_total"`
	RiakPipeVnodesRunning             int64  `json:"riak_pipe_vnodes_running"`
	RingCreationSize                  int64  `json:"ring_creation_size"`
	// RingMembers                       []string `json:"ring_members"`
	RingNumPartitions int64 `json:"ring_num_partitions"`
	// RingOwnership                     string   `json:"ring_ownership"`
	RingsReconciled          int64 `json:"rings_reconciled"`
	RingsReconciledTotal     int64 `json:"rings_reconciled_total"`
	SetActorCounts100        int64 `json:"set_actor_counts_100"`
	SetActorCounts95         int64 `json:"set_actor_counts_95"`
	SetActorCounts99         int64 `json:"set_actor_counts_99"`
	SetActorCountsMean       int64 `json:"set_actor_counts_mean"`
	SetActorCountsMedian     int64 `json:"set_actor_counts_median"`
	SkippedReadRepairs       int64 `json:"skipped_read_repairs"`
	SkippedReadRepairsTotal  int64 `json:"skipped_read_repairs_total"`
	SoftLoadedVnodeMboxTotal int64 `json:"soft_loaded_vnode_mbox_total"`
	// StorageBackend                    string   `json:"storage_backend"`
	// SysDriverVersion                  string   `json:"sys_driver_version"`
	// SysGlobalHeapsSize                string   `json:"sys_global_heaps_size"`
	// SysHeapType                       string   `json:"sys_heap_type"`
	SysLogicalProcessors int64  `json:"sys_logical_processors"`
	SysMonitorCount      int64  `json:"sys_monitor_count"`
	SysOtpRelease        string `json:"sys_otp_release"`
	SysPortCount         int64  `json:"sys_port_count"`
	SysProcessCount      int64  `json:"sys_process_count"`
	// SysSmpSupport                     int64    `json:"sys_smp_support"`
	// SysSystemArchitecture             string   `json:"sys_system_architecture"`
	// SysSystemVersion                  string   `json:"sys_system_version"`
	SysThreadPoolSize int64 `json:"sys_thread_pool_size"`
	// SysThreadsEnabled                 bool     `json:"sys_threads_enabled"`
	SysWordsize                       int64 `json:"sys_wordsize"`
	TictacaaeBranchCompare            int64 `json:"tictacaae_branch_compare"`
	TictacaaeBranchCompareTotal       int64 `json:"tictacaae_branch_compare_total"`
	TictacaaeBucket                   int64 `json:"tictacaae_bucket"`
	TictacaaeBucketTotal              int64 `json:"tictacaae_bucket_total"`
	TictacaaeClockCompare             int64 `json:"tictacaae_clock_compare"`
	TictacaaeClockCompareTotal        int64 `json:"tictacaae_clock_compare_total"`
	TictacaaeError                    int64 `json:"tictacaae_error"`
	TictacaaeErrorTotal               int64 `json:"tictacaae_error_total"`
	TictacaaeExchange                 int64 `json:"tictacaae_exchange"`
	TictacaaeExchangeTotal            int64 `json:"tictacaae_exchange_total"`
	TictacaaeModtime                  int64 `json:"tictacaae_modtime"`
	TictacaaeModtimeTotal             int64 `json:"tictacaae_modtime_total"`
	TictacaaeNotSupported             int64 `json:"tictacaae_not_supported"`
	TictacaaeNotSupportedTotal        int64 `json:"tictacaae_not_supported_total"`
	TictacaaeQueueMicrosecMax         int64 `json:"tictacaae_queue_microsec__max"`
	TictacaaeQueueMicrosecMean        int64 `json:"tictacaae_queue_microsec_mean"`
	TictacaaeRootCompare              int64 `json:"tictacaae_root_compare"`
	TictacaaeRootCompareTotal         int64 `json:"tictacaae_root_compare_total"`
	TictacaaeTimeout                  int64 `json:"tictacaae_timeout"`
	TictacaaeTimeoutTotal             int64 `json:"tictacaae_timeout_total"`
	TtaaefsAllcheckTotal              int64 `json:"ttaaefs_allcheck_total"`
	TtaaefsDaycheckTotal              int64 `json:"ttaaefs_daycheck_total"`
	TtaaefsFailTime100                int64 `json:"ttaaefs_fail_time_100"`
	TtaaefsFailTotal                  int64 `json:"ttaaefs_fail_total"`
	TtaaefsHourcheckTotal             int64 `json:"ttaaefs_hourcheck_total"`
	TtaaefsNosyncTime100              int64 `json:"ttaaefs_nosync_time_100"`
	TtaaefsNosyncTotal                int64 `json:"ttaaefs_nosync_total"`
	TtaaefsRangecheckTotal            int64 `json:"ttaaefs_rangecheck_total"`
	TtaaefsSnkAheadTotal              int64 `json:"ttaaefs_snk_ahead_total"`
	TtaaefsSrcAheadTotal              int64 `json:"ttaaefs_src_ahead_total"`
	TtaaefsSyncTime100                int64 `json:"ttaaefs_sync_time_100"`
	TtaaefsSyncTotal                  int64 `json:"ttaaefs_sync_total"`
	VnodeCounterUpdate                int64 `json:"vnode_counter_update"`
	VnodeCounterUpdateTime100         int64 `json:"vnode_counter_update_time_100"`
	VnodeCounterUpdateTime95          int64 `json:"vnode_counter_update_time_95"`
	VnodeCounterUpdateTime99          int64 `json:"vnode_counter_update_time_99"`
	VnodeCounterUpdateTimeMean        int64 `json:"vnode_counter_update_time_mean"`
	VnodeCounterUpdateTimeMedian      int64 `json:"vnode_counter_update_time_median"`
	VnodeCounterUpdateTotal           int64 `json:"vnode_counter_update_total"`
	VnodeGetFsmTime100                int64 `json:"vnode_get_fsm_time_100"`
	VnodeGetFsmTime95                 int64 `json:"vnode_get_fsm_time_95"`
	VnodeGetFsmTime99                 int64 `json:"vnode_get_fsm_time_99"`
	VnodeGetFsmTimeMean               int64 `json:"vnode_get_fsm_time_mean"`
	VnodeGetFsmTimeMedian             int64 `json:"vnode_get_fsm_time_median"`
	VnodeGets                         int64 `json:"vnode_gets"`
	VnodeGetsTotal                    int64 `json:"vnode_gets_total"`
	VnodeHeadFsmTime100               int64 `json:"vnode_head_fsm_time_100"`
	VnodeHeadFsmTime95                int64 `json:"vnode_head_fsm_time_95"`
	VnodeHeadFsmTime99                int64 `json:"vnode_head_fsm_time_99"`
	VnodeHeadFsmTimeMean              int64 `json:"vnode_head_fsm_time_mean"`
	VnodeHeadFsmTimeMedian            int64 `json:"vnode_head_fsm_time_median"`
	VnodeHeads                        int64 `json:"vnode_heads"`
	VnodeHeadsTotal                   int64 `json:"vnode_heads_total"`
	VnodeHllUpdate                    int64 `json:"vnode_hll_update"`
	VnodeHllUpdateTime100             int64 `json:"vnode_hll_update_time_100"`
	VnodeHllUpdateTime95              int64 `json:"vnode_hll_update_time_95"`
	VnodeHllUpdateTime99              int64 `json:"vnode_hll_update_time_99"`
	VnodeHllUpdateTimeMean            int64 `json:"vnode_hll_update_time_mean"`
	VnodeHllUpdateTimeMedian          int64 `json:"vnode_hll_update_time_median"`
	VnodeHllUpdateTotal               int64 `json:"vnode_hll_update_total"`
	VnodeIndexDeletes                 int64 `json:"vnode_index_deletes"`
	VnodeIndexDeletesPostings         int64 `json:"vnode_index_deletes_postings"`
	VnodeIndexDeletesPostingsTotal    int64 `json:"vnode_index_deletes_postings_total"`
	VnodeIndexDeletesTotal            int64 `json:"vnode_index_deletes_total"`
	VnodeIndexReads                   int64 `json:"vnode_index_reads"`
	VnodeIndexReadsTotal              int64 `json:"vnode_index_reads_total"`
	VnodeIndexRefreshes               int64 `json:"vnode_index_refreshes"`
	VnodeIndexRefreshesTotal          int64 `json:"vnode_index_refreshes_total"`
	VnodeIndexWrites                  int64 `json:"vnode_index_writes"`
	VnodeIndexWritesPostings          int64 `json:"vnode_index_writes_postings"`
	VnodeIndexWritesPostingsTotal     int64 `json:"vnode_index_writes_postings_total"`
	VnodeIndexWritesTotal             int64 `json:"vnode_index_writes_total"`
	VnodeMapUpdate                    int64 `json:"vnode_map_update"`
	VnodeMapUpdateTime100             int64 `json:"vnode_map_update_time_100"`
	VnodeMapUpdateTime95              int64 `json:"vnode_map_update_time_95"`
	VnodeMapUpdateTime99              int64 `json:"vnode_map_update_time_99"`
	VnodeMapUpdateTimeMean            int64 `json:"vnode_map_update_time_mean"`
	VnodeMapUpdateTimeMedian          int64 `json:"vnode_map_update_time_median"`
	VnodeMapUpdateTotal               int64 `json:"vnode_map_update_total"`
	VnodeMboxCheckTimeoutTotal        int64 `json:"vnode_mbox_check_timeout_total"`
	VnodePutFsmTime100                int64 `json:"vnode_put_fsm_time_100"`
	VnodePutFsmTime95                 int64 `json:"vnode_put_fsm_time_95"`
	VnodePutFsmTime99                 int64 `json:"vnode_put_fsm_time_99"`
	VnodePutFsmTimeMean               int64 `json:"vnode_put_fsm_time_mean"`
	VnodePutFsmTimeMedian             int64 `json:"vnode_put_fsm_time_median"`
	VnodePuts                         int64 `json:"vnode_puts"`
	VnodePutsTotal                    int64 `json:"vnode_puts_total"`
	VnodeSetUpdate                    int64 `json:"vnode_set_update"`
	VnodeSetUpdateTime100             int64 `json:"vnode_set_update_time_100"`
	VnodeSetUpdateTime95              int64 `json:"vnode_set_update_time_95"`
	VnodeSetUpdateTime99              int64 `json:"vnode_set_update_time_99"`
	VnodeSetUpdateTimeMean            int64 `json:"vnode_set_update_time_mean"`
	VnodeSetUpdateTimeMedian          int64 `json:"vnode_set_update_time_median"`
	VnodeSetUpdateTotal               int64 `json:"vnode_set_update_total"`
	WorkerAf1PoolQueuetime100         int64 `json:"worker_af1_pool_queuetime_100"`
	WorkerAf1PoolQueuetimeMean        int64 `json:"worker_af1_pool_queuetime_mean"`
	WorkerAf1PoolTotal                int64 `json:"worker_af1_pool_total"`
	WorkerAf1PoolWorktime100          int64 `json:"worker_af1_pool_worktime_100"`
	WorkerAf1PoolWorktimeMean         int64 `json:"worker_af1_pool_worktime_mean"`
	WorkerAf2PoolQueuetime100         int64 `json:"worker_af2_pool_queuetime_100"`
	WorkerAf2PoolQueuetimeMean        int64 `json:"worker_af2_pool_queuetime_mean"`
	WorkerAf2PoolTotal                int64 `json:"worker_af2_pool_total"`
	WorkerAf2PoolWorktime100          int64 `json:"worker_af2_pool_worktime_100"`
	WorkerAf2PoolWorktimeMean         int64 `json:"worker_af2_pool_worktime_mean"`
	WorkerAf3PoolQueuetime100         int64 `json:"worker_af3_pool_queuetime_100"`
	WorkerAf3PoolQueuetimeMean        int64 `json:"worker_af3_pool_queuetime_mean"`
	WorkerAf3PoolTotal                int64 `json:"worker_af3_pool_total"`
	WorkerAf3PoolWorktime100          int64 `json:"worker_af3_pool_worktime_100"`
	WorkerAf3PoolWorktimeMean         int64 `json:"worker_af3_pool_worktime_mean"`
	WorkerAf4PoolQueuetime100         int64 `json:"worker_af4_pool_queuetime_100"`
	WorkerAf4PoolQueuetimeMean        int64 `json:"worker_af4_pool_queuetime_mean"`
	WorkerAf4PoolTotal                int64 `json:"worker_af4_pool_total"`
	WorkerAf4PoolWorktime100          int64 `json:"worker_af4_pool_worktime_100"`
	WorkerAf4PoolWorktimeMean         int64 `json:"worker_af4_pool_worktime_mean"`
	WorkerBePoolQueuetime100          int64 `json:"worker_be_pool_queuetime_100"`
	WorkerBePoolQueuetimeMean         int64 `json:"worker_be_pool_queuetime_mean"`
	WorkerBePoolTotal                 int64 `json:"worker_be_pool_total"`
	WorkerBePoolWorktime100           int64 `json:"worker_be_pool_worktime_100"`
	WorkerBePoolWorktimeMean          int64 `json:"worker_be_pool_worktime_mean"`
	WorkerNodeWorkerPoolQueuetime100  int64 `json:"worker_node_worker_pool_queuetime_100"`
	WorkerNodeWorkerPoolQueuetimeMean int64 `json:"worker_node_worker_pool_queuetime_mean"`
	WorkerNodeWorkerPoolTotal         int64 `json:"worker_node_worker_pool_total"`
	WorkerNodeWorkerPoolWorktime100   int64 `json:"worker_node_worker_pool_worktime_100"`
	WorkerNodeWorkerPoolWorktimeMean  int64 `json:"worker_node_worker_pool_worktime_mean"`
	WorkerUnregisteredQueuetime100    int64 `json:"worker_unregistered_queuetime_100"`
	WorkerUnregisteredQueuetimeMean   int64 `json:"worker_unregistered_queuetime_mean"`
	WorkerUnregisteredTotal           int64 `json:"worker_unregistered_total"`
	WorkerUnregisteredWorktime100     int64 `json:"worker_unregistered_worktime_100"`
	WorkerUnregisteredWorktimeMean    int64 `json:"worker_unregistered_worktime_mean"`
	WorkerVnodePoolQueuetime100       int64 `json:"worker_vnode_pool_queuetime_100"`
	WorkerVnodePoolQueuetimeMean      int64 `json:"worker_vnode_pool_queuetime_mean"`
	WorkerVnodePoolTotal              int64 `json:"worker_vnode_pool_total"`
	WorkerVnodePoolWorktime100        int64 `json:"worker_vnode_pool_worktime_100"`
	WorkerVnodePoolWorktimeMean       int64 `json:"worker_vnode_pool_worktime_mean"`
	WriteOnceMerge                    int64 `json:"write_once_merge"`
	WriteOncePutObjsize100            int64 `json:"write_once_put_objsize_100"`
	WriteOncePutObjsize95             int64 `json:"write_once_put_objsize_95"`
	WriteOncePutObjsize99             int64 `json:"write_once_put_objsize_99"`
	WriteOncePutObjsizeMean           int64 `json:"write_once_put_objsize_mean"`
	WriteOncePutObjsizeMedian         int64 `json:"write_once_put_objsize_median"`
	WriteOncePutTime100               int64 `json:"write_once_put_time_100"`
	WriteOncePutTime95                int64 `json:"write_once_put_time_95"`
	WriteOncePutTime99                int64 `json:"write_once_put_time_99"`
	WriteOncePutTimeMean              int64 `json:"write_once_put_time_mean"`
	WriteOncePutTimeMedian            int64 `json:"write_once_put_time_median"`
	WriteOncePuts                     int64 `json:"write_once_puts"`
	WriteOncePutsTotal                int64 `json:"write_once_puts_total"`
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
	fields := make(map[string]interface{})
	rs := reflect.TypeOf((*riakStats)(nil)).Elem()
	st := reflect.ValueOf(stats)
	for i := 0; i < rs.NumField(); i++ {
		key, err := strconv.Unquote(strings.TrimPrefix(string(rs.Field(i).Tag), "json:"))
		if err != nil {
			return fmt.Errorf("unable to build map of field values: %s", err)
		}
		value := reflect.Indirect(st).FieldByName(rs.Field(i).Name)
		fields[key] = value.Interface()
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
