package ecs

import (
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/require"
)

// codified golden objects for tests

// stats
const pauseStatsKey = "e6af031b91deb3136a2b7c42f262ed2ab554e2fe2736998c7d8edf4afe708dba"
const nginxStatsKey = "fffe894e232d46c76475cfeabf4907f712e8b92618a37fca3ef0805bbbfb0299"

var pauseStatsRead, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:40:00.936081344Z")
var pauseStatsPreRead, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:39:59.933000984Z")

var nginxStatsRead, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:40:00.93733207Z")
var nginxStatsPreRead, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:39:59.934291009Z")

var validStats = map[string]types.StatsJSON{
	pauseStatsKey: {
		Stats: types.Stats{
			Read:    pauseStatsRead,
			PreRead: pauseStatsPreRead,
			BlkioStats: types.BlkioStats{
				IoServiceBytesRecursive: []types.BlkioStatEntry{
					{
						Major: 202,
						Minor: 26368,
						Op:    "Read",
						Value: 790528,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Write",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Sync",
						Value: 790528,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Async",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Total",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Read",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Sync",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Total",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Read",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Sync",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Total",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Read",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Sync",
						Value: 790528,
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Total",
						Value: 790528,
					},
				},
				IoServicedRecursive: []types.BlkioStatEntry{
					{
						Major: 202,
						Minor: 26368,
						Op:    "Read",
						Value: 10,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Write",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Sync",
						Value: 10,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Async",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Total",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Read",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Sync",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Total",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Read",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Sync",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Total",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Read",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Sync",
						Value: 10,
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 4,
						Op:    "Total",
						Value: 10,
					},
				},
			},
			CPUStats: types.CPUStats{
				CPUUsage: types.CPUUsage{
					PercpuUsage: []uint64{
						26426156,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
					},
					UsageInUsermode: 20000000,
					TotalUsage:      26426156,
				},
				SystemUsage:    2336100000000,
				OnlineCPUs:     1,
				ThrottlingData: types.ThrottlingData{},
			},
			PreCPUStats: types.CPUStats{
				CPUUsage: types.CPUUsage{
					PercpuUsage: []uint64{
						26426156,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
					},
					UsageInUsermode: 20000000,
					TotalUsage:      26426156,
				},
				SystemUsage:    2335090000000,
				OnlineCPUs:     1,
				ThrottlingData: types.ThrottlingData{},
			},
			MemoryStats: types.MemoryStats{
				Stats: map[string]uint64{
					"cache":                     790528,
					"mapped_file":               618496,
					"total_inactive_file":       782336,
					"pgpgout":                   1040,
					"rss":                       40960,
					"total_mapped_file":         618496,
					"pgpgin":                    1243,
					"pgmajfault":                6,
					"total_rss":                 40960,
					"hierarchical_memory_limit": 536870912,
					"total_pgfault":             1298,
					"total_active_file":         8192,
					"active_anon":               40960,
					"total_active_anon":         40960,
					"total_pgpgout":             1040,
					"total_cache":               790528,
					"active_file":               8192,
					"pgfault":                   1298,
					"inactive_file":             782336,
					"total_pgpgin":              1243,
					"hierarchical_memsw_limit":  9223372036854772000,
				},
				MaxUsage: 4825088,
				Usage:    1343488,
				Limit:    1033658368,
			},
		},
		Networks: map[string]types.NetworkStats{
			"eth0": {
				RxBytes:   uint64(5338),
				RxDropped: uint64(0),
				RxErrors:  uint64(0),
				RxPackets: uint64(36),
				TxBytes:   uint64(648),
				TxDropped: uint64(0),
				TxErrors:  uint64(0),
				TxPackets: uint64(8),
			},
			"eth5": {
				RxBytes:   uint64(4641),
				RxDropped: uint64(0),
				RxErrors:  uint64(0),
				RxPackets: uint64(26),
				TxBytes:   uint64(690),
				TxDropped: uint64(0),
				TxErrors:  uint64(0),
				TxPackets: uint64(9),
			},
		},
	},
	nginxStatsKey: {
		Stats: types.Stats{
			Read:    nginxStatsRead,
			PreRead: nginxStatsPreRead,
			BlkioStats: types.BlkioStats{
				IoServiceBytesRecursive: []types.BlkioStatEntry{
					{
						Major: 202,
						Minor: 26368,
						Op:    "Read",
						Value: 5730304,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Write",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Sync",
						Value: 5730304,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Async",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Total",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Read",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Sync",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Total",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Read",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Sync",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Total",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Read",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Sync",
						Value: 5730304,
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Total",
						Value: 5730304,
					},
				},
				IoServicedRecursive: []types.BlkioStatEntry{
					{
						Major: 202,
						Minor: 26368,
						Op:    "Read",
						Value: 156,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Write",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Sync",
						Value: 156,
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Async",
					},
					{
						Major: 202,
						Minor: 26368,
						Op:    "Total",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Read",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Sync",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 1,
						Op:    "Total",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Read",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Sync",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 2,
						Op:    "Total",
						Value: 156,
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Read",
						Value: 147,
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Write",
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Sync",
						Value: 147,
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Async",
					},
					{
						Major: 253,
						Minor: 5,
						Op:    "Total",
						Value: 147,
					},
				},
			},
			CPUStats: types.CPUStats{
				CPUUsage: types.CPUUsage{
					PercpuUsage: []uint64{
						65599511,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
					},
					UsageInUsermode:   40000000,
					TotalUsage:        65599511,
					UsageInKernelmode: 10000000,
				},
				SystemUsage:    2336100000000,
				OnlineCPUs:     1,
				ThrottlingData: types.ThrottlingData{},
			},
			PreCPUStats: types.CPUStats{
				CPUUsage: types.CPUUsage{
					PercpuUsage: []uint64{
						65599511,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
						0,
					},
					UsageInUsermode:   40000000,
					TotalUsage:        65599511,
					UsageInKernelmode: 10000000,
				},
				SystemUsage:    2335090000000,
				OnlineCPUs:     1,
				ThrottlingData: types.ThrottlingData{},
			},
			MemoryStats: types.MemoryStats{
				Stats: map[string]uint64{
					"cache":                     5787648,
					"mapped_file":               3616768,
					"total_inactive_file":       4321280,
					"pgpgout":                   1674,
					"rss":                       1597440,
					"total_mapped_file":         3616768,
					"pgpgin":                    3477,
					"pgmajfault":                40,
					"total_rss":                 1597440,
					"total_inactive_anon":       4096,
					"hierarchical_memory_limit": 536870912,
					"total_pgfault":             2924,
					"total_active_file":         1462272,
					"active_anon":               1597440,
					"total_active_anon":         1597440,
					"total_pgpgout":             1674,
					"total_cache":               5787648,
					"inactive_anon":             4096,
					"active_file":               1462272,
					"pgfault":                   2924,
					"inactive_file":             4321280,
					"total_pgpgin":              3477,
					"hierarchical_memsw_limit":  9223372036854772000,
				},
				MaxUsage: 8667136,
				Usage:    8179712,
				Limit:    1033658368,
			},
		},
	},
}

