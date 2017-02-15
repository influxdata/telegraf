package couchbase

import (
	"sync"

	couchbase "github.com/couchbase/go-couchbase"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Couchbase struct {
	Servers []string
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    http://couchbase-0.example.com/
  ##    http://admin:secret@couchbase-0.example.com:8091/
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no protocol is specifed, HTTP is used.
  ## If no port is specified, 8091 is used.
  servers = ["http://localhost:8091"]
`

func (r *Couchbase) SampleConfig() string {
	return sampleConfig
}

func (r *Couchbase) Description() string {
	return "Read metrics from one or many couchbase clusters"
}

// Reads stats from all configured clusters. Accumulates stats.
// Returns one of the errors encountered while gathering stats (if any).
func (r *Couchbase) Gather(acc telegraf.Accumulator) error {
	if len(r.Servers) == 0 {
		r.gatherServer("http://localhost:8091/", acc, nil)
		return nil
	}

	var wg sync.WaitGroup

	var outerr error

	for _, serv := range r.Servers {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = r.gatherServer(serv, acc, nil)
		}(serv)
	}

	wg.Wait()

	return outerr
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
		tags := map[string]string{"cluster": addr, "hostname": node.Hostname}
		fields := make(map[string]interface{})
		fields["memory_free"] = node.MemoryFree
		fields["memory_total"] = node.MemoryTotal
		acc.AddFields("couchbase_node", fields, tags)
	}
	for bucketName, _ := range pool.BucketMap {
		tags := map[string]string{"cluster": addr, "bucket": bucketName}
		bs := pool.BucketMap[bucketName].BasicStats
		fields := make(map[string]interface{})
		fields["quota_percent_used"] = bs["quotaPercentUsed"]
		fields["ops_per_sec"] = bs["opsPerSec"]
		fields["disk_fetches"] = bs["diskFetches"]
		fields["item_count"] = bs["itemCount"]
		fields["disk_used"] = bs["diskUsed"]
		fields["data_used"] = bs["dataUsed"]
		fields["mem_used"] = bs["memUsed"]

		fields["ep_dcp_total_queue"] = bs["epDcpTotalQueue"] //Items within the DCP Queue
		fields["vb_active_num"] = bs["vbActiveNum"]          //Active and Replica vBucket Count
		fields["vb_replica_num"] = bs["vb_ReplicaNum"]       //Active and Replica vBucket Count
		fields["curr_items"] = bs["currItems"]
		fields["ep_bg_fetched"] = bs["epBgFetched"]                             //Number of items fetched from disk (cache misses).
		fields["vb_active_perc_mem_resident"] = bs["vbActivePercMemResident"]   //Percent of active data in a vBucket that is memory resident.
		fields["vb_replica_perc_mem_resident"] = bs["vbReplicaPercMemResident"] //Percent of replica data in a vBucket that is memory resident
		fields["ep_tmp_oom_errors"] = bs["epTmpOomErrors"]                      //Number of times temporary OOMs were sent to a client.  Represents high transient memory pressure within the system.
		fields["ep_oom_errors"] = bs["epOomErrors"]                             //Number of times permanent OOMs were sent to a client.  Represents very high consistent memory pressure within the system.
		fields["ep_queue_size"] = bs["epQueueSize"]                             //The amount of data waiting to be written to disk.
		fields["ep_flusher_todo"] = bs["epFlusherTodo"]                         //The number of items currently being written to disk.
		fields["ep_io_num_read"] = bs["epIoNumRead"]                            //The number of read operations sent to disk.
		fields["ep_io_num_write"] = bs["epIoNumWrite"]                          //The number of write operations sent to disk.
		fields["ep_mem_high_wat "] = bs["epMemHighWatermark "]
		fields["cmd_get"] = bs["cmdGet"]
		fields["ep_kv_size"] = bs["epKvSize"]
		acc.AddFields("couchbase_bucket", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("couchbase", func() telegraf.Input {
		return &Couchbase{}
	})
}
