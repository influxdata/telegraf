package kafka_consumer_legacy

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

// Waits for the metric that was sent to the kafka broker to arrive at the kafka
// consumer
func waitForPoint(acc *testutil.Accumulator, t *testing.T) {
	// Give the kafka container up to 2 seconds to get the point to the consumer
	ticker := time.NewTicker(5 * time.Millisecond)
	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			if counter > 1000 {
				t.Fatal("Waited for 5s, point never arrived to consumer")
			} else if acc.NFields() == 1 {
				return
			}
		}
	}
}
