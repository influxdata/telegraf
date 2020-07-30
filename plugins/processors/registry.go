package processors

import "github.com/influxdata/telegraf"

type Creator func() telegraf.Processor
type StreamingCreator func() telegraf.StreamingProcessor

// all processors are streaming processors.
// telegraf.Processor processors are upgraded to telegraf.StreamingProcessor
var Processors = map[string]StreamingCreator{}

// Add adds a telegraf.Processor processor
func Add(name string, creator Creator) {
	Processors[name] = upgradeToStreamingProcessor(creator)
}

// AddStreaming adds a telegraf.StreamingProcessor streaming processor
func AddStreaming(name string, creator StreamingCreator) {
	Processors[name] = creator
}

func upgradeToStreamingProcessor(oldCreator Creator) StreamingCreator {
	return func() telegraf.StreamingProcessor {
		return NewStreamingProcessorFromProcessor(oldCreator())
	}
}
