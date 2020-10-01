package logzio

import (
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func defaultLogzio() *Logzio {
	return &Logzio{
		CheckDiskSpace: defaultLogzioCheckDiskSpace,
		DiskThreshold:  defaultLogzioDiskThreshold,
		DrainDuration:  defaultLogzioDrainDuration,
		Log:            testutil.Logger{},
		QueueDir: fmt.Sprintf("%s%s%s%s%d", os.TempDir(), string(os.PathSeparator),
			"logzio-queue", string(os.PathSeparator), time.Now().UnixNano()),
		URL: defaultLogzioURL,
	}
}

func TestNewLogzioOutput(t *testing.T) {
	l := defaultLogzio()
	require.Equal(t, l.CheckDiskSpace, defaultLogzioCheckDiskSpace)
	require.Equal(t, l.DiskThreshold, defaultLogzioDiskThreshold)
	require.Equal(t, l.DrainDuration, defaultLogzioDrainDuration)
	require.Equal(t, l.URL, defaultLogzioURL)

	require.Equal(t, l.Token, "")
}

func TestConnectAndWrite(t *testing.T) {
	l := defaultLogzio()
	l.Token = "123456789"

	err := l.Connect()
	require.NoError(t, err)

	err = l.Write(testutil.MockMetrics())
	l.Close()
	require.NoError(t, err)
}

func TestLogzioConnectWitoutToken(t *testing.T) {
	l := defaultLogzio()
	err := l.Connect()
	require.Error(t, err)
}
