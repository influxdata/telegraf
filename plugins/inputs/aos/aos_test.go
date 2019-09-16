package aos_test

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/aos"
	"github.com/influxdata/telegraf/plugins/inputs/aos/aos_streaming"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExtractProbeMessage(t *testing.T) {
	plugin := &aos.Aos{
		Port:            7777,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))
	val := &aos_streaming.ProbeMessage_Int64Value{Int64Value: 10}
	alert := &aos_streaming.ProbeMessage{
		Value: val,
	}
	ssl.ExtractProbeData(alert, "probe_msg_1")
	tag_value := acc.TagValue("probe_message", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "probe_msg_1", "The probe message was not added.")

}

func TestExtractAlertDataForProbeAlert(t *testing.T) {
	plugin := &aos.Aos{
		Port:            7778,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	//plugin.Accumulator = acc
	assert.NoError(t, plugin.Start(&acc))
	pi := "p"
	sn := "s"
	alert := &aos_streaming.ProbeAlert{
		ExpectedInt: new(int64),
		ActualInt:   new(int64),
		ProbeId:     &pi,
		StageName:   &sn,
		//KeyValuePairs: nil,
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
	plugin := &aos.Aos{
		Port:            7779,
		Address:         "blah",
		StreamingType:   []string{"events"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
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
	plugin := &aos.Aos{
		Port:            7780,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
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
	plugin := &aos.Aos{
		Port:            7781,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

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
	ssl.ExtractIntfData(interface_counters, "spine1")
	tag_value := acc.TagValue("interface_counters", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The perfmon data was not added.")
}

func TestExtractSystemInfo(t *testing.T) {
	plugin := &aos.Aos{
		Port:            7782,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

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
	ssl.ExtractSystemInfo(system_info, "spine1")
	tag_value := acc.TagValue("system_info", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The system info was not added.")
}

func TestExtractProcessInfo(t *testing.T) {
	plugin := &aos.Aos{
		Port:            7783,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

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
	ssl.ExtractProcessInfo(processes, "spine1")
	tag_value := acc.TagValue("process_info", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The process info was not added.")
}

func TestExtractFileInfo(t *testing.T) {
	plugin := &aos.Aos{
		Port:            7784,
		Address:         "blah",
		StreamingType:   []string{"alerts"},
		AosServer:       "blah",
		AosPort:         443,
		AosLogin:        "admin",
		AosPassword:     "admin",
		AosProtocol:     "https",
		RefreshInterval: 1000,
	}

	ssl := &aos.StreamAos{
		Listener: nil,
		Aos:      plugin,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, plugin.Start(&acc))

	file_name := "file1"
	file_size := uint64(0)
	file_info := &aos_streaming.FileInfo{
		FileName: &file_name,
		FileSize: &file_size,
	}
	files := []*aos_streaming.FileInfo{}
	files = append(files, file_info)
	ssl.ExtractFileInfo(files, "spine1")
	tag_value := acc.TagValue("file_info", "device")
	plugin.Stop()
	assert.Equal(
		t, tag_value, "spine1", "The file info was not added.")
}
