package telegraf

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const validMs = `
cpu,cpu=cpu0,host=foo,datacenter=us-east usage_idle=99,usage_busy=1 1454105876344540456
`

const invalidMs = `
cpu, cpu=cpu0,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo usage_idle
cpu,host usage_idle=99
cpu,host=foo usage_idle=99 very bad metric
`

const validInvalidMs = `
cpu,cpu=cpu0,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu1,host=foo,datacenter=us-east usage_idle=51,usage_busy=49
cpu,cpu=cpu2,host=foo,datacenter=us-east usage_idle=60,usage_busy=40
cpu,host usage_idle=99
`

func TestParseValidMetrics(t *testing.T) {
	metrics, err := ParseMetrics([]byte(validMs))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	m := metrics[0]

	tags := map[string]string{
		"host":       "foo",
		"datacenter": "us-east",
		"cpu":        "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}

	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, int64(1454105876344540456), m.UnixNano())
}

func TestParseInvalidMetrics(t *testing.T) {
	metrics, err := ParseMetrics([]byte(invalidMs))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)
}

func TestParseValidAndInvalidMetrics(t *testing.T) {
	metrics, err := ParseMetrics([]byte(validInvalidMs))
	assert.Error(t, err)
	assert.Len(t, metrics, 3)
}

func TestNewMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, now, m.Time())
	assert.Equal(t, now.UnixNano(), m.UnixNano())
}

func TestNewMetricString(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
	}
	m, err := NewMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	lineProto := fmt.Sprintf("cpu,host=localhost usage_idle=99 %d",
		now.UnixNano())
	assert.Equal(t, lineProto, m.String())

	lineProtoPrecision := fmt.Sprintf("cpu,host=localhost usage_idle=99 %d",
		now.Unix())
	assert.Equal(t, lineProtoPrecision, m.PrecisionString("s"))
}

func TestNewMetricStringNoTime(t *testing.T) {
	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
	}
	m, err := NewMetric("cpu", tags, fields)
	assert.NoError(t, err)

	lineProto := fmt.Sprintf("cpu,host=localhost usage_idle=99")
	assert.Equal(t, lineProto, m.String())

	lineProtoPrecision := fmt.Sprintf("cpu,host=localhost usage_idle=99")
	assert.Equal(t, lineProtoPrecision, m.PrecisionString("s"))
}

func TestNewMetricFailNaN(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host": "localhost",
	}
	fields := map[string]interface{}{
		"usage_idle": math.NaN(),
	}

	_, err := NewMetric("cpu", tags, fields, now)
	assert.Error(t, err)
}
