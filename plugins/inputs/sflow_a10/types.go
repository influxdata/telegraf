package sflow_a10

import (
	"encoding/xml"
	"fmt"
	"net"
	"strings"
)

type ContainsMetricData interface {
	GetTags() map[string]string
	GetFields() map[string]interface{}
}

// V5Format answers and decoder. Directive capable of decoding sFlow v5 packets in accordance
// with SFlow_A10 v5 specification at https://sflow.org/sflow_version_5.txt
type V5Format struct {
	Version        uint32
	AgentAddress   net.IPAddr
	SubAgentID     uint32
	SequenceNumber uint32
	Uptime         uint32
	Samples        []Sample
}

type SampleType uint32

const (
	SampleTypeCounter SampleType = 2 // sflow_version_5.txt line: 1658
)

type SampleData interface{}

type Sample struct {
	SampleType        SampleType
	SampleCounterData *CounterSample
}

type CounterSample struct {
	SequenceNumber uint32
	SourceID       uint32
	CounterRecords []CounterRecord
}

type CounterFormatType uint32

type CounterRecord struct {
	CounterFormat CounterFormatType
	CounterData   *CounterData
}

type CounterData struct {
	CounterFields map[string]interface{}
}

func (c *CounterData) GetFields() map[string]interface{} {
	return c.CounterFields
}

func (c *CounterData) GetTags(ipDimensions []IPDimension, portDimensions *PortDimension) map[string]string {
	tags := make(map[string]string)

	tags["table_type"] = portDimensions.TableType
	tags["port_number"] = fmt.Sprint(portDimensions.PortNumber)
	tags["port_type"] = portDimensions.PortType
	tags["port_range_end"] = fmt.Sprint(portDimensions.PortRangeEnd)
	tags["ip_address"] = GetAllIPs(ipDimensions)
	//TODO: do we want subnet mask as tag?
	//tags["subnet_mask"] = fmt.Sprint(dimensions.IPDimensions[0].SubnetMask)

	return tags
}

// START - XML FILE DEFINITIONS
// generated from https://www.onlinetool.io/xmltogo/

type Allctrblocks struct {
	XMLName       xml.Name       `xml:"allctrblocks"`
	Text          string         `xml:",chardata"`
	Ctr           string         `xml:"ctr,attr"`
	CounterBlocks []CounterBlock `xml:"counterBlock"`
}

type CounterBlock struct {
	Text              string              `xml:",chardata"`
	MapVersion        string              `xml:"mapVersion"`
	Tag               int                 `xml:"tag"`
	CtrBlkSzMacroName string              `xml:"ctrBlkSzMacroName"`
	CtrBlkType        string              `xml:"ctrBlkType"`
	CtrBlkSz          int                 `xml:"ctrBlkSz"`
	OffsetHeaders     []HeaderDefinition  `xml:"offsetHeader"`
	Counters          []CounterDefinition `xml:"counter"`
}

type HeaderDefinition struct {
	Text      string `xml:",chardata"`
	Dtype     string `xml:"dtype"`
	FieldName string `xml:"fieldName"`
}

type CounterDefinition struct {
	Text      string `xml:",chardata"`
	Offset    int    `xml:"offset"`
	Dtype     string `xml:"dtype"`
	Dsize     string `xml:"dSize"`
	EnumName  string `xml:"enumName"`
	FieldName string `xml:"fieldName"`
}

// END - XML FILE DEFINITIONS

// Validate validates the XML file
func (a *Allctrblocks) Validate() error {
	// used to see if we have two same tags in the XML file
	tagsSet := make(map[int]interface{})
	for _, cb := range a.CounterBlocks {
		if cb.Tag < 0 {
			return fmt.Errorf("tag less than zero, value %d", cb.Tag)
		}

		if _, exists := tagsSet[cb.Tag]; exists {
			return fmt.Errorf("counterBlock with tag %d exists twice in the XML file", cb.Tag)
		}
		tagsSet[cb.Tag] = struct{}{}

		if len(cb.Counters) == 0 {
			return fmt.Errorf("0 counters is not allowed for tag %d", cb.Tag)
		}
		for i := 0; i < len(cb.Counters); i++ {
			c := cb.Counters[i]
			if c.Offset != i {
				return fmt.Errorf("offset %d in counter %s in tag %d has the wrong value", c.Offset, c.EnumName, cb.Tag)
			}
			if c.Dtype != "u8" && c.Dtype != "u16" && c.Dtype != "u32" && c.Dtype != "u64" && c.Dtype != "string" {
				return fmt.Errorf("dtype %s in counter %s in tag %d has the wrong value", c.Dtype, c.EnumName, cb.Tag)
			}
			if c.EnumName == "" && c.FieldName == "" {
				return fmt.Errorf("empty enumname && fieldname in tag %d", cb.Tag)
			}
		}
	}
	return nil
}

