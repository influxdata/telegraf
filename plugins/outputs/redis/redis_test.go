package redis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type T_redisoutput struct {
	Name      string                 `json:"name"`
	Fields    map[string]interface{} `json:"fields"`
	Tags      map[string]interface{} `json:"tags"`
	Timestamp int64                  `json:"timestamp"`
}

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s, _ := serializers.NewJsonSerializer(time.Second)

	idleTimeout := internal.Duration{}
	err := idleTimeout.UnmarshalTOML([]byte("0s"))
	require.NoError(t, err)

	rwTimeout := internal.Duration{}
	err = rwTimeout.UnmarshalTOML([]byte("1s"))
	require.NoError(t, err)

	r := &RedisOutput{
		Server:      "127.0.0.1:6379",
		Password:    "",
		Queue:       "telegraf/redis_test",
		IdleTimeout: idleTimeout,
		Timeout:     rwTimeout,
		serializer:  s,
	}

	err = r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)

	//mockMetrics ->
	//{
	//    "fields": {
	//        "value": 1
	//    },
	//    "name": "test1",
	//    "tags": {
	//        "tag1": "value1"
	//    },
	//    "timestamp": 1257894000
	//}

	bs, err := r.server.RPop(r.Queue).Result()
	require.NoError(t, err)

	m := T_redisoutput{}
	json.Unmarshal([]byte(bs), &m)

	require.Equal(t, m.Fields["value"], float64(1), "field value 1 == 1")
	require.Equal(t, m.Tags["tag1"], "value1", "tag tag1 values == values")

	err = r.Close()
	require.NoError(t, err)
}
