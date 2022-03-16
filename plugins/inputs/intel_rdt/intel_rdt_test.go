//go:build !windows
// +build !windows

package intel_rdt

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

type MockProc struct{}

func (m *MockProc) getAllProcesses() ([]Process, error) {
	procs := []Process{
		{Name: "process", PID: 1000},
		{Name: "process2", PID: 1002},
		{Name: "process2", PID: 1003},
	}
	return procs, nil
}

func TestAssociateProcessesWithPIDs(t *testing.T) {
	log := testutil.Logger{}
	proc := &MockProc{}
	rdt := IntelRDT{
		Log:       log,
		Processor: proc,
	}
	processes := []string{"process"}
	expectedPID := "1000"
	result, err := rdt.associateProcessesWithPIDs(processes)
	assert.Nil(t, err)
	assert.Equal(t, expectedPID, result[processes[0]])

	processes = []string{"process2"}
	expectedPID = "1002,1003"
	result, err = rdt.associateProcessesWithPIDs(processes)
	assert.Nil(t, err)
	assert.Equal(t, expectedPID, result[processes[0]])

	processes = []string{"process1"}
	result, err = rdt.associateProcessesWithPIDs(processes)
	assert.Nil(t, err)
	assert.Len(t, result, 0)
}

func TestSplitCSVLineIntoValues(t *testing.T) {
	line := "2020-08-12 13:34:36,\"45417,29170\",37,44,0.00,0,0.0,0.0,0.0,0.0"
	expectedTimeValue := "2020-08-12 13:34:36"
	expectedMetricsValue := []string{"0.00", "0", "0.0", "0.0", "0.0", "0.0"}
	expectedCoreOrPidsValue := []string{"\"45417", "29170\"", "37", "44"}

	splitCSV, err := splitCSVLineIntoValues(line)
	assert.Nil(t, err)
	assert.Equal(t, expectedTimeValue, splitCSV.timeValue)
	assert.Equal(t, expectedMetricsValue, splitCSV.metricsValues)
	assert.Equal(t, expectedCoreOrPidsValue, splitCSV.coreOrPIDsValues)

	wrongLine := "2020-08-12 13:34:36,37,44,0.00,0,0.0"
	splitCSV, err = splitCSVLineIntoValues(wrongLine)
	assert.NotNil(t, err)
	assert.Equal(t, "", splitCSV.timeValue)
	assert.Nil(t, nil, splitCSV.metricsValues)
	assert.Nil(t, nil, splitCSV.coreOrPIDsValues)
}

func TestFindPIDsInMeasurement(t *testing.T) {
	line := "2020-08-12 13:34:36,\"45417,29170\""
	expected := "45417,29170"
	result, err := findPIDsInMeasurement(line)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)

	line = "pids not included"
	result, err = findPIDsInMeasurement(line)
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestCreateArgsProcesses(t *testing.T) {
	processesPIDs := map[string]string{
		"process": "12345, 99999",
	}
	expected := "--mon-pid=all:[12345, 99999];mbt:[12345, 99999];"
	result := createArgProcess(processesPIDs)
	assert.EqualValues(t, expected, result)

	processesPIDs = map[string]string{
		"process":  "12345, 99999",
		"process2": "44444, 11111",
	}
	expectedPrefix := "--mon-pid="
	expectedSubstring := "all:[12345, 99999];mbt:[12345, 99999];"
	expectedSubstring2 := "all:[44444, 11111];mbt:[44444, 11111];"
	result = createArgProcess(processesPIDs)
	assert.Contains(t, result, expectedPrefix)
	assert.Contains(t, result, expectedSubstring)
	assert.Contains(t, result, expectedSubstring2)
}

func TestCreateArgsCores(t *testing.T) {
	cores := []string{"1,2,3"}
	expected := "--mon-core=all:[1,2,3];mbt:[1,2,3];"
	result := createArgCores(cores)
	assert.EqualValues(t, expected, result)

	cores = []string{"1,2,3", "4,5,6"}
	expectedPrefix := "--mon-core="
	expectedSubstring := "all:[1,2,3];mbt:[1,2,3];"
	expectedSubstring2 := "all:[4,5,6];mbt:[4,5,6];"
	result = createArgCores(cores)
	assert.Contains(t, result, expectedPrefix)
	assert.Contains(t, result, expectedSubstring)
	assert.Contains(t, result, expectedSubstring2)
}

func TestParseCoresConfig(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		var configCores []string
		result, err := parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty string in slice", func(t *testing.T) {
		configCores := []string{""}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("not correct string", func(t *testing.T) {
		configCores := []string{"wrong string"}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("not correct string", func(t *testing.T) {
		configCores := []string{"1,2", "wasd:#$!;"}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("not correct string", func(t *testing.T) {
		configCores := []string{"1,2,2"}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("coma separated cores - positive", func(t *testing.T) {
		configCores := []string{"0,1,2,3,4,5"}
		expected := []string{"0,1,2,3,4,5"}
		result, err := parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)

		configCores = []string{"0,1,2", "3,4,5"}
		expected = []string{"0,1,2", "3,4,5"}
		result, err = parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)

		configCores = []string{"0,4,1", "2,3,5", "9"}
		expected = []string{"0,4,1", "2,3,5", "9"}
		result, err = parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)
	})

	t.Run("coma separated cores - negative", func(t *testing.T) {
		// cannot monitor same cores in different groups
		configCores := []string{"0,1,2", "2"}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		configCores = []string{"0,1,2", "2,3,4"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		configCores = []string{"0,-1,2", "2,3,4"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("dash separated cores - positive", func(t *testing.T) {
		configCores := []string{"0-5"}
		expected := []string{"0,1,2,3,4,5"}
		result, err := parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)

		configCores = []string{"0-5", "7-10"}
		expected = []string{"0,1,2,3,4,5", "7,8,9,10"}
		result, err = parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)

		configCores = []string{"5-5"}
		expected = []string{"5"}
		result, err = parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)
	})

	t.Run("dash separated cores - negative", func(t *testing.T) {
		// cannot monitor same cores in different groups
		configCores := []string{"0-5", "2-7"}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		// more than two values in range
		configCores = []string{"0-5-10"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		// first value cannot be higher than second
		configCores = []string{"12-5"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		configCores = []string{"0-"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("mixed separator - positive", func(t *testing.T) {
		configCores := []string{"0-5,6,7"}
		expected := []string{"0,1,2,3,4,5,6,7"}
		result, err := parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)

		configCores = []string{"0-5,6,7", "8,9,10"}
		expected = []string{"0,1,2,3,4,5,6,7", "8,9,10"}
		result, err = parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)

		configCores = []string{"0-7", "8-10"}
		expected = []string{"0,1,2,3,4,5,6,7", "8,9,10"}
		result, err = parseCoresConfig(configCores)
		assert.Nil(t, err)
		assert.EqualValues(t, expected, result)
	})

	t.Run("mixed separator - negative", func(t *testing.T) {
		// cannot monitor same cores in different groups
		configCores := []string{"0-5,", "2-7"}
		result, err := parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		// cores cannot be duplicated
		configCores = []string{"0-5,5"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)

		// more than two values in range
		configCores = []string{"0-5-6,9"}
		result, err = parseCoresConfig(configCores)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})
}
