package internal_models

import (
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

const (
	// Default number of metrics kept between flushes.
	DEFAULT_METRIC_BUFFER_LIMIT = 1000

	// Limit how many full metric buffers are kept due to failed writes.
	FULL_METRIC_BUFFERS_LIMIT = 100
)

type RunningOutput struct {
	Name                string
	Output              telegraf.Output
	Config              *OutputConfig
	Quiet               bool
	MetricBufferLimit   int
	FlushBufferWhenFull bool

	metrics    []telegraf.Metric
	tmpmetrics map[int][]telegraf.Metric
	overwriteI int
	mapI       int

	sync.Mutex
}

func NewRunningOutput(
	name string,
	output telegraf.Output,
	conf *OutputConfig,
) *RunningOutput {
	ro := &RunningOutput{
		Name:              name,
		metrics:           make([]telegraf.Metric, 0),
		tmpmetrics:        make(map[int][]telegraf.Metric),
		Output:            output,
		Config:            conf,
		MetricBufferLimit: DEFAULT_METRIC_BUFFER_LIMIT,
	}
	return ro
}

// AddMetric adds a metric to the output. This function can also write cached
// points if FlushBufferWhenFull is true.
func (ro *RunningOutput) AddMetric(metric telegraf.Metric) {
	if ro.Config.Filter.IsActive {
		if !ro.Config.Filter.ShouldMetricPass(metric) {
			return
		}
	}
	ro.Lock()
	defer ro.Unlock()

	if len(ro.metrics) < ro.MetricBufferLimit {
		ro.metrics = append(ro.metrics, metric)
	} else {
		if ro.FlushBufferWhenFull {
			ro.metrics = append(ro.metrics, metric)
			tmpmetrics := make([]telegraf.Metric, len(ro.metrics))
			copy(tmpmetrics, ro.metrics)
			ro.metrics = make([]telegraf.Metric, 0)
			err := ro.write(tmpmetrics)
			if err != nil {
				log.Printf("ERROR writing full metric buffer to output %s, %s",
					ro.Name, err)
				if len(ro.tmpmetrics) == FULL_METRIC_BUFFERS_LIMIT {
					ro.mapI = 0
					// overwrite one
					ro.tmpmetrics[ro.mapI] = tmpmetrics
					ro.mapI++
				} else {
					ro.tmpmetrics[ro.mapI] = tmpmetrics
					ro.mapI++
				}
			}
		} else {
			if ro.overwriteI == 0 {
				log.Printf("WARNING: overwriting cached metrics, you may want to " +
					"increase the metric_buffer_limit setting in your [agent] " +
					"config if you do not wish to overwrite metrics.\n")
			}
			if ro.overwriteI == len(ro.metrics) {
				ro.overwriteI = 0
			}
			ro.metrics[ro.overwriteI] = metric
			ro.overwriteI++
		}
	}
}

// Write writes all cached points to this output.
func (ro *RunningOutput) Write() error {
	ro.Lock()
	defer ro.Unlock()
	err := ro.write(ro.metrics)
	if err != nil {
		return err
	} else {
		ro.metrics = make([]telegraf.Metric, 0)
		ro.overwriteI = 0
	}

	// Write any cached metric buffers that failed previously
	for i, tmpmetrics := range ro.tmpmetrics {
		if err := ro.write(tmpmetrics); err != nil {
			return err
		} else {
			delete(ro.tmpmetrics, i)
		}
	}

	return nil
}

func (ro *RunningOutput) write(metrics []telegraf.Metric) error {
	start := time.Now()
	err := ro.Output.Write(metrics)
	elapsed := time.Since(start)
	if err == nil {
		if !ro.Quiet {
			log.Printf("Wrote %d metrics to output %s in %s\n",
				len(metrics), ro.Name, elapsed)
		}
	}
	return err
}

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter
}
