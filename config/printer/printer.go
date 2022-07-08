package printer

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

var (
	// Default sections
	sectionDefaults = []string{"global_tags", "agent", "outputs",
		"processors", "aggregators", "inputs"}

	// Default input plugins
	inputDefaults = []string{"cpu", "mem", "swap", "system", "kernel",
		"processes", "disk", "diskio"}

	// Default output plugins
	outputDefaults = []string{"influxdb"}
)

var header = `# Telegraf Configuration
#
# Telegraf is entirely plugin driven. All metrics are gathered from the
# declared inputs, and sent to the declared outputs.
#
# Plugins must be declared in here to be active.
# To deactivate a plugin, comment out the name and any variables.
#
# Use 'telegraf -config telegraf.conf -test' to see what metrics a config
# file would generate.
#
# Environment variables can be used anywhere in this config file, simply surround
# them with ${}. For strings the variable must be within quotes (ie, "${STR_VAR}"),
# for numbers and booleans they should be plain (ie, ${INT_VAR}, ${BOOL_VAR})

`
var globalTagsConfig = `
# Global tags can be specified here in key="value" format.
[global_tags]
  # dc = "us-east-1" # will tag all metrics with dc=us-east-1
  # rack = "1a"
  ## Environment variables can be used as tags, and throughout the config file
  # user = "$USER"

`

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the agentConfig data.
//go:embed agent.conf
var agentConfig string

var outputHeader = `
###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

`

var processorHeader = `
###############################################################################
#                            PROCESSOR PLUGINS                                #
###############################################################################

`

var aggregatorHeader = `
###############################################################################
#                            AGGREGATOR PLUGINS                               #
###############################################################################

`

var inputHeader = `
###############################################################################
#                            INPUT PLUGINS                                    #
###############################################################################

`

var serviceInputHeader = `
###############################################################################
#                            SERVICE INPUT PLUGINS                            #
###############################################################################

`

func sliceContains(name string, list []string) bool {
	for _, b := range list {
		if b == name {
			return true
		}
	}
	return false
}

// PrintSampleConfig prints the sample config
func PrintSampleConfig(
	sectionFilters []string,
	inputFilters []string,
	outputFilters []string,
	aggregatorFilters []string,
	processorFilters []string,
) {
	// print headers
	fmt.Print(header)

	if len(sectionFilters) == 0 {
		sectionFilters = sectionDefaults
	}
	printFilteredGlobalSections(sectionFilters)

	// print output plugins
	if sliceContains("outputs", sectionFilters) {
		if len(outputFilters) != 0 {
			if len(outputFilters) >= 3 && outputFilters[1] != "none" {
				fmt.Print(outputHeader)
			}
			printFilteredOutputs(outputFilters, false)
		} else {
			fmt.Print(outputHeader)
			printFilteredOutputs(outputDefaults, false)
			// Print non-default outputs, commented
			var pnames []string
			for pname := range outputs.Outputs {
				if !sliceContains(pname, outputDefaults) {
					pnames = append(pnames, pname)
				}
			}
			sort.Strings(pnames)
			printFilteredOutputs(pnames, true)
		}
	}

	// print processor plugins
	if sliceContains("processors", sectionFilters) {
		if len(processorFilters) != 0 {
			if len(processorFilters) >= 3 && processorFilters[1] != "none" {
				fmt.Print(processorHeader)
			}
			printFilteredProcessors(processorFilters, false)
		} else {
			fmt.Print(processorHeader)
			pnames := []string{}
			for pname := range processors.Processors {
				pnames = append(pnames, pname)
			}
			sort.Strings(pnames)
			printFilteredProcessors(pnames, true)
		}
	}

	// print aggregator plugins
	if sliceContains("aggregators", sectionFilters) {
		if len(aggregatorFilters) != 0 {
			if len(aggregatorFilters) >= 3 && aggregatorFilters[1] != "none" {
				fmt.Print(aggregatorHeader)
			}
			printFilteredAggregators(aggregatorFilters, false)
		} else {
			fmt.Print(aggregatorHeader)
			pnames := []string{}
			for pname := range aggregators.Aggregators {
				pnames = append(pnames, pname)
			}
			sort.Strings(pnames)
			printFilteredAggregators(pnames, true)
		}
	}

	// print input plugins
	if sliceContains("inputs", sectionFilters) {
		if len(inputFilters) != 0 {
			if len(inputFilters) >= 3 && inputFilters[1] != "none" {
				fmt.Print(inputHeader)
			}
			printFilteredInputs(inputFilters, false)
		} else {
			fmt.Print(inputHeader)
			printFilteredInputs(inputDefaults, false)
			// Print non-default inputs, commented
			var pnames []string
			for pname := range inputs.Inputs {
				if !sliceContains(pname, inputDefaults) {
					pnames = append(pnames, pname)
				}
			}
			sort.Strings(pnames)
			printFilteredInputs(pnames, true)
		}
	}
}

