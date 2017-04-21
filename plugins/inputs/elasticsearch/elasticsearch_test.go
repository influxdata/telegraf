package elasticsearch

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type transportMock struct {
	statusCode int
	body       string
}

func newTransportMock(statusCode int, body string) http.RoundTripper {
	return &transportMock{
		statusCode: statusCode,
		body:       body,
	}
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {
}

func newElasticsearchWithClient() *Elasticsearch {
	es := NewElasticsearch()
	es.client = &http.Client{}
	return es
}

func checkIsMaster(es *Elasticsearch, expected bool, t *testing.T) {
	if es.localNodeIsMaster != expected {
		msg := fmt.Sprintf("IsMaster set incorrectly")
		assert.Fail(t, msg)
	}
}
func checkNodeStatsResult(t *testing.T, acc *testutil.Accumulator) {
	tags := map[string]string{
		"cluster_name":          "es-testcluster",
		"node_attribute_master": "true",
		"node_id":               "SDFsfSDFsdfFSDSDfSFDSDF",
		"node_name":             "test.host.com",
		"node_host":             "test",
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices", nodestatsIndicesExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_os", nodestatsOsExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_process", nodestatsProcessExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_jvm", nodestatsJvmExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_thread_pool", nodestatsThreadPoolExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_fs", nodestatsFsExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_transport", nodestatsTransportExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_http", nodestatsHttpExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_breakers", nodestatsBreakersExpected, tags)
}

func TestGather(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)

	var acc testutil.Accumulator
	if err := es.Gather(&acc); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, false, t)
	checkNodeStatsResult(t, &acc)
}

func TestGatherNodeStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)

	var acc testutil.Accumulator
	if err := es.gatherNodeStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, false, t)
	checkNodeStatsResult(t, &acc)
}

func TestGatherClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.client.Transport = newTransportMock(http.StatusOK, clusterHealthResponse)

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))

	checkIsMaster(es, false, t)

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": "elasticsearch_telegraf"})

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"name": "v1"})

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"name": "v2"})
}

func TestGatherClusterStatsMaster(t *testing.T) {
	// This needs multiple steps to replicate the multiple calls internally.
	var acc testutil.Accumulator
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{"http://example.com:9200"}

	// first get catMaster
	es.client.Transport = newTransportMock(http.StatusOK, IsMasterResult)
	require.NoError(t, es.gatherCatMaster("junk", &acc))

	if es.masterNodeId != catMasterNodeId {
		msg := fmt.Sprintf("catmaster is incorrect [" + es.masterNodeId + "] vs [" + catMasterNodeId + "]")
		assert.Fail(t, msg)
	}

	// now get node status, which determines whether we're master
	es.Local = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	if err := es.gatherNodeStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, true, t)
	checkNodeStatsResult(t, &acc)

	// now test the clusterstats method
	es.client.Transport = newTransportMock(http.StatusOK, clusterStatsResponse)
	require.NoError(t, es.gatherClusterStats("junk", &acc))

	acc.AssertContainsTaggedFields(t, "elasticsearch_clusterstats", clusterstatsExpected, map[string]string{"name": ""})
	acc.AssertContainsTaggedFields(t, "elasticsearch_clusterstats_nodes", clusterstatsNodesExpected, map[string]string{"name": ""})
	acc.AssertContainsTaggedFields(t, "elasticsearch_clusterstats_indices", clusterstatsIndicesExpected, map[string]string{"name": ""})
}

func TestGatherClusterStatsNonMaster(t *testing.T) {
	// This needs multiple steps to replicate the multiple calls internally.
	var acc testutil.Accumulator
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{"http://example.com:9200"}

	// first get catMaster
	es.client.Transport = newTransportMock(http.StatusOK, IsNotMasterResult)
	require.NoError(t, es.gatherCatMaster("junk", &acc))

	if es.masterNodeId == catMasterNodeId {
		msg := fmt.Sprintf("catmaster is incorrect [" + es.masterNodeId + "] vs [" + catMasterNodeId + "]")
		assert.Fail(t, msg)
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_catmaster", catNotMasterStatsExpected, map[string]string{"name": ""})

	// now get node status, which determines whether we're master
	es.Local = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	if err := es.gatherNodeStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	// ensure flag is clear so Cluster Stats would not be done
	checkIsMaster(es, false, t)
	checkNodeStatsResult(t, &acc)

}

func TestGatherIndicesStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.IndicesIntervalMultiplier = 1
	es.IndicesStats = true
	es.client.Transport = newTransportMock(http.StatusOK, IndicesStatsResponse)

	var acc testutil.Accumulator
	require.NoError(t, es.gatherIndicesStats("junk", &acc, false))

	checkIsMaster(es, false, t)

	shardsTags := map[string]string{"name": ""}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_shards", Indices_ShardStatsExpected, shardsTags)

	allName := map[string]string{"index_name": "all"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_primaries", AllPrimaryStatsExpected, allName)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_total", AllTotalStatsExpected, allName)

	v1Name := map[string]string{"index_name": "test-index-1"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_primaries", V1IndicesPrimaryStatsExpected, v1Name)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_total", V1IndicesTotalStatsExpected, v1Name)

	v2Name := map[string]string{"index_name": "test-index-2"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_primaries", V2IndicesPrimaryStatsExpected, v2Name)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_total", V2IndicesTotalStatsExpected, v2Name)

}

func TestGatherShardsStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.IndicesIntervalMultiplier = 1
	es.IndicesStats = true
	es.ShardsStats = true
	es.client.Transport = newTransportMock(http.StatusOK, IndicesStatsResponse)

	var acc testutil.Accumulator
	require.NoError(t, es.gatherIndicesStats("junk", &acc, true))

	checkIsMaster(es, false, t)

	shardsTags := map[string]string{"name": ""}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_shards", Indices_ShardStatsExpected, shardsTags)

	allName := map[string]string{"index_name": "all"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_primaries", AllPrimaryStatsExpected, allName)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_total", AllTotalStatsExpected, allName)

	v1Name := map[string]string{"index_name": "test-index-1"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_primaries", V1IndicesPrimaryStatsExpected, v1Name)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_total", V1IndicesTotalStatsExpected, v1Name)

	v2Name := map[string]string{"index_name": "test-index-2"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_primaries", V2IndicesPrimaryStatsExpected, v2Name)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_total", V2IndicesTotalStatsExpected, v2Name)

	// shards stats are in index 2
	shardTags0 := map[string]string{"index_name": "test-index-2", "shard_name": "0"}
	shardTags1 := map[string]string{"index_name": "test-index-2", "shard_name": "1"}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_shards_primary", Index2Shard0PrimaryExpected, shardTags0)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_shards_primary", Index2Shard1PrimaryExpected, shardTags1)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_shards_replica", Index2Shard0ReplicaExpected, shardTags0)
	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_shards_replica", Index2Shard1ReplicaExpected, shardTags1)

}
