package statsd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/influxdb/telegraf/testutil"
)

// Invalid lines should return an error
func TestParse_InvalidLines(t *testing.T) {
	s := NewStatsd()
	invalid_lines := []string{
		"i.dont.have.a.pipe:45g",
		"i.dont.have.a.colon45|c",
		"invalid.metric.type:45|e",
		"invalid.plus.minus.non.gauge:+10|c",
		"invalid.plus.minus.non.gauge:+10|s",
		"invalid.plus.minus.non.gauge:+10|ms",
		"invalid.plus.minus.non.gauge:+10|h",
		"invalid.plus.minus.non.gauge:-10|c",
		"invalid.value:foobar|c",
		"invalid.value:d11|c",
		"invalid.value:1d1|c",
	}
	for _, line := range invalid_lines {
		err := s.parseStatsdLine(line)
		if err == nil {
			t.Errorf("Parsing line %s should have resulted in an error\n", line)
		}
	}
}

// Invalid sample rates should be ignored and not applied
func TestParse_InvalidSampleRate(t *testing.T) {
	s := NewStatsd()
	invalid_lines := []string{
		"invalid.sample.rate:45|c|0.1",
		"invalid.sample.rate.2:45|c|@foo",
		"invalid.sample.rate:45|g|@0.1",
		"invalid.sample.rate:45|s|@0.1",
	}

	for _, line := range invalid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	counter_validations := []struct {
		name  string
		value int64
		cache map[string]cachedcounter
	}{
		{
			"invalid_sample_rate",
			45,
			s.counters,
		},
		{
			"invalid_sample_rate_2",
			45,
			s.counters,
		},
	}

	for _, test := range counter_validations {
		err := test_validate_counter(test.name, test.value, test.cache)
		if err != nil {
			t.Error(err.Error())
		}
	}

	err := test_validate_gauge("invalid_sample_rate", 45, s.gauges)
	if err != nil {
		t.Error(err.Error())
	}

	err = test_validate_set("invalid_sample_rate", 1, s.sets)
	if err != nil {
		t.Error(err.Error())
	}
}

