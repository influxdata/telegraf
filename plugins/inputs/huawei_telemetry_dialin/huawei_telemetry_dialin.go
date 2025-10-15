package huawei_telemetry_dialin

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	dialin "github.com/influxdata/telegraf/plugins/inputs/huawei_telemetry_dialin/huawei_dialin"
	huawei_gpb "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb"
	huawei_json "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Path struct {
	Depth int
	Path  string
}

type aaa struct {
	Password string
	Username string
}

type router struct {
	Paths             []Path
	Aaa               aaa
	Address           string
	Encoding          string
	SampleInterval    int
	RequestID         int
	SuppressRedundant bool
	// GRPC TLS settings
	internaltls.ClientConfig
}

// HuaweiTelemetryDialin plugin VRPs
type HuaweiTelemetryDialin struct {
	// Common configuration
	Routers      []router `toml:"routers"`
	Transport    string
	MaxMsgSize   int               `toml:"max_msg_size"`
	Aliases      map[string]string `toml:"aliases"`
	EmbeddedTags []string          `toml:"embedded_tags"`
	Log          telegraf.Logger

	// Internal listener / client handle
	// grpcServer *grpc.Server
	// listener   net.Listener

	// Internal state
	// aliases   map[string]string
	// warned    map[string]struct{}
	// extraTags map[string]map[string]struct{}
	// mutex     sync.Mutex
	acc telegraf.Accumulator
	wg  sync.WaitGroup
}

// Start the Huawei Dialin service
func (d *HuaweiTelemetryDialin) Start(acc telegraf.Accumulator) error {
	d.acc = acc
	// init parser
	parseGPB, err := huawei_gpb.New()
	if err != nil {
		d.acc.AddError(fmt.Errorf("dialin parser init error: %w", err))
		return err
	}
	parseJSON, err := huawei_json.New()
	if err != nil {
		d.acc.AddError(fmt.Errorf("dialin parser init error: %w", err))
		return err
	}
	for _, routerDialinConfig := range d.Routers {
		d.wg.Add(1)
		go d.singleSubscribe(routerDialinConfig, parseGPB, parseJSON)
	}
	return nil
}

// create subscribe args from config of huawei_telemetry_dialin in telegraf.conf
func createSubArgs(router router) *dialin.SubsArgs {
	paths := make([]*dialin.Path, 0, len(router.Paths))

	if len(router.Paths) == 0 {
		return nil
	}

	for _, path := range router.Paths {
		paths = append(paths, &dialin.Path{
			Path:  path.Path,
			Depth: uint32(path.Depth),
		})
	}

	encoding := 0
	if router.Encoding == "json" {
		encoding = 1
	}
	return &dialin.SubsArgs{
		RequestId:         uint64(router.RequestID),
		Encoding:          uint32(encoding),
		Path:              paths,
		SampleInterval:    uint64(router.SampleInterval),
		HeartbeatInterval: 60,
		Suppress:          &dialin.SubsArgs_SuppressRedundant{SuppressRedundant: router.SuppressRedundant},
	}
}

