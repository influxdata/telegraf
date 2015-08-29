package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/influxdb/telegraf"
	_ "github.com/influxdb/telegraf/outputs/all"
	_ "github.com/KPACHbIuLLIAnO4/telegraf/plugins/all"
)

var fDebug = flag.Bool("debug", false,
	"show metrics as they're generated to stdout")
var fTest = flag.Bool("test", false, "gather metrics, print them out, and exit")
var fConfig = flag.String("config", "", "configuration file to load")
var fVersion = flag.Bool("version", false, "display the version")
var fSampleConfig = flag.Bool("sample-config", false,
	"print out full sample configuration")
var fPidfile = flag.String("pidfile", "", "file to write our pid to")
var fPLuginsFilter = flag.String("filter", "",
	"filter the plugins to enable, separator is :")
var fUsage = flag.String("usage", "",
	"print usage for a plugin, ie, 'telegraf -usage mysql'")

// Telegraf version
//	-ldflags "-X main.Version=`git describe --always --tags`"
var Version string

func main() {
	flag.Parse()

	if *fVersion {
		v := fmt.Sprintf("Telegraf - Version %s", Version)
		fmt.Println(v)
		return
	}

	if *fSampleConfig {
		telegraf.PrintSampleConfig()
		return
	}

	if *fUsage != "" {
		if err := telegraf.PrintPluginConfig(*fUsage); err != nil {
			log.Fatal(err)
		}
		return
	}

	var (
		config *telegraf.Config
		err    error
	)

	if *fConfig != "" {
		config, err = telegraf.LoadConfig(*fConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Usage: Telegraf")
		flag.PrintDefaults()
		return
	}

	ag, err := telegraf.NewAgent(config)
	if err != nil {
		log.Fatal(err)
	}

	if *fDebug {
		ag.Debug = true
	}

	outputs, err := ag.LoadOutputs()
	if err != nil {
		log.Fatal(err)
	}
	if len(outputs) == 0 {
		log.Printf("Error: no outputs found, did you provide a valid config file?")
		os.Exit(1)
	}

	plugins, err := ag.LoadPlugins(*fPLuginsFilter)
	if err != nil {
		log.Fatal(err)
	}
	if len(plugins) == 0 {
		log.Printf("Error: no plugins found, did you provide a valid config file?")
		os.Exit(1)
	}

	if *fTest {
		if *fConfig != "" {
			err = ag.Test()
		} else {
			err = ag.TestAllPlugins()
		}

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
	log.Printf("Loaded outputs: %s", strings.Join(outputs, " "))
	log.Printf("Loaded plugins: %s", strings.Join(plugins, " "))
	if ag.Debug {
		log.Printf("Debug: enabled")
		log.Printf("Agent Config: Interval:%s, Debug:%#v, Hostname:%#v\n",
			ag.Interval, ag.Debug, ag.Hostname)
	}
	log.Printf("Tags enabled: %s", config.ListTags())

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
