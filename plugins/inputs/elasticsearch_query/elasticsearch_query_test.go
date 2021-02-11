package elasticsearch_query

import (
	"bufio"
	"context"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	elastic5 "gopkg.in/olivere/elastic.v5"
)

var testindex = "test-elasticsearch_query-" + strconv.Itoa(int(time.Now().Unix()))

func TestElasticsearchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var acc testutil.Accumulator

	type nginxlog struct {
		IPaddress   string    `json:"IP"`
		Timestamp   time.Time `json:"@timestamp"`
		Method      string    `json:"method"`
		URI         string    `json:"URI"`
		Httpversion string    `json:"http_version"`
		Response    string    `json:"response"`
		Size        float64   `json:"size"`
	}

	type expectedResult struct {
		measurement string
		fields      map[string]interface{}
		tags        map[string]string
	}

	e := &ElasticsearchQuery{
		URLs:    []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout: internal.Duration{Duration: time.Second * 30},

		Aggregations: []esAggregation{
			{
				Index:           testindex,
				MeasurementName: "measurement1",
				MetricFields:    []string{"size"},
				FilterQuery:     "product_1",
				MetricFunction:  "avg",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"URI.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement2",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "max",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"URI.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement3",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "sum",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"response.keyword"},
			},
			{
				Index:             testindex,
				MeasurementName:   "measurement4",
				MetricFields:      []string{"size"},
				FilterQuery:       "downloads",
				MetricFunction:    "min",
				DateField:         "@timestamp",
				QueryPeriod:       internal.Duration{Duration: time.Second * 600},
				IncludeMissingTag: true,
				MissingTagValue:   "missing",
				Tags:              []string{"response.keyword", "URI.keyword", "method.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement5",
				FilterQuery:     "product_2",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"URI.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement6",
				FilterQuery:     "response: 200",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"URI.keyword", "response.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement7",
				FilterQuery:     "response: 200",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement8",
				FilterQuery:     "",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"URI.keyword"},
			},
		},
	}

	elasticSearchQueryResults := []expectedResult{
		{
			measurement: "measurement1",
			fields: map[string]interface{}{
				"size_avg":  float64(202.30038022813687),
				"doc_count": int64(263),
			},
			tags: map[string]string{
				"URI_keyword": "/downloads/product_1",
			},
		},
		{
			measurement: "measurement2",
			fields: map[string]interface{}{
				"size_max":  float64(3318),
				"doc_count": int64(237),
			},
			tags: map[string]string{
				"URI_keyword": "/downloads/product_2",
			},
		},
		{
			measurement: "measurement3",
			fields: map[string]interface{}{
				"size_sum":  float64(22790),
				"doc_count": int64(22),
			},
			tags: map[string]string{
				"response_keyword": "200",
			},
		},
		{
			measurement: "measurement3",
			fields: map[string]interface{}{
				"size_sum":  float64(0),
				"doc_count": int64(219),
			},
			tags: map[string]string{
				"response_keyword": "304",
			},
		},
		{
			measurement: "measurement3",
			fields: map[string]interface{}{
				"size_sum":  float64(86932),
				"doc_count": int64(259),
			},
			tags: map[string]string{
				"response_keyword": "404",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(490),
				"doc_count": int64(3),
			},
			tags: map[string]string{
				"response_keyword": "200",
				"URI_keyword":      "/downloads/product_1",
				"method_keyword":   "GET",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(318),
				"doc_count": int64(146),
			},
			tags: map[string]string{
				"response_keyword": "404",
				"URI_keyword":      "/downloads/product_1",
				"method_keyword":   "GET",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(0),
				"doc_count": int64(113),
			},
			tags: map[string]string{
				"response_keyword": "304",
				"URI_keyword":      "/downloads/product_1",
				"method_keyword":   "GET",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(0),
				"doc_count": int64(1),
			},
			tags: map[string]string{
				"response_keyword": "200",
				"URI_keyword":      "/downloads/product_1",
				"method_keyword":   "HEAD",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(490),
				"doc_count": int64(13),
			},
			tags: map[string]string{
				"response_keyword": "200",
				"URI_keyword":      "/downloads/product_2",
				"method_keyword":   "GET",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(0),
				"doc_count": int64(106),
			},
			tags: map[string]string{
				"response_keyword": "304",
				"URI_keyword":      "/downloads/product_2",
				"method_keyword":   "GET",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(0),
				"doc_count": int64(5),
			},
			tags: map[string]string{
				"response_keyword": "200",
				"URI_keyword":      "/downloads/product_2",
				"method_keyword":   "HEAD",
			},
		},
		{
			measurement: "measurement4",
			fields: map[string]interface{}{
				"size_min":  float64(318),
				"doc_count": int64(113),
			},
			tags: map[string]string{
				"response_keyword": "404",
				"URI_keyword":      "/downloads/product_2",
				"method_keyword":   "GET",
			},
		},
		{
			measurement: "measurement5",
			fields: map[string]interface{}{
				"doc_count": int64(237),
			},
			tags: map[string]string{
				"URI_keyword": "/downloads/product_2",
			},
		},
		{
			measurement: "measurement6",
			fields: map[string]interface{}{
				"doc_count": int64(4),
			},
			tags: map[string]string{
				"response_keyword": "200",
				"URI_keyword":      "/downloads/product_1",
			},
		},
		{
			measurement: "measurement6",
			fields: map[string]interface{}{
				"doc_count": int64(18),
			},
			tags: map[string]string{
				"response_keyword": "200",
				"URI_keyword":      "/downloads/product_2",
			},
		},
		{
			measurement: "measurement7",
			fields: map[string]interface{}{
				"doc_count": int64(22),
			},
			tags: map[string]string{},
		},
		{
			measurement: "measurement8",
			fields: map[string]interface{}{
				"doc_count": int64(263),
			},
			tags: map[string]string{
				"URI_keyword": "/downloads/product_1",
			},
		},
		{
			measurement: "measurement8",
			fields: map[string]interface{}{
				"doc_count": int64(237),
			},
			tags: map[string]string{
				"URI_keyword": "/downloads/product_2",
			},
		},
	}

	err := e.connectToES()
	require.NoError(t, err)

	bulkRequest := e.esClient.Bulk()

	// populate elasticsearch with nginx_logs test data file
	file, err := os.Open("testdata/nginx_logs")
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		size, _ := strconv.Atoi(parts[9])

		logline := nginxlog{
			IPaddress:   parts[0],
			Timestamp:   time.Now().UTC(),
			Method:      strings.Replace(parts[5], `"`, "", -1),
			URI:         parts[6],
			Httpversion: strings.Replace(parts[7], `"`, "", -1),
			Response:    parts[8],
			Size:        float64(size),
		}

		bulkRequest.Add(elastic5.NewBulkIndexRequest().
			Index(testindex).
			Type("testquery_data").
			Doc(logline))
	}

	if err = scanner.Err(); err != nil {
		t.Errorf("Error reading testdata file")
	}

	_, err = bulkRequest.Do(context.Background())
	if err != nil {
		t.Errorf("Error sending bulk request to Elasticsearch: %s", err)
	}

	// wait 5s (default) for Elasticsearch to index, so results are consistent
	time.Sleep(time.Second * 5)

	require.NoError(t, e.Init())
	require.NoError(t, e.Gather(&acc))

	if len(acc.Errors) > 0 {
		t.Errorf("%s", acc.Errors)
	}

	for _, r := range elasticSearchQueryResults {
		acc.AssertContainsTaggedFields(t, r.measurement, r.fields, r.tags)
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

	e := &ElasticsearchQuery{
		URLs:                []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout:             internal.Duration{Duration: time.Second * 5},
		HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
	}

	err := e.connectToES()
	require.NoError(t, err)

	tests := []struct {
		name    string
		e       *ElasticsearchQuery
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			"getMetricFields",
			e,
			args{
				context.Background(),
				esAggregation{
					Index:        testindex,
					MetricFields: []string{"URI", "http_version", "method", "response", "size"},
				},
			},
			map[string]string{
				"URI":          "text",
				"http_version": "text",
				"method":       "text",
				"response":     "text",
				"size":         "long",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.getMetricFields(tt.args.ctx, tt.args.aggregation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElasticsearchQuery.getMetricFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ElasticsearchQuery.getMetricFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
