//go:build linux

package nftables

import (
	"testing"

	"github.com/stretchr/testify/require"

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
		acc := new(testutil.Accumulator)
		require.Error(t, parseNftableOutput([]byte(v), acc))
	}
}

func TestParseNftableOutput(t *testing.T) {
	acc := new(testutil.Accumulator)
	err := parseNftableOutput([]byte(singletonTable), acc)
	if err != nil {
		t.Errorf("No Error Expected: %#v", err)
	}
	metrics := acc.Metrics
	if len(metrics) != 2 {
		t.Errorf("Expected 2 measurments. Got: %#v", len(metrics))
	}
	expected := []string{
		"nftables map[chain:test-chain ruleid:test1 table:test] map[bytes:2 pkts:1]",
		"nftables map[chain:test-chain ruleid:test2 table:test] map[bytes:1412296 pkts:24468]",
	}
	for i, v := range metrics {
		if v.String() != expected[i] {
			t.Errorf("Expected measurments to be equal. Expected: %#v, Got: %#v", expected[i], v)
		}
	}
}

func TestParseNftableBadOutput(t *testing.T) {
	acc := new(testutil.Accumulator)
	require.Error(t, parseNftableOutput([]byte("I am not JSON"), acc))
}

func TestNftableBadConfig(t *testing.T) {
	plugin := Nftables{}
	require.NoError(t, plugin.Init())
	acc := new(testutil.Accumulator)
	require.Error(t, acc.GatherError(plugin.Gather))
}
