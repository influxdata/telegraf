package exec

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
	"time"

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

func (r runnerMock) Run(command string, _ time.Duration) ([]byte, []byte, error) {
	return r.out, r.errout, r.err
}

func TestExec(t *testing.T) {
	parser, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "exec",
	})
	e := &Exec{
		Log:      testutil.Logger{},
		runner:   newRunnerMock([]byte(validJson), nil, nil),
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
		runner:   newRunnerMock([]byte(malformedJson), nil, nil),
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

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		bufF func() *bytes.Buffer
		expF func() *bytes.Buffer
	}{
		{
			name: "should not truncate",
			bufF: func() *bytes.Buffer {
				var b bytes.Buffer
				b.WriteString("hello world")
				return &b
			},
			expF: func() *bytes.Buffer {
				var b bytes.Buffer
				b.WriteString("hello world")
				return &b
			},
		},
		{
			name: "should truncate up to the new line",
			bufF: func() *bytes.Buffer {
				var b bytes.Buffer
				b.WriteString("hello world\nand all the people")
				return &b
			},
			expF: func() *bytes.Buffer {
				var b bytes.Buffer
				b.WriteString("hello world...")
				return &b
			},
		},
		{
			name: "should truncate to the MaxStderrBytes",
			bufF: func() *bytes.Buffer {
				var b bytes.Buffer
				for i := 0; i < 2*MaxStderrBytes; i++ {
					b.WriteByte('b')
				}
				return &b
			},
			expF: func() *bytes.Buffer {
				var b bytes.Buffer
				for i := 0; i < MaxStderrBytes; i++ {
					b.WriteByte('b')
				}
				b.WriteString("...")
				return &b
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := truncate(*tt.bufF())
			require.Equal(t, tt.expF().Bytes(), res.Bytes())
		})
	}
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
