package huawei_routers_telemetry

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/DamRCorba/huawei_telemetry_sensors"
	"github.com/DamRCorba/huawei_telemetry_sensors/sensors/huawei-telemetry"

	"github.com/golang/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
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
func HuaweiTelemetryDecoder(body []byte) (*metric.SeriesGrouper, error) {
	msg := &telemetry.Telemetry{}
	err := proto.Unmarshal(body[12:], msg)
	if err != nil {
		fmt.Println("Unable to decode incoming packet: ", err.Error())
		return nil, err
		//panic(err)
	}
	grouper := metric.NewSeriesGrouper()
	for _, gpbkv := range msg.GetDataGpb().GetRow() {
		dataTime := gpbkv.Timestamp
		if dataTime == 0 {
			dataTime = msg.MsgTimestamp
		}
		timestamp := time.Unix(int64(dataTime/1000), int64(dataTime%1000)*1000000)
		sensorMsg := huawei_sensorPath.GetMessageType(msg.GetSensorPath())
		err = proto.Unmarshal(gpbkv.Content, sensorMsg)
		if err != nil {
			fmt.Println("Sensor Error: %s", err.Error())	
			return nil, err		
		} 
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
	return grouper, nil
}

/*
  Listen UDP packets and call the telemetryDecoder.
*/
func (h *packetSocketListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := h.ReadFrom(buf)
		if err != nil {			
			h.Log.Error("Unable to read buffer: %s", err.Error())
			break
		}

		body, err := h.decoder.Decode(buf[:n])
		if err != nil {
			h.Log.Errorf("Unable to decode incoming packet: %s", err.Error())
			continue
		}
		// Telemetry parsing over packet payload
		grouper, err := HuaweiTelemetryDecoder(body)
		if err != nil {
			h.Log.Errorf("Unable to decode telemetry information: %s", err.Error())
			break
		}
		for _, metric := range grouper.Metrics() {
			h.AddMetric(metric)
		}

		if err != nil {
			h.Log.Errorf("Unable to parse incoming packet: %s", err.Error())
		}
	}
}

type HuaweiRoutersTelemetry struct {
	ServicePort     string        `toml:"service_port"`
	ReadBufferSize  internal.Size `toml:"read_buffer_size"`
	ContentEncoding string        `toml:"content_encoding"`
	wg              sync.WaitGroup

	Log telegraf.Logger `toml:"-"`

	telegraf.Accumulator
	io.Closer
	decoder internal.ContentDecoder
}

func (h *HuaweiRoutersTelemetry) Description() string {
	return "Huawei Telemetry UDP model input Plugin"
}

func (h *HuaweiRoutersTelemetry) SampleConfig() string {
	return `
  ## UDP Service Port to capture Telemetry
  # service_port = "8080"

`
}

func (h *HuaweiRoutersTelemetry) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *HuaweiRoutersTelemetry) Start(acc telegraf.Accumulator) error {
	h.Accumulator = acc

	var err error
	h.decoder, err = internal.NewContentDecoder(h.ContentEncoding)
	if err != nil {
		return err
	}

	pc, err := udpListen("udp", ":"+h.ServicePort)
	if err != nil {
		return err
	}

	if h.ReadBufferSize.Size > 0 {
		if srb, ok := pc.(setReadBufferer); ok {
			srb.SetReadBuffer(int(h.ReadBufferSize.Size))
		} else {
			h.Log.Warnf("Unable to set read buffer on a %s socket", "udp")
		}
	}

	h.Log.Infof("Listening on %s://%s", "udp", pc.LocalAddr())

	psl := &packetSocketListener{
		PacketConn:             pc,
		HuaweiRoutersTelemetry: h,
	}

	h.Closer = psl
	h.wg = sync.WaitGroup{}
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
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

func (h *HuaweiRoutersTelemetry) Stop() {
	if h.Closer != nil {
		h.Close()
		h.Closer = nil
	}
	h.wg.Wait()
}


func init() {
	inputs.Add("huawei_routers_telemetry", func() telegraf.Input { return &HuaweiRoutersTelemetry{} })
}
