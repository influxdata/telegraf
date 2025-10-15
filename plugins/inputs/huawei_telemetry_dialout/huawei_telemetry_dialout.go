package huawei_telemetry_dialout

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/influxdata/telegraf"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	dialout "github.com/influxdata/telegraf/plugins/inputs/huawei_telemetry_dialout/huawei_dialout"
	huawei_gpb "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb"
	huawei_json "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials" // Register GRPC gzip decoder to support compressed telemetry
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/peer"
)

const (
	// Maximum telemetry payload size (in bytes) to accept for GRPC dialout transport
	tcpMaxMsgLen uint32 = 0
)

// HuaweiTelemetryDialout plugin VRPs
type HuaweiTelemetryDialout struct {
	// Common configuration
	Transport      string
	ServiceAddress string            `toml:"service_address"`
	MaxMsgSize     int               `toml:"max_msg_size"`
	Aliases        map[string]string `toml:"aliases"`
	EmbeddedTags   []string          `toml:"embedded_tags"`
	Log            telegraf.Logger
	// GRPC TLS settings
	internaltls.ServerConfig

	// Internal listener / client handle
	grpcServer *grpc.Server
	listener   net.Listener

	// Internal state
	aliases   map[string]string
	warned    map[string]struct{}
	extraTags map[string]map[string]struct{}
	metrics   []telegraf.Metric
	mutex     sync.Mutex
	acc       telegraf.Accumulator
	wg        sync.WaitGroup
}

// Start the Huawei Telemetry dialout service
func (c *HuaweiTelemetryDialout) Start(acc telegraf.Accumulator) error {
	var err error
	c.acc = acc
	c.listener, err = net.Listen("tcp", c.ServiceAddress)
	if err != nil {
		return err
	}
	switch c.Transport {
	case "grpc":
		// set tls and max size of packet
		var opts []grpc.ServerOption
		tlsConfig, err := c.ServerConfig.TLSConfig()

		if err != nil {
			c.listener.Close()
			return err
		} else if tlsConfig != nil {
			tlsConfig.ClientAuth = tls.RequestClientCert
			opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		}

		if c.MaxMsgSize > 0 {
			opts = append(opts, grpc.MaxRecvMsgSize(c.MaxMsgSize))
		}
		// create grpc server
		c.grpcServer = grpc.NewServer(opts...)
		// Register the server with the method that receives the data
		dialout.RegisterGRPCDataserviceServer(c.grpcServer, c)

		c.wg.Add(1)
		go func() {
			// Listen on the ServiceAddress port
			c.grpcServer.Serve(c.listener)
			c.wg.Done()
		}()

	default:
		c.listener.Close()
		return fmt.Errorf("invalid Huawei transport: %s", c.Transport)
	}

	return nil
}

// AcceptTCPDialoutClients defines the TCP dialout server main routine
func (c *HuaweiTelemetryDialout) acceptTCPClientsddd() {
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
			c.Log.Debugf("D! Accepted Huawei MDT TCP dialout connection from %s", conn.RemoteAddr())
			if err := c.handleTCPClient(conn); err != nil {
				c.acc.AddError(err)
			}
			c.Log.Debugf("Closed Huawei MDT TCP dialout connection from %s", conn.RemoteAddr())

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
func (c *HuaweiTelemetryDialout) handleTCPClient(conn net.Conn) error {
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
	}
}

// implement the rpc method of huawei-grpc-dialout.proto
func (c *HuaweiTelemetryDialout) DataPublish(stream dialout.GRPCDataservice_DataPublishServer) error {
	peer, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		c.Log.Debugf("Accepted Huawei GRPC dialout connection from %s", peer.Addr)
	}
	// init parser
	parseGpb, err := huawei_gpb.New()
	parseJson, err := huawei_json.New()
	if err != nil {
		c.acc.AddError(fmt.Errorf("Dialout Parser Init error: %s, %v", err))
		return err
	}
	//var chunkBuffer bytes.Buffer
	for {
		packet, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				c.acc.AddError(fmt.Errorf("GRPC dialout receive error: %s, %v", c.listener.Addr(), err))
			}
			break
		}

		if len(packet.Errors) != 0 {
			c.acc.AddError(fmt.Errorf("GRPC dialout error: %s", packet.Errors))
			break
		}

		var metrics []telegraf.Metric
		var errParse error
		// gpb encoding
		if len(packet.GetData()) != 0 {
			c.Log.Debugf("D! data gpb %s", hex.EncodeToString(packet.GetData()))
			metrics, errParse = parseGpb.Parse(packet.GetData())
			if errParse != nil {
				c.acc.AddError(errParse)
				c.stop()
				return fmt.Errorf("[input.huawei_telemetry_dialout] error when parse grpc stream %t", errParse)
			}
		}
		// json encoding

		if len(packet.GetDataJson()) != 0 {
			c.Log.Debugf("D! data str %s", packet.GetDataJson())
			metrics, errParse = parseJson.Parse([]byte(packet.GetDataJson()))
			if errParse != nil {
				c.acc.AddError(errParse)
				c.stop()
				return fmt.Errorf("[input.huawei_telemetry_dialout] error when parse grpc stream %t", errParse)
			}
		}
		for _, metric := range metrics {
			c.acc.AddMetric(metric)
		}
	}
	if peerOK {
		c.Log.Debugf("D! Closed Huawei GRPC dialout connection from %s", peer.Addr)
	}
	return nil
}

func (c *HuaweiTelemetryDialout) Address() net.Addr {
	return c.listener.Addr()
}

// Stop listener and cleanup
func (c *HuaweiTelemetryDialout) Stop() {
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
 ## Address and port to host telemetry listener
 service_address = "10.0.0.1:57000"

 ## Enable TLS; 
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"

 ## Enable TLS client authentication and define allowed CA certificates; grpc
 ##  transport only.
 # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

 ## Define (for certain nested telemetry measurements with embedded tags) which fields are tags

 ## Define aliases to map telemetry encoding paths to simple measurement names
`

// SampleConfig of plugin
func (c *HuaweiTelemetryDialout) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (c *HuaweiTelemetryDialout) Description() string {
	return "Huawei Telemetry For Huawei Router"
}

// Gather plugin measurements (unused)
func (c *HuaweiTelemetryDialout) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("huawei_telemetry_dialout", func() telegraf.Input {
		return &HuaweiTelemetryDialout{}
	})
}

func (c *HuaweiTelemetryDialout) stop() {
	log.SetOutput(os.Stderr)
	log.Printf("I! telegraf stopped because error.")
	os.Exit(1)
}
