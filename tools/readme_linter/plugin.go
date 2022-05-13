package main

import (
	"strings"
)

type plugin int

const (
	pluginNone plugin = iota
	pluginInput
	pluginOutput
	pluginProcessor
	pluginAggregator
	pluginParser
)

func guessPluginType(filename string) plugin {
	switch {
	case strings.Contains(filename, "plugins/inputs/"):
		return pluginInput
	case strings.Contains(filename, "plugins/outputs/"):
		return pluginOutput
	case strings.Contains(filename, "plugins/processors/"):
		return pluginProcessor
	case strings.Contains(filename, "plugins/aggregators/"):
		return pluginAggregator
	case strings.Contains(filename, "plugins/parsers/"):
		return pluginParser
	default:
		return pluginNone
	}
}
