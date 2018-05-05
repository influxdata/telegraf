package lanz_consumer

import (
	"fmt"
	pb "github.com/aristanetworks/goarista/lanz/proto"
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf/testutil"
	"strconv"
	"testing"
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

func TestLanzConsumerGeneratesMetrics(t *testing.T) {

	var acc testutil.Accumulator

	l := NewLanzConsumer()
	l.Clients = make(map[string]*LanzClient)

	c1 := NewLanzClient()
	c1.Server = "tcp://switch01.int.example.com:50001"
	c1.in = make(chan *pb.LanzRecord)
	c1.done = make(chan bool)
	c1.acc = &acc
	l.Clients[c1.Server] = c1

	c2 := NewLanzClient()
	c2.Server = "tcp://switch02.int.example.com:50001"
	c2.in = make(chan *pb.LanzRecord)
	c2.done = make(chan bool)
	c2.acc = &acc
	l.Clients[c2.Server] = c2

	c1.Lock()
	defer c1.Unlock()
	go c1.receiver()

	c2.Lock()
	defer c2.Unlock()
	go c2.receiver()

	c1.in <- testProtoBufCongestionRecord1

	vals1 := map[string]interface{}{
		"timestamp":     int64(100000000000000),
		"queueSize":     int64(1),
		"timeOfMaxQLen": int64(100000000000000),
		"txLatency":     int64(100),
		"qDropCount":    int64(1),
	}
	tags1 := map[string]string{
		"intfName":           "eth1",
		"switchId":           strconv.FormatInt(int64(1), 10),
		"portId":             strconv.FormatInt(int64(1), 10),
		"entryType":          strconv.FormatInt(int64(1), 10),
		"trafficClass":       strconv.FormatInt(int64(1), 10),
		"fabricPeerIntfName": "FabricPeerIntfName1",
		"host":               "switch01.int.example.com:50001",
	}

	acc.AssertContainsTaggedFields(t, "congestionRecord", vals1, tags1)

	c2.in <- testProtoBufCongestionRecord2

	vals2 := map[string]interface{}{
		"timestamp":     int64(200000000000000),
		"queueSize":     int64(2),
		"timeOfMaxQLen": int64(200000000000000),
		"txLatency":     int64(200),
		"qDropCount":    int64(2),
	}
	tags2 := map[string]string{
		"intfName":           "eth2",
		"switchId":           strconv.FormatInt(int64(2), 10),
		"portId":             strconv.FormatInt(int64(2), 10),
		"entryType":          strconv.FormatInt(int64(2), 10),
		"trafficClass":       strconv.FormatInt(int64(2), 10),
		"fabricPeerIntfName": "FabricPeerIntfName2",
		"host":               "switch02.int.example.com:50001",
	}

	fmt.Println(len(acc.Metrics))
	acc.AssertContainsTaggedFields(t, "congestionRecord", vals2, tags2)

	c1.done <- true
	c2.done <- true

}
