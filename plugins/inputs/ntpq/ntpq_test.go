// +build !windows

package ntpq

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSingleNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(singleNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(101),
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestBadIntNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(badIntParseNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(101),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestBadFloatNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(badFloatParseNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(2),
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestDaysNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(whenDaysNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(172800),
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestHoursNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(whenHoursNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(720),
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestMinutesNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(whenMinutesNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(120),
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestBadWhenNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(whenBadNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"stratum":      "2",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestMultiNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(multiNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"delay":  float64(54.033),
		"jitter": float64(449514),
		"offset": float64(243.426),
		"poll":   int64(1024),
		"reach":  int64(377),
		"when":   int64(740),
	}
	tags := map[string]string{
		"refid":   "10.177.80.37",
		"remote":  "83.137.98.96",
		"stratum": "2",
		"type":    "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)

	fields = map[string]interface{}{
		"delay":  float64(60.785),
		"jitter": float64(449539),
		"offset": float64(232.597),
		"poll":   int64(1024),
		"reach":  int64(377),
		"when":   int64(739),
	}
	tags = map[string]string{
		"refid":   "10.177.80.37",
		"remote":  "81.7.16.52",
		"stratum": "2",
		"type":    "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestBadHeaderNTPQ(t *testing.T) {
	resetVars()
	tt := tester{
		ret: []byte(badHeaderNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(101),
		"poll":   int64(256),
		"reach":  int64(37),
		"delay":  float64(51.016),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestMissingDelayColumnNTPQ(t *testing.T) {
	resetVars()
	tt := tester{
		ret: []byte(missingDelayNTPQ),
		err: nil,
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, n.Gather(&acc))

	fields := map[string]interface{}{
		"when":   int64(101),
		"poll":   int64(256),
		"reach":  int64(37),
		"offset": float64(233.010),
		"jitter": float64(17.462),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "*",
		"refid":        "10.177.80.46",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestFailedNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(singleNTPQ),
		err: fmt.Errorf("Test failure"),
	}
	n := &NTPQ{
		runQ: tt.runqTest,
	}

	acc := testutil.Accumulator{}
	assert.Error(t, n.Gather(&acc))
}

type tester struct {
	ret []byte
	err error
}

func (t *tester) runqTest() ([]byte, error) {
	return t.ret, t.err
}

func resetVars() {
	// Mapping of ntpq header names to tag keys
	tagHeaders = map[string]string{
		"remote": "remote",
		"refid":  "refid",
		"st":     "stratum",
		"t":      "type",
	}

	// Mapping of the ntpq tag key to the index in the command output
	tagI = map[string]int{
		"remote":  -1,
		"refid":   -1,
		"stratum": -1,
		"type":    -1,
	}

	// Mapping of float metrics to their index in the command output
	floatI = map[string]int{
		"delay":  -1,
		"offset": -1,
		"jitter": -1,
	}

	// Mapping of int metrics to their index in the command output
	intI = map[string]int{
		"when":  -1,
		"poll":  -1,
		"reach": -1,
	}
}

var singleNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  101  256   37   51.016  233.010  17.462
`

var badHeaderNTPQ = `remote      refid   foobar t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  101  256   37   51.016  233.010  17.462
`

var missingDelayNTPQ = `remote      refid   foobar t when poll reach   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  101  256   37   233.010  17.462
`

var whenDaysNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  2d  256   37   51.016  233.010  17.462
`

var whenHoursNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  2h  256   37   51.016  233.010  17.462
`

var whenMinutesNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  2m  256   37   51.016  233.010  17.462
`

var whenBadNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  2q  256   37   51.016  233.010  17.462
`

var badFloatParseNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  2  256   37   51.016  foobar  17.462
`

var badIntParseNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*uschi5-ntp-002. 10.177.80.46     2 u  101  foobar   37   51.016  233.010  17.462
`

var multiNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
 83.137.98.96    10.177.80.37     2 u  740 1024  377   54.033  243.426 449514.
 81.7.16.52      10.177.80.37     2 u  739 1024  377   60.785  232.597 449539.
 131.188.3.221   10.177.80.37     2 u  783 1024  377  111.820  261.921 449528.
 5.9.29.107      10.177.80.37     2 u  703 1024  377  205.704  160.406 449602.
 91.189.94.4     10.177.80.37     2 u  673 1024  377  143.047  274.726 449445.
`
