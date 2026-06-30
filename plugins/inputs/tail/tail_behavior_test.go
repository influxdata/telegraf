package tail

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

// The tests in this file exercise the tailer engine itself (follow, rotation,
// truncation, deletion, partial lines, symlinks, offset resume) rather than the
// parsing glue. Every scenario runs under both the native watcher and polling
// via forEachWatchMethod so coverage is independent of the watch implementation,
// and acts as an acceptance gate for any replacement of the underlying library.
// Behaviors the current library does not satisfy are kept as skipped targets a
// replacement should clear.

func TestTailFollowAppend(t *testing.T) {
	forEachWatchMethod(t, nil, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		require.NoError(t, os.WriteFile(f, []byte("cpu value=1\n"), 0600))

		plugin := newBehaviorTail(t, watchMethod, "beginning", f)
		var acc testutil.Accumulator
		require.NoError(t, plugin.Start(&acc))
		defer plugin.Stop()

		waitForValues(t, &acc, 1)

		appendLine(t, f, "cpu value=2\n")
		waitForValues(t, &acc, 1, 2)
	})
}

func TestTailRotateRenameRecreate(t *testing.T) {
	skip := func(watchMethod string) string {
		if runtime.GOOS == "windows" {
			return "rename and recreate is not detected under polling on Windows"
		}
		if watchMethod == "inotify" && runtime.GOOS == "darwin" {
			// The kqueue-based native watcher drops rename events under
			// concurrent load, so this case runs under polling locally
			// and through inotify on Linux CI.
			return "macOS native watcher drops rotation/deletion events under load"
		}
		return ""
	}
	forEachWatchMethod(t, skip, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		require.NoError(t, os.WriteFile(f, []byte("cpu value=1\n"), 0600))

		plugin := newBehaviorTail(t, watchMethod, "beginning", f)
		var acc testutil.Accumulator
		require.NoError(t, plugin.Start(&acc))
		defer plugin.Stop()

		waitForValues(t, &acc, 1)

		// Logrotate "create" style: move the active file aside and recreate it.
		require.NoError(t, os.Rename(f, f+".1"))
		require.NoError(t, os.WriteFile(f, []byte("cpu value=2\n"), 0600))

		waitForValues(t, &acc, 1, 2)
	})
}

func TestTailRotateCopytruncate(t *testing.T) {
	skip := func(string) string {
		// The library only treats a size shrink as a truncation, so a
		// truncate-and-rewrite back to the same length is missed under
		// both the native watcher and polling and the new content is
		// lost. A replacement should detect truncation reliably.
		return "current library misses same-length copytruncate"
	}
	forEachWatchMethod(t, skip, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		require.NoError(t, os.WriteFile(f, []byte("cpu value=1\n"), 0600))

		plugin := newBehaviorTail(t, watchMethod, "beginning", f)
		var acc testutil.Accumulator
		require.NoError(t, plugin.Start(&acc))
		defer plugin.Stop()

		waitForValues(t, &acc, 1)

		// Logrotate "copytruncate" style: truncate the file in place, then
		// write fresh content from offset zero.
		require.NoError(t, os.Truncate(f, 0))
		appendLine(t, f, "cpu value=2\n")

		waitForValues(t, &acc, 1, 2)
	})
}

func TestTailDeleteRecreate(t *testing.T) {
	skip := func(watchMethod string) string {
		if watchMethod == "inotify" {
			// The native watcher does not reopen a file that was deleted
			// and recreated at the same path; polling recovers via its
			// wait-for-existence loop. A replacement should reopen under
			// both watch methods.
			return "current library does not reopen a recreated file under the native watcher"
		}
		if runtime.GOOS == "windows" {
			return "delete and recreate is not detected under polling on Windows"
		}
		return ""
	}
	forEachWatchMethod(t, skip, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		require.NoError(t, os.WriteFile(f, []byte("cpu value=1\n"), 0600))

		plugin := newBehaviorTail(t, watchMethod, "beginning", f)
		var acc testutil.Accumulator
		require.NoError(t, plugin.Start(&acc))
		defer plugin.Stop()

		waitForValues(t, &acc, 1)

		// Remove the followed file entirely, then recreate it later.
		require.NoError(t, os.Remove(f))
		require.NoError(t, os.WriteFile(f, []byte("cpu value=2\n"), 0600))

		waitForValues(t, &acc, 1, 2)
	})
}

