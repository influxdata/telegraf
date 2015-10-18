package puppetagent

import (
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator

	pa := PuppetAgent{
		Location: "last_run_summary.yaml",
	}
	pa.Gather(&acc)

	checkInt := []struct {
		name  string
		value int64
	}{
		{"events_failure", 0},
		{"events_total", 0},
		{"events_success", 0},
		{"resources_failed", 0},
		{"resources_scheduled", 0},
		{"resources_changed", 0},
		{"resources_skipped", 0},
		{"resources_total", 109},
		{"resources_failedtorestart", 0},
		{"resources_restarted", 0},
		{"resources_outofsync", 0},
		{"changes_total", 0},
		{"time_lastrun", 1444936531},
		{"version_config", 1444936521},
	}

	for _, c := range checkInt {
		assert.Equal(t, true, acc.CheckValue(c.name, c.value))
	}

	checkFloat := []struct {
		name  string
		value float64
	}{
		{"time_user", 0.004331},
		{"time_schedule", 0.001123},
		{"time_filebucket", 0.000353},
		{"time_file", 0.441472},
		{"time_exec", 0.508123},
		{"time_anchor", 0.000555},
		{"time_sshauthorizedkey", 0.000764},
		{"time_service", 1.807795},
		{"time_package", 1.325788},
		{"time_total", 8.85354707064819},
		{"time_configretrieval", 4.75567007064819},
		{"time_cron", 0.000584},
	}

	for _, f := range checkFloat {
		assert.Equal(t, true, acc.CheckValue(f.name, f.value))
	}

	checkString := []struct {
		name  string
		value string
	}{
		{"version_puppet", "3.7.5"},
	}

	for _, s := range checkString {
		assert.Equal(t, true, acc.CheckValue(s.name, s.value))
	}

}
