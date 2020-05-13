package processors

import "github.com/influxdata/telegraf"

type Creator func() telegraf.Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	Processors[name] = creator
}

type StreamingCreator func() telegraf.StreamingProcessor

var StreamingProcessors = map[string]StreamingCreator{}

func AddStreaming(name string, creator StreamingCreator) {
	StreamingProcessors[name] = creator
}
