package ceph

import (
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCephGenerateMetrics(t *testing.T) {
	p := &CephMetrics{
		Cluster:     "ceph",
		BinLocation: "/usr/bin/ceph",
		SocketDir:   "/var/run/ceph",
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"ceph_pool",
		// "osd_count",
		// "osd_utilization",
		// "total_storage",
		// "used",
		// "available",
		// "client_io_kbs",
		// "client_io_ops",
		// "pool_used",
		// "pool_usedKb",
		// "pool_maxbytes",
		// "pool_utilization",

		// "osd_used",
		// "osd_total",
		// "osd_epoch",

		// "osd_latency_commit",
		// "osd_latency_apply",

		// "op",
		// "op_in_bytes",
		// "op_out_bytes",
		// "op_r",
		// "op_r_out_byes",
		// "op_w",
		// "op_w_in_bytes",
		// "op_rw",
		// "op_rw_in_bytes",
		// "op_rw_out_bytes",

		// "pool_objects",
		// "pg_map_count",
		// "pg_data_bytes",
		// "pg_data_avail",
		// "pg_data_total",
		// "pg_data_used",
		// "pg_distribution",
		// "pg_distribution_pool",
		// "pg_distribution_osd",
	}

	floatMetrics := []string{
		"sum",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric))
	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatValue(metric))
	}
}

func TestCephGenerateMetricsDefault(t *testing.T) {
	p := &CephMetrics{}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	point, ok := acc.Get("ceph_osd_count")
	require.True(t, ok)
	assert.Equal(t, "ceph", point.Tags["cluster"])

}

func TestCephTagsMetricsWithDifferentCluster(t *testing.T) {
	p := &CephMetrics{
		Cluster: "cephTest",
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	point, ok := acc.Get("cephTest_osd_count")
	require.True(t, ok)

	assert.Equal(t, "cephTest", point.Tags["cluster"])
}
