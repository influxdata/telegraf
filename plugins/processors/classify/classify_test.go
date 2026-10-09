package classify

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// testLogger returns a logger that only prints in verbose mode.
func testLogger() telegraf.Logger {
	return testutil.Logger{Quiet: !testing.Verbose()}
}

// runClassifyTest is a test helper that runs Init→Start→Add(metrics)→Stop
// and returns the resulting accumulator. Failures are reported via t.
func runClassifyTest(t *testing.T, cl *Classify, metrics []telegraf.Metric, waitTime ...time.Duration) *testutil.Accumulator {
	t.Helper()
	acc := &testutil.Accumulator{}
	require.NoError(t, cl.Init())
	require.NoError(t, cl.Start(acc))
	for _, m := range metrics {
		require.NoError(t, cl.Add(m, acc))
	}
	if len(waitTime) > 0 {
		time.Sleep(waitTime[0])
	}
	cl.Stop()
	return acc
}

// TestParseFullConfig is a basic smoke test: can the plugin start and classify
// a metric end-to-end with a full, valid configuration?
func TestParseFullConfig(t *testing.T) {
	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
	}
	cl.Log = testLogger()

	m := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "WARNING:  badness happened"},
		time.Now())

	acc := runClassifyTest(t, cl, []telegraf.Metric{m})
	require.Len(t, acc.GetTelegrafMetrics(), 1)

	got := acc.GetTelegrafMetrics()[0]
	resultTag, ok := got.GetTag(cl.ResultTag)
	require.Truef(t, ok, "result tag %q not found in output metric", cl.ResultTag)
	require.Equal(t, "warning", resultTag)
}

// TestParseSelectorItem verifies all valid and invalid combinations of the
// selector_tag / selector_field options.
func TestParseSelectorItem(t *testing.T) {
	msr := map[string][]map[string]any{
		"database": {
			{"ignore": "IGNORE"}, {"okay": "OK"}, {"warning": "WARNING"},
			{"critical": "CRITICAL"}, {"unknown": ".*"},
		},
	}
	sm := []map[string]string{{`pg\d{3}`: "database"}}

	tests := []struct {
		name          string
		selectorTag   string
		selectorField string
		wantErr       bool
	}{
		{name: "no selector"},
		{name: "only selector_tag", selectorTag: "host_tag"},
		{name: "both selector_tag and selector_field", selectorTag: "host_tag", selectorField: "host_field", wantErr: true},
		{name: "only selector_field", selectorField: "host_field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorTag:           tt.selectorTag,
				SelectorField:         tt.selectorField,
				SelectorMapping:       sm,
				DefaultRegexGroup:     "database",
				MatchField:            "message",
				ResultTag:             "severity",
				MappedSelectorRegexes: msr,
			}
			cl.Log = testLogger()
			if tt.wantErr {
				require.Error(t, cl.Init())
				return
			}
			acc := &testutil.Accumulator{}
			require.NoError(t, cl.Init())
			require.NoError(t, cl.Start(acc))
			cl.Stop()
		})
	}
}

// TestParseMatchItem verifies all valid and invalid combinations of the
// match_tag / match_field options.
func TestParseMatchItem(t *testing.T) {
	msr := map[string][]map[string]any{
		"database": {
			{"ignore": "IGNORE"}, {"okay": "OK"}, {"warning": "WARNING"},
			{"critical": "CRITICAL"}, {"unknown": ".*"},
		},
	}
	sm := []map[string]string{{`pg\d{3}`: "database"}}

	tests := []struct {
		name       string
		matchTag   string
		matchField string
		wantErr    bool
	}{
		{name: "no match tag or field", wantErr: true},
		{name: "only match_tag", matchTag: "message_tag"},
		{name: "both match_tag and match_field", matchTag: "message_tag", matchField: "message_field", wantErr: true},
		{name: "only match_field", matchField: "message_field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorMapping:       sm,
				DefaultRegexGroup:     "database",
				MatchTag:              tt.matchTag,
				MatchField:            tt.matchField,
				ResultTag:             "severity",
				MappedSelectorRegexes: msr,
			}
			cl.Log = testLogger()
			if tt.wantErr {
				require.Error(t, cl.Init())
				return
			}
			acc := &testutil.Accumulator{}
			require.NoError(t, cl.Init())
			require.NoError(t, cl.Start(acc))
			cl.Stop()
		})
	}
}

// TestParseResultItem verifies all valid and invalid combinations of the
// result_tag / result_field options.
func TestParseResultItem(t *testing.T) {
	msr := map[string][]map[string]any{
		"database": {
			{"ignore": "IGNORE"}, {"okay": "OK"}, {"warning": "WARNING"},
			{"critical": "CRITICAL"}, {"unknown": ".*"},
		},
	}
	sm := []map[string]string{{`pg\d{3}`: "database"}}

	tests := []struct {
		name        string
		resultTag   string
		resultField string
		wantErr     bool
	}{
		{name: "no result tag or field", wantErr: true},
		{name: "only result_tag", resultTag: "severity_tag"},
		{name: "both result_tag and result_field", resultTag: "severity_tag", resultField: "severity_field", wantErr: true},
		{name: "only result_field", resultField: "severity_field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorMapping:       sm,
				DefaultRegexGroup:     "database",
				MatchField:            "message",
				ResultTag:             tt.resultTag,
				ResultField:           tt.resultField,
				MappedSelectorRegexes: msr,
			}
			cl.Log = testLogger()
			if tt.wantErr {
				require.Error(t, cl.Init())
				return
			}
			acc := &testutil.Accumulator{}
			require.NoError(t, cl.Init())
			require.NoError(t, cl.Start(acc))
			cl.Stop()
		})
	}
}

