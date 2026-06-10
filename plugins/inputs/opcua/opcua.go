//go:generate ../../../tools/readme_config_includer/generator
package opcua

import (
	"context"
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type OpcUA struct {
	readClientConfig
	Log telegraf.Logger `toml:"-"`

	client *readClient

	// Add a consecutive error counter to potentially force reconnection
	consecutiveErrors uint64
}

func (*OpcUA) SampleConfig() string {
	return sampleConfig
}

func (o *OpcUA) Init() (err error) {
	o.client, err = o.readClientConfig.createReadClient(o.Log)
	return err
}

// Start implements the ServiceInput interface so the agent calls Stop on
// shutdown and config reload. The connection itself is established lazily in
// Gather to keep tolerating a server that is unavailable at startup.
func (*OpcUA) Start(telegraf.Accumulator) error {
	return nil
}

func (o *OpcUA) Gather(acc telegraf.Accumulator) error {
	gatherStart := time.Now()
	o.Log.Tracef("Gather starting for %d nodes...", len(o.client.NodeIDs))

	// Force reconnection every time if a threshold is 0
	if o.client.ReconnectErrorThreshold == 0 {
		o.client.forceReconnect = true
	}

	// Will (re)connect if the client is disconnected
	metrics, err := o.client.currentValues()
	if err != nil {
		o.consecutiveErrors++
		o.Log.Tracef("Gather failed after %s: %v (consecutive errors: %d)",
			time.Since(gatherStart), err, o.consecutiveErrors)

		// Force reconnection based on an error threshold: if threshold > 0, reconnect after
		// reaching the specified number of consecutive errors; if a threshold = 0, we already
		// force the reconnection above, so skip this check
		if o.client.ReconnectErrorThreshold > 0 && o.consecutiveErrors >= o.client.ReconnectErrorThreshold {
			o.client.forceReconnect = true
		}
		return err
	}

	// Reset error counter on success
	o.consecutiveErrors = 0

	addStart := time.Now()
	// Parse the resulting data into metrics
	for _, m := range metrics {
		acc.AddMetric(m)
	}
	o.Log.Tracef("Gather complete: %d metrics added to accumulator in %s, total gather time %s",
		len(metrics), time.Since(addStart), time.Since(gatherStart))
	return nil
}

// Stop releases the OPC UA session so it is not orphaned on the server during
// shutdown or config reload. Without this the session lingers until the
// server's session timeout, and repeated reloads can exhaust the session limit.
func (o *OpcUA) Stop() {
	if o.client == nil {
		return
	}
	if state := o.client.State(); state != opcua.Connected && state != opcua.Connecting {
		return
	}

	// Bound the disconnect by the configured request timeout so a stuck server
	// cannot hang shutdown. A zero timeout means "no limit", so fall back to a
	// sane default to avoid an immediately-expired context.
	timeout := time.Duration(o.client.Config.RequestTimeout)
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := o.client.Disconnect(ctx); err != nil {
		o.Log.Warnf("Disconnecting from OPC UA server failed: %v", err)
	}
}

func init() {
	inputs.Add("opcua", func() telegraf.Input {
		return &OpcUA{
			readClientConfig: readClientConfig{
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
			},
		}
	})
}
