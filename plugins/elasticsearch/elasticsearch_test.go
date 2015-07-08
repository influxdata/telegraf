package elasticsearch

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

const indicesResponse = `
{
  "cluster_name": "es-testcluster",
  "nodes": {
    "SDFsfSDFsdfFSDSDfSFDSDF": {
      "timestamp": 1436365550135,
      "name": "test.host.com",
      "transport_address": "inet[/127.0.0.1:9300]",
      "host": "test",
      "ip": [
        "inet[/127.0.0.1:9300]",
        "NONE"
      ],
      "attributes": {
        "master": "true"
      },
      "indices": {
        "docs": {
          "count": 29652,
          "deleted": 5229
        },
        "store": {
          "size_in_bytes": 37715234,
          "throttle_time_in_millis": 215
        },
        "indexing": {
          "index_total": 84790,
          "index_time_in_millis": 29680,
          "index_current": 0,
          "delete_total": 13879,
          "delete_time_in_millis": 1139,
          "delete_current": 0,
          "noop_update_total": 0,
          "is_throttled": false,
          "throttle_time_in_millis": 0
        },
        "get": {
          "total": 1,
          "time_in_millis": 2,
          "exists_total": 0,
          "exists_time_in_millis": 0,
          "missing_total": 1,
          "missing_time_in_millis": 2,
          "current": 0
        },
        "search": {
          "open_contexts": 0,
          "query_total": 1452,
          "query_time_in_millis": 5695,
          "query_current": 0,
          "fetch_total": 414,
          "fetch_time_in_millis": 146,
          "fetch_current": 0
        },
        "merges": {
          "current": 0,
          "current_docs": 0,
          "current_size_in_bytes": 0,
          "total": 133,
          "total_time_in_millis": 21060,
          "total_docs": 203672,
          "total_size_in_bytes": 142900226
        },
        "refresh": {
          "total": 1076,
          "total_time_in_millis": 20078
        },
        "flush": {
          "total": 115,
          "total_time_in_millis": 2401
        },
        "warmer": {
          "current": 0,
          "total": 2319,
          "total_time_in_millis": 448
        },
        "filter_cache": {
          "memory_size_in_bytes": 7384,
          "evictions": 0
        },
        "id_cache": {
          "memory_size_in_bytes": 0
        },
        "fielddata": {
          "memory_size_in_bytes": 12996,
          "evictions": 0
        },
        "percolate": {
          "total": 0,
          "time_in_millis": 0,
          "current": 0,
          "memory_size_in_bytes": -1,
          "memory_size": "-1b",
          "queries": 0
        },
        "completion": {
          "size_in_bytes": 0
        },
        "segments": {
          "count": 134,
          "memory_in_bytes": 1285212,
          "index_writer_memory_in_bytes": 0,
          "index_writer_max_memory_in_bytes": 172368955,
          "version_map_memory_in_bytes": 611844,
          "fixed_bit_set_memory_in_bytes": 0
        },
        "translog": {
          "operations": 17702,
          "size_in_bytes": 17
        },
        "suggest": {
          "total": 0,
          "time_in_millis": 0,
          "current": 0
        },
        "query_cache": {
          "memory_size_in_bytes": 0,
          "evictions": 0,
          "hit_count": 0,
          "miss_count": 0
        },
        "recovery": {
          "current_as_source": 0,
          "current_as_target": 0,
          "throttle_time_in_millis": 0
        }
      }
    }
  }
}
`

var indicesExpected = map[string]int{
	"indices_id_cache_memory_size_in_bytes":             0,
	"indices_completion_size_in_bytes":                  0,
	"indices_suggest_total":                             0,
	"indices_suggest_time_in_millis":                    0,
	"indices_suggest_current":                           0,
	"indices_query_cache_memory_size_in_bytes":          0,
	"indices_query_cache_evictions":                     0,
	"indices_query_cache_hit_count":                     0,
	"indices_query_cache_miss_count":                    0,
	"indices_store_size_in_bytes":                       37715234,
	"indices_store_throttle_time_in_millis":             215,
	"indices_merges_current_docs":                       0,
	"indices_merges_current_size_in_bytes":              0,
	"indices_merges_total":                              133,
	"indices_merges_total_time_in_millis":               21060,
	"indices_merges_total_docs":                         203672,
	"indices_merges_total_size_in_bytes":                142900226,
	"indices_merges_current":                            0,
	"indices_filter_cache_memory_size_in_bytes":         7384,
	"indices_filter_cache_evictions":                    0,
	"indices_indexing_index_total":                      84790,
	"indices_indexing_index_time_in_millis":             29680,
	"indices_indexing_index_current":                    0,
	"indices_indexing_noop_update_total":                0,
	"indices_indexing_throttle_time_in_millis":          0,
	"indices_indexing_delete_total":                     13879,
	"indices_indexing_delete_time_in_millis":            1139,
	"indices_indexing_delete_current":                   0,
	"indices_get_exists_time_in_millis":                 0,
	"indices_get_missing_total":                         1,
	"indices_get_missing_time_in_millis":                2,
	"indices_get_current":                               0,
	"indices_get_total":                                 1,
	"indices_get_time_in_millis":                        2,
	"indices_get_exists_total":                          0,
	"indices_refresh_total":                             1076,
	"indices_refresh_total_time_in_millis":              20078,
	"indices_percolate_current":                         0,
	"indices_percolate_memory_size_in_bytes":            -1,
	"indices_percolate_queries":                         0,
	"indices_percolate_total":                           0,
	"indices_percolate_time_in_millis":                  0,
	"indices_translog_operations":                       17702,
	"indices_translog_size_in_bytes":                    17,
	"indices_recovery_current_as_source":                0,
	"indices_recovery_current_as_target":                0,
	"indices_recovery_throttle_time_in_millis":          0,
	"indices_docs_count":                                29652,
	"indices_docs_deleted":                              5229,
	"indices_flush_total_time_in_millis":                2401,
	"indices_flush_total":                               115,
	"indices_fielddata_memory_size_in_bytes":            12996,
	"indices_fielddata_evictions":                       0,
	"indices_search_fetch_current":                      0,
	"indices_search_open_contexts":                      0,
	"indices_search_query_total":                        1452,
	"indices_search_query_time_in_millis":               5695,
	"indices_search_query_current":                      0,
	"indices_search_fetch_total":                        414,
	"indices_search_fetch_time_in_millis":               146,
	"indices_warmer_current":                            0,
	"indices_warmer_total":                              2319,
	"indices_warmer_total_time_in_millis":               448,
	"indices_segments_count":                            134,
	"indices_segments_memory_in_bytes":                  1285212,
	"indices_segments_index_writer_memory_in_bytes":     0,
	"indices_segments_index_writer_max_memory_in_bytes": 172368955,
	"indices_segments_version_map_memory_in_bytes":      611844,
	"indices_segments_fixed_bit_set_memory_in_bytes":    0,
}

type tranportMock struct {
	statusCode int
	body       string
}

func newTransportMock(statusCode int, body string) http.RoundTripper {
	return &tranportMock{
		statusCode: statusCode,
		body:       body,
	}
}

func (t *tranportMock) RoundTrip(r *http.Request) (*http.Response, error) {
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
	es.client.Transport = newTransportMock(http.StatusOK, indicesResponse)

	var acc testutil.Accumulator
	if err := es.Gather(&acc); err != nil {
		t.Fatal(err)
	}

	tags := map[string]string{
		"node_host":    "test",
		"cluster_name": "es-testcluster",
	}

	for key, val := range indicesExpected {
		assert.NoError(t, acc.ValidateTaggedValue(key, val, tags))
	}
}
