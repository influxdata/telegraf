package nginx_plus_api

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

const processesPayload = `
{
	"respawned": 0
}
`

const connectionsPayload = `
{
	"accepted": 1234567890000,
	"dropped":  2345678900000,
	"active":   345,
	"idle":     567
}
`

const sslPayload = `
{
	"handshakes": 79572,
	"handshakes_failed": 21025,
	"session_reuses": 15762
}
`

const httpRequestsPayload = `
{
	"total": 10624511,
	"current": 4
}
`

const httpServerZonesPayload = `
{
	"site1": {
		"processing": 2,
		"requests": 736395,
		"responses": {
			"1xx": 0,
			"2xx": 727290,
			"3xx": 4614,
			"4xx": 934,
			"5xx": 1535,
			"total": 734373
		},
		"discarded": 2020,
		"received": 180157219,
		"sent": 20183175459
	},
	"site2": {
		"processing": 1,
		"requests": 185307,
		"responses": {
			"1xx": 0,
			"2xx": 112674,
			"3xx": 45383,
			"4xx": 2504,
			"5xx": 4419,
			"total": 164980
		},
		"discarded": 20326,
		"received": 51575327,
		"sent": 2983241510
	}
}
`

const httpUpstreamsPayload = `
{
	"trac-backend": {
		"peers": [
		{
			"id": 0,
			"server": "10.0.0.1:8088",
			"name": "10.0.0.1:8088",
			"backup": false,
			"weight": 5,
			"state": "up",
			"active": 0,
			"requests": 667231,
			"header_time": 20,
			"response_time": 36,
			"responses": {
				"1xx": 0,
				"2xx": 666310,
				"3xx": 0,
				"4xx": 915,
				"5xx": 6,
				"total": 667231
			},
			"sent": 251946292,
			"received": 19222475454,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26214,
				"fails": 0,
				"unhealthy": 0,
				"last_passed": true
			},
			"downtime": 0,
			"downstart": {},
			"selected": {}
		},
		{
			"id": 1,
			"server": "10.0.0.1:8089",
			"name": "10.0.0.1:8089",
			"backup": true,
			"weight": 1,
			"state": "unhealthy",
			"active": 0,
			"requests": 0,
			"responses": {
				"1xx": 0,
				"2xx": 0,
				"3xx": 0,
				"4xx": 0,
				"5xx": 0,
				"total": 0
			},
			"sent": 0,
			"received": 0,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26284,
				"fails": 26284,
				"unhealthy": 1,
				"last_passed": false
			},
			"downtime": 262925617,
			"downstart": {},
			"selected": {}
		}
		],
		"keepalive": 0,
		"zombies": 0,
		"zone": "trac-backend"
	},
	"hg-backend": {
		"peers": [
		{
			"id": 0,
			"server": "10.0.0.1:8088",
			"name": "10.0.0.1:8088",
			"backup": false,
			"weight": 5,
			"state": "up",
			"active": 0,
			"requests": 667231,
			"header_time": 20,
			"response_time": 36,
			"responses": {
				"1xx": 0,
				"2xx": 666310,
				"3xx": 0,
				"4xx": 915,
				"5xx": 6,
				"total": 667231
			},
			"sent": 251946292,
			"received": 19222475454,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26214,
				"fails": 0,
				"unhealthy": 0,
				"last_passed": true
			},
			"downtime": 0,
			"downstart": {},
			"selected": {}
		},
		{
			"id": 1,
			"server": "10.0.0.1:8089",
			"name": "10.0.0.1:8089",
			"backup": true,
			"weight": 1,
			"state": "unhealthy",
			"active": 0,
			"requests": 0,
			"responses": {
				"1xx": 0,
				"2xx": 0,
				"3xx": 0,
				"4xx": 0,
				"5xx": 0,
				"total": 0
			},
			"sent": 0,
			"received": 0,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26284,
				"fails": 26284,
				"unhealthy": 1,
				"last_passed": false
			},
			"downtime": 262925617,
			"downstart": {},
			"selected": {}
		}
		],
		"keepalive": 0,
		"zombies": 0,
		"zone": "hg-backend"
	}
}
`

