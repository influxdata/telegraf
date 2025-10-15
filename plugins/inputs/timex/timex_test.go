//go:build linux

package timex

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestDefaultMetricFormat(t *testing.T) {
	plugin := &Timex{
		Log: &testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}
