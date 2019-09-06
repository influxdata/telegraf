package sflow

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"honnef.co/go/netdb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// SFlowParser is Telegraf parser capable of parsing an sFlow v5 network packet
type SFlowParser struct {
	metricName           string
	snmpCommunity        string
	defaultTags          map[string]string
	maxFlowsPerSample    uint32
	maxCountersPerSample uint32
	maxSamplesPerPacket  uint32
	tagsAsFields         map[string]bool
}

type SFlowParserConfig struct {
	MetricName    string
	SNMPCommunity string
	DefaultTags   map[string]string

	// Optional function to replace default DNS resolution - useful in testing
	DNSLookupFn func(ipAddress string) (string, error)

	// Optional function to replace default port->service name resolution - useful in testing
	ServiceLookupFn func(portNum int) (string, error)
}

// NewParser creats a new SFlowParser
func NewParser(metricName string, snmpCommunity string, defaultTags map[string]string, maxFlowsPerSample uint32, maxCountersPerSample uint32, maxSamplesPerPacket uint32, tagsAsFields map[string]bool) (*SFlowParser, error) {
	if metricName == "" {
		return nil, fmt.Errorf("metric name cannot be empty")
	}
	result := &SFlowParser{metricName: metricName, snmpCommunity: snmpCommunity, maxFlowsPerSample: maxFlowsPerSample, maxCountersPerSample: maxCountersPerSample, maxSamplesPerPacket: maxSamplesPerPacket, tagsAsFields: tagsAsFields}
	if defaultTags != nil {
		result.defaultTags = defaultTags
	}
	return result, nil
}

// Parse takes a byte buffer separated by newlines
// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
// and parses it into telegraf metrics
//
// Must be thread-safe.
func (sfp *SFlowParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	decodedPacket, err := Decode(SFlowFormat(sfp.maxFlowsPerSample, sfp.maxCountersPerSample, sfp.maxSamplesPerPacket), bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

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
									sfp.adPotentialTagsOrFields(decodedPacket, sample, flowRecord, header, tags, fields)
									sfp.adFields(sample, flowRecord, header, fields)
								}
								m, err := metric.New(sfp.metricName, tags, fields, time.Now().Add(time.Duration(nano)))
								nano++
								if err == nil {
									metrics = append(metrics, m)
								} else {
									// DO WHAT?
								}
							} else {
								// header isn't of right type or not right len
							}
						} else {
							// has no header, curious
						}
					}
				} else {
					// has no flowRecords within it, curioius
					fmt.Printf("Sample that is a consider a FlowRecords has no flowRecords member")
				}
			} else {
				// not a flow sample record, no worries
			}
		}
	}

	return metrics, err
}

