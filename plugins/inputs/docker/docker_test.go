package system

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/registry"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestDockerGatherContainerStats(t *testing.T) {
	var acc testutil.Accumulator
	stats := testStats()

	tags := map[string]string{
		"container_name":  "redis",
		"container_image": "redis/image",
	}
	gatherContainerStats(stats, &acc, tags, "123456789")

	// test docker_container_net measurement
	netfields := map[string]interface{}{
		"rx_dropped":   uint64(1),
		"rx_bytes":     uint64(2),
		"rx_errors":    uint64(3),
		"tx_packets":   uint64(4),
		"tx_dropped":   uint64(1),
		"rx_packets":   uint64(2),
		"tx_errors":    uint64(3),
		"tx_bytes":     uint64(4),
		"container_id": "123456789",
	}
	nettags := copyTags(tags)
	nettags["network"] = "eth0"
	acc.AssertContainsTaggedFields(t, "docker_container_net", netfields, nettags)

	// test docker_blkio measurement
	blkiotags := copyTags(tags)
	blkiotags["device"] = "6:0"
	blkiofields := map[string]interface{}{
		"io_service_bytes_recursive_read": uint64(100),
		"io_serviced_recursive_write":     uint64(101),
		"container_id":                    "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_blkio", blkiofields, blkiotags)

	// test docker_container_mem measurement
	memfields := map[string]interface{}{
		"max_usage":                 uint64(1001),
		"usage":                     uint64(1111),
		"fail_count":                uint64(1),
		"limit":                     uint64(2000),
		"total_pgmafault":           uint64(0),
		"cache":                     uint64(0),
		"mapped_file":               uint64(0),
		"total_inactive_file":       uint64(0),
		"pgpgout":                   uint64(0),
		"rss":                       uint64(0),
		"total_mapped_file":         uint64(0),
		"writeback":                 uint64(0),
		"unevictable":               uint64(0),
		"pgpgin":                    uint64(0),
		"total_unevictable":         uint64(0),
		"pgmajfault":                uint64(0),
		"total_rss":                 uint64(44),
		"total_rss_huge":            uint64(444),
		"total_writeback":           uint64(55),
		"total_inactive_anon":       uint64(0),
		"rss_huge":                  uint64(0),
		"hierarchical_memory_limit": uint64(0),
		"total_pgfault":             uint64(0),
		"total_active_file":         uint64(0),
		"active_anon":               uint64(0),
		"total_active_anon":         uint64(0),
		"total_pgpgout":             uint64(0),
		"total_cache":               uint64(0),
		"inactive_anon":             uint64(0),
		"active_file":               uint64(1),
		"pgfault":                   uint64(2),
		"inactive_file":             uint64(3),
		"total_pgpgin":              uint64(4),
		"usage_percent":             float64(55.55),
		"container_id":              "123456789",
	}

	acc.AssertContainsTaggedFields(t, "docker_container_mem", memfields, tags)

	// test docker_container_cpu measurement
	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	cpufields := map[string]interface{}{
		"usage_total":                  uint64(500),
		"usage_in_usermode":            uint64(100),
		"usage_in_kernelmode":          uint64(200),
		"usage_system":                 uint64(100),
		"throttling_periods":           uint64(1),
		"throttling_throttled_periods": uint64(0),
		"throttling_throttled_time":    uint64(0),
		"usage_percent":                float64(400.0),
		"container_id":                 "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpufields, cputags)

	cputags["cpu"] = "cpu0"
	cpu0fields := map[string]interface{}{
		"usage_total":  uint64(1),
		"container_id": "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpu0fields, cputags)

	cputags["cpu"] = "cpu1"
	cpu1fields := map[string]interface{}{
		"usage_total":  uint64(1002),
		"container_id": "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpu1fields, cputags)
}

func testStats() *types.StatsJSON {
	stats := &types.StatsJSON{}
	stats.Read = time.Now()
	stats.Networks = make(map[string]types.NetworkStats)

	stats.CPUStats.CPUUsage.PercpuUsage = []uint64{1, 1002}
	stats.CPUStats.CPUUsage.UsageInUsermode = 100
	stats.CPUStats.CPUUsage.TotalUsage = 500
	stats.CPUStats.CPUUsage.UsageInKernelmode = 200
	stats.CPUStats.SystemUsage = 100
	stats.CPUStats.ThrottlingData.Periods = 1

	stats.PreCPUStats.CPUUsage.TotalUsage = 400
	stats.PreCPUStats.SystemUsage = 50

	stats.MemoryStats.Stats = make(map[string]uint64)
	stats.MemoryStats.Stats["total_pgmajfault"] = 0
	stats.MemoryStats.Stats["cache"] = 0
	stats.MemoryStats.Stats["mapped_file"] = 0
	stats.MemoryStats.Stats["total_inactive_file"] = 0
	stats.MemoryStats.Stats["pagpgout"] = 0
	stats.MemoryStats.Stats["rss"] = 0
	stats.MemoryStats.Stats["total_mapped_file"] = 0
	stats.MemoryStats.Stats["writeback"] = 0
	stats.MemoryStats.Stats["unevictable"] = 0
	stats.MemoryStats.Stats["pgpgin"] = 0
	stats.MemoryStats.Stats["total_unevictable"] = 0
	stats.MemoryStats.Stats["pgmajfault"] = 0
	stats.MemoryStats.Stats["total_rss"] = 44
	stats.MemoryStats.Stats["total_rss_huge"] = 444
	stats.MemoryStats.Stats["total_write_back"] = 55
	stats.MemoryStats.Stats["total_inactive_anon"] = 0
	stats.MemoryStats.Stats["rss_huge"] = 0
	stats.MemoryStats.Stats["hierarchical_memory_limit"] = 0
	stats.MemoryStats.Stats["total_pgfault"] = 0
	stats.MemoryStats.Stats["total_active_file"] = 0
	stats.MemoryStats.Stats["active_anon"] = 0
	stats.MemoryStats.Stats["total_active_anon"] = 0
	stats.MemoryStats.Stats["total_pgpgout"] = 0
	stats.MemoryStats.Stats["total_cache"] = 0
	stats.MemoryStats.Stats["inactive_anon"] = 0
	stats.MemoryStats.Stats["active_file"] = 1
	stats.MemoryStats.Stats["pgfault"] = 2
	stats.MemoryStats.Stats["inactive_file"] = 3
	stats.MemoryStats.Stats["total_pgpgin"] = 4

	stats.MemoryStats.MaxUsage = 1001
	stats.MemoryStats.Usage = 1111
	stats.MemoryStats.Failcnt = 1
	stats.MemoryStats.Limit = 2000

	stats.Networks["eth0"] = types.NetworkStats{
		RxDropped: 1,
		RxBytes:   2,
		RxErrors:  3,
		TxPackets: 4,
		TxDropped: 1,
		RxPackets: 2,
		TxErrors:  3,
		TxBytes:   4,
	}

	sbr := types.BlkioStatEntry{
		Major: 6,
		Minor: 0,
		Op:    "read",
		Value: 100,
	}
	sr := types.BlkioStatEntry{
		Major: 6,
		Minor: 0,
		Op:    "write",
		Value: 101,
	}

	stats.BlkioStats.IoServiceBytesRecursive = append(
		stats.BlkioStats.IoServiceBytesRecursive, sbr)
	stats.BlkioStats.IoServicedRecursive = append(
		stats.BlkioStats.IoServicedRecursive, sr)

	return stats
}

type FakeDockerClient struct {
}

func (d FakeDockerClient) Info(ctx context.Context) (types.Info, error) {
	env := types.Info{
		Containers:         108,
		OomKillDisable:     false,
		SystemTime:         "2016-02-24T00:55:09.15073105-05:00",
		NEventsListener:    0,
		ID:                 "5WQQ:TFWR:FDNG:OKQ3:37Y4:FJWG:QIKK:623T:R3ME:QTKB:A7F7:OLHD",
		Debug:              false,
		LoggingDriver:      "json-file",
		KernelVersion:      "4.3.0-1-amd64",
		IndexServerAddress: "https://index.docker.io/v1/",
		MemTotal:           3840757760,
		Images:             199,
		CPUCfsQuota:        true,
		Name:               "absol",
		SwapLimit:          false,
		IPv4Forwarding:     true,
		ExecutionDriver:    "native-0.2",
		ExperimentalBuild:  false,
		CPUCfsPeriod:       true,
		RegistryConfig: &registry.ServiceConfig{
			IndexConfigs: map[string]*registry.IndexInfo{
				"docker.io": {
					Name:     "docker.io",
					Mirrors:  []string{},
					Official: true,
					Secure:   true,
				},
			}, InsecureRegistryCIDRs: []*registry.NetIPNet{{IP: []byte{127, 0, 0, 0}, Mask: []byte{255, 0, 0, 0}}}, Mirrors: []string{}},
		OperatingSystem:   "Linux Mint LMDE (containerized)",
		BridgeNfIptables:  true,
		HTTPSProxy:        "",
		Labels:            []string{},
		MemoryLimit:       false,
		DriverStatus:      [][2]string{{"Pool Name", "docker-8:1-1182287-pool"}, {"Pool Blocksize", "65.54 kB"}, {"Backing Filesystem", "extfs"}, {"Data file", "/dev/loop0"}, {"Metadata file", "/dev/loop1"}, {"Data Space Used", "17.3 GB"}, {"Data Space Total", "107.4 GB"}, {"Data Space Available", "36.53 GB"}, {"Metadata Space Used", "20.97 MB"}, {"Metadata Space Total", "2.147 GB"}, {"Metadata Space Available", "2.127 GB"}, {"Udev Sync Supported", "true"}, {"Deferred Removal Enabled", "false"}, {"Data loop file", "/var/lib/docker/devicemapper/devicemapper/data"}, {"Metadata loop file", "/var/lib/docker/devicemapper/devicemapper/metadata"}, {"Library Version", "1.02.115 (2016-01-25)"}},
		NFd:               19,
		HTTPProxy:         "",
		Driver:            "devicemapper",
		NGoroutines:       39,
		NCPU:              4,
		DockerRootDir:     "/var/lib/docker",
		NoProxy:           "",
		BridgeNfIP6tables: true,
	}
	return env, nil
}

func (d FakeDockerClient) ContainerList(octx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	container1 := types.Container{
		ID:      "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
		Names:   []string{"/etcd"},
		Image:   "quay.io/coreos/etcd:v2.2.2",
		Command: "/etcd -name etcd0 -advertise-client-urls http://localhost:2379 -listen-client-urls http://0.0.0.0:2379",
		Created: 1455941930,
		Status:  "Up 4 hours",
		Ports: []types.Port{
			types.Port{
				PrivatePort: 7001,
				PublicPort:  0,
				Type:        "tcp",
			},
			types.Port{
				PrivatePort: 4001,
				PublicPort:  0,
				Type:        "tcp",
			},
			types.Port{
				PrivatePort: 2380,
				PublicPort:  0,
				Type:        "tcp",
			},
			types.Port{
				PrivatePort: 2379,
				PublicPort:  2379,
				Type:        "tcp",
				IP:          "0.0.0.0",
			},
		},
		SizeRw:     0,
		SizeRootFs: 0,
	}
	container2 := types.Container{
		ID:      "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		Names:   []string{"/etcd2"},
		Image:   "quay.io/coreos/etcd:v2.2.2",
		Command: "/etcd -name etcd2 -advertise-client-urls http://localhost:2379 -listen-client-urls http://0.0.0.0:2379",
		Created: 1455941933,
		Status:  "Up 4 hours",
		Ports: []types.Port{
			types.Port{
				PrivatePort: 7002,
				PublicPort:  0,
				Type:        "tcp",
			},
			types.Port{
				PrivatePort: 4002,
				PublicPort:  0,
				Type:        "tcp",
			},
			types.Port{
				PrivatePort: 2381,
				PublicPort:  0,
				Type:        "tcp",
			},
			types.Port{
				PrivatePort: 2382,
				PublicPort:  2382,
				Type:        "tcp",
				IP:          "0.0.0.0",
			},
		},
		SizeRw:     0,
		SizeRootFs: 0,
	}

	containers := []types.Container{container1, container2}
	return containers, nil

	//#{e6a96c84ca91a5258b7cb752579fb68826b68b49ff957487695cd4d13c343b44 titilambert/snmpsim /bin/sh -c 'snmpsimd --agent-udpv4-endpoint=0.0.0.0:31161 --process-user=root --process-group=user' 1455724831 Up 4 hours [{31161 31161 udp 0.0.0.0}] 0 0 [/snmp] map[]}]2016/02/24 01:05:01 Gathered metrics, (3s interval), from 1 inputs in 1.233836656s
}

func (d FakeDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (io.ReadCloser, error) {
	var stat io.ReadCloser
	jsonStat := `{"read":"2016-02-24T11:42:27.472459608-05:00","memory_stats":{"stats":{},"limit":18935443456},"blkio_stats":{"io_service_bytes_recursive":[{"major":252,"minor":1,"op":"Read","value":753664},{"major":252,"minor":1,"op":"Write"},{"major":252,"minor":1,"op":"Sync"},{"major":252,"minor":1,"op":"Async","value":753664},{"major":252,"minor":1,"op":"Total","value":753664}],"io_serviced_recursive":[{"major":252,"minor":1,"op":"Read","value":26},{"major":252,"minor":1,"op":"Write"},{"major":252,"minor":1,"op":"Sync"},{"major":252,"minor":1,"op":"Async","value":26},{"major":252,"minor":1,"op":"Total","value":26}]},"cpu_stats":{"cpu_usage":{"percpu_usage":[17871,4959158,1646137,1231652,11829401,244656,369972,0],"usage_in_usermode":10000000,"total_usage":20298847},"system_cpu_usage":24052607520000000,"throttling_data":{}},"precpu_stats":{"cpu_usage":{"percpu_usage":[17871,4959158,1646137,1231652,11829401,244656,369972,0],"usage_in_usermode":10000000,"total_usage":20298847},"system_cpu_usage":24052599550000000,"throttling_data":{}}}`
	stat = ioutil.NopCloser(strings.NewReader(jsonStat))
	return stat, nil
}

func TestDockerGatherInfo(t *testing.T) {
	var acc testutil.Accumulator
	client := FakeDockerClient{}
	d := Docker{client: client}

	err := d.Gather(&acc)

	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"docker",
		map[string]interface{}{
			"n_listener_events":       int(0),
			"n_cpus":                  int(4),
			"n_used_file_descriptors": int(19),
			"n_containers":            int(108),
			"n_images":                int(199),
			"n_goroutines":            int(39),
		},
		map[string]string{},
	)

	acc.AssertContainsTaggedFields(t,
		"docker_data",
		map[string]interface{}{
			"used":      int64(17300000000),
			"total":     int64(107400000000),
			"available": int64(36530000000),
		},
		map[string]string{
			"unit": "bytes",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"docker_container_cpu",
		map[string]interface{}{
			"usage_total":  uint64(1231652),
			"container_id": "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		},
		map[string]string{
			"container_name":  "etcd2",
			"container_image": "quay.io/coreos/etcd:v2.2.2",
			"cpu":             "cpu3",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"docker_container_mem",
		map[string]interface{}{
			"total_pgpgout":             uint64(0),
			"usage_percent":             float64(0),
			"rss":                       uint64(0),
			"total_writeback":           uint64(0),
			"active_anon":               uint64(0),
			"total_pgmafault":           uint64(0),
			"total_rss":                 uint64(0),
			"total_unevictable":         uint64(0),
			"active_file":               uint64(0),
			"total_mapped_file":         uint64(0),
			"pgpgin":                    uint64(0),
			"total_active_file":         uint64(0),
			"total_active_anon":         uint64(0),
			"total_cache":               uint64(0),
			"inactive_anon":             uint64(0),
			"pgmajfault":                uint64(0),
			"total_inactive_anon":       uint64(0),
			"total_rss_huge":            uint64(0),
			"rss_huge":                  uint64(0),
			"hierarchical_memory_limit": uint64(0),
			"pgpgout":                   uint64(0),
			"unevictable":               uint64(0),
			"total_inactive_file":       uint64(0),
			"writeback":                 uint64(0),
			"total_pgfault":             uint64(0),
			"total_pgpgin":              uint64(0),
			"cache":                     uint64(0),
			"mapped_file":               uint64(0),
			"inactive_file":             uint64(0),
			"max_usage":                 uint64(0),
			"fail_count":                uint64(0),
			"pgfault":                   uint64(0),
			"usage":                     uint64(0),
			"limit":                     uint64(18935443456),
			"container_id":              "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		},
		map[string]string{
			"container_name":  "etcd2",
			"container_image": "quay.io/coreos/etcd:v2.2.2",
		},
	)

	//fmt.Print(info)
}
