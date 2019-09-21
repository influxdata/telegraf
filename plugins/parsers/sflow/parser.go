package sflow

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/influxdata/telegraf/plugins/parsers/sflow/decoder"
	"github.com/influxdata/telegraf/plugins/parsers/sflow/protodb"
)

// Parser is Telegraf parser capable of parsing an sFlow v5 network packet
type Parser struct {
	metricName   string
	defaultTags  map[string]string
	tagsAsFields map[string]bool
	sflowFormat  decoder.ItemDecoder
}

// NewParser creates a new SFlow Parser
func NewParser(metricName string, defaultTags map[string]string, sflowConfig V5FormatOptions, tagsAsFields map[string]bool) (*Parser, error) {
	if metricName == "" {
		return nil, fmt.Errorf("metric name cannot be empty")
	}
	result := &Parser{metricName: metricName, sflowFormat: V5Format(sflowConfig), tagsAsFields: tagsAsFields, defaultTags: defaultTags}
	return result, nil
}

// GetTagsAsFields answers a map of _natural_ tags by name that will actually be recored in the metrics as fields
func (sfp *Parser) GetTagsAsFields() map[string]bool {
	return sfp.tagsAsFields
}

// Parse takes a byte buffer separated by newlines
// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
// and parses it into telegraf metrics
//
// Must be thread-safe.
func (sfp *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	decodedPacket, err := decoder.Decode(sfp.sflowFormat, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	// walk the object graph and turn into Metric objects
	nano := 0
	metrics := make([]telegraf.Metric, 0)
	samples, ok := decodedPacket["samples"].([]map[string]interface{})
	if ok {
		for _, sample := range samples {
			iface, ok := sample["flowRecords"]
			if ok {
				// flowRecord
				flowRecords, ok := iface.([]map[string]interface{})
				if ok {
					for _, flowRecord := range flowRecords {
						iface, ok := flowRecord["header"]
						if ok {
							headers, ok := iface.([]map[string]interface{})
							if ok && len(headers) == 1 {
								tags := make(map[string]string)
								fields := make(map[string]interface{})
								header, ok := flowRecord["header"].([]map[string]interface{})
								if ok {
									sfp.addPotentialTagsOrFields(decodedPacket, sample, flowRecord, header, tags, fields)
									sfp.addFields(sample, flowRecord, header, fields)
								}
								m, err := metric.New(sfp.metricName, tags, fields, time.Now().Add(time.Duration(nano)))
								nano++
								if err == nil {
									metrics = append(metrics, m)
								} else {
									return nil, err
								}
							}
						}
					}
				}
			}
		}
	}

	return metrics, err
}

func (sfp *Parser) addPotentialTagsOrFields(decodedPacket map[string]interface{}, sample map[string]interface{}, flowRecord map[string]interface{}, header []map[string]interface{}, tags map[string]string, fields map[string]interface{}) {
	asTagOrField := func(name, value string, natural ...interface{}) {
		if len(natural) > 1 {
			log.Panicf("len(natural) > 1 %d", len(natural))
		}
		if asField, ok := sfp.tagsAsFields[name]; ok && asField {
			if len(natural) == 1 {
				fields[name] = natural[0]
			} else {
				fields[name] = value
			}
		} else {
			tags[name] = value
		}
	}

	// now we can pull out stuff and start appending to result slice
	at, ok := decodedPacket["addressType"].(string)
	a, ok := decodedPacket["agentAddress"].([]byte)
	if ok {
		if at == "IPV4" && len(a) == 4 {
			asTagOrField("agent_ip", fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3]))
		}
		if at == "IPV6" && len(a) == 16 {
			asTagOrField("agent_ip", fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
				a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], a[13], a[14], a[15]))
		}
	}

	v := sample["sourceIdType"]
	ui32, ok := v.(uint32)
	if ok {
		asTagOrField("source_id", fmt.Sprintf("%d", ui32), ui32)
	}

	sourceIDIndex, ok := sample["sourceIdValue"].(uint32)
	if ok {
		asTagOrField("source_id_index", fmt.Sprintf("%d", sourceIDIndex), sourceIDIndex)
	}

	ui32, ok = sample["inputValue"].(uint32)
	if ok {
		format := sample["inputFormat"].(uint32)
		if format == 0 {
			asTagOrField("netif_index_in", fmt.Sprintf("%d", ui32), ui32)
			if sourceIDIndex == ui32 {
				asTagOrField("sample_direction", "ingress")
			}
		}
	}

	ui32, ok = sample["outputValue"].(uint32)
	if ok {
		format := sample["outputFormat"].(uint32)
		if format == 0 {
			asTagOrField("netif_index_out", fmt.Sprintf("%d", ui32), ui32)
			if sourceIDIndex == ui32 {
				asTagOrField("sample_direction", "egress")
			}
		}
	}

	// go into the header itself
	if ok && len(header) == 1 {
		v := header[0]["srcIP"]
		switch t := v.(type) {
		case net.IP:
			asTagOrField("src_ip", t.String())
		case []uint8:
			b := []byte{t[0], t[1], t[2], t[3]}
			ip := net.IP{}
			ip = b
			asTagOrField("src_ip", ip.String())
		}

		v = header[0]["dstIP"]
		switch t := v.(type) {
		case net.IP:
			asTagOrField("dst_ip", t.String())
		case []uint8:
			b := []byte{t[0], t[1], t[2], t[3]}
			ip := net.IP{}
			ip = b
			asTagOrField("dst_ip", ip.String())
		}

		if header[0]["srcPort"] != nil {
			asTagOrField("src_port", fmt.Sprintf("%d", header[0]["srcPort"]), header[0]["srcPort"])
			asTagOrField("src_port_name", serviceNameFromPort(header[0]["srcPort"]), header[0]["srcPort"])
		}
		if header[0]["dstPort"] != nil {
			asTagOrField("dst_port", fmt.Sprintf("%d", header[0]["dstPort"]), header[0]["dstPort"])
			asTagOrField("dst_port_name", serviceNameFromPort(header[0]["dstPort"]))
		}

		v = header[0]["srcMac"]
		switch t := v.(type) {
		case uint64:
			asTagOrField("src_mac", toMACString(t))
		case []uint8:
			asTagOrField("src_mac", toMACString(binary.BigEndian.Uint64(append([]byte{0, 0}, t...))))
		}

		v = header[0]["dstMac"]
		switch t := v.(type) {
		case uint64:
			asTagOrField("dst_mac", toMACString(t))
		case []uint8:
			asTagOrField("dst_mac", toMACString(binary.BigEndian.Uint64(append([]byte{0, 0}, t...))))
		}

		at, ok := header[0]["nextHop.addressType"].(string)
		a, ok := header[0]["nextHop.address"].([]byte)
		if ok {
			if at == "IPV4" && len(a) == 4 {
				asTagOrField("next_hop", fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3]))
			}
			if at == "IPV6" && len(a) == 16 {
				asTagOrField("next_hop", fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
					a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], a[13], a[14], a[15]))
			}
		}

		ui32, ok = header[0]["etype"].(uint32)
		if ok {
			str, ok := etypeAsString[fmt.Sprintf("%d", ui32)]
			if ok {
				asTagOrField("ether_type", str)
			} else {

			}
		} else {

		}

		ui16, ok := header[0]["etype"].(uint16)
		if ok {
			switch ui16 {
			case 0x0800:
				asTagOrField("ether_type", "IPv4")
			case 0x86DD:
				asTagOrField("ether_type", "IPv6")
			}
		}
	}

	ui64, ok := flowRecord["srcVlan"].(uint64)
	if ok {
		asTagOrField("src_vlan", fmt.Sprintf("%d", ui64), ui64)
	}

	ui32, ok = flowRecord["srcPriority"].(uint32)
	if ok {
		asTagOrField("src_priority", fmt.Sprintf("%d", ui32), ui32)
	}

	ui64, ok = flowRecord["dstVlan"].(uint64)
	if ok {
		asTagOrField("dst_vlan", fmt.Sprintf("%d", ui64), ui64)
	}
	ui32, ok = flowRecord["dstPriority"].(uint32)
	if ok {
		asTagOrField("dst_priority", fmt.Sprintf("%d", ui32), ui32)
	}

	protocol, ok := flowRecord["protocol"].(string)
	if ok {
		asTagOrField("header_protocol", protocol)
	}

	if header[0]["dscp"] != nil {
		asTagOrField("ip_dscp", fmt.Sprintf("%d", header[0]["dscp"].(uint16)), header[0]["dscp"])
	}
	if header[0]["ecn"] != nil {
		asTagOrField("ip_ecn", fmt.Sprintf("%d", header[0]["ecn"].(uint16)), header[0]["ecn"])
	}
}

