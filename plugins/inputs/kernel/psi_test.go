//go:build linux

package kernel

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/prometheus/procfs"
	"github.com/stretchr/testify/require"
)

func TestPSIEnabledWrongDir(t *testing.T) {
	k := Kernel{
		psiDir:        "testdata/this_file_does_not_exist",
		ConfigCollect: []string{"psi"},
	}

	require.ErrorContains(t, k.Init(), "PSI is not enabled in this kernel")
}

func TestPSIStats(t *testing.T) {
	var (
		k   *Kernel
		err error
		acc testutil.Accumulator
	)

	mockPressureSome := &procfs.PSILine{
		Avg10:  10,
		Avg60:  60,
		Avg300: 300,
		Total:  114514,
	}
	mockPressureFull := &procfs.PSILine{
		Avg10:  1,
		Avg60:  6,
		Avg300: 30,
		Total:  11451,
	}
	mockPSIStats := procfs.PSIStats{
		Some: mockPressureSome,
		Full: mockPressureFull,
	}
	mockStats := map[string]procfs.PSIStats{
		"cpu":    mockPSIStats,
		"memory": mockPSIStats,
		"io":     mockPSIStats,
	}

	err = k.gatherPressure(&acc)
	require.NoError(t, err)

	// separate fields for gauges and counters
	pressureFields := map[string]map[string]interface{}{
		"some": {
			"avg10":  float64(10),
			"avg60":  float64(60),
			"avg300": float64(300),
		},
		"full": {
			"avg10":  float64(1),
			"avg60":  float64(6),
			"avg300": float64(30),
		},
	}
	pressureTotalFields := map[string]map[string]interface{}{
		"some": {
			"total": uint64(114514),
		},
		"full": {
			"total": uint64(11451),
		},
	}

	acc.ClearMetrics()
	k.uploadPressure(mockStats, &acc)
	for _, typ := range []string{"some", "full"} {
		for _, resource := range []string{"cpu", "memory", "io"} {
			if resource == "cpu" && typ == "full" {
				continue
			}

			tags := map[string]string{
				"resource": resource,
				"type":     typ,
			}

			acc.AssertContainsTaggedFields(t, "pressure", pressureFields[typ], tags)
			acc.AssertContainsTaggedFields(t, "pressure", pressureTotalFields[typ], tags)
		}
	}

	// The combination "resource=cpu,type=full" should NOT appear anywhere
	forbiddenTags := map[string]string{
		"resource": "cpu",
		"type":     "full",
	}
	acc.AssertDoesNotContainsTaggedFields(t, "pressure", pressureFields["full"], forbiddenTags)
	acc.AssertDoesNotContainsTaggedFields(t, "pressure", pressureTotalFields["full"], forbiddenTags)
}
