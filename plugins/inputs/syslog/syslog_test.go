package syslog

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

const (
	address = ":6514"
)

var defaultTime = time.Unix(0, 0)
var maxP = uint8(191)
var maxV = uint16(999)
var maxTS = "2017-12-31T23:59:59.999999+00:00"
var maxH = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqr" +
	"stuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"
var maxA = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"
var maxPID = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"
var maxMID = "abcdefghilmnopqrstuvzabcdefghilm"
var message7681 = strings.Repeat("l", 7681)

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "no address",
			expected: "missing protocol within address",
		},
		{
			name:     "missing protocol",
			address:  "localhost:6514",
			expected: "missing protocol within address",
		},
		{
			name:     "unknown protocol",
			address:  "unsupported://example.com:6514",
			expected: "unknown protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Syslog{
				Address: tt.address,
			}
			var acc testutil.Accumulator
			require.ErrorContains(t, plugin.Start(&acc), tt.expected)
		})
	}
}

func TestAddressUnixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test as unixgram is not supported on Windows")
	}

	sock := filepath.Join(t.TempDir(), "syslog.TestAddress.sock")
	plugin := &Syslog{
		Address: "unixgram://" + sock,
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Equal(t, sock, plugin.Address)
}

func TestAddressDefaultPort(t *testing.T) {
	plugin := &Syslog{
		Address: "tcp://localhost",
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Default port is 6514
	require.Equal(t, "localhost:6514", plugin.Address)
}
