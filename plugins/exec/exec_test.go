package exec

import (
	"fmt"
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
	"time"
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

type runnerMock struct {
	out []byte
	err error
}

type clockMock struct {
	now time.Time
}

func newRunnerMock(out []byte, err error) Runner {
	return &runnerMock{
		out: out,
		err: err,
	}
}

func (r runnerMock) Run(command *Command) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.out, nil
}

func newClockMock(now time.Time) Clock {
	return &clockMock{now: now}
}

func (c clockMock) Now() time.Time {
	return c.now
}

func TestExec(t *testing.T) {
	runner := newRunnerMock([]byte(validJson), nil)
	clock := newClockMock(time.Unix(baseTimeSeconds+20, 0))
	command := Command{
		Command:   "testcommand arg1",
		Name:      "mycollector",
		Interval:  10,
		lastRunAt: time.Unix(baseTimeSeconds, 0),
	}

	e := &Exec{
		runner:   runner,
		clock:    clock,
		Commands: []*Command{&command},
	}

	var acc testutil.Accumulator
	initialPoints := len(acc.Points)
	err := e.Gather(&acc)
	deltaPoints := len(acc.Points) - initialPoints
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

	assert.Equal(t, deltaPoints, 4, "non-numeric measurements should be ignored")
}

func TestExecMalformed(t *testing.T) {
	runner := newRunnerMock([]byte(malformedJson), nil)
	clock := newClockMock(time.Unix(baseTimeSeconds+20, 0))
	command := Command{
		Command:   "badcommand arg1",
		Name:      "mycollector",
		Interval:  10,
		lastRunAt: time.Unix(baseTimeSeconds, 0),
	}

	e := &Exec{
		runner:   runner,
		clock:    clock,
		Commands: []*Command{&command},
	}

	var acc testutil.Accumulator
	initialPoints := len(acc.Points)
	err := e.Gather(&acc)
	deltaPoints := len(acc.Points) - initialPoints
	require.Error(t, err)

	assert.Equal(t, deltaPoints, 0, "No new points should have been added")
}

func TestCommandError(t *testing.T) {
	runner := newRunnerMock(nil, fmt.Errorf("exit status code 1"))
	clock := newClockMock(time.Unix(baseTimeSeconds+20, 0))
	command := Command{
		Command:   "badcommand",
		Name:      "mycollector",
		Interval:  10,
		lastRunAt: time.Unix(baseTimeSeconds, 0),
	}

	e := &Exec{
		runner:   runner,
		clock:    clock,
		Commands: []*Command{&command},
	}

	var acc testutil.Accumulator
	initialPoints := len(acc.Points)
	err := e.Gather(&acc)
	deltaPoints := len(acc.Points) - initialPoints
	require.Error(t, err)

	assert.Equal(t, deltaPoints, 0, "No new points should have been added")
}

func TestExecNotEnoughTime(t *testing.T) {
	runner := newRunnerMock([]byte(validJson), nil)
	clock := newClockMock(time.Unix(baseTimeSeconds+5, 0))
	command := Command{
		Command:   "testcommand arg1",
		Name:      "mycollector",
		Interval:  10,
		lastRunAt: time.Unix(baseTimeSeconds, 0),
	}

	e := &Exec{
		runner:   runner,
		clock:    clock,
		Commands: []*Command{&command},
	}

	var acc testutil.Accumulator
	initialPoints := len(acc.Points)
	err := e.Gather(&acc)
	deltaPoints := len(acc.Points) - initialPoints
	require.NoError(t, err)

	assert.Equal(t, deltaPoints, 0, "No new points should have been added")
}

func TestExecUninitializedLastRunAt(t *testing.T) {
	runner := newRunnerMock([]byte(validJson), nil)
	clock := newClockMock(time.Unix(baseTimeSeconds, 0))
	command := Command{
		Command:  "testcommand arg1",
		Name:     "mycollector",
		Interval: math.MaxInt32,
		// Uninitialized lastRunAt should default to time.Unix(0, 0), so this should
		// run no matter what the interval is
	}

	e := &Exec{
		runner:   runner,
		clock:    clock,
		Commands: []*Command{&command},
	}

	var acc testutil.Accumulator
	initialPoints := len(acc.Points)
	err := e.Gather(&acc)
	deltaPoints := len(acc.Points) - initialPoints
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

	assert.Equal(t, deltaPoints, 4, "non-numeric measurements should be ignored")
}
func TestExecOneNotEnoughTimeAndOneEnoughTime(t *testing.T) {
	runner := newRunnerMock([]byte(validJson), nil)
	clock := newClockMock(time.Unix(baseTimeSeconds+5, 0))
	notEnoughTimeCommand := Command{
		Command:   "testcommand arg1",
		Name:      "mycollector",
		Interval:  10,
		lastRunAt: time.Unix(baseTimeSeconds, 0),
	}
	enoughTimeCommand := Command{
		Command:   "testcommand arg1",
		Name:      "mycollector",
		Interval:  3,
		lastRunAt: time.Unix(baseTimeSeconds, 0),
	}

	e := &Exec{
		runner:   runner,
		clock:    clock,
		Commands: []*Command{&notEnoughTimeCommand, &enoughTimeCommand},
	}

	var acc testutil.Accumulator
	initialPoints := len(acc.Points)
	err := e.Gather(&acc)
	deltaPoints := len(acc.Points) - initialPoints
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

	assert.Equal(t, deltaPoints, 4, "Only one command should have been run")
}
