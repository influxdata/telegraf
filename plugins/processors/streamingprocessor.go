package processors

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
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
	acc       telegraf.Accumulator
	Log       telegraf.Logger
}

func (sp *streamingProcessor) SampleConfig() string {
	return sp.processor.SampleConfig()
}

func (sp *streamingProcessor) Description() string {
	return sp.processor.Description()
}

func (sp *streamingProcessor) Start(acc telegraf.Accumulator) error {
	sp.acc = acc
	return nil
}

func (sp *streamingProcessor) Add(m telegraf.Metric, acc telegraf.Accumulator) error {
	for _, m := range sp.processor.Apply(m) {
		acc.AddMetric(m)
	}
	return nil
}

func (sp *streamingProcessor) Stop() error {
	return nil
}

// Make the streamingProcessor of type Initializer to be able
// to call the Init method of the wrapped processor if
// needed
func (sp *streamingProcessor) Init() error {
	models.SetLoggerOnPlugin(sp.processor, sp.Log)
	if p, ok := sp.processor.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

// Unwrap lets you retrieve the original telegraf.Processor from the
// StreamingProcessor. This is necessary because the toml Unmarshaller won't
// look inside composed types.
func (sp *streamingProcessor) Unwrap() telegraf.Processor {
	return sp.processor
}
