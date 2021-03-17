package elasticsearch_query

import (
	"bufio"
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	elastic5 "gopkg.in/olivere/elastic.v5"
)

var (
	testindex = "test-elasticsearch_query-" + strconv.Itoa(int(time.Now().Unix()))
	setupOnce sync.Once
	e         = &ElasticsearchQuery{
		URLs:    []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout: internal.Duration{Duration: time.Second * 30},
	}
)

type esAggregationQueryTest struct {
	queryName                 string
	testAggregationQueryInput esAggregation
	testAggregationQueryData  []aggregationQueryData
	testMapFields             map[string]string
	expectedMetrics           []expectedMetric
	wantBuildQueryErr         bool
	wantGetMetricFieldsErr    bool
	wantQueryResErr           bool
}

type expectedMetric struct {
	measurement string
	fields      map[string]interface{}
	tags        map[string]string
}

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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{"URI.keyword"},
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
		map[string]string{"size": "long"},
		[]expectedMetric{
			{
				measurement: "measurement1",
				fields:      map[string]interface{}{"size_avg": float64(202.30038022813687), "doc_count": int64(263)},
				tags:        map[string]string{"URI_keyword": "/downloads/product_1"},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{"URI.keyword"},
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
		map[string]string{"size": "long"},
		[]expectedMetric{
			{
				measurement: "measurement2",
				fields:      map[string]interface{}{"size_max": float64(3301), "doc_count": int64(263)},
				tags:        map[string]string{"URI_keyword": "/downloads/product_1"},
			},
			{
				measurement: "measurement2",
				fields:      map[string]interface{}{"size_max": float64(3318), "doc_count": int64(237)},
				tags:        map[string]string{"URI_keyword": "/downloads/product_2"},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{"response.keyword"},
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
		map[string]string{"size": "long"},
		[]expectedMetric{
			{
				measurement: "measurement3",
				fields:      map[string]interface{}{"size_sum": float64(22790), "doc_count": int64(22)},
				tags:        map[string]string{"response_keyword": "200"},
			},
			{
				measurement: "measurement3",
				fields:      map[string]interface{}{"size_sum": float64(0), "doc_count": int64(219)},
				tags:        map[string]string{"response_keyword": "304"},
			},
			{
				measurement: "measurement3",
				fields:      map[string]interface{}{"size_sum": float64(86932), "doc_count": int64(259)},
				tags:        map[string]string{"response_keyword": "404"},
			},
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
			QueryPeriod:       internal.Duration{Duration: time.Second * 600},
			IncludeMissingTag: true,
			MissingTagValue:   "missing",
			Tags:              []string{"response.keyword", "URI.keyword", "method.keyword"},
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
		map[string]string{"size": "long", "response_time": "long"},
		[]expectedMetric{
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(318), "response_time_min": float64(126), "doc_count": int64(146)},
				tags:        map[string]string{"response_keyword": "404", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(0), "response_time_min": float64(71), "doc_count": int64(113)},
				tags:        map[string]string{"response_keyword": "304", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(490), "response_time_min": float64(1514), "doc_count": int64(3)},
				tags:        map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1", "method_keyword": "GET"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(318), "response_time_min": float64(237), "doc_count": int64(113)},
				tags:        map[string]string{"response_keyword": "404", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(0), "response_time_min": float64(134), "doc_count": int64(106)},
				tags:        map[string]string{"response_keyword": "304", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(490), "response_time_min": float64(2), "doc_count": int64(13)},
				tags:        map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2", "method_keyword": "GET"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(0), "response_time_min": float64(8479), "doc_count": int64(1)},
				tags:        map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1", "method_keyword": "HEAD"},
			},
			{
				measurement: "measurement4",
				fields:      map[string]interface{}{"size_min": float64(0), "response_time_min": float64(1059), "doc_count": int64(5)},
				tags:        map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2", "method_keyword": "HEAD"},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{"URI.keyword"},
		},
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement5", name: "URI_keyword", function: "terms", field: "URI.keyword"},
				isParent: true,
			},
		},
		map[string]string{},
		[]expectedMetric{
			{
				measurement: "measurement5",
				fields:      map[string]interface{}{"doc_count": int64(237)},
				tags:        map[string]string{"URI_keyword": "/downloads/product_2"},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{"URI.keyword", "response.keyword"},
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
		map[string]string{},
		[]expectedMetric{
			{
				measurement: "measurement6",
				fields:      map[string]interface{}{"doc_count": int64(4)},
				tags:        map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_1"},
			},
			{
				measurement: "measurement6",
				fields:      map[string]interface{}{"doc_count": int64(18)},
				tags:        map[string]string{"response_keyword": "200", "URI_keyword": "/downloads/product_2"},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{},
		},
		nil,
		map[string]string{},
		[]expectedMetric{
			{
				measurement: "measurement7",
				fields:      map[string]interface{}{"doc_count": int64(22)},
				tags:        map[string]string{},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{},
		},
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement8", name: "size_max", function: "max", field: "size"},
				isParent: true,
			},
		},
		map[string]string{"size": "long"},
		[]expectedMetric{
			{
				measurement: "measurement8",
				fields:      map[string]interface{}{"size_max": float64(3318)},
				tags:        map[string]string{},
			},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{},
		},
		nil,
		map[string]string{"size": "long"},
		nil,
		true,
		false,
		false,
	},
	{
		"query 10 - non-existing metric field",
		esAggregation{
			Index:           testindex,
			MeasurementName: "measurement10",
			MetricFields:    []string{"none"},
			DateField:       "@timestamp",
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{},
		},
		nil,
		map[string]string{},
		nil,
		false,
		false,
		false,
	},
	{
		"query 11 - non-existing index field",
		esAggregation{
			Index:           "notanindex",
			MeasurementName: "measurement11",
			DateField:       "@timestamp",
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{},
		},
		nil,
		map[string]string{},
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
			QueryPeriod:     internal.Duration{Duration: time.Second * 600},
			Tags:            []string{},
		},
		[]aggregationQueryData{
			{
				aggKey:   aggKey{measurement: "measurement12", name: "size_avg", function: "avg", field: "size"},
				isParent: true,
			},
		},
		map[string]string{"size": "long"},
		nil,
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
			QueryPeriod:       internal.Duration{Duration: time.Second * 600},
			IncludeMissingTag: false,
			Tags:              []string{"nothere"},
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
		map[string]string{"size": "long"},
		nil,
		false,
		false,
		false,
	},
}

func setupIntegrationTest() {
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

	err := e.connectToES()
	if err != nil {
		log.Printf("E! error setting up integration test: %s", err.Error())
		return
	}

	bulkRequest := e.esClient.Bulk()

	// populate elasticsearch with nginx_logs test data file
	file, err := os.Open("testdata/nginx_logs")
	if err != nil {
		log.Printf("E! error setting up integration test: %s", err.Error())
		return
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

	if err = scanner.Err(); err != nil {
		log.Printf("E! error setting up integration test: %s", err.Error())
		return
	}

	_, err = bulkRequest.Do(context.Background())
	if err != nil {
		log.Printf("E! error setting up integration test: %s", err.Error())
		return
	}

	// wait 5s (default) for Elasticsearch to index, so results are consistent
	time.Sleep(time.Second * 5)
}

func TestElasticsearchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var acc testutil.Accumulator

	setupOnce.Do(setupIntegrationTest)
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

	for _, result := range testEsAggregationData {
		for _, r := range result.expectedMetrics {
			acc.AssertContainsTaggedFields(t, r.measurement, r.fields, r.tags)
		}
	}

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

	type args struct {
		ctx         context.Context
		aggregation esAggregation
	}

	setupOnce.Do(setupIntegrationTest)

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
			d.testMapFields,
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
	type args struct {
		mapMetricFields map[string]string
		aggregation     esAggregation
	}

	type test struct {
		name    string
		args    args
		want    []aggregationQueryData
		wantErr bool
	}
	var tests []test

	for _, d := range testEsAggregationData {
		tests = append(tests, test{
			"build " + d.queryName,
			args{d.testMapFields, d.testAggregationQueryInput},
			d.testAggregationQueryData,
			d.wantBuildQueryErr,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ElasticsearchQuery{}
			got, err := e.buildAggregationQuery(tt.args.mapMetricFields, tt.args.aggregation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElasticsearchQuery.buildAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(aggKey{}, aggregationQueryData{}),
				cmpopts.IgnoreFields(aggregationQueryData{}, "aggregation"),
				cmpopts.SortSlices(func(x, y aggregationQueryData) bool { return x.aggKey.name > y.aggKey.name }),
			}

			if !cmp.Equal(got, tt.want, opts...) {
				t.Errorf("ElasticsearchQuery.buildAggregationQuery(): %s error = %s ", tt.name, cmp.Diff(got, tt.want, opts...))
			}
		})
	}
}
