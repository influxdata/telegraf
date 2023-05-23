package opensearch_query

import (
	"encoding/json"

	"github.com/influxdata/telegraf"
)

type AggregationResponse struct {
	Hits         *SearchHits  `json:"hits"`
	Aggregations *Aggregation `json:"aggregations"`
}

type SearchHits struct {
	TotalHits *TotalHits `json:"total,omitempty"`
}

type TotalHits struct {
	Relation string `json:"relation"`
	Value    int64  `json:"value"`
}

type MetricAggregation map[string]interface{}

type AggregateValue struct {
	metrics MetricAggregation
	buckets []BucketData
}

type Aggregation map[string]AggregateValue

type BucketData struct {
	DocumentCount int64  `json:"doc_count"`
	Key           string `json:"key"`

	subaggregation Aggregation
}

func (a *AggregationResponse) GetMetrics(acc telegraf.Accumulator, measurement string) error {
	// Simple case (no aggregations)
	if a.Aggregations == nil {
		tags := make(map[string]string)
		fields := map[string]interface{}{
			"doc_count": a.Hits.TotalHits.Value,
		}
		acc.AddFields(measurement, fields, tags)
		return nil
	}

	return a.Aggregations.GetMetrics(acc, measurement, a.Hits.TotalHits.Value, map[string]string{})
}

func (a *Aggregation) GetMetrics(acc telegraf.Accumulator, measurement string, docCount int64, tags map[string]string) error {
	var err error
	fields := make(map[string]interface{})
	for name, agg := range *a {
		if agg.IsAggregation() {
			for _, bucket := range agg.buckets {
				tt := map[string]string{name: bucket.Key}
				for k, v := range tags {
					tt[k] = v
				}
				err = bucket.subaggregation.GetMetrics(acc, measurement, bucket.DocumentCount, tt)
				if err != nil {
					return err
				}
			}
			return nil
		}
		for metric, value := range agg.metrics {
			switch value := value.(type) {
			case map[string]interface{}:
				for k, v := range value {
					fields[name+"_"+metric+"_"+k] = v
				}
			default:
				fields[name+"_"+metric] = value
			}
		}
	}

	fields["doc_count"] = docCount
	acc.AddFields(measurement, fields, tags)

	return nil
}

func (a *AggregateValue) UnmarshalJSON(bytes []byte) error {
	var partial map[string]json.RawMessage
	err := json.Unmarshal(bytes, &partial)
	if err != nil {
		return err
	}

	// We'll continue to unmarshal if we have buckets
	if b, found := partial["buckets"]; found {
		return json.Unmarshal(b, &a.buckets)
	}

	// Use the remaining bytes as metrics
	return json.Unmarshal(bytes, &a.metrics)
}

func (a *AggregateValue) IsAggregation() bool {
	return !(a.buckets == nil)
}

func (b *BucketData) UnmarshalJSON(bytes []byte) error {
	var partial map[string]json.RawMessage
	var err error

	err = json.Unmarshal(bytes, &partial)
	if err != nil {
		return err
	}

	err = json.Unmarshal(partial["doc_count"], &b.DocumentCount)
	if err != nil {
		return err
	}
	delete(partial, "doc_count")

	err = json.Unmarshal(partial["key"], &b.Key)
	if err != nil {
		return err
	}
	delete(partial, "key")

	if b.subaggregation == nil {
		b.subaggregation = make(Aggregation)
	}

	for name, message := range partial {
		var subaggregation AggregateValue
		err = json.Unmarshal(message, &subaggregation)
		if err != nil {
			return err
		}
		b.subaggregation[name] = subaggregation
	}

	return nil
}
