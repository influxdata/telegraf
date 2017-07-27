package runc

import (
	"context"
	"encoding/json"
	"github.com/containerd/go-runc"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Toggle debugging on and off
const DEBUG = false

func TestRuncPlugin(t *testing.T) {
	rc := &Runc{collector: MockCollector{}}
	acc := &testutil.Accumulator{}
	acc.SetDebug(DEBUG)
	assert.NoError(t, rc.Gather(acc))
	// Validate we collected all expected measurements
	assert.True(t, acc.HasMeasurement("cpu"))
	assert.True(t, acc.HasMeasurement("memory"))
	assert.True(t, acc.HasMeasurement("pids"))
	assert.True(t, acc.HasMeasurement("blkio"))
	// Validate some known metrics
	// Make sure we have consistent number of fields
	assert.Equal(t, 279, acc.NFields())
	// Ensure 20 groups of metrics were collected
	assert.Equal(t, uint64(20), acc.NMetrics())
	blkio, ok := acc.Get("blkio")
	assert.True(t, ok)
	// Make sure container-4 has blkio metrics and the are correctly parsed
	assert.Equal(t, blkio.Tags["id"], "container-4")
	assert.True(t, acc.HasUIntField("blkio", "wait_time_recursive_async_8_0"))
	assert.True(t, acc.HasUIntField("blkio", "wait_time_recursive_read_8_0"))
	assert.True(t, acc.HasUIntField("blkio", "wait_time_recursive_sync_8_0"))
	assert.True(t, acc.HasUIntField("blkio", "wait_time_recursive_total_8_0"))
	assert.True(t, acc.HasUIntField("blkio", "wait_time_recursive_write_8_0"))
	// Validate some known metrics from container-1
	assert.Equal(t, "container-1", acc.Metrics[0].Tags["id"])
	cpu, ok := acc.Get("cpu")
	assert.True(t, ok)
	assert.Equal(t, "container-1", cpu.Tags["id"])
	assert.Equal(t, uint64(9494869), cpu.Fields["usage_total"].(uint64))
	assert.Equal(t, uint64(6989690), cpu.Fields["usage_core_0"].(uint64))
	assert.Equal(t, uint64(1663589), cpu.Fields["usage_core_1"].(uint64))
	assert.Equal(t, uint64(428176), cpu.Fields["usage_core_2"].(uint64))
	assert.Equal(t, uint64(413414), cpu.Fields["usage_core_3"].(uint64))
}

func TestRuncPluginExcludeFilter(t *testing.T) {
	rc := &Runc{
		collector: MockCollector{},
		// Ignore container-4
		ContainerExclude: []string{"container-4"},
	}
	acc := &testutil.Accumulator{}
	acc.SetDebug(DEBUG)
	assert.NoError(t, rc.Gather(acc))
	// Should only collect 15 metrics with 3 containers
	assert.Equal(t, uint64(15), acc.NMetrics())
	// Validate collected measurements
	assert.True(t, acc.HasMeasurement("cpu"))
	assert.True(t, acc.HasMeasurement("memory"))
	assert.True(t, acc.HasMeasurement("pids"))
	// Only container-4 has blkio data so it should
	// now be missing
	assert.False(t, acc.HasMeasurement("blkio"))
}

func TestRuncPluginIncludeFilter(t *testing.T) {
	// Clear collected metrics
	rc := &Runc{
		collector: MockCollector{},
		// Include only container-1
		ContainerInclude: []string{"container-1"},
	}
	acc := &testutil.Accumulator{}
	acc.SetDebug(DEBUG)
	assert.NoError(t, rc.Gather(acc))
	// Should only collect 5 metrics with one container
	assert.Equal(t, uint64(5), acc.NMetrics())
	// Validate collected measurements
	assert.True(t, acc.HasMeasurement("cpu"))
	assert.True(t, acc.HasMeasurement("memory"))
	assert.True(t, acc.HasMeasurement("pids"))
}

func TestRuncPluginGlobInclude(t *testing.T) {
	rc := &Runc{
		collector: MockCollector{},
		// Match all containers
		ContainerInclude: []string{"cont*"},
	}
	acc := &testutil.Accumulator{}
	acc.SetDebug(DEBUG)
	assert.NoError(t, rc.Gather(acc))
	assert.Equal(t, 20, int(acc.NMetrics()))
}

// MockCollector implements a fake runc client for testing
type MockCollector struct{}

func (c MockCollector) List(_ context.Context) ([]*runc.Container, error) {
	containers := []*runc.Container{}
	err := json.Unmarshal(fakeData["containers"], &containers)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (c MockCollector) Stats(_ context.Context, id string) (*runc.Stats, error) {
	event := &runc.Event{}
	err := json.Unmarshal(fakeData[id], event)
	if err != nil {
		return nil, err
	}
	return event.Stats, nil
}

var fakeData = map[string][]byte{
	"containers": []byte(`
[{"ociVersion":"1.0.0-rc5-dev","id":"container-1","pid":17860,"status":"running","bundle":"/tmp/container","rootfs":"/tmp/container/rootfs","created":"2017-07-23T13:40:40.756465727Z","owner":"root"},{"ociVersion":"1.0.0-rc5-dev","id":"container-2","pid":17877,"status":"running","bundle":"/tmp/container","rootfs":"/tmp/container/rootfs","created":"2017-07-23T13:40:48.670136498Z","owner":"root"},{"ociVersion":"1.0.0-rc5-dev","id":"container-3","pid":18125,"status":"running","bundle":"/tmp/container","rootfs":"/tmp/container/rootfs","created":"2017-07-23T13:41:43.659479982Z","owner":"root"},{"ociVersion":"1.0.0-rc5-dev","id":"container-4","pid":18262,"status":"running","bundle":"/tmp/container","rootfs":"/tmp/container/rootfs","created":"2017-07-23T13:41:51.859913546Z","owner":"root"}]`),
	"container-1": []byte(`{"type":"stats","id":"container-1","data":{"cpu":{"usage":{"total":9494869,"percpu":[6989690,1663589,428176,413414],"kernel":0,"user":0},"throttling":{}},"memory":{"usage":{"limit":9223372036854771712,"usage":249856,"max":442368,"failcnt":0},"swap":{"limit":9223372036854771712,"usage":249856,"max":442368,"failcnt":0},"kernel":{"limit":9223372036854771712,"usage":188416,"max":188416,"failcnt":0},"kernelTCP":{"limit":9223372036854771712,"failcnt":0},"raw":{"active_anon":61440,"active_file":0,"cache":0,"dirty":0,"hierarchical_memory_limit":9223372036854771712,"hierarchical_memsw_limit":9223372036854771712,"inactive_anon":0,"inactive_file":0,"mapped_file":0,"pgfault":85,"pgmajfault":0,"pgpgin":69,"pgpgout":54,"rss":61440,"rss_huge":0,"swap":0,"total_active_anon":61440,"total_active_file":0,"total_cache":0,"total_dirty":0,"total_inactive_anon":0,"total_inactive_file":0,"total_mapped_file":0,"total_pgfault":85,"total_pgmajfault":0,"total_pgpgin":69,"total_pgpgout":54,"total_rss":61440,"total_rss_huge":0,"total_swap":0,"total_unevictable":0,"total_writeback":0,"unevictable":0,"writeback":0}},"pids":{"current":1},"blkio":{},"hugetlb":{}}}`),
	"container-2": []byte(`{"type":"stats","id":"container-2","data":{"cpu":{"usage":{"total":11063404,"percpu":[2056213,6827556,460880,1718755],"kernel":0,"user":0},"throttling":{}},"memory":{"usage":{"limit":9223372036854771712,"usage":319488,"max":655360,"failcnt":0},"swap":{"limit":9223372036854771712,"usage":319488,"max":655360,"failcnt":0},"kernel":{"limit":9223372036854771712,"usage":258048,"max":258048,"failcnt":0},"kernelTCP":{"limit":9223372036854771712,"failcnt":0},"raw":{"active_anon":61440,"active_file":0,"cache":0,"dirty":0,"hierarchical_memory_limit":9223372036854771712,"hierarchical_memsw_limit":9223372036854771712,"inactive_anon":0,"inactive_file":0,"mapped_file":0,"pgfault":95,"pgmajfault":0,"pgpgin":75,"pgpgout":60,"rss":61440,"rss_huge":0,"swap":0,"total_active_anon":61440,"total_active_file":0,"total_cache":0,"total_dirty":0,"total_inactive_anon":0,"total_inactive_file":0,"total_mapped_file":0,"total_pgfault":95,"total_pgmajfault":0,"total_pgpgin":75,"total_pgpgout":60,"total_rss":61440,"total_rss_huge":0,"total_swap":0,"total_unevictable":0,"total_writeback":0,"unevictable":0,"writeback":0}},"pids":{"current":1},"blkio":{},"hugetlb":{}}}`),
	"container-3": []byte(`{"type":"stats","id":"container-3","data":{"cpu":{"usage":{"total":9962852,"percpu":[159086,742435,8651852,409479],"kernel":0,"user":0},"throttling":{}},"memory":{"usage":{"limit":9223372036854771712,"usage":241664,"max":524288,"failcnt":0},"swap":{"limit":9223372036854771712,"usage":241664,"max":524288,"failcnt":0},"kernel":{"limit":9223372036854771712,"usage":184320,"max":188416,"failcnt":0},"kernelTCP":{"limit":9223372036854771712,"failcnt":0},"raw":{"active_anon":57344,"active_file":0,"cache":0,"dirty":0,"hierarchical_memory_limit":9223372036854771712,"hierarchical_memsw_limit":9223372036854771712,"inactive_anon":0,"inactive_file":0,"mapped_file":0,"pgfault":93,"pgmajfault":0,"pgpgin":74,"pgpgout":60,"rss":57344,"rss_huge":0,"swap":0,"total_active_anon":57344,"total_active_file":0,"total_cache":0,"total_dirty":0,"total_inactive_anon":0,"total_inactive_file":0,"total_mapped_file":0,"total_pgfault":93,"total_pgmajfault":0,"total_pgpgin":74,"total_pgpgout":60,"total_rss":57344,"total_rss_huge":0,"total_swap":0,"total_unevictable":0,"total_writeback":0,"unevictable":0,"writeback":0}},"pids":{"current":1},"blkio":{},"hugetlb":{}}}`),
	"container-4": []byte(`{"type":"stats","id":"container-4","data":{"cpu":{"usage":{"total":43302044,"percpu":[18988408,8814708,8888476,6610452],"kernel":30000000,"user":0},"throttling":{}},"memory":{"cache":8192,"usage":{"limit":9223372036854772000,"usage":798720,"max":1183744,"failcnt":0},"swap":{"limit":9223372036854772000,"usage":798720,"max":1183744,"failcnt":0},"kernel":{"limit":9223372036854772000,"usage":724992,"max":815104,"failcnt":0},"kernelTCP":{"limit":9223372036854772000,"failcnt":0},"raw":{"active_anon":65536,"active_file":4096,"cache":8192,"dirty":4096,"hierarchical_memory_limit":9223372036854772000,"hierarchical_memsw_limit":9223372036854772000,"inactive_anon":0,"inactive_file":4096,"mapped_file":0,"pgfault":972,"pgmajfault":0,"pgpgin":488,"pgpgout":470,"rss":65536,"rss_huge":0,"swap":0,"total_active_anon":65536,"total_active_file":4096,"total_cache":8192,"total_dirty":4096,"total_inactive_anon":0,"total_inactive_file":4096,"total_mapped_file":0,"total_pgfault":972,"total_pgmajfault":0,"total_pgpgin":488,"total_pgpgout":470,"total_rss":65536,"total_rss_huge":0,"total_swap":0,"total_unevictable":0,"total_writeback":0,"unevictable":0,"writeback":0}},"pids":{"current":1},"blkio":{"ioServiceBytesRecursive":[{"major":8,"op":"Read","value":512},{"major":8,"op":"Write"},{"major":8,"op":"Sync","value":512},{"major":8,"op":"Async"},{"major":8,"op":"Total","value":512}],"ioServicedRecursive":[{"major":8,"op":"Read","value":1},{"major":8,"op":"Write"},{"major":8,"op":"Sync","value":1},{"major":8,"op":"Async"},{"major":8,"op":"Total","value":1}],"ioQueueRecursive":[{"major":8,"op":"Read"},{"major":8,"op":"Write"},{"major":8,"op":"Sync"},{"major":8,"op":"Async"},{"major":8,"op":"Total"}],"ioServiceTimeRecursive":[{"major":8,"op":"Read","value":573413},{"major":8,"op":"Write"},{"major":8,"op":"Sync","value":573413},{"major":8,"op":"Async"},{"major":8,"op":"Total","value":573413}],"ioWaitTimeRecursive":[{"major":8,"op":"Read","value":42237},{"major":8,"op":"Write"},{"major":8,"op":"Sync","value":42237},{"major":8,"op":"Async"},{"major":8,"op":"Total","value":42237}],"ioMergedRecursive":[{"major":8,"op":"Read"},{"major":8,"op":"Write"},{"major":8,"op":"Sync"},{"major":8,"op":"Async"},{"major":8,"op":"Total"}],"ioTimeRecursive":[{"major":8,"value":8017998}],"sectorsRecursive":[{"major":8,"value":1}]},"hugetlb":{}}}`)}
