package couchdb_test

import (
	"github.com/influxdata/telegraf/plugins/inputs/couchdb"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasic(t *testing.T) {
	js := `
{
    "couchdb": {
        "auth_cache_misses": {
            "description": "number of authentication cache misses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "database_writes": {
            "description": "number of times a database was changed",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "open_databases": {
            "description": "number of open databases",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "auth_cache_hits": {
            "description": "number of authentication cache hits",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "request_time": {
            "description": "length of a request inside CouchDB without MochiWeb",
            "current": 18.0,
            "sum": 18.0,
            "mean": 18.0,
            "stddev": null,
            "min": 18.0,
            "max": 18.0
        },
        "database_reads": {
            "description": "number of times a document was read from a database",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "open_os_files": {
            "description": "number of file descriptors CouchDB has open",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        }
    },
    "httpd_request_methods": {
        "PUT": {
            "description": "number of HTTP PUT requests",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "GET": {
            "description": "number of HTTP GET requests",
            "current": 2.0,
            "sum": 2.0,
            "mean": 0.25,
            "stddev": 0.70699999999999996181,
            "min": 0,
            "max": 2
        },
        "COPY": {
            "description": "number of HTTP COPY requests",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "DELETE": {
            "description": "number of HTTP DELETE requests",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "POST": {
            "description": "number of HTTP POST requests",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "HEAD": {
            "description": "number of HTTP HEAD requests",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        }
    },
    "httpd_status_codes": {
        "403": {
            "description": "number of HTTP 403 Forbidden responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "202": {
            "description": "number of HTTP 202 Accepted responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "401": {
            "description": "number of HTTP 401 Unauthorized responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "409": {
            "description": "number of HTTP 409 Conflict responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "200": {
            "description": "number of HTTP 200 OK responses",
            "current": 1.0,
            "sum": 1.0,
            "mean": 0.125,
            "stddev": 0.35399999999999998135,
            "min": 0,
            "max": 1
        },
        "405": {
            "description": "number of HTTP 405 Method Not Allowed responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "400": {
            "description": "number of HTTP 400 Bad Request responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "201": {
            "description": "number of HTTP 201 Created responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "404": {
            "description": "number of HTTP 404 Not Found responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "500": {
            "description": "number of HTTP 500 Internal Server Error responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "412": {
            "description": "number of HTTP 412 Precondition Failed responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "301": {
            "description": "number of HTTP 301 Moved Permanently responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "304": {
            "description": "number of HTTP 304 Not Modified responses",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        }
    },
    "httpd": {
        "clients_requesting_changes": {
            "description": "number of clients for continuous _changes",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "temporary_view_reads": {
            "description": "number of temporary view reads",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "requests": {
            "description": "number of HTTP requests",
            "current": 2.0,
            "sum": 2.0,
            "mean": 0.25,
            "stddev": 0.70699999999999996181,
            "min": 0,
            "max": 2
        },
        "bulk_requests": {
            "description": "number of bulk requests",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        },
        "view_reads": {
            "description": "number of view reads",
            "current": null,
            "sum": null,
            "mean": null,
            "stddev": null,
            "min": null,
            "max": null
        }
    }
}

`
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_stats" {
			_, _ = w.Write([]byte(js))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &couchdb.CouchDB{
		HOSTs: []string{fakeServer.URL + "/_stats"},
	}
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Equal(t, 2, int(acc.Metrics[0].Fields["httpd_request_methods_get_max"].(float64)))
	require.True(t, math.Abs(0.707-acc.Metrics[0].Fields["httpd_request_methods_get_stddev"].(float64)) < 0.000001)
}

func TestDbStats(t *testing.T) {
	all_dbs := `
[
"db_1",
"db_2"
]

`
	db1 := `
{
"db_name": "db_1",
"update_seq": "2-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOIS40DSE08fjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9hNSdwCi7j4hdQ8g6kDuywIAikJi-A",
"sizes": {
"file": 42176,
"external": 8,
"active": 584
},
"purge_seq": 0,
"other": {
"data_size": 8
},
"doc_del_count": 0,
"doc_count": 2,
"disk_size": 42176,
"disk_format_version": 6,
"data_size": 584,
"compact_running": false,
"cluster": {
"q": 8,
"n": 1,
"w": 1,
"r": 1
},
"instance_start_time": "0"
}
`
	db2 := `
{
"db_name": "db_2",
"update_seq": "1-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOIS40DSE08fnMSQGrq8arJYwGSDA1ACqhsPiF1CyDq9hNSdwCi7j4hdQ8g6kDuywIAiXxi9w",
"sizes": {
"file": 38048,
"external": 4,
"active": 292
},
"purge_seq": 0,
"other": {
"data_size": 4
},
"doc_del_count": 0,
"doc_count": 1,
"disk_size": 38048,
"disk_format_version": 6,
"data_size": 292,
"compact_running": false,
"cluster": {
"q": 8,
"n": 1,
"w": 1,
"r": 1
},
"instance_start_time": "0"
}
`
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_all_dbs" {
			_, _ = w.Write([]byte(all_dbs))
		} else if r.URL.Path == "/db_1" {
			_, _ = w.Write([]byte(db1))
		} else if r.URL.Path == "/db_2" {
			_, _ = w.Write([]byte(db2))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &couchdb.CouchDB{
		HOSTs: []string{fakeServer.URL + "/_all_dbs"},
	}
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Equal(t, 2, int(acc.Metrics[0].Fields["doc_count"].(float64)))
	require.Equal(t, 42176, int(acc.Metrics[0].Fields["file_size"].(float64)))

	require.Equal(t, 1, int(acc.Metrics[1].Fields["doc_count"].(float64)))
	require.Equal(t, 38048, int(acc.Metrics[1].Fields["file_size"].(float64)))
}
