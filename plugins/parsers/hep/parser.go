package hep

import (
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/json"
)

var headerNames = map[string]int{"version": 1,
	"protocol":  2,
	"src_ip4":   3,
	"dest_ip4":  4,
	"src_ip6":   5,
	"dest_ip6":  6,
	"src_port":  7,
	"dest_port": 8,
	"tsec":      9,
	//"tmsec":      10,
	"proto_type": 11,
	"node_id":    12,
	"mode_pw":    14,
	"cid":        17,
	"vlan":       18,
	"nodename":   19,
}

var headerReverseMap = map[int]string{1: "version",
	2: "protocol",
	3: "src_ip4",
	4: "dest_ip4",
	5: "src_ip6",
	6: "dest_ip6",
	7: "src_port",
	8: "dest_port",
	9: "tsec",
	//10: "tmsec",
	11: "proto_type",
	12: "node_id",
	14: "node_pw",
	17: "cid",
	18: "vlan",
	19: "nodename",
}

type Parser struct {
	//Json fields
	MetricName         string
	HepMeasurementName string
	TagKeys            []string
	JSONNameKey        string
	JSONStringFields   []string
	JSONQuery          string
	JSONTimeKey        string
	JSONTimeFormat     string
	JSONTimezone       string
	DefaultTags        map[string]string
	HepHeader          []string
}

// DecodeHEP returns a parsed HEP message
func DecodeHEP(packet []byte) (*HEP, error) {
	hep := &HEP{}
	err := hep.parse(packet)
	if err != nil {
		return nil, err
	}
	return hep, nil
}

func (h *HEP) parse(packet []byte) error {
	var err error
	err = h.parseHEP(packet)
	if err != nil {
		return err
	}
	t := time.Now()
	if h.ProtoType == 0 {
		return nil
	}

	h.Timestamp = time.Unix(int64(h.Tsec), int64(h.Tmsec*1000))
	d := t.Sub(h.Timestamp)
	if d < 0 || (h.Tsec == 0 && h.Tmsec == 0) {
		h.Timestamp = t
	}

	if h.NodeName == "" {
		h.NodeName = strconv.FormatUint(uint64(h.NodeID), 10)
	}
	return nil
}

func (h *Parser) Parse(packet []byte) ([]telegraf.Metric, error) {

	hep, err := DecodeHEP(packet)

	if err != nil {
		return nil, err
	}
	if len(h.HepMeasurementName) != 0 {
		h.MetricName = h.HepMeasurementName
	} else {
		h.MetricName = "hep"
	}
	headerTags := make(map[string]string)
	jsonParser, err := json.New(
		&json.Config{
			MetricName:   h.MetricName,
			TagKeys:      h.TagKeys,
			NameKey:      h.JSONNameKey,
			StringFields: h.JSONStringFields,
			Query:        h.JSONQuery,
			TimeKey:      h.JSONTimeKey,
			TimeFormat:   h.JSONTimeFormat,
			Timezone:     h.JSONTimezone,
			DefaultTags:  h.DefaultTags,
		},
	)

	if len(h.HepHeader) != 0 {
		var headerArray []int
		for _, v := range h.HepHeader {
			headerArray = append(headerArray, headerNames[v])
			headerTags = h.addHeaders(headerArray, hep)
		}
		headerTags = h.addHeaders(headerArray, hep)
	} else {
		var headerArray []int
		for k := range headerReverseMap {
			headerArray = append(headerArray, k)
			headerTags = h.addHeaders(headerArray, hep)
		}
	}

	if hep.ProtoType >= 2 && hep.Payload != "" && hep.ProtoType != 100 {
		m, err := jsonParser.Parse([]byte(hep.Payload))
		if err != nil {
			return nil, err
		}
		metric := m[0]
		for k, v := range headerTags {
			metric.AddTag(k, v)
		}
		metrics := make([]telegraf.Metric, 0)
		metrics = append(metrics, metric)
		return metrics, nil
	}

	metrics := make([]telegraf.Metric, 0)
	nFields := make(map[string]interface{})
	nFields["protocol_type_field"] = hep.ProtoType
	metric, err := metric.New(h.MetricName, headerTags, nFields, time.Now())
	metrics = append(metrics, metric)
	return metrics, nil

}

//will add the headers found in the hep packet as tag, based upon the
//input array, by default adds all the headers from hep packet.
func (h *Parser) addHeaders(headerNames []int, hep *HEP) map[string]string {
	headerTag := make(map[string]string)
	for _, name := range headerNames {
		switch name {
		case Version:
			headerTag[headerReverseMap[Version]] = strconv.FormatInt(int64(hep.Version), 10)
		case Protocol:
			headerTag[headerReverseMap[Protocol]] = strconv.FormatInt(int64(hep.Protocol), 10)
		case IP4SrcIP:
			headerTag[headerReverseMap[IP4SrcIP]] = hep.SrcIP
		case IP4DstIP:
			headerTag[headerReverseMap[IP4DstIP]] = hep.DstIP
		case IP6SrcIP:
			headerTag[headerReverseMap[IP6SrcIP]] = hep.DstIP
		case IP6DstIP:
			headerTag[headerReverseMap[IP6DstIP]] = hep.DstIP
		case SrcPort:
			headerTag[headerReverseMap[SrcPort]] = strconv.FormatInt(int64(hep.SrcPort), 10)
		case DstPort:
			headerTag[headerReverseMap[DstPort]] = strconv.FormatInt(int64(hep.DstPort), 10)
		case Tsec:
			headerTag[headerReverseMap[Tsec]] = strconv.FormatInt(int64(hep.Tsec), 10)
		//case Tmsec:
		//	headerTag[headerReverseMap[Tmsec]] = strconv.FormatInt(int64(hep.Tmsec), 10)
		case ProtoType:
			headerTag[headerReverseMap[Protocol]] = strconv.FormatInt(int64(hep.ProtoType), 10)
		case NodeID:
			headerTag[headerReverseMap[NodeID]] = strconv.FormatInt(int64(hep.NodeID), 10)
		case NodePW:
			headerTag[headerReverseMap[NodePW]] = string(hep.NodePW)
		case CID:
			headerTag[headerReverseMap[CID]] = string(hep.CID)
		case Vlan:
			headerTag[headerReverseMap[CID]] = strconv.FormatInt(int64(hep.Vlan), 10)
		case NodeName:
			headerTag[headerReverseMap[NodeName]] = string(hep.NodeName)
		default:
			fmt.Println("Could not find the header:", name)
		}
	}
	return headerTag

}

// ParseLine does not use any information in header and assumes DataColumns is set
// it will also not skip any rows
func (h *Parser) ParseLine(line string) (telegraf.Metric, error) {

	metrics, err := h.Parse([]byte(line + "\n"))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("can not parse the line: %s, for data format: json ", line)
	}
	return metrics[0], nil
}

func (h *Parser) SetDefaultTags(tags map[string]string) {
	h.DefaultTags = tags
}
