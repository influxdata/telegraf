//go:build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestWinPerformanceQueryImplIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	query := &PerformanceQueryImpl{}

	err := query.Close()
	require.Error(t, err, "uninitialized query must return errors")

	_, err = query.AddCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	require.ErrorContains(t, err, "uninitialized")

	_, err = query.AddEnglishCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	require.ErrorContains(t, err, "uninitialized")

	err = query.CollectData()
	require.Error(t, err, "uninitialized query must return errors")
	require.ErrorContains(t, err, "uninitialized")

	require.NoError(t, query.Open())

	counterPath := "\\Processor Information(_Total)\\% Processor Time"

	hCounter, err := query.AddCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	require.NoError(t, query.Close())

	require.NoError(t, query.Open())

	hCounter, err = query.AddEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	cp, err := query.GetCounterPath(hCounter)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(cp, counterPath))

	require.NoError(t, query.CollectData())
	time.Sleep(time.Second)

	require.NoError(t, query.CollectData())

	fcounter, err := query.GetFormattedCounterValueDouble(hCounter)
	require.NoError(t, err)
	require.True(t, fcounter > 0)

	rcounter, err := query.GetRawCounterValue(hCounter)
	require.NoError(t, err)
	require.True(t, rcounter > 10000000)

	now := time.Now()
	mtime, err := query.CollectDataWithTime()
	require.NoError(t, err)
	require.True(t, mtime.Sub(now) < time.Second)

	counterPath = "\\Process(*)\\% Processor Time"
	paths, err := query.ExpandWildCardPath(counterPath)
	require.NoError(t, err)
	require.NotNil(t, paths)
	require.True(t, len(paths) > 1)

	counterPath = "\\Process(_Total)\\*"
	paths, err = query.ExpandWildCardPath(counterPath)
	require.NoError(t, err)
	require.NotNil(t, paths)
	require.True(t, len(paths) > 1)

	require.NoError(t, query.Open())

	counterPath = "\\Process(*)\\% Processor Time"
	hCounter, err = query.AddEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	require.NoError(t, query.CollectData())
	time.Sleep(time.Second)

	require.NoError(t, query.CollectData())

	farr, err := query.GetFormattedCounterArrayDouble(hCounter)
	var phdErr *PdhError
	if errors.As(err, &phdErr) && phdErr.ErrorCode != PdhInvalidData && phdErr.ErrorCode != PdhCalcNegativeValue {
		time.Sleep(time.Second)
		farr, err = query.GetFormattedCounterArrayDouble(hCounter)
	}
	require.NoError(t, err)
	require.True(t, len(farr) > 0)

	rarr, err := query.GetRawCounterArray(hCounter)
	require.NoError(t, err)
	require.True(t, len(rarr) > 0, "Too")

	require.NoError(t, query.Close())
}

func TestWinPerfCountersConfigGet1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())
}

