package cloudwatch

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type statisticType int

const (
	statisticTypeNone statisticType = iota
	statisticTypeMax
	statisticTypeMin
	statisticTypeSum
	statisticTypeCount
)

type cloudwatchField interface {
	addValue(sType statisticType, value float64)
	buildDatum() []types.MetricDatum
}

type statisticField struct {
	measurement string
	dimensions  []types.Dimension
	name        string
	values      map[statisticType]float64
	timestamp   time.Time
	resolution  int64
}

func (f *statisticField) addValue(sType statisticType, value float64) {
	if sType != statisticTypeNone {
		f.values[sType] = value
	}
}

func (f *statisticField) buildDatum() []types.MetricDatum {
	if f.hasAllFields() {
		// If we have all required fields, we build datum with StatisticValues
		vmin := f.values[statisticTypeMin]
		vmax := f.values[statisticTypeMax]
		vsum := f.values[statisticTypeSum]
		vcount := f.values[statisticTypeCount]

		datum := types.MetricDatum{
			MetricName: aws.String(strings.Join([]string{f.measurement, f.name}, "_")),
			Dimensions: f.dimensions,
			Timestamp:  aws.Time(f.timestamp),
			StatisticValues: &types.StatisticSet{
				Minimum:     aws.Float64(vmin),
				Maximum:     aws.Float64(vmax),
				Sum:         aws.Float64(vsum),
				SampleCount: aws.Float64(vcount),
			},
			StorageResolution: aws.Int32(int32(f.resolution)),
		}

		return []types.MetricDatum{datum}
	}

	// If we don't have all required fields, we build each field as independent datum
	datums := make([]types.MetricDatum, 0, len(f.values))
	for sType, value := range f.values {
		datum := types.MetricDatum{
			Value:      aws.Float64(value),
			Dimensions: f.dimensions,
			Timestamp:  aws.Time(f.timestamp),
		}

		switch sType {
		case statisticTypeMin:
			datum.MetricName = aws.String(strings.Join([]string{f.measurement, f.name, "min"}, "_"))
		case statisticTypeMax:
			datum.MetricName = aws.String(strings.Join([]string{f.measurement, f.name, "max"}, "_"))
		case statisticTypeSum:
			datum.MetricName = aws.String(strings.Join([]string{f.measurement, f.name, "sum"}, "_"))
		case statisticTypeCount:
			datum.MetricName = aws.String(strings.Join([]string{f.measurement, f.name, "count"}, "_"))
		default:
			// should not be here
			continue
		}

		datums = append(datums, datum)
	}

	return datums
}

func (f *statisticField) hasAllFields() bool {
	_, hasMin := f.values[statisticTypeMin]
	_, hasMax := f.values[statisticTypeMax]
	_, hasSum := f.values[statisticTypeSum]
	_, hasCount := f.values[statisticTypeCount]

	return hasMin && hasMax && hasSum && hasCount
}

type valueField struct {
	measurement string
	dimensions  []types.Dimension
	name        string
	value       float64
	timestamp   time.Time
	resolution  int64
}

func (f *valueField) addValue(sType statisticType, value float64) {
	if sType == statisticTypeNone {
		f.value = value
	}
}

func (f *valueField) buildDatum() []types.MetricDatum {
	return []types.MetricDatum{
		{
			MetricName:        aws.String(strings.Join([]string{f.measurement, f.name}, "_")),
			Value:             aws.Float64(f.value),
			Dimensions:        f.dimensions,
			Timestamp:         aws.Time(f.timestamp),
			StorageResolution: aws.Int32(int32(f.resolution)),
		},
	}
}
