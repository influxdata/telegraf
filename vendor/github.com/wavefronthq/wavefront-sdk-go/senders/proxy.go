package senders

import (
	"fmt"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

type proxySender struct {
	metricHandler internal.ConnectionHandler
	histoHandler  internal.ConnectionHandler
	spanHandler   internal.ConnectionHandler
	defaultSource string
}

// Creates and returns a Wavefront Proxy Sender instance
func NewProxySender(cfg *ProxyConfiguration) (Sender, error) {
	if cfg.FlushIntervalSeconds == 0 {
		cfg.FlushIntervalSeconds = defaultProxyFlushInterval
	}

	var metricHandler internal.ConnectionHandler
	if cfg.MetricsPort != 0 {
		metricHandler = makeConnHandler(cfg.Host, cfg.MetricsPort, cfg.FlushIntervalSeconds)
	}

	var histoHandler internal.ConnectionHandler
	if cfg.DistributionPort != 0 {
		histoHandler = makeConnHandler(cfg.Host, cfg.DistributionPort, cfg.FlushIntervalSeconds)
	}

	var spanHandler internal.ConnectionHandler
	if cfg.TracingPort != 0 {
		spanHandler = makeConnHandler(cfg.Host, cfg.TracingPort, cfg.FlushIntervalSeconds)
	}

	if metricHandler == nil && histoHandler == nil && spanHandler == nil {
		return nil, fmt.Errorf("at least one proxy port should be enabled")
	}

	sender := &proxySender{
		defaultSource: internal.GetHostname("wavefront_proxy_sender"),
		metricHandler: metricHandler,
		histoHandler:  histoHandler,
		spanHandler:   spanHandler,
	}
	sender.Start()
	return sender, nil
}

func makeConnHandler(host string, port, interval int) internal.ConnectionHandler {
	addr := fmt.Sprintf("%s:%d", host, port)
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	return internal.NewProxyConnectionHandler(addr, ticker)
}

func (sender *proxySender) Start() {
	if sender.metricHandler != nil {
		sender.metricHandler.Start()
	}
	if sender.histoHandler != nil {
		sender.histoHandler.Start()
	}
	if sender.spanHandler != nil {
		sender.spanHandler.Start()
	}
}

func (sender *proxySender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	if sender.metricHandler == nil {
		return fmt.Errorf("proxy metrics port not provided, cannot send metric data")
	}

	if !sender.metricHandler.Connected() {
		err := sender.metricHandler.Connect()
		if err != nil {
			return err
		}
	}

	line, err := MetricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.metricHandler.SendData(line)
	return err
}

func (sender *proxySender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	return sender.SendMetric(name, value, 0, source, tags)
}

func (sender *proxySender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	if sender.histoHandler == nil {
		return fmt.Errorf("proxy distribution port not provided, cannot send distribution data")
	}

	if !sender.histoHandler.Connected() {
		err := sender.histoHandler.Connect()
		if err != nil {
			return err
		}
	}

	line, err := HistoLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.histoHandler.SendData(line)
	return err
}

func (sender *proxySender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	if sender.spanHandler == nil {
		return fmt.Errorf("proxy tracing port not provided, cannot send span data")
	}

	if !sender.spanHandler.Connected() {
		err := sender.spanHandler.Connect()
		if err != nil {
			return err
		}
	}

	line, err := SpanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.spanHandler.SendData(line)
	return err
}

func (sender *proxySender) Close() {
	if sender.metricHandler != nil {
		sender.metricHandler.Close()
	}
	if sender.histoHandler != nil {
		sender.histoHandler.Close()
	}
	if sender.spanHandler != nil {
		sender.spanHandler.Close()
	}
}

func (sender *proxySender) Flush() error {
	errStr := ""
	if sender.metricHandler != nil {
		err := sender.metricHandler.Flush()
		if err != nil {
			errStr = errStr + err.Error() + "\n"
		}
	}
	if sender.histoHandler != nil {
		err := sender.histoHandler.Flush()
		if err != nil {
			errStr = errStr + err.Error() + "\n"
		}
	}
	if sender.spanHandler != nil {
		err := sender.spanHandler.Flush()
		if err != nil {
			errStr = errStr + err.Error()
		}
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (sender *proxySender) GetFailureCount() int64 {
	var failures int64
	if sender.metricHandler != nil {
		failures += sender.metricHandler.GetFailureCount()
	}
	if sender.histoHandler != nil {
		failures += sender.histoHandler.GetFailureCount()
	}
	if sender.histoHandler != nil {
		failures += sender.histoHandler.GetFailureCount()
	}
	return failures
}
