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
var maxH = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"
var maxA = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"
var maxPID = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"
var maxMID = "abcdefghilmnopqrstuvzabcdefghilm"
var message7681 = strings.Repeat("l", 7681)

func TestAddress(t *testing.T) {
	var err error
	var rec *Syslog

	rec = &Syslog{
		Address: "localhost:6514",
	}
	err = rec.Start(&testutil.Accumulator{})
	require.EqualError(t, err, "missing protocol within address 'localhost:6514'")
	require.Error(t, err)

	rec = &Syslog{
		Address: "unsupported://example.com:6514",
	}
	err = rec.Start(&testutil.Accumulator{})
	require.EqualError(t, err, "unknown protocol 'unsupported' in 'example.com:6514'")
	require.Error(t, err)

	tmpdir := t.TempDir()
	sock := filepath.Join(tmpdir, "syslog.TestAddress.sock")

	if runtime.GOOS != "windows" {
		// Skipping on Windows, as unixgram sockets are not supported
		rec = &Syslog{
			Address: "unixgram://" + sock,
		}
		err = rec.Start(&testutil.Accumulator{})
		require.NoError(t, err)
		require.Equal(t, sock, rec.Address)
		rec.Stop()
	}

	// Default port is 6514
	rec = &Syslog{
		Address: "tcp://localhost",
	}
	err = rec.Start(&testutil.Accumulator{})
	require.NoError(t, err)
	require.Equal(t, "localhost:6514", rec.Address)
	rec.Stop()
}