const httpCachesPayload = `
{
	"http-cache": {
		"size": 530915328,
		"max_size": 536870912,
		"cold": false,
		"hit": {
			"responses": 254032,
			"bytes": 6685627875
		},
		"stale": {
			"responses": 0,
			"bytes": 0
		},
		"updating": {
			"responses": 0,
			"bytes": 0
		},
		"revalidated": {
			"responses": 0,
			"bytes": 0
		},
		"miss": {
			"responses": 1619201,
			"bytes": 53841943822
		},
		"expired": {
			"responses": 45859,
			"bytes": 1656847080,
			"responses_written": 44992,
			"bytes_written": 1641825173
		},
		"bypass": {
			"responses": 200187,
			"bytes": 5510647548,
			"responses_written": 200173,
			"bytes_written": 44992
		}
	},
	"frontend-cache": {
		"size": 530915328,
		"max_size": 536870912,
		"cold": false,
		"hit": {
			"responses": 254032,
			"bytes": 6685627875
		},
		"stale": {
			"responses": 0,
			"bytes": 0
		},
		"updating": {
			"responses": 0,
			"bytes": 0
		},
		"revalidated": {
			"responses": 0,
			"bytes": 0
		},
		"miss": {
			"responses": 1619201,
			"bytes": 53841943822
		},
		"expired": {
			"responses": 45859,
			"bytes": 1656847080,
			"responses_written": 44992,
			"bytes_written": 1641825173
		},
		"bypass": {
			"responses": 200187,
			"bytes": 5510647548,
			"responses_written": 200173,
			"bytes_written": 44992
		}
	}
}
`

const streamUpstreamsPayload = `
{
	"mysql_backends": {
		"peers": [
		{
			"id": 0,
			"server": "10.0.0.1:12345",
			"name": "10.0.0.1:12345",
			"backup": false,
			"weight": 5,
			"state": "up",
			"active": 0,
			"max_conns": 30,
			"connecions": 1231,
			"sent": 251946292,
			"received": 19222475454,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26214,
				"fails": 0,
				"unhealthy": 0,
				"last_passed": true
			},
			"downtime": 0,
			"downstart": {},
			"selected": {}
		},
		{
			"id": 1,
			"server": "10.0.0.1:12346",
			"name": "10.0.0.1:12346",
			"backup": true,
			"weight": 1,
			"state": "unhealthy",
			"active": 0,
			"max_conns": 30,
			"connections": 0,
			"sent": 0,
			"received": 0,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26284,
				"fails": 26284,
				"unhealthy": 1,
				"last_passed": false
			},
			"downtime": 262925617,
			"downstart": {},
			"selected": {}
		}
		],
		"zombies": 0,
		"zone": "mysql_backends"
	},
	"dns": {
		"peers": [
		{
			"id": 0,
			"server": "10.0.0.1:12347",
			"name": "10.0.0.1:12347",
			"backup": false,
			"weight": 5,
			"state": "up",
			"active": 0,
			"max_conns": 30,
			"connections": 667231,
			"sent": 251946292,
			"received": 19222475454,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26214,
				"fails": 0,
				"unhealthy": 0,
				"last_passed": true
			},
			"downtime": 0,
			"downstart": {},
			"selected": {}
		},
		{
			"id": 1,
			"server": "10.0.0.1:12348",
			"name": "10.0.0.1:12348",
			"backup": true,
			"weight": 1,
			"state": "unhealthy",
			"active": 0,
			"connections": 0,
			"max_conns": 30,
			"sent": 0,
			"received": 0,
			"fails": 0,
			"unavail": 0,
			"health_checks": {
				"checks": 26284,
				"fails": 26284,
				"unhealthy": 1,
				"last_passed": false
			},
			"downtime": 262925617,
			"downstart": {},
			"selected": {}
		}
		],
		"zombies": 0,
		"zone": "dns"
	}
}
`

const streamServerZonesPayload = `
{
	"mysql-frontend": {
		"processing": 2,
		"connections": 270925,
		"sessions": {
			"2xx": 155564,
			"4xx": 0,
			"5xx": 0,
			"total": 270925
		},
		"discarded": 0,
		"received": 28988975,
		"sent": 3879346317
	},
	"dns": {
		"processing": 1,
		"connections": 155569,
		"sessions": {
			"2xx": 155564,
			"4xx": 0,
			"5xx": 0,
			"total": 155569
		},
		"discarded": 0,
		"received": 4200363,
		"sent": 20489184
	}
}
`

func TestGatherProcessesMetrics(t *testing.T) {
	ts, n := prepareEndpoint(processesPath, defaultApiVersion, processesPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherProcessesMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_processes",
		map[string]interface{}{
			"respawned": int(0),
		},
		map[string]string{
			"source": host,
			"port":   port,
		})
}

