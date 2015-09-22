package elasticsearch

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/assert"
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
