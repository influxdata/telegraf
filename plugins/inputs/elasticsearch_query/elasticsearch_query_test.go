package elasticsearch_query

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"
	esapi5 "github.com/elastic/go-elasticsearch/v5/esapi"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/testutil"
)

const (
	servicePort = "9200"
	testindex   = "test-elasticsearch"
)

type nginxlog struct {
	IPaddress    string    `json:"IP"`
	Timestamp    time.Time `json:"@timestamp"`
	Method       string    `json:"method"`
	URI          string    `json:"URI"`
	Httpversion  string    `json:"http_version"`
	Response     string    `json:"response"`
	Size         float64   `json:"size"`
	ResponseTime float64   `json:"response_time"`
}

func TestGatherIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define expectations
	expectedFields := []map[string]string{
		{"size": "long"},
		{"size": "long"},
		{"size": "long"},
		{"size": "long", "response_time": "long"},
		{},
		{},
		{},
		{"size": "long"},
		{"size": "long"},
		{"size": "long"},
	}

	expectedMetrics := []telegraf.Metric{
		metric.New(
			"measurement1",
			map[string]string{"URI_keyword": "/downloads/product_1"},
			map[string]interface{}{"size_avg": float64(202.30038022813687), "doc_count": int64(263)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement2",
			map[string]string{"URI_keyword": "/downloads/product_1"},
			map[string]interface{}{"size_max": float64(3301), "doc_count": int64(263)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement2",
			map[string]string{"URI_keyword": "/downloads/product_2"},
			map[string]interface{}{"size_max": float64(3318), "doc_count": int64(237)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement3",
			map[string]string{"response_keyword": "200"},
			map[string]interface{}{"size_sum": float64(22790), "doc_count": int64(22)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement3",
			map[string]string{"response_keyword": "304"},
			map[string]interface{}{"size_sum": float64(0), "doc_count": int64(219)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement3",
			map[string]string{"response_keyword": "404"},
			map[string]interface{}{"size_sum": float64(86932), "doc_count": int64(259)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "404", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
			map[string]interface{}{"size_min": float64(318), "response_time_min": float64(126), "doc_count": int64(146)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "304", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
			map[string]interface{}{"size_min": float64(0), "response_time_min": float64(71), "doc_count": int64(113)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
			map[string]interface{}{"size_min": float64(490), "response_time_min": float64(1514), "doc_count": int64(3)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "404", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
			map[string]interface{}{"size_min": float64(318), "response_time_min": float64(237), "doc_count": int64(113)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "304", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
			map[string]interface{}{"size_min": float64(0), "response_time_min": float64(134), "doc_count": int64(106)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
			map[string]interface{}{"size_min": float64(490), "response_time_min": float64(2), "doc_count": int64(13)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1", "method_keyword": "HEAD"},
			map[string]interface{}{"size_min": float64(0), "response_time_min": float64(8479), "doc_count": int64(1)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement4",
			map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2", "method_keyword": "HEAD"},
			map[string]interface{}{"size_min": float64(0), "response_time_min": float64(1059), "doc_count": int64(5)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement5",
			map[string]string{"URI_keyword": "/downloads/product_2"},
			map[string]interface{}{"doc_count": int64(237)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement6",
			map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1"},
			map[string]interface{}{"doc_count": int64(4)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement6",
			map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2"},
			map[string]interface{}{"doc_count": int64(18)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement7",
			map[string]string{},
			map[string]interface{}{"doc_count": int64(22)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement8",
			map[string]string{},
			map[string]interface{}{"size_max": float64(3318)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
		metric.New(
			"measurement12",
			map[string]string{},
			map[string]interface{}{"size_avg": float64(0)},
			time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
		),
	}

	// Setup the container
	container := &testutil.Container{
		Image:        "elasticsearch:6.8.23",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type": "single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("] mode [basic] - valid"),
			wait.ForListeningPort(servicePort),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := "http://" + container.Address + ":" + container.Ports[servicePort]

	// Fill the database
	require.NoError(t, sendData(t.Context(), addr))

	// Setup the plugin
	plugin := &ElasticsearchQuery{
		URLs: []string{addr},
		Aggregations: []aggregation{
			{
				Index:           testindex,
				MeasurementName: "measurement1",
				MetricFields:    []string{"size"},
				FilterQuery:     "product_1",
				MetricFunction:  "avg",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
				Tags:            []string{"URI.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement2",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "max",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
				Tags:            []string{"URI.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement3",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "sum",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
				Tags:            []string{"response.keyword"},
			},
			{
				Index:             testindex,
				MeasurementName:   "measurement4",
				MetricFields:      []string{"size", "response_time"},
				FilterQuery:       "downloads",
				MetricFunction:    "min",
				DateField:         "@timestamp",
				QueryPeriod:       config.Duration(time.Second * 600),
				IncludeMissingTag: true,
				MissingTagValue:   "missing",
				Tags:              []string{"response.keyword", "URI.keyword", "method.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement5",
				FilterQuery:     "product_2",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
				Tags:            []string{"URI.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement6",
				FilterQuery:     "response: 200",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
				Tags:            []string{"URI.keyword", "response.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement7",
				FilterQuery:     "response: 200",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			{
				Index:           testindex,
				MeasurementName: "measurement8",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "max",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			{
				Index:           testindex,
				MeasurementName: "measurement12",
				MetricFields:    []string{"size"},
				MetricFunction:  "avg",
				DateField:       "@notatimestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			{
				Index:             testindex,
				MeasurementName:   "measurement13",
				MetricFields:      []string{"size"},
				MetricFunction:    "avg",
				DateField:         "@timestamp",
				QueryPeriod:       config.Duration(time.Second * 600),
				IncludeMissingTag: false,
				Tags:              []string{"nothere"},
			},
		},
		HTTPClientConfig: common_http.HTTPClientConfig{
			Timeout: config.Duration(30 * time.Second),
			TransportConfig: common_http.TransportConfig{
				ResponseHeaderTimeout: config.Duration(30 * time.Second),
			},
		},
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Check the ES field mapping
	for i, agg := range plugin.Aggregations {
		actual := agg.mapMetricFields
		expected := expectedFields[i]
		require.Equalf(t, expected, actual, "mismatch in aggregation %d", i)
	}

	// Collect metrics and check
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Empty(t, acc.Errors)

	// Check the metrics
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.SortMetrics(), testutil.IgnoreTime())
}

func TestGatherFailStartIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup the container
	container := &testutil.Container{
		Image:        "elasticsearch:6.8.23",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type": "single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("] mode [basic] - valid"),
			wait.ForListeningPort(servicePort),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := "http://" + container.Address + ":" + container.Ports[servicePort]

	// Fill the database
	require.NoError(t, sendData(t.Context(), addr))

	tests := []struct {
		name     string
		agg      aggregation
		expected string
	}{
		{
			name: "invalid function",
			agg: aggregation{
				Index:           testindex,
				MeasurementName: "measurement9",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "average",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			expected: `aggregation function "average" not supported`,
		},
		{
			name: "non-existing field",
			agg: aggregation{
				Index:           testindex,
				MeasurementName: "measurement10",
				MetricFields:    []string{"none"},
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			expected: `metric field "none" not found on index`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &ElasticsearchQuery{
				URLs:         []string{addr},
				Aggregations: []aggregation{tt.agg},
				HTTPClientConfig: common_http.HTTPClientConfig{
					Timeout: config.Duration(30 * time.Second),
					TransportConfig: common_http.TransportConfig{
						ResponseHeaderTimeout: config.Duration(30 * time.Second),
					},
				},
				Log: testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.ErrorContains(t, plugin.Start(&acc), tt.expected)
			defer plugin.Stop()
		})
	}
}

func TestGatherFailGatherIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup the container
	container := &testutil.Container{
		Image:        "elasticsearch:6.8.23",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type": "single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("] mode [basic] - valid"),
			wait.ForListeningPort(servicePort),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := "http://" + container.Address + ":" + container.Ports[servicePort]

	// Fill the database
	require.NoError(t, sendData(t.Context(), addr))

	tests := []struct {
		name     string
		agg      aggregation
		expected string
	}{
		{
			name: "invalid index",
			agg: aggregation{
				Index:           "notanindex",
				MeasurementName: "measurement11",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			expected: "Error 404 (Not Found): no such index",
		},
		{
			name: "invalid time format",
			agg: aggregation{
				Index:           testindex,
				MeasurementName: "measurement14",
				DateField:       "@timestamp",
				DateFieldFormat: "yyyy",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			expected: "Error 400 (Bad Request): all shards failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &ElasticsearchQuery{
				URLs:         []string{addr},
				Aggregations: []aggregation{tt.agg},
				HTTPClientConfig: common_http.HTTPClientConfig{
					Timeout: config.Duration(30 * time.Second),
					TransportConfig: common_http.TransportConfig{
						ResponseHeaderTimeout: config.Duration(30 * time.Second),
					},
				},
				Log: testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Gather data and check error
			require.ErrorContains(t, acc.GatherError(plugin.Gather), tt.expected)
		})
	}
}

func TestStartupFailureReleasesClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"version": {"number": "8.1.2"}}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	plugin := &ElasticsearchQuery{
		URLs:                []string{server.URL},
		HealthCheckInterval: config.Duration(10 * time.Second),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.ErrorContains(t, plugin.Start(&acc), "not supported")

	// The failed start must release the client to not leak the
	// health-check goroutine
	require.Nil(t, plugin.client.(*clientV5).client)
}

func sendData(ctx context.Context, url string) error {
	// Read the data
	file, err := os.Open(filepath.Join("testdata", "nginx_logs"))
	if err != nil {
		return fmt.Errorf("reading nginx logs failed: %w", err)
	}
	defer file.Close()

	var logs []nginxlog
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		size, err := strconv.Atoi(parts[9])
		if err != nil {
			return fmt.Errorf("parsing size failed: %w", err)
		}
		responseTime, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			return fmt.Errorf("parsing response-time failed: %w", err)
		}

		logs = append(logs, nginxlog{
			IPaddress:    parts[0],
			Timestamp:    time.Now().UTC(),
			Method:       strings.ReplaceAll(parts[5], `"`, ""),
			URI:          parts[6],
			Httpversion:  strings.ReplaceAll(parts[7], `"`, ""),
			Response:     parts[8],
			Size:         float64(size),
			ResponseTime: float64(responseTime),
		})
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning nginx logs failed: %w", err)
	}

	// Create the client
	client, err := newTestIndexer(ctx, url)
	if err != nil {
		return fmt.Errorf("creating client failed: %w", err)
	}

	// Create bulk request for the data
	if err := client.bulkIndex(ctx, testindex, logs); err != nil {
		return fmt.Errorf("sending bulk request failed: %w", err)
	}

	// Force elastic to refresh indexes to get new batch data
	if err := client.refresh(ctx); err != nil {
		return fmt.Errorf("refreshing indices failed: %w", err)
	}

	return nil
}

type testIndexer struct {
	client *elasticsearch5.Client
	major  int
}

func newTestIndexer(ctx context.Context, baseURL string) (*testIndexer, error) {
	client, err := elasticsearch5.NewClient(elasticsearch5.Config{
		Addresses: []string{baseURL},
		Transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
		},
	})
	if err != nil {
		return nil, err
	}

	idx := &testIndexer{client: client}
	major, err := idx.probeMajor(ctx)
	if err != nil {
		return nil, err
	}
	idx.major = major

	return idx, nil
}

func (idx *testIndexer) bulkIndex(ctx context.Context, index string, docs []nginxlog) error {
	meta := map[string]any{
		"_index": index,
	}
	if idx.major <= 6 {
		meta["_type"] = "testquery_data"
	}
	metaLine, err := json.Marshal(map[string]any{"index": meta})
	if err != nil {
		return err
	}

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	for _, doc := range docs {
		body.Write(metaLine)
		body.WriteByte('\n')
		if err := encoder.Encode(doc); err != nil {
			return err
		}
	}

	var result struct {
		Errors bool `json:"errors"`
	}
	res, err := idx.client.Bulk(
		&body,
		idx.client.Bulk.WithContext(ctx),
	)
	if err := idx.handleResponse(res, err, &result); err != nil {
		return err
	}
	if result.Errors {
		return errors.New("bulk indexing reported item errors")
	}

	return nil
}

func (idx *testIndexer) refresh(ctx context.Context) error {
	res, err := idx.client.Indices.Refresh(
		idx.client.Indices.Refresh.WithContext(ctx),
	)
	return idx.handleResponse(res, err, nil)
}

func (idx *testIndexer) probeMajor(ctx context.Context) (int, error) {
	var info struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}

	res, err := idx.client.Info(idx.client.Info.WithContext(ctx))
	if err := idx.handleResponse(res, err, &info); err != nil {
		return 0, err
	}

	majorText, _, _ := strings.Cut(info.Version.Number, ".")
	major, err := strconv.Atoi(majorText)
	if err != nil {
		return 0, fmt.Errorf("parsing Elasticsearch version %q failed: %w", info.Version.Number, err)
	}

	return major, nil
}

func (*testIndexer) handleResponse(res *esapi5.Response, err error, out any) error {
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: %s", res.Status(), strings.TrimSpace(string(data)))
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}

	return nil
}