func (sfp *Parser) addFields(sample map[string]interface{}, flowRecord map[string]interface{}, header []map[string]interface{}, fields map[string]interface{}) {

	samplingRate, ok := sample["samplingRate"].(uint32)
	if ok {
		fields["packets"] = samplingRate
	}

	fields["ip_fragment_offset"] = header[0]["fragmentOffset"]

	d, ok := sample["drops"].(uint32)
	if ok {
		fields["drops"] = d
	}

	ui32, ok := flowRecord["frameLength"].(uint32)
	if ok {
		fields["frame_length"] = ui32
		fields["bytes"] = ui32 * samplingRate
	}

	ui32, ok = flowRecord["header.length"].(uint32)
	if ok {
		fields["header_length"] = ui32
	}

	v, ok := header[0]["TCPFlags"]
	if ok {
		fields["tcp_flags"] = v
	} else {
		fields["tcp_flags"] = 0
	}

	fields["ip_ttl"] = header[0]["IPTTL"]

	if header[0]["total_length"] != nil {
		fields["ip_total_length"] = header[0]["total_length"]
	}
	if header[0]["flags"] != nil {
		fields["ip_flags"] = header[0]["flags"]
	}
	if header[0]["urgent_pointer"] != nil {
		fields["tcp_urgent_pointer"] = header[0]["urgent_pointer"].(uint16)
	}

	if header[0]["tcp_header_length"] != nil {
		fields["tcp_header_length"] = header[0]["tcp_header_length"].(uint32)
	}
	if header[0]["tcp_window_size"] != nil {
		fields["tcp_window_size"] = header[0]["tcp_window_size"].(uint16)
	}
	if header[0]["udp_length"] != nil {
		fields["udp_length"] = header[0]["udp_length"].(uint16)
	}
}

