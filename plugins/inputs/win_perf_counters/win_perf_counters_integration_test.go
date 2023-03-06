//go:build windows

package win_perf_counters

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWinPerformanceQueryImplIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	var query PerformanceQuery
	var hCounter PDH_HCOUNTER
	var err error
	query = &PerformanceQueryImpl{}

	err = query.Close()
	require.Error(t, err, "uninitialized query must return errors")

	_, err = query.AddCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	require.True(t, strings.Contains(err.Error(), "uninitialized"))

	_, err = query.AddEnglishCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	require.True(t, strings.Contains(err.Error(), "uninitialized"))

	err = query.CollectData()
	require.Error(t, err, "uninitialized query must return errors")
	require.True(t, strings.Contains(err.Error(), "uninitialized"))

	err = query.Open()
	require.NoError(t, err)

	counterPath := "\\Processor Information(_Total)\\% Processor Time"

	hCounter, err = query.AddCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	err = query.Close()
	require.NoError(t, err)

	err = query.Open()
	require.NoError(t, err)

	hCounter, err = query.AddEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	cp, err := query.GetCounterPath(hCounter)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(cp, counterPath))

	err = query.CollectData()
	require.NoError(t, err)
	time.Sleep(time.Second)

	err = query.CollectData()
	require.NoError(t, err)

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

	err = query.Open()
	require.NoError(t, err)

	counterPath = "\\Process(*)\\% Processor Time"
	hCounter, err = query.AddEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, hCounter)

	err = query.CollectData()
	require.NoError(t, err)
	time.Sleep(time.Second)

	err = query.CollectData()
	require.NoError(t, err)

	farr, err := query.GetFormattedCounterArrayDouble(hCounter)
	if phderr, ok := err.(*PdhError); ok && phderr.ErrorCode != PDH_INVALID_DATA && phderr.ErrorCode != PDH_CALC_NEGATIVE_VALUE {
		time.Sleep(time.Second)
		farr, err = query.GetFormattedCounterArrayDouble(hCounter)
	}
	require.NoError(t, err)
	require.True(t, len(farr) > 0)

	rarr, err := query.GetRawCounterArray(hCounter)
	require.NoError(t, err)
	require.True(t, len(rarr) > 0, "Too")

	err = query.Close()
	require.NoError(t, err)

}

func TestWinPerfcountersConfigGet1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)

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

func TestWinPerfcountersConfigGet3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var sources = make([]string, 1)
	var instances = make([]string, 1)
	var counters = make([]string, 2)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"
	counters[1] = "% Idle Time"
	sources[0] = "localhost"

	var measurement = "test"

	PerfObject := perfobject{
		Sources:       sources,
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)

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

func TestWinPerfcountersConfigGet4Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 2)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0,1"
	counters[0] = "% Processor Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)

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

func TestWinPerfcountersConfigGet5Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 2)
	var counters = make([]string, 2)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0,1"
	counters[0] = "% Processor Time"
	counters[1] = "% Idle Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)

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

func TestWinPerfcountersConfigGet6Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "System"
	instances[0] = "------"
	counters[0] = "Context Switches/sec"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)

	_, ok := m.hostCounters["localhost"]
	require.True(t, ok)
}

func TestWinPerfcountersConfigGet7Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 3)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"
	counters[1] = "% Processor TimeERROR"
	counters[2] = "% Idle Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:  objectname,
		Counters:    counters,
		Instances:   instances,
		Measurement: measurement,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.NoError(t, err)

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

func TestWinPerfcountersConfigError1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor InformationERROR"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersConfigError2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor"
	instances[0] = "SuperERROR"
	counters[0] = "% C1 Time"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	var acc testutil.Accumulator
	err = m.Gather(&acc)
	require.Error(t, err)
}

func TestWinPerfcountersConfigError3Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor TimeERROR"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersCollect1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "Parking Status"

	var expectedCounter = "Parking_Status"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:   false,
		Object:       perfobjects,
		queryCreator: &PerformanceQueryCreatorImpl{},
		Log:          testutil.Logger{},
	}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)
	require.NoError(t, err)
	require.Len(t, acc.Metrics, 2)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields[expectedCounter]
		require.True(t, ok)
	}

}
func TestWinPerfcountersCollect2Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 2)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0,0"
	counters[0] = "Performance Limit Flags"

	var expectedCounter = "Performance_Limit_Flags"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:            false,
		UsePerfCounterTime:    true,
		Object:                perfobjects,
		queryCreator:          &PerformanceQueryCreatorImpl{},
		UseWildcardsExpansion: true,
		Log:                   testutil.Logger{},
	}

	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)
	require.NoError(t, err)

	require.Len(t, acc.Metrics, 4)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields[expectedCounter]
		require.True(t, ok)
	}

}

func TestWinPerfcountersCollectRawIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor"
	instances[0] = "*"
	counters[0] = "% Idle Time"

	var expectedCounter = "Percent_Idle_Time_Raw"

	var measurement = "test"

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: true,
		IncludeTotal:  false,
		UseRawValues:  true,
	}

	perfobjects[0] = PerfObject

	m := WinPerfCounters{
		PrintValid:            false,
		Object:                perfobjects,
		queryCreator:          &PerformanceQueryCreatorImpl{},
		UseWildcardsExpansion: true,
		Log:                   testutil.Logger{},
	}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)
	require.NoError(t, err)
	require.True(t, len(acc.Metrics) > 1)

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
		Object:                perfobjects,
		queryCreator:          &PerformanceQueryCreatorImpl{},
		UseWildcardsExpansion: false,
		Log:                   testutil.Logger{},
	}
	var acc2 testutil.Accumulator
	err = m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc2)
	require.NoError(t, err)
	require.True(t, len(acc2.Metrics) > 1)

	for _, metric := range acc2.Metrics {
		val, ok := metric.Fields[expectedCounter]
		require.True(t, ok, "Expected presence of %s field", expectedCounter)
		valInt64, ok := val.(int64)
		require.True(t, ok, fmt.Sprintf("Expected int64, got %T", val))
		require.True(t, valInt64 > 0, fmt.Sprintf("Expected > 0, got %d, for %#v", valInt64, metric))
	}

}
