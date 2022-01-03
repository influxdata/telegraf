//go:build !windows
// +build !windows

package intel_rdt

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var metricsValues = map[string]float64{
	"IPC":        0.5,
	"LLC_Misses": 61650,
	"LLC":        1632,
	"MBL":        0.6,
	"MBR":        0.9,
	"MBT":        1.9,
}

func TestParseCoresMeasurement(t *testing.T) {
	timestamp := "2020-08-12 13:34:36"
	cores := "\"37,44\""

	t.Run("valid measurement string", func(t *testing.T) {
		measurement := fmt.Sprintf("%s,%s,%f,%f,%f,%f,%f,%f",
			timestamp,
			cores,
			metricsValues["IPC"],
			metricsValues["LLC_Misses"],
			metricsValues["LLC"],
			metricsValues["MBL"],
			metricsValues["MBR"],
			metricsValues["MBT"])

		expectedCores := "37,44"
		expectedTimestamp := time.Date(2020, 8, 12, 13, 34, 36, 0, time.Local)

		result, err := parseCoresMeasurement(measurement)

		assert.Nil(t, err)
		assert.Equal(t, expectedCores, result.cores)
		assert.Equal(t, expectedTimestamp, result.time)
		assert.Equal(t, result.values[0], metricsValues["IPC"])
		assert.Equal(t, result.values[1], metricsValues["LLC_Misses"])
		assert.Equal(t, result.values[2], metricsValues["LLC"])
		assert.Equal(t, result.values[3], metricsValues["MBL"])
		assert.Equal(t, result.values[4], metricsValues["MBR"])
		assert.Equal(t, result.values[5], metricsValues["MBT"])
	})
	t.Run("not valid measurement string", func(t *testing.T) {
		measurement := "not, valid, measurement"

		result, err := parseCoresMeasurement(measurement)

		assert.NotNil(t, err)
		assert.Equal(t, "", result.cores)
		assert.Nil(t, result.values)
		assert.Equal(t, time.Time{}, result.time)
	})
	t.Run("not valid values string", func(t *testing.T) {
		measurement := fmt.Sprintf("%s,%s,%s,%s,%f,%f,%f,%f",
			timestamp,
			cores,
			"%d",
			"in",
			metricsValues["LLC"],
			metricsValues["MBL"],
			metricsValues["MBR"],
			metricsValues["MBT"])

		result, err := parseCoresMeasurement(measurement)

		assert.NotNil(t, err)
		assert.Equal(t, "", result.cores)
		assert.Nil(t, result.values)
		assert.Equal(t, time.Time{}, result.time)
	})
	t.Run("not valid timestamp format", func(t *testing.T) {
		invalidTimestamp := "2020-08-12-21 13:34:"
		measurement := fmt.Sprintf("%s,%s,%f,%f,%f,%f,%f,%f",
			invalidTimestamp,
			cores,
			metricsValues["IPC"],
			metricsValues["LLC_Misses"],
			metricsValues["LLC"],
			metricsValues["MBL"],
			metricsValues["MBR"],
			metricsValues["MBT"])

		result, err := parseCoresMeasurement(measurement)

		assert.NotNil(t, err)
		assert.Equal(t, "", result.cores)
		assert.Nil(t, result.values)
		assert.Equal(t, time.Time{}, result.time)
	})
}

func TestParseProcessesMeasurement(t *testing.T) {
	timestamp := "2020-08-12 13:34:36"
	cores := "\"37,44\""
	pids := "\"12345,9999\""
	processName := "process_name"

	t.Run("valid measurement string", func(t *testing.T) {
		measurement := fmt.Sprintf("%s,%s,%s,%f,%f,%f,%f,%f,%f",
			timestamp,
			pids,
			cores,
			metricsValues["IPC"],
			metricsValues["LLC_Misses"],
			metricsValues["LLC"],
			metricsValues["MBL"],
			metricsValues["MBR"],
			metricsValues["MBT"])

		expectedCores := "37,44"
		expectedTimestamp := time.Date(2020, 8, 12, 13, 34, 36, 0, time.Local)

		newMeasurement := processMeasurement{
			name:        processName,
			measurement: measurement,
		}
		result, err := parseProcessesMeasurement(newMeasurement)

		assert.Nil(t, err)
		assert.Equal(t, processName, result.process)
		assert.Equal(t, expectedCores, result.cores)
		assert.Equal(t, expectedTimestamp, result.time)
		assert.Equal(t, result.values[0], metricsValues["IPC"])
		assert.Equal(t, result.values[1], metricsValues["LLC_Misses"])
		assert.Equal(t, result.values[2], metricsValues["LLC"])
		assert.Equal(t, result.values[3], metricsValues["MBL"])
		assert.Equal(t, result.values[4], metricsValues["MBR"])
		assert.Equal(t, result.values[5], metricsValues["MBT"])
	})

	invalidTimestamp := "2020-20-20-31"
	negativeTests := []struct {
		name        string
		measurement string
	}{{
		name:        "not valid measurement string",
		measurement: "invalid,measurement,format",
	}, {
		name: "not valid timestamp format",
		measurement: fmt.Sprintf("%s,%s,%s,%f,%f,%f,%f,%f,%f",
			invalidTimestamp,
			pids,
			cores,
			metricsValues["IPC"],
			metricsValues["LLC_Misses"],
			metricsValues["LLC"],
			metricsValues["MBL"],
			metricsValues["MBR"],
			metricsValues["MBT"]),
	},
		{
			name: "not valid values string",
			measurement: fmt.Sprintf("%s,%s,%s,%s,%s,%f,%f,%f,%f",
				timestamp,
				pids,
				cores,
				"1##",
				"da",
				metricsValues["LLC"],
				metricsValues["MBL"],
				metricsValues["MBR"],
				metricsValues["MBT"]),
		},
		{
			name:        "not valid csv line with quotes",
			measurement: "0000-08-02 0:00:00,,\",,,,,,,,,,,,,,,,,,,,,,,,\",,",
		},
	}

	for _, test := range negativeTests {
		t.Run(test.name, func(t *testing.T) {
			newMeasurement := processMeasurement{
				name:        processName,
				measurement: test.measurement,
			}
			result, err := parseProcessesMeasurement(newMeasurement)

			assert.NotNil(t, err)
			assert.Equal(t, "", result.process)
			assert.Equal(t, "", result.cores)
			assert.Nil(t, result.values)
			assert.Equal(t, time.Time{}, result.time)
		})
	}
}

