package ecs

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

func Test_metastats(t *testing.T) {
	var mockAcc testutil.Accumulator

	tags := map[string]string{
		"test_tag": "test",
	}
	tm := time.Now()

	metastats(nginxStatsKey, validMeta.Containers[1], &mockAcc, tags, tm)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_meta",
		map[string]interface{}{
			"container_id":   nginxStatsKey,
			"docker_name":    "ecs-nginx-2-nginx",
			"image":          "nginx:alpine",
			"image_id":       "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			"desired_status": "RUNNING",
			"known_status":   "RUNNING",
			"limit_cpu":      float64(0),
			"limit_mem":      float64(0),
			"created_at":     metaCreated,
			"started_at":     metaStarted,
			"type":           "NORMAL",
		},
		tags,
	)
}

func Test_memstats(t *testing.T) {
	var mockAcc testutil.Accumulator

	tags := map[string]string{
		"test_tag": "test",
	}
	tm := time.Now()

	memstats(nginxStatsKey, validStats[nginxStatsKey], &mockAcc, tags, tm)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_mem",
		map[string]interface{}{
			"active_anon":               uint64(1597440),
			"active_file":               uint64(1462272),
			"cache":                     uint64(5787648),
			"container_id":              nginxStatsKey,
			"hierarchical_memory_limit": uint64(536870912),
			"inactive_anon":             uint64(4096),
			"inactive_file":             uint64(4321280),
			"limit":                     uint64(1033658368),
			"mapped_file":               uint64(3616768),
			"max_usage":                 uint64(8667136),
			"pgmajfault":                uint64(40),
			"pgpgin":                    uint64(3477),
			"pgpgout":                   uint64(1674),
			"pgfault":                   uint64(2924),
			"rss":                       uint64(1597440),
			"total_active_anon":         uint64(1597440),
			"total_active_file":         uint64(1462272),
			"total_cache":               uint64(5787648),
			"total_inactive_anon":       uint64(4096),
			"total_inactive_file":       uint64(4321280),
			"total_mapped_file":         uint64(3616768),
			"total_pgfault":             uint64(2924),
			"total_pgpgout":             uint64(1674),
			"total_pgpgin":              uint64(3477),
			"total_rss":                 uint64(1597440),
			"usage":                     uint64(2392064),
			"usage_percent":             float64(0.23141727228778164),
		},
		map[string]string{
			"test_tag": "test",
		},
	)
}

func Test_cpustats(t *testing.T) {
	var mockAcc testutil.Accumulator

	tags := map[string]string{
		"test_tag": "test",
	}
	tm := time.Now()

	cpustats(nginxStatsKey, validStats[nginxStatsKey], &mockAcc, tags, tm)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_cpu",
		map[string]interface{}{
			"container_id":                 nginxStatsKey,
			"throttling_periods":           uint64(0),
			"throttling_throttled_periods": uint64(0),
			"throttling_throttled_time":    uint64(0),
			"usage_in_usermode":            uint64(40000000),
			"usage_in_kernelmode":          uint64(10000000),
			"usage_percent":                float64(0),
			"usage_system":                 uint64(2336100000000),
			"usage_total":                  uint64(65599511),
		},
		map[string]string{
			"test_tag": "test",
			"cpu":      "cpu-total",
		},
	)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_cpu",
		map[string]interface{}{
			"container_id": nginxStatsKey,
			"usage_total":  uint64(65599511),
		},
		map[string]string{
			"test_tag": "test",
			"cpu":      "cpu0",
		},
	)
}

func Test_netstats(t *testing.T) {
	var mockAcc testutil.Accumulator

	tags := map[string]string{
		"test_tag": "test",
	}
	tm := time.Now()

	netstats(pauseStatsKey, validStats[pauseStatsKey], &mockAcc, tags, tm)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_net",
		map[string]interface{}{
			"container_id": pauseStatsKey,
			"rx_bytes":     uint64(5338),
			"rx_dropped":   uint64(0),
			"rx_errors":    uint64(0),
			"rx_packets":   uint64(36),
			"tx_bytes":     uint64(648),
			"tx_dropped":   uint64(0),
			"tx_errors":    uint64(0),
			"tx_packets":   uint64(8),
		},
		map[string]string{
			"test_tag": "test",
			"network":  "eth0",
		},
	)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_net",
		map[string]interface{}{
			"container_id": pauseStatsKey,
			"rx_bytes":     uint64(4641),
			"rx_dropped":   uint64(0),
			"rx_errors":    uint64(0),
			"rx_packets":   uint64(26),
			"tx_bytes":     uint64(690),
			"tx_dropped":   uint64(0),
			"tx_errors":    uint64(0),
			"tx_packets":   uint64(9),
		},
		map[string]string{
			"test_tag": "test",
			"network":  "eth5",
		},
	)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_net",
		map[string]interface{}{
			"container_id": pauseStatsKey,
			"rx_bytes":     uint64(9979),
			"rx_dropped":   uint64(0),
			"rx_errors":    uint64(0),
			"rx_packets":   uint64(62),
			"tx_bytes":     uint64(1338),
			"tx_dropped":   uint64(0),
			"tx_errors":    uint64(0),
			"tx_packets":   uint64(17),
		},
		map[string]string{
			"test_tag": "test",
			"network":  "total",
		},
	)
}

func Test_blkstats(t *testing.T) {
	var mockAcc testutil.Accumulator

	tags := map[string]string{
		"test_tag": "test",
	}
	tm := time.Now()

	blkstats(nginxStatsKey, validStats[nginxStatsKey], &mockAcc, tags, tm)
	mockAcc.AssertContainsTaggedFields(
		t,
		"ecs_container_blkio",
		map[string]interface{}{
			"container_id":                     nginxStatsKey,
			"io_service_bytes_recursive_read":  uint64(5730304),
			"io_service_bytes_recursive_write": uint64(0),
			"io_service_bytes_recursive_sync":  uint64(5730304),
			"io_service_bytes_recursive_async": uint64(0),
			"io_service_bytes_recursive_total": uint64(5730304),
			"io_serviced_recursive_read":       uint64(156),
			"io_serviced_recursive_write":      uint64(0),
			"io_serviced_recursive_sync":       uint64(156),
			"io_serviced_recursive_async":      uint64(0),
			"io_serviced_recursive_total":      uint64(156),
		},
		map[string]string{
			"test_tag": "test",
			"device":   "202:26368",
		},
	)
}
