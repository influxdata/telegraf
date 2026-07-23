//go:generate ../../../tools/readme_config_includer/generator
package cisco_telemetry_mdt

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	mdtdialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // Required to allow gzip encoding
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// CiscoTelemetryMDT plugin for IOS XR, IOS XE and NXOS platforms
type CiscoTelemetryMDT struct {
	// Common configuration
	Transport          string                `toml:"transport"`
	ServiceAddress     string                `toml:"service_address"`
	MaxMsgSize         config.Size           `toml:"max_msg_size"`
	Aliases            map[string]string     `toml:"aliases"`
	Dmes               map[string]string     `toml:"dmes"`
	EmbeddedTags       []string              `toml:"embedded_tags"`
	EnforcementPolicy  grpcEnforcementPolicy `toml:"grpc_enforcement_policy"`
	IncludeDeleteField bool                  `toml:"include_delete_field"`
	SourceFieldName    string                `toml:"source_field_name"`
	Log                telegraf.Logger       `toml:"-"`
	common_tls.ServerConfig

	serverOptions []grpc.ServerOption
	grpcServer    *grpc.Server
	listener      net.Listener

	parser *parser

	acc telegraf.Accumulator
	wg  sync.WaitGroup

	// Though unused in the code, required by protoc-gen-go-grpc to maintain compatibility
	mdtdialout.UnimplementedGRPCMdtDialoutServer
}

type grpcEnforcementPolicy struct {
	PermitKeepaliveWithoutCalls bool            `toml:"permit_keepalive_without_calls"`
	KeepaliveMinTime            config.Duration `toml:"keepalive_minimum_time"`
}

func (*CiscoTelemetryMDT) SampleConfig() string {
	return sampleConfig
}

func (c *CiscoTelemetryMDT) Init() error {
	if c.Transport == "" {
		c.Transport = "grpc"
	}

	switch c.Transport {
	case "tcp":
	case "grpc":
		// Prepare the server options
		tlsConfig, err := c.ServerConfig.TLSConfig()
		if err != nil {
			return fmt.Errorf("creating TLS server configuration failed: %w", err)
		}
		if tlsConfig != nil {
			c.serverOptions = append(c.serverOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
		}

		if c.MaxMsgSize > 0 {
			c.serverOptions = append(c.serverOptions, grpc.MaxRecvMsgSize(int(c.MaxMsgSize)))
		}

		// Only set if either parameter does not match defaults
		shouldSet := c.EnforcementPolicy.KeepaliveMinTime != 0
		shouldSet = shouldSet && c.EnforcementPolicy.KeepaliveMinTime != config.Duration(time.Second*300)
		shouldSet = shouldSet || c.EnforcementPolicy.PermitKeepaliveWithoutCalls
		if shouldSet {
			c.serverOptions = append(c.serverOptions, grpc.KeepaliveEnforcementPolicy(
				keepalive.EnforcementPolicy{
					MinTime:             time.Duration(c.EnforcementPolicy.KeepaliveMinTime),
					PermitWithoutStream: c.EnforcementPolicy.PermitKeepaliveWithoutCalls,
				},
			))
		}
	default:
		return fmt.Errorf("invalid transport %q", c.Transport)
	}

	if c.ServiceAddress == "" {
		c.ServiceAddress = "127.0.0.1:57000"
	}

	if c.SourceFieldName == "" {
		c.SourceFieldName = "mdt_source"
	}

	// Initialize parser
	c.parser = newParser(c.IncludeDeleteField, c.Aliases, c.Dmes, c.EmbeddedTags, c.Log)

	return nil
}

// Start the Cisco MDT service
func (c *CiscoTelemetryMDT) Start(acc telegraf.Accumulator) error {
	c.acc = acc

	// Start server
	listener, err := net.Listen("tcp", c.ServiceAddress)
	if err != nil {
		return fmt.Errorf("creating listener on %q failed: %w", c.ServiceAddress, err)
	}
	c.listener = listener

	switch c.Transport {
	case "tcp":
		// TCP dialout server accept routine
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.acceptTCPClients()
		}()
	case "grpc":
		c.grpcServer = grpc.NewServer(c.serverOptions...)
		mdtdialout.RegisterGRPCMdtDialoutServer(c.grpcServer, c)

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			if err := c.grpcServer.Serve(c.listener); err != nil {
				c.Log.Errorf("serving GRPC server failed: %v", err)
			}
		}()
	default:
		return fmt.Errorf("invalid Cisco MDT transport: %s", c.Transport)
	}

	return nil
}

func (c *CiscoTelemetryMDT) Stop() {
	if c.grpcServer != nil {
		// Stop server and terminate all running dialout routines
		c.grpcServer.Stop()
	}
	if c.listener != nil {
		c.listener.Close()
	}
	c.wg.Wait()
}

func (*CiscoTelemetryMDT) Gather(telegraf.Accumulator) error {
	return nil
}

func (c *CiscoTelemetryMDT) handleTelemetry(data []byte) {
	msg := &telemetry.Telemetry{}
	if err := proto.Unmarshal(data, msg); err != nil {
		c.acc.AddError(fmt.Errorf("failed to decode: %w: %s", err, msg.String()))
		return
	}

	if c.Log.Level().Includes(telegraf.Trace) {
		if d, err := protojson.Marshal(msg); err != nil {
			c.Log.Tracef("reencoding message %q failed: %v", hex.EncodeToString(data), err)
		} else {
			c.Log.Tracef("received message: %s", string(d))
		}
	}

	// Create a grouper to accumulate the fields for a series
	grouper := metric.NewSeriesGrouper()

	for _, gpbkv := range msg.DataGpbkv {
		// Top-level field may have measurement timestamp, if not use message timestamp
		measured := gpbkv.Timestamp
		if measured == 0 {
			measured = msg.MsgTimestamp
		}
		timestamp := time.Unix(0, int64(measured)*1000000)

		// Find toplevel GPBKV fields "keys" and "content"
		var keys, content *telemetry.TelemetryField
		for _, field := range gpbkv.Fields {
			switch field.Name {
			case "keys":
				keys = field
			case "content":
				content = field
			}
		}

		if content == nil && !c.IncludeDeleteField {
			c.Log.Debug("Message skipped because no content found and include of delete field not enabled")
			continue
		}

		// Produce metadata tags
		tags := make(map[string]string, 3)
		if keys != nil {
			for _, subfield := range keys.Fields {
				parseKeys(subfield, "", tags)
			}

			// If incoming MDT contains source key, copy to mdt_src
			if _, ok := tags["source"]; ok {
				tags[c.SourceFieldName] = tags["source"]
			}
		}
		tags["source"] = msg.GetNodeIdStr()
		if msgID := msg.GetSubscriptionIdStr(); msgID != "" {
			tags["subscription"] = msgID
		}
		encodingPath := msg.GetEncodingPath()
		tags["path"] = encodingPath

		// Parse the "content" field if any
		errs := c.parser.parse(grouper, content, encodingPath, gpbkv.GetDelete(), tags, timestamp)
		for _, err := range errs {
			c.acc.AddError(err)
		}
	}

	for _, groupedMetric := range grouper.Metrics() {
		c.acc.AddMetric(groupedMetric)
	}
}

func init() {
	inputs.Add("cisco_telemetry_mdt", func() telegraf.Input {
		return &CiscoTelemetryMDT{}
	})
}
