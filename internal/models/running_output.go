package models

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	// Default size of metrics batch size.
	DEFAULT_METRIC_BATCH_SIZE = 1000

	// Default number of metrics kept. It should be a multiple of batch size.
	DEFAULT_METRIC_BUFFER_LIMIT = 10000
)

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter

	FlushInterval     time.Duration
	MetricBufferLimit int
	MetricBatchSize   int
}

// RunningOutput contains the output configuration
type RunningOutput struct {
	// Must be 64-bit aligned
	newMetricsCount int64

	Name              string
	Output            telegraf.Output
	Config            *OutputConfig
	MetricBufferLimit int
	MetricBatchSize   int

	MetricsFiltered selfstat.Stat
	WriteTime       selfstat.Stat

	BatchReady chan time.Time

	buffer *Buffer

	aggMutex sync.Mutex
}

func NewRunningOutput(
	name string,
	output telegraf.Output,
	conf *OutputConfig,
	batchSize int,
	bufferLimit int,
) *RunningOutput {
	if conf.MetricBufferLimit > 0 {
		bufferLimit = conf.MetricBufferLimit
	}
	if bufferLimit == 0 {
		bufferLimit = DEFAULT_METRIC_BUFFER_LIMIT
	}
	if conf.MetricBatchSize > 0 {
		batchSize = conf.MetricBatchSize
	}
	if batchSize == 0 {
		batchSize = DEFAULT_METRIC_BATCH_SIZE
	}
	ro := &RunningOutput{
		Name:              name,
		buffer:            NewBuffer(name, bufferLimit),
		BatchReady:        make(chan time.Time, 1),
		Output:            output,
		Config:            conf,
		MetricBufferLimit: bufferLimit,
		MetricBatchSize:   batchSize,
		MetricsFiltered: selfstat.Register(
			"write",
			"metrics_filtered",
			map[string]string{"output": name},
		),
		WriteTime: selfstat.RegisterTiming(
			"write",
			"write_time_ns",
			map[string]string{"output": name},
		),
	}

	return ro
}

func (ro *RunningOutput) metricFiltered(metric telegraf.Metric) {
	ro.MetricsFiltered.Incr(1)
	metric.Drop()
}

// AddMetric adds a metric to the output.
//
// Takes ownership of metric
func (ro *RunningOutput) AddMetric(metric telegraf.Metric) {
	if ok := ro.Config.Filter.Select(metric); !ok {
		ro.metricFiltered(metric)
		return
	}

	ro.Config.Filter.Modify(metric)
	if len(metric.FieldList()) == 0 {
		ro.metricFiltered(metric)
		return
	}

	if output, ok := ro.Output.(telegraf.AggregatingOutput); ok {
		ro.aggMutex.Lock()
		output.Add(metric)
		ro.aggMutex.Unlock()
		return
	}

	ro.buffer.Add(metric)

	count := atomic.AddInt64(&ro.newMetricsCount, 1)
	if count == int64(ro.MetricBatchSize) {
		atomic.StoreInt64(&ro.newMetricsCount, 0)
		select {
		case ro.BatchReady <- time.Now():
		default:
		}
	}
}

// Write writes all metrics to the output, stopping when all have been sent on
// or error.
func (ro *RunningOutput) Write() error {
	if output, ok := ro.Output.(telegraf.AggregatingOutput); ok {
		ro.aggMutex.Lock()
		metrics := output.Push()
		ro.buffer.Add(metrics...)
		output.Reset()
		ro.aggMutex.Unlock()
	}

	atomic.StoreInt64(&ro.newMetricsCount, 0)

	// Only process the metrics in the buffer now.  Metrics added while we are
	// writing will be sent on the next call.
	nBuffer := ro.buffer.Len()
	nBatches := nBuffer/ro.MetricBatchSize + 1
	for i := 0; i < nBatches; i++ {
		batch := ro.buffer.Batch(ro.MetricBatchSize)
		if len(batch) == 0 {
			break
		}

		err := ro.write(batch)
		if err != nil {
			ro.buffer.Reject(batch)
			return err
		}
		ro.buffer.Accept(batch)
	}
	return nil
}

// WriteBatch writes a single batch of metrics to the output.
func (ro *RunningOutput) WriteBatch() error {
	batch := ro.buffer.Batch(ro.MetricBatchSize)
	if len(batch) == 0 {
		return nil
	}

	err := ro.write(batch)
	if err != nil {
		ro.buffer.Reject(batch)
		return err
	}
	ro.buffer.Accept(batch)

	return nil
}

func (ro *RunningOutput) Close() {
	err := ro.Output.Close()
	if err != nil {
		log.Printf("E! [outputs.%s] Error closing output: %v", ro.Name, err)
	}
}

func (ro *RunningOutput) write(metrics []telegraf.Metric) error {
	start := time.Now()
	err := ro.Output.Write(metrics)
	elapsed := time.Since(start)
	ro.WriteTime.Incr(elapsed.Nanoseconds())

	if err == nil {
		log.Printf("D! [outputs.%s] wrote batch of %d metrics in %s\n",
			ro.Name, len(metrics), elapsed)
	}
	return err
}

func (ro *RunningOutput) LogBufferStatus() {
	nBuffer := ro.buffer.Len()
	log.Printf("D! [outputs.%s] buffer fullness: %d / %d metrics. ",
		ro.Name, nBuffer, ro.MetricBufferLimit)
}
