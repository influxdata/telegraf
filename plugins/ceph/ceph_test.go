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

	sample := p.SampleConfig()
	assert.NotNil(t, sample)
	assert.Equal(t, p.Cluster, "ceph", "Same Cluster")

	intMetrics := []string{
		"total_storage",
		"used_storage",
		"available_storage",
		"client_io_kbs",
		"client_io_ops",

		"pool_used",
		"pool_usedKb",
		"pool_maxbytes",
		"pool_objects",

		"osd_epoch",
		"op_in_bytes",
		"op_out_bytes",
		"op_r",
		"op_w",
		"op_w_in_bytes",
		"op_rw",
		"op_rw_in_bytes",
		"op_rw_out_bytes",

		"pg_map_count",
		"pg_data_bytes",
		"pg_data_total_storage",
		"pg_data_used_storage",
		"pg_distribution",
		"pg_distribution_pool",
		"pg_distribution_osd",
	}

	floatMetrics := []string{
		"osd_utilization",
		"pool_utilization",
		"osd_used_storage",
		"osd_total_storage",
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
	assert.True(t, len(acc.Points) > 0)

	point, ok := acc.Get("op_wip")
	require.True(t, ok)
	assert.Equal(t, "ceph", point.Tags["cluster"])

}
