package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/internal/etcd"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
)

var fDebug = flag.Bool("debug", false,
	"show metrics as they're generated to stdout")
var fQuiet = flag.Bool("quiet", false,
	"run in quiet mode")
var fTest = flag.Bool("test", false, "gather metrics, print them out, and exit")
var fEtcd = flag.String("etcd", "", "etcd urls where configuration is stored (comma separated)")
var fEtcdFolder = flag.String("etcdfolder", "/telegraf", "etcd root folder where configuration is stored")
var fEtcdSendConfigDir = flag.String("etcdwriteconfigdir", "", "store the following config dir to etcd")
var fEtcdSendConfig = flag.String("etcdwriteconfig", "", "store the following config file to etcd")
var fEtcdEraseConfig = flag.Bool("etcderaseconfig", false, "erase all telegraf config in etcd")
var fEtcdSendLabel = flag.String("etcdwritelabel", "", "store config file to etcd with this label")
var fEtcdReadLabels = flag.String("etcdreadlabels", "", "read config from etcd using labels (comma-separated)")
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
var fInputFiltersLegacy = flag.String("filter", "",
	"filter the inputs to enable, separator is :")
var fOutputFiltersLegacy = flag.String("outputfilter", "",
	"filter the outputs to enable, separator is :")
var fConfigDirectoryLegacy = flag.String("configdirectory", "",
	"directory containing additional *.conf files")

// Telegraf version
//	-ldflags "-X main.Version=`git describe --always --tags`"
var Version string

