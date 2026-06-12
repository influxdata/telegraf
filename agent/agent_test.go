package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
	"github.com/influxdata/telegraf/testutil"
)

func TestAgent_OmitHostname(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true
	_ = NewAgent(c)
	require.NotContains(t, c.Tags, "host")
}

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.InputFilters = []string{"mysql"}
	err := c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a := NewAgent(c)
	require.Len(t, a.Config.Inputs, 1)

	c = config.NewConfig()
	c.InputFilters = []string{"foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Empty(t, a.Config.Inputs)

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Inputs, 1)

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "redis"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Inputs, 2)

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo", "redis", "bar"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Inputs, 2)
}

func TestAgent_LoadOutput(t *testing.T) {
	c := config.NewConfig()
	c.OutputFilters = []string{"influxdb"}
	err := c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a := NewAgent(c)
	require.Len(t, a.Config.Outputs, 2)

	c = config.NewConfig()
	c.OutputFilters = []string{"kafka"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Outputs, 1)

	c = config.NewConfig()
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Outputs, 3)

	c = config.NewConfig()
	c.OutputFilters = []string{"foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Empty(t, a.Config.Outputs)

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Outputs, 2)

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "kafka"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	require.Len(t, c.Outputs, 3)
	a = NewAgent(c)
	require.Len(t, a.Config.Outputs, 3)

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo", "kafka", "bar"}
	err = c.LoadConfig("../config/testdata/telegraf-agent.toml")
	require.NoError(t, err)
	a = NewAgent(c)
	require.Len(t, a.Config.Outputs, 3)
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

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		testdataPath := filepath.Join("testcases", fname)
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
		expectedFilename := filepath.Join(testdataPath, "expected.out")

		t.Run(fname, func(t *testing.T) {
			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)
			require.NotEmpty(t, expected)

			// Load the config and inject the mock output to be able to verify
			// the resulting metrics
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadAll(configFilename))
			require.Empty(t, cfg.Outputs, "No output(s) allowed in the config!")

			// Setup the agent and run the agent in "once" mode
			agent := NewAgent(cfg)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			actual, err := collect(ctx, agent, 0)
			require.NoError(t, err)

			// Process expected metrics and compare with resulting metrics
			options := []cmp.Option{
				testutil.IgnoreTags("host"),
				testutil.IgnoreTime(),
			}
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func ptrBool(b bool) *bool { return &b }

// Validate default initialization of skip flags
func TestAgent_DefaultSkipFlags(t *testing.T) {
    c := config.NewConfig()
    a := NewAgent(c)
    require.NotNil(t, a.Config.Agent.SkipProcessorsBeforeAggregators)
    require.NotNil(t, a.Config.Agent.SkipProcessorsAfterAggregators)
    require.False(t, *a.Config.Agent.SkipProcessorsBeforeAggregators)
    require.False(t, *a.Config.Agent.SkipProcessorsAfterAggregators)
}

// Validate error when both flags are true
func TestAgent_BothSkipFlagsError(t *testing.T) {
    c := config.NewConfig()
    c.Agent.SkipProcessorsBeforeAggregators = ptrBool(true)
    c.Agent.SkipProcessorsAfterAggregators = ptrBool(true)
    a := NewAgent(c)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    err := a.runTest(ctx, 0, make(chan telegraf.Metric))
    require.Error(t, err)
    require.Contains(t, err.Error(), "cannot set both SkipProcessorsBeforeAggregators and SkipProcessorsAfterAggregators to true")
}

// Validate runTest with skip-before=true, skip-after=false
func TestAgent_RunTest_SkipBeforeOnly(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    cfg := config.NewConfig()
    cfg.Agent.SkipProcessorsBeforeAggregators = ptrBool(true)
    cfg.Agent.SkipProcessorsAfterAggregators = ptrBool(false)
    cfg.InputFilters = []string{"cpu"}
    cfg.ProcessorFilters = []string{"rename"}
    cfg.AggregatorFilters = []string{"minmax"}
    require.NoError(t, cfg.LoadAll("../config/testdata/telegraf-agent.toml"))

    agent := NewAgent(cfg)
    metrics, err := collect(ctx, agent, 0)
    require.NoError(t, err)

    found := false
    for _, m := range metrics {
        if _, ok := m.Fields()["usage_user_min"]; ok {
            found = true
            break
        }
    }
    require.True(t, found, "expected usage_user_min when skip before is true")
}

// Validate runOnce with skip-after=true, skip-before=false
func TestAgent_RunOnce_SkipAfterOnly(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    cfg := config.NewConfig()
    cfg.Agent.SkipProcessorsBeforeAggregators = ptrBool(false)
    cfg.Agent.SkipProcessorsAfterAggregators = ptrBool(true)
    cfg.InputFilters = []string{"cpu"}
    cfg.ProcessorFilters = []string{"rename"}
    cfg.AggregatorFilters = []string{"minmax"}
    require.NoError(t, cfg.LoadAll("../config/testdata/telegraf-agent.toml"))

    agent := NewAgent(cfg)
    out := &testutil.Accumulator{}
    agent.Config.Outputs = []telegraf.Output{out}

    err := agent.runOnce(ctx, 0)
    require.NoError(t, err)

    found := false
    for _, m := range out.Metrics {
        if _, ok := m.Fields()["user_cpu_min"]; ok {
            found = true
            break
        }
    }
    require.True(t, found, "expected user_cpu_min when skip after is true")
}

// Validate runOnce with both flags false (everything runs)
func TestAgent_RunOnce_BothFalse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := config.NewConfig()
	cfg.Agent.SkipProcessorsBeforeAggregators = ptrBool(false)
	cfg.Agent.SkipProcessorsAfterAggregators = ptrBool(false)
	cfg.InputFilters = []string{"cpu"}
	cfg.ProcessorFilters = []string{"rename"}
	cfg.AggregatorFilters = []string{"minmax"}
	require.NoError(t, cfg.LoadAll("../config/testdata/telegraf-agent.toml"))

	agent := NewAgent(cfg)
	out := &testutil.Accumulator{}
	agent.Config.Outputs = []telegraf.Output{out}

	err := agent.runOnce(ctx, 0)
	require.NoError(t, err)

	found := false
	for _, m := range out.Metrics {
		if _, ok := m.Fields()["user_cpu_min"]; ok {
			found = true
			break
		}
	}
	require.True(t, found, "expected user_cpu_min when both flags are false")
}

// Implement a "test-mode" like call but collect the metrics
func collect(ctx context.Context, a *Agent, wait time.Duration) ([]telegraf.Metric, error) {
	var received []telegraf.Metric
	var mu sync.Mutex

	src := make(chan telegraf.Metric, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for m := range src {
			mu.Lock()
			received = append(received, m)
			mu.Unlock()
			m.Reject()
		}
	}()

	if err := a.runTest(ctx, wait, src); err != nil {
		return nil, err
	}
	wg.Wait()

	if models.GlobalGatherErrors.Get() != 0 {
		return received, fmt.Errorf("input plugins recorded %d errors", models.GlobalGatherErrors.Get())
	}
	return received, nil
}
