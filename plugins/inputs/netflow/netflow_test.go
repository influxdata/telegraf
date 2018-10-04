package netflow

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

func newNetflowV9() (*Netflow, chan inData) {
	in := make(chan inData, 1500)
	netflow := &Netflow{
		ServiceAddress:        ":2055",
		in:                    in,
		done:                  make(chan struct{}),
		v9WriteTemplate:       make(chan *V9TemplateWriteOp),
		v9WriteOptionTemplate: make(chan *V9OptionTemplateWriteOp),
		v9ReadTemplate:        make(chan *V9TemplateReadOp),
		v9ReadFlowField:       make(chan *V9FlowFieldReadOp),
	}
	return netflow, in
}

var testTemplateFlowset = []byte{0, 9, 0, 1, 0, 0, 216, 95, 87, 147, 40, 50, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 40, 1, 0, 0, 8, 0, 8, 0, 4, 0, 12, 0, 4, 0, 95, 0, 4, 0, 7, 0, 2, 0, 11, 0, 2, 0, 4, 0, 1, 0, 1, 0, 4, 0, 2, 0, 4}
var testDataFlowset = []byte{0, 9, 0, 1, 0, 0, 228, 12, 87, 147, 40, 53, 0, 0, 0, 3, 0, 0, 0, 0, 1, 0, 0, 29, 100, 1, 1, 1, 200, 1, 1, 1, 3, 0, 0, 179, 109, 163, 0, 179, 6, 0, 0, 0, 99, 0, 0, 0, 2}

func TestV9Parser(t *testing.T) {
	netflow, in := newNetflowV9()
	acc := testutil.Accumulator{}
	netflow.acc = &acc
	defer close(netflow.done)

	netflow.wg.Add(3)
	go netflow.netflowParser()
	go netflow.v9TemplatePoller()
	go netflow.v9FlowFieldPoller()

	time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testTemplateFlowset}
	//time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testDataFlowset}
	time.Sleep(time.Millisecond * 1000)

	netflow.Gather(&acc)
	fields := map[string]interface{}{
		"application_id":             int64(50331827),
		"transport_source_port":      int64(28067),
		"transport_destination_port": int64(179),
		"ip_protocol":                int64(6),
		"counter_bytes":              int64(99),
		"counter_packets":            int64(2),
		"ipv4_source_address":      "100.1.1.1",
		"ipv4_destination_address": "200.1.1.1",
	}
	tags := map[string]string{
		"exporter":                 "192.168.0.1",
	}
	m, _ := acc.Get("netflow")
	log.Printf("tags: %v", m.Tags)
	log.Printf("fields: %v", m.Fields)
	acc.AssertContainsTaggedFields(t, "netflow", fields, tags)
}

func newNetflowV9ForApplication() (*Netflow, chan inData) {
	in := make(chan inData, 1500)
	netflow := &Netflow{
		ServiceAddress:        ":2055",
		in:                    in,
		done:                  make(chan struct{}),
		v9WriteTemplate:       make(chan *V9TemplateWriteOp),
		v9WriteOptionTemplate: make(chan *V9OptionTemplateWriteOp),
		v9ReadTemplate:        make(chan *V9TemplateReadOp),
		v9ReadFlowField:       make(chan *V9FlowFieldReadOp),

		readApplication:            make(chan *ApplicationReadOp),
		ResolveApplicationNameByID: true,
	}
	return netflow, in
}

func TestResolveApplicationNameById(t *testing.T) {
	netflow, in := newNetflowV9ForApplication()
	acc := testutil.Accumulator{}
	netflow.acc = &acc
	defer close(netflow.done)

	netflow.wg.Add(4)
	go netflow.netflowParser()
	go netflow.v9TemplatePoller()
	go netflow.v9FlowFieldPoller()
	go netflow.applicationPoller()

	time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testTemplateFlowset}
	//time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testDataFlowset}
	time.Sleep(time.Millisecond * 1000)

	netflow.Gather(&acc)
	fields := map[string]interface{}{
		"application_id":             int64(50331827),
		"transport_source_port":      int64(28067),
		"transport_destination_port": int64(179),
		"ip_protocol":                int64(6),
		"counter_bytes":              int64(99),
		"counter_packets":            int64(2),
		"ipv4_source_address":      "100.1.1.1",
		"ipv4_destination_address": "200.1.1.1",
	}
	tags := map[string]string{
		"application_name":         "bgp",
		"exporter":                 "192.168.0.1",
	}
	m, _ := acc.Get("netflow")
	log.Printf("tags: %v", m.Tags)
	log.Printf("fields: %v", m.Fields)
	acc.AssertContainsTaggedFields(t, "netflow", fields, tags)
}

