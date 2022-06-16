// Unit-test routines for the Telegraf "classify" combined processor/aggregator.

package classify

// The tests in this file are designed to verify many aspects of the Telegraf
// "classify" processor plugin.  There are a lot of little things we need to get
// right, so the number of tests included here is much larger than you might see
// in other plugins.

// In this iteration of testing, some collections of sub-tests are combined
// into a single test for the general aspect under scrutiny.  Whether we
// might want to break those up into individual tests is open to opinion.

// Some other aspects of plugin operation not explicitly tested here:
// * logging in general; see [agent] options for that:
//   * where logging output ends up
//   * control of logging levels actually output
//   * log rotation parameters

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log_separator = "----------------------------------------------------------------"

// Routine to be called whenever there is a proposed test or part of a test
// that is not yet implemented, so we don't lose track of the fact that the
// work here is not yet done.
func NotImplemented(t *testing.T, s ...string) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	reg := regexp.MustCompile(`.*/(.*)`)
	function := reg.ReplaceAllString(frame.Function, "${1}")
	if len(s) > 0 {
		t.Fatalf("Test not implemented: %s (%s)", function, s[0])
	} else {
		t.Fatalf("Test not implemented: %s", function)
	}
}

// It would be absurd for us to define a large number of tests and not
// factor out their basic commonality.  Here is that boilerplate.
func RunClassifyTest(t *testing.T, cl *Classify, metrics []telegraf.Metric, wait_time ...time.Duration) (acc *testutil.Accumulator, err error) {
	// If the test panics, let's get the word out to where we can see it,
	// with full details displayed, so we can easily debug the problem.
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("panic: %v\n", p)
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	err = cl.Reset()
	if err != nil {
		err = fmt.Errorf("the Classify object could not be reset:\n%v", err)
		return nil, err
	}
	acc = &testutil.Accumulator{}
	err = cl.Start(acc)
	if err != nil {
		// This particular case, along with any later error return from
		// RunClassifyTest(), risks the cl.wait_group being left with a
		// non-zero count, which (because we have no way to clear that count
		// in a future call to cl.Reset()) will then make it impossible for
		// this particular *Classify object to successfully wait for the
		// aggregation thread to complete in any future tests -- game over.
		// There's not much we can do about that.
		err = fmt.Errorf("the classify plugin could not be started:\n%v", err)
		return acc, err
	}
	for _, metric := range metrics {
		err = cl.Add(metric, acc)
		if err != nil {
			err = fmt.Errorf("a metric could not be added to the accumulator:\n%v", err)
			return acc, err
		}
	}
	if len(wait_time) > 0 {
		time.Sleep(wait_time[0])
	}
	err = cl.Stop()
	if err != nil {
		err = fmt.Errorf("the classify plugin could not be stopped:\n%v", err)
		return acc, err
	}
	return acc, err
}

/*
// Test the internal TOML parsing of a sample detailed config file.
// This is worth implementing but has not been approached yet.
func TestConfigDetailParsing(t *testing.T) {
	NotImplemented(t, "pending some thought")
}
*/

// This is a basic smoke test.  Is the plugin able to get up and running
// at all?  Are we able to parse a config file that includes a full set
// of non-conflicting options, with as much variation of style and content
// as we can stuff into a single configuration?
func TestParseFullConfig(t *testing.T) {
	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "WARNING:  badness happened",
	}, now)
	metrics := make([]telegraf.Metric, 1)
	metrics[0] = m

	acc, err := RunClassifyTest(t, cl, metrics)
	require.NoError(t, err)
	require.Len(t, acc.GetTelegrafMetrics(), 1)

	processedMetric := acc.GetTelegrafMetrics()[0]
	result_tag, ok := processedMetric.GetTag(cl.ResultTag)
	require.Truef(t, ok, "could not find result tag %q in the returned metric", cl.ResultTag)
	require.EqualValuesf(t, "warning", result_tag, "result tag %q value was not %q; output metric is:\n%v\n",
		cl.ResultTag, "warning", processedMetric)
}

// Make sure that all variants of specifying a selector item work as desired.
// * PASS:  No selector tag or selector field defined.
// * PASS:  Only selector tag defined.
// * FAIL:  Both selector tag and selector field defined.
// * PASS:  Only selector field defined.
func TestParseSelectorItem(t *testing.T) {
	cl := &Classify{
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
	}

	acc := &testutil.Accumulator{}
	var err error

	// No selector tag or selector field.
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  no selector tag or selector field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Only selector tag.
	cl.SelectorTag = "host_tag"
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  only selector tag")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Both selector tag and selector field.
	cl.SelectorField = "host_field"
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  both selector tag and selector field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Only selector field.
	cl.SelectorTag = ""
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  only selector field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")
}

// Make sure that all variants of specifying a match item work as desired.
// * FAIL:  No match tag or match field defined.
// * PASS:  Only match tag defined.
// * FAIL:  Both match tag and match field defined.
// * PASS:  Only match field defined.
func TestParseMatchItem(t *testing.T) {
	cl := &Classify{
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
	}

	acc := &testutil.Accumulator{}
	var err error

	// No match tag or match field.
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  no match tag or match field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Only match tag.
	cl.MatchTag = "message_tag"
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  only match tag")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Both match tag and match field.
	cl.MatchField = "message_field"
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  both match tag and match field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Only match field.
	cl.MatchTag = ""
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  only match field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")
}

// Make sure that all variants of specifying a result item work as desired.
// * FAIL:  No result tag or result field defined.
// * PASS:  Only result tag defined.
// * FAIL:  Both result tag and result field defined.
// * PASS:  Only result field defined.
func TestParseResultItem(t *testing.T) {
	cl := &Classify{
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
	}

	acc := &testutil.Accumulator{}
	var err error

	// No result tag or result field.
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  no result tag or result field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Only result tag.
	cl.ResultTag = "severity_tag"
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  only result tag")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Both result tag and result field.
	cl.ResultField = "severity_field"
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  both result tag and result field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Only result field.
	cl.ResultTag = ""
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  only result field")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")
}

