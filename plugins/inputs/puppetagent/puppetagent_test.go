package puppetagent

import (
	"github.com/influxdata/telegraf/testutil"
	"testing"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator

	pa := PuppetAgent{
		Location: "last_run_summary.yaml",
	}
	pa.Gather(&acc)

	tags := map[string]string{"location": "last_run_summary.yaml"}
	fields := map[string]interface{}{
		"events_failure":            int64(0),
		"events_total":              int64(0),
		"events_success":            int64(0),
		"resources_failed":          int64(0),
		"resources_scheduled":       int64(0),
		"resources_changed":         int64(0),
		"resources_skipped":         int64(0),
		"resources_total":           int64(109),
		"resources_failedtorestart": int64(0),
		"resources_restarted":       int64(0),
		"resources_outofsync":       int64(0),
		"changes_total":             int64(0),
		"time_lastrun":              int64(1444936531),
		"version_configstring":      "environment:d6018ce",
		"time_user":                 float64(0.004331),
		"time_schedule":             float64(0.001123),
		"time_filebucket":           float64(0.000353),
		"time_file":                 float64(0.441472),
		"time_exec":                 float64(0.508123),
		"time_anchor":               float64(0.000555),
		"time_sshauthorizedkey":     float64(0.000764),
		"time_service":              float64(1.807795),
		"time_package":              float64(1.325788),
		"time_total":                float64(8.85354707064819),
		"time_configretrieval":      float64(4.75567007064819),
		"time_cron":                 float64(0.000584),
		"version_puppet":            "3.7.5",
	}

	acc.AssertContainsTaggedFields(t, "puppetagent", fields, tags)
}
