//go:build linux

package system

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGatherOSValuesLinux(t *testing.T) {
	etcDir, err := filepath.Abs(filepath.Join("testdata", "os-release"))
	require.NoError(t, err)
	t.Setenv("HOST_ETC", etcDir)

	s := &System{
		Include:    []string{"os"},
		OSCacheTTL: config.Duration(8 * time.Hour),
		Log:        &testutil.Logger{},
	}
	require.NoError(t, s.Init())

	var acc testutil.Accumulator
	require.NoError(t, s.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"system_os",
			map[string]string{},
			map[string]interface{}{
				"os":               "linux",
				"arch":             "amd64",
				"platform":         "telegraftest",
				"platform_family":  "",
				"platform_version": "1.0",
				"kernel_version":   "",
			},
			time.Unix(0, 0),
			telegraf.Untyped,
		),
	}

	testutil.RequireMetricsStructureEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
