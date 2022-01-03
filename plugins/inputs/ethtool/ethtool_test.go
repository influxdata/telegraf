//go:build linux
// +build linux

package ethtool

import (
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

var command *Ethtool
var interfaceMap map[string]*InterfaceMock

type InterfaceMock struct {
	Name        string
	DriverName  string
	Stat        map[string]uint64
	LoopBack    bool
	InterfaceUp bool
}

type CommandEthtoolMock struct {
	InterfaceMap map[string]*InterfaceMock
}

func (c *CommandEthtoolMock) Init() error {
	// Not required for test mock
	return nil
}

func (c *CommandEthtoolMock) DriverName(intf string) (string, error) {
	i := c.InterfaceMap[intf]
	if i != nil {
		return i.DriverName, nil
	}
	return "", errors.New("interface not found")
}

func (c *CommandEthtoolMock) Interfaces() ([]net.Interface, error) {
	interfaceNames := make([]net.Interface, 0)
	for k, v := range c.InterfaceMap {
		var flag net.Flags
		// When interface is up
		if v.InterfaceUp {
			flag |= net.FlagUp
		}
		// For loopback interface
		if v.LoopBack {
			flag |= net.FlagLoopback
		}

		// Create a dummy interface
		iface := net.Interface{
			Index:        0,
			MTU:          1500,
			Name:         k,
			HardwareAddr: nil,
			Flags:        flag,
		}
		interfaceNames = append(interfaceNames, iface)
	}
	return interfaceNames, nil
}

func (c *CommandEthtoolMock) Stats(intf string) (map[string]uint64, error) {
	i := c.InterfaceMap[intf]
	if i != nil {
		return i.Stat, nil
	}
	return nil, errors.New("interface not found")
}

func setup() {
	interfaceMap = make(map[string]*InterfaceMock)

	eth1Stat := map[string]uint64{
		"interface_up":                   1,
		"port_rx_1024_to_15xx":           25167245,
		"port_rx_128_to_255":             1573526387,
		"port_rx_15xx_to_jumbo":          137819058,
		"port_rx_256_to_511":             772038107,
		"port_rx_512_to_1023":            78294457,
		"port_rx_64":                     8798065,
		"port_rx_65_to_127":              450348015,
		"port_rx_bad":                    0,
		"port_rx_bad_bytes":              0,
		"port_rx_bad_gtjumbo":            0,
		"port_rx_broadcast":              6428250,
		"port_rx_bytes":                  893460472634,
		"port_rx_control":                0,
		"port_rx_dp_di_dropped_packets":  2772680304,
		"port_rx_dp_hlb_fetch":           0,
		"port_rx_dp_hlb_wait":            0,
		"port_rx_dp_q_disabled_packets":  0,
		"port_rx_dp_streaming_packets":   0,
		"port_rx_good":                   3045991334,
		"port_rx_good_bytes":             893460472927,
		"port_rx_gtjumbo":                0,
		"port_rx_lt64":                   0,
		"port_rx_multicast":              1639566045,
		"port_rx_nodesc_drops":           0,
		"port_rx_overflow":               0,
		"port_rx_packets":                3045991334,
		"port_rx_pause":                  0,
		"port_rx_pm_discard_bb_overflow": 0,
		"port_rx_pm_discard_mapping":     0,
		"port_rx_pm_discard_qbb":         0,
		"port_rx_pm_discard_vfifo_full":  0,
		"port_rx_pm_trunc_bb_overflow":   0,
		"port_rx_pm_trunc_qbb":           0,
		"port_rx_pm_trunc_vfifo_full":    0,
		"port_rx_unicast":                1399997040,
		"port_tx_1024_to_15xx":           236,
		"port_tx_128_to_255":             275090219,
		"port_tx_15xx_to_jumbo":          926,
		"port_tx_256_to_511":             48567221,
		"port_tx_512_to_1023":            5142016,
		"port_tx_64":                     113903973,
		"port_tx_65_to_127":              161935699,
		"port_tx_broadcast":              8,
		"port_tx_bytes":                  94357131016,
		"port_tx_control":                0,
		"port_tx_lt64":                   0,
		"port_tx_multicast":              325891647,
		"port_tx_packets":                604640290,
		"port_tx_pause":                  0,
		"port_tx_unicast":                278748635,
		"ptp_bad_syncs":                  1,
		"ptp_fast_syncs":                 1,
		"ptp_filter_matches":             0,
		"ptp_good_syncs":                 136151,
		"ptp_invalid_sync_windows":       0,
		"ptp_no_time_syncs":              1,
		"ptp_non_filter_matches":         0,
		"ptp_oversize_sync_windows":      53,
		"ptp_rx_no_timestamp":            0,
		"ptp_rx_timestamp_packets":       0,
		"ptp_sync_timeouts":              1,
		"ptp_timestamp_packets":          0,
		"ptp_tx_timestamp_packets":       0,
		"ptp_undersize_sync_windows":     3,
		"rx-0.rx_packets":                55659234,
		"rx-1.rx_packets":                87880538,
		"rx-2.rx_packets":                26746234,
		"rx-3.rx_packets":                103026471,
		"rx-4.rx_packets":                0,
		"rx_eth_crc_err":                 0,
		"rx_frm_trunc":                   0,
		"rx_inner_ip_hdr_chksum_err":     0,
		"rx_inner_tcp_udp_chksum_err":    0,
		"rx_ip_hdr_chksum_err":           0,
		"rx_mcast_mismatch":              0,
		"rx_merge_events":                0,
		"rx_merge_packets":               0,
		"rx_nodesc_trunc":                0,
		"rx_noskb_drops":                 0,
		"rx_outer_ip_hdr_chksum_err":     0,
		"rx_outer_tcp_udp_chksum_err":    0,
		"rx_reset":                       0,
		"rx_tcp_udp_chksum_err":          0,
		"rx_tobe_disc":                   0,
		"tx-0.tx_packets":                85843565,
		"tx-1.tx_packets":                108642725,
		"tx-2.tx_packets":                202596078,
		"tx-3.tx_packets":                207561010,
		"tx-4.tx_packets":                0,
		"tx_cb_packets":                  4,
		"tx_merge_events":                11025,
		"tx_pio_packets":                 531928114,
		"tx_pushes":                      604643378,
		"tx_tso_bursts":                  0,
		"tx_tso_fallbacks":               0,
		"tx_tso_long_headers":            0,
	}
	eth1 := &InterfaceMock{"eth1", "driver1", eth1Stat, false, true}
	interfaceMap[eth1.Name] = eth1

	eth2Stat := map[string]uint64{
		"interface_up":                   0,
		"port_rx_1024_to_15xx":           11529312,
		"port_rx_128_to_255":             1868952037,
		"port_rx_15xx_to_jumbo":          130339387,
		"port_rx_256_to_511":             843846270,
		"port_rx_512_to_1023":            173194372,
		"port_rx_64":                     9190374,
		"port_rx_65_to_127":              507806115,
		"port_rx_bad":                    0,
		"port_rx_bad_bytes":              0,
		"port_rx_bad_gtjumbo":            0,
		"port_rx_broadcast":              6648019,
		"port_rx_bytes":                  1007358162202,
		"port_rx_control":                0,
		"port_rx_dp_di_dropped_packets":  3164124639,
		"port_rx_dp_hlb_fetch":           0,
		"port_rx_dp_hlb_wait":            0,
		"port_rx_dp_q_disabled_packets":  0,
		"port_rx_dp_streaming_packets":   0,
		"port_rx_good":                   3544857867,
		"port_rx_good_bytes":             1007358162202,
		"port_rx_gtjumbo":                0,
		"port_rx_lt64":                   0,
		"port_rx_multicast":              2231999743,
		"port_rx_nodesc_drops":           0,
		"port_rx_overflow":               0,
		"port_rx_packets":                3544857867,
		"port_rx_pause":                  0,
		"port_rx_pm_discard_bb_overflow": 0,
		"port_rx_pm_discard_mapping":     0,
		"port_rx_pm_discard_qbb":         0,
		"port_rx_pm_discard_vfifo_full":  0,
		"port_rx_pm_trunc_bb_overflow":   0,
		"port_rx_pm_trunc_qbb":           0,
		"port_rx_pm_trunc_vfifo_full":    0,
		"port_rx_unicast":                1306210105,
		"port_tx_1024_to_15xx":           379,
		"port_tx_128_to_255":             202767251,
		"port_tx_15xx_to_jumbo":          558,
		"port_tx_256_to_511":             31454719,
		"port_tx_512_to_1023":            6865731,
		"port_tx_64":                     17268276,
		"port_tx_65_to_127":              272816313,
		"port_tx_broadcast":              6,
		"port_tx_bytes":                  78071946593,
		"port_tx_control":                0,
		"port_tx_lt64":                   0,
		"port_tx_multicast":              239510586,
		"port_tx_packets":                531173227,
		"port_tx_pause":                  0,
		"port_tx_unicast":                291662635,
		"ptp_bad_syncs":                  0,
		"ptp_fast_syncs":                 0,
		"ptp_filter_matches":             0,
		"ptp_good_syncs":                 0,
		"ptp_invalid_sync_windows":       0,
		"ptp_no_time_syncs":              0,
		"ptp_non_filter_matches":         0,
		"ptp_oversize_sync_windows":      0,
		"ptp_rx_no_timestamp":            0,
		"ptp_rx_timestamp_packets":       0,
		"ptp_sync_timeouts":              0,
		"ptp_timestamp_packets":          0,
		"ptp_tx_timestamp_packets":       0,
		"ptp_undersize_sync_windows":     0,
		"rx-0.rx_packets":                84587075,
		"rx-1.rx_packets":                74029305,
		"rx-2.rx_packets":                134586471,
		"rx-3.rx_packets":                87531322,
		"rx-4.rx_packets":                0,
		"rx_eth_crc_err":                 0,
		"rx_frm_trunc":                   0,
		"rx_inner_ip_hdr_chksum_err":     0,
		"rx_inner_tcp_udp_chksum_err":    0,
		"rx_ip_hdr_chksum_err":           0,
		"rx_mcast_mismatch":              0,
		"rx_merge_events":                0,
		"rx_merge_packets":               0,
		"rx_nodesc_trunc":                0,
		"rx_noskb_drops":                 0,
		"rx_outer_ip_hdr_chksum_err":     0,
		"rx_outer_tcp_udp_chksum_err":    0,
		"rx_reset":                       0,
		"rx_tcp_udp_chksum_err":          0,
		"rx_tobe_disc":                   0,
		"tx-0.tx_packets":                232521451,
		"tx-1.tx_packets":                97876137,
		"tx-2.tx_packets":                106822111,
		"tx-3.tx_packets":                93955050,
		"tx-4.tx_packets":                0,
		"tx_cb_packets":                  1,
		"tx_merge_events":                8402,
		"tx_pio_packets":                 481040054,
		"tx_pushes":                      531174491,
		"tx_tso_bursts":                  128,
		"tx_tso_fallbacks":               0,
		"tx_tso_long_headers":            0,
	}
	eth2 := &InterfaceMock{"eth2", "driver1", eth2Stat, false, false}
	interfaceMap[eth2.Name] = eth2

	// dummy loopback including dummy stat to ensure that the ignore feature is working
	lo0Stat := map[string]uint64{
		"dummy": 0,
	}
	lo0 := &InterfaceMock{"lo0", "", lo0Stat, true, true}
	interfaceMap[lo0.Name] = lo0

	c := &CommandEthtoolMock{interfaceMap}
	command = &Ethtool{
		InterfaceInclude: []string{},
		InterfaceExclude: []string{},
		command:          c,
	}
}

func toStringMapInterface(in map[string]uint64) map[string]interface{} {
	var m = map[string]interface{}{}
	for k, v := range in {
		m[k] = v
	}
	return m
}

func TestGather(t *testing.T) {
	setup()
	var acc testutil.Accumulator

	err := command.Gather(&acc)
	require.NoError(t, err)
	require.Len(t, acc.Metrics, 2)

	expectedFieldsEth1 := toStringMapInterface(interfaceMap["eth1"].Stat)
	expectedTagsEth1 := map[string]string{
		"interface": "eth1",
		"driver":    "driver1",
	}
	acc.AssertContainsTaggedFields(t, pluginName, expectedFieldsEth1, expectedTagsEth1)
	expectedFieldsEth2 := toStringMapInterface(interfaceMap["eth2"].Stat)
	expectedTagsEth2 := map[string]string{
		"interface": "eth2",
		"driver":    "driver1",
	}
	acc.AssertContainsTaggedFields(t, pluginName, expectedFieldsEth2, expectedTagsEth2)
}

func TestGatherIncludeInterfaces(t *testing.T) {
	setup()
	var acc testutil.Accumulator

	command.InterfaceInclude = append(command.InterfaceInclude, "eth1")

	err := command.Gather(&acc)
	require.NoError(t, err)
	require.Len(t, acc.Metrics, 1)

	// Should contain eth1
	expectedFieldsEth1 := toStringMapInterface(interfaceMap["eth1"].Stat)
	expectedTagsEth1 := map[string]string{
		"interface": "eth1",
		"driver":    "driver1",
	}
	acc.AssertContainsTaggedFields(t, pluginName, expectedFieldsEth1, expectedTagsEth1)

	// Should not contain eth2
	expectedFieldsEth2 := toStringMapInterface(interfaceMap["eth2"].Stat)
	expectedTagsEth2 := map[string]string{
		"interface": "eth2",
		"driver":    "driver1",
	}
	acc.AssertDoesNotContainsTaggedFields(t, pluginName, expectedFieldsEth2, expectedTagsEth2)
}

func TestGatherIgnoreInterfaces(t *testing.T) {
	setup()
	var acc testutil.Accumulator

	command.InterfaceExclude = append(command.InterfaceExclude, "eth1")

	err := command.Gather(&acc)
	require.NoError(t, err)
	require.Len(t, acc.Metrics, 1)

	// Should not contain eth1
	expectedFieldsEth1 := toStringMapInterface(interfaceMap["eth1"].Stat)
	expectedTagsEth1 := map[string]string{
		"interface": "eth1",
		"driver":    "driver1",
	}
	acc.AssertDoesNotContainsTaggedFields(t, pluginName, expectedFieldsEth1, expectedTagsEth1)

	// Should contain eth2
	expectedFieldsEth2 := toStringMapInterface(interfaceMap["eth2"].Stat)
	expectedTagsEth2 := map[string]string{
		"interface": "eth2",
		"driver":    "driver1",
	}
	acc.AssertContainsTaggedFields(t, pluginName, expectedFieldsEth2, expectedTagsEth2)
}

type TestCase struct {
	normalization  []string
	stats          map[string]uint64
	expectedFields map[string]uint64
}

func TestNormalizedKeys(t *testing.T) {
	cases := []TestCase{
		{
			normalization: []string{"underscore"},
			stats: map[string]uint64{
				"port rx":      1,
				" Port_tx":     0,
				"interface_up": 0,
			},
			expectedFields: map[string]uint64{
				"port_rx":      1,
				"_Port_tx":     0,
				"interface_up": 0,
			},
		},
		{
			normalization: []string{"underscore", "lower"},
			stats: map[string]uint64{
				"Port rx":      1,
				" Port_tx":     0,
				"interface_up": 0,
			},
			expectedFields: map[string]uint64{
				"port_rx":      1,
				"_port_tx":     0,
				"interface_up": 0,
			},
		},
		{
			normalization: []string{"underscore", "lower", "trim"},
			stats: map[string]uint64{
				"  Port RX ":   1,
				" Port_tx":     0,
				"interface_up": 0,
			},
			expectedFields: map[string]uint64{
				"port_rx":      1,
				"port_tx":      0,
				"interface_up": 0,
			},
		},
		{
			normalization: []string{"underscore", "lower", "snakecase", "trim"},
			stats: map[string]uint64{
				"  Port RX ":   1,
				" Port_tx":     0,
				"interface_up": 0,
			},
			expectedFields: map[string]uint64{
				"port_rx":      1,
				"port_tx":      0,
				"interface_up": 0,
			},
		},
		{
			normalization: []string{"snakecase"},
			stats: map[string]uint64{
				"  PortRX ":    1,
				" PortTX":      0,
				"interface_up": 0,
			},
			expectedFields: map[string]uint64{
				"port_rx":      1,
				"port_tx":      0,
				"interface_up": 0,
			},
		},
		{
			normalization: []string{},
			stats: map[string]uint64{
				"  Port RX ":   1,
				" Port_tx":     0,
				"interface_up": 0,
			},
			expectedFields: map[string]uint64{
				"  Port RX ":   1,
				" Port_tx":     0,
				"interface_up": 0,
			},
		},
	}
	for _, c := range cases {
		eth0 := &InterfaceMock{"eth0", "e1000e", c.stats, false, true}
		expectedTags := map[string]string{
			"interface": eth0.Name,
			"driver":    eth0.DriverName,
		}

		interfaceMap = make(map[string]*InterfaceMock)
		interfaceMap[eth0.Name] = eth0

		cmd := &CommandEthtoolMock{interfaceMap}
		command = &Ethtool{
			InterfaceInclude: []string{},
			InterfaceExclude: []string{},
			NormalizeKeys:    c.normalization,
			command:          cmd,
		}

		var acc testutil.Accumulator
		err := command.Gather(&acc)

		require.NoError(t, err)
		require.Len(t, acc.Metrics, 1)

		acc.AssertContainsFields(t, pluginName, toStringMapInterface(c.expectedFields))
		acc.AssertContainsTaggedFields(t, pluginName, toStringMapInterface(c.expectedFields), expectedTags)
	}
}
