package nginx_vts

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
    "serverZones": {
        "example.com": {
            "requestCounter": 1415887,
            "inBytes": 1296356607,
            "outBytes": 4404939605,
            "responses": {
                "1xx": 100,
                "2xx": 200,
                "3xx": 300,
                "4xx": 400,
                "5xx": 500,
                "miss": 14,
                "bypass": 15,
                "expired": 16,
                "stale": 17,
                "updating": 18,
                "revalidated": 19,
                "hit": 20,
                "scarce": 21
            },
            "requestMsec": 13
        },
        "other.example.com": {
            "requestCounter": 505,
            "inBytes": 171388,
            "outBytes": 1273382,
            "responses": {
                "1xx": 101,
                "2xx": 201,
                "3xx": 301,
                "4xx": 401,
                "5xx": 501,
                "miss": 22,
                "bypass": 23,
                "expired": 24,
                "stale": 25,
                "updating": 26,
                "revalidated": 27,
                "hit": 28,
                "scarce": 29
            },
            "requestMsec": 12
        }
    },
	"filterZones": {
		"country": {
			"FI": {
				"requestCounter": 60,
				"inBytes": 2570,
				"outBytes": 53597,
				"responses": {
					"1xx": 106,
					"2xx": 206,
					"3xx": 306,
					"4xx": 406,
					"5xx": 506,
					"miss": 61,
					"bypass": 62,
					"expired": 63,
					"stale": 64,
					"updating": 65,
					"revalidated": 66,
					"hit": 67,
					"scarce": 68
				},
				"requestMsec": 69
			}
		}
	},
    "upstreamZones": {
        "backend_cluster": [
            {
                "server": "127.0.0.1:6000",
                "requestCounter": 2103849,
                "inBytes": 1774680141,
                "outBytes": 11727669190,
                "responses": {
                    "1xx": 103,
                    "2xx": 203,
                    "3xx": 303,
                    "4xx": 403,
                    "5xx": 503
                },
                "requestMsec": 30,
                "responseMsec": 31,
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
                "requestCounter": 8,
                "inBytes": 5013,
                "outBytes": 487585,
                "responses": {
                    "1xx": 104,
                    "2xx": 204,
                    "3xx": 304,
                    "4xx": 404,
                    "5xx": 504
                },
                "requestMsec": 34,
                "responseMsec": 35,
                "weight": 36,
                "maxFails": 37,
                "failTimeout": 38,
                "backup": true,
                "down": false
            },
            {
                "server": "127.0.0.1:8080",
                "requestCounter": 7,
                "inBytes": 2926,
                "outBytes": 3846638,
                "responses": {
                    "1xx": 105,
                    "2xx": 205,
                    "3xx": 305,
                    "4xx": 405,
                    "5xx": 505
                },
                "requestMsec": 39,
                "responseMsec": 40,
                "weight": 41,
                "maxFails": 42,
                "failTimeout": 43,
                "backup": true,
                "down": true
            }
        ]
    },
    "cacheZones": {
        "example": {
            "maxSize": 9223372036854776000,
            "usedSize": 68639232,
            "inBytes": 697138673,
            "outBytes": 11305044106,
            "responses": {
                "miss": 44,
                "bypass": 45,
                "expired": 46,
                "stale": 47,
                "updating": 48,
                "revalidated": 49,
                "hit": 50,
                "scarce": 51
            }
        },
        "static": {
            "maxSize": 9223372036854776000,
            "usedSize": 569856,
            "inBytes": 551652333,
            "outBytes": 1114889271,
            "responses": {
                "miss": 52,
                "bypass": 53,
                "expired": 54,
                "stale": 55,
                "updating": 56,
                "revalidated": 57,
                "hit": 58,
                "scarce": 59
            }
        }
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

	n := &NginxVTS{
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
		"nginx_vts_connections",
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
		"nginx_vts_server",
		map[string]interface{}{
			"requests":     uint64(1415887),
			"request_time": uint64(13),
			"in_bytes":     uint64(1296356607),
			"out_bytes":    uint64(4404939605),

			"response_1xx_count": uint64(100),
			"response_2xx_count": uint64(200),
			"response_3xx_count": uint64(300),
			"response_4xx_count": uint64(400),
			"response_5xx_count": uint64(500),

			"cache_miss":        uint64(14),
			"cache_bypass":      uint64(15),
			"cache_expired":     uint64(16),
			"cache_stale":       uint64(17),
			"cache_updating":    uint64(18),
			"cache_revalidated": uint64(19),
			"cache_hit":         uint64(20),
			"cache_scarce":      uint64(21),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "example.com",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_vts_filter",
		map[string]interface{}{
			"requests":     uint64(60),
			"request_time": uint64(69),
			"in_bytes":     uint64(2570),
			"out_bytes":    uint64(53597),

			"response_1xx_count": uint64(106),
			"response_2xx_count": uint64(206),
			"response_3xx_count": uint64(306),
			"response_4xx_count": uint64(406),
			"response_5xx_count": uint64(506),

			"cache_miss":        uint64(61),
			"cache_bypass":      uint64(62),
			"cache_expired":     uint64(63),
			"cache_stale":       uint64(64),
			"cache_updating":    uint64(65),
			"cache_revalidated": uint64(66),
			"cache_hit":         uint64(67),
			"cache_scarce":      uint64(68),
		},
		map[string]string{
			"source":      host,
			"port":        port,
			"filter_key":  "FI",
			"filter_name": "country",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_vts_server",
		map[string]interface{}{
			"requests":     uint64(505),
			"request_time": uint64(12),
			"in_bytes":     uint64(171388),
			"out_bytes":    uint64(1273382),

			"response_1xx_count": uint64(101),
			"response_2xx_count": uint64(201),
			"response_3xx_count": uint64(301),
			"response_4xx_count": uint64(401),
			"response_5xx_count": uint64(501),

			"cache_miss":        uint64(22),
			"cache_bypass":      uint64(23),
			"cache_expired":     uint64(24),
			"cache_stale":       uint64(25),
			"cache_updating":    uint64(26),
			"cache_revalidated": uint64(27),
			"cache_hit":         uint64(28),
			"cache_scarce":      uint64(29),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "other.example.com",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_vts_upstream",
		map[string]interface{}{
			"requests":      uint64(2103849),
			"request_time":  uint64(30),
			"response_time": uint64(31),
			"in_bytes":      uint64(1774680141),
			"out_bytes":     uint64(11727669190),

			"response_1xx_count": uint64(103),
			"response_2xx_count": uint64(203),
			"response_3xx_count": uint64(303),
			"response_4xx_count": uint64(403),
			"response_5xx_count": uint64(503),

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
		"nginx_vts_upstream",
		map[string]interface{}{
			"requests":      uint64(8),
			"request_time":  uint64(34),
			"response_time": uint64(35),
			"in_bytes":      uint64(5013),
			"out_bytes":     uint64(487585),

			"response_1xx_count": uint64(104),
			"response_2xx_count": uint64(204),
			"response_3xx_count": uint64(304),
			"response_4xx_count": uint64(404),
			"response_5xx_count": uint64(504),

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
		"nginx_vts_upstream",
		map[string]interface{}{
			"requests":      uint64(7),
			"request_time":  uint64(39),
			"response_time": uint64(40),
			"in_bytes":      uint64(2926),
			"out_bytes":     uint64(3846638),

			"response_1xx_count": uint64(105),
			"response_2xx_count": uint64(205),
			"response_3xx_count": uint64(305),
			"response_4xx_count": uint64(405),
			"response_5xx_count": uint64(505),

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

	acc.AssertContainsTaggedFields(
		t,
		"nginx_vts_cache",
		map[string]interface{}{
			"max_bytes":  uint64(9223372036854776000),
			"used_bytes": uint64(68639232),
			"in_bytes":   uint64(697138673),
			"out_bytes":  uint64(11305044106),

			"miss":        uint64(44),
			"bypass":      uint64(45),
			"expired":     uint64(46),
			"stale":       uint64(47),
			"updating":    uint64(48),
			"revalidated": uint64(49),
			"hit":         uint64(50),
			"scarce":      uint64(51),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "example",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_vts_cache",
		map[string]interface{}{
			"max_bytes":  uint64(9223372036854776000),
			"used_bytes": uint64(569856),
			"in_bytes":   uint64(551652333),
			"out_bytes":  uint64(1114889271),

			"miss":        uint64(52),
			"bypass":      uint64(53),
			"expired":     uint64(54),
			"stale":       uint64(55),
			"updating":    uint64(56),
			"revalidated": uint64(57),
			"hit":         uint64(58),
			"scarce":      uint64(59),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "static",
		})
}
