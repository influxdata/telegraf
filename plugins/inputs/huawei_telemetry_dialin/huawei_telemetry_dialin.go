package huawei_telemetry_dialin

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/influxdata/telegraf"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	huawei_dialin "github.com/influxdata/telegraf/plugins/inputs/huawei_telemetry_dialin/huawei_dialin"
	huawei_gpb "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb"
	huawei_json "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	Paths              []Path
	Aaa                aaa
	Address            string
	Encoding           string
	Sample_interval    int
	Request_id         int
	Suppress_redundant bool
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

// Start the Huawei Dialin service wt git
func (dialin *HuaweiTelemetryDialin) Start(acc telegraf.Accumulator) error {
	dialin.acc = acc
	// init parser
	parseGpb, err := huawei_gpb.New()
	parseJson, err := huawei_json.New()
	if err != nil {
		dialin.acc.AddError(fmt.Errorf("DialIn Parser Init error: %s", err))
		return err
	}
	for _, routerDialinConfig := range dialin.Routers {
		dialin.wg.Add(1)
		go dialin.singleSubscribe(routerDialinConfig, parseGpb, parseJson)
	}
	return nil
}

// create subscribe args from config of huawei_telemetry_dialin in telegraf.conf
func createSubArgs(router router) *huawei_dialin.SubsArgs {

	var paths []*huawei_dialin.Path

	if len(router.Paths) <= 0 {
		return nil
	}

	for _, path := range router.Paths {
		paths = append(paths, &huawei_dialin.Path{
			Path:  path.Path,
			Depth: uint32(path.Depth),
		})
	}

	encoding := 0
	if router.Encoding == "json" {
		encoding = 1
	}
	return &huawei_dialin.SubsArgs{
		RequestId:         uint64(router.Request_id),
		Encoding:          uint32(encoding),
		Path:              paths,
		SampleInterval:    uint64(router.Sample_interval),
		HeartbeatInterval: 60,
		Suppress:          &huawei_dialin.SubsArgs_SuppressRedundant{SuppressRedundant: router.Suppress_redundant},
	}
}

func (dialin *HuaweiTelemetryDialin) singleSubscribe(dialinConfig router, parserGpb, parserJson telegraf.Parser) {
	var opts []grpc.DialOption
	tlsConfig, errTls := dialinConfig.ClientConfig.TLSConfig()
	if errTls != nil {
		dialin.Log.Errorf("E! [single Subscribe] tlsConfig %s", errTls)
		//dialin.stop()
	} else if tlsConfig != nil {
		// tls
		tlsConfig.ServerName = dialinConfig.ServerName
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// no tls
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(dialinConfig.Address, opts...)
	if err != nil {
		dialin.Log.Errorf("E! [single Subscribe] invalid Huawei Dialin remoteServer:ng TLS PEM %s,device address : %s, request_id:%s", dialin.Transport, dialinConfig.Address, dialinConfig.Request_id)
		return
	}
	//defer conn.Close()
	client := huawei_dialin.NewGRPCConfigOperClient(conn)
	var cancel context.CancelFunc
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	if dialinConfig.Aaa.Username != "" && dialinConfig.Aaa.Password != "" {
		// input aaa as context
		//dec_password, _ := base64.URLEncoding.DecodeString(dialinConfig.Aaa.Password)
		//ctx = metadata.AppendToOutgoingContext(context.TODO(), "username", dialinConfig.Aaa.Username, "password", string(dec_password))
		ctx = metadata.AppendToOutgoingContext(context.TODO(), "username", dialinConfig.Aaa.Username, "password", dialinConfig.Aaa.Password)
	} else {
		dialin.Log.Errorf("E! [single Subscribe] no aaa configuration checked , device address : %s, request_id:%s", dialinConfig.Address, dialinConfig.Request_id)
	}
	stream, err := client.Subscribe(ctx, createSubArgs(dialinConfig))
	if err != nil {
		cancel()
		dialin.Log.Errorf("E! [single Subscribe] Huawei Dialin connection failed %s request_id %s, connection error %v", dialinConfig.Address, dialinConfig.Request_id, err)
		return
	}
	defer cancel() // 确保context被正确取消
	if stream != nil {
		for {
			packet, err := stream.Recv()
			if err != nil || packet == nil {
				dialin.Log.Errorf("E! [single Subscribe] Huawei Dialin device address %s, request_id %s, stream recv() %t", dialinConfig.Address, dialinConfig.Request_id, err)
				return
			}
			isData := checkValidData(packet)
			if isData {
				var metrics []telegraf.Metric
				var errParse error
				if len(packet.GetMessage()) != 0 {
					dialin.Log.Debugf("D! data gpb %s", hex.EncodeToString(packet.GetMessage()))
					metrics, errParse = parserGpb.Parse(packet.GetMessage())
					if errParse != nil {
						dialin.Log.Errorf("E! [Huawei Dialin] device address %s , request_id %s,Packet Parse%t", dialinConfig.Address, dialinConfig.Request_id, errParse)
						return
					}
				}
				if len(packet.GetMessageJson()) != 0 {
					dialin.Log.Debugf("D! data str %s", packet.GetMessageJson())
					metrics, errParse = parserJson.Parse([]byte(packet.GetMessageJson()))
					if errParse != nil {
						dialin.Log.Errorf("E! [Huawei Dialin] device address %s ,request_id %s, Packet Json Parse %t", dialinConfig.Address, dialinConfig.Request_id, errParse)
						return
					}
				}
				for _, metric := range metrics {
					dialin.acc.AddMetric(metric)
				}
			} else {
				dialin.Log.Errorf("the %s [request_id %s] reply with packet :\n %s ", dialinConfig.Address, dialinConfig.Request_id, packet)
			}
		}
	}
	dialin.wg.Done()
}

// Stop Client Subscribe, Close connection
func (dialin *HuaweiTelemetryDialin) Stop() {
	dialin.wg.Wait()
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
func (dialin *HuaweiTelemetryDialin) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (dialin *HuaweiTelemetryDialin) Description() string {
	return "Huawei model-driven telemetry (MDT) input plugin for dialin"
}

// Gather plugin measurements (unused)
func (dialin *HuaweiTelemetryDialin) Gather(_ telegraf.Accumulator) error {
	return nil
}

// check if the telemetry is valid
func checkValidData(reply *huawei_dialin.SubsReply) bool {
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

func (dialin *HuaweiTelemetryDialin) stop() {
	log.SetOutput(os.Stderr)
	log.Printf("I! telegraf stopped because error.")
	os.Exit(1)
}
