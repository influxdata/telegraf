// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCounter struct {
	handle PDH_HCOUNTER
	path   string
	value  float64
	status uint32 // allows for tests against specific pdh_error codes, rather than assuming all cases of "value == 0" to indicate error conditions
}
type FakePerformanceQuery struct {
	counters      map[string]testCounter
	vistaAndNewer bool
	expandPaths   map[string][]string
	openCalled    bool
}

var MetricTime = time.Date(2018, 5, 28, 12, 0, 0, 0, time.UTC)

func (m *testCounter) ToCounterValue() *CounterValue {
	_, inst, _, _ := extractCounterInfoFromCounterPath(m.path)
	if inst == "" {
		inst = "--"
	}
	return &CounterValue{inst, m.value}
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
			if counter.status > 0 {
				return 0, NewPdhError(counter.status)
			}
			return counter.value, nil
		}
	}
	return 0, fmt.Errorf("GetFormattedCounterValueDouble: invalid handle: %d", counterHandle)
}
func (m *FakePerformanceQuery) findCounterByPath(counterPath string) *testCounter {
	for _, c := range m.counters {
		if c.path == counterPath {
			return &c
		}
	}
	return nil
}

func (m *FakePerformanceQuery) findCounterByHandle(counterHandle PDH_HCOUNTER) *testCounter {
	for _, c := range m.counters {
		if c.handle == counterHandle {
			return &c
		}
	}
	return nil
}

func (m *FakePerformanceQuery) GetFormattedCounterArrayDouble(hCounter PDH_HCOUNTER) ([]CounterValue, error) {
	if !m.openCalled {
		return nil, errors.New("GetFormattedCounterArrayDouble: uninitialised query")
	}
	for _, c := range m.counters {
		if c.handle == hCounter {
			if e, ok := m.expandPaths[c.path]; ok {
				counters := make([]CounterValue, 0, len(e))
				for _, p := range e {
					counter := m.findCounterByPath(p)
					if counter != nil {
						if counter.status > 0 {
							return nil, NewPdhError(counter.status)
						}
						counters = append(counters, *counter.ToCounterValue())
					} else {
						return nil, fmt.Errorf("GetFormattedCounterArrayDouble: invalid counter : %s", p)
					}
				}
				return counters, nil
			} else {
				return nil, fmt.Errorf("GetFormattedCounterArrayDouble: invalid counter : %d", hCounter)
			}
		}
	}
	return nil, fmt.Errorf("GetFormattedCounterArrayDouble: invalid counter : %d, no paths found", hCounter)
}

func (m *FakePerformanceQuery) CollectData() error {
	if !m.openCalled {
		return errors.New("CollectData: uninitialised query")
	}
	return nil
}

func (m *FakePerformanceQuery) CollectDataWithTime() (time.Time, error) {
	if !m.openCalled {
		return time.Now(), errors.New("CollectData: uninitialised query")
	}
	return MetricTime, nil
}