var testOptionTemplateFlowsetAndRecords = []byte{0, 9, 0, 5, 0, 95, 215, 188, 87, 149, 225, 8, 0, 0, 5, 82, 0, 0, 0, 0, 0, 1, 0, 26, 1, 0, 0, 4, 0, 12, 0, 1, 0, 4, 0, 10, 0, 4, 0, 82, 0, 32, 0, 83, 0, 64, 1, 0, 1, 164, 10, 71, 134, 65, 0, 0, 0, 1, 69, 116, 48, 47, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 71, 134, 65, 0, 0, 0, 2, 69, 116, 48, 47, 49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 71, 134, 65, 0, 0, 0, 3, 69, 116, 48, 47, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 71, 134, 65, 0, 0, 0, 4, 69, 116, 48, 47, 51, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 51, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var testTemplateFlowsetAndDataFlowset = []byte{0, 9, 0, 2, 0, 0, 131, 28, 87, 151, 21, 16, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 40, 1, 2, 0, 8, 0, 8, 0, 4, 0, 12, 0, 4, 0, 10, 0, 4, 0, 7, 0, 2, 0, 11, 0, 2, 0, 4, 0, 1, 0, 1, 0, 4, 0, 2, 0, 4, 1, 2, 0, 29, 192, 168, 1, 1, 192, 168, 1, 254, 0, 0, 0, 1, 0, 0, 0, 0, 88, 0, 0, 1, 164, 0, 0, 0, 6}

func newNetflowV9ForIfname() (*Netflow, chan inData) {
	in := make(chan inData, 1500)
	netflow := &Netflow{
		ServiceAddress:        ":2055",
		in:                    in,
		done:                  make(chan struct{}),
		v9WriteTemplate:       make(chan *V9TemplateWriteOp),
		v9WriteOptionTemplate: make(chan *V9OptionTemplateWriteOp),
		v9ReadTemplate:        make(chan *V9TemplateReadOp),
		v9ReadFlowField:       make(chan *V9FlowFieldReadOp),

		writeIfname:            make(chan *IfnameWriteOp),
		readIfname:             make(chan *IfnameReadOp),
		ResolveIfnameByIfindex: true,
	}
	return netflow, in
}

func TestResolveIfnameByIfindex(t *testing.T) {
	netflow, in := newNetflowV9ForIfname()
	acc := testutil.Accumulator{}
	netflow.acc = &acc
	defer close(netflow.done)

	netflow.wg.Add(4)
	go netflow.netflowParser()
	go netflow.v9TemplatePoller()
	go netflow.v9FlowFieldPoller()
	go netflow.ifnamePoller()

	time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testOptionTemplateFlowsetAndRecords}
	//time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testTemplateFlowsetAndDataFlowset}
	time.Sleep(time.Millisecond * 1000)

	netflow.Gather(&acc)
	b := []byte{69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} // "Ethernet0/0"
	fields := map[string]interface{}{
		"transport_source_port":      int64(0),
		"transport_destination_port": int64(0),
		"ip_protocol":                int64(88),
		"counter_bytes":              int64(420),
		"counter_packets":            int64(6),
		"interface_input_snmp":       int64(1),
		"ipv4_source_address":      "192.168.1.1",
		"ipv4_destination_address": "192.168.1.254",
	}
	tags := map[string]string{
		"exporter":                 "192.168.0.1",
		"interface_input_name":       string(b),
	}
	m, _ := acc.Get("netflow")
	log.Printf("tags: %v", m.Tags)
	log.Printf("fields: %v", m.Fields)
	acc.AssertContainsTaggedFields(t, "netflow", fields, tags)
}

