package tivan

import (
	"testing"

	"github.com/influxdb/tivan/plugins"
	"github.com/stretchr/testify/require"
	"github.com/vektra/cypress"
)

func TestAgent_DrivesMetrics(t *testing.T) {
	var (
		plugin  plugins.MockPlugin
		metrics MockMetrics
	)

	defer plugin.AssertExpectations(t)
	defer metrics.AssertExpectations(t)

	a := &Agent{
		plugins: []plugins.Plugin{&plugin},
		metrics: &metrics,
		Config:  &Config{},
	}

	m1 := cypress.Metric()
	m1.Add("name", "foo")
	m1.Add("value", 1.2)

	m2 := cypress.Metric()
	m2.Add("name", "bar")
	m2.Add("value", 888)

	msgs := []*cypress.Message{m1, m2}

	plugin.On("Read").Return(msgs, nil)
	metrics.On("Receive", m1).Return(nil)
	metrics.On("Receive", m2).Return(nil)

	err := a.crank()
	require.NoError(t, err)
}

func TestAgent_AppliesTags(t *testing.T) {
	var (
		plugin  plugins.MockPlugin
		metrics MockMetrics
	)

	defer plugin.AssertExpectations(t)
	defer metrics.AssertExpectations(t)

	a := &Agent{
		plugins: []plugins.Plugin{&plugin},
		metrics: &metrics,
		Config: &Config{
			Tags: map[string]string{
				"dc": "us-west-1",
			},
		},
	}

	m1 := cypress.Metric()
	m1.Add("name", "foo")
	m1.Add("value", 1.2)

	msgs := []*cypress.Message{m1}

	m2 := cypress.Metric()
	m2.Timestamp = m1.Timestamp
	m2.Add("name", "foo")
	m2.Add("value", 1.2)
	m2.AddTag("dc", "us-west-1")

	plugin.On("Read").Return(msgs, nil)
	metrics.On("Receive", m2).Return(nil)

	err := a.crank()
	require.NoError(t, err)
}
