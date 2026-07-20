//go:generate ../../../tools/readme_config_includer/generator
package cisco_telemetry_mdt

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"path"
	"strings"
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

	// Internal listener / client handle
	grpcServer *grpc.Server
	listener   net.Listener

	// Internal state
	internalAliases map[string]string
	warned          map[string]bool
	dmesFuncs       map[string]string
	extraTags       map[string]map[string]bool
	nxpathMap       map[string]map[string]string // per path map
	propMap         map[string]func(*telemetry.TelemetryField) interface{}

	serverOptions []grpc.ServerOption

	acc   telegraf.Accumulator
	mutex sync.Mutex
	wg    sync.WaitGroup

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

	// Invert aliases list
	c.warned = make(map[string]bool)
	c.internalAliases = make(map[string]string, len(c.Aliases))
	for alias, encodingPath := range c.Aliases {
		c.internalAliases[encodingPath] = alias
	}

	// Initialize the path mappings
	c.nxpathMap = createDatabase()

	// Initialize property conversion map
	c.propMap = make(map[string]func(field *telemetry.TelemetryField) interface{}, len(c.Dmes)+4)
	c.propMap["test"] = nxosValueXformUint64Toint64
	c.propMap["asn"] = nxosValueXformUint64ToString            // uint64 to string.
	c.propMap["subscriptionId"] = nxosValueXformUint64ToString // uint64 to string.
	c.propMap["operState"] = nxosValueXformUint64ToString      // uint64 to string.

	c.dmesFuncs = make(map[string]string, len(c.Dmes))
	for dme, dmeKey := range c.Dmes {
		c.dmesFuncs[dmeKey] = dme
		switch dmeKey {
		case "uint64 to int":
			c.propMap[dme] = nxosValueXformUint64Toint64
		case "uint64 to string":
			c.propMap[dme] = nxosValueXformUint64ToString
		case "string to float64":
			c.propMap[dme] = nxosValueXformStringTofloat
		case "string to uint64":
			c.propMap[dme] = nxosValueXformStringToUint64
		case "string to int64":
			c.propMap[dme] = nxosValueXformStringToInt64
		case "auto-float-xfrom":
			c.propMap[dme] = nxosValueAutoXformFloatProp
		default:
			if !strings.HasPrefix(dme, "dnpath") {
				// Ignore non-path based property map
				continue
			}

			var payload nxPayloadXfromStructure
			if err := json.Unmarshal([]byte(dmeKey), &payload); err != nil {
				continue
			}

			// Build 2 level Hash nxpathMap Key = jsStruct.Name, Value = map of jsStruct.Prop
			// It will override the default of code if same path is provided in configuration.
			c.nxpathMap[payload.Name] = make(map[string]string, len(payload.Prop))
			for _, prop := range payload.Prop {
				c.nxpathMap[payload.Name][prop.Key] = prop.Value
			}
		}
	}

	// Fill extra tags
	c.extraTags = make(map[string]map[string]bool)
	for _, tag := range c.EmbeddedTags {
		dir := strings.ReplaceAll(path.Dir(tag), "-", "_")
		if _, found := c.extraTags[dir]; !found {
			c.extraTags[dir] = make(map[string]bool)
		}
		c.extraTags[dir][path.Base(tag)] = true
	}

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
		var tags map[string]string
		if keys != nil {
			tags = make(map[string]string, len(keys.Fields)+3)
			for _, subfield := range keys.Fields {
				c.parseKeyField(tags, subfield, "")
			}

			// If incoming MDT contains source key, copy to mdt_src
			if _, ok := tags["source"]; ok {
				tags[c.SourceFieldName] = tags["source"]
			}
		} else {
			tags = make(map[string]string, 3)
		}

		// Parse keys
		tags["source"] = msg.GetNodeIdStr()
		if msgID := msg.GetSubscriptionIdStr(); msgID != "" {
			tags["subscription"] = msgID
		}
		encodingPath := msg.GetEncodingPath()
		tags["path"] = encodingPath

		if content != nil {
			// Parse values
			for _, subfield := range content.Fields {
				prefix := ""
				switch subfield.Name {
				case "operation-metric":
					if len(subfield.Fields[0].Fields) > 0 {
						prefix = subfield.Fields[0].Fields[0].GetStringValue()
					}
				case "class-stats":
					if len(subfield.Fields[0].Fields) > 1 {
						prefix = subfield.Fields[0].Fields[1].GetStringValue()
					}
				}
				// Parse the content with and without prefix
				c.parseContentField(grouper, subfield, prefix, encodingPath, tags, timestamp)
				c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
			}
		}
		if c.IncludeDeleteField {
			grouper.Add(c.getMeasurementName(encodingPath), tags, timestamp, "delete", gpbkv.GetDelete())
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
