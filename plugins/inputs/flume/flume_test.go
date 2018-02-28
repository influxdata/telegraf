package flume

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"strings"
)

const flumeSampleResponse = `
{
    "CHANNEL.c1": {
        "ChannelCapacity": "100000",
        "ChannelFillPercentage": "0.44999999999999996",
        "ChannelSize": "450",
        "EventPutAttemptCount": "16760030",
        "EventPutSuccessCount": "16760030",
        "EventTakeAttemptCount": "16760032",
        "EventTakeSuccessCount": "16759000",
        "StartTime": "1519632359600",
        "StopTime": "0",
        "Type": "CHANNEL"
    },
    "SINK.k1": {
        "BatchCompleteCount": "8370",
        "BatchEmptyCount": "0",
        "BatchUnderflowCount": "0",
        "ConnectionClosedCount": "0",
        "ConnectionCreatedCount": "1",
        "ConnectionFailedCount": "0",
        "EventDrainAttemptCount": "8370000",
        "EventDrainSuccessCount": "8370000",
        "StartTime": "1519632360106",
        "StopTime": "0",
        "Type": "SINK"
    },
    "SINK.k2": {
        "BatchCompleteCount": "8389",
        "BatchEmptyCount": "0",
        "BatchUnderflowCount": "0",
        "ConnectionClosedCount": "0",
        "ConnectionCreatedCount": "1",
        "ConnectionFailedCount": "0",
        "EventDrainAttemptCount": "8389000",
        "EventDrainSuccessCount": "8389000",
        "StartTime": "1519632360110",
        "StopTime": "0",
        "Type": "SINK"
    },
    "SOURCE.r1": {
        "AppendAcceptedCount": "0",
        "AppendBatchAcceptedCount": "0",
        "AppendBatchReceivedCount": "0",
        "AppendReceivedCount": "0",
        "EventAcceptedCount": "16760030",
        "EventReceivedCount": "16760030",
        "KafkaCommitTimer": "3157997",
        "KafkaEventGetTimer": "171215457",
        "OpenConnectionCount": "0",
        "StartTime": "1519632388215",
        "StopTime": "0",
        "Type": "SOURCE"
    }
}
`

func TestNginxGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/metrics" {
			rsp = flumeSampleResponse
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &Flume{
		Servers: []string{fmt.Sprintf("%s/metrics", ts.URL)},
	}

	var accFlume testutil.Accumulator

	errFlume := accFlume.GatherError(n.Gather)

	require.NoError(t, errFlume)

	flumeSampleMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(flumeSampleResponse), &flumeSampleMap)
	if err != nil {
		panic(err)
	}

	for c, _ := range flumeSampleMap {

		tags := map[string]string{"component": c, "server": ts.URL + "/metrics"}

		component := strings.Split(c, ".")[0]
		accFlume.AssertContainsTaggedFields(t, "flume_"+component, flumeSampleMap[c].(map[string]interface{}), tags)
	}

}
