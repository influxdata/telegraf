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

// TODO: Wil be deleted with: https://github.com/influxdata/telegraf/pull/11656
var (
	version string
	commit  string
	branch  string
)

// TODO: Wil be deleted with: https://github.com/influxdata/telegraf/pull/11656
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
	var fSubSectionFilters, fSubInputFilters, fSubOutputFilters, fsubAggregatorFilters, fSubProcessorFilters string

	app := &cli.App{
		Name:   "Telegraf",
		Usage:  "The plugin-driven server agent for collecting & reporting metrics.",
		Writer: outputBuffer,
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
