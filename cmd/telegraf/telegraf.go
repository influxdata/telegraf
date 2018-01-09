package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // Comment this line to disable pprof endpoint.
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/logger"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
	"github.com/kardianos/service"
)

var fDebug = flag.Bool("debug", false,
	"turn on debug logging")
var pprofAddr = flag.String("pprof-addr", "",
	"pprof address to listen on, not activate pprof if empty")
var fQuiet = flag.Bool("quiet", false,
	"run in quiet mode")
var fTest = flag.Bool("test", false, "gather metrics, print them out, and exit")
var fConfig = flag.String("config", "", "configuration file to load")
var fConfigDirectory = flag.String("config-directory", "",
	"directory containing additional *.conf files")
var fVersion = flag.Bool("version", false, "display the version")
var fSampleConfig = flag.Bool("sample-config", false,
	"print out full sample configuration")
var fPidfile = flag.String("pidfile", "", "file to write our pid to")
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
var fService = flag.String("service", "",
	"operate on the service")

var (
	nextVersion = "1.5.0"
	version     string
	commit      string
	branch      string
)

func init() {
	// If commit or branch are not set, make that clear.
	if commit == "" {
		commit = "unknown"
	}
	if branch == "" {
		branch = "unknown"
	}
}

const usage = `Telegraf, The plugin-driven server agent for collecting and reporting metrics.

Usage:

  telegraf [commands|flags]

The commands & flags are:

  config              print out full sample configuration to stdout
  version             print the version to stdout

  --config <file>     configuration file to load
  --test              gather metrics once, print them to stdout, and exit
  --config-directory  directory containing additional *.conf files
  --input-filter      filter the input plugins to enable, separator is :
  --output-filter     filter the output plugins to enable, separator is :
  --usage             print usage for a plugin, ie, 'telegraf --usage mysql'
  --debug             print metrics as they're generated to stdout
  --pprof-addr        pprof address to listen on, format: localhost:6060 or :6060
  --quiet             run in quiet mode

Examples:

  # generate a telegraf config file:
  telegraf config > telegraf.conf

  # generate config with only cpu input & influxdb output plugins defined
  telegraf --input-filter cpu --output-filter influxdb config

  # run a single telegraf collection, outputing metrics to stdout
  telegraf --config telegraf.conf --test

  # run telegraf with all plugins defined in config file
  telegraf --config telegraf.conf

  # run telegraf, enabling the cpu & memory input, and influxdb output plugins
  telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb

  # run telegraf with pprof
  telegraf --config telegraf.conf --pprof-addr localhost:6060
`

var stop chan struct{}

func reloadLoop(
	stop chan struct{},
	inputFilters []string,
	outputFilters []string,
	aggregatorFilters []string,
	processorFilters []string,
) {
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false

		// If no other options are specified, load the config file and run.
		c := config.NewConfig()
		c.OutputFilters = outputFilters
		c.InputFilters = inputFilters
		err := c.LoadConfig(*fConfig)
		if err != nil {
			log.Fatal("E! " + err.Error())
		}

		if *fConfigDirectory != "" {
			err = c.LoadDirectory(*fConfigDirectory)
			if err != nil {
				log.Fatal("E! " + err.Error())
			}
		}
		if !*fTest && len(c.Outputs) == 0 {
			log.Fatalf("E! Error: no outputs found, did you provide a valid config file?")
		}
		if len(c.Inputs) == 0 {
			log.Fatalf("E! Error: no inputs found, did you provide a valid config file?")
		}

		if int64(c.Agent.Interval.Duration) <= 0 {
			log.Fatalf("E! Agent interval must be positive, found %s",
				c.Agent.Interval.Duration)
		}

		if int64(c.Agent.FlushInterval.Duration) <= 0 {
			log.Fatalf("E! Agent flush_interval must be positive; found %s",
				c.Agent.Interval.Duration)
		}

		ag, err := agent.NewAgent(c)
		if err != nil {
			log.Fatal("E! " + err.Error())
		}

		// Setup logging
		logger.SetupLogging(
			ag.Config.Agent.Debug || *fDebug,
			ag.Config.Agent.Quiet || *fQuiet,
			ag.Config.Agent.Logfile,
		)

		if *fTest {
			err = ag.Test()
			if err != nil {
				log.Fatal("E! " + err.Error())
			}
			os.Exit(0)
		}

		err = ag.Connect()
		if err != nil {
			log.Fatal("E! " + err.Error())
		}

		shutdown := make(chan struct{})
		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP)
		go func() {
			select {
			case sig := <-signals:
				if sig == os.Interrupt {
					close(shutdown)
				}
				if sig == syscall.SIGHUP {
					log.Printf("I! Reloading Telegraf config\n")
					<-reload
					reload <- true
					close(shutdown)
				}
			case <-stop:
				close(shutdown)
			}
		}()

		log.Printf("I! Starting Telegraf %s\n", displayVersion())
		log.Printf("I! Loaded outputs: %s", strings.Join(c.OutputNames(), " "))
		log.Printf("I! Loaded inputs: %s", strings.Join(c.InputNames(), " "))
		log.Printf("I! Tags enabled: %s", c.ListTags())

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

		ag.Run(shutdown)
	}
}

