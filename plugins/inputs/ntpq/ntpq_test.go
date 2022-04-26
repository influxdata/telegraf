package ntpq

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestSingleNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(singleNTPQ),
		err: nil,
	}
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.Error(t, acc.GatherError(n.Gather))

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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.Error(t, acc.GatherError(n.Gather))

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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

	fields := map[string]interface{}{
		"when":   int64(7200),
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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.Error(t, acc.GatherError(n.Gather))

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

// TestParserNTPQ - realated to:
// https://github.com/influxdata/telegraf/issues/2386
func TestParserNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(multiParserNTPQ),
		err: nil,
	}

	n := newNTPQ()
	n.runQ = tt.runqTest
	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

	fields := map[string]interface{}{
		"poll":   int64(64),
		"when":   int64(60),
		"reach":  int64(377),
		"delay":  float64(0.0),
		"offset": float64(0.045),
		"jitter": float64(1.012),
	}
	tags := map[string]string{
		"remote":       "SHM(0)",
		"state_prefix": "*",
		"refid":        ".PPS.",
		"stratum":      "1",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)

	fields = map[string]interface{}{
		"poll":   int64(128),
		"when":   int64(121),
		"reach":  int64(377),
		"delay":  float64(0.0),
		"offset": float64(10.105),
		"jitter": float64(2.012),
	}
	tags = map[string]string{
		"remote":       "SHM(1)",
		"state_prefix": "-",
		"refid":        ".GPS.",
		"stratum":      "1",
		"type":         "u",
	}
	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)

	fields = map[string]interface{}{
		"poll":   int64(1024),
		"when":   int64(10),
		"reach":  int64(377),
		"delay":  float64(1.748),
		"offset": float64(0.373),
		"jitter": float64(0.101),
	}
	tags = map[string]string{
		"remote":       "37.58.57.238",
		"state_prefix": "+",
		"refid":        "192.53.103.103",
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
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

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
	tt := tester{
		ret: []byte(badHeaderNTPQ),
		err: nil,
	}
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

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
	tt := tester{
		ret: []byte(missingDelayNTPQ),
		err: nil,
	}
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

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

func TestLongPoll(t *testing.T) {
	tt := tester{
		ret: []byte(longPollTime),
		err: nil,
	}
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(n.Gather))

	fields := map[string]interface{}{
		"when":   int64(617),
		"poll":   int64(4080),
		"reach":  int64(377),
		"offset": float64(2.849),
		"jitter": float64(1.192),
		"delay":  float64(9.145),
	}
	tags := map[string]string{
		"remote":       "uschi5-ntp-002.",
		"state_prefix": "-",
		"refid":        "10.177.80.46",
		"type":         "u",
		"stratum":      "3",
	}

	acc.AssertContainsTaggedFields(t, "ntpq", fields, tags)
}

func TestFailedNTPQ(t *testing.T) {
	tt := tester{
		ret: []byte(singleNTPQ),
		err: fmt.Errorf("test failure"),
	}
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{}
	require.Error(t, acc.GatherError(n.Gather))
}

// It is possible for the output of ntqp to be missing the refid column.  This
// is believed to be http://bugs.ntp.org/show_bug.cgi?id=3484 which is fixed
// in ntp-4.2.8p12 (included first in Debian Buster).
func TestNoRefID(t *testing.T) {
	now := time.Now()
	expected := []telegraf.Metric{
		testutil.MustMetric("ntpq",
			map[string]string{
				"refid":   "10.177.80.37",
				"remote":  "83.137.98.96",
				"stratum": "2",
				"type":    "u",
			},
			map[string]interface{}{
				"delay":  float64(54.033),
				"jitter": float64(449514),
				"offset": float64(243.426),
				"poll":   int64(1024),
				"reach":  int64(377),
				"when":   int64(740),
			},
			now),
		testutil.MustMetric("ntpq",
			map[string]string{
				"refid":   "10.177.80.37",
				"remote":  "131.188.3.221",
				"stratum": "2",
				"type":    "u",
			},
			map[string]interface{}{
				"delay":  float64(111.820),
				"jitter": float64(449528),
				"offset": float64(261.921),
				"poll":   int64(1024),
				"reach":  int64(377),
				"when":   int64(783),
			},
			now),
	}

	tt := tester{
		ret: []byte(noRefID),
		err: nil,
	}
	n := newNTPQ()
	n.runQ = tt.runqTest

	acc := testutil.Accumulator{
		TimeFunc: func() time.Time { return now },
	}

	require.NoError(t, acc.GatherError(n.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

type tester struct {
	ret []byte
	err error
}

func (t *tester) runqTest() ([]byte, error) {
	return t.ret, t.err
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

var multiParserNTPQ = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
*SHM(0)          .PPS.                          1 u   60  64   377    0.000    0.045   1.012
+37.58.57.238 (d 192.53.103.103			2 u   10 1024  377    1.748    0.373   0.101
+37.58.57.238 (domain) 192.53.103.103   2 u   10 1024  377    1.748    0.373   0.101
+37.58.57.238 ( 192.53.103.103			2 u   10 1024  377    1.748    0.373   0.101
-SHM(1)          .GPS.                          1 u   121 128  377    0.000   10.105   2.012
`

var noRefID = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
 83.137.98.96    10.177.80.37     2 u  740 1024  377   54.033  243.426 449514.
 91.189.94.4                      2 u  673 1024  377  143.047  274.726 449445.
 131.188.3.221   10.177.80.37     2 u  783 1024  377  111.820  261.921 449528.
`

var longPollTime = `     remote           refid      st t when poll reach   delay   offset  jitter
==============================================================================
-uschi5-ntp-002. 10.177.80.46     3 u  617 68m 377 9.145 +2.849 1.192
`
