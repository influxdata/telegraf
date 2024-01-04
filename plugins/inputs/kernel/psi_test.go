//go:build linux

package kernel

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestPSIEnabledWrongDir(t *testing.T) {
	k := Kernel{
		psiDir:        "testdata/this_directory_does_not_exist/stub",
		ConfigCollect: []string{"psi"},
	}

	require.ErrorContains(t, k.Init(), "failed to initialize procfs on ")
}

func TestPSIStats(t *testing.T) {
	var acc testutil.Accumulator

	k := Kernel{
		psiDir:        "testdata/pressure",
		ConfigCollect: []string{"psi"},
	}
	require.NoError(t, k.Init())
	require.NoError(t, k.gatherPressure(&acc))

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
