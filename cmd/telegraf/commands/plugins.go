package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/plugins/serializers"
)

func pluginsCommand() *cli.Command {
	subcommands := []*cli.Command{
		{
			Name:   "inputs",
			Usage:  "Print available input plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("inputs", inputs.Inputs, inputs.Deprecations),
		},
		{
			Name:   "outputs",
			Usage:  "Print available output plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("outputs", outputs.Outputs, outputs.Deprecations),
		},
		{
			Name:   "processors",
			Usage:  "Print available processor plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("processors", processors.Processors, processors.Deprecations),
		},
		{
			Name:   "aggregators",
			Usage:  "Print available aggregator plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("aggregators", aggregators.Aggregators, aggregators.Deprecations),
		},
		{
			Name:   "secretstores",
			Usage:  "Print available secretstore plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("secretstores", secretstores.SecretStores, secretstores.Deprecations),
		},
		{
			Name:   "parsers",
			Usage:  "Print available parser plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("parsers", parsers.Parsers, parsers.Deprecations),
		},
		{
			Name:   "serializers",
			Usage:  "Print available serializer plugins",
			Flags:  pluginFlags(),
			Action: listPlugins("serializers", serializers.Serializers, serializers.Deprecations),
		},
	}
	return &cli.Command{
		Name:  "plugins",
		Usage: "commands for printing available plugins",
		Flags: pluginFlags(),
		Action: func(ctx *cli.Context) error {
			for _, cmd := range subcommands {
				if err := cmd.Action(ctx); err != nil {
					return err
				}
			}
			return nil
		},
		Subcommands: subcommands,
	}
}

func pluginFlags() []cli.Flag {
	return []cli.Flag{&cli.BoolFlag{Name: "deprecated", Usage: "print only deprecated plugins"}}
}

func listPlugins[V, D any](prefix string, plugins map[string]V, deprecated map[string]D) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		if ctx.Bool("deprecated") {
			_, err := ctx.App.Writer.Write(pluginNames(deprecated, prefix))
			return err
		}
		_, err := ctx.App.Writer.Write(pluginNames(plugins, prefix))
		return err
	}
}

func pluginNames[M ~map[string]V, V any](m M, prefix string) []byte {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, fmt.Sprintf("%s.%s\n", prefix, k))
	}
	sort.Strings(names)
	return []byte(strings.Join(names, ""))
}
