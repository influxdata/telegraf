package telegraf

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdb/telegraf/outputs"
	"github.com/influxdb/telegraf/plugins"
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
	Interval Duration

	// Option for outputting data in UTC
	UTC bool `toml:"utc"`

	// Precision to write data at
	// Valid values for Precision are n, u, ms, s, m, and h
	Precision string

	// Option for running in debug mode
	Debug    bool
	Hostname string

	Config *Config

	outputs []*runningOutput
	plugins []*runningPlugin
}

// NewAgent returns an Agent struct based off the given Config
func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{
		Config:    config,
		Interval:  Duration{10 * time.Second},
		UTC:       true,
		Precision: "s",
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

	if config.Tags == nil {
		config.Tags = map[string]string{}
	}

	config.Tags["host"] = agent.Hostname

	return agent, nil
}

// Connect connects to all configured outputs
func (a *Agent) Connect() error {
	for _, o := range a.outputs {
		err := o.output.Connect()
		if err != nil {
			return err
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
	}
	return err
}

// LoadOutputs loads the agent's outputs
func (a *Agent) LoadOutputs() ([]string, error) {
	var names []string

	for _, name := range a.Config.OutputsDeclared() {
		creator, ok := outputs.Outputs[name]
		if !ok {
			return nil, fmt.Errorf("Undefined but requested output: %s", name)
		}

		output := creator()

		err := a.Config.ApplyOutput(name, output)
		if err != nil {
			return nil, err
		}

		a.outputs = append(a.outputs, &runningOutput{name, output})
		names = append(names, name)
	}

	sort.Strings(names)

	return names, nil
}

// LoadPlugins loads the agent's plugins
func (a *Agent) LoadPlugins(pluginsFilter string) ([]string, error) {
	var names []string
	var filters []string

	pluginsFilter = strings.TrimSpace(pluginsFilter)
	if pluginsFilter != "" {
		filters = strings.Split(":"+pluginsFilter+":", ":")
	}

	for _, name := range a.Config.PluginsDeclared() {
		creator, ok := plugins.Plugins[name]
		if !ok {
			return nil, fmt.Errorf("Undefined but requested plugin: %s", name)
		}

		isPluginEnabled := false
		if len(filters) > 0 {
			for _, runeValue := range filters {
				if runeValue != "" && strings.ToLower(runeValue) == strings.ToLower(name) {
					fmt.Printf("plugin [%s] is enabled (filter options)\n", name)
					isPluginEnabled = true
					break
				}
			}
		} else {
			// if no filter, we ALWAYS accept the plugin
			isPluginEnabled = true
		}

		if isPluginEnabled {
			plugin := creator()
			config, err := a.Config.ApplyPlugin(name, plugin)
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

// crankParallel runs the plugins that are using the same reporting interval
// as the telegraf agent.
func (a *Agent) crankParallel() error {
	points := make(chan *BatchPoints, len(a.plugins))

	var wg sync.WaitGroup

	for _, plugin := range a.plugins {
		if plugin.config.Interval != 0 {
			continue
		}

		wg.Add(1)
		go func(plugin *runningPlugin) {
			defer wg.Done()

			var acc BatchPoints
			acc.Debug = a.Debug
			acc.Prefix = plugin.name + "_"
			acc.Config = plugin.config

			if err := plugin.plugin.Gather(&acc); err != nil {
				log.Printf("Error in plugin [%s]: %s", plugin.name, err)
			}

			points <- &acc
		}(plugin)
	}

	wg.Wait()

	close(points)

	var bp BatchPoints
	bp.Time = time.Now()
	if a.UTC {
		bp.Time = bp.Time.UTC()
	}
	bp.Tags = a.Config.Tags
	bp.Precision = a.Precision

	for sub := range points {
		bp.Points = append(bp.Points, sub.Points...)
	}

	return a.flush(&bp)
}

// crank is mostly for test purposes.
func (a *Agent) crank() error {
	var bp BatchPoints

	bp.Debug = a.Debug

	for _, plugin := range a.plugins {
		bp.Prefix = plugin.name + "_"
		bp.Config = plugin.config
		err := plugin.plugin.Gather(&bp)
		if err != nil {
			return err
		}
	}

	bp.Tags = a.Config.Tags
	bp.Time = time.Now()
	if a.UTC {
		bp.Time = bp.Time.UTC()
	}
	bp.Precision = a.Precision

	return a.flush(&bp)
}

// crankSeparate runs the plugins that have been configured with their own
// reporting interval.
func (a *Agent) crankSeparate(shutdown chan struct{}, plugin *runningPlugin) error {
	ticker := time.NewTicker(plugin.config.Interval)

	for {
		var bp BatchPoints
		var outerr error

		bp.Debug = a.Debug

		bp.Prefix = plugin.name + "_"
		bp.Config = plugin.config

		if err := plugin.plugin.Gather(&bp); err != nil {
			log.Printf("Error in plugin [%s]: %s", plugin.name, err)
			outerr = errors.New("Error encountered processing plugins & outputs")
		}

		bp.Tags = a.Config.Tags
		bp.Time = time.Now()
		if a.UTC {
			bp.Time = bp.Time.UTC()
		}
		bp.Precision = a.Precision

		if err := a.flush(&bp); err != nil {
			outerr = errors.New("Error encountered processing plugins & outputs")
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

func (a *Agent) flush(bp *BatchPoints) error {
	var wg sync.WaitGroup
	var outerr error

	for _, o := range a.outputs {
		wg.Add(1)
		go func(ro *runningOutput) {
			defer wg.Done()
			// Log all output errors:
			if err := ro.output.Write(bp.BatchPoints); err != nil {
				log.Printf("Error in output [%s]: %s", ro.name, err)
				outerr = errors.New("Error encountered flushing outputs")
			}
		}(o)
	}

	wg.Wait()
	return outerr
}

// TestAllPlugins verifies that we can 'Gather' from all plugins with the
// default configuration
func (a *Agent) TestAllPlugins() error {
	var names []string

	for name := range plugins.Plugins {
		names = append(names, name)
	}

	sort.Strings(names)

	var acc BatchPoints
	acc.Debug = true

	fmt.Printf("* Testing all plugins with default configuration\n")

	for _, name := range names {
		plugin := plugins.Plugins[name]()

		fmt.Printf("* Plugin: %s\n", name)

		acc.Prefix = name + "_"
		if err := plugin.Gather(&acc); err != nil {
			return err
		}
	}

	return nil
}

// Test verifies that we can 'Gather' from all plugins with their configured
// Config struct
func (a *Agent) Test() error {
	var acc BatchPoints

	acc.Debug = true

	for _, plugin := range a.plugins {
		acc.Prefix = plugin.name + "_"
		acc.Config = plugin.config

		fmt.Printf("* Plugin: %s\n", plugin.name)
		if plugin.config.Interval != 0 {
			fmt.Printf("* Internal: %s\n", plugin.config.Interval)
		}

		if err := plugin.plugin.Gather(&acc); err != nil {
			return err
		}
	}

	return nil
}

// Run runs the agent daemon, gathering every Interval
func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup

	for _, plugin := range a.plugins {
		if plugin.config.Interval != 0 {
			wg.Add(1)
			go func(plugin *runningPlugin) {
				defer wg.Done()
				if err := a.crankSeparate(shutdown, plugin); err != nil {
					log.Printf(err.Error())
				}
			}(plugin)
		}
	}

	defer wg.Wait()

	ticker := time.NewTicker(a.Interval.Duration)

	for {
		if err := a.crankParallel(); err != nil {
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
