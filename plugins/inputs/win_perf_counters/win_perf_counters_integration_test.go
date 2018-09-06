// +build windows

package win_perf_counters

import (
	"errors"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
)

func TestWinPerformanceQueryImpl(t *testing.T) {
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
	assert.True(t, strings.Contains(err.Error(), "uninitialised"))

	_, err = query.AddEnglishCounterToQuery("")
	require.Error(t, err, "uninitialized query must return errors")
	assert.True(t, strings.Contains(err.Error(), "uninitialised"))

	err = query.CollectData()
	require.Error(t, err, "uninitialized query must return errors")
	assert.True(t, strings.Contains(err.Error(), "uninitialised"))

	err = query.Open()
	require.NoError(t, err)

	counterPath := "\\Processor Information(_Total)\\% Processor Time"

	hCounter, err = query.AddCounterToQuery(counterPath)
	require.NoError(t, err)
	assert.NotEqual(t, 0, hCounter)

	err = query.Close()
	require.NoError(t, err)

	err = query.Open()
	require.NoError(t, err)

	hCounter, err = query.AddEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	assert.NotEqual(t, 0, hCounter)

	cp, err := query.GetCounterPath(hCounter)
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(cp, counterPath))

	err = query.CollectData()
	require.NoError(t, err)
	time.Sleep(time.Second)

	err = query.CollectData()
	require.NoError(t, err)

	_, err = query.GetFormattedCounterValueDouble(hCounter)
	require.NoError(t, err)

	now := time.Now()
	mtime, err := query.CollectDataWithTime()
	require.NoError(t, err)
	assert.True(t, mtime.Sub(now) < time.Second)

	counterPath = "\\Process(*)\\% Processor Time"
	paths, err := query.ExpandWildCardPath(counterPath)
	require.NoError(t, err)
	require.NotNil(t, paths)
	assert.True(t, len(paths) > 1)

	counterPath = "\\Process(_Total)\\*"
	paths, err = query.ExpandWildCardPath(counterPath)
	require.NoError(t, err)
	require.NotNil(t, paths)
	assert.True(t, len(paths) > 1)

	err = query.Open()
	require.NoError(t, err)

	counterPath = "\\Process(*)\\% Processor Time"
	hCounter, err = query.AddEnglishCounterToQuery(counterPath)
	require.NoError(t, err)
	assert.NotEqual(t, 0, hCounter)

	err = query.CollectData()
	require.NoError(t, err)
	time.Sleep(time.Second)

	err = query.CollectData()
	require.NoError(t, err)

	arr, err := query.GetFormattedCounterArrayDouble(hCounter)
	if phderr, ok := err.(*PdhError); ok && phderr.ErrorCode != PDH_INVALID_DATA && phderr.ErrorCode != PDH_CALC_NEGATIVE_VALUE {
		time.Sleep(time.Second)
		arr, err = query.GetFormattedCounterArrayDouble(hCounter)
	}
	require.NoError(t, err)
	assert.True(t, len(arr) > 0, "Too")

	err = query.Close()
	require.NoError(t, err)

}

func TestWinPerfcountersConfigGet1(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet2(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	if len(m.counters) == 1 {
		require.NoError(t, nil)
	} else if len(m.counters) == 0 {
		var errorstring1 = "No results returned from the counterPath: " + string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(m.counters) > 1 {
		var errorstring1 = "Too many results returned from the counterPath: " + string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var instances = make([]string, 1)
	var counters = make([]string, 2)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	if len(m.counters) == 2 {
		require.NoError(t, nil)
	} else if len(m.counters) < 2 {

		var errorstring1 = "Too few results returned from the counterPath. " + string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(m.counters) > 2 {

		var errorstring1 = "Too many results returned from the counterPath: " + string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet4(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	if len(m.counters) == 2 {
		require.NoError(t, nil)
	} else if len(m.counters) < 2 {

		var errorstring1 = "Too few results returned from the counterPath: " + string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(m.counters) > 2 {

		var errorstring1 = "Too many results returned from the counterPath: " + string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet5(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	if len(m.counters) == 4 {
		require.NoError(t, nil)
	} else if len(m.counters) < 4 {
		var errorstring1 = "Too few results returned from the counterPath: " +
			string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(m.counters) > 4 {
		var errorstring1 = "Too many results returned from the counterPath: " +
			string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet6(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet7(t *testing.T) {
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
		objectname,
		counters,
		instances,
		measurement,
		false,
		false,
		false,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	if len(m.counters) == 2 {
		require.NoError(t, nil)
	} else if len(m.counters) < 2 {
		var errorstring1 = "Too few results returned from the counterPath: " +
			string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(m.counters) > 2 {
		var errorstring1 = "Too many results returned from the counterPath: " +
			string(len(m.counters))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigError1(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersConfigError2(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	var acc testutil.Accumulator
	err = m.Gather(&acc)
	require.Error(t, err)
}

func TestWinPerfcountersConfigError3(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersCollect1(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)
	require.NoError(t, err)
	assert.Len(t, acc.Metrics, 2)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields[expectedCounter]
		assert.True(t, ok)
	}

}
func TestWinPerfcountersCollect2(t *testing.T) {
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

	m := Win_PerfCounters{PrintValid: false, UsePerfCounterTime: true, Object: perfobjects, query: &PerformanceQueryImpl{}, UseWildcardsExpansion: true}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)
	require.NoError(t, err)

	assert.Len(t, acc.Metrics, 4)

	for _, metric := range acc.Metrics {
		_, ok := metric.Fields[expectedCounter]
		assert.True(t, ok)
	}

}