func usageExit(rc int) {
	fmt.Println(usage)
	os.Exit(rc)
}

type program struct {
	inputFilters      []string
	outputFilters     []string
	aggregatorFilters []string
	processorFilters  []string
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {
	stop = make(chan struct{})
	reloadLoop(
		stop,
		p.inputFilters,
		p.outputFilters,
		p.aggregatorFilters,
		p.processorFilters,
	)
}
func (p *program) Stop(s service.Service) error {
	close(stop)
	return nil
}

func displayVersion() string {
	if version == "" {
		return fmt.Sprintf("v%s~%s", nextVersion, commit)
	}
	return "v" + version
}

func main() {
	flag.Usage = func() { usageExit(0) }
	flag.Parse()
	args := flag.Args()

	inputFilters, outputFilters := []string{}, []string{}
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
			fmt.Printf("Telegraf %s (git: %s %s)\n", displayVersion(), branch, commit)
			return
		case "config":
			config.PrintSampleConfig(
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
	case *fOutputList:
		fmt.Println("Available Output Plugins:")
		for k, _ := range outputs.Outputs {
			fmt.Printf("  %s\n", k)
		}
		return
	case *fInputList:
		fmt.Println("Available Input Plugins:")
		for k, _ := range inputs.Inputs {
			fmt.Printf("  %s\n", k)
		}
		return
	case *fVersion:
		fmt.Printf("Telegraf %s (git: %s %s)\n", displayVersion(), branch, commit)
		return
	case *fSampleConfig:
		config.PrintSampleConfig(
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

	if runtime.GOOS == "windows" {
		svcConfig := &service.Config{
			Name:        "telegraf",
			DisplayName: "Telegraf Data Collector Service",
			Description: "Collects data using a series of plugins and publishes it to" +
				"another series of plugins.",
			Arguments: []string{"-config", "C:\\Program Files\\Telegraf\\telegraf.conf"},
		}

		prg := &program{
			inputFilters:      inputFilters,
			outputFilters:     outputFilters,
			aggregatorFilters: aggregatorFilters,
			processorFilters:  processorFilters,
		}
		s, err := service.New(prg, svcConfig)
		if err != nil {
			log.Fatal("E! " + err.Error())
		}
		// Handle the -service flag here to prevent any issues with tooling that
		// may not have an interactive session, e.g. installing from Ansible.
		if *fService != "" {
			if *fConfig != "" {
				(*svcConfig).Arguments = []string{"-config", *fConfig}
			}
			if *fConfigDirectory != "" {
				(*svcConfig).Arguments = append((*svcConfig).Arguments, "-config-directory", *fConfigDirectory)
			}
			err := service.Control(s, *fService)
			if err != nil {
				log.Fatal("E! " + err.Error())
			}
			os.Exit(0)
		} else {
			err = s.Run()
			if err != nil {
				log.Println("E! " + err.Error())
			}
		}
	} else {
		stop = make(chan struct{})
		reloadLoop(
			stop,
			inputFilters,
			outputFilters,
			aggregatorFilters,
			processorFilters,
		)
	}
}
