package ubazaar

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestSerializeMetricFloat(t *testing.T) {
	// Setup test metric
	now := time.Now()
	tags := map[string]string{
		"customer_id":     "testCustomer",
		"service":         "MP",
		"unit_of_measure": "na-net-gb",
	}
	fields := map[string]interface{}{
		"quantity":   25.0,
		"start_time": now.Add(time.Second * -30).Format(time.RFC3339),
	}
	m, err := metric.New("net", tags, fields, now)
	assert.NoError(t, err)

	// Serialize Metric
	s, _ := NewSerializer()
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

	// Get UUID from serialized metric
	ubazaarEvent := &event{}
	err = json.Unmarshal(buf, ubazaarEvent)
	assert.NoError(t, err)

	expectedMetric := event{
		EventID:           ubazaarEvent.EventID,
		ServiceCustomerID: "testCustomer",
		Service:           "MP",
		UnitOfMeasure:     "na-net-gb",
		Quantity:          25.0,
		StartTime:         now.Add(time.Second * -30).Format(time.RFC3339),
		EndTime:           now.Format(time.RFC3339),
		MetaData:          make(map[string]string),
	}

	exp, err := json.Marshal(expectedMetric)
	assert.NoError(t, err)

	assert.Equal(t, string(exp)+"\n", string(buf))
}
