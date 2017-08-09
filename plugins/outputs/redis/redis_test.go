package redis

import (
	"testing"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	r := &RedisOutput{
		Addr:     "127.0.0.1:6379",
		Password: "",
		Queue:    "",
	}

	err := r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)

	//mockMetrics -> {"body":{"value":1},"tag1":"value1"}
	conn := r.server.Get()
	defer conn.Close()

	jstr, err := redigo.String(conn.Do("RPOP", r.Queue))
	require.NoError(t, err)

	value := gjson.Get(jstr, "body.value").Int()
	tag1 := gjson.Get(jstr, "tag1").String()

	require.Equal(t, int64(1), value, "metrics field body.value 1 = 1")
	require.Equal(t, "value1", tag1, "tag tag1 value1=value1")

	err = r.Close()
	require.NoError(t, err)
}
