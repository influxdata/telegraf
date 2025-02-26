//go:build linux

package infiniband

import (
	"testing"

	"github.com/Mellanox/rdmamap"

	"github.com/influxdata/telegraf/testutil"
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

func TestInfinibandRDMA(t *testing.T) {
	fields := map[string]interface{}{
		"duplicate_request":          uint64(0),
		"implied_nak_seq_err":        uint64(0),
		"lifespan":                   uint64(10),
		"local_ack_timeout_err":      uint64(38),
		"np_cnp_sent":                uint64(10284520),
		"np_ecn_marked_roce_packets": uint64(286733949),
		"out_of_buffer":              uint64(1149772),
		"out_of_sequence":            uint64(44),
		"packet_seq_err":             uint64(1),
		"req_cqe_error":              uint64(10776),
		"req_cqe_flush_error":        uint64(2173),
		"req_remote_access_errors":   uint64(0),
		"req_remote_invalid_request": uint64(0),
		"resp_cqe_error":             uint64(759),
		"resp_cqe_flush_error":       uint64(759),
		"resp_local_length_error":    uint64(0),
		"resp_remote_access_errors":  uint64(0),
		"rnr_nak_retry_err":          uint64(0),
		"roce_adp_retrans":           uint64(0),
		"roce_adp_retrans_to":        uint64(0),
		"roce_slow_restart":          uint64(0),
		"roce_slow_restart_cnps":     uint64(0),
		"roce_slow_restart_trans":    uint64(0),
		"rp_cnp_handled":             uint64(1),
		"rp_cnp_ignored":             uint64(0),
		"rx_atomic_requests":         uint64(0),
		"rx_icrc_encapsulated":       uint64(0),
		"rx_read_requests":           uint64(488228),
		"rx_write_requests":          uint64(3928699),
	}

	tags := map[string]string{
		"device": "m1x5_0",
		"port":   "1",
	}

	sampleRdmaHwStatsEntries := []rdmamap.RdmaStatEntry{
		{
			Name:  "duplicate_request",
			Value: uint64(0),
		},
		{
			Name:  "implied_nak_seq_err",
			Value: uint64(0),
		},
		{
			Name:  "lifespan",
			Value: uint64(10),
		},
		{
			Name:  "local_ack_timeout_err",
			Value: uint64(38),
		},
		{
			Name:  "np_cnp_sent",
			Value: uint64(10284520),
		},
		{
			Name:  "np_ecn_marked_roce_packets",
			Value: uint64(286733949),
		},
		{
			Name:  "out_of_buffer",
			Value: uint64(1149772),
		},
		{
			Name:  "out_of_sequence",
			Value: uint64(44),
		},
		{
			Name:  "packet_seq_err",
			Value: uint64(1),
		},
		{
			Name:  "req_cqe_error",
			Value: uint64(10776),
		},
		{
			Name:  "req_cqe_flush_error",
			Value: uint64(2173),
		},
		{
			Name:  "req_remote_access_errors",
			Value: uint64(0),
		},
		{
			Name:  "req_remote_invalid_request",
			Value: uint64(0),
		},
		{
			Name:  "resp_cqe_error",
			Value: uint64(759),
		},
		{
			Name:  "resp_cqe_flush_error",
			Value: uint64(759),
		},
		{
			Name:  "resp_local_length_error",
			Value: uint64(0),
		},
		{
			Name:  "resp_remote_access_errors",
			Value: uint64(0),
		},
		{
			Name:  "rnr_nak_retry_err",
			Value: uint64(0),
		},
		{
			Name:  "roce_adp_retrans",
			Value: uint64(0),
		},
		{
			Name:  "roce_adp_retrans_to",
			Value: uint64(0),
		},
		{
			Name:  "roce_slow_restart",
			Value: uint64(0),
		},
		{
			Name:  "roce_slow_restart_cnps",
			Value: uint64(0),
		},
		{
			Name:  "roce_slow_restart_trans",
			Value: uint64(0),
		},
		{
			Name:  "rp_cnp_handled",
			Value: uint64(1),
		},
		{
			Name:  "rp_cnp_ignored",
			Value: uint64(0),
		},
		{
			Name:  "rx_atomic_requests",
			Value: uint64(0),
		},
		{
			Name:  "rx_icrc_encapsulated",
			Value: uint64(0),
		},
		{
			Name:  "rx_read_requests",
			Value: uint64(488228),
		},
		{
			Name:  "rx_write_requests",
			Value: uint64(3928699),
		},
	}

	var acc testutil.Accumulator

	addStats("m1x5_0", "1", sampleRdmaHwStatsEntries, &acc)

	acc.AssertContainsTaggedFields(t, "infiniband", fields, tags)
}
