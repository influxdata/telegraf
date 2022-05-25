package net

import (
	"syscall"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/netstat"
	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/stretchr/testify/require"
)

func TestNetStats(t *testing.T) {
	var mps system.MockPS
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	netio := net.IOCountersStat{
		Name:        "eth0",
		BytesSent:   1123,
		BytesRecv:   8734422,
		PacketsSent: 781,
		PacketsRecv: 23456,
		Errin:       832,
		Errout:      8,
		Dropin:      7,
		Dropout:     1,
	}

	mps.On("NetIO").Return([]net.IOCountersStat{netio}, nil)

	netprotos := []net.ProtoCountersStat{
		{
			Protocol: "Udp",
			Stats: map[string]int64{
				"InDatagrams": 4655,
				"NoPorts":     892592,
			},
		},
	}
	mps.On("NetProto").Return(netprotos, nil)

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

	err = (&NetIOStats{ps: &mps, skipChecks: true}).Gather(&acc)
	require.NoError(t, err)

	ntags := map[string]string{
		"interface": "eth0",
	}

	fields1 := map[string]interface{}{
		"bytes_sent":   uint64(1123),
		"bytes_recv":   uint64(8734422),
		"packets_sent": uint64(781),
		"packets_recv": uint64(23456),
		"err_in":       uint64(832),
		"err_out":      uint64(8),
		"drop_in":      uint64(7),
		"drop_out":     uint64(1),
	}
	acc.AssertContainsTaggedFields(t, "net", fields1, ntags)

	fields2 := map[string]interface{}{
		"udp_noports":     int64(892592),
		"udp_indatagrams": int64(4655),
	}
	ntags = map[string]string{
		"interface": "all",
	}
	acc.AssertContainsTaggedFields(t, "net", fields2, ntags)

	acc.Metrics = nil

	err = (&netstat.NetStats{
		PS: &mps,
	}).Gather(&acc)
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

	acc.Metrics = nil
	err = (&NetIOStats{ps: &mps, IgnoreProtocolStats: true}).Gather(&acc)
	require.NoError(t, err)

	acc.AssertDoesNotContainsTaggedFields(t, "netstat", fields3, make(map[string]string))
}
