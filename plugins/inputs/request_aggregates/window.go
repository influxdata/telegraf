package request_aggregates

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"sort"
	"strings"
	"time"
)

const (
	MeasurementTime        = "request_aggregates_total"
	MeasurementTimeFail    = "request_aggregates_fail"
	MeasurementTimeSuccess = "request_aggregates_success"
	FieldTimeRequests      = "requests"
	FieldTimeMin           = "time_min"
	FieldTimeMax           = "time_max"
	FieldTimeMean          = "time_mean"
	FieldTimePerc          = "time_perc_"

	MeasurementThroughput = "request_aggregates_throughput"
	FieldThroughputTotal  = "requests_total"
	FieldThroughputFailed = "requests_failed"
)

type Window interface {
	Aggregate() ([]telegraf.Metric, error)
	Add(request *Request) error
	Start() time.Time
	End() time.Time
}

type TimeWindow struct {
	StartTime    time.Time
	EndTime      time.Time
	TimesTotal   []float64
	TimesSuccess []float64
	TimesFail    []float64
	Percentiles  []float32
	OnlyTotal    bool
}

type ThroughputWindow struct {
	StartTime     time.Time
	EndTime       time.Time
	RequestsTotal int64
	RequestsFail  int64
}

func (tw *TimeWindow) Aggregate() ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 3)

	var err error
	metrics[0], err = aggregateTimes(MeasurementTime, tw.TimesTotal, tw.Percentiles, tw.EndTime)
	if err != nil {
		return metrics, err
	}
	if !tw.OnlyTotal {
		metrics[1], err = aggregateTimes(MeasurementTimeFail, tw.TimesFail, tw.Percentiles, tw.EndTime)
		if err != nil {
			return metrics, err
		}
		metrics[2], err = aggregateTimes(MeasurementTimeSuccess, tw.TimesSuccess, tw.Percentiles, tw.EndTime)
	} else {
		metrics = metrics[:1]
	}

	return metrics, err
}

func (tw *TimeWindow) Add(request *Request) error {
	tw.TimesTotal = append(tw.TimesTotal, request.Time)
	if !tw.OnlyTotal {
		if request.Failure {
			tw.TimesFail = append(tw.TimesFail, request.Time)
		} else {
			tw.TimesSuccess = append(tw.TimesSuccess, request.Time)
		}
	}
	return nil
}

func (tw *TimeWindow) Start() time.Time {
	return tw.StartTime
}

func (tw *TimeWindow) End() time.Time {
	return tw.EndTime
}

func (tw *ThroughputWindow) Aggregate() ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 1)

	metric, err := telegraf.NewMetric(MeasurementThroughput, nil, map[string]interface{}{
		FieldThroughputTotal:  tw.RequestsTotal,
		FieldThroughputFailed: tw.RequestsFail}, tw.EndTime)
	metrics[0] = metric

	return metrics, err
}

func (tw *ThroughputWindow) Add(request *Request) error {
	tw.RequestsTotal++
	if request.Failure {
		tw.RequestsFail++
	}
	return nil
}

func (tw *ThroughputWindow) Start() time.Time {
	return tw.StartTime
}

func (tw *ThroughputWindow) End() time.Time {
	return tw.EndTime
}

// Produces a metric with the aggregates for the given times and percentiles
func aggregateTimes(name string, times []float64, percentiles []float32, endTime time.Time) (telegraf.Metric, error) {
	sort.Float64s(times)

	fields := map[string]interface{}{FieldTimeRequests: len(times)}
	if len(times) > 0 {
		fields[FieldTimeMin] = times[0]
		fields[FieldTimeMax] = times[len(times)-1]
		totalSum := float64(0)
		for _, time := range times {
			totalSum += time
		}
		fields[FieldTimeMean] = totalSum / float64(len(times))

		for _, perc := range percentiles {
			i := int(float64(len(times)) * float64(perc) / float64(100))
			if i < 0 {
				i = 0
			}
			fields[FieldTimePerc+strings.Replace(fmt.Sprintf("%v", perc), ".", "_", -1)] = times[i]
		}
	}

	return telegraf.NewMetric(name, nil, fields, endTime)
}
