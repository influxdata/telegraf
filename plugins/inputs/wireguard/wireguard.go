package wireguard

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	measurementDevice = "wireguard_device"
	measurementPeer   = "wireguard_peer"
)

var (
	deviceTypeNames = map[wgtypes.DeviceType]string{
		wgtypes.Unknown:     "unknown",
		wgtypes.LinuxKernel: "linux_kernel",
		wgtypes.Userspace:   "userspace",
	}
)

// Wireguard is an input that enumerates all Wireguard interfaces/devices on
// the host, and reports gauge metrics for the device itself and its peers.
type Wireguard struct {
	Devices []string `toml:"devices"`

	client *wgctrl.Client
}

func (wg *Wireguard) Description() string {
	return "Collect Wireguard server interface and peer statistics"
}

func (wg *Wireguard) SampleConfig() string {
	return `
  ## Optional list of Wireguard device/interface names to query.
  ## If omitted, all Wireguard interfaces are queried.
  # devices = ["wg0"]
`
}

func (wg *Wireguard) Init() error {
	var err error

	wg.client, err = wgctrl.New()

	return err
}

func (wg *Wireguard) Gather(acc telegraf.Accumulator) error {
	devices, err := wg.enumerateDevices()
	if err != nil {
		return fmt.Errorf("error enumerating Wireguard devices: %v", err)
	}

	for _, device := range devices {
		wg.gatherDeviceMetrics(acc, device)

		for _, peer := range device.Peers {
			wg.gatherDevicePeerMetrics(acc, device, peer)
		}
	}

	return nil
}

func (wg *Wireguard) enumerateDevices() ([]*wgtypes.Device, error) {
	var devices []*wgtypes.Device

	// If no device names are specified, defer to the library to enumerate
	// all of them
	if len(wg.Devices) == 0 {
		return wg.client.Devices()
	}

	// Otherwise, explicitly populate only device names specified in config
	for _, name := range wg.Devices {
		dev, err := wg.client.Device(name)
		if err != nil {
			log.Printf("W! [inputs.wireguard] No Wireguard device found with name %s", name)
			continue
		}

		devices = append(devices, dev)
	}

	return devices, nil
}

func (wg *Wireguard) gatherDeviceMetrics(acc telegraf.Accumulator, device *wgtypes.Device) {
	fields := map[string]interface{}{
		"listen_port":   device.ListenPort,
		"firewall_mark": device.FirewallMark,
	}

	gauges := map[string]interface{}{
		"peers": len(device.Peers),
	}

	tags := map[string]string{
		"name": device.Name,
		"type": deviceTypeNames[device.Type],
	}

	acc.AddFields(measurementDevice, fields, tags)
	acc.AddGauge(measurementDevice, gauges, tags)
}

func (wg *Wireguard) gatherDevicePeerMetrics(acc telegraf.Accumulator, device *wgtypes.Device, peer wgtypes.Peer) {
	fields := map[string]interface{}{
		"persistent_keepalive_interval_ns": peer.PersistentKeepaliveInterval.Nanoseconds(),
		"protocol_version":                 peer.ProtocolVersion,
		"allowed_ips":                      len(peer.AllowedIPs),
	}

	gauges := map[string]interface{}{
		"last_handshake_time_ns": peer.LastHandshakeTime.UnixNano(),
		"rx_bytes":               peer.ReceiveBytes,
		"tx_bytes":               peer.TransmitBytes,
	}

	tags := map[string]string{
		"device":     device.Name,
		"public_key": peer.PublicKey.String(),
	}

	acc.AddFields(measurementPeer, fields, tags)
	acc.AddGauge(measurementPeer, gauges, tags)
}

func init() {
	inputs.Add("wireguard", func() telegraf.Input {
		return &Wireguard{}
	})
}
