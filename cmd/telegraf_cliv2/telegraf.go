package main

import (
	"fmt"
	"io"
	"log" //nolint:revive
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/config/printer"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/goplugin"
	"github.com/influxdata/telegraf/logger"
	"github.com/urfave/cli/v2"

	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
)

var (
	version string
	commit  string
	branch  string
)

func formatFullVersion() string {
	var parts = []string{"Telegraf"}

	if version != "" {
		parts = append(parts, version)
	} else {
		parts = append(parts, "unknown")
	}

	if branch != "" || commit != "" {
		if branch == "" {
			branch = "unknown"
		}
		if commit == "" {
			commit = "unknown"
		}
		git := fmt.Sprintf("(git: %s %s)", branch, commit)
		parts = append(parts, git)
	}

	return strings.Join(parts, " ")
}

type TelegrafConfig interface {
	CollectDeprecationInfos([]string, []string, []string, []string) map[string][]config.PluginDeprecationInfo
	PrintDeprecationList([]config.PluginDeprecationInfo)
}

type Server interface {
	Start(string)
	ErrChan() <-chan error
}

type PprofServer struct {
	err chan error
}

func NewPprofServer() *PprofServer {
	return &PprofServer{
		err: make(chan error),
	}
}

func (p *PprofServer) Start(address string) {
	go func() {
		pprofHostPort := address
		parts := strings.Split(pprofHostPort, ":")
		if len(parts) == 2 && parts[0] == "" {
			pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
		}
		pprofHostPort = "http://" + pprofHostPort + "/debug/pprof"

		log.Printf("I! Starting pprof HTTP server at: %s", pprofHostPort)

		if err := http.ListenAndServe(address, nil); err != nil {
			p.err <- fmt.Errorf("E! %w", err)
		}
		close(p.err)
	}()
}

func (p *PprofServer) ErrChan() <-chan error {
	return p.err
}

type Filters struct {
	section    []string
	input      []string
	output     []string
	aggregator []string
	processor  []string
}

func processFilterFlags(section, input, output, aggregator, processor string) Filters {
	sectionFilters := deleteEmpty(strings.Split(section, ":"))
	inputFilters := deleteEmpty(strings.Split(input, ":"))
	outputFilters := deleteEmpty(strings.Split(output, ":"))
	aggregatorFilters := deleteEmpty(strings.Split(aggregator, ":"))
	processorFilters := deleteEmpty(strings.Split(processor, ":"))
	return Filters{sectionFilters, inputFilters, outputFilters, aggregatorFilters, processorFilters}
}