func (d *HuaweiTelemetryDialin) singleSubscribe(dialinConfig router, parserGPB, parserJSON telegraf.Parser) {
	var opts []grpc.DialOption
	tlsConfig, errTLS := dialinConfig.ClientConfig.TLSConfig()
	if errTLS != nil {
		d.Log.Errorf("[single subscribe] tlsConfig error: %s", errTLS)
		// d.stop()
	} else if tlsConfig != nil {
		// TLS
		tlsConfig.ServerName = dialinConfig.ServerName
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// No TLS
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(dialinConfig.Address, opts...)
	if err != nil {
		d.Log.Errorf("[single subscribe] invalid Huawei Dialin remoteServer: device address: %s, request_id: %d", dialinConfig.Address, dialinConfig.RequestID)
		return
	}
	defer conn.Close()
	client := dialin.NewGRPCConfigOperClient(conn)
	var cancel context.CancelFunc
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	if dialinConfig.Aaa.Username != "" && dialinConfig.Aaa.Password != "" {
		// Input AAA as context
		ctx = metadata.AppendToOutgoingContext(context.TODO(), "username", dialinConfig.Aaa.Username, "password", dialinConfig.Aaa.Password)
	} else {
		d.Log.Errorf("[single subscribe] no AAA configuration checked, device address: %s, request_id: %d",
			dialinConfig.Address, dialinConfig.RequestID)
	}
	stream, err := client.Subscribe(ctx, createSubArgs(dialinConfig))
	if err != nil {
		cancel()
		d.Log.Errorf("[single subscribe] Huawei Dialin connection failed %s request_id %d, connection error %v",
			dialinConfig.Address, dialinConfig.RequestID, err)
		return
	}
	defer cancel() // 确保context被正确取消
	if stream != nil {
		for {
			packet, err := stream.Recv()
			if err != nil || packet == nil {
				d.Log.Errorf("[single subscribe] Huawei Dialin device address %s, request_id %d, stream recv() %v",
					dialinConfig.Address, dialinConfig.RequestID, err)
				return
			}
			isData := checkValidData(packet)
			if isData {
				var metrics []telegraf.Metric
				var errParse error
				if len(packet.GetMessage()) != 0 {
					d.Log.Debugf("data gpb %s", hex.EncodeToString(packet.GetMessage()))
					metrics, errParse = parserGPB.Parse(packet.GetMessage())
					if errParse != nil {
						d.Log.Errorf("[huawei dialin] device address %s, request_id %d, packet parse error: %v",
							dialinConfig.Address, dialinConfig.RequestID, errParse)
						return
					}
				}
				if len(packet.GetMessageJson()) != 0 {
					d.Log.Debugf("data str %s", packet.GetMessageJson())
					metrics, errParse = parserJSON.Parse([]byte(packet.GetMessageJson()))
					if errParse != nil {
						d.Log.Errorf("[huawei dialin] device address %s, request_id %d, packet JSON parse error: %v",
							dialinConfig.Address, dialinConfig.RequestID, errParse)
						return
					}
				}
				for _, metric := range metrics {
					d.acc.AddMetric(metric)
				}
			} else {
				d.Log.Errorf("device %s [request_id %d] reply with packet: %v", dialinConfig.Address, dialinConfig.RequestID, packet)
			}
		}
	}
	d.wg.Done()
}

// Stop Client Subscribe, Close connection
func (d *HuaweiTelemetryDialin) Stop() {
	d.wg.Wait()
}

const sampleConfig = `
## Address and port to host telemetry listener
 service_address = "10.0.0.1:57000"

 ## Enable TLS; grpc transport only.
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"

 ## Enable TLS client authentication and define allowed CA certificates; grpc
 ##  transport only.
 # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

 ## Define (for certain nested telemetry measurements with embedded tags) which fields are tags

 ## Define aliases to map telemetry encoding paths to simple measurement names
 `

// SampleConfig of plugin
func (*HuaweiTelemetryDialin) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (*HuaweiTelemetryDialin) Description() string {
	return "Huawei model-driven telemetry (MDT) input plugin for dialin"
}

// Gather plugin measurements (unused)
func (*HuaweiTelemetryDialin) Gather(_ telegraf.Accumulator) error {
	return nil
}

// check if the telemetry is valid
func checkValidData(reply *dialin.SubsReply) bool {
	respCode := reply.ResponseCode
	if respCode == "" {
		return true
	}
	// return error
	if respCode != "200" && respCode != "" {
		return false
	}
	// respCode is "ok" doesn't mean this message is a telemetry data message ,just an answer message
	if respCode == "ok" {
		return false
	}
	return false
}

func init() {
	inputs.Add("huawei_telemetry_dialin", func() telegraf.Input {
		return &HuaweiTelemetryDialin{}
	})
}
