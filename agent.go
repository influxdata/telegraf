package telegraf

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
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

// Agent runs telegraf and collects data based on the given config
type Agent struct {

	// Interval at which to gather information
	Interval Duration

	// Run in debug mode?
	Debug    bool
	Hostname string

	Config *Config

	plugins []*runningPlugin

	conn *client.Client
}

// NewAgent returns an Agent struct based off the given Config
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

// Connect connects to the agent's config URL
func (a *Agent) Connect() error {
	config := a.Config

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

	_, err = c.Query(client.Query{
		Command: fmt.Sprintf("CREATE DATABASE telegraf"),
	})

	if err != nil && !strings.Contains(err.Error(), "database already exists") {
		log.Fatal(err)
	}

	a.conn = c

	return nil
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
		err := plugin.Gather(&acc)
		if err != nil {
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

		err := plugin.plugin.Gather(&acc)
		if err != nil {
			return err
		}
	}

	return nil
}

// Run runs the agent daemon, gathering every Interval
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
