package commands

import "github.com/urfave/cli/v2"

type ConfigActions struct {
	Config  cli.ActionFunc
	Check   cli.ActionFunc
	Create  cli.ActionFunc
	Migrate cli.ActionFunc
}

func configCommands(actions ConfigActions, configHandlingFlags []cli.Flag) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "config",
			Usage:  "commands for generating and migrating configurations",
			Flags:  configHandlingFlags,
			Action: actions.Config,
			Subcommands: []*cli.Command{
				{
					Name:  "check",
					Usage: "check configuration file(s) for issues",
					Description: `
		The 'check' command reads the configuration files specified via '--config' or
		'--config-directory' and tries to initialize, but not start, the plugins.
		Syntax and semantic errors detectable without starting the plugins will
		be reported.
		If no configuration file is	explicitly specified the command reads the
		default locations and uses those configuration files.

		To check the file 'mysettings.conf' use

		> telegraf config check --config mysettings.conf
		`,
					Flags:  configHandlingFlags,
					Action: actions.Check,
				},
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
					Flags:  configHandlingFlags,
					Action: actions.Create,
				},
				{
					Name:  "migrate",
					Usage: "migrate deprecated plugins and options of the configuration(s)",
					Description: `
The 'migrate' command reads the configuration files specified via '--config' or
'--config-directory' and tries to migrate plugins or options that are currently
deprecated using the recommended replacements. If no configuration file is
explicitly specified the command reads the default locations and uses those
configuration files. Migrated files are stored with a '.migrated' suffix at the
location of the  inputs. If you are migrating remote configurations the migrated
configurations is stored in the current directory using the filename of the URL
with a '.migrated' suffix.
It is highly recommended to test those migrated configurations before using
those files unattended!

To migrate the file 'mysettings.conf' use

> telegraf config migrate --config mysettings.conf
`,
					Flags: append(configHandlingFlags,
						&cli.BoolFlag{
							Name:  "force",
							Usage: "forces overwriting of an existing migration file",
						},
					),
					Action: actions.Migrate,
				},
			},
		},
	}
}
