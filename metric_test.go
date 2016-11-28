package telegraf

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

	assert.Equal(t, Untyped, m.Type())
	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, now, m.Time())
	assert.Equal(t, now.UnixNano(), m.UnixNano())
}

func TestNewGaugeMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := NewGaugeMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, Gauge, m.Type())
	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, now, m.Time())
	assert.Equal(t, now.UnixNano(), m.UnixNano())
}

func TestNewCounterMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := NewCounterMetric("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, Counter, m.Type())
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
