package systemd_units

import (
	"fmt"
	"testing"
)

func TestSubcommandList(t *testing.T) {
	tests := []TestDef{
		{
			Name: "example loaded active running",
			Line: "example.service                loaded active running example service description",
			Tags: map[string]string{
				"name":   "example.service",
				"load":   "loaded",
				"active": "active",
				"sub":    "running",
			},
			Fields: map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    0,
			},
		},
		{
			Name: "example loaded active exited",
			Line: "example.service                loaded active exited  example service description",
			Tags: map[string]string{
				"name":   "example.service",
				"load":   "loaded",
				"active": "active",
				"sub":    "exited",
			},
			Fields: map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    4,
			},
		},
		{
			Name: "example loaded failed failed",
			Line: "example.service                loaded failed failed  example service description",
			Tags: map[string]string{"name": "example.service", "load": "loaded", "active": "failed", "sub": "failed"},
			Fields: map[string]interface{}{
				"load_code":   0,
				"active_code": 3,
				"sub_code":    12,
			},
		},
		{
			Name: "example not-found inactive dead",
			Line: "example.service                not-found inactive dead  example service description",
			Tags: map[string]string{
				"name":   "example.service",
				"load":   "not-found",
				"active": "inactive",
				"sub":    "dead",
			},
			Fields: map[string]interface{}{
				"load_code":   2,
				"active_code": 2,
				"sub_code":    1,
			},
		},
		{
			Name: "example unknown unknown unknown",
			Line: "example.service                unknown unknown unknown  example service description",
			Err:  fmt.Errorf("parsing field 'load' failed, value not in map: %s", "unknown"),
		},
		{
			Name: "example too few fields",
			Line: "example.service                loaded fai",
			Err:  fmt.Errorf("parsing line failed (expected at least 4 fields): %s", "example.service                loaded fai"),
		},
	}

	dut := initSubcommandListUnits()

	runParserTests(t, tests, dut)
}

func TestCommandlineList(t *testing.T) {
	// Test using the default pattern (no pattern)
	paramsTemplate := []string{
		"list-units",
		"--all",
		"--plain",
		"--no-legend",
		"--type=service",
	}

	dut := initSubcommandListUnits()
	systemdUnits := SystemdUnits{
		UnitType: "service",
	}

	runCommandLineTest(t, paramsTemplate, dut, &systemdUnits)

	// Test using a more complex pattern
	paramsTemplate = []string{
		"list-units",
		"unita.service",
		"*.timer",
		"--all",
		"--plain",
		"--no-legend",
		"--type=service",
	}

	systemdUnits = SystemdUnits{
		UnitType: "service",
		Pattern:  "unita.service *.timer",
	}

	runCommandLineTest(t, paramsTemplate, dut, &systemdUnits)
}
