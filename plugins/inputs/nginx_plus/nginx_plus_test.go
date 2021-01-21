package nginx_plus

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleStatusResponse = `
{
    "version": 6,
    "nginx_version":  "1.22.333",
    "address":        "1.2.3.4",
    "generation":     88,
    "load_timestamp": 1451606400000,
    "timestamp":      1451606400000,
    "pid":            9999,
    "processes": {
        "respawned": 9999
     },
    "connections": {
        "accepted": 1234567890000,
        "dropped":  2345678900000,
        "active":   345,
        "idle":     567
    },
    "ssl": {
        "handshakes":        1234567800000,
        "handshakes_failed": 5432100000000,
        "session_reuses":    6543210000000
    },
    "requests": {
        "total":   9876543210000,
        "current": 98
    },
    "server_zones": {
        "zone.a_80": {
            "processing": 12,
            "requests": 34,
            "responses": {
                "1xx": 111,
                "2xx": 222,
                "3xx": 333,
                "4xx": 444,
                "5xx": 555,
                "total": 999
            },
            "discarded": 11,
            "received": 22,
            "sent": 33
        },
        "zone.a_443": {
            "processing": 45,
            "requests": 67,
            "responses": {
                "1xx": 1111,
                "2xx": 2222,
                "3xx": 3333,
                "4xx": 4444,
                "5xx": 5555,
                "total": 999
            },
            "discarded": 44,
            "received": 55,
            "sent": 66
        }
    },
    "upstreams": {
        "first_upstream": {
            "peers": [
                {
                    "id": 0,
                    "server": "1.2.3.123:80",
                    "backup": false,
                    "weight": 1,
                    "state": "up",
                    "active": 0,
                    "requests": 9876,
                    "responses": {
                        "1xx": 1111,
                        "2xx": 2222,
                        "3xx": 3333,
                        "4xx": 4444,
                        "5xx": 5555,
                        "total": 987654
                    },
                    "sent": 987654321,
                    "received": 87654321,
                    "fails": 98,
                    "unavail": 65,
                    "health_checks": {
                        "checks": 54,
                        "fails": 32,
                        "unhealthy": 21
                    },
                    "downtime": 5432,
                    "downstart": 4321,
                    "selected": 1451606400000
                },
                {
                    "id": 1,
                    "server": "1.2.3.123:80",
                    "backup": true,
                    "weight": 1,
                    "state": "up",
                    "active": 0,
                    "requests": 8765,
                    "responses": {
                        "1xx": 1112,
                        "2xx": 2223,
                        "3xx": 3334,
                        "4xx": 4445,
                        "5xx": 5556,
                        "total": 987655
                    },
                    "sent": 987654322,
                    "received": 87654322,
                    "fails": 99,
                    "unavail": 88,
                    "health_checks": {
                        "checks": 77,
                        "fails": 66,
                        "unhealthy": 55
                    },
                    "downtime": 5433,
                    "downstart": 4322,
                    "selected": 1451606400000
                }
            ],
            "keepalive": 1,
            "zombies": 2
        }
    },
    "caches": {
        "cache_01": {
            "size": 12,
            "max_size": 23,
            "cold": false,
            "hit": {
                "responses": 34,
                "bytes": 45
            },
            "stale": {
                "responses": 56,
                "bytes": 67
            },
            "updating": {
                "responses": 78,
                "bytes": 89
            },
            "revalidated": {
                "responses": 90,
                "bytes": 98
            },
            "miss": {
                "responses": 87,
                "bytes": 76,
                "responses_written": 65,
                "bytes_written": 54
            },
            "expired": {
                "responses": 43,
                "bytes": 32,
                "responses_written": 21,
                "bytes_written": 10
            },
            "bypass": {
                "responses": 13,
                "bytes": 35,
                "responses_written": 57,
                "bytes_written": 79
            }
        }
    },
    "stream": {
        "server_zones": {
            "stream.zone.01": {
                "processing": 24,
                "connections": 46,
                "received": 68,
                "sent": 80
            },
            "stream.zone.02": {
                "processing": 96,
                "connections": 63,
                "received": 31,
                "sent": 25
            }
        },
        "upstreams": {
            "upstream.01": {
                "peers": [
                    {
                        "id": 0,
                        "server": "4.3.2.1:2345",
                        "backup": false,
                        "weight": 1,
                        "state": "up",
                        "active": 0,
                        "connections": 0,
                        "sent": 0,
                        "received": 0,
                        "fails": 0,
                        "unavail": 0,
                        "health_checks": {
                            "checks": 40848,
                            "fails": 0,
                            "unhealthy": 0,
                            "last_passed": true
                        },
                        "downtime": 0,
                        "downstart": 0,
                        "selected": 0
                    },
                    {
                        "id": 1,
                        "server": "5.4.3.2:2345",
                        "backup": false,
                        "weight": 1,
                        "state": "up",
                        "active": 0,
                        "connections": 0,
                        "sent": 0,
                        "received": 0,
                        "fails": 0,
                        "unavail": 0,
                        "health_checks": {
                            "checks": 40851,
                            "fails": 0,
                            "unhealthy": 0,
                            "last_passed": true
                        },
                        "downtime": 0,
                        "downstart": 0,
                        "selected": 0
                    }
                ],
                "zombies": 0
            }
        }
    }
}
`

func TestNginxPlusGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/status" {
			rsp = sampleStatusResponse
			w.Header()["Content-Type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &NginxPlus{
		Urls: []string{fmt.Sprintf("%s/status", ts.URL)},
	}

	var acc testutil.Accumulator

	err_nginx := n.Gather(&acc)

	require.NoError(t, err_nginx)

	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	host, port, err := net.SplitHostPort(addr.Host)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_processes",
		map[string]interface{}{
			"respawned": int(9999),
		},
		map[string]string{
			"server": host,
			"port":   port,
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_connections",
		map[string]interface{}{
			"accepted": int64(1234567890000),
			"dropped":  int64(2345678900000),
			"active":   int64(345),
			"idle":     int64(567),
		},
		map[string]string{
			"server": host,
			"port":   port,
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_ssl",
		map[string]interface{}{
			"handshakes":        int64(1234567800000),
			"handshakes_failed": int64(5432100000000),
			"session_reuses":    int64(6543210000000),
		},
		map[string]string{
			"server": host,
			"port":   port,
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_requests",
		map[string]interface{}{
			"total":   int64(9876543210000),
			"current": int(98),
		},
		map[string]string{
			"server": host,
			"port":   port,
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_zone",
		map[string]interface{}{
			"processing":      int(12),
			"requests":        int64(34),
			"responses_1xx":   int64(111),
			"responses_2xx":   int64(222),
			"responses_3xx":   int64(333),
			"responses_4xx":   int64(444),
			"responses_5xx":   int64(555),
			"responses_total": int64(999),
			"discarded":       int64(11),
			"received":        int64(22),
			"sent":            int64(33),
		},
		map[string]string{
			"server": host,
			"port":   port,
			"zone":   "zone.a_80",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_upstream",
		map[string]interface{}{
			"keepalive": int(1),
			"zombies":   int(2),
		},
		map[string]string{
			"server":   host,
			"port":     port,
			"upstream": "first_upstream",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_upstream_peer",
		map[string]interface{}{
			"backup":                 false,
			"weight":                 int(1),
			"state":                  "up",
			"active":                 int(0),
			"requests":               int64(9876),
			"responses_1xx":          int64(1111),
			"responses_2xx":          int64(2222),
			"responses_3xx":          int64(3333),
			"responses_4xx":          int64(4444),
			"responses_5xx":          int64(5555),
			"responses_total":        int64(987654),
			"sent":                   int64(987654321),
			"received":               int64(87654321),
			"fails":                  int64(98),
			"unavail":                int64(65),
			"healthchecks_checks":    int64(54),
			"healthchecks_fails":     int64(32),
			"healthchecks_unhealthy": int64(21),
			"downtime":               int64(5432),
			"downstart":              int64(4321),
			"selected":               int64(1451606400000),
		},
		map[string]string{
			"server":           host,
			"port":             port,
			"upstream":         "first_upstream",
			"upstream_address": "1.2.3.123:80",
			"id":               "0",
		})

}
