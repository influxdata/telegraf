// Command handling for configuration "plugins" command
package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/cmd/telegraf/commands"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/plugins/serializers"
)

func pluginNames[M ~map[string]V, V any](m M, prefix string) []byte {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, fmt.Sprintf("%s.%s\n", prefix, k))
	}
	sort.Strings(names)
	return []byte(strings.Join(names, ""))
}

func getPluginActions(outputBuffer io.Writer) commands.PluginActions {
	return commands.PluginActions{
		Plugins: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(inputs.Deprecations, "inputs"))
				outputBuffer.Write(pluginNames(outputs.Deprecations, "outputs"))
				outputBuffer.Write(pluginNames(processors.Deprecations, "processors"))
				outputBuffer.Write(pluginNames(aggregators.Deprecations, "aggregators"))
				outputBuffer.Write(pluginNames(secretstores.Deprecations, "secretstores"))
				outputBuffer.Write(pluginNames(parsers.Deprecations, "parsers"))
				outputBuffer.Write(pluginNames(serializers.Deprecations, "serializers"))
			} else {
				outputBuffer.Write(pluginNames(inputs.Inputs, "inputs"))
				outputBuffer.Write(pluginNames(outputs.Outputs, "outputs"))
				outputBuffer.Write(pluginNames(processors.Processors, "processors"))
				outputBuffer.Write(pluginNames(aggregators.Aggregators, "aggregators"))
				outputBuffer.Write(pluginNames(secretstores.SecretStores, "secretstores"))
				outputBuffer.Write(pluginNames(parsers.Parsers, "parsers"))
				outputBuffer.Write(pluginNames(serializers.Serializers, "serializers"))
			}

			return nil
		},
		Inputs: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(inputs.Deprecations, "inputs"))
			} else {
				outputBuffer.Write(pluginNames(inputs.Inputs, "inputs"))
			}
			return nil
		},
		Outputs: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(outputs.Deprecations, "outputs"))
			} else {
				outputBuffer.Write(pluginNames(outputs.Outputs, "outputs"))
			}
			return nil
		},
		Processors: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(processors.Deprecations, "processors"))
			} else {
				outputBuffer.Write(pluginNames(processors.Processors, "processors"))
			}
			return nil
		},
		Aggregators: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(aggregators.Deprecations, "aggregators"))
			} else {
				outputBuffer.Write(pluginNames(aggregators.Aggregators, "aggregators"))
			}
			return nil
		},
		Secretstores: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(secretstores.Deprecations, "secretstores"))
			} else {
				outputBuffer.Write(pluginNames(secretstores.SecretStores, "secretstores"))
			}
			return nil
		},
		Parsers: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(parsers.Deprecations, "parsers"))
			} else {
				outputBuffer.Write(pluginNames(parsers.Parsers, "parsers"))
			}
			return nil
		},
		Serializers: func(cCtx *cli.Context) error {
			if cCtx.Bool("deprecated") {
				outputBuffer.Write(pluginNames(serializers.Deprecations, "serializers"))
			} else {
				outputBuffer.Write(pluginNames(serializers.Serializers, "serializers"))
			}
			return nil
		},
	}
}
