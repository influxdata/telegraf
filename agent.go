package telegraf

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/influxdb/telegraf/internal/config"
	"github.com/influxdb/telegraf/outputs"
	"github.com/influxdb/telegraf/plugins"

	"github.com/influxdb/influxdb/client/v2"
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
		case outputs.ServiceOutput:
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
			log.Printf("Failed to connect to output %s, retrying in 15s\n", o.Name)
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
		case outputs.ServiceOutput:
			ot.Stop()
		}
	}
	return err
}

// gatherParallel runs the plugins that are using the same reporting interval
// as the telegraf agent.
func (a *Agent) gatherParallel(pointChan chan *client.Point) error {
	var wg sync.WaitGroup

	start := time.Now()
	counter := 0
	for _, plugin := range a.Config.Plugins {
		if plugin.Config.Interval != 0 {
			continue
		}

		wg.Add(1)
		counter++
		go func(plugin *config.RunningPlugin) {
			defer wg.Done()

			acc := NewAccumulator(plugin.Config, pointChan)
			acc.SetDebug(a.Config.Agent.Debug)
			acc.SetPrefix(plugin.Name + "_")
			acc.SetDefaultTags(a.Config.Tags)

			if err := plugin.Plugin.Gather(acc); err != nil {
				log.Printf("Error in plugin [%s]: %s", plugin.Name, err)
			}

		}(plugin)
	}

	if counter == 0 {
		return nil
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("Gathered metrics, (%s interval), from %d plugins in %s\n",
		a.Config.Agent.Interval, counter, elapsed)
	return nil
}

// gatherSeparate runs the plugins that have been configured with their own
// reporting interval.
func (a *Agent) gatherSeparate(
	shutdown chan struct{},
	plugin *config.RunningPlugin,
	pointChan chan *client.Point,
) error {
	ticker := time.NewTicker(plugin.Config.Interval)

	for {
		var outerr error
		start := time.Now()

		acc := NewAccumulator(plugin.Config, pointChan)
		acc.SetDebug(a.Config.Agent.Debug)
		acc.SetPrefix(plugin.Name + "_")
		acc.SetDefaultTags(a.Config.Tags)

		if err := plugin.Plugin.Gather(acc); err != nil {
			log.Printf("Error in plugin [%s]: %s", plugin.Name, err)
		}

		elapsed := time.Since(start)
		log.Printf("Gathered metrics, (separate %s interval), from %s in %s\n",
			plugin.Config.Interval, plugin.Name, elapsed)

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

// Test verifies that we can 'Gather' from all plugins with their configured
// Config struct
func (a *Agent) Test() error {
	shutdown := make(chan struct{})
	defer close(shutdown)
	pointChan := make(chan *client.Point)

	// dummy receiver for the point channel
	go func() {
		for {
			select {
			case <-pointChan:
				// do nothing
			case <-shutdown:
				return
			}
		}
	}()

	for _, plugin := range a.Config.Plugins {
		acc := NewAccumulator(plugin.Config, pointChan)
		acc.SetDebug(true)
		acc.SetPrefix(plugin.Name + "_")

		fmt.Printf("* Plugin: %s, Collection 1\n", plugin.Name)
		if plugin.Config.Interval != 0 {
			fmt.Printf("* Internal: %s\n", plugin.Config.Interval)
		}

		if err := plugin.Plugin.Gather(acc); err != nil {
			return err
		}

		// Special instructions for some plugins. cpu, for example, needs to be
		// run twice in order to return cpu usage percentages.
		switch plugin.Name {
		case "cpu", "mongodb":
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("* Plugin: %s, Collection 2\n", plugin.Name)
			if err := plugin.Plugin.Gather(acc); err != nil {
				return err
			}
		}

	}
	return nil
}

// writeOutput writes a list of points to a single output, with retries.
// Optionally takes a `done` channel to indicate that it is done writing.
func (a *Agent) writeOutput(
	points []*client.Point,
	ro *config.RunningOutput,
	shutdown chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if len(points) == 0 {
		return
	}
	retry := 0
	retries := a.Config.Agent.FlushRetries
	start := time.Now()

	for {
		err := ro.Output.Write(points)
		if err == nil {
			// Write successful
			elapsed := time.Since(start)
			log.Printf("Flushed %d metrics to output %s in %s\n",
				len(points), ro.Name, elapsed)
			return
		}

		select {
		case <-shutdown:
			return
		default:
			if retry >= retries {
				// No more retries
				msg := "FATAL: Write to output [%s] failed %d times, dropping" +
					" %d metrics\n"
				log.Printf(msg, ro.Name, retries+1, len(points))
				return
			} else if err != nil {
				// Sleep for a retry
				log.Printf("Error in output [%s]: %s, retrying in %s",
					ro.Name, err.Error(), a.Config.Agent.FlushInterval.Duration)
				time.Sleep(a.Config.Agent.FlushInterval.Duration)
			}
		}

		retry++
	}
}

// flush writes a list of points to all configured outputs
func (a *Agent) flush(
	points []*client.Point,
	shutdown chan struct{},
	wait bool,
) {
	var wg sync.WaitGroup
	for _, o := range a.Config.Outputs {
		wg.Add(1)
		go a.writeOutput(points, o, shutdown, &wg)
	}
	if wait {
		wg.Wait()
	}
}

// flusher monitors the points input channel and flushes on the minimum interval
func (a *Agent) flusher(shutdown chan struct{}, pointChan chan *client.Point) error {
	// Inelegant, but this sleep is to allow the Gather threads to run, so that
	// the flusher will flush after metrics are collected.
	time.Sleep(time.Millisecond * 100)

	ticker := time.NewTicker(a.Config.Agent.FlushInterval.Duration)
	points := make([]*client.Point, 0)

	for {
		select {
		case <-shutdown:
			log.Println("Hang on, flushing any cached points before shutdown")
			a.flush(points, shutdown, true)
			return nil
		case <-ticker.C:
			a.flush(points, shutdown, false)
			points = make([]*client.Point, 0)
		case pt := <-pointChan:
			points = append(points, pt)
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
		if j, err := rand.Int(rand.Reader, maxjitter); err == nil {
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

	a.Config.Agent.FlushInterval.Duration = jitterInterval(a.Config.Agent.FlushInterval.Duration,
		a.Config.Agent.FlushJitter.Duration)

	log.Printf("Agent Config: Interval:%s, Debug:%#v, Hostname:%#v, "+
		"Flush Interval:%s\n",
		a.Config.Agent.Interval, a.Config.Agent.Debug,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval)

	// channel shared between all plugin threads for accumulating points
	pointChan := make(chan *client.Point, 1000)

	// Round collection to nearest interval by sleeping
	if a.Config.Agent.RoundInterval {
		i := int64(a.Config.Agent.Interval.Duration)
		time.Sleep(time.Duration(i - (time.Now().UnixNano() % i)))
	}
	ticker := time.NewTicker(a.Config.Agent.Interval.Duration)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(shutdown, pointChan); err != nil {
			log.Printf("Flusher routine failed, exiting: %s\n", err.Error())
			close(shutdown)
		}
	}()

	for _, plugin := range a.Config.Plugins {

		// Start service of any ServicePlugins
		switch p := plugin.Plugin.(type) {
		case plugins.ServicePlugin:
			if err := p.Start(); err != nil {
				log.Printf("Service for plugin %s failed to start, exiting\n%s\n",
					plugin.Name, err.Error())
				return err
			}
			defer p.Stop()
		}

		// Special handling for plugins that have their own collection interval
		// configured. Default intervals are handled below with gatherParallel
		if plugin.Config.Interval != 0 {
			wg.Add(1)
			go func(plugin *config.RunningPlugin) {
				defer wg.Done()
				if err := a.gatherSeparate(shutdown, plugin, pointChan); err != nil {
					log.Printf(err.Error())
				}
			}(plugin)
		}
	}

	defer wg.Wait()

	for {
		if err := a.gatherParallel(pointChan); err != nil {
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
