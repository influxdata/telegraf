package agent

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/internal/models"
)

// Agent runs telegraf and collects data based on the given config
type Agent struct {
	Config *config.Config
}

// NewAgent returns an Agent struct based off the given Config
func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}

	if !a.Config.Agent.OmitHostname {
		if a.Config.Agent.Hostname == "" {
			hostname, err := os.Hostname()
			if err != nil {
				return nil, err
			}

			a.Config.Agent.Hostname = hostname
		}

		config.Tags["host"] = a.Config.Agent.Hostname
	}

	return a, nil
}

// Connect connects to all configured outputs
func (a *Agent) Connect() error {
	for _, o := range a.Config.Outputs {
		o.Quiet = a.Config.Agent.Quiet

		switch ot := o.Output.(type) {
		case telegraf.ServiceOutput:
			if err := ot.Start(); err != nil {
				log.Printf("Service for output %s failed to start, exiting\n%s\n",
					o.Name, err.Error())
				return err
			}
		}

		if a.Config.Agent.Debug {
			log.Printf("Attempting connection to output: %s\n", o.Name)
		}
		err := o.Output.Connect()
		if err != nil {
			log.Printf("Failed to connect to output %s, retrying in 15s, "+
				"error was '%s' \n", o.Name, err)
			time.Sleep(15 * time.Second)
			err = o.Output.Connect()
			if err != nil {
				return err
			}
		}
		if a.Config.Agent.Debug {
			log.Printf("Successfully connected to output: %s\n", o.Name)
		}
	}
	return nil
}

// Close closes the connection to all configured outputs
func (a *Agent) Close() error {
	var err error
	for _, o := range a.Config.Outputs {
		err = o.Output.Close()
		switch ot := o.Output.(type) {
		case telegraf.ServiceOutput:
			ot.Stop()
		}
	}
	return err
}

func panicRecover(input *internal_models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("FATAL: Input [%s] panicked: %s, Stack:\n%s\n",
			input.Name, err, trace)
		log.Println("PLEASE REPORT THIS PANIC ON GITHUB with " +
			"stack trace, configuration, and OS information: " +
			"https://github.com/influxdata/telegraf/issues/new")
	}
}

// gatherer runs the inputs that have been configured with their own
// reporting interval.
func (a *Agent) gatherer(
	shutdown chan struct{},
	input *internal_models.RunningInput,
	interval time.Duration,
	metricC chan telegraf.Metric,
) error {
	defer panicRecover(input)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		var outerr error

		acc := NewAccumulator(input.Config, metricC)
		acc.SetDebug(a.Config.Agent.Debug)
		acc.SetPrecision(a.Config.Agent.Precision.Duration,
			a.Config.Agent.Interval.Duration)
		acc.setDefaultTags(a.Config.Tags)

		internal.RandomSleep(a.Config.Agent.CollectionJitter.Duration, shutdown)

		start := time.Now()
		gatherWithTimeout(shutdown, input, acc, interval)
		elapsed := time.Since(start)

		if outerr != nil {
			return outerr
		}
		if a.Config.Agent.Debug {
			log.Printf("Input [%s] gathered metrics, (%s interval) in %s\n",
				input.Name, interval, elapsed)
		}

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
}

// gatherWithTimeout gathers from the given input, with the given timeout.
//   when the given timeout is reached, gatherWithTimeout logs an error message
//   but continues waiting for it to return. This is to avoid leaving behind
//   hung processes, and to prevent re-calling the same hung process over and
//   over.
func gatherWithTimeout(
	shutdown chan struct{},
	input *internal_models.RunningInput,
	acc *accumulator,
	timeout time.Duration,
) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	done := make(chan error)
	go func() {
		done <- input.Input.Gather(acc)
	}()

	for {
		select {
		case err := <-done:
			if err != nil {
				log.Printf("ERROR in input [%s]: %s", input.Name, err)
			}
			return
		case <-ticker.C:
			log.Printf("ERROR: input [%s] took longer to collect than "+
				"collection interval (%s)",
				input.Name, timeout)
			continue
		case <-shutdown:
			return
		}
	}
}

