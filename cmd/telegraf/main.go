package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"github.com/awnumar/memguard"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/goplugin"
	"github.com/influxdata/telegraf/logger"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
	_ "github.com/influxdata/telegraf/plugins/secretstores/all"
)

type TelegrafConfig interface {
	CollectDeprecationInfos([]string, []string, []string, []string) map[string][]config.PluginDeprecationInfo
	PrintDeprecationList([]config.PluginDeprecationInfo)
}

type Filters struct {
	section     []string
	input       []string
	output      []string
	aggregator  []string
	processor   []string
	secretstore []string
}

func appendFilter(a, b string) string {
	if a != "" && b != "" {
		return fmt.Sprintf("%s:%s", a, b)
	}
	if a != "" {
		return a
	}
	return b
}

func processFilterFlags(ctx *cli.Context) Filters {
	var section, input, output, aggregator, processor, secretstore string

	// Support defining filters before and after the command
	// The old style was:
	// ./telegraf --section-filter inputs --input-filter cpu config >test.conf
	// The new style is:
	// ./telegraf config --section-filter inputs --input-filter cpu >test.conf
	// To support the old style, check if the parent context has the filter flags defined
	if len(ctx.Lineage()) >= 2 {
		parent := ctx.Lineage()[1] // ancestor contexts in order from child to parent
		section = parent.String("section-filter")
		input = parent.String("input-filter")
		output = parent.String("output-filter")
		aggregator = parent.String("aggregator-filter")
		processor = parent.String("processor-filter")
		secretstore = parent.String("secretstore-filter")
	}

	// If both the parent and command filters are defined, append them together
	section = appendFilter(section, ctx.String("section-filter"))
	input = appendFilter(input, ctx.String("input-filter"))
	output = appendFilter(output, ctx.String("output-filter"))
	aggregator = appendFilter(aggregator, ctx.String("aggregator-filter"))
	processor = appendFilter(processor, ctx.String("processor-filter"))
	secretstore = appendFilter(secretstore, ctx.String("secretstore-filter"))

	sectionFilters := deleteEmpty(strings.Split(section, ":"))
	inputFilters := deleteEmpty(strings.Split(input, ":"))
	outputFilters := deleteEmpty(strings.Split(output, ":"))
	aggregatorFilters := deleteEmpty(strings.Split(aggregator, ":"))
	processorFilters := deleteEmpty(strings.Split(processor, ":"))
	secretstoreFilters := deleteEmpty(strings.Split(secretstore, ":"))
	return Filters{sectionFilters, inputFilters, outputFilters, aggregatorFilters, processorFilters, secretstoreFilters}
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// runApp defines all the subcommands and flags for Telegraf
// this abstraction is used for testing, so outputBuffer and args can be changed
func runApp(args []string, outputBuffer io.Writer, pprof Server, c TelegrafConfig, m App) error {
	pluginFilterFlags := []cli.Flag{
		&cli.StringFlag{
			Name: "section-filter",
			Usage: "filter the sections to print, separator is ':'. " +
				"Valid values are 'agent', 'global_tags', 'outputs', 'processors', 'aggregators' and 'inputs'",
		},
		&cli.StringFlag{
			Name:  "input-filter",
			Usage: "filter the inputs to enable, separator is ':'",
		},
		&cli.StringFlag{
			Name:  "output-filter",
			Usage: "filter the outputs to enable, separator is ':'",
		},
		&cli.StringFlag{
			Name:  "aggregator-filter",
			Usage: "filter the aggregators to enable, separator is ':'",
		},
		&cli.StringFlag{
			Name:  "processor-filter",
			Usage: "filter the processors to enable, separator is ':'",
		},
		&cli.StringFlag{
			Name:  "secretstore-filter",
			Usage: "filter the secret-stores to enable, separator is ':'",
		},
	}

	extraFlags := append(pluginFilterFlags, cliFlags()...)

	// This function is used when Telegraf is run with only flags

	action := func(cCtx *cli.Context) error {
		err := logger.SetupLogging(logger.LogConfig{})
		if err != nil {
			return err
		}

		// Deprecated: Use execd instead
		// Load external plugins, if requested.
		if cCtx.String("plugin-directory") != "" {
			log.Printf("I! Loading external plugins from: %s", cCtx.String("plugin-directory"))
			if err := goplugin.LoadExternalPlugins(cCtx.String("plugin-directory")); err != nil {
				return err
			}
		}

		// switch for flags which just do something and exit immediately
		switch {
		// print available input plugins
		case cCtx.Bool("deprecation-list"):
			filters := processFilterFlags(cCtx)
			infos := c.CollectDeprecationInfos(
				filters.input, filters.output, filters.aggregator, filters.processor,
			)
			outputBuffer.Write([]byte("Deprecated Input Plugins:\n"))
			c.PrintDeprecationList(infos["inputs"])
			outputBuffer.Write([]byte("Deprecated Output Plugins:\n"))
			c.PrintDeprecationList(infos["outputs"])
			outputBuffer.Write([]byte("Deprecated Processor Plugins:\n"))
			c.PrintDeprecationList(infos["processors"])
			outputBuffer.Write([]byte("Deprecated Aggregator Plugins:\n"))
			c.PrintDeprecationList(infos["aggregators"])
			return nil
		// print available output plugins
		case cCtx.Bool("output-list"):
			outputBuffer.Write([]byte("Available Output Plugins:\n"))
			names := make([]string, 0, len(outputs.Outputs))
			for k := range outputs.Outputs {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				outputBuffer.Write([]byte(fmt.Sprintf("  %s\n", k)))
			}
			return nil
		// print available input plugins
		case cCtx.Bool("input-list"):
			outputBuffer.Write([]byte("Available Input Plugins:\n"))
			names := make([]string, 0, len(inputs.Inputs))
			for k := range inputs.Inputs {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				outputBuffer.Write([]byte(fmt.Sprintf("  %s\n", k)))
			}
			return nil
		// print usage for a plugin, ie, 'telegraf --usage mysql'
		case cCtx.String("usage") != "":
			err := PrintInputConfig(cCtx.String("usage"), outputBuffer)
			err2 := PrintOutputConfig(cCtx.String("usage"), outputBuffer)
			if err != nil && err2 != nil {
				return fmt.Errorf("%s and %s", err, err2)
			}
			return nil
		// DEPRECATED
		case cCtx.Bool("version"):
			outputBuffer.Write([]byte(fmt.Sprintf("%s\n", internal.FormatFullVersion())))
			return nil
		// DEPRECATED
		case cCtx.Bool("sample-config"):
			filters := processFilterFlags(cCtx)

			printSampleConfig(
				outputBuffer,
				filters.section,
				filters.input,
				filters.output,
				filters.aggregator,
				filters.processor,
				filters.secretstore,
			)
			return nil
		}

		if cCtx.String("pprof-addr") != "" {
			pprof.Start(cCtx.String("pprof-addr"))
		}

		filters := processFilterFlags(cCtx)

		g := GlobalFlags{
			config:      cCtx.StringSlice("config"),
			configDir:   cCtx.StringSlice("config-directory"),
			testWait:    cCtx.Int("test-wait"),
			watchConfig: cCtx.String("watch-config"),
			pidFile:     cCtx.String("pidfile"),
			plugindDir:  cCtx.String("plugin-directory"),
			test:        cCtx.Bool("test"),
			debug:       cCtx.Bool("debug"),
			once:        cCtx.Bool("once"),
			quiet:       cCtx.Bool("quiet"),
		}

		w := WindowFlags{
			service:             cCtx.String("service"),
			serviceName:         cCtx.String("service-name"),
			serviceDisplayName:  cCtx.String("service-display-name"),
			serviceRestartDelay: cCtx.String("service-restart-delay"),
			serviceAutoRestart:  cCtx.Bool("service-auto-restart"),
			console:             cCtx.Bool("console"),
		}

		m.Init(pprof.ErrChan(), filters, g, w)
		return m.Run()
	}

	app := &cli.App{
		Name:   "Telegraf",
		Usage:  "The plugin-driven server agent for collecting & reporting metrics.",
		Writer: outputBuffer,
		Flags: append(
			[]cli.Flag{
				// String slice flags
				&cli.StringSliceFlag{
					Name:  "config",
					Usage: "configuration file to load",
				},
				&cli.StringSliceFlag{
					Name:  "config-directory",
					Usage: "directory containing additional *.conf files",
				},
				// Int flags
				&cli.IntFlag{
					Name:  "test-wait",
					Usage: "wait up to this many seconds for service inputs to complete in test mode",
				},
				//
				// String flags
				&cli.StringFlag{
					Name:  "usage",
					Usage: "print usage for a plugin, ie, 'telegraf --usage mysql'",
				},
				&cli.StringFlag{
					Name:  "pprof-addr",
					Usage: "pprof host/IP and port to listen on (e.g. 'localhost:6060')",
				},
				&cli.StringFlag{
					Name:  "watch-config",
					Usage: "monitoring config changes [notify, poll]",
				},
				&cli.StringFlag{
					Name:  "pidfile",
					Usage: "file to write our pid to",
				},
				//
				// Bool flags
				&cli.BoolFlag{
					Name:  "once",
					Usage: "run one gather and exit",
				},
				&cli.BoolFlag{
					Name:  "debug",
					Usage: "turn on debug logging",
				},
				&cli.BoolFlag{
					Name:  "quiet",
					Usage: "run in quiet mode",
				},
				&cli.BoolFlag{
					Name: "test",
					Usage: "enable test mode: gather metrics, print them out, and exit. " +
						"Note: Test mode only runs inputs, not processors, aggregators, or outputs",
				},
				// TODO: Change "deprecation-list, input-list, output-list" flags to become a subcommand "list" that takes
				// "input,output,aggregator,processor, deprecated" as parameters
				&cli.BoolFlag{
					Name:  "deprecation-list",
					Usage: "print all deprecated plugins or plugin options",
				},
				&cli.BoolFlag{
					Name:  "input-list",
					Usage: "print available input plugins",
				},
				&cli.BoolFlag{
					Name:  "output-list",
					Usage: "print available output plugins",
				},
				//
				// !!! The following flags are DEPRECATED !!!
				// Already covered with the subcommand `./telegraf version`
				&cli.BoolFlag{
					Name:  "version",
					Usage: "DEPRECATED: display the version and exit",
				},
				// Already covered with the subcommand `./telegraf config`
				&cli.BoolFlag{
					Name:  "sample-config",
					Usage: "DEPRECATED: print out full sample configuration",
				},
				// Using execd plugin to add external plugins is preferred (less size impact, easier for end user)
				&cli.StringFlag{
					Name:  "plugin-directory",
					Usage: "DEPRECATED: path to directory containing external plugins",
				},
				// !!!
			}, extraFlags...),
		Action: action,
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "print out full sample configuration to stdout",
				Flags: pluginFilterFlags,
				Action: func(cCtx *cli.Context) error {
					// The sub_Filters are populated when the filter flags are set after the subcommand config
					// e.g. telegraf config --section-filter inputs
					filters := processFilterFlags(cCtx)

					printSampleConfig(
						outputBuffer,
						filters.section,
						filters.input,
						filters.output,
						filters.aggregator,
						filters.processor,
						filters.secretstore,
					)
					return nil
				},
			},
			{
				Name:  "list-secretstores",
				Usage: "list all known secret-stores",
				Description: `
The 'list-secretstores' command requires passing in your configuration file
containing the secret-store definitions you want to access. To get a
list of available secret-store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.

For help on how to define secret-stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following commands to list all available secret-stores

> telegraf list-secretstores

It will print a list of all known secret-store IDs that can be used	for
accessing the secrets in that store.`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "secretstore-filter",
						Usage: "filter the secret-stores to enable, separator is ':'",
					},
				},
				Action: func(cCtx *cli.Context) error {
					// Only load the secret-stores
					cCtx.Set("section-filter", "inputs:outpus:processors:aggregators")
					filters := processFilterFlags(cCtx)
					g := GlobalFlags{
						config:     cCtx.StringSlice("config"),
						configDir:  cCtx.StringSlice("config-directory"),
						plugindDir: cCtx.String("plugin-directory"),
						debug:      cCtx.Bool("debug"),
					}
					w := WindowFlags{}
					m.Init(nil, filters, g, w)

					ids, err := m.ListSecretStores()
					if err != nil {
						return err
					}
					sort.Strings(ids)

					_, _ = fmt.Println("Known secret stores:")
					for _, id := range ids {
						_, _ = fmt.Printf("    %s\n", id)
					}

					return nil
				},
			},
			{
				Name:  "list-secrets",
				Usage: "list known secrets in the given store(s)",
				Description: `
The 'list-secrets' command requires passing in your configuration file
containing the secret-store definitions you want to access. To get a
list of available secret-store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.
and use 'list-secretstores' to get the IDs of available stores.

For help on how to define secret-stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to list the keys of all known secrets in ALL
available stores

> telegraf list-secrets

To get the keys of all known secrets in a particular store, you can run

> telegraf list-secrets mystore

To also reveal the actual secret, i.e. the value, you can pass the
'--reveal-secret' flag.
`,
				ArgsUsage: "[secret-store ID]...[secret-store ID]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "reveal-secret",
						Usage: "also print the secret value",
					},
				},
				Action: func(cCtx *cli.Context) error {
					// Only load the secret-stores
					cCtx.Set("section-filter", "inputs:outpus:processors:aggregators")
					filters := processFilterFlags(cCtx)
					g := GlobalFlags{
						config:     cCtx.StringSlice("config"),
						configDir:  cCtx.StringSlice("config-directory"),
						plugindDir: cCtx.String("plugin-directory"),
						debug:      cCtx.Bool("debug"),
					}
					w := WindowFlags{}
					m.Init(nil, filters, g, w)

					args := cCtx.Args()
					var storeIDs []string
					if args.Present() {
						storeIDs = args.Slice()
					} else {
						ids, err := m.ListSecretStores()
						if err != nil {
							return fmt.Errorf("unable to determine secret-store IDs: %w", err)
						}
						storeIDs = ids
					}
					sort.Strings(storeIDs)

					reveal := cCtx.Bool("reveal-secret")
					for _, storeID := range storeIDs {
						store, err := m.GetSecretStore(storeID)
						if err != nil {
							return fmt.Errorf("unable to get secret-store %q: %w", storeID, err)
						}
						keys, err := store.List()
						if err != nil {
							return fmt.Errorf("unable to get secrets from store %q: %w", storeID, err)
						}
						sort.Strings(keys)

						_, _ = fmt.Printf("Known secrets for store %q:\n", storeID)
						for _, k := range keys {
							var v string
							if reveal {
								if v, err = store.Get(k); err != nil {
									return fmt.Errorf("unable to get value of secret %q from store %q: %w", k, storeID, err)
								}
							}
							_, _ = fmt.Printf("    %-30s  %s\n", k, v)
						}
					}

					return nil
				},
			},
			{
				Name:  "get-secret",
				Usage: "retrieves value of given secret from given store",
				Description: `
The 'get-secret' command requires passing in your configuration file
containing the secret-store definitions you want to access. To get a
list of available secret-store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.
and use 'list-secretstores' to get the IDs of available stores and
'list-secret' to get the key(s) of available secrets.

For help on how to define secret-stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to retrieve a secret from a secret store
available stores

> telegraf get-secret mystore mysecretkey

This will fetch the secret with the key 'mysecretkey' from the secret-store
with the ID 'mystore'.
`,
				ArgsUsage: "<secret-store ID> <secret key>",
				Action: func(cCtx *cli.Context) error {
					// Only load the secret-stores
					cCtx.Set("section-filter", "inputs:outpus:processors:aggregators")
					filters := processFilterFlags(cCtx)
					g := GlobalFlags{
						config:     cCtx.StringSlice("config"),
						configDir:  cCtx.StringSlice("config-directory"),
						plugindDir: cCtx.String("plugin-directory"),
						debug:      cCtx.Bool("debug"),
					}
					w := WindowFlags{}
					m.Init(nil, filters, g, w)

					args := cCtx.Args()
					if !args.Present() || args.Len() != 2 {
						return errors.New("invalid number of arguments")
					}

					storeID := args.First()
					key := args.Get(1)

					store, err := m.GetSecretStore(storeID)
					if err != nil {
						return fmt.Errorf("unable to get secret-store: %w", err)
					}
					value, err := store.Get(key)
					if err != nil {
						return fmt.Errorf("unable to get secret: %w", err)
					}
					_, _ = fmt.Printf("%s:%s = %s\n", storeID, key, value)

					return nil
				},
			},
			{
				Name:  "set-secret",
				Usage: "create or modify a secret in the given store",
				Description: `
The 'set-secret' command requires passing in your configuration file
containing the secret-store definitions you want to access. To get a
list of available secret-store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.
and use 'list-secretstores' to get the IDs of available stores.

For help on how to define secret-stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to create a secret in anm available secret-store

> telegraf set-secret mystore mysecretkey mysecretvalue

This will create a secret with the key 'mysecretkey' in the secret-store
with the ID 'mystore' with the value being set to 'mysecretvalue'. If a
secret with that key ('mysecretkey') already existed in that store, its
value will be modified.

When you leave out the value of the secret like

> telegraf set-secret mystore mysecretkey

you will be prompted to enter the value of the secret.
`,
				ArgsUsage: "<secret-store ID> <secret key>",
				Action: func(cCtx *cli.Context) error {
					// Only load the secret-stores
					cCtx.Set("section-filter", "inputs:outpus:processors:aggregators")
					filters := processFilterFlags(cCtx)
					g := GlobalFlags{
						config:     cCtx.StringSlice("config"),
						configDir:  cCtx.StringSlice("config-directory"),
						plugindDir: cCtx.String("plugin-directory"),
						debug:      cCtx.Bool("debug"),
					}
					w := WindowFlags{}
					m.Init(nil, filters, g, w)

					args := cCtx.Args()
					if !args.Present() || args.Len() < 2 {
						return errors.New("invalid number of arguments")
					}

					storeID := args.First()
					key := args.Get(1)
					value := args.Get(2)
					if value == "" {
						fmt.Printf("enter secret: ")
						b, err := term.ReadPassword(int(os.Stdin.Fd()))
						if err != nil {
							return err
						}
						fmt.Println()
						value = string(b)
					}

					store, err := m.GetSecretStore(storeID)
					if err != nil {
						return fmt.Errorf("unable to get secret-store: %w", err)
					}
					if err := store.Set(key, value); err != nil {
						return fmt.Errorf("unable to set secret: %w", err)
					}

					return nil
				},
			},
			{
				Name:  "version",
				Usage: "print current version to stdout",
				Action: func(cCtx *cli.Context) error {
					outputBuffer.Write([]byte(fmt.Sprintf("%s\n", internal.FormatFullVersion())))
					return nil
				},
			},
		},
	}

	// Make sure we safely erase secrets
	memguard.CatchInterrupt()
	defer memguard.Purge()

	return app.Run(args)
}

func main() {
	agent := Telegraf{}
	pprof := NewPprofServer()
	c := config.NewConfig()
	err := runApp(os.Args, os.Stdout, pprof, c, &agent)
	if err != nil {
		log.Fatalf("E! %s", err)
	}
}
