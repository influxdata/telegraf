package internal

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
	"log"
	"os/exec"
	"regexp"
	"sync"
	"testing"
	"time"

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
		t.Run(test.input, func(t *testing.T) {
			require.Equal(t, test.output, SnakeCase(test.input))
		})
	}
}

func TestRunTimeout(t *testing.T) {
	t.Skip("Skipping test due to random failures & a data race when running test-all.")

	sleepbin, err := exec.LookPath("sleep")
	if err != nil || sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}

	cmd := exec.Command(sleepbin, "10")
	start := time.Now()
	err = RunTimeout(cmd, time.Millisecond*20)
	elapsed := time.Since(start)

	require.Equal(t, ErrTimeout, err)
	// Verify that command gets killed in 20ms, with some breathing room
	require.Less(t, elapsed, time.Millisecond*75)
}

// Verifies behavior of a command that doesn't get killed.
func TestRunTimeoutFastExit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test due to random failures.")
	}
	echobin, err := exec.LookPath("echo")
	if err != nil || echobin == "" {
		t.Skip("'echo' binary not available on OS, skipping.")
	}
	cmd := exec.Command(echobin)
	start := time.Now()
	err = RunTimeout(cmd, time.Millisecond*20)
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	elapsed := time.Since(start)

	require.NoError(t, err)
	// Verify that command gets killed in 20ms, with some breathing room
	require.Less(t, elapsed, time.Millisecond*75)

	// Verify "process already finished" log doesn't occur.
	time.Sleep(time.Millisecond * 75)
	require.Equal(t, "", buf.String())
}

func TestCombinedOutputTimeout(t *testing.T) {
	// TODO: Fix this test
	t.Skip("Test failing too often, skip for now and revisit later.")
	sleepbin, err := exec.LookPath("sleep")
	if err != nil || sleepbin == "" {
		t.Skip("'sleep' binary not available on OS, skipping.")
	}
	cmd := exec.Command(sleepbin, "10")
	start := time.Now()
	_, err = CombinedOutputTimeout(cmd, time.Millisecond*20)
	elapsed := time.Since(start)

	require.Equal(t, ErrTimeout, err)
	// Verify that command gets killed in 20ms, with some breathing room
	require.Less(t, elapsed, time.Millisecond*75)
}

func TestCombinedOutput(t *testing.T) {
	echobin, err := exec.LookPath("echo")
	if err != nil || echobin == "" {
		t.Skip("'echo' binary not available on OS, skipping.")
	}
	cmd := exec.Command(echobin, "foo")
	out, err := CombinedOutputTimeout(cmd, time.Second)

	require.NoError(t, err)
	require.Equal(t, "foo\n", string(out))
}

// test that CombinedOutputTimeout and exec.Cmd.CombinedOutput return
// the same output from a failed command.
func TestCombinedOutputError(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil || shell == "" {
		t.Skip("'sh' binary not available on OS, skipping.")
	}
	cmd := exec.Command(shell, "-c", "false")
	expected, err := cmd.CombinedOutput()
	require.Error(t, err)

	cmd2 := exec.Command(shell, "-c", "false")
	actual, err := CombinedOutputTimeout(cmd2, time.Second)

	require.Error(t, err)
	require.Equal(t, expected, actual)
}

func TestRunError(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil || shell == "" {
		t.Skip("'sh' binary not available on OS, skipping.")
	}
	cmd := exec.Command(shell, "-c", "false")
	err = RunTimeout(cmd, time.Second)

	require.Error(t, err)
}

func TestRandomSleep(t *testing.T) {
	// TODO: Fix this test
	t.Skip("Test failing too often, skip for now and revisit later.")
	// test that zero max returns immediately
	s := time.Now()
	RandomSleep(time.Duration(0), make(chan struct{}))
	elapsed := time.Since(s)
	require.Less(t, elapsed, time.Millisecond)

	// test that max sleep is respected
	s = time.Now()
	RandomSleep(time.Millisecond*50, make(chan struct{}))
	elapsed = time.Since(s)
	require.Less(t, elapsed, time.Millisecond*100)

	// test that shutdown is respected
	s = time.Now()
	shutdown := make(chan struct{})
	go func() {
		time.Sleep(time.Millisecond * 100)
		close(shutdown)
	}()
	RandomSleep(time.Second, shutdown)
	elapsed = time.Since(s)
	require.Less(t, elapsed, time.Millisecond*150)
}

