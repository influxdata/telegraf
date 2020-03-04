package lanz

import (
	"net/url"
	"strconv"
	"testing"

	pb "github.com/aristanetworks/goarista/lanz/proto"
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf/testutil"
)

var testProtoBufCongestionRecord1 = &pb.LanzRecord{
	CongestionRecord: &pb.CongestionRecord{
		Timestamp:          proto.Uint64(100000000000000),
		IntfName:           proto.String("eth1"),
		SwitchId:           proto.Uint32(1),
		PortId:             proto.Uint32(1),
		QueueSize:          proto.Uint32(1),
		EntryType:          pb.CongestionRecord_EntryType.Enum(1),
		TrafficClass:       proto.Uint32(1),
		TimeOfMaxQLen:      proto.Uint64(100000000000000),
		TxLatency:          proto.Uint32(100),
		QDropCount:         proto.Uint32(1),
		FabricPeerIntfName: proto.String("FabricPeerIntfName1"),
	},
}
var testProtoBufCongestionRecord2 = &pb.LanzRecord{
	CongestionRecord: &pb.CongestionRecord{
		Timestamp:          proto.Uint64(200000000000000),
		IntfName:           proto.String("eth2"),
		SwitchId:           proto.Uint32(2),
		PortId:             proto.Uint32(2),
		QueueSize:          proto.Uint32(2),
		EntryType:          pb.CongestionRecord_EntryType.Enum(2),
		TrafficClass:       proto.Uint32(2),
		TimeOfMaxQLen:      proto.Uint64(200000000000000),
		TxLatency:          proto.Uint32(200),
		QDropCount:         proto.Uint32(2),
		FabricPeerIntfName: proto.String("FabricPeerIntfName2"),
	},
}

func TestLanzGeneratesMetrics(t *testing.T) {

	var acc testutil.Accumulator

	l := NewLanz()

	l.Servers = append(l.Servers, "tcp://switch01.int.example.com:50001")
	l.Servers = append(l.Servers, "tcp://switch02.int.example.com:50001")

	in1 := make(chan *pb.LanzRecord)
	in2 := make(chan *pb.LanzRecord)

	deviceUrl1, err := url.Parse(l.Servers[0])
	if err != nil {
		t.Fail()
	}
	deviceUrl2, err := url.Parse(l.Servers[1])
	if err != nil {
		t.Fail()
	}
	go receive(&acc, in1, deviceUrl1)
	go receive(&acc, in2, deviceUrl2)

	in1 <- testProtoBufCongestionRecord1
	acc.Wait(1)

	vals1 := map[string]interface{}{
		"timestamp":        int64(100000000000000),
		"queue_size":       int64(1),
		"time_of_max_qlen": int64(100000000000000),
		"tx_latency":       int64(100),
		"q_drop_count":     int64(1),
	}
	tags1 := map[string]string{
		"intf_name":             "eth1",
		"switch_id":             strconv.FormatInt(int64(1), 10),
		"port_id":               strconv.FormatInt(int64(1), 10),
		"entry_type":            strconv.FormatInt(int64(1), 10),
		"traffic_class":         strconv.FormatInt(int64(1), 10),
		"fabric_peer_intf_name": "FabricPeerIntfName1",
		"source":                "switch01.int.example.com",
		"port":                  "50001",
	}

	acc.AssertContainsTaggedFields(t, "lanz_congestion_record", vals1, tags1)

	acc.ClearMetrics()
	in2 <- testProtoBufCongestionRecord2
	acc.Wait(1)

	vals2 := map[string]interface{}{
		"timestamp":        int64(200000000000000),
		"queue_size":       int64(2),
		"time_of_max_qlen": int64(200000000000000),
		"tx_latency":       int64(200),
		"q_drop_count":     int64(2),
	}
	tags2 := map[string]string{
		"intf_name":             "eth2",
		"switch_id":             strconv.FormatInt(int64(2), 10),
		"port_id":               strconv.FormatInt(int64(2), 10),
		"entry_type":            strconv.FormatInt(int64(2), 10),
		"traffic_class":         strconv.FormatInt(int64(2), 10),
		"fabric_peer_intf_name": "FabricPeerIntfName2",
		"source":                "switch02.int.example.com",
		"port":                  "50001",
	}

	acc.AssertContainsFields(t, "lanz_congestion_record", vals2)
	acc.AssertContainsTaggedFields(t, "lanz_congestion_record", vals2, tags2)

}
