// Command handling for configuration "config" command
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/migrations"
)

func getConfigCommands(pluginFilterFlags []cli.Flag, outputBuffer io.Writer) []*cli.Command {
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

> telegraf --config mysettings.conf config migrate
`,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "force",
							Usage: "forces overwriting of an existing migration file",
						},
					},
					Action: func(cCtx *cli.Context) error {
						// Setup logging
						logConfig := logger.Config{Debug: cCtx.Bool("debug")}
						if err := logger.SetupLogging(logConfig); err != nil {
							return err
						}

						// Check if we have migrations at all. There might be
						// none if you run a custom build without migrations
						// enabled.
						if len(migrations.PluginMigrations) == 0 {
							return errors.New("no migrations available")
						}
						log.Printf("%d plugin migration(s) available", len(migrations.PluginMigrations))

						// Collect the given configuration files
						configFiles := cCtx.StringSlice("config")
						configDir := cCtx.StringSlice("config-directory")
						for _, fConfigDirectory := range configDir {
							files, err := config.WalkDirectory(fConfigDirectory)
							if err != nil {
								return err
							}
							configFiles = append(configFiles, files...)
						}

						// If no "config" or "config-directory" flag(s) was
						// provided we should load default configuration files
						if len(configFiles) == 0 {
							paths, err := config.GetDefaultConfigPath()
							if err != nil {
								return err
							}
							configFiles = paths
						}

						for _, fn := range configFiles {
							log.Printf("D! Trying to migrate %q...", fn)

							// Read and parse the config file
							data, remote, err := config.LoadConfigFile(fn)
							if err != nil {
								return fmt.Errorf("opening input %q failed: %w", fn, err)
							}

							out, applied, err := config.ApplyMigrations(data)
							if err != nil {
								return err
							}

							// Do not write a migration file if nothing was done
							if applied == 0 {
								log.Printf("I! No migration applied for %q", fn)
								continue
							}

							// Construct the output filename
							// For remote locations we just save the filename
							// with the migrated suffix.
							outfn := fn + ".migrated"
							if remote {
								u, err := url.Parse(fn)
								if err != nil {
									return fmt.Errorf("parsing remote config URL %q failed: %w", fn, err)
								}
								outfn = filepath.Base(u.Path) + ".migrated"
							}

							log.Printf("I! %d migration applied for %q, writing result as %q", applied, fn, outfn)

							// Make sure the file does not exist yet if we should not overwrite
							if !cCtx.Bool("force") {
								if _, err := os.Stat(outfn); !errors.Is(err, os.ErrNotExist) {
									return fmt.Errorf("output file %q already exists", outfn)
								}
							}

							// Write the output file
							if err := os.WriteFile(outfn, out, 0640); err != nil {
								return fmt.Errorf("writing output %q failed: %w", outfn, err)
							}
						}
						return nil
					},
				},
			},
		},
	}
}
