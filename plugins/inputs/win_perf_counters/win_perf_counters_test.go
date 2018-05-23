// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type testCounter struct {
	handle PDH_HCOUNTER
	path   string
	value  float64
}
type FakePerformanceQuery struct {
	counters            map[string]testCounter
	addEnglishSupported bool
	expandPaths         map[string][]string
	openCalled          bool
}

func (m *FakePerformanceQuery) Open() error {
	if m.openCalled {
		err := m.Close()
		if err != nil {
			return err
		}
	}
	m.openCalled = true
	return nil
}

func (m *FakePerformanceQuery) Close() error {
	if !m.openCalled {
		return errors.New("CloSe: uninitialised query")
	}
	m.openCalled = false
	return nil
}

func (m *FakePerformanceQuery) AddCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	if !m.openCalled {
		return 0, errors.New("AddCounterToQuery: uninitialised query")
	}
	if c, ok := m.counters[counterPath]; ok {
		return c.handle, nil
	} else {
		return 0, errors.New(fmt.Sprintf("AddCounterToQuery: invalid counter path: %s", counterPath))
	}
}

func (m *FakePerformanceQuery) AddEnglishCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	if !m.openCalled {
		return 0, errors.New("AddEnglishCounterToQuery: uninitialised query")
	}
	if c, ok := m.counters[counterPath]; ok {
		return c.handle, nil
	} else {
		return 0, fmt.Errorf("AddEnglishCounterToQuery: invalid counter path: %s", counterPath)
	}
}

func (m *FakePerformanceQuery) GetCounterPath(counterHandle PDH_HCOUNTER) (string, error) {
	for _, counter := range m.counters {
		if counter.handle == counterHandle {
			return counter.path, nil
		}
	}
	return "", fmt.Errorf("GetCounterPath: invalid handle: %d", counterHandle)
}

func (m *FakePerformanceQuery) ExpandWildCardPath(counterPath string) ([]string, error) {
	if e, ok := m.expandPaths[counterPath]; ok {
		return e, nil
	} else {
		return []string{}, fmt.Errorf("ExpandWildCardPath: invalid counter path: %s", counterPath)
	}
}

func (m *FakePerformanceQuery) GetFormattedCounterValueDouble(counterHandle PDH_HCOUNTER) (float64, error) {
	if !m.openCalled {
		return 0, errors.New("GetFormattedCounterValueDouble: uninitialised query")
	}
	for _, counter := range m.counters {
		if counter.handle == counterHandle {
			if counter.value > 0 {
				return counter.value, nil
			} else {
				if counter.value == 0 {
					return 0, NewPdhError(PDH_INVALID_DATA)
				} else {
					return 0, NewPdhError(PDH_CALC_NEGATIVE_VALUE)
				}
			}
		}
	}
	return 0, fmt.Errorf("GetFormattedCounterValueDouble: invalid handle: %d", counterHandle)
}

func (m *FakePerformanceQuery) CollectData() error {
	if !m.openCalled {
		return errors.New("CollectData: uninitialised query")
	}
	return nil
}

func (m *FakePerformanceQuery) AddEnglishCounterSupported() bool {
	return m.addEnglishSupported
}

func createPerfObject(measurement string, object string, instances []string, counters []string, failOnMissing bool, includeTotal bool) []perfobject {
	PerfObject := perfobject{
		ObjectName:    object,
		Instances:     instances,
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: false,
		FailOnMissing: failOnMissing,
		IncludeTotal:  includeTotal,
	}
	perfobjects := []perfobject{PerfObject}
	return perfobjects
}

func createCounterMap(counterPaths []string, values []float64) map[string]testCounter {
	counters := make(map[string]testCounter)
	for i, cp := range counterPaths {
		counters[cp] = testCounter{
			PDH_HCOUNTER(i),
			cp,
			values[i],
		}
	}
	return counters
}

