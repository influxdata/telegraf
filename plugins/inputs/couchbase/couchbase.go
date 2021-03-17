package couchbase

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"time"

	couchbase "github.com/couchbase/go-couchbase"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Couchbase struct {
	Servers []string

	// We use this to know how to filter out our stats results.
	lastInterval time.Time

	// Zoom defines how wide of a time period we ask Couchbase for stats over.
	zoom string

	// Zoom levels mapped to the time intervals of their metrics.
	couchbaseTimeIntervals map[string]time.Duration
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
`

var regexpURI = regexp.MustCompile(`(\S+://)?(\S+\:\S+@)`)
var client = &http.Client{Timeout: 10 * time.Second}

func (r *Couchbase) SampleConfig() string {
	return sampleConfig
}

func (r *Couchbase) Description() string {
	return "Read metrics from one or many couchbase clusters"
}

// Reads stats from all configured clusters. Accumulates stats.
// Returns one of the errors encountered while gathering stats (if any).
func (r *Couchbase) Gather(acc telegraf.Accumulator) error {
	// We skip the first interval to get a measurement on how long our intervals will be (how far back to look for metrics).
	if r.lastInterval.IsZero() {
		r.lastInterval = time.Now()
		return nil
	}

	if len(r.Servers) == 0 {
		r.gatherServer("http://localhost:8091/", acc, nil)
		return nil
	}

	var wg sync.WaitGroup
	for _, serv := range r.Servers {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			acc.AddError(r.gatherServer(serv, acc, nil))
		}(serv)
	}

	wg.Wait()

	return nil
}

func (r *Couchbase) gatherServer(addr string, acc telegraf.Accumulator, pool *couchbase.Pool) error {
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
		fields["memory_free"] = node.MemoryFree
		fields["memory_total"] = node.MemoryTotal
		acc.AddFields("couchbase_node", fields, tags)
	}

	for bucketName := range pool.BucketMap {
		tags := map[string]string{"cluster": regexpURI.ReplaceAllString(addr, "${1}"), "bucket": bucketName}
		bs := pool.BucketMap[bucketName].BasicStats
		fields := make(map[string]interface{})
		fields["quota_percent_used"] = bs["quotaPercentUsed"]
		fields["ops_per_sec"] = bs["opsPerSec"]
		fields["disk_fetches"] = bs["diskFetches"]
		fields["item_count"] = bs["itemCount"]
		fields["disk_used"] = bs["diskUsed"]
		fields["data_used"] = bs["dataUsed"]
		fields["mem_used"] = bs["memUsed"]
		acc.AddFields("couchbase_bucket", fields, tags)

		// Move on to gathering 'extended' couchbase stats.
		extendedBucketStats := &BucketStats{}
		err := r.extendedBucketStats(addr, bucketName, extendedBucketStats)
		if err != nil {
			return err
		}

		r.gatherExtendedBucketStats(acc, extendedBucketStats, tags)
	}

	return nil
}