func TestTailPartialLine(t *testing.T) {
	forEachWatchMethod(t, nil, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		// A complete first line establishes the tailer at a non-zero offset.
		require.NoError(t, os.WriteFile(f, []byte("cpu value=1\n"), 0600))

		plugin := newBehaviorTail(t, watchMethod, "beginning", f)
		var acc testutil.Accumulator
		require.NoError(t, plugin.Start(&acc))
		defer plugin.Stop()

		waitForValues(t, &acc, 1)

		// Write a line with no terminator yet, then complete it. The
		// partial line must be held and only emitted once its newline
		// arrives; a premature emit would parse "cpu value=" as a
		// malformed line and the completed value would never form.
		appendLine(t, f, "cpu value=")
		appendLine(t, f, "2\n")
		waitForValues(t, &acc, 1, 2)
	})
}

func TestTailSymlink(t *testing.T) {
	skip := func(watchMethod string) string {
		if runtime.GOOS == "windows" {
			return "symlink creation requires elevated privileges on Windows"
		}
		if watchMethod == "inotify" && runtime.GOOS == "darwin" {
			// The native (kqueue/FSEvents) watcher registered on the
			// symlink path does not receive events for appends to the
			// symlink target. A replacement should follow the resolved
			// target on all platforms.
			return "current library misses target writes via the native watcher on macOS"
		}
		return ""
	}
	forEachWatchMethod(t, skip, func(t *testing.T, watchMethod, dir string) {
		target := filepath.Join(dir, "real.log")
		link := filepath.Join(dir, "link.log")
		require.NoError(t, os.WriteFile(target, []byte("cpu value=1\n"), 0600))
		require.NoError(t, os.Symlink(target, link))

		plugin := newBehaviorTail(t, watchMethod, "beginning", link)
		var acc testutil.Accumulator
		require.NoError(t, plugin.Start(&acc))
		defer plugin.Stop()

		waitForValues(t, &acc, 1)

		appendLine(t, target, "cpu value=2\n")
		waitForValues(t, &acc, 1, 2)
	})
}

func TestTailResumeOffset(t *testing.T) {
	forEachWatchMethod(t, nil, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		require.NoError(t, os.WriteFile(f, []byte("cpu value=1\ncpu value=2\n"), 0600))

		// First run reads the file and persists its offset on stop.
		first := newBehaviorTail(t, watchMethod, "saved-or-beginning", f)
		var firstAcc testutil.Accumulator
		require.NoError(t, first.Start(&firstAcc))
		waitForValues(t, &firstAcc, 1, 2)
		state := first.GetState()
		first.Stop()

		// New content arrives while no tailer is running.
		appendLine(t, f, "cpu value=3\n")

		// Second run resumes from the persisted offset and must only see
		// the new line, never re-reading the already-delivered ones.
		second := newBehaviorTail(t, watchMethod, "saved-or-beginning", f)
		require.NoError(t, second.SetState(state))
		var secondAcc testutil.Accumulator
		require.NoError(t, second.Start(&secondAcc))
		defer second.Stop()

		waitForValues(t, &secondAcc, 3)
		// Reading is sequential, so if 1 or 2 were going to reappear they
		// would arrive before 3; asserting their absence needs no extra wait.
		require.ElementsMatch(t, []float64{3}, tailFieldValues(&secondAcc))
	})
}

func TestTailResumeMidMultiline(t *testing.T) {
	skip := func(string) string {
		// On stop the offset is taken from tailer.Tell() (the last byte
		// read), so a stop in the middle of a multi-line entry persists an
		// offset past the buffered start lines. On resume the remaining
		// lines are read as a standalone fragment and the entry is
		// delivered truncated. A replacement should persist the offset of
		// the last fully delivered entry instead.
		return "current library persists read offset, not delivered offset, corrupting mid-multiline resume"
	}
	forEachWatchMethod(t, skip, func(t *testing.T, watchMethod, dir string) {
		f := filepath.Join(dir, "test.log")
		// First entry is complete; the second entry's start line is present
		// but its continuation has not arrived yet, so it is only buffered.
		run1 := "[04/Jun/2016:12:41:45 +0100] DEBUG firstline\n" +
			" firstcont\n" +
			"[04/Jun/2016:12:41:46 +0100] INFO secondline\n"
		require.NoError(t, os.WriteFile(f, []byte(run1), 0600))

		first := newMultilineTail(t, watchMethod, f)
		var firstAcc testutil.Accumulator
		require.NoError(t, first.Start(&firstAcc))

		// The first entry is emitted once the second entry's start line is read.
		require.Eventuallyf(t, func() bool {
			return containsEntry(&firstAcc, "DEBUG|firstline firstcont")
		}, 10*time.Second, 50*time.Millisecond, "entry1 not delivered, got %v", tailGrokEntries(&firstAcc))

		state := first.GetState()
		first.Stop()

		// The rest of the second entry plus a third entry arrive while stopped.
		appendLine(t, f, " secondcont\n[04/Jun/2016:12:41:47 +0100] WARN thirdline\n")

		second := newMultilineTail(t, watchMethod, f)
		require.NoError(t, second.SetState(state))
		var secondAcc testutil.Accumulator
		require.NoError(t, second.Start(&secondAcc))
		defer second.Stop()

		// The second entry must be delivered intact once its terminator is read.
		require.Eventuallyf(t, func() bool {
			return containsEntry(&secondAcc, "INFO|secondline secondcont")
		}, 10*time.Second, 50*time.Millisecond, "entry2 not delivered intact, got %v", tailGrokEntries(&secondAcc))

		// Stop flushes the trailing third entry.
		second.Stop()

		got := append(tailGrokEntries(&firstAcc), tailGrokEntries(&secondAcc)...)
		expected := []string{
			"DEBUG|firstline firstcont",
			"INFO|secondline secondcont",
			"WARN|thirdline",
		}
		require.ElementsMatch(t, expected, got)
	})
}