func TestWinPerfCountersConfigGet2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)

	if len(hostCounters.counters) == 1 {
		require.NoError(t, nil)
	} else if len(hostCounters.counters) == 0 {
		err2 := fmt.Errorf("no results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	} else if len(hostCounters.counters) > 1 {
		err2 := fmt.Errorf("too many results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	}
}

func TestWinPerfCountersConfigGet3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sources := []string{"localhost"}
	instances := []string{"_Total"}
	counters := []string{"% Processor Time", "% Idle Time"}
	perfObjects := []perfObject{{
		Sources:       sources,
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)

	if len(hostCounters.counters) == 2 {
		require.NoError(t, nil)
	} else if len(hostCounters.counters) < 2 {
		err2 := fmt.Errorf("too few results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	} else if len(hostCounters.counters) > 2 {
		err2 := fmt.Errorf("too many results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	}
}

func TestWinPerfCountersConfigGet4Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total", "0,1"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)

	if len(hostCounters.counters) == 2 {
		require.NoError(t, nil)
	} else if len(hostCounters.counters) < 2 {
		err2 := fmt.Errorf("too few results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	} else if len(hostCounters.counters) > 2 {
		err2 := fmt.Errorf("too many results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	}
}

func TestWinPerfCountersConfigGet5Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total", "0,1"}
	counters := []string{"% Processor Time", "% Idle Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)

	if len(hostCounters.counters) == 4 {
		require.NoError(t, nil)
	} else if len(hostCounters.counters) < 4 {
		err2 := fmt.Errorf("too few results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	} else if len(hostCounters.counters) > 4 {
		err2 := fmt.Errorf("too many results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	}
}

func TestWinPerfCountersConfigGet6Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"------"}
	counters := []string{"Context Switches/sec"}
	perfObjects := []perfObject{{
		ObjectName:    "System",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())

	_, ok := m.hostCounters["localhost"]
	require.True(t, ok)
}

func TestWinPerfCountersConfigGet7Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time", "% Processor TimeERROR", "% Idle Time"}
	perfObjects := []perfObject{{
		ObjectName:  "Processor Information",
		Counters:    counters,
		Instances:   instances,
		Measurement: "test",
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())

	hostCounters, ok := m.hostCounters["localhost"]
	require.True(t, ok)

	if len(hostCounters.counters) == 2 {
		require.NoError(t, nil)
	} else if len(hostCounters.counters) < 2 {
		err2 := fmt.Errorf("too few results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	} else if len(hostCounters.counters) > 2 {
		err2 := fmt.Errorf("too many results returned from the counterPath: %v", len(hostCounters.counters))
		require.NoError(t, err2)
	}
}

func TestWinPerfCountersConfigError1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor InformationERROR",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.Error(t, m.ParseConfig())
}

func TestWinPerfCountersConfigError2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"SuperERROR"}
	counters := []string{"% C1 Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.NoError(t, m.ParseConfig())
	var acc testutil.Accumulator
	require.Error(t, m.Gather(&acc))
}

func TestWinPerfCountersConfigError3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"% Processor TimeERROR"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	require.Error(t, m.ParseConfig())
}

func TestWinPerfCountersCollect1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total"}
	counters := []string{"Parking Status"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfObjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc))
	require.Len(t, acc.Metrics, 2)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields["Parking_Status"]
		require.True(t, ok)
	}
}

func TestWinPerfCountersCollect2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"_Total", "0,0"}
	counters := []string{"Performance Limit Flags"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor Information",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}}

	m := WinPerfCounters{
		PrintValid:            false,
		UsePerfCounterTime:    true,
		Object:                perfObjects,
		queryCreator:          &PerformanceQueryCreatorImpl{},
		UseWildcardsExpansion: true,
		Log:                   testutil.Logger{},
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc))

	require.Len(t, acc.Metrics, 4)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields["Performance_Limit_Flags"]
		require.True(t, ok)
	}
}

func TestWinPerfCountersCollectRawIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	instances := []string{"*"}
	counters := []string{"% Idle Time"}
	perfObjects := []perfObject{{
		ObjectName:    "Processor",
		Instances:     instances,
		Counters:      counters,
		Measurement:   "test",
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
		UseRawValues:  true,
	}}

	m := WinPerfCounters{
		PrintValid:            false,
		Object:                perfObjects,
		queryCreator:          &PerformanceQueryCreatorImpl{},
		UseWildcardsExpansion: true,
		Log:                   testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc))
	require.True(t, len(acc.Metrics) > 1)

	expectedCounter := "Percent_Idle_Time_Raw"
	for _, metric := range acc.Metrics {
		val, ok := metric.Fields[expectedCounter]
		require.True(t, ok, "Expected presence of %s field", expectedCounter)
		valInt64, ok := val.(int64)
		require.True(t, ok, fmt.Sprintf("Expected int64, got %T", val))
		require.True(t, valInt64 > 0, fmt.Sprintf("Expected > 0, got %d, for %#v", valInt64, metric))
	}

	// Test *Array way
	m = WinPerfCounters{
		PrintValid:            false,
		Object:                perfObjects,
		queryCreator:          &PerformanceQueryCreatorImpl{},
		UseWildcardsExpansion: false,
		Log:                   testutil.Logger{},
	}
	var acc2 testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	time.Sleep(2000 * time.Millisecond)
	require.NoError(t, m.Gather(&acc2))
	require.True(t, len(acc2.Metrics) > 1)

	for _, metric := range acc2.Metrics {
		val, ok := metric.Fields[expectedCounter]
		require.True(t, ok, "Expected presence of %s field", expectedCounter)
		valInt64, ok := val.(int64)
		require.True(t, ok, fmt.Sprintf("Expected int64, got %T", val))
		require.True(t, valInt64 > 0, fmt.Sprintf("Expected > 0, got %d, for %#v", valInt64, metric))
	}
}
