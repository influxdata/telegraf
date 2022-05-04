package cisco_telemetry_mdt

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	dialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // Register GRPC gzip decoder to support compressed telemetry
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// Maximum telemetry payload size (in bytes) to accept for GRPC dialout transport
	tcpMaxMsgLen uint32 = 1024 * 1024
)

// CiscoTelemetryMDT plugin for IOS XR, IOS XE and NXOS platforms
type CiscoTelemetryMDT struct {
	// Common configuration
	Transport      string
	ServiceAddress string            `toml:"service_address"`
	MaxMsgSize     int               `toml:"max_msg_size"`
	Aliases        map[string]string `toml:"aliases"`
	Dmes           map[string]string `toml:"dmes"`
	EmbeddedTags   []string          `toml:"embedded_tags"`

	Log telegraf.Logger

	// GRPC TLS settings
	internaltls.ServerConfig

	// Internal listener / client handle
	grpcServer *grpc.Server
	listener   net.Listener

	// Internal state
	internalAliases map[string]string
	dmesFuncs       map[string]string
	warned          map[string]struct{}
	extraTags       map[string]map[string]struct{}
	nxpathMap       map[string]map[string]string //per path map
	propMap         map[string]func(field *telemetry.TelemetryField, value interface{}) interface{}
	mutex           sync.Mutex
	acc             telegraf.Accumulator
	wg              sync.WaitGroup

	// Though unused in the code, required by protoc-gen-go-grpc to maintain compatibility
	dialout.UnimplementedGRPCMdtDialoutServer
}

