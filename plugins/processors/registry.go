package processors

import "github.com/influxdata/telegraf"

// Creator is a function that returns a new instance of a telegraf.Processor.
type Creator func() telegraf.Processor

// StreamingCreator is a function that returns a new instance of a telegraf.StreamingProcessor.
type StreamingCreator func() telegraf.StreamingProcessor

// HasUnwrap indicates the presence of an Unwrap() function to retrieve the underlying telegraf.Processor.
type HasUnwrap interface {
	// Unwrap returns the underlying telegraf.Processor.
	Unwrap() telegraf.Processor
}

// Processors is a map of processor names to their respective creator functions.
// All processors are streaming processors.
// telegraf.Processor processors are upgraded to telegraf.StreamingProcessor.
var Processors = make(map[string]StreamingCreator)

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
