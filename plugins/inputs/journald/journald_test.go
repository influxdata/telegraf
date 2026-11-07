//go:build linux

package journald

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

// Environment variables controlling the journalctl mock, see TestMain
const (
	envEntries     = "JOURNALD_MOCK_ENTRIES"
	envInvocations = "JOURNALD_MOCK_INVOCATIONS"
	envExit        = "JOURNALD_MOCK_EXIT"
)

func TestCases(t *testing.T) {
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	inputs.Add("journald", func() telegraf.Input {
		return &Journald{}
	})

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())

		t.Run(f.Name(), func(t *testing.T) {
			expected, err := testutil.ParseMetricsFromFile(filepath.Join(testcasePath, "expected.out"), parser)
			require.NoError(t, err)

			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(filepath.Join(testcasePath, "telegraf.conf")))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Journald)

			mock(t, filepath.Join(testcasePath, "entries.json"), false)
			plugin.Binary = mockBinary(t)
			plugin.Log = &testutil.Logger{}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected))
			}, 5*time.Second, 10*time.Millisecond)
			require.Empty(t, acc.Errors)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.SortMetrics())
		})
	}
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Journald
		expected string
	}{
		{
			name:     "invalid offset",
			plugin:   &Journald{InitialReadOffset: "middle"},
			expected: `invalid 'initial_read_offset' setting "middle"`,
		},
		{
			name:     "invalid priority name",
			plugin:   &Journald{Priority: "fatal"},
			expected: `invalid priority "fatal"`,
		},
		{
			name:     "invalid priority number",
			plugin:   &Journald{Priority: "8"},
			expected: `invalid priority "8"`,
		},
		{
			name:     "invalid priority range",
			plugin:   &Journald{Priority: "err..warning..info"},
			expected: `invalid priority "err..warning..info"`,
		},
		{
			name:     "negative undelivered entries",
			plugin:   &Journald{MaxUndeliveredEntries: -1},
			expected: "'max_undelivered_entries' must be positive",
		},
		{
			name:     "negative restart delay",
			plugin:   &Journald{RestartDelay: config.Duration(-time.Second)},
			expected: "'restart_delay' must not be negative",
		},
		{
			name:     "empty match",
			plugin:   &Journald{Matches: []string{"_TRANSPORT=kernel", " "}},
			expected: "empty entry in 'matches'",
		},
		{
			name:     "empty tag field",
			plugin:   &Journald{TagFields: []string{""}},
			expected: "empty entry in 'tag_fields'",
		},
		{
			name:     "missing binary",
			plugin:   &Journald{Binary: "/non/existing/journalctl"},
			expected: `looking up "/non/existing/journalctl" failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.plugin.Binary == "" {
				tt.plugin.Binary = mockBinary(t)
			}
			tt.plugin.Log = &testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestArguments(t *testing.T) {
	base := []string{"--output=json", "--all", "--follow", "--no-pager"}
	tests := []struct {
		name     string
		plugin   *Journald
		cursor   string
		expected []string
	}{
		{
			name:     "defaults",
			plugin:   &Journald{},
			expected: append(slices.Clone(base), "--lines=0"),
		},
		{
			name:     "beginning",
			plugin:   &Journald{InitialReadOffset: "saved-or-beginning"},
			expected: append(slices.Clone(base), "--lines=all"),
		},
		{
			name:     "cursor",
			plugin:   &Journald{InitialReadOffset: "beginning"},
			cursor:   "s=1;i=2",
			expected: append(slices.Clone(base), "--after-cursor=s=1;i=2"),
		},
		{
			name: "filters",
			plugin: &Journald{
				Units:     []string{"nginx.service", "docker-*"},
				UserUnits: []string{"backup.service"},
				Priority:  "err..notice",
				Directory: "/var/log/journal/remote",
				Namespace: "app",
				Matches:   []string{"_TRANSPORT=kernel", "+", "_COMM=sshd"},
			},
			expected: append(slices.Clone(base),
				"--unit=nginx.service",
				"--unit=docker-*",
				"--user-unit=backup.service",
				"--priority=err..notice",
				"--directory=/var/log/journal/remote",
				"--namespace=app",
				"--lines=0",
				"_TRANSPORT=kernel", "+", "_COMM=sshd",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Binary = mockBinary(t)
			tt.plugin.Log = &testutil.Logger{}
			require.NoError(t, tt.plugin.Init())
			tt.plugin.readCursor = tt.cursor
			require.Equal(t, tt.expected, tt.plugin.args())
		})
	}
}

func TestSudo(t *testing.T) {
	plugin := &Journald{
		Binary:  "/usr/bin/journalctl",
		UseSudo: true,
		Log:     &testutil.Logger{},
	}
	if err := plugin.Init(); err != nil {
		t.Skipf("sudo not available: %v", err)
	}
	require.Equal(t, "sudo", plugin.command)
	require.Equal(t, "/usr/bin/journalctl", plugin.args()[0])
}

func TestStartAtEnd(t *testing.T) {
	invocations := mock(t, filepath.Join("testcases", "basic", "entries.json"), false)

	plugin := &Journald{
		Binary: mockBinary(t),
		Log:    &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	require.Eventually(t, func() bool {
		return len(readInvocations(t, invocations)) == 1
	}, 5*time.Second, 10*time.Millisecond)
	plugin.Stop()

	require.Contains(t, readInvocations(t, invocations)[0], "--lines=0")
	require.Empty(t, acc.GetTelegrafMetrics())
	require.Empty(t, acc.Errors)
}

func TestResumeFromState(t *testing.T) {
	entries := filepath.Join("testcases", "basic", "entries.json")
	cursors := readCursors(t, entries)

	for _, offset := range []string{"saved-or-beginning", "saved-or-end"} {
		t.Run(offset, func(t *testing.T) {
			invocations := mock(t, entries, false)

			plugin := &Journald{
				Binary:            mockBinary(t),
				InitialReadOffset: offset,
				Log:               &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.SetState(cursors[0]))

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Only the entries after the persisted cursor are read
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= 2
			}, 5*time.Second, 10*time.Millisecond)
			require.Equal(t, []string{"--after-cursor=" + cursors[0]}, cursorArgs(readInvocations(t, invocations)[0]))
			metrics := acc.GetTelegrafMetrics()
			require.Len(t, metrics, 2)
			require.Equal(t, "kernel", metrics[0].Tags()["identifier"])

			// The state only changes once the entries are delivered
			require.Equal(t, cursors[0], plugin.GetState())
			for _, m := range metrics {
				m.Accept()
			}
			require.Eventually(t, func() bool {
				return plugin.GetState() == cursors[2]
			}, 5*time.Second, 10*time.Millisecond)
		})
	}
}

func TestStateIgnored(t *testing.T) {
	entries := filepath.Join("testcases", "basic", "entries.json")
	invocations := mock(t, entries, false)

	plugin := &Journald{
		Binary:            mockBinary(t),
		InitialReadOffset: "beginning",
		Log:               &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.SetState(readCursors(t, entries)[1]))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= 3
	}, 5*time.Second, 10*time.Millisecond)
	args := readInvocations(t, invocations)[0]
	require.Contains(t, args, "--lines=all")
	require.Empty(t, cursorArgs(args))
}

func TestStateAdvancesInOrder(t *testing.T) {
	entries := filepath.Join("testcases", "basic", "entries.json")
	cursors := readCursors(t, entries)
	mock(t, entries, false)

	plugin := &Journald{
		Binary:            mockBinary(t),
		InitialReadOffset: "beginning",
		Log:               &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= 3
	}, 5*time.Second, 10*time.Millisecond)
	metrics := acc.GetTelegrafMetrics()
	require.Empty(t, plugin.GetState())

	// Delivering a later entry must not advance the cursor past an entry
	// still in flight
	metrics[2].Accept()
	metrics[0].Reject()
	require.Eventually(t, func() bool {
		return plugin.GetState() == cursors[0]
	}, 5*time.Second, 10*time.Millisecond)
	require.Never(t, func() bool {
		return plugin.GetState() != cursors[0]
	}, 100*time.Millisecond, 10*time.Millisecond)

	metrics[1].Accept()
	require.Eventually(t, func() bool {
		return plugin.GetState() == cursors[2]
	}, 5*time.Second, 10*time.Millisecond)
}

func TestMaxUndeliveredEntries(t *testing.T) {
	entries := filepath.Join("testcases", "basic", "entries.json")
	mock(t, entries, false)

	plugin := &Journald{
		Binary:                mockBinary(t),
		InitialReadOffset:     "beginning",
		MaxUndeliveredEntries: 2,
		Log:                   &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Reading blocks until an entry is delivered
	require.Eventually(t, func() bool {
		return acc.NMetrics() == 2
	}, 5*time.Second, 10*time.Millisecond)
	require.Never(t, func() bool {
		return acc.NMetrics() > 2
	}, 200*time.Millisecond, 10*time.Millisecond)

	acc.GetTelegrafMetrics()[0].Accept()
	require.Eventually(t, func() bool {
		return acc.NMetrics() == 3
	}, 5*time.Second, 10*time.Millisecond)
}

func TestRestartContinuesAfterLastEntry(t *testing.T) {
	entries := filepath.Join("testcases", "basic", "entries.json")
	cursors := readCursors(t, entries)
	invocations := mock(t, entries, true)

	plugin := &Journald{
		Binary:            mockBinary(t),
		InitialReadOffset: "beginning",
		RestartDelay:      config.Duration(10 * time.Millisecond),
		Log:               &testutil.Logger{Quiet: true},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Eventually(t, func() bool {
		return len(readInvocations(t, invocations)) >= 3
	}, 5*time.Second, 10*time.Millisecond)

	// The restarted journalctl continues after the last entry read even
	// though nothing was delivered yet, so no entry is duplicated.
	calls := readInvocations(t, invocations)
	require.Contains(t, calls[0], "--lines=all")
	require.Equal(t, []string{"--after-cursor=" + cursors[2]}, cursorArgs(calls[1]))
	require.Equal(t, []string{"--after-cursor=" + cursors[2]}, cursorArgs(calls[2]))
	require.Len(t, acc.GetTelegrafMetrics(), 3)
}

func TestInvalidCursor(t *testing.T) {
	entries := filepath.Join("testcases", "basic", "entries.json")
	invocations := mock(t, entries, false)

	plugin := &Journald{
		Binary:            mockBinary(t),
		InitialReadOffset: "saved-or-beginning",
		RestartDelay:      config.Duration(10 * time.Millisecond),
		Log:               &testutil.Logger{Quiet: true},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.SetState("invalid"))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// journalctl fails to seek to the cursor so the plugin falls back to the
	// initial position
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= 3
	}, 5*time.Second, 10*time.Millisecond)
	calls := readInvocations(t, invocations)
	require.Len(t, calls, 2)
	require.Equal(t, []string{"--after-cursor=invalid"}, cursorArgs(calls[0]))
	require.Contains(t, calls[1], "--lines=all")
	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "journalctl failed")
}

func TestInvalidEntry(t *testing.T) {
	entries := filepath.Join(t.TempDir(), "entries.json")
	basic, err := os.ReadFile(filepath.Join("testcases", "basic", "entries.json"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(entries, append([]byte("{\"MESSAGE\": broken\n"), basic...), 0o600))
	mock(t, entries, false)

	plugin := &Journald{
		Binary:            mockBinary(t),
		InitialReadOffset: "beginning",
		Log:               &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// The broken entry is reported and skipped
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= 3
	}, 5*time.Second, 10*time.Millisecond)
	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "parsing journal entry failed")
}

func TestDecodeValue(t *testing.T) {
	tests := []struct {
		raw      string
		expected string
		ok       bool
		err      bool
	}{
		{raw: `"plain"`, expected: "plain", ok: true},
		{raw: `""`, expected: "", ok: true},
		{raw: `null`},
		{raw: `[104,105]`, expected: "hi", ok: true},
		{raw: `[]`, expected: "", ok: true},
		{raw: `["first","second"]`, expected: "first", ok: true},
		{raw: `[[104,105],"second"]`, expected: "hi", ok: true},
		{raw: `[null,"second"]`},
		{raw: `[300]`, err: true},
		{raw: `42`, err: true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			v, ok, err := decodeValue(json.RawMessage(tt.raw))
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.expected, v)
		})
	}
}

func TestCursorTracker(t *testing.T) {
	tr := newCursorTracker()
	tr.setCommitted("c0")

	tr.add(1, "c1")
	tr.add(2, "")
	tr.add(3, "c3")
	require.Equal(t, "c0", tr.committed())

	// Out-of-order delivery
	tr.done(3)
	require.Equal(t, "c0", tr.committed())
	tr.done(1)
	require.Equal(t, "c1", tr.committed())
	tr.done(2)
	require.Equal(t, "c3", tr.committed())

	// Delivery reported before the entry is registered
	tr.done(4)
	tr.add(4, "c4")
	require.Equal(t, "c4", tr.committed())
	require.Empty(t, tr.pending)
	require.Empty(t, tr.index)
	require.Empty(t, tr.early)
}

func BenchmarkParse(b *testing.B) {
	plugin := &Journald{
		Binary:        mockBinary(b),
		TagFields:     []string{"_BOOT_ID"},
		JournalFields: []string{"_EXE", "_COMM"},
		Log:           &testutil.Logger{},
	}
	require.NoError(b, plugin.Init())

	f, err := os.Open(filepath.Join("testcases", "basic", "entries.json"))
	require.NoError(b, err)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	require.True(b, scanner.Scan())
	line := slices.Clone(scanner.Bytes())

	b.SetBytes(int64(len(line)))
	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := plugin.parse(line); err != nil {
			b.Fatal(err)
		}
	}
}

// mock configures the journalctl mock for the current test and returns the
// file recording its invocations
func mock(t *testing.T, entries string, exit bool) string {
	t.Helper()
	abs, err := filepath.Abs(entries)
	require.NoError(t, err)
	invocations := filepath.Join(t.TempDir(), "invocations")
	t.Setenv(envEntries, abs)
	t.Setenv(envInvocations, invocations)
	if exit {
		t.Setenv(envExit, "true")
	} else {
		t.Setenv(envExit, "")
	}
	return invocations
}

func mockBinary(t testing.TB) string {
	t.Helper()
	exe, err := os.Executable()
	require.NoError(t, err)
	return exe
}

var invocationsMu sync.Mutex

func readInvocations(t *testing.T, filename string) [][]string {
	t.Helper()
	invocationsMu.Lock()
	defer invocationsMu.Unlock()

	buf, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)

	var calls [][]string
	for line := range strings.Lines(string(buf)) {
		var args []string
		require.NoError(t, json.Unmarshal([]byte(line), &args))
		calls = append(calls, args)
	}
	return calls
}

func readCursors(t *testing.T, filename string) []string {
	t.Helper()
	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	var cursors []string
	for line := range strings.Lines(string(buf)) {
		var entry struct {
			Cursor string `json:"__CURSOR"`
		}
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		cursors = append(cursors, entry.Cursor)
	}
	return cursors
}

// cursorArgs returns the --after-cursor arguments of a journalctl invocation
func cursorArgs(args []string) []string {
	var filtered []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--after-cursor=") {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// TestMain runs the test binary as a journalctl mock if the corresponding
// environment variables are set. The mock records its arguments, serves the
// entries of the configured file according to the position arguments and
// then either exits or, like journalctl --follow, waits until terminated.
func TestMain(m *testing.M) {
	entries := os.Getenv(envEntries)
	if entries == "" {
		os.Exit(m.Run())
	}
	os.Exit(runMock(entries, os.Getenv(envInvocations), os.Getenv(envExit) != ""))
}

func runMock(entries, invocations string, exit bool) int {
	args := os.Args[1:]

	// Record the invocation
	buf, err := json.Marshal(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	f, err := os.OpenFile(invocations, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	_, err = f.Write(append(buf, '\n'))
	f.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, required := range []string{"--output=json", "--all", "--follow", "--no-pager"} {
		if !slices.Contains(args, required) {
			fmt.Fprintf(os.Stderr, "missing argument %q\n", required)
			return 1
		}
	}

	content, err := os.ReadFile(entries)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	lines := slices.Collect(strings.Lines(string(content)))

	// Determine the starting position
	start := len(lines)
	for _, arg := range args {
		switch {
		case arg == "--lines=all":
			start = 0
		case strings.HasPrefix(arg, "--after-cursor="):
			cursor := strings.TrimPrefix(arg, "--after-cursor=")
			if !strings.HasPrefix(cursor, "s=") {
				fmt.Fprintln(os.Stderr, "Failed to seek to cursor: Invalid argument")
				return 1
			}
			for i, line := range lines {
				if strings.Contains(line, `"__CURSOR":"`+cursor+`"`) {
					start = i + 1
					break
				}
			}
		}
	}
	for _, line := range lines[start:] {
		fmt.Print(line)
	}

	if exit {
		return 0
	}

	// Follow mode, wait until terminated
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	<-signals
	return 0
}
