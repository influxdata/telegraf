package clickhouse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClusterIncludeExcludeFilter(t *testing.T) {
	ch := ClickHouse{}
	if assert.Equal(t, "", ch.clusterIncludeExcludeFilter()) {
		ch.ClusterExclude = []string{"test_cluster"}
		assert.Equal(t, "WHERE cluster NOT IN ('test_cluster')", ch.clusterIncludeExcludeFilter())

		ch.ClusterExclude = []string{"test_cluster"}
		ch.ClusterInclude = []string{"cluster"}
		assert.Equal(t, "WHERE cluster IN ('cluster') OR cluster NOT IN ('test_cluster')", ch.clusterIncludeExcludeFilter())

		ch.ClusterExclude = []string{}
		ch.ClusterInclude = []string{"cluster1", "cluster2"}
		assert.Equal(t, "WHERE cluster IN ('cluster1', 'cluster2')", ch.clusterIncludeExcludeFilter())

		ch.ClusterExclude = []string{"cluster1", "cluster2"}
		ch.ClusterInclude = []string{}
		assert.Equal(t, "WHERE cluster NOT IN ('cluster1', 'cluster2')", ch.clusterIncludeExcludeFilter())
	}
}

func TestChInt64(t *testing.T) {
	assets := map[string]uint64{
		`"1"`:                  1,
		"1":                    1,
		"42":                   42,
		`"42"`:                 42,
		"18446743937525109187": 18446743937525109187,
	}
	for src, expected := range assets {
		var v chUInt64
		if err := v.UnmarshalJSON([]byte(src)); assert.NoError(t, err) {
			assert.Equal(t, expected, uint64(v))
		}
	}
}

func TestGather(t *testing.T) {
	var (
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			type result struct {
				Data interface{} `json:"data"`
			}
			enc := json.NewEncoder(w)
			switch query := r.URL.Query().Get("query"); {
			case strings.Contains(query, "system.parts"):
				enc.Encode(result{
					Data: []struct {
						Database string   `json:"database"`
						Table    string   `json:"table"`
						Bytes    chUInt64 `json:"bytes"`
						Parts    chUInt64 `json:"parts"`
						Rows     chUInt64 `json:"rows"`
					}{
						{
							Database: "test_database",
							Table:    "test_table",
							Bytes:    1,
							Parts:    10,
							Rows:     100,
						},
					},
				})
			case strings.Contains(query, "system.events"):
				enc.Encode(result{
					Data: []struct {
						Metric string   `json:"metric"`
						Value  chUInt64 `json:"value"`
					}{
						{
							Metric: "TestSystemEvent",
							Value:  1000,
						},
						{
							Metric: "TestSystemEvent2",
							Value:  2000,
						},
					},
				})
			case strings.Contains(query, "system.metrics"):
				enc.Encode(result{
					Data: []struct {
						Metric string   `json:"metric"`
						Value  chUInt64 `json:"value"`
					}{
						{
							Metric: "TestSystemMetric",
							Value:  1000,
						},
						{
							Metric: "TestSystemMetric2",
							Value:  2000,
						},
					},
				})
			case strings.Contains(query, "system.asynchronous_metrics"):
				enc.Encode(result{
					Data: []struct {
						Metric string   `json:"metric"`
						Value  chUInt64 `json:"value"`
					}{
						{
							Metric: "TestSystemAsynchronousMetric",
							Value:  1000,
						},
						{
							Metric: "TestSystemAsynchronousMetric2",
							Value:  2000,
						},
					},
				})
			}
		}))
		ch = &ClickHouse{
			Servers: []string{
				ts.URL,
			},
		}
		acc = &testutil.Accumulator{}
	)
	defer ts.Close()
	ch.Gather(acc)

	acc.AssertContainsFields(t, "clickhouse_tables",
		map[string]interface{}{
			"bytes": uint64(1),
			"parts": uint64(10),
			"rows":  uint64(100),
		},
	)
	acc.AssertContainsFields(t, "clickhouse_events",
		map[string]interface{}{
			"test_system_event":  uint64(1000),
			"test_system_event2": uint64(2000),
		},
	)
	acc.AssertContainsFields(t, "clickhouse_metrics",
		map[string]interface{}{
			"test_system_metric":  uint64(1000),
			"test_system_metric2": uint64(2000),
		},
	)
	acc.AssertContainsFields(t, "clickhouse_asynchronous_metrics",
		map[string]interface{}{
			"test_system_asynchronous_metric":  uint64(1000),
			"test_system_asynchronous_metric2": uint64(2000),
		},
	)
}
