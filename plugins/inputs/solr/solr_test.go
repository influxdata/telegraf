package solr

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

func TestGatherClusterStats(t *testing.T) {
	solr := newSolrWithClient()
	solr.Servers = []string{"http://example.com:8983"}
	solr.Cores = []string{"main"}
	solr.client.Transport = newTransportMock(http.StatusOK, coreStatsResponse)

	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_mbean_metrics",
		solrCoreExpected,
		map[string]string{"core": "main", "type": "core", "handler": "searcher"})

	solr.client.Transport = newTransportMock(http.StatusOK, queryHandlerStatsResponse)
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_mbean_metrics",
		solrQueryHandlerExpected,
		map[string]string{"core": "main", "type": "queryhandler", "handler": "org.apache.solr.handler.component.SearchHandler"})

	solr.client.Transport = newTransportMock(http.StatusOK, updateHandlerStatsResponse)
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_mbean_metrics",
		solrUpdateHandlerExpected,
		map[string]string{"core": "main", "type": "updatehandler", "handler": "updateHandler"})

	solr.client.Transport = newTransportMock(http.StatusOK, cacheStatsResponse)
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_mbean_metrics",
		solrCacheExpected,
		map[string]string{"core": "main", "type": "cache", "handler": "filterCache"})
}

func newSolrWithClient() *Solr {
	solr := NewSolr()
	solr.client = &http.Client{}
	return solr
}
