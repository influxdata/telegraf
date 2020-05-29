package cisco_telemetry_mdt

import (
	"bytes"
	"encoding/binary"
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
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
	internaltls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials" // Register GRPC gzip decoder to support compressed telemetry
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/peer"
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
	EmbeddedTags   []string          `toml:"embedded_tags"`

	Log telegraf.Logger

	// GRPC TLS settings
	internaltls.ServerConfig

	// Internal listener / client handle
	grpcServer *grpc.Server
	listener   net.Listener

	// Internal state
	aliases   map[string]string
	warned    map[string]struct{}
	extraTags map[string]map[string]struct{}
	mutex     sync.Mutex
	acc       telegraf.Accumulator
	wg        sync.WaitGroup
}

// Start the Cisco MDT service
func (c *CiscoTelemetryMDT) Start(acc telegraf.Accumulator) error {
	var err error
	c.acc = acc
	c.listener, err = net.Listen("tcp", c.ServiceAddress)
	if err != nil {
		return err
	}

	// Invert aliases list
	c.warned = make(map[string]struct{})
	c.aliases = make(map[string]string, len(c.Aliases))
	for alias, path := range c.Aliases {
		c.aliases[path] = alias
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
			c.grpcServer.Serve(c.listener)
			c.wg.Done()
		}()

	default:
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

			conn.Close()
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
	peer, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		c.Log.Debugf("Accepted Cisco MDT GRPC dialout connection from %s", peer.Addr)
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
			chunkBuffer.Write(packet.Data)
			if chunkBuffer.Len() >= int(packet.TotalSize) {
				c.handleTelemetry(chunkBuffer.Bytes())
				chunkBuffer.Reset()
			}
		} else {
			c.acc.AddError(fmt.Errorf("dropped too large packet: %dB > %dB", packet.TotalSize, c.MaxMsgSize))
		}
	}

	if peerOK {
		c.Log.Debugf("Closed Cisco MDT GRPC dialout connection from %s", peer.Addr)
	}

	return nil
}

// Handle telemetry packet from any transport, decode and add as measurement
func (c *CiscoTelemetryMDT) handleTelemetry(data []byte) {
	msg := &telemetry.Telemetry{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		c.acc.AddError(fmt.Errorf("Cisco MDT failed to decode: %v", err))
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

		if keys == nil || content == nil {
			c.Log.Infof("Message from %s missing keys or content", msg.GetNodeIdStr())
			continue
		}

		// Parse keys
		tags = make(map[string]string, len(keys.Fields)+3)
		tags["source"] = msg.GetNodeIdStr()
		tags["subscription"] = msg.GetSubscriptionIdStr()
		tags["path"] = msg.GetEncodingPath()

		for _, subfield := range keys.Fields {
			c.parseKeyField(tags, subfield, "")
		}

		// Parse values
		for _, subfield := range content.Fields {
			c.parseContentField(grouper, subfield, "", msg.EncodingPath, tags, timestamp)
		}
	}

	for _, metric := range grouper.Metrics() {
		c.acc.AddMetric(metric)
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

func (c *CiscoTelemetryMDT) parseContentField(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField, prefix string,
	path string, tags map[string]string, timestamp time.Time) {
	name := strings.Replace(field.Name, "-", "_", -1)
	if len(name) == 0 {
		name = prefix
	} else if len(prefix) > 0 {
		name = prefix + "/" + name
	}

	extraTags := c.extraTags[strings.Replace(path, "-", "_", -1)+"/"+name]

	if value := decodeValue(field); value != nil {
		// Do alias lookup, to shorten measurement names
		measurement := path
		if alias, ok := c.aliases[path]; ok {
			measurement = alias
		} else {
			c.mutex.Lock()
			if _, haveWarned := c.warned[path]; !haveWarned {
				c.Log.Debugf("No measurement alias for encoding path: %s", path)
				c.warned[path] = struct{}{}
			}
			c.mutex.Unlock()
		}

		grouper.Add(measurement, tags, timestamp, name, value)
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
	isNXOS := !strings.ContainsRune(path, ':') // IOS-XR and IOS-XE have a colon in their encoding path, NX-OS does not
	for _, subfield := range field.Fields {
		if isNXOS && subfield.Name == "attributes" && len(subfield.Fields) > 0 {
			nxAttributes = subfield.Fields[0]
		} else if isNXOS && subfield.Name == "children" && len(subfield.Fields) > 0 {
			nxChildren = subfield
		} else if isNXOS && strings.HasPrefix(subfield.Name, "ROW_") {
			nxRows = subfield
		} else if _, isExtraTag := extraTags[subfield.Name]; !isExtraTag { // Regular telemetry decoding
			c.parseContentField(grouper, subfield, name, path, tags, timestamp)
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
				} else {
					c.parseContentField(grouper, subfield, "", path, tags, timestamp)
				}
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
			c.parseContentField(grouper, subfield, "", path, tags, timestamp)
		}
	}

	if nxChildren != nil {
		// This is a nested structure, children will inherit relative name keys of parent
		for _, subfield := range nxChildren.Fields {
			c.parseContentField(grouper, subfield, prefix, path, tags, timestamp)
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
		c.grpcServer.Stop()
	}
	if c.listener != nil {
		c.listener.Close()
	}
	c.wg.Wait()
}

const sampleConfig = `
 ## Telemetry transport can be "tcp" or "grpc".  TLS is only supported when
 ## using the grpc transport.
 transport = "grpc"

 ## Address and port to host telemetry listener
 service_address = ":57000"

 ## Enable TLS; grpc transport only.
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"

 ## Enable TLS client authentication and define allowed CA certificates; grpc
 ##  transport only.
 # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

 ## Define (for certain nested telemetry measurements with embedded tags) which fields are tags
 # embedded_tags = ["Cisco-IOS-XR-qos-ma-oper:qos/interface-table/interface/input/service-policy-names/service-policy-instance/statistics/class-stats/class-name"]

 ## Define aliases to map telemetry encoding paths to simple measurement names
 [inputs.cisco_telemetry_mdt.aliases]
   ifstats = "ietf-interfaces:interfaces-state/interface/statistics"
`

// SampleConfig of plugin
func (c *CiscoTelemetryMDT) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (c *CiscoTelemetryMDT) Description() string {
	return "Cisco model-driven telemetry (MDT) input plugin for IOS XR, IOS XE and NX-OS platforms"
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
