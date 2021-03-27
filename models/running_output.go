package models

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	// Default size of metrics batch size.
	DefaultMetricBatchSize = 1000

	// Default number of metrics kept. It should be a multiple of batch size.
	DefaultMetricBufferLimit = 10000
)

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Alias  string
	Filter Filter

	FlushInterval     time.Duration
	FlushJitter       time.Duration
	MetricBufferLimit int
	MetricBatchSize   int

	NameOverride string
	NamePrefix   string
	NameSuffix   string
}

// RunningOutput contains the output configuration
type RunningOutput struct {
	// Must be 64-bit aligned
	newMetricsCount int64
	droppedMetrics  int64

	Output            telegraf.Output
	Config            *OutputConfig
	MetricBufferLimit int
	MetricBatchSize   int

	MetricsFiltered selfstat.Stat
	WriteTime       selfstat.Stat

	BatchReady chan time.Time

	buffer *Buffer
	log    telegraf.Logger

	aggMutex sync.Mutex
}

func NewRunningOutput(
	output telegraf.Output,
	config *OutputConfig,
	batchSize int,
	bufferLimit int,
) *RunningOutput {
	tags := map[string]string{"output": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	writeErrorsRegister := selfstat.Register("write", "errors", tags)
	logger := NewLogger("outputs", config.Name, config.Alias)
	logger.OnErr(func() {
		writeErrorsRegister.Incr(1)
	})
	SetLoggerOnPlugin(output, logger)

	if config.MetricBufferLimit > 0 {
		bufferLimit = config.MetricBufferLimit
	}
	if bufferLimit == 0 {
		bufferLimit = DefaultMetricBufferLimit
	}
	if config.MetricBatchSize > 0 {
		batchSize = config.MetricBatchSize
	}
	if batchSize == 0 {
		batchSize = DefaultMetricBatchSize
	}

	ro := &RunningOutput{
		buffer:            NewBuffer(config.Name, config.Alias, bufferLimit),
		BatchReady:        make(chan time.Time, 1),
		Output:            output,
		Config:            config,
		MetricBufferLimit: bufferLimit,
		MetricBatchSize:   batchSize,
		MetricsFiltered: selfstat.Register(
			"write",
			"metrics_filtered",
			tags,
		),
		WriteTime: selfstat.RegisterTiming(
			"write",
			"write_time_ns",
			tags,
		),
		log: logger,
	}

	return ro
}

func (r *RunningOutput) LogName() string {
	return logName("outputs", r.Config.Name, r.Config.Alias)
}

func (r *RunningOutput) metricFiltered(metric telegraf.Metric) {
	r.MetricsFiltered.Incr(1)
	metric.Drop()
}

func (r *RunningOutput) Init() error {
	if p, ok := r.Output.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

// AddMetric adds a metric to the output.
//
// Takes ownership of metric
func (r *RunningOutput) AddMetric(metric telegraf.Metric) {
	if ok := r.Config.Filter.Select(metric); !ok {
		r.metricFiltered(metric)
		return
	}

	r.Config.Filter.Modify(metric)
	if len(metric.FieldList()) == 0 {
		r.metricFiltered(metric)
		return
	}

	if output, ok := r.Output.(telegraf.AggregatingOutput); ok {
		r.aggMutex.Lock()
		output.Add(metric)
		r.aggMutex.Unlock()
		return
	}

	if len(r.Config.NameOverride) > 0 {
		metric.SetName(r.Config.NameOverride)
	}

	if len(r.Config.NamePrefix) > 0 {
		metric.AddPrefix(r.Config.NamePrefix)
	}

	if len(r.Config.NameSuffix) > 0 {
		metric.AddSuffix(r.Config.NameSuffix)
	}

	dropped := r.buffer.Add(metric)
	atomic.AddInt64(&r.droppedMetrics, int64(dropped))

	count := atomic.AddInt64(&r.newMetricsCount, 1)
	if count == int64(r.MetricBatchSize) {
		atomic.StoreInt64(&r.newMetricsCount, 0)
		select {
		case r.BatchReady <- time.Now():
		default:
		}
	}
}

// Write writes all metrics to the output, stopping when all have been sent on
// or error.
func (r *RunningOutput) Write() error {
	if output, ok := r.Output.(telegraf.AggregatingOutput); ok {
		r.aggMutex.Lock()
		metrics := output.Push()
		r.buffer.Add(metrics...)
		output.Reset()
		r.aggMutex.Unlock()
	}

	atomic.StoreInt64(&r.newMetricsCount, 0)

	// Only process the metrics in the buffer now.  Metrics added while we are
	// writing will be sent on the next call.
	nBuffer := r.buffer.Len()
	nBatches := nBuffer/r.MetricBatchSize + 1
	for i := 0; i < nBatches; i++ {
		batch := r.buffer.Batch(r.MetricBatchSize)
		if len(batch) == 0 {
			break
		}

		err := r.write(batch)
		if err != nil {
			r.buffer.Reject(batch)
			return err
		}
		r.buffer.Accept(batch)
	}
	return nil
}

// WriteBatch writes a single batch of metrics to the output.
func (r *RunningOutput) WriteBatch() error {
	batch := r.buffer.Batch(r.MetricBatchSize)
	if len(batch) == 0 {
		return nil
	}

	err := r.write(batch)
	if err != nil {
		r.buffer.Reject(batch)
		return err
	}
	r.buffer.Accept(batch)

	return nil
}

// Close closes the output
func (r *RunningOutput) Close() {
	err := r.Output.Close()
	if err != nil {
		r.log.Errorf("Error closing output: %v", err)
	}
}

func (r *RunningOutput) write(metrics []telegraf.Metric) error {
	dropped := atomic.LoadInt64(&r.droppedMetrics)
	if dropped > 0 {
		r.log.Warnf("Metric buffer overflow; %d metrics have been dropped", dropped)
		atomic.StoreInt64(&r.droppedMetrics, 0)
	}

	start := time.Now()
	err := r.Output.Write(metrics)
	elapsed := time.Since(start)
	r.WriteTime.Incr(elapsed.Nanoseconds())

	if err == nil {
		r.log.Debugf("Wrote batch of %d metrics in %s", len(metrics), elapsed)
	}
	return err
}

func (r *RunningOutput) LogBufferStatus() {
	nBuffer := r.buffer.Len()
	r.log.Debugf("Buffer fullness: %d / %d metrics", nBuffer, r.MetricBufferLimit)
}

func (r *RunningOutput) Log() telegraf.Logger {
	return r.log
}

func (r *RunningOutput) BufferLength() int {
	return r.buffer.Len()
}
