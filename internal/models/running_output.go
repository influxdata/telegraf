package models

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/buffer"
)

const (
	// Default size of metrics batch size.
	DEFAULT_METRIC_BATCH_SIZE = 1000

	// Default number of metrics kept. It should be a multiple of batch size.
	DEFAULT_METRIC_BUFFER_LIMIT = 10000
)

// RunningOutput contains the output configuration
type RunningOutput struct {
	Name              string
	Output            telegraf.Output
	Config            *OutputConfig
	Quiet             bool
	MetricBufferLimit int
	MetricBatchSize   int

	metrics     *buffer.Buffer
	failMetrics *buffer.Buffer
}

func NewRunningOutput(
	name string,
	output telegraf.Output,
	conf *OutputConfig,
	batchSize int,
	bufferLimit int,
) *RunningOutput {
	if bufferLimit == 0 {
		bufferLimit = DEFAULT_METRIC_BUFFER_LIMIT
	}
	if batchSize == 0 {
		batchSize = DEFAULT_METRIC_BATCH_SIZE
	}
	ro := &RunningOutput{
		Name:              name,
		metrics:           buffer.NewBuffer(batchSize),
		failMetrics:       buffer.NewBuffer(bufferLimit),
		Output:            output,
		Config:            conf,
		MetricBufferLimit: bufferLimit,
		MetricBatchSize:   batchSize,
	}
	return ro
}

// AddMetric adds a metric to the output. This function can also write cached
// points if FlushBufferWhenFull is true.
func (ro *RunningOutput) AddMetric(metric telegraf.Metric) {
	// Filter any tagexclude/taginclude parameters before adding metric
	if ro.Config.Filter.IsActive() {
		// In order to filter out tags, we need to create a new metric, since
		// metrics are immutable once created.
		name := metric.Name()
		tags := metric.Tags()
		fields := metric.Fields()
		t := metric.Time()
		if ok := ro.Config.Filter.Apply(name, fields, tags); !ok {
			return
		}
		// error is not possible if creating from another metric, so ignore.
		metric, _ = telegraf.NewMetric(name, tags, fields, t)
	}

	ro.metrics.Add(metric)
	if ro.metrics.Len() == ro.MetricBatchSize {
		batch := ro.metrics.Batch(ro.MetricBatchSize)
		err := ro.write(batch)
		if err != nil {
			ro.failMetrics.Add(batch...)
		}
	}
}

// Write writes all cached points to this output.
func (ro *RunningOutput) Write() error {
	if !ro.Quiet {
		log.Printf("I! Output [%s] buffer fullness: %d / %d metrics. "+
			"Total gathered metrics: %d. Total dropped metrics: %d.",
			ro.Name,
			ro.failMetrics.Len()+ro.metrics.Len(),
			ro.MetricBufferLimit,
			ro.metrics.Total(),
			ro.metrics.Drops()+ro.failMetrics.Drops())
	}

	var err error
	if !ro.failMetrics.IsEmpty() {
		bufLen := ro.failMetrics.Len()
		// how many batches of failed writes we need to write.
		nBatches := bufLen/ro.MetricBatchSize + 1
		batchSize := ro.MetricBatchSize

		for i := 0; i < nBatches; i++ {
			// If it's the last batch, only grab the metrics that have not had
			// a write attempt already (this is primarily to preserve order).
			if i == nBatches-1 {
				batchSize = bufLen % ro.MetricBatchSize
			}
			batch := ro.failMetrics.Batch(batchSize)
			// If we've already failed previous writes, don't bother trying to
			// write to this output again. We are not exiting the loop just so
			// that we can rotate the metrics to preserve order.
			if err == nil {
				err = ro.write(batch)
			}
			if err != nil {
				ro.failMetrics.Add(batch...)
			}
		}
	}

	batch := ro.metrics.Batch(ro.MetricBatchSize)
	// see comment above about not trying to write to an already failed output.
	// if ro.failMetrics is empty then err will always be nil at this point.
	if err == nil {
		err = ro.write(batch)
	}
	if err != nil {
		ro.failMetrics.Add(batch...)
		return err
	}
	return nil
}

func (ro *RunningOutput) write(metrics []telegraf.Metric) error {
	if metrics == nil || len(metrics) == 0 {
		return nil
	}
	start := time.Now()
	err := ro.Output.Write(metrics)
	elapsed := time.Since(start)
	if err == nil {
		if !ro.Quiet {
			log.Printf("I! Output [%s] wrote batch of %d metrics in %s\n",
				ro.Name, len(metrics), elapsed)
		}
	}
	return err
}

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter
}
