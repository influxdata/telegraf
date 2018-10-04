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

type V9TemplateWriteOp struct {
	Exporter   *net.UDPAddr
	TemplateID uint16
	Value      V9Template
	Resp       chan bool
}

type V9OptionTemplateWriteOp struct {
	Exporter   *net.UDPAddr
	TemplateID uint16
	Value      V9OptionTemplate
	Resp       chan bool
}

type V9TemplateReadOp struct {
	Exporter           *net.UDPAddr
	TemplateID         uint16
	TemplateResp       chan V9Template
	OptionTemplateResp chan V9OptionTemplate
	Fail               chan bool
}

type V9FlowFieldReadOp struct {
	Key  uint16
	Resp chan V9FlowField
	Fail chan bool
}

func (n *Netflow) v9TemplatePoller() error {
	defer n.wg.Done()

	var templateTable = map[string]map[uint16]V9Template{}
	var optionTemplateTable = map[string]map[uint16]V9OptionTemplate{}
	for {
		select {
		case <-n.done:
			return nil
		case read := <-n.v9ReadTemplate:
			log.Println("I! read template")
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
			log.Printf("W! not found: exporter=%s, templateID=%d", exporterAddr, read.TemplateID)
			read.Fail <- false
		case write := <-n.v9WriteTemplate:
			log.Println("I! write template")
			exporterAddr := write.Exporter.IP.String()
			if _, ok := templateTable[exporterAddr]; ok {
				templateTable[exporterAddr][write.TemplateID] = write.Value
			} else {
				templateTable[exporterAddr] = map[uint16]V9Template{}
				templateTable[exporterAddr][write.TemplateID] = write.Value
			}
			write.Resp <- true
		case write := <-n.v9WriteOptionTemplate:
			log.Println("I! write option template")
			exporterAddr := write.Exporter.IP.String()
			if _, ok := optionTemplateTable[exporterAddr]; ok {
				optionTemplateTable[exporterAddr][write.TemplateID] = write.Value
			} else {
				optionTemplateTable[exporterAddr] = map[uint16]V9OptionTemplate{}
				optionTemplateTable[exporterAddr][write.TemplateID] = write.Value
			}
			write.Resp <- true
		}
	}
}

func (n *Netflow) v9FlowFieldPoller() error {
	defer n.wg.Done()
	var fieldTable = v9FieldTable
	for {
		select {
		case <-n.done:
			return nil
		case read := <-n.v9ReadFlowField:
			field, ok := fieldTable[read.Key]
			if ok {
				read.Resp <- field
			} else {
				read.Resp <- NewFlowField("unknown")
			}
		}
	}
}

type V9Header struct {
	Version        uint16
	Count          uint16
	SysUptime      uint32
	UNIXSeconds    uint32
	SequenceNumber uint32
	SourceID       uint32
}

type V9Field struct {
	Type   uint16
	Length uint16
}

type V9Template struct {
	ID         uint16
	FieldCount uint16
	Fields     []*V9Field
}

func (t *V9Template) TotalLength() uint16 {
	var total_len uint16
	for _, f := range t.Fields {
		total_len += f.Length
	}
	return total_len
}

func (t *V9Template) PrintFields() {
	for _, f := range t.Fields {
		log.Printf("D! {Length=%d, Type=%d}", f.Length, f.Type)
	}
}

type V9OptionTemplate struct {
	ID                uint16
	OptionScopeLength uint16
	OptionLength      uint16
	ScopeFields       []*V9Field
	OptionFields      []*V9Field
}

func (t *V9OptionTemplate) TotalLength() uint16 {
	var total_len uint16
	for _, f := range t.ScopeFields {
		total_len += f.Length
	}
	for _, f := range t.OptionFields {
		total_len += f.Length
	}
	return total_len
}

func (t *V9OptionTemplate) PrintFields() {
	for _, f := range t.ScopeFields {
		log.Printf("D! Scope Field: {Length=%d, Type=%d}", f.Length, f.Type)
	}
	for _, f := range t.OptionFields {
		log.Printf("D! Option Field: {Length=%d, Type=%d}", f.Length, f.Type)
	}
}

func (n *Netflow) parseV9TemplateFlowset(frame *bytes.Buffer, exporter *net.UDPAddr, fsLen uint16) uint16 {
	var templateCount uint16 = 0
	var current uint16 = 4 // flowset id + flowset length
	for current < fsLen {
		// no need to consider padding
		var template = new(V9Template)
		if err := binary.Read(frame, binary.BigEndian, &template.ID); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2
		if err := binary.Read(frame, binary.BigEndian, &template.FieldCount); err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		current += 2
		log.Printf("D! Template ID=%d", template.ID)
		log.Printf("D! Field Number=%d", template.FieldCount)
		for i := uint16(0); i < template.FieldCount; i++ {
			var field = new(V9Field)
			if err := binary.Read(frame, binary.BigEndian, &field.Type); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			if err := binary.Read(frame, binary.BigEndian, &field.Length); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			current += 2
			template.Fields = append(template.Fields, field)
		}
		template.PrintFields()
		writeOp := &V9TemplateWriteOp{TemplateID: template.ID, Exporter: exporter, Value: *template, Resp: make(chan bool)}
		n.v9WriteTemplate <- writeOp
		<-writeOp.Resp
		templateCount++
	}
	return templateCount
}

