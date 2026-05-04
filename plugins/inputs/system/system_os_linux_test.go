//go:build linux

package system

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const testOSRelease = `NAME="Telegraf Test OS"
ID=telegraftest
VERSION_ID="1.0"
PRETTY_NAME="Telegraf Test OS 1.0"
`

// setupOS points gopsutil at a synthetic os-release file via HOST_ETC.
// Kernel fields still come from the live uname syscall.
func setupOS(t testing.TB) bool {
	t.Helper()
	mockOSRelease(t, testOSRelease)
	return true
}

func mockOSRelease(t testing.TB, content string) string {
	t.Helper()
	etcDir := os.Getenv("HOST_ETC")
	if etcDir == "" {
		etcDir = filepath.Join(t.TempDir(), "etc")
		require.NoError(t, os.MkdirAll(etcDir, 0750))
		t.Setenv("HOST_ETC", etcDir)
	}
	writeOSRelease(t, etcDir, content)
	return etcDir
}

func writeOSRelease(t testing.TB, etcDir, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(etcDir, "os-release"), []byte(content), 0640))
}

func newOSPlugin(ttl time.Duration) *System {
	return &System{
		Include:    []string{"os"},
		OSCacheTTL: config.Duration(ttl),
		Log:        &testutil.Logger{},
	}
}

func TestGatherOSValuesLinux(t *testing.T) {
	setupOS(t)

	s := newOSPlugin(defaultOSCacheTTL)
	require.NoError(t, s.Init())

	var acc testutil.Accumulator
	require.NoError(t, s.Gather(&acc))

	m, found := acc.Get("system_os")
	require.True(t, found, "system_os metric not produced")

	require.Equal(t, "linux", m.Fields["os"])
	require.Equal(t, "telegraftest", m.Fields["platform"])
	require.Equal(t, "", m.Fields["platform_family"])
	require.Equal(t, "1.0", m.Fields["platform_version"])
	require.IsType(t, "", m.Fields["kernel_version"])
	require.NotEmpty(t, m.Fields["kernel_version"])
	require.IsType(t, "", m.Fields["kernel_arch"])
	require.NotEmpty(t, m.Fields["kernel_arch"])
}

func TestGatherOSMissingOSReleaseLinux(t *testing.T) {
	t.Setenv("HOST_ETC", t.TempDir())

	s := newOSPlugin(defaultOSCacheTTL)
	require.NoError(t, s.Init())

	var acc testutil.Accumulator
	require.NoError(t, s.Gather(&acc))
}

func TestGatherOSCacheLinux(t *testing.T) {
	tests := []struct {
		name             string
		ttl              time.Duration
		sleep            time.Duration
		expectedPlatform string
	}{
		{
			name:             "refresh after expiry",
			ttl:              time.Millisecond,
			sleep:            5 * time.Millisecond,
			expectedPlatform: "upgraded",
		},
		{
			name:             "forever with zero ttl",
			ttl:              0,
			sleep:            5 * time.Millisecond,
			expectedPlatform: "telegraftest",
		},
		{
			name:             "refresh on every gather with tiny ttl",
			ttl:              time.Nanosecond,
			expectedPlatform: "upgraded",
		},
		{
			name:             "served from cache within ttl",
			ttl:              defaultOSCacheTTL,
			expectedPlatform: "telegraftest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupOS(t)

			s := newOSPlugin(tt.ttl)
			require.NoError(t, s.Init())

			var acc testutil.Accumulator
			require.NoError(t, s.Gather(&acc))

			mockOSRelease(t, "ID=upgraded\nVERSION_ID=\"2.0\"\n")
			if tt.sleep > 0 {
				time.Sleep(tt.sleep)
			}

			acc.ClearMetrics()
			require.NoError(t, s.Gather(&acc))

			m, found := acc.Get("system_os")
			require.True(t, found)
			require.Equal(t, tt.expectedPlatform, m.Fields["platform"])
		})
	}
}

func BenchmarkGatherOS(b *testing.B) {
	setupOS(b)

	s := newOSPlugin(defaultOSCacheTTL)
	require.NoError(b, s.Init())

	var acc testutil.Accumulator
	for b.Loop() {
		acc.ClearMetrics()
		if err := s.Gather(&acc); err != nil {
			b.Fatal(err)
		}
	}
}
