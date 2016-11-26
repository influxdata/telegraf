package laravel_redis_queue

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestRedisConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf(testutil.GetLocalHost() + ":6379")

	r := &LaravelRedisQueue{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)
}

func TestLaravelRedisQueue_ParseMetrics(t *testing.T) {
	var acc testutil.Accumulator
	tags := map[string]string{"host": "redis.net"}
	rdr := bufio.NewReader(strings.NewReader(testOutput))

	err := gatherInfoOutput(rdr, &acc, tags, "pushed_count")
	require.NoError(t, err)

	tags = map[string]string{"host": "redis.net"}
	fields := map[string]interface{}{
		"pushed_count": uint64(1),
	}

	// We have to test rdb_last_save_time_offset manually because the value is based on the time when gathered
	for _, m := range acc.Metrics {
		for k, v := range m.Fields {
			fields[k] = v
		}
	}

	acc.AssertContainsTaggedFields(t, "laravel_redis_queue", fields, tags)
}

const testOutput = `1`
