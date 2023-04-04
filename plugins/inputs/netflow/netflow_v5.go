package netflow

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/netsampler/goflow2/decoders/netflowlegacy"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Decoder structure
type netflowv5Decoder struct{}

func (d *netflowv5Decoder) Init() error {
	if err := initL4ProtoMapping(); err != nil {
		return fmt.Errorf("initializing layer 4 protocol mapping failed: %w", err)
	}
	return nil
}

func (d *netflowv5Decoder) Decode(srcIP net.IP, payload []byte) ([]telegraf.Metric, error) {
	src := srcIP.String()

	// Decode the message
	buf := bytes.NewBuffer(payload)
	packet, err := netflowlegacy.DecodeMessage(buf)
	if err != nil {
		return nil, err
	}

	// Extract metrics
	msg, ok := packet.(netflowlegacy.PacketNetFlowV5)
	if !ok {
		return nil, fmt.Errorf("unexpected message type %T", packet)
	}

	t := time.Unix(int64(msg.UnixSecs), int64(msg.UnixNSecs))
	metrics := make([]telegraf.Metric, 0, len(msg.Records))
	for _, record := range msg.Records {
		tags := map[string]string{
			"source":  src,
			"version": "NetFlowV5",
		}
		fields := map[string]interface{}{
			"flows":             msg.Count,
			"sys_uptime":        msg.SysUptime,
			"seq_number":        msg.FlowSequence,
			"engine_type":       mapEngineType(msg.EngineType),
			"engine_id":         decodeHex([]byte{msg.EngineId}),
			"sampling_interval": msg.SamplingInterval,
			"src":               decodeIPFromUint32(record.SrcAddr),
			"dst":               decodeIPFromUint32(record.DstAddr),
			"next_hop":          decodeIPFromUint32(record.NextHop),
			"in_snmp":           record.Input,
			"out_snmp":          record.Output,
			"in_packets":        record.DPkts,
			"in_bytes":          record.DOctets,
			"first_switched":    record.First,
			"last_switched":     record.Last,
			"src_port":          record.SrcPort,
			"dst_port":          record.DstPort,
			"tcp_flags":         mapTCPFlags(record.TCPFlags),
			"protocol":          mapL4Proto(record.Proto),
			"src_tos":           decodeHex([]byte{record.Tos}),
			"bgp_src_as":        record.SrcAS,
			"bgp_dst_as":        record.DstAS,
			"src_mask":          record.SrcMask,
			"dst_mask":          record.DstMask,
		}
		metrics = append(metrics, metric.New("netflow", tags, fields, t))
	}

	return metrics, nil
}
