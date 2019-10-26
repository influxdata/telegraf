package solr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGatherStats(t *testing.T) {
	ts := createMockServer()
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminMainCoreStatusExpected,
		map[string]string{"core": "main"})

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminCore1StatusExpected,
		map[string]string{"core": "core1"})

	acc.AssertContainsTaggedFields(t, "solr_core",
		solrCoreExpected,
		map[string]string{"core": "main", "handler": "searcher"})

	acc.AssertContainsTaggedFields(t, "solr_queryhandler",
		solrQueryHandlerExpected,
		map[string]string{"core": "main", "handler": "org.apache.solr.handler.component.SearchHandler"})

	acc.AssertContainsTaggedFields(t, "solr_updatehandler",
		solrUpdateHandlerExpected,
		map[string]string{"core": "main", "handler": "updateHandler"})

	acc.AssertContainsTaggedFields(t, "solr_cache",
		solrCacheExpected,
		map[string]string{"core": "main", "handler": "filterCache"})
}

func TestSolr7MbeansStats(t *testing.T) {
	ts := createMockSolr7Server()
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))
	acc.AssertContainsTaggedFields(t, "solr_cache",
		solr7CacheExpected,
		map[string]string{"core": "main", "handler": "documentCache"})
}

func TestSolr3GatherStats(t *testing.T) {
	ts := createMockSolr3Server()
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminMainCoreStatusExpected,
		map[string]string{"core": "main"})

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminCore1StatusExpected,
		map[string]string{"core": "core1"})

	acc.AssertContainsTaggedFields(t, "solr_core",
		solr3CoreExpected,
		map[string]string{"core": "main", "handler": "searcher"})

	acc.AssertContainsTaggedFields(t, "solr_queryhandler",
		solr3QueryHandlerExpected,
		map[string]string{"core": "main", "handler": "org.apache.solr.handler.component.SearchHandler"})

	acc.AssertContainsTaggedFields(t, "solr_updatehandler",
		solr3UpdateHandlerExpected,
		map[string]string{"core": "main", "handler": "updateHandler"})

	acc.AssertContainsTaggedFields(t, "solr_cache",
		solr3CacheExpected,
		map[string]string{"core": "main", "handler": "filterCache"})
}
func TestNoCoreDataHandling(t *testing.T) {
	ts := createMockNoCoreDataServer()
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminMainCoreStatusExpected,
		map[string]string{"core": "main"})

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminCore1StatusExpected,
		map[string]string{"core": "core1"})

	acc.AssertDoesNotContainMeasurement(t, "solr_core")
	acc.AssertDoesNotContainMeasurement(t, "solr_queryhandler")
	acc.AssertDoesNotContainMeasurement(t, "solr_updatehandler")
	acc.AssertDoesNotContainMeasurement(t, "solr_handler")

}

func createMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansMainResponse)
		} else if strings.Contains(r.URL.Path, "solr/core1/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansCore1Response)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}

func createMockNoCoreDataServer() *httptest.Server {
	var nodata string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, nodata)
		} else if strings.Contains(r.URL.Path, "solr/core1/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, nodata)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}

func createMockSolr3Server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansSolr3MainResponse)
		} else if strings.Contains(r.URL.Path, "solr/core1/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansSolr3MainResponse)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}

func createMockSolr7Server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansSolr7Response)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}
