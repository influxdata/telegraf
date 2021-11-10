//go:build linux
// +build linux

package infiniband

import (
	"github.com/Mellanox/rdmamap"
	"github.com/influxdata/telegraf/testutil"
	"testing"
)

func TestInfiniband(t *testing.T) {
	fields := map[string]interface{}{
		"excessive_buffer_overrun_errors": uint64(0),
		"link_downed":                     uint64(0),
		"link_error_recovery":             uint64(0),
		"local_link_integrity_errors":     uint64(0),
		"multicast_rcv_packets":           uint64(0),
		"multicast_xmit_packets":          uint64(0),
		"port_rcv_constraint_errors":      uint64(0),
		"port_rcv_data":                   uint64(237159415345822),
		"port_rcv_errors":                 uint64(0),
		"port_rcv_packets":                uint64(801977655075),
		"port_rcv_remote_physical_errors": uint64(0),
		"port_rcv_switch_relay_errors":    uint64(0),
		"port_xmit_constraint_errors":     uint64(0),
		"port_xmit_data":                  uint64(238334949937759),
		"port_xmit_discards":              uint64(0),
		"port_xmit_packets":               uint64(803162651391),
		"port_xmit_wait":                  uint64(4294967295),
		"symbol_error":                    uint64(0),
		"unicast_rcv_packets":             uint64(801977655075),
		"unicast_xmit_packets":            uint64(803162651391),
		"VL15_dropped":                    uint64(0),
	}

	tags := map[string]string{
		"device": "m1x5_0",
		"port":   "1",
	}

	sampleRdmastatsEntries := []rdmamap.RdmaStatEntry{
		{
			Name:  "excessive_buffer_overrun_errors",
			Value: uint64(0),
		},
		{
			Name:  "link_downed",
			Value: uint64(0),
		},
		{
			Name:  "link_error_recovery",
			Value: uint64(0),
		},
		{
			Name:  "local_link_integrity_errors",
			Value: uint64(0),
		},
		{
			Name:  "multicast_rcv_packets",
			Value: uint64(0),
		},
		{
			Name:  "multicast_xmit_packets",
			Value: uint64(0),
		},
		{
			Name:  "port_rcv_constraint_errors",
			Value: uint64(0),
		},
		{
			Name:  "port_rcv_data",
			Value: uint64(237159415345822),
		},
		{
			Name:  "port_rcv_errors",
			Value: uint64(0),
		},
		{
			Name:  "port_rcv_packets",
			Value: uint64(801977655075),
		},
		{
			Name:  "port_rcv_remote_physical_errors",
			Value: uint64(0),
		},
		{
			Name:  "port_rcv_switch_relay_errors",
			Value: uint64(0),
		},
		{
			Name:  "port_xmit_constraint_errors",
			Value: uint64(0),
		},
		{
			Name:  "port_xmit_data",
			Value: uint64(238334949937759),
		},
		{
			Name:  "port_xmit_discards",
			Value: uint64(0),
		},
		{
			Name:  "port_xmit_packets",
			Value: uint64(803162651391),
		},
		{
			Name:  "port_xmit_wait",
			Value: uint64(4294967295),
		},
		{
			Name:  "symbol_error",
			Value: uint64(0),
		},
		{
			Name:  "unicast_rcv_packets",
			Value: uint64(801977655075),
		},
		{
			Name:  "unicast_xmit_packets",
			Value: uint64(803162651391),
		},
		{
			Name:  "VL15_dropped",
			Value: uint64(0),
		},
	}

	var acc testutil.Accumulator

	addStats("m1x5_0", "1", sampleRdmastatsEntries, &acc)

	acc.AssertContainsTaggedFields(t, "infiniband", fields, tags)
}
