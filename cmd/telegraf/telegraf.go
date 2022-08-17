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
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/goplugin"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/plugins/aggregators"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	"github.com/influxdata/telegraf/plugins/processors"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
)

type TelegrafConfig interface {
	CollectDeprecationInfos([]string, []string, []string, []string) map[string][]config.PluginDeprecationInfo
	PrintDeprecationList([]config.PluginDeprecationInfo)
}

	for _, fConfigDirectory := range fConfigDirs {
		err = c.LoadDirectory(fConfigDirectory)
		if err != nil {
			return err
		}
	}

	if !(*fTest || *fTestWait != 0) && len(c.Outputs) == 0 {
		return errors.New("Error: no outputs found, did you provide a valid config file?")
	}
	if *fPlugins == "" && len(c.Inputs) == 0 {
		return errors.New("Error: no inputs found, did you provide a valid config file?")
	}

	if int64(c.Agent.Interval) <= 0 {
		return fmt.Errorf("Agent interval must be positive, found %v", c.Agent.Interval)
	}

	if int64(c.Agent.FlushInterval) <= 0 {
		return fmt.Errorf("Agent flush_interval must be positive; found %v", c.Agent.Interval)
	}

	// Setup logging as configured.
	telegraf.Debug = c.Agent.Debug || *fDebug
	logConfig := logger.LogConfig{
		Debug:               telegraf.Debug,
		Quiet:               c.Agent.Quiet || *fQuiet,
		LogTarget:           c.Agent.LogTarget,
		Logfile:             c.Agent.Logfile,
		RotationInterval:    c.Agent.LogfileRotationInterval,
		RotationMaxSize:     c.Agent.LogfileRotationMaxSize,
		RotationMaxArchives: c.Agent.LogfileRotationMaxArchives,
		LogWithTimezone:     c.Agent.LogWithTimezone,
	}

	logger.SetupLogging(logConfig)

	log.Printf("I! Starting Telegraf %s%s", internal.Version(), internal.Customized)
	log.Printf("I! Available plugins: %d inputs, %d aggregators, %d processors, %d parsers, %d outputs",
		len(inputs.Inputs),
		len(aggregators.Aggregators),
		len(processors.Processors),
		len(parsers.Parsers),
		len(outputs.Outputs),
	)
	log.Printf("I! Loaded inputs: %s", strings.Join(c.InputNames(), " "))
	log.Printf("I! Loaded aggregators: %s", strings.Join(c.AggregatorNames(), " "))
	log.Printf("I! Loaded processors: %s", strings.Join(c.ProcessorNames(), " "))
	if !*fRunOnce && (*fTest || *fTestWait != 0) {
		log.Print("W! " + color.RedString("Outputs are not used in testing mode!"))
	} else {
		log.Printf("I! Loaded outputs: %s", strings.Join(c.OutputNames(), " "))
	}
	log.Printf("I! Tags enabled: %s", c.ListTags())

	if count, found := c.Deprecations["inputs"]; found && (count[0] > 0 || count[1] > 0) {
		log.Printf("W! Deprecated inputs: %d and %d options", count[0], count[1])
	}
	if count, found := c.Deprecations["aggregators"]; found && (count[0] > 0 || count[1] > 0) {
		log.Printf("W! Deprecated aggregators: %d and %d options", count[0], count[1])
	}
	if count, found := c.Deprecations["processors"]; found && (count[0] > 0 || count[1] > 0) {
		log.Printf("W! Deprecated processors: %d and %d options", count[0], count[1])
	}
	if count, found := c.Deprecations["outputs"]; found && (count[0] > 0 || count[1] > 0) {
		log.Printf("W! Deprecated outputs: %d and %d options", count[0], count[1])
	}

	ag, err := agent.NewAgent(c)
	if err != nil {
		return err
	}

	// Notify systemd that telegraf is ready
	// SdNotify() only tries to notify if the NOTIFY_SOCKET environment is set, so it's safe to call when systemd isn't present.
	// Ignore the return values here because they're not valid for platforms that don't use systemd.
	// For platforms that use systemd, telegraf doesn't log if the notification failed.
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)

	if *fRunOnce {
		wait := time.Duration(*fTestWait) * time.Second
		return ag.Once(ctx, wait)
	}

	if *fTest || *fTestWait != 0 {
		wait := time.Duration(*fTestWait) * time.Second
		return ag.Test(ctx, wait)
	}

	if *fPidfile != "" {
		f, err := os.OpenFile(*fPidfile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("E! Unable to create pidfile: %s", err)
		} else {
			fmt.Fprintf(f, "%d\n", os.Getpid())

			f.Close()

			defer func() {
				err := os.Remove(*fPidfile)
				if err != nil {
					log.Printf("E! Unable to remove pidfile: %s", err)
				}
			}()
		}
	}

	return ag.Run(ctx)
}

func usageExit(rc int) {
	fmt.Println(internal.Usage)
	os.Exit(rc)
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

func main() {
	flag.Var(&fConfigs, "config", "configuration file to load")
	flag.Var(&fConfigDirs, "config-directory", "directory containing additional *.conf files")

	flag.Usage = func() { usageExit(0) }
	flag.Parse()
	args := flag.Args()

	sectionFilters, inputFilters, outputFilters := []string{}, []string{}, []string{}
	if *fSectionFilters != "" {
		sectionFilters = strings.Split(":"+strings.TrimSpace(*fSectionFilters)+":", ":")
	}
	if *fInputFilters != "" {
		inputFilters = strings.Split(":"+strings.TrimSpace(*fInputFilters)+":", ":")
	}
	if *fOutputFilters != "" {
		outputFilters = strings.Split(":"+strings.TrimSpace(*fOutputFilters)+":", ":")
	}

	aggregatorFilters, processorFilters := []string{}, []string{}
	if *fAggregatorFilters != "" {
		aggregatorFilters = strings.Split(":"+strings.TrimSpace(*fAggregatorFilters)+":", ":")
	}
	if *fProcessorFilters != "" {
		processorFilters = strings.Split(":"+strings.TrimSpace(*fProcessorFilters)+":", ":")
	}

	logger.SetupLogging(logger.LogConfig{})

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
				err := PrintInputConfig(cCtx.String("usage"), outputBuffer)
				err2 := PrintOutputConfig(cCtx.String("usage"), outputBuffer)
				if err != nil && err2 != nil {
					return fmt.Errorf("E! %s and %s", err, err2)
				}
				return nil
			// DEPRECATED
			case cCtx.Bool("version"):
				fmt.Println(internal.FormatFullVersion())
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
					_, _ = outputBuffer.Write([]byte(internal.FormatFullVersion()))
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
