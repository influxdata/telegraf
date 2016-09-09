package solr

import (
	"fmt"
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

func TestSetDefaults(t *testing.T) {
	solr := newSolrWithClient()
	solr.Servers = []string{"http://example.com:8983"}
	solr.client.Transport = newTransportMock(http.StatusOK, adminCoresResponse)

	cores, max, err := solr.setDefaults()
	if max != 5 {
		err = fmt.Errorf("Received unexpected error: max number of cores: %v, expected 5", max)
	}

	require.NoError(t, err)
}

func TestGatherClusterStats(t *testing.T) {
	solr := newSolrWithClient()
	solr.Servers = []string{"http://example.com:8983"}
	solr.Cores = []string{"main"}
	solr.client.Transport = newTransportMock(http.StatusOK, coreStatsResponse)

	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_core",
		solrCoreExpected,
		map[string]string{"core": "main", "handler": "searcher"})

	solr.client.Transport = newTransportMock(http.StatusOK, queryHandlerStatsResponse)
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_queryhandler",
		solrQueryHandlerExpected,
		map[string]string{"core": "main", "handler": "org.apache.solr.handler.component.SearchHandler"})

	solr.client.Transport = newTransportMock(http.StatusOK, updateHandlerStatsResponse)
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_updatehandler",
		solrUpdateHandlerExpected,
		map[string]string{"core": "main", "handler": "updateHandler"})

	solr.client.Transport = newTransportMock(http.StatusOK, cacheStatsResponse)
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_cache",
		solrCacheExpected,
		map[string]string{"core": "main", "handler": "filterCache"})
}

func newSolrWithClient() *Solr {
	solr := NewSolr()
	solr.client = &http.Client{}
	return solr
}
