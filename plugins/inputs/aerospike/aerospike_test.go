package aerospike

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAerospikeStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	a := &Aerospike{
		Servers: []string{testutil.GetLocalHost() + ":3000"},
	}

	var acc testutil.Accumulator

	err := a.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasIntField("aerospike_node", "batch_error"))
}

func TestAerospikeStatisticsPartialErr(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
			testutil.GetLocalHost() + ":9999",
		},
	}

	var acc testutil.Accumulator

	err := a.Gather(&acc)
	require.Error(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasIntField("aerospike_node", "batch_error"))
}
