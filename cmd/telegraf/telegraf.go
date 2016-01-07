package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/influxdb/telegraf"
	"github.com/influxdb/telegraf/internal/config"
	_ "github.com/influxdb/telegraf/plugins/inputs/all"
	_ "github.com/influxdb/telegraf/plugins/outputs/all"
)

var fDebug = flag.Bool("debug", false,
	"show metrics as they're generated to stdout")
var fTest = flag.Bool("test", false, "gather metrics, print them out, and exit")
var fConfig = flag.String("config", "", "configuration file to load")
var fConfigDirectory = flag.String("config-directory", "",
	"directory containing additional *.conf files")
var fVersion = flag.Bool("version", false, "display the version")
var fSampleConfig = flag.Bool("sample-config", false,
	"print out full sample configuration")
var fPidfile = flag.String("pidfile", "", "file to write our pid to")
var fInputFilters = flag.String("input-filter", "",
	"filter the plugins to enable, separator is :")
var fOutputFilters = flag.String("output-filter", "",
	"filter the outputs to enable, separator is :")
var fUsage = flag.String("usage", "",
	"print usage for a plugin, ie, 'telegraf -usage mysql'")

// Telegraf version
//	-ldflags "-X main.Version=`git describe --always --tags`"
var Version string

const usage = `Telegraf, The plugin-driven server agent for reporting metrics into InfluxDB

Usage:

    telegraf <flags>

The flags are:

    -config <file>     configuration file to load
    -test              gather metrics once, print them to stdout, and exit
    -sample-config     print out full sample configuration to stdout
    -config-directory  directory containing additional *.conf files
    -input-filter      filter the input plugins to enable, separator is :
    -output-filter     filter the output plugins to enable, separator is :
    -usage             print usage for a plugin, ie, 'telegraf -usage mysql'
    -version           print the version to stdout

Examples:

    # generate a telegraf config file:
    telegraf -sample-config > telegraf.conf

    # generate a telegraf config file with only cpu input and influxdb output enabled
    telegraf -sample-config -input-filter cpu -output-filter influxdb

    # run a single telegraf collection, outputting metrics to stdout
    telegraf -config telegraf.conf -test

    # run telegraf with all plugins defined in config file
    telegraf -config telegraf.conf

    # run telegraf, enabling only the cpu and memory inputs and influxdb output
    telegraf -config telegraf.conf -input-filter cpu:mem -output-filter influxdb
`

func main() {
	flag.Usage = usageExit
	flag.Parse()

	if flag.NFlag() == 0 {
		usageExit()
	}

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

	if *fVersion {
		v := fmt.Sprintf("Telegraf - Version %s", Version)
		fmt.Println(v)
		return
	}

	if *fSampleConfig {
		config.PrintSampleConfig(inputFilters, outputFilters)
		return
	}

	if *fUsage != "" {
		if err := config.PrintInputConfig(*fUsage); err != nil {
			if err2 := config.PrintOutputConfig(*fUsage); err2 != nil {
				log.Fatalf("%s and %s", err, err2)
			}
		}
		return
	}

	var (
		c   *config.Config
		err error
	)

	if *fConfig != "" {
		c = config.NewConfig()
		c.OutputFilters = outputFilters
		c.InputFilters = inputFilters
		err = c.LoadConfig(*fConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Usage: Telegraf")
		flag.PrintDefaults()
		return
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
		log.Fatalf("Error: no plugins found, did you provide a valid config file?")
	}

	ag, err := telegraf.NewAgent(c)
	if err != nil {
		log.Fatal(err)
	}

	if *fDebug {
		ag.Config.Agent.Debug = true
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
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(shutdown)
	}()

	log.Printf("Starting Telegraf (version %s)\n", Version)
	log.Printf("Loaded outputs: %s", strings.Join(c.OutputNames(), " "))
	log.Printf("Loaded plugins: %s", strings.Join(c.InputNames(), " "))
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

func usageExit() {
	fmt.Println(usage)
	os.Exit(0)
}
