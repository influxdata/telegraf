package system

import (
	"syscall"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetStats(t *testing.T) {
	var mps MockPS
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	netio := net.NetIOCountersStat{
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

	mps.On("NetIO").Return([]net.NetIOCountersStat{netio}, nil)

	netprotos := []net.NetProtoCountersStat{
		net.NetProtoCountersStat{
			Protocol: "Udp",
			Stats: map[string]int64{
				"InDatagrams": 4655,
				"NoPorts":     892592,
			},
		},
	}
	mps.On("NetProto").Return(netprotos, nil)

	netstats := []net.NetConnectionStat{
		net.NetConnectionStat{
			Type: syscall.SOCK_DGRAM,
		},
		net.NetConnectionStat{
			Status: "ESTABLISHED",
		},
		net.NetConnectionStat{
			Status: "ESTABLISHED",
		},
		net.NetConnectionStat{
			Status: "CLOSE",
		},
	}

	mps.On("NetConnections").Return(netstats, nil)

	err = (&NetIOStats{ps: &mps, skipChecks: true}).Gather(&acc)
	require.NoError(t, err)

	ntags := map[string]string{
		"interface": "eth0",
	}

	assert.NoError(t, acc.ValidateTaggedValue("bytes_sent", uint64(1123), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("bytes_recv", uint64(8734422), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("packets_sent", uint64(781), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("packets_recv", uint64(23456), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("err_in", uint64(832), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("err_out", uint64(8), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("drop_in", uint64(7), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("drop_out", uint64(1), ntags))
	assert.NoError(t, acc.ValidateValue("udp_noports", int64(892592)))
	assert.NoError(t, acc.ValidateValue("udp_indatagrams", int64(4655)))

	acc.Points = nil

	err = (&NetStats{&mps}).Gather(&acc)
	require.NoError(t, err)
	netstattags := map[string]string(nil)

	assert.NoError(t, acc.ValidateTaggedValue("tcp_established", 2, netstattags))
	assert.NoError(t, acc.ValidateTaggedValue("tcp_close", 1, netstattags))
	assert.NoError(t, acc.ValidateTaggedValue("udp_socket", 1, netstattags))
}
