package aos

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf/plugins/inputs/aos/aos_streaming"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var PORT uint16 = 9999

func TestExtractProbeMessage(t *testing.T) {
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"perfmon"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "probe_msg_1"

	probe_id := "1234-5678"
	stage_name := "test_stage"
	blueprint_id := "test_blueprint"
	item_id := "test_route"
	probe_label := "probe_msg_label"
	probe_value := int64(10)

	perfmon := &aos_streaming.ProbeMessage{
		Value:       &aos_streaming.ProbeMessage_Int64Value{probe_value},
		ProbeId:     &probe_id,
		StageName:   &stage_name,
		BlueprintId: &blueprint_id,
		ItemId:      &item_id,
		ProbeLabel:  &probe_label,
	}
	ssl.ExtractProbeData(perfmon, tags)
	tag_value := acc.TagValue("probe_message", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "probe_msg_1", "The probe message was not added.")
}

func TestExtractAlertDataForProbeAlert(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	//plugin.Accumulator = acc
	assert.NoError(t, plugin.Start(&acc))
	pi := "p"
	sn := "s"
	item_id := "1234"
	alert := &aos_streaming.ProbeAlert{
		ExpectedInt: new(int64),
		ActualInt:   new(int64),
		ProbeId:     &pi,
		StageName:   &sn,
		ItemId:      &item_id,
	}
	tags := make(map[string]string)
	tags["device"] = "probe_alert"
	ssl.ExtractAlertData("probe_alert", tags, alert, true)
	tag_value := acc.TagValue("alert_probe", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "probe_alert", "The alert message was not added.")
}

func TestExtractEventDataForStreamingEvent(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"events"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))
	aoshost := "AOSHOST"
	streaming_type := aos_streaming.StreamingType_STREAMING_TYPE_PERFMON
	streaming_protocol := aos_streaming.StreamingProtocol_STREAMING_PROTOCOL_PROTOBUF_OVER_TCP
	streaming_status := aos_streaming.StreamingStatus_STREAMING_STATUS_UP
	event := &aos_streaming.StreamingEvent{
		AosServer:     &aoshost,
		StreamingType: &streaming_type,
		Protocol:      &streaming_protocol,
		Status:        &streaming_status,
	}
	tags := make(map[string]string)
	tags["device"] = "10.1.252.130:7777"
	ssl.ExtractEventData("streaming", tags, event)
	tag_value := acc.TagValue("event_streaming", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "10.1.252.130:7777", "The event message was not added.")
}

func TestExtractAlertDataForStreamingAlert(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))
	aoshost := "AOSHOST"
	streaming_type := aos_streaming.StreamingType_STREAMING_TYPE_PERFMON
	streaming_protocol := aos_streaming.StreamingProtocol_STREAMING_PROTOCOL_PROTOBUF_OVER_TCP
	streaming_alert_reason := aos_streaming.StreamingAlertReason_STREAMING_ALERT_REASON_FAILED_CONNECTION
	alert := &aos_streaming.StreamingAlert{
		AosServer:     &aoshost,
		StreamingType: &streaming_type,
		Protocol:      &streaming_protocol,
		Reason:        &streaming_alert_reason,
	}
	tags := make(map[string]string)
	tags["device"] = "10.1.252.130:7777"
	ssl.ExtractAlertData("streaming_alert", tags, alert, true)
	tag_value := acc.TagValue("alert_streaming", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "10.1.252.130:7777", "The alert message was not added.")
}

