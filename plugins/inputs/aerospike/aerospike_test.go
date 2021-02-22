package aerospike

import (
	"testing"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAerospikeStatisticsIntegration(t *testing.T) {
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
	assert.True(t, acc.HasInt64Field("aerospike_node", "batch_index_error"))

	namespaceName := acc.TagValue("aerospike_namespace", "namespace")
	assert.Equal(t, namespaceName, "test")

}

func TestAerospikeStatisticsPartialErrIntegration(t *testing.T) {
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
	err := acc.GatherError(a.Gather)

	require.Error(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasInt64Field("aerospike_node", "batch_index_error"))
	namespaceName := acc.TagSetValue("aerospike_namespace", "namespace")
	assert.Equal(t, namespaceName, "test")
}

func TestSelectNamepsacesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}

	// Select nonexistent namespace
	a := &Aerospike{
		Servers:    []string{testutil.GetLocalHost() + ":3000"},
		Namespaces: []string{"notTest"},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasTag("aerospike_node", "node_name"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
	assert.True(t, acc.HasTag("aerospike_namespace", "node_name"))

	// Expect only 1 namespace
	count := 0
	for _, p := range acc.Metrics {
		if p.Measurement == "aerospike_namespace" {
			count++
		}
	}
	assert.Equal(t, count, 1)

	// expect namespace to have no fields as nonexistent
	assert.False(t, acc.HasInt64Field("aerospke_namespace", "appeals_tx_remaining"))
}

func TestDisableQueryNamespacesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}

	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		DisableQueryNamespaces: true,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.False(t, acc.HasMeasurement("aerospike_namespace"))

	a.DisableQueryNamespaces = false
	err = acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_node"))
	assert.True(t, acc.HasMeasurement("aerospike_namespace"))
}

func TestQuerySetsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}

	// create a set
	// test is the default namespace from aerospike
	policy := as.NewClientPolicy()
	client, err := as.NewClientWithPolicy(policy, testutil.GetLocalHost(), 3000)

	key, err := as.NewKey("test", "foo", 123)
	require.NoError(t, err)
	bins := as.BinMap{
		"e":  2,
		"pi": 3,
	}
	err = client.Add(nil, key, bins)
	require.NoError(t, err)

	key, err = as.NewKey("test", "bar", 1234)
	require.NoError(t, err)
	bins = as.BinMap{
		"e":  2,
		"pi": 3,
	}
	err = client.Add(nil, key, bins)
	require.NoError(t, err)

	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		QuerySets:              true,
		DisableQueryNamespaces: true,
	}

	var acc testutil.Accumulator
	err = acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, FindTagValue(&acc, "aerospike_set", "set", "test/foo"))
	assert.True(t, FindTagValue(&acc, "aerospike_set", "set", "test/bar"))

	assert.True(t, acc.HasMeasurement("aerospike_set"))
	assert.True(t, acc.HasTag("aerospike_set", "set"))
	assert.True(t, acc.HasInt64Field("aerospike_set", "memory_data_bytes"))

}

func TestSelectQuerySetsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}

	// create a set
	// test is the default namespace from aerospike
	policy := as.NewClientPolicy()
	client, err := as.NewClientWithPolicy(policy, testutil.GetLocalHost(), 3000)

	key, err := as.NewKey("test", "foo", 123)
	require.NoError(t, err)
	bins := as.BinMap{
		"e":  2,
		"pi": 3,
	}
	err = client.Add(nil, key, bins)
	require.NoError(t, err)

	key, err = as.NewKey("test", "bar", 1234)
	require.NoError(t, err)
	bins = as.BinMap{
		"e":  2,
		"pi": 3,
	}
	err = client.Add(nil, key, bins)
	require.NoError(t, err)

	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		QuerySets:              true,
		Sets:                   []string{"test/foo"},
		DisableQueryNamespaces: true,
	}

	var acc testutil.Accumulator
	err = acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, FindTagValue(&acc, "aerospike_set", "set", "test/foo"))
	assert.False(t, FindTagValue(&acc, "aerospike_set", "set", "test/bar"))

	assert.True(t, acc.HasMeasurement("aerospike_set"))
	assert.True(t, acc.HasTag("aerospike_set", "set"))
	assert.True(t, acc.HasInt64Field("aerospike_set", "memory_data_bytes"))

}

func TestDisableTTLHistogramIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}
	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		QuerySets:          true,
		EnableTTLHistogram: false,
	}
	/*
		No measurement exists
	*/
	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.False(t, acc.HasMeasurement("aerospike_histogram_ttl"))
}
func TestTTLHistogramIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	} else {
		t.Skip("Skipping, only passes if the aerospike db has been running for at least 1 hour")
	}
	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		QuerySets:          true,
		EnableTTLHistogram: true,
	}
	/*
		Produces histogram
		Measurment exists
		Has appropriate tags (node name etc)
		Has appropriate keys (time:value)
		may be able to leverage histogram plugin
	*/
	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("aerospike_histogram_ttl"))
	assert.True(t, FindTagValue(&acc, "aerospike_histogram_ttl", "namespace", "test"))

}
func TestDisableObjectSizeLinearHistogramIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	}
	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		QuerySets:                       true,
		EnableObjectSizeLinearHistogram: false,
	}
	/*
		No Measurement
	*/
	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	assert.False(t, acc.HasMeasurement("aerospike_histogram_object_size_linear"))
}
func TestObjectSizeLinearHistogramIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping aerospike integration tests.")
	} else {
		t.Skip("Skipping, only passes if the aerospike db has been running for at least 1 hour")
	}
	a := &Aerospike{
		Servers: []string{
			testutil.GetLocalHost() + ":3000",
		},
		QuerySets:                       true,
		EnableObjectSizeLinearHistogram: true,
	}
	/*
		Produces histogram
		Measurment exists
		Has appropriate tags (node name etc)
		Has appropriate keys (time:value)

	*/
	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)
	assert.True(t, acc.HasMeasurement("aerospike_histogram_object_size_linear"))
	assert.True(t, FindTagValue(&acc, "aerospike_histogram_object_size_linear", "namespace", "test"))
}

func TestParseNodeInfo(t *testing.T) {
	a := &Aerospike{}
	var acc testutil.Accumulator

	stats := map[string]string{
		"early_tsvc_from_proxy_error": "0",
		"cluster_principal":           "BB9020012AC4202",
		"cluster_is_member":           "true",
	}

	expectedFields := map[string]interface{}{
		"early_tsvc_from_proxy_error": int64(0),
		"cluster_principal":           "BB9020012AC4202",
		"cluster_is_member":           true,
	}

	expectedTags := map[string]string{
		"aerospike_host": "127.0.0.1:3000",
		"node_name":      "TestNodeName",
	}

	a.parseNodeInfo(stats, "127.0.0.1:3000", "TestNodeName", &acc)
	acc.AssertContainsTaggedFields(t, "aerospike_node", expectedFields, expectedTags)
}

func TestParseNamespaceInfo(t *testing.T) {
	a := &Aerospike{}
	var acc testutil.Accumulator

	stats := map[string]string{
		"namespace/test": "ns_cluster_size=1;effective_replication_factor=1;objects=2;tombstones=0;master_objects=2",
	}

	expectedFields := map[string]interface{}{
		"ns_cluster_size":              int64(1),
		"effective_replication_factor": int64(1),
		"tombstones":                   int64(0),
		"objects":                      int64(2),
		"master_objects":               int64(2),
	}

	expectedTags := map[string]string{
		"aerospike_host": "127.0.0.1:3000",
		"node_name":      "TestNodeName",
		"namespace":      "test",
	}

	a.parseNamespaceInfo(stats, "127.0.0.1:3000", "test", "TestNodeName", &acc)
	acc.AssertContainsTaggedFields(t, "aerospike_namespace", expectedFields, expectedTags)
}