func TestCompressWithGzip(t *testing.T) {
	testData := "the quick brown fox jumps over the lazy dog"
	inputBuffer := bytes.NewBufferString(testData)

	outputBuffer := CompressWithGzip(inputBuffer)
	gzipReader, err := gzip.NewReader(outputBuffer)
	require.NoError(t, err)
	defer gzipReader.Close()

	output, err := io.ReadAll(gzipReader)
	require.NoError(t, err)

	require.Equal(t, testData, string(output))
}

type mockReader struct {
	err    error
	ncalls uint64 // record the number of calls to Read
	msg    []byte
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	r.ncalls++

	if len(r.msg) > 0 {
		n, err = copy(p, r.msg), io.EOF
	} else {
		n, err = rand.Read(p)
	}
	if r.err == nil {
		return n, err
	}
	return n, r.err
}

func TestCompressWithGzipEarlyClose(t *testing.T) {
	mr := &mockReader{}

	rc := CompressWithGzip(mr)
	n, err := io.CopyN(io.Discard, rc, 10000)
	require.NoError(t, err)
	require.Equal(t, int64(10000), n)

	r1 := mr.ncalls
	require.NoError(t, rc.Close())

	n, err = io.CopyN(io.Discard, rc, 10000)
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Equal(t, int64(0), n)

	r2 := mr.ncalls
	// no more read to the source after closing
	require.Equal(t, r1, r2)
}

func TestCompressWithGzipErrorPropagationCopy(t *testing.T) {
	errs := []error{io.ErrClosedPipe, io.ErrNoProgress, io.ErrUnexpectedEOF}
	for _, expected := range errs {
		r := &mockReader{msg: []byte("this is a test"), err: expected}

		rc := CompressWithGzip(r)
		n, err := io.Copy(io.Discard, rc)
		require.Greater(t, n, int64(0))
		require.ErrorIs(t, err, expected)
		require.NoError(t, rc.Close())
	}
}

