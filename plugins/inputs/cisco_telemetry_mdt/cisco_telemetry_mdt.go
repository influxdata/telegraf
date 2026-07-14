//go:generate ../../../tools/readme_config_includer/generator
package cisco_telemetry_mdt

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path"
	"strconv"
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

// Recursively parse tag fields
func (c *CiscoTelemetryMDT) parseKeyField(tags map[string]string, field *telemetry.TelemetryField, prefix string) {
	localname := strings.ReplaceAll(field.Name, "-", "_")
	name := localname
	if len(localname) == 0 {
		name = prefix
	} else if len(prefix) > 0 {
		name = prefix + "/" + localname
	}

	if tag := decodeTag(field); len(name) > 0 && len(tag) > 0 {
		if _, exists := tags[localname]; !exists { // Use short keys whenever possible
			tags[localname] = tag
		} else {
			tags[name] = tag
		}
	}

	for _, subfield := range field.Fields {
		c.parseKeyField(tags, subfield, name)
	}
}

func parseRib(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	// RIB
	measurement := encodingPath
	for _, subfield := range field.Fields {
		// For Every table fill the keys which are vrfName, address and masklen
		switch subfield.Name {
		case "vrfName", "address", "maskLen":
			tags[subfield.Name] = decodeTag(subfield)
		}
		if value := decode(subfield); value != nil {
			grouper.Add(measurement, tags, timestamp, subfield.Name, value)
		}
		if subfield.Name != "nextHop" {
			continue
		}
		// For next hop table fill the keys in the tag - which is address and vrfname
		for _, subf := range subfield.Fields {
			for _, ff := range subf.Fields {
				switch ff.Name {
				case "address", "vrfName":
					key := "nextHop/" + ff.Name
					tags[key] = decodeTag(ff)
				}
				if value := decode(ff); value != nil {
					name := "nextHop/" + ff.Name
					grouper.Add(measurement, tags, timestamp, name, value)
				}
			}
		}
	}
}

func parseMicroburst(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	var nxMicro *telemetry.TelemetryField
	var nxMicro1 *telemetry.TelemetryField
	// Microburst
	measurement := encodingPath
	if len(field.Fields) > 3 {
		nxMicro = field.Fields[2]
		if len(nxMicro.Fields) > 0 {
			nxMicro1 = nxMicro.Fields[0]
			if len(nxMicro1.Fields) >= 3 {
				nxMicro = nxMicro1.Fields[3]
			}
		}
	}
	for _, subfield := range nxMicro.Fields {
		if subfield.Name == "interfaceName" {
			tags[subfield.Name] = decodeTag(subfield)
		}

		for _, subf := range subfield.Fields {
			switch subf.Name {
			case "sourceName":
				newstr := strings.Split(decodeTag(subf), "-[")
				if len(newstr) <= 2 {
					tags[subf.Name] = decodeTag(subf)
				} else {
					intfName := strings.Split(newstr[1], "]")
					queue := strings.Split(newstr[2], "]")
					tags["interface_name"] = intfName[0]
					tags["queue_number"] = queue[0]
				}
			case "startTs":
				tags[subf.Name] = decodeTag(subf)
			}
			if value := decode(subf); value != nil {
				grouper.Add(measurement, tags, timestamp, subf.Name, value)
			}
		}
	}
}

func (c *CiscoTelemetryMDT) parseClassAttributeField(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	var nxAttributes *telemetry.TelemetryField
	isDme := strings.Contains(encodingPath, "sys/")
	if encodingPath == "rib" {
		// handle native data path rib
		parseRib(grouper, field, encodingPath, tags, timestamp)
		return
	}
	if encodingPath == "microburst" {
		// dump microburst
		parseMicroburst(grouper, field, encodingPath, tags, timestamp)
		return
	}
	if field == nil || !isDme || len(field.Fields) == 0 || len(field.Fields[0].Fields) == 0 || len(field.Fields[0].Fields[0].Fields) == 0 {
		return
	}

	if field.Fields[0] != nil && field.Fields[0].Fields != nil && field.Fields[0].Fields[0] != nil && field.Fields[0].Fields[0].Fields[0].Name != "attributes" {
		return
	}
	nxAttributes = field.Fields[0].Fields[0].Fields[0].Fields[0]

	// Find dn tag among list of attributes
	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "dn" {
			tags["dn"] = decodeTag(subfield)
			break
		}
	}
	// Add attributes to grouper with consistent dn tag
	for _, subfield := range nxAttributes.Fields {
		c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
	}
	// Delete dn tag to prevent it from being added to the next node's attributes
	delete(tags, "dn")
}

func (c *CiscoTelemetryMDT) getMeasurementName(encodingPath string) string {
	// Do alias lookup, to shorten measurement names
	measurement := encodingPath
	if alias, ok := c.internalAliases[encodingPath]; ok {
		measurement = alias
	} else {
		c.mutex.Lock()
		if !c.warned[encodingPath] {
			c.Log.Debugf("No measurement alias for encoding path: %s", encodingPath)
			c.warned[encodingPath] = true
		}
		c.mutex.Unlock()
	}
	return measurement
}