func TestAddItemSimple(t *testing.T) {
	var err error
	cps1 := []string{"\\O(I)\\C"}
	m := Win_PerfCounters{PrintValid: false, Object: nil, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1}),
		expandPaths: map[string][]string{
			cps1[0]: cps1,
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.AddItem(cps1[0], "I", "test", false)
	require.NoError(t, err)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestAddItemInvalidCountPath(t *testing.T) {
	var err error
	cps1 := []string{"\\O\\C"}
	m := Win_PerfCounters{PrintValid: false, Object: nil, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1}),
		expandPaths: map[string][]string{
			cps1[0]: {"\\O/C"},
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.AddItem("\\O\\C", "*", "test", false)
	require.Error(t, err)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigBasic(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"I1", "I2"}, []string{"C1", "C2"}, false, false)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1, 1.2, 1.3, 1.4}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
			cps1[3]: {cps1[3]},
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 4)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigNoInstance(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"------"}, []string{"C1", "C2"}, false, false)
	cps1 := []string{"\\O\\C1", "\\O\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1, 1.2}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 2)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigInvalidCounterError(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"I1", "I2"}, []string{"C1", "C2"}, true, false)
	cps1 := []string{"\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1, 1.2, 1.3}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.Error(t, err)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigInvalidCounterNoError(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"I1", "I2"}, []string{"C1", "C2"}, false, false)
	cps1 := []string{"\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1, 1.2, 1.3}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigTotal(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"*"}, []string{"*"}, true, true)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(_Total)\\C1", "\\O(_Total)\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 4)
	err = m.query.Close()
	require.NoError(t, err)

	perfObjects[0].IncludeTotal = false

	m = Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 2)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigExpand(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"*"}, []string{"*"}, false, false)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		addEnglishSupported: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 4)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestSimpleGather(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"I"}, []string{"C"}, false, false)
	cp1 := "\\O(I)\\C"
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap([]string{cp1}, []float64{1.2}),
		expandPaths: map[string][]string{
			cp1: {cp1},
		},
		addEnglishSupported: false,
	}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	require.NoError(t, err)

	fields1 := map[string]interface{}{
		"C": float32(1.2),
	}
	tags1 := map[string]string{
		"instance":   "I",
		"objectname": "O",
	}
	acc1.AssertContainsTaggedFields(t, measurement, fields1, tags1)
}

func TestGatherInvalidDataIgnore(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"I"}, []string{"C1", "C2", "C3"}, false, false)
	cps1 := []string{"\\O(I)\\C1", "\\O(I)\\C2", "\\O(I)\\C3"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.2, -1, 0}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
		},
		addEnglishSupported: false,
	}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	require.NoError(t, err)

	fields1 := map[string]interface{}{
		"C1": float32(1.2),
	}
	tags1 := map[string]string{
		"instance":   "I",
		"objectname": "O",
	}
	acc1.AssertContainsTaggedFields(t, measurement, fields1, tags1)
}

func TestGatherRefreshing(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"*"}, []string{"*"}, false, false)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	fpm := &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		addEnglishSupported: true,
	}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: fpm, CountersRefreshInterval: internal.Duration{Duration: time.Second * 10}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	assert.Len(t, m.counters, 4)
	require.NoError(t, err)
	assert.Len(t, acc1.Metrics, 2)

	fields1 := map[string]interface{}{
		"C1": float32(1.1),
		"C2": float32(1.2),
	}
	tags1 := map[string]string{
		"instance":   "I1",
		"objectname": "O",
	}
	acc1.AssertContainsTaggedFields(t, measurement, fields1, tags1)

	fields2 := map[string]interface{}{
		"C1": float32(1.3),
		"C2": float32(1.4),
	}
	tags2 := map[string]string{
		"instance":   "I2",
		"objectname": "O",
	}
	acc1.AssertContainsTaggedFields(t, measurement, fields2, tags2)
	cps2 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2", "\\O(I3)\\C1", "\\O(I3)\\C2"}
	fpm = &FakePerformanceQuery{
		counters: createCounterMap(append(cps2, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps2,
		},
		addEnglishSupported: true,
	}
	m.query = fpm
	fpm.Open()
	var acc2 testutil.Accumulator

	fields3 := map[string]interface{}{
		"C1": float32(1.5),
		"C2": float32(1.6),
	}
	tags3 := map[string]string{
		"instance":   "I3",
		"objectname": "O",
	}

	//test before elapsing CounterRefreshRate counters are not refreshed
	err = m.Gather(&acc2)
	require.NoError(t, err)
	assert.Len(t, m.counters, 4)
	assert.Len(t, acc2.Metrics, 2)

	acc2.AssertContainsTaggedFields(t, measurement, fields1, tags1)
	acc2.AssertContainsTaggedFields(t, measurement, fields2, tags2)
	acc2.AssertDoesNotContainsTaggedFields(t, measurement, fields3, tags3)
	time.Sleep(m.CountersRefreshInterval.Duration)

	var acc3 testutil.Accumulator
	err = m.Gather(&acc3)
	require.NoError(t, err)
	assert.Len(t, acc3.Metrics, 3)

	acc3.AssertContainsTaggedFields(t, measurement, fields1, tags1)
	acc3.AssertContainsTaggedFields(t, measurement, fields2, tags2)

	acc3.AssertContainsTaggedFields(t, measurement, fields3, tags3)

}