// TestReturnSampleConfig verifies that SampleConfig returns non-empty content.
func TestReturnSampleConfig(t *testing.T) {
	cl := &Classify{}
	require.NotEmpty(t, cl.SampleConfig(), "SampleConfig must return non-empty content")
}

// TestBadSelectorRegex verifies that invalid selector_mapping regexes are rejected.
func TestBadSelectorRegex(t *testing.T) {
	msr := map[string][]map[string]any{"database": {{"ignore": "IGNORE"}}}

	tests := []struct {
		name    string
		regex   string
		wantErr bool
	}{
		{name: "valid selector regex", regex: `pg\d{3}`},
		{name: "bad selector regex", regex: `*pg\d{3}`, wantErr: true},
		{name: "empty selector regex", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorTag:           "host",
				SelectorMapping:       []map[string]string{{tt.regex: "database"}},
				MatchField:            "message",
				ResultTag:             "severity",
				MappedSelectorRegexes: msr,
			}
			cl.Log = testLogger()
			if tt.wantErr {
				require.Error(t, cl.Init())
			} else {
				require.NoError(t, cl.Init())
			}
		})
	}
}

// TestBadCategoryRegexType verifies that all supported and unsupported value
// types for mapped_selector_regexes category entries are handled correctly.
func TestBadCategoryRegexType(t *testing.T) {
	myString := "IGNORE"

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "single string regex", value: "IGNORE"},
		{name: "multi-line string regex", value: "    IGNORE\n    DO NOT CARE\n    FUGGEDDABOUDIT\n    "},
		{name: "multi-line string with leading whitespace", value: "  \n    IGNORE\n    DO NOT CARE\n    "},
		{name: "multi-line string with no regexes", value: "    "},
		{name: "array of strings regex", value: []string{"IGNORE", "DO NOT CARE", "FUGGEDDABOUDIT"}},
		{name: "nil regex value", value: nil, wantErr: true},
		{name: "pointer-to-string regex value", value: &myString, wantErr: true},
		{name: "integer regex value", value: 42, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorTag:     "host",
				SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
				MatchField:      "message",
				ResultTag:       "severity",
				MappedSelectorRegexes: map[string][]map[string]any{
					"test-group": {{"ignore": tt.value}},
				},
			}
			cl.Log = testLogger()
			if tt.wantErr {
				require.Error(t, cl.Init())
				return
			}
			acc := &testutil.Accumulator{}
			require.NoError(t, cl.Init())
			require.NoError(t, cl.Start(acc))
			cl.Stop()
		})
	}
}

// TestBadCategoryRegex verifies that invalid category regex content is rejected.
func TestBadCategoryRegex(t *testing.T) {
	tests := []struct {
		name    string
		entries []map[string]any
	}{
		{
			name:    "duplicate category in same group",
			entries: []map[string]any{{"ignore": "foobar"}, {"ignore": "barfoo"}},
		},
		{
			name:    "bad category regex",
			entries: []map[string]any{{"ignore": "*foobar"}},
		},
		{
			name:    "empty category regex",
			entries: []map[string]any{{"ignore": ""}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorTag:     "host",
				SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
				MatchField:      "message",
				ResultTag:       "severity",
				MappedSelectorRegexes: map[string][]map[string]any{
					"test-group": tt.entries,
				},
			}
			cl.Log = testLogger()
			require.Error(t, cl.Init())
		})
	}
}

// TestSelectorMapping exercises the full range of selector_mapping behaviours.
func TestSelectorMapping(t *testing.T) {
	msr := map[string][]map[string]any{
		"database": {
			{"ignore": "IGNORE"}, {"okay": "OK"}, {"warning": "WARNING"},
			{"critical": "CRITICAL"}, {"unknown": ".*"},
		},
	}
	m := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "WARNING:  badness happened"},
		time.Now())

	tests := []struct {
		name            string
		selectorMapping []map[string]string
		defaultGroup    string
		wantCount       int
		wantInitErr     bool
	}{
		{
			name:      "no selector_mapping and no default_regex_group",
			wantCount: 0,
		},
		{
			name:         "default_regex_group names nonexistent group",
			defaultGroup: "foobar",
			wantInitErr:  true,
		},
		{
			name:         "default_regex_group names existing group",
			defaultGroup: "database",
			wantCount:    1,
		},
		{
			name:            "selector matches entry",
			selectorMapping: []map[string]string{{`pg123`: "database"}},
			wantCount:       1,
		},
		{
			name:            "selector matches nothing, no default",
			selectorMapping: []map[string]string{{`abcde`: "database"}},
			wantCount:       0,
		},
		{
			name:            "selector matches nothing, valid default",
			selectorMapping: []map[string]string{{`abcde`: "database"}},
			defaultGroup:    "database",
			wantCount:       1,
		},
		{
			name:            "selector maps to empty string",
			selectorMapping: []map[string]string{{`pg123`: ""}},
			wantCount:       0,
		},
		{
			name:            "selector maps to *, no default",
			selectorMapping: []map[string]string{{`pg123`: "*"}},
			wantCount:       0,
		},
		{
			name:            "selector maps to *, valid default",
			selectorMapping: []map[string]string{{`pg123`: "*"}},
			defaultGroup:    "database",
			wantCount:       1,
		},
		{
			name:            "selector maps to unknown group",
			selectorMapping: []map[string]string{{`pg123`: "foobar"}},
			wantCount:       0,
		},
		{
			name:            "selector uses valid regex",
			selectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
			wantCount:       1,
		},
		{
			name: "multiple ordered selector entries",
			selectorMapping: []map[string]string{
				{`fire\d{3}`: "firewall"},
				{`desk\d{3}`: "desktop"},
				{`pg\d{3}`: "database"},
			},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorTag:           "host",
				SelectorMapping:       tt.selectorMapping,
				DefaultRegexGroup:     tt.defaultGroup,
				MatchField:            "message",
				ResultTag:             "severity",
				MappedSelectorRegexes: msr,
			}
			cl.Log = testLogger()
			if tt.wantInitErr {
				require.ErrorContains(t, cl.Init(), "default_regex_group")
				return
			}
			acc := runClassifyTest(t, cl, []telegraf.Metric{m})
			require.Len(t, acc.GetTelegrafMetrics(), tt.wantCount)
		})
	}
}

