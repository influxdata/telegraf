package syslog

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
)

func TestSyslogMapperWithDefaults(t *testing.T) {
	s := newSyslog()
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	hostname, err := os.Hostname()
	require.NoError(t, err)
	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<13>1 2010-11-10T23:00:00Z "+hostname+" Telegraf - testmetric -", str, "Wrong syslog message")
}

func TestSyslogMapperWithHostname(t *testing.T) {
	s := newSyslog()
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{
			"hostname": "testhost",
			"source":   "sourcevalue",
			"host":     "hostvalue",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<13>1 2010-11-10T23:00:00Z testhost Telegraf - testmetric -", str, "Wrong syslog message")
}
func TestSyslogMapperWithHostnameSourceFallback(t *testing.T) {
	s := newSyslog()
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{
			"source": "sourcevalue",
			"host":   "hostvalue",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<13>1 2010-11-10T23:00:00Z sourcevalue Telegraf - testmetric -", str, "Wrong syslog message")
}

func TestSyslogMapperWithHostnameHostFallback(t *testing.T) {
	s := newSyslog()
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{
			"host": "hostvalue",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<13>1 2010-11-10T23:00:00Z hostvalue Telegraf - testmetric -", str, "Wrong syslog message")
}

func TestSyslogMapperWithDefaultSdid(t *testing.T) {
	s := newSyslog()
	s.DefaultSdid = "default@32473"
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{
			"appname":            "testapp",
			"hostname":           "testhost",
			"tag1":               "bar",
			"default@32473_tag2": "foobar",
		},
		map[string]interface{}{
			"severity_code":        uint64(3),
			"facility_code":        uint64(3),
			"msg":                  "Test message",
			"procid":               uint64(25),
			"version":              uint16(2),
			"msgid":                int64(555),
			"timestamp":            time.Date(2010, time.November, 10, 23, 30, 0, 0, time.UTC).UnixNano(),
			"value1":               int64(2),
			"default@32473_value2": "foo",
			"value3":               float64(1.2),
		},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<27>2 2010-11-10T23:30:00Z testhost testapp 25 555 [default@32473 tag1=\"bar\" tag2=\"foobar\" value1=\"2\" value2=\"foo\" value3=\"1.2\"] Test message", str, "Wrong syslog message")
}

func TestSyslogMapperWithDefaultSdidAndOtherSdids(t *testing.T) {
	s := newSyslog()
	s.DefaultSdid = "default@32473"
	s.Sdids = []string{"bar@123", "foo@456"}
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{
			"appname":            "testapp",
			"hostname":           "testhost",
			"tag1":               "bar",
			"default@32473_tag2": "foobar",
			"bar@123_tag3":       "barfoobar",
		},
		map[string]interface{}{
			"severity_code":        uint64(1),
			"facility_code":        uint64(3),
			"msg":                  "Test message",
			"procid":               uint64(25),
			"version":              uint16(2),
			"msgid":                int64(555),
			"timestamp":            time.Date(2010, time.November, 10, 23, 30, 0, 0, time.UTC).UnixNano(),
			"value1":               int64(2),
			"default@32473_value2": "default",
			"bar@123_value3":       int64(2),
			"foo@456_value4":       "foo",
		},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<25>2 2010-11-10T23:30:00Z testhost testapp 25 555 [bar@123 tag3=\"barfoobar\" value3=\"2\"][default@32473 tag1=\"bar\" tag2=\"foobar\" value1=\"2\" value2=\"default\"][foo@456 value4=\"foo\"] Test message", str, "Wrong syslog message")
}

func TestSyslogMapperWithNoSdids(t *testing.T) {
	// Init mapper
	s := newSyslog()
	s.initializeSyslogMapper()

	// Init metrics
	m1 := metric.New(
		"testmetric",
		map[string]string{
			"appname":            "testapp",
			"hostname":           "testhost",
			"tag1":               "bar",
			"default@32473_tag2": "foobar",
			"bar@123_tag3":       "barfoobar",
			"foo@456_tag4":       "foobarfoo",
		},
		map[string]interface{}{
			"severity_code":        uint64(2),
			"facility_code":        uint64(3),
			"msg":                  "Test message",
			"procid":               uint64(25),
			"version":              uint16(2),
			"msgid":                int64(555),
			"timestamp":            time.Date(2010, time.November, 10, 23, 30, 0, 0, time.UTC).UnixNano(),
			"value1":               int64(2),
			"default@32473_value2": "default",
			"bar@123_value3":       int64(2),
			"foo@456_value4":       "foo",
		},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	syslogMessage, err := s.mapper.MapMetricToSyslogMessage(m1)
	require.NoError(t, err)
	str, _ := syslogMessage.String()
	require.Equal(t, "<26>2 2010-11-10T23:30:00Z testhost testapp 25 555 - Test message", str, "Wrong syslog message")
}
