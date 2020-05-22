package processors

import (
	"github.com/influxdata/telegraf"
)

// NewStreamingProcessorFromProcessor is a converter that turns a standard
// processor into a streaming processor
func NewStreamingProcessorFromProcessor(p telegraf.Processor) telegraf.StreamingProcessor {
	sp := &streamingProcessor{
		processor: p,
	}
	return sp
}

type streamingProcessor struct {
	processor telegraf.Processor
	acc       telegraf.MetricStreamAccumulator
}

func (sp *streamingProcessor) SampleConfig() string {
	return sp.processor.SampleConfig()
}

func (sp *streamingProcessor) Description() string {
	return sp.processor.Description()
}

func (sp *streamingProcessor) Start(acc telegraf.MetricStreamAccumulator) error {
	sp.acc = acc
	return nil
}
func (sp *streamingProcessor) Add(m telegraf.Metric) {
	for _, m := range sp.processor.Apply(m) {
		sp.acc.PassMetric(m)
	}
}
func (sp *streamingProcessor) Stop() error {
	return nil
}