var testIpfixTemplateFlowset = []byte{0, 10, 0, 56, 87, 175, 11, 114, 0, 0, 0, 7, 0, 0, 0, 0, 0, 2, 0, 40, 1, 0, 0, 8, 0, 8, 0, 4, 0, 12, 0, 4, 0, 95, 0, 4, 0, 7, 0, 2, 0, 11, 0, 2, 0, 4, 0, 1, 0, 1, 0, 4, 0, 2, 0, 4}
var testIpfixDataFlowset = []byte{0, 10, 0, 45, 87, 175, 11, 122, 0, 0, 0, 7, 0, 0, 0, 0, 1, 0, 0, 29, 100, 1, 1, 1, 200, 1, 1, 1, 13, 0, 0, 1, 87, 41, 0, 179, 6, 0, 0, 0, 44, 0, 0, 0, 1}

func newTestNetflowIpfix() (*Netflow, chan inData) {
	in := make(chan inData, 1500)
	netflow := &Netflow{
		ServiceAddress:              ":2055",
		in:                          in,
		done:                        make(chan struct{}),
		ipfixWriteTemplate:          make(chan *IpfixTemplateWriteOp),
		ipfixWriteOptionTemplate:    make(chan *IpfixOptionTemplateWriteOp),
		ipfixReadTemplate:           make(chan *IpfixTemplateReadOp),
		ipfixReadInformationElement: make(chan *IpfixInformationElementReadOp),
	}
	return netflow, in
}

func TestIpfixParser(t *testing.T) {
	netflow, in := newTestNetflowIpfix()
	acc := testutil.Accumulator{}
	netflow.acc = &acc
	defer close(netflow.done)

	netflow.wg.Add(3)
	go netflow.netflowParser()
	go netflow.ipfixTemplatePoller()
	go netflow.ipfixInformationElementPoller()

	time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testIpfixTemplateFlowset}
	//time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testIpfixDataFlowset}
	time.Sleep(time.Millisecond * 1000)

	netflow.Gather(&acc)
	fields := map[string]interface{}{
		"transport_source_port":      int64(22313),
		"transport_destination_port": int64(179),
		"ip_protocol":                int64(6),
		"counter_bytes":              int64(44),
		"counter_packets":            int64(1),
		"application_id":             int64(218103809),
		"ipv4_source_address":        "100.1.1.1",
		"ipv4_destination_address":   "200.1.1.1",
	}
	tags := map[string]string{
		"exporter": "192.168.0.1",
	}
	acc.AssertContainsTaggedFields(t, "netflow", fields, tags)
}

func newNetflowIpfixForApplication() (*Netflow, chan inData) {
	in := make(chan inData, 1500)
	netflow := &Netflow{
		ServiceAddress:              ":2055",
		in:                          in,
		done:                        make(chan struct{}),
		ipfixWriteTemplate:          make(chan *IpfixTemplateWriteOp),
		ipfixWriteOptionTemplate:    make(chan *IpfixOptionTemplateWriteOp),
		ipfixReadTemplate:           make(chan *IpfixTemplateReadOp),
		ipfixReadInformationElement: make(chan *IpfixInformationElementReadOp),
		readApplication:             make(chan *ApplicationReadOp),
		ResolveApplicationNameByID:  true,
	}
	return netflow, in
}

func TestIpfixResolveApplicationNameById(t *testing.T) {
	netflow, in := newNetflowIpfixForApplication()
	acc := testutil.Accumulator{}
	netflow.acc = &acc
	defer close(netflow.done)

	netflow.wg.Add(4)
	go netflow.netflowParser()
	go netflow.ipfixTemplatePoller()
	go netflow.ipfixInformationElementPoller()
	go netflow.applicationPoller()

	time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testIpfixTemplateFlowset}
	//time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testIpfixDataFlowset}
	time.Sleep(time.Millisecond * 1000)

	netflow.Gather(&acc)
	fields := map[string]interface{}{
		"transport_source_port":      int64(22313),
		"transport_destination_port": int64(179),
		"ip_protocol":                int64(6),
		"counter_bytes":              int64(44),
		"counter_packets":            int64(1),
		"application_id":             int64(218103809),
		"ipv4_source_address":        "100.1.1.1",
		"ipv4_destination_address":   "200.1.1.1",
	}
	tags := map[string]string{
		"application_name": "unknown",
		"exporter":         "192.168.0.1",
	}
	acc.AssertContainsTaggedFields(t, "netflow", fields, tags)
}

