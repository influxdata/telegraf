package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/cmd/telegraf/commands"
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
	_ "github.com/influxdata/telegraf/plugins/serializers/all"
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
	// This function is used when Telegraf is run with only flags
	action := func(cCtx *cli.Context) error {
		// We do not expect any arguments this is likely a misspelling of
		// a command...
		if cCtx.NArg() > 0 {
			return fmt.Errorf("unknown command %q", cCtx.Args().First())
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
			outputBuffer.Write([]byte("DEPRECATED: use telegraf plugins outputs\n"))
			outputBuffer.Write([]byte("Available Output Plugins:\n"))
			names := make([]string, 0, len(outputs.Outputs))
			for k := range outputs.Outputs {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				fmt.Fprintf(outputBuffer, "  %s\n", k)
			}
			return nil
		// print available input plugins
		case cCtx.Bool("input-list"):
			outputBuffer.Write([]byte("DEPRECATED: use telegraf plugins inputs\n"))
			outputBuffer.Write([]byte("Available Input Plugins:\n"))
			names := make([]string, 0, len(inputs.Inputs))
			for k := range inputs.Inputs {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				fmt.Fprintf(outputBuffer, "  %s\n", k)
			}
			return nil
		// print usage for a plugin, ie, 'telegraf --usage mysql'
		case cCtx.String("usage") != "":
			err := PrintInputConfig(cCtx.String("usage"), outputBuffer)
			err2 := PrintOutputConfig(cCtx.String("usage"), outputBuffer)
			if err != nil && err2 != nil {
				return fmt.Errorf("%w and %w", err, err2)
			}
			return nil
		// DEPRECATED
		case cCtx.Bool("version"):
			fmt.Fprintf(outputBuffer, "%s\n", internal.FormatFullVersion())
			return nil
		// DEPRECATED
		case cCtx.Bool("sample-config"):
			filters := processFilterFlags(cCtx)

			printSampleConfig(outputBuffer, filters)
			return nil
		}

		if cCtx.String("pprof-addr") != "" {
			pprof.Start(cCtx.String("pprof-addr"))
		}

		if cCtx.Bool("strict-env-handling") && cCtx.Bool("non-strict-env-handling") {
			return errors.New("flags --strict-env-handling and --non-strict-env-handling cannot be used together")
		}
		if !cCtx.Bool("strict-env-handling") && !cCtx.Bool("non-strict-env-handling") {
			msg := "Strict environment variable handling is the new default starting with v1.38.0! " +
				"If your configuration does not work with strict handling please explicitly add " +
				"the --non-strict-env-handling flag to switch to the previous behavior!"
			log.Println("W! " + color.YellowString(msg))
		}

		if err := config.SetPluginLabelSelections(cCtx.StringSlice("select")); err != nil {
			return err
		}

		filters := processFilterFlags(cCtx)

		g := GlobalFlags{
			config:                  cCtx.StringSlice("config"),
			configDir:               cCtx.StringSlice("config-directory"),
			testWait:                cCtx.Int("test-wait"),
			configURLRetryAttempts:  cCtx.Int("config-url-retry-attempts"),
			configURLWatchInterval:  cCtx.Duration("config-url-watch-interval"),
			watchConfig:             cCtx.String("watch-config"),
			watchInterval:           cCtx.Duration("watch-interval"),
			watchDebounceInterval:   cCtx.Duration("watch-debounce-interval"),
			pidFile:                 cCtx.String("pidfile"),
			plugindDir:              cCtx.String("plugin-directory"),
			password:                cCtx.String("password"),
			oldEnvBehavior:          cCtx.Bool("old-env-behavior"),
			nonStrictEnvVars:        cCtx.Bool("non-strict-env-handling"),
			printPluginConfigSource: cCtx.Bool("print-plugin-config-source"),
			test:                    cCtx.Bool("test"),
			debug:                   cCtx.Bool("debug"),
			once:                    cCtx.Bool("once"),
			quiet:                   cCtx.Bool("quiet"),
			unprotected:             cCtx.Bool("unprotected"),
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

	printConfig := func(ctx *cli.Context) error {
		printSampleConfig(outputBuffer, processFilterFlags(ctx))
		return nil
	}
	openStores := func(ctx *cli.Context) commands.SecretStores {
		filters := Filters{
			section:     []string{"inputs", "outputs", "processors", "aggregators"},
			input:       []string{"-"},
			output:      []string{"-"},
			processor:   []string{"-"},
			aggregator:  []string{"-"},
			secretstore: processFilterFlags(ctx).secretstore,
		}
		g := GlobalFlags{
			config:     ctx.StringSlice("config"),
			configDir:  ctx.StringSlice("config-directory"),
			plugindDir: ctx.String("plugin-directory"),
			password:   ctx.String("password"),
			debug:      ctx.Bool("debug"),
		}
		m.Init(nil, filters, g, WindowFlags{})
		return m
	}
	app := &cli.App{
		Name:     "Telegraf",
		Usage:    "The plugin-driven server agent for collecting & reporting metrics.",
		Writer:   outputBuffer,
		Flags:    commands.RootFlags(runtime.GOOS),
		Commands: commands.Subcommands(runtime.GOOS, printConfig, openStores),
		Action:   action,
	}
	// Make sure we safely erase secrets
	defer memguard.Purge()
	defer logger.CloseLogging()

	if err := app.Run(args); err != nil {
		log.Printf("E! %s", err)
		return err
	}
	return nil
}

func main() {
	// #13481: disables gh:99designs/keyring kwallet.go from connecting to dbus
	os.Setenv("DISABLE_KWALLET", "1")

	agent := Telegraf{}
	pprof := NewPprofServer()
	c := config.NewConfig()
	if err := runApp(os.Args, os.Stdout, pprof, c, &agent); err != nil {
		os.Exit(1)
	}
}
