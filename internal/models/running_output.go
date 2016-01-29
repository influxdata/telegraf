package internal_models

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
)

const DEFAULT_POINT_BUFFER_LIMIT = 10000

type RunningOutput struct {
	Name             string
	Output           telegraf.Output
	Config           *OutputConfig
	Quiet            bool
	PointBufferLimit int

	metrics          []telegraf.Metric
	overwriteCounter int
}

func NewRunningOutput(
	name string,
	output telegraf.Output,
	conf *OutputConfig,
) *RunningOutput {
	ro := &RunningOutput{
		Name:             name,
		metrics:          make([]telegraf.Metric, 0),
		Output:           output,
		Config:           conf,
		PointBufferLimit: DEFAULT_POINT_BUFFER_LIMIT,
	}
	return ro
}

func (ro *RunningOutput) AddPoint(point telegraf.Metric) {
	if ro.Config.Filter.IsActive {
		if !ro.Config.Filter.ShouldMetricPass(point) {
			return
		}
	}

	if len(ro.metrics) < ro.PointBufferLimit {
		ro.metrics = append(ro.metrics, point)
	} else {
		log.Printf("WARNING: overwriting cached metrics, you may want to " +
			"increase the metric_buffer_limit setting in your [agent] config " +
			"if you do not wish to overwrite metrics.\n")
		if ro.overwriteCounter == len(ro.metrics) {
			ro.overwriteCounter = 0
		}
		ro.metrics[ro.overwriteCounter] = point
		ro.overwriteCounter++
	}
}

func (ro *RunningOutput) Write() error {
	start := time.Now()
	err := ro.Output.Write(ro.metrics)
	elapsed := time.Since(start)
	if err == nil {
		if !ro.Quiet {
			log.Printf("Wrote %d metrics to output %s in %s\n",
				len(ro.metrics), ro.Name, elapsed)
		}
		ro.metrics = make([]telegraf.Metric, 0)
		ro.overwriteCounter = 0
	}
	return err
}

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter
}