// TestDefaultCategory verifies that default_category is applied when no regex
// matches, and that metrics are dropped when it is not configured.
func TestDefaultCategory(t *testing.T) {
	msr := map[string][]map[string]any{
		"database": {
			{"ignore": "IGNORE"}, {"okay": "OKAY"},
			{"warning": "WARNING"}, {"critical": "CRITICAL"}, {"unknown": "UNKNOWN"},
		},
	}
	m := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "this message contains no category name"},
		time.Now())

	tests := []struct {
		name            string
		defaultCategory string
		wantCount       int
	}{
		{name: "no default_category", wantCount: 0},
		{name: "default_category set", defaultCategory: "unmatched", wantCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &Classify{
				SelectorTag:           "host",
				SelectorMapping:       []map[string]string{{`pg\d{3}`: "database"}},
				MatchField:            "message",
				ResultTag:             "severity",
				DefaultCategory:       tt.defaultCategory,
				MappedSelectorRegexes: msr,
			}
			cl.Log = testLogger()
			acc := runClassifyTest(t, cl, []telegraf.Metric{m})
			require.Len(t, acc.GetTelegrafMetrics(), tt.wantCount)
			if tt.wantCount > 0 {
				got := acc.GetTelegrafMetrics()[0]
				resultTag, ok := got.GetTag(cl.ResultTag)
				require.Truef(t, ok, "result tag %q not found", cl.ResultTag)
				require.Equal(t, tt.defaultCategory, resultTag)
			}
		})
	}
}

// TestAggregationSummary verifies basic summary aggregation: counters are
// emitted at the end of a period and the metric has the expected shape.
func TestAggregationSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aggregation test in short mode")
	}

	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		DropCategories:  []string{"ignore", "unknown"},
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       config.Duration(5 * time.Second),
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSummaryTag:   "summary",
		AggregationSummaryValue: "full",
		AggregationSummaryFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	cl.Log = testLogger()

	m := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "nothing to see here, move along"},
		time.Now())

	acc := runClassifyTest(t, cl, []telegraf.Metric{m}, 10*time.Second)

	allMetrics := acc.GetTelegrafMetrics()
	errMsg := metricsErrMsg(allMetrics)
	require.Len(t, allMetrics, 1, errMsg)

	got := allMetrics[0]
	require.Equal(t, "status", got.Name(), errMsg)
	require.Lenf(t, got.TagList(), 1, "tag count; %s", errMsg)
	require.Lenf(t, got.FieldList(), 3, "field count; %s", errMsg)

	summaryTag, ok := got.GetTag("summary")
	require.Truef(t, ok, "summary tag missing; %s", errMsg)
	require.Equal(t, "full", summaryTag, errMsg)

	dropped, ok := got.GetField("dropped")
	require.Truef(t, ok, "dropped field missing; %s", errMsg)
	require.EqualValues(t, 1, dropped, errMsg)

	total, ok := got.GetField("total")
	require.Truef(t, ok, "total field missing; %s", errMsg)
	require.EqualValues(t, 1, total, errMsg)

	unknown, ok := got.GetField("unknown")
	require.Truef(t, ok, "unknown field missing; %s", errMsg)
	require.EqualValues(t, 1, unknown, errMsg)
}

