// Command handling for configuration "plugins" command
package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/urfave/cli/v2"
)

var deprecatedFlag = cli.BoolFlag{
	Name:  "deprecated",
	Usage: "print only deprecated plugins",
}

func getInputNames() []byte {
	inputNames := make([]string, 0, len(inputs.Inputs))
	for i := range inputs.Inputs {
		inputNames = append(inputNames, fmt.Sprintf("inputs.%s\n", i))
	}
	sort.Strings(inputNames)

	return []byte(strings.Join(inputNames, ""))
}

func getOutputNames() []byte {
	outputNames := make([]string, 0, len(outputs.Outputs))
	for i := range outputs.Outputs {
		outputNames = append(outputNames, fmt.Sprintf("outputs.%s\n", i))
	}
	sort.Strings(outputNames)

	return []byte(strings.Join(outputNames, ""))
}

func getProcessorNames() []byte {
	processorNames := make([]string, 0, len(processors.Processors))
	for i := range processors.Processors {
		processorNames = append(processorNames, fmt.Sprintf("processors.%s\n", i))
	}
	sort.Strings(processorNames)

	return []byte(strings.Join(processorNames, ""))
}

func getAggregatorNames() []byte {
	aggregatorNames := make([]string, 0, len(aggregators.Aggregators))
	for i := range aggregators.Aggregators {
		aggregatorNames = append(aggregatorNames, fmt.Sprintf("aggregators.%s\n", i))
	}
	sort.Strings(aggregatorNames)

	return []byte(strings.Join(aggregatorNames, ""))
}

func getSecretstoreNames() []byte {
	stores := make([]string, 0, len(secretstores.SecretStores))
	for i := range secretstores.SecretStores {
		stores = append(stores, fmt.Sprintf("secretstores.%s\n", i))
	}
	sort.Strings(stores)

	return []byte(strings.Join(stores, ""))
}

func getDeprecatedInputNames() []byte {
	inputNames := make([]string, 0, len(inputs.Deprecations))
	for i := range inputs.Deprecations {
		inputNames = append(inputNames, fmt.Sprintf("inputs.%s\n", i))
	}
	sort.Strings(inputNames)

	return []byte(strings.Join(inputNames, ""))
}

func getDeprecatedOutputNames() []byte {
	outputNames := make([]string, 0, len(outputs.Deprecations))
	for i := range outputs.Deprecations {
		outputNames = append(outputNames, fmt.Sprintf("outputs.%s\n", i))
	}
	sort.Strings(outputNames)

	return []byte(strings.Join(outputNames, ""))
}

func getDeprecatedProcessorNames() []byte {
	processorNames := make([]string, 0, len(processors.Deprecations))
	for i := range processors.Deprecations {
		processorNames = append(processorNames, fmt.Sprintf("processors.%s\n", i))
	}
	sort.Strings(processorNames)

	return []byte(strings.Join(processorNames, ""))
}

func getDeprecatedAggregatorNames() []byte {
	aggregatorNames := make([]string, 0, len(aggregators.Deprecations))
	for i := range aggregators.Deprecations {
		aggregatorNames = append(aggregatorNames, fmt.Sprintf("aggregators.%s\n", i))
	}
	sort.Strings(aggregatorNames)

	return []byte(strings.Join(aggregatorNames, ""))
}

func getDeprecatedSecretstoreNames() []byte {
	stores := make([]string, 0, len(secretstores.Deprecations))
	for i := range secretstores.Deprecations {
		stores = append(stores, fmt.Sprintf("secretstores.%s\n", i))
	}
	sort.Strings(stores)

	return []byte(strings.Join(stores, ""))
}

func getPluginCommands(outputBuffer io.Writer) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "plugins",
			Usage: "commands for printing available plugins",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "deprecated",
					Usage: "print only deprecated plugins",
				},
			},
			Action: func(cCtx *cli.Context) error {
				if cCtx.Bool("deprecated") {
					outputBuffer.Write(getDeprecatedInputNames())
					outputBuffer.Write(getDeprecatedOutputNames())
					outputBuffer.Write(getDeprecatedProcessorNames())
					outputBuffer.Write(getDeprecatedAggregatorNames())
					outputBuffer.Write(getDeprecatedSecretstoreNames())
				} else {
					outputBuffer.Write(getInputNames())
					outputBuffer.Write(getOutputNames())
					outputBuffer.Write(getProcessorNames())
					outputBuffer.Write(getAggregatorNames())
					outputBuffer.Write(getSecretstoreNames())
				}

				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name:  "inputs",
					Usage: "Print available input plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(getDeprecatedInputNames())
						} else {
							outputBuffer.Write(getInputNames())
						}
						return nil
					},
				},
				{
					Name:  "outputs",
					Usage: "Print available output plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(getDeprecatedOutputNames())
						} else {
							outputBuffer.Write(getOutputNames())
						}
						return nil
					},
				},
				{
					Name:  "processors",
					Usage: "Print available processor plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(getDeprecatedProcessorNames())
						} else {
							outputBuffer.Write(getProcessorNames())
						}
						return nil
					},
				},
				{
					Name:  "aggregators",
					Usage: "Print available aggregator plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(getDeprecatedAggregatorNames())
						} else {
							outputBuffer.Write(getAggregatorNames())
						}
						return nil
					},
				},
				{
					Name:  "secretstores",
					Usage: "Print available secretstore plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: func(cCtx *cli.Context) error {
						if cCtx.Bool("deprecated") {
							outputBuffer.Write(getDeprecatedSecretstoreNames())
						} else {
							outputBuffer.Write(getSecretstoreNames())
						}
						return nil
					},
				},
			},
		},
	}
}