func TestExtractIntfDataForPerfmon(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "spine1"

	tx_unicast_packets := uint64(10)
	tx_broadcast_packets := uint64(10)
	tx_multicast_packets := uint64(10)
	tx_bytes := uint64(10)
	rx_unicast_packets := uint64(5)
	rx_broadcast_packets := uint64(5)
	rx_multicast_packets := uint64(5)
	rx_bytes := uint64(10)
	tx_error_packets := uint64(0)
	tx_discard_packets := uint64(0)
	rx_error_packets := uint64(0)
	rx_discard_packets := uint64(0)
	alignment_errors := uint64(0)
	fcs_errors := uint64(0)
	symbol_errors := uint64(0)
	runts := uint64(0)
	giants := uint64(0)
	delta_seconds := uint64(5)
	interface_counters := &aos_streaming.InterfaceCounters{
		TxUnicastPackets:   &tx_unicast_packets,
		TxBroadcastPackets: &tx_broadcast_packets,
		TxMulticastPackets: &tx_multicast_packets,
		TxBytes:            &tx_bytes,
		RxUnicastPackets:   &rx_unicast_packets,
		RxBroadcastPackets: &rx_broadcast_packets,
		RxMulticastPackets: &rx_multicast_packets,
		RxBytes:            &rx_bytes,
		TxErrorPackets:     &tx_error_packets,
		TxDiscardPackets:   &tx_discard_packets,
		RxErrorPackets:     &rx_error_packets,
		RxDiscardPackets:   &rx_discard_packets,
		AlignmentErrors:    &alignment_errors,
		FcsErrors:          &fcs_errors,
		SymbolErrors:       &symbol_errors,
		Runts:              &runts,
		Giants:             &giants,
		DeltaSeconds:       &delta_seconds,
	}
	ssl.ExtractIntfData(interface_counters, tags)
	tag_value := acc.TagValue("interface_counters", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The perfmon data was not added.")
}

func TestExtractIntfDataWithDefaultDeltaSeconds(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "spine1"

	tx_unicast_packets := uint64(10)
	tx_broadcast_packets := uint64(10)
	tx_multicast_packets := uint64(10)
	tx_bytes := uint64(10)
	rx_unicast_packets := uint64(5)
	rx_broadcast_packets := uint64(5)
	rx_multicast_packets := uint64(5)
	rx_bytes := uint64(10)
	tx_error_packets := uint64(0)
	tx_discard_packets := uint64(0)
	rx_error_packets := uint64(0)
	rx_discard_packets := uint64(0)
	alignment_errors := uint64(0)
	fcs_errors := uint64(0)
	symbol_errors := uint64(0)
	runts := uint64(0)
	giants := uint64(0)
	interface_counters := &aos_streaming.InterfaceCounters{
		TxUnicastPackets:   &tx_unicast_packets,
		TxBroadcastPackets: &tx_broadcast_packets,
		TxMulticastPackets: &tx_multicast_packets,
		TxBytes:            &tx_bytes,
		RxUnicastPackets:   &rx_unicast_packets,
		RxBroadcastPackets: &rx_broadcast_packets,
		RxMulticastPackets: &rx_multicast_packets,
		RxBytes:            &rx_bytes,
		TxErrorPackets:     &tx_error_packets,
		TxDiscardPackets:   &tx_discard_packets,
		RxErrorPackets:     &rx_error_packets,
		RxDiscardPackets:   &rx_discard_packets,
		AlignmentErrors:    &alignment_errors,
		FcsErrors:          &fcs_errors,
		SymbolErrors:       &symbol_errors,
		Runts:              &runts,
		Giants:             &giants,
	}
	ssl.ExtractIntfData(interface_counters, tags)
	tag_value := acc.TagValue("interface_counters", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The perfmon data was not added.")
}

func TestExtractIntfDataWithZeroDeltaSeconds(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "spine1"

	tx_unicast_packets := uint64(10)
	tx_broadcast_packets := uint64(10)
	tx_multicast_packets := uint64(10)
	tx_bytes := uint64(10)
	rx_unicast_packets := uint64(5)
	rx_broadcast_packets := uint64(5)
	rx_multicast_packets := uint64(5)
	rx_bytes := uint64(10)
	tx_error_packets := uint64(0)
	tx_discard_packets := uint64(0)
	rx_error_packets := uint64(0)
	rx_discard_packets := uint64(0)
	alignment_errors := uint64(0)
	fcs_errors := uint64(0)
	symbol_errors := uint64(0)
	runts := uint64(0)
	giants := uint64(0)
	delta_seconds := uint64(0)
	interface_counters := &aos_streaming.InterfaceCounters{
		TxUnicastPackets:   &tx_unicast_packets,
		TxBroadcastPackets: &tx_broadcast_packets,
		TxMulticastPackets: &tx_multicast_packets,
		TxBytes:            &tx_bytes,
		RxUnicastPackets:   &rx_unicast_packets,
		RxBroadcastPackets: &rx_broadcast_packets,
		RxMulticastPackets: &rx_multicast_packets,
		RxBytes:            &rx_bytes,
		TxErrorPackets:     &tx_error_packets,
		TxDiscardPackets:   &tx_discard_packets,
		RxErrorPackets:     &rx_error_packets,
		RxDiscardPackets:   &rx_discard_packets,
		AlignmentErrors:    &alignment_errors,
		FcsErrors:          &fcs_errors,
		SymbolErrors:       &symbol_errors,
		Runts:              &runts,
		Giants:             &giants,
		DeltaSeconds:       &delta_seconds,
	}
	ssl.ExtractIntfData(interface_counters, tags)
	fields := make(map[string]interface{})
	plugin.Stop()
	acc.AssertDoesNotContainsTaggedFields(t, "interface_counters", fields, tags)
}

func TestExtractSystemInfo(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "spine1"

	cpu_user := float32(0)
	cpu_system := float32(0)
	cpu_idle := float32(0)
	memory_used := uint64(0)
	memory_total := uint64(0)
	system_info := &aos_streaming.SystemInfo{
		CpuUser:     &cpu_user,
		CpuSystem:   &cpu_system,
		CpuIdle:     &cpu_idle,
		MemoryUsed:  &memory_used,
		MemoryTotal: &memory_total,
	}
	ssl.ExtractSystemInfo(system_info, tags)
	tag_value := acc.TagValue("system_info", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The system info was not added.")
}

func TestExtractProcessInfo(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "spine1"

	process_name := "proc1"
	cpu_user := float32(0)
	cpu_system := float32(0)
	memory_used := uint64(0)
	process_info := &aos_streaming.ProcessInfo{
		ProcessName: &process_name,
		CpuUser:     &cpu_user,
		CpuSystem:   &cpu_system,
		MemoryUsed:  &memory_used,
	}
	processes := []*aos_streaming.ProcessInfo{}
	processes = append(processes, process_info)
	ssl.ExtractProcessInfo(processes, tags)
	tag_value := acc.TagValue("process_info", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The process info was not added.")
}

func TestExtractFileInfo(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["device"] = "spine1"

	file_name := "file1"
	file_size := uint64(0)
	file_info := &aos_streaming.FileInfo{
		FileName: &file_name,
		FileSize: &file_size,
	}
	files := []*aos_streaming.FileInfo{}
	files = append(files, file_info)
	ssl.ExtractFileInfo(files, tags)
	tag_value := acc.TagValue("file_info", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The file info was not added.")
}

func TestPerfmonOverTCP(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"perfmon"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	perfmon := CreatePerfmonMsg(uint64(99), uint64(5))
	socket := fmt.Sprintf("127.0.0.1:%v", PORT)
	conn, err := net.Dial("tcp", socket)
	assert.NoError(t, err)
	_, err = conn.Write(perfmon)
	assert.NoError(t, err)

	tags := make(map[string]string)
	tags["blueprint"] = "dev-testing"
	tags["device_name"] = "spine99"
	tags["device"] = "spine99"
	tags["role"] = "spine"
	tags["device_key"] = "52540077DE72"

	fields := map[string]interface{}{
		"tx_bytes":             uint64(10),
		"tx_unicast_packets":   uint64(10),
		"tx_broadcast_packets": uint64(10),
		"tx_multicast_packets": uint64(10),
		"tx_error_packets":     uint64(0),
		"tx_discard_packets":   uint64(0),
		"rx_bytes":             uint64(10),
		"rx_unicast_packets":   uint64(5),
		"rx_broadcast_packets": uint64(5),
		"rx_multicast_packets": uint64(5),
		"rx_error_packets":     uint64(0),
		"rx_discard_packets":   uint64(0),
		"alignment_errors":     uint64(0),
		"fcs_errors":           uint64(0),
		"symbol_errors":        uint64(0),
		"runts":                uint64(0),
		"giants":               uint64(0),
		"delta_seconds":        uint64(5),
		"tx_bps":               uint64(16),
		"tx_unicast_pps":       uint64(2),
		"tx_broadcast_pps":     uint64(2),
		"tx_multicast_pps":     uint64(2),
		"tx_error_pps":         uint64(0),
		"tx_discard_pps":       uint64(0),
		"rx_bps":               uint64(16),
		"rx_unicast_pps":       uint64(1),
		"rx_broadcast_pps":     uint64(1),
		"rx_multicast_pps":     uint64(1),
		"rx_error_pps":         uint64(0),
		"rx_discard_pps":       uint64(0),
	}

	acc.Wait(2)
	acc.AssertContainsTaggedFields(t, "interface_counters", fields, tags)
}

func TestEventOverTCP(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"event"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// streaming event
	aoshost := "AOSHOST"
	streaming_type := aos_streaming.StreamingType_STREAMING_TYPE_EVENTS
	streaming_protocol := aos_streaming.StreamingProtocol_STREAMING_PROTOCOL_PROTOBUF_OVER_TCP
	streaming_status := aos_streaming.StreamingStatus_STREAMING_STATUS_DOWN
	streaming_event := &aos_streaming.StreamingEvent{
		AosServer:     &aoshost,
		StreamingType: &streaming_type,
		Protocol:      &streaming_protocol,
		Status:        &streaming_status,
	}

	// event
	event_id := "1234"
	event := &aos_streaming.Event{
		Id:   &event_id,
		Data: &aos_streaming.Event_Streaming{Streaming: streaming_event},
	}

	// aos message
	now := time.Now()
	secs := now.Unix()
	timestamp := uint64(secs)
	origin_name := "52540077DE72"
	origin_hostname := "spine99"
	origin_label := "dev-testing"
	origin_role := "spine"
	aos_message := &aos_streaming.AosMessage{
		Timestamp:      &timestamp,
		OriginName:     &origin_name,
		OriginHostname: &origin_hostname,
		BlueprintLabel: &origin_label,
		OriginRole:     &origin_role,
		Data:           &aos_streaming.AosMessage_Event{Event: event},
	}

	data, err := proto.Marshal(aos_message)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	// aos sequenced message
	sequence_number := uint64(1)
	aos_sequenced_message := &aos_streaming.AosSequencedMessage{
		SeqNum:   &sequence_number,
		AosProto: data,
	}

	seq_msg_data, err := proto.Marshal(aos_sequenced_message)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	msg_size := uint16(len(seq_msg_data))
	size_buff := make([]byte, 2)
	binary.BigEndian.PutUint16(size_buff, msg_size)
	send_buff := append(size_buff, seq_msg_data...)

	socket := fmt.Sprintf("127.0.0.1:%v", PORT)
	conn, err := net.Dial("tcp", socket)
	assert.NoError(t, err)
	_, err = conn.Write(send_buff)
	assert.NoError(t, err)

	tags := make(map[string]string)
	tags["blueprint"] = "dev-testing"
	tags["device_name"] = "spine99"
	tags["device"] = "spine99"
	tags["role"] = "spine"
	tags["device_key"] = "52540077DE72"
	tags["aos_server"] = "AOSHOST"
	tags["streaming_type"] = "STREAMING_TYPE_EVENTS"
	tags["protocol"] = "STREAMING_PROTOCOL_PROTOBUF_OVER_TCP"
	tags["status"] = "STREAMING_STATUS_DOWN"

	fields := map[string]interface{}{
		"event": 1,
	}

	acc.Wait(2)
	acc.AssertContainsTaggedFields(t, "event_streaming", fields, tags)
}

func CreatePerfmonMsg(seq_num uint64, delta_seconds uint64) []byte {
	tx_unicast_packets := uint64(10)
	tx_broadcast_packets := uint64(10)
	tx_multicast_packets := uint64(10)
	tx_bytes := uint64(10)
	rx_unicast_packets := uint64(5)
	rx_broadcast_packets := uint64(5)
	rx_multicast_packets := uint64(5)
	rx_bytes := uint64(10)
	tx_error_packets := uint64(0)
	tx_discard_packets := uint64(0)
	rx_error_packets := uint64(0)
	rx_discard_packets := uint64(0)
	alignment_errors := uint64(0)
	fcs_errors := uint64(0)
	symbol_errors := uint64(0)
	runts := uint64(0)
	giants := uint64(0)
	interface_counters := &aos_streaming.InterfaceCounters{
		TxUnicastPackets:   &tx_unicast_packets,
		TxBroadcastPackets: &tx_broadcast_packets,
		TxMulticastPackets: &tx_multicast_packets,
		TxBytes:            &tx_bytes,
		RxUnicastPackets:   &rx_unicast_packets,
		RxBroadcastPackets: &rx_broadcast_packets,
		RxMulticastPackets: &rx_multicast_packets,
		RxBytes:            &rx_bytes,
		TxErrorPackets:     &tx_error_packets,
		TxDiscardPackets:   &tx_discard_packets,
		RxErrorPackets:     &rx_error_packets,
		RxDiscardPackets:   &rx_discard_packets,
		AlignmentErrors:    &alignment_errors,
		FcsErrors:          &fcs_errors,
		SymbolErrors:       &symbol_errors,
		Runts:              &runts,
		Giants:             &giants,
		DeltaSeconds:       &delta_seconds,
	}

	// perfmon
	perfmon := &aos_streaming.PerfMon{
		Data: &aos_streaming.PerfMon_InterfaceCounters{InterfaceCounters: interface_counters},
	}

	// aos message
	now := time.Now()
	secs := now.Unix()
	timestamp := uint64(secs)
	origin_name := "52540077DE72"
	origin_hostname := "spine99"
	blueprint_label := "dev-testing"
	origin_role := "spine"
	aos_message := &aos_streaming.AosMessage{
		Timestamp:      &timestamp,
		OriginName:     &origin_name,
		OriginHostname: &origin_hostname,
		BlueprintLabel: &blueprint_label,
		OriginRole:     &origin_role,
		Data:           &aos_streaming.AosMessage_PerfMon{PerfMon: perfmon},
	}

	data, err := proto.Marshal(aos_message)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	// aos sequenced message
	aos_sequenced_message := &aos_streaming.AosSequencedMessage{
		SeqNum:   &seq_num,
		AosProto: data,
	}

	seq_msg_data, err := proto.Marshal(aos_sequenced_message)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	msg_size := uint16(len(seq_msg_data))
	size_buff := make([]byte, 2)
	binary.BigEndian.PutUint16(size_buff, msg_size)
	send_buff := append(size_buff, seq_msg_data...)
	return send_buff
}

func TestReportMessageLoss(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"perfmon"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["message_type"] = "perfmon"

	ssl.reportMessageLoss("perfmon", 1, 10)

	fields := make(map[string]interface{})
	fields["perfmon"] = uint64(9)
	acc.AssertContainsTaggedFields(t, "message_loss", fields, tags)
}

func TestReportMessageLossOnReset(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"perfmon"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	tags := make(map[string]string)
	tags["message_type"] = "perfmon"

	ssl.reportMessageLoss("perfmon", 5, 0)

	fields := make(map[string]interface{})
	acc.AssertDoesNotContainsTaggedFields(t, "message_loss", fields, tags)
}

func TestSequencedVersion(t *testing.T) {
	fakeAosServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/version" {
				_, _ = w.Write([]byte(aosVersionSequencedJSON))
			} else if r.URL.Path == "/api/user/login" {
				w.WriteHeader(201)
			} else if r.URL.Path == "/api/streaming-config" {
				w.WriteHeader(201)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	defer fakeAosServer.Close()
	log.Printf("D! Fake AOS server listening on %v", fakeAosServer.URL)

	url_params := strings.Split(fakeAosServer.URL, ":")
	ip_addr := strings.Split(url_params[1], "//")
	aos_server := ip_addr[1]
	aos_port, _ := strconv.Atoi(url_params[2])

	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       aos_server,
		AosPort:         aos_port,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "http",
		RefreshInterval: 1000,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, plugin.Start(&acc))
}

func TestUnsequencedVersion(t *testing.T) {
	fakeAosServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/version" {
				_, _ = w.Write([]byte(aosVersionUnsequencedJSON))
			} else if r.URL.Path == "/api/blueprints" {
				_, _ = w.Write([]byte(blueprintJSON))
			} else if r.URL.Path == "/api/blueprints/rack-based-blueprint-a3a1802f/qe" {
				_, _ = w.Write([]byte(blueprintSystemsJSON))
			} else if r.URL.Path == "/api/systems" {
				_, _ = w.Write([]byte(systemsJSON))
			} else if r.URL.Path == "/api/user/login" {
				w.WriteHeader(201)
			} else if r.URL.Path == "/api/streaming-config" {
				w.WriteHeader(201)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	defer fakeAosServer.Close()
	log.Printf("D! Fake AOS server listening on %v", fakeAosServer.URL)

	url_params := strings.Split(fakeAosServer.URL, ":")
	ip_addr := strings.Split(url_params[1], "//")
	aos_server := ip_addr[1]
	aos_port, _ := strconv.Atoi(url_params[2])

	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"alerts"},
		AosServer:       aos_server,
		AosPort:         aos_port,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "http",
		RefreshInterval: 1000,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, plugin.Start(&acc))
}

func TestProbeMessageOverTCP(t *testing.T) {
	PORT++
	plugin := &Aos{
		Port:            PORT,
		Address:         "127.0.0.1",
		StreamingType:   []string{"perfmon"},
		AosServer:       "127.0.0.1",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	acc := testutil.Accumulator{}
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	perfmon := CreateEvpnType3ProbeMessage()
	socket := fmt.Sprintf("127.0.0.1:%v", PORT)
	conn, err := net.Dial("tcp", socket)
	assert.NoError(t, err)
	_, err = conn.Write(perfmon)
	assert.NoError(t, err)

	tags := make(map[string]string)
	tags["blueprint"] = "dev-testing"
	tags["device_name"] = "spine99"
	tags["device"] = "spine99"
	tags["role"] = "spine"
	tags["device_key"] = "52540077DE72"

	probe_id := "1234-5678"
	stage_name := "test_stage"
	blueprint_id := "test_blueprint"
	item_id := "test_route"
	probe_label := "test_probe_label"

	fields := make(map[string]interface{})
	fields["probe_label"] = probe_label
	fields["blueprint_id"] = blueprint_id
	fields["item_id"] = item_id
	fields["probe_id"] = probe_id
	fields["stage_name"] = stage_name
	fields["property"] = "[name:\"property_name\" value:\"property_value\" ]"
	fields["value"] = "&{state:ROUTE_ADD system_id:\"1234567\" vni:5 next_hop:\"10.1.1.1\" rd:\"100\" rt:\"100\" }"

	acc.Wait(2)
	acc.AssertContainsTaggedFields(t, "probe_message", fields, tags)
}

func CreateEvpnType3ProbeMessage() []byte {
	state := aos_streaming.RouteState_ROUTE_ADD
	system_id := "1234567"
	vni := uint32(5)
	next_hop := "10.1.1.1"
	rd := "100"
	rt := "100"
	evpn_type_3_event := &aos_streaming.EvpnType3RouteEvent{
		State:    &state,
		SystemId: &system_id,
		Vni:      &vni,
		NextHop:  &next_hop,
		Rd:       &rd,
		Rt:       &rt,
	}

	property_name := "property_name"
	property_value := "property_value"
	probe_property := aos_streaming.ProbeProperty{
		Name:  &property_name,
		Value: &property_value,
	}

	probe_id := "1234-5678"
	stage_name := "test_stage"
	blueprint_id := "test_blueprint"
	item_id := "test_route"
	probe_label := "test_probe_label"
	probe_message := &aos_streaming.ProbeMessage{
		Property:    []*aos_streaming.ProbeProperty{&probe_property},
		Value:       &aos_streaming.ProbeMessage_EvpnType3RouteState{evpn_type_3_event},
		ProbeId:     &probe_id,
		StageName:   &stage_name,
		BlueprintId: &blueprint_id,
		ItemId:      &item_id,
		ProbeLabel:  &probe_label,
	}

	// perfmon
	perfmon := &aos_streaming.PerfMon{
		Data: &aos_streaming.PerfMon_ProbeMessage{
			ProbeMessage: probe_message,
		},
	}

	// aos message
	now := time.Now()
	secs := now.Unix()
	timestamp := uint64(secs)
	origin_name := "52540077DE72"
	origin_hostname := "spine99"
	origin_label := "dev-testing"
	origin_role := "spine"
	aos_message := &aos_streaming.AosMessage{
		Timestamp:      &timestamp,
		OriginName:     &origin_name,
		OriginHostname: &origin_hostname,
		BlueprintLabel: &origin_label,
		OriginRole:     &origin_role,
		Data:           &aos_streaming.AosMessage_PerfMon{perfmon},
	}

	data, err := proto.Marshal(aos_message)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	// aos sequenced message
	sequence_number := uint64(99)
	aos_sequenced_message := &aos_streaming.AosSequencedMessage{
		SeqNum:   &sequence_number,
		AosProto: data,
	}

	seq_msg_data, err := proto.Marshal(aos_sequenced_message)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	msg_size := uint16(len(seq_msg_data))
	fmt.Printf("length of message %d\n", msg_size)

	size_buff := make([]byte, 2)
	binary.BigEndian.PutUint16(size_buff, msg_size)

	send_buff := append(size_buff, seq_msg_data...)
	return send_buff
}

const aosVersionSequencedJSON = `
{
  "major": "3",
  "version": "3.2.0",
  "build": "0",
  "minor": "2"
}`

const aosVersionUnsequencedJSON = `
{
  "major": "2",
  "version": "2.3.1",
  "build": "1",
  "minor": "3"
}`

const blueprintJSON = `
{
  "items": [
    {
      "top_level_root_cause_count": 0,
      "spine_count": 3,
      "anomaly_counts": {
        "arp": 0,
        "mlag": 0,
        "probe": 0,
        "hostname": 0,
        "streaming": 0,
        "series": 0,
        "cabling": 3,
        "route": 3,
        "counter": 0,
        "all": 12,
        "bgp": 6,
        "blueprint_rendering": 0,
        "mac": 0,
        "headroom": 0,
        "deployment": 0,
        "interface": 0,
        "liveness": 0,
        "config": 0,
        "lag": 0
      },
      "last_modified_at": "2019-09-20T15:46:55.789657Z",
      "label": "rack-based-blueprint-a3a1802f",
      "version": 38,
      "design": "two_stage_l3clos",
      "leaf_count": 3,
      "root_cause_count": 0,
      "external_router_count": 1,
      "l2_server_count": 0,
      "l3_server_count": 3,
      "id": "rack-based-blueprint-a3a1802f",
      "superspine_count": 0,
      "deployment_status": {
        "service_config": {
          "num_succeeded": 9,
          "num_failed": 0,
          "num_pending": 0
        },
        "discovery2_config": {
          "num_succeeded": 0,
          "num_failed": 0,
          "num_pending": 0
        }
      }
    }
  ]
}`

const blueprintSystemsJSON = `
{
  "count": 10,
  "items": [
    {
      "system_node": {
        "tags": null,
        "position_data": {
          "position": 2,
          "region": 0,
          "plane": 0,
          "pod": 0
        },
        "property_set": null,
        "hostname": "spine3",
        "group_label": null,
        "label": "spine3",
        "role": "spine",
        "system_type": "switch",
        "deploy_mode": "deploy",
        "system_id": "525400B7D07B",
        "type": "system",
        "id": "761158a4-e1fe-46c1-8fd4-4bfaf444cb96"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": "racktype-1-002-leaf1",
        "group_label": "leaf1",
        "label": "racktype_1_002_leaf1",
        "role": "leaf",
        "system_type": "switch",
        "deploy_mode": "deploy",
        "system_id": "5054004D3E93",
        "type": "system",
        "id": "19eda7c6-05c1-42ee-9d0b-c46a5019c21d"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": "racktype-1-001-leaf1",
        "group_label": "leaf1",
        "label": "racktype_1_001_leaf1",
        "role": "leaf",
        "system_type": "switch",
        "deploy_mode": "deploy",
        "system_id": "525400F70835",
        "type": "system",
        "id": "ff05bd03-681d-4b32-928b-80c78272296a"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": "racktype-1-003-server001",
        "group_label": "server1",
        "label": "racktype_1_003_server001",
        "role": "l3_server",
        "system_type": "server",
        "deploy_mode": "deploy",
        "system_id": "525400736842",
        "type": "system",
        "id": "29b3a7a9-4813-4ce5-bdfc-5809e8cd55b9"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": "racktype-1-001-server001",
        "group_label": "server1",
        "label": "racktype_1_001_server001",
        "role": "l3_server",
        "system_type": "server",
        "deploy_mode": "deploy",
        "system_id": "525400CA9587",
        "type": "system",
        "id": "9b93fc4a-0a0d-4e2a-b356-9e720ba3392e"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": "racktype-1-003-leaf1",
        "group_label": "leaf1",
        "label": "racktype_1_003_leaf1",
        "role": "leaf",
        "system_type": "switch",
        "deploy_mode": "deploy",
        "system_id": "525400D22438",
        "type": "system",
        "id": "fa042132-5fd0-4a7b-a72f-a7ecb3b945f2"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": "racktype-1-002-server001",
        "group_label": "server1",
        "label": "racktype_1_002_server001",
        "role": "l3_server",
        "system_type": "server",
        "deploy_mode": "deploy",
        "system_id": "52540077DE72",
        "type": "system",
        "id": "053d3bfa-9af8-45c8-beb3-389ccf8930f7"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": {
          "position": 0,
          "region": 0,
          "plane": 0,
          "pod": 0
        },
        "property_set": null,
        "hostname": "spine1",
        "group_label": null,
        "label": "spine1",
        "role": "spine",
        "system_type": "switch",
        "deploy_mode": "deploy",
        "system_id": "52540078FB3B",
        "type": "system",
        "id": "5dd25fca-0fee-4045-ac46-a30f2d1308bf"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": null,
        "property_set": null,
        "hostname": null,
        "group_label": null,
        "label": "ext_router_ad3ae11b",
        "role": "external_router",
        "system_type": "switch",
        "deploy_mode": null,
        "system_id": null,
        "type": "system",
        "id": "f57345ef-bd12-477b-b954-1abfa40c8b5a"
      }
    },
    {
      "system_node": {
        "tags": null,
        "position_data": {
          "position": 1,
          "region": 0,
          "plane": 0,
          "pod": 0
        },
        "property_set": null,
        "hostname": "spine2",
        "group_label": null,
        "label": "spine2",
        "role": "spine",
        "system_type": "switch",
        "deploy_mode": "deploy",
        "system_id": "505400CF4CB9",
        "type": "system",
        "id": "e93ea3b7-c80e-47cd-8aa3-f69830b3c294"
      }
    }
  ]
}`

const systemsJSON = `
{
  "items": [
    {
      "device_key": "5054004D3E93",
      "facts": {
        "aos_hcl_model": "Arista_vEOS",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "50:54:00:e1:47:39-50:54:00:e1:47:39",
        "hw_model": "vEOS",
        "hw_version": "",
        "mgmt_ifname": "Management1",
        "mgmt_ipaddr": "172.20.32.11",
        "mgmt_macaddr": "50:54:00:E1:47:39",
        "os_arch": "x86_64",
        "os_family": "EOS",
        "os_version": "4.20.11M",
        "os_version_info": {
          "build": "11M",
          "major": "4",
          "minor": "20"
        },
        "serial_number": "505400E14739",
        "vendor": "Arista"
      },
      "id": "5054004D3E93",
      "status": {
        "agent_start_time": "2019-08-19T15:41:34.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-08-19T15:25:56.950331Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "racktype-1-002-leaf1",
        "hostname": "racktype-1-002-leaf1",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Arista_vEOS",
        "location": "some_location"
      }
    },
    {
      "device_key": "525400736842",
      "facts": {
        "aos_hcl_model": "Generic_Server_1RU_1x1G",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "",
        "hw_model": "Generic Model",
        "hw_version": "",
        "mgmt_ifname": "eth0",
        "mgmt_ipaddr": "172.20.32.7",
        "mgmt_macaddr": "52:54:00:73:68:42",
        "os_arch": "x86_64",
        "os_family": "Ubuntu GNU/Linux",
        "os_version": "16.04 LTS",
        "os_version_info": {
          "build": "",
          "major": "16",
          "minor": "04"
        },
        "serial_number": "525400736842",
        "vendor": "Generic Manufacturer"
      },
      "id": "525400736842",
      "status": {
        "agent_start_time": "2019-09-20T15:19:02.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-09-20T15:18:48.580655Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "racktype-1-003-server001",
        "hostname": "racktype-1-003-server001",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Generic_Server_1RU_2x10G",
        "location": "some_location"
      }
    },
    {
      "device_key": "525400F70835",
      "facts": {
        "aos_hcl_model": "Cisco_NXOSv",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "52:54:00:f7:08:35-52:54:00:f7:08:bc",
        "hw_model": "NX-OSv",
        "hw_version": "0.0",
        "mgmt_ifname": "mgmt0",
        "mgmt_ipaddr": "172.20.32.10",
        "mgmt_macaddr": "52:54:00:F7:08:35",
        "os_arch": "x86_64",
        "os_family": "NXOS",
        "os_version": "7.0(3)I7(4)",
        "os_version_info": {
          "build": "(3)I7(4)",
          "major": "7",
          "minor": "0"
        },
        "serial_number": "525400F70835",
        "vendor": "Cisco"
      },
      "id": "525400F70835",
      "status": {
        "agent_start_time": "2019-08-30T20:31:21.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-08-30T20:26:47.911768Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "racktype-1-001-leaf1",
        "hostname": "racktype-1-001-leaf1",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Cisco_NXOSv",
        "location": "some_location"
      }
    },
    {
      "device_key": "525400CA9587",
      "facts": {
        "aos_hcl_model": "Generic_Server_1RU_1x1G",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "",
        "hw_model": "Generic Model",
        "hw_version": "",
        "mgmt_ifname": "eth0",
        "mgmt_ipaddr": "172.20.32.8",
        "mgmt_macaddr": "52:54:00:ca:95:87",
        "os_arch": "x86_64",
        "os_family": "Ubuntu GNU/Linux",
        "os_version": "16.04 LTS",
        "os_version_info": {
          "build": "",
          "major": "16",
          "minor": "04"
        },
        "serial_number": "525400CA9587",
        "vendor": "Generic Manufacturer"
      },
      "id": "525400CA9587",
      "status": {
        "agent_start_time": "2019-09-20T15:23:23.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-09-20T15:23:09.840520Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "racktype-1-001-server001",
        "hostname": "racktype-1-001-server001",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Generic_Server_1RU_2x10G",
        "location": "some_location"
      }
    },
    {
      "device_key": "52540077DE72",
      "facts": {
        "aos_hcl_model": "Generic_Server_1RU_1x1G",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "",
        "hw_model": "Generic Model",
        "hw_version": "",
        "mgmt_ifname": "eth0",
        "mgmt_ipaddr": "172.20.32.6",
        "mgmt_macaddr": "52:54:00:77:de:72",
        "os_arch": "x86_64",
        "os_family": "Ubuntu GNU/Linux",
        "os_version": "16.04 LTS",
        "os_version_info": {
          "build": "",
          "major": "16",
          "minor": "04"
        },
        "serial_number": "52540077DE72",
        "vendor": "Generic Manufacturer"
      },
      "id": "52540077DE72",
      "status": {
        "agent_start_time": "2019-09-20T15:22:23.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-09-20T15:22:04.931116Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "racktype-1-002-server001",
        "hostname": "racktype-1-002-server001",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Generic_Server_1RU_2x10G",
        "location": "some_location"
      }
    },
    {
      "device_key": "525400B7D07B",
      "facts": {
        "aos_hcl_model": "Cisco_NXOSv",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "52:54:00:b7:d0:7b-52:54:00:b7:d1:02",
        "hw_model": "NX-OSv",
        "hw_version": "0.0",
        "mgmt_ifname": "mgmt0",
        "mgmt_ipaddr": "172.20.32.13",
        "mgmt_macaddr": "52:54:00:B7:D0:7B",
        "os_arch": "x86_64",
        "os_family": "NXOS",
        "os_version": "7.0(3)I7(4)",
        "os_version_info": {
          "build": "(3)I7(4)",
          "major": "7",
          "minor": "0"
        },
        "serial_number": "525400B7D07B",
        "vendor": "Cisco"
      },
      "id": "525400B7D07B",
      "status": {
        "agent_start_time": "2019-09-28T02:51:13.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-09-28T02:46:45.374478Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "spine3",
        "hostname": "spine3",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Cisco_NXOSv",
        "location": "some_location"
      }
    },
    {
      "device_key": "505400CF4CB9",
      "facts": {
        "aos_hcl_model": "Arista_vEOS",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "50:54:00:5d:69:b6-50:54:00:5d:69:b6",
        "hw_model": "vEOS",
        "hw_version": "",
        "mgmt_ifname": "Management1",
        "mgmt_ipaddr": "172.20.32.14",
        "mgmt_macaddr": "50:54:00:5D:69:B6",
        "os_arch": "x86_64",
        "os_family": "EOS",
        "os_version": "4.20.11M",
        "os_version_info": {
          "build": "11M",
          "major": "4",
          "minor": "20"
        },
        "serial_number": "5054005D69B6",
        "vendor": "Arista"
      },
      "id": "505400CF4CB9",
      "status": {
        "agent_start_time": "2019-08-19T15:42:01.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-08-19T15:26:44.138617Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "spine2",
        "hostname": "spine2",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Arista_vEOS",
        "location": "some_location"
      }
    },
    {
      "device_key": "52540078FB3B",
      "facts": {
        "aos_hcl_model": "Cumulus_VX",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "52:54:00:78:fb:3b-52:54:00:78:fb:43",
        "hw_model": "VX",
        "hw_version": "3",
        "mgmt_ifname": "eth0",
        "mgmt_ipaddr": "172.20.32.12",
        "mgmt_macaddr": "52:54:00:78:FB:3B",
        "os_arch": "x86_64",
        "os_family": "Cumulus",
        "os_version": "3.7.5",
        "os_version_info": {
          "build": "5",
          "major": "3",
          "minor": "7"
        },
        "serial_number": "52540078FB3B",
        "vendor": "Cumulus"
      },
      "id": "52540078FB3B",
      "status": {
        "agent_start_time": "2019-08-19T15:36:51.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-08-19T15:26:07.398068Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "spine1",
        "hostname": "spine1",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Cumulus_VX",
        "location": "some_location"
      }
    },
    {
      "device_key": "525400D22438",
      "facts": {
        "aos_hcl_model": "Cumulus_VX",
        "aos_server": "172.20.32.3",
        "aos_version": "AOS_3.1.0_OB.181",
        "chassis_mac_ranges": "52:54:00:d2:24:38-52:54:00:d2:24:40",
        "hw_model": "VX",
        "hw_version": "3",
        "mgmt_ifname": "eth0",
        "mgmt_ipaddr": "172.20.32.15",
        "mgmt_macaddr": "52:54:00:D2:24:38",
        "os_arch": "x86_64",
        "os_family": "Cumulus",
        "os_version": "3.7.5",
        "os_version_info": {
          "build": "5",
          "major": "3",
          "minor": "7"
        },
        "serial_number": "525400D22438",
        "vendor": "Cumulus"
      },
      "id": "525400D22438",
      "status": {
        "agent_start_time": "2019-08-19T15:36:53.000000Z",
        "blueprint_active": true,
        "blueprint_id": "rack-based-blueprint-a3a1802f",
        "comm_state": "on",
        "device_start_time": "2019-08-19T15:26:46.033647Z",
        "domain_name": "",
        "error_message": "",
        "fqdn": "racktype-1-003-leaf1",
        "hostname": "racktype-1-003-leaf1",
        "is_acknowledged": true,
        "operation_mode": "full_control",
        "pool_id": "default_pool",
        "state": "IS-ACTIVE"
      },
      "user_config": {
        "admin_state": "normal",
        "aos_hcl_model": "Cumulus_VX",
        "location": "some_location"
      }
    }
  ]
}`