func TestGatherConnectioinsMetrics(t *testing.T) {
	ts, n := prepareEndpoint(connectionsPath, defaultApiVersion, connectionsPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherConnectionsMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_connections",
		map[string]interface{}{
			"accepted": int64(1234567890000),
			"dropped":  int64(2345678900000),
			"active":   int64(345),
			"idle":     int64(567),
		},
		map[string]string{
			"source": host,
			"port":   port,
		})
}

func TestGatherSslMetrics(t *testing.T) {
	ts, n := prepareEndpoint(sslPath, defaultApiVersion, sslPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherSslMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_ssl",
		map[string]interface{}{
			"handshakes":        int64(79572),
			"handshakes_failed": int64(21025),
			"session_reuses":    int64(15762),
		},
		map[string]string{
			"source": host,
			"port":   port,
		})
}

func TestGatherHttpRequestsMetrics(t *testing.T) {
	ts, n := prepareEndpoint(httpRequestsPath, defaultApiVersion, httpRequestsPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherHttpRequestsMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_requests",
		map[string]interface{}{
			"total":   int64(10624511),
			"current": int64(4),
		},
		map[string]string{
			"source": host,
			"port":   port,
		})
}

func TestGatherHttpServerZonesMetrics(t *testing.T) {
	ts, n := prepareEndpoint(httpServerZonesPath, defaultApiVersion, httpServerZonesPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherHttpServerZonesMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_server_zones",
		map[string]interface{}{
			"discarded":       int64(2020),
			"processing":      int(2),
			"received":        int64(180157219),
			"requests":        int64(736395),
			"responses_1xx":   int64(0),
			"responses_2xx":   int64(727290),
			"responses_3xx":   int64(4614),
			"responses_4xx":   int64(934),
			"responses_5xx":   int64(1535),
			"responses_total": int64(734373),
			"sent":            int64(20183175459),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "site1",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_server_zones",
		map[string]interface{}{
			"discarded":       int64(20326),
			"processing":      int(1),
			"received":        int64(51575327),
			"requests":        int64(185307),
			"responses_1xx":   int64(0),
			"responses_2xx":   int64(112674),
			"responses_3xx":   int64(45383),
			"responses_4xx":   int64(2504),
			"responses_5xx":   int64(4419),
			"responses_total": int64(164980),
			"sent":            int64(2983241510),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "site2",
		})
}

func TestHatherHttpUpstreamsMetrics(t *testing.T) {
	ts, n := prepareEndpoint(httpUpstreamsPath, defaultApiVersion, httpUpstreamsPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherHttpUpstreamsMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_upstreams",
		map[string]interface{}{
			"keepalive": int(0),
			"zombies":   int(0),
		},
		map[string]string{
			"source":   host,
			"port":     port,
			"upstream": "trac-backend",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_upstreams",
		map[string]interface{}{
			"keepalive": int(0),
			"zombies":   int(0),
		},
		map[string]string{
			"source":   host,
			"port":     port,
			"upstream": "hg-backend",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   false,
			"downtime":                 int64(0),
			"fails":                    int64(0),
			"header_time":              int64(20),
			"healthchecks_checks":      int64(26214),
			"healthchecks_fails":       int64(0),
			"healthchecks_last_passed": true,
			"healthchecks_unhealthy":   int64(0),
			"received":                 int64(19222475454),
			"requests":                 int64(667231),
			"response_time":            int64(36),
			"responses_1xx":            int64(0),
			"responses_2xx":            int64(666310),
			"responses_3xx":            int64(0),
			"responses_4xx":            int64(915),
			"responses_5xx":            int64(6),
			"responses_total":          int64(667231),
			"sent":                     int64(251946292),
			"state":                    "up",
			"unavail":                  int64(0),
			"weight":                   int(5),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "trac-backend",
			"upstream_address": "10.0.0.1:8088",
			"id":               "0",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   true,
			"downtime":                 int64(262925617),
			"fails":                    int64(0),
			"healthchecks_checks":      int64(26284),
			"healthchecks_fails":       int64(26284),
			"healthchecks_last_passed": false,
			"healthchecks_unhealthy":   int64(1),
			"received":                 int64(0),
			"requests":                 int64(0),
			"responses_1xx":            int64(0),
			"responses_2xx":            int64(0),
			"responses_3xx":            int64(0),
			"responses_4xx":            int64(0),
			"responses_5xx":            int64(0),
			"responses_total":          int64(0),
			"sent":                     int64(0),
			"state":                    "unhealthy",
			"unavail":                  int64(0),
			"weight":                   int(1),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "trac-backend",
			"upstream_address": "10.0.0.1:8089",
			"id":               "1",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   false,
			"downtime":                 int64(0),
			"fails":                    int64(0),
			"header_time":              int64(20),
			"healthchecks_checks":      int64(26214),
			"healthchecks_fails":       int64(0),
			"healthchecks_last_passed": true,
			"healthchecks_unhealthy":   int64(0),
			"received":                 int64(19222475454),
			"requests":                 int64(667231),
			"response_time":            int64(36),
			"responses_1xx":            int64(0),
			"responses_2xx":            int64(666310),
			"responses_3xx":            int64(0),
			"responses_4xx":            int64(915),
			"responses_5xx":            int64(6),
			"responses_total":          int64(667231),
			"sent":                     int64(251946292),
			"state":                    "up",
			"unavail":                  int64(0),
			"weight":                   int(5),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "hg-backend",
			"upstream_address": "10.0.0.1:8088",
			"id":               "0",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   true,
			"downtime":                 int64(262925617),
			"fails":                    int64(0),
			"healthchecks_checks":      int64(26284),
			"healthchecks_fails":       int64(26284),
			"healthchecks_last_passed": false,
			"healthchecks_unhealthy":   int64(1),
			"received":                 int64(0),
			"requests":                 int64(0),
			"responses_1xx":            int64(0),
			"responses_2xx":            int64(0),
			"responses_3xx":            int64(0),
			"responses_4xx":            int64(0),
			"responses_5xx":            int64(0),
			"responses_total":          int64(0),
			"sent":                     int64(0),
			"state":                    "unhealthy",
			"unavail":                  int64(0),
			"weight":                   int(1),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "hg-backend",
			"upstream_address": "10.0.0.1:8089",
			"id":               "1",
		})
}

func TestGatherHttpCachesMetrics(t *testing.T) {
	ts, n := prepareEndpoint(httpCachesPath, defaultApiVersion, httpCachesPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherHttpCachesMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_caches",
		map[string]interface{}{
			"bypass_bytes":              int64(5510647548),
			"bypass_bytes_written":      int64(44992),
			"bypass_responses":          int64(200187),
			"bypass_responses_written":  int64(200173),
			"cold":                      false,
			"expired_bytes":             int64(1656847080),
			"expired_bytes_written":     int64(1641825173),
			"expired_responses":         int64(45859),
			"expired_responses_written": int64(44992),
			"hit_bytes":                 int64(6685627875),
			"hit_responses":             int64(254032),
			"max_size":                  int64(536870912),
			"miss_bytes":                int64(53841943822),
			"miss_bytes_written":        int64(0),
			"miss_responses":            int64(1619201),
			"miss_responses_written":    int64(0),
			"revalidated_bytes":         int64(0),
			"revalidated_responses":     int64(0),
			"size":                      int64(530915328),
			"stale_bytes":               int64(0),
			"stale_responses":           int64(0),
			"updating_bytes":            int64(0),
			"updating_responses":        int64(0),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"cache":  "http-cache",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_http_caches",
		map[string]interface{}{
			"bypass_bytes":              int64(5510647548),
			"bypass_bytes_written":      int64(44992),
			"bypass_responses":          int64(200187),
			"bypass_responses_written":  int64(200173),
			"cold":                      false,
			"expired_bytes":             int64(1656847080),
			"expired_bytes_written":     int64(1641825173),
			"expired_responses":         int64(45859),
			"expired_responses_written": int64(44992),
			"hit_bytes":                 int64(6685627875),
			"hit_responses":             int64(254032),
			"max_size":                  int64(536870912),
			"miss_bytes":                int64(53841943822),
			"miss_bytes_written":        int64(0),
			"miss_responses":            int64(1619201),
			"miss_responses_written":    int64(0),
			"revalidated_bytes":         int64(0),
			"revalidated_responses":     int64(0),
			"size":                      int64(530915328),
			"stale_bytes":               int64(0),
			"stale_responses":           int64(0),
			"updating_bytes":            int64(0),
			"updating_responses":        int64(0),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"cache":  "frontend-cache",
		})
}

func TestGatherStreamUpstreams(t *testing.T) {
	ts, n := prepareEndpoint(streamUpstreamsPath, defaultApiVersion, streamUpstreamsPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherStreamUpstreamsMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_upstreams",
		map[string]interface{}{
			"zombies": int(0),
		},
		map[string]string{
			"source":   host,
			"port":     port,
			"upstream": "mysql_backends",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_upstreams",
		map[string]interface{}{
			"zombies": int(0),
		},
		map[string]string{
			"source":   host,
			"port":     port,
			"upstream": "dns",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   false,
			"connections":              int64(0),
			"downtime":                 int64(0),
			"fails":                    int64(0),
			"healthchecks_checks":      int64(26214),
			"healthchecks_fails":       int64(0),
			"healthchecks_last_passed": true,
			"healthchecks_unhealthy":   int64(0),
			"received":                 int64(19222475454),
			"sent":                     int64(251946292),
			"state":                    "up",
			"unavail":                  int64(0),
			"weight":                   int(5),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "mysql_backends",
			"upstream_address": "10.0.0.1:12345",
			"id":               "0",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   true,
			"connections":              int64(0),
			"downtime":                 int64(262925617),
			"fails":                    int64(0),
			"healthchecks_checks":      int64(26284),
			"healthchecks_fails":       int64(26284),
			"healthchecks_last_passed": false,
			"healthchecks_unhealthy":   int64(1),
			"received":                 int64(0),
			"sent":                     int64(0),
			"state":                    "unhealthy",
			"unavail":                  int64(0),
			"weight":                   int(1),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "mysql_backends",
			"upstream_address": "10.0.0.1:12346",
			"id":               "1",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   false,
			"connections":              int64(667231),
			"downtime":                 int64(0),
			"fails":                    int64(0),
			"healthchecks_checks":      int64(26214),
			"healthchecks_fails":       int64(0),
			"healthchecks_last_passed": true,
			"healthchecks_unhealthy":   int64(0),
			"received":                 int64(19222475454),
			"sent":                     int64(251946292),
			"state":                    "up",
			"unavail":                  int64(0),
			"weight":                   int(5),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "dns",
			"upstream_address": "10.0.0.1:12347",
			"id":               "0",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_upstream_peers",
		map[string]interface{}{
			"active":                   int(0),
			"backup":                   true,
			"connections":              int64(0),
			"downtime":                 int64(262925617),
			"fails":                    int64(0),
			"healthchecks_checks":      int64(26284),
			"healthchecks_fails":       int64(26284),
			"healthchecks_last_passed": false,
			"healthchecks_unhealthy":   int64(1),
			"received":                 int64(0),
			"sent":                     int64(0),
			"state":                    "unhealthy",
			"unavail":                  int64(0),
			"weight":                   int(1),
		},
		map[string]string{
			"source":           host,
			"port":             port,
			"upstream":         "dns",
			"upstream_address": "10.0.0.1:12348",
			"id":               "1",
		})

}

func TestGatherStreamServerZonesMatrics(t *testing.T) {
	ts, n := prepareEndpoint(streamServerZonesPath, defaultApiVersion, streamServerZonesPayload)
	defer ts.Close()

	var acc testutil.Accumulator
	addr, host, port := prepareAddr(ts)

	require.NoError(t, n.gatherStreamServerZonesMetrics(addr, &acc))

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_server_zones",
		map[string]interface{}{
			"connections": int(270925),
			"processing":  int(2),
			"received":    int64(28988975),
			"sent":        int64(3879346317),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "mysql-frontend",
		})

	acc.AssertContainsTaggedFields(
		t,
		"nginx_plus_api_stream_server_zones",
		map[string]interface{}{
			"connections": int(155569),
			"processing":  int(1),
			"received":    int64(4200363),
			"sent":        int64(20489184),
		},
		map[string]string{
			"source": host,
			"port":   port,
			"zone":   "dns",
		})
}

func prepareAddr(ts *httptest.Server) (*url.URL, string, string) {
	addr, err := url.Parse(fmt.Sprintf("%s/api", ts.URL))
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

	return addr, host, port
}

func prepareEndpoint(path string, apiVersion int64, payload string) (*httptest.Server, *NginxPlusApi) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == fmt.Sprintf("/api/%d/%s", apiVersion, path) {
			rsp = payload
			w.Header()["Content-Type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))

	n := &NginxPlusApi{
		Urls:       []string{fmt.Sprintf("%s/api", ts.URL)},
		ApiVersion: apiVersion,
	}

	client, err := n.createHttpClient()
	if err != nil {
		panic(err)
	}
	n.client = client

	return ts, n
}