type NxPayloadXfromStructure struct {
	Name string `json:"Name"`
	Prop []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"prop"`
}

// Start the Cisco MDT service
func (c *CiscoTelemetryMDT) Start(acc telegraf.Accumulator) error {
	var err error
	c.acc = acc
	c.listener, err = net.Listen("tcp", c.ServiceAddress)
	if err != nil {
		return err
	}

	c.propMap = make(map[string]func(field *telemetry.TelemetryField, value interface{}) interface{}, 100)
	c.propMap["test"] = nxosValueXformUint64Toint64
	c.propMap["asn"] = nxosValueXformUint64ToString            //uint64 to string.
	c.propMap["subscriptionId"] = nxosValueXformUint64ToString //uint64 to string.
	c.propMap["operState"] = nxosValueXformUint64ToString      //uint64 to string.

	// Invert aliases list
	c.warned = make(map[string]struct{})
	c.internalAliases = make(map[string]string, len(c.Aliases))
	for alias, encodingPath := range c.Aliases {
		c.internalAliases[encodingPath] = alias
	}
	c.initDb()

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
			if !strings.HasPrefix(dme, "dnpath") { // not path based property map
				continue
			}

			var jsStruct NxPayloadXfromStructure
			err := json.Unmarshal([]byte(dmeKey), &jsStruct)
			if err != nil {
				continue
			}

			// Build 2 level Hash nxpathMap Key = jsStruct.Name, Value = map of jsStruct.Prop
			// It will override the default of code if same path is provided in configuration.
			c.nxpathMap[jsStruct.Name] = make(map[string]string, len(jsStruct.Prop))
			for _, prop := range jsStruct.Prop {
				c.nxpathMap[jsStruct.Name][prop.Key] = prop.Value
			}
		}
	}

	// Fill extra tags
	c.extraTags = make(map[string]map[string]struct{})
	for _, tag := range c.EmbeddedTags {
		dir := strings.Replace(path.Dir(tag), "-", "_", -1)
		if _, hasKey := c.extraTags[dir]; !hasKey {
			c.extraTags[dir] = make(map[string]struct{})
		}
		c.extraTags[dir][path.Base(tag)] = struct{}{}
	}

	switch c.Transport {
	case "tcp":
		// TCP dialout server accept routine
		c.wg.Add(1)
		go func() {
			c.acceptTCPClients()
			c.wg.Done()
		}()

	case "grpc":
		var opts []grpc.ServerOption
		tlsConfig, err := c.ServerConfig.TLSConfig()
		if err != nil {
			//nolint:errcheck,revive // we cannot do anything if the closing fails
			c.listener.Close()
			return err
		} else if tlsConfig != nil {
			opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		}

		if c.MaxMsgSize > 0 {
			opts = append(opts, grpc.MaxRecvMsgSize(c.MaxMsgSize))
		}

		c.grpcServer = grpc.NewServer(opts...)
		dialout.RegisterGRPCMdtDialoutServer(c.grpcServer, c)

		c.wg.Add(1)
		go func() {
			if err := c.grpcServer.Serve(c.listener); err != nil {
				c.Log.Errorf("serving GRPC server failed: %v", err)
			}
			c.wg.Done()
		}()

	default:
		//nolint:errcheck,revive // we cannot do anything if the closing fails
		c.listener.Close()
		return fmt.Errorf("invalid Cisco MDT transport: %s", c.Transport)
	}

	return nil
}

// AcceptTCPDialoutClients defines the TCP dialout server main routine
func (c *CiscoTelemetryMDT) acceptTCPClients() {
	// Keep track of all active connections, so we can close them if necessary
	var mutex sync.Mutex
	clients := make(map[net.Conn]struct{})

	for {
		conn, err := c.listener.Accept()
		if neterr, ok := err.(*net.OpError); ok && (neterr.Timeout() || neterr.Temporary()) {
			continue
		} else if err != nil {
			break // Stop() will close the connection so Accept() will fail here
		}

		mutex.Lock()
		clients[conn] = struct{}{}
		mutex.Unlock()

		// Individual client connection routine
		c.wg.Add(1)
		go func() {
			c.Log.Debugf("Accepted Cisco MDT TCP dialout connection from %s", conn.RemoteAddr())
			if err := c.handleTCPClient(conn); err != nil {
				c.acc.AddError(err)
			}
			c.Log.Debugf("Closed Cisco MDT TCP dialout connection from %s", conn.RemoteAddr())

			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()

			if err := conn.Close(); err != nil {
				c.Log.Warnf("closing connection failed: %v", err)
			}
			c.wg.Done()
		}()
	}

	// Close all remaining client connections
	mutex.Lock()
	for client := range clients {
		if err := client.Close(); err != nil {
			c.Log.Errorf("Failed to close TCP dialout client: %v", err)
		}
	}
	mutex.Unlock()
}

// Handle a TCP telemetry client
func (c *CiscoTelemetryMDT) handleTCPClient(conn net.Conn) error {
	// TCP Dialout telemetry framing header
	var hdr struct {
		MsgType       uint16
		MsgEncap      uint16
		MsgHdrVersion uint16
		MsgFlags      uint16
		MsgLen        uint32
	}

	var payload bytes.Buffer

	for {
		// Read and validate dialout telemetry header
		if err := binary.Read(conn, binary.BigEndian, &hdr); err != nil {
			return err
		}

		maxMsgSize := tcpMaxMsgLen
		if c.MaxMsgSize > 0 {
			maxMsgSize = uint32(c.MaxMsgSize)
		}

		if hdr.MsgLen > maxMsgSize {
			return fmt.Errorf("dialout packet too long: %v", hdr.MsgLen)
		} else if hdr.MsgFlags != 0 {
			return fmt.Errorf("invalid dialout flags: %v", hdr.MsgFlags)
		}

		// Read and handle telemetry packet
		payload.Reset()
		if size, err := payload.ReadFrom(io.LimitReader(conn, int64(hdr.MsgLen))); size != int64(hdr.MsgLen) {
			if err != nil {
				return err
			}
			return fmt.Errorf("TCP dialout premature EOF")
		}

		c.handleTelemetry(payload.Bytes())
	}
}

// MdtDialout RPC server method for grpc-dialout transport
func (c *CiscoTelemetryMDT) MdtDialout(stream dialout.GRPCMdtDialout_MdtDialoutServer) error {
	peerInCtx, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		c.Log.Debugf("Accepted Cisco MDT GRPC dialout connection from %s", peerInCtx.Addr)
	}

	var chunkBuffer bytes.Buffer

	for {
		packet, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				c.acc.AddError(fmt.Errorf("GRPC dialout receive error: %v", err))
			}
			break
		}

		if len(packet.Data) == 0 && len(packet.Errors) != 0 {
			c.acc.AddError(fmt.Errorf("GRPC dialout error: %s", packet.Errors))
			break
		}

		// Reassemble chunked telemetry data received from NX-OS
		if packet.TotalSize == 0 {
			c.handleTelemetry(packet.Data)
		} else if int(packet.TotalSize) <= c.MaxMsgSize {
			if _, err := chunkBuffer.Write(packet.Data); err != nil {
				c.acc.AddError(fmt.Errorf("writing packet %q failed: %v", packet.Data, err))
			}
			if chunkBuffer.Len() >= int(packet.TotalSize) {
				c.handleTelemetry(chunkBuffer.Bytes())
				chunkBuffer.Reset()
			}
		} else {
			c.acc.AddError(fmt.Errorf("dropped too large packet: %dB > %dB", packet.TotalSize, c.MaxMsgSize))
		}
	}

	if peerOK {
		c.Log.Debugf("Closed Cisco MDT GRPC dialout connection from %s", peerInCtx.Addr)
	}

	return nil
}

// Handle telemetry packet from any transport, decode and add as measurement
func (c *CiscoTelemetryMDT) handleTelemetry(data []byte) {
	msg := &telemetry.Telemetry{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		c.acc.AddError(fmt.Errorf("failed to decode: %v", err))
		return
	}

	grouper := metric.NewSeriesGrouper()
	for _, gpbkv := range msg.DataGpbkv {
		// Produce metadata tags
		var tags map[string]string

		// Top-level field may have measurement timestamp, if not use message timestamp
		measured := gpbkv.Timestamp
		if measured == 0 {
			measured = msg.MsgTimestamp
		}

		timestamp := time.Unix(int64(measured/1000), int64(measured%1000)*1000000)

		// Find toplevel GPBKV fields "keys" and "content"
		var keys, content *telemetry.TelemetryField = nil, nil
		for _, field := range gpbkv.Fields {
			if field.Name == "keys" {
				keys = field
			} else if field.Name == "content" {
				content = field
			}
		}

		// if the keys and content fields are missing, skip the message as it
		// does not have parsable data used by Telegraf
		if keys == nil || content == nil {
			continue
		}

		// Parse keys
		tags = make(map[string]string, len(keys.Fields)+3)
		tags["source"] = msg.GetNodeIdStr()
		if msgID := msg.GetSubscriptionIdStr(); msgID != "" {
			tags["subscription"] = msgID
		}
		tags["path"] = msg.GetEncodingPath()

		for _, subfield := range keys.Fields {
			c.parseKeyField(tags, subfield, "")
		}

		// Parse values
		for _, subfield := range content.Fields {
			c.parseContentField(grouper, subfield, "", msg.EncodingPath, tags, timestamp)
		}
	}

	for _, groupedMetric := range grouper.Metrics() {
		c.acc.AddMetric(groupedMetric)
	}
}

func decodeValue(field *telemetry.TelemetryField) interface{} {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_BytesValue:
		return val.BytesValue
	case *telemetry.TelemetryField_StringValue:
		if len(val.StringValue) > 0 {
			return val.StringValue
		}
	case *telemetry.TelemetryField_BoolValue:
		return val.BoolValue
	case *telemetry.TelemetryField_Uint32Value:
		return val.Uint32Value
	case *telemetry.TelemetryField_Uint64Value:
		return val.Uint64Value
	case *telemetry.TelemetryField_Sint32Value:
		return val.Sint32Value
	case *telemetry.TelemetryField_Sint64Value:
		return val.Sint64Value
	case *telemetry.TelemetryField_DoubleValue:
		return val.DoubleValue
	case *telemetry.TelemetryField_FloatValue:
		return val.FloatValue
	}
	return nil
}

func decodeTag(field *telemetry.TelemetryField) string {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_BytesValue:
		return string(val.BytesValue)
	case *telemetry.TelemetryField_StringValue:
		return val.StringValue
	case *telemetry.TelemetryField_BoolValue:
		if val.BoolValue {
			return "true"
		}
		return "false"
	case *telemetry.TelemetryField_Uint32Value:
		return strconv.FormatUint(uint64(val.Uint32Value), 10)
	case *telemetry.TelemetryField_Uint64Value:
		return strconv.FormatUint(val.Uint64Value, 10)
	case *telemetry.TelemetryField_Sint32Value:
		return strconv.FormatInt(int64(val.Sint32Value), 10)
	case *telemetry.TelemetryField_Sint64Value:
		return strconv.FormatInt(val.Sint64Value, 10)
	case *telemetry.TelemetryField_DoubleValue:
		return strconv.FormatFloat(val.DoubleValue, 'f', -1, 64)
	case *telemetry.TelemetryField_FloatValue:
		return strconv.FormatFloat(float64(val.FloatValue), 'f', -1, 32)
	default:
		return ""
	}
}

// Recursively parse tag fields
func (c *CiscoTelemetryMDT) parseKeyField(tags map[string]string, field *telemetry.TelemetryField, prefix string) {
	localname := strings.Replace(field.Name, "-", "_", -1)
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

func (c *CiscoTelemetryMDT) parseRib(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField,
	encodingPath string, tags map[string]string, timestamp time.Time) {
	// RIB
	measurement := encodingPath
	for _, subfield := range field.Fields {
		//For Every table fill the keys which are vrfName, address and masklen
		switch subfield.Name {
		case "vrfName", "address", "maskLen":
			tags[subfield.Name] = decodeTag(subfield)
		}
		if value := decodeValue(subfield); value != nil {
			if err := grouper.Add(measurement, tags, timestamp, subfield.Name, value); err != nil {
				c.Log.Errorf("adding field %q to group failed: %v", subfield.Name, err)
			}
		}
		if subfield.Name != "nextHop" {
			continue
		}
		//For next hop table fill the keys in the tag - which is address and vrfname
		for _, subf := range subfield.Fields {
			for _, ff := range subf.Fields {
				switch ff.Name {
				case "address", "vrfName":
					key := "nextHop/" + ff.Name
					tags[key] = decodeTag(ff)
				}
				if value := decodeValue(ff); value != nil {
					name := "nextHop/" + ff.Name
					if err := grouper.Add(measurement, tags, timestamp, name, value); err != nil {
						c.Log.Errorf("adding field %q to group failed: %v", name, err)
					}
				}
			}
		}
	}
}

func (c *CiscoTelemetryMDT) parseClassAttributeField(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField,
	encodingPath string, tags map[string]string, timestamp time.Time) {
	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	var nxAttributes *telemetry.TelemetryField
	isDme := strings.Contains(encodingPath, "sys/")
	if encodingPath == "rib" {
		//handle native data path rib
		c.parseRib(grouper, field, encodingPath, tags, timestamp)
		return
	}
	if field == nil || !isDme || len(field.Fields) == 0 || len(field.Fields[0].Fields) == 0 || len(field.Fields[0].Fields[0].Fields) == 0 {
		return
	}

	if field.Fields[0] != nil && field.Fields[0].Fields != nil && field.Fields[0].Fields[0] != nil && field.Fields[0].Fields[0].Fields[0].Name != "attributes" {
		return
	}
	nxAttributes = field.Fields[0].Fields[0].Fields[0].Fields[0]

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "dn" {
			tags["dn"] = decodeTag(subfield)
		} else {
			c.parseContentField(grouper, subfield, "", encodingPath, tags, timestamp)
		}
	}
}

func (c *CiscoTelemetryMDT) parseContentField(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField, prefix string,
	encodingPath string, tags map[string]string, timestamp time.Time) {
	name := strings.Replace(field.Name, "-", "_", -1)

	if (name == "modTs" || name == "createTs") && decodeValue(field) == "never" {
		return
	}
	if len(name) == 0 {
		name = prefix
	} else if len(prefix) > 0 {
		name = prefix + "/" + name
	}

	extraTags := c.extraTags[strings.Replace(encodingPath, "-", "_", -1)+"/"+name]

	if value := decodeValue(field); value != nil {
		// Do alias lookup, to shorten measurement names
		measurement := encodingPath
		if alias, ok := c.internalAliases[encodingPath]; ok {
			measurement = alias
		} else {
			c.mutex.Lock()
			if _, haveWarned := c.warned[encodingPath]; !haveWarned {
				c.Log.Debugf("No measurement alias for encoding path: %s", encodingPath)
				c.warned[encodingPath] = struct{}{}
			}
			c.mutex.Unlock()
		}

		if val := c.nxosValueXform(field, value, encodingPath); val != nil {
			if err := grouper.Add(measurement, tags, timestamp, name, val); err != nil {
				c.Log.Errorf("adding field %q to group failed: %v", name, err)
			}
		} else {
			if err := grouper.Add(measurement, tags, timestamp, name, value); err != nil {
				c.Log.Errorf("adding field %q to group failed: %v", name, err)
			}
		}
		return
	}

	if len(extraTags) > 0 {
		for _, subfield := range field.Fields {
			if _, isExtraTag := extraTags[subfield.Name]; isExtraTag {
				tags[name+"/"+strings.Replace(subfield.Name, "-", "_", -1)] = decodeTag(subfield)
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
				if len(sub) > 0 && sub[0] != nil && sub[0].Fields[0].Name == "subscriptionId" && len(sub[0].Fields) >= 2 {
					nxAttributes = sub[0].Fields[1].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
				}
			}
			//if nxAttributes == NULL then class based query.
			if nxAttributes == nil {
				//call function walking over walking list.
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
	} else if nxRows != nil {
		// NXAPI structure: https://developer.cisco.com/docs/cisco-nexus-9000-series-nx-api-cli-reference-release-9-2x/
		for _, row := range nxRows.Fields {
			for i, subfield := range row.Fields {
				if i == 0 { // First subfield contains the index, promote it from value to tag
					tags[prefix] = decodeTag(subfield)
					//We can have subfield so recursively handle it.
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
	rn := ""
	dn := false

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "rn" {
			rn = decodeTag(subfield)
		} else if subfield.Name == "dn" {
			dn = true
		}
	}

	if len(rn) > 0 {
		tags[prefix] = rn
	} else if !dn { // Check for distinguished name being present
		c.acc.AddError(fmt.Errorf("NX-OS decoding failed: missing dn field"))
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

func (c *CiscoTelemetryMDT) Address() net.Addr {
	return c.listener.Addr()
}

// Stop listener and cleanup
func (c *CiscoTelemetryMDT) Stop() {
	if c.grpcServer != nil {
		// Stop server and terminate all running dialout routines
		//nolint:errcheck,revive // we cannot do anything if the stopping fails
		c.grpcServer.Stop()
	}
	if c.listener != nil {
		//nolint:errcheck,revive // we cannot do anything if the closing fails
		c.listener.Close()
	}
	c.wg.Wait()
}

// Gather plugin measurements (unused)
func (c *CiscoTelemetryMDT) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("cisco_telemetry_mdt", func() telegraf.Input {
		return &CiscoTelemetryMDT{
			Transport:      "grpc",
			ServiceAddress: "127.0.0.1:57000",
		}
	})
}
