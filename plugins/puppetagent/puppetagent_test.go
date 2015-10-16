package puppetagent

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdb/telegraf/testutil"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator

	pa := PuppetAgent{
		Location: "last_run_summary.yaml",
	}
	pa.Gather(&acc)

	assert.True(t, acc.HasIntValue("puppetagent_events_failure"))
	assert.True(t, acc.HasIntValue("puppetagent_events_total"))
	assert.True(t, acc.HasIntValue("puppetagent_events_success"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_failed"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_scheduled"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_changed"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_skipped"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_total"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_failedtorestart"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_restarted"))
	assert.True(t, acc.HasIntValue("puppetagent_resources_outofsync"))
	assert.True(t, acc.HasIntValue("puppetagent_changes_total"))
	assert.True(t, acc.HasIntValue("puppetagent_time_service"))
	assert.True(t, acc.HasIntValue("puppetagent_time_lastrun"))
	assert.True(t, acc.HasIntValue("puppetagent_version_config"))

	assert.True(t, acc.HasFloatValue("puppetagent_time_user"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_schedule"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_filebucket"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_file"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_exec"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_anchor"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_sshauthorizedkey"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_package"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_total"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_configretrieval"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_lastrun"))
	assert.True(t, acc.HasFloatValue("puppetagent_time_cron"))

	assert.False(t, acc.HasFloatValue("puppetagent_version_config"))
	assert.False(t, acc.HasIntValue("puppetagent_version_config"))
}
