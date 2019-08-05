package timestamp

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestTimestampTag(t *testing.T) {
	sut := Timestamp{
		FieldKey: "timestamp",
	}

	currentTime := time.Now()

	met, _ := metric.New("foo", nil, nil, currentTime)

	sut.Apply(met)

	timestamp, _ := met.GetField("timestamp")

	assert.Equal(t, timestamp, currentTime.UnixNano())
}