// PluginNameCounts returns a list of sorted plugin names and their count
func PluginNameCounts(plugins []string) []string {
	names := make(map[string]int)
	for _, plugin := range plugins {
		names[plugin]++
	}

	var namecount []string
	for name, count := range names {
		if count == 1 {
			namecount = append(namecount, name)
		} else {
			namecount = append(namecount, fmt.Sprintf("%s (%dx)", name, count))
		}
	}

	sort.Strings(namecount)
	return namecount
}

func printFilteredProcessors(processorFilters []string, commented bool) {
	// Filter processors
	var pnames []string
	for pname := range processors.Processors {
		if sliceContains(pname, processorFilters) {
			pnames = append(pnames, pname)
		}
	}
	sort.Strings(pnames)

	// Print Outputs
	for _, pname := range pnames {
		creator := processors.Processors[pname]
		output := creator()
		printConfig(pname, output, "processors", commented, processors.Deprecations[pname])
	}
}

func printFilteredAggregators(aggregatorFilters []string, commented bool) {
	// Filter outputs
	var anames []string
	for aname := range aggregators.Aggregators {
		if sliceContains(aname, aggregatorFilters) {
			anames = append(anames, aname)
		}
	}
	sort.Strings(anames)

	// Print Outputs
	for _, aname := range anames {
		creator := aggregators.Aggregators[aname]
		output := creator()
		printConfig(aname, output, "aggregators", commented, aggregators.Deprecations[aname])
	}
}

func printFilteredInputs(inputFilters []string, commented bool) {
	// Filter inputs
	var pnames []string
	for pname := range inputs.Inputs {
		if sliceContains(pname, inputFilters) {
			pnames = append(pnames, pname)
		}
	}
	sort.Strings(pnames)

	// cache service inputs to print them at the end
	servInputs := make(map[string]telegraf.ServiceInput)
	// for alphabetical looping:
	servInputNames := []string{}

	// Print Inputs
	for _, pname := range pnames {
		// Skip inputs that are registered twice for backward compatibility
		switch pname {
		case "cisco_telemetry_gnmi", "io", "KNXListener":
			continue
		}
		creator := inputs.Inputs[pname]
		input := creator()

		if p, ok := input.(telegraf.ServiceInput); ok {
			servInputs[pname] = p
			servInputNames = append(servInputNames, pname)
			continue
		}

		printConfig(pname, input, "inputs", commented, inputs.Deprecations[pname])
	}

	// Print Service Inputs
	if len(servInputs) == 0 {
		return
	}
	sort.Strings(servInputNames)

	fmt.Print(serviceInputHeader)
	for _, name := range servInputNames {
		printConfig(name, servInputs[name], "inputs", commented, inputs.Deprecations[name])
	}
}

func printFilteredOutputs(outputFilters []string, commented bool) {
	// Filter outputs
	var onames []string
	for oname := range outputs.Outputs {
		if sliceContains(oname, outputFilters) {
			onames = append(onames, oname)
		}
	}
	sort.Strings(onames)

	// Print Outputs
	for _, oname := range onames {
		creator := outputs.Outputs[oname]
		output := creator()
		printConfig(oname, output, "outputs", commented, outputs.Deprecations[oname])
	}
}

func printFilteredGlobalSections(sectionFilters []string) {
	if sliceContains("global_tags", sectionFilters) {
		fmt.Print(globalTagsConfig)
	}

	if sliceContains("agent", sectionFilters) {
		fmt.Print(agentConfig)
	}
}

func printConfig(name string, p telegraf.PluginDescriber, op string, commented bool, di telegraf.DeprecationInfo) {
	comment := ""
	if commented {
		comment = "# "
	}

	if di.Since != "" {
		removalNote := ""
		if di.RemovalIn != "" {
			removalNote = " and will be removed in " + di.RemovalIn
		}
		fmt.Printf("\n%s ## DEPRECATED: The '%s' plugin is deprecated in version %s%s, %s.", comment, name, di.Since, removalNote, di.Notice)
	}

	config := p.SampleConfig()
	if config == "" {
		fmt.Printf("\n#[[%s.%s]]", op, name)
		fmt.Printf("\n%s  # no configuration\n\n", comment)
	} else {
		lines := strings.Split(config, "\n")
		fmt.Print("\n")
		for i, line := range lines {
			if i == len(lines)-1 {
				fmt.Print("\n")
				continue
			}
			fmt.Print(strings.TrimRight(comment+line, " ") + "\n")
		}
	}
}

// PrintInputConfig prints the config usage of a single input.
func PrintInputConfig(name string) error {
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("input %s not found", name)
	}

	printConfig(name, creator(), "inputs", false, inputs.Deprecations[name])
	return nil
}

// PrintOutputConfig prints the config usage of a single output.
func PrintOutputConfig(name string) error {
	creator, ok := outputs.Outputs[name]
	if !ok {
		return fmt.Errorf("output %s not found", name)
	}

	printConfig(name, creator(), "outputs", false, outputs.Deprecations[name])
	return nil
}
