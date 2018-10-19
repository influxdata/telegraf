package models

import (
	"log"
	"sync"
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
}

// RunningOutput contains the output configuration
type RunningOutput struct {
	Name              string
	Output            telegraf.Output
	Config            *OutputConfig
	MetricBufferLimit int
	MetricBatchSize   int

	MetricsFiltered selfstat.Stat
	MetricsWritten  selfstat.Stat
	BufferSize      selfstat.Stat
	BufferLimit     selfstat.Stat
	WriteTime       selfstat.Stat

	batch      []telegraf.Metric
	buffer     *Buffer
	BatchReady chan time.Time

	aggMutex sync.Mutex
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
		batch:             make([]telegraf.Metric, 0, batchSize),
		buffer:            NewBuffer(bufferLimit),
		BatchReady:        make(chan time.Time, 1),
		Output:            output,
		Config:            conf,
		MetricBufferLimit: bufferLimit,
		MetricBatchSize:   batchSize,
		MetricsWritten: selfstat.Register(
			"write",
			"metrics_written",
			map[string]string{"output": name},
		),
		MetricsFiltered: selfstat.Register(
			"write",
			"metrics_filtered",
			map[string]string{"output": name},
		),
		BufferSize: selfstat.Register(
			"write",
			"buffer_size",
			map[string]string{"output": name},
		),
		BufferLimit: selfstat.Register(
			"write",
			"buffer_limit",
			map[string]string{"output": name},
		),
		WriteTime: selfstat.RegisterTiming(
			"write",
			"write_time_ns",
			map[string]string{"output": name},
		),
	}
	ro.BufferLimit.Set(int64(ro.MetricBufferLimit))
	return ro
}

func (ro *RunningOutput) metricFiltered(metric telegraf.Metric) {
	ro.MetricsFiltered.Incr(1)
	metric.Accept()
}

// AddMetric adds a metric to the output. This function can also write cached
// points if FlushBufferWhenFull is true.
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

	ro.batch = append(ro.batch, metric)
	if len(ro.batch) == ro.MetricBatchSize {
		ro.addBatchToBuffer()

		select {
		case ro.BatchReady <- time.Now():
		default:
		}
	}
}

// AddBatchToBuffer moves the metrics from the batch into the metric buffer.
func (ro *RunningOutput) addBatchToBuffer() {
	ro.buffer.Add(ro.batch...)
	ro.batch = ro.batch[:0]
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
	ro.addBatchToBuffer()

	nBuffer := ro.buffer.Len()

	ro.BufferSize.Set(int64(nBuffer))
	log.Printf("D! Output [%s] buffer fullness: %d / %d metrics. ",
		ro.Name, nBuffer, ro.MetricBufferLimit)

	// Only process the metrics in the buffer now.  Metrics added while we are
	// writing will be sent on the next call.
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

		for _, m := range batch {
			m.Accept()
		}

		ro.buffer.Ack()
	}
	return nil
}

// WriteBatch writes only the batch metrics to the output.
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

	for _, m := range batch {
		m.Accept()
	}

	ro.buffer.Ack()
	return nil
}

func (ro *RunningOutput) write(metrics []telegraf.Metric) error {
	start := time.Now()
	err := ro.Output.Write(metrics)
	elapsed := time.Since(start)
	ro.WriteTime.Incr(elapsed.Nanoseconds())

	if err == nil {
		nMetrics := len(metrics)
		log.Printf("D! Output [%s] wrote batch of %d metrics in %s\n",
			ro.Name, nMetrics, elapsed)
		ro.MetricsWritten.Incr(int64(nMetrics))
	}
	return err
}