// Test the ability of the plugin to output a sample configuration file.
func TestReturnSampleConfig(t *testing.T) {
	cl := &Classify{}
	sample_config := cl.SampleConfig()
	require.Len(t, sample_config, 275, "length of the sample configuration is not as expected")
	require.Contains(t, sample_config, "detailed configuration data for the classify plugin",
		"content of sample configuration is not as expected")
}

/*
// Run a a basic end-to-end smoke test, as a means of proving out the
// overall logic flow rather than any particular aspect of it.
// (The question is, would such a test be useful in some way that we
// have not already dealt with in other tests?)
func TestMatchMetric(t *testing.T) {
	NotImplemented(t, "pending some thought")
}
*/

// Test various forms of selector mapping, both valid and invalid, at this point
// consolidated to run all in one overall test:
//
// * no selector_mapping at all is provided, and no default_regex_group is defined
// * XXXX:  No selector mapping at all is defined.                             (selector item value should be used unchanged)

// * no selector_mapping at all is provided, and a default_regex_group is provided
//   that does not name one of the mapped_selector_regexes groups
// * XXXX:  No selector mapping at all is defined.                             (selector item value should be used unchanged)

// * no selector_mapping at all is provided, and a default_regex_group is provided
//   that names one of the mapped_selector_regexes groups
// * XXXX:  No selector mapping at all is defined.                             (selector item value should be used unchanged)

// * XXXX:  Selector map regex is an empty string.                             (the configuration should be rejected)
// * XXXX:  Selector map regex is a non-empty literal string.                  (this alone is not any kind of special condition)
// * XXXX:  Selector map regex is unparseable as a regex.                      (the configuration should be rejected)
// * XXXX:  Selector map regex is a valid regex.                               (this alone is not any kind of special condition)
// * XXXX:  Selector matches; map value is an empty string.                    (input data point should be dropped)
// * XXXX:  Selector matches; map value is the special string "*".             (selector item value should be used unchanged)
// * XXXX:  Selector matches; map value matches some match-regex group.        (use the mapped selector item value as the regex group)
// * XXXX:  Selector matches; map value does not match any match-regex group.  (SPECIAL HANDLING:  use the last configured regex group)
// * XXXX:  Selector mapping is non-empty; selector does not match any key.    (SPECIAL HANDLING:  use the last configured regex group)

// * selector item value matches a selector_mapping entry, and no default_regex_group is provided
// * selector item value does not match any selector_mapping entry, and no default_regex_group is provided
// * selector item value does not match any selector_mapping entry, and a default_regex_group
//   is provided that names one of the mapped_selector_regexes groups
// * selector item maps to an empty string
// * selector item maps to '*', and there is no default regex group in play
// * selector item maps to '*', and there is a valid default regex group in play
// * selector item maps to some value which is not one of the mapped_selector_regexes groups
// * selector map uses a valid regex for matching
// * selector map uses Listed Order (with multiple mapping elements)
//
// In all of these tests, we ought to somehow be able to test the result of just
// the mapping, not just the whole-metric-matching result.  That said, I'm not
// sure there is any way to do that, unless we capture intermediate results in
// Classify{} structure elements for test purposes, so they can be examined by
// unit-test code.  We would also need to be careful to zero out those elements
// between metrics, so as not to get confused by the intermediate results from
// earlier metrics.
//
func TestSelectorMapping(t *testing.T) {
	cl := &Classify{
		SelectorTag: "host",
		MatchField:  "message",
		ResultTag:   "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "WARNING:  badness happened",
	}, now)
	metrics := make([]telegraf.Metric, 1)
	metrics[0] = m

	var acc *testutil.Accumulator
	var err error

	// In theory, we could use better after-action tests here, to verify whether the
	// expected regex group got used when the data point was not dropped.  However,
	// in these tests we are only supplying one regex group, so that is effectively
	// checked for us even though we don't have any direct mechanism for detecting
	// that internal transient result of the processing.

	// NOTE:  Some of the log messages recorded here are out of date, but they give
	// the general flavor of what gets logged.

	// * no selector_mapping at all is provided (meaning, the user did not define
	//   selector_mapping in the config file, so cl.SelectorMapping will be nil),
	//   and no default_regex_group is defined
	// get back in the log:
	// selector item value "pg123" does not match anything in the selector_mapping
	// dropping point (selector item value "pg123" maps to an empty string)
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  no selector_mapping at all is provided, and no default_regex_group is defined")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * no selector_mapping at all is provided, and a default_regex_group is provided
	//   that does not name one of the mapped_selector_regexes groups
	// get back in the log:
	// selector item value "pg123" does not match anything in the selector_mapping
	// selector item value "pg123" maps to "foobar", which does not match any mapped_selector_regexes group
	// dropping point (selector item mapped value "foobar" does not match any mapped_selector_regexes group)
	cl.DefaultRegexGroup = "foobar"
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  no selector_mapping at all is provided, and default_regex_group does not name an existing group")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * no selector_mapping at all is provided, and a default_regex_group is provided
	//   that names one of the mapped_selector_regexes groups
	// get back in the log:
	// selector item value "pg123" does not match anything in the selector_mapping
	cl.DefaultRegexGroup = "database"
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  no selector_mapping at all is provided, and default_regex_group names an existing group")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item value matches a selector_mapping entry, and no default_regex_group is provided
	// get back in the log:
	// (nothing additional logged for this sub-test)
	cl.SelectorMapping = []map[string]string{{`pg123`: "database"}}
	cl.DefaultRegexGroup = ""
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item value matches a selector_mapping entry")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item value does not match any selector_mapping entry, and no default_regex_group is provided
	// get back in the log:
	// selector item value "pg123" does not match anything in the selector_mapping
	// dropping point (selector item value "pg123" maps to an empty string)
	cl.SelectorMapping = []map[string]string{{`abcde`: "database"}}
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item value does not match any selector_mapping entry")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item value does not match any selector_mapping entry, and a default_regex_group
	//   is provided that names one of the mapped_selector_regexes groups
	// get back in the log:
	// selector item value "pg123" does not match anything in the selector_mapping
	cl.SelectorMapping = []map[string]string{{`abcde`: "database"}}
	cl.DefaultRegexGroup = "database"
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item value does not match any selector_mapping entry")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item maps to an empty string
	// get back in the log:
	// dropping point (selector item value "pg123" maps to an empty string)
	cl.SelectorMapping = []map[string]string{{`pg123`: ""}}
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item maps to an empty string")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item maps to '*', and there is no default regex group in play
	// get back in the log:
	// selector item value "pg123" maps to "pg123", which does not match any mapped_selector_regexes group
	// dropping point (selector item mapped value "pg123" does not match any mapped_selector_regexes group)
	cl.SelectorMapping = []map[string]string{{`pg123`: "*"}}
	cl.DefaultRegexGroup = ""
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item maps to '*', and there is no default regex group in play")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item maps to '*', and there is a valid default regex group in play
	// get back in the log:
	// selector item value "pg123" maps to "pg123", which does not match any mapped_selector_regexes group
	// attempting category regex matches
	// selector item value "pg123" mapped to regex group "database", which has 5 categories
	// matching category "ignore", which has 1 regexes
	// matching against regex "IGNORE"
	// matching category "okay", which has 1 regexes
	// matching against regex "OK"
	// matching category "warning", which has 1 regexes
	// matching against regex "WARNING"
	// found match
	// matched category "warning"
	// setting result tag "severity" to "warning"
	cl.SelectorMapping = []map[string]string{{`pg123`: "*"}}
	cl.DefaultRegexGroup = "database"
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item maps to '*', and there is a valid default regex group in play")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector item maps to some value which is not one of the mapped_selector_regexes groups
	// get back in the log:
	// dropping point (selector item value "pg123" maps to "foobar", which does not match any mapped_selector_regexes group)
	cl.SelectorMapping = []map[string]string{{`pg123`: "foobar"}}
	cl.DefaultRegexGroup = ""
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector item maps to some value which is not one of the mapped_selector_regexes groups")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector map uses a valid regex for matching
	// get back in the log:
	// (nothing additional logged for this sub-test)
	cl.SelectorMapping = []map[string]string{{`pg\d{3}`: "database"}}
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector map uses a valid regex for matching")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	// * selector map uses Listed Order (with multiple mapping elements)
	//   (well, because of difficulty in parsing polymorphic forms of this
	//   mapping, we have not restricted ourselves to only supporting the
	//   Listed Order format, but at least here in this test we supply more
	//   than one mapping)
	// get back in the log:
	// (nothing additional logged for this sub-test)
	cl.SelectorMapping = []map[string]string{
		{`fire\d{3}`: "firewall"},
		{`desk\d{3}`: "desktop"},
		{`pg\d{3}`: "database"},
	}
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  selector map uses Listed Order")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}

