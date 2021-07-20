package agent

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent_OmitHostname(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true
	require.NotContains(t, c.Tags, "host")
}

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.InputFilters = []string{"mysql"}
	a := NewAgent(context.Background(), c)
	c.SetAgent(a)
	err := c.LoadConfig(context.Background(), context.Background(), "../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(a.Config.Inputs()))

	c = config.NewConfig()
	c.InputFilters = []string{"foo"}
	a = NewAgent(context.Background(), c)
	c.SetAgent(a)
	err = c.LoadConfig(context.Background(), context.Background(), "../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	require.Equal(t, 0, len(a.Config.Inputs()))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo"}
	a = NewAgent(context.Background(), c)
	c.SetAgent(a)
	err = c.LoadConfig(context.Background(), context.Background(), "../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(a.Config.Inputs()))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "redis"}
	a = NewAgent(context.Background(), c)
	c.SetAgent(a)
	err = c.LoadConfig(context.Background(), context.Background(), "../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	require.Equal(t, 2, len(a.Config.Inputs()))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo", "redis", "bar"}
	a = NewAgent(context.Background(), c)
	c.SetAgent(a)
	err = c.LoadConfig(context.Background(), context.Background(), "../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	require.Equal(t, 2, len(a.Config.Inputs()))
}

func TestAgent_LoadOutput(t *testing.T) {
	tests := []struct {
		OutputFilters   []string
		ExpectedOutputs int
	}{
		{OutputFilters: []string{"influxdb"}, ExpectedOutputs: 2}, // two instances in toml
		{OutputFilters: []string{"kafka"}, ExpectedOutputs: 1},
		{OutputFilters: []string{}, ExpectedOutputs: 3},
		{OutputFilters: []string{"foo"}, ExpectedOutputs: 0},
		{OutputFilters: []string{"influxdb", "foo"}, ExpectedOutputs: 2},
		{OutputFilters: []string{"influxdb", "kafka"}, ExpectedOutputs: 3},
		{OutputFilters: []string{"influxdb", "foo", "kafka", "bar"}, ExpectedOutputs: 3},
	}
	for _, test := range tests {
		name := strings.Join(test.OutputFilters, "-")
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c := config.NewConfig()
			c.OutputFilters = test.OutputFilters
			a := NewAgent(ctx, c)
			c.SetAgent(a)
			err := c.LoadConfig(context.Background(), context.Background(), "../config/testdata/telegraf-agent.toml")
			require.NoError(t, err)
			require.Len(t, a.Config.Outputs(), test.ExpectedOutputs)
		})
	}
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