func (m *FakePerformanceQuery) IsVistaOrNewer() bool {
	return m.vistaAndNewer
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

func createCounterMap(counterPaths []string, values []float64, status []uint32) map[string]testCounter {
	counters := make(map[string]testCounter)
	for i, cp := range counterPaths {
		counters[cp] = testCounter{
			PDH_HCOUNTER(i),
			cp,
			values[i],
			status[i],
		}
	}
	return counters
}

var counterPathsAndRes = map[string][]string{
	"\\O\\CT":                           {"O", "", "CT"},
	"\\O\\CT(i)":                        {"O", "", "CT(i)"},
	"\\O\\CT(d:\\f\\i)":                 {"O", "", "CT(d:\\f\\i)"},
	"\\\\CM\\O\\CT":                     {"O", "", "CT"},
	"\\O(I)\\CT":                        {"O", "I", "CT"},
	"\\O(I)\\CT(i)":                     {"O", "I", "CT(i)"},
	"\\O(I)\\CT(i)x":                    {"O", "I", "CT(i)x"},
	"\\O(I)\\CT(d:\\f\\i)":              {"O", "I", "CT(d:\\f\\i)"},
	"\\\\CM\\O(I)\\CT":                  {"O", "I", "CT"},
	"\\O(d:\\f\\I)\\CT":                 {"O", "d:\\f\\I", "CT"},
	"\\O(d:\\f\\I(d))\\CT":              {"O", "d:\\f\\I(d)", "CT"},
	"\\O(d:\\f\\I(d)x)\\CT":             {"O", "d:\\f\\I(d)x", "CT"},
	"\\O(d:\\f\\I)\\CT(i)":              {"O", "d:\\f\\I", "CT(i)"},
	"\\O(d:\\f\\I)\\CT(d:\\f\\i)":       {"O", "d:\\f\\I", "CT(d:\\f\\i)"},
	"\\\\CM\\O(d:\\f\\I)\\CT":           {"O", "d:\\f\\I", "CT"},
	"\\\\CM\\O(d:\\f\\I)\\CT(d:\\f\\i)": {"O", "d:\\f\\I", "CT(d:\\f\\i)"},
	"\\O(I(info))\\CT":                  {"O", "I(info)", "CT"},
	"\\\\CM\\O(I(info))\\CT":            {"O", "I(info)", "CT"},
}

var invalidCounterPaths = []string{
	"\\O(I\\C",
	"\\OI)\\C",
	"\\O(I\\C",
	"\\O/C",
	"\\O(I/C",
	"\\O(I/C)",
	"\\O(I\\)C",
	"\\O(I\\C)",
}

func TestCounterPathParsing(t *testing.T) {
	for path, vals := range counterPathsAndRes {
		o, i, c, err := extractCounterInfoFromCounterPath(path)
		require.NoError(t, err)
		require.True(t, assert.ObjectsAreEqual(vals, []string{o, i, c}), "arrays: %#v and %#v are not equal", vals, []string{o, i, c})
	}
	for _, path := range invalidCounterPaths {
		_, _, _, err := extractCounterInfoFromCounterPath(path)
		require.Error(t, err)
	}
}

func TestAddItemSimple(t *testing.T) {
	var err error
	cps1 := []string{"\\O(I)\\C"}
	m := Win_PerfCounters{PrintValid: false, Object: nil, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1}, []uint32{0}),
		expandPaths: map[string][]string{
			cps1[0]: cps1,
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.AddItem(cps1[0], "O", "I", "c", "test", false)
	require.NoError(t, err)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestAddItemInvalidCountPath(t *testing.T) {
	var err error
	cps1 := []string{"\\O\\C"}
	m := Win_PerfCounters{PrintValid: false, Object: nil, UseWildcardsExpansion: true, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1}, []uint32{0}),
		expandPaths: map[string][]string{
			cps1[0]: {"\\O/C"},
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.AddItem("\\O\\C", "O", "------", "C", "test", false)
	require.Error(t, err)
	err = m.query.Close()
	require.NoError(t, err)
}

func TestParseConfigBasic(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"I1", "I2"}, []string{"C1", "C2"}, false, false)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1, 1.2, 1.3, 1.4}, []uint32{0, 0, 0, 0}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
			cps1[3]: {cps1[3]},
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 4)
	err = m.query.Close()
	require.NoError(t, err)

	m.UseWildcardsExpansion = true
	m.counters = nil

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
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, UseWildcardsExpansion: false, query: &FakePerformanceQuery{
		counters: createCounterMap(cps1, []float64{1.1, 1.2}, []uint32{0, 0}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 2)
	err = m.query.Close()
	require.NoError(t, err)

	m.UseWildcardsExpansion = true
	m.counters = nil

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
		counters: createCounterMap(cps1, []float64{1.1, 1.2, 1.3}, []uint32{0, 0, 0}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.Error(t, err)
	err = m.query.Close()
	require.NoError(t, err)

	m.UseWildcardsExpansion = true
	m.counters = nil

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
		counters: createCounterMap(cps1, []float64{1.1, 1.2, 1.3}, []uint32{0, 0, 0}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	err = m.query.Close()
	require.NoError(t, err)

	m.UseWildcardsExpansion = true
	m.counters = nil

	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	err = m.query.Close()
	require.NoError(t, err)

}

func TestParseConfigTotalExpansion(t *testing.T) {
	var err error
	perfObjects := createPerfObject("m", "O", []string{"*"}, []string{"*"}, true, true)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(_Total)\\C1", "\\O(_Total)\\C2"}
	m := Win_PerfCounters{PrintValid: false, UseWildcardsExpansion: true, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}, []uint32{0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		vistaAndNewer: true,
	}}
	err = m.query.Open()
	require.NoError(t, err)
	err = m.ParseConfig()
	require.NoError(t, err)
	assert.Len(t, m.counters, 4)
	err = m.query.Close()
	require.NoError(t, err)

	perfObjects[0].IncludeTotal = false

	m = Win_PerfCounters{PrintValid: false, UseWildcardsExpansion: true, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}, []uint32{0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		vistaAndNewer: true,
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
	m := Win_PerfCounters{PrintValid: false, UseWildcardsExpansion: true, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}, []uint32{0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		vistaAndNewer: true,
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
		counters: createCounterMap([]string{cp1}, []float64{1.2}, []uint32{0}),
		expandPaths: map[string][]string{
			cp1: {cp1},
		},
		vistaAndNewer: false,
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

	m.UseWildcardsExpansion = true
	m.counters = nil
	m.lastRefreshed = time.Time{}

	var acc2 testutil.Accumulator

	err = m.Gather(&acc2)
	require.NoError(t, err)
	acc1.AssertContainsTaggedFields(t, measurement, fields1, tags1)
}

func TestSimpleGatherNoData(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"I"}, []string{"C"}, false, false)
	cp1 := "\\O(I)\\C"
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap([]string{cp1}, []float64{1.2}, []uint32{PDH_NO_DATA}),
		expandPaths: map[string][]string{
			cp1: {cp1},
		},
		vistaAndNewer: false,
	}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	// this "PDH_NO_DATA" error should not be returned to caller, but checked, and handled
	require.NoError(t, err)

	// fields would contain if the error was ignored, and we simply added garbage
	fields1 := map[string]interface{}{
		"C": float32(1.2),
	}
	// tags would contain if the error was ignored, and we simply added garbage
	tags1 := map[string]string{
		"instance":   "I",
		"objectname": "O",
	}
	acc1.AssertDoesNotContainsTaggedFields(t, measurement, fields1, tags1)

	m.UseWildcardsExpansion = true
	m.counters = nil
	m.lastRefreshed = time.Time{}

	var acc2 testutil.Accumulator

	err = m.Gather(&acc2)
	require.NoError(t, err)
	acc1.AssertDoesNotContainsTaggedFields(t, measurement, fields1, tags1)
}