func TestAddToAccumulatorCores(t *testing.T) {
	t.Run("shortened false", func(t *testing.T) {
		var acc testutil.Accumulator
		publisher := Publisher{acc: &acc}

		cores := "1,2,3"
		metricsValues := []float64{1, 2, 3, 4, 5, 6}
		timestamp := time.Date(2020, 8, 12, 13, 34, 36, 0, time.Local)

		publisher.addToAccumulatorCores(parsedCoresMeasurement{cores, metricsValues, timestamp})

		for _, test := range testCoreMetrics {
			acc.AssertContainsTaggedFields(t, "rdt_metric", test.fields, test.tags)
		}
	})
	t.Run("shortened true", func(t *testing.T) {
		var acc testutil.Accumulator
		publisher := Publisher{acc: &acc, shortenedMetrics: true}

		cores := "1,2,3"
		metricsValues := []float64{1, 2, 3, 4, 5, 6}
		timestamp := time.Date(2020, 8, 12, 13, 34, 36, 0, time.Local)

		publisher.addToAccumulatorCores(parsedCoresMeasurement{cores, metricsValues, timestamp})

		for _, test := range testCoreMetricsShortened {
			acc.AssertDoesNotContainsTaggedFields(t, "rdt_metric", test.fields, test.tags)
		}
	})
}

func TestAddToAccumulatorProcesses(t *testing.T) {
	t.Run("shortened false", func(t *testing.T) {
		var acc testutil.Accumulator
		publisher := Publisher{acc: &acc}

		process := "process_name"
		cores := "1,2,3"
		metricsValues := []float64{1, 2, 3, 4, 5, 6}
		timestamp := time.Date(2020, 8, 12, 13, 34, 36, 0, time.Local)

		publisher.addToAccumulatorProcesses(parsedProcessMeasurement{process, cores, metricsValues, timestamp})

		for _, test := range testCoreProcesses {
			acc.AssertContainsTaggedFields(t, "rdt_metric", test.fields, test.tags)
		}
	})
	t.Run("shortened true", func(t *testing.T) {
		var acc testutil.Accumulator
		publisher := Publisher{acc: &acc, shortenedMetrics: true}

		process := "process_name"
		cores := "1,2,3"
		metricsValues := []float64{1, 2, 3, 4, 5, 6}
		timestamp := time.Date(2020, 8, 12, 13, 34, 36, 0, time.Local)

		publisher.addToAccumulatorProcesses(parsedProcessMeasurement{process, cores, metricsValues, timestamp})

		for _, test := range testCoreProcessesShortened {
			acc.AssertDoesNotContainsTaggedFields(t, "rdt_metric", test.fields, test.tags)
		}
	})
}

var (
	testCoreMetrics = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value": float64(1),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "IPC",
			},
		},
		{
			map[string]interface{}{
				"value": float64(2),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "LLC_Misses",
			},
		},
		{
			map[string]interface{}{
				"value": float64(3),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "LLC",
			},
		},
		{
			map[string]interface{}{
				"value": float64(4),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "MBL",
			},
		},
		{
			map[string]interface{}{
				"value": float64(5),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "MBR",
			},
		},
		{
			map[string]interface{}{
				"value": float64(6),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "MBT",
			},
		},
	}
	testCoreMetricsShortened = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value": float64(1),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "IPC",
			},
		},
		{
			map[string]interface{}{
				"value": float64(2),
			},
			map[string]string{
				"cores": "1,2,3",
				"name":  "LLC_Misses",
			},
		},
	}
	testCoreProcesses = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value": float64(1),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "IPC",
				"process": "process_name",
			},
		},
		{
			map[string]interface{}{
				"value": float64(2),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "LLC_Misses",
				"process": "process_name",
			},
		},
		{
			map[string]interface{}{
				"value": float64(3),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "LLC",
				"process": "process_name",
			},
		},
		{
			map[string]interface{}{
				"value": float64(4),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "MBL",
				"process": "process_name",
			},
		},
		{
			map[string]interface{}{
				"value": float64(5),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "MBR",
				"process": "process_name",
			},
		},
		{
			map[string]interface{}{
				"value": float64(6),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "MBT",
				"process": "process_name",
			},
		},
	}
	testCoreProcessesShortened = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"value": float64(1),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "IPC",
				"process": "process_name",
			},
		},
		{
			map[string]interface{}{
				"value": float64(2),
			},
			map[string]string{
				"cores":   "1,2,3",
				"name":    "LLC_Misses",
				"process": "process_name",
			},
		},
	}
)
