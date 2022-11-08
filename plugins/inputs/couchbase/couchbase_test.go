package couchbase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

func TestGatherServer(t *testing.T) {
	bucket := "blastro-df"
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pools" {
			_, _ = w.Write(readJSON(t, "testdata/pools_response.json"))
		} else if r.URL.Path == "/pools/default" {
			_, _ = w.Write(readJSON(t, "testdata/pools_default_response.json"))
		} else if r.URL.Path == "/pools/default/buckets" {
			_, _ = w.Write(readJSON(t, "testdata/bucket_response.json"))
		} else if r.URL.Path == "/pools/default/buckets/"+bucket+"/stats" {
			_, _ = w.Write(readJSON(t, "testdata/bucket_stats_response.json"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	cb := Couchbase{
		ClusterBucketStats:  true,
		BucketStatsIncluded: []string{"quota_percent_used", "ops_per_sec", "disk_fetches", "item_count", "disk_used", "data_used", "mem_used"},
	}
	require.NoError(t, cb.Init())

	var acc testutil.Accumulator
	require.NoError(t, cb.gatherServer(&acc, fakeServer.URL))

	acc.AssertContainsTaggedFields(t, "couchbase_node",
		map[string]interface{}{"memory_free": 23181365248.0, "memory_total": 64424656896.0},
		map[string]string{"cluster": fakeServer.URL, "hostname": "172.16.10.187:8091"})
	acc.AssertContainsTaggedFields(t, "couchbase_node",
		map[string]interface{}{"memory_free": 23665811456.0, "memory_total": 64424656896.0},
		map[string]string{"cluster": fakeServer.URL, "hostname": "172.16.10.65:8091"})
	acc.AssertContainsTaggedFields(t, "couchbase_bucket",
		map[string]interface{}{
			"quota_percent_used": 68.85424936294555,
			"ops_per_sec":        5686.789686789687,
			"disk_fetches":       0.0,
			"item_count":         943239752.0,
			"disk_used":          409178772321.0,
			"data_used":          212179309111.0,
			"mem_used":           202156957464.0,
		},
		map[string]string{"cluster": fakeServer.URL, "bucket": "blastro-df"})
}

func TestSanitizeURI(t *testing.T) {
	var sanitizeTest = []struct {
		input    string
		expected string
	}{
		{"http://user:password@localhost:121", "http://localhost:121"},
		{"user:password@localhost:12/endpoint", "localhost:12/endpoint"},
		{"https://mail@address.com:password@localhost", "https://localhost"},
		{"localhost", "localhost"},
		{"user:password@localhost:2321", "localhost:2321"},
		{"http://user:password@couchbase-0.example.com:8091/endpoint", "http://couchbase-0.example.com:8091/endpoint"},
		{" ", " "},
	}

	for _, test := range sanitizeTest {
		result := regexpURI.ReplaceAllString(test.input, "${1}")

		if result != test.expected {
			t.Errorf("TestSanitizeAddress: input %s, expected %s, actual %s", test.input, test.expected, result)
		}
	}
}

func TestGatherDetailedBucketMetrics(t *testing.T) {
	bucket := "Ducks"
	node := "172.94.77.2:8091"

	bucketStatsResponse := readJSON(t, "testdata/bucket_stats_response.json")
	bucketStatsResponseWithMissing := readJSON(t, "testdata/bucket_stats_response_with_missing.json")
	nodeBucketStatsResponse := readJSON(t, "testdata/node_bucket_stats_response.json")

	tests := []struct {
		name     string
		node     *string
		response []byte
	}{
		{
			name:     "cluster-level with all fields",
			response: bucketStatsResponse,
		},
		{
			name:     "cluster-level with missing fields",
			response: bucketStatsResponseWithMissing,
		},
		{
			name:     "node-level with all fields",
			response: nodeBucketStatsResponse,
			node:     &node,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/pools/default/buckets/"+bucket+"/stats" || r.URL.Path == "/pools/default/buckets/"+bucket+"/nodes/"+node+"/stats" {
					_, _ = w.Write(test.response)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))

			var err error
			var cb Couchbase
			cb.BucketStatsIncluded = []string{"couch_total_disk_size"}
			cb.ClientConfig = tls.ClientConfig{
				InsecureSkipVerify: true,
			}
			err = cb.Init()
			require.NoError(t, err)
			var acc testutil.Accumulator
			bucketStats := &BucketStats{}
			if err := json.Unmarshal(test.response, bucketStats); err != nil {
				t.Fatal("parse bucketResponse", err)
			}

			fields := make(map[string]interface{})
			err = cb.gatherDetailedBucketStats(fakeServer.URL, bucket, test.node, fields)
			require.NoError(t, err)

			acc.AddFields("couchbase_bucket", fields, nil)

			// Ensure we gathered only one metric (the one that we configured).
			require.Equal(t, len(acc.Metrics), 1)
			require.Equal(t, len(acc.Metrics[0].Fields), 1)
		})
	}
}

func TestGatherNodeOnly(t *testing.T) {
	faker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pools" {
			_, _ = w.Write(readJSON(t, "testdata/pools_response.json"))
		} else if r.URL.Path == "/pools/default" {
			_, _ = w.Write(readJSON(t, "testdata/pools_default_response.json"))
		} else if r.URL.Path == "/pools/default/buckets" {
			_, _ = w.Write(readJSON(t, "testdata/bucket_response.json"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	cb := Couchbase{
		Servers: []string{faker.URL},
	}
	require.NoError(t, cb.Init())

	var acc testutil.Accumulator
	require.NoError(t, cb.gatherServer(&acc, faker.URL))

	require.Equal(t, 0, len(acc.Errors))
	require.Equal(t, 7, len(acc.Metrics))
	acc.AssertDoesNotContainMeasurement(t, "couchbase_bucket")
}

func readJSON(t *testing.T, jsonFilePath string) []byte {
	data, err := os.ReadFile(jsonFilePath)
	require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)

	return data
}
