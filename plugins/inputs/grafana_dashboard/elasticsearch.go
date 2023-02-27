package grafana_dashboard

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aquasecurity/esquery"
	"github.com/grafana-tools/sdk"
	"github.com/influxdata/telegraf"
)

type ElasticsearchDateHistogramAggregation struct {
	name        string
	field       string
	minDocCount uint64
	interval    string
	format      string
	aggs        []esquery.Aggregation
}

type ElasticsearchResponseItemAggregationBucket struct {
	DocCount uint64      `json:"doc_count,omitempty"`
	Key      interface{} `json:"key,omitempty"`
}

type ElasticsearchResponseItemAggregation struct {
	Buckets []map[string]interface{} `json:"buckets,omitempty"`
}

type ElasticsearchResponseItemShards struct {
	Failed     uint `json:"failed"`
	Skipped    uint `json:"skipped"`
	Successful uint `json:"successful"`
	Total      uint `json:"total"`
}

type ElasticsearchResponseItem struct {
	Aggregations map[string]interface{}          `json:"aggregations,omitempty"`
	Hits         map[string]interface{}          `json:"hits"`
	Status       uint                            `json:"status"`
	TimedOut     bool                            `json:"timed_out"`
	Took         uint                            `json:"took"`
	Shards       ElasticsearchResponseItemShards `json:"_shards"`
}

type ElasticsearchResponse struct {
	Responses []*ElasticsearchResponseItem `json:"responses"`
}

type ElasticsearchType struct {
	SearchType        string `json:"search_type"`
	IgnoreUnavailable bool   `json:"ignore_unavailable"`
	Index             string `json:"index"`
}

type ElasticsearchQueryString struct {
	Query           string `json:"query"`
	AnalyzeWildcard bool   `json:"analyze_wildcard"`
}

type Elasticsearch struct {
	log     telegraf.Logger
	grafana *Grafana
}

func ElasticsearchDateHistogramAgg(name, field string) *ElasticsearchDateHistogramAggregation {
	return &ElasticsearchDateHistogramAggregation{
		name:  name,
		field: field,
	}
}

func (agg *ElasticsearchDateHistogramAggregation) Name() string {
	return agg.name
}

func (agg *ElasticsearchDateHistogramAggregation) MinDocCount(d uint64) *ElasticsearchDateHistogramAggregation {
	agg.minDocCount = d
	return agg
}

func (agg *ElasticsearchDateHistogramAggregation) Interval(s string) *ElasticsearchDateHistogramAggregation {
	agg.interval = s
	return agg
}

func (agg *ElasticsearchDateHistogramAggregation) Format(s string) *ElasticsearchDateHistogramAggregation {
	agg.format = s
	return agg
}

func (agg *ElasticsearchDateHistogramAggregation) Aggs(aggs ...esquery.Aggregation) *ElasticsearchDateHistogramAggregation {
	agg.aggs = aggs
	return agg
}

func (agg *ElasticsearchDateHistogramAggregation) Map() map[string]interface{} {
	innerMap := map[string]interface{}{
		"field": agg.field,
	}

	innerMap["min_doc_count"] = agg.minDocCount
	if agg.interval != "" {
		innerMap["interval"] = agg.interval
	}
	if agg.format != "" {
		innerMap["format"] = agg.format
	}

	outerMap := map[string]interface{}{
		"date_histogram": innerMap,
	}
	if len(agg.aggs) > 0 {
		subAggs := make(map[string]map[string]interface{})
		for _, sub := range agg.aggs {
			subAggs[sub.Name()] = sub.Map()
		}
		outerMap["aggs"] = subAggs
	}

	return outerMap
}

func (esqs *ElasticsearchQueryString) Map() map[string]interface{} {
	return map[string]interface{}{
		"query_string": esqs,
	}
}

