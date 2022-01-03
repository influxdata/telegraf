package wireguard

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/influxdata/telegraf/testutil"
)

func TestWireguard_gatherDeviceMetrics(t *testing.T) {
	var acc testutil.Accumulator

	wg := &Wireguard{}
	device := &wgtypes.Device{
		Name:         "wg0",
		Type:         wgtypes.LinuxKernel,
		ListenPort:   1,
		FirewallMark: 2,
		Peers:        []wgtypes.Peer{{}, {}},
	}

	expectFields := map[string]interface{}{
		"listen_port":   1,
		"firewall_mark": 2,
	}
	expectGauges := map[string]interface{}{
		"peers": 2,
	}
	expectTags := map[string]string{
		"name": "wg0",
		"type": "linux_kernel",
	}

	wg.gatherDeviceMetrics(&acc, device)

	require.Equal(t, 3, acc.NFields())
	acc.AssertDoesNotContainMeasurement(t, measurementPeer)
	acc.AssertContainsTaggedFields(t, measurementDevice, expectFields, expectTags)
	acc.AssertContainsTaggedFields(t, measurementDevice, expectGauges, expectTags)
}

func TestWireguard_gatherDevicePeerMetrics(t *testing.T) {
	var acc testutil.Accumulator
	pubkey, _ := wgtypes.ParseKey("NZTRIrv/ClTcQoNAnChEot+WL7OH7uEGQmx8oAN9rWE=")

	wg := &Wireguard{}
	device := &wgtypes.Device{
		Name: "wg0",
	}
	peer := wgtypes.Peer{
		PublicKey:                   pubkey,
		PersistentKeepaliveInterval: 1 * time.Minute,
		LastHandshakeTime:           time.Unix(100, 0),
		ReceiveBytes:                int64(40),
		TransmitBytes:               int64(60),
		AllowedIPs:                  []net.IPNet{{}, {}},
		ProtocolVersion:             0,
	}

	expectFields := map[string]interface{}{
		"persistent_keepalive_interval_ns": int64(60000000000),
		"protocol_version":                 0,
		"allowed_ips":                      2,
	}
	expectGauges := map[string]interface{}{
		"last_handshake_time_ns": int64(100000000000),
		"rx_bytes":               int64(40),
		"tx_bytes":               int64(60),
	}
	expectTags := map[string]string{
		"device":     "wg0",
		"public_key": pubkey.String(),
	}

	wg.gatherDevicePeerMetrics(&acc, device, peer)

	require.Equal(t, 6, acc.NFields())
	acc.AssertDoesNotContainMeasurement(t, measurementDevice)
	acc.AssertContainsTaggedFields(t, measurementPeer, expectFields, expectTags)
	acc.AssertContainsTaggedFields(t, measurementPeer, expectGauges, expectTags)
}
