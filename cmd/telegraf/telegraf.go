package main

import (
	"context"
	"errors"
	"fmt"
	"log" //nolint:revive
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/fatih/color"
	"github.com/influxdata/tail/watch"
	"gopkg.in/tomb.v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

var stop chan struct{}

type GlobalFlags struct {
	config      []string
	configDir   []string
	testWait    int
	watchConfig string
	pidFile     string
	plugindDir  string
	test        bool
	debug       bool
	once        bool
	quiet       bool
}

type WindowFlags struct {
	service             string
	serviceName         string
	serviceDisplayName  string
	serviceRestartDelay string
	serviceAutoRestart  bool
	console             bool
}

type App interface {
	Init(<-chan error, Filters, GlobalFlags, WindowFlags)
	Run() error
}

type Telegraf struct {
	pprofErr <-chan error

	inputFilters  []string
	outputFilters []string

	GlobalFlags
	WindowFlags
}

func (t *Telegraf) Init(pprofErr <-chan error, f Filters, g GlobalFlags, w WindowFlags) {
	t.pprofErr = pprofErr
	t.inputFilters = f.input
	t.outputFilters = f.output
	t.GlobalFlags = g
	t.WindowFlags = w
}

func (t *Telegraf) reloadLoop() error {
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false
		ctx, cancel := context.WithCancel(context.Background())

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
			syscall.SIGTERM, syscall.SIGINT)
		if t.watchConfig != "" {
			for _, fConfig := range t.config {
				if _, err := os.Stat(fConfig); err == nil {
					go t.watchLocalConfig(signals, fConfig)
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
			case err := <-t.pprofErr:
				log.Printf("E! pprof server failed: %v", err)
				cancel()
			case <-stop:
				cancel()
			}
		}()

		err := t.runAgent(ctx)
		if err != nil && err != context.Canceled {
			return fmt.Errorf("[telegraf] Error running agent: %v", err)
		}
	}

	return nil
}

func (t *Telegraf) watchLocalConfig(signals chan os.Signal, fConfig string) {
	var mytomb tomb.Tomb
	var watcher watch.FileWatcher
	if t.watchConfig == "poll" {
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

func (t *Telegraf) runAgent(ctx context.Context) error {
	// If no other options are specified, load the config file and run.
	c := config.NewConfig()
	c.OutputFilters = t.outputFilters
	c.InputFilters = t.inputFilters
	var err error
	// providing no "config" flag should load default config
	if len(t.config) == 0 {
		err = c.LoadConfig("")
		if err != nil {
			return err
		}
	}
	for _, fConfig := range t.config {
		err = c.LoadConfig(fConfig)
		if err != nil {
			return err
		}
	}

	for _, fConfigDirectory := range t.configDir {
		err = c.LoadDirectory(fConfigDirectory)
		if err != nil {
			return err
		}
	}

	if !(t.test || t.testWait != 0) && len(c.Outputs) == 0 {
		return errors.New("error: no outputs found, did you provide a valid config file?")
	}
	if t.plugindDir == "" && len(c.Inputs) == 0 {
		return errors.New("error: no inputs found, did you provide a valid config file?")
	}

	if int64(c.Agent.Interval) <= 0 {
		return fmt.Errorf("agent interval must be positive, found %v", c.Agent.Interval)
	}

	if int64(c.Agent.FlushInterval) <= 0 {
		return fmt.Errorf("agent flush_interval must be positive; found %v", c.Agent.Interval)
	}

	// Setup logging as configured.
	telegraf.Debug = c.Agent.Debug || t.debug
	logConfig := logger.LogConfig{
		Debug:               telegraf.Debug,
		Quiet:               c.Agent.Quiet || t.quiet,
		LogTarget:           c.Agent.LogTarget,
		Logfile:             c.Agent.Logfile,
		RotationInterval:    c.Agent.LogfileRotationInterval,
		RotationMaxSize:     c.Agent.LogfileRotationMaxSize,
		RotationMaxArchives: c.Agent.LogfileRotationMaxArchives,
		LogWithTimezone:     c.Agent.LogWithTimezone,
	}

	err = logger.SetupLogging(logConfig)
	if err != nil {
		return err
	}

	log.Printf("I! Starting Telegraf %s%s", internal.Version, internal.Customized)
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
	if !t.once && (t.test || t.testWait != 0) {
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

	if t.once {
		wait := time.Duration(t.testWait) * time.Second
		return ag.Once(ctx, wait)
	}

	if t.test || t.testWait != 0 {
		wait := time.Duration(t.testWait) * time.Second
		return ag.Test(ctx, wait)
	}

	if t.pidFile != "" {
		f, err := os.OpenFile(t.pidFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("E! Unable to create pidfile: %s", err)
		} else {
			_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())

			err = f.Close()
			if err != nil {
				return err
			}

			defer func() {
				err := os.Remove(t.pidFile)
				if err != nil {
					log.Printf("E! Unable to remove pidfile: %s", err)
				}
			}()
		}
	}

	return ag.Run(ctx)
}
