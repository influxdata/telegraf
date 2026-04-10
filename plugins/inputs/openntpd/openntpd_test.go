package openntpd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func openntpdCTL(output string) func(string, config.Duration, bool) (*bytes.Buffer, error) {
	return func(string, config.Duration, bool) (*bytes.Buffer, error) {
		return bytes.NewBufferString(output), nil
	}
}

func TestParseSimpleOutput(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(simpleOutput),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "212.129.9.36",
				"stratum": "3",
			},
			map[string]interface{}{
				"wt":     int64(1),
				"tl":     int64(10),
				"next":   int64(56),
				"poll":   int64(63),
				"offset": float64(9.271),
				"delay":  float64(44.662),
				"jitter": float64(2.678),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseSimpleOutputwithStatePrefix(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(simpleOutputwithStatePrefix),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd",
			map[string]string{
				"remote":       "92.243.6.5",
				"stratum":      "2",
				"state_prefix": "*",
			},
			map[string]interface{}{
				"wt":     int64(1),
				"tl":     int64(10),
				"next":   int64(45),
				"poll":   int64(980),
				"offset": float64(-9.901),
				"delay":  float64(67.573),
				"jitter": float64(29.350),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseSimpleOutputInvalidPeer(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(simpleOutputInvalidPeer),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "178.33.111.49",
				"stratum": "-",
			},
			map[string]interface{}{
				"wt":   int64(1),
				"tl":   int64(2),
				"next": int64(203),
				"poll": int64(300),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseSimpleOutputServersDNSError(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(simpleOutputServersDNSError),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "pool.nl.ntp.org",
				"stratum": "-",
			},
			map[string]interface{}{
				"wt":   int64(1),
				"tl":   int64(2),
				"next": int64(2),
				"poll": int64(15),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseSimpleOutputServerDNSError(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(simpleOutputServerDNSError),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "pool.fr.ntp.org",
				"stratum": "-",
			},
			map[string]interface{}{
				"wt":   int64(1),
				"tl":   int64(2),
				"next": int64(12),
				"poll": int64(15),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseFullOutput(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(fullOutput),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd",
			map[string]string{"remote": "212.129.9.36", "stratum": "3"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(56), "poll": int64(63),
				"offset": float64(9.271), "delay": float64(44.662), "jitter": float64(2.678),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "163.172.25.19", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(21), "poll": int64(64),
				"offset": float64(-0.103), "delay": float64(53.199), "jitter": float64(9.046),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "92.243.6.5", "stratum": "2", "state_prefix": "*"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(45), "poll": int64(980),
				"offset": float64(-9.901), "delay": float64(67.573), "jitter": float64(29.350),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "178.33.111.49", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(203), "poll": int64(300),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "62.210.122.129", "stratum": "3"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(4), "poll": int64(60),
				"offset": float64(5.372), "delay": float64(53.690), "jitter": float64(14.700),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "163.172.225.159", "stratum": "3"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(38), "poll": int64(61),
				"offset": float64(12.276), "delay": float64(40.631), "jitter": float64(1.282),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "5.196.192.58", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(0), "poll": int64(300),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "129.250.35.250", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(28), "poll": int64(63),
				"offset": float64(11.236), "delay": float64(43.874), "jitter": float64(1.381),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "2001:41d0:a:5a7::1", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(5), "poll": int64(15),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "2001:41d0:8:188d::16", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(3), "poll": int64(15),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "2001:4b98:dc0:41:216:3eff:fe69:46e3", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(14), "poll": int64(15),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "2a01:e0d:1:3:58bf:fa61:0:1", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(9), "poll": int64(15),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "163.172.179.38", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(51), "poll": int64(65),
				"offset": float64(-19.229), "delay": float64(85.404), "jitter": float64(48.734),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "5.135.3.88", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(173), "poll": int64(300),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "195.154.41.195", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(84), "poll": int64(1004),
				"offset": float64(-3.956), "delay": float64(54.549), "jitter": float64(13.658),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "62.210.81.130", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(158), "poll": int64(1043),
				"offset": float64(-42.593), "delay": float64(124.353), "jitter": float64(94.230),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "149.202.97.123", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(205), "poll": int64(300),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "51.15.175.224", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(9), "poll": int64(64),
				"offset": float64(8.861), "delay": float64(46.640), "jitter": float64(0.668),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "37.187.5.167", "stratum": "-"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(2), "next": int64(105), "poll": int64(300),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{"remote": "194.57.169.1", "stratum": "2"},
			map[string]interface{}{
				"wt": int64(1), "tl": int64(10), "next": int64(32), "poll": int64(63),
				"offset": float64(6.589), "delay": float64(52.051), "jitter": float64(2.057),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseFullOutputAll(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(fullOutputAll),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd_status",
			map[string]string{},
			map[string]interface{}{
				"peers_valid":         int64(12),
				"peers_total":         int64(12),
				"sensors_valid":       int64(1),
				"sensors_total":       int64(1),
				"constraint_offset_s": int64(-1),
				"clock_synced":        int64(1),
				"stratum":             int64(1),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "212.129.9.36",
				"stratum": "3",
			},
			map[string]interface{}{
				"wt":     int64(1),
				"tl":     int64(10),
				"next":   int64(56),
				"poll":   int64(63),
				"offset": float64(9.271),
				"delay":  float64(44.662),
				"jitter": float64(2.678),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "163.172.25.19",
				"stratum": "2",
			},
			map[string]interface{}{
				"wt":     int64(1),
				"tl":     int64(10),
				"next":   int64(21),
				"poll":   int64(64),
				"offset": float64(-0.103),
				"delay":  float64(53.199),
				"jitter": float64(9.046),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{
				"remote":       "92.243.6.5",
				"stratum":      "2",
				"state_prefix": "*",
			},
			map[string]interface{}{
				"wt":     int64(1),
				"tl":     int64(10),
				"next":   int64(45),
				"poll":   int64(980),
				"offset": float64(-9.901),
				"delay":  float64(67.573),
				"jitter": float64(29.350),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "178.33.111.49",
				"stratum": "-",
			},
			map[string]interface{}{
				"wt":   int64(1),
				"tl":   int64(2),
				"next": int64(203),
				"poll": int64(300),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd_sensors",
			map[string]string{
				"sensor":       "nmea0",
				"refid":        "GPS",
				"state_prefix": "*",
			},
			map[string]interface{}{
				"wt":         int64(10),
				"gd":         int64(1),
				"st":         int64(0),
				"next":       int64(1),
				"poll":       int64(15),
				"offset":     float64(-0.673),
				"correction": float64(0.600),
			},
			time.Unix(0, 0),
		),
	}

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t,
		expected, actual,
		testutil.SortMetrics(),
		testutil.IgnoreTime(),
	)
}

func TestParseStatusLineClockNotSynced(t *testing.T) {
	plugin := &Openntpd{
		run: openntpdCTL(outputNoSync),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openntpd_status",
			map[string]string{},
			map[string]interface{}{
				"peers_valid":         int64(0),
				"peers_total":         int64(4),
				"constraint_offset_s": int64(0),
				"clock_synced":        int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openntpd",
			map[string]string{
				"remote":  "212.129.9.36",
				"stratum": "-",
			},
			map[string]interface{}{
				"wt":   int64(1),
				"tl":   int64(2),
				"next": int64(5),
				"poll": int64(15),
			},
			time.Unix(0, 0),
		),
	}

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t,
		expected, actual,
		testutil.SortMetrics(),
		testutil.IgnoreTime(),
	)
}

var simpleOutput = `peer
wt tl st  next  poll          offset       delay      jitter
212.129.9.36 from pool 0.debian.pool.ntp.org
1 10  3   56s   63s         9.271ms    44.662ms     2.678ms`

var simpleOutputwithStatePrefix = `peer
wt tl st  next  poll          offset       delay      jitter
92.243.6.5 from pool 0.debian.pool.ntp.org
*  1 10  2   45s  980s        -9.901ms    67.573ms    29.350ms`

var simpleOutputInvalidPeer = `peer
wt tl st  next  poll          offset       delay      jitter
178.33.111.49 from pool 0.debian.pool.ntp.org
1  2  -  203s  300s             ---- peer not valid ----`

var simpleOutputServersDNSError = `peer
wt tl st  next  poll          offset       delay      jitter
not resolved from pool pool.nl.ntp.org
1  2  -    2s   15s             ---- peer not valid ----
`
var simpleOutputServerDNSError = `peer
wt tl st  next  poll          offset       delay      jitter
not resolved pool.fr.ntp.org
1  2  -   12s   15s             ---- peer not valid ----
`

var fullOutput = `peer
wt tl st  next  poll          offset       delay      jitter
212.129.9.36 from pool 0.debian.pool.ntp.org
1 10  3   56s   63s         9.271ms    44.662ms     2.678ms
163.172.25.19 from pool 0.debian.pool.ntp.org
1 10  2   21s   64s        -0.103ms    53.199ms     9.046ms
92.243.6.5 from pool 0.debian.pool.ntp.org
*  1 10  2   45s  980s        -9.901ms    67.573ms    29.350ms
178.33.111.49 from pool 0.debian.pool.ntp.org
1  2  -  203s  300s             ---- peer not valid ----
62.210.122.129 from pool 1.debian.pool.ntp.org
1 10  3    4s   60s         5.372ms    53.690ms    14.700ms
163.172.225.159 from pool 1.debian.pool.ntp.org
1 10  3   38s   61s        12.276ms    40.631ms     1.282ms
5.196.192.58 from pool 1.debian.pool.ntp.org
1  2  -    0s  300s             ---- peer not valid ----
129.250.35.250 from pool 1.debian.pool.ntp.org
1 10  2   28s   63s        11.236ms    43.874ms     1.381ms
2001:41d0:a:5a7::1 from pool 2.debian.pool.ntp.org
1  2  -    5s   15s             ---- peer not valid ----
2001:41d0:8:188d::16 from pool 2.debian.pool.ntp.org
1  2  -    3s   15s             ---- peer not valid ----
2001:4b98:dc0:41:216:3eff:fe69:46e3 from pool 2.debian.pool.ntp.org
1  2  -   14s   15s             ---- peer not valid ----
2a01:e0d:1:3:58bf:fa61:0:1 from pool 2.debian.pool.ntp.org
1  2  -    9s   15s             ---- peer not valid ----
163.172.179.38 from pool 2.debian.pool.ntp.org
1 10  2   51s   65s       -19.229ms    85.404ms    48.734ms
5.135.3.88 from pool 2.debian.pool.ntp.org
1  2  -  173s  300s             ---- peer not valid ----
195.154.41.195 from pool 2.debian.pool.ntp.org
1 10  2   84s 1004s        -3.956ms    54.549ms    13.658ms
62.210.81.130 from pool 2.debian.pool.ntp.org
1 10  2  158s 1043s       -42.593ms   124.353ms    94.230ms
149.202.97.123 from pool 3.debian.pool.ntp.org
1  2  -  205s  300s             ---- peer not valid ----
51.15.175.224 from pool 3.debian.pool.ntp.org
1 10  2    9s   64s         8.861ms    46.640ms     0.668ms
37.187.5.167 from pool 3.debian.pool.ntp.org
1  2  -  105s  300s             ---- peer not valid ----
194.57.169.1 from pool 3.debian.pool.ntp.org
1 10  2   32s   63s         6.589ms    52.051ms     2.057ms`

// fullOutputAll represents the output of `ntpctl -s all` with status line,
// peers and a sensor section.
var fullOutputAll = `12/12 peers valid, 1/1 sensors valid, constraint offset -1s, clock synced, stratum 1

peer
   wt tl st  next  poll          offset       delay      jitter
212.129.9.36 from pool 0.debian.pool.ntp.org
    1 10  3   56s   63s         9.271ms    44.662ms     2.678ms
163.172.25.19 from pool 0.debian.pool.ntp.org
    1 10  2   21s   64s        -0.103ms    53.199ms     9.046ms
92.243.6.5 from pool 0.debian.pool.ntp.org
 *  1 10  2   45s  980s        -9.901ms    67.573ms    29.350ms
178.33.111.49 from pool 0.debian.pool.ntp.org
    1  2  -  203s  300s             ---- peer not valid ----

sensor
   wt gd st  next  poll          offset  correction
nmea0  GPS
 * 10  1  0    1s   15s        -0.673ms     0.600ms`

// outputNoSync represents a status line where the clock is not yet synced.
var outputNoSync = `0/4 peers valid, constraint offset 0s, clock unsynced

peer
   wt tl st  next  poll          offset       delay      jitter
212.129.9.36 from pool 0.debian.pool.ntp.org
    1  2  -    5s   15s             ---- peer not valid ----`
