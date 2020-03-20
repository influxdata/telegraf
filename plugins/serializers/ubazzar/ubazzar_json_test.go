package ubazzar

import (
	"fmt"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"event_id" : "test",
		"customer_id": "testCustomer",
		"unit_of_measure": "na-net-gb",
	}
	fields := map[string]interface{}{
		"quantity": 25.0,
		"start_time": now.Add(time.Second * -30).Format(time.RFC3339),
	}
	m, err := metric.New("net", tags, fields, now)
	assert.NoError(t, err)

	s, _ := NewSerializer(0)
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
	expS := []byte(fmt.Sprintf(`{"event_id":"test","service_customer_id":"testCustomer","service":"-","unit_of_measure":"na-net-gb","quantity":25,"start_time":"%s","end_time":"%s"}`, now.Add(time.Second * -30).Format(time.RFC3339), now.Format(time.RFC3339)) + "\n")
	assert.Equal(t, string(expS), string(buf))
}