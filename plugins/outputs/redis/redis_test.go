package redis

import (
	"encoding/json"
	"testing"
	"time"

	redigo "github.com/garyburd/redigo/redis"
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
	r := &RedisOutput{
		Addr:       "127.0.0.1:6379",
		Password:   "",
		Queue:      "",
		serializer: s,
	}

	err := r.Connect()
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
	conn := r.server.Get()
	defer conn.Close()

	bs, err := redigo.Bytes(conn.Do("RPOP", r.Queue))
	require.NoError(t, err)

	m := T_redisoutput{}
	json.Unmarshal(bs, &m)

	require.Equal(t, m.Fields["value"], float64(1), "field value 1 == 1")
	require.Equal(t, m.Tags["tag1"], "value1", "tag tag1 values == values")

	err = r.Close()
	require.NoError(t, err)
}
