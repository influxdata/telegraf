// Command handling for configuration "plugins" command
package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/urfave/cli/v2"
)

func pluginNames[M ~map[string]V, V any](m M, prefix string) []byte {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, fmt.Sprintf("%s.%s\n", prefix, k))
	}
	sort.Strings(names)
	return []byte(strings.Join(names, ""))
}

func getPluginCommands(outputBuffer io.Writer) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "plugins",
			Usage: "commands for printing available plugins",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "deprecated",
					Usage: "print only deprecated plugins",
				},
			},
			Action: func(cCtx *cli.Context) error {
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
			Subcommands: []*cli.Command{
				{
					Name:  "inputs",
					Usage: "Print available input plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(inputs.Deprecations, "inputs"))
						} else {
							outputBuffer.Write(pluginNames(inputs.Inputs, "inputs"))
						}
						return nil
					},
				},
				{
					Name:  "outputs",
					Usage: "Print available output plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(outputs.Deprecations, "outputs"))
						} else {
							outputBuffer.Write(pluginNames(outputs.Outputs, "outputs"))
						}
						return nil
					},
				},
				{
					Name:  "processors",
					Usage: "Print available processor plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(processors.Deprecations, "processors"))
						} else {
							outputBuffer.Write(pluginNames(processors.Processors, "processors"))
						}
						return nil
					},
				},
				{
					Name:  "aggregators",
					Usage: "Print available aggregator plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(aggregators.Deprecations, "aggregators"))
						} else {
							outputBuffer.Write(pluginNames(aggregators.Aggregators, "aggregators"))
						}
						return nil
					},
				},
				{
					Name:  "secretstores",
					Usage: "Print available secretstore plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(secretstores.Deprecations, "secretstores"))
						} else {
							outputBuffer.Write(pluginNames(secretstores.SecretStores, "secretstores"))
						}
						return nil
					},
				},
				{
					Name:  "parsers",
					Usage: "Print available parser plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(parsers.Deprecations, "parsers"))
						} else {
							outputBuffer.Write(pluginNames(parsers.Parsers, "parsers"))
						}
						return nil
					},
				},
				{
					Name:  "serializers",
					Usage: "Print available serializer plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(pluginNames(serializers.Deprecations, "serializers"))
						} else {
							outputBuffer.Write(pluginNames(serializers.Serializers, "serializers"))
						}
						return nil
					},
				},
			},
		},
	}
}