func TestParseSetInfo(t *testing.T) {
	a := &Aerospike{}

	var acc testutil.Accumulator

	stats := map[string]string{
		"sets/test/foo": "objects=1:tombstones=0:memory_data_bytes=26;",
	}

	expectedFields := map[string]interface{}{
		"objects":           int64(1),
		"tombstones":        int64(0),
		"memory_data_bytes": int64(26),
	}

	expectedTags := map[string]string{
		"aerospike_host": "127.0.0.1:3000",
		"node_name":      "TestNodeName",
		"set":            "test/foo",
	}
	a.parseSetInfo(stats, "127.0.0.1:3000", "test/foo", "TestNodeName", &acc)
	acc.AssertContainsTaggedFields(t, "aerospike_set", expectedFields, expectedTags)
}

func TestParseHistogramSet(t *testing.T) {
	a := &Aerospike{
		NumberHistogramBuckets: 10,
	}

	var acc testutil.Accumulator

	stats := map[string]string{
		"histogram:type=object-size-linear;namespace=test;set=foo": "units=bytes:hist-width=1048576:bucket-width=1024:buckets=0,1,3,1,6,1,9,1,12,1,15,1,18",
	}

	expectedFields := map[string]interface{}{
		"0": int64(1),
		"1": int64(4),
		"2": int64(7),
		"3": int64(10),
		"4": int64(13),
		"5": int64(16),
		"6": int64(18),
	}

	expectedTags := map[string]string{
		"aerospike_host": "127.0.0.1:3000",
		"node_name":      "TestNodeName",
		"namespace":      "test",
		"set":            "foo",
	}

	a.parseHistogram(stats, "127.0.0.1:3000", "test", "foo", "object-size-linear", "TestNodeName", &acc)
	acc.AssertContainsTaggedFields(t, "aerospike_histogram_object_size_linear", expectedFields, expectedTags)

}
func TestParseHistogramNamespace(t *testing.T) {
	a := &Aerospike{
		NumberHistogramBuckets: 10,
	}

	var acc testutil.Accumulator

	stats := map[string]string{
		"histogram:type=object-size-linear;namespace=test;set=foo": " units=bytes:hist-width=1048576:bucket-width=1024:buckets=0,1,3,1,6,1,9,1,12,1,15,1,18",
	}

	expectedFields := map[string]interface{}{
		"0": int64(1),
		"1": int64(4),
		"2": int64(7),
		"3": int64(10),
		"4": int64(13),
		"5": int64(16),
		"6": int64(18),
	}

	expectedTags := map[string]string{
		"aerospike_host": "127.0.0.1:3000",
		"node_name":      "TestNodeName",
		"namespace":      "test",
	}

	a.parseHistogram(stats, "127.0.0.1:3000", "test", "", "object-size-linear", "TestNodeName", &acc)
	acc.AssertContainsTaggedFields(t, "aerospike_histogram_object_size_linear", expectedFields, expectedTags)

}
func TestAerospikeParseValue(t *testing.T) {
	// uint64 with value bigger than int64 max
	val := parseAerospikeValue("", "18446744041841121751")
	require.Equal(t, val, uint64(18446744041841121751))

	val = parseAerospikeValue("", "true")
	require.Equal(t, val, true)

	// int values
	val = parseAerospikeValue("", "42")
	require.Equal(t, int64(42), val, "must be parsed as an int64")

	// string values
	val = parseAerospikeValue("", "BB977942A2CA502")
	require.Equal(t, `BB977942A2CA502`, val, "must be left as a string")

	// all digit hex values, unprotected
	val = parseAerospikeValue("", "1992929191")
	require.Equal(t, int64(1992929191), val, "must be parsed as an int64")

	// all digit hex values, protected
	val = parseAerospikeValue("node_name", "1992929191")
	require.Equal(t, `1992929191`, val, "must be left as a string")
}

func FindTagValue(acc *testutil.Accumulator, measurement string, key string, value string) bool {
	for _, p := range acc.Metrics {
		if p.Measurement == measurement {
			v, ok := p.Tags[key]
			if ok && v == value {
				return true
			}

		}
	}
	return false
}
