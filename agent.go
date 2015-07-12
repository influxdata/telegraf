package telegraf

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/telegraf/plugins"
)

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

	plugins []*runningPlugin

	conn *client.Client
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

func (agent *Agent) Connect() error {
	config := agent.Config

	u, err := url.Parse(config.URL)
	if err != nil {
		return err
	}

	c, err := client.NewClient(client.Config{
		URL:       *u,
		Username:  config.Username,
		Password:  config.Password,
		UserAgent: config.UserAgent,
		Timeout:   config.Timeout.Duration,
	})

	if err != nil {
		return err
	}

	agent.conn = c

	return nil
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

	var acc BatchPoints
	acc.Tags = a.Config.Tags
	acc.Time = time.Now()
	acc.Database = a.Config.Database

	for sub := range points {
		acc.Points = append(acc.Points, sub.Points...)
	}

	_, err := a.conn.Write(acc.BatchPoints)
	return err
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
	acc.Database = a.Config.Database

	_, err := a.conn.Write(acc.BatchPoints)
	return err
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
		acc.Database = a.Config.Database

		a.conn.Write(acc.BatchPoints)

		select {
		case <-shutdown:
			return nil
		case <-ticker.C:
			continue
		}
	}
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
	if a.conn == nil {
		err := a.Connect()
		if err != nil {
			return err
		}
	}

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
