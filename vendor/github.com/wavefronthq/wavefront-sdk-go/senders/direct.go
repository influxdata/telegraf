package senders

import (
	"fmt"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

const (
	metricFormat    = "wavefront"
	histogramFormat = "histogram"
	traceFormat     = "trace"
)

type directSender struct {
	reporter      internal.Reporter
	defaultSource string
	mtx           sync.Mutex
	pointHandler  *internal.LineHandler
	histoHandler  *internal.LineHandler
	spanHandler   *internal.LineHandler
}

// Creates and returns a Wavefront Direct Ingestion Sender instance
func NewDirectSender(cfg *DirectConfiguration) (Sender, error) {
	if cfg.Server == "" || cfg.Token == "" {
		return nil, fmt.Errorf("server and token cannot be empty")
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.MaxBufferSize == 0 {
		cfg.MaxBufferSize = defaultBufferSize
	}
	if cfg.FlushIntervalSeconds == 0 {
		cfg.FlushIntervalSeconds = defaultFlushInterval
	}
	reporter := internal.NewDirectReporter(cfg.Server, cfg.Token)
	sender := &directSender{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
		pointHandler:  makeLineHandler(reporter, cfg, metricFormat),
		histoHandler:  makeLineHandler(reporter, cfg, histogramFormat),
		spanHandler:   makeLineHandler(reporter, cfg, traceFormat),
	}
	sender.Start()
	return sender, nil
}

func makeLineHandler(reporter internal.Reporter, cfg *DirectConfiguration, format string) *internal.LineHandler {
	return &internal.LineHandler{
		Reporter:      reporter,
		BatchSize:     cfg.BatchSize,
		MaxBufferSize: cfg.MaxBufferSize,
		FlushTicker:   time.NewTicker(time.Second * time.Duration(cfg.FlushIntervalSeconds)),
		Format:        format,
	}
}

func (sender *directSender) Start() {
	sender.pointHandler.Start()
	sender.histoHandler.Start()
	sender.spanHandler.Start()
}

func (sender *directSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	line, err := MetricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.pointHandler.HandleLine(line)
}

func (sender *directSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	return sender.SendMetric(name, value, 0, source, tags)
}

func (sender *directSender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	line, err := HistoLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.histoHandler.HandleLine(line)
}

func (sender *directSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	line, err := SpanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.spanHandler.HandleLine(line)
}

func (sender *directSender) Close() {
	sender.pointHandler.Stop()
	sender.histoHandler.Stop()
	sender.spanHandler.Stop()
}

func (sender *directSender) Flush() error {
	errStr := ""
	err := sender.pointHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.histoHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.spanHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (sender *directSender) GetFailureCount() int64 {
	return sender.pointHandler.GetFailureCount() +
		sender.histoHandler.GetFailureCount() +
		sender.spanHandler.GetFailureCount()
}