// meta
var metaPauseCreated, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:31:26.641964373Z")
var metaPauseStarted, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:31:27.035698679Z")
var metaCreated, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:31:27.614884084Z")
var metaStarted, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:31:27.975996351Z")
var metaPullStart, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:31:27.197327103Z")
var metaPullStop, _ = time.Parse(time.RFC3339Nano, "2018-11-19T15:31:27.609089471Z")

var validMeta = Task{
	Cluster:       "test",
	TaskARN:       "arn:aws:ecs:aws-region-1:012345678901:task/a1234abc-a0a0-0a01-ab01-0abc012a0a0a",
	Family:        "nginx",
	Revision:      "2",
	DesiredStatus: "RUNNING",
	KnownStatus:   "RUNNING",
	Containers: []Container{
		{
			ID:         pauseStatsKey,
			Name:       "~internal~ecs~pause",
			DockerName: "ecs-nginx-2-internalecspause",
			Image:      "amazon/amazon-ecs-pause:0.1.0",
			ImageID:    "",
			Labels: map[string]string{
				"com.amazonaws.ecs.cluster":                 "test",
				"com.amazonaws.ecs.container-name":          "~internal~ecs~pause",
				"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:aws-region-1:012345678901:task/a1234abc-a0a0-0a01-ab01-0abc012a0a0a",
				"com.amazonaws.ecs.task-definition-family":  "nginx",
				"com.amazonaws.ecs.task-definition-version": "2",
			},
			DesiredStatus: "RESOURCES_PROVISIONED",
			KnownStatus:   "RESOURCES_PROVISIONED",
			Limits: map[string]float64{
				"CPU":    0,
				"Memory": 0,
			},
			CreatedAt: metaPauseCreated,
			StartedAt: metaPauseStarted,
			Type:      "CNI_PAUSE",
			Networks: []Network{
				{
					NetworkMode: "awsvpc",
					IPv4Addresses: []string{
						"172.31.25.181",
					},
				},
			},
		},
		{
			ID:         nginxStatsKey,
			Name:       "nginx",
			DockerName: "ecs-nginx-2-nginx",
			Image:      "nginx:alpine",
			ImageID:    "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Labels: map[string]string{
				"com.amazonaws.ecs.cluster":                 "test",
				"com.amazonaws.ecs.container-name":          "nginx",
				"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:aws-region-1:012345678901:task/a1234abc-a0a0-0a01-ab01-0abc012a0a0a",
				"com.amazonaws.ecs.task-definition-family":  "nginx",
				"com.amazonaws.ecs.task-definition-version": "2",
			},
			DesiredStatus: "RUNNING",
			KnownStatus:   "RUNNING",
			Limits: map[string]float64{
				"CPU":    0,
				"Memory": 0,
			},
			CreatedAt: metaCreated,
			StartedAt: metaStarted,
			Type:      "NORMAL",
			Networks: []Network{
				{
					NetworkMode: "awsvpc",
					IPv4Addresses: []string{
						"172.31.25.181",
					},
				},
			},
		},
	},
	Limits: map[string]float64{
		"CPU":    0.5,
		"Memory": 512,
	},
	PullStartedAt: metaPullStart,
	PullStoppedAt: metaPullStop,
}

func TestResolveEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		given  Ecs
		exp    Ecs
		preF   func()
		afterF func()
	}{
		{
			name: "Endpoint is explicitly set => use v2 metadata",
			given: Ecs{
				EndpointURL: "192.162.0.1/custom_endpoint",
			},
			exp: Ecs{
				EndpointURL:     "192.162.0.1/custom_endpoint",
				metadataVersion: 2,
			},
		},
		{
			name: "Endpoint is not set, ECS_CONTAINER_METADATA_URI is not set => use v2 metadata",
			given: Ecs{
				EndpointURL: "",
			},
			exp: Ecs{
				EndpointURL:     v2Endpoint,
				metadataVersion: 2,
			},
		},
		{
			name: "Endpoint is not set, ECS_CONTAINER_METADATA_URI is set => use v3 metadata",
			preF: func() {
				require.NoError(t, os.Setenv("ECS_CONTAINER_METADATA_URI", "v3-endpoint.local"))
			},
			afterF: func() {
				require.NoError(t, os.Unsetenv("ECS_CONTAINER_METADATA_URI"))
			},
			given: Ecs{
				EndpointURL: "",
			},
			exp: Ecs{
				EndpointURL:     "v3-endpoint.local",
				metadataVersion: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preF != nil {
				tt.preF()
			}
			if tt.afterF != nil {
				defer tt.afterF()
			}

			act := tt.given
			resolveEndpoint(&act)
			require.Equal(t, tt.exp, act)
		})
	}
}
