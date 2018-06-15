package splunkhec

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	//MockMetrics returns a single event that matches this:
	const validResult = "{\"time\":1257894000,\"event\":\"metric\",\"source\":\"telegraf\",\"host\":\"\",\"fields\":{\"_value\":1,\"metric_name\":\"test1.value\",\"tag1\":\"value1\"}}"

	d := &SplunkHEC{}

	v, _ := json.Marshal(validResult)

	hecMs, err := buildMetrics(testutil.MockMetrics()[0], d)
	require.NoError(t, err)
	b, err := json.Marshal(hecMs)
	require.NoError(t, err)
	require.Equal(t, v, b)
}
