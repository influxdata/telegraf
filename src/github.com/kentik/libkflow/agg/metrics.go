package agg

import "github.com/kentik/go-metrics"

type Metrics struct {
	TotalFlowsIn   metrics.Meter
	TotalFlowsOut  metrics.Meter
	OrigSampleRate metrics.Histogram
	NewSampleRate  metrics.Histogram
	RateLimitDrops metrics.Meter
}
