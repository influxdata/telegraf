package sum

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func MustMetric(name string, tags map[string]string, fields map[string]interface{}, metricTime time.Time) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m, _ := metric.New(name, tags, fields, metricTime)
	return m
}

func TestSum(t *testing.T) {
	s := Sum{
		FieldKey: "total_net_usage",
		FieldSum: []string{"bytes_sent", "bytes_recv"},
	}

	err := s.Init()
	require.NoError(t, err)

	currentTime := time.Now()

	m  := MustMetric("", nil, map[string]interface{}{"bytes_sent": 10, "bytes_recv": 10}, currentTime)
	m2 := s.Apply(m)

	sum, _ := m2[0].GetField("total_net_usage")
	assert.Equal(t, int64(20), sum)


}