package processors

import "github.com/influxdata/telegraf"

type Creator func() telegraf.Processor
type StreamingCreator func() telegraf.StreamingProcessor

// all processors are now streaming processors.
var Processors = map[string]StreamingCreator{}

// Add adds a processor
func Add(name string, creator Creator) {
	Processors[name] = upgradeToStreamingProcessor(creator)
}

// AddStreaming adds a streaming processor
func AddStreaming(name string, creator StreamingCreator) {
	Processors[name] = creator
}

func upgradeToStreamingProcessor(oldCreator Creator) StreamingCreator {
	return func() telegraf.StreamingProcessor {
		return telegraf.NewProcessorWrapper(oldCreator())
	}
}
