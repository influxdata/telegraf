//Copyright (c) 2019, Groupon, Inc.
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are
//met:
//
//Redistributions of source code must retain the above copyright notice,
//this list of conditions and the following disclaimer.
//
//Redistributions in binary form must reproduce the above copyright
//notice, this list of conditions and the following disclaimer in the
//documentation and/or other materials provided with the distribution.
//
//Neither the name of GROUPON nor the names of its contributors may be
//used to endorse or promote products derived from this software without
//specific prior written permission.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
//IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
//TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
//PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
//TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
//PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
//LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
//NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package tdigestagg

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/bucketing"
	"github.com/influxdata/telegraf/selfstat"
	"time"
)

var noValuePresent = "none"

type CacheKey struct {
	bucket string
	time   time.Time
}

type TDigestAgg struct {
	cache       map[CacheKey]*Aggregation
	Compression float64                  `toml:"compression"`
	Bucketing   []bucketing.BucketConfig `toml:"bucketing"`
}

func NewTDigestAgg() *TDigestAgg {
	td := &TDigestAgg{}

	td.Reset()
	return td
}

var sampleOutput = ` Pretty formatted for readability
{
  "fields": {
    "sum._utility": 1230.0,
    "centroids": "[{97.97979797979798 1} {97.97979797979798 1} {98 1} {98 1} {98 1} {98 1} {98 1} {98 1} {98.00990099009901 2} {98.01980198019803 2} {98.01980198019803 2} {98.01980198019803 2} {98.98989898989899 1} {98.98989898989899 2} {99 1} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99 2} {99.00990099009901 2} {99.00990099009901 2} {99.00990099009901 2} {100 2} {100 2} {100 1} {100 1} {100 1} {100 1} {100 1} {100 1} {100 1}]",
    "compression": 30
  },
  "name": "cpu_usage_idle",
  "tags": {
    "cpu": "cpu1",
    "source": "C02S121GG8WL.group.on",
    "az": "snc1",
    "env": "dev",
    "service": "awesome",
    "aggregates": "max,min,count,p99,p95,avg,med",
    "bucket_key": "cpu_usage_idle_awesome_snc1_dev"
  },
  "timestamp": 1532630290113371000
}
`

var sampleTDigestJson = `[{97.97979797979798 1} {98 1} {98.00990099009901 2} {99 1} {100 1}]`

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "60s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true

  ## TDigest Compression
  ## This value corresponds to the number of centroids the histogram will use
  ## Higher values increase size of data but also precision of calculated percentiles
  compression = 30.0

  [[aggregators.tdigestagg.bucketing]]
	## List of tags that will not be considered for aggregation and not emitted
	exclude_tags=[host]
	## 
	source_tag_key=service
	## Optional: Default value is "atom"
	atom_replacement_tag_key=service

  ## Subsequent bucketing configurations will all ingest the same points
  [[aggregators.tdigest.bucketing]]
	exclude_tags=[]
	source_tag_key=service

  ## All supported macro names should be added here
  ## This logic could potentially be supported w/o config but this functionality already
  [aggregators.tdigestagg.tagpass]
    _rollup = ["timer*", "counter*", "gauge*", "local*", "default*"]
`

func (agg *TDigestAgg) SampleConfig() string {
	return sampleConfig
}

func (agg *TDigestAgg) Description() string {
	return "Keep a histogram representation of each metric passing through, if appropriately tagged."
}

func (agg *TDigestAgg) Add(input telegraf.Metric) {
	for measureName, stringValue := range input.Fields() {
		if floatValue, ok := convert(stringValue); ok {
			baseName := input.Name() + "_" + measureName
			minuteTime := input.Time().Round(time.Minute)

			bucketDataChannel := make(chan bucketing.BucketData)
			go generateBucketData(baseName, input.Tags(), agg, bucketDataChannel)

			// TODO: logic to prevent a data prevent a point from being put in the same config twice by
			//			separate Bucketing Configs.  i.e. config2 excludes a tag that is not present on the metric
			for range agg.Bucketing {
				bt := <-bucketDataChannel
				cacheKey := CacheKey{bt.BucketKey(), minuteTime}

				bucket, bucketExists := agg.cache[cacheKey]
				if bucketExists {
					(*bucket).addValue(floatValue)
				} else {
					aggregationData := newAggregationData(baseName, bt.OutputTags(), agg.Compression, floatValue, minuteTime)

					var aggregation Aggregation
					if bt.Aggregates() == noValuePresent {
						// TODO: Evaluate - Local aggregations add late data into "current" config or drop it
						aggregation = &LocalAggregation{aggregationData}
					} else {
						aggregation = &CentroidAggregation{aggregationData}
					}

					agg.cache[cacheKey] = &aggregation
				}
			}
		}
	}
}

func (agg *TDigestAgg) Push(acc telegraf.Accumulator) {

	for _, agg := range agg.cache {
		(*agg).emit(acc)

		delay := fmt.Sprintf("%.1f", time.Since((*agg).time()).Minutes()-1)
		selfstat.Register("tdigest", "window_cardinality", map[string]string{"delay": delay}).Incr(1)
	}
}

func (agg *TDigestAgg) Reset() {
	agg.cache = make(map[CacheKey]*Aggregation)
}

func convert(input interface{}) (float64, bool) {
	switch inputType := input.(type) {
	case float64:
		return inputType, true
	case int64:
		return float64(inputType), true
	default:
		return 0, false
	}
}

// TODO: Create ticket for adding init check
//func (agg *TDigestAgg) Init() error {
//	if len(agg.Bucketing) == 0 {
//		return errors.New("no bucket configurations defined")
//	}
//
//	for _, bucket := range agg.Bucketing {
//		if "" == bucket.SourceTagKey {
//			return errors.New("every bucket configuration must define \"source_tag_key\"")
//		}
//	}
//
//	return nil
//}

func init() {
	aggregators.Add("tdigestagg", func() telegraf.Aggregator {
		return NewTDigestAgg()
	})
}
