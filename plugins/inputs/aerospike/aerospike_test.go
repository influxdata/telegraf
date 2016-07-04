package aerospike

import (
	"bytes"
	"encoding/binary"
	"reflect"
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
		Server: testutil.GetLocalHost() + ":3000",
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
		assert.True(t, acc.HasIntField("aerospike", metric), metric)
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

	fields_ := make(map[string]interface{})
	readAerospikeStats(stats, fields_, "host1", "")
	tags_ := map[string]string{
		"aerospike_host": "host1",
		"namespace":      "_service",
	}
	acc.AddFields("aerospike", fields_, tags_)

	fields := map[string]interface{}{
		"stat_write_errs": int64(12345),
		"stat_read_reqs":  int64(12345),
	}
	tags := map[string]string{
		"aerospike_host": "host1",
		"namespace":      "_service",
	}
	acc.AssertContainsTaggedFields(t, "aerospike", fields, tags)
}

func TestReadAerospikeStatsNamespace(t *testing.T) {
	var acc testutil.Accumulator
	stats := map[string]string{
		"stat_write_errs": "12345",
		"stat_read_reqs":  "12345",
	}

	fields_ := make(map[string]interface{})
	readAerospikeStats(stats, fields_, "host1", "test")
	tags_ := map[string]string{
		"aerospike_host": "host1",
		"namespace":      "test",
	}

	acc.AddFields("aerospike", fields_, tags_)

	fields := map[string]interface{}{
		"stat_write_errs": int64(12345),
		"stat_read_reqs":  int64(12345),
	}
	tags := map[string]string{
		"aerospike_host": "host1",
		"namespace":      "test",
	}
	acc.AssertContainsTaggedFields(t, "aerospike", fields, tags)
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

func TestFieldSerialization(t *testing.T) {
	typeId := byte(12)
	data := []byte("test string")

	f := newField(typeId, data)

	buf := bytes.Buffer{}
	f.WriteToBuf(&buf)

	var sz int32
	var rType byte

	binary.Read(&buf, binary.BigEndian, &sz)
	binary.Read(&buf, binary.BigEndian, &rType)

	assert.True(t, int(sz) == len(data)+1)
	assert.True(t, rType == typeId)

	assert.True(t, bytes.Equal(data, buf.Bytes()))

}
