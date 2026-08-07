package tail

import (
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/grok"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestAppend(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Add more data to the file
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("cpu value=2i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestPartialWrite(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content with a partial second line
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\ncpu "), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Add the remaining data for the second line to the file
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("value=2i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestLogRotateCreate(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			switch {
			case runtime.GOOS == "windows" && method == "inotify":
				t.Skip("Windows does not support inotify!")
			case runtime.GOOS == "windows" && method == "poll":
				// Should potentially be supported in the plugin
				t.Skip("[to be fixed] Renaming and recreating is not detected with polling on Windows!")
			case runtime.GOOS == "darwin" && method == "inotify":
				// The kqueue-based native watcher drops rename events under
				// concurrent load, so this case runs under polling locally
				// and through inotify on Linux CI.
				t.Skip("macOS native watcher drops rotation and deletion events under load!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Logrotate "create" style: move the active file aside and recreate it with the same name
			require.NoError(t, os.Rename(fn, fn+".1"))
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=2i\n"), 0600))

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestLogRotateCopyTruncateSmaller(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(3)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\ncpu value=2i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Logrotate "copytruncate" style: truncate the file in place, then
			// write fresh content from offset zero.
			buf, err := os.ReadFile(fn)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(fn+".1", buf, 0600))
			require.NoError(t, os.Truncate(fn, 0))
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("cpu value=3i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestLogRotateCopytruncateSameSize(t *testing.T) {
	t.Skip("[to be fixed] Plugin currently does not detect truncation if new file has same size!")

	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Logrotate "copytruncate" style: truncate the file in place, then
			// write fresh content from offset zero.
			buf, err := os.ReadFile(fn)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(fn+".1", buf, 0600))
			require.NoError(t, os.Truncate(fn, 0))
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("cpu value=2i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestLogRotateCopytruncateLarger(t *testing.T) {
	t.Skip("[to be fixed] Plugin currently does not detect truncation if new file is larger!")

	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(99)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Logrotate "copytruncate" style: truncate the file in place, then
			// write fresh content from offset zero.
			buf, err := os.ReadFile(fn)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(fn+".1", buf, 0600))
			require.NoError(t, os.Truncate(fn, 0))
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("cpu value=99i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestDelete(t *testing.T) {
	t.Skip("[to be fixed] Plugin is currently unable to read the remaining data of the deleted file!")

	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Add some more data to the file and delete the file immediately afterwards
			f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("cpu value=2i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())
			require.NoError(t, os.Remove(fn))

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestDeleteRecreate(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}
			if runtime.GOOS == "windows" && method == "poll" {
				t.Skip("[to be fixed] Plugin currently doesn't detect rename and recreate under Windows!")
			}

			if method == "inotify" {
				// The native watcher does not reopen a file that was deleted
				// and recreated at the same path; polling recovers via its
				// wait-for-existence loop. A replacement should reopen under
				// both watch methods.
				t.Skip("[to be fixed] Plugin currently doesn't detect deletion and recreation with inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": fn},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\n"), 0600))

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Remove tailed file entirely, then recreate it with new content
			require.NoError(t, os.Remove(fn))
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=2i\n"), 0600))

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestSymlink(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}
			if runtime.GOOS == "windows" {
				t.Skip("Windows requires elevated privileges for symlink creation")
			}
			if runtime.GOOS == "darwin" && method == "inotify" {
				// The native (kqueue/FSEvents) watcher registered on the
				// symlink path does not receive events for appends to the
				// symlink target. A replacement should follow the resolved
				// target on all platforms.
				t.Skip("[to be fixed] Plugin currently misses target writes via the native watcher on macOS!")
			}

			// Create the initial file content and link via symlink which we
			// will follow later
			dir := t.TempDir()
			target := filepath.Join(dir, "test.log")
			link := filepath.Join(dir, "link.log")
			require.NoError(t, os.WriteFile(target, []byte("cpu value=1i\n"), 0600))
			require.NoError(t, os.Symlink(target, link))

			// Define the expected metrics in each step
			expected := [][]telegraf.Metric{
				{
					// Initial step
					metric.New(
						"cpu",
						map[string]string{"path": link},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
				},
				{
					// After append
					metric.New(
						"cpu",
						map[string]string{"path": link},
						map[string]interface{}{"value": int64(1)},
						time.Unix(0, 0),
					),
					metric.New(
						"cpu",
						map[string]string{"path": link},
						map[string]interface{}{"value": int64(2)},
						time.Unix(0, 0),
					),
				},
			}

			// Setup the plugin and start it
			plugin := &Tail{
				Files:               []string{link},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics of the initial reading
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[0]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[0], acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Add more data to the file
			f, err := os.OpenFile(target, os.O_APPEND|os.O_WRONLY, 0600)
			require.NoError(t, err)
			defer f.Close()
			_, err = f.WriteString("cpu value=2i\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Check the metrics after appending
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected[1]))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected[1], acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestResumeFromPreviousOffset(t *testing.T) {
	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{"path": fn},
					map[string]interface{}{"value": int64(3)},
					time.Unix(0, 0),
				),
			}

			// Create the initial file content
			require.NoError(t, os.WriteFile(fn, []byte("cpu value=1i\ncpu value=2i\ncpu value=3i\n"), 0600))

			// Setup the plugin
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "saved-or-beginning",
				WatchMethod:         method,
				PathTag:             "path",
				Log:                 testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			// Set the initial state to after the first two lines so we should
			// only see the last line
			require.NoError(t, plugin.SetState(map[string]int64{fn: 24}))

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics read
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestStopIncompleteMultiline(t *testing.T) {
	// On stop the offset is taken from tailer.Tell() (the last byte read), so
	// a stop in the middle of a multi-line entry persists an offset past the
	// buffered start lines. On resume the remaining lines are read as a
	// standalone fragment and the entry is delivered truncated. A replacement
	// should persist the offset of the last fully delivered entry instead.
	t.Skip("[to be fixed] Plugin currently persists read offset, not delivered offset, corrupting mid-multiline resume!")

	for _, method := range []string{"inotify", "poll"} {
		t.Run(method, func(t *testing.T) {
			if runtime.GOOS == "windows" && method == "inotify" {
				t.Skip("Windows does not support inotify!")
			}

			// Get a temporary filename for testing
			fn := filepath.Join(t.TempDir(), "test.log")

			// Define the expected metrics in each step
			expected := []telegraf.Metric{
				metric.New(
					"tail_grok",
					map[string]string{
						"path":     fn,
						"loglevel": "DEBUG",
					},
					map[string]interface{}{
						"message": "firstline firstcont",
					},
					time.Unix(0, 0),
				),
			}

			// Create the initial file content with multi-line content
			// The first entry is complete but the second entry is incomplete so
			// it cannot be fully processed yet.
			data := "[04/Jun/2016:12:41:45 +0100] DEBUG firstline\n" +
				" firstcont\n" +
				"[04/Jun/2016:12:41:46 +0100] INFO secondline\n"
			require.NoError(t, os.WriteFile(fn, []byte(data), 0600))

			// Setup the plugin
			plugin := &Tail{
				Files:               []string{fn},
				MaxUndeliveredLines: 1000,
				InitialReadOffset:   "beginning",
				WatchMethod:         method,
				PathTag:             "path",
				MultilineConfig: multilineConfig{
					Pattern:        `^[^\[]`,
					MatchWhichLine: previous,
					Timeout:        new(config.Duration(100 * time.Second)),
				},
				Log: testutil.Logger{},

				offsets: maps.Clone(offsets),
			}
			plugin.SetParserFunc(func() (telegraf.Parser, error) {
				parser := &grok.Parser{
					Measurement:        "tail_grok",
					Patterns:           []string{"%{TEST_LOG_MULTILINE}"},
					CustomPatternFiles: []string{filepath.Join("testdata", "test-patterns")},
					Log:                testutil.Logger{},
				}
				err := parser.Init()
				return parser, err
			})
			require.NoError(t, plugin.Init())

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Check the metrics read
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())

			// Make sure that the returned state marks the end of the fully
			// processed metric
			plugin.Stop()
			state := plugin.GetState()
			require.Equal(t, map[string]int64{fn: 45}, state)
		})
	}
}
