package system

import (
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestUniqueUsers(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		data     []host.UserStat
	}{
		{
			name:     "single entry",
			expected: 1,
			data: []host.UserStat{
				{User: "root"},
			},
		},
		{
			name:     "empty entry",
			expected: 0,
		},
		{
			name:     "all duplicates",
			expected: 1,
			data: []host.UserStat{
				{User: "root"},
				{User: "root"},
				{User: "root"},
			},
		},
		{
			name:     "all unique",
			expected: 3,
			data: []host.UserStat{
				{User: "root"},
				{User: "ubuntu"},
				{User: "ec2-user"},
			},
		},
		{
			name:     "mix of dups",
			expected: 3,
			data: []host.UserStat{
				{User: "root"},
				{User: "ubuntu"},
				{User: "ubuntu"},
				{User: "ubuntu"},
				{User: "ec2-user"},
				{User: "ec2-user"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := findUniqueUsers(tt.data)
			require.Equal(t, tt.expected, actual, tt.name)
		})
	}
}

func TestInitAllValidOptions(t *testing.T) {
	// cpus/legacy_cpus and uptime/legacy_uptime are mutually exclusive,
	// so cover all six valid values across two configurations.
	tests := []struct {
		name    string
		include []string
	}{
		{"new", []string{"load", "users", "cpus", "uptime"}},
		{"legacy", []string{"load", "users", "legacy_cpus", "legacy_uptime"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &System{Include: tt.include, Log: &testutil.Logger{}}
			require.NoError(t, s.Init())
		})
	}
}

func TestInitErrors(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		errMsg  string
	}{
		{
			name:    "invalid option",
			include: []string{"invalid"},
			errMsg:  `invalid 'include' option "invalid"`,
		},
		{
			name:    "cpus mutually exclusive",
			include: []string{"cpus", "legacy_cpus"},
			errMsg:  "mutually exclusive",
		},
		{
			name:    "uptime mutually exclusive",
			include: []string{"uptime", "legacy_uptime"},
			errMsg:  "mutually exclusive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &System{
				Include: tt.include,
				Log:     &testutil.Logger{},
			}
			require.ErrorContains(t, s.Init(), tt.errMsg)
		})
	}
}

func TestGather(t *testing.T) {
	// host.Users() depends on /var/run/utmp which is not available on every
	// runner. On Linux we mock it via HOST_VAR; on other platforms we probe
	// at runtime and skip relevant cases when the call cannot be satisfied.
	usersAvailable := setupUsers(t)

	tests := []struct {
		name         string
		include      []string
		expected     []telegraf.Metric
		requireUsers bool
	}{
		{
			name:         "default",
			include:      nil,
			requireUsers: true,
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{
						"load1":           float64(0),
						"load5":           float64(0),
						"load15":          float64(0),
						"n_users":         0,
						"n_unique_users":  0,
						"n_cpus":          0,
						"n_physical_cpus": 0,
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime": uint64(0)},
					time.Unix(0, 0),
					telegraf.Counter,
				),
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime_format": string("")},
					time.Unix(0, 0),
					telegraf.Untyped,
				),
			},
		},
		{
			name:    "cpus",
			include: []string{"cpus"},
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{
						"n_virtual_cpus":  0,
						"n_physical_cpus": 0,
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:    "uptime as gauge field",
			include: []string{"uptime"},
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime": uint64(0)},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:         "all new options",
			include:      []string{"load", "users", "cpus", "uptime"},
			requireUsers: true,
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{
						"load1":           float64(0),
						"load5":           float64(0),
						"load15":          float64(0),
						"n_users":         0,
						"n_unique_users":  0,
						"n_virtual_cpus":  0,
						"n_physical_cpus": 0,
						"uptime":          uint64(0),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:    "legacy_uptime only",
			include: []string{"legacy_uptime"},
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime": uint64(0)},
					time.Unix(0, 0),
					telegraf.Counter,
				),
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime_format": string("")},
					time.Unix(0, 0),
					telegraf.Untyped,
				),
			},
		},
		{
			name:         "users only",
			include:      []string{"users"},
			requireUsers: true,
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{
						"n_users":        0,
						"n_unique_users": 0,
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:    "duplicates are de-duplicated",
			include: []string{"legacy_uptime", "legacy_uptime", "cpus", "cpus"},
			expected: []telegraf.Metric{
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{
						"n_virtual_cpus":  0,
						"n_physical_cpus": 0,
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime": uint64(0)},
					time.Unix(0, 0),
					telegraf.Counter,
				),
				metric.New(
					"system",
					map[string]string{},
					map[string]interface{}{"uptime_format": string("")},
					time.Unix(0, 0),
					telegraf.Untyped,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.requireUsers && !usersAvailable {
				t.Skip("host.Users() not mockable on this platform")
			}
			s := &System{
				Include: tt.include,
				Log:     &testutil.Logger{},
			}
			require.NoError(t, s.Init())

			var acc testutil.Accumulator
			require.NoError(t, s.Gather(&acc))

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsStructureEqual(t, tt.expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}
