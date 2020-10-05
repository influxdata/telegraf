package devo

import (
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevoMapperWithDefaults(t *testing.T) {
	dw := newDevoWriter()
	dw.initializeDevoMapper()

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	hostname, err := os.Hostname()
	assert.NoError(t, err)
  bs, err := dw.Serialize(m1)
  bs, _ = dw.mapper.devoMapper(m1, bs)
  str := string(bs)
	assert.Equal(t, "<13>2010-11-10T23:00:00Z "+hostname+" my.app.telegraf.untagged: ", str, "Wrong syslog message")
}

func TestDevoMapperWithHostname(t *testing.T) {
	dw := newDevoWriter()
	dw.initializeDevoMapper()

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{
			"hostname": "testhost",
			"source":   "sourcevalue",
			"host":     "hostvalue",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
  bs, _ := dw.Serialize(m1)
  bs, _ = dw.mapper.devoMapper(m1, bs)
  str := string(bs)
	assert.Equal(t, "<13>2010-11-10T23:00:00Z testhost my.app.telegraf.untagged: ", str, "Wrong syslog message")
}
func TestDevoMapperWithHostnameSourceFallback(t *testing.T) {
	dw := newDevoWriter()
	dw.initializeDevoMapper()

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{
			"source": "sourcevalue",
			"host":   "hostvalue",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
  bs, _ := dw.Serialize(m1)
  bs, _ = dw.mapper.devoMapper(m1, bs)
  str := string(bs)
	assert.Equal(t, "<13>2010-11-10T23:00:00Z sourcevalue my.app.telegraf.untagged: ", str, "Wrong syslog message")
}

func TestDevoMapperWithHostnameHostFallback(t *testing.T) {
	dw := newDevoWriter()
	dw.initializeDevoMapper()

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{
			"host": "hostvalue",
		},
		map[string]interface{}{},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
  bs, _ := dw.Serialize(m1)
  bs, _ = dw.mapper.devoMapper(m1, bs)
  str := string(bs)
	assert.Equal(t, "<13>2010-11-10T23:00:00Z hostvalue my.app.telegraf.untagged: ", str, "Wrong syslog message")
}

func TestDevoMapperWithNoErrors(t *testing.T) {
	// Init mapper
	dw := newDevoWriter()
	dw.initializeDevoMapper()

	// Init metrics
	m1, _ := metric.New(
		"testmetric",
		map[string]string{
			"appname":            "testapp",
      "devo_tag":           "my.app.telegraf.devotagapp",
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

  bs, err := dw.Serialize(m1)
  bs, err = dw.mapper.devoMapper(m1, bs)
	require.NoError(t, err)
}
