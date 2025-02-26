package elasticsearch

import (
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
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

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{
		statusCode: http.StatusOK,
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
	res.Body = io.NopCloser(strings.NewReader(t.body))
	return res, nil
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
	es.client.Transport = newTransportMock(nodeStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(es.Gather))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")
	checkNodeStatsResult(t, &acc)
}

func TestGatherIndividualStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.NodeStats = []string{"jvm", "process"}
	es.client.Transport = newTransportMock(nodeStatsResponseJVMProcess)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(es.Gather))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")

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

func TestGatherEnrichStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.EnrichStats = true
	es.client.Transport = newTransportMock(enrichStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(es.Gather))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")

	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 8)
}

func TestGatherNodeStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(nodeStatsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherNodeStats("junk", &acc))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")
	checkNodeStatsResult(t, &acc)
}

func TestGatherClusterHealthEmptyClusterHealth(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.ClusterHealthLevel = ""
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")

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
	es.client.Transport = newTransportMock(clusterHealthResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")

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
	es.client.Transport = newTransportMock(clusterHealthResponseWithIndices)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherClusterHealth("junk", &acc))
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")

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
	es.client.Transport = newTransportMock(IsMasterResult)
	masterID, err := es.getCatMaster("junk")
	require.NoError(t, err)
	info.masterID = masterID
	es.serverInfo["http://example.com:9200"] = info

	isMasterResultTokens := strings.Split(IsMasterResult, " ")
	require.Equal(t, masterID, isMasterResultTokens[0], "catmaster is incorrect")

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(nodeStatsResponse)
	require.NoError(t, es.gatherNodeStats("junk", &acc))
	require.True(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")
	checkNodeStatsResult(t, &acc)

	// now test the clusterstats method
	es.client.Transport = newTransportMock(clusterStatsResponse)
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
	es.client.Transport = newTransportMock(IsNotMasterResult)
	masterID, err := es.getCatMaster("junk")
	require.NoError(t, err)

	isNotMasterResultTokens := strings.Split(IsNotMasterResult, " ")
	require.Equal(t, masterID, isNotMasterResultTokens[0], "catmaster is incorrect")

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(nodeStatsResponse)
	require.NoError(t, es.gatherNodeStats("junk", &acc))

	// ensure flag is clear so Cluster Stats would not be done
	require.False(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")
	checkNodeStatsResult(t, &acc)
}

func TestGatherCCRStatsMaster(t *testing.T) {
	// This needs multiple steps to replicate the multiple calls internally.
	es := newElasticsearchWithClient()
	es.CCRStats = true
	es.Servers = []string{"http://example.com:9200"}
	es.serverInfo = make(map[string]serverInfo)
	info := serverInfo{nodeID: "SDFsfSDFsdfFSDSDfSFDSDF", masterID: ""}

	// first get catMaster
	es.client.Transport = newTransportMock(IsMasterResult)
	masterID, err := es.getCatMaster("junk")
	require.NoError(t, err)
	info.masterID = masterID
	es.serverInfo["http://example.com:9200"] = info

	isMasterResultTokens := strings.Split(IsMasterResult, " ")
	require.Equal(t, masterID, isMasterResultTokens[0], "catmaster is incorrect")

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(nodeStatsResponse)
	require.NoError(t, es.gatherNodeStats("junk", &acc))
	require.True(t, es.serverInfo[es.Servers[0]].isMaster(), "IsMaster set incorrectly")
	checkNodeStatsResult(t, &acc)

	tags := map[string]string{}

	// now test the ccr leader method
	es.client.Transport = newTransportMock(ccrLeaderResponse)
	require.NoError(t, es.gatherCCRLeaderStats("junk", &acc))

	acc.AssertContainsTaggedFields(t, "elasticsearch_ccr_stats_leader", ccrLeaderStatsExpected, tags)

	// now test the ccr follower method
	es.client.Transport = newTransportMock(ccrFollowerResponse)
	require.NoError(t, es.gatherCCRFollowerStats("junk", &acc))

	acc.AssertContainsTaggedFields(t, "elasticsearch_ccr_stats_follower", ccrFollowerStatsExpected, tags)
}

func TestGatherClusterIndicesStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.IndicesInclude = []string{"_all"}
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(clusterIndicesResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherIndicesStats("junk", &acc))

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "twitter"})
}

func TestGatherDateStampedIndicesStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.IndicesInclude = []string{"twitter*", "influx*", "penguins"}
	es.NumMostRecentIndices = 2
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(dateStampedIndicesResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()
	require.NoError(t, es.Init())

	var acc testutil.Accumulator
	require.NoError(t, es.gatherIndicesStats(es.Servers[0]+"/"+strings.Join(es.IndicesInclude, ",")+"/_stats", &acc))

	// includes 2 most recent indices for "twitter", only expect the most recent two.
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "twitter_2020_08_02"})
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "twitter_2020_08_01"})
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "twitter_2020_07_31"})

	// includes 2 most recent indices for "influx", only expect the most recent two.
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "influx2021.01.02"})
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "influx2021.01.01"})
	acc.AssertDoesNotContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "influx2020.12.31"})

	// not configured to sort the 'penguins' index, but ensure it is also included.
	acc.AssertContainsTaggedFields(t, "elasticsearch_indices_stats_primaries",
		clusterIndicesExpected,
		map[string]string{"index_name": "penguins"})
}

func TestGatherClusterIndiceShardsStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.IndicesLevel = "shards"
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(clusterIndicesShardsResponse)
	es.serverInfo = make(map[string]serverInfo)
	es.serverInfo["http://example.com:9200"] = defaultServerInfo()

	var acc testutil.Accumulator
	require.NoError(t, es.gatherIndicesStats("junk", &acc))

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

func TestGatherRemoteStoreStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.RemoteStoreStats = true
	es.Servers = []string{"http://example.com:9200"}
	es.IndicesInclude = []string{"remote-index"}
	es.client.Transport = newTransportMock(remoteStoreResponse)
	var acc testutil.Accumulator
	url := "http://example.com:9200/_remotestore/stats/remote-index"
	err := es.gatherRemoteStoreStats(url, "remote-index", &acc)
	require.NoError(t, err)

	globalTags := map[string]string{"index_name": "remote-index"}
	expectedGlobalFields := map[string]interface{}{
		"total":      float64(4),
		"successful": float64(4),
		"failed":     float64(0),
	}
	acc.AssertContainsTaggedFields(t, "elasticsearch_remotestore_global", expectedGlobalFields, globalTags)

	primaryTags := map[string]string{
		"index_name":    "remote-index",
		"shard_id":      "0",
		"routing_state": "STARTED",
		"shard_type":    "primary",
		"node_id":       "q1VxWZnCTICrfRc2bRW3nw",
	}
	acc.AssertContainsTaggedFields(t, "elasticsearch_remotestore_stats_shards", remoteStoreIndicesPrimaryShardsExpected, primaryTags)

	replicaTags := map[string]string{
		"index_name":    "remote-index",
		"shard_id":      "1",
		"routing_state": "STARTED",
		"shard_type":    "replica",
		"node_id":       "q1VxWZnCTICrfRc2bRW3nw",
	}
	acc.AssertContainsTaggedFields(t, "elasticsearch_remotestore_stats_shards", remoteStoreIndicesReplicaShardsExpected, replicaTags)
}

func newElasticsearchWithClient() *Elasticsearch {
	es := newElasticsearch()
	es.client = &http.Client{}
	return es
}