// Names should be parsed like . -> _ and - -> __
func TestParse_DefaultNameParsing(t *testing.T) {
	s := NewStatsd()
	valid_lines := []string{
		"valid:1|c",
		"valid.foo-bar:11|c",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"valid",
			1,
		},
		{
			"valid_foo__bar",
			11,
		},
	}

	for _, test := range validations {
		err := test_validate_counter(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Test that template name transformation works
func TestParse_Template(t *testing.T) {
	s := NewStatsd()
	s.Templates = []string{
		"measurement.measurement.host.service",
	}

	lines := []string{
		"cpu.idle.localhost:1|c",
		"cpu.busy.host01.myservice:11|c",
	}

	for _, line := range lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"cpu_idle",
			1,
		},
		{
			"cpu_busy",
			11,
		},
	}

	// Validate counters
	for _, test := range validations {
		err := test_validate_counter(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Test that template filters properly
func TestParse_TemplateFilter(t *testing.T) {
	s := NewStatsd()
	s.Templates = []string{
		"cpu.idle.* measurement.measurement.host",
	}

	lines := []string{
		"cpu.idle.localhost:1|c",
		"cpu.busy.host01.myservice:11|c",
	}

	for _, line := range lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"cpu_idle",
			1,
		},
		{
			"cpu_busy_host01_myservice",
			11,
		},
	}

	// Validate counters
	for _, test := range validations {
		err := test_validate_counter(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Test that most specific template is chosen
func TestParse_TemplateSpecificity(t *testing.T) {
	s := NewStatsd()
	s.Templates = []string{
		"cpu.* measurement.foo.host",
		"cpu.idle.* measurement.measurement.host",
	}

	lines := []string{
		"cpu.idle.localhost:1|c",
	}

	for _, line := range lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"cpu_idle",
			1,
		},
	}

	// Validate counters
	for _, test := range validations {
		err := test_validate_counter(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Test that fields are parsed correctly
func TestParse_Fields(t *testing.T) {
	if false {
		t.Errorf("TODO")
	}
}

// Test that tags within the bucket are parsed correctly
func TestParse_Tags(t *testing.T) {
	s := NewStatsd()

	tests := []struct {
		bucket string
		name   string
		tags   map[string]string
	}{
		{
			"cpu.idle,host=localhost",
			"cpu_idle",
			map[string]string{
				"host": "localhost",
			},
		},
		{
			"cpu.idle,host=localhost,region=west",
			"cpu_idle",
			map[string]string{
				"host":   "localhost",
				"region": "west",
			},
		},
		{
			"cpu.idle,host=localhost,color=red,region=west",
			"cpu_idle",
			map[string]string{
				"host":   "localhost",
				"region": "west",
				"color":  "red",
			},
		},
	}

	for _, test := range tests {
		name, tags := s.parseName(test.bucket)
		if name != test.name {
			t.Errorf("Expected: %s, got %s", test.name, name)
		}

		for k, v := range test.tags {
			actual, ok := tags[k]
			if !ok {
				t.Errorf("Expected key: %s not found", k)
			}
			if actual != v {
				t.Errorf("Expected %s, got %s", v, actual)
			}
		}
	}
}

// Test that measurements with the same name, but different tags, are treated
// as different outputs
func TestParse_MeasurementsWithSameName(t *testing.T) {
	s := NewStatsd()

	// Test that counters work
	valid_lines := []string{
		"test.counter,host=localhost:1|c",
		"test.counter,host=localhost,region=west:1|c",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	if len(s.counters) != 2 {
		t.Errorf("Expected 2 separate measurements, found %d", len(s.counters))
	}
}

// Test that measurements with multiple bits, are treated as different outputs
// but are equal to their single-measurement representation
func TestParse_MeasurementsWithMultipleValues(t *testing.T) {
	single_lines := []string{
		"valid.multiple:0|ms|@0.1",
		"valid.multiple:0|ms|",
		"valid.multiple:1|ms",
		"valid.multiple.duplicate:1|c",
		"valid.multiple.duplicate:1|c",
		"valid.multiple.duplicate:2|c",
		"valid.multiple.duplicate:1|c",
		"valid.multiple.duplicate:1|h",
		"valid.multiple.duplicate:1|h",
		"valid.multiple.duplicate:2|h",
		"valid.multiple.duplicate:1|h",
		"valid.multiple.duplicate:1|s",
		"valid.multiple.duplicate:1|s",
		"valid.multiple.duplicate:2|s",
		"valid.multiple.duplicate:1|s",
		"valid.multiple.duplicate:1|g",
		"valid.multiple.duplicate:1|g",
		"valid.multiple.duplicate:2|g",
		"valid.multiple.duplicate:1|g",
		"valid.multiple.mixed:1|c",
		"valid.multiple.mixed:1|ms",
		"valid.multiple.mixed:2|s",
		"valid.multiple.mixed:1|g",
	}

	multiple_lines := []string{
		"valid.multiple:0|ms|@0.1:0|ms|:1|ms",
		"valid.multiple.duplicate:1|c:1|c:2|c:1|c",
		"valid.multiple.duplicate:1|h:1|h:2|h:1|h",
		"valid.multiple.duplicate:1|s:1|s:2|s:1|s",
		"valid.multiple.duplicate:1|g:1|g:2|g:1|g",
		"valid.multiple.mixed:1|c:1|ms:2|s:1|g",
	}

	s_single := NewStatsd()
	s_multiple := NewStatsd()

	for _, line := range single_lines {
		err := s_single.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	for _, line := range multiple_lines {
		err := s_multiple.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	if len(s_single.timings) != 3 {
		t.Errorf("Expected 3 measurement, found %d", len(s_single.timings))
	}

	if cachedtiming, ok := s_single.timings["metric_type=timingvalid_multiple"]; !ok {
		t.Errorf("Expected cached measurement with hash 'metric_type=timingvalid_multiple' not found")
	} else {
		if cachedtiming.name != "valid_multiple" {
			t.Errorf("Expected the name to be 'valid_multiple', got %s", cachedtiming.name)
		}

		// A 0 at samplerate 0.1 will add 10 values of 0,
		// A 0 with invalid samplerate will add a single 0,
		// plus the last bit of value 1
		// which adds up to 12 individual datapoints to be cached
		if cachedtiming.stats.n != 12 {
			t.Errorf("Expected 11 additions, got %d", cachedtiming.stats.n)
		}

		if cachedtiming.stats.upper != 1 {
			t.Errorf("Expected max input to be 1, got %f", cachedtiming.stats.upper)
		}
	}

	// test if s_single and s_multiple did compute the same stats for valid.multiple.duplicate
	if err := test_validate_set("valid_multiple_duplicate", 2, s_single.sets); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_set("valid_multiple_duplicate", 2, s_multiple.sets); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_counter("valid_multiple_duplicate", 5, s_single.counters); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_counter("valid_multiple_duplicate", 5, s_multiple.counters); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_gauge("valid_multiple_duplicate", 1, s_single.gauges); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_gauge("valid_multiple_duplicate", 1, s_multiple.gauges); err != nil {
		t.Error(err.Error())
	}

	// test if s_single and s_multiple did compute the same stats for valid.multiple.mixed
	if err := test_validate_set("valid_multiple_mixed", 1, s_single.sets); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_set("valid_multiple_mixed", 1, s_multiple.sets); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_counter("valid_multiple_mixed", 1, s_single.counters); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_counter("valid_multiple_mixed", 1, s_multiple.counters); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_gauge("valid_multiple_mixed", 1, s_single.gauges); err != nil {
		t.Error(err.Error())
	}

	if err := test_validate_gauge("valid_multiple_mixed", 1, s_multiple.gauges); err != nil {
		t.Error(err.Error())
	}
}

// Valid lines should be parsed and their values should be cached
func TestParse_ValidLines(t *testing.T) {
	s := NewStatsd()
	valid_lines := []string{
		"valid:45|c",
		"valid:45|s",
		"valid:45|g",
		"valid.timer:45|ms",
		"valid.timer:45|h",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}
}

// Tests low-level functionality of gauges
func TestParse_Gauges(t *testing.T) {
	s := NewStatsd()

	// Test that gauge +- values work
	valid_lines := []string{
		"plus.minus:100|g",
		"plus.minus:-10|g",
		"plus.minus:+30|g",
		"plus.plus:100|g",
		"plus.plus:+100|g",
		"plus.plus:+100|g",
		"minus.minus:100|g",
		"minus.minus:-100|g",
		"minus.minus:-100|g",
		"lone.plus:+100|g",
		"lone.minus:-100|g",
		"overwrite:100|g",
		"overwrite:300|g",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value float64
	}{
		{
			"plus_minus",
			120,
		},
		{
			"plus_plus",
			300,
		},
		{
			"minus_minus",
			-100,
		},
		{
			"lone_plus",
			100,
		},
		{
			"lone_minus",
			-100,
		},
		{
			"overwrite",
			300,
		},
	}

	for _, test := range validations {
		err := test_validate_gauge(test.name, test.value, s.gauges)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Tests low-level functionality of sets
func TestParse_Sets(t *testing.T) {
	s := NewStatsd()

	// Test that sets work
	valid_lines := []string{
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:100|s",
		"unique.user.ids:101|s",
		"unique.user.ids:102|s",
		"unique.user.ids:102|s",
		"unique.user.ids:123456789|s",
		"oneuser.id:100|s",
		"oneuser.id:100|s",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"unique_user_ids",
			4,
		},
		{
			"oneuser_id",
			1,
		},
	}

	for _, test := range validations {
		err := test_validate_set(test.name, test.value, s.sets)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Tests low-level functionality of counters
func TestParse_Counters(t *testing.T) {
	s := NewStatsd()

	// Test that counters work
	valid_lines := []string{
		"small.inc:1|c",
		"big.inc:100|c",
		"big.inc:1|c",
		"big.inc:100000|c",
		"big.inc:1000000|c",
		"small.inc:1|c",
		"zero.init:0|c",
		"sample.rate:1|c|@0.1",
		"sample.rate:1|c",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	validations := []struct {
		name  string
		value int64
	}{
		{
			"small_inc",
			2,
		},
		{
			"big_inc",
			1100101,
		},
		{
			"zero_init",
			0,
		},
		{
			"sample_rate",
			11,
		},
	}

	for _, test := range validations {
		err := test_validate_counter(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Tests low-level functionality of timings
func TestParse_Timings(t *testing.T) {
	s := NewStatsd()
	s.Percentiles = []int{90}
	acc := &testutil.Accumulator{}

	// Test that counters work
	valid_lines := []string{
		"test.timing:1|ms",
		"test.timing:1|ms",
		"test.timing:1|ms",
		"test.timing:1|ms",
		"test.timing:1|ms",
	}

	for _, line := range valid_lines {
		err := s.parseStatsdLine(line)
		if err != nil {
			t.Errorf("Parsing line %s should not have resulted in an error\n", line)
		}
	}

	s.Gather(acc)

	tests := []struct {
		name  string
		value interface{}
	}{
		{
			"test_timing_mean",
			float64(1),
		},
		{
			"test_timing_stddev",
			float64(0),
		},
		{
			"test_timing_upper",
			float64(1),
		},
		{
			"test_timing_lower",
			float64(1),
		},
		{
			"test_timing_count",
			int64(5),
		},
		{
			"test_timing_percentile_90",
			float64(1),
		},
	}

	for _, test := range tests {
		acc.AssertContainsFields(t, test.name,
			map[string]interface{}{"value": test.value})
	}
}

func TestParse_Timings_Delete(t *testing.T) {
	s := NewStatsd()
	s.DeleteTimings = true
	fakeacc := &testutil.Accumulator{}
	var err error

	line := "timing:100|ms"
	err = s.parseStatsdLine(line)
	if err != nil {
		t.Errorf("Parsing line %s should not have resulted in an error\n", line)
	}

	if len(s.timings) != 1 {
		t.Errorf("Should be 1 timing, found %d", len(s.timings))
	}

	s.Gather(fakeacc)

	if len(s.timings) != 0 {
		t.Errorf("All timings should have been deleted, found %d", len(s.timings))
	}
}

// Tests the delete_gauges option
func TestParse_Gauges_Delete(t *testing.T) {
	s := NewStatsd()
	s.DeleteGauges = true
	fakeacc := &testutil.Accumulator{}
	var err error

	line := "current.users:100|g"
	err = s.parseStatsdLine(line)
	if err != nil {
		t.Errorf("Parsing line %s should not have resulted in an error\n", line)
	}

	err = test_validate_gauge("current_users", 100, s.gauges)
	if err != nil {
		t.Error(err.Error())
	}

	s.Gather(fakeacc)

	err = test_validate_gauge("current_users", 100, s.gauges)
	if err == nil {
		t.Error("current_users_gauge metric should have been deleted")
	}
}

// Tests the delete_sets option
func TestParse_Sets_Delete(t *testing.T) {
	s := NewStatsd()
	s.DeleteSets = true
	fakeacc := &testutil.Accumulator{}
	var err error

	line := "unique.user.ids:100|s"
	err = s.parseStatsdLine(line)
	if err != nil {
		t.Errorf("Parsing line %s should not have resulted in an error\n", line)
	}

	err = test_validate_set("unique_user_ids", 1, s.sets)
	if err != nil {
		t.Error(err.Error())
	}

	s.Gather(fakeacc)

	err = test_validate_set("unique_user_ids", 1, s.sets)
	if err == nil {
		t.Error("unique_user_ids_set metric should have been deleted")
	}
}

// Tests the delete_counters option
func TestParse_Counters_Delete(t *testing.T) {
	s := NewStatsd()
	s.DeleteCounters = true
	fakeacc := &testutil.Accumulator{}
	var err error

	line := "total.users:100|c"
	err = s.parseStatsdLine(line)
	if err != nil {
		t.Errorf("Parsing line %s should not have resulted in an error\n", line)
	}

	err = test_validate_counter("total_users", 100, s.counters)
	if err != nil {
		t.Error(err.Error())
	}

	s.Gather(fakeacc)

	err = test_validate_counter("total_users", 100, s.counters)
	if err == nil {
		t.Error("total_users_counter metric should have been deleted")
	}
}

func TestParseKeyValue(t *testing.T) {
	k, v := parseKeyValue("foo=bar")
	if k != "foo" {
		t.Errorf("Expected %s, got %s", "foo", k)
	}
	if v != "bar" {
		t.Errorf("Expected %s, got %s", "bar", v)
	}

	k2, v2 := parseKeyValue("baz")
	if k2 != "" {
		t.Errorf("Expected %s, got %s", "", k2)
	}
	if v2 != "baz" {
		t.Errorf("Expected %s, got %s", "baz", v2)
	}
}

// Test utility functions

func test_validate_set(
	name string,
	value int64,
	cache map[string]cachedset,
) error {
	var metric cachedset
	var found bool
	for _, v := range cache {
		if v.name == name {
			metric = v
			found = true
			break
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Test Error: Metric name %s not found\n", name))
	}

	if value != int64(len(metric.set)) {
		return errors.New(fmt.Sprintf("Measurement: %s, expected %d, actual %d\n",
			name, value, len(metric.set)))
	}
	return nil
}

func test_validate_counter(
	name string,
	value int64,
	cache map[string]cachedcounter,
) error {
	var metric cachedcounter
	var found bool
	for _, v := range cache {
		if v.name == name {
			metric = v
			found = true
			break
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Test Error: Metric name %s not found\n", name))
	}

	if value != metric.value {
		return errors.New(fmt.Sprintf("Measurement: %s, expected %d, actual %d\n",
			name, value, metric.value))
	}
	return nil
}

func test_validate_gauge(
	name string,
	value float64,
	cache map[string]cachedgauge,
) error {
	var metric cachedgauge
	var found bool
	for _, v := range cache {
		if v.name == name {
			metric = v
			found = true
			break
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Test Error: Metric name %s not found\n", name))
	}

	if value != metric.value {
		return errors.New(fmt.Sprintf("Measurement: %s, expected %f, actual %f\n",
			name, value, metric.value))
	}
	return nil
}
