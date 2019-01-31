package pulsar_consumer

import (
	"context"
	"testing"
	"time"

	"github.com/Comcast/pulsar-client-go"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	testMsg    = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
	invalidMsg = "cpu_load_short,host=server01 1422568543702900257\n"
)

func newTestPulsarConsumer() *pulsarConsumer {
	p := &pulsarConsumer{
		URL:                    "pulsar://" + testutil.GetLocalHost() + ":6650",
		Topic:                  "telegraf",
		Name:                   "telegraf-consumer",
		QueueSize:              1000,
		DialTimeout:            "15s",
		RecvTimeout:            "15s",
		PingFrequency:          "30s",
		PingTimeout:            "15s",
		InitialReconnectDelay:  "15s",
		MaxReconnectDelay:      "15s",
		NewConsumerTimeout:     "15s",
		MaxUndeliveredMessages: 1000,
	}
	pr, _ := parsers.NewInfluxParser()
	p.SetParser(pr)

	return p
}

func newTestPulsarProducer() *pulsar.ManagedProducer {
	cp := pulsar.NewManagedClientPool()

	conf := pulsar.ManagedProducerConfig{
		Topic: "telegraf",
		Name:  "telegraf-producer",
	}
	conf.Addr = "pulsar://" + testutil.GetLocalHost() + ":6650"
	return pulsar.NewManagedProducer(cp, conf)
}

func TestReadMetricsFromPulsar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	acc := &testutil.Accumulator{}

	pc := newTestPulsarConsumer()
	require.NotNil(t, pc)
	require.IsType(t, &pulsarConsumer{}, pc)

	err := pc.Start(acc)
	require.NoError(t, err)

	<-time.After(time.Second)

	pp := newTestPulsarProducer()
	for i := 0; i < 1000; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = pp.Send(ctx, []byte(testMsg))
		require.NoError(t, err)
	}

	acc.Wait(1000)
	pc.Stop()
	require.Equal(t, uint64(1000), acc.NMetrics())
}