func (es *Elasticsearch) GetData(t *sdk.Target, ds *sdk.Datasource, period *GrafanaDashboardPeriod, push GrafanaDatasourcePushFunc) error {

	t1, t2 := period.StartEnd()
	start := int(t1.UTC().UnixMilli())
	end := int(t2.UTC().UnixMilli())

	est := ElasticsearchType{
		SearchType:        "query_then_fetch",
		IgnoreUnavailable: true,
		Index:             "*",
	}
	b1, err := json.Marshal(est)
	if err != nil {
		return err
	}

	fields := make(map[string]string)
	var aggs []esquery.Aggregation
	var last esquery.Aggregation

	for _, ba := range t.BucketAggs {

		field := "timestamp"
		if t.TimeField != "" {
			field = t.TimeField
		}
		if ba.Field != "" {
			field = ba.Field
		}

		var minDocCount uint64
		if ba.Settings.MinDocCount != nil {
			mdc, ok := ba.Settings.MinDocCount.(uint64)
			if ok {
				minDocCount = mdc
			}
		}

		var agg esquery.Aggregation
		if ba.Type == "terms" {
			ta := esquery.TermsAgg(ba.ID, field)
			ss, err := strconv.ParseUint(ba.Settings.Size, 10, 64)
			if err == nil && ss > 0 {
				ta.Size(ss)
			}
			if ba.Settings.OrderBy != "" {
				order := make(map[string]string)
				order[ba.Settings.OrderBy] = ba.Settings.Order
				ta.Order(order)
			}
			ta.Map()["min_doc_count"] = minDocCount
			agg = ta
		} else if ba.Type == "date_histogram" {

			dh := ElasticsearchDateHistogramAgg(ba.ID, field)
			if ba.Settings.Interval != "" {
				if ba.Settings.Interval == "auto" {
					dh.Interval(t.Interval)
				} else {
					dh.Interval(ba.Settings.Interval)
				}
			} else {
				dh.Interval(t.Interval)
			}
			dh.Format("epoch_millis")
			dh.MinDocCount(uint64(minDocCount))
			agg = dh
		} else {
			es.log.Debugf("Elasticsearch type %s is not supported", ba.Type)
		}

		if agg == nil {
			continue
		}

		if len(aggs) == 0 {
			aggs = append(aggs, agg)
		} else {

			ta, ok := last.(*esquery.TermsAggregation)
			if ok {
				ta.Aggs(agg)
			}

			dh, ok := last.(*ElasticsearchDateHistogramAggregation)
			if ok {
				dh.Aggs(agg)
			}
		}
		last = agg
		fields[ba.ID] = field
	}

	if last != nil {

		var aggs []esquery.Aggregation

		if len(t.Metrics) > 0 {

			for _, ms := range t.Metrics {

				var agg esquery.Aggregation
				switch ms.Type {
				case "avg":
					agg = esquery.Avg(ms.ID, ms.Field)
				case "min":
					agg = esquery.Min(ms.ID, ms.Field)
				case "max":
					agg = esquery.Max(ms.ID, ms.Field)
				case "sum":
					agg = esquery.Sum(ms.ID, ms.Field)
				case "count":
					//agg = esquery.CustomAgg("", make(map[string]interface{}))
				}
				if agg != nil {
					aggs = append(aggs, agg)
					fields[ms.ID] = ms.Field
				}
			}
			dh, ok := last.(*ElasticsearchDateHistogramAggregation)
			if ok {
				dh.Aggs(aggs...)
			}
		}
	}

	q := "*"
	if t.Query != "" {
		q = t.Query
	}
	var filters []esquery.Mappable
	filters = append(filters, esquery.Range("timestamp").Gte(start).Lte(end).Format("epoch_millis"))
	filters = append(filters, &ElasticsearchQueryString{
		AnalyzeWildcard: true,
		Query:           q,
	})
	esq := esquery.Query(esquery.Bool().Filter(filters...)).Aggs(aggs...)
	esq.Size(0)

	b2, err := json.Marshal(esq)
	if err != nil {
		return err
	}

	s := fmt.Sprintf("%s\n%s\n", string(b1), string(b2))
	es.log.Debugf("Elasticsearch request body => %s", s)

	when := time.Now()

	URL := fmt.Sprintf("/api/datasources/proxy/%d/_msearch", ds.ID)
	raw, code, err := es.grafana.httpPost(URL, nil, []byte(s))
	if err != nil {
		return nil
	}
	if code != 200 {
		return fmt.Errorf("Elasticsearch HTTP error %d: returns %s", code, raw)
	}
	var res ElasticsearchResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return err
	}

	if len(res.Responses) == 0 {
		return fmt.Errorf("Elasticsearch has no fields")
	}

	var tags map[string]string

	for _, r := range res.Responses {

		if r == nil {
			continue
		}

		for k, a := range r.Aggregations {

			if a == nil {
				continue
			}
			if fields[k] == "" {
				continue
			}

			if t.Alias != "" {
				tags = make(map[string]string)
				tags["alias"] = t.Alias
			}
			es.processAggregation(when, k, a, tags, fields, push)
		}
	}
	return nil
}

func (es *Elasticsearch) processAggregation(when time.Time, key string, agg interface{}, tags, fields map[string]string, push GrafanaDatasourcePushFunc) {

	m, ok := agg.(map[string]interface{})
	if ok && m["buckets"] != nil {

		if tags == nil {
			tags = make(map[string]string)
		}

		items, ok := m["buckets"].([]interface{})
		if ok {

			for _, item := range items {

				if item == nil {
					continue
				}

				im, ok := item.(map[string]interface{})
				if ok {

					var ky float64
					kys := ""
					if im["key"] != nil {
						k, ok := im["key"].(string)
						if ok {
							tags[fields[key]] = k
						}
						ki, ok := im["key"].(float64)
						if ok {
							ky = ki
							kys = fmt.Sprintf("%.0f", ki)
						}
					}

					kas := ""
					if im["key_as_string"] != nil {
						k, ok := im["key_as_string"].(string)
						if ok {
							kas = k
						}
					}

					var dc float64
					if im["doc_count"] != nil {
						d, ok := im["doc_count"].(float64)
						if ok {
							dc = d
						}
					}

					if ky > 0 && kys == kas {
						push(when, tags, time.UnixMilli(int64(ky)), dc)
						continue
					}

					for k, a := range im {
						if a == nil {
							continue
						}
						if fields[k] == "" {
							continue
						}
						es.processAggregation(when, k, a, tags, fields, push)
					}

				}
			}
		}
	}
}

func NewElasticsearch(log telegraf.Logger, grafana *Grafana) *Elasticsearch {

	return &Elasticsearch{
		log:     log,
		grafana: grafana,
	}
}