// Test verifies that we can 'Gather' from all inputs with their configured
// Config struct
func (a *Agent) Test() error {
	shutdown := make(chan struct{})
	defer close(shutdown)
	metricC := make(chan telegraf.Metric)

	// dummy receiver for the point channel
	go func() {
		for {
			select {
			case <-metricC:
				// do nothing
			case <-shutdown:
				return
			}
		}
	}()

	for _, input := range a.Config.Inputs {
		acc := NewAccumulator(input.Config, metricC)
		acc.SetTrace(true)
		acc.SetPrecision(a.Config.Agent.Precision.Duration,
			a.Config.Agent.Interval.Duration)
		acc.setDefaultTags(a.Config.Tags)

		fmt.Printf("* Plugin: %s, Collection 1\n", input.Name)
		if input.Config.Interval != 0 {
			fmt.Printf("* Internal: %s\n", input.Config.Interval)
		}

		if err := input.Input.Gather(acc); err != nil {
			return err
		}

		// Special instructions for some inputs. cpu, for example, needs to be
		// run twice in order to return cpu usage percentages.
		switch input.Name {
		case "cpu", "mongodb", "procstat":
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("* Plugin: %s, Collection 2\n", input.Name)
			if err := input.Input.Gather(acc); err != nil {
				return err
			}
		}

	}
	return nil
}

// flush writes a list of metrics to all configured outputs
func (a *Agent) flush() {
	var wg sync.WaitGroup

	wg.Add(len(a.Config.Outputs))
	for _, o := range a.Config.Outputs {
		go func(output *internal_models.RunningOutput) {
			defer wg.Done()
			err := output.Write()
			if err != nil {
				log.Printf("Error writing to output [%s]: %s\n",
					output.Name, err.Error())
			}
		}(o)
	}

	wg.Wait()
}

// flusher monitors the metrics input channel and flushes on the minimum interval
func (a *Agent) flusher(shutdown chan struct{}, metricC chan telegraf.Metric) error {
	// Inelegant, but this sleep is to allow the Gather threads to run, so that
	// the flusher will flush after metrics are collected.
	time.Sleep(time.Millisecond * 200)

	ticker := time.NewTicker(a.Config.Agent.FlushInterval.Duration)

	for {
		select {
		case <-shutdown:
			log.Println("Hang on, flushing any cached metrics before shutdown")
			a.flush()
			return nil
		case <-ticker.C:
			internal.RandomSleep(a.Config.Agent.FlushJitter.Duration, shutdown)
			a.flush()
		case m := <-metricC:
			for i, o := range a.Config.Outputs {
				if i == len(a.Config.Outputs)-1 {
					o.AddMetric(m)
				} else {
					o.AddMetric(copyMetric(m))
				}
			}
		}
	}
}

func copyMetric(m telegraf.Metric) telegraf.Metric {
	t := time.Time(m.Time())

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	for k, v := range m.Tags() {
		tags[k] = v
	}
	for k, v := range m.Fields() {
		fields[k] = v
	}

	out, _ := telegraf.NewMetric(m.Name(), tags, fields, t)
	return out
}

// Run runs the agent daemon, gathering every Interval
func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup

	log.Printf("Agent Config: Interval:%s, Debug:%#v, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s \n",
		a.Config.Agent.Interval.Duration, a.Config.Agent.Debug, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	// channel shared between all input threads for accumulating metrics
	metricC := make(chan telegraf.Metric, 10000)

	for _, input := range a.Config.Inputs {
		// Start service of any ServicePlugins
		switch p := input.Input.(type) {
		case telegraf.ServiceInput:
			acc := NewAccumulator(input.Config, metricC)
			acc.SetDebug(a.Config.Agent.Debug)
			// Service input plugins should set their own precision of their
			// metrics.
			acc.DisablePrecision()
			acc.setDefaultTags(a.Config.Tags)
			if err := p.Start(acc); err != nil {
				log.Printf("Service for input %s failed to start, exiting\n%s\n",
					input.Name, err.Error())
				return err
			}
			defer p.Stop()
		}
	}

	// Round collection to nearest interval by sleeping
	if a.Config.Agent.RoundInterval {
		i := int64(a.Config.Agent.Interval.Duration)
		time.Sleep(time.Duration(i - (time.Now().UnixNano() % i)))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(shutdown, metricC); err != nil {
			log.Printf("Flusher routine failed, exiting: %s\n", err.Error())
			close(shutdown)
		}
	}()

	wg.Add(len(a.Config.Inputs))
	for _, input := range a.Config.Inputs {
		interval := a.Config.Agent.Interval.Duration
		// overwrite global interval if this plugin has it's own.
		if input.Config.Interval != 0 {
			interval = input.Config.Interval
		}
		go func(in *internal_models.RunningInput, interv time.Duration) {
			defer wg.Done()
			if err := a.gatherer(shutdown, in, interv, metricC); err != nil {
				log.Printf(err.Error())
			}
		}(input, interval)
	}

	wg.Wait()
	return nil
}
