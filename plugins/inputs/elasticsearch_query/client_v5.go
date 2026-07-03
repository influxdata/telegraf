package elasticsearch_query

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	elastic5 "gopkg.in/olivere/elastic.v5"

	"github.com/influxdata/telegraf"
)

type clientV5 struct {
	url        string
	httpClient *http.Client
	client     *elastic5.Client
	log        telegraf.Logger
}

func (e *ElasticsearchQuery) newClientV5() (client, error) {
	// Make sure the HTTP client exists
	httpClient, err := e.HTTPClientConfig.CreateClient(context.Background(), e.Log)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client failed: %w", err)
	}

	// Create a new ElasticSearch client
	clientOptions := []elastic5.ClientOptionFunc{
		elastic5.SetHttpClient(httpClient),
		elastic5.SetSniff(e.EnableSniffer),
		elastic5.SetURL(e.URLs...),
		elastic5.SetHealthcheckInterval(time.Duration(e.HealthCheckInterval)),
	}
	if e.Username != "" {
		clientOptions = append(clientOptions, elastic5.SetBasicAuth(e.Username, e.Password))
	}
	if time.Duration(e.HealthCheckInterval) == 0 {
		clientOptions = append(clientOptions, elastic5.SetHealthcheck(false))
	}

	c, err := elastic5.NewClient(clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("creating ElasticSearch client failed: %w", err)
	}

	return &clientV5{
		url:        e.URLs[0],
		httpClient: httpClient,
		client:     c,
		log:        e.Log,
	}, nil
}

func (c *clientV5) close() {
	if c.client != nil {
		c.client.Stop()
		c.client = nil
	}
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
		c.httpClient = nil
	}
}

func (c *clientV5) version() (string, error) {
	return c.client.ElasticsearchVersion(c.url)
}

func (c *clientV5) isRunning() bool {
	return c.client.IsRunning()
}

func (c *clientV5) getFieldMapping(ctx context.Context, index, field string) (map[string]interface{}, error) {
	return c.client.GetFieldMapping().Index(index).Field(field).Do(ctx)
}

type queryDataV5 struct {
	measurement string
	name        string
	function    string
	field       string
	isParent    bool
	aggregation elastic5.Aggregation
}

func (*clientV5) buildQueries(aggregation *aggregation) error {
	// Create one aggregation per metric field found or function defined for
	// numeric fields
	queries := make([]queryDataV5, 0, len(aggregation.mapMetricFields)+len(aggregation.Tags))
	for k, v := range aggregation.mapMetricFields {
		switch v {
		case "long", "float", "integer", "short", "double", "scaled_float":
		default:
			continue
		}

		var agg elastic5.Aggregation
		switch aggregation.MetricFunction {
		case "avg":
			agg = elastic5.NewAvgAggregation().Field(k)
		case "sum":
			agg = elastic5.NewSumAggregation().Field(k)
		case "min":
			agg = elastic5.NewMinAggregation().Field(k)
		case "max":
			agg = elastic5.NewMaxAggregation().Field(k)
		default:
			return fmt.Errorf("aggregation function %q not supported", aggregation.MetricFunction)
		}

		query := queryDataV5{
			measurement: aggregation.MeasurementName,
			function:    aggregation.MetricFunction,
			field:       k,
			name:        strings.ReplaceAll(k, ".", "_") + "_" + aggregation.MetricFunction,
			isParent:    true,
			aggregation: agg,
		}
		queries = append(queries, query)
	}

	// Create a terms aggregation per tag
	for _, term := range aggregation.Tags {
		agg := elastic5.NewTermsAggregation()
		if aggregation.IncludeMissingTag && aggregation.MissingTagValue != "" {
			agg.Missing(aggregation.MissingTagValue)
		}
		agg.Field(term).Size(1000)

		// add each previous parent aggregations as subaggregations of this terms aggregation
		for key, aggMap := range queries {
			if !aggMap.isParent {
				continue
			}

			agg.Field(term).SubAggregation(aggMap.name, aggMap.aggregation).Size(1000)

			// Update subaggregation map with parent information
			queries[key].isParent = false
		}

		query := queryDataV5{
			measurement: aggregation.MeasurementName,
			function:    "terms",
			field:       term,
			name:        strings.ReplaceAll(term, ".", "_"),
			isParent:    true,
			aggregation: agg,
		}
		queries = append(queries, query)
	}
	aggregation.queries = queries

	// Prepare measurement mapping to organize the aggregation query data
	// by measurement
	measurements := make(map[string]map[string]string, len(queries))
	for _, query := range queries {
		if measurements[query.measurement] == nil {
			measurements[query.measurement] = map[string]string{
				query.name: query.function,
			}
		} else {
			measurements[query.measurement][query.name] = query.function
		}
	}
	aggregation.measurements = measurements

	return nil
}

