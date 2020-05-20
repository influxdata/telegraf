package processors

import (
	"sync"

	"github.com/influxdata/telegraf"
)

// NewStreamingProcessorFromProcessor is a converter that turns a standard
// processor into a streaming processor
func NewStreamingProcessorFromProcessor(p telegraf.Processor) telegraf.StreamingProcessor {
	sp := &streamingProcessor{
		processor: p,
		wg:        sync.WaitGroup{},
	}
	sp.wg.Add(1)
	return sp
}

type streamingProcessor struct {
	wg        sync.WaitGroup
	processor telegraf.Processor
}

func (sp *streamingProcessor) SampleConfig() string {
	return sp.processor.SampleConfig()
}

func (sp *streamingProcessor) Description() string {
	return sp.processor.Description()
}

func (sp *streamingProcessor) Start(acc telegraf.MetricStream) error {
	defer sp.wg.Done()
	for {
		m := acc.GetNextMetric()
		if m == nil {
			if acc.IsStreamClosed() {
				return nil
			}
			continue
		}
		for _, metric := range sp.processor.Apply(m) {
			acc.PassMetric(metric)
		}
	}
}

func (sp *streamingProcessor) Stop() {
	sp.wg.Wait()
}
