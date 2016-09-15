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
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/kardianos/service"
)

var fDebug = flag.Bool("debug", false,
	"show metrics as they're generated to stdout")
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

  telegraf <flags>

The flags are:

  -config <file>     configuration file to load
  -test              gather metrics once, print them to stdout, and exit
  -sample-config     print out full sample configuration to stdout
  -config-directory  directory containing additional *.conf files
  -input-filter      filter the input plugins to enable, separator is :
  -input-list        print all the plugins inputs
  -output-filter     filter the output plugins to enable, separator is :
  -output-list       print all the available outputs
  -usage             print usage for a plugin, ie, 'telegraf -usage mysql'
  -debug             print metrics as they're generated to stdout
  -quiet             run in quiet mode
  -version           print the version to stdout
  -service           Control the service, ie, 'telegraf -service install (windows only)'

In addition to the -config flag, telegraf will also load the config file from
an environment variable or default location. Precedence is:
  1. -config flag
  2. $TELEGRAF_CONFIG_PATH environment variable
  3. $HOME/.telegraf/telegraf.conf
  4. /etc/telegraf/telegraf.conf

Examples:

  # generate a telegraf config file:
  telegraf -sample-config > telegraf.conf

  # generate config with only cpu input & influxdb output plugins defined
  telegraf -sample-config -input-filter cpu -output-filter influxdb

  # run a single telegraf collection, outputing metrics to stdout
  telegraf -config telegraf.conf -test

  # run telegraf with all plugins defined in config file
  telegraf -config telegraf.conf

  # run telegraf, enabling the cpu & memory input, and influxdb output plugins
  telegraf -config telegraf.conf -input-filter cpu:mem -output-filter influxdb
`

var logger service.Logger

var stop chan struct{}

var srvc service.Service
var svcConfig *service.Config

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
		flag.Usage = func() { usageExit(0) }
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

		if len(args) > 0 {
			switch args[0] {
			case "version":
				fmt.Printf("Telegraf v%s (git: %s %s)\n", version, branch, commit)
				return
			case "config":
				config.PrintSampleConfig(inputFilters, outputFilters)
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
			config.PrintSampleConfig(inputFilters, outputFilters)
			return
		case *fUsage != "":
			if err := config.PrintInputConfig(*fUsage); err != nil {
				if err2 := config.PrintOutputConfig(*fUsage); err2 != nil {
					log.Fatalf("%s and %s", err, err2)
				}
			}
			return
		case *fService != "" && runtime.GOOS == "windows":
			if *fConfig != "" {
				(*svcConfig).Arguments = []string{"-config", *fConfig}
			}
			err := service.Control(s, *fService)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		// If no other options are specified, load the config file and run.
		c := config.NewConfig()
		c.OutputFilters = outputFilters
		c.InputFilters = inputFilters
		err := c.LoadConfig(*fConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if *fConfigDirectory != "" {
			err = c.LoadDirectory(*fConfigDirectory)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(c.Outputs) == 0 {
			log.Fatalf("Error: no outputs found, did you provide a valid config file?")
		}
		if len(c.Inputs) == 0 {
			log.Fatalf("Error: no inputs found, did you provide a valid config file?")
		}

		ag, err := agent.NewAgent(c)
		if err != nil {
			log.Fatal(err)
		}

		if *fDebug {
			ag.Config.Agent.Debug = true
		}

		if *fQuiet {
			ag.Config.Agent.Quiet = true
		}

		if *fTest {
			err = ag.Test()
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		err = ag.Connect()
		if err != nil {
			log.Fatal(err)
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
					log.Printf("Reloading Telegraf config\n")
					<-reload
					reload <- true
					close(shutdown)
				}
			case <-stop:
				close(shutdown)
			}
		}()

		log.Printf("Starting Telegraf (version %s)\n", version)
		log.Printf("Loaded outputs: %s", strings.Join(c.OutputNames(), " "))
		log.Printf("Loaded inputs: %s", strings.Join(c.InputNames(), " "))
		log.Printf("Tags enabled: %s", c.ListTags())

		if *fPidfile != "" {
			f, err := os.Create(*fPidfile)
			if err != nil {
				log.Fatalf("Unable to create pidfile: %s", err)
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
	if runtime.GOOS == "windows" {
		svcConfig = &service.Config{
			Name:        "telegraf",
			DisplayName: "Telegraf Data Collector Service",
			Description: "Collects data using a series of plugins and publishes it to" +
				"another series of plugins.",
			Arguments: []string{"-config", "C:\\Program Files\\Telegraf\\telegraf.conf"},
		}

		prg := &program{}
		s, err := service.New(prg, svcConfig)
		if err != nil {
			log.Fatal(err)
		}
		logger, err = s.Logger(nil)
		if err != nil {
			log.Fatal(err)
		}
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
	} else {
		stop = make(chan struct{})
		reloadLoop(stop, nil)
	}
}
