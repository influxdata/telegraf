package couchbase

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"time"

	couchbase "github.com/couchbase/go-couchbase"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Couchbase struct {
	Servers []string

	BucketMetricType string `toml:"bucket_metric_type"`

	FieldsIncluded []string `toml:"fields_included"`

	FieldsExcluded []string `toml:"fields_excluded"`

	includeExclude filter.Filter
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    http://couchbase-0.example.com/
  ##    http://admin:secret@couchbase-0.example.com:8091/
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no protocol is specified, HTTP is used.
  ## If no port is specified, 8091 is used.
  servers = ["http://localhost:8091"]

  ## Use "basic" for a limited number of basic bucket stats. Use "detailed" for a more comprehensive list of all bucket stats.
  bucket_metric_type = "basic"

  ## Filter fields to include only here.
  # fields_included = []

  ## Filter fields to exclude only here.
  # fields_excluded = []
`

var regexpURI = regexp.MustCompile(`(\S+://)?(\S+\:\S+@)`)
var client = &http.Client{Timeout: 10 * time.Second}

func (cb *Couchbase) SampleConfig() string {
	return sampleConfig
}

func (cb *Couchbase) Description() string {
	return "Read metrics from one or many couchbase clusters"
}

// Reads stats from all configured clusters. Accumulates stats.
// Returns one of the errors encountered while gathering stats (if any).
func (cb *Couchbase) Gather(acc telegraf.Accumulator) error {
	if len(cb.Servers) == 0 {
		return cb.gatherServer("http://localhost:8091/", acc, nil)
	}

	var wg sync.WaitGroup
	for _, serv := range cb.Servers {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			acc.AddError(cb.gatherServer(serv, acc, nil))
		}(serv)
	}

	wg.Wait()

	return nil
}

func (cb *Couchbase) gatherServer(addr string, acc telegraf.Accumulator, pool *couchbase.Pool) error {
	if pool == nil {
		client, err := couchbase.Connect(addr)
		if err != nil {
			return err
		}

		// `default` is the only possible pool name. It's a
		// placeholder for a possible future Couchbase feature. See
		// http://stackoverflow.com/a/16990911/17498.
		p, err := client.GetPool("default")
		if err != nil {
			return err
		}
		pool = &p
	}

	for i := 0; i < len(pool.Nodes); i++ {
		node := pool.Nodes[i]
		tags := map[string]string{"cluster": regexpURI.ReplaceAllString(addr, "${1}"), "hostname": node.Hostname}
		fields := make(map[string]interface{})
		cb.addField(fields, "memory_free", node.MemoryFree)
		cb.addField(fields, "memory_total", node.MemoryTotal)
		acc.AddFields("couchbase_node", fields, tags)
	}

	for bucketName := range pool.BucketMap {
		tags := map[string]string{"cluster": regexpURI.ReplaceAllString(addr, "${1}"), "bucket": bucketName}
		bs := pool.BucketMap[bucketName].BasicStats
		fields := make(map[string]interface{})
		cb.addField(fields, "quota_percent_used", bs["quotaPercentUsed"])
		cb.addField(fields, "ops_per_sec", bs["opsPerSec"])
		cb.addField(fields, "disk_fetches", bs["diskFetches"])
		cb.addField(fields, "item_count", bs["itemCount"])
		cb.addField(fields, "disk_used", bs["diskUsed"])
		cb.addField(fields, "data_used", bs["dataUsed"])
		cb.addField(fields, "mem_used", bs["memUsed"])

		if cb.BucketMetricType == "detailed" {
			err := cb.gatherDetailedBucketStats(addr, bucketName, fields)
			if err != nil {
				return err
			}
		}

		acc.AddFields("couchbase_bucket", fields, tags)
	}

	return nil
}

func (cb *Couchbase) gatherDetailedBucketStats(server, bucket string, fields map[string]interface{}) error {
	extendedBucketStats := &BucketStats{}
	err := cb.queryDetailedBucketStats(server, bucket, extendedBucketStats)
	if err != nil {
		return err
	}

	// Use length of any set of metrics, they will all be the same length.
	lastEntry := len(extendedBucketStats.Op.Samples.CouchTotalDiskSize) - 1
	cb.addField(fields, "couch_total_disk_size", extendedBucketStats.Op.Samples.CouchTotalDiskSize[lastEntry])
	cb.addField(fields, "couch_docs_fragmentation", extendedBucketStats.Op.Samples.CouchDocsFragmentation[lastEntry])
	cb.addField(fields, "couch_views_fragmentation", extendedBucketStats.Op.Samples.CouchViewsFragmentation[lastEntry])
	cb.addField(fields, "hit_ratio", extendedBucketStats.Op.Samples.HitRatio[lastEntry])
	cb.addField(fields, "ep_cache_miss_rate", extendedBucketStats.Op.Samples.EpCacheMissRate[lastEntry])
	cb.addField(fields, "ep_resident_items_rate", extendedBucketStats.Op.Samples.EpResidentItemsRate[lastEntry])
	cb.addField(fields, "vb_avg_active_queue_age", extendedBucketStats.Op.Samples.VbAvgActiveQueueAge[lastEntry])
	cb.addField(fields, "vb_avg_replica_queue_age", extendedBucketStats.Op.Samples.VbAvgReplicaQueueAge[lastEntry])
	cb.addField(fields, "vb_avg_pending_queue_age", extendedBucketStats.Op.Samples.VbAvgPendingQueueAge[lastEntry])
	cb.addField(fields, "vb_avg_total_queue_age", extendedBucketStats.Op.Samples.VbAvgTotalQueueAge[lastEntry])
	cb.addField(fields, "vb_active_resident_items_ratio", extendedBucketStats.Op.Samples.VbActiveResidentItemsRatio[lastEntry])
	cb.addField(fields, "vb_replica_resident_items_ratio", extendedBucketStats.Op.Samples.VbReplicaResidentItemsRatio[lastEntry])
	cb.addField(fields, "vb_pending_resident_items_ratio", extendedBucketStats.Op.Samples.VbPendingResidentItemsRatio[lastEntry])
	cb.addField(fields, "avg_disk_update_time", extendedBucketStats.Op.Samples.AvgDiskUpdateTime[lastEntry])
	cb.addField(fields, "avg_disk_commit_time", extendedBucketStats.Op.Samples.AvgDiskCommitTime[lastEntry])
	cb.addField(fields, "avg_bg_wait_time", extendedBucketStats.Op.Samples.AvgBgWaitTime[lastEntry])
	cb.addField(fields, "avg_active_timestamp_drift", extendedBucketStats.Op.Samples.AvgActiveTimestampDrift[lastEntry])
	cb.addField(fields, "avg_replica_timestamp_drift", extendedBucketStats.Op.Samples.AvgReplicaTimestampDrift[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_count", extendedBucketStats.Op.Samples.EpDcpViewsIndexesCount[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_items_remaining", extendedBucketStats.Op.Samples.EpDcpViewsIndexesItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_producer_count", extendedBucketStats.Op.Samples.EpDcpViewsIndexesProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpViewsIndexesTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_items_sent", extendedBucketStats.Op.Samples.EpDcpViewsIndexesItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_total_bytes", extendedBucketStats.Op.Samples.EpDcpViewsIndexesTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_views+indexes_backoff", extendedBucketStats.Op.Samples.EpDcpViewsIndexesBackoff[lastEntry])
	cb.addField(fields, "bg_wait_count", extendedBucketStats.Op.Samples.BgWaitCount[lastEntry])
	cb.addField(fields, "bg_wait_total", extendedBucketStats.Op.Samples.BgWaitTotal[lastEntry])
	cb.addField(fields, "bytes_read", extendedBucketStats.Op.Samples.BytesRead[lastEntry])
	cb.addField(fields, "bytes_written", extendedBucketStats.Op.Samples.BytesWritten[lastEntry])
	cb.addField(fields, "cas_badval", extendedBucketStats.Op.Samples.CasBadval[lastEntry])
	cb.addField(fields, "cas_hits", extendedBucketStats.Op.Samples.CasHits[lastEntry])
	cb.addField(fields, "cas_misses", extendedBucketStats.Op.Samples.CasMisses[lastEntry])
	cb.addField(fields, "cmd_get", extendedBucketStats.Op.Samples.CmdGet[lastEntry])
	cb.addField(fields, "cmd_lookup", extendedBucketStats.Op.Samples.CmdLookup[lastEntry])
	cb.addField(fields, "cmd_set", extendedBucketStats.Op.Samples.CmdSet[lastEntry])
	cb.addField(fields, "couch_docs_actual_disk_size", extendedBucketStats.Op.Samples.CouchDocsActualDiskSize[lastEntry])
	cb.addField(fields, "couch_docs_data_size", extendedBucketStats.Op.Samples.CouchDocsDataSize[lastEntry])
	cb.addField(fields, "couch_docs_disk_size", extendedBucketStats.Op.Samples.CouchDocsDiskSize[lastEntry])
	cb.addField(fields, "couch_spatial_data_size", extendedBucketStats.Op.Samples.CouchSpatialDataSize[lastEntry])
	cb.addField(fields, "couch_spatial_disk_size", extendedBucketStats.Op.Samples.CouchSpatialDiskSize[lastEntry])
	cb.addField(fields, "couch_spatial_ops", extendedBucketStats.Op.Samples.CouchSpatialOps[lastEntry])
	cb.addField(fields, "couch_views_actual_disk_size", extendedBucketStats.Op.Samples.CouchViewsActualDiskSize[lastEntry])
	cb.addField(fields, "couch_views_data_size", extendedBucketStats.Op.Samples.CouchViewsDataSize[lastEntry])
	cb.addField(fields, "couch_views_disk_size", extendedBucketStats.Op.Samples.CouchViewsDiskSize[lastEntry])
	cb.addField(fields, "couch_views_ops", extendedBucketStats.Op.Samples.CouchViewsOps[lastEntry])
	cb.addField(fields, "curr_connections", extendedBucketStats.Op.Samples.CurrConnections[lastEntry])
	cb.addField(fields, "curr_items", extendedBucketStats.Op.Samples.CurrItems[lastEntry])
	cb.addField(fields, "curr_items_tot", extendedBucketStats.Op.Samples.CurrItemsTot[lastEntry])
	cb.addField(fields, "decr_hits", extendedBucketStats.Op.Samples.DecrHits[lastEntry])
	cb.addField(fields, "decr_misses", extendedBucketStats.Op.Samples.DecrMisses[lastEntry])
	cb.addField(fields, "delete_hits", extendedBucketStats.Op.Samples.DeleteHits[lastEntry])
	cb.addField(fields, "delete_misses", extendedBucketStats.Op.Samples.DeleteMisses[lastEntry])
	cb.addField(fields, "disk_commit_count", extendedBucketStats.Op.Samples.DiskCommitCount[lastEntry])
	cb.addField(fields, "disk_commit_total", extendedBucketStats.Op.Samples.DiskCommitTotal[lastEntry])
	cb.addField(fields, "disk_update_count", extendedBucketStats.Op.Samples.DiskUpdateCount[lastEntry])
	cb.addField(fields, "disk_update_total", extendedBucketStats.Op.Samples.DiskUpdateTotal[lastEntry])
	cb.addField(fields, "disk_write_queue", extendedBucketStats.Op.Samples.DiskWriteQueue[lastEntry])
	cb.addField(fields, "ep_active_ahead_exceptions", extendedBucketStats.Op.Samples.EpActiveAheadExceptions[lastEntry])
	cb.addField(fields, "ep_active_hlc_drift", extendedBucketStats.Op.Samples.EpActiveHlcDrift[lastEntry])
	cb.addField(fields, "ep_active_hlc_drift_count", extendedBucketStats.Op.Samples.EpActiveHlcDriftCount[lastEntry])
	cb.addField(fields, "ep_bg_fetched", extendedBucketStats.Op.Samples.EpBgFetched[lastEntry])
	cb.addField(fields, "ep_clock_cas_drift_threshold_exceeded", extendedBucketStats.Op.Samples.EpClockCasDriftThresholdExceeded[lastEntry])
	cb.addField(fields, "ep_data_read_failed", extendedBucketStats.Op.Samples.EpDataReadFailed[lastEntry])
	cb.addField(fields, "ep_data_write_failed", extendedBucketStats.Op.Samples.EpDataWriteFailed[lastEntry])
	cb.addField(fields, "ep_dcp_2i_backoff", extendedBucketStats.Op.Samples.EpDcp2IBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_2i_count", extendedBucketStats.Op.Samples.EpDcp2ICount[lastEntry])
	cb.addField(fields, "ep_dcp_2i_items_remaining", extendedBucketStats.Op.Samples.EpDcp2IItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_2i_items_sent", extendedBucketStats.Op.Samples.EpDcp2IItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_2i_producer_count", extendedBucketStats.Op.Samples.EpDcp2IProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_2i_total_backlog_size", extendedBucketStats.Op.Samples.EpDcp2ITotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_2i_total_bytes", extendedBucketStats.Op.Samples.EpDcp2ITotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_backoff", extendedBucketStats.Op.Samples.EpDcpCbasBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_count", extendedBucketStats.Op.Samples.EpDcpCbasCount[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_items_remaining", extendedBucketStats.Op.Samples.EpDcpCbasItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_items_sent", extendedBucketStats.Op.Samples.EpDcpCbasItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_producer_count", extendedBucketStats.Op.Samples.EpDcpCbasProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpCbasTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_cbas_total_bytes", extendedBucketStats.Op.Samples.EpDcpCbasTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_backoff", extendedBucketStats.Op.Samples.EpDcpEventingBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_count", extendedBucketStats.Op.Samples.EpDcpEventingCount[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_items_remaining", extendedBucketStats.Op.Samples.EpDcpEventingItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_items_sent", extendedBucketStats.Op.Samples.EpDcpEventingItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_producer_count", extendedBucketStats.Op.Samples.EpDcpEventingProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpEventingTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_eventing_total_bytes", extendedBucketStats.Op.Samples.EpDcpEventingTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_fts_backoff", extendedBucketStats.Op.Samples.EpDcpFtsBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_fts_count", extendedBucketStats.Op.Samples.EpDcpFtsCount[lastEntry])
	cb.addField(fields, "ep_dcp_fts_items_remaining", extendedBucketStats.Op.Samples.EpDcpFtsItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_fts_items_sent", extendedBucketStats.Op.Samples.EpDcpFtsItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_fts_producer_count", extendedBucketStats.Op.Samples.EpDcpFtsProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_fts_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpFtsTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_fts_total_bytes", extendedBucketStats.Op.Samples.EpDcpFtsTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_other_backoff", extendedBucketStats.Op.Samples.EpDcpOtherBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_other_count", extendedBucketStats.Op.Samples.EpDcpOtherCount[lastEntry])
	cb.addField(fields, "ep_dcp_other_items_remaining", extendedBucketStats.Op.Samples.EpDcpOtherItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_other_items_sent", extendedBucketStats.Op.Samples.EpDcpOtherItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_other_producer_count", extendedBucketStats.Op.Samples.EpDcpOtherProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_other_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpOtherTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_other_total_bytes", extendedBucketStats.Op.Samples.EpDcpOtherTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_replica_backoff", extendedBucketStats.Op.Samples.EpDcpReplicaBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_replica_count", extendedBucketStats.Op.Samples.EpDcpReplicaCount[lastEntry])
	cb.addField(fields, "ep_dcp_replica_items_remaining", extendedBucketStats.Op.Samples.EpDcpReplicaItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_replica_items_sent", extendedBucketStats.Op.Samples.EpDcpReplicaItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_replica_producer_count", extendedBucketStats.Op.Samples.EpDcpReplicaProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_replica_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpReplicaTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_replica_total_bytes", extendedBucketStats.Op.Samples.EpDcpReplicaTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_views_backoff", extendedBucketStats.Op.Samples.EpDcpViewsBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_views_count", extendedBucketStats.Op.Samples.EpDcpViewsCount[lastEntry])
	cb.addField(fields, "ep_dcp_views_items_remaining", extendedBucketStats.Op.Samples.EpDcpViewsItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_views_items_sent", extendedBucketStats.Op.Samples.EpDcpViewsItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_views_producer_count", extendedBucketStats.Op.Samples.EpDcpViewsProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_views_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpViewsTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_views_total_bytes", extendedBucketStats.Op.Samples.EpDcpViewsTotalBytes[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_backoff", extendedBucketStats.Op.Samples.EpDcpXdcrBackoff[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_count", extendedBucketStats.Op.Samples.EpDcpXdcrCount[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_items_remaining", extendedBucketStats.Op.Samples.EpDcpXdcrItemsRemaining[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_items_sent", extendedBucketStats.Op.Samples.EpDcpXdcrItemsSent[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_producer_count", extendedBucketStats.Op.Samples.EpDcpXdcrProducerCount[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpXdcrTotalBacklogSize[lastEntry])
	cb.addField(fields, "ep_dcp_xdcr_total_bytes", extendedBucketStats.Op.Samples.EpDcpXdcrTotalBytes[lastEntry])
	cb.addField(fields, "ep_diskqueue_drain", extendedBucketStats.Op.Samples.EpDiskqueueDrain[lastEntry])
	cb.addField(fields, "ep_diskqueue_fill", extendedBucketStats.Op.Samples.EpDiskqueueFill[lastEntry])
	cb.addField(fields, "ep_diskqueue_items", extendedBucketStats.Op.Samples.EpDiskqueueItems[lastEntry])
	cb.addField(fields, "ep_flusher_todo", extendedBucketStats.Op.Samples.EpFlusherTodo[lastEntry])
	cb.addField(fields, "ep_item_commit_failed", extendedBucketStats.Op.Samples.EpItemCommitFailed[lastEntry])
	cb.addField(fields, "ep_kv_size", extendedBucketStats.Op.Samples.EpKvSize[lastEntry])
	cb.addField(fields, "ep_max_size", extendedBucketStats.Op.Samples.EpMaxSize[lastEntry])
	cb.addField(fields, "ep_mem_high_wat", extendedBucketStats.Op.Samples.EpMemHighWat[lastEntry])
	cb.addField(fields, "ep_mem_low_wat", extendedBucketStats.Op.Samples.EpMemLowWat[lastEntry])
	cb.addField(fields, "ep_meta_data_memory", extendedBucketStats.Op.Samples.EpMetaDataMemory[lastEntry])
	cb.addField(fields, "ep_num_non_resident", extendedBucketStats.Op.Samples.EpNumNonResident[lastEntry])
	cb.addField(fields, "ep_num_ops_del_meta", extendedBucketStats.Op.Samples.EpNumOpsDelMeta[lastEntry])
	cb.addField(fields, "ep_num_ops_del_ret_meta", extendedBucketStats.Op.Samples.EpNumOpsDelRetMeta[lastEntry])
	cb.addField(fields, "ep_num_ops_get_meta", extendedBucketStats.Op.Samples.EpNumOpsGetMeta[lastEntry])
	cb.addField(fields, "ep_num_ops_set_meta", extendedBucketStats.Op.Samples.EpNumOpsSetMeta[lastEntry])
	cb.addField(fields, "ep_num_ops_set_ret_meta", extendedBucketStats.Op.Samples.EpNumOpsSetRetMeta[lastEntry])
	cb.addField(fields, "ep_num_value_ejects", extendedBucketStats.Op.Samples.EpNumValueEjects[lastEntry])
	cb.addField(fields, "ep_oom_errors", extendedBucketStats.Op.Samples.EpOomErrors[lastEntry])
	cb.addField(fields, "ep_ops_create", extendedBucketStats.Op.Samples.EpOpsCreate[lastEntry])
	cb.addField(fields, "ep_ops_update", extendedBucketStats.Op.Samples.EpOpsUpdate[lastEntry])
	cb.addField(fields, "ep_overhead", extendedBucketStats.Op.Samples.EpOverhead[lastEntry])
	cb.addField(fields, "ep_queue_size", extendedBucketStats.Op.Samples.EpQueueSize[lastEntry])
	cb.addField(fields, "ep_replica_ahead_exceptions", extendedBucketStats.Op.Samples.EpReplicaAheadExceptions[lastEntry])
	cb.addField(fields, "ep_replica_hlc_drift", extendedBucketStats.Op.Samples.EpReplicaHlcDrift[lastEntry])
	cb.addField(fields, "ep_replica_hlc_drift_count", extendedBucketStats.Op.Samples.EpReplicaHlcDriftCount[lastEntry])
	cb.addField(fields, "ep_tmp_oom_errors", extendedBucketStats.Op.Samples.EpTmpOomErrors[lastEntry])
	cb.addField(fields, "ep_vb_total", extendedBucketStats.Op.Samples.EpVbTotal[lastEntry])
	cb.addField(fields, "evictions", extendedBucketStats.Op.Samples.Evictions[lastEntry])
	cb.addField(fields, "get_hits", extendedBucketStats.Op.Samples.GetHits[lastEntry])
	cb.addField(fields, "get_misses", extendedBucketStats.Op.Samples.GetMisses[lastEntry])
	cb.addField(fields, "incr_hits", extendedBucketStats.Op.Samples.IncrHits[lastEntry])
	cb.addField(fields, "incr_misses", extendedBucketStats.Op.Samples.IncrMisses[lastEntry])
	cb.addField(fields, "misses", extendedBucketStats.Op.Samples.Misses[lastEntry])
	cb.addField(fields, "ops", extendedBucketStats.Op.Samples.Ops[lastEntry])
	cb.addField(fields, "timestamp", extendedBucketStats.Op.Samples.Timestamp[lastEntry])
	cb.addField(fields, "vb_active_eject", extendedBucketStats.Op.Samples.VbActiveEject[lastEntry])
	cb.addField(fields, "vb_active_itm_memory", extendedBucketStats.Op.Samples.VbActiveItmMemory[lastEntry])
	cb.addField(fields, "vb_active_meta_data_memory", extendedBucketStats.Op.Samples.VbActiveMetaDataMemory[lastEntry])
	cb.addField(fields, "vb_active_num", extendedBucketStats.Op.Samples.VbActiveNum[lastEntry])
	cb.addField(fields, "vb_active_num_non_resident", extendedBucketStats.Op.Samples.VbActiveNumNonResident[lastEntry])
	cb.addField(fields, "vb_active_ops_create", extendedBucketStats.Op.Samples.VbActiveOpsCreate[lastEntry])
	cb.addField(fields, "vb_active_ops_update", extendedBucketStats.Op.Samples.VbActiveOpsUpdate[lastEntry])
	cb.addField(fields, "vb_active_queue_age", extendedBucketStats.Op.Samples.VbActiveQueueAge[lastEntry])
	cb.addField(fields, "vb_active_queue_drain", extendedBucketStats.Op.Samples.VbActiveQueueDrain[lastEntry])
	cb.addField(fields, "vb_active_queue_fill", extendedBucketStats.Op.Samples.VbActiveQueueFill[lastEntry])
	cb.addField(fields, "vb_active_queue_size", extendedBucketStats.Op.Samples.VbActiveQueueSize[lastEntry])
	cb.addField(fields, "vb_active_sync_write_aborted_count", extendedBucketStats.Op.Samples.VbActiveSyncWriteAbortedCount[lastEntry])
	cb.addField(fields, "vb_active_sync_write_accepted_count", extendedBucketStats.Op.Samples.VbActiveSyncWriteAcceptedCount[lastEntry])
	cb.addField(fields, "vb_active_sync_write_committed_count", extendedBucketStats.Op.Samples.VbActiveSyncWriteCommittedCount[lastEntry])
	cb.addField(fields, "vb_pending_curr_items", extendedBucketStats.Op.Samples.VbPendingCurrItems[lastEntry])
	cb.addField(fields, "vb_pending_eject", extendedBucketStats.Op.Samples.VbPendingEject[lastEntry])
	cb.addField(fields, "vb_pending_itm_memory", extendedBucketStats.Op.Samples.VbPendingItmMemory[lastEntry])
	cb.addField(fields, "vb_pending_meta_data_memory", extendedBucketStats.Op.Samples.VbPendingMetaDataMemory[lastEntry])
	cb.addField(fields, "vb_pending_num", extendedBucketStats.Op.Samples.VbPendingNum[lastEntry])
	cb.addField(fields, "vb_pending_num_non_resident", extendedBucketStats.Op.Samples.VbPendingNumNonResident[lastEntry])
	cb.addField(fields, "vb_pending_ops_create", extendedBucketStats.Op.Samples.VbPendingOpsCreate[lastEntry])
	cb.addField(fields, "vb_pending_ops_update", extendedBucketStats.Op.Samples.VbPendingOpsUpdate[lastEntry])
	cb.addField(fields, "vb_pending_queue_age", extendedBucketStats.Op.Samples.VbPendingQueueAge[lastEntry])
	cb.addField(fields, "vb_pending_queue_drain", extendedBucketStats.Op.Samples.VbPendingQueueDrain[lastEntry])
	cb.addField(fields, "vb_pending_queue_fill", extendedBucketStats.Op.Samples.VbPendingQueueFill[lastEntry])
	cb.addField(fields, "vb_pending_queue_size", extendedBucketStats.Op.Samples.VbPendingQueueSize[lastEntry])
	cb.addField(fields, "vb_replica_curr_items", extendedBucketStats.Op.Samples.VbReplicaCurrItems[lastEntry])
	cb.addField(fields, "vb_replica_eject", extendedBucketStats.Op.Samples.VbReplicaEject[lastEntry])
	cb.addField(fields, "vb_replica_itm_memory", extendedBucketStats.Op.Samples.VbReplicaItmMemory[lastEntry])
	cb.addField(fields, "vb_replica_meta_data_memory", extendedBucketStats.Op.Samples.VbReplicaMetaDataMemory[lastEntry])
	cb.addField(fields, "vb_replica_num", extendedBucketStats.Op.Samples.VbReplicaNum[lastEntry])
	cb.addField(fields, "vb_replica_num_non_resident", extendedBucketStats.Op.Samples.VbReplicaNumNonResident[lastEntry])
	cb.addField(fields, "vb_replica_ops_create", extendedBucketStats.Op.Samples.VbReplicaOpsCreate[lastEntry])
	cb.addField(fields, "vb_replica_ops_update", extendedBucketStats.Op.Samples.VbReplicaOpsUpdate[lastEntry])
	cb.addField(fields, "vb_replica_queue_age", extendedBucketStats.Op.Samples.VbReplicaQueueAge[lastEntry])
	cb.addField(fields, "vb_replica_queue_drain", extendedBucketStats.Op.Samples.VbReplicaQueueDrain[lastEntry])
	cb.addField(fields, "vb_replica_queue_fill", extendedBucketStats.Op.Samples.VbReplicaQueueFill[lastEntry])
	cb.addField(fields, "vb_replica_queue_size", extendedBucketStats.Op.Samples.VbReplicaQueueSize[lastEntry])
	cb.addField(fields, "vb_total_queue_age", extendedBucketStats.Op.Samples.VbTotalQueueAge[lastEntry])
	cb.addField(fields, "xdc_ops", extendedBucketStats.Op.Samples.XdcOps[lastEntry])
	cb.addField(fields, "allocstall", extendedBucketStats.Op.Samples.Allocstall[lastEntry])
	cb.addField(fields, "cpu_cores_available", extendedBucketStats.Op.Samples.CPUCoresAvailable[lastEntry])
	cb.addField(fields, "cpu_irq_rate", extendedBucketStats.Op.Samples.CPUIrqRate[lastEntry])
	cb.addField(fields, "cpu_stolen_rate", extendedBucketStats.Op.Samples.CPUStolenRate[lastEntry])
	cb.addField(fields, "cpu_sys_rate", extendedBucketStats.Op.Samples.CPUSysRate[lastEntry])
	cb.addField(fields, "cpu_user_rate", extendedBucketStats.Op.Samples.CPUUserRate[lastEntry])
	cb.addField(fields, "cpu_utilization_rate", extendedBucketStats.Op.Samples.CPUUtilizationRate[lastEntry])
	cb.addField(fields, "hibernated_requests", extendedBucketStats.Op.Samples.HibernatedRequests[lastEntry])
	cb.addField(fields, "hibernated_waked", extendedBucketStats.Op.Samples.HibernatedWaked[lastEntry])
	cb.addField(fields, "mem_actual_free", extendedBucketStats.Op.Samples.MemActualFree[lastEntry])
	cb.addField(fields, "mem_actual_used", extendedBucketStats.Op.Samples.MemActualUsed[lastEntry])
	cb.addField(fields, "mem_free", extendedBucketStats.Op.Samples.MemFree[lastEntry])
	cb.addField(fields, "mem_limit", extendedBucketStats.Op.Samples.MemLimit[lastEntry])
	cb.addField(fields, "mem_total", extendedBucketStats.Op.Samples.MemTotal[lastEntry])
	cb.addField(fields, "mem_used_sys", extendedBucketStats.Op.Samples.MemUsedSys[lastEntry])
	cb.addField(fields, "odp_report_failed", extendedBucketStats.Op.Samples.OdpReportFailed[lastEntry])
	cb.addField(fields, "rest_requests", extendedBucketStats.Op.Samples.RestRequests[lastEntry])
	cb.addField(fields, "swap_total", extendedBucketStats.Op.Samples.SwapTotal[lastEntry])
	cb.addField(fields, "swap_used", extendedBucketStats.Op.Samples.SwapUsed[lastEntry])

	return nil
}

func (cb *Couchbase) addField(fields map[string]interface{}, fieldKey string, value interface{}) {
	if !cb.includeExclude.Match(fieldKey) {
		return
	}

	fields[fieldKey] = value
}

func (cb *Couchbase) queryDetailedBucketStats(server, bucket string, bucketStats *BucketStats) error {
	// Set up an HTTP request to get the complete set of bucket stats.
	req, err := http.NewRequest("GET", server+"/pools/default/buckets/"+bucket+"/stats?", nil)
	if err != nil {
		return err
	}

	r, err := client.Do(req)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(bucketStats)
}

func (cb *Couchbase) Init() error {
	filter, err := filter.NewIncludeExcludeFilter(cb.FieldsIncluded, cb.FieldsExcluded)
	if err != nil {
		return err
	}

	cb.includeExclude = filter

	return nil
}

func init() {
	inputs.Add("couchbase", func() telegraf.Input {
		return &Couchbase{
			BucketMetricType: "basic",
		}
	})
}