// TestAggregationSummaryCycles verifies that counters reset between periods
// and that two successive periods produce independent metrics.
func TestAggregationSummaryCycles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aggregation test in short mode")
	}

	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		DropCategories:  []string{"ignore", "unknown"},
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       config.Duration(5 * time.Second),
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSummaryTag:   "summary",
		AggregationSummaryValue: "full",
		AggregationSummaryFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	cl.Log = testLogger()

	m1 := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "nothing to see here, move along"},
		time.Now())
	m2 := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "CRITICAL:  second message from the same host"},
		time.Now())

	acc := &testutil.Accumulator{}
	require.NoError(t, cl.Init())
	require.NoError(t, cl.Start(acc))

	require.NoError(t, cl.Add(m1, acc))
	time.Sleep(7 * time.Second)
	require.NoError(t, cl.Add(m2, acc))
	cl.Stop()

	allMetrics := acc.GetTelegrafMetrics()
	errMsg := metricsErrMsg(allMetrics)
	require.Len(t, allMetrics, 3, errMsg)

	// First metric: summary for first period (m1 dropped as unknown).
	got := allMetrics[0]
	require.Equal(t, "status", got.Name(), errMsg)
	require.Lenf(t, got.TagList(), 1, "tag count; %s", errMsg)
	require.Lenf(t, got.FieldList(), 3, "field count; %s", errMsg)
	summaryTag, ok := got.GetTag("summary")
	require.Truef(t, ok, "summary tag missing; %s", errMsg)
	require.Equal(t, "full", summaryTag, errMsg)
	dropped, ok := got.GetField("dropped")
	require.Truef(t, ok, "dropped missing; %s", errMsg)
	require.EqualValues(t, 1, dropped, errMsg)
	total, ok := got.GetField("total")
	require.Truef(t, ok, "total missing; %s", errMsg)
	require.EqualValues(t, 1, total, errMsg)

	// Second metric: m2 passed through as "critical".
	got = allMetrics[1]
	require.Equal(t, "datapoint", got.Name(), errMsg)
	require.Lenf(t, got.TagList(), 2, "tag count; %s", errMsg)
	require.Lenf(t, got.FieldList(), 1, "field count; %s", errMsg)
	hostTag, ok := got.GetTag("host")
	require.Truef(t, ok, "host tag missing; %s", errMsg)
	require.Equal(t, "pg123", hostTag, errMsg)
	resultTag, ok := got.GetTag("severity")
	require.Truef(t, ok, "severity tag missing; %s", errMsg)
	require.Equal(t, "critical", resultTag, errMsg)

	// Third metric: summary for second period (m2 classified as critical).
	got = allMetrics[2]
	require.Equal(t, "status", got.Name(), errMsg)
	require.Lenf(t, got.TagList(), 1, "tag count; %s", errMsg)
	require.Lenf(t, got.FieldList(), 2, "field count; %s", errMsg)
	summaryTag, ok = got.GetTag("summary")
	require.Truef(t, ok, "summary tag missing; %s", errMsg)
	require.Equal(t, "full", summaryTag, errMsg)
	critical, ok := got.GetField("critical")
	require.Truef(t, ok, "critical field missing; %s", errMsg)
	require.EqualValues(t, 1, critical, errMsg)
	total, ok = got.GetField("total")
	require.Truef(t, ok, "total missing; %s", errMsg)
	require.EqualValues(t, 1, total, errMsg)
}

