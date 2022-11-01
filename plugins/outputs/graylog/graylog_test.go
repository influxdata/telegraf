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

		require.Equal(t, obj["version"], "1.1")
		require.Equal(t, obj["_name"], "testing")
		require.Equal(t, obj["_verb"], "GET")
		require.Equal(t, obj["host"], "hostname")
		require.Equal(t, obj["full_message"], "full")
		require.Equal(t, obj["short_message"], "short")
		require.Equal(t, obj["level"], "1")
		require.Equal(t, obj["facility"], "demo")
		require.Equal(t, obj["line"], "42")
		require.Equal(t, obj["file"], "graylog.go")
	}
}
