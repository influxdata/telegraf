package main

import (
	"flag"
	"fmt"
	"log"
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
	"print usage for a plugin, ie, 'telegraf -usage mysql'")
var fService = flag.String("service", "",
	"operate on the service")

// Telegraf version, populated linker.
//   ie, -ldflags "-X main.version=`git describe --always --tags`"
var (
	version string
	commit  string
	branch  string
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

  config             print out full sample configuration to stdout
  version            print the version to stdout

  --config <file>     configuration file to load
  --test              gather metrics once, print them to stdout, and exit
  --config-directory  directory containing additional *.conf files
  --input-filter      filter the input plugins to enable, separator is :
  --output-filter     filter the output plugins to enable, separator is :
  --usage             print usage for a plugin, ie, 'telegraf --usage mysql'
  --debug             print metrics as they're generated to stdout
  --quiet             run in quiet mode

Examples:

  # generate a telegraf config file:
  telegraf config > telegraf.conf

  # generate config with only cpu input & influxdb output plugins defined
  telegraf --input-filter cpu --output-filter influxdb config

  # run a single telegraf collection, outputing metrics to stdout
  telegraf --config telegraf.conf -test

  # run telegraf with all plugins defined in config file
  telegraf --config telegraf.conf

  # run telegraf, enabling the cpu & memory input, and influxdb output plugins
  telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb
`

var stop chan struct{}

var srvc service.Service

type program struct{}

func reloadLoop(stop chan struct{}, s service.Service) {
	defer func() {
		if service.Interactive() {
			os.Exit(0)
		}
		return
	}()
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false
		flag.Parse()
		args := flag.Args()

		var inputFilters []string
		if *fInputFilters != "" {
			inputFilter := strings.TrimSpace(*fInputFilters)
			inputFilters = strings.Split(":"+inputFilter+":", ":")
		}
		var outputFilters []string
		if *fOutputFilters != "" {
			outputFilter := strings.TrimSpace(*fOutputFilters)
			outputFilters = strings.Split(":"+outputFilter+":", ":")
		}
		var aggregatorFilters []string
		if *fAggregatorFilters != "" {
			aggregatorFilter := strings.TrimSpace(*fAggregatorFilters)
			aggregatorFilters = strings.Split(":"+aggregatorFilter+":", ":")
		}
		var processorFilters []string
		if *fProcessorFilters != "" {
			processorFilter := strings.TrimSpace(*fProcessorFilters)
			processorFilters = strings.Split(":"+processorFilter+":", ":")
		}

		if len(args) > 0 {
			switch args[0] {
			case "version":
				fmt.Printf("Telegraf v%s (git: %s %s)\n", version, branch, commit)
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
			fmt.Printf("Telegraf v%s (git: %s %s)\n", version, branch, commit)
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
			if err := config.PrintInputConfig(*fUsage); err != nil {
				if err2 := config.PrintOutputConfig(*fUsage); err2 != nil {
					log.Fatalf("E! %s and %s", err, err2)
				}
			}
			return
		}

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
		if len(c.Outputs) == 0 {
			log.Fatalf("E! Error: no outputs found, did you provide a valid config file?")
		}
		if len(c.Inputs) == 0 {
			log.Fatalf("E! Error: no inputs found, did you provide a valid config file?")
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
			return
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

		log.Printf("I! Starting Telegraf (version %s)\n", version)
		log.Printf("I! Loaded outputs: %s", strings.Join(c.OutputNames(), " "))
		log.Printf("I! Loaded inputs: %s", strings.Join(c.InputNames(), " "))
		log.Printf("I! Tags enabled: %s", c.ListTags())

		if *fPidfile != "" {
			f, err := os.Create(*fPidfile)
			if err != nil {
				log.Fatalf("E! Unable to create pidfile: %s", err)
			}

			fmt.Fprintf(f, "%d\n", os.Getpid())

			f.Close()
		}

		ag.Run(shutdown)
	}
}

func usageExit(rc int) {
	fmt.Println(usage)
	os.Exit(rc)
}

func (p *program) Start(s service.Service) error {
	srvc = s
	go p.run()
	return nil
}
func (p *program) run() {
	stop = make(chan struct{})
	reloadLoop(stop, srvc)
}
func (p *program) Stop(s service.Service) error {
	close(stop)
	return nil
}

func main() {
	flag.Usage = func() { usageExit(0) }
	flag.Parse()
	if runtime.GOOS == "windows" {
		svcConfig := &service.Config{
			Name:        "telegraf",
			DisplayName: "Telegraf Data Collector Service",
			Description: "Collects data using a series of plugins and publishes it to" +
				"another series of plugins.",
			Arguments: []string{"-config", "C:\\Program Files\\Telegraf\\telegraf.conf"},
		}

		prg := &program{}
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
			err := service.Control(s, *fService)
			if err != nil {
				log.Fatal("E! " + err.Error())
			}
		} else {
			err = s.Run()
			if err != nil {
				log.Println("E! " + err.Error())
			}
		}
	} else {
		stop = make(chan struct{})
		reloadLoop(stop, nil)
	}
}
