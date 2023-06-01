// Command handling for configuration "config" command
package main

import (
	"io"

	"github.com/urfave/cli/v2"
)

func getConfigCommands(m App, pluginFilterFlags []cli.Flag, outputBuffer io.Writer) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "config",
			Usage: "commands for generating and migrating configurations",
			Flags: pluginFilterFlags,
			Action: func(cCtx *cli.Context) error {
				// The sub_Filters are populated when the filter flags are set after the subcommand config
				// e.g. telegraf config --section-filter inputs
				filters := processFilterFlags(cCtx)

				printSampleConfig(outputBuffer, filters)
				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name:  "create",
					Usage: "create a full sample configuration and show it",
					Description: `
The 'create' produces a full configuration containing all plugins as an example
and shows it on the console. You may apply 'section' or 'plugin' filtering
to reduce the output to the plugins you need

Create the full configuration

> telegraf config create

To produce a configuration only containing a Modbus input plugin and an
InfluxDB v2 output plugin use

> telegraf config create --section-filter "inputs:outputs" --input-filter "modbus" --output-filter "influxdb_v2"
`,
					Flags: pluginFilterFlags,
					Action: func(cCtx *cli.Context) error {
						filters := processFilterFlags(cCtx)

						printSampleConfig(outputBuffer, filters)
						return nil
					},
				},
			},
		},
	}
}
