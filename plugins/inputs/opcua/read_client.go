package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/selfstat"
)

type ReadClientWorkarounds struct {
	UseUnregisteredReads bool `toml:"use_unregistered_reads"`
}

type ReadClientConfig struct {
	ReadClientWorkarounds ReadClientWorkarounds `toml:"request_workarounds"`
	input.InputClientConfig
}

// ReadClient Requests the current values from the required nodes when gather is called.
type ReadClient struct {
	*input.OpcUAInputClient

	ReadSuccess selfstat.Stat
	ReadError   selfstat.Stat
	Workarounds ReadClientWorkarounds

	// internal values
	reqIDs []*ua.ReadValueID
	ctx    context.Context
}

func (rc *ReadClientConfig) CreateReadClient(log telegraf.Logger) (*ReadClient, error) {
	inputClient, err := rc.InputClientConfig.CreateInputClient(log)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		"endpoint": inputClient.Config.OpcUAClientConfig.Endpoint,
	}

	return &ReadClient{
		OpcUAInputClient: inputClient,
		ReadSuccess:      selfstat.Register("opcua", "read_success", tags),
		ReadError:        selfstat.Register("opcua", "read_error", tags),
		Workarounds:      rc.ReadClientWorkarounds,
	}, nil
}

func (o *ReadClient) Connect() error {
	o.ctx = context.Background()

	if err := o.OpcUAClient.Connect(o.ctx); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	// Make sure we setup the node-ids correctly after reconnect
	// as the server might be restarted and IDs changed
	if err := o.OpcUAInputClient.InitNodeIDs(); err != nil {
		return fmt.Errorf("initializing node IDs failed: %w", err)
	}

	o.reqIDs = make([]*ua.ReadValueID, 0, len(o.NodeIDs))
	if o.Workarounds.UseUnregisteredReads {
		for _, nid := range o.NodeIDs {
			o.reqIDs = append(o.reqIDs, &ua.ReadValueID{NodeID: nid})
		}
	} else {
		regResp, err := o.Client.RegisterNodes(o.ctx, &ua.RegisterNodesRequest{
			NodesToRegister: o.NodeIDs,
		})
		if err != nil {
			return fmt.Errorf("registering nodes failed: %w", err)
		}

		for _, v := range regResp.RegisteredNodeIDs {
			o.reqIDs = append(o.reqIDs, &ua.ReadValueID{NodeID: v})
		}
	}

	if err := o.read(); err != nil {
		return fmt.Errorf("get data failed: %w", err)
	}

	return nil
}

func (o *ReadClient) ensureConnected() error {
	if o.State() == opcua.Disconnected || o.State() == opcua.Closed {
		return o.Connect()
	}
	return nil
}

func (o *ReadClient) CurrentValues() ([]telegraf.Metric, error) {
	if err := o.ensureConnected(); err != nil {
		return nil, err
	}

	if state := o.State(); state != opcua.Connected {
		return nil, fmt.Errorf("not connected, in state %q", state)
	}

	if err := o.read(); err != nil {
		// We do not return the disconnect error, as this would mask the
		// original problem, but we do log it
		if derr := o.Disconnect(context.Background()); derr != nil {
			o.Log.Debug("Error while disconnecting: ", derr)
		}

		return nil, err
	}

	metrics := make([]telegraf.Metric, 0, len(o.NodeMetricMapping))
	// Parse the resulting data into metrics
	for i := range o.NodeIDs {
		if !o.StatusCodeOK(o.LastReceivedData[i].Quality) {
			continue
		}

		metrics = append(metrics, o.MetricForNode(i))
	}

	return metrics, nil
}

func (o *ReadClient) read() error {
	req := &ua.ReadRequest{
		MaxAge:             2000,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		NodesToRead:        o.reqIDs,
	}

	resp, err := o.Client.Read(o.ctx, req)
	if err != nil {
		o.ReadError.Incr(1)
		return fmt.Errorf("RegisterNodes Read failed: %w", err)
	}
	o.ReadSuccess.Incr(1)
	for i, d := range resp.Results {
		o.UpdateNodeValue(i, d)
	}
	return nil
}
