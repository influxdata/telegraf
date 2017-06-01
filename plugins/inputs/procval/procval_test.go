// +build !windows

package procval

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

// Test that Gather function works on a valid proc value
func TestProcGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Procval{
		Files: map[string]string{"valid": "testdata/valid.procval"},
	}

	err := p.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]interface{}{
		"valid": 42,
	}
	acc.AssertContainsFields(t, "procval", fields)
}

// Test Gather function with invalid proc values or paths
func TestEmptyProcGather(t *testing.T) {
	var tests = []struct {
		fieldName string
		procPath  string
	}{
		{
			"empty",
			"testdata/empty.procval",
		},
		{
			"float",
			"testdata/float.procval",
		},
		{
			"string",
			"testdata/string.procval",
		},
		{
			"notExisting",
			"testdata/does_not_exist.procval",
		},
	}

	var acc testutil.Accumulator
	for _, test := range tests {
		p := Procval{
			Files: map[string]string{test.fieldName: test.procPath},
		}
		err := p.Gather(&acc)
		if err == nil {
			t.Errorf("no error produced")
		}
	}
}
