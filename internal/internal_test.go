package internal

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
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

// Verifies behavior of a command that doesn't get killed.
func TestRunTimeoutFastExit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test due to random failures.")
	}
	if echobin == "" {
		t.Skip("'echo' binary not available on OS, skipping.")
	}
	cmd := exec.Command(echobin)
	start := time.Now()
	err := RunTimeout(cmd, time.Millisecond*20)
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	elapsed := time.Since(start)

	require.NoError(t, err)
	// Verify that command gets killed in 20ms, with some breathing room
	assert.True(t, elapsed < time.Millisecond*75)

	// Verify "process already finished" log doesn't occur.
	time.Sleep(time.Millisecond * 75)
	require.Equal(t, "", buf.String())
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

type mockReader struct {
	readN uint64 // record the number of calls to Read
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	r.readN++
	return rand.Read(p)
}

func TestCompressWithGzipEarlyClose(t *testing.T) {
	mr := &mockReader{}

	rc, err := CompressWithGzip(mr)
	assert.NoError(t, err)

	n, err := io.CopyN(ioutil.Discard, rc, 10000)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000), n)

	r1 := mr.readN
	err = rc.Close()
	assert.NoError(t, err)

	n, err = io.CopyN(ioutil.Discard, rc, 10000)
	assert.Error(t, io.EOF, err)
	assert.Equal(t, int64(0), n)

	r2 := mr.readN
	// no more read to the source after closing
	assert.Equal(t, r1, r2)
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

func TestAlignTime(t *testing.T) {
	rfc3339 := func(value string) time.Time {
		t, _ := time.Parse(time.RFC3339, value)
		return t
	}

	tests := []struct {
		name     string
		now      time.Time
		interval time.Duration
		expected time.Time
	}{
		{
			name:     "aligned",
			now:      rfc3339("2018-01-01T01:01:00Z"),
			interval: 10 * time.Second,
			expected: rfc3339("2018-01-01T01:01:00Z"),
		},
		{
			name:     "aligned",
			now:      rfc3339("2018-01-01T01:01:01Z"),
			interval: 10 * time.Second,
			expected: rfc3339("2018-01-01T01:01:10Z"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := AlignTime(tt.now, tt.interval)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	rfc3339 := func(value string) time.Time {
		tm, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			panic(err)
		}
		return tm
	}

	tests := []struct {
		name      string
		format    string
		timestamp interface{}
		location  string
		expected  time.Time
		err       bool
	}{
		{
			name:      "parse layout string in utc",
			format:    "2006-01-02 15:04:05",
			timestamp: "2019-02-20 21:50:34",
			location:  "UTC",
			expected:  rfc3339("2019-02-20T21:50:34Z"),
		},
		{
			name:      "parse layout string with invalid timezone",
			format:    "2006-01-02 15:04:05",
			timestamp: "2019-02-20 21:50:34",
			location:  "InvalidTimeZone",
			err:       true,
		},
		{
			name:      "layout regression 6386",
			format:    "02.01.2006 15:04:05",
			timestamp: "09.07.2019 00:11:00",
			expected:  rfc3339("2019-07-09T00:11:00Z"),
		},
		{
			name:      "default location is utc",
			format:    "2006-01-02 15:04:05",
			timestamp: "2019-02-20 21:50:34",
			expected:  rfc3339("2019-02-20T21:50:34Z"),
		},
		{
			name:      "unix seconds without fractional",
			format:    "unix",
			timestamp: "1568338208",
			expected:  rfc3339("2019-09-13T01:30:08Z"),
		},
		{
			name:      "unix seconds with fractional",
			format:    "unix",
			timestamp: "1568338208.500",
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix seconds with fractional and comma decimal point",
			format:    "unix",
			timestamp: "1568338208,500",
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix seconds extra precision",
			format:    "unix",
			timestamp: "1568338208.00000050042",
			expected:  rfc3339("2019-09-13T01:30:08.000000500Z"),
		},
		{
			name:      "unix seconds integer",
			format:    "unix",
			timestamp: int64(1568338208),
			expected:  rfc3339("2019-09-13T01:30:08Z"),
		},
		{
			name:      "unix seconds float",
			format:    "unix",
			timestamp: float64(1568338208.500),
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix milliseconds",
			format:    "unix_ms",
			timestamp: "1568338208500",
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix milliseconds with fractional is ignored",
			format:    "unix_ms",
			timestamp: "1568338208500.42",
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix microseconds",
			format:    "unix_us",
			timestamp: "1568338208000500",
			expected:  rfc3339("2019-09-13T01:30:08.000500Z"),
		},
		{
			name:      "unix nanoseconds",
			format:    "unix_ns",
			timestamp: "1568338208000000500",
			expected:  rfc3339("2019-09-13T01:30:08.000000500Z"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, err := ParseTimestamp(tt.format, tt.timestamp, tt.location)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, tm)
			}
		})
	}
}

func TestProductToken(t *testing.T) {
	token := ProductToken()
	// Telegraf version depends on the call to SetVersion, it cannot be set
	// multiple times and is not thread-safe.
	re := regexp.MustCompile(`^Telegraf/[^\s]+ Go/\d+.\d+.\d+$`)
	require.True(t, re.MatchString(token), token)
}