func (n *Netflow) parseV9OptionTemplateFlowset(frame *bytes.Buffer, exporter *net.UDPAddr, fsLen uint16) uint16 {
	var templateCount uint16 = 0
	var current uint16 = 4 // flowset id + flowset length
	for current < fsLen {
		// need to consider padding
		var template = new(V9OptionTemplate)
		if err := binary.Read(frame, binary.BigEndian, &template.ID); err != nil {
			log.Fatal(err)
		}
		current += 2
		if err := binary.Read(frame, binary.BigEndian, &template.OptionScopeLength); err != nil {
			log.Fatal(err)
		}
		current += 2
		if err := binary.Read(frame, binary.BigEndian, &template.OptionLength); err != nil {
			log.Fatal(err)
		}
		current += 2
		log.Printf("D! Template ID=%d", template.ID)
		log.Printf("D! Option Scope Length=%d", template.OptionScopeLength)
		log.Printf("D! Option Length=%d", template.OptionLength)
		for i := uint16(0); i < template.OptionScopeLength / 4; i++ {
			var field = new(V9Field)
			if err := binary.Read(frame, binary.BigEndian, &field.Type); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			if err := binary.Read(frame, binary.BigEndian, &field.Length); err != nil {
				log.Printf("E!%s\n", err.Error())
			}
			template.ScopeFields = append(template.ScopeFields, field)
		}
		current += template.OptionScopeLength
		for i := uint16(0); i < template.OptionLength / 4; i++ {
			var field = new(V9Field)
			if err := binary.Read(frame, binary.BigEndian, &field.Type); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			if err := binary.Read(frame, binary.BigEndian, &field.Length); err != nil {
				log.Printf("E! %s\n", err.Error())
			}
			template.OptionFields = append(template.OptionFields, field)
		}
		current += template.OptionLength
		template.PrintFields()
		writeOp := &V9OptionTemplateWriteOp{TemplateID: template.ID, Exporter: exporter, Value: *template, Resp: make(chan bool)}
		n.v9WriteOptionTemplate <- writeOp
		<-writeOp.Resp
		templateCount++
	}
	return templateCount
}

func (n *Netflow) parseV9DataFlowset(frame *bytes.Buffer, exporter *net.UDPAddr, fsID uint16, fsLen uint16) (uint16, []telegraf.Metric, bool) {
	var recordCount uint16
	readOp := &V9TemplateReadOp{TemplateID: fsID, Exporter: exporter, TemplateResp: make(chan V9Template), OptionTemplateResp: make(chan V9OptionTemplate), Fail: make(chan bool)}
	n.v9ReadTemplate <- readOp
	select {
	case template := <-readOp.TemplateResp:
		recordCount, metrics := n.parseV9DataFlowsetInternal(frame, exporter, fsLen, template)
		return recordCount, metrics, true
	case template := <-readOp.OptionTemplateResp:
		recordCount = n.parseV9OptionDataRecordInternal(frame, fsLen, template)
		return recordCount, nil, true
	case <-readOp.Fail:
		log.Printf("W! Template for id=%d is missing", fsID)
		return 0, nil, false
	}
}

func (n *Netflow) resolveIfname(fields map[string]interface{}, tags map[string]string, ifIndexLabel string, ifNameLabel string) {
	ifIndex, ok := fields[ifIndexLabel]
	//log.Printf("D! resolve ifname by ifIndex=%v", ifIndex)
	if ok {
		ifIndexID, ok := ifIndex.(uint32)
		log.Printf("D! resolve ifname by ifIndex=%d", ifIndexID)
		if ok && ifIndexID != 0 {
			readOp := &IfnameReadOp{Key: ifIndexID, Resp: make(chan string)}
			n.readIfname <- readOp
			select {
			case ifname := <-readOp.Resp:
				tags[ifNameLabel] = ifname
			case <-readOp.Fail:
				log.Printf("W! unknown ifindex=%d", ifIndexID)
			}
		}
	}
}

