package docker

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
)

var info = types.Info{
	Containers:         108,
	ContainersRunning:  98,
	ContainersStopped:  6,
	ContainersPaused:   3,
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
	ServerVersion:     "17.09.0-ce",
}

var containerList = []types.Container{
	types.Container{
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
		Labels: map[string]string{
			"label1": "test_value_1",
			"label2": "test_value_2",
		},
		SizeRw:     0,
		SizeRootFs: 0,
	},
	types.Container{
		ID:      "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		Names:   []string{"/etcd2"},
		Image:   "quay.io:4443/coreos/etcd:v2.2.2",
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
		Labels: map[string]string{
			"label1": "test_value_1",
			"label2": "test_value_2",
		},
		SizeRw:     0,
		SizeRootFs: 0,
	},
}

var two = uint64(2)
var ServiceList = []swarm.Service{
	swarm.Service{
		ID: "qolkls9g5iasdiuihcyz9rnx2",
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name: "test1",
			},
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: &two,
				},
			},
		},
	},
	swarm.Service{
		ID: "qolkls9g5iasdiuihcyz9rn3",
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name: "test2",
			},
			Mode: swarm.ServiceMode{
				Global: &swarm.GlobalService{},
			},
		},
	},
}

var TaskList = []swarm.Task{
	swarm.Task{
		ID:        "kwh0lv7hwwbh",
		ServiceID: "qolkls9g5iasdiuihcyz9rnx2",
		NodeID:    "0cl4jturcyd1ks3fwpd010kor",
		Status: swarm.TaskStatus{
			State: "running",
		},
		DesiredState: "running",
	},
	swarm.Task{
		ID:        "u78m5ojbivc3",
		ServiceID: "qolkls9g5iasdiuihcyz9rnx2",
		NodeID:    "0cl4jturcyd1ks3fwpd010kor",
		Status: swarm.TaskStatus{
			State: "running",
		},
		DesiredState: "running",
	},
	swarm.Task{
		ID:        "1n1uilkhr98l",
		ServiceID: "qolkls9g5iasdiuihcyz9rn3",
		NodeID:    "0cl4jturcyd1ks3fwpd010kor",
		Status: swarm.TaskStatus{
			State: "running",
		},
		DesiredState: "running",
	},
}

var NodeList = []swarm.Node{
	swarm.Node{
		ID: "0cl4jturcyd1ks3fwpd010kor",
		Status: swarm.NodeStatus{
			State: "ready",
		},
	},
	swarm.Node{
		ID: "0cl4jturcyd1ks3fwpd010kor",
		Status: swarm.NodeStatus{
			State: "ready",
		},
	},
}

func containerStats() types.ContainerStats {
	var stat types.ContainerStats
	jsonStat := `
{
    "blkio_stats": {
        "io_service_bytes_recursive": [
            {
                "major": 252,
                "minor": 1,
                "op": "Read",
                "value": 753664
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Write"
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Sync"
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Async",
                "value": 753664
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Total",
                "value": 753664
            }
        ],
        "io_serviced_recursive": [
            {
                "major": 252,
                "minor": 1,
                "op": "Read",
                "value": 26
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Write"
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Sync"
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Async",
                "value": 26
            },
            {
                "major": 252,
                "minor": 1,
                "op": "Total",
                "value": 26
            }
        ]
    },
    "cpu_stats": {
        "cpu_usage": {
            "percpu_usage": [
                17871,
                4959158,
                1646137,
                1231652,
                11829401,
                244656,
                369972,
                0
            ],
            "total_usage": 20298847,
            "usage_in_usermode": 10000000
        },
        "system_cpu_usage": 24052607520000000,
        "throttling_data": {}
    },
    "memory_stats": {
        "limit": 18935443456,
        "stats": {}
    },
    "precpu_stats": {
        "cpu_usage": {
            "percpu_usage": [
                17871,
                4959158,
                1646137,
                1231652,
                11829401,
                244656,
                369972,
                0
            ],
            "total_usage": 20298847,
            "usage_in_usermode": 10000000
        },
        "system_cpu_usage": 24052599550000000,
        "throttling_data": {}
    },
    "read": "2016-02-24T11:42:27.472459608-05:00"
}`
	stat.Body = ioutil.NopCloser(strings.NewReader(jsonStat))
	return stat
}

