package puppetagent

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator

	pa := PuppetAgent{
		Location: "last_run_summary.yaml",
	}
	require.NoError(t, pa.Gather(&acc))

	tags := map[string]string{"location": "last_run_summary.yaml"}
	fields := map[string]interface{}{
		"events_failure":             int64(0),
		"events_noop":                int64(0),
		"events_success":             int64(0),
		"events_total":               int64(0),
		"resources_changed":          int64(0),
		"resources_correctivechange": int64(0),
		"resources_failed":           int64(0),
		"resources_failedtorestart":  int64(0),
		"resources_outofsync":        int64(0),
		"resources_restarted":        int64(0),
		"resources_scheduled":        int64(0),
		"resources_skipped":          int64(0),
		"resources_total":            int64(109),
		"changes_total":              int64(0),
		"time_anchor":                float64(0.000555),
		"time_catalogapplication":    float64(0.010555),
		"time_configretrieval":       float64(4.75567007064819),
		"time_convertcatalog":        float64(1.3),
		"time_cron":                  float64(0.000584),
		"time_exec":                  float64(0.508123),
		"time_factgeneration":        float64(0.34),
		"time_file":                  float64(0.441472),
		"time_filebucket":            float64(0.000353),
		"time_group":                 float64(0),
		"time_lastrun":               int64(1444936531),
		"time_noderetrieval":         float64(1.235),
		"time_notify":                float64(0.00035),
		"time_package":               float64(1.325788),
		"time_pluginsync":            float64(0.325788),
		"time_schedule":              float64(0.001123),
		"time_service":               float64(1.807795),
		"time_sshauthorizedkey":      float64(0.000764),
		"time_total":                 float64(8.85354707064819),
		"time_transactionevaluation": float64(4.69765),
		"time_user":                  float64(0.004331),
		"version_configstring":       "environment:d6018ce",
		"version_puppet":             "3.7.5",
	}

	acc.AssertContainsTaggedFields(t, "puppetagent", fields, tags)
}
