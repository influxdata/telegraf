package telegraf

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/influxdb/telegraf/duration"
	"github.com/influxdb/telegraf/outputs"
	"github.com/influxdb/telegraf/plugins"

	"github.com/influxdb/influxdb/client/v2"
)

type runningOutput struct {
	name   string
	output outputs.Output
}

type runningPlugin struct {
	name   string
	plugin plugins.Plugin
	config *ConfiguredPlugin
}

// Agent runs telegraf and collects data based on the given config
type Agent struct {

	// Interval at which to gather information
	Interval duration.Duration

	// RoundInterval rounds collection interval to 'interval'.
	//     ie, if Interval=10s then always collect on :00, :10, :20, etc.
	RoundInterval bool

	// Interval at which to flush data
	FlushInterval duration.Duration

	// FlushRetries is the number of times to retry each data flush
	FlushRetries int

	// FlushJitter tells
	FlushJitter duration.Duration

	// TODO(cam): Remove UTC and Precision parameters, they are no longer
	// valid for the agent config. Leaving them here for now for backwards-
	// compatability

	// Option for outputting data in UTC
	UTC bool `toml:"utc"`

	// Precision to write data at
	// Valid values for Precision are n, u, ms, s, m, and h
	Precision string

	// Option for running in debug mode
	Debug    bool
	Hostname string

	Tags map[string]string

	outputs []*runningOutput
	plugins []*runningPlugin
}

// NewAgent returns an Agent struct based off the given Config
func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{
		Tags:          make(map[string]string),
		Interval:      duration.Duration{10 * time.Second},
		RoundInterval: true,
		FlushInterval: duration.Duration{10 * time.Second},
		FlushRetries:  2,
		FlushJitter:   duration.Duration{5 * time.Second},
	}

	// Apply the toml table to the agent config, overriding defaults
	err := config.ApplyAgent(agent)
	if err != nil {
		return nil, err
	}

	if agent.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		agent.Hostname = hostname
	}

	agent.Tags["host"] = agent.Hostname

	return agent, nil
}

// Connect connects to all configured outputs
func (a *Agent) Connect() error {
	for _, o := range a.outputs {
		switch ot := o.output.(type) {
		case outputs.ServiceOutput:
			if err := ot.Start(); err != nil {
				log.Printf("Service for output %s failed to start, exiting\n%s\n",
					o.name, err.Error())
				return err
			}
		}

		if a.Debug {
			log.Printf("Attempting connection to output: %s\n", o.name)
		}
		err := o.output.Connect()
		if err != nil {
			log.Printf("Failed to connect to output %s, retrying in 15s\n", o.name)
			time.Sleep(15 * time.Second)
			err = o.output.Connect()
			if err != nil {
				return err
			}
		}
		if a.Debug {
			log.Printf("Successfully connected to output: %s\n", o.name)
		}
	}
	return nil
}

// Close closes the connection to all configured outputs
func (a *Agent) Close() error {
	var err error
	for _, o := range a.outputs {
		err = o.output.Close()
		switch ot := o.output.(type) {
		case outputs.ServiceOutput:
			ot.Stop()
		}
	}
	return err
}

// LoadOutputs loads the agent's outputs
func (a *Agent) LoadOutputs(filters []string, config *Config) ([]string, error) {
	var names []string

	for _, name := range config.OutputsDeclared() {
		creator, ok := outputs.Outputs[name]
		if !ok {
			return nil, fmt.Errorf("Undefined but requested output: %s", name)
		}

		if sliceContains(name, filters) || len(filters) == 0 {
			if a.Debug {
				log.Println("Output Enabled: ", name)
			}
			output := creator()

			err := config.ApplyOutput(name, output)
			if err != nil {
				return nil, err
			}

			a.outputs = append(a.outputs, &runningOutput{name, output})
			names = append(names, name)
		}
	}

	sort.Strings(names)

	return names, nil
}

// LoadPlugins loads the agent's plugins
func (a *Agent) LoadPlugins(filters []string, config *Config) ([]string, error) {
	var names []string

	for _, name := range config.PluginsDeclared() {
		creator, ok := plugins.Plugins[name]
		if !ok {
			return nil, fmt.Errorf("Undefined but requested plugin: %s", name)
		}

		if sliceContains(name, filters) || len(filters) == 0 {
			plugin := creator()

			config, err := config.ApplyPlugin(name, plugin)
			if err != nil {
				return nil, err
			}

			a.plugins = append(a.plugins, &runningPlugin{name, plugin, config})
			names = append(names, name)
		}
	}

	sort.Strings(names)

	return names, nil
}

