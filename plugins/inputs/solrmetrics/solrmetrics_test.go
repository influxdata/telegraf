package solrmetrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const path = "/solr/admin/metrics"

// Test an error messgae if both keys and prefix are empty
func TestEmptyParameters(t *testing.T) {
	solr := NewSolr()
	var acc testutil.Accumulator
	require.Errorf(t, solr.Gather(&acc), "'keys' and 'prefix' are empty")
}

// test a request with prefixes
func TestSolrPrefixMetrics(t *testing.T) {
	ts := createMockSolrServer(true)
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	solr.Prefixes = testPrefixes
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))
	port := solr.reqs[0].URL.Port()

	acc.AssertContainsTaggedFields(
		t,
		"core",
		solrCol01MetricsExpected,
		map[string]string{"port": port, "collection": "collection01", "shard": "shard1", "replica": "replica1"})
	acc.AssertContainsTaggedFields(
		t,
		"core",
		solrCol02MetricsExpected,
		map[string]string{"port": port, "collection": "collection02", "shard": "shard1", "replica": "replica1"})
	acc.AssertContainsTaggedFields(
		t,
		"core",
		solrCol03MetricsExpected,
		map[string]string{"port": port, "collection": "collection03", "shard": "shard1", "replica": "replica1"})
	acc.AssertContainsTaggedFields(
		t,
		"core",
		solrCol04MetricsExpected,
		map[string]string{"port": port, "collection": "collection04", "shard": "shard1", "replica": "replica1"})
}

// Test a request with keys
func TestSolrKeysMetrics(t *testing.T) {
	ts := createMockSolrServer(false)
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	solr.Keys = testKeys
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))
	port := solr.reqs[0].URL.Port()

	acc.AssertContainsTaggedFields(
		t,
		"jetty",
		solrJettyMetricsExpected,
		map[string]string{"port": port})

	acc.AssertContainsTaggedFields(
		t,
		"jvm",
		solrJVMMetricsExpected,
		map[string]string{"port": port})

	acc.AssertContainsTaggedFields(
		t,
		"node",
		solrNodeMetricsExpected,
		map[string]string{"port": port})
}

// Create http server and return response for Prefixes or Keys
func createMockSolrServer(prx bool) *httptest.Server {
	var response string
	if prx {
		response = metricsPrefixesResponse
	} else {
		response = metricsKeysResponse
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, path) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, response)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}
