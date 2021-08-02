package processors

import (
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestMetricType(t *testing.T) {
	mt := NewMetricType()

	now := time.Now()
	m := metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1", metricTypeTagName: "counter"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	err := mt.Process(m)

	assert.Nil(t, err)

	_, set := m.GetTag(metricTypeTagName)
	assert.Equal(t, false, set)
	assert.True(t, strings.HasSuffix(m.FieldList()[0].Key, ".counter"))
	assert.True(t, strings.HasSuffix(m.FieldList()[1].Key, ".counter"))
	assert.True(t, strings.HasSuffix(m.FieldList()[2].Key, ".counter"))
}
