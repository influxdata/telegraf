package internal

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SnakeTest struct {
	input  string
	output string
}

var tests = []SnakeTest{
	{"a", "a"},
	{"snake", "snake"},
	{"A", "a"},
	{"ID", "id"},
	{"MOTD", "motd"},
	{"Snake", "snake"},
	{"SnakeTest", "snake_test"},
	{"APIResponse", "api_response"},
	{"SnakeID", "snake_id"},
	{"SnakeIDGoogle", "snake_id_google"},
	{"LinuxMOTD", "linux_motd"},
	{"OMGWTFBBQ", "omgwtfbbq"},
	{"omg_wtf_bbq", "omg_wtf_bbq"},
}

func TestSnakeCase(t *testing.T) {
	for _, test := range tests {
		if SnakeCase(test.input) != test.output {
			t.Errorf(`SnakeCase("%s"), wanted "%s", got \%s"`, test.input, test.output, SnakeCase(test.input))
		}
	}
}

var (
	sleepbin, _ = exec.LookPath("sleep")
	echobin, _  = exec.LookPath("echo")
)

func TestRunTimeout(t *testing.T) {
	if sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}
	cmd := exec.Command(sleepbin, "10")
	start := time.Now()
	err := RunTimeout(cmd, time.Millisecond*20)
	elapsed := time.Since(start)

	assert.Equal(t, TimeoutErr, err)
	// Verify that command gets killed in 20ms, with some breathing room
	assert.True(t, elapsed < time.Millisecond*75)
}

func TestCombinedOutputTimeout(t *testing.T) {
	if sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}
	cmd := exec.Command(sleepbin, "10")
	start := time.Now()
	_, err := CombinedOutputTimeout(cmd, time.Millisecond*20)
	elapsed := time.Since(start)

	assert.Equal(t, TimeoutErr, err)
	// Verify that command gets killed in 20ms, with some breathing room
	assert.True(t, elapsed < time.Millisecond*75)
}

func TestCombinedOutput(t *testing.T) {
	if echobin == "" {
		t.Skip("'echo' binary not available on OS, skipping.")
	}
	cmd := exec.Command(echobin, "foo")
	out, err := CombinedOutputTimeout(cmd, time.Second)

	assert.NoError(t, err)
	assert.Equal(t, "foo\n", string(out))
}

// test that CombinedOutputTimeout and exec.Cmd.CombinedOutput return
// the same output from a failed command.
func TestCombinedOutputError(t *testing.T) {
	if sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}
	cmd := exec.Command(sleepbin, "foo")
	expected, err := cmd.CombinedOutput()

	cmd2 := exec.Command(sleepbin, "foo")
	actual, err := CombinedOutputTimeout(cmd2, time.Second)

	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

func TestRunError(t *testing.T) {
	if sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}
	cmd := exec.Command(sleepbin, "foo")
	err := RunTimeout(cmd, time.Second)

	assert.Error(t, err)
}
