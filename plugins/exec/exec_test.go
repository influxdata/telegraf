package exec

import (
	"fmt"
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const validJson = `
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

const malformedJson = `
{
    "status": "green",
`

type runnerMock struct {
	out []byte
	err error
}

func newRunnerMock(out []byte, err error) Runner {
	return &runnerMock{out: out, err: err}
}

func (r runnerMock) Run(command string, args ...string) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.out, nil
}

func TestExec(t *testing.T) {
	runner := newRunnerMock([]byte(validJson), nil)
	command := Command{Command: "testcommand arg1", Name: "mycollector"}
	e := &Exec{runner: runner, Commands: []*Command{&command}}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.NoError(t, err)

	checkFloat := []struct {
		name  string
		value float64
	}{
		{"mycollector_num_processes", 82},
		{"mycollector_cpu_used", 8234},
		{"mycollector_cpu_free", 32},
		{"mycollector_percent", 0.81},
	}

	for _, c := range checkFloat {
		assert.True(t, acc.CheckValue(c.name, c.value))
	}

	assert.Equal(t, len(acc.Points), 4, "non-numeric measurements should be ignored")
}

func TestExecMalformed(t *testing.T) {
	runner := newRunnerMock([]byte(malformedJson), nil)
	command := Command{Command: "badcommand arg1", Name: "mycollector"}
	e := &Exec{runner: runner, Commands: []*Command{&command}}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.Error(t, err)
}

func TestCommandError(t *testing.T) {
	runner := newRunnerMock(nil, fmt.Errorf("exit status code 1"))
	command := Command{Command: "badcommand", Name: "mycollector"}
	e := &Exec{runner: runner, Commands: []*Command{&command}}
	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.Error(t, err)
}
