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
	"github.groupondev.com/metrics/telegraf-tdigest-plugin/plugins/aggregators/tdigestagg/constants"
	"strings"
	"time"
)

var weight = 1.

type Aggregation interface {
	histogram() *TDigest
	tags() map[string]string
	basicName() string
	sum() float64
	time() time.Time

	addValue(floatValue float64)
	emit(acc telegraf.Accumulator)
}

type AggregationData struct {
	_histogram *TDigest
	_tags      map[string]string
	_basicName string
	_sum       float64
	_time      time.Time
}

type CentralAggregation struct {
	AggregationData
}

func (ca *CentralAggregation) emit(acc telegraf.Accumulator) {
	fields := map[string]interface{}{}

	if strings.Contains(ca._tags[constants.TagKeyAggregates], constants.FieldSum) {
		fields[constants.FieldSum+constants.UtilityFieldModifier] = ca._sum
	}

	var histOut = ca._histogram.ForJson()
	fields[constants.FieldCompression] = histOut.Compression
	// TODO: Create formatter to ensure consistent behaviour
	fields[constants.FieldCentroids] = fmt.Sprint(histOut.Centroids)

	acc.AddFields(ca._basicName, fields, ca._tags, ca._time)
}

type LocalAggregation struct {
	AggregationData
}

func (la *LocalAggregation) emit(acc telegraf.Accumulator) {
	fields := map[string]interface{}{}

	delete(la._tags, constants.TagKeyBucketKey)
	delete(la._tags, constants.TagKeyAggregates)
	la._tags[constants.TagKeySource] = la._tags[constants.TagKeyHost]

	// local aggregations
	fields[constants.FieldMaximum] = la._histogram.Max()
	fields[constants.FieldMinimum] = la._histogram.Min()
	fields[constants.FieldCount] = la._histogram.Count()
	fields[constants.FieldMedian] = la._histogram.Quantile(0.50)

	acc.AddFields(la._basicName, fields, la._tags)
}

func (ad *AggregationData) histogram() *TDigest     { return ad._histogram }
func (ad *AggregationData) tags() map[string]string { return ad._tags }
func (ad *AggregationData) basicName() string       { return ad._basicName }
func (ad *AggregationData) sum() float64            { return ad._sum }
func (ad *AggregationData) time() time.Time         { return ad._time }

func newAggregationData(name string, tags map[string]string, compression float64, firstValue float64, time time.Time) AggregationData {
	aggregationData := AggregationData{
		_histogram: NewTDigest(compression),
		_tags:      tags,
		_basicName: name,
		_sum:       0,
		_time:      time,
	}

	aggregationData.addValue(firstValue)
	return aggregationData
}

func (ad *AggregationData) addValue(floatValue float64) {
	ad._histogram.Add(floatValue, weight)
	ad._sum += floatValue
}

func (ad *AggregationData) getAggregation(name string) float64 {
	switch name {
	case constants.FieldMinimum:
		return ad._histogram.Min()
	case constants.FieldMaximum:
		return ad._histogram.Max()
	case constants.FieldSum:
		return ad._sum
	case constants.FieldCount:
		return ad._histogram.Count()
	case constants.FieldMedian:
		return ad._histogram.Quantile(0.50)
	case constants.FieldPercentile90:
		return ad._histogram.Quantile(0.90)
	case constants.FieldPercentile95:
		return ad._histogram.Quantile(0.95)
	case constants.FieldPercentile99:
		return ad._histogram.Quantile(0.99)
	}

	fmt.Println("Unsupported Aggregation: " + name)
	return -1
}
