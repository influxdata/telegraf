package agent

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
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

func TestAgent_isIgnoreInput(t *testing.T) {
	c := config.NewConfig()
	assert.False(t, c.Agent.IgnoreErrorInputs)
	a, err := NewAgent(c)
	assert.NoError(t, err)
	input := &models.RunningInput{
		Config: &models.InputConfig{},
	}
	assert.False(t, a.Config.Agent.IgnoreErrorInputs)
	assert.False(t, input.Config.IgnoreInitError)
	// default: input.ignore_init_error=false and agent.ignore_error_inputs=false
	assert.False(t, a.isIgnoreInput(input))

	a.Config.Agent.IgnoreErrorInputs = true
	// input.ignore_init_error=false and agent.ignore_error_inputs=true
	assert.True(t, a.isIgnoreInput(input))

	input.Config.IgnoreInitError = true
	// input.ignore_init_error=true and agent.ignore_error_inputs=true
	assert.True(t, a.isIgnoreInput(input))

	a.Config.Agent.IgnoreErrorInputs = false
	// input.ignore_init_error=false and agent.ignore_error_inputs=false
	assert.True(t, a.isIgnoreInput(input))

}

type testIgnoreErrorInput struct {
}

func (i *testIgnoreErrorInput) Init() error {
	return fmt.Errorf("could not initialize input: test error")
}

func (i *testIgnoreErrorInput) Gather(telegraf.Accumulator) error {
	return nil
}
func (i *testIgnoreErrorInput) SampleConfig() string {
	return ""
}

func TestAgent_IgnoreErrorInputs(t *testing.T) {
	c := config.NewConfig()
	assert.False(t, c.Agent.IgnoreErrorInputs)
	c.Inputs = []*models.RunningInput{{}}
	a, err := NewAgent(c)
	assert.NoError(t, err)
	err = a.initPlugins()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(c.Inputs))

	c.Inputs = []*models.RunningInput{{
		Config: &models.InputConfig{
			Name:     "test error input",
			Alias:    "test alias",
			Interval: 10 * time.Second,
		},
		Input: &testIgnoreErrorInput{},
	}}
	a, err = NewAgent(c)
	assert.NoError(t, err)
	err = a.initPlugins()
	assert.Error(t, err)

	assert.Equal(t, 1, len(c.Inputs))
	c.Agent.IgnoreErrorInputs = true
	err = a.initPlugins()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(c.Inputs))
}

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.InputFilters = []string{"mysql"}
	err := c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "redis"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo", "redis", "bar"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))
}

func TestAgent_LoadOutput(t *testing.T) {
	c := config.NewConfig()
	c.OutputFilters = []string{"influxdb"}
	err := c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"kafka"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "kafka"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(c.Outputs))
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo", "kafka", "bar"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
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
