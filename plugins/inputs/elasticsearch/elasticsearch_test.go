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

func checkIsMaster(es *Elasticsearch, expected bool, t *testing.T) {
	if es.isMaster != expected {
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
	clusterName, err := es.gatherNodeStats("junk", &acc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "es-testcluster", clusterName)

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

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices",
		v1IndexExpected,
		map[string]string{"index": "v1"})

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices",
		v2IndexExpected,
		map[string]string{"index": "v2"})
}

func TestGatherClusterStatsMaster(t *testing.T) {
	// This needs multiple steps to replicate the multiple calls internally.
	es := newElasticsearchWithClient()
	es.ClusterStats = true
	es.Servers = []string{"http://example.com:9200"}

	// first get catMaster
	es.client.Transport = newTransportMock(http.StatusOK, IsMasterResult)
	require.NoError(t, es.setCatMaster("junk"))

	IsMasterResultTokens := strings.Split(string(IsMasterResult), " ")
	if es.catMasterResponseTokens[0] != IsMasterResultTokens[0] {
		msg := fmt.Sprintf("catmaster is incorrect")
		assert.Fail(t, msg)
	}

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	_, err := es.gatherNodeStats("junk", &acc)
	if err != nil {
		t.Fatal(err)
	}

	checkIsMaster(es, true, t)
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

	// first get catMaster
	es.client.Transport = newTransportMock(http.StatusOK, IsNotMasterResult)
	require.NoError(t, es.setCatMaster("junk"))

	IsNotMasterResultTokens := strings.Split(string(IsNotMasterResult), " ")
	if es.catMasterResponseTokens[0] != IsNotMasterResultTokens[0] {
		msg := fmt.Sprintf("catmaster is incorrect")
		assert.Fail(t, msg)
	}

	// now get node status, which determines whether we're master
	var acc testutil.Accumulator
	es.Local = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)
	_, err := es.gatherNodeStats("junk", &acc)
	if err != nil {
		t.Fatal(err)
	}

	// ensure flag is clear so Cluster Stats would not be done
	checkIsMaster(es, false, t)
	checkNodeStatsResult(t, &acc)

}

func TestGatherIndicesStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.IndicesStats = true
	es.client.Transport = newTransportMock(http.StatusOK, indicesStatsResponse)

	var acc testutil.Accumulator
	require.NoError(t, es.gatherIndicesStats("es-testcluster", "junk", &acc))

	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats",
		indicesStatsShardsExpected,
		map[string]string{
			"cluster_name": "es-testcluster",
			"stat_name":    "shards",
		})

	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats",
		indicesStatsAllExpected,
		map[string]string{
			"cluster_name": "es-testcluster",
			"stat_name":    "all",
		})

	acc.AssertContainsTaggedFields(t, "elasticsearch_indicesstats_indices",
		v1IndexStatsExpected,
		map[string]string{
			"cluster_name": "es-testcluster",
			"index":        "v1",
		})

}

func TestGetClusterName(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.IndicesStats = true
	es.client.Transport = newTransportMock(http.StatusOK, nodeStatsResponse)

	clusterName, err := es.getClusterName("junk")
	require.NoError(t, err)

	assert.Equal(t, "es-testcluster", clusterName)
}

func newElasticsearchWithClient() *Elasticsearch {
	es := NewElasticsearch()
	es.client = &http.Client{}
	return es
}
