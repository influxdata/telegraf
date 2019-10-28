package wireguard

import (
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const sampleConfig = `
  # if none are provided, all of the available ones will be scraped
  interfaces = ["wg0"]
`

const description = "Reads metrics from a Wireguard interface"

// Wireguard holds the configuration of the plugin.
type Wireguard struct {
	Interfaces []string `toml:"interfaces"`
}

// Description returns description of the plugin.
func (w *Wireguard) Description() string {
	return description
}

// SampleConfig returns configuration sample for the plugin.
func (w *Wireguard) SampleConfig() string {
	return sampleConfig
}

// Gather adds metrics into the accumulator
func (w *Wireguard) Gather(acc telegraf.Accumulator) error {
	// Initializing wireguard client
	cli, err := wgctrl.New()
	if err != nil {
		acc.AddError(fmt.Errorf("Error creating the wireguard client: %v", err))
		return nil
	}
	defer cli.Close()
	// Gathering devices
	devices := []*wgtypes.Device{}
	if len(w.Interfaces) == 0 {
		devices, err = cli.Devices()
		if err != nil {
			acc.AddError(fmt.Errorf("Error getting all Wireguard interfaces: %v", err))
			return nil
		}
	} else {
		for _, interfaceName := range w.Interfaces {
			if interfaceName == "" {
				acc.AddError(fmt.Errorf("The interface name cannot be empty"))
				return nil
			}
			device, err := cli.Device(interfaceName)
			if err != nil {
				acc.AddError(fmt.Errorf("Error getting %s Wireguard interface: %v", interfaceName, err))
				return nil
			}
			devices = append(devices, device)
		}
	}
	if len(devices) == 0 {
		acc.AddError(fmt.Errorf("There are no Wireguard interfaces on this node or the user doesn't have access to any of them"))
		return nil
	}
	// Getting metrics from devices
	for _, device := range devices {
		tags := map[string]string{
			"name":           device.Name,
			"type":           device.Type.String(),
			"serverpublikey": device.PublicKey.String(),
			"listenport":     strconv.Itoa(device.ListenPort),
			"firewallmark":   strconv.Itoa(device.FirewallMark),
		}
		for _, peer := range device.Peers {
			tags["peerpublickey"] = peer.PublicKey.String()
			tags["endpoint"] = peer.Endpoint.String()
			fields := map[string]interface{}{
				"received_bytes":               peer.ReceiveBytes,
				"transmit_bytes":               peer.TransmitBytes,
				"protocol":                     peer.ProtocolVersion,
				"last_hanshake_time":           peer.LastHandshakeTime.Unix(),
				"persisten_keepalive_interval": int(peer.PersistentKeepaliveInterval.Seconds()),
			}
			acc.AddFields("wireguard", fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("wireguard", func() telegraf.Input {
		return &Wireguard{
			Interfaces: []string{},
		}
	})
}
