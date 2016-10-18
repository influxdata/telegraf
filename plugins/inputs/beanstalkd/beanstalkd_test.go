package beanstalkd

import (
	"bufio"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeanstalkdGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Beanstalkd{
		Servers: []string{testutil.GetLocalHost()},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{"current-jobs-urgent", "current-jobs-ready", "current-jobs-reserved",
		"current-jobs-delayed", "current-jobs-buried", "cmd-put", "cmd-peek", "cmd-peek-ready",
		"cmd-peek-delayed", "cmd-peek-buried", "cmd-reserve", "cmd-reserve-with-timeout",
		"cmd-delete", "cmd-release", "cmd-use", "cmd-watch", "cmd-ignore",
		"cmd-bury", "cmd-kick", "cmd-touch", "cmd-stats", "cmd-stats-job",
		"cmd-stats-tube", "cmd-list-tubes", "cmd-list-tube-used", "cmd-list-tubes-watched",
		"cmd-pause-tube", "job-timeouts", "total-jobs",
		"current-tubes", "current-connections", "current-producers",
		"current-workers", "current-waiting", "total-connections", "uptime",
		"binlog-oldest-index", "binlog-current-index", "binlog-records-migrated",
		"binlog-records-written", "binlog-max-size",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("beanstalkd", metric), metric)
	}
}

func TestBeanstalkdParseMetrics(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(beanstalkdStats))
	values, err := parseResponse(r)
	require.NoError(t, err, "Error parsing beanstalkd response")

	tests := []struct {
		key   string
		value string
	}{
		{"current-jobs-urgent", "0"},
		{"current-jobs-ready", "0"},
		{"current-jobs-reserved", "0"},
		{"current-jobs-delayed", "0"},
		{"current-jobs-buried", "0"},
		{"cmd-put", "613"},
		{"cmd-peek", "0"},
		{"cmd-peek-ready", "0"},
		{"cmd-peek-delayed", "0"},
		{"cmd-peek-buried", "0"},
		{"cmd-reserve", "805"},
		{"cmd-reserve-with-timeout", "0"},
		{"cmd-delete", "613"},
		{"cmd-release", "0"},
		{"cmd-use", "613"},
		{"cmd-watch", "805"},
		{"cmd-ignore", "805"},
		{"cmd-bury", "0"},
		{"cmd-kick", "0"},
		{"cmd-touch", "0"},
		{"cmd-stats", "5"},
		{"cmd-stats-job", "0"},
		{"cmd-stats-tube", "0"},
		{"cmd-list-tubes", "0"},
		{"cmd-list-tube-used", "0"},
		{"cmd-list-tubes-watched", "0"},
		{"cmd-pause-tube", "0"},
		{"job-timeouts", "0"},
		{"total-jobs", "613"},
		{"current-tubes", "1"},
		{"current-connections", "1"},
		{"current-producers", "0"},
		{"current-workers", "0"},
		{"current-waiting", "0"},
		{"total-connections", "809"},
		{"uptime", "8007"},
		{"binlog-oldest-index", "10"},
		{"binlog-current-index", "10"},
		{"binlog-records-migrated", "0"},
		{"binlog-records-written", "1226"},
		{"binlog-max-size", "10485760"},
	}

	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("Did not find key for metric %s in values", test.key)
			continue
		}
		if value != test.value {
			t.Errorf("Metric: %s, Expected: %s, actual: %s",
				test.key, test.value, value)
		}
	}
}

var beanstalkdStats = `---
current-jobs-urgent: 0
current-jobs-ready: 0
current-jobs-reserved: 0
current-jobs-delayed: 0
current-jobs-buried: 0
cmd-put: 613
cmd-peek: 0
cmd-peek-ready: 0
cmd-peek-delayed: 0
cmd-peek-buried: 0
cmd-reserve: 805
cmd-reserve-with-timeout: 0
cmd-delete: 613
cmd-release: 0
cmd-use: 613
cmd-watch: 805
cmd-ignore: 805
cmd-bury: 0
cmd-kick: 0
cmd-touch: 0
cmd-stats: 1
cmd-stats-job: 0
cmd-stats-tube: 0
cmd-list-tubes: 0
cmd-list-tube-used: 0
cmd-list-tubes-watched: 0
cmd-pause-tube: 0
job-timeouts: 0
total-jobs: 613
current-tubes: 1
current-connections: 1
current-producers: 0
current-workers: 0
current-waiting: 0
total-connections: 806
version: 1.10
rusage-utime: 0.060000
rusage-stime: 0.212000
uptime: 7754
binlog-oldest-index: 10
binlog-current-index: 10
binlog-records-migrated: 0
binlog-records-written: 1226
binlog-max-size: 10485760
id: 95dd7f24ee126dd6
hostname: irrlab

`
