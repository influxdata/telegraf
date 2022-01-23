package nginx_sts

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const sampleStatusResponse = `
{
    "hostName": "test.example.com",
    "nginxVersion": "1.12.2",
    "loadMsec": 1518180328331,
    "nowMsec": 1518256058416,
    "connections": {
        "active": 111,
        "reading": 222,
        "writing": 333,
        "waiting": 444,
        "accepted": 555,
        "handled": 666,
        "requests": 777
    },
    "streamServerZones": {
        "example.com": {
            "connectCounter": 1415887,
            "inBytes": 1296356607,
            "outBytes": 4404939605,
            "responses": {
                "1xx": 100,
                "2xx": 200,
                "3xx": 300,
                "4xx": 400,
                "5xx": 500
            },
            "sessionMsecCounter": 13,
            "sessionMsec": 14
        },
        "other.example.com": {
            "connectCounter": 505,
            "inBytes": 171388,
            "outBytes": 1273382,
            "responses": {
                "1xx": 101,
                "2xx": 201,
                "3xx": 301,
                "4xx": 401,
                "5xx": 501
            },
            "sessionMsecCounter": 12,
            "sessionMsec": 15
        }
    },
	"streamFilterZones": {
		"country": {
			"FI": {
				"connectCounter": 60,
				"inBytes": 2570,
				"outBytes": 53597,
				"responses": {
					"1xx": 106,
					"2xx": 206,
					"3xx": 306,
					"4xx": 406,
					"5xx": 506
				},
                "sessionMsecCounter": 12,
                "sessionMsec": 15
			}
		}
	},
    "streamUpstreamZones": {
        "backend_cluster": [
            {
                "server": "127.0.0.1:6000",
                "connectCounter": 2103849,
                "inBytes": 1774680141,
                "outBytes": 11727669190,
                "responses": {
                    "1xx": 103,
                    "2xx": 203,
                    "3xx": 303,
                    "4xx": 403,
                    "5xx": 503
                },
                "sessionMsecCounter": 31,
                "sessionMsec": 131,
                "uSessionMsecCounter": 32,
                "uSessionMsec": 132,
                "uConnectMsecCounter": 33,
                "uConnectMsec": 130,
                "uFirstByteMsecCounter": 34,
                "uFirstByteMsec": 129,
                "weight": 32,
                "maxFails": 33,
                "failTimeout": 34,
                "backup": false,
                "down": false
			}
        ],
        "::nogroups": [
            {
                "server": "127.0.0.1:4433",
                "connectCounter": 8,
                "inBytes": 5013,
                "outBytes": 487585,
                "responses": {
                    "1xx": 104,
                    "2xx": 204,
                    "3xx": 304,
                    "4xx": 404,
                    "5xx": 504
                },
                "sessionMsecCounter": 31,
                "sessionMsec": 131,
                "uSessionMsecCounter": 32,
                "uSessionMsec": 132,
                "uConnectMsecCounter": 33,
                "uConnectMsec": 130,
                "uFirstByteMsecCounter": 34,
                "uFirstByteMsec": 129,
                "weight": 36,
                "maxFails": 37,
                "failTimeout": 38,
                "backup": true,
                "down": false
            },
            {
                "server": "127.0.0.1:8080",
                "connectCounter": 7,
                "inBytes": 2926,
                "outBytes": 3846638,
                "responses": {
                    "1xx": 105,
                    "2xx": 205,
                    "3xx": 305,
                    "4xx": 405,
                    "5xx": 505
                },
                "sessionMsecCounter": 31,
                "sessionMsec": 131,
                "uSessionMsecCounter": 32,
                "uSessionMsec": 132,
                "uConnectMsecCounter": 33,
                "uConnectMsec": 130,
                "uFirstByteMsecCounter": 34,
                "uFirstByteMsec": 129,
                "weight": 41,
                "maxFails": 42,
                "failTimeout": 43,
                "backup": true,
                "down": true
            }
        ]
    }
}
`

func TestNginxPlusGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		require.Equal(t, r.URL.Path, "/status", "Cannot handle request")

		rsp = sampleStatusResponse
		w.Header()["Content-Type"] = []string{"application/json"}

		_, err := fmt.Fprintln(w, rsp)
		require.NoError(t, err)
	}))
	defer ts.Close()

	n := &NginxSTS{
		Urls: []string{fmt.Sprintf("%s/status", ts.URL)},
	}

	var acc testutil.Accumulator

	err := n.Gather(&acc)
	require.NoError(t, err)

	addr, err := url.Parse(ts.URL)
	require.NoError(t, err)

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
		"nginx_sts_connections",
		map[string]interface{}{
			"accepted": uint64(555),
			"active":   uint64(111),
			"handled":  uint64(666),
			"reading":  uint64(222),
			"requests": uint64(777),
			"waiting":  uint64(444),
			"writing":  uint64(333),
		},
		map[string]string{
			"source": host,
			"port":   port,
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_sts_server",
		map[string]interface{}{
			"connects":             uint64(1415887),
			"in_bytes":             uint64(1296356607),
			"out_bytes":            uint64(4404939605),
			"session_msec_counter": uint64(13),
			"session_msec":         uint64(14),

			"response_1xx_count": uint64(100),
			"response_2xx_count": uint64(200),
			"response_3xx_count": uint64(300),
			"response_4xx_count": uint64(400),
			"response_5xx_count": uint64(500),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "example.com",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_sts_filter",
		map[string]interface{}{
			"connects":             uint64(60),
			"in_bytes":             uint64(2570),
			"out_bytes":            uint64(53597),
			"session_msec_counter": uint64(12),
			"session_msec":         uint64(15),

			"response_1xx_count": uint64(106),
			"response_2xx_count": uint64(206),
			"response_3xx_count": uint64(306),
			"response_4xx_count": uint64(406),
			"response_5xx_count": uint64(506),
		},
		map[string]string{
			"source":      host,
			"port":        port,
			"filter_key":  "FI",
			"filter_name": "country",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_sts_server",
		map[string]interface{}{
			"connects":             uint64(505),
			"in_bytes":             uint64(171388),
			"out_bytes":            uint64(1273382),
			"session_msec_counter": uint64(12),
			"session_msec":         uint64(15),

			"response_1xx_count": uint64(101),
			"response_2xx_count": uint64(201),
			"response_3xx_count": uint64(301),
			"response_4xx_count": uint64(401),
			"response_5xx_count": uint64(501),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "other.example.com",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_sts_upstream",
		map[string]interface{}{
			"connects":  uint64(2103849),
			"in_bytes":  uint64(1774680141),
			"out_bytes": uint64(11727669190),

			"response_1xx_count": uint64(103),
			"response_2xx_count": uint64(203),
			"response_3xx_count": uint64(303),
			"response_4xx_count": uint64(403),
			"response_5xx_count": uint64(503),

			"session_msec_counter":            uint64(31),
			"session_msec":                    uint64(131),
			"upstream_session_msec_counter":   uint64(32),
			"upstream_session_msec":           uint64(132),
			"upstream_connect_msec_counter":   uint64(33),
			"upstream_connect_msec":           uint64(130),
			"upstream_firstbyte_msec_counter": uint64(34),
			"upstream_firstbyte_msec":         uint64(129),

			"weight":       uint64(32),
			"max_fails":    uint64(33),
			"fail_timeout": uint64(34),
			"backup":       bool(false),
			"down":         bool(false),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "backend_cluster",
			"upstream_address": "127.0.0.1:6000",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_sts_upstream",
		map[string]interface{}{
			"connects":  uint64(8),
			"in_bytes":  uint64(5013),
			"out_bytes": uint64(487585),

			"response_1xx_count": uint64(104),
			"response_2xx_count": uint64(204),
			"response_3xx_count": uint64(304),
			"response_4xx_count": uint64(404),
			"response_5xx_count": uint64(504),

			"session_msec_counter":            uint64(31),
			"session_msec":                    uint64(131),
			"upstream_session_msec_counter":   uint64(32),
			"upstream_session_msec":           uint64(132),
			"upstream_connect_msec_counter":   uint64(33),
			"upstream_connect_msec":           uint64(130),
			"upstream_firstbyte_msec_counter": uint64(34),
			"upstream_firstbyte_msec":         uint64(129),

			"weight":       uint64(36),
			"max_fails":    uint64(37),
			"fail_timeout": uint64(38),
			"backup":       bool(true),
			"down":         bool(false),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "::nogroups",
			"upstream_address": "127.0.0.1:4433",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_sts_upstream",
		map[string]interface{}{
			"connects":  uint64(7),
			"in_bytes":  uint64(2926),
			"out_bytes": uint64(3846638),

			"response_1xx_count": uint64(105),
			"response_2xx_count": uint64(205),
			"response_3xx_count": uint64(305),
			"response_4xx_count": uint64(405),
			"response_5xx_count": uint64(505),

			"session_msec_counter":            uint64(31),
			"session_msec":                    uint64(131),
			"upstream_session_msec_counter":   uint64(32),
			"upstream_session_msec":           uint64(132),
			"upstream_connect_msec_counter":   uint64(33),
			"upstream_connect_msec":           uint64(130),
			"upstream_firstbyte_msec_counter": uint64(34),
			"upstream_firstbyte_msec":         uint64(129),

			"weight":       uint64(41),
			"max_fails":    uint64(42),
			"fail_timeout": uint64(43),
			"backup":       bool(true),
			"down":         bool(true),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "::nogroups",
			"upstream_address": "127.0.0.1:8080",
		})
}
