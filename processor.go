package telegraf

import (
	"sync"
)

// Processor is a processor plugin interface for defining new inline processors.
// these are extremely efficient and should be used over StreamingProcessor if
// you do not need asynchronous metric writes.
type Processor interface {
	PluginDescriber

	// Apply the filter to the given metric.
	Apply(in ...Metric) []Metric
}

// StreamingProcessor is a processor that can take in a stream of messages
type StreamingProcessor interface {
	PluginDescriber
	StreamingStartStopper
}

// StreamingStartStopper is the main implementation for building streaming processors.
type StreamingStartStopper interface {
	// Start is called when the processor should start.
	// The StreamingAccumulator may be retained and used until Stop returns.
	// Start is only called once per plugin instance, and never in parallel.
	// Start should exit when acc.IsStreamClosed() returns true.
	// Start should not exit until the processor is ready to quit and the stream
	// is empty.
	Start(acc StreamingAccumulator) error

	// Stop is called when the plugin should stop processing.
	// at this point no new metrics will be coming in to the StreamingAccumulator,
	// you can finish up processing the remaining metrics until IsStreamClosed()
	// returns true. Wait for this to happen, then return from Stop. After Stop()
	// returns, the reference to the StreamingAccumulator should not be used.
	Stop()
}

// NewProcessorWrapper turns a standard processor into a streaming processor
func NewProcessorWrapper(p Processor) StreamingProcessor {
	spw := &streamingProcessorWrapper{
		processor: p,
		wg:        sync.WaitGroup{},
	}
	spw.wg.Add(1)
	return spw
}

type streamingProcessorWrapper struct {
	wg        sync.WaitGroup
	processor Processor
}

func (spw *streamingProcessorWrapper) SampleConfig() string {
	return spw.processor.SampleConfig()
}

func (spw *streamingProcessorWrapper) Description() string {
	return spw.processor.Description()
}

func (spw *streamingProcessorWrapper) Start(acc StreamingAccumulator) error {
	defer spw.wg.Done()
	for {
		m := acc.GetNextMetric()
		if m == nil {
			if acc.IsStreamClosed() {
				return nil
			}
			continue
		}
		for _, metric := range spw.processor.Apply(m) {
			acc.PassMetric(metric)
		}
	}
}

func (spw *streamingProcessorWrapper) Stop() {
	spw.wg.Wait()
}
