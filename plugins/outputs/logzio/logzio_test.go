package logzio

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewLogzioOutput(t *testing.T) {
	l := NewLogzioOutput()
	require.Equal(t, l.CheckDiskSpace, defaultLogzioCheckDiskSpace)
	require.Equal(t, l.DiskThreshold, defaultLogzioDiskThreshold)
	require.Equal(t, l.DrainDuration, defaultLogzioDrainDuration)
	require.Equal(t, l.URL, defaultLogzioURL)

	require.Equal(t, l.Token, "")
}

func TestConnectAndWrite(t *testing.T) {
	l := NewLogzioOutput()
	l.Token = "123456789"

	err := l.Connect()
	require.NoError(t, err)

	err = l.Write(testutil.MockMetrics())
	l.Close()
	require.NoError(t, err)
}

func TestLogzioConnectWitoutToken(t *testing.T) {
	l := NewLogzioOutput()
	err := l.Connect()
	require.Error(t, err)
}
