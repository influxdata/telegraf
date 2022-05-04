package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // Comment this line to disable pprof endpoint.
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"

	"github.com/fatih/color"

	"github.com/influxdata/tail/watch"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
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
	"gopkg.in/tomb.v1"
)

type sliceFlags []string

func (i *sliceFlags) String() string {
	s := strings.Join(*i, " ")
	return "[" + s + "]"
}

func (i *sliceFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// If you update these, update usage.go and usage_windows.go
var fDebug = flag.Bool("debug", false,
	"turn on debug logging")
var pprofAddr = flag.String("pprof-addr", "",
	"pprof address to listen on, not activate pprof if empty")
var fQuiet = flag.Bool("quiet", false,
	"run in quiet mode")
var fTest = flag.Bool("test", false, "enable test mode: gather metrics, print them out, and exit. Note: Test mode only runs inputs, not processors, aggregators, or outputs")
var fTestWait = flag.Int("test-wait", 0, "wait up to this many seconds for service inputs to complete in test mode")

var fConfigs sliceFlags
var fConfigDirs sliceFlags
var fWatchConfig = flag.String("watch-config", "", "Monitoring config changes [notify, poll]")
var fVersion = flag.Bool("version", false, "display the version and exit")
var fSampleConfig = flag.Bool("sample-config", false,
	"print out full sample configuration")
var fPidfile = flag.String("pidfile", "", "file to write our pid to")
var fDeprecationList = flag.Bool("deprecation-list", false,
	"print all deprecated plugins or plugin options.")
var fSectionFilters = flag.String("section-filter", "",
	"filter the sections to print, separator is ':'. Valid values are 'agent', 'global_tags', 'outputs', 'processors', 'aggregators' and 'inputs'")
var fInputFilters = flag.String("input-filter", "",
	"filter the inputs to enable, separator is :")
var fInputList = flag.Bool("input-list", false,
	"print available input plugins.")
var fOutputFilters = flag.String("output-filter", "",
	"filter the outputs to enable, separator is :")
var fOutputList = flag.Bool("output-list", false,
	"print available output plugins.")
var fAggregatorFilters = flag.String("aggregator-filter", "",
	"filter the aggregators to enable, separator is :")
var fProcessorFilters = flag.String("processor-filter", "",
	"filter the processors to enable, separator is :")
var fUsage = flag.String("usage", "",
	"print usage for a plugin, ie, 'telegraf --usage mysql'")

//nolint:varcheck,unused // False positive - this var is used for non-default build tag: windows
var fService = flag.String("service", "",
	"operate on the service (windows only)")

//nolint:varcheck,unused // False positive - this var is used for non-default build tag: windows
var fServiceName = flag.String("service-name", "telegraf",
	"service name (windows only)")

//nolint:varcheck,unused // False positive - this var is used for non-default build tag: windows
var fServiceDisplayName = flag.String("service-display-name", "Telegraf Data Collector Service",
	"service display name (windows only)")

//nolint:varcheck,unused // False positive - this var is used for non-default build tag: windows
var fServiceAutoRestart = flag.Bool("service-auto-restart", false,
	"auto restart service on failure (windows only)")

//nolint:varcheck,unused // False positive - this var is used for non-default build tag: windows
var fServiceRestartDelay = flag.String("service-restart-delay", "5m",
	"delay before service auto restart, default is 5m (windows only)")

//nolint:varcheck,unused // False positive - this var is used for non-default build tag: windows
var fRunAsConsole = flag.Bool("console", false,
	"run as console application (windows only)")
var fPlugins = flag.String("plugin-directory", "",
	"path to directory containing external plugins")
var fRunOnce = flag.Bool("once", false, "run one gather and exit")

var (
	version string
	commit  string
	branch  string
)

var stop chan struct{}

func reloadLoop(
	inputFilters []string,
	outputFilters []string,
) {
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false
		ctx, cancel := context.WithCancel(context.Background())

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
			syscall.SIGTERM, syscall.SIGINT)
		if *fWatchConfig != "" {
			for _, fConfig := range fConfigs {
				if _, err := os.Stat(fConfig); err == nil {
					go watchLocalConfig(signals, fConfig)
				} else {
					log.Printf("W! Cannot watch config %s: %s", fConfig, err)
				}
			}
		}
		go func() {
			select {
			case sig := <-signals:
				if sig == syscall.SIGHUP {
					log.Printf("I! Reloading Telegraf config")
					<-reload
					reload <- true
				}
				cancel()
			case <-stop:
				cancel()
			}
		}()

		err := runAgent(ctx, inputFilters, outputFilters)
		if err != nil && err != context.Canceled {
			log.Fatalf("E! [telegraf] Error running agent: %v", err)
		}
	}
}

