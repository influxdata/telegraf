package aerospike

import (
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
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

	// Only use a few of the metrics
	asMetrics := []string{
		"transactions",
		"stat_write_errs",
		"stat_read_reqs",
		"stat_write_reqs",
	}

	for _, metric := range asMetrics {
		assert.True(t, acc.HasIntValue(metric), metric)
	}

}

func TestAerospikeMsgLenFromToBytes(t *testing.T) {
	var i int64 = 8
	assert.True(t, i == msgLenFromBytes(msgLenToBytes(i)))
}

func TestReadAerospikeStatsNoNamespace(t *testing.T) {
	// Also test for re-writing
	var acc testutil.Accumulator
	stats := map[string]string{
		"stat-write-errs": "12345",
		"stat_read_reqs":  "12345",
	}
	readAerospikeStats(stats, &acc, "host1", "")
	for k := range stats {
		if k == "stat-write-errs" {
			k = "stat_write_errs"
		}
		assert.True(t, acc.HasMeasurement(k))
		assert.True(t, acc.CheckValue(k, int64(12345)))
	}
}

func TestReadAerospikeStatsNamespace(t *testing.T) {
	var acc testutil.Accumulator
	stats := map[string]string{
		"stat_write_errs": "12345",
		"stat_read_reqs":  "12345",
	}
	readAerospikeStats(stats, &acc, "host1", "test")

	tags := map[string]string{
		"aerospike_host": "host1",
		"namespace":      "test",
	}
	for k := range stats {
		assert.True(t, acc.ValidateTaggedValue(k, int64(12345), tags) == nil)
	}
}

func TestAerospikeUnmarshalList(t *testing.T) {
	i := map[string]string{
		"test": "one;two;three",
	}

	expected := []string{"one", "two", "three"}

	list, err := unmarshalListInfo(i, "test2")
	assert.True(t, err != nil)

	list, err = unmarshalListInfo(i, "test")
	assert.True(t, err == nil)
	equal := true
	for ix := range expected {
		if list[ix] != expected[ix] {
			equal = false
			break
		}
	}
	assert.True(t, equal)
}

func TestAerospikeUnmarshalMap(t *testing.T) {
	i := map[string]string{
		"test": "key1=value1;key2=value2",
	}

	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	m, err := unmarshalMapInfo(i, "test")
	assert.True(t, err == nil)
	assert.True(t, reflect.DeepEqual(m, expected))
}
