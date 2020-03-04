package huawei_routers_telemetry

import (
	"fmt"
	"github.com/DamRCorba/huawei_telemetry_sensors"
	"github.com/DamRCorba/huawei_telemetry_sensors/sensors/huawei-telemetry"
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type streamSocketListener struct {
	net.Listener
	*HuaweiRoutersTelemetry

	sockType string

	connections    map[string]net.Conn
	connectionsMtx sync.Mutex
}

type packetSocketListener struct {
	net.PacketConn
	*HuaweiRoutersTelemetry
}

/*
  Telemetry Decoder.

*/
func HuaweiTelemetryDecoder(body []byte) *metric.SeriesGrouper {
	msg := &telemetry.Telemetry{}
	err := proto.Unmarshal(body[12:], msg)
	if err != nil {
		fmt.Println("Error", err)
		panic(err)
	}
	grouper := metric.NewSeriesGrouper()
	for _, gpbkv := range msg.GetDataGpb().GetRow() {
		dataTime := gpbkv.Timestamp
		if dataTime == 0 {
			dataTime = msg.MsgTimestamp
		}
		timestamp := time.Unix(int64(dataTime/1000), int64(dataTime%1000)*1000000)
		sensorType := strings.Split(msg.GetSensorPath(), ":")[0]
		sensorMsg := huawei_sensorPath.GetMessageType(sensorType)
		err = proto.Unmarshal(gpbkv.Content, sensorMsg)
		if err != nil {
			fmt.Println("Sensor Error", err)
		} else {
			fields, vals := huawei_sensorPath.SearchKey(gpbkv, msg.GetSensorPath())
			tags := make(map[string]string, len(fields)+3)
			tags["source"] = msg.GetNodeIdStr()
			tags["subscription"] = msg.GetSubscriptionIdStr()
			tags["path"] = msg.GetSensorPath()
			// Search for Tags
			for i := 0; i < len(fields); i++ {
				tags = huawei_sensorPath.AppendTags(fields[i], vals[i], tags, msg.GetSensorPath())
			}
			// Create Metrics
			for i := 0; i < len(fields); i++ {
				huawei_sensorPath.CreateMetrics(grouper, tags, timestamp, msg.GetSensorPath(), fields[i], vals[i])
			}
		}
	}
	return grouper
}

/*
  Listen UDP packets and call the telemetryDecoder.
*/
func (psl *packetSocketListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.Log.Error(err.Error())
			}
			break
		}

		body, err := psl.decoder.Decode(buf[:n])
		if err != nil {
			psl.Log.Errorf("Unable to decode incoming packet: %s", err.Error())
		}
		// Inicia el manejo de la telemetria
		grouper := HuaweiTelemetryDecoder(body)
		for _, metric := range grouper.Metrics() {
			psl.AddMetric(metric)
		}

		if err != nil {
			psl.Log.Errorf("Unable to parse incoming packet: %s", err.Error())

			continue
		}
	}
}

type HuaweiRoutersTelemetry struct {
	ServicePort     string        `toml:"service_port"`
	ReadBufferSize  internal.Size `toml:"read_buffer_size"`
	ContentEncoding string        `toml:"content_encoding"`
	wg              sync.WaitGroup

	Log telegraf.Logger

	telegraf.Accumulator
	io.Closer
	decoder internal.ContentDecoder
}

func (sl *HuaweiRoutersTelemetry) Description() string {
	return "Huawei Telemetry UDP model input Plugin"
}

func (sl *HuaweiRoutersTelemetry) SampleConfig() string {
	return `
  ## UDP Service Port to capture Telemetry
  service_port = "8080"

`
}

func (sl *HuaweiRoutersTelemetry) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (hrt *HuaweiRoutersTelemetry) Start(acc telegraf.Accumulator) error {
	hrt.Accumulator = acc

	var err error
	hrt.decoder, err = internal.NewContentDecoder(hrt.ContentEncoding)
	if err != nil {
		return err
	}

	pc, err := udpListen("udp", ":"+hrt.ServicePort)
	if err != nil {
		return err
	}

	if hrt.ReadBufferSize.Size > 0 {
		if srb, ok := pc.(setReadBufferer); ok {
			srb.SetReadBuffer(int(hrt.ReadBufferSize.Size))
		} else {
			hrt.Log.Warnf("Unable to set read buffer on a %s socket", "udp")
		}
	}

	hrt.Log.Infof("Listening on %s://%s", "udp", pc.LocalAddr())

	psl := &packetSocketListener{
		PacketConn:             pc,
		HuaweiRoutersTelemetry: hrt,
	}

	hrt.Closer = psl
	hrt.wg = sync.WaitGroup{}
	hrt.wg.Add(1)
	go func() {
		defer hrt.wg.Done()
		psl.listen()
	}()
	return nil
}

func udpListen(network string, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		var addr *net.UDPAddr
		var err error
		var ifi *net.Interface
		if spl := strings.SplitN(address, "%", 2); len(spl) == 2 {
			address = spl[0]
			ifi, err = net.InterfaceByName(spl[1])
			if err != nil {
				return nil, err
			}
		}
		addr, err = net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if addr.IP.IsMulticast() {
			return net.ListenMulticastUDP(network, ifi, addr)
		}
		return net.ListenUDP(network, addr)
	}
	return net.ListenPacket(network, address)
}

func (hrt *HuaweiRoutersTelemetry) Stop() {
	if hrt.Closer != nil {
		hrt.Close()
		hrt.Closer = nil
	}
	hrt.wg.Wait()
}

func newHuaweiRoutersTelemetry() *HuaweiRoutersTelemetry {
	return &HuaweiRoutersTelemetry{}
}

type unixCloser struct {
	path   string
	closer io.Closer
}

func (uc unixCloser) Close() error {
	err := uc.closer.Close()
	os.Remove(uc.path) // ignore error
	return err
}

func init() {
	inputs.Add("huawei_routers_telemetry", func() telegraf.Input { return newHuaweiRoutersTelemetry() })
}
