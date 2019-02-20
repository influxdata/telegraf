package internal

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	shell, _    = exec.LookPath("sh")
)

func TestRunTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test due to random failures.")
	}
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
	// TODO: Fix this test
	t.Skip("Test failing too often, skip for now and revisit later.")
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
	if shell == "" {
		t.Skip("'sh' binary not available on OS, skipping.")
	}
	cmd := exec.Command(shell, "-c", "false")
	expected, err := cmd.CombinedOutput()

	cmd2 := exec.Command(shell, "-c", "false")
	actual, err := CombinedOutputTimeout(cmd2, time.Second)

	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

func TestRunError(t *testing.T) {
	if shell == "" {
		t.Skip("'sh' binary not available on OS, skipping.")
	}
	cmd := exec.Command(shell, "-c", "false")
	err := RunTimeout(cmd, time.Second)

	assert.Error(t, err)
}

func TestRandomSleep(t *testing.T) {
	// TODO: Fix this test
	t.Skip("Test failing too often, skip for now and revisit later.")
	// test that zero max returns immediately
	s := time.Now()
	RandomSleep(time.Duration(0), make(chan struct{}))
	elapsed := time.Since(s)
	assert.True(t, elapsed < time.Millisecond)

	// test that max sleep is respected
	s = time.Now()
	RandomSleep(time.Millisecond*50, make(chan struct{}))
	elapsed = time.Since(s)
	assert.True(t, elapsed < time.Millisecond*100)

	// test that shutdown is respected
	s = time.Now()
	shutdown := make(chan struct{})
	go func() {
		time.Sleep(time.Millisecond * 100)
		close(shutdown)
	}()
	RandomSleep(time.Second, shutdown)
	elapsed = time.Since(s)
	assert.True(t, elapsed < time.Millisecond*150)
}

func TestDuration(t *testing.T) {
	var d Duration

	d.UnmarshalTOML([]byte(`"1s"`))
	assert.Equal(t, time.Second, d.Duration)

	d = Duration{}
	d.UnmarshalTOML([]byte(`1s`))
	assert.Equal(t, time.Second, d.Duration)

	d = Duration{}
	d.UnmarshalTOML([]byte(`'1s'`))
	assert.Equal(t, time.Second, d.Duration)

	d = Duration{}
	d.UnmarshalTOML([]byte(`10`))
	assert.Equal(t, 10*time.Second, d.Duration)

	d = Duration{}
	d.UnmarshalTOML([]byte(`1.5`))
	assert.Equal(t, time.Second, d.Duration)
}

func TestSize(t *testing.T) {
	var s Size

	s.UnmarshalTOML([]byte(`"1B"`))
	assert.Equal(t, int64(1), s.Size)

	s = Size{}
	s.UnmarshalTOML([]byte(`1`))
	assert.Equal(t, int64(1), s.Size)

	s = Size{}
	s.UnmarshalTOML([]byte(`'1'`))
	assert.Equal(t, int64(1), s.Size)

	s = Size{}
	s.UnmarshalTOML([]byte(`"1GB"`))
	assert.Equal(t, int64(1000*1000*1000), s.Size)

	s = Size{}
	s.UnmarshalTOML([]byte(`"12GiB"`))
	assert.Equal(t, int64(12*1024*1024*1024), s.Size)
}

func TestCompressWithGzip(t *testing.T) {
	testData := "the quick brown fox jumps over the lazy dog"
	inputBuffer := bytes.NewBuffer([]byte(testData))

	outputBuffer, err := CompressWithGzip(inputBuffer)
	assert.NoError(t, err)

	gzipReader, err := gzip.NewReader(outputBuffer)
	assert.NoError(t, err)
	defer gzipReader.Close()

	output, err := ioutil.ReadAll(gzipReader)
	assert.NoError(t, err)

	assert.Equal(t, testData, string(output))
}

func TestVersionAlreadySet(t *testing.T) {
	err := SetVersion("foo")
	assert.Nil(t, err)

	err = SetVersion("bar")

	assert.NotNil(t, err)
	assert.IsType(t, VersionAlreadySetError, err)

	assert.Equal(t, "foo", Version())
}

func TestAlignDuration(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		interval time.Duration
		expected time.Duration
	}{
		{
			name:     "aligned",
			now:      time.Date(2018, 1, 1, 1, 1, 0, 0, time.UTC),
			interval: 10 * time.Second,
			expected: 0 * time.Second,
		},
		{
			name:     "standard interval",
			now:      time.Date(2018, 1, 1, 1, 1, 1, 0, time.UTC),
			interval: 10 * time.Second,
			expected: 9 * time.Second,
		},
		{
			name:     "odd interval",
			now:      time.Date(2018, 1, 1, 1, 1, 1, 0, time.UTC),
			interval: 3 * time.Second,
			expected: 2 * time.Second,
		},
		{
			name:     "sub second interval",
			now:      time.Date(2018, 1, 1, 1, 1, 0, 5e8, time.UTC),
			interval: 1 * time.Second,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "non divisible not aligned on minutes",
			now:      time.Date(2018, 1, 1, 1, 0, 0, 0, time.UTC),
			interval: 1*time.Second + 100*time.Millisecond,
			expected: 400 * time.Millisecond,
		},
		{
			name:     "long interval",
			now:      time.Date(2018, 1, 1, 1, 1, 0, 0, time.UTC),
			interval: 1 * time.Hour,
			expected: 59 * time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := AlignDuration(tt.now, tt.interval)
			require.Equal(t, tt.expected, actual)
		})
	}
}