func (c *CiscoTelemetryMDT) parseContentField(
	grouper *metric.SeriesGrouper,
	field *telemetry.TelemetryField,
	prefix, encodingPath string,
	tags map[string]string,
	timestamp time.Time,
) {
	name := strings.ReplaceAll(field.Name, "-", "_")

	if (name == "modTs" || name == "createTs") && decode(field) == "never" {
		return
	}
	if len(name) == 0 {
		name = prefix
	} else if prefix != "" {
		name = prefix + "/" + name
	}

	extraTags := c.extraTags[strings.ReplaceAll(encodingPath, "-", "_")+"/"+name]
	if value := decode(field); value != nil {
		measurement := c.getMeasurementName(encodingPath)
		if val := c.nxosValueXform(field, encodingPath); val != nil {
			grouper.Add(measurement, tags, timestamp, name, val)
		} else {
			grouper.Add(measurement, tags, timestamp, name, value)
		}
		return
	}

	if len(extraTags) > 0 {
		for _, subfield := range field.Fields {
			if _, isExtraTag := extraTags[subfield.Name]; isExtraTag {
				tags[name+"/"+strings.ReplaceAll(subfield.Name, "-", "_")] = decodeTag(subfield)
			}
		}
	}

	var nxAttributes, nxChildren, nxRows *telemetry.TelemetryField
	isNXOS := !strings.ContainsRune(encodingPath, ':') // IOS-XR and IOS-XE have a colon in their encoding path, NX-OS does not
	isEVENT := isNXOS && strings.Contains(encodingPath, "EVENT-LIST")
	nxChildren = nil
	nxAttributes = nil
	for _, subfield := range field.Fields {
		if isNXOS && subfield.Name == "attributes" && len(subfield.Fields) > 0 {
			nxAttributes = subfield.Fields[0]
		} else if isNXOS && subfield.Name == "children" && len(subfield.Fields) > 0 {
			if !isEVENT {
				nxChildren = subfield
			} else {
				sub := subfield.Fields
				if len(sub) > 0 && sub[0] != nil && len(sub[0].Fields) >= 2 {
					if sub[0].Fields[0].Name == "subscriptionId" {
						nxAttributes = sub[0].Fields[1].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
					} else if sub[0].Fields[1].Name == "subscriptionId" {
						nxAttributes = sub[0].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
					}
				}
			}
			// if nxAttributes == NULL then class based query.
			if nxAttributes == nil {
				// call function walking over walking list.
				for _, sub := range subfield.Fields {
					c.parseClassAttributeField(grouper, sub, encodingPath, tags, timestamp)
				}
			}
		} else if isNXOS && strings.HasPrefix(subfield.Name, "ROW_") {
			nxRows = subfield
		} else if _, isExtraTag := extraTags[subfield.Name]; !isExtraTag { // Regular telemetry decoding
			c.parseContentField(grouper, subfield, name, encodingPath, tags, timestamp)
		}
	}

	if nxAttributes == nil && nxRows == nil {
		return
	}
	if nxRows != nil {
		// NXAPI structure: https://developer.cisco.com/docs/cisco-nexus-9000-series-nx-api-cli-reference-release-9-2x/
		for _, row := range nxRows.Fields {
			for i, subfield := range row.Fields {
				if i == 0 { // First subfield contains the index, promote it from value to tag
					tags[prefix] = decodeTag(subfield)
					// We can have subfield so recursively handle it.
					if len(row.Fields) == 1 {
						tags["row_number"] = strconv.FormatInt(int64(i), 10)
						c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
					}
				} else {
					c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
				}
				// Nxapi we can't identify keys always from prefix
				tags["row_number"] = strconv.FormatInt(int64(i), 10)
			}
			delete(tags, prefix)
		}
		return
	}

	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	var rn string
	var dn bool
	for _, subfield := range nxAttributes.Fields {
		switch subfield.Name {
		case "rn":
			rn = decodeTag(subfield)
		case "dn":
			dn = true
		}
	}

	if len(rn) > 0 {
		tags[prefix] = rn
	} else if !dn { // Check for distinguished name being present
		c.acc.AddError(errors.New("failed while decoding NX-OS: missing 'dn' field"))
		return
	}

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name != "rn" {
			c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
		}
	}

	if nxChildren != nil {
		// This is a nested structure, children will inherit relative name keys of parent
		for _, subfield := range nxChildren.Fields {
			c.parseContentField(grouper, subfield, prefix, encodingPath, tags, timestamp)
		}
	}
	delete(tags, prefix)
}

func init() {
	inputs.Add("cisco_telemetry_mdt", func() telegraf.Input {
		return &CiscoTelemetryMDT{}
	})
}
