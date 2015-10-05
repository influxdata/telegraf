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
		"invalid.value:1.1|c",
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

	validations := []struct {
		name  string
		value int64
		cache map[string]cachedmetric
	}{
		{
			"invalid_sample_rate_counter",
			45,
			s.counters,
		},
		{
			"invalid_sample_rate_2_counter",
			45,
			s.counters,
		},
		{
			"invalid_sample_rate_gauge",
			45,
			s.gauges,
		},
		{
			"invalid_sample_rate_set",
			1,
			s.sets,
		},
	}

	for _, test := range validations {
		err := test_validate_value(test.name, test.value, test.cache)
		if err != nil {
			t.Error(err.Error())
		}
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
			"valid_counter",
			1,
		},
		{
			"valid_foo__bar_counter",
			11,
		},
	}

	for _, test := range validations {
		err := test_validate_value(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Test that name mappings match and work
func TestParse_NameMap(t *testing.T) {
	if false {
		t.Errorf("TODO")
	}
}

// Test that name map tags are applied properly
func TestParse_NameMapTags(t *testing.T) {
	if false {
		t.Errorf("TODO")
	}
}

// Valid lines should be parsed and their values should be cached
func TestParse_ValidLines(t *testing.T) {
	s := NewStatsd()
	valid_lines := []string{
		"valid:45|c",
		"valid:45|s",
		"valid:45|g",
		// TODO(cam): timings
		//"valid.timer:45|ms",
		//"valid.timer:45|h",
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
		cache map[string]cachedmetric
	}{
		{
			"valid_counter",
			45,
			s.counters,
		},
		{
			"valid_set",
			1,
			s.sets,
		},
		{
			"valid_gauge",
			45,
			s.gauges,
		},
	}

	for _, test := range validations {
		err := test_validate_value(test.name, test.value, test.cache)
		if err != nil {
			t.Error(err.Error())
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
		value int64
	}{
		{
			"plus_minus_gauge",
			120,
		},
		{
			"plus_plus_gauge",
			300,
		},
		{
			"minus_minus_gauge",
			-100,
		},
		{
			"lone_plus_gauge",
			100,
		},
		{
			"lone_minus_gauge",
			-100,
		},
		{
			"overwrite_gauge",
			300,
		},
	}

	for _, test := range validations {
		err := test_validate_value(test.name, test.value, s.gauges)
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
			"unique_user_ids_set",
			4,
		},
		{
			"oneuser_id_set",
			1,
		},
	}

	for _, test := range validations {
		err := test_validate_value(test.name, test.value, s.sets)
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
			"small_inc_counter",
			2,
		},
		{
			"big_inc_counter",
			1100101,
		},
		{
			"zero_init_counter",
			0,
		},
		{
			"sample_rate_counter",
			11,
		},
	}

	for _, test := range validations {
		err := test_validate_value(test.name, test.value, s.counters)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

// Tests low-level functionality of timings
func TestParse_Timings(t *testing.T) {
	if false {
		t.Errorf("TODO")
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

	err = test_validate_value("current_users_gauge", 100, s.gauges)
	if err != nil {
		t.Error(err.Error())
	}

	s.Gather(fakeacc)

	err = test_validate_value("current_users_gauge", 100, s.gauges)
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

	err = test_validate_value("unique_user_ids_set", 1, s.sets)
	if err != nil {
		t.Error(err.Error())
	}

	s.Gather(fakeacc)

	err = test_validate_value("unique_user_ids_set", 1, s.sets)
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

	err = test_validate_value("total_users_counter", 100, s.counters)
	if err != nil {
		t.Error(err.Error())
	}

	s.Gather(fakeacc)

	err = test_validate_value("total_users_counter", 100, s.counters)
	if err == nil {
		t.Error("total_users_counter metric should have been deleted")
	}
}

// Integration test the listener starting up and receiving UDP packets
func TestListen(t *testing.T) {
	if false {
		t.Errorf("TODO")
	}
}

// Test utility functions

func test_validate_value(name string, value int64, cache map[string]cachedmetric) error {
	metric, ok := cache[name]
	if !ok {
		return errors.New(fmt.Sprintf("Test Error: Metric name %s not found\n", name))
	}

	if value != metric.value {
		return errors.New(fmt.Sprintf("Measurement: %s, expected %d, actual %d",
			name, value, metric.value))
	}
	return nil
}
