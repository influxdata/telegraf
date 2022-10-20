//go:generate ../../../tools/readme_config_includer/generator
package opcua

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"strings"
	"time"

	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

type OpcuaWorkarounds struct {
	UseUnregisteredReads bool `toml:"use_unregistered_reads"`
}

// OpcUA type
type OpcUA struct {
	Workarounds OpcuaWorkarounds `toml:"workarounds"`
	Log         telegraf.Logger  `toml:"-"`

	// status
	ReadSuccess selfstat.Stat `toml:"-"`
	ReadError   selfstat.Stat `toml:"-"`

	// internal values
	client *input.OpcUAInputClient
	req    *ua.ReadRequest
}

func (*OpcUA) SampleConfig() string {
	return sampleConfig
}

// Init will initialize all tags
func (o *OpcUA) Init() error {
	err := o.client.Init()
	if err != nil {
		return err
	}

	err = o.client.InitNodeMetricMapping()
	if err != nil {
		return err
	}

	tags := map[string]string{
		"endpoint": o.client.Config.Endpoint,
	}
	o.ReadError = selfstat.Register("opcua", "read_error", tags)
	o.ReadSuccess = selfstat.Register("opcua", "read_success", tags)

	return nil
}

// Connect to a OPCUA device
func Connect(o *OpcUA) error {
	err := o.client.Connect()
	if err != nil {
		return err
	}

	if !o.Workarounds.UseUnregisteredReads {
		regResp, err := o.client.Client.RegisterNodes(&ua.RegisterNodesRequest{
			NodesToRegister: o.client.NodeIDs,
		})
		if err != nil {
			return fmt.Errorf("registerNodes failed: %v", err)
		}

		o.req = &ua.ReadRequest{
			MaxAge:             2000,
			TimestampsToReturn: ua.TimestampsToReturnBoth,
			NodesToRead:        readvalues(regResp.RegisteredNodeIDs),
		}
	} else {
		var nodesToRead []*ua.ReadValueID

		for _, nid := range o.client.NodeIDs {
			nodesToRead = append(nodesToRead, &ua.ReadValueID{NodeID: nid})
		}

		o.req = &ua.ReadRequest{
			MaxAge:             2000,
			TimestampsToReturn: ua.TimestampsToReturnBoth,
			NodesToRead:        nodesToRead,
		}
	}

	err = o.getData()
	if err != nil {
		return fmt.Errorf("get Data Failed: %v", err)
	}

	return nil
}

func (o *OpcUA) getData() error {
	resp, err := o.client.Client.Read(o.req)
	if err != nil {
		o.ReadError.Incr(1)
		return fmt.Errorf("Read failed: %w", err)
	}
	o.ReadSuccess.Incr(1)
	for i, d := range resp.Results {
		o.client.LastReceivedData[i].Quality = d.Status
		if !o.client.StatusCodeOK(d.Status) {
			// removed error logging because of easier merging
			continue
		}

		o.client.LastReceivedData[i].TagName = o.client.NodeMetricMapping[i].Tag.FieldName
		if d.Value != nil {
			o.client.LastReceivedData[i].Value = d.Value.Value()
			o.client.LastReceivedData[i].DataType = d.Value.Type()
		}
		o.client.LastReceivedData[i].Quality = d.Status
		o.client.LastReceivedData[i].ServerTime = d.ServerTimestamp
		o.client.LastReceivedData[i].SourceTime = d.SourceTimestamp
	}
	return nil
}

func readvalues(ids []*ua.NodeID) []*ua.ReadValueID {
	rvids := make([]*ua.ReadValueID, len(ids))
	for i, v := range ids {
		rvids[i] = &ua.ReadValueID{NodeID: v}
	}
	return rvids
}

// Gather defines what data the plugin will gather.
func (o *OpcUA) Gather(acc telegraf.Accumulator) error {
	if o.client.State == opcua.Disconnected {
		err := o.client.Connect()
		if err != nil {
			return err
		}
	}

	err := o.getData()
	if err != nil && o.client.State == opcua.Connected {
		// Ignore returned error to not mask the original problem
		//nolint:errcheck,revive
		o.client.Disconnect(context.Background())
		return err
	}

	for i, n := range o.client.NodeMetricMapping {
		if o.client.StatusCodeOK(o.client.LastReceivedData[i].Quality) {
			fields := make(map[string]interface{})
			tags := map[string]string{
				"id": n.Tag.NodeID(),
			}
			for k, v := range n.MetricTags {
				tags[k] = v
			}

			fields[o.client.LastReceivedData[i].TagName] = o.client.LastReceivedData[i].Value
			fields["Quality"] = strings.TrimSpace(fmt.Sprint(o.client.LastReceivedData[i].Quality))

			switch o.client.Config.Timestamp {
			case "server":
				acc.AddFields(n.Tag.FieldName, fields, tags, o.client.LastReceivedData[i].ServerTime)
			case "source":
				acc.AddFields(n.Tag.FieldName, fields, tags, o.client.LastReceivedData[i].SourceTime)
			default:
				acc.AddFields(n.Tag.FieldName, fields, tags)
			}
		}
	}
	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("opcua", func() telegraf.Input {
		return &OpcUA{
			client: &input.OpcUAInputClient{
				OpcUAClient: &opcua.OpcUAClient{
					Config: &opcua.OpcUAClientConfig{
						Endpoint:       "opc.tcp://localhost:4840",
						SecurityPolicy: "auto",
						SecurityMode:   "auto",
						RequestTimeout: config.Duration(5 * time.Second),
						ConnectTimeout: config.Duration(10 * time.Second),
						Certificate:    "/etc/telegraf/cert.pem",
						PrivateKey:     "/etc/telegraf/key.pem",
						AuthMethod:     "Anonymous",
						Username:       "",
						Password:       "",
						Workarounds:    opcua.OpcUAWorkarounds{},
					},
				},
				Config: input.InputClientConfig{
					OpcUAClientConfig: opcua.OpcUAClientConfig{},
					MetricName:        "opcua",
					Timestamp:         "gather",
					RootNodes:         nil,
					Groups:            nil,
				},
			},
		}
	})
}
