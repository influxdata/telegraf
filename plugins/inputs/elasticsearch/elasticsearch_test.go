package elasticsearch

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"

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

func TestElasticsearch(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.client.Transport = newTransportMock(http.StatusOK, statsResponse)

	var acc testutil.Accumulator
	if err := es.Gather(&acc); err != nil {
		t.Fatal(err)
	}

	tags := map[string]string{
		"cluster_name":          "es-testcluster",
		"node_attribute_master": "true",
		"node_id":               "SDFsfSDFsdfFSDSDfSFDSDF",
		"node_name":             "test.host.com",
		"node_host":             "test",
	}

	acc.AssertContainsTaggedFields(t, "elasticsearch_indices", indicesExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_os", osExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_process", processExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_jvm", jvmExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_thread_pool", threadPoolExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_fs", fsExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_transport", transportExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_http", httpExpected, tags)
	acc.AssertContainsTaggedFields(t, "elasticsearch_breakers", breakersExpected, tags)
}

func TestGatherClusterStats(t *testing.T) {
	es := newElasticsearchWithClient()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.client.Transport = newTransportMock(http.StatusOK, clusterResponse)

	var acc testutil.Accumulator
	require.NoError(t, es.Gather(&acc))

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

func newElasticsearchWithClient() *Elasticsearch {
	es := NewElasticsearch()
	es.client = &http.Client{}
	return es
}
