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
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasIntField("aerospike_node", "batch_error"))
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
	assert.True(t, acc.HasIntField("aerospike_node", "batch_error"))
}

func TestAerospikeParseValue(t *testing.T) {
	// uint64 with value bigger than int64 max
	val, err := parseValue("18446744041841121751")
	assert.Nil(t, val)
	assert.Error(t, err)

	// int values
	val, err = parseValue("42")
	assert.NoError(t, err)
	assert.Equal(t, val, int64(42), "must be parsed as int")

	// string values
	val, err = parseValue("BB977942A2CA502")
	assert.NoError(t, err)
	assert.Equal(t, val, `BB977942A2CA502`, "must be left as string")
}