// Test what happens if the user supplies an invalid selector regex.
func TestBadSelectorRegex(t *testing.T) {
	cl := &Classify{
		SelectorTag: "host",
		MatchField:  "message",
		ResultTag:   "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
			},
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	acc := &testutil.Accumulator{}

	// Test the behavior if a selector_mapping pattern compiles as a regex.
	cl.SelectorMapping = []map[string]string{{`pg\d{3}`: "database"}}
	err := cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  good selector mapping regex")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Test the behavior if a selector_mapping pattern won't compile as a regex.
	cl.SelectorMapping = []map[string]string{{`*pg\d{3}`: "database"}}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad selector_mapping regex [bad repetition]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Test the behavior if an empty selector_mapping pattern is specified.
	cl.SelectorMapping = []map[string]string{{``: "database"}}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad selector_mapping regex [empty regex]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")
}

// Test the error handling if a category regex is not specified as either a
// single string, a multi-line string, or an array of strings -- perhaps
// the user specifies some other TOML type, like an integer, instead.
func TestBadCategoryRegexType(t *testing.T) {
	cl := &Classify{
		SelectorTag:           "host",
		SelectorMapping:       []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:            "message",
		ResultTag:             "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	acc := &testutil.Accumulator{}

	// Single string.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": "IGNORE"},
	}
	err := cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  good category regex type [single-string]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Multi-line string.
	// Here we simulate what we will get back from a TOML Multi-line literal
	// string that the user has specified using whitespace indentation for
	// regexes and placing the terminating delimeter on its own line.
	// Our plugin's internal parsing of that string, taking it apart into
	// its constituent regexes, should be able to hand this.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": "    IGNORE\n    DO NOT CARE\n    FUGGEDDABOUDIT\n    "},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  good category regex type [multi-line string]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Multi-line string, containing some invisible whitespace immediately
	// after the opening delimiter in a multi-line literal string, before
	// the newline at the end of that line.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": "  \n    IGNORE\n    DO NOT CARE\n    FUGGEDDABOUDIT\n    "},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  good category regex type [multi-line string with invisible leading whitespace]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Multi-line string, containing just the whitespace at the beginning of a line.
	// This would be the case if the user used the form of a multi-line literal
	// string, but did not include any regexes within that string, so there is
	// only the opening delimiter (and its following newline, which is suppressed),
	// and whitespace on the next line before the closing delimiter.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": "    "},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  good category regex type [multi-line string containing no regexes]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Array of strings.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": []string{"IGNORE", "DO NOT CARE", "FUGGEDDABOUDIT"}},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "subtest:  good category regex type [array-of-strings]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": nil},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad category regex type [nil-value]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	my_string := "IGNORE"
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": &my_string},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad category regex type [ptr-to-string]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": 42},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad category regex type [integer]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")
}

