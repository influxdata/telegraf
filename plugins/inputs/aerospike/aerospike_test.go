package aerospike

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAerospikeStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}

	a := &Aerospike{
		Servers: []string{testutil.GetLocalHost() + ":3000"},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasTag("aerospike_node", "node_name"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasTag("aerospike_namespace", "node_name"))
	assert.True(t, acc.HasInt64Field("aerospike_node", "batch_error"))
}

func TestAerospikeStatisticsPartialErr(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}

	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
			testutil.GetLocalHost() + ":9999",
		},
	}

	var acc testutil.Accumulator

	require.Error(t, acc.GatherError(a.Gather))

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasInt64Field("aerospike_node", "batch_error"))
}

func TestAerospikeParseValue(t *testing.T) {
	// uint64 with value bigger than int64 max
	val := parseValue("18446744041841121751")
	require.Equal(t, uint64(18446744041841121751), val)

	// int values
	val = parseValue("42")
	require.Equal(t, val, int64(42), "must be parsed as int")

	// string values
	val = parseValue("BB977942A2CA502")
	require.Equal(t, val, `BB977942A2CA502`, "must be left as string")
}
