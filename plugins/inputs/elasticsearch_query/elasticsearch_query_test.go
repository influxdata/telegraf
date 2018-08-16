package elasticsearch_query

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	elastic "gopkg.in/olivere/elastic.v5"
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

	e := &ElasticsearchQuery{
		URLs:                []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout:             internal.Duration{Duration: time.Second * 30},
		HealthCheckInterval: internal.Duration{Duration: time.Second * 30},
		acc:                 &acc,

		Aggregations: []Aggregation{
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
				Index:           testindex,
				MeasurementName: "measurement4",
				MetricFields:    []string{"size"},
				FilterQuery:     "downloads",
				MetricFunction:  "min",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"response.keyword"},
			},
			{
				Index:           testindex,
				MeasurementName: "measurement5",
				FilterQuery:     "product_2",
				DateField:       "@timestamp",
				QueryPeriod:     internal.Duration{Duration: time.Second * 600},
				Tags:            []string{"URI.keyword"},
			},
		},
	}

	fields1 := map[string]interface{}{
		"size_avg":  float64(202.30038022813687),
		"doc_count": int64(263),
	}

	fields2 := map[string]interface{}{
		"size_max":  float64(3318),
		"doc_count": int64(237),
	}

	fields3 := map[string]interface{}{
		"size_sum":  float64(22790),
		"doc_count": int64(22),
	}

	fields4 := map[string]interface{}{
		"size_min":  float64(0),
		"doc_count": int64(22),
	}

	fields5 := map[string]interface{}{
		"doc_count": int64(237),
	}

	err := e.connectToES()
	if err != nil {
		fmt.Printf("Error connecting to Elasticsearch")
	}

	bulkRequest := e.ESClient.Bulk()

	// populate elasticsearch with nginx_logs test data file
	file, err := os.Open("testdata/nginx_logs")
	if err != nil {
		fmt.Printf("Error opening testdata file")
	}

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

		bulkRequest.Add(elastic.NewBulkIndexRequest().
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

	// wait 5s for Elasticsearch to index, so results are consistent
	time.Sleep(time.Second * 5)

	require.NoError(t, e.Gather(&acc))

	if len(acc.Errors) > 0 {
		t.Errorf("%s", acc.Errors)
	}

	tags := map[string]string{
		"URI_keyword": "/downloads/product_1",
	}

	acc.AssertContainsTaggedFields(t, "measurement1", fields1, tags)

	tags = map[string]string{
		"URI_keyword": "/downloads/product_2",
	}

	acc.AssertContainsTaggedFields(t, "measurement2", fields2, tags)
	acc.AssertContainsTaggedFields(t, "measurement5", fields5, tags)

	tags = map[string]string{
		"response_keyword": "200",
	}

	acc.AssertContainsTaggedFields(t, "measurement3", fields3, tags)
	acc.AssertContainsTaggedFields(t, "measurement4", fields4, tags)

}

func TestElasticsearchQuery_getMetricFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	type args struct {
		ctx         context.Context
		aggregation Aggregation
	}

	e := &ElasticsearchQuery{
		URLs:                []string{"http://" + testutil.GetLocalHost() + ":9200"},
		Timeout:             internal.Duration{Duration: time.Second * 5},
		HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
	}

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
				Aggregation{
					Index:        testindex,
					MetricFields: []string{"URI", "http_version", "method", "response", "size"},
				},
			},
			map[string]string{
				"URI":          "text",
				"http_version": "text",
				"method":       "text",
				"response":     "text",
				"size":         "long"},
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
