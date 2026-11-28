package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"

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
			filters := commands.ProcessFilterFlags(cCtx)
			infos := c.CollectDeprecationInfos(
				filters.Input, filters.Output, filters.Aggregator, filters.Processor,
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
			err := commands.PrintInputConfig(cCtx.String("usage"), outputBuffer)
			err2 := commands.PrintOutputConfig(cCtx.String("usage"), outputBuffer)
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
			return commands.ConfigCreateAction(cCtx)
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

		filters := commands.ProcessFilterFlags(cCtx)

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

	app := &cli.App{
		Name:     "Telegraf",
		Usage:    "The plugin-driven server agent for collecting & reporting metrics.",
		Writer:   outputBuffer,
		Flags:    commands.RootFlags(runtime.GOOS),
		Commands: commands.Subcommands(runtime.GOOS, m),
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