// gatherParallel runs the plugins that are using the same reporting interval
// as the telegraf agent.
func (a *Agent) gatherParallel(pointChan chan *client.Point) error {
	var wg sync.WaitGroup

	start := time.Now()
	counter := 0
	for _, plugin := range a.plugins {
		if plugin.config.Interval != 0 {
			continue
		}

		wg.Add(1)
		counter++
		go func(plugin *runningPlugin) {
			defer wg.Done()

			acc := NewAccumulator(plugin.config, pointChan)
			acc.SetDebug(a.Debug)
			acc.SetPrefix(plugin.name + "_")
			acc.SetDefaultTags(a.Tags)

			if err := plugin.plugin.Gather(acc); err != nil {
				log.Printf("Error in plugin [%s]: %s", plugin.name, err)
			}

		}(plugin)
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("Gathered metrics, (%s interval), from %d plugins in %s\n",
		a.Interval, counter, elapsed)
	return nil
}

// gatherSeparate runs the plugins that have been configured with their own
// reporting interval.
func (a *Agent) gatherSeparate(
	shutdown chan struct{},
	plugin *runningPlugin,
	pointChan chan *client.Point,
) error {
	ticker := time.NewTicker(plugin.config.Interval)

	for {
		var outerr error
		start := time.Now()

		acc := NewAccumulator(plugin.config, pointChan)
		acc.SetDebug(a.Debug)
		acc.SetPrefix(plugin.name + "_")
		acc.SetDefaultTags(a.Tags)

		if err := plugin.plugin.Gather(acc); err != nil {
			log.Printf("Error in plugin [%s]: %s", plugin.name, err)
		}

		elapsed := time.Since(start)
		log.Printf("Gathered metrics, (separate %s interval), from %s in %s\n",
			plugin.config.Interval, plugin.name, elapsed)

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

	for _, plugin := range a.plugins {
		acc := NewAccumulator(plugin.config, pointChan)
		acc.SetDebug(true)
		acc.SetPrefix(plugin.name + "_")

		fmt.Printf("* Plugin: %s, Collection 1\n", plugin.name)
		if plugin.config.Interval != 0 {
			fmt.Printf("* Internal: %s\n", plugin.config.Interval)
		}

		if err := plugin.plugin.Gather(acc); err != nil {
			return err
		}

		// Special instructions for some plugins. cpu, for example, needs to be
		// run twice in order to return cpu usage percentages.
		switch plugin.name {
		case "cpu", "mongodb":
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("* Plugin: %s, Collection 2\n", plugin.name)
			if err := plugin.plugin.Gather(acc); err != nil {
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
	ro *runningOutput,
	shutdown chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if len(points) == 0 {
		return
	}
	retry := 0
	retries := a.FlushRetries
	start := time.Now()

	for {
		err := ro.output.Write(points)
		if err == nil {
			// Write successful
			elapsed := time.Since(start)
			log.Printf("Flushed %d metrics to output %s in %s\n",
				len(points), ro.name, elapsed)
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
				log.Printf(msg, ro.name, retries+1, len(points))
				return
			} else if err != nil {
				// Sleep for a retry
				log.Printf("Error in output [%s]: %s, retrying in %s",
					ro.name, err.Error(), a.FlushInterval.Duration)
				time.Sleep(a.FlushInterval.Duration)
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
	for _, o := range a.outputs {
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

	ticker := time.NewTicker(a.FlushInterval.Duration)
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

	a.FlushInterval.Duration = jitterInterval(a.FlushInterval.Duration,
		a.FlushJitter.Duration)

	log.Printf("Agent Config: Interval:%s, Debug:%#v, Hostname:%#v, "+
		"Flush Interval:%s\n",
		a.Interval, a.Debug, a.Hostname, a.FlushInterval)

	// channel shared between all plugin threads for accumulating points
	pointChan := make(chan *client.Point, 1000)

	// Round collection to nearest interval by sleeping
	if a.RoundInterval {
		i := int64(a.Interval.Duration)
		time.Sleep(time.Duration(i - (time.Now().UnixNano() % i)))
	}
	ticker := time.NewTicker(a.Interval.Duration)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(shutdown, pointChan); err != nil {
			log.Printf("Flusher routine failed, exiting: %s\n", err.Error())
			close(shutdown)
		}
	}()

	for _, plugin := range a.plugins {

		// Start service of any ServicePlugins
		switch p := plugin.plugin.(type) {
		case plugins.ServicePlugin:
			if err := p.Start(); err != nil {
				log.Printf("Service for plugin %s failed to start, exiting\n%s\n",
					plugin.name, err.Error())
				return err
			}
			defer p.Stop()
		}

		// Special handling for plugins that have their own collection interval
		// configured. Default intervals are handled below with gatherParallel
		if plugin.config.Interval != 0 {
			wg.Add(1)
			go func(plugin *runningPlugin) {
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
