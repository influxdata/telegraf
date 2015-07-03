package ceph

import (
	"encoding/json"
	"fmt"
	"github.com/influxdb/telegraf/plugins"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var sampleConfig = `
# Gather metrics for CEPH
# Specify cluster name  
#cluster="ceph"
# Specify CEPH Bin Location
#binLocation="/usr/bin/ceph"
# Specify CEPH Socket Directory
#socketDir="/var/run/ceph"
`

type CephMetrics struct {
	Cluster     string
	BinLocation string
	SocketDir   string
}

func (_ *CephMetrics) SampleConfig() string {
	return sampleConfig
}

func (_ *CephMetrics) Description() string {
	return "Reading Ceph Metrics"
}

func (ceph *CephMetrics) Gather(acc plugins.Accumulator) error {

	if ceph.Cluster == "" {
		ceph.Cluster = "ceph"
	}

	if ceph.BinLocation == "" {
		ceph.BinLocation = "/usr/bin/ceph"
	}

	if ceph.SocketDir == "" {
		ceph.SocketDir = "/var/run/ceph"
	}

	ceph.gatherMetrics(acc)
	return nil
}

func init() {
	plugins.Add("ceph", func() plugins.Plugin {
		return &CephMetrics{}
	})
}

func (ceph *CephMetrics) gatherMetrics(acc plugins.Accumulator) {

	var quorum QuorumStat

	hostname, err := os.Hostname()

	if err != nil {
		return
	}

	if err := ceph.cephCommand(&quorum, "quorum_status"); err != nil {
		return
	}

	ceph.getOSDPerf(acc)

	if strings.TrimSpace(hostname) != strings.TrimSpace(quorum.LeaderName) {
		fmt.Printf("Not a leader: Quorum leader %s, Host %s", quorum.LeaderName, hostname)
		return
	}

	ceph.getCommon(acc)
	ceph.getPool(acc)
	ceph.getPg(acc)
	ceph.getOSDDaemon(acc)

}

func (ceph *CephMetrics) getCommon(acc plugins.Accumulator) {
	var health CephHealth
	var quorum QuorumStat
	var poolStatsList []PoolStats
	var cephDf CephDF
	var cephStatus CephStatus

	if err := ceph.cephCommand(&cephStatus, "status"); err != nil {
		return
	}

	if err := ceph.cephCommand(&health, "health"); err != nil {
		return
	}

	if err := ceph.cephCommand(&quorum, "quorum_status"); err != nil {
		return
	}

	if err := ceph.cephCommand(&poolStatsList, "osd", "pool", "stats"); err != nil {
		return
	}

	if err := ceph.cephCommand(&cephDf, "df"); err != nil {
		return
	}

	tags := map[string]string{"cluster": ceph.Cluster}

	//Monitors
	monitors := quorum.MonitorMap.Mons
	monitorNames := make([]string, len(monitors))
	monitorValueMap := make(map[string]interface{})
	monitorValueMap["value"] = len(monitors)

	for i, value := range monitors {
		monitorNames[i] = value.Name
	}

	monitorValueMap["name"] = strings.Join(monitorNames, ",")

	//Quorum Names
	quorum_name := quorum.QuorumName
	quorumValueMap := make(map[string]interface{})
	quorumValueMap["value"] = len(quorum_name)
	quorumValueMap["members"] = strings.Join(quorum_name, ",")

	//clientIOs
	sumOps := int64(0)
	sumWrs := int64(0)
	for _, stat := range poolStatsList {
		sumOps += int64(stat.ClientIoRate.OpsPerSec)
		sumWrs += int64(stat.ClientIoRate.WriteBytesPerSecond) / 1024
	}

	// OSD Epoch
	epoch := cephStatus.OSDMap.OSDMap.Epoch
	acc.Add("osd_epoch", epoch, map[string]string{"cluster": ceph.Cluster})
	acc.Add("health", health.OverallStatus, tags)
	acc.Add("total_storage", cephDf.Stats.TotalBytes, tags)
	acc.Add("used_storage", cephDf.Stats.TotalUsedBytes, tags)
	acc.Add("available_storage", cephDf.Stats.TotalAvailableBytes, tags)
	acc.Add("client_io_kbs", sumWrs, tags)
	acc.Add("client_io_ops", sumOps, tags)
	acc.AddValuesWithTime("monitor", monitorValueMap, tags, time.Now())
	acc.AddValuesWithTime("quorum", quorumValueMap, tags, time.Now())
}

func (ceph *CephMetrics) getPool(acc plugins.Accumulator) {
	var cephDf CephDF
	var poolStatsList []PoolStats
	var pgDump PgDump

	if err := ceph.cephCommand(&poolStatsList, "osd", "pool", "stats"); err != nil {
		return
	}

	if err := ceph.cephCommand(&cephDf, "df"); err != nil {
		return
	}

	if err := ceph.cephCommand(&pgDump, "pg", "dump"); err != nil {
		return
	}

	for _, pool := range cephDf.Pools {
		poolId := pool.Id
		numObjects := pool.Stats.Objects
		numBytes := pool.Stats.UsedBytes
		maxAvail := pool.Stats.Available
		numKb := pool.Stats.UsedKb

		utilized := 0.0
		if numBytes != 0 {
			utilized = (float64(numBytes) / float64(maxAvail)) * 100.0
		}

		var quota PoolQuota

		err := ceph.cephCommand(&quota, "osd", "pool", "get-quota", pool.Name)
		if err != nil {
			continue
		}
		maxObjects := quota.MaxObjects
		maxBytes := quota.MaxBytes

		tags := map[string]string{"cluster": ceph.Cluster, "pool": fmt.Sprintf("%d", poolId)}
		acc.Add("pool_objects", numObjects, tags)
		acc.Add("pool_used", numBytes, tags)
		acc.Add("pool_usedKb", numKb, tags)
		acc.Add("pool_max_objects", maxObjects, tags)
		acc.Add("pool_maxbytes", maxBytes, tags)
		acc.Add("pool_utilization", utilized, tags)
	}

	acc.Add("pool", fmt.Sprintf("%d", len(cephDf.Pools)), map[string]string{"cluster": ceph.Cluster})

	for _, stat := range poolStatsList {
		poolId := stat.PoolId
		kbs := stat.ClientIoRate.WriteBytesPerSecond / 1024
		ops := stat.ClientIoRate.OpsPerSec

		tags := map[string]string{"cluster": ceph.Cluster, "pool": fmt.Sprintf("%d", poolId)}
		acc.Add("pool_io_kbs", kbs, tags)
		acc.Add("pool_io_ops", ops, tags)
	}

	for _, pgPoolStat := range pgDump.PoolStats {
		tags := map[string]string{"cluster": ceph.Cluster, "pool": fmt.Sprintf("%d", pgPoolStat.PoolId)}
		for k, v := range pgPoolStat.StatSum {
			acc.Add(fmt.Sprintf("pool_%s", k), fmt.Sprint(v), tags)
		}
	}

}

func (ceph *CephMetrics) getPg(acc plugins.Accumulator) {

	var cephStatus CephStatus
	if err := ceph.cephCommand(&cephStatus, "status"); err != nil {
		return
	}

	pgMap := cephStatus.PgMap

	for _, value := range pgMap.PgByState {
		tags := map[string]string{"cluster": ceph.Cluster, "state": value.Name}
		acc.Add("pg_count", value.Count, tags)
	}

	tags := map[string]string{"cluster": ceph.Cluster}
	acc.Add("pg_map_count", pgMap.PgCount, tags)
	acc.Add("pg_data_bytes", pgMap.DataBytes, tags)
	acc.Add("pg_data_available_storage", pgMap.BytesAvail, tags)
	acc.Add("pg_data_total_storage", pgMap.BytesTotal, tags)
	acc.Add("pg_data_used_storage", pgMap.BytesUsed, tags)

	var pgDump PgDump
	if err := ceph.cephCommand(&pgDump, "pg", "dump"); err != nil {
		return
	}

	poolOsdPgMap := make(PoolOsdPgMap, len(pgDump.PoolStats))
	totalOsdPgs := make(map[int64]int64, len(pgDump.OsdStats))

	for _, pgStat := range pgDump.PgStats {
		poolId, _ := strconv.ParseInt(strings.Split(pgStat.PgId, ".")[0], 10, 64)

		osdPgMap := poolOsdPgMap[poolId]
		if osdPgMap == nil {
			osdPgMap = make(map[int64]int64, len(pgDump.OsdStats))
			poolOsdPgMap[poolId] = osdPgMap
		}

		for _, osd := range pgStat.Up {
			osdPgMap[osd] = int64(osdPgMap[osd] + 1)
			totalOsdPgs[osd] = int64(totalOsdPgs[osd] + 1)
		}
	}

	for poolId, osdPgMap := range poolOsdPgMap {
		poolPg := int64(0)
		for osdId, pgs := range osdPgMap {
			tags := map[string]string{"cluster": ceph.Cluster, "pool": fmt.Sprintf("%d", poolId), "osd": fmt.Sprintf("%d", osdId)}
			poolPg += pgs
			acc.Add("pg_distribution", pgs, tags)
		}

		tags := map[string]string{"cluster": ceph.Cluster, "pool": fmt.Sprintf("%d", poolId)}
		acc.Add("pg_distribution_pool", poolPg, tags)
	}

	for osd, pg := range totalOsdPgs {
		tags := map[string]string{"cluster": ceph.Cluster, "osd": fmt.Sprintf("%d", osd)}
		acc.Add("pg_distribution_osd", pg, tags)
	}

	clusterTag := map[string]string{"cluster": ceph.Cluster}
	for k, v := range pgDump.PgStatSum.StatSum {
		acc.Add(fmt.Sprintf("pg_stats_%s", k), fmt.Sprintf("%d", v), clusterTag)
	}

}

func (ceph *CephMetrics) getOSDDaemon(acc plugins.Accumulator) {

	var osd OsdDump
	var osdPerf OsdPerf
	var pgDump PgDump

	if err := ceph.cephCommand(&pgDump, "pg", "dump"); err != nil {
		return
	}

	if err := ceph.cephCommand(&osdPerf, "osd", "perf"); err != nil {
		return
	}

	if err := ceph.cephCommand(&osd, "osd", "dump"); err != nil {
		return
	}

	up := 0
	in := 0
	down := 0
	out := 0
	osds := osd.Osds

	for _, osdStat := range osds {

		if osdStat.Up == 1 {
			up += 1
		} else {
			down += 1
		}

		if osdStat.In == 1 {
			in += 1
		} else {
			out += 1
		}
	}

	acc.Add("osd_count", len(osd.Osds), map[string]string{"cluster": ceph.Cluster})
	acc.Add("osd_count", in, map[string]string{"cluster": ceph.Cluster, "state": "in"})
	acc.Add("osd_count", out, map[string]string{"cluster": ceph.Cluster, "state": "out"})
	acc.Add("osd_count", up, map[string]string{"cluster": ceph.Cluster, "state": "up"})
	acc.Add("osd_count", down, map[string]string{"cluster": ceph.Cluster, "state": "down"})

	// OSD Utilization
	for _, osdStat := range pgDump.OsdStats {
		osdNum := osdStat.Osd
		used := float64(osdStat.UsedKb)
		total := float64(osdStat.TotalKb)
		utilized := (used / total) * 100.0

		tag := map[string]string{"cluster": ceph.Cluster, "osd": fmt.Sprintf("%d", osdNum)}
		acc.Add("osd_utilization", utilized, tag)
		acc.Add("osd_used", utilized, tag)
		acc.Add("osd_total", total, tag)
	}

	//OSD Commit and Apply Latency
	for _, perf := range osdPerf.PerfInfo {
		osdNum := perf.Id
		commit := perf.Stats.CommitLatency
		apply := perf.Stats.ApplyLatency

		acc.Add("osd_latency_commit", commit, map[string]string{"cluster": ceph.Cluster, "osd": fmt.Sprintf("%d", osdNum)})
		acc.Add("osd_latency_apply", apply, map[string]string{"cluster": ceph.Cluster, "osd": fmt.Sprintf("%d", osdNum)})
	}

}

func (ceph *CephMetrics) getOSDPerf(acc plugins.Accumulator) {
	var osdPerf OsdPerfDump

	osdsArray, err := ioutil.ReadDir(ceph.SocketDir)
	if err != nil {
		return
	}

	for _, file := range osdsArray {
		name := file.Name()

		if !strings.HasPrefix(name, ceph.Cluster+"-osd") {
			continue
		}

		location := fmt.Sprint(ceph.SocketDir, "/", name)
		args := []string{"--admin-daemon", location, "perf", "dump"}

		if err := ceph.cephCommand(&osdPerf, args...); err != nil {
			return
		}

		osdId := string(name[strings.LastIndex(name, ".")-1])
		tags := map[string]string{"cluster": ceph.Cluster, "osd": osdId}
		osd := osdPerf.Osd

		// osd-<id>.osd.recovery_ops ?
		acc.Add("op_wip", osd.OpWip, tags)
		acc.Add("op", osd.Op, tags)
		acc.Add("op_in_bytes", osd.OpInBytes, tags)
		acc.Add("op_out_bytes", osd.OpOutBytes, tags)
		acc.Add("op_r", osd.OpRead, tags)
		acc.Add("op_r_out_bytes", osd.OpReadOutBytes, tags)
		acc.Add("op_w", osd.OpWrite, tags)
		acc.Add("op_w_in_bytes", osd.OpWriteInBytes, tags)
		acc.Add("op_rw", osd.OpReadWrite, tags)
		acc.Add("op_rw_in_bytes", osd.OpReadWriteInBytes, tags)
		acc.Add("op_rw_out_bytes", osd.OpReadWriteOutBytes, tags)
		acc.AddValuesWithTime("op_latency", getOSDLatencyCalc(&osd.OpLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_process_latency", getOSDLatencyCalc(&osd.OpProcessLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_r", getOSDLatencyCalc(&osd.OpReadLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_r_process_latency", getOSDLatencyCalc(&osd.OpReadProcessLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_w_latency", getOSDLatencyCalc(&osd.OpWriteLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_w_process_latency", getOSDLatencyCalc(&osd.OpWriteProcessLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_rw_latency", getOSDLatencyCalc(&osd.OpReadWriteLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_rw_process_latency", getOSDLatencyCalc(&osd.OpReadWriteProcessLatency.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_rw_rlat", getOSDLatencyCalc(&osd.OpReadWriteRlat.OSDLatencyCalc), tags, time.Now())
		acc.AddValuesWithTime("op_w_rlat", getOSDLatencyCalc(&osd.OpWriteRlat.OSDLatencyCalc), tags, time.Now())
	}
}

func getOSDLatencyCalc(osdLatency *OSDLatency) map[string]interface{} {
	latencyMap := make(map[string]interface{})
	latencyMap["avgcount"] = osdLatency.AvgCount
	latencyMap["sum"] = osdLatency.Sum
	return latencyMap
}

func (ceph *CephMetrics) cephCommand(v interface{}, args ...string) error {
	args = append(args, "-f", "json", "--cluster="+ceph.Cluster)
	out, err := exec.Command(ceph.BinLocation, args...).Output()
	if err != nil {
		return err
	}
	return json.Unmarshal(out, v)
}
