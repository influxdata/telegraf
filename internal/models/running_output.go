package internal_models

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/influxdb/client/v2"
)

const DEFAULT_POINT_BUFFER_LIMIT = 10000

type RunningOutput struct {
	Name             string
	Output           telegraf.Output
	Config           *OutputConfig
	Quiet            bool
	PointBufferLimit int

	points           []*client.Point
	overwriteCounter int
}

func NewRunningOutput(
	name string,
	output telegraf.Output,
	conf *OutputConfig,
) *RunningOutput {
	ro := &RunningOutput{
		Name:             name,
		points:           make([]*client.Point, 0),
		Output:           output,
		Config:           conf,
		PointBufferLimit: DEFAULT_POINT_BUFFER_LIMIT,
	}
	return ro
}

func (ro *RunningOutput) AddPoint(point *client.Point) {
	if ro.Config.Filter.IsActive {
		if !ro.Config.Filter.ShouldPointPass(point) {
			return
		}
	}

	if len(ro.points) < ro.PointBufferLimit {
		ro.points = append(ro.points, point)
	} else {
		if ro.overwriteCounter == len(ro.points) {
			ro.overwriteCounter = 0
		}
		ro.points[ro.overwriteCounter] = point
		ro.overwriteCounter++
	}
}

func (ro *RunningOutput) Write() error {
	start := time.Now()
	err := ro.Output.Write(ro.points)
	elapsed := time.Since(start)
	if err == nil {
		if !ro.Quiet {
			log.Printf("Wrote %d metrics to output %s in %s\n",
				len(ro.points), ro.Name, elapsed)
		}
		ro.points = make([]*client.Point, 0)
		ro.overwriteCounter = 0
	}
	return err
}

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter
}
