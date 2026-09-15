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

func TestAgent_SkipProcessorsAfterAggregatorsDefaultsFalse(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true

	a := NewAgent(c)
	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Second)
	defer cancel()

	require.NoError(t, a.Run(ctx))
	require.NotNil(t, c.Agent.SkipProcessorsAfterAggregators)
	require.False(t, *c.Agent.SkipProcessorsAfterAggregators)
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

// pushRecorder reports the wall-clock time of the first Push.
type pushRecorder struct {
	pushed chan time.Time
}

func (*pushRecorder) SampleConfig() string { return "" }
func (*pushRecorder) Add(telegraf.Metric)  {}
func (*pushRecorder) Reset()               {}
func (r *pushRecorder) Push(telegraf.Accumulator) {
	select {
	case r.pushed <- time.Now():
	default:
	}
}

// pushAfterWindowShift runs the aggregator push loop against a window that
// ends 500ms out, then moves the window end forward by shift while the loop
// is asleep on its timer. From the loop's point of view that is the same as
// the wall clock stepping back by shift while it waited. It returns the time
// of the first Push and the shifted window end.
func pushAfterWindowShift(t *testing.T, shift time.Duration) (pushed, windowEnd time.Time) {
	rec := &pushRecorder{pushed: make(chan time.Time, 1)}
	ra := models.NewRunningAggregator(rec, &models.AggregatorConfig{
		Name:   "test",
		Filter: models.Filter{NamePass: []string{"*"}},
		Period: time.Second,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	// Strip the monotonic reading, as AlignTime does under round_interval.
	end := time.Now().Add(500 * time.Millisecond).Truncate(-1)
	ra.UpdateWindow(end.Add(-ra.Period()), end)

	ctx, cancel := context.WithCancel(context.Background())
	var acc testutil.Accumulator
	done := make(chan struct{})
	go func() {
		defer close(done)
		(&Agent{}).push(ctx, ra, &acc)
	}()

	// Give the loop time to arm its timer before moving the deadline under
	// it; the shift has to land while the timer is pending.
	time.Sleep(50 * time.Millisecond)
	windowEnd = end.Add(shift)
	ra.Lock()
	ra.UpdateWindow(windowEnd.Add(-ra.Period()), windowEnd)
	ra.Unlock()

	select {
	case pushed = <-rec.pushed:
	case <-time.After(5 * time.Second):
		t.Fatal("no push within 5s")
	}
	cancel()
	<-done
	return pushed, windowEnd
}

func TestAggregatorPushWaitsForWindowEndOnEarlyTimer(t *testing.T) {
	pushed, windowEnd := pushAfterWindowShift(t, 500*time.Millisecond)
	require.False(t, pushed.Before(windowEnd), "pushed at %v, before the window end %v", pushed, windowEnd)
}

func TestAggregatorPushDoesNotWaitOutClockAdjustment(t *testing.T) {
	pushed, windowEnd := pushAfterWindowShift(t, 2*time.Second)
	require.True(t, pushed.Before(windowEnd), "pushed at %v, waited for the window end %v", pushed, windowEnd)
}
