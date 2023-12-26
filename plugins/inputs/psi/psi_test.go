//go:build linux

package psi

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/prometheus/procfs"
	"github.com/stretchr/testify/require"
)

func TestPSIStats(t *testing.T) {
	var (
		psi *Psi
		err error
		acc testutil.Accumulator
	)

	mockPressure := &procfs.PSILine{
		Avg10:  10,
		Avg60:  60,
		Avg300: 300,
		Total:  114514,
	}

	mockPSIStats := procfs.PSIStats{
		Some: mockPressure,
		Full: mockPressure,
	}

	mockStats := map[string]procfs.PSIStats{
		"cpu":    mockPSIStats,
		"memory": mockPSIStats,
		"io":     mockPSIStats,
	}

	err = psi.Gather(&acc)
	require.NoError(t, err)

	pressureFields := map[string]interface{}{
		"avg10":  float64(10),
		"avg60":  float64(60),
		"avg300": float64(300),
	}
	pressureTotalFields := map[string]interface{}{
		"total": uint64(114514),
	}
	acc.ClearMetrics()
	psi.uploadPressure(mockStats, &acc)
	for _, typ := range []string{"some", "full"} {
		for _, resource := range []string{"cpu", "memory", "io"} {
			if resource == "cpu" && typ == "full" {
				continue
			}

			// "pressure" should contain what it should
			acc.AssertContainsTaggedFields(t, "pressure", pressureFields, map[string]string{
				"resource": resource,
				"type":     typ,
			})

			// "pressure" should NOT contain a "total" field
			acc.AssertDoesNotContainsTaggedFields(t, "pressure", pressureTotalFields, map[string]string{
				"resource": resource,
				"type":     typ,
			})
		}
	}

	acc.AssertDoesNotContainsTaggedFields(t, "pressure", pressureFields, map[string]string{
		"resource": "cpu",
		"type":     "full",
	})
}