func TestCompressWithGzipErrorPropagationReadAll(t *testing.T) {
	errs := []error{io.ErrClosedPipe, io.ErrNoProgress, io.ErrUnexpectedEOF}
	for _, expected := range errs {
		r := &mockReader{msg: []byte("this is a test"), err: expected}

		rc := CompressWithGzip(r)
		buf, err := io.ReadAll(rc)
		require.NotEmpty(t, buf)
		require.ErrorIs(t, err, expected)
		require.NoError(t, rc.Close())
	}
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
		tt, err := time.Parse(time.RFC3339, value)
		require.NoError(t, err)
		return tt
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

	rubydate := func(value string) time.Time {
		tm, err := time.Parse(time.RubyDate, value)
		require.NoError(t, err)
		return tm
	}

	rfc822z := func(value string) time.Time {
		tm, err := time.Parse(time.RFC822Z, value)
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
		separator []string
		expected  time.Time
	}{
		{
			name:      "parse layout string in utc",
			format:    "2006-01-02 15:04:05",
			timestamp: "2019-02-20 21:50:34",
			location:  "UTC",
			expected:  rfc3339("2019-02-20T21:50:34Z"),
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
			name:      "unix seconds with thousand separator only (dot)",
			format:    "unix",
			timestamp: "1.568.338.208",
			separator: []string{","},
			expected:  rfc3339("2019-09-13T01:30:08Z"),
		},
		{
			name:      "unix seconds with thousand separator only (comma)",
			format:    "unix",
			timestamp: "1,568,338,208",
			separator: []string{"."},
			expected:  rfc3339("2019-09-13T01:30:08Z"),
		},
		{
			name:      "unix seconds with thousand separator only (space)",
			format:    "unix",
			timestamp: "1 568 338 208",
			separator: []string{"."},
			expected:  rfc3339("2019-09-13T01:30:08Z"),
		},
		{
			name:      "unix seconds with thousand separator only (underscore)",
			format:    "unix",
			timestamp: "1_568_338_208",
			separator: []string{"."},
			expected:  rfc3339("2019-09-13T01:30:08Z"),
		},
		{
			name:      "unix seconds with thousand and decimal separator (US)",
			format:    "unix",
			timestamp: "1,568,338,208.500",
			separator: []string{"."},
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix seconds with thousand and decimal separator (EU)",
			format:    "unix",
			timestamp: "1.568.338.208,500",
			separator: []string{","},
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
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
			name:      "unix seconds float exponential",
			format:    "unix",
			timestamp: float64(1.5683382085e+9),
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix milliseconds",
			format:    "unix_ms",
			timestamp: "1568338208500",
			expected:  rfc3339("2019-09-13T01:30:08.500Z"),
		},
		{
			name:      "unix milliseconds with fractional",
			format:    "unix_ms",
			timestamp: "1568338208500.42",
			expected:  rfc3339("2019-09-13T01:30:08.50042Z"),
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
			name:      "unix nanoseconds exponential",
			format:    "unix_ns",
			timestamp: "1.5683382080000005e+18",
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
			expected:  time.Unix(1136239445, 0),
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
			expected:  time.Unix(1136239440, 0),
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
			expected:  time.Unix(1136239445, 0),
			location:  "Local",
		},

		{
			name:      "RFC1123",
			format:    "RFC1123",
			timestamp: "Mon, 02 Jan 2006 15:04:05 MST",
			expected:  time.Unix(1136239445, 0),
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

		{
			name:      "RFC850",
			format:    "RFC850",
			timestamp: "Monday, 02-Jan-06 15:04:05 MST",
			expected:  time.Unix(1136239445, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure any one-time warnings are printed for each test
			once = sync.Once{}

			// Ensure the warnings are captured and not to stdout
			var buf bytes.Buffer
			backup := log.Writer()
			log.SetOutput(&buf)
			defer log.SetOutput(backup)

			var loc *time.Location
			if tt.location != "" {
				var err error
				loc, err = time.LoadLocation(tt.location)
				require.NoError(t, err)
			}
			tm, err := ParseTimestamp(tt.format, tt.timestamp, loc, tt.separator...)
			require.NoError(t, err)
			require.Equal(t, tt.expected.Unix(), tm.Unix())
		})
	}
}

func TestParseTimestampInvalid(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		timestamp interface{}
		expected  string
	}{
		{
			name:      "too few digits",
			format:    "2006-01-02 15:04:05",
			timestamp: "2019-02-20 21:50",
			expected:  "cannot parse \"\" as \":\"",
		},
		{
			name:      "invalid layout",
			format:    "rfc3399",
			timestamp: "09.07.2019 00:11:00",
			expected:  "cannot parse \"09.07.2019 00:11:00\" as \"rfc\"",
		},
		{
			name:      "layout not matching time",
			format:    "rfc3339",
			timestamp: "09.07.2019 00:11:00",
			expected:  "parsing time \"09.07.2019 00:11:00\" as \"2006-01-02T15:04:05Z07:00\": cannot parse",
		},
		{
			name:      "unix wrong type",
			format:    "unix",
			timestamp: true,
			expected:  "unsupported type",
		},
		{
			name:      "unix multiple separators (dot)",
			format:    "unix",
			timestamp: "1568338.208.500",
			expected:  "invalid number",
		},
		{
			name:      "unix multiple separators (comma)",
			format:    "unix",
			timestamp: "1568338,208,500",
			expected:  "invalid number",
		},
		{
			name:      "unix multiple separators (mixed)",
			format:    "unix",
			timestamp: "1,568,338,208.500",
			expected:  "invalid number",
		},
		{
			name:      "invalid timezone abbreviation",
			format:    "RFC850",
			timestamp: "Monday, 02-Jan-06 15:04:05 CDT",
			expected:  "cannot resolve timezone abbreviation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure any one-time warnings are printed for each test
			once = sync.Once{}

			// Ensure the warnings are captured and not to stdout
			var buf bytes.Buffer
			backup := log.Writer()
			log.SetOutput(&buf)
			defer log.SetOutput(backup)

			_, err := ParseTimestamp(tt.format, tt.timestamp, nil)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestTimestampAbbrevWarning(t *testing.T) {
	// Ensure any one-time warnings are printed for each test
	once = sync.Once{}

	// Ensure the warnings are captured and not to stdout
	var buf bytes.Buffer
	backup := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(backup)

	// Try multiple timestamps with abbreviated timezones in case a user
	// is actually in one of the timezones.
	ts, err := ParseTimestamp("RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", nil)
	require.NoError(t, err)
	require.EqualValues(t, 1136239445, ts.Unix())

	ts2, err := ParseTimestamp("RFC1123", "Mon, 02 Jan 2006 15:04:05 EST", nil)
	require.NoError(t, err)
	require.EqualValues(t, 1136232245, ts2.Unix())

	require.Contains(t, buf.String(), "Your config is using abbreviated timezones and parsing was changed in v1.27.0")
}

func TestProductToken(t *testing.T) {
	token := ProductToken()
	// Telegraf version depends on the call to SetVersion, it cannot be set
	// multiple times and is not thread-safe.
	re := regexp.MustCompile(`^Telegraf/[^\s]+ Go/\d+.\d+(.\d+)?$`)
	require.True(t, re.MatchString(token), token)
}
