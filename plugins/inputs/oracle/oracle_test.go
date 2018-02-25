package oracle

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOracleGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Oracle{
		Connection:           testConnection(),
		InstanceStateMetrics: true,
		SystemMetrics:        true,
		TablespaceMetrics:    true,
		WaitClassMetrics:     true,
		WaitEventMetrics:     true,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(o.Gather)
	require.NoError(t, err)

	tablespaceFields, err := readTestDataFile("testdata/expected/tablespace_fields.txt")
	require.NoError(t, err)
	for _, field := range tablespaceFields {
		assert.True(t, acc.HasField("oracle_tablespace", field))
	}

	instanceStateFields, err := readTestDataFile("testdata/expected/instance_state_fields.txt")
	require.NoError(t, err)
	for _, field := range instanceStateFields {
		assert.True(t, acc.HasField("oracle_instance_state", field))
	}

	waitClassFields, err := readTestDataFile("testdata/expected/wait_class_fields.txt")
	require.NoError(t, err)
	for _, field := range waitClassFields {
		assert.True(t, acc.HasField("oracle_wait_class", field))
	}

	waitEventFields, err := readTestDataFile("testdata/expected/wait_event_fields.txt")
	require.NoError(t, err)
	for _, field := range waitEventFields {
		assert.True(t, acc.HasField("oracle_wait_event", field))
	}
}

func TestOracleCommonTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Oracle{
		Connection:           testConnection(),
		InstanceStateMetrics: true,
		SystemMetrics:        true,
		TablespaceMetrics:    true,
		WaitClassMetrics:     true,
		WaitEventMetrics:     true,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(o.Gather)
	require.NoError(t, err)

	commonTags, err := readTestDataFile("testdata/expected/common_tags.txt")
	require.NoError(t, err)

	for _, tag := range commonTags {
		assert.True(t, acc.HasTag("oracle_instance_state", tag))
	}
	for _, tag := range commonTags {
		assert.True(t, acc.HasTag("oracle_system", tag))
	}
	for _, tag := range commonTags {
		assert.True(t, acc.HasTag("oracle_tablespace", tag))
	}
	for _, tag := range commonTags {
		assert.True(t, acc.HasTag("oracle_wait_class", tag))
	}
	for _, tag := range commonTags {
		assert.True(t, acc.HasTag("oracle_wait_event", tag))
	}
}

func TestOracleTablespaceTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Oracle{
		Connection:           testConnection(),
		InstanceStateMetrics: false,
		SystemMetrics:        false,
		TablespaceMetrics:    true,
		WaitClassMetrics:     false,
		WaitEventMetrics:     false,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(o.Gather)
	require.NoError(t, err)

	tablespaceTags, err := readTestDataFile("testdata/expected/tablespace_tags.txt")
	require.NoError(t, err)

	for _, tag := range tablespaceTags {
		assert.True(t, acc.HasTag("oracle_tablespace", tag))
	}
}

func TestOracleWaitClassTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Oracle{
		Connection:           testConnection(),
		InstanceStateMetrics: false,
		SystemMetrics:        false,
		TablespaceMetrics:    false,
		WaitClassMetrics:     true,
		WaitEventMetrics:     false,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(o.Gather)
	require.NoError(t, err)

	waitClassTags, err := readTestDataFile("testdata/expected/wait_class_tags.txt")
	require.NoError(t, err)

	for _, tag := range waitClassTags {
		assert.True(t, acc.HasTag("oracle_wait_class", tag))
	}
}

func TestOracleWaitEventTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Oracle{
		Connection:           testConnection(),
		InstanceStateMetrics: false,
		SystemMetrics:        false,
		TablespaceMetrics:    false,
		WaitClassMetrics:     false,
		WaitEventMetrics:     true,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(o.Gather)
	require.NoError(t, err)

	waitEventTags, err := readTestDataFile("testdata/expected/wait_event_tags.txt")
	require.NoError(t, err)

	for _, tag := range waitEventTags {
		assert.True(t, acc.HasTag("oracle_wait_event", tag))
	}
}

func TestStringSanitization(t *testing.T) {
	tests := []struct {
		Dirty     string
		Sanitized string
	}{
		{
			"Parameter File I/O",
			"parameter_file_io",
		},
		{
			"SQL*Net message from client",
			"sql_net_message_from_client",
		},
		{
			"enq: JS - queue lock",
			"enq_js_queue_lock",
		},
		{
			"Disk file Mirror/Media Repair Write",
			"disk_file_mirror_media_repair_write",
		},
		{
			"direct path read",
			"direct_path_read",
		},
		{
			"System I/O",
			"system_io",
		},
		{
			"usage_%",
			"usage_percent",
		},
	}

	for _, test := range tests {
		actual := sanitize(test.Dirty)
		assert.Equal(t, test.Sanitized, actual)
	}
}

func readTestDataFile(path string) ([]string, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(string(d), "\n"), nil
}

func testConnection() Connection {
	return Connection{
		SID:           testutil.GetLocalHost() + ":1521/xe.oracle.docker",
		Username:      "system",
		Password:      "oracle",
		MaxLifetime:   internal.Duration{},
		MinSessions:   10,
		MaxSessions:   20,
		PoolIncrement: 1,
	}
}
