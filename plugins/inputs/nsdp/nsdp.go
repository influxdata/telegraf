//go:generate ../../../tools/readme_config_includer/generator
package nsdp

import (
	_ "embed"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tdrn-org/go-nsdp"
)

//go:embed sample.conf
var sampleConfig string

type NSDP struct {
	Address     string          `toml:"address"`
	DeviceLimit uint            `toml:"device_limit"`
	Timeout     config.Duration `toml:"timeout"`

	Log telegraf.Logger `toml:"-"`

	conn *nsdp.Conn
}

func (*NSDP) SampleConfig() string {
	return sampleConfig
}

func (n *NSDP) Init() error {
	if n.Address == "" {
		n.Address = nsdp.IPv4BroadcastTarget
	}
	if n.Timeout == 0 {
		n.Timeout = config.Duration(2 * time.Second)
	}
	return nil
}

func (n *NSDP) Gather(acc telegraf.Accumulator) error {
	if n.conn == nil {
		conn, err := nsdp.NewConn(n.Address, n.Log.Level().Includes(telegraf.Trace))
		if err != nil {
			return fmt.Errorf("failed to create connection to address %s: %s", n.Address, err)
		}
		conn.ReceiveDeviceLimit = n.DeviceLimit
		conn.ReceiveTimeout = time.Duration(n.Timeout)
		n.conn = conn
	}
	responses, err := n.conn.SendReceiveMessage(n.newGatherRequest())
	if err != nil {
		// Close malfunctioning connection and re-connect on next Gather call
		n.conn.Close()
		n.conn = nil
		return fmt.Errorf("failed to query address %s: %w", n.Address, err)
	}
	for device, response := range responses {
		n.Log.Tracef("Processing device: %s", device)
		n.gatherDevice(acc, device, response)
	}
	return nil
}

func (n *NSDP) newGatherRequest() *nsdp.Message {
	request := nsdp.NewMessage(nsdp.ReadRequest)
	request.AppendTLV(nsdp.EmptyDeviceModel())
	request.AppendTLV(nsdp.EmptyDeviceName())
	request.AppendTLV(nsdp.EmptyDeviceIP())
	request.AppendTLV(nsdp.EmptyPortStatistic())
	return request
}

func (n *NSDP) gatherDevice(acc telegraf.Accumulator, device string, response *nsdp.Message) {
	var deviceModel string
	var deviceName string
	var deviceIP net.IP
	portStats := make(map[uint8]*nsdp.PortStatistic, 0)
	for _, tlv := range response.Body {
		switch tlv.Type() {
		case nsdp.TypeDeviceModel:
			deviceModel = tlv.(*nsdp.DeviceModel).Model
		case nsdp.TypeDeviceName:
			deviceName = tlv.(*nsdp.DeviceName).Name
		case nsdp.TypeDeviceIP:
			deviceIP = tlv.(*nsdp.DeviceIP).IP
		case nsdp.TypePortStatistic:
			portStat := tlv.(*nsdp.PortStatistic)
			portStats[portStat.Port] = portStat
		}
	}
	for port, stat := range portStats {
		tags := map[string]string{
			"device":       device,
			"device_ip":    deviceIP.String(),
			"device_name":  deviceName,
			"device_model": deviceModel,
			"device_port":  strconv.FormatUint(uint64(port), 10),
		}
		fields := map[string]interface{}{
			"bytes_sent":       stat.Sent,
			"bytes_recv":       stat.Received,
			"packets_total":    stat.Packets,
			"broadcasts_total": stat.Broadcasts,
			"multicasts_total": stat.Multicasts,
			"errors_total":     stat.Errors,
		}
		acc.AddCounter("nsdp_device_port", fields, tags)
	}
}

func init() {
	inputs.Add("nsdp", func() telegraf.Input {
		return &NSDP{}
	})
}
