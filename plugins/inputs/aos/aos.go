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
	"strings"
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

func (ssl *StreamAos) ExtractProbeData(newProbeMessage interface{}, originName string) {
	myValue := reflect.ValueOf(newProbeMessage).Elem()
	myType := myValue.Type()
	propType := proto.GetProperties(myType)

	serie := "probe_message"
	fields := make(map[string]interface{})
	tags := ssl.GetTags(originName)

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
	}

	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) ExtractIntfData(intfData interface{}, originName string) {
	myValue := reflect.ValueOf(intfData).Elem()
	myType := myValue.Type()
	propType := proto.GetProperties(myType)

	serie := "interface_counters"
	fields := make(map[string]interface{})
	tags := ssl.GetTags(originName)

	for i := 0; i < myValue.NumField(); i++ {

		myField := myValue.Field(i)
		field_name := propType.Prop[i].OrigName

		// Skip field with XXX_
		if strings.Contains(field_name, "XXX_") {
			continue
		}

		fields[propType.Prop[i].OrigName] = reflect.Indirect(myField).Interface()
	}

	ssl.Aos.Accumulator.AddFields(serie, fields, tags)
}

func (ssl *StreamAos) ExtractSystemInfo(systemInfo interface{}, originName string) {
	// Prepare value. type and property
	myValue := reflect.ValueOf(systemInfo).Elem()
	myType := myValue.Type()
	propType := proto.GetProperties(myType)

	serie := "system_info"
	fields := make(map[string]interface{})
	tags := ssl.GetTags(originName)

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

func (ssl *StreamAos) ExtractProcessInfo(processInfo []*aos_streaming.ProcessInfo, originName string) {
	for _, p := range processInfo {

		myValue := reflect.ValueOf(p).Elem()
		myType := myValue.Type()
		propType := proto.GetProperties(myType)

		process_name := p.ProcessName
		serie := "process_info"
		fields := make(map[string]interface{})
		tags := ssl.GetTags(originName)

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

func (ssl *StreamAos) ExtractFileInfo(fileInfo []*aos_streaming.FileInfo, originName string) {
	for _, f := range fileInfo {
		file_name := f.FileName
		file_size := f.FileSize

		serie := "file_info"
		fields := make(map[string]interface{})
		tags := ssl.GetTags(originName)

		tags["file_name"] = *file_name
		fields["size"] = *file_size

		ssl.Aos.Accumulator.AddFields(serie, fields, tags)
	}
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

		// Create new aos_streaming.AosMessage and deserialize protobuf
		newMsg := new(aos_streaming.AosMessage)
		err = proto.Unmarshal(msgBuf, newMsg)

		if err != nil {
			log.Printf("W! Error unmarshaling: %v", err)
			return
		}

		// ----------------------------------------------------------------
		// Extract all Types of data
		// ----------------------------------------------------------------
		newPerfMonData := newMsg.GetPerfMon()
		newEvent := newMsg.GetEvent()
		newAlert := newMsg.GetAlert()

		// ----------------------------------------------------------------
		// Extract Timestamp and Device Name
		// ----------------------------------------------------------------
		originName := newMsg.GetOriginName()

		if newPerfMonData != nil {

			newIntCounter := newPerfMonData.GetInterfaceCounters()
			newResourceCounter := newPerfMonData.GetSystemResourceCounters()
			newGenericPerfMon := newPerfMonData.GetGeneric()
			newProbeMessage := newPerfMonData.GetProbeMessage()

			// ----------------------------------------------------------------
			// Interface Counters
			// ----------------------------------------------------------------
			if newIntCounter != nil {
				ssl.ExtractIntfData(newIntCounter, originName)
			}

			// ----------------------------------------------------------------
			// Resource Counters
			// ----------------------------------------------------------------
			if newResourceCounter != nil {

				systemInfo := newResourceCounter.GetSystemInfo()
				processInfo := newResourceCounter.GetProcessInfo()
				fileInfo := newResourceCounter.GetFileInfo()

				if systemInfo != nil {
					ssl.ExtractSystemInfo(systemInfo, originName)
				}

				if processInfo != nil {
					ssl.ExtractProcessInfo(processInfo, originName)
				}

				if fileInfo != nil {
					ssl.ExtractFileInfo(fileInfo, originName)
				}

			}

			if newProbeMessage != nil {
				ssl.ExtractProbeData(newProbeMessage, originName)
			}

			if newGenericPerfMon != nil {

				serie := "perfmon_generic_undefined"
				fields := make(map[string]interface{})
				tags := ssl.GetTags(originName)

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

		if newEvent != nil {

			// ----------------------------------------------------------------
			// Collect all type of events
			// ----------------------------------------------------------------
			myEventValue := reflect.ValueOf(newEvent.Data).Elem()
			myEventType := myEventValue.Type()
			propType := proto.GetProperties(myEventType)

			eventTypeName := propType.Prop[0].OrigName

			tags := ssl.GetTags(originName)

			switch eventTypeName {
			case "device_state":
				myEventData := newEvent.GetDeviceState()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "streaming":
				myEventData := newEvent.GetStreaming()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "cable_peer":
				myEventData := newEvent.GetCablePeer()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "bgp_neighbor":
				myEventData := newEvent.GetBgpNeighbor()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "link_status":
				myEventData := newEvent.GetLinkStatus()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "traffic":
				myEventData := newEvent.GetTraffic()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "mac_state":
				myEventData := newEvent.GetMacState()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "arp_state":
				myEventData := newEvent.GetArpState()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "lag_state":
				myEventData := newEvent.GetLagState()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "mlag_state":
				myEventData := newEvent.GetMlagState()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "extensible_event":
				myEventData := newEvent.GetExtensibleEvent()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)
			case "route_state":
				myEventData := newEvent.GetRouteState()
				ssl.ExtractEventData(eventTypeName, tags, myEventData)

			default:
				log.Printf("W! Event Type - %s, not supported yet", eventTypeName)
			}
		}

		if newAlert != nil {

			myAlertValue := reflect.ValueOf(newAlert.Data).Elem()
			myAlertType := myAlertValue.Type()
			propAlertType := proto.GetProperties(myAlertType)

			alertTypeName := propAlertType.Prop[0].OrigName

			tags := ssl.GetTags(originName)

			tags["severity"] = fmt.Sprintf("%v", newAlert.Severity)

			raise := *newAlert.Raised

			switch alertTypeName {
			case "config_deviation_alert":
				myAlertData := newAlert.GetConfigDeviationAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "streaming_alert":
				myAlertData := newAlert.GetStreamingAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "cable_peer_mismatch_alert":
				myAlertData := newAlert.GetCablePeerMismatchAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "bgp_neighbor_mismatch_alert":
				myAlertData := newAlert.GetBgpNeighborMismatchAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "interface_link_status_mismatch_alert":
				myAlertData := newAlert.GetInterfaceLinkStatusMismatchAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "hostname_alert":
				myAlertData := newAlert.GetHostnameAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "route_alert":
				myAlertData := newAlert.GetRouteAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "liveness_alert":
				myAlertData := newAlert.GetLivenessAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "deployment_alert":
				myAlertData := newAlert.GetDeploymentAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "blueprint_rendering_alert":
				myAlertData := newAlert.GetBlueprintRenderingAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "counters_alert":
				myAlertData := newAlert.GetCountersAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "mac_alert":
				myAlertData := newAlert.GetMacAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "arp_alert":
				myAlertData := newAlert.GetArpAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "headroom_alert":
				myAlertData := newAlert.GetHeadroomAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "lag_alert":
				myAlertData := newAlert.GetLagAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "mlag_alert":
				myAlertData := newAlert.GetMlagAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "probe_alert":
				myAlertData := newAlert.GetProbeAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "config_mismatch_alert":
				myAlertData := newAlert.GetConfigMismatchAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "extensible_alert":
				myAlertData := newAlert.GetExtensibleAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)
			case "test_alert":
				myAlertData := newAlert.GetTestAlert()
				ssl.ExtractAlertData(alertTypeName, tags, myAlertData, raise)

			default:
				log.Printf("W! Alert Type - %s, Not Supported Yet", alertTypeName)
			}
		}
	}
}

// ----------------------------------------------------------------
// Aos "Class"
// ----------------------------------------------------------------
type Aos struct {
	Port          int
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
  # refresh_interval = 30

  ## Streaming Type Can be "perfmon", "alerts" or "events".
  streaming_type = [ "perfmon", "alerts" ]

  ## Define parameter to join the AOS Server using the REST API.
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

// Continuous Query that will refresh data every 15 sec
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
	} else {
		log.Printf("I! Session to AOS server Opened on %v://%v:%v", aos.AosProtocol, aos.AosServer, aos.AosPort)
	}

	// --------------------------------------------
	// Collect Blueprint and System info
	// --------------------------------------------
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

	// --------------------------------------------
	// Configure Streaming on Server
	// --------------------------------------------
	for _, st := range aos.StreamingType {
		err = aos.api.StartStreaming(st, aos.Address, aos.Port)

		if err != nil {
			log.Printf("W! Unable to configure Streaming %v to %v:%v - %v", st, aos.Address, aos.Port, err)
		} else {
			log.Printf("I! Streaming of %v configured to %v:%v", st, aos.Address, aos.Port)
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
		}
	})
}
