package internal

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
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
	sleepbin, _ = exec.LookPath("sleep") //nolint:unused // Used in skipped tests
	echobin, _  = exec.LookPath("echo")
	shell, _    = exec.LookPath("sh")
)

func TestRunTimeout(t *testing.T) {
	t.Skip("Skipping test due to random failures & a data race when running test-all.")

	if sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}
	cmd := exec.Command(sleepbin, "10")
	start := time.Now()
	err := RunTimeout(cmd, time.Millisecond*20)
	elapsed := time.Since(start)

	assert.Equal(t, ErrTimeout, err)
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

	assert.Equal(t, ErrTimeout, err)
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

func TestCompressWithGzip(t *testing.T) {
	testData := "the quick brown fox jumps over the lazy dog"
	inputBuffer := bytes.NewBuffer([]byte(testData))

	outputBuffer, err := CompressWithGzip(inputBuffer)
	assert.NoError(t, err)

	gzipReader, err := gzip.NewReader(outputBuffer)
	assert.NoError(t, err)
	defer gzipReader.Close()

	output, err := io.ReadAll(gzipReader)
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

	n, err := io.CopyN(io.Discard, rc, 10000)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000), n)

	r1 := mr.readN
	err = rc.Close()
	assert.NoError(t, err)

	n, err = io.CopyN(io.Discard, rc, 10000)
	assert.Error(t, io.EOF, err)
	assert.Equal(t, int64(0), n)

	r2 := mr.readN
	// no more read to the source after closing
	assert.Equal(t, r1, r2)
}

func TestVersionAlreadySet(t *testing.T) {
	err := SetVersion("foo")
	assert.NoError(t, err)

	err = SetVersion("bar")

	assert.Error(t, err)
	assert.IsType(t, ErrorVersionAlreadySet, err)

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
		require.NoError(t, err)
		return tm
	}
	ansic := func(value string) time.Time {
		tm, err := time.Parse(time.ANSIC, value)
		require.NoError(t, err)
		return tm
	}

	unixdate := func(value string) time.Time {
		tm, err := time.Parse(time.UnixDate, value)
		require.NoError(t, err)
		return tm
	}

	rubydate := func(value string) time.Time {
		tm, err := time.Parse(time.RubyDate, value)
		require.NoError(t, err)
		return tm
	}

	rfc822 := func(value string) time.Time {
		tm, err := time.Parse(time.RFC822, value)
		require.NoError(t, err)
		return tm
	}

	rfc822z := func(value string) time.Time {
		tm, err := time.Parse(time.RFC822Z, value)
		require.NoError(t, err)
		return tm
	}

	rfc850 := func(value string) time.Time {
		tm, err := time.Parse(time.RFC850, value)
		require.NoError(t, err)
		return tm
	}

	rfc1123 := func(value string) time.Time {
		tm, err := time.Parse(time.RFC1123, value)
		require.NoError(t, err)
		return tm
	}

	rfc1123z := func(value string) time.Time {
		tm, err := time.Parse(time.RFC1123Z, value)
		require.NoError(t, err)
		return tm
	}

	rfc3339nano := func(value string) time.Time {
		tm, err := time.Parse(time.RFC3339Nano, value)
		require.NoError(t, err)
		return tm
	}

	stamp := func(value string) time.Time {
		tm, err := time.Parse(time.Stamp, value)
		require.NoError(t, err)
		return tm
	}

	stampmilli := func(value string) time.Time {
		tm, err := time.Parse(time.StampMilli, value)
		require.NoError(t, err)
		return tm
	}

	stampmicro := func(value string) time.Time {
		tm, err := time.Parse(time.StampMicro, value)
		require.NoError(t, err)
		return tm
	}

	stampnano := func(value string) time.Time {
		tm, err := time.Parse(time.StampNano, value)
		require.NoError(t, err)
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
		{
			name:      "rfc339 test",
			format:    "RFC3339",
			timestamp: "2018-10-26T13:30:33Z",
			expected:  rfc3339("2018-10-26T13:30:33Z"),
		},

		{
			name:      "ANSIC",
			format:    "ANSIC",
			timestamp: "Mon Jan 2 15:04:05 2006",
			expected:  ansic("Mon Jan 2 15:04:05 2006"),
		},

		{
			name:      "UnixDate",
			format:    "UnixDate",
			timestamp: "Mon Jan 2 15:04:05 MST 2006",
			expected:  unixdate("Mon Jan 2 15:04:05 MST 2006"),
			location:  "Local",
		},

		{
			name:      "RubyDate",
			format:    "RubyDate",
			timestamp: "Mon Jan 02 15:04:05 -0700 2006",
			expected:  rubydate("Mon Jan 02 15:04:05 -0700 2006"),
			location:  "Local",
		},

		{
			name:      "RFC822",
			format:    "RFC822",
			timestamp: "02 Jan 06 15:04 MST",
			expected:  rfc822("02 Jan 06 15:04 MST"),
			location:  "Local",
		},

		{
			name:      "RFC822Z",
			format:    "RFC822Z",
			timestamp: "02 Jan 06 15:04 -0700",
			expected:  rfc822z("02 Jan 06 15:04 -0700"),
			location:  "Local",
		},

		{
			name:      "RFC850",
			format:    "RFC850",
			timestamp: "Monday, 02-Jan-06 15:04:05 MST",
			expected:  rfc850("Monday, 02-Jan-06 15:04:05 MST"),
			location:  "Local",
		},

		{
			name:      "RFC1123",
			format:    "RFC1123",
			timestamp: "Mon, 02 Jan 2006 15:04:05 MST",
			expected:  rfc1123("Mon, 02 Jan 2006 15:04:05 MST"),
			location:  "Local",
		},

		{
			name:      "RFC1123Z",
			format:    "RFC1123Z",
			timestamp: "Mon, 02 Jan 2006 15:04:05 -0700",
			expected:  rfc1123z("Mon, 02 Jan 2006 15:04:05 -0700"),
			location:  "Local",
		},

		{
			name:      "RFC3339Nano",
			format:    "RFC3339Nano",
			timestamp: "2006-01-02T15:04:05.999999999-07:00",
			expected:  rfc3339nano("2006-01-02T15:04:05.999999999-07:00"),
			location:  "Local",
		},

		{
			name:      "Stamp",
			format:    "Stamp",
			timestamp: "Jan 2 15:04:05",
			expected:  stamp("Jan 2 15:04:05"),
		},

		{
			name:      "StampMilli",
			format:    "StampMilli",
			timestamp: "Jan 2 15:04:05.000",
			expected:  stampmilli("Jan 2 15:04:05.000"),
		},

		{
			name:      "StampMicro",
			format:    "StampMicro",
			timestamp: "Jan 2 15:04:05.000000",
			expected:  stampmicro("Jan 2 15:04:05.000000"),
		},

		{
			name:      "StampNano",
			format:    "StampNano",
			timestamp: "Jan 2 15:04:05.000000000",
			expected:  stampnano("Jan 2 15:04:05.000000000"),
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
	re := regexp.MustCompile(`^Telegraf/[^\s]+ Go/\d+.\d+(.\d+)?$`)
	require.True(t, re.MatchString(token), token)
}
