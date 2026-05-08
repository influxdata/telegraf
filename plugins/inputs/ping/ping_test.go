package ping

import (
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestNativeIDsAssignment(t *testing.T) {
	// Generate a target list
	targets := slices.Repeat([]string{"localhost"}, 100)

	// Add a number of plugin instances that need to share the IDs
	plugins := make([]*Ping, 0, 10)
	for range 10 {
		plugin := &Ping{
			Method: "native",
			Urls:   targets,
			Log:    &testutil.Logger{},
		}
		require.NoError(t, plugin.Init())
		require.NoError(t, plugin.Start(nil))
		plugins = append(plugins, plugin)
	}

	// Check for the IDs being missing
	require.Len(t, availableIDs, math.MaxUint16-1000)

	// Check the plugins for different ranges and stop them afterwards to
	// free the IDs. This is crucial when stop/starting the plugin e.g. during
	// reloading configurations
	for i, plugin := range plugins {
		require.Len(t, plugin.availableIDs, 100)
		require.Equal(t, i*100, plugin.availableIDs[0])
		plugin.Stop()
		require.Empty(t, plugin.availableIDs)
	}

	// Check that all reserved IDs were returned
	require.Len(t, availableIDs, math.MaxUint16)
}

func TestNativeIDsOverflow(t *testing.T) {
	// Create plugin with too many URLS
	plugin := &Ping{
		Method: "native",
		Urls:   slices.Repeat([]string{"localhost"}, 70000),

		Log: &testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "too many URLs (70000)")
}

func TestNativeIDsOverflowAcrossPlugins(t *testing.T) {
	// Generate a target list
	targets := slices.Repeat([]string{"localhost"}, 10000)

	// Add plugin instances staying within the number of available IDs
	plugins := make([]*Ping, 0, 10)
	defer func() {
		for _, plugin := range plugins {
			plugin.Stop()
		}
	}()
	for range 6 {
		plugin := &Ping{
			Method: "native",
			Urls:   targets,
			Log:    &testutil.Logger{},
		}
		require.NoError(t, plugin.Init())
		require.NoError(t, plugin.Start(nil))
		plugins = append(plugins, plugin)
	}

	// This instance should cause the overflow
	plugin := &Ping{
		Method: "native",
		Urls:   targets,
		Log:    &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.ErrorContains(t, plugin.Start(nil), "too many URLs across all plugin instances")
}

func Benchmark(b *testing.B) {
	// Generate a target list
	targets := slices.Repeat([]string{"localhost"}, 100)

	plugin := &Ping{
		Method: "native",
		Urls:   targets,
		Log:    &testutil.Logger{},
	}
	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Start(nil))
	defer plugin.Stop()

	acc := &testutil.Accumulator{Discard: true}
	for b.Loop() {
		require.NoError(b, plugin.Gather(acc))
	}
}
