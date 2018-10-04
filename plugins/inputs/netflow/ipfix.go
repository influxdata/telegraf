package netflow

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs/netflow/utils"
)

type IpfixHeader struct {
	Version             uint16
	Length              uint16
	ExportTime          uint32
	SequenceNumber      uint32
	ObservationDomainID uint32
}

type IpfixSetHeader struct {
	SetID  uint16
	Length uint16
}

type IpfixInformationElement struct {
	ID               uint16
	Length           uint16
	EnterpriseNumber uint32
}

type IpfixTemplate struct {
	ID                  uint16
	FieldCount          uint16
	InformationElements []*IpfixInformationElement
}

func (t *IpfixTemplate) TotalLength() uint16 {
	var totalLen uint16
	for _, ie := range t.InformationElements {
		totalLen += ie.Length
	}
	return totalLen
}

func (t *IpfixTemplate) PrintInformationElements() {
	for _, ie := range t.InformationElements {
		log.Printf("D! Information Element: {ID=%d, Length=%d} ", ie.ID, ie.Length)
	}
}

type IpfixOptionTemplate struct {
	ID              uint16
	FieldCount      uint16
	ScopeFieldCount uint16
	ScopeFields     []*IpfixInformationElement
	OptionFields    []*IpfixInformationElement
}

func (t *IpfixOptionTemplate) TotalLength() uint16 {
	var totalLen uint16
	for _, ie := range t.ScopeFields {
		totalLen += ie.Length
	}
	for _, ie := range t.OptionFields {
		totalLen += ie.Length
	}
	return totalLen
}

func (t *IpfixOptionTemplate) PrintFields() {
	for _, ie := range t.ScopeFields {
		log.Printf("D! ScopeField: {ID=%d, Length=%d} ", ie.ID, ie.Length)
	}
	for _, ie := range t.OptionFields {
		log.Printf("D! Option Field: {ID=%d, Length=%d} ", ie.ID, ie.Length)
	}
}

type IpfixTemplateWriteOp struct {
	Exporter   *net.UDPAddr
	TemplateID uint16
	Value      IpfixTemplate
	Resp       chan bool
}

type IpfixOptionTemplateWriteOp struct {
	Exporter   *net.UDPAddr
	TemplateID uint16
	Value      IpfixOptionTemplate
	Resp       chan bool
}

type IpfixTemplateReadOp struct {
	Exporter           *net.UDPAddr
	TemplateID         uint16
	TemplateResp       chan IpfixTemplate
	OptionTemplateResp chan IpfixOptionTemplate
	Fail               chan bool
}

type IpfixInformationElementReadOp struct {
	Key  uint16
	Resp chan IpfixInformationElement2
	Fail chan bool
}

func (n *Netflow) ipfixTemplatePoller() error {
	defer n.wg.Done()

	var templateTable = map[string]map[uint16]IpfixTemplate{}
	var optionTemplateTable = map[string]map[uint16]IpfixOptionTemplate{}
	for {
		select {
		case <-n.done:
			return nil
		case read := <-n.ipfixReadTemplate:
			log.Println("I! ipfix read template")
			exporterAddr := read.Exporter.IP.String()
			if _, ok := templateTable[exporterAddr]; ok {
				if template, ok := templateTable[exporterAddr][read.TemplateID]; ok {
					log.Printf("I! found template: id=%d", template.ID)
					read.TemplateResp <- template
					continue
				}
			}
			if _, ok := optionTemplateTable[exporterAddr]; ok {
				if template, ok := optionTemplateTable[exporterAddr][read.TemplateID]; ok {
					log.Printf("I! found option template: id=%d", template.ID)
					read.OptionTemplateResp <- template
					continue
				}
			}
			log.Printf("I! not found: exporter=%s, templateId=%d", exporterAddr, read.TemplateID)
			read.Fail <- false
		case write := <-n.ipfixWriteTemplate:
			log.Printf("I! ipfix write template")
			exporterAddr := write.Exporter.IP.String()
			if _, ok := templateTable[exporterAddr]; ok {
				templateTable[exporterAddr][write.TemplateID] = write.Value
			} else {
				templateTable[exporterAddr] = map[uint16]IpfixTemplate{}
				templateTable[exporterAddr][write.TemplateID] = write.Value
			}
			write.Resp <- true
		case write := <-n.ipfixWriteOptionTemplate:
			log.Printf("I! ipfix write option template")
			exporterAddr := write.Exporter.IP.String()
			if _, ok := optionTemplateTable[exporterAddr]; ok {
				optionTemplateTable[exporterAddr][write.TemplateID] = write.Value
			} else {
				optionTemplateTable[exporterAddr] = map[uint16]IpfixOptionTemplate{}
				optionTemplateTable[exporterAddr][write.TemplateID] = write.Value
			}
			write.Resp <- true
		}
	}
}

