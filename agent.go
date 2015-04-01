package tivan

import (
	"fmt"
	"log"
	"sort"

	"github.com/influxdb/tivan/plugins"
	"github.com/vektra/cypress"
	"github.com/vektra/cypress/plugins/metrics"
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
	metrics Metrics

	eachInternal []func()
}

func NewAgent(config *Config) *Agent {
	m := metrics.NewMetricSink()

	agent := &Agent{Config: config, metrics: m}

	err := config.Apply("agent", agent)
	if err != nil {
		panic(err)
	}

	if config.URL != "" {
		icfg := metrics.DefaultInfluxConfig()
		icfg.URL = config.URL
		icfg.Username = config.Username
		icfg.Password = config.Password
		icfg.Database = config.Database
		icfg.UserAgent = config.UserAgent

		agent.eachInternal = append(agent.eachInternal, func() {
			if agent.Debug {
				log.Printf("flushing to influxdb")
			}

			m.FlushInflux(icfg)
		})
	}

	return agent
}

type HTTPInterface interface {
	RunHTTP(string) error
}

func (a *Agent) RunHTTP(addr string) {
	a.metrics.(HTTPInterface).RunHTTP(addr)
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
	for _, plugin := range a.plugins {
		msgs, err := plugin.Read()
		if err != nil {
			return err
		}

		for _, m := range msgs {
			for k, v := range a.Config.Tags {
				m.AddTag(k, v)
			}

			if a.Debug {
				fmt.Println(m.KVString())
			}

			err = a.metrics.Receive(m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Agent) Run(shutdown chan struct{}) {
	if a.HTTP != "" {
		go a.RunHTTP(a.HTTP)
	}

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