// forEachWatchMethod runs fn as a subtest under each supported watch method so a
// behavior is verified for both the native watcher and polling. skip returns a
// non-empty reason to skip the given watch method on the current platform (and
// an empty string to run it); pass nil to always run.
func forEachWatchMethod(t *testing.T, skip func(watchMethod string) string, fn func(t *testing.T, watchMethod, dir string)) {
	t.Helper()
	methods := []string{"inotify", "poll"}
	if runtime.GOOS == "windows" {
		methods = []string{"poll"}
	}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			if skip != nil {
				if reason := skip(method); reason != "" {
					t.Skip(reason)
				}
			}
			fn(t, method, t.TempDir())
		})
	}
}

func newBehaviorTail(t *testing.T, watchMethod, initialReadOffset string, files ...string) *Tail {
	t.Helper()
	plugin := newTestTail()
	plugin.Log = testutil.Logger{}
	plugin.WatchMethod = watchMethod
	plugin.InitialReadOffset = initialReadOffset
	plugin.Files = files
	plugin.SetParserFunc(newInfluxParser)
	require.NoError(t, plugin.Init())
	return plugin
}

func newMultilineTail(t *testing.T, watchMethod, file string) *Tail {
	t.Helper()
	timeout := config.Duration(100 * time.Second)
	plugin := newTestTail()
	plugin.Log = testutil.Logger{}
	plugin.WatchMethod = watchMethod
	plugin.InitialReadOffset = "saved-or-beginning"
	plugin.Files = []string{file}
	plugin.MultilineConfig = multilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: previous,
		Timeout:        &timeout,
	}
	plugin.SetParserFunc(createGrokParser)
	require.NoError(t, plugin.Init())
	return plugin
}

func appendLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	require.NoError(t, err)
	_, err = f.WriteString(line)
	require.NoError(t, err)
	require.NoError(t, f.Close())
}

func tailFieldValues(acc *testutil.Accumulator) []float64 {
	var out []float64
	for _, m := range acc.GetTelegrafMetrics() {
		if v, ok := m.Fields()["value"]; ok {
			if f, ok := v.(float64); ok {
				out = append(out, f)
			}
		}
	}
	return out
}

// tailGrokEntries returns the delivered multi-line entries as "loglevel|message"
// so a whole entry (start line plus its continuations) can be asserted as a unit.
func tailGrokEntries(acc *testutil.Accumulator) []string {
	metrics := acc.GetTelegrafMetrics()
	out := make([]string, 0, len(metrics))
	for _, m := range metrics {
		level, _ := m.GetTag("loglevel")
		var message string
		if v, ok := m.Fields()["message"]; ok {
			message, _ = v.(string)
		}
		out = append(out, level+"|"+message)
	}
	return out
}

func containsEntry(acc *testutil.Accumulator, entry string) bool {
	return slices.Contains(tailGrokEntries(acc), entry)
}

func waitForValues(t *testing.T, acc *testutil.Accumulator, want ...float64) {
	t.Helper()
	require.Eventuallyf(t, func() bool {
		got := make(map[float64]bool)
		for _, v := range tailFieldValues(acc) {
			got[v] = true
		}
		for _, w := range want {
			if !got[w] {
				return false
			}
		}
		return true
	}, 10*time.Second, 50*time.Millisecond, "did not observe values %v, got %v", want, tailFieldValues(acc))
}