func (n *Netflow) ipfixInformationElementPoller() error {
	defer n.wg.Done()
	var fieldTable = ipfixFieldTable
	for {
		select {
		case <-n.done:
			return nil
		case read := <-n.ipfixReadInformationElement:
			field, ok := fieldTable[read.Key]
			if ok {
				read.Resp <- field
			} else {
				read.Resp <- NewInformationElement2("unknown")
			}
		}

	}
}

func (n *Netflow) parseIpfixTemplateSet(frame *bytes.Buffer, exporter *net.UDPAddr, fsLen uint16) {
	var current uint16 = 4 // set id + set length
	for current < fsLen {
		// need to consider padding later
		var template = new(IpfixTemplate)
		if err := binary.Read(frame, binary.BigEndian, &template.ID); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2
		if err := binary.Read(frame, binary.BigEndian, &template.FieldCount); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2

		log.Printf("D! Template ID=%d", template.ID)
		log.Printf("D! Template Field Count=%d", template.FieldCount)

		for i := uint16(0); i < template.FieldCount; i++ {
			var ie = new(IpfixInformationElement)
			if err := binary.Read(frame, binary.BigEndian, &ie.ID); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if err := binary.Read(frame, binary.BigEndian, &ie.Length); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if ie.ID >= 32768 {
				if err := binary.Read(frame, binary.BigEndian, &ie.EnterpriseNumber); err != nil {
					log.Printf("E!%s\n", err.Error())
				}
				current += 4
			}
			template.InformationElements = append(template.InformationElements, ie)
		}
		template.PrintInformationElements()
		writeOp := &IpfixTemplateWriteOp{TemplateID: template.ID, Exporter: exporter, Value: *template, Resp: make(chan bool)}
		n.ipfixWriteTemplate <- writeOp
		<-writeOp.Resp
	}
}

func (n *Netflow) parseIpfixOptionTemplateSet(frame *bytes.Buffer, exporter *net.UDPAddr, fsLen uint16) {
	var current uint16 = 4 // set id + flowset length
	for current < fsLen {
		// need to consider padding
		var template = new(IpfixOptionTemplate)
		if err := binary.Read(frame, binary.BigEndian, &template.ID); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2
		if err := binary.Read(frame, binary.BigEndian, &template.FieldCount); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2
		if err := binary.Read(frame, binary.BigEndian, &template.ScopeFieldCount); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2
		log.Printf("D! Template ID=%d", template.ID)
		log.Printf("D! Option Field Count=%d", template.FieldCount)
		log.Printf("D! Option Scope Field Count=%d", template.ScopeFieldCount)

		for i := uint16(0); i < template.ScopeFieldCount; i++ {
			var ie = new(IpfixInformationElement)
			if err := binary.Read(frame, binary.BigEndian, &ie.ID); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if err := binary.Read(frame, binary.BigEndian, &ie.Length); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if ie.ID >= 32768 {
				if err := binary.Read(frame, binary.BigEndian, &ie.EnterpriseNumber); err != nil {
					log.Printf("E! %s\n", err.Error())
				}
				current += 4
			}
			template.ScopeFields = append(template.ScopeFields, ie)
		}

		for i := uint16(0); i < template.FieldCount-template.ScopeFieldCount; i++ {
			var ie = new(IpfixInformationElement)
			if err := binary.Read(frame, binary.BigEndian, &ie.ID); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if err := binary.Read(frame, binary.BigEndian, &ie.Length); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if ie.ID >= 32768 {
				if err := binary.Read(frame, binary.BigEndian, &ie.EnterpriseNumber); err != nil {
					log.Printf("E! %s\n", err.Error())
				}
				current += 4
			}
			template.OptionFields = append(template.OptionFields, ie)
		}
		template.PrintFields()
		writeOp := &IpfixOptionTemplateWriteOp{TemplateID: template.ID, Exporter: exporter, Value: *template, Resp: make(chan bool)}
		n.ipfixWriteOptionTemplate <- writeOp
		<-writeOp.Resp
	}
}

