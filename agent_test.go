package telegraf

import (
	"github.com/stretchr/testify/assert"
	"testing"

	// needing to load the plugins
	_ "github.com/influxdb/telegraf/plugins/all"
	// needing to load the outputs
	_ "github.com/influxdb/telegraf/outputs/all"
)

func TestAgent_LoadPlugin(t *testing.T) {

	// load a dedicated configuration file
	config, _ := LoadConfig("./testdata/telegraf-agent.toml")
	a, _ := NewAgent(config)

	pluginsEnabled, _ := a.LoadPlugins([]string{"mysql"}, config)
	assert.Equal(t, 1, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"foo"}, config)
	assert.Equal(t, 0, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"mysql", "foo"}, config)
	assert.Equal(t, 1, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"mysql", "redis"}, config)
	assert.Equal(t, 2, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"mysql", "foo", "redis", "bar"}, config)
	assert.Equal(t, 2, len(pluginsEnabled))
}

func TestAgent_LoadOutput(t *testing.T) {
	// load a dedicated configuration file
	config, _ := LoadConfig("./testdata/telegraf-agent.toml")
	a, _ := NewAgent(config)

	outputsEnabled, _ := a.LoadOutputs([]string{"influxdb"}, config)
	assert.Equal(t, 1, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{}, config)
	assert.Equal(t, 2, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"foo"}, config)
	assert.Equal(t, 0, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"influxdb", "foo"}, config)
	assert.Equal(t, 1, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"influxdb", "kafka"}, config)
	assert.Equal(t, 2, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"influxdb", "foo", "kafka", "bar"}, config)
	assert.Equal(t, 2, len(outputsEnabled))
}
