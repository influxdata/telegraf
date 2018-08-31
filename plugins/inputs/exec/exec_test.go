package exec

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"

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

const lineProtocol = "cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1\n"
const lineProtocolEmpty = ""
const lineProtocolShort = "ab"

const lineProtocolMulti = `
cpu,cpu=cpu0,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu1,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu2,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu5,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu6,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

type CarriageReturnTest struct {
	input  []byte
	output []byte
}

var crTests = []CarriageReturnTest{
	{[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0d, 0x0a, 0x4c, 0x69,
		0x6e, 0x65, 0x20, 0x32, 0x0d, 0x0a, 0x4c, 0x69, 0x6e, 0x65,
		0x20, 0x33},
		[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0a, 0x4c, 0x69, 0x6e,
			0x65, 0x20, 0x32, 0x0a, 0x4c, 0x69, 0x6e, 0x65, 0x20, 0x33}},
	{[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0a, 0x4c, 0x69, 0x6e,
		0x65, 0x20, 0x32, 0x0a, 0x4c, 0x69, 0x6e, 0x65, 0x20, 0x33},
		[]byte{0x4c, 0x69, 0x6e, 0x65, 0x20, 0x31, 0x0a, 0x4c, 0x69, 0x6e,
			0x65, 0x20, 0x32, 0x0a, 0x4c, 0x69, 0x6e, 0x65, 0x20, 0x33}},
	{[]byte{0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x6c,
		0x6c, 0x20, 0x6f, 0x6e, 0x65, 0x20, 0x62, 0x69, 0x67, 0x20,
		0x6c, 0x69, 0x6e, 0x65},
		[]byte{0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x6c,
			0x6c, 0x20, 0x6f, 0x6e, 0x65, 0x20, 0x62, 0x69, 0x67, 0x20,
			0x6c, 0x69, 0x6e, 0x65}},
}

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

func (r runnerMock) Run(e *Exec, command string, acc telegraf.Accumulator) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.out, nil
}

func TestExec(t *testing.T) {
	parser, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "exec",
	})
	e := &Exec{
		runner:   newRunnerMock([]byte(validJson), nil),
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
		runner:   newRunnerMock([]byte(malformedJson), nil),
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
		runner:   newRunnerMock(nil, fmt.Errorf("exit status code 1")),
		Commands: []string{"badcommand"},
		parser:   parser,
	}

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(e.Gather))
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}

func TestExecCommandWithGlob(t *testing.T) {
	parser, _ := parsers.NewValueParser("metric", "string", nil)
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
	parser, _ := parsers.NewValueParser("metric", "string", nil)
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
	parser, _ := parsers.NewValueParser("metric", "string", nil)
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

func TestRemoveCarriageReturns(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Test that all carriage returns are removed
		for _, test := range crTests {
			b := bytes.NewBuffer(test.input)
			out := removeCarriageReturns(*b)
			assert.True(t, bytes.Equal(test.output, out.Bytes()))
		}
	} else {
		// Test that the buffer is returned unaltered
		for _, test := range crTests {
			b := bytes.NewBuffer(test.input)
			out := removeCarriageReturns(*b)
			assert.True(t, bytes.Equal(test.input, out.Bytes()))
		}
	}
}