// Test what happens if the user supplies an invalid category regex.
func TestBadCategoryRegex(t *testing.T) {
	cl := &Classify{
		SelectorTag:           "host",
		SelectorMapping:       []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:            "message",
		ResultTag:             "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	acc := &testutil.Accumulator{}

	// Test the behavior if a duplicate category is listed for a single
	// mapped_selector_regexes group.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": "foobar"},
		{"ignore": "barfoo"},
	}
	err := cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  duplicate category for the same mapped_selector_regexes group [bad repetition]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Test the behavior if a mapped_selector_regexes regex won't compile as a regex.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": "*foobar"},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad mapped_selector_regexes regex [bad repetition]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// Test the behavior if an empty mapped_selector_regexes regex is specified.
	cl.MappedSelectorRegexes["test-group"] = []map[string]interface{}{
		{"ignore": ""},
	}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.Error(t, err, "subtest:  bad mapped_selector_regexes regex [empty regex]")
	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")
}

// Test whatever it is we want to do with the default_category option.
// The only sensible thing is to have an input data point that does not
// match any category regexes, then test what happens with and without
// the default_category option defined.  If it is not defined, the data
// point should be dropped.  If it is defined as a non-empty string, the
// default_category option value should be used as the result item value.
func TestDefaultCategory(t *testing.T) {
	cl := &Classify{
		SelectorTag:     "host",
		SelectorMapping: []map[string]string{{`pg\d{3}`: "database"}},
		MatchField:      "message",
		ResultTag:       "severity",
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OKAY"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": "UNKNOWN"},
			},
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "this message contains no category name",
	}, now)
	metrics := make([]telegraf.Metric, 1)
	metrics[0] = m

	acc, err := RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  default_category is not supplied")
	require.Len(t, acc.GetTelegrafMetrics(), 0)
	if cl.logger != nil {
		cl.logger.Info(log_separator)
	}

	cl.DefaultCategory = "unmatched"
	acc, err = RunClassifyTest(t, cl, metrics)
	require.NoError(t, err, "subtest:  default_category is supplied as a non-empty string")
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	processedMetric := acc.GetTelegrafMetrics()[0]
	result_tag, ok := processedMetric.GetTag(cl.ResultTag)
	require.Truef(t, ok, "could not find result tag %q in the returned metric", cl.ResultTag)
	require.EqualValuesf(t, cl.DefaultCategory, result_tag, "result tag %q value was not %q", cl.ResultTag, cl.DefaultCategory)
}

/*
// Test various ways in which a metric might be dropped.
// * metric matched, but its match category was in drop_categories
// * (list other cases as well, as they occur to me)
func TestDroppedMetric(t *testing.T) {
	NotImplemented(t, "pending some thought")
}
*/

/*
// Check that the plugin shuts down smoothly, both with and without
// aggregation in play.  We should be able to see some logging output
// that indicates certain code paths have been executed.
func TestStopPlugin(t *testing.T) {
	NotImplemented(t, "pending some thought")
}
*/

// To keep the test execution time within sensible limits, we use a small
// aggregation_period for all aggregation testing.  That said, to allow
// cutting down even further, all aggregation tests will be skipped in
// short-test mode.

// Test essential operation of an aggregation summary.  Also test varying
// the set of output fields listed in the aggregation_summary_fields option.
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
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       "5s",
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSummaryTag:   "summary",
		AggregationSummaryValue: "full",
		AggregationSummaryFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)
	metrics := make([]telegraf.Metric, 1)
	metrics[0] = m

	// Our configured aggregation_period is one minute, so if this code
	// does not either wait for that interval to expire or force the
	// aggregation thread to shut down early and flush its data, we
	// will only get back the input data point, not the aggregation-data
	// metric as well.
	wait_duration, err := time.ParseDuration("10s")
	require.NoError(t, err)
	acc, err := RunClassifyTest(t, cl, metrics, wait_duration)
	require.NoError(t, err)

	// The original input data point should be dropped.
	// What we get back instead should be just the summary metric.
	//
	// For error reporting, if we have any, we dump out all the accumulator
	// items one by one on separate lines into a more descriptive error
	// message, not all in one run-on sentence that is hard to read.
	//
	all_metrics := acc.GetTelegrafMetrics()
	err_msg := "output metrics are:\n"
	for _, output_metric := range all_metrics {
		err_msg += fmt.Sprintf("%v\n", output_metric)
	}
	require.Equal(t, 1, len(all_metrics), err_msg)

	// At this point, we should have (except for a different timestamp value, of course):
	// status map[summary:full] map[dropped:1 total:1 unknown:1] 1655615640000241981
	processedMetric := all_metrics[0]

	measurement := processedMetric.Name()
	require.Equal(t, "status", measurement, err_msg)

	tag_count := len(processedMetric.TagList())
	require.EqualValuesf(t, 1, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count := len(processedMetric.FieldList())
	require.EqualValuesf(t, 3, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	summary_tag, ok := processedMetric.GetTag("summary")
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", "summary", err_msg)
	require.EqualValuesf(t, "full", summary_tag, "tag %q value was not %q; %s", "summary", "full", err_msg)

	dropped_field, ok := processedMetric.GetField("dropped")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "dropped", err_msg)
	require.EqualValuesf(t, 1, dropped_field, "field %q value was not %q: %s", "dropped", 1, err_msg)

	total_field, ok := processedMetric.GetField("total")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "total", err_msg)
	require.EqualValuesf(t, 1, total_field, "field %q value was not %q; %s", "total", 1, err_msg)

	unknown_field, ok := processedMetric.GetField("unknown")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "unknown", err_msg)
	require.EqualValuesf(t, 1, unknown_field, "field %q value was not %q; %s", "unknown", 1, err_msg)
}

