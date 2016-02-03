package exec

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Midnight 9/22/2015
const baseTimeSeconds = 1442905200

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

const lineProtocol = "cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1"

const lineProtocolMulti = `
cpu,cpu=cpu0,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu1,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu2,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu5,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu6,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

type runnerMock struct {
	out []byte
	err error
}

func newRunnerMock(out []byte, err error) Runner {
	return &runnerMock{
		out: out,
		err: err,
	}
}

func (r runnerMock) Run(e *Exec, command string) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.out, nil
}

func TestExec(t *testing.T) {
	e := &Exec{
		runner:   newRunnerMock([]byte(validJson), nil),
		Commands: []string{"testcommand arg1"},
	}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
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
	e := &Exec{
		runner:   newRunnerMock([]byte(malformedJson), nil),
		Commands: []string{"badcommand arg1"},
	}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.Error(t, err)
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}

func TestCommandError(t *testing.T) {
	e := &Exec{
		runner:   newRunnerMock(nil, fmt.Errorf("exit status code 1")),
		Commands: []string{"badcommand"},
	}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.Error(t, err)
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}

func TestLineProtocolParse(t *testing.T) {
	e := &Exec{
		runner:     newRunnerMock([]byte(lineProtocol), nil),
		Commands:   []string{"line-protocol"},
		DataFormat: "influx",
	}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	tags := map[string]string{
		"host":       "foo",
		"datacenter": "us-east",
	}
	acc.AssertContainsTaggedFields(t, "cpu", fields, tags)
}

func TestLineProtocolParseMultiple(t *testing.T) {
	e := &Exec{
		runner:     newRunnerMock([]byte(lineProtocolMulti), nil),
		Commands:   []string{"line-protocol"},
		DataFormat: "influx",
	}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	tags := map[string]string{
		"host":       "foo",
		"datacenter": "us-east",
	}
	cpuTags := []string{"cpu0", "cpu1", "cpu2", "cpu3", "cpu4", "cpu5", "cpu6"}

	for _, cpu := range cpuTags {
		tags["cpu"] = cpu
		acc.AssertContainsTaggedFields(t, "cpu", fields, tags)
	}
}

func TestInvalidDataFormat(t *testing.T) {
	e := &Exec{
		runner:     newRunnerMock([]byte(lineProtocol), nil),
		Commands:   []string{"bad data format"},
		DataFormat: "FooBar",
	}

	var acc testutil.Accumulator
	err := e.Gather(&acc)
	require.Error(t, err)
}
