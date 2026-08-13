package internet_speed

import (
	"net"
	"testing"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestSelectServer(t *testing.T) {
	tests := []struct {
		name       string
		include    []string
		exclude    []string
		servers    speedtest.Servers
		expectedID string
	}{
		{
			name: "no filter selects lowest latency",
			servers: speedtest.Servers{
				{ID: "1", Latency: 50 * time.Millisecond},
				{ID: "2", Latency: 10 * time.Millisecond},
				{ID: "3", Latency: 30 * time.Millisecond},
			},
			expectedID: "2",
		},
		{
			name:    "include filter selects matching server",
			include: []string{"3"},
			servers: speedtest.Servers{
				{ID: "1", Latency: 10 * time.Millisecond},
				{ID: "2", Latency: 20 * time.Millisecond},
				{ID: "3", Latency: 30 * time.Millisecond},
			},
			expectedID: "3",
		},
		{
			name:    "include filter with multiple matches selects lowest latency",
			include: []string{"2", "3"},
			servers: speedtest.Servers{
				{ID: "1", Latency: 5 * time.Millisecond},
				{ID: "2", Latency: 30 * time.Millisecond},
				{ID: "3", Latency: 15 * time.Millisecond},
			},
			expectedID: "3",
		},
		{
			name:    "exclude filter skips excluded server",
			exclude: []string{"1"},
			servers: speedtest.Servers{
				{ID: "1", Latency: 5 * time.Millisecond},
				{ID: "2", Latency: 10 * time.Millisecond},
				{ID: "3", Latency: 30 * time.Millisecond},
			},
			expectedID: "2",
		},
		{
			name:    "no latency info selects first matching server",
			include: []string{"2", "3"},
			servers: speedtest.Servers{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
			},
			expectedID: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &InternetSpeed{
				ServerIDInclude: tt.include,
				ServerIDExclude: tt.exclude,
				Log:             testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			plugin.servers = tt.servers

			require.NoError(t, plugin.selectServer())
			require.Equal(t, tt.expectedID, plugin.server.ID)
		})
	}
}

func TestSelectServerError(t *testing.T) {
	plugin := &InternetSpeed{
		ServerIDInclude: []string{"99"},
		Log:             testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	plugin.servers = speedtest.Servers{
		{ID: "1", Latency: 10 * time.Millisecond},
		{ID: "2", Latency: 20 * time.Millisecond},
	}

	require.ErrorContains(t, plugin.selectServer(), "filter excluded all servers")
}

func TestLocalAddress(t *testing.T) {
	tests := []struct {
		name         string
		localAddress string
		expected     *net.TCPAddr
	}{
		{
			name:     "unset",
			expected: nil,
		},
		{
			name:         "IPv4 address",
			localAddress: "127.0.0.1",
			expected:     &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)},
		},
		{
			name:         "IPv6 address",
			localAddress: "::1",
			expected:     &net.TCPAddr{IP: net.IPv6loopback},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &InternetSpeed{
				LocalAddress: tt.localAddress,
				Log:          testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			require.Equal(t, tt.expected, plugin.localAddr)
		})
	}
}

func TestLocalAddressInvalid(t *testing.T) {
	tests := []struct {
		name         string
		localAddress string
	}{
		{
			// The underlying library only accepts an interface name on Linux and
			// silently falls back to the default route everywhere else
			name:         "interface name",
			localAddress: "eth0",
		},
		{
			name:         "address with port",
			localAddress: "127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &InternetSpeed{
				LocalAddress: tt.localAddress,
				Log:          testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), "resolving local address failed")
		})
	}
}

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	internetSpeed := &InternetSpeed{
		MemorySavingMode: true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, internetSpeed.Init())

	var acc testutil.Accumulator
	require.NoError(t, internetSpeed.Gather(&acc))
}

func TestDataGen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	internetSpeed := &InternetSpeed{
		MemorySavingMode: true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, internetSpeed.Init())

	var acc testutil.Accumulator
	require.NoError(t, internetSpeed.Gather(&acc))

	metric, ok := acc.Get("internet_speed")
	require.True(t, ok)
	acc.AssertContainsTaggedFields(t, "internet_speed", metric.Fields, metric.Tags)
}