// Make sure that aggregation counters get cleared when an aggregation period
// expires.  Also test the clock phase of the reported timestamps, to see if
// we can get an implementation that syncs up with "natural" boundaries given
// whatever aggregation_period you have specified.
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
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       "5s",
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSummaryTag:   "summary",
		AggregationSummaryValue: "full",
		AggregationSummaryFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m1 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)

	now = time.Now()
	m2 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "CRITICAL:  second message from the same host",
	}, now)

	// Our configured aggregation_period is one minute, so if this code
	// does not either wait for that interval to expire or force the
	// aggregation thread to shut down early and flush its data, we
	// will only get back the input data point, not the aggregation-data
	// metric as well.
	wait_duration, err := time.ParseDuration("7s")
	require.NoError(t, err)

	acc := &testutil.Accumulator{}
	err = cl.Reset()
	require.NoError(t, err, "the Classify object could not be reset")
	err = cl.Start(acc)
	require.NoError(t, err, "the classify plugin could not be started")

	err = cl.Add(m1, acc)
	require.NoError(t, err, "a metric could not be added to the accumulator")

	time.Sleep(wait_duration)

	err = cl.Add(m2, acc)
	require.NoError(t, err, "a metric could not be added to the accumulator")

	// The cl.Stop() call should shut down the aggregation thread and
	// effectively flush all the pending aggregation counters to an
	// output metric, before the current aggregation_period has expired.

	err = cl.Stop()
	require.NoError(t, err, "the classify plugin could not be stopped")

	// One of the original input data points should have been dropped.  What
	// we get back instead for that data point should be just the summary line.
	//
	// The other input data point should be classified as expected.
	//
	// The overall effect is that we should end up with 3 output points,
	// 1 original and 2 summary.
	//
	// For error reporting, if we have any, we dump out all the accumulator
	// items one by one on separate lines into a more descriptive error
	// message, not all in one run-on sentence that is hard to read.
	//
	all_metrics := acc.GetTelegrafMetrics()
	err_msg := "output metrics are:\n"
	for _, output_metric := range all_metrics {
		err_msg += fmt.Sprintf("%v\n", output_metric)
	}
	require.Equal(t, 3, len(all_metrics), err_msg)

	// At this point, we should have (except for different timestamp values, of course):
	// status map[summary:full] map[dropped:1 total:1 unknown:1] 1655617150000240511
	// datapoint map[host:pg123 severity:critical] map[message:CRITICAL:  second message from the same host] 1655617149411068742
	// status map[summary:full] map[critical:1 total:1] 1655617156413777891

	processedMetric := all_metrics[0]

	measurement := processedMetric.Name()
	require.Equal(t, "status", measurement, err_msg)

	tag_count := len(processedMetric.TagList())
	require.EqualValuesf(t, 1, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count := len(processedMetric.FieldList())
	require.EqualValuesf(t, 3, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	summary_tag, ok := processedMetric.GetTag("summary")
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", "summary", err_msg)
	require.EqualValuesf(t, "full", summary_tag, "tag %q value was not %q; %s", "summary", "full", err_msg)

	dropped_field, ok := processedMetric.GetField("dropped")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "dropped", err_msg)
	require.EqualValuesf(t, 1, dropped_field, "field %q value was not %q: %s", "dropped", 1, err_msg)

	total_field, ok := processedMetric.GetField("total")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "total", err_msg)
	require.EqualValuesf(t, 1, total_field, "field %q value was not %q; %s", "total", 1, err_msg)

	unknown_field, ok := processedMetric.GetField("unknown")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "unknown", err_msg)
	require.EqualValuesf(t, 1, unknown_field, "field %q value was not %q; %s", "unknown", 1, err_msg)

	processedMetric = all_metrics[1]

	measurement = processedMetric.Name()
	require.Equal(t, "datapoint", measurement, err_msg)

	tag_count = len(processedMetric.TagList())
	require.EqualValuesf(t, 2, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count = len(processedMetric.FieldList())
	require.EqualValuesf(t, 1, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	host_tag, ok := processedMetric.GetTag("host")
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", "host", err_msg)
	require.EqualValuesf(t, "pg123", host_tag, "tag %q value was not %q; %s", "host", "pg123", err_msg)

	result_tag, ok := processedMetric.GetTag(cl.ResultTag)
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.ResultTag, err_msg)
	require.EqualValuesf(t, "critical", result_tag, "tag %q value was not %q; %s", cl.ResultTag, "critical", err_msg)

	match_field, ok := processedMetric.GetField(cl.MatchField)
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.MatchField, err_msg)
	require.EqualValuesf(t, "CRITICAL:  second message from the same host", match_field, "field %q value was not %q; %s",
		cl.MatchField, "CRITICAL:  second message from the same host", err_msg)

	processedMetric = all_metrics[2]

	measurement = processedMetric.Name()
	require.Equal(t, "status", measurement, err_msg)

	tag_count = len(processedMetric.TagList())
	require.EqualValuesf(t, 1, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count = len(processedMetric.FieldList())
	require.EqualValuesf(t, 2, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	summary_tag, ok = processedMetric.GetTag("summary")
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", "summary", err_msg)
	require.EqualValuesf(t, "full", summary_tag, "tag %q value was not %q; %s", "summary", "full", err_msg)

	critical_field, ok := processedMetric.GetField("critical")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "critical", err_msg)
	require.EqualValuesf(t, 1, critical_field, "field %q value was not %q: %s", "critical", 1, err_msg)

	total_field, ok = processedMetric.GetField("total")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "total", err_msg)
	require.EqualValuesf(t, 1, total_field, "field %q value was not %q; %s", "total", 1, err_msg)
}

