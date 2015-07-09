package ceph

type QuorumStat struct {
	LeaderName string   `json:"quorum_leader_name"`
	QuorumName []string `json:"quorum_names"`
	MonitorMap struct {
		Epoch int64 `json:"election_epoch"`
		Mons  []struct {
			Name    string `json:"name"`
			Address string `json:"addr"`
		} `json:"mons"`
	} `json:"monmap"`
}

type CephHealth struct {
	OverallStatus string `json:"overall_status"`
}
type CephStatus struct {
	Quorum []int64 `json:"quorum"`
	OSDMap struct {
		OSDMap struct {
			Epoch int64 `json:"epoch"`
		} `json:"osdmap"`
	} `json:"osdmap"`
	Health struct {
		OverallStatus string `json:"overall_status"`
	} `json:"health"`
	PgMap struct {
		PgByState []struct {
			Name  string `json:"state_name"`
			Count int64  `json:"count"`
		} `json:"pgs_by_state"`
		PgCount    int64 `json:"num_pgs"`
		DataBytes  int64 `json:"data_bytes"`
		BytesUsed  int64 `json:"bytes_used"`
		BytesAvail int64 `json:"bytes_avail"`
		BytesTotal int64 `json:"bytes_total"`
	} `json:"pgmap"`
}

type CephDF struct {
	Stats struct {
		TotalBytes          int64 `json:"total_bytes"`
		TotalUsedBytes      int64 `json:"total_used_bytes"`
		TotalAvailableBytes int64 `json:"total_avail_bytes"`
	} `json:"stats"`
	Pools []struct {
		Name  string `json:"name"`
		Id    int64  `json:"id"`
		Stats struct {
			UsedKb    int64 `json:"kb_used"`
			UsedBytes int64 `json:"bytes_used"`
			Available int64 `json:"max_avail"`
			Objects   int64 `json::"objects"`
		} `json:"stats"`
	} `json:"pools"`
}

type PoolStats struct {
	PoolName     string `json:"pool_name"`
	PoolId       int64  `json:"pool_id"`
	ClientIoRate struct {
		WriteBytesPerSecond int64 `json:"write_bytes_sec"`
		OpsPerSec           int64 `json:"op_per_sec"`
	} `json:"client_io_rate"`
}

type PoolQuota struct {
	PoolName   string `json:"pool_name"`
	PoolId     int64  `json:"pool_id"`
	MaxObjects int64  `json:"quota_max_objects"`
	MaxBytes   int64  `json:"quota_max_bytes"`
}

type OsdDump struct {
	Osds []struct {
		OsdNum   int64    `json:"osd"`
		Uuid     string   `json:"uuid"`
		Up       int64    `json:"up"`
		In       int64    `json:"in"`
		OsdState []string `json:"state"`
	} `json:"osds"`
}

type OsdPerf struct {
	PerfInfo []struct {
		Id    int64 `json:"id"`
		Stats struct {
			CommitLatency int64 `json:"commit_latency_ms"`
			ApplyLatency  int64 `json:"apply_latency_ms"`
		} `json:"perf_stats"`
	} `json:"osd_perf_infos"`
}

type PgDump struct {
	PgStatSum struct {
		StatSum map[string]int64 `json:"stat_sum"`
	} `json:"pg_stats_sum"`
	PoolStats []struct {
		PoolId  int64                  `json:"poolid"`
		StatSum map[string]interface{} `json:"stat_sum"`
	} `json:"pool_stats"`
	PgStats []struct {
		PgId          string  `json:"pgid"`
		Up            []int64 `json:"up"`
		Acting        []int64 `json:"acting"`
		UpPrimary     int64   `json:"up_primary"`
		ActingPrimary int64   `json:"acting_primary"`
	} `json:"pg_stats"`
	OsdStats []struct {
		Osd         int64 `json:"osd"`
		TotalKb     int64 `json:"kb"`
		UsedKb      int64 `json:"kb_used"`
		AvailableKb int64 `json:"kb_avail"`
	} `json:"osd_stats"`
}

type OsdPerfDump struct {
	Osd struct {
		RecoveryOps         int64
		OpWip               int64 `json:"op_wip"`
		Op                  int64 `json:"op"`
		OpInBytes           int64 `json:"op_in_bytes"`
		OpOutBytes          int64 `json:"op_out_bytes"`
		OpRead              int64 `json:"op_r"`
		OpReadOutBytes      int64 `json:"op_r_out_bytes"`
		OpWrite             int64 `json:"op_w"`
		OpWriteInBytes      int64 `json:"op_w_in_bytes"`
		OpReadWrite         int64 `json:"op_rw"`
		OpReadWriteInBytes  int64 `json:"op_rw_in_btyes"`
		OpReadWriteOutBytes int64 `json:"op_rw_out_bytes"`

		OpLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_latency"`
		OpProcessLatency struct {
			OSDLatencyCalc OSDLatency
		}
		OpReadLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_r_latency"`
		OpReadProcessLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_r_process_latency"`
		OpWriteRlat struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_w_rlat"`
		OpWriteLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_w_latency"`
		OpWriteProcessLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_w_process_latency"`
		OpReadWriteRlat struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_rw_rlat"`
		OpReadWriteLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_rw_latency"`
		OpReadWriteProcessLatency struct {
			OSDLatencyCalc OSDLatency
		} `json:"op_rw_process_latency"`
	} `json:"osd"`
}

type OSDLatency struct {
	AvgCount int64   `json:"avgcount"`
	Sum      float64 `json:"sum"`
}

type PoolOsdPgMap map[int64]map[int64]int64