func TestSimpleGatherWithTimestamp(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"I"}, []string{"C"}, false, false)
	cp1 := "\\O(I)\\C"
	m := Win_PerfCounters{PrintValid: false, UsePerfCounterTime: true, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap([]string{cp1}, []float64{1.2}, []uint32{0}),
		expandPaths: map[string][]string{
			cp1: {cp1},
		},
		vistaAndNewer: true,
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
	assert.True(t, acc1.HasTimestamp(measurement, MetricTime))
}

func TestGatherError(t *testing.T) {
	var err error
	expected_error := "error while getting value for counter \\O(I)\\C: The information passed is not valid.\r\n"
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"I"}, []string{"C"}, false, false)
	cp1 := "\\O(I)\\C"
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap([]string{cp1}, []float64{-2}, []uint32{PDH_PLA_VALIDATION_WARNING}),
		expandPaths: map[string][]string{
			cp1: {cp1},
		},
		vistaAndNewer: false,
	}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	require.Error(t, err)
	require.Equal(t, expected_error, err.Error())

	m.UseWildcardsExpansion = true
	m.counters = nil
	m.lastRefreshed = time.Time{}

	var acc2 testutil.Accumulator

	err = m.Gather(&acc2)
	require.Error(t, err)
	require.Equal(t, expected_error, err.Error())
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
		counters: createCounterMap(cps1, []float64{1.2, 1, 0}, []uint32{0, PDH_INVALID_DATA, 0}),
		expandPaths: map[string][]string{
			cps1[0]: {cps1[0]},
			cps1[1]: {cps1[1]},
			cps1[2]: {cps1[2]},
		},
		vistaAndNewer: false,
	}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	require.NoError(t, err)

	fields1 := map[string]interface{}{
		"C1": float32(1.2),
		"C3": float32(0),
	}
	tags1 := map[string]string{
		"instance":   "I",
		"objectname": "O",
	}
	acc1.AssertContainsTaggedFields(t, measurement, fields1, tags1)

	m.UseWildcardsExpansion = true
	m.counters = nil
	m.lastRefreshed = time.Time{}

	var acc2 testutil.Accumulator
	err = m.Gather(&acc2)
	require.NoError(t, err)
	acc1.AssertContainsTaggedFields(t, measurement, fields1, tags1)
}

