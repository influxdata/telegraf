package agent

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func panicOnReceive(signals chan os.Signal, tick <-chan time.Time) {
	for {
		select {
		case signals <- syscall.SIGINT:
			// if we reach this, it means that something was listening on the chan, and should not have been
			panic("listener found where none should exist")
		case <-tick:
			// we're done
			return
		}
	}
}

func TestExtractNumeric(t *testing.T) {
	d, ok, err := extractNumeric("v1")
	require.Nil(t, err)
	require.Equal(t, true, ok)
	require.Equal(t, int64(1), d)

	d, ok, err = extractNumeric("1Garbage2")
	require.Nil(t, err)
	require.Equal(t, true, ok)
	require.Equal(t, int64(1), d)

	d, ok, err = extractNumeric("mon~2keyGa-rbage.")
	require.Nil(t, err)
	require.Equal(t, false, ok)
	require.Equal(t, int64(0), d)

	d, ok, err = extractNumeric("13~123bvdbc1")
	require.Nil(t, err)
	require.Equal(t, true, ok)
	require.Equal(t, int64(13), d)
}

func TestGetNumericVersion(t *testing.T) {
	numerics, err := getNumericVersion("v1.2.3")
	require.Nil(t, err)
	require.Equal(t, []int64{1, 2, 3}, numerics)

	numerics, err = getNumericVersion("1.2.3~123bvdbc1")
	require.Nil(t, err)
	require.Equal(t, []int64{1, 2, 3}, numerics)

	numerics, err = getNumericVersion("1.1.9-alpha")
	require.Nil(t, err)
	require.Equal(t, []int64{1, 1, 9}, numerics)

	numerics, err = getNumericVersion("1.2-banana")
	require.NotNil(t, err)
	require.Equal(t, "version string is not of expected semantic format: 1.2-banana", err.Error())
	require.Equal(t, []int64{}, numerics)

	numerics, err = getNumericVersion("1.2")
	require.NotNil(t, err)
	require.Equal(t, "version string is not of expected semantic format: 1.2", err.Error())
	require.Equal(t, []int64{}, numerics)

	numerics, err = getNumericVersion("1.banana.2")
	require.NotNil(t, err)
	require.Equal(t, "version string contains nonsensical character sequence: 1.banana.2", err.Error())
	require.Equal(t, []int64{}, numerics)
}

func TestAgentMetaDataCreate(t *testing.T) {
	c := config.NewConfig()
	outgoing := make(chan telegraf.Metric, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)

	_, err := NewAgentMonitor(ctx, c, signals, outgoing)
	// no version string exists yet, so object instantiation should have failed
	require.NotNil(t, err)

	// so let's add the version, and retry
	c.Agent.Version = "1.20.30"
	_, err = NewAgentMonitor(ctx, c, signals, outgoing)
	require.Nil(t, err)
	cancel()
}

