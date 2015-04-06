package tivan

import (
	"log"
	"net/url"
	"sort"

	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/tivan/plugins"
	"github.com/vektra/cypress"
)
import "time"

type Metrics interface {
	Receive(*cypress.Message) error
}

type Agent struct {
	Interval Duration
	Debug    bool
	HTTP     string

	Config *Config

	plugins []plugins.Plugin

	conn *client.Client

	eachInternal []func()
}

func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{Config: config}

	err := config.Apply("agent", agent)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	c, err := client.NewClient(client.Config{
		URL:       *u,
		Username:  config.Username,
		Password:  config.Password,
		UserAgent: config.UserAgent,
	})

	if err != nil {
		return nil, err
	}

	agent.conn = c

	return agent, nil
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

func (a *Agent) Run(shutdown chan struct{}) {
	ticker := time.NewTicker(a.Interval.Duration)

	for {
		err := a.crank()
		if err != nil {
			log.Printf("Error in plugins: %s", err)
		}

		for _, f := range a.eachInternal {
			f()
		}

		select {
		case <-shutdown:
			return
		case <-ticker.C:
			continue
		}
	}
}
