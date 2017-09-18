package tile38

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/tile38/controller"
)

var (
	host     = testutil.GetLocalHost()
	port     = rand.Int()%20000 + 20000
	password = "passw0rd"
)

func init() {
	var t *testing.T
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	mockTile38()
}

func mockTile38() {
	mockCleanup()

	dir := fmt.Sprintf("data-mock-%d", port)
	fmt.Printf("Starting test server at port %d\n", port)
	go func() {
		if err := controller.ListenAndServe(host, port, dir, false); err != nil {
			log.Fatal(err)
		}
	}()
	conn, _ := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	conn.Do("SET", "TEST_KEY", "TRUNK1", "POINT", "33.5123", "-112.2693")
	conn.Do("CONFIG", "SET", "requirepass", password)
}

func mockCleanup() {
	fmt.Printf("Cleanup: may take some time... ")
	files, _ := ioutil.ReadDir(".")
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "data-mock-") {
			os.RemoveAll(file.Name())
		}
	}
	fmt.Printf("OK\n")
}

func TestTile38ConnectWithAuthentication(t *testing.T) {
	addr := fmt.Sprintf("tcp://:%s@%s:%d", password, host, port)
	url, _ := url.Parse(addr)

	client := initTile38(url)
	conn := client.Get()
	defer client.Close()
	result, err := redis.String(conn.Do("PING"))

	require.NoError(t, err)
	require.Equal(t, "pong", gjson.Get(result, "ping").String())
}

func TestTile38ParseMetrics(t *testing.T) {
	addr := fmt.Sprintf("tcp://:%s@%s:%d", password, host, port)

	tt := &Tile38{
		Servers: []string{addr},
		Stats:   true,
	}

	var acc testutil.Accumulator

	err := acc.GatherError(tt.Gather)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("tile38_server"))

	tags := [...]string{"server", "port", "id"}
	fields := [...]string{
		"aof_size",
		"avg_item_size",
		"heap_released",
		"heap_size",
		"http_transport",
		"in_memory_size",
		"max_heap_size",
		"mem_alloc",
		"num_collections",
		"num_hooks",
		"num_objects",
		"num_points",
		"num_strings",
		"pid",
		"pointer_size",
		"read_only"}

	for _, tag := range tags {
		require.True(t, acc.HasTag("tile38_server", tag))
	}
	for _, field := range fields {
		require.True(t, acc.HasField("tile38_server", field))
	}

	require.True(t, acc.HasMeasurement("tile38_stats"))
	keytags := [...]string{"server", "port", "id", "key"}
	keyfields := [...]string{
		"in_memory_size",
		"num_objects",
		"num_points",
		"num_strings"}

	for _, tag := range keytags {
		require.True(t, acc.HasTag("tile38_stats", tag))
	}
	for _, field := range keyfields {
		require.True(t, acc.HasField("tile38_stats", field))
	}
}
