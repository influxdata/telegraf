// +build !linux

package netstat

import (
	"syscall"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/require"
)

func TestNetStats(t *testing.T) {
	var mps system.MockPS
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	netstats := []net.ConnectionStat{
		{
			Type: syscall.SOCK_DGRAM,
		},
		{
			Status: "ESTABLISHED",
		},
		{
			Status: "ESTABLISHED",
		},
		{
			Status: "CLOSE",
		},
	}

	mps.On("NetConnections").Return(netstats, nil)

	acc.Metrics = nil
	err = (&NetStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	fields3 := map[string]interface{}{
		"tcp_established": 2,
		"tcp_syn_sent":    0,
		"tcp_syn_recv":    0,
		"tcp_fin_wait1":   0,
		"tcp_fin_wait2":   0,
		"tcp_time_wait":   0,
		"tcp_close":       1,
		"tcp_close_wait":  0,
		"tcp_last_ack":    0,
		"tcp_listen":      0,
		"tcp_closing":     0,
		"tcp_none":        0,
		"udp_socket":      1,
	}
	acc.AssertContainsTaggedFields(t, "netstat", fields3, make(map[string]string))
}