// ParseLine takes a single string metric
// ie, "cpu.usage.idle 90"
// and parses it into a telegraf metric.
//
// Must be thread-safe.
func (sfp *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := sfp.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

// SetDefaultTags tells the parser to add all of the given tags
// to each parsed metric.
// NOTE: do _not_ modify the map after you've passed it here!!
func (sfp *Parser) SetDefaultTags(tags map[string]string) {
	sfp.defaultTags = tags
}

var etypeAsString = map[string]string{
	"2048": "IPv4",
}

var portNumberStrToServiceName = map[string]string{
	"22":  "ssh",
	"80":  "http",
	"443": "https",
}

func toMACString(val uint64) string {
	pair1 := val & 0x00000000000000ff
	val = val >> 8
	pair2 := val & 0x00000000000000ff
	val = val >> 8
	pair3 := val & 0x00000000000000ff
	val = val >> 8
	pair4 := val & 0x00000000000000ff
	val = val >> 8
	pair5 := val & 0x00000000000000ff
	val = val >> 8
	pair6 := val & 0x00000000000000ff
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", pair6, pair5, pair4, pair3, pair2, pair1)
}

func serviceNameFromPort(value interface{}) string {
	portNum, ok := value.(uint16)
	if ok {
		if service, ok := protodb.GetServByPort("tcp", int(portNum)); ok {
			return service
		}
	}
	return fmt.Sprintf("%v", portNum)
}
