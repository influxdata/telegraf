// +build windows

package wpc

import (
	"errors"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWPCConfigGet1(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(_Total)\\%% Processor Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigGet1", Template: templates}

	err := m.parseConfig()
	require.NoError(t, err)
}

func TestWPCConfigGet2(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(_Total)\\%% Processor Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigGet2", Template: templates}

	err := m.parseConfig()
	require.NoError(t, err)

	require.Equal(t, 1, len(gTaskList), "Wrong number of tasks defined.")
	require.Equal(t, 1, len(gCounterCache), "Wrong number of counters opened.")
	require.Equal(t, 1, len(gTaskList[0].fields), "Wrong number of field mappings defined.")
	require.Equal(t, "test", gTaskList[0].measurement, "Wrong measurement saved.")
}

func TestWPCConfigGet3(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(_Total)\\%% Processor Time"},
		[]string{"bar", "\\Processor Information(_Total)\\%% Processor Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigGet3", Template: templates}

	err := m.parseConfig()
	require.NoError(t, err)

	require.Equal(t, 1, len(gTaskList), "Wrong number of tasks defined.")
	require.Equal(t, 1, len(gCounterCache), "Wrong number of counters opened.")
	require.Equal(t, 2, len(gTaskList[0].fields), "Wrong number of field mappings defined.")
	require.Equal(t, "test", gTaskList[0].measurement, "Wrong measurement saved.")
}

func TestWPCConfigGet4(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(_Total)\\%% Processor Time"},
		[]string{"bar", "\\System\\Context Switches/sec"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigGet4", Template: templates}

	err := m.parseConfig()
	require.NoError(t, err)

	require.Equal(t, 1, len(gTaskList), "Wrong number of tasks defined.")
	require.Equal(t, 2, len(gCounterCache), "Wrong number of counters opened.")
	require.Equal(t, 2, len(gTaskList[0].fields), "Wrong number of field mappings defined.")
	require.Equal(t, "test", gTaskList[0].measurement, "Wrong measurement saved.")
}

func TestWPCConfigGet5(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(*)\\%% Processor Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigGet5", Template: templates}

	err := m.parseConfig()
	require.NoError(t, err)

	require.Equal(t, 1, len(gTaskList), "Wrong number of tasks defined.")
	require.Equal(t, 1, len(gCounterCache), "Wrong number of counters opened.")
	require.Equal(t, 1, len(gTaskList[0].fields), "Wrong number of field mappings defined.")
	require.Equal(t, "test", gTaskList[0].measurement, "Wrong measurement saved.")
}

func TestWPCConfigGet6(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(*)\\%% Processor TimeERROR"},
		[]string{"bar", "\\Processor Information(*)\\%% Idle Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = false

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigGet6", Template: templates}

	err := m.parseConfig()
	require.NoError(t, err)

	require.Equal(t, 1, len(gTaskList), "Wrong number of tasks defined.")
	require.Equal(t, 1, len(gCounterCache), "Wrong number of counters opened.")
	require.Equal(t, 1, len(gTaskList[0].fields), "Wrong number of field mappings defined.")
	require.Equal(t, "test", gTaskList[0].measurement, "Wrong measurement saved.")
}

func TestWPCConfigError1(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor InformationERROR(*)\\%% Processor Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = false

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigError1", Template: templates}

	err := m.parseConfig()
	require.Error(t, err)
}

func TestWPCConfigError2(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor(SuperERROR)\\%% C1 Time"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = false

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigError2", Template: templates}

	err := m.parseConfig()
	require.Error(t, err)
}

func TestWPCConfigError3(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"foo", "\\Processor Information(*)\\%% Processor TimeERROR"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "ConfigError3", Template: templates}

	err := m.parseConfig()
	require.Error(t, err)
}

func TestWPCCollect1(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"park", "\\Processor Information(_Total)\\Parking Status"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "Collect1", Template: templates}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)

	tags := map[string]string{
		"instance": "_Total",
	}
	fields := map[string]interface{}{
		"park": float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
}

func TestWPCCollect2(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"park", "\\Processor Information(_Total)\\Parking Status"},
		[]string{"plim", "\\Processor Information(_Total)\\Performance Limit Flags"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "Collect2", Template: templates}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)

	tags := map[string]string{
		"instance": "_Total",
	}
	fields := map[string]interface{}{
		"park": float32(0),
		"plim": float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
}

func TestWPCCollect3(t *testing.T) {
	var templates = make([]template, 1)

	counters := [][]string{
		[]string{"park", "\\Processor Information(_Total)\\Parking Status"},
		[]string{"sys_cs_rate", "\\System\\Context Switches/sec"},
	}

	var measurement string = "test"
	var warnonmissing bool = false
	var failonmissing bool = true

	tmpl := template{
		Counters:      counters,
		Measurement:   measurement,
		WarnOnMissing: warnonmissing,
		FailOnMissing: failonmissing,
	}

	templates[0] = tmpl

	m := WindowsPerformanceCounter{PrintValid: false, TestName: "Collect3", Template: templates}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	time.Sleep(2000 * time.Millisecond)
	err = m.Gather(&acc)

	tags := map[string]string{
		"instance": "_Total",
	}
	fields := map[string]interface{}{
		"park": float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)

	tags = map[string]string{
		"instance": "",
	}
	fields = map[string]interface{}{
		"sys_cs_rate": float32(0),
	}
	acc.AssertContainsTaggedFields(t, measurement, fields, tags)
}
