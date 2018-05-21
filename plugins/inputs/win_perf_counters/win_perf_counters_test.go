// +build windows

package win_perf_counters

import (
	"errors"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type testCounter struct {
	handle PDH_HCOUNTER
	path string
	value float32
}
type FakePerformanceQuery struct {
	counters map[string]testCounter
	addEnglishSupported bool
	expanded map[string][]string
}

func (m *FakePerformanceQuery) Open() error {
	return nil
}

func (m *FakePerformanceQuery) Close() error {
	return nil
}

func (m *FakePerformanceQuery) AddCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	if c, ok := m.counters[counterPath]; ok {
		return c.handle, nil
	} else {
		return 0, errors.New("invalid path")
	}
}

func (m *FakePerformanceQuery) AddEnglishCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	if c, ok := m.counters[counterPath]; ok {
		return c.handle, nil
	} else {
		return 0, errors.New("invalid path")
	}
}

func (m *FakePerformanceQuery) GetCounterPath(counterHandle PDH_HCOUNTER) (string, error) {
	for _, counter := range m.counters {
		if counter.handle == counterHandle {
			return counter.path, nil
		}
	}
	return "", errors.New("invalid handle")
}

func (m *FakePerformanceQuery) ExpandWildCardPath(counterPath string) ([]string, error) {
	if e, ok := m.expanded[counterPath]; ok {
		return e, nil
	} else {
		return []string{}, errors.New("invalid path")
	}
}

func (m *FakePerformanceQuery) GetFormattedCounterValueDouble(hCounter PDH_HCOUNTER) (float64, error) {
	panic("implement me")
}

func (m *FakePerformanceQuery) CollectData() error {
	return nil
}

func (m *FakePerformanceQuery) AddEnglishCounterSupported() bool {
	return m.addEnglishSupported
}

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


	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

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
	var includetotal bool = true

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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 1 {
		require.NoError(t, nil)
	} else if len(parsedItems) == 0 {
		var errorstring1 string = "No results returned from the counterPath: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 1 {
		var errorstring1 string = "Too many results returned from the counterPath: " + string(len(parsedItems))
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
	var includetotal bool = true

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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {

		var errorstring1 string = "Too few results returned from the counterPath. " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {

		var errorstring1 string = "Too many results returned from the counterPath: " + string(len(parsedItems))
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
	var includetotal bool = true

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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {

		var errorstring1 string = "Too few results returned from the counterPath: " + string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {

		var errorstring1 string = "Too many results returned from the counterPath: " + string(len(parsedItems))
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {
		var errorstring1 string = "Too few results returned from the counterPath: " +
			string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {
		var errorstring1 string = "Too many results returned from the counterPath: " +
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

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

	PerfObject := perfobject{
		objectname,
		counters,
		instances,
		measurement,
		false,
		false,
		true,
	}

	perfobjects[0] = PerfObject

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	require.NoError(t, err)

	var parsedItems = m.GetParsedItemsForTesting()

	if len(parsedItems) == 2 {
		require.NoError(t, nil)
	} else if len(parsedItems) < 2 {
		var errorstring1 string = "Too few results returned from the counterPath: " +
			string(len(parsedItems))
		err2 := errors.New(errorstring1)
		require.NoError(t, err2)
	} else if len(parsedItems) > 2 {
		var errorstring1 string = "Too many results returned from the counterPath: " +
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

	err := m.ParseConfig()
	var acc testutil.Accumulator
	err = m.Gather(&acc)
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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
	m.query.Open()

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
	var includetotal bool = true

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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
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
	var includetotal bool = true

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

	m := Win_PerfCounters{PrintValid: false, Object: perfobjects, query: &PerformanceQueryImpl{}}
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
		expectedCounter: float32(2),
	}

	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
	tags = map[string]string{
		"instance":   instances[1],
		"objectname": objectname,
	}
	fields = map[string]interface{}{
		expectedCounter: float32(2),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)

}
