//go:build !windows
// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package exec

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSON = `
{
    "status": "green",
    "num_processes": 82,
    "cpu": {
        "status": "red",
        "nil_status": null,
        "used": 8234,
        "free": 32
    },
    "percent": 0.81,
    "users": [0, 1, 2, 3]
}`

const malformedJSON = `
{
    "status": "green",
`

type runnerMock struct {
	out    []byte
	errout []byte
	err    error
}

func newRunnerMock(out []byte, errout []byte, err error) Runner {
	return &runnerMock{
		out:    out,
		errout: errout,
		err:    err,
	}
}

func (r runnerMock) Run(_ string, _ []string, _ []string, _ []byte, _ time.Duration) ([]byte, []byte, error) {
	return r.out, r.errout, r.err
}

func TestExec(t *testing.T) {
	parser, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "exec",
	})
	e := &Exec{
		Log:      testutil.Logger{},
		runner:   newRunnerMock([]byte(validJSON), nil, nil),
		Commands: []string{"testcommand arg1"},
		parser:   parser,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(e.Gather)
	require.NoError(t, err)
	assert.Equal(t, acc.NFields(), 8, "non-numeric measurements should be ignored")

	fields := map[string]interface{}{
		"num_processes": float64(82),
		"cpu_used":      float64(8234),
		"cpu_free":      float64(32),
		"percent":       float64(0.81),
		"users_0":       float64(0),
		"users_1":       float64(1),
		"users_2":       float64(2),
		"users_3":       float64(3),
	}
	acc.AssertContainsFields(t, "exec", fields)
}

func TestExecMalformed(t *testing.T) {
	parser, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "exec",
	})
	e := &Exec{
		Log:      testutil.Logger{},
		runner:   newRunnerMock([]byte(malformedJSON), nil, nil),
		Commands: []string{"badcommand arg1"},
		parser:   parser,
	}

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(e.Gather))
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}

func TestCommandError(t *testing.T) {
	parser, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "exec",
	})
	e := &Exec{
		Log:      testutil.Logger{},
		runner:   newRunnerMock(nil, nil, fmt.Errorf("exit status code 1")),
		Commands: []string{"badcommand"},
		parser:   parser,
	}

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(e.Gather))
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}

func TestExecCommandWithGlob(t *testing.T) {
	parser, _ := parsers.NewValueParser("metric", "string", "", nil)
	e := NewExec()
	e.Commands = []string{"/bin/ech* metric_value"}
	e.SetParser(parser)

	var acc testutil.Accumulator
	err := acc.GatherError(e.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"value": "metric_value",
	}
	acc.AssertContainsFields(t, "metric", fields)
}

func TestExecCommandWithoutGlob(t *testing.T) {
	parser, _ := parsers.NewValueParser("metric", "string", "", nil)
	e := NewExec()
	e.Commands = []string{"/bin/echo metric_value"}
	e.SetParser(parser)

	var acc testutil.Accumulator
	err := acc.GatherError(e.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"value": "metric_value",
	}
	acc.AssertContainsFields(t, "metric", fields)
}

func TestExecCommandWithoutGlobAndPath(t *testing.T) {
	parser, _ := parsers.NewValueParser("metric", "string", "", nil)
	e := NewExec()
	e.Commands = []string{"echo metric_value"}
	e.SetParser(parser)

	var acc testutil.Accumulator
	err := acc.GatherError(e.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"value": "metric_value",
	}
	acc.AssertContainsFields(t, "metric", fields)
}