func (couchbase *Couchbase) gatherExtendedBucketStats(acc telegraf.Accumulator, extendedBucketStats *BucketStats, tags map[string]string) {
	// Iterate backwards over the extended stats entries, so as to get the latest metric, and then the metric before that, etc.
	timestamp := time.Unix(0, extendedBucketStats.Op.Lasttstamp*int64(time.Millisecond))

	// To get an accurate size of the number of metrics returned, we get the length of an arbitrary choice of the metric arrays.
	for i := len(extendedBucketStats.Op.Samples.CouchTotalDiskSize) - 1; i >= 0; i-- {
		extendedFields := make(map[string]interface{})
		extendedFields["couch_total_disk_size"] = extendedBucketStats.Op.Samples.CouchTotalDiskSize[i]
		extendedFields["couch_docs_fragmentation"] = extendedBucketStats.Op.Samples.CouchDocsFragmentation[i]
		extendedFields["couch_views_fragmentation"] = extendedBucketStats.Op.Samples.CouchViewsFragmentation[i]
		extendedFields["hit_ratio"] = extendedBucketStats.Op.Samples.HitRatio[i]
		extendedFields["ep_cache_miss_rate"] = extendedBucketStats.Op.Samples.EpCacheMissRate[i]
		extendedFields["ep_resident_items_rate"] = extendedBucketStats.Op.Samples.EpResidentItemsRate[i]
		extendedFields["vb_avg_active_queue_age"] = extendedBucketStats.Op.Samples.VbAvgActiveQueueAge[i]
		extendedFields["vb_avg_replica_queue_age"] = extendedBucketStats.Op.Samples.VbAvgReplicaQueueAge[i]
		extendedFields["vb_avg_pending_queue_age"] = extendedBucketStats.Op.Samples.VbAvgPendingQueueAge
		extendedFields["vb_avg_total_queue_age"] = extendedBucketStats.Op.Samples.VbAvgTotalQueueAge[i]
		extendedFields["vb_active_resident_items_ratio"] = extendedBucketStats.Op.Samples.VbActiveResidentItemsRatio[i]
		extendedFields["vb_replica_resident_items_ratio"] = extendedBucketStats.Op.Samples.VbReplicaResidentItemsRatio[i]
		extendedFields["vb_pending_resident_items_ratio"] = extendedBucketStats.Op.Samples.VbPendingResidentItemsRatio[i]
		extendedFields["avg_disk_update_time"] = extendedBucketStats.Op.Samples.AvgDiskUpdateTime[i]
		extendedFields["avg_disk_commit_time"] = extendedBucketStats.Op.Samples.AvgDiskCommitTime[i]
		extendedFields["avg_bg_wait_time"] = extendedBucketStats.Op.Samples.AvgBgWaitTime[i]
		extendedFields["avg_active_timestamp_drift"] = extendedBucketStats.Op.Samples.AvgActiveTimestampDrift[i]
		extendedFields["avg_replica_timestamp_drift"] = extendedBucketStats.Op.Samples.AvgReplicaTimestampDrift[i]
		extendedFields["ep_dcp_views+indexes_count"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesCount[i]
		extendedFields["ep_dcp_views+indexes_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesItemsRemaining[i]
		extendedFields["ep_dcp_views+indexes_producer_count"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesProducerCount[i]
		extendedFields["ep_dcp_views+indexes_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesTotalBacklogSize[i]
		extendedFields["ep_dcp_views+indexes_items_sent"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesItemsSent[i]
		extendedFields["ep_dcp_views+indexes_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesTotalBytes[i]
		extendedFields["ep_dcp_views+indexes_backoff"] = extendedBucketStats.Op.Samples.EpDcpViewsIndexesBackoff[i]
		extendedFields["bg_wait_count"] = extendedBucketStats.Op.Samples.BgWaitCount[i]
		extendedFields["bg_wait_total"] = extendedBucketStats.Op.Samples.BgWaitTotal[i]
		extendedFields["bytes_read"] = extendedBucketStats.Op.Samples.BytesRead[i]
		extendedFields["bytes_written"] = extendedBucketStats.Op.Samples.BytesWritten[i]
		extendedFields["cas_badval"] = extendedBucketStats.Op.Samples.CasBadval[i]
		extendedFields["cas_hits"] = extendedBucketStats.Op.Samples.CasHits[i]
		extendedFields["cas_misses"] = extendedBucketStats.Op.Samples.CasMisses[i]
		extendedFields["cmd_get"] = extendedBucketStats.Op.Samples.CmdGet[i]
		extendedFields["cmd_lookup"] = extendedBucketStats.Op.Samples.CmdLookup[i]
		extendedFields["cmd_set"] = extendedBucketStats.Op.Samples.CmdSet[i]
		extendedFields["couch_docs_actual_disk_size"] = extendedBucketStats.Op.Samples.CouchDocsActualDiskSize[i]
		extendedFields["couch_docs_data_size"] = extendedBucketStats.Op.Samples.CouchDocsDataSize[i]
		extendedFields["couch_docs_disk_size"] = extendedBucketStats.Op.Samples.CouchDocsDiskSize[i]
		extendedFields["couch_spatial_data_size"] = extendedBucketStats.Op.Samples.CouchSpatialDataSize[i]
		extendedFields["couch_spatial_disk_size"] = extendedBucketStats.Op.Samples.CouchSpatialDiskSize[i]
		extendedFields["couch_spatial_ops"] = extendedBucketStats.Op.Samples.CouchSpatialOps[i]
		extendedFields["couch_views_actual_disk_size"] = extendedBucketStats.Op.Samples.CouchViewsActualDiskSize[i]
		extendedFields["couch_views_data_size"] = extendedBucketStats.Op.Samples.CouchViewsDataSize[i]
		extendedFields["couch_views_disk_size"] = extendedBucketStats.Op.Samples.CouchViewsDiskSize[i]
		extendedFields["couch_views_ops"] = extendedBucketStats.Op.Samples.CouchViewsOps[i]
		extendedFields["curr_connections"] = extendedBucketStats.Op.Samples.CurrConnections[i]
		extendedFields["curr_items"] = extendedBucketStats.Op.Samples.CurrItems[i]
		extendedFields["curr_items_tot"] = extendedBucketStats.Op.Samples.CurrItemsTot[i]
		extendedFields["decr_hits"] = extendedBucketStats.Op.Samples.DecrHits[i]
		extendedFields["decr_misses"] = extendedBucketStats.Op.Samples.DecrMisses[i]
		extendedFields["delete_hits"] = extendedBucketStats.Op.Samples.DeleteHits[i]
		extendedFields["delete_misses"] = extendedBucketStats.Op.Samples.DeleteMisses[i]
		extendedFields["disk_commit_count"] = extendedBucketStats.Op.Samples.DiskCommitCount[i]
		extendedFields["disk_commit_total"] = extendedBucketStats.Op.Samples.DiskCommitTotal[i]
		extendedFields["disk_update_count"] = extendedBucketStats.Op.Samples.DiskUpdateCount[i]
		extendedFields["disk_update_total"] = extendedBucketStats.Op.Samples.DiskUpdateTotal[i]
		extendedFields["disk_write_queue"] = extendedBucketStats.Op.Samples.DiskWriteQueue[i]
		extendedFields["ep_active_ahead_exceptions"] = extendedBucketStats.Op.Samples.EpActiveAheadExceptions[i]
		extendedFields["ep_active_hlc_drift"] = extendedBucketStats.Op.Samples.EpActiveHlcDrift[i]
		extendedFields["ep_active_hlc_drift_count"] = extendedBucketStats.Op.Samples.EpActiveHlcDriftCount[i]
		extendedFields["ep_bg_fetched"] = extendedBucketStats.Op.Samples.EpBgFetched[i]
		extendedFields["ep_clock_cas_drift_threshold_exceeded"] = extendedBucketStats.Op.Samples.EpClockCasDriftThresholdExceeded[i]
		extendedFields["ep_data_read_failed"] = extendedBucketStats.Op.Samples.EpDataReadFailed[i]
		extendedFields["ep_data_write_failed"] = extendedBucketStats.Op.Samples.EpDataWriteFailed[i]
		extendedFields["ep_dcp_2i_backoff"] = extendedBucketStats.Op.Samples.EpDcp2IBackoff[i]
		extendedFields["ep_dcp_2i_count"] = extendedBucketStats.Op.Samples.EpDcp2ICount[i]
		extendedFields["ep_dcp_2i_items_remaining"] = extendedBucketStats.Op.Samples.EpDcp2IItemsRemaining[i]
		extendedFields["ep_dcp_2i_items_sent"] = extendedBucketStats.Op.Samples.EpDcp2IItemsSent[i]
		extendedFields["ep_dcp_2i_producer_count"] = extendedBucketStats.Op.Samples.EpDcp2IProducerCount[i]
		extendedFields["ep_dcp_2i_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcp2ITotalBacklogSize[i]
		extendedFields["ep_dcp_2i_total_bytes"] = extendedBucketStats.Op.Samples.EpDcp2ITotalBytes[i]
		extendedFields["ep_dcp_cbas_backoff"] = extendedBucketStats.Op.Samples.EpDcpCbasBackoff[i]
		extendedFields["ep_dcp_cbas_count"] = extendedBucketStats.Op.Samples.EpDcpCbasCount[i]
		extendedFields["ep_dcp_cbas_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpCbasItemsRemaining[i]
		extendedFields["ep_dcp_cbas_items_sent"] = extendedBucketStats.Op.Samples.EpDcpCbasItemsSent[i]
		extendedFields["ep_dcp_cbas_producer_count"] = extendedBucketStats.Op.Samples.EpDcpCbasProducerCount[i]
		extendedFields["ep_dcp_cbas_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpCbasTotalBacklogSize[i]
		extendedFields["ep_dcp_cbas_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpCbasTotalBytes[i]
		extendedFields["ep_dcp_eventing_backoff"] = extendedBucketStats.Op.Samples.EpDcpEventingBackoff[i]
		extendedFields["ep_dcp_eventing_count"] = extendedBucketStats.Op.Samples.EpDcpEventingCount[i]
		extendedFields["ep_dcp_eventing_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpEventingItemsRemaining[i]
		extendedFields["ep_dcp_eventing_items_sent"] = extendedBucketStats.Op.Samples.EpDcpEventingItemsSent[i]
		extendedFields["ep_dcp_eventing_producer_count"] = extendedBucketStats.Op.Samples.EpDcpEventingProducerCount[i]
		extendedFields["ep_dcp_eventing_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpEventingTotalBacklogSize[i]
		extendedFields["ep_dcp_eventing_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpEventingTotalBytes[i]
		extendedFields["ep_dcp_fts_backoff"] = extendedBucketStats.Op.Samples.EpDcpFtsBackoff[i]
		extendedFields["ep_dcp_fts_count"] = extendedBucketStats.Op.Samples.EpDcpFtsCount[i]
		extendedFields["ep_dcp_fts_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpFtsItemsRemaining[i]
		extendedFields["ep_dcp_fts_items_sent"] = extendedBucketStats.Op.Samples.EpDcpFtsItemsSent[i]
		extendedFields["ep_dcp_fts_producer_count"] = extendedBucketStats.Op.Samples.EpDcpFtsProducerCount[i]
		extendedFields["ep_dcp_fts_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpFtsTotalBacklogSize[i]
		extendedFields["ep_dcp_fts_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpFtsTotalBytes[i]
		extendedFields["ep_dcp_other_backoff"] = extendedBucketStats.Op.Samples.EpDcpOtherBackoff[i]
		extendedFields["ep_dcp_other_count"] = extendedBucketStats.Op.Samples.EpDcpOtherCount[i]
		extendedFields["ep_dcp_other_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpOtherItemsRemaining[i]
		extendedFields["ep_dcp_other_items_sent"] = extendedBucketStats.Op.Samples.EpDcpOtherItemsSent[i]
		extendedFields["ep_dcp_other_producer_count"] = extendedBucketStats.Op.Samples.EpDcpOtherProducerCount[i]
		extendedFields["ep_dcp_other_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpOtherTotalBacklogSize[i]
		extendedFields["ep_dcp_other_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpOtherTotalBytes[i]
		extendedFields["ep_dcp_replica_backoff"] = extendedBucketStats.Op.Samples.EpDcpReplicaBackoff[i]
		extendedFields["ep_dcp_replica_count"] = extendedBucketStats.Op.Samples.EpDcpReplicaCount[i]
		extendedFields["ep_dcp_replica_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpReplicaItemsRemaining[i]
		extendedFields["ep_dcp_replica_items_sent"] = extendedBucketStats.Op.Samples.EpDcpReplicaItemsSent[i]
		extendedFields["ep_dcp_replica_producer_count"] = extendedBucketStats.Op.Samples.EpDcpReplicaProducerCount[i]
		extendedFields["ep_dcp_replica_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpReplicaTotalBacklogSize[i]
		extendedFields["ep_dcp_replica_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpReplicaTotalBytes[i]
		extendedFields["ep_dcp_views_backoff"] = extendedBucketStats.Op.Samples.EpDcpViewsBackoff[i]
		extendedFields["ep_dcp_views_count"] = extendedBucketStats.Op.Samples.EpDcpViewsCount[i]
		extendedFields["ep_dcp_views_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpViewsItemsRemaining[i]
		extendedFields["ep_dcp_views_items_sent"] = extendedBucketStats.Op.Samples.EpDcpViewsItemsSent[i]
		extendedFields["ep_dcp_views_producer_count"] = extendedBucketStats.Op.Samples.EpDcpViewsProducerCount[i]
		extendedFields["ep_dcp_views_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpViewsTotalBacklogSize[i]
		extendedFields["ep_dcp_views_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpViewsTotalBytes[i]
		extendedFields["ep_dcp_xdcr_backoff"] = extendedBucketStats.Op.Samples.EpDcpXdcrBackoff[i]
		extendedFields["ep_dcp_xdcr_count"] = extendedBucketStats.Op.Samples.EpDcpXdcrCount[i]
		extendedFields["ep_dcp_xdcr_items_remaining"] = extendedBucketStats.Op.Samples.EpDcpXdcrItemsRemaining[i]
		extendedFields["ep_dcp_xdcr_items_sent"] = extendedBucketStats.Op.Samples.EpDcpXdcrItemsSent[i]
		extendedFields["ep_dcp_xdcr_producer_count"] = extendedBucketStats.Op.Samples.EpDcpXdcrProducerCount[i]
		extendedFields["ep_dcp_xdcr_total_backlog_size"] = extendedBucketStats.Op.Samples.EpDcpXdcrTotalBacklogSize[i]
		extendedFields["ep_dcp_xdcr_total_bytes"] = extendedBucketStats.Op.Samples.EpDcpXdcrTotalBytes[i]
		extendedFields["ep_diskqueue_drain"] = extendedBucketStats.Op.Samples.EpDiskqueueDrain[i]
		extendedFields["ep_diskqueue_fill"] = extendedBucketStats.Op.Samples.EpDiskqueueFill[i]
		extendedFields["ep_diskqueue_items"] = extendedBucketStats.Op.Samples.EpDiskqueueItems[i]
		extendedFields["ep_flusher_todo"] = extendedBucketStats.Op.Samples.EpFlusherTodo[i]
		extendedFields["ep_item_commit_failed"] = extendedBucketStats.Op.Samples.EpItemCommitFailed[i]
		extendedFields["ep_kv_size"] = extendedBucketStats.Op.Samples.EpKvSize[i]
		extendedFields["ep_max_size"] = extendedBucketStats.Op.Samples.EpMaxSize[i]
		extendedFields["ep_mem_high_wat"] = extendedBucketStats.Op.Samples.EpMemHighWat[i]
		extendedFields["ep_mem_low_wat"] = extendedBucketStats.Op.Samples.EpMemLowWat[i]
		extendedFields["ep_meta_data_memory"] = extendedBucketStats.Op.Samples.EpMetaDataMemory[i]
		extendedFields["ep_num_non_resident"] = extendedBucketStats.Op.Samples.EpNumNonResident[i]
		extendedFields["ep_num_ops_del_meta"] = extendedBucketStats.Op.Samples.EpNumOpsDelMeta[i]
		extendedFields["ep_num_ops_del_ret_meta"] = extendedBucketStats.Op.Samples.EpNumOpsDelRetMeta[i]
		extendedFields["ep_num_ops_get_meta"] = extendedBucketStats.Op.Samples.EpNumOpsGetMeta[i]
		extendedFields["ep_num_ops_set_meta"] = extendedBucketStats.Op.Samples.EpNumOpsSetMeta[i]
		extendedFields["ep_num_ops_set_ret_meta"] = extendedBucketStats.Op.Samples.EpNumOpsSetRetMeta[i]
		extendedFields["ep_num_value_ejects"] = extendedBucketStats.Op.Samples.EpNumValueEjects[i]
		extendedFields["ep_oom_errors"] = extendedBucketStats.Op.Samples.EpOomErrors[i]
		extendedFields["ep_ops_create"] = extendedBucketStats.Op.Samples.EpOpsCreate[i]
		extendedFields["ep_ops_update"] = extendedBucketStats.Op.Samples.EpOpsUpdate[i]
		extendedFields["ep_overhead"] = extendedBucketStats.Op.Samples.EpOverhead[i]
		extendedFields["ep_queue_size"] = extendedBucketStats.Op.Samples.EpQueueSize[i]
		extendedFields["ep_replica_ahead_exceptions"] = extendedBucketStats.Op.Samples.EpReplicaAheadExceptions[i]
		extendedFields["ep_replica_hlc_drift"] = extendedBucketStats.Op.Samples.EpReplicaHlcDrift[i]
		extendedFields["ep_replica_hlc_drift_count"] = extendedBucketStats.Op.Samples.EpReplicaHlcDriftCount[i]
		extendedFields["ep_tmp_oom_errors"] = extendedBucketStats.Op.Samples.EpTmpOomErrors[i]
		extendedFields["ep_vb_total"] = extendedBucketStats.Op.Samples.EpVbTotal[i]
		extendedFields["evictions"] = extendedBucketStats.Op.Samples.Evictions[i]
		extendedFields["get_hits"] = extendedBucketStats.Op.Samples.GetHits[i]
		extendedFields["get_misses"] = extendedBucketStats.Op.Samples.GetMisses[i]
		extendedFields["incr_hits"] = extendedBucketStats.Op.Samples.IncrHits[i]
		extendedFields["incr_misses"] = extendedBucketStats.Op.Samples.IncrMisses[i]
		extendedFields["misses"] = extendedBucketStats.Op.Samples.Misses[i]
		extendedFields["ops"] = extendedBucketStats.Op.Samples.Ops[i]
		extendedFields["timestamp"] = extendedBucketStats.Op.Samples.Timestamp[i]
		extendedFields["vb_active_eject"] = extendedBucketStats.Op.Samples.VbActiveEject[i]
		extendedFields["vb_active_itm_memory"] = extendedBucketStats.Op.Samples.VbActiveItmMemory[i]
		extendedFields["vb_active_meta_data_memory"] = extendedBucketStats.Op.Samples.VbActiveMetaDataMemory[i]
		extendedFields["vb_active_num"] = extendedBucketStats.Op.Samples.VbActiveNum[i]
		extendedFields["vb_active_num_non_resident"] = extendedBucketStats.Op.Samples.VbActiveNumNonResident[i]
		extendedFields["vb_active_ops_create"] = extendedBucketStats.Op.Samples.VbActiveOpsCreate[i]
		extendedFields["vb_active_ops_update"] = extendedBucketStats.Op.Samples.VbActiveOpsUpdate[i]
		extendedFields["vb_active_queue_age"] = extendedBucketStats.Op.Samples.VbActiveQueueAge[i]
		extendedFields["vb_active_queue_drain"] = extendedBucketStats.Op.Samples.VbActiveQueueDrain[i]
		extendedFields["vb_active_queue_fill"] = extendedBucketStats.Op.Samples.VbActiveQueueFill[i]
		extendedFields["vb_active_queue_size"] = extendedBucketStats.Op.Samples.VbActiveQueueSize[i]
		extendedFields["vb_active_sync_write_aborted_count"] = extendedBucketStats.Op.Samples.VbActiveSyncWriteAbortedCount[i]
		extendedFields["vb_active_sync_write_accepted_count"] = extendedBucketStats.Op.Samples.VbActiveSyncWriteAcceptedCount[i]
		extendedFields["vb_active_sync_write_committed_count"] = extendedBucketStats.Op.Samples.VbActiveSyncWriteCommittedCount[i]
		extendedFields["vb_pending_curr_items"] = extendedBucketStats.Op.Samples.VbPendingCurrItems[i]
		extendedFields["vb_pending_eject"] = extendedBucketStats.Op.Samples.VbPendingEject[i]
		extendedFields["vb_pending_itm_memory"] = extendedBucketStats.Op.Samples.VbPendingItmMemory[i]
		extendedFields["vb_pending_meta_data_memory"] = extendedBucketStats.Op.Samples.VbPendingMetaDataMemory[i]
		extendedFields["vb_pending_num"] = extendedBucketStats.Op.Samples.VbPendingNum[i]
		extendedFields["vb_pending_num_non_resident"] = extendedBucketStats.Op.Samples.VbPendingNumNonResident[i]
		extendedFields["vb_pending_ops_create"] = extendedBucketStats.Op.Samples.VbPendingOpsCreate[i]
		extendedFields["vb_pending_ops_update"] = extendedBucketStats.Op.Samples.VbPendingOpsUpdate[i]
		extendedFields["vb_pending_queue_age"] = extendedBucketStats.Op.Samples.VbPendingQueueAge[i]
		extendedFields["vb_pending_queue_drain"] = extendedBucketStats.Op.Samples.VbPendingQueueDrain[i]
		extendedFields["vb_pending_queue_fill"] = extendedBucketStats.Op.Samples.VbPendingQueueFill[i]
		extendedFields["vb_pending_queue_size"] = extendedBucketStats.Op.Samples.VbPendingQueueSize[i]
		extendedFields["vb_replica_curr_items"] = extendedBucketStats.Op.Samples.VbReplicaCurrItems[i]
		extendedFields["vb_replica_eject"] = extendedBucketStats.Op.Samples.VbReplicaEject[i]
		extendedFields["vb_replica_itm_memory"] = extendedBucketStats.Op.Samples.VbReplicaItmMemory[i]
		extendedFields["vb_replica_meta_data_memory"] = extendedBucketStats.Op.Samples.VbReplicaMetaDataMemory[i]
		extendedFields["vb_replica_num"] = extendedBucketStats.Op.Samples.VbReplicaNum[i]
		extendedFields["vb_replica_num_non_resident"] = extendedBucketStats.Op.Samples.VbReplicaNumNonResident[i]
		extendedFields["vb_replica_ops_create"] = extendedBucketStats.Op.Samples.VbReplicaOpsCreate[i]
		extendedFields["vb_replica_ops_update"] = extendedBucketStats.Op.Samples.VbReplicaOpsUpdate[i]
		extendedFields["vb_replica_queue_age"] = extendedBucketStats.Op.Samples.VbReplicaQueueAge[i]
		extendedFields["vb_replica_queue_drain"] = extendedBucketStats.Op.Samples.VbReplicaQueueDrain[i]
		extendedFields["vb_replica_queue_fill"] = extendedBucketStats.Op.Samples.VbReplicaQueueFill[i]
		extendedFields["vb_replica_queue_size"] = extendedBucketStats.Op.Samples.VbReplicaQueueSize[i]
		extendedFields["vb_total_queue_age"] = extendedBucketStats.Op.Samples.VbTotalQueueAge[i]
		extendedFields["xdc_ops"] = extendedBucketStats.Op.Samples.XdcOps[i]
		extendedFields["allocstall"] = extendedBucketStats.Op.Samples.Allocstall[i]
		extendedFields["cpu_cores_available"] = extendedBucketStats.Op.Samples.CPUCoresAvailable[i]
		extendedFields["cpu_irq_rate"] = extendedBucketStats.Op.Samples.CPUIrqRate[i]
		extendedFields["cpu_stolen_rate"] = extendedBucketStats.Op.Samples.CPUStolenRate[i]
		extendedFields["cpu_sys_rate"] = extendedBucketStats.Op.Samples.CPUSysRate[i]
		extendedFields["cpu_user_rate"] = extendedBucketStats.Op.Samples.CPUUserRate[i]
		extendedFields["cpu_utilization_rate"] = extendedBucketStats.Op.Samples.CPUUtilizationRate[i]
		extendedFields["hibernated_requests"] = extendedBucketStats.Op.Samples.HibernatedRequests[i]
		extendedFields["hibernated_waked"] = extendedBucketStats.Op.Samples.HibernatedWaked[i]
		extendedFields["mem_actual_free"] = extendedBucketStats.Op.Samples.MemActualFree[i]
		extendedFields["mem_actual_used"] = extendedBucketStats.Op.Samples.MemActualUsed[i]
		extendedFields["mem_free"] = extendedBucketStats.Op.Samples.MemFree[i]
		extendedFields["mem_limit"] = extendedBucketStats.Op.Samples.MemLimit[i]
		extendedFields["mem_total"] = extendedBucketStats.Op.Samples.MemTotal[i]
		extendedFields["mem_used_sys"] = extendedBucketStats.Op.Samples.MemUsedSys[i]
		extendedFields["odp_report_failed"] = extendedBucketStats.Op.Samples.OdpReportFailed[i]
		extendedFields["rest_requests"] = extendedBucketStats.Op.Samples.RestRequests[i]
		extendedFields["swap_total"] = extendedBucketStats.Op.Samples.SwapTotal[i]
		extendedFields["swap_used"] = extendedBucketStats.Op.Samples.SwapUsed[i]
		acc.AddFields("couchbase_bucket", extendedFields, tags, timestamp)

		// Set timestamp back by the couchbase metric interval, so that we have an accurate timestamp for the previous set of metrics.
		timestamp = timestamp.Add(-1 * couchbase.couchbaseTimeIntervals[couchbase.zoom])

		// If we've set the timestamp back before the last interval time, we've collected all the metrics since then. Time to stop.
		if timestamp.Before(couchbase.lastInterval) || timestamp.Equal(couchbase.lastInterval) {
			break
		}
	}
}

func (couchbase *Couchbase) extendedBucketStats(server, bucket string, bucketStats *BucketStats) error {
	// Set up an HTTP request to get the complete set of bucket stats.
	req, err := http.NewRequest("GET", server+"/pools/default/buckets/"+bucket+"/stats?", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()

	couchbase.zoom = couchbase.resolveZoomLevel()
	q.Add("zoom", couchbase.zoom)
	req.URL.RawQuery = q.Encode()

	r, err := client.Do(req)
	if err != nil {
		return err
	}

	// Set last interval so we know how far back to look into the metrics at the next run.
	couchbase.lastInterval = time.Now()
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(bucketStats)
}

func (couchbase *Couchbase) resolveZoomLevel() string {
	timeSinceLastInterval := time.Now().Sub(couchbase.lastInterval)
	if timeSinceLastInterval <= (59 * time.Second) {
		return "minute"
	} else if timeSinceLastInterval <= (59 * time.Minute) {
		return "hour"
	} else if timeSinceLastInterval <= (23*time.Hour + 59*time.Minute) {
		return "day"
	} else if timeSinceLastInterval <= (191*time.Hour + 59*time.Minute) {
		// To couchbase, a week is 8 days. So we have the cutoff be just before 8 days.
		return "week"
	} else {
		return "year"
	}
}

func (couchbase *Couchbase) Init() error {
	couchbase.couchbaseTimeIntervals = map[string]time.Duration{
		"minute": 1 * time.Second,
		"hour":   4 * time.Second,
		"day":    1 * time.Minute,
		"week":   10 * time.Minute,
		"year":   6 * time.Hour,
	}

	return nil
}

func init() {
	inputs.Add("couchbase", func() telegraf.Input {
		return &Couchbase{}
	})
}