func (c *clientV5) query(ctx context.Context, aggregation *aggregation) (interface{}, int64, error) {
	// Type assert client-specific data
	// This can only be the expected type as it is set by this code with exactly
	// the expected type. If the assertion fails code will panic here but that's
	// okay as it can only happen on programming errors e.g. by mixing different
	// client versions!
	queries := aggregation.queries.([]queryDataV5)

	now := time.Now().UTC()
	from := now.Add(time.Duration(-aggregation.QueryPeriod))

	query := elastic5.NewBoolQuery()
	query = query.Filter(elastic5.NewQueryStringQuery(aggregation.FilterQuery))
	query = query.Filter(elastic5.NewRangeQuery(aggregation.DateField).From(from).To(now).Format(aggregation.DateFieldFormat))

	src, err := query.Source()
	if err != nil {
		return nil, 0, fmt.Errorf("getting query source failed: %w", err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil, 0, fmt.Errorf("unmarshal response failed: %w", err)
	}
	c.log.Debugf("{\"query\": %s}", string(data))

	// Add only parent elastic.Aggregations to the search request, all the rest
	// are subaggregations of these
	search := c.client.Search().Index(aggregation.Index).Query(query).Size(0)
	for _, v := range queries {
		if v.isParent && v.aggregation != nil {
			search.Aggregation(v.name, v.aggregation)
		}
	}

	result, err := search.Do(ctx)
	if err != nil {
		if result != nil {
			return result.Aggregations, result.Hits.TotalHits, fmt.Errorf("%s - %s", result.Error.Type, result.Error.Reason)
		}
		return nil, 0, err
	}

	if len(result.Aggregations) == 0 {
		return nil, result.Hits.TotalHits, nil
	}

	return result.Aggregations, result.Hits.TotalHits, nil
}

func (*clientV5) aggregate(acc telegraf.Accumulator, measurement string, nameFunction map[string]string, response interface{}) error {
	// Type assert client-specific data
	r := response.(elastic5.Aggregations)

	m := &iteratorV5{
		name:   measurement,
		fields: make(map[string]interface{}),
		tags:   make(map[string]string),
	}

	return m.iterate(acc, nameFunction, r)
}

type iteratorV5 struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

func (m *iteratorV5) iterate(acc telegraf.Accumulator, nameFunction map[string]string, response elastic5.Aggregations) error {
	names := make([]string, 0, len(response))
	for k := range response {
		if k != "key" && k != "doc_count" {
			names = append(names, k)
		}
	}
	if len(names) == 0 {
		// We've reached a single bucket or response without aggregation, i.e.
		// we've reached a leaf node. Add the accumulated metric and reset it
		if len(m.fields) > 0 {
			acc.AddFields(m.name, m.fields, m.tags)
			m.fields = make(map[string]interface{})
		}
		return nil
	}

	// Metrics aggregations response can contain multiple field values, so we
	// iterate over them
	for _, name := range names {
		function, found := nameFunction[name]
		if !found {
			return fmt.Errorf("child aggregation function %q not found %v", name, nameFunction)
		}

		// Execute the aggregation function
		var result interface{}
		switch function {
		case "avg":
			result, _ = response.Avg(name)
		case "sum":
			result, _ = response.Sum(name)
		case "min":
			result, _ = response.Min(name)
		case "max":
			result, _ = response.Max(name)
		case "terms":
			result, _ = response.Terms(name)
		default:
			return fmt.Errorf("aggregation %q not supported", function)
		}

		switch r := result.(type) {
		case *elastic5.AggregationBucketKeyItems:
			// We've found a terms aggregation, iterate over the buckets and try
			// to retrieve the inner aggregation values
			for _, bucket := range r.Buckets {
				s, ok := bucket.Key.(string)
				if !ok {
					return fmt.Errorf("bucket key is not a string (%s, %s)", name, function)
				}
				m.tags[name] = s
				m.fields["doc_count"] = bucket.DocCount

				// We need to recurse down through the buckets, as it may
				// contain another terms aggregation
				if err := m.iterate(acc, nameFunction, bucket.Aggregations); err != nil {
					return err
				}
				delete(m.tags, name)
			}
		case *elastic5.AggregationValueMetric:
			if r.Value != nil {
				m.fields[name] = *r.Value
			} else {
				m.fields[name] = float64(0)
			}
		default:
			return fmt.Errorf("aggregation type %T not supported", r)
		}
	}

	// If there are fields here it comes from a metrics aggregation without a
	// parent terms aggregation
	if len(m.fields) > 0 {
		acc.AddFields(m.name, m.fields, m.tags)
		m.fields = make(map[string]interface{})
	}

	return nil
}
