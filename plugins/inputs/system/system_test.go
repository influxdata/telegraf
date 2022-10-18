package system

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAddFields(t *testing.T) {
	tests := []struct {
		MergeMetrics bool
		expectedLen  int
	}{
		{
			false,
			3,
		},
		{
			true,
			1,
		},
	}

	for _, test := range tests {
		var testAcc testutil.Accumulator

		loadFields := map[string]interface{}{
			"load1":   3.72,
			"load5":   2.4,
			"load15":  2.1,
			"n_cpus":  4,
			"n_users": 3,
		}
		uptimeFields := map[string]interface{}{
			"uptime": uint64(1249632),
		}
		uptimeFormatfields := map[string]interface{}{
			"uptimeFormat": "14 days, 11:07",
		}

		var s SystemStats
		s.MergeMetrics = test.MergeMetrics
		s.addFields(&testAcc, loadFields, uptimeFields, uptimeFormatfields)

		require.Equal(t, test.expectedLen, len(testAcc.GetTelegrafMetrics()))
	}
}