func (sfp *SFlowParser) adPotentialTagsOrFields(decodedPacket map[string]interface{}, sample map[string]interface{}, flowRecord map[string]interface{}, header []map[string]interface{}, tags map[string]string, fields map[string]interface{}) {
	asTagOrField := func(name, value string) {
		if asField, ok := sfp.tagsAsFields[name]; ok && asField {
			fields[name] = value
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

		//tags["host"] = ipToName(tags["agent_ip"])
	}

	v := sample["sourceIdType"]
	ui32, ok := v.(uint32)
	if ok {
		asTagOrField("source_id", fmt.Sprintf("%d", ui32))
	} else {
		fmt.Println("couldn't find sourceIdType", v, ok, sample)
	}
	sourceIDIndex, ok := sample["sourceIdValue"].(uint32)
	if ok {
		asTagOrField("source_id_index", fmt.Sprintf("%d", sourceIDIndex))
		//									tags["source_id_name"] = ifIndexToIfName(sfp.snmpCommunity, tags["agent_ip"], sourceIDIndex)
	} else {
		fmt.Println("couldn't find sourceIdValue")
	}

	ui32, ok = sample["inputValue"].(uint32)
	if ok {
		// need to do some maths to extract format and value
		// most significant 2 bits are format, rest is value
		//format := ui32 >> 30
		//value := ui32 & 0x0fffffff
		format := sample["inputFormat"].(uint32)
		if format == 0 {
			asTagOrField("netif_index_in", fmt.Sprintf("%d", ui32))
			//tags["netif_name_in"] = ifIndexToIfName(sfp.snmpCommunity, tags["agent_ip"], ui32)
			if sourceIDIndex == ui32 {
				asTagOrField("sample_direction", "ingress")
			}
		} // WHAT IF SOMETHING ELSE?
	} else {
		// WHAT IF EXPANDED FORMAT - should probbaly do this processing in decoder to they are normalized
		fmt.Println("couldn't find inputValue") // questions this from Rob!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! formt or value
	}

	ui32, ok = sample["outputValue"].(uint32)
	if ok {
		format := sample["outputFormat"].(uint32)
		// need to do some maths to extract format and value
		// most significant 2 bits are format, rest is value
		//format := ui32 >> 30
		//value := ui32 & 0x0fffffff
		if format == 0 {
			asTagOrField("netif_index_out", fmt.Sprintf("%d", ui32))
			//tags["netif_name_out"] = ifIndexToIfName(sfp.snmpCommunity, tags["agent_ip"], ui32)
			if sourceIDIndex == ui32 {
				asTagOrField("sample_direction", "egress")
			}
		} // WHAT IF SOMETHING ELSE?
	} else {
		fmt.Println("couldn't find outputValue")
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

		//tags["src_ip"] = header[0]["srcIP"].(net.IP).String()
		//tags["src_host"] = ipToName(tags["src_ip"])
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
		//tags["dst_ip"] = header[0]["dstIP"].(net.IP).String()
		//tags["dst_host"] = ipToName(tags["dst_ip"])
		if header[0]["srcPort"] != nil {
			asTagOrField("src_port", fmt.Sprintf("%d", header[0]["srcPort"]))
			asTagOrField("src_port_name", serviceNameFromPort(header[0]["srcPort"]))
		}
		if header[0]["dstPort"] != nil {
			asTagOrField("dst_port", fmt.Sprintf("%d", header[0]["dstPort"]))
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

		// tag ip_protocol = proto
		// tag ip_version = IPVersion

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
				tags["ether_type"] = str
			} else {

			}
		} else {

		}
		// NEW HEADER
		ui16, ok := header[0]["etype"].(uint16)
		if ok {
			switch ui16 {
			case 0x0800:
				asTagOrField("ether_type", "IPv4")
			case 0x86DD:
				asTagOrField("ether_type", "IPv6")
			}
		}
	} else {

	}

	ui64, ok := flowRecord["srcVlan"].(uint64)
	if ok {
		asTagOrField("src_vlan", fmt.Sprintf("%d", ui64))
	}

	//addFieldUint64(taf,flowRecord,"srcVlan","src_vlan")

	ui32, ok = flowRecord["srcPriority"].(uint32)
	if ok {
		asTagOrField("src_priority", fmt.Sprintf("%d", ui32))
	}

	ui64, ok = flowRecord["dstVlan"].(uint64)
	if ok {
		asTagOrField("dst_vlan", fmt.Sprintf("%d", ui64))
	}
	ui32, ok = flowRecord["dstPriority"].(uint32)
	if ok {
		asTagOrField("dst_priority", fmt.Sprintf("%d", ui32))
	}

	// ui32("srcMaskLen"),
	// ui32("dstMaskLen"),

	//header_protocol
	protocol, ok := flowRecord["protocol"].(string)
	if ok {
		asTagOrField("header_protocol", protocol)
	}

	if header[0]["dscp"] != nil {
		tags["ip_dscp"] = fmt.Sprintf("%d", header[0]["dscp"].(uint16))
	}
	if header[0]["ecn"] != nil {
		tags["ip_ecn"] = fmt.Sprintf("%d", header[0]["ecn"].(uint16))
	}
}

func (sfp *SFlowParser) adFields(sample map[string]interface{}, flowRecord map[string]interface{}, header []map[string]interface{}, fields map[string]interface{}) {

	samplingRate, ok := sample["samplingRate"].(uint32)

	if ok {
		fields["packets"] = samplingRate
	}

	fields["ip_fragment_offset"] = header[0]["fragmentOffset"]

	// ingress or egress
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
		fields["tcp_flags"] = 0 // this is filling a gap in the tests, don't think it is really required
	}

	fields["ip_ttl"] = header[0]["IPTTL"]

	if header[0]["total_length"] != nil {
		fields["ip_total_length"] = header[0]["total_length"] //.(uint32)
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

/*
func (sfp *SFlowParser) asTagOrField(name, value string, tags map[string]string, fields map[string]interface{}) {
	if asField, ok := sfp.tagsAsFields[name]; ok && asField {
		fields[name] = value
	} else {
		tags[name] = value
	}
}
*/

// ParseLine takes a single string metric
// ie, "cpu.usage.idle 90"
// and parses it into a telegraf metric.
//
// Must be thread-safe.
func (sfp *SFlowParser) ParseLine(line string) (telegraf.Metric, error) {
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
func (sfp *SFlowParser) SetDefaultTags(tags map[string]string) {
	sfp.defaultTags = tags
}

/* MOVE ALL THESE TO sFLow */
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
	if value == nil {
		return "nil"
	}
	var portNum int
	switch t := value.(type) {
	case uint32:
		portNum = int(t)
	case uint16:
		portNum = int(t)
	default:
		portNum = -1
	}

	if portNum >= 0 {
		proto := netdb.GetProtoByName("tcp")
		serv := netdb.GetServByPort(portNum, proto)
		if serv != nil {
			//if serv.Name != "" {
			return fmt.Sprintf("%s", serv.Name)
			/*} else {
				log.Printf("E! [parsers.sflow] example of serv.Name = `` %d\n", serv.Port)
				return fmt.Sprintf("%v", value)
			}*/
		}
	}
	return fmt.Sprintf("%v", value)
}
