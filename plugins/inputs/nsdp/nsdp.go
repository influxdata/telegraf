//go:generate ../../../tools/readme_config_includer/generator
package nsdp

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/tdrn-org/go-nsdp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type NSDP struct {
	Address     string          `toml:"address"`
	DeviceLimit uint            `toml:"device_limit"`
	Timeout     config.Duration `toml:"timeout"`
	Log         telegraf.Logger `toml:"-"`

	conn *nsdp.Conn
}

func (*NSDP) SampleConfig() string {
	return sampleConfig
}

func (n *NSDP) Init() error {
	if n.Address == "" {
		n.Address = nsdp.IPv4BroadcastTarget
	}
	if n.Timeout <= 0 {
		return errors.New("timeout must be greater than zero")
	}
	return nil
}

func (n *NSDP) Start(telegraf.Accumulator) error {
	conn, err := nsdp.NewConn(n.Address, n.Log.Level().Includes(telegraf.Trace))
	if err != nil {
		return fmt.Errorf("failed to create connection to address %s: %w", n.Address, err)
	}
	conn.ReceiveDeviceLimit = n.DeviceLimit
	conn.ReceiveTimeout = time.Duration(n.Timeout)
	n.conn = conn
	return nil
}

func (n *NSDP) Stop() {
	if n.conn == nil {
		return
	}
	n.conn.Close()
	n.conn = nil
}

func (n *NSDP) Gather(acc telegraf.Accumulator) error {
	if n.conn == nil {
		if err := n.Start(nil); err != nil {
			return err
		}
	}

	// Send request to query devices including infos (model, name, IP) and status (port statistics)
	request := nsdp.NewMessage(nsdp.ReadRequest)
	request.AppendTLV(nsdp.EmptyDeviceModel())
	request.AppendTLV(nsdp.EmptyDeviceName())
	request.AppendTLV(nsdp.EmptyDeviceIP())
	request.AppendTLV(nsdp.EmptyPortStatistic())
	responses, err := n.conn.SendReceiveMessage(request)
	if err != nil {
		// Close malfunctioning connection and re-connect on next Gather call
		n.Stop()
		return fmt.Errorf("failed to query address %s: %w", n.Address, err)
	}

	// Create metrics for each responding device
	for device, response := range responses {
		n.Log.Tracef("Processing device: %s", device)
		gatherDevice(acc, device, response)
	}
	return nil
}

func gatherDevice(acc telegraf.Accumulator, device string, response *nsdp.Message) {
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
		return &NSDP{Timeout: config.Duration(2 * time.Second)}
	})
}
