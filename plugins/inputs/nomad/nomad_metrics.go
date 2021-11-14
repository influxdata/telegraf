package nomad

import (
	"time"
)

type MetricsSummary struct {
	Timestamp string
	Gauges    []GaugeValue
	Points    []PointValue
	Counters  []SampledValue
	Samples   []SampledValue
}

type GaugeValue struct {
	Name  string
	Hash  string `json:"-"`
	Value float32

	Labels        []Label           `json:"-"`
	DisplayLabels map[string]string `json:"Labels"`
}

type PointValue struct {
	Name   string
	Points []float32
}

type SampledValue struct {
	Name string
	Hash string `json:"-"`
	*AggregateSample
	Mean   float64
	Stddev float64

	Labels        []Label           `json:"-"`
	DisplayLabels map[string]string `json:"Labels"`
}

type AggregateSample struct {
	Count       int
	Rate        float64
	Sum         float64
	SumSq       float64 `json:"-"`
	Min         float64
	Max         float64
	LastUpdated time.Time `json:"-"`
}

type Label struct {
	Name  string
	Value string
}
