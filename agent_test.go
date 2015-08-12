package telegraf

import (
	"github.com/stretchr/testify/assert"
	"testing"

	// needing to load the plugins
	_ "github.com/influxdb/telegraf/plugins/all"
)

func TestAgent_LoadPlugin(t *testing.T) {

	// load a dedicated configuration file
	config, _ := LoadConfig("./testdata/telegraf-agent.toml")
	a, _ := NewAgent(config)

	pluginsEnabled, _ := a.LoadPlugins("mysql")
	assert.Equal(t, 1, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins("foo")
	assert.Equal(t, 0, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins("mysql:foo")
	assert.Equal(t, 1, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins("mysql:redis")
	assert.Equal(t, 2, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins(":mysql:foo:redis:bar")
	assert.Equal(t, 2, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins("")
	assert.Equal(t, 24, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins(" ")
	assert.Equal(t, 24, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins("		")
	assert.Equal(t, 24, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins("\n\t")
	assert.Equal(t, 24, len(pluginsEnabled))
}

/*
func TestAgent_DrivesMetrics(t *testing.T) {
	var (
		plugin  plugins.MockPlugin
	)

	defer plugin.AssertExpectations(t)
	defer metrics.AssertExpectations(t)

	a := &Agent{
		plugins: []plugins.Plugin{&plugin},
		Config:  &Config{},
	}

	plugin.On("Add", "foo", 1.2, nil).Return(nil)
	plugin.On("Add", "bar", 888, nil).Return(nil)

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
*/
