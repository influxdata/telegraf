package couchdb_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/couchdb"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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
		Hosts: []string{fakeServer.URL + "/_stats"},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
}
