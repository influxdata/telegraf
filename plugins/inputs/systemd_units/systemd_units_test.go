package systemd_units

import (
	"bytes"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

// Global test definitions and structure.
// Tests are located within `subcommand_list_test.go` and
// `subcommand_show_test.go`.

type TestDef struct {
	Name   string
	Line   string
	Lines  []string
	Tags   map[string]string
	Fields map[string]interface{}
	Status int
	Err    error
}

func runParserTests(t *testing.T, tests []TestDef, dut *subCommandInfo) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			acc := new(testutil.Accumulator)

			var line string
			if len(tt.Lines) > 0 && len(tt.Line) == 0 {
				line = strings.Join(tt.Lines, "\n")
			} else if len(tt.Lines) == 0 && len(tt.Line) > 0 {
				line = tt.Line
			} else {
				t.Error("property Line and Lines set in test definition")
			}

			dut.parseResult(acc, bytes.NewBufferString(line))
			err := acc.FirstError()

			if !reflect.DeepEqual(tt.Err, err) {
				t.Errorf("%s: expected error '%#v' got '%#v'", tt.Name, tt.Err, err)
			}
			if len(acc.Metrics) > 0 {
				m := acc.Metrics[0]
				if !reflect.DeepEqual(m.Measurement, measurement) {
					t.Errorf("%s: expected measurement '%#v' got '%#v'\n", tt.Name, measurement, m.Measurement)
				}
				if !reflect.DeepEqual(m.Tags, tt.Tags) {
					t.Errorf("%s: expected tags\n%#v got\n%#v\n", tt.Name, tt.Tags, m.Tags)
				}
				if !reflect.DeepEqual(m.Fields, tt.Fields) {
					t.Errorf("%s: expected fields\n%#v got\n%#v\n", tt.Name, tt.Fields, m.Fields)
				}
			}
		})
	}
}

func runCommandLineTest(t *testing.T, paramsTemplate []string, dut *subCommandInfo, systemdUnits *SystemdUnits) {
	params := *dut.getParameters(systemdUnits)

	// Because we sort the params and the template array before comparison
	// we have to compare the positional parameters first.
	for i, v := range paramsTemplate {
		if strings.HasPrefix(v, "--") {
			break
		}

		if v != params[i] {
			t.Errorf("Positional parameter %d is '%s'. Expected '%s'.", i, params[i], v)
		}
	}
	// Because the maps do not lead to a stable order of the "--property"
	// arguments sort all the command line arguments and compare them.
	sort.Strings(params)
	sort.Strings(paramsTemplate)
	if !reflect.DeepEqual(params, paramsTemplate) {
		t.Errorf("Generated list of command line arguments '%#v' do not match expected list command line arguments '%#v'", params, paramsTemplate)
	}
}
