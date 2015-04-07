package tivan

import (
	"log"
	"net/url"
	"sort"
	"time"

	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/tivan/plugins"
)

type Agent struct {
	Interval Duration
	Debug    bool
	HTTP     string

	Config *Config

	plugins []plugins.Plugin

	conn *client.Client
}

func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{Config: config}

	err := config.Apply("agent", agent)
	if err != nil {
		return nil, err
	}

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
	})

	if err != nil {
		return err
	}

	agent.conn = c

	return nil
}

func (a *Agent) LoadPlugins() ([]string, error) {
	var names []string

	for name, creator := range plugins.Plugins {
		a.plugins = append(a.plugins, creator())
		names = append(names, name)
	}

	sort.Strings(names)

	return names, nil
}

func (a *Agent) crank() error {
	var acc BatchPoints

	acc.Debug = a.Debug

	for _, plugin := range a.plugins {
		err := plugin.Gather(&acc)
		if err != nil {
			return err
		}
	}

	acc.Tags = a.Config.Tags
	acc.Timestamp = time.Now()
	acc.Database = a.Config.Database

	_, err := a.conn.Write(acc.BatchPoints)
	return err
}

func (a *Agent) Test() error {
	var acc BatchPoints

	acc.Debug = true

	for _, plugin := range a.plugins {
		err := plugin.Gather(&acc)
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

	ticker := time.NewTicker(a.Interval.Duration)

	for {
		err := a.crank()
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
