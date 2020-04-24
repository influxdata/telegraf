package agent

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal/config"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent_OmitHostname(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true
	_, err := NewAgent(c)
	assert.NoError(t, err)
	assert.NotContains(t, c.Tags, "host")
}

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.InputFilters = []string{"mysql"}
	err := c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "redis"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo", "redis", "bar"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))
}

func TestAgent_LoadOutput(t *testing.T) {
	c := config.NewConfig()
	c.OutputFilters = []string{"influxdb"}
	err := c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"kafka"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "kafka"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(c.Outputs))
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo", "kafka", "bar"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))
}

func TestWindow(t *testing.T) {
	parse := func(s string) time.Time {
		tm, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return tm
	}

	tests := []struct {
		name          string
		start         time.Time
		roundInterval bool
		period        time.Duration
		since         time.Time
		until         time.Time
	}{
		{
			name:          "round with exact alignment",
			start:         parse("2018-03-27T00:00:00Z"),
			roundInterval: true,
			period:        30 * time.Second,
			since:         parse("2018-03-27T00:00:00Z"),
			until:         parse("2018-03-27T00:00:30Z"),
		},
		{
			name:          "round with alignment needed",
			start:         parse("2018-03-27T00:00:05Z"),
			roundInterval: true,
			period:        30 * time.Second,
			since:         parse("2018-03-27T00:00:00Z"),
			until:         parse("2018-03-27T00:00:30Z"),
		},
		{
			name:          "no round with exact alignment",
			start:         parse("2018-03-27T00:00:00Z"),
			roundInterval: false,
			period:        30 * time.Second,
			since:         parse("2018-03-27T00:00:00Z"),
			until:         parse("2018-03-27T00:00:30Z"),
		},
		{
			name:          "no found with alignment needed",
			start:         parse("2018-03-27T00:00:05Z"),
			roundInterval: false,
			period:        30 * time.Second,
			since:         parse("2018-03-27T00:00:05Z"),
			until:         parse("2018-03-27T00:00:35Z"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			since, until := updateWindow(tt.start, tt.roundInterval, tt.period)
			require.Equal(t, tt.since, since, "since")
			require.Equal(t, tt.until, until, "until")
		})
	}
}
