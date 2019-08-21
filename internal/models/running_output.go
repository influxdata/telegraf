package models

import (
	"log"
	"reflect"
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
	Alias  string
	Filter Filter

	FlushInterval     time.Duration
	MetricBufferLimit int
	MetricBatchSize   int
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
		buffer:            NewBuffer(conf.LogName(), bufferLimit),
		BatchReady:        make(chan time.Time, 1),
		Output:            output,
		Config:            conf,
		MetricBufferLimit: bufferLimit,
		MetricBatchSize:   batchSize,
		MetricsFiltered: selfstat.Register(
			"write",
			"metrics_filtered",
			map[string]string{"output": conf.Name, "alias": conf.Alias},
		),
		WriteTime: selfstat.RegisterTiming(
			"write",
			"write_time_ns",
			map[string]string{"output": conf.Name, "alias": conf.Alias},
		),
	}

	return ro
}

func (c *OutputConfig) LogName() string {
	if c.Alias == "" {
		return c.Name
	}
	return c.Name + "::" + c.Alias
}

func (ro *RunningOutput) Name() string {
	return "outputs." + ro.Config.Name
}

func (ro *RunningOutput) LogName() string {
	if ro.Config.Alias == "" {
		return ro.Name()
	}
	return ro.Name() + "::" + ro.Config.Alias
}

func (ro *RunningOutput) metricFiltered(metric telegraf.Metric) {
	ro.MetricsFiltered.Incr(1)
	metric.Drop()
}

func (r *RunningOutput) Init() error {
	setLogIfExist(r.Output, &Logger{
		Name: r.LogName(),
		Errs: selfstat.Register("write", "errors",
			map[string]string{"output": r.Config.Name, "alias": r.Config.Alias}),
	})

	if p, ok := r.Output.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}

	}
	return nil
}

func setLogIfExist(i interface{}, log telegraf.Logger) {
	valI := reflect.ValueOf(i)

	if valI.Type().Kind() != reflect.Ptr {
		valI = reflect.New(reflect.TypeOf(i))
	}

	field := valI.Elem().FieldByName("Log")
	if !field.IsValid() {
		return
	}

	switch field.Type().String() {
	case "telegraf.Logger":
		if field.CanSet() {
			field.Set(reflect.ValueOf(log))
		}
	}

	return
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

	dropped := ro.buffer.Add(metric)
	atomic.AddInt64(&ro.droppedMetrics, int64(dropped))

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
		log.Printf("E! [%s] Error closing output: %v", ro.LogName(), err)
	}
}

func (ro *RunningOutput) write(metrics []telegraf.Metric) error {
	dropped := atomic.LoadInt64(&ro.droppedMetrics)
	if dropped > 0 {
		log.Printf("W! [%s] Metric buffer overflow; %d metrics have been dropped",
			ro.LogName(), dropped)
		atomic.StoreInt64(&ro.droppedMetrics, 0)
	}

	start := time.Now()
	err := ro.Output.Write(metrics)
	elapsed := time.Since(start)
	ro.WriteTime.Incr(elapsed.Nanoseconds())

	if err == nil {
		log.Printf("D! [%s] wrote batch of %d metrics in %s",
			ro.LogName(), len(metrics), elapsed)
	}
	return err
}

func (ro *RunningOutput) LogBufferStatus() {
	nBuffer := ro.buffer.Len()
	log.Printf("D! [%s] buffer fullness: %d / %d metrics",
		ro.LogName(), nBuffer, ro.MetricBufferLimit)
}