// Test essential operation of aggregating statistics by regex group.  Vary
// the set of output fields listed in the aggregation_group_fields option.
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
		MappedSelectorRegexes: map[string][]map[string]interface{}{
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
		AggregationPeriod:       "5s",
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationGroupTag:     "by_machine_type",
		AggregationGroupFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m0 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "WARNING:  situation is crazy",
	}, now)
	m1 := metric.New("datapoint", map[string]string{
		"host": "pg124",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)
	m2 := metric.New("datapoint", map[string]string{
		"host": "fire567",
	}, map[string]interface{}{
		"message": "INTRUSION:  assets at risk",
	}, now)
	metrics := make([]telegraf.Metric, 3)
	metrics[0] = m0
	metrics[1] = m1
	metrics[2] = m2

	// Our configured aggregation_period is one minute, so if this code
	// does not either wait for that interval to expire or force the
	// aggregation thread to shut down early and flush its data, we
	// will only get back the input data point, not the aggregation-data
	// metric as well.
	wait_duration, err := time.ParseDuration("10s")
	require.NoError(t, err)
	acc, err := RunClassifyTest(t, cl, metrics, wait_duration)
	require.NoError(t, err)

	// The original input data point should be dropped.
	// What we get back instead should be just the summary metric.
	//
	// For error reporting, if we have any, we dump out all the accumulator
	// items one by one on separate lines into a more descriptive error
	// message, not all in one run-on sentence that is hard to read.
	//
	all_metrics := acc.GetTelegrafMetrics()
	err_msg := "output metrics are:\n"
	for _, output_metric := range all_metrics {
		err_msg += fmt.Sprintf("%v\n", output_metric)
	}
	require.Equal(t, 4, len(all_metrics), err_msg)

	// At this point, we should have (except for different timestamp values, of course):
	// datapoint map[host:pg123 severity:warning] map[message:WARNING:  situation is crazy] 1655659203578691212
	// datapoint map[host:fire567 severity:critical] map[message:INTRUSION:  assets at risk] 1655659203578691212
	// status map[by_machine_type:database] map[dropped:1 total:2 unknown:1 warning:1] 1655659205001563076
	// status map[by_machine_type:firewall] map[critical:1 total:1] 1655659205001563076

	// The input-data items may appear in either order, so we have to deal with that in the logic here.
	distinct_selector := make(map[string]int)
	for index := 0; index <= 1; index++ {
		processedMetric := all_metrics[index]

		measurement := processedMetric.Name()
		require.Equal(t, "datapoint", measurement, err_msg)

		tag_count := len(processedMetric.TagList())
		require.EqualValuesf(t, 2, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)

		selector_tag, ok := processedMetric.GetTag(cl.SelectorTag)
		require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.SelectorTag, err_msg)

		result_tag, ok := processedMetric.GetTag(cl.ResultTag)
		require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.ResultTag, err_msg)

		field_count := len(processedMetric.FieldList())
		require.EqualValuesf(t, 1, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

		match_field, ok := processedMetric.GetField(cl.MatchField)
		require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.MatchField, err_msg)

		distinct_selector[selector_tag]++
		switch selector_tag {
		case "pg123":
			require.EqualValuesf(t, "warning", result_tag, "tag %q value was not %q; %s", cl.ResultTag, "warning", err_msg)
			require.EqualValuesf(t, "WARNING:  situation is crazy", match_field, "field %q value was not %q; %s",
				cl.MatchField, "WARNING:  situation is crazy", err_msg)
		case "fire567":
			require.EqualValuesf(t, "critical", result_tag, "tag %q value was not %q; %s", cl.ResultTag, "critical", err_msg)
			require.EqualValuesf(t, "INTRUSION:  assets at risk", match_field, "field %q value was not %q; %s",
				cl.MatchField, "INTRUSION:  assets at risk", err_msg)
		default:
			require.FailNowf(t, "an unexpected selector tag value appears", "%q tag value %q is unexpected; %s",
				cl.AggregationSelectorTag, selector_tag, err_msg)
		}
	}
	require.EqualValuesf(t, 2, len(distinct_selector), "have not seen the expected set of selectors in agggregation data points; %s", err_msg)

	// The aggregation-data items may appear in either order, so we have to deal with that in the logic here.
	distinct_group := make(map[string]int)
	for index := 2; index <= 3; index++ {
		processedMetric := all_metrics[index]

		measurement := processedMetric.Name()
		require.Equal(t, cl.AggregationMeasurement, measurement, err_msg)

		tag_count := len(processedMetric.TagList())
		require.EqualValuesf(t, 1, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)

		group_tag, ok := processedMetric.GetTag(cl.AggregationGroupTag)
		require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.AggregationGroupTag, err_msg)

		distinct_group[group_tag]++
		switch group_tag {
		case "database":
			field_count := len(processedMetric.FieldList())
			require.EqualValuesf(t, 4, field_count, "measurement %q selector %q has %d fields; %s", measurement, group_tag, field_count, err_msg)

			dropped_field, ok := processedMetric.GetField(cl.AggregationDroppedField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationDroppedField, err_msg)
			require.EqualValuesf(t, 1, dropped_field, "field %q value was not %q: %s", cl.AggregationDroppedField, 1, err_msg)

			total_field, ok := processedMetric.GetField(cl.AggregationTotalField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationTotalField, err_msg)
			require.EqualValuesf(t, 2, total_field, "field %q value was not %q; %s", cl.AggregationTotalField, 2, err_msg)

			unknown_field, ok := processedMetric.GetField("unknown")
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", "unknown", err_msg)
			require.EqualValuesf(t, 1, unknown_field, "field %q value was not %q; %s", "unknown", 1, err_msg)

			warning_field, ok := processedMetric.GetField("warning")
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", "warning", err_msg)
			require.EqualValuesf(t, 1, warning_field, "field %q value was not %q; %s", "warning", 1, err_msg)
		case "firewall":
			field_count := len(processedMetric.FieldList())
			require.EqualValuesf(t, 2, field_count, "measurement %q selector %q has %d fields; %s", measurement, group_tag, field_count, err_msg)

			critical_field, ok := processedMetric.GetField("critical")
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", "critical", err_msg)
			require.EqualValuesf(t, 1, critical_field, "field %q value was not %q; %s", "critical", 1, err_msg)

			total_field, ok := processedMetric.GetField(cl.AggregationTotalField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationTotalField, err_msg)
			require.EqualValuesf(t, 1, total_field, "field %q value was not %q; %s", cl.AggregationTotalField, 1, err_msg)
		default:
			require.FailNowf(t, "an unexpected group tag value appears", "%q tag value %q is unexpected; %s",
				cl.AggregationGroupTag, group_tag, err_msg)
		}
	}
	require.EqualValuesf(t, 2, len(distinct_group), "have not seen the expected set of groups in agggregation data points; %s", err_msg)
}