// DimensionsPerSourceID contains Port and IP information for each SourceID
// Port and IP information is obtained by processing counter records tagged 260 and 271/272 respectively
type DimensionsPerSourceID struct {
	PortDimensions *PortDimension
	IPDimensions   []IPDimension
}

// GetAllIPs concatenates all IPs in the DimensionsPerSourceID and returns them
func GetAllIPs(ipDimensions []IPDimension) string {
	var ips []string
	for _, ip := range ipDimensions {
		ips = append(ips, ip.IPAddress)
	}
	return strings.Join(ips, "_")
}

// PortDimension contains port information. Obtained from parsing counter record 260
type PortDimension struct {
	TableType    string
	PortNumber   int
	PortType     string
	PortRangeEnd int
}

// IPDimension contains IP information. Obtained from parsing counter record 271/272 (ipv4/ipv6)
type IPDimension struct {
	IPAddress  string
	SubnetMask uint8
}

// Validate returns true if all fields of the DimensionsPerSourceID struct are valid
func (d *DimensionsPerSourceID) Validate() error {
	if d.PortDimensions == nil {
		return fmt.Errorf("PortDimension is nil")
	} else if d.IPDimensions == nil {
		return fmt.Errorf("IPDimensions is nil")
	} else if len(d.IPDimensions) == 0 {
		return fmt.Errorf("IPDimensions has zero length")
	}
	return nil
}

func tableTypeIntToString(tableType uint8) string {
	switch tableType {
	case 0:
		return "INVALID"
	case 1:
		return "DST"
	case 5:
		return "Zone"
	default:
		return "Unknown"
	}
}

func portTypeIntToString(portType uint8) string {
	switch portType {
	case 0:
		return "INVALID"
	case 1:
		return "UDP"
	case 2:
		return "TCP"
	case 3:
		return "ICMP"
	case 4:
		return "OTHER"
	case 5:
		return "HTTP"
	case 6:
		return "DNS_TCP"
	case 7:
		return "DNS_UDP"
	case 8:
		return "SSL"
	case 9:
		return "UDP (SRC_PORT)"
	case 10:
		return "TCP (SRC_PORT)"
	case 11:
		return "SIP_TCP"
	case 12:
		return "SIP_UDP"
	case 13:
		return "QUIC"
	default:
		return "Unknown"
	}
}

// readA10XMLData parses the A10 XML definitions file and returns a map with tag as key and counter information as value
// moreover, it does some processing on the FieldName strings so they are compatible with different timeseries storage backends
func (s *SFlow_A10) readA10XMLData(data []byte) (map[uint32]CounterBlock, error) {
	var allCounterBlocks Allctrblocks
	if err := xml.Unmarshal([]byte(data), &allCounterBlocks); err != nil {
		return nil, err
	}

	err := allCounterBlocks.Validate()
	if err != nil {
		return nil, err
	}

	counterBlocks := make(map[uint32]CounterBlock)
	for i := 0; i < len(allCounterBlocks.CounterBlocks); i++ {
		counterBlock := allCounterBlocks.CounterBlocks[i]

		s.Log.Debugf("adding tag %d to the tags map", counterBlock.Tag)

		// assign it to the global map
		counterBlocks[uint32(counterBlock.Tag)] = counterBlock

		// modifying field-name so it's backend compatible (lowercase, no spaces, underscores etc.)
		for i := 0; i < len(counterBlock.Counters); i++ {
			fieldName := counterBlock.Counters[i].FieldName
			fieldName = strings.ToLower(fieldName)
			fieldName = strings.ReplaceAll(fieldName, ": ", "_")
			fieldName = strings.ReplaceAll(fieldName, " ", "_")
			fieldName = strings.ReplaceAll(fieldName, ":", "_")
			fieldName = strings.ReplaceAll(fieldName, "-", "_")
			fieldName = strings.ReplaceAll(fieldName, "/", "_")
			counterBlock.Counters[i].FieldName = fieldName
		}
	}

	return counterBlocks, nil
}