// TestAggregationByGroup verifies per-regex-group aggregation counters.
func TestAggregationByGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aggregation test in short mode")
	}

	cl := &Classify{
		SelectorTag: "host",
		SelectorMapping: []map[string]string{
			{`pg\d{3}`: "database"},
			{`fire\d{3}`: "firewall"},
		},
		MatchField:     "message",
		DropCategories: []string{"ignore", "unknown"},
		ResultTag:      "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
			"firewall": {
				{"ignore": "IGNORE"},
				{"okay": "OKAY"},
				{"warning": "LOGIN"},
				{"critical": "INTRUSION"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       config.Duration(5 * time.Second),
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationGroupTag:     "by_machine_type",
		AggregationGroupFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	cl.Log = testLogger()

	now := time.Now()
	metrics := []telegraf.Metric{
		metric.New("datapoint",
			map[string]string{"host": "pg123"},
			map[string]any{"message": "WARNING:  situation is crazy"},
			now),
		metric.New("datapoint",
			map[string]string{"host": "pg124"},
			map[string]any{"message": "nothing to see here, move along"},
			now),
		metric.New("datapoint",
			map[string]string{"host": "fire567"},
			map[string]any{"message": "INTRUSION:  assets at risk"},
			now),
	}

	acc := runClassifyTest(t, cl, metrics, 10*time.Second)

	allMetrics := acc.GetTelegrafMetrics()
	errMsg := metricsErrMsg(allMetrics)
	require.Len(t, allMetrics, 4, errMsg)

	// First two are passthrough datapoints (order may vary).
	distinctSelector := make(map[string]int)
	for i := 0; i <= 1; i++ {
		got := allMetrics[i]
		require.Equal(t, "datapoint", got.Name(), errMsg)
		require.Lenf(t, got.TagList(), 2, "tag count; %s", errMsg)

		selectorTag, ok := got.GetTag(cl.SelectorTag)
		require.Truef(t, ok, "host tag missing; %s", errMsg)
		resultTag, ok := got.GetTag(cl.ResultTag)
		require.Truef(t, ok, "severity tag missing; %s", errMsg)

		distinctSelector[selectorTag]++
		switch selectorTag {
		case "pg123":
			require.Equal(t, "warning", resultTag, errMsg)
		case "fire567":
			require.Equal(t, "critical", resultTag, errMsg)
		default:
			require.FailNowf(t, "unexpected selector tag value %q", selectorTag)
		}
	}
	require.Len(t, distinctSelector, 2, errMsg)

	// Last two are aggregation metrics by group (order may vary).
	distinctGroup := make(map[string]int)
	for i := 2; i <= 3; i++ {
		got := allMetrics[i]
		require.Equal(t, cl.AggregationMeasurement, got.Name(), errMsg)
		require.Lenf(t, got.TagList(), 1, "tag count; %s", errMsg)

		groupTag, ok := got.GetTag(cl.AggregationGroupTag)
		require.Truef(t, ok, "group tag missing; %s", errMsg)

		distinctGroup[groupTag]++
		switch groupTag {
		case "database":
			require.Lenf(t, got.FieldList(), 4, "database field count; %s", errMsg)
			dropped, ok := got.GetField(cl.AggregationDroppedField)
			require.Truef(t, ok, "dropped missing; %s", errMsg)
			require.EqualValues(t, 1, dropped, errMsg)
			total, ok := got.GetField(cl.AggregationTotalField)
			require.Truef(t, ok, "total missing; %s", errMsg)
			require.EqualValues(t, 2, total, errMsg)
		case "firewall":
			require.Lenf(t, got.FieldList(), 2, "firewall field count; %s", errMsg)
			critical, ok := got.GetField("critical")
			require.Truef(t, ok, "critical missing; %s", errMsg)
			require.EqualValues(t, 1, critical, errMsg)
		default:
			require.FailNowf(t, "unexpected group tag value %q", groupTag)
		}
	}
	require.Len(t, distinctGroup, 2, errMsg)
}

// TestAggregationBySelector verifies per-selector aggregation counters.
func TestAggregationBySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aggregation test in short mode")
	}

	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		DropCategories:  []string{"ignore", "unknown"},
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       config.Duration(5 * time.Second),
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSelectorTag:  "by_host",
		AggregationSelectorFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	cl.Log = testLogger()

	now := time.Now()
	metrics := []telegraf.Metric{
		metric.New("datapoint",
			map[string]string{"host": "pg123"},
			map[string]any{"message": "WARNING:  situation is crazy"},
			now),
		metric.New("datapoint",
			map[string]string{"host": "pg123"},
			map[string]any{"message": "nothing to see here, move along"},
			now),
		metric.New("datapoint",
			map[string]string{"host": "pg124"},
			map[string]any{"message": "nothing to see here, move along"},
			now),
	}

	acc := runClassifyTest(t, cl, metrics, 10*time.Second)

	allMetrics := acc.GetTelegrafMetrics()
	errMsg := metricsErrMsg(allMetrics)
	require.Len(t, allMetrics, 3, errMsg)

	// First is the passthrough datapoint.
	got := allMetrics[0]
	require.Equal(t, "datapoint", got.Name(), errMsg)
	selectorTag, ok := got.GetTag(cl.SelectorTag)
	require.Truef(t, ok, "host tag missing; %s", errMsg)
	require.Equal(t, "pg123", selectorTag, errMsg)
	resultTag, ok := got.GetTag(cl.ResultTag)
	require.Truef(t, ok, "severity tag missing; %s", errMsg)
	require.Equal(t, "warning", resultTag, errMsg)

	// Remaining two are per-selector aggregation metrics (order may vary).
	distinctSelector := make(map[string]int)
	for i := 1; i <= 2; i++ {
		got = allMetrics[i]
		require.Equal(t, cl.AggregationMeasurement, got.Name(), errMsg)
		require.Lenf(t, got.TagList(), 1, "tag count; %s", errMsg)

		sel, ok := got.GetTag(cl.AggregationSelectorTag)
		require.Truef(t, ok, "selector tag missing; %s", errMsg)
		distinctSelector[sel]++

		switch sel {
		case "pg123":
			require.Lenf(t, got.FieldList(), 4, "pg123 field count; %s", errMsg)
			dropped, ok := got.GetField(cl.AggregationDroppedField)
			require.Truef(t, ok, "dropped missing; %s", errMsg)
			require.EqualValues(t, 1, dropped, errMsg)
			total, ok := got.GetField(cl.AggregationTotalField)
			require.Truef(t, ok, "total missing; %s", errMsg)
			require.EqualValues(t, 2, total, errMsg)
		case "pg124":
			require.Lenf(t, got.FieldList(), 3, "pg124 field count; %s", errMsg)
			dropped, ok := got.GetField(cl.AggregationDroppedField)
			require.Truef(t, ok, "dropped missing; %s", errMsg)
			require.EqualValues(t, 1, dropped, errMsg)
		default:
			require.FailNowf(t, "unexpected selector tag value %q", sel)
		}
	}
	require.Len(t, distinctSelector, 2, errMsg)
}

// TestAggregationDroppedAndTotalFields verifies that dropped and total counters
// track the right counts across a mix of passed and dropped metrics.
func TestAggregationDroppedAndTotalFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aggregation test in short mode")
	}

	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		DropCategories:  []string{"ignore", "unknown"},
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       config.Duration(5 * time.Second),
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSummaryTag:   "summary",
		AggregationSummaryValue: "full",
		AggregationSummaryFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	cl.Log = testLogger()

	now := time.Now()
	metrics := []telegraf.Metric{
		metric.New("datapoint",
			map[string]string{"host": "pg123"},
			map[string]any{"message": "WARNING:  situation is crazy"},
			now),
		metric.New("datapoint",
			map[string]string{"host": "pg123"},
			map[string]any{"message": "nothing to see here, move along"},
			now),
		metric.New("datapoint",
			map[string]string{"host": "pg123"},
			map[string]any{"message": "nothing to see here, move along"},
			now),
	}

	acc := runClassifyTest(t, cl, metrics, 10*time.Second)

	allMetrics := acc.GetTelegrafMetrics()
	errMsg := metricsErrMsg(allMetrics)
	require.Len(t, allMetrics, 2, errMsg)

	// First: the passed-through warning metric.
	got := allMetrics[0]
	require.Equal(t, "datapoint", got.Name(), errMsg)
	require.Lenf(t, got.TagList(), 2, "tag count; %s", errMsg)
	require.Lenf(t, got.FieldList(), 1, "field count; %s", errMsg)
	resultTag, ok := got.GetTag(cl.ResultTag)
	require.Truef(t, ok, "severity tag missing; %s", errMsg)
	require.Equal(t, "warning", resultTag, errMsg)

	// Second: the summary aggregation metric.
	got = allMetrics[1]
	require.Equal(t, cl.AggregationMeasurement, got.Name(), errMsg)
	require.Lenf(t, got.TagList(), 1, "tag count; %s", errMsg)
	require.Lenf(t, got.FieldList(), 4, "field count; %s", errMsg)

	summaryTag, ok := got.GetTag(cl.AggregationSummaryTag)
	require.Truef(t, ok, "summary tag missing; %s", errMsg)
	require.Equal(t, cl.AggregationSummaryValue, summaryTag, errMsg)

	dropped, ok := got.GetField(cl.AggregationDroppedField)
	require.Truef(t, ok, "dropped missing; %s", errMsg)
	require.EqualValues(t, 2, dropped, errMsg)

	total, ok := got.GetField(cl.AggregationTotalField)
	require.Truef(t, ok, "total missing; %s", errMsg)
	require.EqualValues(t, 3, total, errMsg)

	unknown, ok := got.GetField("unknown")
	require.Truef(t, ok, "unknown missing; %s", errMsg)
	require.EqualValues(t, 2, unknown, errMsg)

	warning, ok := got.GetField("warning")
	require.Truef(t, ok, "warning missing; %s", errMsg)
	require.EqualValues(t, 1, warning, errMsg)
}