var testIpfixOptionTemplateSetAndRecords = []byte{0, 10, 1, 186, 87, 174, 196, 110, 0, 2, 11, 254, 0, 0, 0, 0, 0, 3, 0, 22, 1, 0, 0, 3, 0, 1, 0, 10, 0, 4, 0, 82, 0, 32, 0, 83, 0, 64, 1, 0, 1, 148, 0, 0, 0, 1, 69, 116, 48, 47, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 69, 116, 48, 47, 49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 69, 116, 48, 47, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 69, 116, 48, 47, 51, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 51, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var testIpfixTemplateSetAndDataSet = []byte{0, 10, 0, 85, 87, 174, 214, 117, 0, 0, 0, 10, 0, 0, 0, 0, 0, 2, 0, 40, 1, 1, 0, 8, 0, 8, 0, 4, 0, 12, 0, 4, 0, 10, 0, 4, 0, 7, 0, 2, 0, 11, 0, 2, 0, 4, 0, 1, 0, 1, 0, 4, 0, 2, 0, 4, 1, 1, 0, 29, 100, 1, 1, 1, 200, 1, 1, 1, 0, 0, 0, 1, 84, 168, 0, 179, 6, 0, 0, 0, 44, 0, 0, 0, 1}

func newNetflowIpfixForIfname() (*Netflow, chan inData) {
	in := make(chan inData, 1500)
	netflow := &Netflow{
		ServiceAddress:              ":2055",
		in:                          in,
		done:                        make(chan struct{}),
		ipfixWriteTemplate:          make(chan *IpfixTemplateWriteOp),
		ipfixWriteOptionTemplate:    make(chan *IpfixOptionTemplateWriteOp),
		ipfixReadTemplate:           make(chan *IpfixTemplateReadOp),
		ipfixReadInformationElement: make(chan *IpfixInformationElementReadOp),

		writeIfname:            make(chan *IfnameWriteOp),
		readIfname:             make(chan *IfnameReadOp),
		ResolveIfnameByIfindex: true,
	}
	return netflow, in
}

func TestIpfixResolveIfnameByIfindex(t *testing.T) {
	netflow, in := newNetflowIpfixForIfname()
	acc := testutil.Accumulator{}
	netflow.acc = &acc
	defer close(netflow.done)

	netflow.wg.Add(4)
	go netflow.netflowParser()
	go netflow.ipfixTemplatePoller()
	go netflow.ipfixInformationElementPoller()
	go netflow.ifnamePoller()

	time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testIpfixOptionTemplateSetAndRecords}
	//time.Sleep(time.Millisecond * 1000)
	in <- inData{remote: &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 65535}, data: testIpfixTemplateSetAndDataSet}
	time.Sleep(time.Millisecond * 1000)

	netflow.Gather(&acc)
	b := []byte{69, 116, 104, 101, 114, 110, 101, 116, 48, 47, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} // "Ethernet0/0"
	fields := map[string]interface{}{
		"transport_source_port":      int64(21672),
		"transport_destination_port": int64(179),
		"ip_protocol":                int64(6),
		"counter_bytes":              int64(44),
		"counter_packets":            int64(1),
		"interface_input_snmp":       int64(1),
		"ipv4_source_address":        "100.1.1.1",
		"ipv4_destination_address":   "200.1.1.1",
	}
	tags := map[string]string{
		"exporter":             "192.168.0.1",
		"interface_input_name": string(b),
	}
	log.Printf("%v", acc.Metrics)
	m, _ := acc.Get("netflow")
	log.Printf("tags: %v", m.Tags)
	log.Printf("fields: %v", m.Fields)
	acc.AssertContainsTaggedFields(t, "netflow", fields, tags)
}
