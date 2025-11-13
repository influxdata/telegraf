//go:build linux

package nftables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var singletonTable = `{
  "nftables": [
    {
      "metainfo": {
        "version": "1.0.2",
        "release_name": "Lester Gooch",
        "json_schema_version": 1
      }
    },
    {
      "table": {
        "family": "inet",
        "name": "test",
        "handle": 5
      }
    },
    {
      "chain": {
        "family": "inet",
        "table": "test",
        "name": "test-chain",
        "handle": 1,
        "type": "filter",
        "hook": "input",
        "prio": 0,
        "policy": "accept"
      } }, {
      "rule": {
        "family": "inet",
        "table": "test",
        "chain": "no_counter",
        "handle": 2,
        "comment": "test1",
        "expr": [
          {
            "match": {
              "op": "==",
              "left": {
                "payload": {
                  "protocol": "tcp",
                  "field": "dport"
                }
              },
              "right": 22
            }
          },
          {
            "accept": null
          }
        ]
      }
    },
    {
      "rule": {
        "family": "inet",
        "table": "test",
        "chain": "test-chain",
        "handle": 2,
        "comment": "test1",
        "expr": [
          {
            "match": {
              "op": "==",
              "left": {
                "payload": {
                  "protocol": "tcp",
                  "field": "dport"
                }
              },
              "right": 22
            }
          },
          {
            "counter": {
              "packets": 1,
              "bytes": 2
            }
          },
          {
            "accept": null
          }
        ]
      }
    },
    {
      "rule": {
        "family": "inet",
        "table": "test",
        "chain": "no_comment",
        "handle": 2,
        "expr": [
          {
            "match": {
              "op": "==",
              "left": {
                "payload": {
                  "protocol": "tcp",
                  "field": "dport"
                }
              },
              "right": 22
            }
          },
          {
            "accept": null
          }
        ]
      }
    },
    {
      "rule": {
        "family": "inet",
        "table": "test",
        "chain": "test-chain",
        "handle": 2,
        "comment": "test2",
        "expr": [
          {
            "match": {
              "op": "==",
              "left": {
                "payload": {
                  "protocol": "tcp",
                  "field": "dport"
                }
              },
              "right": 22
            }
          },
          {
            "counter": {
              "packets": 24468,
              "bytes": 1412296
            }
          },
          {
            "accept": null
          }
        ]
      }
    }
  ]
}`

func TestParseNftableBadRule(t *testing.T) {
	badrules := []string{
		`{ "nftables": [
    {
      "rule": "bad"
    }
  ]
}`,
		`{ "nftables": [
    {
      "rule": {}
    },
}`,
		`{ "nftables": [
    {
      "rule": []
    },
}`}
	for _, v := range badrules {
		var acc testutil.Accumulator
		require.Error(t, parseNftableOutput(&acc, []byte(v)))
	}
}

func TestParseNftableOutput(t *testing.T) {
	var acc testutil.Accumulator
	require.NoError(t, parseNftableOutput(&acc, []byte(singletonTable)))
	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 2)
	defaultTime := time.Unix(0, 0)
	expected := []telegraf.Metric{
		testutil.MustMetric("nftables",
			map[string]string{
				"chain": "test-chain",
				"rule":  "test1",
				"table": "test",
			},
			map[string]interface{}{
				"bytes": 2,
				"pkts":  1,
			}, defaultTime),
		testutil.MustMetric("nftables",
			map[string]string{
				"chain": "test-chain",
				"rule":  "test2",
				"table": "test",
			},
			map[string]interface{}{
				"bytes": 1412296,
				"pkts":  24468,
			}, defaultTime),
	}
	for i, v := range metrics {
		testutil.RequireMetricEqual(t, expected[i], v, testutil.IgnoreTime())
	}
}

func TestParseNftableBadOutput(t *testing.T) {
	var acc testutil.Accumulator
	require.Error(t, parseNftableOutput(&acc, []byte("I am not JSON")))
}
