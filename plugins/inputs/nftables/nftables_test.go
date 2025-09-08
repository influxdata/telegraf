package nftables

import (
	"errors"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var singleton_table = `{
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
	bad_rules := []string{
		`{ "nftables": [
  {
    "rule": "I am a weird rule"
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
	bad_errors := []string{
		"Error Parsing: { \"nftables\": [\n  {\n    \"rule\": \"I am a weird rule\"\n  }\n  ]\n}, Error: Unable to parse Rule: Unable to Unmarshal: \"I am a weird rule\"",
		"Error Parsing: { \"nftables\": [\n    {\n      \"rule\": {}\n    },\n}, Error: invalid character '}' looking for beginning of value",
		"Error Parsing: { \"nftables\": [\n    {\n      \"rule\": []\n    },\n}, Error: invalid character '}' looking for beginning of value",
	}
	for i, v := range bad_rules {
		acc := new(testutil.Accumulator)
		err := parseNftableOutput([]byte(v), acc)
		if err.Error() != bad_errors[i] {
			t.Errorf("Expected Error %#v, but got %#v", bad_errors[i], err.Error())
		}
	}
}

func TestParseNftableOutput(t *testing.T) {
	acc := new(testutil.Accumulator)
	err := parseNftableOutput([]byte(singleton_table), acc)
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
	errFoo := errors.New("Error Parsing: I am not JSON, Error: invalid character 'I' looking for beginning of value")
	acc := new(testutil.Accumulator)
	err := parseNftableOutput([]byte("I am not JSON"), acc)
	if !reflect.DeepEqual(err, errFoo) {
		t.Errorf("Expected error %#v got\n%#v\n", errFoo, err)
	}
}

func TestNftableBadConfig(t *testing.T) {
	errFoo := errors.New("Invalid Configuration. Expected a `Tables` entry with list of nftables to monitor")
	ft := Nftables{}
	acc := new(testutil.Accumulator)
	err := acc.GatherError(ft.Gather)
	if !reflect.DeepEqual(err, errFoo) {
		t.Errorf("Expected error %#v got\n%#v\n", errFoo, err)
	}
}
