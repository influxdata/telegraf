package classify

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/testutil"
)

var now = time.Unix(1700000000, 0)

// newClassify returns a plugin with a database and a firewall group,
// selected by the "host" tag and matching the "message" field.
func newClassify() *Classify {
	return &Classify{
		SelectorTag: "host",
		SelectorMapping: []selectorMapping{
			{Pattern: `^fire\d{3}$`, Group: "firewall"},
			{Pattern: `^pg\d{3}$`, Group: "database"},
			{Pattern: `^test`, Group: ""},
			{Pattern: `^(database|firewall|other)$`, Group: "*"},
		},
		MatchField: "message",
		ResultTag:  "status",
		MappedSelectorRegexes: map[string][]category{
			"database": {
				{Category: "ignore", Patterns: []string{"DB client connected"}},
				{Category: "warning", Patterns: []string{`Tablespace \w+ free space is low`}},
				{Category: "critical", Patterns: []string{"Database is shutting down"}},
			},
			"firewall": {
				{Category: "warning", Patterns: []string{"login attempt"}},
				{Category: "critical", Patterns: []string{"intrusion detected", "port ?scan"}},
			},
		},
		Log: testutil.Logger{},
	}
}

func message(host, msg string) telegraf.Metric {
	return metric.New("syslog", map[string]string{"host": host}, map[string]any{"message": msg}, now)
}

func classified(host, msg, status string) telegraf.Metric {
	return metric.New("syslog", map[string]string{"host": host, "status": status}, map[string]any{"message": msg}, now)
}

func TestSampleConfig(t *testing.T) {
	c := config.NewConfig()
	require.NoError(t, c.LoadConfigData(testutil.DefaultSampleConfig((&Classify{}).SampleConfig()), config.EmptySourcePath))
	require.Len(t, c.Processors, 1)
	require.NoError(t, c.Processors[0].Processor.(telegraf.Initializer).Init())
}

