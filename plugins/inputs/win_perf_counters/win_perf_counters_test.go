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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet2(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 1 {
		require.NoError(t, nil)
	} else if len(parsedItems) == 0 {
		var errorstring1 string = "No results returned from the query: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 1 {
		var errorstring1 string = "Too many results returned from the query: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet3(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {

		var errorstring1 string = "Too few results returned from the query. " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {

		var errorstring1 string = "Too many results returned from the query: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet4(t *testing.T) {

	var instances = make([]string, 2)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0,1"
	counters[0] = "% Processor Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {

		var errorstring1 string = "Too few results returned from the query: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {

		var errorstring1 string = "Too many results returned from the query: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet5(t *testing.T) {

	var instances = make([]string, 2)
	var counters = make([]string, 2)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	instances[1] = "0,1"
	counters[0] = "% Processor Time"
	counters[1] = "% Idle Time"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 4 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 4 {
		var errorstring1 string = "Too few results returned from the query: " +
			string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 4 {
		var errorstring1 string = "Too many results returned from the query: " +
			string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigGet6(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)
}

func TestWinPerfcountersConfigGet7(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {
		var errorstring1 string = "Too few results returned from the query: " +
			string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {
		var errorstring1 string = "Too many results returned from the query: " +
			string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	}
}

func TestWinPerfcountersConfigError1(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersConfigError2(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersConfigError3(t *testing.T) {

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

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}

	err := m.ParseConfig()
	require.Error(t, err)
}

func TestWinPerfcountersCollect1(t *testing.T) {

	var instances = make([]string, 1)
	var counters = make([]string, 1)
	var perfobjects = make([]perfobject, 1)

	objectname := "Processor Information"
	instances[0] = "_Total"
	counters[0] = "Parking Status"

	var expectedCounter string = "Parking_Status"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}
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
		expectedCounter: float32(0),
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

	var expectedCounter string = "Performance_Limit_Flags"

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true
	var includetotal bool = false

	PerfObject := perfobject{
		ObjectName:    objectname,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
		IncludeTotal:  includetotal,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects}
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
		expectedCounter: float32(0),
	}

	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
	tags = map[string]string{
		"instance":   instances[1],
		"objectname": objectname,
	}
	fields = map[string]interface{}{
		expectedCounter: float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)

}
