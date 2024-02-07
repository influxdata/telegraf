package zabbix

import (
	"testing"
	"time"

	"github.com/datadope-io/go-zabbix/v2"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

// TestZabbixAutoregisterDisabledAdd tests that Add does not store information if autoregister is disabled.
func TestZabbixAutoregisterDisabledAdd(t *testing.T) {
	z := Zabbix{
		autoregisterLastSend: make(map[string]time.Time),
	}

	z.autoregisterAdd("hostname")
	require.Empty(t, z.autoregisterLastSend)
}

// TestZabbixAutoregisterEnabledAdd tests that Add stores information if autoregister is enabled.
func TestZabbixAutoregisterEnabledAdd(t *testing.T) {
	z := Zabbix{
		Autoregister:         "autoregister",
		autoregisterLastSend: make(map[string]time.Time),
	}

	z.autoregisterAdd("hostname")
	require.Len(t, z.autoregisterLastSend, 1)

	require.Contains(t, z.autoregisterLastSend, "hostname")
}

// TestZabbixAutoregisterPush tests different cases of Push with a table oriented test.
func TestZabbixAutoregisterPush(t *testing.T) {
	zabbixSender := &mockZabbixSender{}
	z := Zabbix{
		Log:                        testutil.Logger{},
		AutoregisterResendInterval: config.Duration(1 * time.Second),
		autoregisterLastSend:       make(map[string]time.Time),
		sender:                     zabbixSender,
	}

	// Test that nothing is sent if autoregister is disabled.
	z.autoregisterPush()
	require.Empty(t, z.autoregisterLastSend)

	// Test that nothing is sent if autoregister is enabled but no host is added.
	z.Autoregister = "autoregister"
	z.autoregisterPush()
	require.Empty(t, z.autoregisterLastSend)

	// Test that autoregister is sent if autoregister is enabled and a host is added.
	z.Autoregister = "autoregister"
	z.autoregisterAdd("hostname")
	z.autoregisterPush()
	require.Len(t, z.autoregisterLastSend, 1)
	require.Equal(t, "hostname", zabbixSender.hostname)
	require.Equal(t, "autoregister", zabbixSender.hostMetadata)

	// Test that autoregister is not sent if the last send was less than AutoregisterResendInterval ago.
	z.Autoregister = "autoregister"
	z.autoregisterAdd("hostname")
	z.autoregisterLastSend["hostname"] = time.Now().Add(time.Hour)
	zabbixSender.Reset()
	z.autoregisterPush()
	require.Len(t, z.autoregisterLastSend, 1)
	require.Equal(t, "", zabbixSender.hostname)
	require.Equal(t, "", zabbixSender.hostMetadata)

	// Test that autoregister is sent if last send was more than autoregisterSendPeriod ago.
	z.Autoregister = "autoregister"
	z.autoregisterAdd("hostname")
	z.autoregisterLastSend["hostname"] = time.Now().Add(-24 * time.Hour)
	zabbixSender.Reset()
	z.autoregisterPush()
	require.Len(t, z.autoregisterLastSend, 1)
	require.Equal(t, "hostname", zabbixSender.hostname)
	require.Equal(t, "autoregister", zabbixSender.hostMetadata)
}

// mockZabbixSender is a mock of ZabbixAutoregisterSender.
type mockZabbixSender struct {
	hostname     string
	hostMetadata string
	sendMetrics  []*zabbix.Metric
	sendPackets  []*zabbix.Packet
}

// Reset resets the mock.
func (m *mockZabbixSender) Reset() {
	m.hostname = ""
	m.hostMetadata = ""
	m.sendMetrics = nil
	m.sendPackets = nil
}

// RegisterHost is a mock of ZabbixAutoregisterSender.RegisterHost.
func (m *mockZabbixSender) RegisterHost(hostname, hostMetadata string) error {
	m.hostname = hostname
	m.hostMetadata = hostMetadata

	return nil
}

// RegisterHost is a mock of ZabbixAutoregisterSender.RegisterHost.
func (m *mockZabbixSender) Send(packet *zabbix.Packet) (res zabbix.Response, err error) {
	m.sendPackets = append(m.sendPackets, packet)
	return zabbix.Response{}, nil
}

func (m *mockZabbixSender) SendMetrics(metrics []*zabbix.Metric) (
	resActive zabbix.Response,
	resTrapper zabbix.Response,
	err error,
) {
	m.sendMetrics = append(m.sendMetrics, metrics...)
	return zabbix.Response{}, zabbix.Response{}, nil
}
