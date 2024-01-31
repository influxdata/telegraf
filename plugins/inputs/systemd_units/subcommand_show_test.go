package systemd_units

import (
	"fmt"
	"testing"
)

func TestSubcommandShow(t *testing.T) {
	tests := []TestDef{
		{
			Name: "example loaded active running",
			Lines: []string{
				"Id=example.service",
				"LoadState=loaded",
				"ActiveState=active",
				"SubState=running",
				"UnitFileState=enabled",
				"UnitFilePreset=disabled",
				"StatusErrno=0",
				"NRestarts=1",
				"MemoryCurrent=1000",
				"MemoryPeak=2000",
				"MemorySwapCurrent=3000",
				"MemorySwapPeak=4000",
				"MemoryAvailable=5000",
				"MainPID=9999",
			},
			Tags: map[string]string{
				"name":      "example.service",
				"load":      "loaded",
				"active":    "active",
				"sub":       "running",
				"uf_state":  "enabled",
				"uf_preset": "disabled",
			},
			Fields: map[string]interface{}{
				"load_code":    0,
				"active_code":  0,
				"sub_code":     0,
				"status_errno": 0,
				"restarts":     1,
				"mem_current":  1000,
				"mem_peak":     2000,
				"swap_current": 3000,
				"swap_peak":    4000,
				"mem_avail":    5000,
				"pid":          9999,
			},
		},
		{
			Name: "example loaded active exited",
			Lines: []string{
				"Id=example.service",
				"LoadState=loaded",
				"ActiveState=active",
				"SubState=exited",
				"UnitFileState=enabled",
				"UnitFilePreset=disabled",
				"StatusErrno=0",
				"NRestarts=0",
			},
			Tags: map[string]string{
				"name":      "example.service",
				"load":      "loaded",
				"active":    "active",
				"sub":       "exited",
				"uf_state":  "enabled",
				"uf_preset": "disabled",
			},
			Fields: map[string]interface{}{
				"load_code":    0,
				"active_code":  0,
				"sub_code":     4,
				"status_errno": 0,
				"restarts":     0,
			},
		},
		{
			Name: "example loaded failed failed",
			Lines: []string{
				"Id=example.service",
				"LoadState=loaded",
				"ActiveState=failed",
				"SubState=failed",
				"UnitFileState=enabled",
				"UnitFilePreset=disabled",
				"StatusErrno=10",
				"NRestarts=1",
				"MemoryCurrent=1000",
				"MemoryPeak=2000",
				"MemorySwapCurrent=3000",
				"MemorySwapPeak=4000",
				"MemoryAvailable=5000",
			},
			Tags: map[string]string{
				"name":      "example.service",
				"load":      "loaded",
				"active":    "failed",
				"sub":       "failed",
				"uf_state":  "enabled",
				"uf_preset": "disabled",
			},
			Fields: map[string]interface{}{
				"load_code":    0,
				"active_code":  3,
				"sub_code":     12,
				"status_errno": 10,
				"restarts":     1,
				"mem_current":  1000,
				"mem_peak":     2000,
				"swap_current": 3000,
				"swap_peak":    4000,
				"mem_avail":    5000,
			},
		},
		{
			Name: "example not-found inactive dead",
			Lines: []string{
				"Id=example.service",
				"LoadState=not-found",
				"ActiveState=inactive",
				"SubState=dead",
				"UnitFileState=enabled",
				"UnitFilePreset=disabled",
				"StatusErrno=[not set]",
				"NRestarts=[not set]",
				"MemoryCurrent=[not set]",
				"MemoryPeak=[not set]",
				"MemorySwapCurrent=[not set]",
				"MemorySwapPeak=[not set]",
				"MemoryAvailable=[not set]",
				"MainPID=[not set]",
			},
			Tags: map[string]string{
				"name":      "example.service",
				"load":      "not-found",
				"active":    "inactive",
				"sub":       "dead",
				"uf_state":  "enabled",
				"uf_preset": "disabled",
			},
			Fields: map[string]interface{}{
				"load_code":   2,
				"active_code": 2,
				"sub_code":    1,
			},
		},
		{
			Name: "example unknown unknown unknown",
			Lines: []string{
				"Id=example.service",
				"LoadState=unknown",
				"ActiveState=unknown",
				"SubState=unknown",
				"UnitFileState=unknown",
				"UnitFilePreset=unknown",
			},
			Err: fmt.Errorf("error parsing field '%s', value '%s' not in map", "LoadState", "unknown"),
		},
		{
			Name: "example no key value pair",
			Lines: []string{
				"Id=example.service",
				"LoadState",
				"ActiveState=active",
			},
			Err: fmt.Errorf("error parsing line (expected key=value): %s", "LoadState"),
			Tags: map[string]string{
				"name":   "example.service",
				"active": "active",
			},
			Fields: map[string]interface{}{
				"active_code": 0,
			},
		},
	}

	dut := initSubcommandShow()

	runParserTests(t, tests, dut)
}

func TestCommandlineShow(t *testing.T) {
	propertiesTemplate := []string{
		"--all",
		"--type=service",
		"--property=Id",
		"--property=LoadState",
		"--property=ActiveState",
		"--property=SubState",
		"--property=StatusErrno",
		"--property=UnitFileState",
		"--property=UnitFilePreset",
		"--property=NRestarts",
		"--property=MemoryCurrent",
		"--property=MemoryPeak",
		"--property=MemorySwapCurrent",
		"--property=MemorySwapPeak",
		"--property=MemoryAvailable",
		"--property=MainPID",
	}

	// Test with the default patern
	paramsTemplate := append([]string{
		"show",
		"*",
	}, propertiesTemplate...)

	dut := initSubcommandShow()
	systemdUnits := SystemdUnits{
		UnitType: "service",
	}

	runCommandLineTest(t, paramsTemplate, dut, &systemdUnits)

	// Test using a more komplex pattern
	paramsTemplate = append([]string{
		"show",
		"unita.service",
		"*.timer",
	}, propertiesTemplate...)

	systemdUnits = SystemdUnits{
		UnitType: "service",
		Pattern:  "unita.service *.timer",
	}

	runCommandLineTest(t, paramsTemplate, dut, &systemdUnits)
}
