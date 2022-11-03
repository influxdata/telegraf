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
	req *ua.ReadRequest
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
	err := o.OpcUAClient.Connect()
	if err != nil {
		return err
	}

	readValueIds := make([]*ua.ReadValueID, len(o.NodeIDs))
	if o.Workarounds.UseUnregisteredReads {
		for i, nid := range o.NodeIDs {
			readValueIds[i] = &ua.ReadValueID{NodeID: nid}
		}
	} else {
		regResp, err := o.Client.RegisterNodes(&ua.RegisterNodesRequest{
			NodesToRegister: o.NodeIDs,
		})
		if err != nil {
			return fmt.Errorf("registerNodes failed: %v", err)
		}

		for i, v := range regResp.RegisteredNodeIDs {
			readValueIds[i] = &ua.ReadValueID{NodeID: v}
		}
	}

	o.req = &ua.ReadRequest{
		MaxAge:             2000,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		NodesToRead:        readValueIds,
	}

	err = o.read()
	if err != nil {
		return fmt.Errorf("get Data Failed: %v", err)
	}

	return nil
}

func (o *ReadClient) ensureConnected() error {
	if o.State == opcua.Disconnected {
		err := o.Connect()
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *ReadClient) CurrentValues() ([]telegraf.Metric, error) {
	err := o.ensureConnected()
	if err != nil {
		return nil, err
	}

	err = o.read()
	if err != nil && o.State == opcua.Connected {
		// We do not return the disconnect error, as this would mask the
		// original problem, but we do log it
		disconnectErr := o.Disconnect(context.Background())
		if disconnectErr != nil {
			o.Log.Debug("Error while disconnecting: ", disconnectErr)
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
	resp, err := o.Client.Read(o.req)
	if err != nil {
		o.ReadError.Incr(1)
		return fmt.Errorf("RegisterNodes Read failed: %v", err)
	}
	o.ReadSuccess.Incr(1)
	for i, d := range resp.Results {
		o.UpdateNodeValue(i, d)
	}
	return nil
}
