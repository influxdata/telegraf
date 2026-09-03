//go:build linux && (386 || amd64 || arm || arm64)

package ras

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestUpdateCounters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ras-mc_event.db")
	require.NoError(t, os.WriteFile(path, nil, 0600))

	ras := &Ras{
		DBPath: path,
		Log:    &testutil.Logger{},
	}
	require.NoError(t, ras.Init())

	testData := []machineCheckError{
		{
			timestamp:    "2020-05-20 07:34:53 +0200",
			socketID:     0,
			errorMsg:     "MEMORY CONTROLLER RD_CHANNEL0_ERR Transaction: Memory read error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 07:35:11 +0200",
			socketID:     0,
			errorMsg:     "MEMORY CONTROLLER RD_CHANNEL0_ERR Transaction: Memory read error",
			mciStatusMsg: "Uncorrected_error",
		},
		{
			timestamp:    "2020-05-20 07:37:50 +0200",
			socketID:     0,
			errorMsg:     "MEMORY CONTROLLER RD_CHANNEL2_ERR Transaction: Memory write error",
			mciStatusMsg: "Uncorrected_error",
		},
		{
			timestamp:    "2020-05-20 08:14:51 +0200",
			socketID:     0,
			errorMsg:     "MEMORY CONTROLLER WR_CHANNEL2_ERR Transaction: Memory write error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:15:31 +0200",
			socketID:     0,
			errorMsg:     "corrected filtering (some unreported errors in same region) Instruction CACHE Level-0 Read Error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:16:32 +0200",
			socketID:     0,
			errorMsg:     "Instruction TLB Level-0 Error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:16:56 +0200",
			socketID:     0,
			errorMsg:     "No Error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:17:24 +0200",
			socketID:     0,
			errorMsg:     "Unclassified",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:17:41 +0200",
			socketID:     0,
			errorMsg:     "Microcode ROM parity error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:17:48 +0200",
			socketID:     0,
			errorMsg:     "FRC error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:18:18 +0200",
			socketID:     0,
			errorMsg:     "Internal parity error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:18:34 +0200",
			socketID:     0,
			errorMsg:     "SMM Handler Code Access Violation",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:18:54 +0200",
			socketID:     0,
			errorMsg:     "Internal Timer error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:21:23 +0200",
			socketID:     0,
			errorMsg:     "BUS Level-3 Generic Generic IO Request-did-not-timeout Error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:23:23 +0200",
			socketID:     0,
			errorMsg:     "External error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:25:31 +0200",
			socketID:     0,
			errorMsg:     "UPI: COR LL Rx detected CRC error - successful LLR without Phy Reinit",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:25:55 +0200",
			socketID:     0,
			errorMsg:     "Instruction CACHE Level-2 Generic Error",
			mciStatusMsg: "Error_overflow Corrected_error",
		},
		{
			timestamp:    "2020-05-20 08:25:56 +0200",
			socketID:     0,
			errorMsg:     "Corrected error, no action required.",
			mciStatusMsg: "Error_overflow CECC",
			mcaStatusMsg: "DRAM ECC error. Ext Err Code: 0 Memory Error 'mem-tx: generic read, tx: generic, level: L3/generic'",
		},
		{
			timestamp:    "2020-05-20 08:25:58 +0200",
			socketID:     0,
			errorMsg:     "Uncorrected error.",
			mciStatusMsg: "Error_overflow CECC",
			mcaStatusMsg: "DRAM ECC error. Ext Err Code: 0 Memory Error 'mem-tx: generic read, tx: generic, level: L3/generic'",
		},
	}
	for i := range testData {
		ras.updateCounters(&testData[i])
	}

	require.Len(t, ras.cpuSocketCounters, 1, "Should contain counters only for single socket")

	for metric, value := range ras.cpuSocketCounters[0] {
		if metric == "processor_base_errors" {
			// processor_base_errors is sum of other seven errors: internal_timer_errors, smm_handler_code_access_violation_errors,
			// internal_parity_errors, frc_errors, external_mce_errors, microcode_rom_parity_errors and unclassified_mce_errors
			require.Equal(t, int64(7), value, "processor_base_errors should have value of 7")
		} else {
			require.Equal(t, int64(1), value, metric+" should have value of 1")
		}
	}

	for metric, value := range ras.serverCounters {
		require.Equal(t, int64(1), value, metric+" should have value of 1")
	}
}

func TestUpdateLatestTimestamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ras-mc_event.db")
	require.NoError(t, os.WriteFile(path, nil, 0600))

	ras := &Ras{
		DBPath: path,
		Log:    &testutil.Logger{},
	}
	require.NoError(t, ras.Init())

	timestamps := []string{
		"2020-05-20 07:34:53 +0200",
		"2020-05-20 07:35:11 +0200",
		"2020-05-20 07:37:50 +0200",
		"2020-05-20 08:14:51 +0200",
		"2020-05-20 08:15:31 +0200",
		"2020-05-20 08:16:32 +0200",
		"2020-05-20 08:16:56 +0200",
		"2020-05-20 08:17:24 +0200",
		"2020-05-20 08:17:41 +0200",
		"2020-05-20 08:17:48 +0200",
		"2020-05-20 08:18:18 +0200",
		"2020-05-20 08:18:34 +0200",
		"2020-05-20 08:18:54 +0200",
		"2020-05-20 08:21:23 +0200",
		"2020-05-20 08:23:23 +0200",
		"2020-05-20 08:25:31 +0200",
		"2020-05-20 08:25:55 +0200",
		"2019-05-20 08:25:55 +0200",
		"2018-02-21 12:27:22 +0200",
		"2020-08-01 15:13:27 +0200",
	}
	for _, timestamp := range timestamps {
		err := ras.updateLatestTimestamp(timestamp)
		require.NoError(t, err)
	}
	require.Equal(t, int64(1596287607000000000), ras.latestTimestamp.UnixNano())
}

func TestMultipleSockets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ras-mc_event.db")
	require.NoError(t, os.WriteFile(path, nil, 0600))

	ras := &Ras{
		DBPath: path,
		Log:    &testutil.Logger{},
	}
	require.NoError(t, ras.Init())

	cacheL2 := "Instruction CACHE Level-2 Generic Error"
	overflow := "Error_overflow Corrected_error"
	testData := []machineCheckError{
		{
			timestamp:    "2019-05-20 08:25:55 +0200",
			socketID:     0,
			errorMsg:     cacheL2,
			mciStatusMsg: overflow,
		},
		{
			timestamp:    "2018-02-21 12:27:22 +0200",
			socketID:     1,
			errorMsg:     cacheL2,
			mciStatusMsg: overflow,
		},
		{
			timestamp:    "2020-03-21 14:17:28 +0200",
			socketID:     2,
			errorMsg:     cacheL2,
			mciStatusMsg: overflow,
		},
		{
			timestamp:    "2020-03-21 17:24:18 +0200",
			socketID:     3,
			errorMsg:     cacheL2,
			mciStatusMsg: overflow,
		},
	}
	for i := range testData {
		ras.updateCounters(&testData[i])
	}
	require.Len(t, ras.cpuSocketCounters, 4, "Should contain counters for four sockets")

	for _, metricData := range ras.cpuSocketCounters {
		for metric, value := range metricData {
			if metric == "cache_l2_errors" {
				require.Equal(t, int64(1), value, "cache_l2_errors should have value of 1")
			} else {
				require.Equal(t, int64(0), value, metric+" should have value of 0")
			}
		}
	}
}

func TestMissingDatabase(t *testing.T) {
	ras := &Ras{
		DBPath: "/nonexistent/ras.db",
		Log:    &testutil.Logger{},
	}
	require.ErrorContains(t, ras.Init(), "does not exist")
}

func TestEmptyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ras-mc_event.db")
	require.NoError(t, os.WriteFile(path, nil, 0600))

	ras := &Ras{
		DBPath: path,
		Log:    &testutil.Logger{},
	}
	require.NoError(t, ras.Init())

	require.Len(t, ras.cpuSocketCounters, 1, "Should contain default counters for one socket")
	require.Len(t, ras.serverCounters, 2, "Should contain default counters for server")

	for metric, value := range ras.cpuSocketCounters[0] {
		require.Equal(t, int64(0), value, metric+" should have value of 0")
	}

	for metric, value := range ras.serverCounters {
		require.Equal(t, int64(0), value, metric+" should have value of 0")
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("ras", func() telegraf.Input { return &Ras{} })

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	// Set the testing options
	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.IgnoreType(),
		testutil.SortMetrics(),
	}

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the expected output
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Configure and initialize the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			plugin := cfg.Inputs[0].Input.(*Ras)
			plugin.DBPath = filepath.Join(testcasePath, "ras-mc_event.db")
			require.NoError(t, plugin.Init())

			// Start the plugin
			require.NoError(t, plugin.Start(nil))
			defer plugin.Stop()

			// Collect the metrics
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			// Check the result
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}