func TestConfigParsing(t *testing.T) {
	c := config.NewConfig()
	cfg := []byte(`
[[processors.classify]]
  selector_tag = "host"
  selector_mapping = [
    { pattern = 'fire\d{3}', group = "firewall" },
    { pattern = '.*', group = "*" },
  ]
  match_field = "message"
  drop_categories = ["ignore"]
  result_tag = "status"
  [processors.classify.mapped_selector_regexes]
    firewall = [
      { category = "ignore", patterns = ["low-priority traffic"] },
      { category = "critical", patterns = ["intrusion detected", 'port ?scan'] },
    ]
`)
	require.NoError(t, c.LoadConfigData(cfg, config.EmptySourcePath))
	require.Len(t, c.Processors, 1)

	wrapped, ok := c.Processors[0].Processor.(interface{ Unwrap() telegraf.Processor })
	require.True(t, ok)
	plugin, ok := wrapped.Unwrap().(*Classify)
	require.True(t, ok)
	require.Equal(t, []selectorMapping{
		{Pattern: `fire\d{3}`, Group: "firewall"},
		{Pattern: ".*", Group: "*"},
	}, plugin.SelectorMapping)
	require.Equal(t, map[string][]category{
		"firewall": {
			{Category: "ignore", Patterns: []string{"low-priority traffic"}},
			{Category: "critical", Patterns: []string{"intrusion detected", `port ?scan`}},
		},
	}, plugin.MappedSelectorRegexes)
	require.Equal(t, []string{"ignore"}, plugin.DropCategories)
}

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Classify)
	}{
		{
			name:   "selector",
			modify: func(*Classify) {},
		},
		{
			name: "default group only",
			modify: func(cl *Classify) {
				cl.SelectorTag = ""
				cl.SelectorMapping = nil
				cl.DefaultRegexGroup = "database"
			},
		},
		{
			name: "selector field and match tag",
			modify: func(cl *Classify) {
				cl.SelectorTag, cl.SelectorField = "", "source"
				cl.MatchField, cl.MatchTag = "", "msg"
			},
		},
		{
			name: "drop categories",
			modify: func(cl *Classify) {
				cl.DefaultCategory = "unknown"
				cl.DropCategories = []string{"ignore", "unknown"}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := newClassify()
			tt.modify(cl)
			require.NoError(t, cl.Init())
		})
	}
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		modify   func(*Classify)
		expected string
	}{
		{
			name:     "selector tag and field",
			modify:   func(cl *Classify) { cl.SelectorField = "source" },
			expected: "selector_tag and selector_field cannot both be defined",
		},
		{
			name:     "no match item",
			modify:   func(cl *Classify) { cl.MatchField = "" },
			expected: "either match_tag or match_field must be defined",
		},
		{
			name:     "match tag and field",
			modify:   func(cl *Classify) { cl.MatchTag = "msg" },
			expected: "match_tag and match_field cannot both be defined",
		},
		{
			name:     "no result item",
			modify:   func(cl *Classify) { cl.ResultTag = "" },
			expected: "either result_tag or result_field must be defined",
		},
		{
			name:     "result tag and field",
			modify:   func(cl *Classify) { cl.ResultField = "status" },
			expected: "result_tag and result_field cannot both be defined",
		},
		{
			name:     "no groups",
			modify:   func(cl *Classify) { cl.MappedSelectorRegexes = nil },
			expected: "mapped_selector_regexes has no groups defined",
		},
		{
			name:     "group without categories",
			modify:   func(cl *Classify) { cl.MappedSelectorRegexes["router"] = nil },
			expected: `mapped_selector_regexes group "router" has no categories`,
		},
		{
			name: "empty category name",
			modify: func(cl *Classify) {
				cl.MappedSelectorRegexes["firewall"][0].Category = ""
			},
			expected: `empty category name in mapped_selector_regexes group "firewall"`,
		},
		{
			name: "duplicate category",
			modify: func(cl *Classify) {
				cl.MappedSelectorRegexes["firewall"][1].Category = "warning"
			},
			expected: `duplicate category "warning" in mapped_selector_regexes group "firewall"`,
		},
		{
			name: "category without patterns",
			modify: func(cl *Classify) {
				cl.MappedSelectorRegexes["firewall"][0].Patterns = nil
			},
			expected: `mapped_selector_regexes group "firewall" category "warning" has no patterns`,
		},
		{
			name: "empty pattern",
			modify: func(cl *Classify) {
				cl.MappedSelectorRegexes["firewall"][0].Patterns = []string{""}
			},
			expected: `empty pattern in mapped_selector_regexes group "firewall" category "warning"`,
		},
		{
			name: "invalid pattern",
			modify: func(cl *Classify) {
				cl.MappedSelectorRegexes["firewall"][0].Patterns = []string{"("}
			},
			expected: `invalid pattern "(" in mapped_selector_regexes group "firewall" category "warning"`,
		},
		{
			name:     "unknown default group",
			modify:   func(cl *Classify) { cl.DefaultRegexGroup = "router" },
			expected: `default_regex_group "router" is not a group in mapped_selector_regexes`,
		},
		{
			name: "no selector and no default group",
			modify: func(cl *Classify) {
				cl.SelectorTag = ""
			},
			expected: "default_regex_group must be set when neither selector_tag nor selector_field is defined",
		},
		{
			name: "empty selector pattern",
			modify: func(cl *Classify) {
				cl.SelectorMapping[0].Pattern = ""
			},
			expected: `empty pattern in selector_mapping for group "firewall"`,
		},
		{
			name: "duplicate selector pattern",
			modify: func(cl *Classify) {
				cl.SelectorMapping[1].Pattern = cl.SelectorMapping[0].Pattern
			},
			expected: `duplicate selector_mapping pattern "^fire\\d{3}$"`,
		},
		{
			name: "unknown selector group",
			modify: func(cl *Classify) {
				cl.SelectorMapping[0].Group = "firewal"
			},
			expected: `selector_mapping group "firewal" is not a group in mapped_selector_regexes`,
		},
		{
			name: "invalid selector pattern",
			modify: func(cl *Classify) {
				cl.SelectorMapping[0].Pattern = "("
			},
			expected: `invalid selector_mapping pattern "(" for group "firewall"`,
		},
		{
			name:     "empty drop category",
			modify:   func(cl *Classify) { cl.DropCategories = []string{""} },
			expected: "drop_categories contains an empty string",
		},
		{
			name:     "unknown drop category",
			modify:   func(cl *Classify) { cl.DropCategories = []string{"unknown"} },
			expected: `"unknown" in drop_categories is not a known category or default_category`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := newClassify()
			tt.modify(cl)
			require.ErrorContains(t, cl.Init(), tt.expected)
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name     string
		modify   func(*Classify)
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "first matching category wins",
			input: []telegraf.Metric{
				message("fire001", "intrusion detected after login attempt"),
				message("fire002", "login attempt from 10.0.0.1"),
				message("pg001", "Tablespace users free space is low"),
			},
			expected: []telegraf.Metric{
				classified("fire001", "intrusion detected after login attempt", "warning"),
				classified("fire002", "login attempt from 10.0.0.1", "warning"),
				classified("pg001", "Tablespace users free space is low", "warning"),
			},
		},
		{
			name: "any pattern of a category matches",
			input: []telegraf.Metric{
				message("fire001", "portscan from 10.0.0.1"),
			},
			expected: []telegraf.Metric{
				classified("fire001", "portscan from 10.0.0.1", "critical"),
			},
		},
		{
			name: "no match without default category drops",
			input: []telegraf.Metric{
				message("pg001", "checkpoint complete"),
			},
		},
		{
			name:   "no match with default category",
			modify: func(cl *Classify) { cl.DefaultCategory = "unknown" },
			input: []telegraf.Metric{
				message("pg001", "checkpoint complete"),
			},
			expected: []telegraf.Metric{
				classified("pg001", "checkpoint complete", "unknown"),
			},
		},
		{
			name: "drop categories",
			modify: func(cl *Classify) {
				cl.DefaultCategory = "unknown"
				cl.DropCategories = []string{"ignore", "unknown"}
			},
			input: []telegraf.Metric{
				message("pg001", "DB client connected"),
				message("pg001", "checkpoint complete"),
				message("pg001", "Database is shutting down"),
			},
			expected: []telegraf.Metric{
				classified("pg001", "Database is shutting down", "critical"),
			},
		},
		{
			name: "selector passthrough",
			input: []telegraf.Metric{
				message("firewall", "login attempt"),
				message("database", "Database is shutting down"),
			},
			expected: []telegraf.Metric{
				classified("firewall", "login attempt", "warning"),
				classified("database", "Database is shutting down", "critical"),
			},
		},
		{
			name: "passthrough to an unknown group drops without default group",
			input: []telegraf.Metric{
				message("other", "login attempt"),
			},
		},
		{
			name:   "passthrough to an unknown group uses the default group",
			modify: func(cl *Classify) { cl.DefaultRegexGroup = "firewall" },
			input: []telegraf.Metric{
				message("other", "login attempt"),
			},
			expected: []telegraf.Metric{
				classified("other", "login attempt", "warning"),
			},
		},
		{
			name:   "empty group drops even with default group",
			modify: func(cl *Classify) { cl.DefaultRegexGroup = "firewall" },
			input: []telegraf.Metric{
				message("test01", "login attempt"),
			},
		},
		{
			name:   "unmatched selector uses the default group",
			modify: func(cl *Classify) { cl.DefaultRegexGroup = "database" },
			input: []telegraf.Metric{
				message("web01", "Database is shutting down"),
			},
			expected: []telegraf.Metric{
				classified("web01", "Database is shutting down", "critical"),
			},
		},
		{
			name: "unmatched selector drops without default group",
			input: []telegraf.Metric{
				message("web01", "Database is shutting down"),
			},
		},
		{
			name: "missing selector or match item drops",
			input: []telegraf.Metric{
				metric.New("syslog", nil, map[string]any{"message": "login attempt"}, now),
				metric.New("syslog", map[string]string{"host": "fire001"}, map[string]any{"value": 1}, now),
				metric.New("syslog", map[string]string{"host": "fire001"}, map[string]any{"message": 1}, now),
			},
		},
		{
			name: "default group without selector",
			modify: func(cl *Classify) {
				cl.SelectorTag = ""
				cl.SelectorMapping = nil
				cl.DefaultRegexGroup = "firewall"
			},
			input: []telegraf.Metric{
				message("pg001", "login attempt"),
			},
			expected: []telegraf.Metric{
				classified("pg001", "login attempt", "warning"),
			},
		},
		{
			name: "selector field, match tag and result field",
			modify: func(cl *Classify) {
				cl.SelectorTag, cl.SelectorField = "", "source"
				cl.MatchField, cl.MatchTag = "", "msg"
				cl.ResultTag, cl.ResultField = "", "status"
			},
			input: []telegraf.Metric{
				metric.New("syslog", map[string]string{"msg": "login attempt"}, map[string]any{"source": "fire001"}, now),
			},
			expected: []telegraf.Metric{
				metric.New("syslog", map[string]string{"msg": "login attempt"}, map[string]any{"source": "fire001", "status": "warning"}, now),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := newClassify()
			if tt.modify != nil {
				tt.modify(cl)
			}
			require.NoError(t, cl.Init())
			actual := cl.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestApplyTracking(t *testing.T) {
	var delivered int
	notify := func(telegraf.DeliveryInfo) { delivered++ }
	input := []telegraf.Metric{
		message("fire001", "login attempt"),
		message("web01", "login attempt"),
	}
	for i, m := range input {
		input[i], _ = metric.WithTracking(m, notify)
	}

	cl := newClassify()
	require.NoError(t, cl.Init())
	actual := cl.Apply(input...)
	require.Len(t, actual, 1)
	require.Equal(t, 1, delivered, "the dropped metric must be marked as delivered")
	for _, m := range actual {
		m.Accept()
	}
	require.Equal(t, 2, delivered)
}

func TestRegistered(t *testing.T) {
	require.Contains(t, processors.Processors, "classify")
}