func (n *Netflow) parseV9DataFlowsetInternal(frame *bytes.Buffer, exporter *net.UDPAddr, fsLen uint16, template V9Template) (uint16, []telegraf.Metric) {
	var metrics []telegraf.Metric
	recordCount := (fsLen - 4) / template.TotalLength()
	padding := (fsLen - 4) % template.TotalLength()
	for current := uint16(4); current < fsLen - padding; {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		for _, f := range template.Fields {
			var byteArray = make([]byte, int(f.Length))
			binary.Read(frame, binary.BigEndian, &byteArray)

			readOp := &V9FlowFieldReadOp{Key: f.Type, Resp: make(chan V9FlowField)}
			n.v9ReadFlowField <- readOp
			flowField := <-readOp.Resp

			switch flowField.Type {
			case "ipv4_addr":
				fields[flowField.Name] = fmt.Sprintf("%d.%d.%d.%d", int(byteArray[0]), int(byteArray[1]), int(byteArray[2]), int(byteArray[3]))
			case "integer":
				if f.Length <= 8 {
					if flowField.Name == "flow_direction" {
						val := utils.ReadIntFromByteArray(byteArray, f.Length)
						val2 := val.(uint8)
						log.Println("D! Flow Direction: ", val2)
						if val2 == 0 {
							log.Println("D! Flow Direction: ingress")
							tags[flowField.Name] = "ingress"
						} else {
							log.Println("D! Flow Direction: egress")
							tags[flowField.Name] = "egress"
						}
					} else {
						fields[flowField.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
					}
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "float":
				if f.Length <= 8 {
					bits := binary.BigEndian.Uint64(byteArray)
					fields[flowField.Name] = math.Float64frombits(bits)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "string":
				fields[flowField.Name] = string(byteArray[:f.Length])
			default: // unknown
				log.Println("W! unsupported flow field type")
				if f.Length <= 8 {
					fields[flowField.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			}
			current += f.Length
		}

		tags["exporter"] = exporter.IP.String()

		if n.ResolveApplicationNameByID {
			applicationID, ok := fields["application_id"]
			log.Printf("I! resolve application name by id=%d", applicationID)
			if ok {
				appID, ok := applicationID.(uint32)
				log.Printf("I! resolve application name by id=%d", appID)
				if ok {
					id := fmt.Sprintf("%d:%d", appID / 16777216, appID % 16777216) // 2^24 = 16777216
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
		m, err := metric.New("netflow", tags, fields, time.Now())
		if err != nil {
			log.Printf("E! %s\n", err.Error())
		}
		log.Printf("D! fields: %v", fields)
		log.Printf("D! tags: %v", tags)
		metrics = append(metrics, m)
	}
	return recordCount, metrics
}

func (n *Netflow) parseV9OptionDataRecordInternal(frame *bytes.Buffer, fsLen uint16, template V9OptionTemplate) uint16 {
	recordCount := (fsLen - 4) / template.TotalLength()
	padding := (fsLen - 4) % template.TotalLength()
	log.Printf("D! fsLen=%d, totalLength=%d, padding=%d", fsLen, template.TotalLength(), padding)
	for current := uint16(4); current < fsLen - padding; {
		//scopeFields := make(map[string]interface{})
		for _, f := range template.ScopeFields {
			var byteArray = make([]byte, int(f.Length))
			binary.Read(frame, binary.BigEndian, &byteArray)
			// Todo: need change for the scope later
			readOp := &V9FlowFieldReadOp{Key: f.Type, Resp: make(chan V9FlowField)}
			n.v9ReadFlowField <- readOp
			flowField := <-readOp.Resp
			log.Printf("D! scope field=%s", flowField.Name)
			current += f.Length
		}
		optionFields := make(map[string]interface{})
		for _, f := range template.OptionFields {
			var byteArray = make([]byte, int(f.Length))
			binary.Read(frame, binary.BigEndian, &byteArray)
			readOp := &V9FlowFieldReadOp{Key: f.Type, Resp: make(chan V9FlowField)}
			n.v9ReadFlowField <- readOp
			flowField := <-readOp.Resp
			switch flowField.Type {
			case "ipv4_addr":
				optionFields[flowField.Name] = fmt.Sprintf("%d.%d.%d.%d", int(byteArray[0]), int(byteArray[1]), int(byteArray[2]), int(byteArray[3]))
			case "integer":
				if f.Length <= 8 {
					optionFields[flowField.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "float":
				if f.Length <= 8 {
					bits := binary.BigEndian.Uint64(byteArray)
					optionFields[flowField.Name] = math.Float64frombits(bits)
				} else {
					log.Println("W! unsupported flow field length")
				}
			case "string":
				optionFields[flowField.Name] = string(byteArray[:f.Length])
			default: // unknown
				log.Println("W! unsupported flow field type")
				if f.Length <= 8 {
					optionFields[flowField.Name] = utils.ReadIntFromByteArray(byteArray, f.Length)
				} else {
					log.Println("W! unsupported flow field length")
				}
			}
			log.Printf("D! %s: %v", flowField.Name, optionFields[flowField.Name])
			// register ifname
			if n.ResolveIfnameByIfindex {
				inputSnmp, ok1 := optionFields["interface_input_snmp"]
				ifIndex, ok2 := inputSnmp.(uint32)
				interfaceName, ok3 := optionFields["interface_name_long"]
				ifName, ok4 := interfaceName.(string)
				if ok1 && ok2 && ok3 && ok4 {
					writeOp := &IfnameWriteOp{Key: ifIndex, Value: ifName, Resp: make(chan bool)}
					n.writeIfname <- writeOp
					<-writeOp.Resp
				}
			}
			current += f.Length
		}
	}
	return recordCount
}
