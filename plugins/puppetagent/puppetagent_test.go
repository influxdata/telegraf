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

	assert.True(t, acc.HasIntValue("events_failure"))
	assert.True(t, acc.HasIntValue("events_total"))
	assert.True(t, acc.HasIntValue("events_success"))
	assert.True(t, acc.HasIntValue("resources_failed"))
	assert.True(t, acc.HasIntValue("resources_scheduled"))
	assert.True(t, acc.HasIntValue("resources_changed"))
	assert.True(t, acc.HasIntValue("resources_skipped"))
	assert.True(t, acc.HasIntValue("resources_total"))
	assert.True(t, acc.HasIntValue("resources_failedtorestart"))
	assert.True(t, acc.HasIntValue("resources_restarted"))
	assert.True(t, acc.HasIntValue("resources_outofsync"))
	assert.True(t, acc.HasIntValue("changes_total"))

	assert.True(t, acc.HasIntValue("time_lastrun"))
	assert.True(t, acc.HasIntValue("version_config"))

	assert.True(t, acc.HasFloatValue("time_user"))
	assert.True(t, acc.HasFloatValue("time_schedule"))
	assert.True(t, acc.HasFloatValue("time_filebucket"))
	assert.True(t, acc.HasFloatValue("time_file"))
	assert.True(t, acc.HasFloatValue("time_exec"))
	assert.True(t, acc.HasFloatValue("time_anchor"))
	assert.True(t, acc.HasFloatValue("time_sshauthorizedkey"))
	assert.True(t, acc.HasFloatValue("time_service"))
	assert.True(t, acc.HasFloatValue("time_package"))
	assert.True(t, acc.HasFloatValue("time_total"))
	assert.True(t, acc.HasFloatValue("time_configretrieval"))
	assert.True(t, acc.HasFloatValue("time_cron"))
}
