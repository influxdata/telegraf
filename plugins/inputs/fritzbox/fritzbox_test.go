package fritzbox

import (
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-tr064/mock"
)

func TestInitDefaults(t *testing.T) {
	plugin := defaultFritzbox()
	err := plugin.Init()
	require.NoError(t, err)
	require.Equal(t, 0, len(plugin.Devices))
	require.True(t, plugin.DeviceInfo)
	require.True(t, plugin.WanInfo)
	require.True(t, plugin.PppInfo)
	require.True(t, plugin.DslInfo)
	require.True(t, plugin.WlanInfo)
	require.False(t, plugin.HostsInfo)
	require.Equal(t, 30, plugin.FullQueryCycle)
	require.Equal(t, defaultTimeout, plugin.Timeout)
	require.False(t, plugin.TlsSkipVerify)
	require.NotNil(t, plugin.Log)
}

func TestConfig(t *testing.T) {
	conf, err := os.ReadFile("testdata/fritzbox.conf")
	require.NoError(t, err)
	var plugin = defaultFritzbox()
	err = toml.Unmarshal(conf, plugin)
	require.NoError(t, err)
	err = plugin.Init()
	require.NoError(t, err)
	require.Equal(t, 2, len(plugin.Devices))
	require.False(t, plugin.DeviceInfo)
	require.False(t, plugin.WanInfo)
	require.False(t, plugin.PppInfo)
	require.False(t, plugin.DslInfo)
	require.False(t, plugin.WlanInfo)
	require.True(t, plugin.HostsInfo)
	require.Equal(t, 6, plugin.FullQueryCycle)
	require.Equal(t, config.Duration(60*time.Second), plugin.Timeout)
	require.True(t, plugin.TlsSkipVerify)
}

var testMocks = []*mock.ServiceMock{
	mock.ServiceMockFromFile("/deviceinfo", "testdata/DeviceInfo.xml"),
	mock.ServiceMockFromFile("/wancommonifconfig", "testdata/WANCommonInterfaceConfig1.xml"),
	mock.ServiceMockFromFile("/WANCommonIFC1", "testdata/WANCommonInterfaceConfig2.xml"),
	mock.ServiceMockFromFile("/wanpppconn", "testdata/WANPPPConnection.xml"),
	mock.ServiceMockFromFile("/wandslifconfig", "testdata/WANDSLInterfaceConfig.xml"),
	mock.ServiceMockFromFile("/wlanconfig", "testdata/WLANConfiguration.xml"),
	mock.ServiceMockFromFile("/hosts", "testdata/Hosts.xml"),
}

func TestGatherDeviceInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Actual test
	plugin := defaultFritzbox()
	plugin.Devices = append(plugin.Devices, tr064Server.Server().String())
	plugin.DeviceInfo = true
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMeasurement(t, acc, "fritzbox_device", []string{"source", "service"}, []string{"uptime", "model_name", "serial_number", "hardware_version", "software_version"})
}

func TestGatherWanInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Actual test
	plugin := defaultFritzbox()
	plugin.Devices = append(plugin.Devices, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = true
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMeasurement(t, acc, "fritzbox_wan", []string{"source", "service"}, []string{"layer1_upstream_max_bit_rate", "layer1_downstream_max_bit_rate", "upstream_current_max_speed", "downstream_current_max_speed", "total_bytes_sent", "total_bytes_received"})
}

func TestGatherPppInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Actual test
	plugin := defaultFritzbox()
	plugin.Devices = append(plugin.Devices, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = true
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMeasurement(t, acc, "fritzbox_ppp", []string{"source", "service"}, []string{"uptime", "upstream_max_bit_rate", "downstream_max_bit_rate"})
}

func TestGatherDslInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Actual test
	plugin := defaultFritzbox()
	plugin.Devices = append(plugin.Devices, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = true
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMeasurement(t, acc, "fritzbox_dsl", []string{"source", "service"}, []string{"upstream_curr_rate", "downstream_curr_rate", "upstream_max_rate", "downstream_max_rate", "upstream_noise_margin", "downstream_noise_margin", "upstream_attenuation", "downstream_attenuation", "upstream_power", "downstream_power", "receive_blocks", "transmit_blocks", "cell_delin", "link_retrain", "init_errors", "init_timeouts", "loss_of_framing", "errored_secs", "severly_errored_secs", "fec_errors", "atuc_fec_errors", "hec_errors", "atuc_hec_errors", "crc_errors", "atuc_crc_errors"})
}

func TestGatherWlanInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Actual test
	plugin := defaultFritzbox()
	plugin.Devices = append(plugin.Devices, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = true
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMeasurement(t, acc, "fritzbox_wlan", []string{"source", "service", "ssid", "channel", "band"}, []string{"total_associations"})
}

func TestGatherHostsInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Actual test
	plugin := defaultFritzbox()
	plugin.Devices = append(plugin.Devices, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = true
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMeasurement(t, acc, "fritzbox_host", []string{"source", "service", "host", "host_role", "host_ap", "host_ap_role", "link_type", "link_name"}, []string{"max_data_rate_tx", "max_data_rate_rx", "cur_data_rate_tx", "cur_data_rate_rx"})
}

func testMeasurement(t *testing.T, acc *testutil.Accumulator, measurement string, tags []string, fields []string) {
	require.Truef(t, acc.HasMeasurement(measurement), "measurement: %s", measurement)
	for _, tag := range tags {
		require.Truef(t, acc.HasTag(measurement, tag), "measurement: %s tag: %s", measurement, tag)
	}
	for _, field := range fields {
		require.True(t, acc.HasField(measurement, field), "measurement: %s field: %s", measurement, field)
	}
}
