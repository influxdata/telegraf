package elasticsearch_query

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	elastic5 "gopkg.in/olivere/elastic.v5"

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
	expectedData := [][]aggregationQueryData{
		{
			{
				aggKey:   aggKey{measurement: "measurement1", name: "size_avg", function: "avg", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement1", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement2", name: "size_max", function: "max", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement2", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement3", name: "size_sum", function: "sum", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement3", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement4", name: "size_min", function: "min", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement4", name: "response_time_min", function: "min", field: "response_time"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement4", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement4", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement4", name: "method_keyword", function: "terms", field: "method.keyword"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement5", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement6", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement6", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: true,
			},
		},
		nil,
		{
			{
				aggKey:   aggKey{measurement: "measurement8", name: "size_max", function: "max", field: "size"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement12", name: "size_avg", function: "avg", field: "size"},
				isParent: true,
			},
		},
		{
			{
				aggKey:   aggKey{measurement: "measurement13", name: "size_avg", function: "avg", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement13", name: "nothere", function: "terms", field: "nothere"},
				isParent: true,
			},
		},
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
		Aggregations: []esAggregation{
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
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	// Check the query data
	opts := []cmp.Option{
		cmp.AllowUnexported(aggKey{}, aggregationQueryData{}),
		cmpopts.IgnoreFields(aggregationQueryData{}, "aggregation"),
		cmpopts.SortSlices(func(x, y aggregationQueryData) bool { return x.aggKey.name > y.aggKey.name }),
	}

	for i, agg := range plugin.Aggregations {
		actual := agg.aggregationQueryList
		expected := expectedData[i]
		require.Truef(t, cmp.Equal(expected, actual, opts...), "mismatch in aggregation %d\nexpected:%v\nactual:%v\n", i, expected, actual)
	}

	// Check the metrics
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.SortMetrics(), testutil.IgnoreTime())
}

func TestGatherFailIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define expectations
	expected := []string{
		`elasticsearch query aggregation measurement9: aggregation function "average" not supported`,
		`elasticsearch query aggregation measurement10: metric field "none" not found on index "test-elasticsearch"`,
		"elasticsearch query aggregation measurement11: elastic: Error 404 (Not Found): no such index [type=index_not_found_exception]",
		"elasticsearch query aggregation measurement14: elastic: Error 400 (Bad Request): all shards failed [type=search_phase_execution_exception]",
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

	// Setup plugin
	plugin := &ElasticsearchQuery{
		URLs: []string{addr},
		Aggregations: []esAggregation{
			{
				Index:           testindex,
				MeasurementName: "measurement9",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "average",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			{
				Index:           testindex,
				MeasurementName: "measurement10",
				MetricFields:    []string{"none"},
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			{
				Index:           "notanindex",
				MeasurementName: "measurement11",
				DateField:       "@timestamp",
				QueryPeriod:     config.Duration(time.Second * 600),
			},
			{
				Index:           testindex,
				MeasurementName: "measurement14",
				DateField:       "@timestamp",
				DateFieldFormat: "yyyy",
				QueryPeriod:     config.Duration(time.Second * 600),
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

	// Collect data and check the errors
	require.NoError(t, plugin.Gather(&acc))

	// Check the errors
	actual := make([]string, 0, len(acc.Errors))
	for _, err := range acc.Errors {
		actual = append(actual, err.Error())
	}
	require.ElementsMatch(t, expected, actual)
}

func sendData(ctx context.Context, url string) error {
	// Read the data
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
	options := []elastic5.ClientOptionFunc{
		elastic5.SetSniff(false),
		elastic5.SetURL(url),
		elastic5.SetHealthcheckInterval(10 * time.Second),
	}
	client, err := elastic5.NewClient(options...)
	if err != nil {
		return fmt.Errorf("creating client failed: %w", err)
	}

	// Create bulk request for the data
	bulkRequest := client.Bulk()
	for _, logline := range logs {
		bulkRequest.Add(elastic5.NewBulkIndexRequest().
			Index(testindex).
			Type("testquery_data").
			Doc(logline),
		)
	}
	if _, err := bulkRequest.Do(ctx); err != nil {
		return fmt.Errorf("sending bulk request failed: %w", err)
	}

	// Force elastic to refresh indexes to get new batch data
	if _, err := client.Refresh().Do(ctx); err != nil {
		return fmt.Errorf("refreshing indices failed: %w", err)
	}

	return nil
}