func (n *Netflow) parseIpfixDataSet(frame *bytes.Buffer, exporter *net.UDPAddr, fsId uint16, fsLen uint16) []telegraf.Metric {
	readOp := &IpfixTemplateReadOp{TemplateID: fsId, Exporter: exporter, TemplateResp: make(chan IpfixTemplate), OptionTemplateResp: make(chan IpfixOptionTemplate), Fail: make(chan bool)}
	n.ipfixReadTemplate <- readOp
	select {
	case template := <-readOp.TemplateResp:
		metrics := n.parseIpfixDataSetInternal(frame, exporter, fsLen, template)
		return metrics
	case template := <-readOp.OptionTemplateResp:
		n.parseIpfixOptionDataSetInternal(frame, fsLen, template)
		return nil
	case <-readOp.Fail:
		log.Printf("W!Template for id=%d is missing", fsId)
		return nil
	}
}

func (n *Netflow) parseIpfixDataSetInternal(frame *bytes.Buffer, exporter *net.UDPAddr, fsLen uint16, template IpfixTemplate) []telegraf.Metric {
	var metrics []telegraf.Metric
	padding := (fsLen - 4) % template.TotalLength()
	log.Printf("D! fsLen=%d", fsLen)
	log.Printf("D! template length=%d", template.TotalLength())
	log.Printf("D! Padding=%d", padding)
	for current := uint16(4); current < fsLen-padding; {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		for _, ie := range template.InformationElements {
			var byteArray = make([]byte, int(ie.Length))
			binary.Read(frame, binary.BigEndian, &byteArray)

			readOp := &IpfixInformationElementReadOp{Key: ie.ID, Resp: make(chan IpfixInformationElement2)}
			n.ipfixReadInformationElement <- readOp
			ie2 := <-readOp.Resp

			log.Printf("D! information element=%s", ie2.Name)
			switch ie2.Type {
			case "ipv4_addr":
				fields[ie2.Name] = fmt.Sprintf("%d.%d.%d.%d", int(byteArray[0]), int(byteArray[1]), int(byteArray[2]), int(byteArray[3]))
			case "integer":
				if ie.Length <= 8 {
					fields[ie2.Name] = utils.ReadIntFromByteArray(byteArray, ie.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "float":
				if ie.Length <= 8 {
					bits := binary.BigEndian.Uint64(byteArray)
					fields[ie2.Name] = math.Float64frombits(bits)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "string":
				fields[ie2.Name] = string(byteArray[:ie.Length])
			default: // unknown
				log.Println("W! unsupported flow field type")
				if ie.Length <= 8 {
					fields[ie2.Name] = utils.ReadIntFromByteArray(byteArray, ie.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			}
			current += ie.Length
		}

		tags["exporter"] = exporter.IP.String()

		if n.ResolveApplicationNameByID {
			applicationID, ok := fields["application_id"]
			log.Printf("I! resolve application name by id=%d", applicationID)
			if ok {
				appID, ok := applicationID.(uint32)
				log.Printf("I! resolve application name by id=%d", appID)
				if ok {
					id := fmt.Sprintf("D! %d:%d", appID / 16777216, appID % 16777216) // 2^24 = 16777216
					log.Printf("D! id=%s", id)
					readOp := &ApplicationReadOp{Key: id, Resp: make(chan Application)}
					n.readApplication <- readOp
					select {
					case application := <-readOp.Resp:
						tags["application_name"] = application.Name
					case <-readOp.Fail:
						log.Printf("W! unknown applicaton id=%s", id)
					}
				}
			}
		}
		if n.ResolveIfnameByIfindex {
			n.resolveIfname(fields, tags, "interface_input_snmp", "interface_input_name")
			n.resolveIfname(fields, tags, "interface_output_snmp", "interface_output_name")
		}
		metric, err := metric.New("netflow", tags, fields, time.Now())
		if err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		log.Printf("D! fields: %v", fields)
		log.Printf("D! tags: %v", tags)
		metrics = append(metrics, metric)
	}
	return metrics
}

func (n *Netflow) parseIpfixOptionDataSetInternal(frame *bytes.Buffer, fsLen uint16, template IpfixOptionTemplate) {
	padding := (fsLen - 4) % template.TotalLength()
	log.Printf("D! fsLen=%d, ID=%d, padding=%d", fsLen, template.ID, padding)
	for current := uint16(4); current < fsLen-padding; {
		scopeFields := make(map[string]interface{})
		for _, f := range template.ScopeFields {
			var byteArray = make([]byte, int(f.Length))
			binary.Read(frame, binary.BigEndian, &byteArray)
			// change for scope later
			readOp := &IpfixInformationElementReadOp{Key: f.ID, Resp: make(chan IpfixInformationElement2)}
			n.ipfixReadInformationElement <- readOp
			ie2 := <-readOp.Resp
			switch ie2.Type {
			case "ipv4_addr":
				scopeFields[ie2.Name] = fmt.Sprintf("%d.%d.%d.%d", int(byteArray[0]), int(byteArray[1]), int(byteArray[2]), int(byteArray[3]))
			case "integer":
				if f.Length <= 8 {
					scopeFields[ie2.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "float":
				if f.Length <= 8 {
					bits := binary.BigEndian.Uint64(byteArray)
					scopeFields[ie2.Name] = math.Float64frombits(bits)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "string":
				scopeFields[ie2.Name] = string(byteArray[:f.Length])
			default: // unknown
				log.Println("W! unsupported flow field type")
				if f.Length <= 8 {
					scopeFields[ie2.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			}
			log.Printf("D! %s: %v", ie2.Name, scopeFields[ie2.Name])
			current += f.Length
		}
		optionFields := make(map[string]interface{})
		for _, f := range template.OptionFields {
			var byteArray = make([]byte, int(f.Length))
			binary.Read(frame, binary.BigEndian, &byteArray)

			readOp := &IpfixInformationElementReadOp{Key: f.ID, Resp: make(chan IpfixInformationElement2)}
			n.ipfixReadInformationElement <- readOp
			ie2 := <-readOp.Resp

			switch ie2.Type {
			case "ipv4_addr":
				optionFields[ie2.Name] = fmt.Sprintf("%d.%d.%d.%d", int(byteArray[0]), int(byteArray[1]), int(byteArray[2]), int(byteArray[3]))
			case "integer":
				if f.Length <= 8 {
					optionFields[ie2.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "float":
				if f.Length <= 8 {
					bits := binary.BigEndian.Uint64(byteArray)
					optionFields[ie2.Name] = math.Float64frombits(bits)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "string":
				optionFields[ie2.Name] = string(byteArray[:f.Length])
			default: // unknown
				log.Println("W! unsupported flow field type")
				if f.Length <= 8 {
					optionFields[ie2.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			}
			log.Printf("D! %s: %v", ie2.Name, optionFields[ie2.Name])
			current += f.Length
		}
		// register ifname
		if n.ResolveIfnameByIfindex {
			inputSnmp, ok1 := scopeFields["interface_input_snmp"]
			ifIndex, ok2 := inputSnmp.(uint32)
			interfaceName, ok3 := optionFields["interface_name_long"]
			ifName, ok4 := interfaceName.(string)
			if ok1 && ok2 && ok3 && ok4 {
				writeOp := &IfnameWriteOp{Key: ifIndex, Value: ifName, Resp: make(chan bool)}
				n.writeIfname <- writeOp
				<-writeOp.Resp
			}
		}
	}
}
