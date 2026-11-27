package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/internal"
)

// Filters selects the configuration sections and plugins used by a command.
type Filters struct {
	Section     []string
	Input       []string
	Output      []string
	Aggregator  []string
	Processor   []string
	Secretstore []string
}

type commandActions struct {
	version             cli.ActionFunc
	configCreate        cli.ActionFunc
	configCheck         cli.ActionFunc
	configMigrate       cli.ActionFunc
	secretsList         cli.ActionFunc
	secretsGet          cli.ActionFunc
	secretsSet          cli.ActionFunc
	secretsRemove       cli.ActionFunc
	plugins             cli.ActionFunc
	pluginsInputs       cli.ActionFunc
	pluginsOutputs      cli.ActionFunc
	pluginsProcessors   cli.ActionFunc
	pluginsAggregators  cli.ActionFunc
	pluginsSecretstores cli.ActionFunc
	pluginsParsers      cli.ActionFunc
	pluginsSerializers  cli.ActionFunc
	service             serviceActions
}

// ProcessFilterFlags combines filters from the command and its parent.
func ProcessFilterFlags(ctx *cli.Context) Filters {
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

	return Filters{
		Section:     splitFilter(section),
		Input:       splitFilter(input),
		Output:      splitFilter(output),
		Aggregator:  splitFilter(aggregator),
		Processor:   splitFilter(processor),
		Secretstore: splitFilter(secretstore),
	}
}

func versionAction(ctx *cli.Context) error {
	fmt.Fprintln(ctx.App.Writer, internal.FormatFullVersion())
	return nil
}

func getGlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "config",
			Usage: "configuration file to load",
		},
		&cli.StringSliceFlag{
			Name:  "config-directory",
			Usage: "directory containing additional *.conf files",
		},
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
			Usage: "filter the secret stores to enable, separator is ':'",
		},
		&cli.BoolFlag{
			Name:  "strict-env-handling",
			Usage: "enforces strict and secure handling of environment variables; will not work with non-string settings",
		},
		&cli.BoolFlag{
			Name:  "non-strict-env-handling",
			Usage: "allow unsafe non-strict handling of environment variables to replace non-string settings",
		},
	}
}

func getRunFlags() []cli.Flag {
	return []cli.Flag{
		// Int flags
		&cli.IntFlag{
			Name:  "test-wait",
			Usage: "wait up to this many seconds for service inputs to complete in test mode",
		},
		&cli.IntFlag{
			Name: "config-url-retry-attempts",
			Usage: "Number of attempts to obtain a remote configuration via a URL during startup. " +
				"Set to -1 for unlimited attempts.",
			DefaultText: "3",
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
			Name: "watch-config",
			Usage: "monitoring config changes [notify, poll] of --config and --config-directory options. " +
				"Notify supports linux, *bsd, and macOS. Poll is required for Windows and checks every 250ms.",
		},
		&cli.DurationFlag{
			Name:        "watch-debounce-interval",
			Usage:       "Time duration to wait after a config change before reloading",
			DefaultText: "0s",
			Value:       0,
		},
		&cli.StringFlag{
			Name:  "pidfile",
			Usage: "file to write our pid to",
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "password to unlock secret stores",
		},
		//
		// Bool flags
		&cli.BoolFlag{
			Name:  "old-env-behavior",
			Usage: "switch back to pre v1.27 environment replacement behavior",
		},
		&cli.BoolFlag{
			Name:  "print-plugin-config-source",
			Usage: "print the source for a given plugin",
		},
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
			Name:  "unprotected",
			Usage: "do not protect secrets in memory",
		},
		&cli.BoolFlag{
			Name: "test",
			Usage: "enable test mode: gather metrics, print them out, and exit. " +
				"Note: Test mode only runs inputs, processors, and aggregators, but not outputs",
		},
		&cli.StringSliceFlag{
			Name: "select",
			Usage: "enable only plugins with labels matching the given key-value selection. " +
				"If no selectors are provided, all plugins are enabled. Multiple key-value pairs " +
				"in an option will be combined by AND, multiple options are combined by OR. " +
				"Key and value are separated by an equal sign, multiple pairs are separated by " +
				"semi-colon, values do accept wildcards.",
		},
		//
		// Duration flags
		&cli.DurationFlag{
			Name: "watch-interval",
			Usage: "Time duration to check for updates to config files specified by --config and " +
				"--config-directory options. Use with '--watch-config poll'",
			DefaultText: "disabled",
		},
		&cli.DurationFlag{
			Name:        "config-url-watch-interval",
			Usage:       "Time duration to check for updates to URL based configuration files",
			DefaultText: "disabled",
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
	}
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

func splitFilter(value string) []string {
	return slices.DeleteFunc(strings.Split(value, ":"), func(s string) bool { return s == "" })
}
