package agent

import (
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
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

	if a.Config.Agent.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		a.Config.Agent.Hostname = hostname
	}

	config.Tags["host"] = a.Config.Agent.Hostname

	return a, nil
}

// Connect connects to all configured outputs
func (a *Agent) Connect() error {
	for _, o := range a.Config.Outputs {
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
			log.Printf("Failed to connect to output %s, retrying in 15s, error was '%s' \n", o.Name, err)
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

// gatherParallel runs the inputs that are using the same reporting interval
// as the telegraf agent.
func (a *Agent) gatherParallel(metricC chan telegraf.Metric) error {
	var wg sync.WaitGroup

	start := time.Now()
	counter := 0
	jitter := a.Config.Agent.CollectionJitter.Duration.Nanoseconds()
	for _, input := range a.Config.Inputs {
		if input.Config.Interval != 0 {
			continue
		}

		wg.Add(1)
		counter++
		go func(input *internal_models.RunningInput) {
			defer panicRecover(input)
			defer wg.Done()

			acc := NewAccumulator(input.Config, metricC, a.Config.Agent.Interval.Duration)
			acc.SetDebug(a.Config.Agent.Debug)
			acc.setDefaultTags(a.Config.Tags)

			if jitter != 0 {
				nanoSleep := rand.Int63n(jitter)
				d, err := time.ParseDuration(fmt.Sprintf("%dns", nanoSleep))
				if err != nil {
					log.Printf("Jittering collection interval failed for plugin %s",
						input.Name)
				} else {
					time.Sleep(d)
				}
			}

			if err := input.Input.Gather(acc); err != nil {
				log.Printf("Error in input [%s]: %s", input.Name, err)
			}

		}(input)
	}

	if counter == 0 {
		return nil
	}

	wg.Wait()

	elapsed := time.Since(start)
	if !a.Config.Agent.Quiet {
		log.Printf("Gathered metrics, (%s interval), from %d inputs in %s\n",
			a.Config.Agent.Interval.Duration, counter, elapsed)
	}
	return nil
}

// gatherSeparate runs the inputs that have been configured with their own
// reporting interval.
func (a *Agent) gatherSeparate(
	shutdown chan struct{},
	input *internal_models.RunningInput,
	metricC chan telegraf.Metric,
) error {
	defer panicRecover(input)

	ticker := time.NewTicker(input.Config.Interval)

	for {
		var outerr error
		start := time.Now()

		acc := NewAccumulator(input.Config, metricC, input.Config.Interval)
		acc.SetDebug(a.Config.Agent.Debug)
		acc.setDefaultTags(a.Config.Tags)

		if err := input.Input.Gather(acc); err != nil {
			log.Printf("Error in input [%s]: %s", input.Name, err)
		}

		elapsed := time.Since(start)
		if !a.Config.Agent.Quiet {
			log.Printf("Gathered metrics, (separate %s interval), from %s in %s\n",
				input.Config.Interval, input.Name, elapsed)
		}

		if outerr != nil {
			return outerr
		}

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
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
		acc := NewAccumulator(input.Config, metricC, input.Config.Interval)
		acc.SetDebug(true)

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

// flush writes a list of points to all configured outputs
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

// flusher monitors the points input channel and flushes on the minimum interval
func (a *Agent) flusher(shutdown chan struct{}, metricC chan telegraf.Metric) error {
	// Inelegant, but this sleep is to allow the Gather threads to run, so that
	// the flusher will flush after metrics are collected.
	time.Sleep(time.Millisecond * 200)

	ticker := time.NewTicker(a.Config.Agent.FlushInterval.Duration)

	for {
		select {
		case <-shutdown:
			log.Println("Hang on, flushing any cached points before shutdown")
			a.flush()
			return nil
		case <-ticker.C:
			a.flush()
		case m := <-metricC:
			for _, o := range a.Config.Outputs {
				o.AddPoint(m)
			}
		}
	}
}

// jitterInterval applies the the interval jitter to the flush interval using
// crypto/rand number generator
func jitterInterval(ininterval, injitter time.Duration) time.Duration {
	var jitter int64
	outinterval := ininterval
	if injitter.Nanoseconds() != 0 {
		maxjitter := big.NewInt(injitter.Nanoseconds())
		if j, err := cryptorand.Int(cryptorand.Reader, maxjitter); err == nil {
			jitter = j.Int64()
		}
		outinterval = time.Duration(jitter + ininterval.Nanoseconds())
	}

	if outinterval.Nanoseconds() < time.Duration(500*time.Millisecond).Nanoseconds() {
		log.Printf("Flush interval %s too low, setting to 500ms\n", outinterval)
		outinterval = time.Duration(500 * time.Millisecond)
	}

	return outinterval
}

// Run runs the agent daemon, gathering every Interval
func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup

	a.Config.Agent.FlushInterval.Duration = jitterInterval(
		a.Config.Agent.FlushInterval.Duration,
		a.Config.Agent.FlushJitter.Duration)

	log.Printf("Agent Config: Interval:%s, Debug:%#v, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s \n",
		a.Config.Agent.Interval.Duration, a.Config.Agent.Debug, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	// channel shared between all input threads for accumulating points
	metricC := make(chan telegraf.Metric, 1000)

	// Round collection to nearest interval by sleeping
	if a.Config.Agent.RoundInterval {
		i := int64(a.Config.Agent.Interval.Duration)
		time.Sleep(time.Duration(i - (time.Now().UnixNano() % i)))
	}
	ticker := time.NewTicker(a.Config.Agent.Interval.Duration)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(shutdown, metricC); err != nil {
			log.Printf("Flusher routine failed, exiting: %s\n", err.Error())
			close(shutdown)
		}
	}()

	for _, input := range a.Config.Inputs {

		// Start service of any ServicePlugins
		switch p := input.Input.(type) {
		case telegraf.ServiceInput:
			if err := p.Start(); err != nil {
				log.Printf("Service for input %s failed to start, exiting\n%s\n",
					input.Name, err.Error())
				return err
			}
			defer p.Stop()
		}

		// Special handling for inputs that have their own collection interval
		// configured. Default intervals are handled below with gatherParallel
		if input.Config.Interval != 0 {
			wg.Add(1)
			go func(input *internal_models.RunningInput) {
				defer wg.Done()
				if err := a.gatherSeparate(shutdown, input, metricC); err != nil {
					log.Printf(err.Error())
				}
			}(input)
		}
	}

	defer wg.Wait()

	for {
		if err := a.gatherParallel(metricC); err != nil {
			log.Printf(err.Error())
		}

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
}