//tests with expansion
func TestGatherRefreshingWithExpansion(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"*"}, []string{"*"}, true, false)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	fpm := &FakePerformanceQuery{
		counters: createCounterMap(append(cps1, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 0}, []uint32{0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps1,
		},
		vistaAndNewer: true,
	}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, UseWildcardsExpansion: true, query: fpm, CountersRefreshInterval: internal.Duration{Duration: time.Second * 10}}
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
		counters: createCounterMap(append(cps2, "\\O(*)\\*"), []float64{1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 0}, []uint32{0, 0, 0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\*": cps2,
		},
		vistaAndNewer: true,
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

func TestGatherRefreshingWithoutExpansion(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping long taking test in short mode")
	}
	measurement := "test"
	perfObjects := createPerfObject(measurement, "O", []string{"*"}, []string{"C1", "C2"}, true, false)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2"}
	fpm := &FakePerformanceQuery{
		counters: createCounterMap(append([]string{"\\O(*)\\C1", "\\O(*)\\C2"}, cps1...), []float64{0, 0, 1.1, 1.2, 1.3, 1.4}, []uint32{0, 0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\C1": {cps1[0], cps1[2]},
			"\\O(*)\\C2": {cps1[1], cps1[3]},
		},
		vistaAndNewer: true,
	}
	m := Win_PerfCounters{PrintValid: false, Object: perfObjects, UseWildcardsExpansion: false, query: fpm, CountersRefreshInterval: internal.Duration{Duration: time.Second * 10}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	assert.Len(t, m.counters, 2)
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
	//test finding new instance
	cps2 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I2)\\C1", "\\O(I2)\\C2", "\\O(I3)\\C1", "\\O(I3)\\C2"}
	fpm = &FakePerformanceQuery{
		counters: createCounterMap(append([]string{"\\O(*)\\C1", "\\O(*)\\C2"}, cps2...), []float64{0, 0, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6}, []uint32{0, 0, 0, 0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\C1": {cps2[0], cps2[2], cps2[4]},
			"\\O(*)\\C2": {cps2[1], cps2[3], cps2[5]},
		},
		vistaAndNewer: true,
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
	assert.Len(t, m.counters, 2)
	assert.Len(t, acc2.Metrics, 3)

	acc2.AssertContainsTaggedFields(t, measurement, fields1, tags1)
	acc2.AssertContainsTaggedFields(t, measurement, fields2, tags2)
	acc2.AssertContainsTaggedFields(t, measurement, fields3, tags3)
	//test changed configuration
	perfObjects = createPerfObject(measurement, "O", []string{"*"}, []string{"C1", "C2", "C3"}, true, false)
	cps3 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(I1)\\C3", "\\O(I2)\\C1", "\\O(I2)\\C2", "\\O(I2)\\C3"}
	fpm = &FakePerformanceQuery{
		counters: createCounterMap(append([]string{"\\O(*)\\C1", "\\O(*)\\C2", "\\O(*)\\C3"}, cps3...), []float64{0, 0, 0, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6}, []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\C1": {cps3[0], cps3[3]},
			"\\O(*)\\C2": {cps3[1], cps3[4]},
			"\\O(*)\\C3": {cps3[2], cps3[5]},
		},
		vistaAndNewer: true,
	}
	m.query = fpm
	m.Object = perfObjects

	fpm.Open()

	time.Sleep(m.CountersRefreshInterval.Duration)

	var acc3 testutil.Accumulator
	err = m.Gather(&acc3)
	require.NoError(t, err)
	assert.Len(t, acc3.Metrics, 2)
	fields4 := map[string]interface{}{
		"C1": float32(1.1),
		"C2": float32(1.2),
		"C3": float32(1.3),
	}
	tags4 := map[string]string{
		"instance":   "I1",
		"objectname": "O",
	}
	fields5 := map[string]interface{}{
		"C1": float32(1.4),
		"C2": float32(1.5),
		"C3": float32(1.6),
	}
	tags5 := map[string]string{
		"instance":   "I2",
		"objectname": "O",
	}

	acc3.AssertContainsTaggedFields(t, measurement, fields4, tags4)
	acc3.AssertContainsTaggedFields(t, measurement, fields5, tags5)

}

func TestGatherTotalNoExpansion(t *testing.T) {
	var err error
	measurement := "m"
	perfObjects := createPerfObject(measurement, "O", []string{"*"}, []string{"C1", "C2"}, true, true)
	cps1 := []string{"\\O(I1)\\C1", "\\O(I1)\\C2", "\\O(_Total)\\C1", "\\O(_Total)\\C2"}
	m := Win_PerfCounters{PrintValid: false, UseWildcardsExpansion: false, Object: perfObjects, query: &FakePerformanceQuery{
		counters: createCounterMap(append([]string{"\\O(*)\\C1", "\\O(*)\\C2"}, cps1...), []float64{0, 0, 1.1, 1.2, 1.3, 1.4}, []uint32{0, 0, 0, 0, 0, 0}),
		expandPaths: map[string][]string{
			"\\O(*)\\C1": {cps1[0], cps1[2]},
			"\\O(*)\\C2": {cps1[1], cps1[3]},
		},
		vistaAndNewer: true,
	}}
	var acc1 testutil.Accumulator
	err = m.Gather(&acc1)
	require.NoError(t, err)
	assert.Len(t, m.counters, 2)
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
		"instance":   "_Total",
		"objectname": "O",
	}
	acc1.AssertContainsTaggedFields(t, measurement, fields2, tags2)

	perfObjects[0].IncludeTotal = false

	m.counters = nil
	m.lastRefreshed = time.Time{}

	var acc2 testutil.Accumulator
	err = m.Gather(&acc2)
	require.NoError(t, err)
	assert.Len(t, m.counters, 2)
	assert.Len(t, acc2.Metrics, 1)

	acc2.AssertContainsTaggedFields(t, measurement, fields1, tags1)

	acc2.AssertDoesNotContainsTaggedFields(t, measurement, fields2, tags2)
}

// list of nul terminated strings from WinAPI
var unicodeStringListWithEnglishChars = []uint16{0x5c, 0x5c, 0x54, 0x34, 0x38, 0x30, 0x5c, 0x50, 0x68, 0x79, 0x73, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x69, 0x73, 0x6b, 0x28, 0x30, 0x20, 0x43, 0x3a, 0x29, 0x5c, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x20, 0x44, 0x69, 0x73, 0x6b, 0x20, 0x51, 0x75, 0x65, 0x75, 0x65, 0x20, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x0, 0x5c, 0x5c, 0x54, 0x34, 0x38, 0x30, 0x5c, 0x50, 0x68, 0x79, 0x73, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x69, 0x73, 0x6b, 0x28, 0x5f, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x29, 0x5c, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x20, 0x44, 0x69, 0x73, 0x6b, 0x20, 0x51, 0x75, 0x65, 0x75, 0x65, 0x20, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x0, 0x0}
var unicodeStringListWithCzechChars = []uint16{0x5c, 0x5c, 0x54, 0x34, 0x38, 0x30, 0x5c, 0x46, 0x79, 0x7a, 0x69, 0x63, 0x6b, 0xfd, 0x20, 0x64, 0x69, 0x73, 0x6b, 0x28, 0x30, 0x20, 0x43, 0x3a, 0x29, 0x5c, 0x41, 0x6b, 0x74, 0x75, 0xe1, 0x6c, 0x6e, 0xed, 0x20, 0x64, 0xe9, 0x6c, 0x6b, 0x61, 0x20, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x79, 0x20, 0x64, 0x69, 0x73, 0x6b, 0x75, 0x0, 0x5c, 0x5c, 0x54, 0x34, 0x38, 0x30, 0x5c, 0x46, 0x79, 0x7a, 0x69, 0x63, 0x6b, 0xfd, 0x20, 0x64, 0x69, 0x73, 0x6b, 0x28, 0x5f, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x29, 0x5c, 0x41, 0x6b, 0x74, 0x75, 0xe1, 0x6c, 0x6e, 0xed, 0x20, 0x64, 0xe9, 0x6c, 0x6b, 0x61, 0x20, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x79, 0x20, 0x64, 0x69, 0x73, 0x6b, 0x75, 0x0, 0x0}
var unicodeStringListSingleItem = []uint16{0x5c, 0x5c, 0x54, 0x34, 0x38, 0x30, 0x5c, 0x50, 0x68, 0x79, 0x73, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x69, 0x73, 0x6b, 0x28, 0x30, 0x20, 0x43, 0x3a, 0x29, 0x5c, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x20, 0x44, 0x69, 0x73, 0x6b, 0x20, 0x51, 0x75, 0x65, 0x75, 0x65, 0x20, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x0, 0x0}
var unicodeStringListNoItem = []uint16{0x0}

var stringArrayWithEnglishChars = []string{
	"\\\\T480\\PhysicalDisk(0 C:)\\Current Disk Queue Length",
	"\\\\T480\\PhysicalDisk(_Total)\\Current Disk Queue Length",
}
var stringArrayWithCzechChars = []string{
	"\\\\T480\\Fyzick\u00fd disk(0 C:)\\Aktu\u00e1ln\u00ed d\u00e9lka fronty disku",
	"\\\\T480\\Fyzick\u00fd disk(_Total)\\Aktu\u00e1ln\u00ed d\u00e9lka fronty disku",
}

var stringArraySingleItem = []string{
	"\\\\T480\\PhysicalDisk(0 C:)\\Current Disk Queue Length",
}

func TestUTF16ToStringArray(t *testing.T) {
	singleItem := UTF16ToStringArray(unicodeStringListSingleItem)
	assert.True(t, assert.ObjectsAreEqual(singleItem, stringArraySingleItem), "Not equal single arrays")

	noItem := UTF16ToStringArray(unicodeStringListNoItem)
	assert.Nil(t, noItem)

	engStrings := UTF16ToStringArray(unicodeStringListWithEnglishChars)
	assert.True(t, assert.ObjectsAreEqual(engStrings, stringArrayWithEnglishChars), "Not equal eng arrays")

	czechStrings := UTF16ToStringArray(unicodeStringListWithCzechChars)
	assert.True(t, assert.ObjectsAreEqual(czechStrings, stringArrayWithCzechChars), "Not equal czech arrays")
}
