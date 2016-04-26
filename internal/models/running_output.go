package internal_models

import (
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

const (

	// Default size of metrics batch size.
	DEFAULT_METRIC_BATCH_SIZE = 1000

	// Default number of metrics kept. It should be a multiple of batch size.
	DEFAULT_METRIC_BUFFER_LIMIT = 10000
)

// tmpmetrics point to batch of metrics ready to be wrote to output.
// readI point to the oldest batch of metrics (the first to sent to output). It
// may point to nil value if tmpmetrics is empty.
// writeI point to the next slot to buffer a batch of metrics is output fail to
// write.
type RunningOutput struct {
	Name                string
	Output              telegraf.Output
	Config              *OutputConfig
	Quiet               bool
	MetricBufferLimit   int
	MetricBatchSize     int
	FlushBufferWhenFull bool

	metrics    []telegraf.Metric
	tmpmetrics []([]telegraf.Metric)
	writeI     int
	readI      int

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
		Output:            output,
		Config:            conf,
		MetricBufferLimit: DEFAULT_METRIC_BUFFER_LIMIT,
		MetricBatchSize:   DEFAULT_METRIC_BATCH_SIZE,
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

	if ro.tmpmetrics == nil {
		size := ro.MetricBufferLimit / ro.MetricBatchSize
		// ro.metrics already contains one batch
		size = size - 1

		if size < 1 {
			size = 1
		}
		ro.tmpmetrics = make([]([]telegraf.Metric), size)
	}

	// Filter any tagexclude/taginclude parameters before adding metric
	if len(ro.Config.Filter.TagExclude) != 0 || len(ro.Config.Filter.TagInclude) != 0 {
		// In order to filter out tags, we need to create a new metric, since
		// metrics are immutable once created.
		tags := metric.Tags()
		fields := metric.Fields()
		t := metric.Time()
		name := metric.Name()
		ro.Config.Filter.FilterTags(tags)
		// error is not possible if creating from another metric, so ignore.
		metric, _ = telegraf.NewMetric(name, tags, fields, t)
	}

	if len(ro.metrics) < ro.MetricBatchSize {
		ro.metrics = append(ro.metrics, metric)
	} else {
		flushSuccess := true
		if ro.FlushBufferWhenFull {
			err := ro.write(ro.metrics)
			if err != nil {
				log.Printf("ERROR writing full metric buffer to output %s, %s",
					ro.Name, err)
				flushSuccess = false
			}
		} else {
			flushSuccess = false
		}
		if !flushSuccess {
			if ro.tmpmetrics[ro.writeI] != nil && ro.writeI == ro.readI {
				log.Printf("WARNING: overwriting cached metrics, you may want to " +
					"increase the metric_buffer_limit setting in your [agent] " +
					"config if you do not wish to overwrite metrics.\n")
				ro.readI = (ro.readI + 1) % cap(ro.tmpmetrics)
			}
			ro.tmpmetrics[ro.writeI] = ro.metrics
			ro.writeI = (ro.writeI + 1) % cap(ro.tmpmetrics)
		}
		ro.metrics = make([]telegraf.Metric, 0)
		ro.metrics = append(ro.metrics, metric)
	}
}

// Write writes all cached points to this output.
func (ro *RunningOutput) Write() error {
	ro.Lock()
	defer ro.Unlock()

	if ro.tmpmetrics == nil {
		size := ro.MetricBufferLimit / ro.MetricBatchSize
		// ro.metrics already contains one batch
		size = size - 1

		if size < 1 {
			size = 1
		}
		ro.tmpmetrics = make([]([]telegraf.Metric), size)
	}

	// Write any cached metric buffers before, as those metrics are the
	// oldest
	for ro.tmpmetrics[ro.readI] != nil {
		if err := ro.write(ro.tmpmetrics[ro.readI]); err != nil {
			return err
		} else {
			ro.tmpmetrics[ro.readI] = nil
			ro.readI = (ro.readI + 1) % cap(ro.tmpmetrics)
		}
	}

	err := ro.write(ro.metrics)
	if err != nil {
		return err
	} else {
		ro.metrics = make([]telegraf.Metric, 0)
	}

	return nil
}

func (ro *RunningOutput) write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
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