func testStats() *types.StatsJSON {
	stats := &types.StatsJSON{}
	stats.Read = time.Now()
	stats.Networks = make(map[string]types.NetworkStats)
	stats.CPUStats.OnlineCPUs = 2
	stats.CPUStats.CPUUsage.PercpuUsage = []uint64{1, 1002, 0, 0}
	stats.CPUStats.CPUUsage.UsageInUsermode = 100
	stats.CPUStats.CPUUsage.TotalUsage = 500
	stats.CPUStats.CPUUsage.UsageInKernelmode = 200
	stats.CPUStats.SystemUsage = 100
	stats.CPUStats.ThrottlingData.Periods = 1

	stats.PreCPUStats.CPUUsage.TotalUsage = 400
	stats.PreCPUStats.SystemUsage = 50

	stats.MemoryStats.Stats = make(map[string]uint64)
	stats.MemoryStats.Stats["active_anon"] = 0
	stats.MemoryStats.Stats["active_file"] = 1
	stats.MemoryStats.Stats["cache"] = 0
	stats.MemoryStats.Stats["hierarchical_memory_limit"] = 0
	stats.MemoryStats.Stats["inactive_anon"] = 0
	stats.MemoryStats.Stats["inactive_file"] = 3
	stats.MemoryStats.Stats["mapped_file"] = 0
	stats.MemoryStats.Stats["pgfault"] = 2
	stats.MemoryStats.Stats["pgmajfault"] = 0
	stats.MemoryStats.Stats["pgpgin"] = 0
	stats.MemoryStats.Stats["pgpgout"] = 0
	stats.MemoryStats.Stats["rss"] = 0
	stats.MemoryStats.Stats["rss_huge"] = 0
	stats.MemoryStats.Stats["total_active_anon"] = 0
	stats.MemoryStats.Stats["total_active_file"] = 0
	stats.MemoryStats.Stats["total_cache"] = 0
	stats.MemoryStats.Stats["total_inactive_anon"] = 0
	stats.MemoryStats.Stats["total_inactive_file"] = 0
	stats.MemoryStats.Stats["total_mapped_file"] = 0
	stats.MemoryStats.Stats["total_pgfault"] = 0
	stats.MemoryStats.Stats["total_pgmajfault"] = 0
	stats.MemoryStats.Stats["total_pgpgin"] = 4
	stats.MemoryStats.Stats["total_pgpgout"] = 0
	stats.MemoryStats.Stats["total_rss"] = 44
	stats.MemoryStats.Stats["total_rss_huge"] = 444
	stats.MemoryStats.Stats["total_unevictable"] = 0
	stats.MemoryStats.Stats["total_writeback"] = 55
	stats.MemoryStats.Stats["unevictable"] = 0
	stats.MemoryStats.Stats["writeback"] = 0

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

	stats.Networks["eth1"] = types.NetworkStats{
		RxDropped: 5,
		RxBytes:   6,
		RxErrors:  7,
		TxPackets: 8,
		TxDropped: 5,
		RxPackets: 6,
		TxErrors:  7,
		TxBytes:   8,
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
	sr2 := types.BlkioStatEntry{
		Major: 6,
		Minor: 1,
		Op:    "write",
		Value: 201,
	}

	stats.BlkioStats.IoServiceBytesRecursive = append(
		stats.BlkioStats.IoServiceBytesRecursive, sbr)
	stats.BlkioStats.IoServicedRecursive = append(
		stats.BlkioStats.IoServicedRecursive, sr)
	stats.BlkioStats.IoServicedRecursive = append(
		stats.BlkioStats.IoServicedRecursive, sr2)

	return stats
}

func containerStatsWindows() types.ContainerStats {
	var stat types.ContainerStats
	jsonStat := `
{
	"read":"2017-01-11T08:32:46.2413794Z",
	"preread":"0001-01-01T00:00:00Z",
	"num_procs":64,
	"cpu_stats":{
		"cpu_usage":{
			"total_usage":536718750,
			"usage_in_kernelmode":390468750,
			"usage_in_usermode":390468750
		},
		"throttling_data":{
			"periods":0,
			"throttled_periods":0,
			"throttled_time":0
		}
	},
	"precpu_stats":{
		"cpu_usage":{
			"total_usage":0,
			"usage_in_kernelmode":0,
			"usage_in_usermode":0
		},
		"throttling_data":{
			"periods":0,
			"throttled_periods":0,
			"throttled_time":0
		}
	},
	"memory_stats":{
		"commitbytes":77160448,
		"commitpeakbytes":105000960,
		"privateworkingset":59961344
	},
	"name":"/gt_test_iis",
}`
	stat.Body = ioutil.NopCloser(strings.NewReader(jsonStat))
	return stat
}

var containerInspect = types.ContainerJSON{
	Config: &container.Config{
		Env: []string{
			"ENVVAR1=loremipsum",
			"ENVVAR1FOO=loremipsum",
			"ENVVAR2=dolorsitamet",
			"ENVVAR3==ubuntu:10.04",
			"ENVVAR4",
			"ENVVAR5=",
			"ENVVAR6= ",
			"ENVVAR7=ENVVAR8=ENVVAR9",
			"PATH=/bin:/sbin",
		},
	},
	ContainerJSONBase: &types.ContainerJSONBase{
		State: &types.ContainerState{
			Health: &types.Health{
				FailingStreak: 1,
				Status:        "Unhealthy",
			},
			Status:     "running",
			OOMKilled:  false,
			Pid:        1234,
			ExitCode:   0,
			StartedAt:  "2018-06-14T05:48:53.266176036Z",
			FinishedAt: "0001-01-01T00:00:00Z",
		},
	},
}
