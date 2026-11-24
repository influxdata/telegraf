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

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/cmd/telegraf/commands"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/migrations"
)

func getConfigActions(outputBuffer io.Writer) commands.ConfigActions {
	return commands.ConfigActions{
		Config: func(cCtx *cli.Context) error {
			// The sub_Filters are populated when the filter flags are set after the subcommand config
			// e.g. telegraf config --section-filter inputs
			filters := processFilterFlags(cCtx)

			printSampleConfig(outputBuffer, filters)
			return nil
		},
		Check: func(cCtx *cli.Context) error {
			// Setup logging
			logConfig := &logger.Config{Debug: cCtx.Bool("debug")}
			if err := logger.SetupLogging(logConfig); err != nil {
				return err
			}

			// Set the environment variables handling mode
			if cCtx.Bool("strict-env-handling") && cCtx.Bool("non-strict-env-handling") {
				return errors.New("flags --strict-env-handling and --non-strict-env-handling cannot be used together")
			}
			if !cCtx.Bool("strict-env-handling") && !cCtx.Bool("non-strict-env-handling") {
				msg := "Strict environment variable handling will be the new default starting with v1.38.0! " +
					"If your configuration works with strict handling or you don't use environment variables it is safe " +
					"to ignore this warning. Otherwise please explicitly add the --non-strict-env-handling flag!"
				log.Println("W! " + color.YellowString(msg))
			}
			config.NonStrictEnvVarHandling = !cCtx.Bool("strict-env-handling")

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

			// Load the config and try to initialize the plugins
			c := config.NewConfig()
			c.Agent.Quiet = cCtx.Bool("quiet")
			if err := c.LoadAll(configFiles...); err != nil {
				return err
			}

			ag := agent.NewAgent(c)
			return ag.InitPlugins()
		},
		Create: func(cCtx *cli.Context) error {
			filters := processFilterFlags(cCtx)

			printSampleConfig(outputBuffer, filters)
			return nil
		},
		Migrate: func(cCtx *cli.Context) error {
			// Setup logging
			logConfig := &logger.Config{Debug: cCtx.Bool("debug")}
			if err := logger.SetupLogging(logConfig); err != nil {
				return err
			}

			// Check if we have migrations at all. There might be
			// none if you run a custom build without migrations
			// enabled.
			migrationsGeneral := len(migrations.GeneralMigrations) + len(migrations.GlobalMigrations)
			migrationsPlugins := len(migrations.PluginMigrations)
			migrationsOptions := len(migrations.PluginOptionMigrations)
			if migrationsGeneral+migrationsPlugins+migrationsOptions == 0 {
				return errors.New("no migrations available")
			}
			log.Printf(
				"%d general, %d plugin and %d plugin-option migrations available",
				migrationsGeneral, migrationsPlugins, migrationsOptions,
			)

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
	}
}