const usage = `Telegraf, The plugin-driven server agent for collecting and reporting metrics.

Usage:

  telegraf <flags>

The flags are:

  -config <file>      configuration file to load
  -test               gather metrics once, print them to stdout, and exit
  -sample-config      print out full sample configuration to stdout
  -config-directory   directory containing additional *.conf files
  -etcd               etcd urls where configuration is stored (comma separated)
  -etcdfolder         etcd folder where configuration is stored and read
  -etcdwriteconfigdir store the following config dir to etcd
  -etcdwriteconfig    store the following config file to etcd
  -etcdwritelabel     store config file to etcd with this label
  -etcdreadlabels     read config from etcd using labels (comma-separated)
  -etcderaseconfig    erase all telegraf config in etcd
  -input-filter       filter the input plugins to enable, separator is :
  -input-list         print all the plugins inputs
  -output-filter      filter the output plugins to enable, separator is :
  -output-list        print all the available outputs
  -usage              print usage for a plugin, ie, 'telegraf -usage mysql'
  -debug              print metrics as they're generated to stdout
  -quiet              run in quiet mode
  -version            print the version to stdout

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

func main() {
	// Read flags
	flag.Usage = func() { usageExit(0) }
	flag.Parse()
	args := flag.Args()
	if flag.NFlag() == 0 && len(args) == 0 {
		usageExit(0)
	}

	// Prepare signals handling
	reload := make(chan bool, 1)
	reload <- true
	shutdown := make(chan struct{})
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP)

	// Prepare etcd if needed
	var e *etcd.EtcdClient
	if *fEtcd != "" {
		e = etcd.NewEtcdClient(*fEtcd, *fEtcdFolder)
		if *fEtcdSendConfig == "" && *fEtcdSendLabel == "" && *fEtcdSendConfigDir == "" {
			go e.LaunchWatcher(shutdown, signals)
		}
	}

	// Handle signals
	go func() {
		for {
			sig := <-signals
			if sig == os.Interrupt {
				close(shutdown)
			} else if sig == syscall.SIGHUP {
				log.Print("Reloading Telegraf config\n")
				<-reload
				reload <- true
				close(shutdown)
			}
		}
	}()

	// Prepare inputs
	var inputFilters []string
	if *fInputFiltersLegacy != "" {
		inputFilter := strings.TrimSpace(*fInputFiltersLegacy)
		inputFilters = strings.Split(":"+inputFilter+":", ":")
	}
	if *fInputFilters != "" {
		inputFilter := strings.TrimSpace(*fInputFilters)
		inputFilters = strings.Split(":"+inputFilter+":", ":")
	}

	// Prepare outputs
	var outputFilters []string
	if *fOutputFiltersLegacy != "" {
		outputFilter := strings.TrimSpace(*fOutputFiltersLegacy)
		outputFilters = strings.Split(":"+outputFilter+":", ":")
	}
	if *fOutputFilters != "" {
		outputFilter := strings.TrimSpace(*fOutputFilters)
		outputFilters = strings.Split(":"+outputFilter+":", ":")
	}

	if len(args) > 0 {
		switch args[0] {
		case "version":
			v := fmt.Sprintf("Telegraf - Version %s", Version)
			fmt.Println(v)
			return
		case "config":
			config.PrintSampleConfig(inputFilters, outputFilters)
			return
		}
	}

	if *fOutputList {
		fmt.Println("Available Output Plugins:")
		for k, _ := range outputs.Outputs {
			fmt.Printf("  %s\n", k)
		}
		return
	}

	if *fInputList {
		fmt.Println("Available Input Plugins:")
		for k, _ := range inputs.Inputs {
			fmt.Printf("  %s\n", k)
		}
		return
	}

	// Print version
	if *fVersion {
		v := fmt.Sprintf("Telegraf - Version %s", Version)
		fmt.Println(v)
		return
	}

	// Print sample config
	if *fSampleConfig {
		config.PrintSampleConfig(inputFilters, outputFilters)
		return
	}

	// Print usage
	if *fUsage != "" {
		if err := config.PrintInputConfig(*fUsage); err != nil {
			if err2 := config.PrintOutputConfig(*fUsage); err2 != nil {
				log.Fatalf("%s and %s", err, err2)
			}
		}
		return
	}

	for <-reload {
		// Reset signal handler vars
		shutdown = make(chan struct{})
		reload <- false

		// Prepare config
		var (
			c   *config.Config
			err error
		)

		if *fEtcd != "" {
			c = config.NewConfig()
			c.OutputFilters = outputFilters
			c.InputFilters = inputFilters

			if *fEtcdSendConfigDir != "" {
				// TODO check config format before write it
				// Erase config in etcd
				if *fEtcdEraseConfig {
					err = e.DeleteConfig("")
					if err != nil {
						err = fmt.Errorf("Error erasing Telegraf Etcd Config: %s", err)
						log.Fatal(err)
					}
				}
				// Write config dir to etcd
				err = c.LoadDirectory(*fEtcdSendConfigDir)
				if err != nil {
					log.Fatal(err)
				}
				err = e.WriteConfigDir(*fEtcdSendConfigDir)
				if err != nil {
					log.Fatal(err)
				}
				return
			} else if *fEtcdSendConfig != "" && *fEtcdSendLabel != "" {
				// TODO check config format before write it
				// Write config to etcd
				err = c.LoadConfig(*fEtcdSendConfig)
				if err != nil {
					log.Fatal(err)
				}
				err = e.WriteLabelConfig(*fEtcdSendLabel, *fEtcdSendConfig)
				if err != nil {
					log.Fatal(err)
				}
				return
			} else if *fEtcdEraseConfig {
				// Erase config in etcd
				err = e.DeleteConfig("")
				if err != nil {
					err = fmt.Errorf("Error erasing Telegraf Etcd Config: %s", err)
					log.Fatal(err)
				}
				return
			} else {
				// Read config to etcd
				log.Printf("Config read from etcd with labels %s\n", *fEtcdReadLabels)
				c, err = e.ReadConfig(c, *fEtcdReadLabels)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else if *fConfig != "" {
			// Read config from file
			c = config.NewConfig()
			c.OutputFilters = outputFilters
			c.InputFilters = inputFilters
			err = c.LoadConfig(*fConfig)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("You must specify a config file. See telegraf --help")
			os.Exit(1)
		}

		// Read config dir
		if *fConfigDirectoryLegacy != "" {
			err = c.LoadDirectory(*fConfigDirectoryLegacy)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Read config dir
		if *fConfigDirectory != "" {
			err = c.LoadDirectory(*fConfigDirectory)
			if err != nil {
				log.Fatal(err)
			}
		}
		// check config
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

		log.Printf("Starting Telegraf (version %s)\n", Version)
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