// metricsErrMsg formats all metrics into a readable string for assertion messages.
func metricsErrMsg(metrics []telegraf.Metric) string {
	var msg strings.Builder
	msg.WriteString("output metrics:\n")
	for _, m := range metrics {
		fmt.Fprintf(&msg, "  %v\n", m)
	}
	return msg.String()
}

// TestDropCategories loads drop_categories through the TOML decoder, as a real
// configuration file does: a TOML array is decoded as []any, not []string.
func TestDropCategories(t *testing.T) {
	const base = `
selector_tag = "host"
selector_mapping = [{ "pg\\d{3}" = "database" }]
match_field = "message"
result_tag = "severity"
[mapped_selector_regexes]
  database = [
    { ignore = "IGNORE" },
    { okay = "OKAY" },
    { warning = "WARNING" },
  ]
`
	tests := []struct {
		name    string
		drop    string
		want    []string // messages that survive
		wantErr string
	}{
		{name: "string", drop: `drop_categories = "ignore"`, want: []string{"OKAY", "WARNING"}},
		{name: "array", drop: `drop_categories = ["ignore", "warning"]`, want: []string{"OKAY"}},
		{name: "empty array", drop: `drop_categories = []`, want: []string{"IGNORE", "OKAY", "WARNING"}},
		{name: "non-string element", drop: `drop_categories = [1, 2]`, wantErr: "must be a string or array of strings"},
		{name: "not a string or array", drop: `drop_categories = 1`, wantErr: "must be a string or array of strings"},
		{name: "unknown category", drop: `drop_categories = ["ignore", "bogus"]`, wantErr: `"bogus" in drop_categories`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// top-level keys must precede the [mapped_selector_regexes] table
			cl := &Classify{Log: testLogger()}
			require.NoError(t, toml.Unmarshal([]byte(tt.drop+"\n"+base), cl))
			if tt.wantErr != "" {
				require.ErrorContains(t, cl.Init(), tt.wantErr)
				return
			}
			messages := []string{"IGNORE", "OKAY", "WARNING"}
			metrics := make([]telegraf.Metric, 0, len(messages))
			for _, msg := range messages {
				metrics = append(metrics, metric.New("datapoint",
					map[string]string{"host": "pg123"},
					map[string]any{"message": msg},
					time.Now()))
			}
			acc := runClassifyTest(t, cl, metrics)
			got := make([]string, 0, len(acc.GetTelegrafMetrics()))
			for _, m := range acc.GetTelegrafMetrics() {
				msg, _ := m.GetField("message")
				got = append(got, msg.(string))
			}
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

// TestSelectorMappingFirstMatchWins verifies that the first matching
// selector_mapping entry decides the regex group.
func TestSelectorMappingFirstMatchWins(t *testing.T) {
	cl := &Classify{
		SelectorTag: "host",
		SelectorMapping: []map[string]string{
			{`pg\d+`: "database"},
			{`.*`: "generic"},
		},
		MatchField: "message",
		ResultTag:  "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {{"db": ".*"}},
			"generic":  {{"other": ".*"}},
		},
		Log: testLogger(),
	}
	m := metric.New("datapoint",
		map[string]string{"host": "pg123"},
		map[string]any{"message": "anything"},
		time.Now())

	acc := runClassifyTest(t, cl, []telegraf.Metric{m})
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	severity, ok := acc.GetTelegrafMetrics()[0].GetTag("severity")
	require.True(t, ok)
	require.Equal(t, "db", severity)
}

// TestMultiKeyCategoryEntry verifies that a mapped_selector_regexes element
// with more than one category is rejected, since its order would be random.
func TestMultiKeyCategoryEntry(t *testing.T) {
	cl := &Classify{
		MatchField: "message",
		ResultTag:  "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {{"warning": "x", "critical": "x"}},
		},
		Log: testLogger(),
	}
	require.ErrorContains(t, cl.Init(), "more than one key")
}

// aggregationConfig returns a config with one regex group whose categories
// match their own upper-cased names, and summary/group aggregation enabled.
func aggregationConfig(summaryFields, groupFields []string, includeZeroes bool) *Classify {
	return &Classify{
		DefaultRegexGroup: "database",
		MatchField:        "message",
		ResultTag:         "severity",
		DefaultCategory:   "unknown",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {{"okay": "OKAY"}, {"warning": "WARNING"}, {"critical": "CRITICAL"}},
		},
		AggregationPeriod:         config.Duration(time.Hour),
		AggregationMeasurement:    "status",
		AggregationTotalField:     "total",
		AggregationSummaryTag:     "summary",
		AggregationSummaryValue:   "full",
		AggregationSummaryFields:  summaryFields,
		AggregationGroupTag:       "host_type",
		AggregationGroupFields:    groupFields,
		AggregationIncludesZeroes: includeZeroes,
		Log:                       testLogger(),
	}
}

