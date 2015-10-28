package elasticsearch

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdb/telegraf/testutil"
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

func TestElasticsearch(t *testing.T) {
	es := NewElasticsearch()
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

	testTables := []map[string]float64{
		indicesExpected,
		osExpected,
		processExpected,
		jvmExpected,
		threadPoolExpected,
		networkExpected,
		fsExpected,
		transportExpected,
		httpExpected,
		breakersExpected,
	}

	for _, testTable := range testTables {
		for k, v := range testTable {
			assert.NoError(t, acc.ValidateTaggedValue(k, v, tags))
		}
	}
}

func TestGatherClusterStats(t *testing.T) {
	es := NewElasticsearch()
	es.Servers = []string{"http://example.com:9200"}
	es.ClusterHealth = true
	es.client.Transport = newTransportMock(http.StatusOK, clusterResponse)

	var acc testutil.Accumulator
	require.NoError(t, es.Gather(&acc))

	var clusterHealthTests = []struct {
		measurement string
		fields      map[string]interface{}
		tags        map[string]string
	}{
		{
			"cluster_health",
			clusterHealthExpected,
			map[string]string{"name": "elasticsearch_telegraf"},
		},
		{
			"indices",
			v1IndexExpected,
			map[string]string{"index": "v1"},
		},
		{
			"indices",
			v2IndexExpected,
			map[string]string{"index": "v2"},
		},
	}

	for _, exp := range clusterHealthTests {
		assert.NoError(t, acc.ValidateTaggedFields(exp.measurement, exp.fields, exp.tags))
	}
}
