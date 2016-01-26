package models

import (
	"log"
	"time"

	emodels "github.com/influxdata/telegraf/models"
)

const DEFAULT_POINT_BUFFER_LIMIT = 10000

type RunningOutput struct {
	Name             string
	Output           emodels.Output
	Config           *OutputConfig
	Quiet            bool
	PointBufferLimit int

	metrics          []emodels.Metric
	overwriteCounter int
}

func NewRunningOutput(
	name string,
	output emodels.Output,
	conf *OutputConfig,
) *RunningOutput {
	ro := &RunningOutput{
		Name:             name,
		metrics:          make([]emodels.Metric, 0),
		Output:           output,
		Config:           conf,
		PointBufferLimit: DEFAULT_POINT_BUFFER_LIMIT,
	}
	return ro
}

func (ro *RunningOutput) AddPoint(metric emodels.Metric) {
	if ro.Config.Filter.IsActive {
		if !ro.Config.Filter.ShouldPointPass(metric) {
			return
		}
	}

	if len(ro.metrics) < ro.PointBufferLimit {
		ro.metrics = append(ro.metrics, metric)
	} else {
		if ro.overwriteCounter == len(ro.metrics) {
			ro.overwriteCounter = 0
		}
		ro.metrics[ro.overwriteCounter] = metric
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
		ro.metrics = make([]emodels.Metric, 0)
		ro.overwriteCounter = 0
	}
	return err
}

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter
}