// deleteEmpty will create a new slice without any empty strings
// useful when using strings.Split(s, sep), when `sep` is provided
// but no `sep“ is found, returns slice of length 1 containing s
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
func runApp(args []string, outputBuffer io.Writer, pprof Server, c TelegrafConfig, m Manager) error {
	pluginFilterFlags := []cli.Flag{
		&cli.StringFlag{
			Name:  "section-filter",
			Usage: "filter the sections to print, separator is ':'. Valid values are 'agent', 'global_tags', 'outputs', 'processors', 'aggregators' and 'inputs'",
		},
		&cli.StringFlag{
			Name:  "input-filter",
			Usage: "filter the inputs to enable, separator is :",
		},
		&cli.StringFlag{
			Name:  "output-filter",
			Usage: "filter the outputs to enable, separator is :",
		},
		&cli.StringFlag{
			Name:  "aggregator-filter",
			Usage: "filter the aggregators to enable, separator is :",
		},
		&cli.StringFlag{
			Name:  "processor-filter",
			Usage: "filter the processors to enable, separator is :",
		},
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
				// Windows only string & bool flags
				&cli.StringFlag{
					Name:  "service",
					Usage: "operate on the service (windows only)",
				},
				&cli.StringFlag{
					Name:        "service-name",
					DefaultText: "telegraf",
					Usage:       "service name (windows only)",
				},
				&cli.StringFlag{
					Name:        "service-display-name",
					DefaultText: "Telegraf Data Collector Service",
					Usage:       "service display name (windows only)",
				},
				&cli.StringFlag{
					Name:        "service-restart-delay",
					DefaultText: "5m",
				},
				&cli.BoolFlag{
					Name:  "service-auto-restart",
					Usage: "auto restart service on failure (windows only)",
				},
				&cli.BoolFlag{
					Name:  "console",
					Usage: "run as console application (windows only)",
				},
				//
				// String flags
				&cli.StringFlag{
					Name:  "usage",
					Usage: "print usage for a plugin, ie, 'telegraf --usage mysql'",
				},
				&cli.StringFlag{
					Name:  "pprof-addr",
					Usage: "pprof address to listen on, not activate pprof if empty",
				},
				&cli.StringFlag{
					Name:  "watch-config",
					Usage: "Monitoring config changes [notify, poll]",
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
					Name:  "test",
					Usage: "enable test mode: gather metrics, print them out, and exit. Note: Test mode only runs inputs, not processors, aggregators, or outputs",
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
				// Using execd plugin to add external plugins is preffered (less size impact, easier for end user)
				&cli.StringFlag{
					Name:  "plugin-directory",
					Usage: "path to directory containing external plugins",
				},
				// !!!
			}, pluginFilterFlags...),
		Action: func(cCtx *cli.Context) error {
			logger.SetupLogging(logger.LogConfig{})

			// Configure version
			if err := internal.SetVersion(version); err != nil {
				log.Println("Telegraf version already configured to: " + internal.Version())
			}

			// Deprecated: Use execd instead
			// Load external plugins, if requested.
			if cCtx.String("plugin-directory") != "" {
				log.Printf("I! Loading external plugins from: %s", cCtx.String("plugin-directory"))
				if err := goplugin.LoadExternalPlugins(cCtx.String("plugin-directory")); err != nil {
					return fmt.Errorf("E! %w", err)
				}
			}

			// switch for flags which just do something and exit immediately
			switch {
			// print available input plugins
			case cCtx.Bool("deprecation-list"):
				filters := processFilterFlags(
					cCtx.String("section-filter"),
					cCtx.String("input-filter"),
					cCtx.String("output-filter"),
					cCtx.String("aggregator-filter"),
					cCtx.String("processor-filter"),
				)
				infos := c.CollectDeprecationInfos(
					filters.input, filters.output, filters.aggregator, filters.processor,
				)
				//nolint:revive // We will notice if Println fails
				_, _ = outputBuffer.Write([]byte("Deprecated Input Plugins:\n"))
				c.PrintDeprecationList(infos["inputs"])
				//nolint:revive // We will notice if Println fails
				_, _ = outputBuffer.Write([]byte("Deprecated Output Plugins:\n"))
				c.PrintDeprecationList(infos["outputs"])
				//nolint:revive // We will notice if Println fails
				_, _ = outputBuffer.Write([]byte("Deprecated Processor Plugins:\n"))
				c.PrintDeprecationList(infos["processors"])
				//nolint:revive // We will notice if Println fails
				_, _ = outputBuffer.Write([]byte("Deprecated Aggregator Plugins:\n"))
				c.PrintDeprecationList(infos["aggregators"])
				return nil
			// print available output plugins
			case cCtx.Bool("output-list"):
				_, _ = outputBuffer.Write([]byte("Available Output Plugins:\n"))
				names := make([]string, 0, len(outputs.Outputs))
				for k := range outputs.Outputs {
					names = append(names, k)
				}
				sort.Strings(names)
				for _, k := range names {
					_, _ = outputBuffer.Write([]byte(fmt.Sprintf("  %s\n", k)))
				}
				return nil
			// print available input plugins
			case cCtx.Bool("input-list"):
				_, _ = outputBuffer.Write([]byte("Available Input Plugins:\n"))
				names := make([]string, 0, len(inputs.Inputs))
				for k := range inputs.Inputs {
					names = append(names, k)
				}
				sort.Strings(names)
				for _, k := range names {
					_, _ = outputBuffer.Write([]byte(fmt.Sprintf("  %s\n", k)))
				}
				return nil
			// print usage for a plugin, ie, 'telegraf --usage mysql'
			case cCtx.String("usage") != "":
				err := printer.PrintInputConfig(cCtx.String("usage"), outputBuffer)
				err2 := printer.PrintOutputConfig(cCtx.String("usage"), outputBuffer)
				if err != nil && err2 != nil {
					return fmt.Errorf("E! %s and %s", err, err2)
				}
				return nil
			// DEPRECATED
			case cCtx.Bool("version"):
				fmt.Println(formatFullVersion())
				return nil
			// DEPRECATED
			case cCtx.Bool("sample-config"):
				filters := processFilterFlags(
					cCtx.String("section-filter"),
					cCtx.String("input-filter"),
					cCtx.String("output-filter"),
					cCtx.String("aggregator-filter"),
					cCtx.String("processor-filter"),
				)

				printSampleConfig(
					outputBuffer,
					filters.section,
					filters.input,
					filters.output,
					filters.aggregator,
					filters.processor,
				)
				return nil
			}

			if cCtx.String("pprof-addr") != "" {
				pprof.Start(cCtx.String("pprof-addr"))
			}

			filters := processFilterFlags(
				cCtx.String("section-filter"),
				cCtx.String("input-filter"),
				cCtx.String("output-filter"),
				cCtx.String("aggregator-filter"),
				cCtx.String("processor-filter"),
			)

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
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "print out full sample configuration to stdout",
				Flags: pluginFilterFlags,
				Action: func(cCtx *cli.Context) error {
					// The sub_Filters are populated when the filter flags are set after the subcommand config
					// e.g. telegraf config --section-filter inputs
					filters := processFilterFlags(
						cCtx.String("section-filter"),
						cCtx.String("input-filter"),
						cCtx.String("output-filter"),
						cCtx.String("aggregator-filter"),
						cCtx.String("processor-filter"),
					)

					printSampleConfig(
						outputBuffer,
						filters.section,
						filters.input,
						filters.output,
						filters.aggregator,
						filters.processor,
					)
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "print current version to stdout",
				Action: func(cCtx *cli.Context) error {
					_, _ = outputBuffer.Write([]byte(formatFullVersion()))
					return nil
				},
			},
		},
	}

	return app.Run(args)
}

func main() {
	agent := AgentManager{}
	pprof := NewPprofServer()
	c := config.NewConfig()
	err := runApp(os.Args, os.Stdout, pprof, c, &agent)
	if err != nil {
		log.Fatalf("E! %s", err)
	}
}
