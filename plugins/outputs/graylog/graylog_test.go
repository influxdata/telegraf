package graylog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
)

func TestSerializer(t *testing.T) {
	m1 := metric.New("testing",
		map[string]string{
			"verb": "GET",
			"host": "hostname",
		},
		map[string]interface{}{
			"full_message":  "full",
			"short_message": "short",
			"level":         "1",
			"facility":      "demo",
			"line":          "42",
			"file":          "graylog.go",
		},
		time.Now(),
	)

	graylog := Graylog{}
	result, err := graylog.serialize(m1)

	require.NoError(t, err)

	for _, r := range result {
		obj := make(map[string]interface{})
		err = json.Unmarshal([]byte(r), &obj)
		require.NoError(t, err)

		require.Equal(t, "1.1", obj["version"])
		require.Equal(t, "testing", obj["_name"])
		require.Equal(t, "GET", obj["_verb"])
		require.Equal(t, "hostname", obj["host"])
		require.Equal(t, "full", obj["full_message"])
		require.Equal(t, "short", obj["short_message"])
		require.Equal(t, "1", obj["level"])
		require.Equal(t, "demo", obj["facility"])
		require.Equal(t, "42", obj["line"])
		require.Equal(t, "graylog.go", obj["file"])
	}
}
