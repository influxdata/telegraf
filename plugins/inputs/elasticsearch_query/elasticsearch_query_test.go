package elasticsearch_query

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	elastic5 "gopkg.in/olivere/elastic.v5"
)

var (
	testindex = "test-elasticsearch_query-" + strconv.Itoa(int(time.Now().Unix()))
	setupOnce sync.Once
)

type esAggregationQueryTest struct {
	queryName                 string
	testAggregationQueryInput esAggregation
	testAggregationQueryData  []aggregationQueryData
	expectedMetrics           []telegraf.Metric
	wantBuildQueryErr         bool
	wantGetMetricFieldsErr    bool
	wantQueryResErr           bool
}

var queryPeriod = config.Duration(time.Second * 600)

var testEsAggregationData = []esAggregationQueryTest{
	{
		"query 1",
		esAggregation{
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
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement1", name: "size_avg", function: "avg", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement1", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
			testutil.MustMetric(
				"measurement1",
				map[string]string{"URI_keyword": "/downloads/product_1"},
				map[string]interface{}{"size_avg": float64(202.30038022813687), "doc_count": int64(263)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
		false,
		false,
		false,
	},
	{
		"query 2",
		esAggregation{
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
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement2", name: "size_max", function: "max", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement2", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
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
		false,
		false,
		false,
	},
	{
		"query 3",
		esAggregation{
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
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement3", name: "size_sum", function: "sum", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement3", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
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
		false,
		false,
		false,
	},
	{
		"query 4",
		esAggregation{
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
		[]aggregationQueryData{
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
		[]telegraf.Metric{
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
		false,
		false,
		false,
	},
	{
		"query 5",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement5",
			FilterQuery:     "product_2",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"URI.keyword"},
			mapMetricFields: map[string]string{},
		},
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement5", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
			testutil.MustMetric(
				"measurement5",
				map[string]string{"URI_keyword": "/downloads/product_2"},
				map[string]interface{}{"doc_count": int64(237)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
		false,
		false,
		false,
	},
	{
		"query 6",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement6",
			FilterQuery:     "response: 200",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{"URI.keyword", "response.keyword"},
			mapMetricFields: map[string]string{},
		},
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement6", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement6", name: "response_keyword", function: "terms", field: "response.keyword"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
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
		false,
		false,
		false,
	},
	{
		"query 7 - simple query",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement7",
			FilterQuery:     "response: 200",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		nil,
		[]telegraf.Metric{
			testutil.MustMetric(
				"measurement7",
				map[string]string{},
				map[string]interface{}{"doc_count": int64(22)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
		false,
		false,
		false,
	},
	{
		"query 8",
		esAggregation{
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
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement8", name: "size_max", function: "max", field: "size"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
			testutil.MustMetric(
				"measurement8",
				map[string]string{},
				map[string]interface{}{"size_max": float64(3318)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
		false,
		false,
		false,
	},
	{
		"query 9 - invalid function",
		esAggregation{
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
		nil,
		nil,
		true,
		false,
		true,
	},
	{
		"query 10 - non-existing metric field",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement10",
			MetricFields:    []string{"none"},
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		nil,
		nil,
		false,
		false,
		true,
	},
	{
		"query 11 - non-existing index field",
		esAggregation{
			Index:           "notanindex",
			MeasurementName: "measurement11",
			DateField:       "@timestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		nil,
		nil,
		false,
		false,
		true,
	},
	{
		"query 12 - non-existing timestamp field",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement12",
			MetricFields:    []string{"size"},
			MetricFunction:  "avg",
			DateField:       "@notatimestamp",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{"size": "long"},
		},
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement12", name: "size_avg", function: "avg", field: "size"},
				isParent: true,
			},
		},
		[]telegraf.Metric{
			testutil.MustMetric(
				"measurement12",
				map[string]string{},
				map[string]interface{}{"size_avg": float64(0)},
				time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			),
		},
		false,
		false,
		false,
	},
	{
		"query 13 - non-existing tag field",
		esAggregation{
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
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement13", name: "size_avg", function: "avg", field: "size"},
				isParent: false,
			},
			{
				aggKey:   aggKey{measurement: "measurement13", name: "nothere", function: "terms", field: "nothere"},
				isParent: true,
			},
		},
		nil,
		false,
		false,
		false,
	},
	{
		"query 14 - non-existing custom date/time format",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement14",
			DateField:       "@timestamp",
			DateFieldFormat: "yyyy",
			QueryPeriod:     queryPeriod,
			Tags:            []string{},
			mapMetricFields: map[string]string{},
		},
		nil,
		nil,
		false,
		false,
		true,
	},
}

func setupIntegrationTest() error {
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

	e := &ElasticsearchQuery{
		URLs:    []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout: config.Duration(time.Second * 30),
		Log:     testutil.Logger{},
	}

	err := e.connectToES()
	if err != nil {
		return err
	}

	bulkRequest := e.esClient.Bulk()

	// populate elasticsearch with nginx_logs test data file
	file, err := os.Open("testdata/nginx_logs")
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		size, _ := strconv.Atoi(parts[9])
		responseTime, _ := strconv.Atoi(parts[len(parts)-1])

		logline := nginxlog{
			IPaddress:    parts[0],
			Timestamp:    time.Now().UTC(),
			Method:       strings.Replace(parts[5], `"`, "", -1),
			URI:          parts[6],
			Httpversion:  strings.Replace(parts[7], `"`, "", -1),
			Response:     parts[8],
			Size:         float64(size),
			ResponseTime: float64(responseTime),
		}

		bulkRequest.Add(elastic5.NewBulkIndexRequest().
			Index(testindex).
			Type("testquery_data").
			Doc(logline))
	}
	if scanner.Err() != nil {
		return err
	}

	_, err = bulkRequest.Do(context.Background())
	if err != nil {
		return err
	}

	// wait 5s (default) for Elasticsearch to index, so results are consistent
	time.Sleep(time.Second * 5)
	return nil
}

func TestElasticsearchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupOnce.Do(func() {
		err := setupIntegrationTest()
		require.NoError(t, err)
	})

	var acc testutil.Accumulator
	e := &ElasticsearchQuery{
		URLs:    []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout: config.Duration(time.Second * 30),
		Log:     testutil.Logger{},
	}

	err := e.connectToES()
	require.NoError(t, err)

	var aggs []esAggregation
	var aggsErr []esAggregation

	for _, agg := range testEsAggregationData {
		if !agg.wantQueryResErr {
			aggs = append(aggs, agg.testAggregationQueryInput)
		}
	}
	e.Aggregations = aggs

	require.NoError(t, e.Init())
	require.NoError(t, e.Gather(&acc))

	if len(acc.Errors) > 0 {
		t.Errorf("%s", acc.Errors)
	}

	var expectedMetrics []telegraf.Metric
	for _, result := range testEsAggregationData {
		expectedMetrics = append(expectedMetrics, result.expectedMetrics...)
	}
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.SortMetrics(), testutil.IgnoreTime())

	// aggregations that should return an error
	for _, agg := range testEsAggregationData {
		if agg.wantQueryResErr {
			aggsErr = append(aggsErr, agg.testAggregationQueryInput)
		}
	}
	e.Aggregations = aggsErr
	require.NoError(t, e.Init())
	require.NoError(t, e.Gather(&acc))

	if len(acc.Errors) != len(aggsErr) {
		t.Errorf("expecting %v query result errors, got %v: %s", len(aggsErr), len(acc.Errors), acc.Errors)
	}
}

func TestElasticsearchQuery_getMetricFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupOnce.Do(func() {
		err := setupIntegrationTest()
		require.NoError(t, err)
	})

	type args struct {
		ctx         context.Context
		aggregation esAggregation
	}

	e := &ElasticsearchQuery{
		URLs:    []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout: config.Duration(time.Second * 30),
		Log:     testutil.Logger{},
	}

	err := e.connectToES()
	require.NoError(t, err)

	type test struct {
		name    string
		e       *ElasticsearchQuery
		args    args
		want    map[string]string
		wantErr bool
	}

	var tests []test

	for _, d := range testEsAggregationData {
		tests = append(tests, test{
			"getMetricFields " + d.queryName,
			e,
			args{context.Background(), d.testAggregationQueryInput},
			d.testAggregationQueryInput.mapMetricFields,
			d.wantGetMetricFieldsErr,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.getMetricFields(tt.args.ctx, tt.args.aggregation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElasticsearchQuery.buildAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("ElasticsearchQuery.getMetricFields() = error = %s", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestElasticsearchQuery_buildAggregationQuery(t *testing.T) {
	type test struct {
		name        string
		aggregation esAggregation
		want        []aggregationQueryData
		wantErr     bool
	}
	var tests []test

	for _, d := range testEsAggregationData {
		tests = append(tests, test{
			"build " + d.queryName,
			d.testAggregationQueryInput,
			d.testAggregationQueryData,
			d.wantBuildQueryErr,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.aggregation.buildAggregationQuery()
			if (err != nil) != tt.wantErr {
				t.Errorf("ElasticsearchQuery.buildAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(aggKey{}, aggregationQueryData{}),
				cmpopts.IgnoreFields(aggregationQueryData{}, "aggregation"),
				cmpopts.SortSlices(func(x, y aggregationQueryData) bool { return x.aggKey.name > y.aggKey.name }),
			}

			if !cmp.Equal(tt.aggregation.aggregationQueryList, tt.want, opts...) {
				t.Errorf("ElasticsearchQuery.buildAggregationQuery(): %s error = %s ", tt.name, cmp.Diff(tt.aggregation.aggregationQueryList, tt.want, opts...))
			}
		})
	}
}
