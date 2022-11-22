package netflow

import (
	"bytes"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/netsampler/goflow2/decoders/netflowlegacy"
)

// Decoder structure
type netflowv5Decoder struct{}

func (d *netflowv5Decoder) Init() error {
	return nil
}

func (d *netflowv5Decoder) Decode(srcIP net.IP, payload []byte) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric

	src := srcIP.String()

	// Decode the message
	buf := bytes.NewBuffer(payload)
	packet, err := netflowlegacy.DecodeMessage(buf)
	if err != nil {
		return nil, err
	}

	// Extract metrics
	switch msg := packet.(type) {
	case netflowlegacy.PacketNetFlowV5:
		t := time.Unix(int64(msg.UnixSecs), int64(msg.UnixNSecs))
		for _, record := range msg.Records {
			tags := map[string]string{
				"source":  src,
				"version": "NetFlowV5",
			}
			fields := map[string]interface{}{
				"version":           msg.Version,
				"flows":             msg.Count,
				"sys_uptime":        msg.SysUptime,
				"seq_number":        msg.FlowSequence,
				"engine_type":       mapEngineType(msg.EngineType),
				"engine_id":         msg.EngineId,
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
				"src_tos":           record.Tos,
				"bgp_src_as":        record.SrcAS,
				"bgp_dst_as":        record.DstAS,
				"src_mask":          record.SrcMask,
				"dst_mask":          record.DstMask,
			}
			metrics = append(metrics, metric.New("netflow", tags, fields, t))
		}
	}

	return metrics, nil
}
