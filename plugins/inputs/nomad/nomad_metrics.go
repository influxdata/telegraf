package nomad

import (
	"time"
)

type MetricsSummary struct {
	Timestamp string         `json:"timestamp"`
	Gauges    []GaugeValue   `json:"gauges"`
	Points    []PointValue   `json:"points"`
	Counters  []SampledValue `json:"counters"`
	Samples   []SampledValue `json:"samples"`
}

type GaugeValue struct {
	Name  string  `json:"name"`
	Hash  string  `json:"-"`
	Value float32 `json:"value"`

	Labels        []Label           `json:"-"`
	DisplayLabels map[string]string `json:"Labels"`
}

type PointValue struct {
	Name   string    `json:"name"`
	Points []float32 `json:"points"`
}

type SampledValue struct {
	Name string `json:"name"`
	Hash string `json:"-"`
	*AggregateSample
	Mean   float64 `json:"mean"`
	Stddev float64 `json:"stddev"`

	Labels        []Label           `json:"-"`
	DisplayLabels map[string]string `json:"Labels"`
}

type AggregateSample struct {
	Count       int       `json:"count"`
	Rate        float64   `json:"rate"`
	Sum         float64   `json:"sum"`
	SumSq       float64   `json:"-"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	LastUpdated time.Time `json:"-"`
}

type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
