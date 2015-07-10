package telegraf

import (
	"fmt"
	"log"
	"os"
	"sort"
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

type Agent struct {
	Interval Duration
	Debug    bool
	Hostname string

	Config *Config

	outputs []*runningOutput
	plugins []*runningPlugin
}

func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{Config: config, Interval: Duration{10 * time.Second}}

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

func (a *Agent) Connect() error {
	for _, o := range a.outputs {
		err := o.output.Connect()
		if err != nil {
			return err
		}
	}
	return nil
}

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

func (a *Agent) LoadPlugins() ([]string, error) {
	var names []string

	for _, name := range a.Config.PluginsDeclared() {
		creator, ok := plugins.Plugins[name]
		if !ok {
			return nil, fmt.Errorf("Undefined but requested plugin: %s", name)
		}

		plugin := creator()

		config, err := a.Config.ApplyPlugin(name, plugin)
		if err != nil {
			return nil, err
		}

		a.plugins = append(a.plugins, &runningPlugin{name, plugin, config})
		names = append(names, name)
	}

	sort.Strings(names)

	return names, nil
}

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

			plugin.plugin.Gather(&acc)

			points <- &acc
		}(plugin)
	}

	wg.Wait()

	close(points)

	var bp BatchPoints
	bp.Tags = a.Config.Tags
	bp.Time = time.Now()

	for sub := range points {
		bp.Points = append(bp.Points, sub.Points...)
	}

	return a.flush(bp)
}

func (a *Agent) crank() error {
	var acc BatchPoints

	acc.Debug = a.Debug

	for _, plugin := range a.plugins {
		acc.Prefix = plugin.name + "_"
		acc.Config = plugin.config
		err := plugin.plugin.Gather(&acc)
		if err != nil {
			return err
		}
	}

	acc.Tags = a.Config.Tags
	acc.Time = time.Now()

	return a.flush(acc)
}

func (a *Agent) crankSeparate(shutdown chan struct{}, plugin *runningPlugin) error {
	ticker := time.NewTicker(plugin.config.Interval)

	for {
		var acc BatchPoints

		acc.Debug = a.Debug

		acc.Prefix = plugin.name + "_"
		acc.Config = plugin.config
		err := plugin.plugin.Gather(&acc)
		if err != nil {
			return err
		}

		acc.Tags = a.Config.Tags
		acc.Time = time.Now()

		err = a.flush(acc)
		if err != nil {
			return err
		}

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (a *Agent) flush(bp BatchPoints) error {
	var wg sync.WaitGroup
	var outerr error
	for _, o := range a.outputs {
		wg.Add(1)
		go func(output *runningOutput) {
			defer wg.Done()
			outerr = o.output.Write(bp.BatchPoints)
		}(o)
	}

	wg.Wait()

	return outerr
}

func (a *Agent) TestAllPlugins() error {
	var names []string

	for name, _ := range plugins.Plugins {
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
		err := plugin.Gather(&acc)
		if err != nil {
			return err
		}
	}

	return nil
}

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

		err := plugin.plugin.Gather(&acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup

	for _, plugin := range a.plugins {
		if plugin.config.Interval != 0 {
			wg.Add(1)
			go func(plugin *runningPlugin) {
				defer wg.Done()
				a.crankSeparate(shutdown, plugin)
			}(plugin)
		}
	}

	defer wg.Wait()

	ticker := time.NewTicker(a.Interval.Duration)

	for {
		err := a.crankParallel()
		if err != nil {
			log.Printf("Error in plugins: %s", err)
		}

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
}
