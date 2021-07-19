/*
The MIT License (MIT)

Copyright 2014-present, Apstra, Inc. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package aos

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/aos/aos_streaming"
	aosrestapi "github.com/influxdata/telegraf/plugins/inputs/aos/restapi"
)

// ----------------------------------------------------------------
// StreamAos "Class"
// ----------------------------------------------------------------
type StreamAos struct {
	net.Listener
	*Aos
}

func (ssl *StreamAos) listen() {

	for {
		conn, err := ssl.Listener.Accept()
		if err != nil {
			log.Printf("W! Accepting Conn: " + err.Error())
			continue
		}

		go ssl.MsgReader(conn)
	}

}

func (ssl *StreamAos) ExtractEventData(eventType string, tags map[string]string, eventData interface{}) {

	myEventDataValue := reflect.ValueOf(eventData).Elem()
	myEventDataType := myEventDataValue.Type()
	propDataType := proto.GetProperties(myEventDataType)

	serie := "event_" + eventType
	fields := make(map[string]interface{})

	fields["event"] = 1

	for i := 0; i < myEventDataValue.NumField(); i++ {
		myField := myEventDataValue.Field(i)
		if myField.IsNil() {
			continue
		}
		field_name := propDataType.Prop[i].OrigName

		// Skip field with XXX_
		if strings.Contains(field_name, "XXX_") {
			continue
		}

		if propDataType.Prop[i].Enum != "" {
			field_value := fmt.Sprintf("%v", myField.Elem().Interface().(fmt.Stringer).String())
			tags[field_name] = field_value
		} else {
			field_value := fmt.Sprintf("%v", reflect.Indirect(myField).Interface())
			tags[field_name] = field_value
		}
	}

	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) ExtractAlertData(alertType string, tags map[string]string, alertData interface{}, raised bool) {

	myAlertDataValue := reflect.ValueOf(alertData).Elem()
	myAlertDataType := myAlertDataValue.Type()
	propDataType := proto.GetProperties(myAlertDataType)

	serie := "alert_" + strings.Replace(alertType, "_alert", "", -1)
	fields := make(map[string]interface{})

	if raised {
		fields["status"] = 1
	} else {
		fields["status"] = 0
	}

	for i := 0; i < myAlertDataValue.NumField(); i++ {
		myField := myAlertDataValue.Field(i)
		if myField.IsNil() {
			continue
		}
		field_name := propDataType.Prop[i].OrigName

		// Skip field with XXX_
		if strings.Contains(field_name, "XXX_") {
			continue
		}

		if propDataType.Prop[i].Enum != "" {
			field_value := fmt.Sprintf("%v", myField.Elem().Interface().(fmt.Stringer).String())
			tags[field_name] = field_value
		} else {
			field_value := fmt.Sprintf("%v", reflect.Indirect(myField).Interface())
			tags[field_name] = field_value
		}
	}
	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) GetTags(deviceKey string) map[string]string {

	tags := make(map[string]string)

	// search for :: in string and split if found
	if strings.Contains(deviceKey, "::") {
		devInt := strings.Split(deviceKey, "::")
		deviceKey = devInt[0]
		tags["interface"] = devInt[1]
	}

	tags["device_key"] = deviceKey

	if ssl.IsSequencedStream {
		// If the version supports sequenced streams, then we also have augmented messages.
		// There's no need to query the server for additional info.
		return tags
	}

	system := ssl.Aos.api.GetSystemByKey(deviceKey)

	if system != nil {
		if system.Blueprint.Role != "" {
			tags["role"] = system.Blueprint.Role
		}

		if system.Status.BlueprintId != "" {
			blueprint := ssl.Aos.api.GetBlueprintById(system.Status.BlueprintId)
			if blueprint != nil {
				tags["blueprint"] = blueprint.Name
			}
		}

		if system.Blueprint.Name != "" {
			tags["device_name"] = system.Blueprint.Name
			tags["device"] = system.Blueprint.Name
		} else {
			tags["device"] = deviceKey
		}
	} else {
		tags["device"] = deviceKey
	}

	return tags
}

func (ssl *StreamAos) ExtractProbeData(newProbeMessage interface{}, tags map[string]string) {
	myValue := reflect.ValueOf(newProbeMessage).Elem()
	myType := myValue.Type()
	propType := proto.GetProperties(myType)

	serie := "probe_message"
	fields := make(map[string]interface{})

	for i := 0; i < myValue.NumField(); i++ {
		myField := myValue.Field(i)
		field_name := propType.Prop[i].OrigName

		if strings.Contains(field_name, "XXX_") {
			continue
		}

		temp := reflect.Indirect(myField)
		if temp == reflect.ValueOf(nil) {
			continue
		}
		fields[field_name] = temp.Interface()
		if _, ok := temp.Interface().(string); !ok {
			tmpString := fmt.Sprintf("%v", temp.Interface())
			fields[field_name] = tmpString
		}
	}

	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) CreateBpsMetrics(fields map[string]interface{}) {
	delta_seconds := fields["delta_seconds"].(uint64)
	if delta_seconds == 0 {
		return
	}

	fields["tx_bps"] = fields["tx_bytes"].(uint64) * 8 / delta_seconds
	fields["tx_unicast_pps"] = fields["tx_unicast_packets"].(uint64) / delta_seconds
	fields["tx_broadcast_pps"] = fields["tx_broadcast_packets"].(uint64) / delta_seconds
	fields["tx_multicast_pps"] = fields["tx_multicast_packets"].(uint64) / delta_seconds
	fields["tx_error_pps"] = fields["tx_error_packets"].(uint64) / delta_seconds
	fields["tx_discard_pps"] = fields["tx_discard_packets"].(uint64) / delta_seconds
	fields["rx_bps"] = fields["rx_bytes"].(uint64) * 8 / delta_seconds
	fields["rx_unicast_pps"] = fields["rx_unicast_packets"].(uint64) / delta_seconds
	fields["rx_broadcast_pps"] = fields["rx_broadcast_packets"].(uint64) / delta_seconds
	fields["rx_multicast_pps"] = fields["rx_multicast_packets"].(uint64) / delta_seconds
	fields["rx_error_pps"] = fields["rx_error_packets"].(uint64) / delta_seconds
	fields["rx_discard_pps"] = fields["rx_discard_packets"].(uint64) / delta_seconds
}

func (ssl *StreamAos) ExtractIntfData(intfData interface{}, tags map[string]string) {
	myValue := reflect.ValueOf(intfData).Elem()
	myType := myValue.Type()
	propType := proto.GetProperties(myType)

	serie := "interface_counters"
	fields := make(map[string]interface{})

	for i := 0; i < myValue.NumField(); i++ {
		myField := myValue.Field(i)
		field_name := propType.Prop[i].OrigName

		// Skip field with XXX_
		if strings.Contains(field_name, "XXX_") {
			continue
		}

		if !myField.IsNil() {
			fields[propType.Prop[i].OrigName] = reflect.Indirect(myField).Interface()
		} else if field_name == "delta_seconds" {
			// try to use the default value
			defaultDeltaSeconds, _ := strconv.Atoi(propType.Prop[i].Default)
			fields[propType.Prop[i].OrigName] = uint64(defaultDeltaSeconds)
		} else {
			log.Printf("W!: Found nil field without default value %v", field_name)
		}
	}
	ssl.CreateBpsMetrics(fields)
	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) ExtractSystemInfo(systemInfo interface{}, tags map[string]string) {
	// Prepare value. type and property
	myValue := reflect.ValueOf(systemInfo).Elem()
	myType := myValue.Type()
	propType := proto.GetProperties(myType)

	serie := "system_info"
	fields := make(map[string]interface{})

	for i := 0; i < myValue.NumField(); i++ {
		myField := myValue.Field(i)
		field_name := propType.Prop[i].OrigName

		// Skip field with XXX_
		if strings.Contains(field_name, "XXX_") {
			continue
		}

		fields[field_name] = reflect.Indirect(myField).Interface()
	}

	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) ExtractProcessInfo(processInfo []*aos_streaming.ProcessInfo, tags map[string]string) {
	for _, p := range processInfo {

		myValue := reflect.ValueOf(p).Elem()
		myType := myValue.Type()
		propType := proto.GetProperties(myType)

		process_name := p.ProcessName
		serie := "process_info"
		fields := make(map[string]interface{})

		tags["process_name"] = *process_name

		for i := 0; i < myValue.NumField(); i++ {
			myField := myValue.Field(i)
			field_name := propType.Prop[i].OrigName

			// Skip field with XXX_ and process_name
			if strings.Contains(field_name, "XXX_") {
				continue
			}
			if strings.Contains(field_name, "process_name") {
				continue
			}

			fields[field_name] = reflect.Indirect(myField).Interface()
		}

		ssl.Aos.Accumulator.AddFields(serie, fields, tags)
	}
}

func (ssl *StreamAos) ExtractFileInfo(fileInfo []*aos_streaming.FileInfo, tags map[string]string) {
	for _, f := range fileInfo {
		file_name := f.FileName
		file_size := f.FileSize

		serie := "file_info"
		fields := make(map[string]interface{})

		tags["file_name"] = *file_name
		fields["size"] = *file_size

		ssl.Aos.Accumulator.AddFields(serie, fields, tags)
	}
}

func (ssl *StreamAos) reportMessageLoss(msg_type string, expected uint64, actual uint64) {
	ssl.msgLoss.Lock()
	defer ssl.msgLoss.Unlock()
	if actual < expected {
		log.Printf("W! %v sequence number less than expected sequence number.", msg_type)
		return
	}

	loss_count := actual - expected
	serie := "message_loss"

	tags := make(map[string]string)
	tags["message_type"] = msg_type

	fields := make(map[string]interface{})
	fields[msg_type] = loss_count

	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) MsgReader(r io.Reader) {
	var msgSize uint16

	log.Printf("D! New TCP Session Opened .. ")

	for {
		sizeReader := io.LimitReader(r, 2)
		sizeBuf, err := ioutil.ReadAll(sizeReader)

		if err != nil {
			log.Printf("W! Reading Size failed: %v", err)
			return
		}

		err = binary.Read(
			bytes.NewReader(sizeBuf),
			binary.BigEndian,
			&msgSize)

		if err != nil {
			log.Printf("W! binary.Read failed: %v", err)
			return
		}

		IoMsgReader := io.LimitReader(r, int64(msgSize))
		msgBuf, err := ioutil.ReadAll(IoMsgReader)

		if err != nil {
			log.Printf("W! Reading message failed: %v", err)
			return
		}

		ssl.handleTelemetry(msgBuf)
	}
}

func getFuncName(msg_name string) string {
	result := "Get"
	tokens := strings.Split(msg_name, "_")
	for _, token := range tokens {
		result += strings.Title(token)
	}
	return result
}

func (ssl *StreamAos) handleTelemetry(msgBuf []byte) {
	// Create new aos_streaming.AosMessage and deserialize protobuf
	newMsg := new(aos_streaming.AosMessage)
	sequencedMsg := new(aos_streaming.AosSequencedMessage)

	ssl.sequencingMode.Do(func() {
		// Automatically detect if we are receiving sequenced or unsequenced protobuf messages
		err := proto.Unmarshal(msgBuf, sequencedMsg)
		if err != nil {
			log.Printf("W! Error unmarshaling sequenced message: %v", err)
			return
		} else {
			err = proto.Unmarshal(sequencedMsg.GetAosProto(), newMsg)
			if err != nil {
				log.Printf("W! Message seq_num not supported: %v trying unsequenced", err)
				err = proto.Unmarshal(msgBuf, newMsg)
				if err != nil {
					log.Printf("W! Error unmarshaling: %v", err)
					return
				}
				ssl.IsSequencedStream = false
			} else {
				ssl.IsSequencedStream = true
			}
		}
	})

	if ssl.IsSequencedStream {
		err := proto.Unmarshal(msgBuf, sequencedMsg)
		if err != nil {
			log.Printf("W! Error unmarshaling sequenced message: %v", err)
			return
		} else {
			err = proto.Unmarshal(sequencedMsg.GetAosProto(), newMsg)
			if err != nil {
				log.Printf("W! Error unmarshaling sequenced aos_proto message: %v", err)
			}
		}
	} else {
		// Handle the original unsequenced message format
		err := proto.Unmarshal(msgBuf, newMsg)
		if err != nil {
			log.Printf("W! Error unmarshaling: %v", err)
			return
		}
	}

	// ----------------------------------------------------------------
	// Extract all Types of data
	// ----------------------------------------------------------------
	newPerfMonData := newMsg.GetPerfMon()
	newEventData := newMsg.GetEvent()
	newAlertData := newMsg.GetAlert()

	// ----------------------------------------------------------------
	// Set tags
	// ----------------------------------------------------------------
	originName := newMsg.GetOriginName()
	tags := ssl.GetTags(originName)
	if ssl.IsSequencedStream {
		tags["blueprint"] = newMsg.GetBlueprintLabel()
		tags["device_name"] = newMsg.GetOriginHostname()
		tags["device"] = newMsg.GetOriginHostname()
		tags["role"] = newMsg.GetOriginRole()
	}

	if newPerfMonData != nil {
		if ssl.IsSequencedStream {
			// Initialize the sequence number from the first packet we see
			ssl.perfmonSeqInit.Do(func() {
				ssl.PerfmonSequence = sequencedMsg.GetSeqNum()
				ssl.PerfmonSeqGap = 0
				ssl.reportMessageLoss("perfmon", 0, 0)
			})
			// The gap is used trigger a message loss event on transition, e.g.: 0 -> 2 -> 0
			seqGap := sequencedMsg.GetSeqNum() - ssl.PerfmonSequence
			if seqGap != ssl.PerfmonSeqGap {
				log.Printf("W! Perfmon sequence number expected %d got %d", ssl.PerfmonSequence, sequencedMsg.GetSeqNum())
				ssl.reportMessageLoss("perfmon", ssl.PerfmonSequence, sequencedMsg.GetSeqNum())
				ssl.PerfmonSequence = sequencedMsg.GetSeqNum() + 1
				ssl.PerfmonSeqGap = seqGap
			} else {
				ssl.PerfmonSequence++
			}

		}

		newIntCounter := newPerfMonData.GetInterfaceCounters()
		newResourceCounter := newPerfMonData.GetSystemResourceCounters()
		newGenericPerfMon := newPerfMonData.GetGeneric()
		newProbeMessage := newPerfMonData.GetProbeMessage()

		// ----------------------------------------------------------------
		// Interface Counters
		// ----------------------------------------------------------------
		if newIntCounter != nil {
			ssl.ExtractIntfData(newIntCounter, tags)
		}

		// ----------------------------------------------------------------
		// Resource Counters
		// ----------------------------------------------------------------
		if newResourceCounter != nil {
			systemInfo := newResourceCounter.GetSystemInfo()
			processInfo := newResourceCounter.GetProcessInfo()
			fileInfo := newResourceCounter.GetFileInfo()

			if systemInfo != nil {
				ssl.ExtractSystemInfo(systemInfo, tags)
			}

			if processInfo != nil {
				ssl.ExtractProcessInfo(processInfo, tags)
			}

			if fileInfo != nil {
				ssl.ExtractFileInfo(fileInfo, tags)
			}

		}

		if newProbeMessage != nil {
			ssl.ExtractProbeData(newProbeMessage, tags)
		}

		if newGenericPerfMon != nil {
			serie := "perfmon_generic_undefined"
			fields := make(map[string]interface{})

			for _, t := range newGenericPerfMon.GetTags() {
				tName := t.GetName()
				tValue := t.GetValue()

				myValueOfName := reflect.ValueOf(tValue).Elem()
				myType := myValueOfName.Type().String()

				// Intercept the special tag "data_type"
				if tName == "data_type" {
					serie = t.GetStringValue()
					continue
				}

				switch myType {
				case "aos_streaming.Tag_StringValue":
					tags[tName] = t.GetStringValue()
				case "aos_streaming.Tag_FloatValue":
					log.Printf("W! Perfmon Generic - Tag can only be of type String, %v is type Float", tName)
				case "aos_streaming.Tag_Int64Value":
					log.Printf("W! Perfmon Generic - Tag can only be of type String, %v is type Int64", tName)
				}
			}
			for _, f := range newGenericPerfMon.GetFields() {
				fName := f.GetName()
				fValue := f.GetValue()

				myValueOfValue := reflect.ValueOf(fValue).Elem()
				myType := myValueOfValue.Type().String()

				switch myType {
				case "aos_streaming.Field_FloatValue":
					fields[fName] = f.GetFloatValue()
				case "aos_streaming.Field_Int64Value":
					fields[fName] = f.GetInt64Value()
				case "aos_streaming.Field_StringValue":
					log.Printf("W! Perfmon Generic - Field %v can't be of type String, must be Float of Int64", fName)
				}
			}

			ssl.Aos.Accumulator.AddFields(serie, fields, tags)
		}
	}

	if newEventData != nil {

		// ----------------------------------------------------------------
		// Collect all type of events
		// ----------------------------------------------------------------
		if ssl.IsSequencedStream {
			ssl.eventSeqInit.Do(func() {
				ssl.EventSequence = sequencedMsg.GetSeqNum()
				ssl.EventSeqGap = 0
				ssl.reportMessageLoss("event", 0, 0)
			})
			seqGap := sequencedMsg.GetSeqNum() - ssl.EventSequence
			if seqGap != ssl.EventSeqGap {
				log.Printf("W! Event sequence number expected %d got %d", ssl.EventSequence, sequencedMsg.GetSeqNum())
				ssl.reportMessageLoss("event", ssl.EventSequence, sequencedMsg.GetSeqNum())
				ssl.EventSequence = sequencedMsg.GetSeqNum() + 1
				ssl.EventSeqGap = seqGap
			} else {
				ssl.EventSequence++
			}
		}

		myEventValue := reflect.ValueOf(newEventData.Data).Elem()
		myEventType := myEventValue.Type()
		propType := proto.GetProperties(myEventType)
		eventTypeName := propType.Prop[0].OrigName

		// Convert event name to method name to eliminate lengthy switch statement
		// E.g., bgp_neighbor => GetBgpNeighbor
		method := getFuncName(eventTypeName)
		checkMethod := reflect.ValueOf(newEventData).MethodByName(method)
		if !checkMethod.IsValid() {
			log.Printf("W! Event Type - %s, not supported yet", eventTypeName)
		} else {
			myEventData := reflect.ValueOf(newEventData).MethodByName(method).Call([]reflect.Value{})
			ssl.ExtractEventData(eventTypeName, tags, myEventData[0].Interface())
		}
	}

	if newAlertData != nil {
		if ssl.IsSequencedStream {
			ssl.alertSeqInit.Do(func() {
				ssl.AlertSequence = sequencedMsg.GetSeqNum()
				ssl.AlertSeqGap = 0
				ssl.reportMessageLoss("alert", 0, 0)
			})
			seqGap := sequencedMsg.GetSeqNum() - ssl.AlertSequence
			if seqGap != ssl.AlertSeqGap {
				log.Printf("W! Alert sequence number expected %d got %d", ssl.AlertSequence, sequencedMsg.GetSeqNum())
				ssl.reportMessageLoss("alert", ssl.AlertSequence, sequencedMsg.GetSeqNum())
				ssl.AlertSequence = sequencedMsg.GetSeqNum() + 1
				ssl.AlertSeqGap = seqGap
			} else {
				ssl.AlertSequence++
			}
		}

		myAlertValue := reflect.ValueOf(newAlertData.Data).Elem()
		myAlertType := myAlertValue.Type()
		propAlertType := proto.GetProperties(myAlertType)
		alertTypeName := propAlertType.Prop[0].OrigName
		tags["severity"] = fmt.Sprintf("%v", newAlertData.Severity)
		raise := *newAlertData.Raised

		method := getFuncName(alertTypeName)
		checkMethod := reflect.ValueOf(newAlertData).MethodByName(method)
		if !checkMethod.IsValid() {
			log.Printf("W! Alert Type - %s, not supported yet", alertTypeName)
		} else {
			myAlertData := reflect.ValueOf(newAlertData).MethodByName(method).Call([]reflect.Value{})
			ssl.ExtractAlertData(alertTypeName, tags, myAlertData[0].Interface(), raise)
		}
	}
}

// ----------------------------------------------------------------
// Aos "Class"
// ----------------------------------------------------------------
type Aos struct {
	Port          uint16
	Address       string
	StreamingType []string

	AosServer   string
	AosPort     int
	AosLogin    string
	AosPassword string
	AosProtocol string

	RefreshInterval int

	api *aosrestapi.AosServerApi
	telegraf.Accumulator
	io.Closer
	EventSequence     uint64
	AlertSequence     uint64
	PerfmonSequence   uint64
	EventSeqGap       uint64
	AlertSeqGap       uint64
	PerfmonSeqGap     uint64
	IsSequencedStream bool
	CredentialsValid  bool
	sequencingMode    sync.Once
	perfmonSeqInit    sync.Once
	alertSeqInit      sync.Once
	eventSeqInit      sync.Once
	msgLoss           sync.Mutex
}

func (aos *Aos) Description() string {
	return "input Plugin for Apstra AOS Telemetry Streaming"
}

func (aos *Aos) SampleConfig() string {
	return `

  ## TCP Port to listen for incoming sessions from the AOS Server.
  port = 7777

  ## Address of the server running Telegraf, it needs to be reacheable from AOS.
  address = "192.168.59.1"

  ## Interval to refresh content from the AOS server (in sec).
  ## Only used with AOS versions prior to 3.2
  # refresh_interval = 30

  ## Streaming Type Can be "perfmon", "alerts" and/or "events".
  streaming_type = [ "perfmon", "alerts" ]

  ## Define parameter to join the AOS Server using the REST API.
  ## These parameters are necessary for the receiver to configure the endpoints in AOS
  ## or if you are using versions of AOS prior to 3.2.
  aos_server = "192.168.59.250"
  aos_port = 443
  aos_login = "admin"
  aos_password = "admin"
  aos_protocol = "https"
`
}

func (aos *Aos) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Continuous Query that will refresh data every aos.RefreshInterval sec
func (aos *Aos) RefreshData() {
	for {
		time.Sleep(time.Duration(aos.RefreshInterval) * time.Second)
		aos.api.GetBlueprints()
		aos.api.GetSystems()
		log.Printf("D! Finished to Refresh Data, will sleep for %v sec", aos.RefreshInterval)
	}
}

func (aos *Aos) Start(acc telegraf.Accumulator) error {
	aos.Accumulator = acc

	log.Printf("D! Starting input:aos, will connect to AOS server %v:%v ", aos.AosServer, aos.AosPort)

	// --------------------------------------------
	// Open Session to Rest API
	// --------------------------------------------
	aos.api = aosrestapi.NewAosServerApi(aos.AosServer, aos.AosPort, aos.AosLogin, aos.AosPassword, aos.AosProtocol)

	err := aos.api.Login()
	if err != nil {
		log.Printf("W! Error %+v", err)
		aos.CredentialsValid = false
	} else {
		log.Printf("I! Session to AOS server Opened on %v://%v:%v", aos.AosProtocol, aos.AosServer, aos.AosPort)
		aos.CredentialsValid = true
	}

	if aos.CredentialsValid {
		err = aos.api.GetVersion()
		if err != nil {
			log.Printf("W! Can't determine AOS Version: %v", err)
		} else {
			major, _ := strconv.Atoi(aos.api.AosVersion.Major)
			minor, _ := strconv.Atoi(aos.api.AosVersion.Minor)
			if major > 3 || (major == 3 && minor >= 2) {
				aos.IsSequencedStream = true
				log.Printf("I! AOS Version: %v supports sequenced messaging", aos.api.AosVersion.Version)
			} else {
				aos.IsSequencedStream = false
				log.Printf("I! AOS Version: %v does not support sequenced messaging", aos.api.AosVersion.Version)
			}
		}

		if !aos.IsSequencedStream {
			// Collect Blueprint and System info to augment messages when AOS does not support sequenced messaging and augmented messages
			err = aos.api.GetBlueprints()
			if err != nil {
				log.Printf("W! Error fetching GetBlueprints: %v", err)
			}

			err = aos.api.GetSystems()
			if err != nil {
				log.Printf("W! Error fetching GetSystems: %v", err)
			}

			for _, system := range aos.api.Systems {

				if system.Status.BlueprintId != "" {
					log.Printf("I! Id: %v - %v %s | %v", system.DeviceKey, system.UserConfig.AdminState, system.Status.BlueprintId, system.Blueprint.Role)
				} else {
					log.Printf("I! Id: %v - %v", system.DeviceKey, system.UserConfig.AdminState)
				}
			}

			// Launch Data Refresh in the Background
			go aos.RefreshData()
		}
	}

	// --------------------------------------------
	// Start Listening on TCP port
	// --------------------------------------------

	listenOn := fmt.Sprintf("0.0.0.0:%v", aos.Port)
	l, err := net.Listen("tcp", listenOn)
	if err != nil {
		return err
	}

	log.Printf("I! Listening on port %v", aos.Port)

	ssl := &StreamAos{
		Listener: l,
		Aos:      aos,
	}

	if aos.CredentialsValid {
		// --------------------------------------------
		// Configure Streaming on Server
		// --------------------------------------------
		for _, st := range aos.StreamingType {
			err = aos.api.StartStreaming(st, aos.Address, aos.Port, aos.IsSequencedStream)

			if err != nil {
				log.Printf("W! Unable to configure Streaming %v to %v:%v - %v", st, aos.Address, aos.Port, err)
			} else {
				log.Printf("I! Streaming of %v configured to %v:%v", st, aos.Address, aos.Port)
			}
		}
	}

	go ssl.listen()
	return nil
}

func (aos *Aos) Stop() {
	if aos.Closer != nil {
		aos.Close()
		aos.Closer = nil
	}

	err := aos.api.StopStreaming()
	if err != nil {
		log.Printf("W! Error while stopping Streaming - %v", err)
	} else {
		log.Printf("I! Streaming stopped Successfully")
	}
}

func init() {
	inputs.Add("aos", func() telegraf.Input {
		return &Aos{
			RefreshInterval: 30,
			AosPort:         443,
			AosProtocol:     "https",
			AosLogin:        "admin",
			AosPassword:     "admin",
			EventSequence:   0,
			AlertSequence:   0,
			PerfmonSequence: 0,
		}
	})
}
