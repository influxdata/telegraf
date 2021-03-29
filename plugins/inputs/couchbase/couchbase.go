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

	BucketStatsIncluded []string `toml:"bucket_stats_included"`

	bucketInclude filter.Filter
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

  ## Filter fields to include only here.
  # bucket_stats_included = ["quota_percent_used", "ops_per_sec", "disk_fetches", "item_count", "disk_used", "data_used", "mem_used"]
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
		fields["memory_free"] = node.MemoryFree
		fields["memory_total"] = node.MemoryTotal
		acc.AddFields("couchbase_node", fields, tags)
	}

	for bucketName := range pool.BucketMap {
		tags := map[string]string{"cluster": regexpURI.ReplaceAllString(addr, "${1}"), "bucket": bucketName}
		bs := pool.BucketMap[bucketName].BasicStats
		fields := make(map[string]interface{})
		cb.addBucketField(fields, "quota_percent_used", bs["quotaPercentUsed"])
		cb.addBucketField(fields, "ops_per_sec", bs["opsPerSec"])
		cb.addBucketField(fields, "disk_fetches", bs["diskFetches"])
		cb.addBucketField(fields, "item_count", bs["itemCount"])
		cb.addBucketField(fields, "disk_used", bs["diskUsed"])
		cb.addBucketField(fields, "data_used", bs["dataUsed"])
		cb.addBucketField(fields, "mem_used", bs["memUsed"])

		err := cb.gatherDetailedBucketStats(addr, bucketName, fields)
		if err != nil {
			return err
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
	cb.addBucketField(fields, "couch_total_disk_size", extendedBucketStats.Op.Samples.CouchTotalDiskSize[lastEntry])
	cb.addBucketField(fields, "couch_docs_fragmentation", extendedBucketStats.Op.Samples.CouchDocsFragmentation[lastEntry])
	cb.addBucketField(fields, "couch_views_fragmentation", extendedBucketStats.Op.Samples.CouchViewsFragmentation[lastEntry])
	cb.addBucketField(fields, "hit_ratio", extendedBucketStats.Op.Samples.HitRatio[lastEntry])
	cb.addBucketField(fields, "ep_cache_miss_rate", extendedBucketStats.Op.Samples.EpCacheMissRate[lastEntry])
	cb.addBucketField(fields, "ep_resident_items_rate", extendedBucketStats.Op.Samples.EpResidentItemsRate[lastEntry])
	cb.addBucketField(fields, "vb_avg_active_queue_age", extendedBucketStats.Op.Samples.VbAvgActiveQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_avg_replica_queue_age", extendedBucketStats.Op.Samples.VbAvgReplicaQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_avg_pending_queue_age", extendedBucketStats.Op.Samples.VbAvgPendingQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_avg_total_queue_age", extendedBucketStats.Op.Samples.VbAvgTotalQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_active_resident_items_ratio", extendedBucketStats.Op.Samples.VbActiveResidentItemsRatio[lastEntry])
	cb.addBucketField(fields, "vb_replica_resident_items_ratio", extendedBucketStats.Op.Samples.VbReplicaResidentItemsRatio[lastEntry])
	cb.addBucketField(fields, "vb_pending_resident_items_ratio", extendedBucketStats.Op.Samples.VbPendingResidentItemsRatio[lastEntry])
	cb.addBucketField(fields, "avg_disk_update_time", extendedBucketStats.Op.Samples.AvgDiskUpdateTime[lastEntry])
	cb.addBucketField(fields, "avg_disk_commit_time", extendedBucketStats.Op.Samples.AvgDiskCommitTime[lastEntry])
	cb.addBucketField(fields, "avg_bg_wait_time", extendedBucketStats.Op.Samples.AvgBgWaitTime[lastEntry])
	cb.addBucketField(fields, "avg_active_timestamp_drift", extendedBucketStats.Op.Samples.AvgActiveTimestampDrift[lastEntry])
	cb.addBucketField(fields, "avg_replica_timestamp_drift", extendedBucketStats.Op.Samples.AvgReplicaTimestampDrift[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_count", extendedBucketStats.Op.Samples.EpDcpViewsIndexesCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_items_remaining", extendedBucketStats.Op.Samples.EpDcpViewsIndexesItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_producer_count", extendedBucketStats.Op.Samples.EpDcpViewsIndexesProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpViewsIndexesTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_items_sent", extendedBucketStats.Op.Samples.EpDcpViewsIndexesItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_total_bytes", extendedBucketStats.Op.Samples.EpDcpViewsIndexesTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views+indexes_backoff", extendedBucketStats.Op.Samples.EpDcpViewsIndexesBackoff[lastEntry])
	cb.addBucketField(fields, "bg_wait_count", extendedBucketStats.Op.Samples.BgWaitCount[lastEntry])
	cb.addBucketField(fields, "bg_wait_total", extendedBucketStats.Op.Samples.BgWaitTotal[lastEntry])
	cb.addBucketField(fields, "bytes_read", extendedBucketStats.Op.Samples.BytesRead[lastEntry])
	cb.addBucketField(fields, "bytes_written", extendedBucketStats.Op.Samples.BytesWritten[lastEntry])
	cb.addBucketField(fields, "cas_badval", extendedBucketStats.Op.Samples.CasBadval[lastEntry])
	cb.addBucketField(fields, "cas_hits", extendedBucketStats.Op.Samples.CasHits[lastEntry])
	cb.addBucketField(fields, "cas_misses", extendedBucketStats.Op.Samples.CasMisses[lastEntry])
	cb.addBucketField(fields, "cmd_get", extendedBucketStats.Op.Samples.CmdGet[lastEntry])
	cb.addBucketField(fields, "cmd_lookup", extendedBucketStats.Op.Samples.CmdLookup[lastEntry])
	cb.addBucketField(fields, "cmd_set", extendedBucketStats.Op.Samples.CmdSet[lastEntry])
	cb.addBucketField(fields, "couch_docs_actual_disk_size", extendedBucketStats.Op.Samples.CouchDocsActualDiskSize[lastEntry])
	cb.addBucketField(fields, "couch_docs_data_size", extendedBucketStats.Op.Samples.CouchDocsDataSize[lastEntry])
	cb.addBucketField(fields, "couch_docs_disk_size", extendedBucketStats.Op.Samples.CouchDocsDiskSize[lastEntry])
	cb.addBucketField(fields, "couch_spatial_data_size", extendedBucketStats.Op.Samples.CouchSpatialDataSize[lastEntry])
	cb.addBucketField(fields, "couch_spatial_disk_size", extendedBucketStats.Op.Samples.CouchSpatialDiskSize[lastEntry])
	cb.addBucketField(fields, "couch_spatial_ops", extendedBucketStats.Op.Samples.CouchSpatialOps[lastEntry])
	cb.addBucketField(fields, "couch_views_actual_disk_size", extendedBucketStats.Op.Samples.CouchViewsActualDiskSize[lastEntry])
	cb.addBucketField(fields, "couch_views_data_size", extendedBucketStats.Op.Samples.CouchViewsDataSize[lastEntry])
	cb.addBucketField(fields, "couch_views_disk_size", extendedBucketStats.Op.Samples.CouchViewsDiskSize[lastEntry])
	cb.addBucketField(fields, "couch_views_ops", extendedBucketStats.Op.Samples.CouchViewsOps[lastEntry])
	cb.addBucketField(fields, "curr_connections", extendedBucketStats.Op.Samples.CurrConnections[lastEntry])
	cb.addBucketField(fields, "curr_items", extendedBucketStats.Op.Samples.CurrItems[lastEntry])
	cb.addBucketField(fields, "curr_items_tot", extendedBucketStats.Op.Samples.CurrItemsTot[lastEntry])
	cb.addBucketField(fields, "decr_hits", extendedBucketStats.Op.Samples.DecrHits[lastEntry])
	cb.addBucketField(fields, "decr_misses", extendedBucketStats.Op.Samples.DecrMisses[lastEntry])
	cb.addBucketField(fields, "delete_hits", extendedBucketStats.Op.Samples.DeleteHits[lastEntry])
	cb.addBucketField(fields, "delete_misses", extendedBucketStats.Op.Samples.DeleteMisses[lastEntry])
	cb.addBucketField(fields, "disk_commit_count", extendedBucketStats.Op.Samples.DiskCommitCount[lastEntry])
	cb.addBucketField(fields, "disk_commit_total", extendedBucketStats.Op.Samples.DiskCommitTotal[lastEntry])
	cb.addBucketField(fields, "disk_update_count", extendedBucketStats.Op.Samples.DiskUpdateCount[lastEntry])
	cb.addBucketField(fields, "disk_update_total", extendedBucketStats.Op.Samples.DiskUpdateTotal[lastEntry])
	cb.addBucketField(fields, "disk_write_queue", extendedBucketStats.Op.Samples.DiskWriteQueue[lastEntry])
	cb.addBucketField(fields, "ep_active_ahead_exceptions", extendedBucketStats.Op.Samples.EpActiveAheadExceptions[lastEntry])
	cb.addBucketField(fields, "ep_active_hlc_drift", extendedBucketStats.Op.Samples.EpActiveHlcDrift[lastEntry])
	cb.addBucketField(fields, "ep_active_hlc_drift_count", extendedBucketStats.Op.Samples.EpActiveHlcDriftCount[lastEntry])
	cb.addBucketField(fields, "ep_bg_fetched", extendedBucketStats.Op.Samples.EpBgFetched[lastEntry])
	cb.addBucketField(fields, "ep_clock_cas_drift_threshold_exceeded", extendedBucketStats.Op.Samples.EpClockCasDriftThresholdExceeded[lastEntry])
	cb.addBucketField(fields, "ep_data_read_failed", extendedBucketStats.Op.Samples.EpDataReadFailed[lastEntry])
	cb.addBucketField(fields, "ep_data_write_failed", extendedBucketStats.Op.Samples.EpDataWriteFailed[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_backoff", extendedBucketStats.Op.Samples.EpDcp2IBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_count", extendedBucketStats.Op.Samples.EpDcp2ICount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_items_remaining", extendedBucketStats.Op.Samples.EpDcp2IItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_items_sent", extendedBucketStats.Op.Samples.EpDcp2IItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_producer_count", extendedBucketStats.Op.Samples.EpDcp2IProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_total_backlog_size", extendedBucketStats.Op.Samples.EpDcp2ITotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_2i_total_bytes", extendedBucketStats.Op.Samples.EpDcp2ITotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_backoff", extendedBucketStats.Op.Samples.EpDcpCbasBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_count", extendedBucketStats.Op.Samples.EpDcpCbasCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_items_remaining", extendedBucketStats.Op.Samples.EpDcpCbasItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_items_sent", extendedBucketStats.Op.Samples.EpDcpCbasItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_producer_count", extendedBucketStats.Op.Samples.EpDcpCbasProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpCbasTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_cbas_total_bytes", extendedBucketStats.Op.Samples.EpDcpCbasTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_backoff", extendedBucketStats.Op.Samples.EpDcpEventingBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_count", extendedBucketStats.Op.Samples.EpDcpEventingCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_items_remaining", extendedBucketStats.Op.Samples.EpDcpEventingItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_items_sent", extendedBucketStats.Op.Samples.EpDcpEventingItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_producer_count", extendedBucketStats.Op.Samples.EpDcpEventingProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpEventingTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_eventing_total_bytes", extendedBucketStats.Op.Samples.EpDcpEventingTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_backoff", extendedBucketStats.Op.Samples.EpDcpFtsBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_count", extendedBucketStats.Op.Samples.EpDcpFtsCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_items_remaining", extendedBucketStats.Op.Samples.EpDcpFtsItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_items_sent", extendedBucketStats.Op.Samples.EpDcpFtsItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_producer_count", extendedBucketStats.Op.Samples.EpDcpFtsProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpFtsTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_fts_total_bytes", extendedBucketStats.Op.Samples.EpDcpFtsTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_backoff", extendedBucketStats.Op.Samples.EpDcpOtherBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_count", extendedBucketStats.Op.Samples.EpDcpOtherCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_items_remaining", extendedBucketStats.Op.Samples.EpDcpOtherItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_items_sent", extendedBucketStats.Op.Samples.EpDcpOtherItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_producer_count", extendedBucketStats.Op.Samples.EpDcpOtherProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpOtherTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_other_total_bytes", extendedBucketStats.Op.Samples.EpDcpOtherTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_backoff", extendedBucketStats.Op.Samples.EpDcpReplicaBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_count", extendedBucketStats.Op.Samples.EpDcpReplicaCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_items_remaining", extendedBucketStats.Op.Samples.EpDcpReplicaItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_items_sent", extendedBucketStats.Op.Samples.EpDcpReplicaItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_producer_count", extendedBucketStats.Op.Samples.EpDcpReplicaProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpReplicaTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_replica_total_bytes", extendedBucketStats.Op.Samples.EpDcpReplicaTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_backoff", extendedBucketStats.Op.Samples.EpDcpViewsBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_count", extendedBucketStats.Op.Samples.EpDcpViewsCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_items_remaining", extendedBucketStats.Op.Samples.EpDcpViewsItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_items_sent", extendedBucketStats.Op.Samples.EpDcpViewsItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_producer_count", extendedBucketStats.Op.Samples.EpDcpViewsProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpViewsTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_views_total_bytes", extendedBucketStats.Op.Samples.EpDcpViewsTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_backoff", extendedBucketStats.Op.Samples.EpDcpXdcrBackoff[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_count", extendedBucketStats.Op.Samples.EpDcpXdcrCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_items_remaining", extendedBucketStats.Op.Samples.EpDcpXdcrItemsRemaining[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_items_sent", extendedBucketStats.Op.Samples.EpDcpXdcrItemsSent[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_producer_count", extendedBucketStats.Op.Samples.EpDcpXdcrProducerCount[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_total_backlog_size", extendedBucketStats.Op.Samples.EpDcpXdcrTotalBacklogSize[lastEntry])
	cb.addBucketField(fields, "ep_dcp_xdcr_total_bytes", extendedBucketStats.Op.Samples.EpDcpXdcrTotalBytes[lastEntry])
	cb.addBucketField(fields, "ep_diskqueue_drain", extendedBucketStats.Op.Samples.EpDiskqueueDrain[lastEntry])
	cb.addBucketField(fields, "ep_diskqueue_fill", extendedBucketStats.Op.Samples.EpDiskqueueFill[lastEntry])
	cb.addBucketField(fields, "ep_diskqueue_items", extendedBucketStats.Op.Samples.EpDiskqueueItems[lastEntry])
	cb.addBucketField(fields, "ep_flusher_todo", extendedBucketStats.Op.Samples.EpFlusherTodo[lastEntry])
	cb.addBucketField(fields, "ep_item_commit_failed", extendedBucketStats.Op.Samples.EpItemCommitFailed[lastEntry])
	cb.addBucketField(fields, "ep_kv_size", extendedBucketStats.Op.Samples.EpKvSize[lastEntry])
	cb.addBucketField(fields, "ep_max_size", extendedBucketStats.Op.Samples.EpMaxSize[lastEntry])
	cb.addBucketField(fields, "ep_mem_high_wat", extendedBucketStats.Op.Samples.EpMemHighWat[lastEntry])
	cb.addBucketField(fields, "ep_mem_low_wat", extendedBucketStats.Op.Samples.EpMemLowWat[lastEntry])
	cb.addBucketField(fields, "ep_meta_data_memory", extendedBucketStats.Op.Samples.EpMetaDataMemory[lastEntry])
	cb.addBucketField(fields, "ep_num_non_resident", extendedBucketStats.Op.Samples.EpNumNonResident[lastEntry])
	cb.addBucketField(fields, "ep_num_ops_del_meta", extendedBucketStats.Op.Samples.EpNumOpsDelMeta[lastEntry])
	cb.addBucketField(fields, "ep_num_ops_del_ret_meta", extendedBucketStats.Op.Samples.EpNumOpsDelRetMeta[lastEntry])
	cb.addBucketField(fields, "ep_num_ops_get_meta", extendedBucketStats.Op.Samples.EpNumOpsGetMeta[lastEntry])
	cb.addBucketField(fields, "ep_num_ops_set_meta", extendedBucketStats.Op.Samples.EpNumOpsSetMeta[lastEntry])
	cb.addBucketField(fields, "ep_num_ops_set_ret_meta", extendedBucketStats.Op.Samples.EpNumOpsSetRetMeta[lastEntry])
	cb.addBucketField(fields, "ep_num_value_ejects", extendedBucketStats.Op.Samples.EpNumValueEjects[lastEntry])
	cb.addBucketField(fields, "ep_oom_errors", extendedBucketStats.Op.Samples.EpOomErrors[lastEntry])
	cb.addBucketField(fields, "ep_ops_create", extendedBucketStats.Op.Samples.EpOpsCreate[lastEntry])
	cb.addBucketField(fields, "ep_ops_update", extendedBucketStats.Op.Samples.EpOpsUpdate[lastEntry])
	cb.addBucketField(fields, "ep_overhead", extendedBucketStats.Op.Samples.EpOverhead[lastEntry])
	cb.addBucketField(fields, "ep_queue_size", extendedBucketStats.Op.Samples.EpQueueSize[lastEntry])
	cb.addBucketField(fields, "ep_replica_ahead_exceptions", extendedBucketStats.Op.Samples.EpReplicaAheadExceptions[lastEntry])
	cb.addBucketField(fields, "ep_replica_hlc_drift", extendedBucketStats.Op.Samples.EpReplicaHlcDrift[lastEntry])
	cb.addBucketField(fields, "ep_replica_hlc_drift_count", extendedBucketStats.Op.Samples.EpReplicaHlcDriftCount[lastEntry])
	cb.addBucketField(fields, "ep_tmp_oom_errors", extendedBucketStats.Op.Samples.EpTmpOomErrors[lastEntry])
	cb.addBucketField(fields, "ep_vb_total", extendedBucketStats.Op.Samples.EpVbTotal[lastEntry])
	cb.addBucketField(fields, "evictions", extendedBucketStats.Op.Samples.Evictions[lastEntry])
	cb.addBucketField(fields, "get_hits", extendedBucketStats.Op.Samples.GetHits[lastEntry])
	cb.addBucketField(fields, "get_misses", extendedBucketStats.Op.Samples.GetMisses[lastEntry])
	cb.addBucketField(fields, "incr_hits", extendedBucketStats.Op.Samples.IncrHits[lastEntry])
	cb.addBucketField(fields, "incr_misses", extendedBucketStats.Op.Samples.IncrMisses[lastEntry])
	cb.addBucketField(fields, "misses", extendedBucketStats.Op.Samples.Misses[lastEntry])
	cb.addBucketField(fields, "ops", extendedBucketStats.Op.Samples.Ops[lastEntry])
	cb.addBucketField(fields, "timestamp", extendedBucketStats.Op.Samples.Timestamp[lastEntry])
	cb.addBucketField(fields, "vb_active_eject", extendedBucketStats.Op.Samples.VbActiveEject[lastEntry])
	cb.addBucketField(fields, "vb_active_itm_memory", extendedBucketStats.Op.Samples.VbActiveItmMemory[lastEntry])
	cb.addBucketField(fields, "vb_active_meta_data_memory", extendedBucketStats.Op.Samples.VbActiveMetaDataMemory[lastEntry])
	cb.addBucketField(fields, "vb_active_num", extendedBucketStats.Op.Samples.VbActiveNum[lastEntry])
	cb.addBucketField(fields, "vb_active_num_non_resident", extendedBucketStats.Op.Samples.VbActiveNumNonResident[lastEntry])
	cb.addBucketField(fields, "vb_active_ops_create", extendedBucketStats.Op.Samples.VbActiveOpsCreate[lastEntry])
	cb.addBucketField(fields, "vb_active_ops_update", extendedBucketStats.Op.Samples.VbActiveOpsUpdate[lastEntry])
	cb.addBucketField(fields, "vb_active_queue_age", extendedBucketStats.Op.Samples.VbActiveQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_active_queue_drain", extendedBucketStats.Op.Samples.VbActiveQueueDrain[lastEntry])
	cb.addBucketField(fields, "vb_active_queue_fill", extendedBucketStats.Op.Samples.VbActiveQueueFill[lastEntry])
	cb.addBucketField(fields, "vb_active_queue_size", extendedBucketStats.Op.Samples.VbActiveQueueSize[lastEntry])
	cb.addBucketField(fields, "vb_active_sync_write_aborted_count", extendedBucketStats.Op.Samples.VbActiveSyncWriteAbortedCount[lastEntry])
	cb.addBucketField(fields, "vb_active_sync_write_accepted_count", extendedBucketStats.Op.Samples.VbActiveSyncWriteAcceptedCount[lastEntry])
	cb.addBucketField(fields, "vb_active_sync_write_committed_count", extendedBucketStats.Op.Samples.VbActiveSyncWriteCommittedCount[lastEntry])
	cb.addBucketField(fields, "vb_pending_curr_items", extendedBucketStats.Op.Samples.VbPendingCurrItems[lastEntry])
	cb.addBucketField(fields, "vb_pending_eject", extendedBucketStats.Op.Samples.VbPendingEject[lastEntry])
	cb.addBucketField(fields, "vb_pending_itm_memory", extendedBucketStats.Op.Samples.VbPendingItmMemory[lastEntry])
	cb.addBucketField(fields, "vb_pending_meta_data_memory", extendedBucketStats.Op.Samples.VbPendingMetaDataMemory[lastEntry])
	cb.addBucketField(fields, "vb_pending_num", extendedBucketStats.Op.Samples.VbPendingNum[lastEntry])
	cb.addBucketField(fields, "vb_pending_num_non_resident", extendedBucketStats.Op.Samples.VbPendingNumNonResident[lastEntry])
	cb.addBucketField(fields, "vb_pending_ops_create", extendedBucketStats.Op.Samples.VbPendingOpsCreate[lastEntry])
	cb.addBucketField(fields, "vb_pending_ops_update", extendedBucketStats.Op.Samples.VbPendingOpsUpdate[lastEntry])
	cb.addBucketField(fields, "vb_pending_queue_age", extendedBucketStats.Op.Samples.VbPendingQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_pending_queue_drain", extendedBucketStats.Op.Samples.VbPendingQueueDrain[lastEntry])
	cb.addBucketField(fields, "vb_pending_queue_fill", extendedBucketStats.Op.Samples.VbPendingQueueFill[lastEntry])
	cb.addBucketField(fields, "vb_pending_queue_size", extendedBucketStats.Op.Samples.VbPendingQueueSize[lastEntry])
	cb.addBucketField(fields, "vb_replica_curr_items", extendedBucketStats.Op.Samples.VbReplicaCurrItems[lastEntry])
	cb.addBucketField(fields, "vb_replica_eject", extendedBucketStats.Op.Samples.VbReplicaEject[lastEntry])
	cb.addBucketField(fields, "vb_replica_itm_memory", extendedBucketStats.Op.Samples.VbReplicaItmMemory[lastEntry])
	cb.addBucketField(fields, "vb_replica_meta_data_memory", extendedBucketStats.Op.Samples.VbReplicaMetaDataMemory[lastEntry])
	cb.addBucketField(fields, "vb_replica_num", extendedBucketStats.Op.Samples.VbReplicaNum[lastEntry])
	cb.addBucketField(fields, "vb_replica_num_non_resident", extendedBucketStats.Op.Samples.VbReplicaNumNonResident[lastEntry])
	cb.addBucketField(fields, "vb_replica_ops_create", extendedBucketStats.Op.Samples.VbReplicaOpsCreate[lastEntry])
	cb.addBucketField(fields, "vb_replica_ops_update", extendedBucketStats.Op.Samples.VbReplicaOpsUpdate[lastEntry])
	cb.addBucketField(fields, "vb_replica_queue_age", extendedBucketStats.Op.Samples.VbReplicaQueueAge[lastEntry])
	cb.addBucketField(fields, "vb_replica_queue_drain", extendedBucketStats.Op.Samples.VbReplicaQueueDrain[lastEntry])
	cb.addBucketField(fields, "vb_replica_queue_fill", extendedBucketStats.Op.Samples.VbReplicaQueueFill[lastEntry])
	cb.addBucketField(fields, "vb_replica_queue_size", extendedBucketStats.Op.Samples.VbReplicaQueueSize[lastEntry])
	cb.addBucketField(fields, "vb_total_queue_age", extendedBucketStats.Op.Samples.VbTotalQueueAge[lastEntry])
	cb.addBucketField(fields, "xdc_ops", extendedBucketStats.Op.Samples.XdcOps[lastEntry])
	cb.addBucketField(fields, "allocstall", extendedBucketStats.Op.Samples.Allocstall[lastEntry])
	cb.addBucketField(fields, "cpu_cores_available", extendedBucketStats.Op.Samples.CPUCoresAvailable[lastEntry])
	cb.addBucketField(fields, "cpu_irq_rate", extendedBucketStats.Op.Samples.CPUIrqRate[lastEntry])
	cb.addBucketField(fields, "cpu_stolen_rate", extendedBucketStats.Op.Samples.CPUStolenRate[lastEntry])
	cb.addBucketField(fields, "cpu_sys_rate", extendedBucketStats.Op.Samples.CPUSysRate[lastEntry])
	cb.addBucketField(fields, "cpu_user_rate", extendedBucketStats.Op.Samples.CPUUserRate[lastEntry])
	cb.addBucketField(fields, "cpu_utilization_rate", extendedBucketStats.Op.Samples.CPUUtilizationRate[lastEntry])
	cb.addBucketField(fields, "hibernated_requests", extendedBucketStats.Op.Samples.HibernatedRequests[lastEntry])
	cb.addBucketField(fields, "hibernated_waked", extendedBucketStats.Op.Samples.HibernatedWaked[lastEntry])
	cb.addBucketField(fields, "mem_actual_free", extendedBucketStats.Op.Samples.MemActualFree[lastEntry])
	cb.addBucketField(fields, "mem_actual_used", extendedBucketStats.Op.Samples.MemActualUsed[lastEntry])
	cb.addBucketField(fields, "mem_free", extendedBucketStats.Op.Samples.MemFree[lastEntry])
	cb.addBucketField(fields, "mem_limit", extendedBucketStats.Op.Samples.MemLimit[lastEntry])
	cb.addBucketField(fields, "mem_total", extendedBucketStats.Op.Samples.MemTotal[lastEntry])
	cb.addBucketField(fields, "mem_used_sys", extendedBucketStats.Op.Samples.MemUsedSys[lastEntry])
	cb.addBucketField(fields, "odp_report_failed", extendedBucketStats.Op.Samples.OdpReportFailed[lastEntry])
	cb.addBucketField(fields, "rest_requests", extendedBucketStats.Op.Samples.RestRequests[lastEntry])
	cb.addBucketField(fields, "swap_total", extendedBucketStats.Op.Samples.SwapTotal[lastEntry])
	cb.addBucketField(fields, "swap_used", extendedBucketStats.Op.Samples.SwapUsed[lastEntry])

	return nil
}

func (cb *Couchbase) addBucketField(fields map[string]interface{}, fieldKey string, value interface{}) {
	if !cb.bucketInclude.Match(fieldKey) {
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
	f, err := filter.NewIncludeExcludeFilter(cb.BucketStatsIncluded, []string{})
	if err != nil {
		return err
	}

	cb.bucketInclude = f

	return nil
}

func init() {
	inputs.Add("couchbase", func() telegraf.Input {
		return &Couchbase{
			BucketStatsIncluded: []string{"quota_percent_used", "ops_per_sec", "disk_fetches", "item_count", "disk_used", "data_used", "mem_used"},
		}
	})
}
