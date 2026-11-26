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

func pluginsCommand(actions commandActions) *cli.Command {
	subcommands := []*cli.Command{
		{
			Name:   "inputs",
			Usage:  "Print available input plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsInputs,
		},
		{
			Name:   "outputs",
			Usage:  "Print available output plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsOutputs,
		},
		{
			Name:   "processors",
			Usage:  "Print available processor plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsProcessors,
		},
		{
			Name:   "aggregators",
			Usage:  "Print available aggregator plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsAggregators,
		},
		{
			Name:   "secretstores",
			Usage:  "Print available secretstore plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsSecretstores,
		},
		{
			Name:   "parsers",
			Usage:  "Print available parser plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsParsers,
		},
		{
			Name:   "serializers",
			Usage:  "Print available serializer plugins",
			Flags:  pluginFlags(),
			Action: actions.pluginsSerializers,
		},
	}
	return &cli.Command{
		Name:        "plugins",
		Usage:       "commands for printing available plugins",
		Flags:       pluginFlags(),
		Action:      actions.plugins,
		Subcommands: subcommands,
	}
}

func pluginFlags() []cli.Flag {
	return []cli.Flag{&cli.BoolFlag{Name: "deprecated", Usage: "print only deprecated plugins"}}
}

func listPlugins[V, D any](ctx *cli.Context, prefix string, plugins map[string]V, deprecated map[string]D) error {
	if ctx.Bool("deprecated") {
		_, err := ctx.App.Writer.Write(pluginNames(deprecated, prefix))
		return err
	}
	_, err := ctx.App.Writer.Write(pluginNames(plugins, prefix))
	return err
}

func pluginNames[M ~map[string]V, V any](m M, prefix string) []byte {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, fmt.Sprintf("%s.%s\n", prefix, k))
	}
	sort.Strings(names)
	return []byte(strings.Join(names, ""))
}

func pluginsAction(ctx *cli.Context) error {
	for _, cmd := range ctx.Command.Subcommands {
		// urfave/cli adds a help subcommand when it sets up the command.
		if cmd.Name == "help" {
			continue
		}
		if err := cmd.Action(ctx); err != nil {
			return err
		}
	}
	return nil
}

func pluginsInputsAction(ctx *cli.Context) error {
	return listPlugins(ctx, "inputs", inputs.Inputs, inputs.Deprecations)
}

func pluginsOutputsAction(ctx *cli.Context) error {
	return listPlugins(ctx, "outputs", outputs.Outputs, outputs.Deprecations)
}

func pluginsProcessorsAction(ctx *cli.Context) error {
	return listPlugins(ctx, "processors", processors.Processors, processors.Deprecations)
}

func pluginsAggregatorsAction(ctx *cli.Context) error {
	return listPlugins(ctx, "aggregators", aggregators.Aggregators, aggregators.Deprecations)
}

func pluginsSecretstoresAction(ctx *cli.Context) error {
	return listPlugins(ctx, "secretstores", secretstores.SecretStores, secretstores.Deprecations)
}

func pluginsParsersAction(ctx *cli.Context) error {
	return listPlugins(ctx, "parsers", parsers.Parsers, parsers.Deprecations)
}

func pluginsSerializersAction(ctx *cli.Context) error {
	return listPlugins(ctx, "serializers", serializers.Serializers, serializers.Deprecations)
}