// Test essential operation of aggregating statistics by selector value.  Vary
// the set of output fields listed in the aggregation_selector_fields option.
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
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       "5s",
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSelectorTag:  "by_host",
		AggregationSelectorFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m0 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "WARNING:  situation is crazy",
	}, now)
	m1 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)
	m2 := metric.New("datapoint", map[string]string{
		"host": "pg124",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)
	metrics := make([]telegraf.Metric, 3)
	metrics[0] = m0
	metrics[1] = m1
	metrics[2] = m2

	// Our configured aggregation_period is one minute, so if this code
	// does not either wait for that interval to expire or force the
	// aggregation thread to shut down early and flush its data, we
	// will only get back the input data point, not the aggregation-data
	// metric as well.
	wait_duration, err := time.ParseDuration("10s")
	require.NoError(t, err)
	acc, err := RunClassifyTest(t, cl, metrics, wait_duration)
	require.NoError(t, err)

	// The original input data point should be dropped.
	// What we get back instead should be just the summary metric.
	//
	// For error reporting, if we have any, we dump out all the accumulator
	// items one by one on separate lines into a more descriptive error
	// message, not all in one run-on sentence that is hard to read.
	//
	all_metrics := acc.GetTelegrafMetrics()
	err_msg := "output metrics are:\n"
	for _, output_metric := range all_metrics {
		err_msg += fmt.Sprintf("%v\n", output_metric)
	}
	require.Equal(t, 3, len(all_metrics), err_msg)

	// At this point, we should have (except for different timestamp values, of course):
	// datapoint map[host:pg123 severity:warning] map[message:WARNING:  situation is crazy] 1655627120351552490
	// status map[by_host:pg123] map[dropped:1 total:2 unknown:1 warning:1] 1655627125003105632
	// status map[by_host:pg124] map[dropped:1 total:1 unknown:1] 1655627125003105632

	processedMetric := all_metrics[0]

	measurement := processedMetric.Name()
	require.Equal(t, "datapoint", measurement, err_msg)

	tag_count := len(processedMetric.TagList())
	require.EqualValuesf(t, 2, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count := len(processedMetric.FieldList())
	require.EqualValuesf(t, 1, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	selector_tag, ok := processedMetric.GetTag(cl.SelectorTag)
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.SelectorTag, err_msg)
	require.EqualValuesf(t, "pg123", selector_tag, "tag %q value was not %q; %s", cl.SelectorTag, "pg123", err_msg)

	result_tag, ok := processedMetric.GetTag(cl.ResultTag)
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.ResultTag, err_msg)
	require.EqualValuesf(t, "warning", result_tag, "tag %q value was not %q; %s", cl.ResultTag, "warning", err_msg)

	match_field, ok := processedMetric.GetField(cl.MatchField)
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.MatchField, err_msg)
	require.EqualValuesf(t, "WARNING:  situation is crazy", match_field, "field %q value was not %q; %s",
		cl.MatchField, "WARNING:  situation is crazy", err_msg)

	// The aggregation-data items may appear in either order, so we have to deal with that in the logic here.
	distinct_selector := make(map[string]int)
	for index := 1; index <= 2; index++ {
		processedMetric = all_metrics[index]

		measurement = processedMetric.Name()
		require.Equal(t, cl.AggregationMeasurement, measurement, err_msg)

		tag_count = len(processedMetric.TagList())
		require.EqualValuesf(t, 1, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)

		selector_tag, ok := processedMetric.GetTag(cl.AggregationSelectorTag)
		require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.AggregationSelectorTag, err_msg)

		distinct_selector[selector_tag]++
		switch selector_tag {
		case "pg123":
			field_count = len(processedMetric.FieldList())
			require.EqualValuesf(t, 4, field_count, "measurement %q selector %q has %d fields; %s", measurement, selector_tag, field_count, err_msg)

			dropped_field, ok := processedMetric.GetField(cl.AggregationDroppedField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationDroppedField, err_msg)
			require.EqualValuesf(t, 1, dropped_field, "field %q value was not %q: %s", cl.AggregationDroppedField, 1, err_msg)

			total_field, ok := processedMetric.GetField(cl.AggregationTotalField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationTotalField, err_msg)
			require.EqualValuesf(t, 2, total_field, "field %q value was not %q; %s", cl.AggregationTotalField, 2, err_msg)

			unknown_field, ok := processedMetric.GetField("unknown")
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", "unknown", err_msg)
			require.EqualValuesf(t, 1, unknown_field, "field %q value was not %q; %s", "unknown", 1, err_msg)

			warning_field, ok := processedMetric.GetField("warning")
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", "warning", err_msg)
			require.EqualValuesf(t, 1, warning_field, "field %q value was not %q; %s", "warning", 1, err_msg)
		case "pg124":
			field_count = len(processedMetric.FieldList())
			require.EqualValuesf(t, 3, field_count, "measurement %q selector %q has %d fields; %s", measurement, selector_tag, field_count, err_msg)

			dropped_field, ok := processedMetric.GetField(cl.AggregationDroppedField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationDroppedField, err_msg)
			require.EqualValuesf(t, 1, dropped_field, "field %q value was not %q: %s", cl.AggregationDroppedField, 1, err_msg)

			total_field, ok := processedMetric.GetField(cl.AggregationTotalField)
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationTotalField, err_msg)
			require.EqualValuesf(t, 1, total_field, "field %q value was not %q; %s", cl.AggregationTotalField, 1, err_msg)

			unknown_field, ok := processedMetric.GetField("unknown")
			require.Truef(t, ok, "could not find %q field in the returned metric; %s", "unknown", err_msg)
			require.EqualValuesf(t, 1, unknown_field, "field %q value was not %q; %s", "unknown", 1, err_msg)
		default:
			require.FailNowf(t, "an unexpected selector tag value appears", "%q tag value %q is unexpected; %s",
				cl.AggregationSelectorTag, selector_tag, err_msg)
		}
	}
	require.EqualValuesf(t, 2, len(distinct_selector), "have not seen the expected set of selectors in agggregation data points; %s", err_msg)
}

