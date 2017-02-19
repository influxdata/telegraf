// +build windows

package win_perf_counters

import (
	"errors"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWinPerfcountersConfigGet1(t *testing.T) {
	validmetrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet1", Object: perfobjects}

	err := m.ParseConfig(&validmetrics)
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet2(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet2", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.NoError(t, err)

	if len(metrics.items) == 1 {
		require.NoError(t, nil)
	} else if len(metrics.items) == 0 {
		var errorstring1 string = "No results returned from the query: " + string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(metrics.items) > 1 {
		var errorstring1 string = "Too many results returned from the query: " + string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet3(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 2)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"
	counters[1] = "% Idle Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet3", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.NoError(t, err)

	if len(metrics.items) == 2 {
		require.NoError(t, nil)
	} else if len(metrics.items) < 2 {

		var errorstring1 string = "Too few results returned from the query. " + string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(metrics.items) > 2 {

		var errorstring1 string = "Too many results returned from the query: " + string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet4(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 2)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0"
	counters[0] = "% Processor Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet4", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.NoError(t, err)

	if len(metrics.items) == 2 {
		require.NoError(t, nil)
	} else if len(metrics.items) < 2 {

		var errorstring1 string = "Too few results returned from the query: " + string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(metrics.items) > 2 {

		var errorstring1 string = "Too many results returned from the query: " + string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet5(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 2)
	var counters = make([]string, 2)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0"
	counters[0] = "% Processor Time"
	counters[1] = "% Idle Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet5", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.NoError(t, err)

	if len(metrics.items) == 4 {
		require.NoError(t, nil)
	} else if len(metrics.items) < 4 {
		var errorstring1 string = "Too few results returned from the query: " +
			string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(metrics.items) > 4 {
		var errorstring1 string = "Too many results returned from the query: " +
			string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet6(t *testing.T) {
	validmetrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "System"
	instances[0] = "------"
	counters[0] = "Context Switches/sec"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet6", Object: perfobjects}

	err := m.ParseConfig(&validmetrics)
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet7(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 3)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"
	counters[1] = "% Processor TimeERROR"
	counters[2] = "% Idle Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = false
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigGet7", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.NoError(t, err)

	if len(metrics.items) == 2 {
		require.NoError(t, nil)
	} else if len(metrics.items) < 2 {
		var errorstring1 string = "Too few results returned from the query: " +
			string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(metrics.items) > 2 {
		var errorstring1 string = "Too many results returned from the query: " +
			string(len(metrics.items))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigError1(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor InformationERROR"
	instances[0] = "_Total"
	counters[0] = "% Processor Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigError1", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.Error(t, err)
}

func TestWinPerfcountersConfigError2(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor"
	instances[0] = "SuperERROR"
	counters[0] = "% C1 Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigError2", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.Error(t, err)
}

func TestWinPerfcountersConfigError3(t *testing.T) {
	metrics := itemList{}

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "% Processor TimeERROR"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "ConfigError3", Object: perfobjects}

	err := m.ParseConfig(&metrics)
	require.Error(t, err)
}

func TestWinPerfcountersCollect1(t *testing.T) {

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "Parking Status"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "Collect1", Object: perfobjects}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)

	tags := map[string]string{
		"instance":   instances[0],
		"objectname": objectname,
	}
	fields := map[string]interface{}{
		counters[0]: float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)

}
func TestWinPerfcountersCollect2(t *testing.T) {

	var instances = make([]string, 2)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0,0"
	counters[0] = "Performance Limit Flags"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = false

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "Collect2", Object: perfobjects}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)

	tags := map[string]string{
		"instance":   instances[0],
		"objectname": objectname,
	}
	fields := map[string]interface{}{
		counters[0]: float32(0),
	}

	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
	tags = map[string]string{
		"instance":   instances[1],
		"objectname": objectname,
	}
	fields = map[string]interface{}{
		counters[0]: float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
}

func TestWinPerfcountersCollectExcludeObjectName(t *testing.T) {

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "Parking Status"

	var expectedCounterName = "Parking_Status"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false
	var excludeobjectname bool = true

	PerfObject := perfobject{
		ObjectName:        objectname,
		Instances:         instances,
		Counters:          counters,
		Measurement:       measurement,
		WarnOnMissing:     warnonmissing,
		FailOnMissing:     failonmissing,
		IncludeTotal:      includetotal,
		ExcludeObjectName: excludeobjectname,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, TestName: "CollectExcludeObjectName", Object: perfobjects}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)

	tags := map[string]string{
		"instance": instances[0],
	}
	fields := map[string]interface{}{
		expectedCounterName: float32(0),
	}

	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
}
