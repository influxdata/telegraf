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

func TestParseMetric(t *testing.T) {
	l := defaultLogzio()
	for _, tm := range testutil.MockMetrics() {
		lm := l.parseMetric(tm)
		require.Equal(t, tm.Fields(), lm.Metric[tm.Name()])
		require.Equal(t, logzioType, lm.Type)
		require.Equal(t, tm.Tags(), lm.Dimensions)
		require.Equal(t, tm.Time(), lm.Time)
	}
}

func TestLogzioConnectWitoutToken(t *testing.T) {
	l := defaultLogzio()
	err := l.Connect()
	require.Error(t, err)
}