// Test the counting of dropped items and the total total number of processed
// items (whether dropped or not), along with the category names specified by
// the aggregation_dropped_field and aggregation_total_field options.
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
		MappedSelectorRegexes: map[string][]map[string]interface{}{
			"database": {
				{"ignore": "IGNORE"},
				{"okay": "OK"},
				{"warning": "WARNING"},
				{"critical": "CRITICAL"},
				{"unknown": ".*"},
			},
		},
		AggregationPeriod:       "5s",
		AggregationMeasurement:  "status",
		AggregationDroppedField: "dropped",
		AggregationTotalField:   "total",
		AggregationSummaryTag:   "summary",
		AggregationSummaryValue: "full",
		AggregationSummaryFields: []string{
			"ignore", "okay", "warning", "critical", "unknown", "dropped", "total",
		},
	}
	if testing.Verbose() {
		cl.logger = &testutil.Logger{}
	}

	now := time.Now()
	m0 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "WARNING:  situation is crazy",
	}, now)
	m1 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)
	m2 := metric.New("datapoint", map[string]string{
		"host": "pg123",
	}, map[string]interface{}{
		"message": "nothing to see here, move along",
	}, now)
	metrics := make([]telegraf.Metric, 3)
	metrics[0] = m0
	metrics[1] = m1
	metrics[2] = m2

	// Our configured aggregation_period is one minute, so if this code
	// does not either wait for that interval to expire or force the
	// aggregation thread to shut down early and flush its data, we
	// will only get back the input data point, not the aggregation-data
	// metric as well.
	wait_duration, err := time.ParseDuration("10s")
	require.NoError(t, err)
	acc, err := RunClassifyTest(t, cl, metrics, wait_duration)
	require.NoError(t, err)

	// The original input data point should be dropped.
	// What we get back instead should be just the summary metric.
	//
	// For error reporting, if we have any, we dump out all the accumulator
	// items one by one on separate lines into a more descriptive error
	// message, not all in one run-on sentence that is hard to read.
	//
	all_metrics := acc.GetTelegrafMetrics()
	err_msg := "output metrics are:\n"
	for _, output_metric := range all_metrics {
		err_msg += fmt.Sprintf("%v\n", output_metric)
	}
	require.Equal(t, 2, len(all_metrics), err_msg)

	// At this point, we should have (except for different timestamp values, of course):
	// datapoint map[host:pg123 severity:warning] map[message:WARNING:  situation is crazy] 1655624937598389523
	// status map[summary:full] map[dropped:2 total:3 unknown:2 warning:1] 1655624940002217562

	processedMetric := all_metrics[0]

	measurement := processedMetric.Name()
	require.Equal(t, "datapoint", measurement, err_msg)

	tag_count := len(processedMetric.TagList())
	require.EqualValuesf(t, 2, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count := len(processedMetric.FieldList())
	require.EqualValuesf(t, 1, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	selector_tag, ok := processedMetric.GetTag(cl.SelectorTag)
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.SelectorTag, err_msg)
	require.EqualValuesf(t, "pg123", selector_tag, "tag %q value was not %q; %s", cl.SelectorTag, "pg123", err_msg)

	result_tag, ok := processedMetric.GetTag(cl.ResultTag)
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.ResultTag, err_msg)
	require.EqualValuesf(t, "warning", result_tag, "tag %q value was not %q; %s", cl.ResultTag, "warning", err_msg)

	match_field, ok := processedMetric.GetField(cl.MatchField)
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.MatchField, err_msg)
	require.EqualValuesf(t, "WARNING:  situation is crazy", match_field, "field %q value was not %q; %s",
		cl.MatchField, "WARNING:  situation is crazy", err_msg)

	processedMetric = all_metrics[1]

	measurement = processedMetric.Name()
	require.Equal(t, cl.AggregationMeasurement, measurement, err_msg)

	tag_count = len(processedMetric.TagList())
	require.EqualValuesf(t, 1, tag_count, "measurement %q has %d tags; %s", measurement, tag_count, err_msg)
	field_count = len(processedMetric.FieldList())
	require.EqualValuesf(t, 4, field_count, "measurement %q has %d fields; %s", measurement, field_count, err_msg)

	summary_tag, ok := processedMetric.GetTag(cl.AggregationSummaryTag)
	require.Truef(t, ok, "could not find %q tag in the returned metric; %s", cl.AggregationSummaryTag, err_msg)
	require.EqualValuesf(t, cl.AggregationSummaryValue, summary_tag, "tag %q value was not %q; %s",
		cl.AggregationSummaryTag, cl.AggregationSummaryValue, err_msg)

	dropped_field, ok := processedMetric.GetField(cl.AggregationDroppedField)
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationDroppedField, err_msg)
	require.EqualValuesf(t, 2, dropped_field, "field %q value was not %q: %s", cl.AggregationDroppedField, 2, err_msg)

	total_field, ok := processedMetric.GetField(cl.AggregationTotalField)
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", cl.AggregationTotalField, err_msg)
	require.EqualValuesf(t, 3, total_field, "field %q value was not %q; %s", cl.AggregationTotalField, 3, err_msg)

	unknown_field, ok := processedMetric.GetField("unknown")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "unknown", err_msg)
	require.EqualValuesf(t, 2, unknown_field, "field %q value was not %q; %s", "unknown", 2, err_msg)

	warning_field, ok := processedMetric.GetField("warning")
	require.Truef(t, ok, "could not find %q field in the returned metric; %s", "warning", err_msg)
	require.EqualValuesf(t, 1, warning_field, "field %q value was not %q; %s", "warning", 1, err_msg)
}