// aggregate classifies the given messages and flushes one aggregation period
// directly, without starting the aggregation goroutine.
func aggregate(t *testing.T, cl *Classify, messages ...string) []telegraf.Metric {
	t.Helper()
	require.NoError(t, cl.Init())
	acc := &testutil.Accumulator{}
	cl.acc = acc
	for _, msg := range messages {
		require.NoError(t, cl.Add(metric.New("datapoint", nil,
			map[string]any{"message": msg}, time.Now()), acc))
	}
	acc.ClearMetrics()
	cl.outputAggregationData(time.Unix(0, 0))
	return acc.GetTelegrafMetrics()
}

// TestAggregationFieldsFilter verifies that aggregation output carries only the
// configured fields, with zeroes added on request for every aggregation type.
func TestAggregationFieldsFilter(t *testing.T) {
	tests := []struct {
		name          string
		includeZeroes bool
		want          []telegraf.Metric
	}{
		{
			name: "without zeroes",
			want: []telegraf.Metric{
				metric.New("status", map[string]string{"summary": "full"},
					map[string]any{"warning": 2}, time.Unix(0, 0), telegraf.Counter),
				metric.New("status", map[string]string{"host_type": "database"},
					map[string]any{"warning": 2}, time.Unix(0, 0), telegraf.Counter),
			},
		},
		{
			name:          "with zeroes",
			includeZeroes: true,
			want: []telegraf.Metric{
				metric.New("status", map[string]string{"summary": "full"},
					map[string]any{"warning": 2}, time.Unix(0, 0), telegraf.Counter),
				metric.New("status", map[string]string{"host_type": "database"},
					map[string]any{"warning": 2, "critical": 0}, time.Unix(0, 0), telegraf.Counter),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := aggregationConfig([]string{"warning"}, []string{"warning", "critical"}, tt.includeZeroes)
			got := aggregate(t, cl, "WARNING", "WARNING", "OKAY")
			testutil.RequireMetricsEqual(t, tt.want, got, testutil.SortMetrics())
		})
	}
}

// TestAggregationAllZeroSuppressed verifies that a point whose configured
// fields are all zero is not emitted, even with aggregation_includes_zeroes.
func TestAggregationAllZeroSuppressed(t *testing.T) {
	cl := aggregationConfig([]string{"critical"}, []string{"critical"}, true)
	require.Empty(t, aggregate(t, cl, "OKAY", "WARNING"))
}

// TestAggregationDefaultCategoryField verifies that default_category may be
// listed as an aggregation field and is counted.
func TestAggregationDefaultCategoryField(t *testing.T) {
	cl := aggregationConfig([]string{"unknown", "total"}, []string{"unknown"}, false)
	want := []telegraf.Metric{
		metric.New("status", map[string]string{"summary": "full"},
			map[string]any{"unknown": 1, "total": 2}, time.Unix(0, 0), telegraf.Counter),
		metric.New("status", map[string]string{"host_type": "database"},
			map[string]any{"unknown": 1}, time.Unix(0, 0), telegraf.Counter),
	}
	got := aggregate(t, cl, "no match here", "OKAY")
	testutil.RequireMetricsEqual(t, want, got, testutil.SortMetrics())
}

// TestAggregationPartialConfig verifies that half-set aggregation options are
// rejected, while omitting aggregation_period alone disables aggregation.
func TestAggregationPartialConfig(t *testing.T) {
	fields := []string{"okay"}
	tests := []struct {
		name    string
		modify  func(cl *Classify)
		wantErr string
	}{
		{name: "group tag without fields", modify: func(cl *Classify) { cl.AggregationGroupFields = nil },
			wantErr: "aggregation_group_tag and aggregation_group_fields"},
		{name: "selector fields without tag", modify: func(cl *Classify) { cl.AggregationSelectorFields = fields },
			wantErr: "aggregation_selector_tag and aggregation_selector_fields"},
		{name: "summary fields without tag", modify: func(cl *Classify) {
			cl.AggregationSummaryTag, cl.AggregationSummaryValue = "", ""
		}, wantErr: "aggregation_summary_tag and aggregation_summary_fields"},
		{name: "period without measurement", modify: func(cl *Classify) { cl.AggregationMeasurement = "" },
			wantErr: "aggregation_measurement must be set"},
		{name: "no period disables aggregation", modify: func(cl *Classify) { cl.AggregationPeriod = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := aggregationConfig(fields, fields, false)
			tt.modify(cl)
			if tt.wantErr != "" {
				require.ErrorContains(t, cl.Init(), tt.wantErr)
				return
			}
			require.NoError(t, cl.Init())
			require.Empty(t, cl.aggregators)
		})
	}
}