func TestAgentMetaDataRunNoJitter(t *testing.T) {
	c := config.NewConfig()
	outgoing := make(chan telegraf.Metric, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	c.Agent.Version = "1.20.30"
	agentMonitor, err := NewAgentMonitor(ctx, c, signals, outgoing)
	// no version string exists yet, so object instantiation should work
	require.Nil(t, err)

	// start the monitor
	go agentMonitor.Run()

	// grab the "startup signal" - sent on startup
	signal := <-outgoing
	require.Contains(t, "agent_statechange", signal.Name())
	var expected = map[string]interface{}{
		"state": "started",
	}
	actual := signal.Fields()
	require.Equal(t, expected, actual)

	// grab the version it just sent on startup
	metric := <-outgoing
	require.Contains(t, "agent_meta_data", metric.Name())

	// send it an interupt
	signals <- syscall.SIGINT

	// grab the signal
	metric = <-outgoing
	require.Contains(t, "agent_statechange", metric.Name())

	// cancel the agent, which should cause the agentmonitor to quietly die
	cancel()

	// sleep to give cancel() time to work
	time.Sleep(500 * time.Millisecond)

	// verify that it is indeed dead by trying to send another signal to it - which should panic
	ticker := time.NewTicker(100 * time.Millisecond)
	assert.NotPanics(t, func() { panicOnReceive(signals, ticker.C) })
	ticker.Stop()

}
func TestAgentMetaDataVersion(t *testing.T) {
	c := config.NewConfig()
	outgoing := make(chan telegraf.Metric, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	// invalid string
	c.Agent.Version = "1.banana.orangejuice"
	agentMonitor, err := NewAgentMonitor(ctx, c, signals, outgoing)
	require.NotNil(t, err)

	// now with a valid version
	c.Agent.Version = "1.20.30"
	agentMonitor, err = NewAgentMonitor(ctx, c, signals, outgoing)
	require.Nil(t, err)

	// change the jitter value to be crazy short - allowing consistent tests as must be shorter than ticker
	agentMonitor.jitter = time.Duration(10 * time.Millisecond)

	// start the monitor
	go agentMonitor.Run()

	// grab the startup signal - sent on startup
	signal := <-outgoing
	require.Contains(t, "agent_statechange", signal.Name())
	var expected = map[string]interface{}{
		"state": "started",
	}
	actual := signal.Fields()
	require.Equal(t, expected, actual)

	// grab the version it just sent on startup
	metric := <-outgoing
	require.Contains(t, "agent_meta_data", metric.Name())
	expected = map[string]interface{}{
		"version_string": "1.20.30",
		"major_version":  int64(1),
		"minor_version":  int64(20),
		"patch_version":  int64(30),
		"number_inputs":  int64(0),
		"number_outputs": int64(0),
	}
	actual = metric.Fields()
	require.Equal(t, expected, actual)

	// cancel the context, which should cause the agentmonitor to quietly die
	cancel()

	// sleep to give cancel() time to work
	time.Sleep(500 * time.Millisecond)

	// verify that it is indeed dead by trying to send another signal to it - which should panic
	ticker := time.NewTicker(100 * time.Millisecond)
	assert.NotPanics(t, func() { panicOnReceive(signals, ticker.C) })
	ticker.Stop()

}

func TestAgentMetaDataEmptyVersion(t *testing.T) {
	c := config.NewConfig()
	outgoing := make(chan telegraf.Metric, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	// empty string
	c.Agent.Version = ""
	// confirm that default of not to ignore empty version string works as expected
	agentMonitor, err := NewAgentMonitor(ctx, c, signals, outgoing)
	require.NotNil(t, err)

	// now with the setting changed to ignore
	c.Agent.IgnoreInvalidVersion = true
	agentMonitor, err = NewAgentMonitor(ctx, c, signals, outgoing)
	require.Nil(t, err)

	// change the jitter value to be crazy short - allowing consistent tests as must be shorter than ticker
	agentMonitor.jitter = time.Duration(10 * time.Millisecond)

	// start the monitor
	go agentMonitor.Run()

	// grab the startup signal - sent on startup
	signal := <-outgoing
	require.Contains(t, "agent_statechange", signal.Name())
	var expected = map[string]interface{}{
		"state": "started",
	}
	actual := signal.Fields()
	require.Equal(t, expected, actual)

	// grab the version it just sent on startup
	metric := <-outgoing
	require.Contains(t, "agent_meta_data", metric.Name())
	expected = map[string]interface{}{
		"version_string": "none",
		"number_inputs":  int64(0),
		"number_outputs": int64(0),
	}
	actual = metric.Fields()
	require.Equal(t, expected, actual)

	// cancel the context, which should cause the agentmonitor to quietly die
	cancel()

	// sleep to give cancel() time to work
	time.Sleep(500 * time.Millisecond)

	// verify that it is indeed dead by trying to send another signal to it - which should panic
	ticker := time.NewTicker(100 * time.Millisecond)
	assert.NotPanics(t, func() { panicOnReceive(signals, ticker.C) })
	ticker.Stop()

}

func TestAgentMetaData(t *testing.T) {
	c := config.NewConfig()
	outgoing := make(chan telegraf.Metric, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	c.Agent.Version = "1.20.30"

	// make our fake inputs here
	fi := models.NewRunningInput(&fakeInput{name: "sheeps"}, &models.InputConfig{
		Name: "VeryFakeRunningInput",
	})

	c.Inputs = append(c.Inputs, fi)

	agentMonitor, err := NewAgentMonitor(ctx, c, signals, outgoing)
	// no version string exists yet, so object instantiation should work
	require.Nil(t, err)

	// change the jitter value to be crazy short - allowing consistent tests as must be shorter than ticker
	agentMonitor.jitter = time.Duration(10 * time.Millisecond)

	// start the monitor
	go agentMonitor.Run()

	// grab the startup signal - sent on startup
	signal := <-outgoing
	require.Contains(t, "agent_statechange", signal.Name())
	var expected = map[string]interface{}{
		"state": "started",
	}
	actual := signal.Fields()
	require.Equal(t, expected, actual)

	// now grab meta data
	metric := <-outgoing
	require.Contains(t, "agent_meta_data", metric.Name())
	expected = map[string]interface{}{
		"version_string": "1.20.30",
		"major_version":  int64(1),
		"minor_version":  int64(20),
		"patch_version":  int64(30),
		"number_inputs":  int64(1),
		"number_outputs": int64(0),
	}
	actual = metric.Fields()
	require.Equal(t, expected, actual)

	// cancel the context, which should cause the agentmonitor to quietly die
	cancel()

	// sleep to give cancel() time to work
	time.Sleep(500 * time.Millisecond)

	// verify that it is indeed dead by trying to send another signal to it - which should panic
	ticker := time.NewTicker(100 * time.Millisecond)
	assert.NotPanics(t, func() { panicOnReceive(signals, ticker.C) })
	ticker.Stop()

}

type fakeInput struct {
	name string
}

func (f *fakeInput) SampleConfig() string                  { return f.name }
func (f *fakeInput) Description() string                   { return "description for: " + f.name }
func (f *fakeInput) Gather(acc telegraf.Accumulator) error { return nil }
