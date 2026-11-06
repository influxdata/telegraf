//go:build !windows

package logger

import (
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newSyslogListener(t *testing.T) net.PacketConn {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return conn
}

func readSyslogMessage(t *testing.T, conn net.PacketConn, timeout time.Duration) (string, error) {
	t.Helper()

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(timeout)))
	buf := make([]byte, 4096)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func redirectSyslog(t *testing.T, network, address string) {
	t.Helper()

	syslogNetwork, syslogAddress = network, address
	t.Cleanup(func() {
		syslogNetwork, syslogAddress = "", ""
	})
}

func setupSyslogLogging(t *testing.T, conn net.PacketConn, cfg *Config) {
	t.Helper()

	redirectSyslog(t, "udp", conn.LocalAddr().String())
	cfg.LogFormat = "syslog"
	require.NoError(t, SetupLogging(cfg))
	t.Cleanup(func() {
		require.NoError(t, CloseLogging())
	})
}

func TestSyslogLevels(t *testing.T) {
	conn := newSyslogListener(t)
	setupSyslogLogging(t, conn, &Config{Debug: true})

	l := New("inputs", "test", "")
	tests := []struct {
		name     string
		log      func(args ...any)
		priority string
	}{
		{name: "error", log: l.Error, priority: "<27>"},
		{name: "warning", log: l.Warn, priority: "<28>"},
		{name: "info", log: l.Info, priority: "<30>"},
		{name: "debug", log: l.Debug, priority: "<31>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.log("message")

			msg, err := readSyslogMessage(t, conn, 2*time.Second)
			require.NoError(t, err)
			require.True(t, strings.HasPrefix(msg, tt.priority), "unexpected priority in %q", msg)
			require.Contains(t, msg, " telegraf[")
			require.True(t, strings.HasSuffix(strings.TrimSpace(msg), "[inputs.test] message"), "unexpected message %q", msg)
		})
	}
}

func TestSyslogLevelFiltering(t *testing.T) {
	conn := newSyslogListener(t)
	setupSyslogLogging(t, conn, &Config{})

	l := New("inputs", "test", "")
	l.Debug("debug message")

	_, err := readSyslogMessage(t, conn, 200*time.Millisecond)
	require.ErrorIs(t, err, os.ErrDeadlineExceeded)

	l.Info("info message")
	msg, err := readSyslogMessage(t, conn, 2*time.Second)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg, "<30>"), "unexpected priority in %q", msg)
	require.True(t, strings.HasSuffix(strings.TrimSpace(msg), "[inputs.test] info message"), "unexpected message %q", msg)
}

func TestSyslogInstanceName(t *testing.T) {
	conn := newSyslogListener(t)
	setupSyslogLogging(t, conn, &Config{InstanceName: "mytag"})

	New("inputs", "test", "").Info("message")

	msg, err := readSyslogMessage(t, conn, 2*time.Second)
	require.NoError(t, err)
	require.Contains(t, msg, " mytag[")
	require.NotContains(t, msg, " telegraf[")
}

func TestSyslogStdlogRedirect(t *testing.T) {
	conn := newSyslogListener(t)
	setupSyslogLogging(t, conn, &Config{})

	log.Println("W! Warn message")

	msg, err := readSyslogMessage(t, conn, 2*time.Second)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(msg, "<28>"), "unexpected priority in %q", msg)
	require.True(t, strings.HasSuffix(strings.TrimSpace(msg), "Warn message"), "unexpected message %q", msg)
}

func TestSyslogCloseTwice(t *testing.T) {
	conn := newSyslogListener(t)

	redirectSyslog(t, "udp", conn.LocalAddr().String())

	s, err := createSyslogLogger(&Config{InstanceName: "telegraf"})
	require.NoError(t, err)

	l, ok := s.(*syslogLogger)
	require.True(t, ok)
	require.NoError(t, l.Close())
	require.NoError(t, l.Close())
}

func TestSyslogConnectFailure(t *testing.T) {
	redirectSyslog(t, "unix", t.TempDir()+"/nonexistent.sock")

	_, err := createSyslogLogger(&Config{InstanceName: "telegraf"})
	require.ErrorContains(t, err, "connecting to syslog failed")
}
