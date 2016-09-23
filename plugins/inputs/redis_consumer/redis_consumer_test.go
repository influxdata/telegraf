package redis_consumer

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/redis.v4"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChannels(t *testing.T) {
	psubChannel := `channel[1_3]`
	plainSubChannel := `channel`
	escapeSubChannel := `channel\*`

	tests := []struct {
		testName  string
		channels  []string
		psubCount int
		subCount  int
	}{
		{
			testName:  "normal sub",
			channels:  []string{plainSubChannel},
			psubCount: 0,
			subCount:  1,
		},
		{
			testName:  "psub",
			channels:  []string{psubChannel},
			psubCount: 1,
			subCount:  0,
		},
		{
			testName:  "escaped sub",
			channels:  []string{escapeSubChannel},
			psubCount: 0,
			subCount:  1,
		},
		{
			testName:  "all",
			channels:  []string{escapeSubChannel, psubChannel, plainSubChannel},
			psubCount: 1,
			subCount:  2,
		},
	}

	for _, parseTest := range tests {
		s, p, e := parseChannels(parseTest.channels)

		if e != nil {
			t.Errorf("Test %s had unexpected error %v", parseTest.testName, e)
		}

		if parseTest.subCount != len(s) {
			t.Errorf("Test %s subchanel count. Expected %d Actual %d", parseTest.testName, parseTest.subCount, len(s))
		}

		if parseTest.psubCount != len(p) {
			t.Errorf("Test %s psubchanel count. Expected %d Actual %d", parseTest.testName, parseTest.psubCount, len(p))
		}
	}
}

func TestRedisConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf(testutil.GetLocalHost() + ":6379")
	c, err := createClient(addr)
	require.NoError(t, err)
	c.Close()
}

func TestCreateSubscriptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf(testutil.GetLocalHost() + ":6379")
	c, _ := createClient(addr)
	defer c.Close()
	r := &RedisConsumer{
		Channels: []string{"test_channel_1", "test_channel_2, test_channel_[1_3]"},
		clients:  []*redis.Client{c},
	}

	pubsubs, err := r.createSubscriptions()
	require.NoError(t, err)
	assert.Equal(t, 2, len(pubsubs))

}

func TestRedisReceive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testChannel := "test_channel"
	testMsg := "cpu_load,host=server01,region=us-west value=1.0 1444444444"
	addr := fmt.Sprintf(testutil.GetLocalHost() + ":6379")
	testClient, err := createClient(addr)
	defer testClient.Close()

	parser, _ := parsers.NewInfluxParser()
	var acc testutil.Accumulator

	require.NoError(t, err)
	r := &RedisConsumer{
		Servers:  []string{addr},
		Channels: []string{testChannel},
	}

	r.SetParser(parser)

	if err = r.Start(&acc); err != nil {
		t.Fatal(err.Error())
	}
	defer r.Stop()

	testClient.Publish(testChannel, testMsg)
	waitForPoint(&acc, 2, t)

	if len(acc.Metrics) != 1 {
		t.Error("Metric no receieved")
	}
}

// Waits for the metric to arrive in the accumulator
func waitForPoint(acc *testutil.Accumulator, waitSeconds int, t *testing.T) {
	intervalMS := 5
	threshold := (waitSeconds * 1000) / intervalMS
	ticker := time.NewTicker(time.Duration(intervalMS) * time.Millisecond)
	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			if counter > threshold {
				t.Fatalf("Waited for %ds, point never arrived to accumulator", waitSeconds)
			} else if acc.NFields() == 1 {
				return
			}
		}
	}
}
