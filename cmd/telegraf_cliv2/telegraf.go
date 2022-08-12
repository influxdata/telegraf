package main

import (
	"fmt"
	"io"
	"log" //nolint:revive
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
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

// deleteEmpty will create a new slice without any empty strings
// useful when using strings.Split(s, sep), when `sep` is provided
// but no `sep`` is found, returns slice of length 1 containing s
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
func runApp(args []string, outputBuffer io.Writer) error {
	var fConfigs, fConfigDirs cli.StringSlice
	var fTestWait int
	var fServiceWin, fServiceNameWin, fServiceDisplayNameWin, fServiceRestartDelay string
	var fServiceAutoRestartWin, fRunAsConsole bool
	var fUsage, fPlugins, fPprofAddr, fWatchConfig, fPidfile string
	var fRunOnce, fDebug, fQuiet, fTest, fDeprecationList, fInputList, fOutputList bool
	var fSubSectionFilters, fSubInputFilters, fSubOutputFilters, fsubAggregatorFilters, fSubProcessorFilters string

	// !!! The following flags are DEPRECATED !!!
	var fVersion bool      // Already covered with the subcommand `./telegraf version`
	var fSampleConfig bool // Already covered with the subcommand `./telegraf config`
	// !!!

	app := &cli.App{
		Name:   "Telegraf",
		Usage:  "The plugin-driven server agent for collecting & reporting metrics.",
		Writer: outputBuffer,
		Flags: []cli.Flag{
			// String slice flags
			&cli.StringSliceFlag{
				Name:        "config",
				Usage:       "configuration file to load",
				Destination: &fConfigs,
			},
			&cli.StringSliceFlag{
				Name:        "config-directory",
				Usage:       "directory containing additional *.conf files",
				Destination: &fConfigDirs,
			},
			// Int flags
			&cli.IntFlag{
				Name:        "test-wait",
				Usage:       "wait up to this many seconds for service inputs to complete in test mode",
				Destination: &fTestWait,
			},
			// Windows only string & bool flags
			&cli.StringFlag{
				Name:        "service",
				Usage:       "operate on the service (windows only)",
				Destination: &fServiceWin,
			},
			&cli.StringFlag{
				Name:        "service-name",
				DefaultText: "telegraf",
				Usage:       "service name (windows only)",
				Destination: &fServiceNameWin,
			},
			&cli.StringFlag{
				Name:        "service-display-name",
				DefaultText: "Telegraf Data Collector Service",
				Usage:       "service display name (windows only)",
				Destination: &fServiceDisplayNameWin,
			},
			&cli.StringFlag{
				Name:        "service-restart-delay",
				DefaultText: "5m",
				Usage:       "delay before service auto restart, default is 5m (windows only)",
				Destination: &fServiceRestartDelay,
			},
			&cli.BoolFlag{
				Name:        "service-restart-delay",
				Usage:       "auto restart service on failure (windows only)",
				Destination: &fServiceAutoRestartWin,
			},
			&cli.BoolFlag{
				Name:        "console",
				Usage:       "run as console application (windows only)",
				Destination: &fRunAsConsole,
			},
			//
			// String flags
			&cli.StringFlag{
				Name:        "usage",
				Usage:       "print usage for a plugin, ie, 'telegraf --usage mysql'",
				Destination: &fUsage,
			},
			&cli.StringFlag{
				Name:        "plugin-directory",
				Usage:       "path to directory containing external plugins",
				Destination: &fPlugins,
			},
			&cli.StringFlag{
				Name:        "pprof-addr",
				Usage:       "pprof address to listen on, not activate pprof if empty",
				Destination: &fPprofAddr,
			},
			&cli.StringFlag{
				Name:        "watch-config",
				Usage:       "Monitoring config changes [notify, poll]",
				Destination: &fWatchConfig,
			},
			&cli.StringFlag{
				Name:        "pidfile",
				Usage:       "file to write our pid to",
				Destination: &fPidfile,
			},
			//
			// Bool flags
			&cli.BoolFlag{
				Name:        "once",
				Usage:       "run one gather and exit",
				Destination: &fRunOnce,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "turn on debug logging",
				Destination: &fDebug,
			},
			&cli.BoolFlag{
				Name:        "quiet",
				Usage:       "run in quiet mode",
				Destination: &fQuiet,
			},
			&cli.BoolFlag{
				Name:        "test",
				Usage:       "enable test mode: gather metrics, print them out, and exit. Note: Test mode only runs inputs, not processors, aggregators, or outputs",
				Destination: &fTest,
			},
			&cli.BoolFlag{
				Name:        "deprecation-list",
				Usage:       "print all deprecated plugins or plugin options.",
				Destination: &fDeprecationList,
			},
			&cli.BoolFlag{
				Name:        "input-list",
				Usage:       "print available input plugins.",
				Destination: &fInputList,
			},
			&cli.BoolFlag{
				Name:        "output-list",
				Usage:       "print available output plugins.",
				Destination: &fOutputList,
			},
			//
			// TODO: These are missing flags, only input and output listing is possible
			// Unkown if anyone wants this, perhaps the right solution is to remove input-list and output-list
			// &cli.BoolFlag{
			// 	Name:        "aggregator-list",
			// 	Usage:       "print available aggregator plugins.",
			// 	Destination: &fAggregatorList,
			// },
			// &cli.BoolFlag{
			// 	Name:        "processor-list",
			// 	Usage:       "print available processor plugins.",
			// 	Destination: &fProcessorList,
			// },
			//
			// !!! The following flags are DEPRECATED !!!
			// The destination variables have comments explaining why they are deprecated
			&cli.BoolFlag{
				Name:        "version",
				Usage:       "DEPRECATED: display the version and exit",
				Destination: &fVersion,
			},
			&cli.BoolFlag{
				Name:        "sample-config",
				Usage:       "DEPRECATED: print out full sample configuration",
				Destination: &fSampleConfig,
			},
			// !!!
		},
		Action: func(*cli.Context) error {
			fmt.Println("boom! I say!")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "print out full sample configuration to stdout",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "section-filter",
						Usage:       "filter the sections to print, separator is ':'. Valid values are 'agent', 'global_tags', 'outputs', 'processors', 'aggregators' and 'inputs'",
						Destination: &fSubSectionFilters,
					},
					&cli.StringFlag{
						Name:        "input-filter",
						Usage:       "filter the inputs to enable, separator is :",
						Destination: &fSubInputFilters,
					},
					&cli.StringFlag{
						Name:        "output-filter",
						Usage:       "filter the outputs to enable, separator is :",
						Destination: &fSubOutputFilters,
					},
					&cli.StringFlag{
						Name:        "aggregator-filter",
						Usage:       "filter the aggregators to enable, separator is :",
						Destination: &fsubAggregatorFilters,
					},
					&cli.StringFlag{
						Name:        "processor-filter",
						Usage:       "filter the processors to enable, separator is :",
						Destination: &fSubProcessorFilters,
					},
				},
				Action: func(cCtx *cli.Context) error {
					// The sub_Filters are populated when the filter flags are set after the subcommand config
					// e.g. telegraf config --section-filter inputs
					sectionFilters := deleteEmpty(strings.Split(fSubSectionFilters, ":"))
					inputFilters := deleteEmpty(strings.Split(strings.TrimSpace(fSubInputFilters), ":"))
					outputFilters := deleteEmpty(strings.Split(strings.TrimSpace(fSubOutputFilters), ":"))
					aggregatorFilters := deleteEmpty(strings.Split(strings.TrimSpace(fsubAggregatorFilters), ":"))
					processorFilters := deleteEmpty(strings.Split(strings.TrimSpace(fSubProcessorFilters), ":"))

					printSampleConfig(
						outputBuffer,
						sectionFilters,
						inputFilters,
						outputFilters,
						aggregatorFilters,
						processorFilters,
					)
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "print current version to stdout.",
				Action: func(cCtx *cli.Context) error {
					fmt.Println(formatFullVersion())
					return nil
				},
			},
		},
	}

	return app.Run(args)
}

func main() {
	err := runApp(os.Args, os.Stdout)
	if err != nil {
		log.Fatalf("E! %s", err)
	}
}