// panickingAccumulator panics on its first aggregation output only.
type panickingAccumulator struct {
	testutil.Accumulator
	panicked atomic.Bool
}

func (a *panickingAccumulator) AddCounter(measurement string, fields map[string]any, tags map[string]string, t ...time.Time) {
	if a.panicked.CompareAndSwap(false, true) {
		panic("boom")
	}
	a.Accumulator.AddCounter(measurement, fields, tags, t...)
}

// TestAggregationSurvivesPanic verifies that a panic while emitting
// aggregation data does not stop later periods from being emitted, and that
// Stop still returns.
func TestAggregationSurvivesPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping aggregation test in short mode")
	}
	cl := aggregationConfig([]string{"okay"}, nil, false)
	cl.AggregationGroupTag = ""
	cl.AggregationPeriod = config.Duration(time.Second)
	require.NoError(t, cl.Init())

	acc := &panickingAccumulator{}
	require.NoError(t, cl.Start(acc))
	okay := func() telegraf.Metric {
		return metric.New("datapoint", nil, map[string]any{"message": "OKAY"}, time.Now())
	}
	require.NoError(t, cl.Add(okay(), acc))
	require.Eventually(t, acc.panicked.Load, 3*time.Second, 50*time.Millisecond)

	require.NoError(t, cl.Add(okay(), acc))
	require.Eventually(t, func() bool { return acc.NMetrics() > 0 }, 3*time.Second, 50*time.Millisecond,
		"no aggregation output after the panic")

	stopped := make(chan struct{})
	go func() {
		cl.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		require.FailNow(t, "Stop did not return")
	}
}

// TestNoUsableRegexGroup verifies that a config without a selector must name
// a default_regex_group.
func TestNoUsableRegexGroup(t *testing.T) {
	cl := &Classify{
		MatchField: "message",
		ResultTag:  "severity",
		MappedSelectorRegexes: map[string][]map[string]any{
			"database": {{"okay": "OKAY"}},
		},
		Log: testLogger(),
	}
	require.ErrorContains(t, cl.Init(), "default_regex_group must be set")
}

// TestUnknownGroupNotAggregated verifies that a selector value passed through
// by "*" that names no regex group does not become a per-group aggregation key.
func TestUnknownGroupNotAggregated(t *testing.T) {
	cl := aggregationConfig([]string{"total"}, []string{"total"}, false)
	cl.DefaultRegexGroup = ""
	cl.SelectorTag = "host"
	cl.SelectorMapping = []map[string]string{{".*": "*"}}
	require.NoError(t, cl.Init())
	acc := &testutil.Accumulator{}
	cl.acc = acc
	for _, host := range []string{"database", "web01"} {
		require.NoError(t, cl.Add(metric.New("datapoint", map[string]string{"host": host},
			map[string]any{"message": "OKAY"}, time.Now()), acc))
	}
	acc.ClearMetrics()
	cl.outputAggregationData(time.Unix(0, 0))

	want := []telegraf.Metric{
		metric.New("status", map[string]string{"summary": "full"},
			map[string]any{"total": 2}, time.Unix(0, 0), telegraf.Counter),
		metric.New("status", map[string]string{"host_type": "database"},
			map[string]any{"total": 1}, time.Unix(0, 0), telegraf.Counter),
	}
	testutil.RequireMetricsEqual(t, want, acc.GetTelegrafMetrics(), testutil.SortMetrics())
}

// TestEmptyCategoryName verifies that an empty category name is rejected.
func TestEmptyCategoryName(t *testing.T) {
	cl := aggregationConfig(nil, nil, false)
	cl.AggregationSummaryTag, cl.AggregationSummaryValue, cl.AggregationGroupTag = "", "", ""
	cl.MappedSelectorRegexes["database"] = append(cl.MappedSelectorRegexes["database"],
		map[string]any{"": "timeout"})
	require.ErrorContains(t, cl.Init(), "empty category name")
}

// TestSelectorAggregationWithoutSelector verifies that selector aggregation
// requires a selector to bin by.
func TestSelectorAggregationWithoutSelector(t *testing.T) {
	cl := aggregationConfig([]string{"okay"}, []string{"okay"}, false)
	cl.AggregationSelectorTag = "host"
	cl.AggregationSelectorFields = []string{"okay"}
	require.ErrorContains(t, cl.Init(), "aggregation_selector_tag requires selector_tag or selector_field")

	cl.SelectorTag = "host"
	require.NoError(t, cl.Init())
	require.Len(t, cl.aggregators, 3)
}

// TestCategoryWithoutRegexes verifies that a category declared without regexes
// never matches but may still be named in drop_categories and aggregation fields.
func TestCategoryWithoutRegexes(t *testing.T) {
	cl := aggregationConfig([]string{"warning", "placeholder", "total"}, []string{"total"}, false)
	cl.MappedSelectorRegexes["database"] = append(cl.MappedSelectorRegexes["database"],
		map[string]any{"placeholder": make([]any, 0)})
	cl.DropCategories = "placeholder"
	want := []telegraf.Metric{
		metric.New("status", map[string]string{"summary": "full"},
			map[string]any{"warning": 1, "total": 1}, time.Unix(0, 0), telegraf.Counter),
		metric.New("status", map[string]string{"host_type": "database"},
			map[string]any{"total": 1}, time.Unix(0, 0), telegraf.Counter),
	}
	got := aggregate(t, cl, "WARNING")
	testutil.RequireMetricsEqual(t, want, got, testutil.SortMetrics())
}
