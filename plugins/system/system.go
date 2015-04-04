package system

import (
	"github.com/influxdb/tivan/plugins"
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/vektra/cypress"
)

type PS interface {
	LoadAvg() (*load.LoadAvgStat, error)
}

type SystemStats struct {
	ps   PS
	tags map[string]string
}

func (s *SystemStats) Read() ([]*cypress.Message, error) {
	lv, err := s.ps.LoadAvg()
	if err != nil {
		return nil, err
	}

	m1 := cypress.Metric()
	m1.Add("type", "gauge")
	m1.Add("name", "load1")
	m1.Add("value", lv.Load1)

	for k, v := range s.tags {
		m1.AddTag(k, v)
	}

	m2 := cypress.Metric()
	m2.Add("type", "gauge")
	m2.Add("name", "load5")
	m2.Add("value", lv.Load5)

	for k, v := range s.tags {
		m2.AddTag(k, v)
	}

	m3 := cypress.Metric()
	m3.Add("type", "gauge")
	m3.Add("name", "load15")
	m3.Add("value", lv.Load15)

	for k, v := range s.tags {
		m3.AddTag(k, v)
	}

	return []*cypress.Message{m1, m2, m3}, nil
}

type systemPS struct{}

func (s *systemPS) LoadAvg() (*load.LoadAvgStat, error) {
	return load.LoadAvg()
}

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{ps: &systemPS{}}
	})
}
