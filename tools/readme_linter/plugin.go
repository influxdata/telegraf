package main

import (
	"path/filepath"
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
	// Switch takes `plugins/inputs/amd_rocm_smi/README.md` and converts it to
	// `plugins/inputs`. This avoids parsing READMEs that are under a plugin
	// like those found in test folders as actual plugin readmes.
	switch filepath.Dir(filepath.Dir(filename)) {
	case "plugins/inputs":
		return pluginInput
	case "plugins/outputs":
		return pluginOutput
	case "plugins/processors":
		return pluginProcessor
	case "plugins/aggregators":
		return pluginAggregator
	case "plugins/parsers":
		return pluginParser
	default:
		return pluginNone
	}
}
