//go:build linux && (386 || amd64 || arm || arm64)
// +build linux
// +build 386 amd64 arm arm64

package ras

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestUpdateCounters(t *testing.T) {
	ras := newRas()
	for _, mce := range testData {
		ras.updateCounters(&mce)
	}

	require.Equal(t, 1, len(ras.cpuSocketCounters), "Should contain counters only for single socket")

	for metric, value := range ras.cpuSocketCounters[0] {
		if metric == processorBase {
			// processor_base_errors is sum of other seven errors: internal_timer_errors, smm_handler_code_access_violation_errors,
			// internal_parity_errors, frc_errors, external_mce_errors, microcode_rom_parity_errors and unclassified_mce_errors
			require.Equal(t, int64(7), value, fmt.Sprintf("%s should have value of 7", processorBase))
		} else {
			require.Equal(t, int64(1), value, fmt.Sprintf("%s should have value of 1", metric))
		}
	}

	for metric, value := range ras.serverCounters {
		require.Equal(t, int64(1), value, fmt.Sprintf("%s should have value of 1", metric))
	}
}

func TestUpdateLatestTimestamp(t *testing.T) {
	ras := newRas()
	ts := "2020-08-01 15:13:27 +0200"
	testData = append(testData, []machineCheckError{
		{
			Timestamp:    "2019-05-20 08:25:55 +0200",
			SocketID:     0,
			ErrorMsg:     "",
			MciStatusMsg: "",
		},
		{
			Timestamp:    "2018-02-21 12:27:22 +0200",
			SocketID:     0,
			ErrorMsg:     "",
			MciStatusMsg: "",
		},
		{
			Timestamp:    ts,
			SocketID:     0,
			ErrorMsg:     "",
			MciStatusMsg: "",
		},
	}...)
	for _, mce := range testData {
		err := ras.updateLatestTimestamp(mce.Timestamp)
		require.NoError(t, err)
	}
	require.Equal(t, ts, ras.latestTimestamp.Format(dateLayout))
}

func TestMultipleSockets(t *testing.T) {
	ras := newRas()
	cacheL2 := "Instruction CACHE Level-2 Generic Error"
	overflow := "Error_overflow Corrected_error"
	testData = []machineCheckError{
		{
			Timestamp:    "2019-05-20 08:25:55 +0200",
			SocketID:     0,
			ErrorMsg:     cacheL2,
			MciStatusMsg: overflow,
		},
		{
			Timestamp:    "2018-02-21 12:27:22 +0200",
			SocketID:     1,
			ErrorMsg:     cacheL2,
			MciStatusMsg: overflow,
		},
		{
			Timestamp:    "2020-03-21 14:17:28 +0200",
			SocketID:     2,
			ErrorMsg:     cacheL2,
			MciStatusMsg: overflow,
		},
		{
			Timestamp:    "2020-03-21 17:24:18 +0200",
			SocketID:     3,
			ErrorMsg:     cacheL2,
			MciStatusMsg: overflow,
		},
	}
	for _, mce := range testData {
		ras.updateCounters(&mce)
	}
	require.Equal(t, 4, len(ras.cpuSocketCounters), "Should contain counters for four sockets")

	for _, metricData := range ras.cpuSocketCounters {
		for metric, value := range metricData {
			if metric == levelTwoCache {
				require.Equal(t, int64(1), value, fmt.Sprintf("%s should have value of 1", levelTwoCache))
			} else {
				require.Equal(t, int64(0), value, fmt.Sprintf("%s should have value of 0", metric))
			}
		}
	}
}

func TestMissingDatabase(t *testing.T) {
	var acc testutil.Accumulator
	ras := newRas()
	ras.DBPath = "/nonexistent/ras.db"
	err := ras.Start(&acc)
	require.Error(t, err)
}

func TestEmptyDatabase(t *testing.T) {
	ras := newRas()

	require.Equal(t, 1, len(ras.cpuSocketCounters), "Should contain default counters for one socket")
	require.Equal(t, 2, len(ras.serverCounters), "Should contain default counters for server")

	for metric, value := range ras.cpuSocketCounters[0] {
		require.Equal(t, int64(0), value, fmt.Sprintf("%s should have value of 0", metric))
	}

	for metric, value := range ras.serverCounters {
		require.Equal(t, int64(0), value, fmt.Sprintf("%s should have value of 0", metric))
	}
}

func newRas() *Ras {
	defaultTimestamp, _ := parseDate("1970-01-01 00:00:01 -0700")
	return &Ras{
		DBPath:          defaultDbPath,
		latestTimestamp: defaultTimestamp,
		cpuSocketCounters: map[int]metricCounters{
			0: *newMetricCounters(),
		},
		serverCounters: map[string]int64{
			levelTwoCache: 0,
			upi:           0,
		},
	}
}

var testData = []machineCheckError{
	{
		Timestamp:    "2020-05-20 07:34:53 +0200",
		SocketID:     0,
		ErrorMsg:     "MEMORY CONTROLLER RD_CHANNEL0_ERR Transaction: Memory read error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 07:35:11 +0200",
		SocketID:     0,
		ErrorMsg:     "MEMORY CONTROLLER RD_CHANNEL0_ERR Transaction: Memory read error",
		MciStatusMsg: "Uncorrected_error",
	},
	{
		Timestamp:    "2020-05-20 07:37:50 +0200",
		SocketID:     0,
		ErrorMsg:     "MEMORY CONTROLLER RD_CHANNEL2_ERR Transaction: Memory write error",
		MciStatusMsg: "Uncorrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:14:51 +0200",
		SocketID:     0,
		ErrorMsg:     "MEMORY CONTROLLER WR_CHANNEL2_ERR Transaction: Memory write error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:15:31 +0200",
		SocketID:     0,
		ErrorMsg:     "corrected filtering (some unreported errors in same region) Instruction CACHE Level-0 Read Error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:16:32 +0200",
		SocketID:     0,
		ErrorMsg:     "Instruction TLB Level-0 Error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:16:56 +0200",
		SocketID:     0,
		ErrorMsg:     "No Error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:17:24 +0200",
		SocketID:     0,
		ErrorMsg:     "Unclassified",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:17:41 +0200",
		SocketID:     0,
		ErrorMsg:     "Microcode ROM parity error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:17:48 +0200",
		SocketID:     0,
		ErrorMsg:     "FRC error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:18:18 +0200",
		SocketID:     0,
		ErrorMsg:     "Internal parity error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:18:34 +0200",
		SocketID:     0,
		ErrorMsg:     "SMM Handler Code Access Violation",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:18:54 +0200",
		SocketID:     0,
		ErrorMsg:     "Internal Timer error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:21:23 +0200",
		SocketID:     0,
		ErrorMsg:     "BUS Level-3 Generic Generic IO Request-did-not-timeout Error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:23:23 +0200",
		SocketID:     0,
		ErrorMsg:     "External error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:25:31 +0200",
		SocketID:     0,
		ErrorMsg:     "UPI: COR LL Rx detected CRC error - successful LLR without Phy Reinit",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
	{
		Timestamp:    "2020-05-20 08:25:55 +0200",
		SocketID:     0,
		ErrorMsg:     "Instruction CACHE Level-2 Generic Error",
		MciStatusMsg: "Error_overflow Corrected_error",
	},
}