func watchLocalConfig(signals chan os.Signal, fConfig string) {
	var mytomb tomb.Tomb
	var watcher watch.FileWatcher
	if *fWatchConfig == "poll" {
		watcher = watch.NewPollingFileWatcher(fConfig)
	} else {
		watcher = watch.NewInotifyFileWatcher(fConfig)
	}
	changes, err := watcher.ChangeEvents(&mytomb, 0)
	if err != nil {
		log.Printf("E! Error watching config: %s\n", err)
		return
	}
	log.Println("I! Config watcher started")
	select {
	case <-changes.Modified:
		log.Println("I! Config file modified")
	case <-changes.Deleted:
		// deleted can mean moved. wait a bit a check existence
		<-time.After(time.Second)
		if _, err := os.Stat(fConfig); err == nil {
			log.Println("I! Config file overwritten")
		} else {
			log.Println("W! Config file deleted")
			if err := watcher.BlockUntilExists(&mytomb); err != nil {
				log.Printf("E! Cannot watch for config: %s\n", err.Error())
				return
			}
			log.Println("I! Config file appeared")
		}
	case <-changes.Truncated:
		log.Println("I! Config file truncated")
	case <-mytomb.Dying():
		log.Println("I! Config watcher ended")
		return
	}
	mytomb.Done()
	signals <- syscall.SIGHUP
}

func runAgent(ctx context.Context,
	inputFilters []string,
	outputFilters []string,
) error {
	// If no other options are specified, load the config file and run.
	c := config.NewConfig()
	c.OutputFilters = outputFilters
	c.InputFilters = inputFilters
	var err error
	// providing no "config" flag should load default config
	if len(fConfigs) == 0 {
		err = c.LoadConfig("")
		if err != nil {
			return err
		}
	}
	for _, fConfig := range fConfigs {
		err = c.LoadConfig(fConfig)
		if err != nil {
			return err
		}
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

	log.Printf("I! Starting Telegraf %s", version)
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

	// Configure version
	if err := internal.SetVersion(version); err != nil {
		log.Println("Telegraf version already configured to: " + internal.Version())
	}

	// Load external plugins, if requested.
	if *fPlugins != "" {
		log.Printf("I! Loading external plugins from: %s", *fPlugins)
		if err := goplugin.LoadExternalPlugins(*fPlugins); err != nil {
			log.Fatal("E! " + err.Error())
		}
	}

	if *pprofAddr != "" {
		go func() {
			pprofHostPort := *pprofAddr
			parts := strings.Split(pprofHostPort, ":")
			if len(parts) == 2 && parts[0] == "" {
				pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
			}
			pprofHostPort = "http://" + pprofHostPort + "/debug/pprof"

			log.Printf("I! Starting pprof HTTP server at: %s", pprofHostPort)

			if err := http.ListenAndServe(*pprofAddr, nil); err != nil {
				log.Fatal("E! " + err.Error())
			}
		}()
	}

	if len(args) > 0 {
		switch args[0] {
		case "version":
			fmt.Println(formatFullVersion())
			return
		case "config":
			config.PrintSampleConfig(
				sectionFilters,
				inputFilters,
				outputFilters,
				aggregatorFilters,
				processorFilters,
			)
			return
		}
	}

	// switch for flags which just do something and exit immediately
	switch {
	case *fDeprecationList:
		c := config.NewConfig()
		infos := c.CollectDeprecationInfos(
			inputFilters,
			outputFilters,
			aggregatorFilters,
			processorFilters,
		)
		//nolint:revive // We will notice if Println fails
		fmt.Println("Deprecated Input Plugins: ")
		c.PrintDeprecationList(infos["inputs"])
		//nolint:revive // We will notice if Println fails
		fmt.Println("Deprecated Output Plugins: ")
		c.PrintDeprecationList(infos["outputs"])
		//nolint:revive // We will notice if Println fails
		fmt.Println("Deprecated Processor Plugins: ")
		c.PrintDeprecationList(infos["processors"])
		//nolint:revive // We will notice if Println fails
		fmt.Println("Deprecated Aggregator Plugins: ")
		c.PrintDeprecationList(infos["aggregators"])
		return
	case *fOutputList:
		fmt.Println("Available Output Plugins: ")
		names := make([]string, 0, len(outputs.Outputs))
		for k := range outputs.Outputs {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Printf("  %s\n", k)
		}
		return
	case *fInputList:
		fmt.Println("Available Input Plugins:")
		names := make([]string, 0, len(inputs.Inputs))
		for k := range inputs.Inputs {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Printf("  %s\n", k)
		}
		return
	case *fVersion:
		fmt.Println(formatFullVersion())
		return
	case *fSampleConfig:
		config.PrintSampleConfig(
			sectionFilters,
			inputFilters,
			outputFilters,
			aggregatorFilters,
			processorFilters,
		)
		return
	case *fUsage != "":
		err := config.PrintInputConfig(*fUsage)
		err2 := config.PrintOutputConfig(*fUsage)
		if err != nil && err2 != nil {
			log.Fatalf("E! %s and %s", err, err2)
		}
		return
	}

	run(
		inputFilters,
		outputFilters,
	)
}
