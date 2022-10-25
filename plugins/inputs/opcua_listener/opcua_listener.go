package opcua_listener

import (
	"context"
	_ "embed"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
)

type OpcUaListener struct {
	SubscribeClientConfig
	client *SubscribeClient
	Log    telegraf.Logger `toml:"-"`
}

//go:embed sample.conf
var sampleConfig string

func (*OpcUaListener) SampleConfig() string {
	return sampleConfig
}

func (o *OpcUaListener) Init() (err error) {
	o.client, err = o.SubscribeClientConfig.CreateSubscribeClient(o.Log)
	return err
}

func (o *OpcUaListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (o *OpcUaListener) Start(acc telegraf.Accumulator) error {
	ctx := context.Background()
	ch, err := o.client.StartStreamValues(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			m, ok := <-ch
			if !ok {
				o.Log.Debug("Metric collection stopped due to closed channel")
				return
			}
			acc.AddMetric(m)
		}
	}()

	return nil
}

func (o *OpcUaListener) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	select {
	case <-o.client.Stop(ctx):
		o.Log.Infof("Unsubscribed OPC UA successfully")
	case <-ctx.Done(): // Timeout context
		o.Log.Warn("Timeout while stopping OPC UA subscription")
	}
	cancel()
}

// Add this plugin to telegraf
func init() {
	inputs.Add("opcua_listener", func() telegraf.Input {
		return &OpcUaListener{
			SubscribeClientConfig: SubscribeClientConfig{
				InputClientConfig: input.InputClientConfig{
					OpcUAClientConfig: opcua.OpcUAClientConfig{
						Endpoint:       "opc.tcp://localhost:4840",
						SecurityPolicy: "auto",
						SecurityMode:   "auto",
						Certificate:    "/etc/telegraf/cert.pem",
						PrivateKey:     "/etc/telegraf/key.pem",
						AuthMethod:     "Anonymous",
						ConnectTimeout: config.Duration(5 * time.Second),
						RequestTimeout: config.Duration(10 * time.Second),
					},
					MetricName: "opcua",
					Timestamp:  input.TimestampSourceTelegraf,
				},
				SubscriptionInterval: config.Duration(100 * time.Millisecond),
			},
		}
	})
}
