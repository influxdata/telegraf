package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/influxdb/tivan"
	_ "github.com/influxdb/tivan/plugins/all"
)

var fDebug = flag.Bool("debug", false, "show metrics as they're generated to stdout")
var fTest = flag.Bool("test", false, "gather metrics, print them out, and exit")
var fConfig = flag.String("config", "", "configuration file to load")
var fVersion = flag.Bool("version", false, "display the version")

var Version = "unreleased"

func main() {
	flag.Parse()

	if *fVersion {
		fmt.Printf("InfluxDB Tivan agent - Version %s\n", Version)
		return
	}

	var (
		config *tivan.Config
		err    error
	)

	if *fConfig != "" {
		config, err = tivan.LoadConfig(*fConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		config = tivan.DefaultConfig()
	}

	ag, err := tivan.NewAgent(config)
	if err != nil {
		log.Fatal(err)
	}

	if *fDebug {
		ag.Debug = true
	}

	plugins, err := ag.LoadPlugins()
	if err != nil {
		log.Fatal(err)
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

	log.Print("InfluxDB Agent running")
	log.Printf("Loaded plugins: %s", strings.Join(plugins, " "))
	if ag.Debug {
		log.Printf("Debug: enabled")
		log.Printf("Agent Config: %#v", ag)
	}

	if config.URL != "" {
		log.Printf("Sending metrics to: %s", config.URL)
		log.Printf("Tags enabled: %v", config.ListTags())
	}

	ag.Run(shutdown)
}
