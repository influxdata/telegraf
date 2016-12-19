package kinesis

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFormatMetric(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	k := &KinesisOutput{
		Format: "string",
	}

	p := testutil.MockMetrics()[0]

	valid_string := "test1,tag1=value1 value=1 1257894000000000000\n"
	func_string, err := FormatMetric(k, p)

	if func_string != valid_string {
		t.Error("Expected ", valid_string)
	}
	require.NoError(t, err)

	k = &KinesisOutput{
		Format: "json",
	}

	valid_json := "{\"fields\":{\"value\":1},\"name\":\"test1\",\"tags\":{\"tag1\":\"value1\"},\"timestamp\":1257894000}"
	func_json, err := FormatMetric(k, p)

	if func_json != valid_json {
		t.Error("Expected ", valid_json)
		t.Error("Found ", func_json)
	}
	require.NoError(t, err)

	k = &KinesisOutput{
		Format: "custom",
	}

	valid_custom := "test1,map[tag1:value1],test1,tag1=value1 value=1 1257894000000000000\n"
	func_custom, err := FormatMetric(k, p)

	if func_custom != valid_custom {
		t.Error("Expected ", valid_custom)
	}
	require.NoError(t, err)
}
