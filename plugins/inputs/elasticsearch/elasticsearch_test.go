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

func defaultTags() map[string]string {
	return map[string]string{
		"cluster_name":          "es-testcluster",
		"node_attribute_master": "true",
		"node_id":               "SDFsfSDFsdfFSDSDfSFDSDF",
		"node_name":             "test.host.com",
		"node_host":             "test",
		"node_roles":            "data,ingest,master",
	}
}
func defaultServerInfo() serverInfo {
	return serverInfo{nodeID: "", masterID: "SDFsfSDFsdfFSDSDfSFDSDF"}
}

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

func checkIsMaster(es *Elasticsearch, server string, expected bool, t *testing.T) {
	if es.serverInfo[server].isMaster() != expected {
		msg := fmt.Sprintf("IsMaster set incorrectly")
		assert.Fail(t, msg)
	}
}

func checkNodeStatsResult(t *testing.T, acc *testutil.Accumulator) {
	tags := defaultTags()
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices", nodestatsIndicesExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_os", nodestatsOsExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_process", nodestatsProcessExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_jvm", nodestatsJvmExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_thread_pool", nodestatsThreadPoolExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_fs", nodestatsFsExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_transport", nodestatsTransportExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_http", nodestatsHTTPExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_breakers", nodestatsBreakersExpected, tags)
}

func TestGather(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	if err := acc.GatherError(es.Gather); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], false, t)
	checkNodeStatsResult(t, &acc)
}

func TestGatherIndividualStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.NodeStats = []string{"jvm", "process"}
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponseJVMProcess)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	if err := acc.GatherError(es.Gather); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], false, t)

	tags := defaultTags()
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_indices", nodestatsIndicesExpected, tags)
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_os", nodestatsOsExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_process", nodestatsProcessExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_jvm", nodestatsJvmExpected, tags)
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_thread_pool", nodestatsThreadPoolExpected, tags)
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_fs", nodestatsFsExpected, tags)
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_transport", nodestatsTransportExpected, tags)
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_http", nodestatsHTTPExpected, tags)
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_breakers", nodestatsBreakersExpected, tags)
}

func TestGatherNodeStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	if err := es.gatherNodeStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], false, t)
	checkNodeStatsResult(t, &acc)
}

func TestGatherClusterHealthEmptyClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.ClusterHealthLevel = ""
	es.client.Transport = newTransportMock(http.StatusOK, clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))

	checkIsMaster(es, es.Servers[0], false, t)

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": "elasticsearch_telegraf"})

	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1"})

	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2"})
}

func TestGatherClusterHealthSpecificClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.ClusterHealthLevel = "cluster"
	es.client.Transport = newTransportMock(http.StatusOK, clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))

	checkIsMaster(es, es.Servers[0], false, t)

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": "elasticsearch_telegraf"})

	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1"})

	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2"})
}

func TestGatherClusterHealthAlsoIndicesHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.ClusterHealthLevel = "indices"
	es.client.Transport = newTransportMock(http.StatusOK, clusterHealthResponseWithIndices)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))

	checkIsMaster(es, es.Servers[0], false, t)

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health",
		clusterHealthExpected,
		map[string]string{"name": "elasticsearch_telegraf"})

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v1IndexExpected,
		map[string]string{"index": "v1", "name": "elasticsearch_telegraf"})

	acc.AssertContainsTaggedFields(t, "elasticsearch_cluster_health_indices",
		v2IndexExpected,
		map[string]string{"index": "v2", "name": "elasticsearch_telegraf"})
}

func TestGatherClusterStatsMaster(t *testing.T) {
	// This needs multiple steps to replicate the multiple calls internally.
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{"http://example.com:9200"}
	es.serverInfo = make(map[string]serverInfo)
	info := serverInfo{nodeID: "SDFsfSDFsdfFSDSDfSFDSDF", masterID: ""}

	// first get catMaster
	es.client.Transport = newTransportMock(http.StatusOK, IsMasterResult)
	masterID, err := es.getCatMaster("junk")
	require.NoError(t, err)
	info.masterID = masterID
	es.serverInfo["http://example.com:9200"] = info

	IsMasterResultTokens := strings.Split(string(IsMasterResult), " ")
	if masterID != IsMasterResultTokens[0] {
		msg := fmt.Sprintf("catmaster is incorrect")
		assert.Fail(t, msg)
	}

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	if err := es.gatherNodeStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, es.Servers[0], true, t)
	checkNodeStatsResult(t, &acc)

	// now test the clusterstats method
	es.client.Transport = newTransportMock(http.StatusOK, clusterStatsResponse)
	require.NoError(t, es.gatherClusterStats("junk", &acc))

	tags := map[string]string{
		"cluster_name": "es-testcluster",
		"node_name":    "test.host.com",
		"status":       "red",
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_clusterstats_nodes", clusterstatsNodesExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_clusterstats_indices", clusterstatsIndicesExpected, tags)
}

func TestGatherClusterStatsNonMaster(t *testing.T) {
	// This needs multiple steps to replicate the multiple calls internally.
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{"http://example.com:9200"}
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = serverInfo{nodeID: "SDFsfSDFsdfFSDSDfSFDSDF", masterID: ""}

	// first get catMaster
	es.client.Transport = newTransportMock(http.StatusOK, IsNotMasterResult)
	masterID, err := es.getCatMaster("junk")
	require.NoError(t, err)

	IsNotMasterResultTokens := strings.Split(string(IsNotMasterResult), " ")
	if masterID != IsNotMasterResultTokens[0] {
		msg := fmt.Sprintf("catmaster is incorrect")
		assert.Fail(t, msg)
	}

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	if err := es.gatherNodeStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	// ensure flag is clear so Cluster Stats would not be done
	checkIsMaster(es, es.Servers[0], false, t)
	checkNodeStatsResult(t, &acc)
}

func TestGatherClusterIndicesStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.IndicesInclude = []string{"_all"}
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, clusterIndicesResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	if err := es.gatherIndicesStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "twitter"})
}

func TestGatherClusterIndiceShardsStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.IndicesLevel = "shards"
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, clusterIndicesShardsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	if err := es.gatherIndicesStats("junk", &acc); err != nil {
		t.Fatal(err)
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "twitter"})

	primaryTags := map[string]string{
		"index_name": "twitter",
		"node_id":    "oqvR8I1dTpONvwRM30etww",
		"shard_name": "0",
		"type":       "primary",
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_shards",
		clusterIndicesPrimaryShardsExpected,
		primaryTags)

	replicaTags := map[string]string{
		"index_name": "twitter",
		"node_id":    "oqvR8I1dTpONvwRM30etww",
		"shard_name": "1",
		"type":       "replica",
	}
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_shards",
		clusterIndicesReplicaShardsExpected,
		replicaTags)
}

func newElasticsearchWithClient() *Elasticsearch {
	es := NewElasticsearch()
	es.client = &http.Client{}
	return es
}
