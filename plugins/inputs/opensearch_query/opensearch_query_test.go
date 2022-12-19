package opensearch_query

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	servicePort = "9200"
	testindex   = "test-opensearch"
)

type osAggregationQueryTest struct {
	queryName                 string
	testAggregationQueryInput osAggregation
	testAggregationQueryData  []aggregationQueryData
	expectedMetrics           []telegraf.Metric
	wantBuildQueryErr         bool
	wantGetMetricFieldsErr    bool
	wantQueryResErr           bool
	wantInitErr               bool
}

var queryPeriod = config.Duration(time.Second * 600)

var testOpensearchAggregationData = []osAggregationQueryTest{
	{
		queryName: "query 1",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement1",
			MetricFields:    []string{"size"},
			FilterQuery:     "product_1",
			MetricFunction:  "avg",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"URI.keyword"},
			mapMetricFields: map[string]string{"size": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement1", name: "size_avg", function: "avg", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement1", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement1",
				map[string]string{"URI_keyword": "/downloads/product_1"},
				map[string]interface{}{"size_avg": float64(202.30038022813687), "doc_count": int64(263)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 2",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement2",
			MetricFields:    []string{"size"},
			FilterQuery:     "downloads",
			MetricFunction:  "max",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"URI.keyword"},
			mapMetricFields: map[string]string{"size": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement2", name: "size_max", function: "max", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement2", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement2",
				map[string]string{"URI_keyword": "/downloads/product_1"},
				map[string]interface{}{"size_max": float64(3301), "doc_count": int64(263)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement2",
				map[string]string{"URI_keyword": "/downloads/product_2"},
				map[string]interface{}{"size_max": float64(3318), "doc_count": int64(237)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 3",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement3",
			MetricFields:    []string{"size"},
			FilterQuery:     "downloads",
			MetricFunction:  "sum",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"response.keyword"},
			mapMetricFields: map[string]string{"size": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement3", name: "size_sum", function: "sum", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement3", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement3",
				map[string]string{"response_keyword": "200"},
				map[string]interface{}{"size_sum": float64(22790), "doc_count": int64(22)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement3",
				map[string]string{"response_keyword": "304"},
				map[string]interface{}{"size_sum": float64(0), "doc_count": int64(219)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement3",
				map[string]string{"response_keyword": "404"},
				map[string]interface{}{"size_sum": float64(86932), "doc_count": int64(259)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 4",
		testAggregationQueryInput: osAggregation{
			Index:             testindex,
			MeasurementName:   "measurement4",
			MetricFields:      []string{"size", "response_time"},
			FilterQuery:       "downloads",
			MetricFunction:    "min",
			DateField:         "@timestamp",
			QueryPeriod:       queryPeriod,
			IncludeMissingTag: true,
			MissingTagValue:   "missing",
			Tags:              []string{"response.keyword", "URI.keyword", "method.keyword"},
			mapMetricFields:   map[string]string{"size": "long", "response_time": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
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
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "404", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
				map[string]interface{}{"size_min": float64(318), "response_time_min": float64(126), "doc_count": int64(146)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "304", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
				map[string]interface{}{"size_min": float64(0), "response_time_min": float64(71), "doc_count": int64(113)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
				map[string]interface{}{"size_min": float64(490), "response_time_min": float64(1514), "doc_count": int64(3)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "404", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
				map[string]interface{}{"size_min": float64(318), "response_time_min": float64(237), "doc_count": int64(113)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "304", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
				map[string]interface{}{"size_min": float64(0), "response_time_min": float64(134), "doc_count": int64(106)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
				map[string]interface{}{"size_min": float64(490), "response_time_min": float64(2), "doc_count": int64(13)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1", "method_keyword": "HEAD"},
				map[string]interface{}{"size_min": float64(0), "response_time_min": float64(8479), "doc_count": int64(1)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement4",
				map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2", "method_keyword": "HEAD"},
				map[string]interface{}{"size_min": float64(0), "response_time_min": float64(1059), "doc_count": int64(5)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 5",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement5",
			FilterQuery:     "product_2",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"URI.keyword"},
			mapMetricFields: map[string]string{},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement5", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement5",
				map[string]string{"URI_keyword": "/downloads/product_2"},
				map[string]interface{}{"doc_count": int64(237)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 6",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement6",
			FilterQuery:     "response: 200",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"URI.keyword", "response.keyword"},
			mapMetricFields: map[string]string{},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement6", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement6", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement6",
				map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1"},
				map[string]interface{}{"doc_count": int64(4)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
			testutil.MustMetric(
				"measurement6",
				map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2"},
				map[string]interface{}{"doc_count": int64(18)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 7 - simple query",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement7",
			FilterQuery:     "response: 200",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement7",
				map[string]string{},
				map[string]interface{}{"doc_count": int64(22)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 8",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement8",
			MetricFields:    []string{"size"},
			FilterQuery:     "downloads",
			MetricFunction:  "max",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{"size": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement8", name: "size_max", function: "max", field: "size"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement8",
				map[string]string{},
				map[string]interface{}{"size_max": float64(3318)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 9 - invalid function",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement9",
			MetricFields:    []string{"size"},
			FilterQuery:     "downloads",
			MetricFunction:  "average",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{"size": "long"},
		},
		wantBuildQueryErr: true,
		wantInitErr:       true,
	},
	{
		queryName: "query 10 - non-existing metric field",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement10",
			MetricFields:    []string{"none"},
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		wantQueryResErr: true,
		wantInitErr:     true,
	},
	{
		queryName: "query 11 - non-existing index field",
		testAggregationQueryInput: osAggregation{
			Index:           "notanindex",
			MeasurementName: "measurement11",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		wantQueryResErr: true,
	},
	{
		queryName: "query 12 - non-existing timestamp field",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement12",
			MetricFields:    []string{"size"},
			MetricFunction:  "avg",
			DateField:       "@notatimestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{"size": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement12", name: "size_avg", function: "avg", field: "size"},
				isParent: true,
			},
		},
		expectedMetrics: []telegraf.Metric{
			testutil.MustMetric(
				"measurement12",
				map[string]string{},
				map[string]interface{}{"size_avg": float64(0)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
	},
	{
		queryName: "query 13 - non-existing tag field",
		testAggregationQueryInput: osAggregation{
			Index:             testindex,
			MeasurementName:   "measurement13",
			MetricFields:      []string{"size"},
			MetricFunction:    "avg",
			DateField:         "@timestamp",
			QueryPeriod:       queryPeriod,
			IncludeMissingTag: false,
			Tags:              []string{"nothere"},
			mapMetricFields:   map[string]string{"size": "long"},
		},
		testAggregationQueryData: []aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement13", name: "size_avg", function: "avg", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement13", name: "nothere", function: "terms", field: "nothere"},
				isParent: true,
			},
		},
	},
	{
		queryName: "query 14 - non-existing custom date/time format",
		testAggregationQueryInput: osAggregation{
			Index:           testindex,
			MeasurementName: "measurement14",
			DateField:       "@timestamp",
			DateFieldFormat: "yyyy",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		wantQueryResErr: true,
	},
}

func setupIntegrationTest(t *testing.T) (*testutil.Container, error) {
	var err error

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

	container := testutil.Container{
		Image:        "opensearchproject/opensearch:2.4.0",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type": "single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog(".opendistro_security is used as internal security index."),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")

	url := fmt.Sprintf("https://%s:%s", container.Address, container.Ports[servicePort])

	o := &OpensearchQuery{
		URLs:         []string{url},
		Timeout:      config.Duration(time.Second * 30),
		Log:          testutil.Logger{},
		Username:     "admin",
		Password:     "admin",
		ClientConfig: tls.ClientConfig{InsecureSkipVerify: true},
	}

	err = o.connectToOpensearch()
	if err != nil {
		return &container, err
	}

	// parse and build query
	file, err := os.Open("testdata/nginx_logs")
	if err != nil {
		return &container, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	indexer, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Client:  o.osClient,
		Index:   testindex,
		Refresh: "true",
	})
	if err != nil {
		return &container, err
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		size, _ := strconv.Atoi(parts[9])
		responseTime, _ := strconv.Atoi(parts[len(parts)-1])

		logline := nginxlog{
			IPaddress:    parts[0],
			Timestamp:    time.Now().UTC(),
			Method:       strings.ReplaceAll(parts[5], `"`, ""),
			URI:          parts[6],
			Httpversion:  strings.ReplaceAll(parts[7], `"`, ""),
			Response:     parts[8],
			Size:         float64(size),
			ResponseTime: float64(responseTime),
		}

		body, e := json.Marshal(logline)
		if e != nil {
			return &container, e
		}

		e = indexer.Add(
			context.Background(),
			opensearchutil.BulkIndexerItem{
				Index:  testindex,
				Action: "index",
				Body:   strings.NewReader(string(body)),
			})
		if e != nil {
			return &container, e
		}
	}

	if scanner.Err() != nil {
		return &container, err
	}

	if err = indexer.Close(context.Background()); err != nil {
		return &container, err
	}

	return &container, nil
}

func TestOpensearchQueryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container, err := setupIntegrationTest(t)
	require.NoError(t, err)
	defer container.Terminate()

	o := &OpensearchQuery{
		URLs: []string{
			fmt.Sprintf("https://%s:%s", container.Address, container.Ports[servicePort]),
		},
		Timeout:      config.Duration(time.Second * 30),
		Log:          testutil.Logger{},
		Username:     "admin",
		Password:     "admin",
		ClientConfig: tls.ClientConfig{InsecureSkipVerify: true},
	}

	err = o.connectToOpensearch()
	require.NoError(t, err)

	for _, tt := range testOpensearchAggregationData {
		t.Run(tt.queryName, func(t *testing.T) {
			var err error
			var acc testutil.Accumulator

			o.Aggregations = []osAggregation{tt.testAggregationQueryInput}
			err = o.Init()
			if (err != nil) != tt.wantInitErr {
				t.Errorf("OpensearchQuery.Init() error = %v, wantInitErr %v", err, tt.wantInitErr)
				return
			} else if err != nil {
				// Init() failures mean we're done
				return
			}

			err = o.Gather(&acc)
			if (len(acc.Errors) > 0) && !tt.wantQueryResErr {
				for _, err = range acc.Errors {
					t.Errorf("OpensearchQuery.Gather() error: %v, wantQueryResErr %v", err, tt.wantQueryResErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error %v", err)
				return
			}

			testutil.RequireMetricsEqual(t, tt.expectedMetrics, acc.GetTelegrafMetrics(), testutil.SortMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestOpensearchQueryIntegration_getMetricFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container, err := setupIntegrationTest(t)
	require.NoError(t, err)
	defer container.Terminate()

	type args struct {
		ctx         context.Context
		aggregation osAggregation
	}

	e := &OpensearchQuery{
		URLs: []string{
			fmt.Sprintf("https://%s:%s", container.Address, container.Ports[servicePort]),
		},
		Timeout:      config.Duration(time.Second * 30),
		Log:          testutil.Logger{},
		Username:     "admin",
		Password:     "admin",
		ClientConfig: tls.ClientConfig{InsecureSkipVerify: true},
	}

	err = e.connectToOpensearch()
	require.NoError(t, err)

	type test struct {
		name    string
		e       *OpensearchQuery
		args    args
		want    map[string]string
		wantErr bool
	}

	tests := make([]test, 0, len(testOpensearchAggregationData))
	for _, d := range testOpensearchAggregationData {
		tests = append(tests, test{
			name:    "getMetricFields " + d.queryName,
			e:       e,
			args:    args{context.Background(), d.testAggregationQueryInput},
			want:    d.testAggregationQueryInput.mapMetricFields,
			wantErr: d.wantGetMetricFieldsErr,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.getMetricFields(tt.args.ctx, tt.args.aggregation)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpensearchQuery.buildAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("OpensearchQuery.getMetricFields() = error = %s", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestOpensearchQuery_buildAggregationQuery(t *testing.T) {
	type test struct {
		name        string
		aggregation osAggregation
		want        []aggregationQueryData
		wantErr     bool
	}

	tests := make([]test, 0, len(testOpensearchAggregationData))
	for _, d := range testOpensearchAggregationData {
		tests = append(tests, test{
			name:        "build " + d.queryName,
			aggregation: d.testAggregationQueryInput,
			want:        d.testAggregationQueryData,
			wantErr:     d.wantBuildQueryErr,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.aggregation.buildAggregationQuery()
			if (err != nil) != tt.wantErr {
				t.Errorf("OpensearchQuery.buildAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(aggKey{}, aggregationQueryData{}),
				cmpopts.IgnoreFields(aggregationQueryData{}, "aggregation"),
				cmpopts.SortSlices(func(x, y aggregationQueryData) bool { return x.aggKey.name > y.aggKey.name }),
			}

			if !cmp.Equal(tt.aggregation.aggregationQueryList, tt.want, opts...) {
				t.Errorf("OpensearchQuery.buildAggregationQuery(): %s error = %s ", tt.name, cmp.Diff(tt.aggregation.aggregationQueryList, tt.want, opts...))
			}
		})
	}
}

func TestAggregationSearchBody(t *testing.T) {
	aa := aggregationSearchBody{
		Size:         0,
		Aggregations: make(map[name]aggType),
	}

	ss, err := json.Marshal(aa)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ss))
}
