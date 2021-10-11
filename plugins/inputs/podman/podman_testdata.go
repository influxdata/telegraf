package podman

import (
	"time"

	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/cri-o/ocicni/pkg/ocicni"
)

var (
	container_test_1 = "nginx"
	container_test_2 = "blissful_lewin"
)

var info = define.Info{
	Host: &define.HostInfo{
		Distribution: define.DistributionInfo{
			Distribution: "fedora",
		},
		CPUs: 8,
	},
	Store: &define.StoreInfo{
		ContainerStore: define.ContainerStore{
			Number:  2,
			Paused:  0,
			Running: 2,
			Stopped: 0,
		},
		ImageStore: define.ImageStore{
			Number: 10,
		},
	},
	Version: define.Version{
		Version: "3.2.0",
	},
}

var containerList = []entities.ListContainer{
	{
		AutoRemove: true,
		Command:    []string{"nginx", "-g", "daemon off;"},
		Created:    time.Unix(870337891, 0),
		ID:         "9a4f6929b45ee0171b781233ce4c68acd2b7ede4fdf8d1dbe17edc3b07446854",
		Image:      "docker.io/library/nginx:latest",
		ImageID:    "08b152afcfae220e9709f00767054b824361c742ea03a9fe936271ba520a0a4b",
		Labels:     map[string]string{"maintainer": "NGINX Docker Maintainers <docker-maint@nginx.com>"},
		Names:      []string{"nginx"},
		Pid:        5429,
		Pod:        "",
		PodName:    "",
		Ports:      []ocicni.PortMapping{{HostPort: 8080, ContainerPort: 80, Protocol: "tcp", HostIP: ""}},
		State:      "running",
		Status:     "",
	},
	{
		AutoRemove: false,
		Command:    []string{"top"},
		Created:    time.Unix(870337893, 0), Exited: false, ExitCode: 0,
		ID:         "59897a61355010568bb67c3c4150163b7246648ceae6f64fac77da590dacdc3d",
		Image:      "docker.io/library/ubuntu:latest",
		ImageID:    "1318b700e415001198d1bf66d260b07f67ca8a552b61b0da02b3832c778f221b",
		IsInfra:    false,
		Labels:     map[string]string(nil),
		Mounts:     []string{},
		Names:      []string{"blissful_lewin"},
		Namespaces: entities.ListContainerNamespaces{MNT: "", Cgroup: "", IPC: "", NET: "", PIDNS: "", UTS: "", User: ""},
		Networks:   []string(nil),
		Pid:        33082,
		Pod:        "0af86f5aca6d5a8fc5628733b99d3e136beb9d859c56c702bc96aab7e5f6e4b7",
		PodName:    "elastic_pare",
		Ports:      []ocicni.PortMapping(nil),
		State:      "running",
		Status:     "",
	},
}

var containerStats_blissful_lewin = define.ContainerStats{
	ContainerID:   "59897a61355010568bb67c3c4150163b7246648ceae6f64fac77da590dacdc3d",
	Name:          "blissful_lewin",
	PerCPU:        []uint64(nil),
	CPU:           3.687353584388549e-08,
	CPUNano:       0x23cda258,
	CPUSystemNano: 0x76443,
	SystemNano:    0x169b7551393be314,
	MemUsage:      0x32d000,
	MemLimit:      0x1dacff000,
	MemPerc:       0.04180311811154799,
	NetInput:      0x0,
	NetOutput:     0x0,
	BlockInput:    0x0,
	BlockOutput:   0x0,
	PIDs:          0x1,
}

var containerStats_nginx = define.ContainerStats{
	ContainerID:   "9a4f6929b45ee0171b781233ce4c68acd2b7ede4fdf8d1dbe17edc3b07446854",
	Name:          "nginx",
	PerCPU:        []uint64(nil),
	CPU:           2.1863381926526583e-09,
	CPUNano:       0x21f7500,
	CPUSystemNano: 0x8b20,
	SystemNano:    0x169b75c89d2e4917,
	MemUsage:      0x3d4000,
	MemLimit:      0x1dacff000,
	MemPerc:       0.05038998247148467,
	NetInput:      0x0,
	NetOutput:     0x0,
	BlockInput:    0x0,
	BlockOutput:   0x0,
	PIDs:          0x9,
}
