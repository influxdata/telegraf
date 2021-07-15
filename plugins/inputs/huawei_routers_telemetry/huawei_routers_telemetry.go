package huawei_routers_telemetry

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/DamRCorba/huawei_sensors/huawei-telemetry"
	//"github.com/DamRCorba/huawei_sensors/bfd_rten"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-bgp"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-devm"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-driver"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-ifm"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-isis"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-mpls"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-ospfv2"
	//"github.com/DamRCorba/huawei_sensors/huaweiV8R10-ospfv3"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-qos"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-sem"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-telemEmdi"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R10-trafficmng"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R12-debug"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R12-devm"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R12-ifm"
	"github.com/DamRCorba/huawei_sensors/huaweiV8R12-qos"

	"github.com/golang/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type HuaweiRoutersTelemetry struct {
	ServicePort     string          `toml:"service_port"`
	VrpVersion      string          `toml:"vrp_version"`
	ReadBufferSize  config.Size     `toml:"read_buffer_size"`
	ContentEncoding string          `toml:"content_encoding"`
	Log             telegraf.Logger `toml:"-"`
	connection      net.PacketConn
	decoder         internal.ContentDecoder

	wg sync.WaitGroup

	acc telegraf.Accumulator
	io.Closer
}

/*
  Telemetry Decoder.
  @params {byte[]} body - message body
  @params {HuaweiRoutersTelemetry} h - HuaweiRoutersTelemetry structure
*/
func HuaweiTelemetryDecoder(body []byte, h *HuaweiRoutersTelemetry) (*metric.SeriesGrouper, error) {
	msg := &huawei_telemetry.Telemetry{}
	err := proto.Unmarshal(body[12:], msg)
	if err != nil {
		h.Log.Error("Unable to decode incoming packet:  %v", err)
		return nil, err
	}
	grouper := metric.NewSeriesGrouper()
	for _, gpbkv := range msg.GetDataGpb().GetRow() {
		dataTime := gpbkv.Timestamp
		if dataTime == 0 {
			dataTime = msg.MsgTimestamp
		}
		timestamp := time.Unix(0, int64(dataTime)*1000000)
		//sensorMsg := huawei_routers_sensorPath.GetMessageType(msg.GetSensorPath())
		sensorMsg := GetMessageType(msg.GetSensorPath(), h.VrpVersion)
		err = proto.Unmarshal(gpbkv.Content, sensorMsg)
		if err != nil {
			h.Log.Error("Sensor Error:  %v", err)
			return nil, err
		}
		//fields, vals := huawei_routers_sensorPath.SearchKey(gpbkv, msg.GetSensorPath())
		fields, vals := SearchKey(gpbkv, msg.GetSensorPath(), h.VrpVersion)
		tags := make(map[string]string, len(fields)+3)
		tags["source"] = msg.GetNodeIdStr()
		tags["subscription"] = msg.GetSubscriptionIdStr()
		tags["path"] = msg.GetSensorPath()
		// Search for Tags
		for i := 0; i < len(fields); i++ {
			//tags = huawei_routers_sensorPath.AppendTags(fields[i], vals[i], tags, msg.GetSensorPath())
			tags = AppendTags(fields[i], vals[i], tags, msg.GetSensorPath(), h.VrpVersion)
		}
		// Create Metrics
		for i := 0; i < len(fields); i++ {
			CreateMetrics(grouper, tags, timestamp, msg.GetSensorPath(), fields[i], vals[i], h.VrpVersion)
		}
	}
	return grouper, nil
}

/*
  Listen UDP packets and call the telemetryDecoder.
*/
func (h *HuaweiRoutersTelemetry) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := h.connection.ReadFrom(buf)
		if err != nil {
			h.Log.Error("Unable to read buffer: %v", err)
			break
		}

		body, err := h.decoder.Decode(buf[:n])
		if err != nil {
			h.Log.Errorf("Unable to decode incoming packet: %v", err)
			continue
		}
		// Telemetry parsing over packet payload
		grouper, err := HuaweiTelemetryDecoder(body, h)
		if err != nil {
			h.Log.Errorf("Unable to decode telemetry information: %v", err)
			break
		}
		for _, metric := range grouper.Metrics() {
			h.acc.AddMetric(metric)
		}

		if err != nil {
			h.Log.Errorf("Unable to parse incoming packet: %v", err)
		}
	}
}

/* Returns the protoMessage of the sensor path.
Huawei have only a few sensors paths for metrics.
The sensors could be known with the command. "display telemetry sensor-path "
@params: path (string) - The head of the sensor path. Example: "huawei-ifm"
@returns: sensor-path proto message
*/
func GetMessageType(path string, version string) proto.Message {
	sensorType := strings.Split(path, ":")
	switch sensorType[0] {
	/* case "huawei-bfd":
	   return &huaweiV8R10_bfd.Bfd{}
	*/

	case "huawei-bgp":
		switch sensorType[1] {
		case "ESTABLISHED":
			return &huaweiV8R10_bgp.ESTABLISHED{}
		case "BACKWARD":
			return &huaweiV8R10_bgp.BACKWARD{}
		}
		return &huaweiV8R10_bgp.ESTABLISHED{}

	case "huawei-devm":
		if version == "V8R10" {
			return &huaweiV8R10_devm.Devm{}
		} else {
			return &__huaweiV8R12_devm.Devm{}
		}

	case "huawei-driver":
		switch sensorType[1] {
		case "hwEntityInvalid":
			return &huaweiV8R10_driver.HwEntityInvalid{}
		case "hwEntityResume":
			return &huaweiV8R10_driver.HwEntityResume{}
		case "hwOpticalInvalid":
			return &huaweiV8R10_driver.HwOpticalInvalid{}
		case "hwOpticalInvalidResume":
			return &huaweiV8R10_driver.HwOpticalInvalidResume{}
		}
		return &huaweiV8R10_driver.HwEntityInvalid{}

	case "huawei-ifm":
		if version == "V8R10" {
			return &huaweiV8R10_ifm.Ifm{}
		} else {
			return &huaweiV8R12_ifm.Ifm{}
		}

	case "huawei-isis":
	case "huawei-isiscomm":
		return &huaweiV8R10_isiscomm.IsisAdjacencyChange{}

	case "huawei-mpls":
		return &huaweiV8R10_mpls.Mpls{}

	case "huawei-ospfv2":
		switch sensorType[1] {
		case "ospfNbrStateChange":
			return &huaweiV8R10_ospfv2.OspfNbrStateChange{}
		case "ospfVirtNbrStateChange":
			return &huaweiV8R10_ospfv2.OspfVirtNbrStateChange{}
		}
		return &huaweiV8R10_ospfv2.OspfNbrStateChange{}

	//case "huawei-ospfv3":
	//  return &huaweiV8R10_ospfv3.Ospfv3NbrStateChange{}

	case "huawei-qos":
		if version == "V8R10" {
			return &huaweiV8R10_qos.Qos{}
		} else {
			return &huaweiV8R12_qos.Qos{}
		}

	case "huawei-sem":
		switch sensorType[1] {
		case "hwCPUUtilizationResume":
			return &huaweiV8R10_sem.HwStorageUtilizationResume{}
		case "hwCPUUtilizationRisingAlarm":
			return &huaweiV8R10_sem.HwCPUUtilizationRisingAlarm{}
		case "hwStorageUtilizationResume":
			return &huaweiV8R10_sem.HwStorageUtilizationResume{}
		case "hwStorageUtilizationRisingAlarm":
			return &huaweiV8R10_sem.HwStorageUtilizationRisingAlarm{}
		}
		return &huaweiV8R10_sem.HwStorageUtilizationResume{}

	case "huawei-telmEmdi":
	case "huawei-emdi":
		return &huaweiV8R10_telemEmdi.TelemEmdi{}

	case "huawei-trafficmng":
		return &huaweiV8R10_trafficmng.Trafficmng{}

	case "huawei-debug":
		if len(sensorType) == 1 {
			return &huaweiV8R12_debug.Debug{}
		} else {
			switch sensorType[1] {
			case "debug/cpu-infos/cpu-info":
				return &huaweiV8R12_debug.Debug{}
			case "debug/memory-infos/memory-info":
				return &huaweiV8R12_debug.Debug{}

			default:
				return &huaweiV8R12_debug.Debug{}
			}
		}
	default:
		//	fmt.Println("Error Sensor Desconocido en GetMessageType", path)
		return &huaweiV8R10_devm.Devm{}
	}
	return &huaweiV8R10_devm.Devm{}
}

/*
  Get the types of the Telemetry EndPoint
  @Params: a string with the telemetry complete path
  @Returns: a Map with keys and types of the endpoint
*/
func GetTypeValue(path string, version string) map[string]reflect.Type {
	resolve := make(map[string]reflect.Type)
	splited := strings.Split(path, ":")
	switch splited[0] {
	case "huawei-bfd":
		// TODO bfd is not working
		return resolve

	case "huawei-bgp":
		switch splited[1] {
		case "ESTABLISHED":
			fooType := reflect.TypeOf(huaweiV8R10_bgp.ESTABLISHED{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "BACKWARD":
			fooType := reflect.TypeOf(huaweiV8R10_bgp.BACKWARD{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break

		}
		return resolve

	case "huawei-devm":
		if version != "V8R10" {
			fooType := reflect.TypeOf(__huaweiV8R12_devm.Devm{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			return resolve
		}
		switch splited[1] {
		case "devm/cpuInfos/cpuInfo":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_CpuInfos_CpuInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "devm/fans/fan":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_Fans_Fan{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "devm/memoryInfos/memoryInfo":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_MemoryInfos_MemoryInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "devm/ports/port":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_Ports_Port{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "devm/ports/port/opticalInfo":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_Ports_Port{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			fooType = reflect.TypeOf(huaweiV8R10_devm.Devm_Ports_Port_OpticalInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "RxPower" || keys.Name == "TxPower" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf(1.0)
				} else {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			break
		case "devm/powerSupplys/powerSupply/powerEnvironments/powerEnvironment":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_PowerSupplys_PowerSupply_PowerEnvironments_PowerEnvironment{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "PowerValue" || keys.Name == "VoltageValue" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf(1.0)
				} else {
					if keys.Name == "PemIndex" {
						resolve[LcFirst(keys.Name)] = reflect.TypeOf("")
					} else {
						resolve[LcFirst(keys.Name)] = keys.Type
					}
				}
			}
			break
		case "devm/temperatureInfos/temperatureInfo":
			fooType := reflect.TypeOf(huaweiV8R10_devm.Devm_TemperatureInfos_TemperatureInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "I2c" || keys.Name == "Channel" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf("")
				} else {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			break
		}

		return resolve

	case "huawei-driver":
		switch splited[1] {
		case "hwEntityInvalid":
			fooType := reflect.TypeOf(huaweiV8R10_driver.HwEntityInvalid{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "I2c" || keys.Name == "Channel" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf("")
				} else {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			break
		case "hwEntityResume":
			fooType := reflect.TypeOf(huaweiV8R10_driver.HwEntityResume{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "I2c" || keys.Name == "Channel" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf("")
				} else {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			break
		case "hwOpticalInvalid":
			fooType := reflect.TypeOf(huaweiV8R10_driver.HwOpticalInvalid{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "I2c" || keys.Name == "Channel" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf("")
				} else {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			break
		case "hwOpticalInvalidResume":
			fooType := reflect.TypeOf(huaweiV8R10_driver.HwOpticalInvalidResume{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "I2c" || keys.Name == "Channel" {
					resolve[LcFirst(keys.Name)] = reflect.TypeOf("")
				} else {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			break

		}
		return resolve

	case "huawei-ifm":
		if version == "V8R10" {
			fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "IfName" {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			switch splited[1] {
			case "ifm/interfaces/interface": // No trae data mas que IfIndex, IfName e IfAdminStatus_UP si la interface esta Down no devuevle el campo.
				fooType = reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifClearedStat":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfClearedStat{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifDynamicInfo":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfDynamicInfo{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifStatistics":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfStatistics{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifStatistics/ethPortErrSts":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfStatistics_EthPortErrSts{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			}
			return resolve
		} else { // V8R11 & V8R12
			fooType := reflect.TypeOf(huaweiV8R12_ifm.Ifm_Interfaces_Interface{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				if keys.Name == "Name" {
					resolve[LcFirst(keys.Name)] = keys.Type
				}
			}
			switch splited[1] {
			case "ifm/interfaces/interface": // No trae data mas que IfIndex, IfName e IfAdminStatus_UP si la interface esta Down no devuevle el campo.
				fooType = reflect.TypeOf(huaweiV8R12_ifm.Ifm_Interfaces_Interface{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifClearedStat":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfClearedStat{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifDynamicInfo":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfDynamicInfo{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifStatistics":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfStatistics{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			case "ifm/interfaces/interface/ifStatistics/ethPortErrSts":
				fooType := reflect.TypeOf(huaweiV8R10_ifm.Ifm_Interfaces_Interface_IfStatistics_EthPortErrSts{})
				for i := 0; i < fooType.NumField(); i++ {
					keys := fooType.Field(i)
					resolve[LcFirst(keys.Name)] = keys.Type
				}
				break
			}
			return resolve
		}
	case "huawei-isiscomm":
	case "huawei-isis":
		fooType := reflect.TypeOf(huaweiV8R10_isiscomm.IsisAdjacencyChange{})
		for i := 0; i < fooType.NumField(); i++ {
			keys := fooType.Field(i)
			resolve[LcFirst(keys.Name)] = keys.Type
		}
		return resolve

	case "huawei-mpls":
		fooType := reflect.TypeOf(huaweiV8R10_mpls.Mpls{})
		for i := 0; i < fooType.NumField(); i++ {
			keys := fooType.Field(i)
			resolve[LcFirst(keys.Name)] = keys.Type
		}
		return resolve

	case "huawei-ospfv2":
		switch splited[1] {
		case "ospfNbrStateChange":
			fooType := reflect.TypeOf(huaweiV8R10_ospfv2.OspfNbrStateChange{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "ospfVirtNbrStateChange":
			fooType := reflect.TypeOf(huaweiV8R10_ospfv2.OspfVirtNbrStateChange{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		}
		return resolve

		/* case "huawei-ospfv3":
		   fooType := reflect.TypeOf(huaweiV8R10_ospfv3.Ospfv3NbrStateChange{})
		   for i := 0; i < fooType.NumField(); i ++ {
		     keys := fooType.Field(i)
		     resolve[LcFirst(keys.Name)] = keys.Type
		     }
		   return resolve
		*/
	case "huawei-qos":
		switch splited[1] {
		case "qos/qosBuffers/qosBuffer":
			fooType := reflect.TypeOf(huaweiV8R10_qos.Qos_QosBuffers_QosBuffer{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "qos/qosIfQoss/qosIfQos/qosPolicyApplys/qosPolicyApply/qosPolicyStats/qosPolicyStat/qosRuleStats/qosRuleStat":
			fooType := reflect.TypeOf(huaweiV8R10_qos.Qos_QosIfQoss_QosIfQos_QosPolicyApplys_QosPolicyApply_QosPolicyStats_QosPolicyStat_QosRuleStats_QosRuleStat{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "qos/qosPortQueueStatInfos/qosPortQueueStatInfo":
			fooType := reflect.TypeOf(huaweiV8R10_qos.Qos_QosPortQueueStatInfos_QosPortQueueStatInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		}
		return resolve

	case "huawei-sem":
		switch splited[1] {
		case "hwCPUUtilizationResume":
			fooType := reflect.TypeOf(huaweiV8R10_sem.HwStorageUtilizationResume{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "hwCPUUtilizationRisingAlarm":
			fooType := reflect.TypeOf(huaweiV8R10_sem.HwCPUUtilizationRisingAlarm{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "hwStorageUtilizationResume":
			fooType := reflect.TypeOf(huaweiV8R10_sem.HwStorageUtilizationResume{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "hwStorageUtilizationRisingAlarm":
			fooType := reflect.TypeOf(huaweiV8R10_sem.HwStorageUtilizationRisingAlarm{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		}
		return resolve

	case "huawei-telmEmdi":
	case "huawei-emdi":
		switch splited[1] {
		case "emdi/emdiTelemReps/emdiTelemRep":
			fooType := reflect.TypeOf(huaweiV8R10_telemEmdi.TelemEmdi_EmdiTelemReps_EmdiTelemRep{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "emdi/emdiTelemRtps/emdiTelemRtp":
			fooType := reflect.TypeOf(huaweiV8R10_telemEmdi.TelemEmdi_EmdiTelemRtps_EmdiTelemRtp{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		}
		return resolve

	case "huawei-trafficmng":
		switch splited[1] {
		case "trafficmng/tmSlotSFUs/tmSlotSFU/sfuStatisticss/sfuStatistics":
			fooType := reflect.TypeOf(huaweiV8R10_trafficmng.Trafficmng{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			fooType = reflect.TypeOf(huaweiV8R10_trafficmng.Trafficmng_TmSlotSFUs{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			fooType = reflect.TypeOf(huaweiV8R10_trafficmng.Trafficmng_TmSlotSFUs_TmSlotSFU{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			fooType = reflect.TypeOf(huaweiV8R10_trafficmng.Trafficmng_TmSlotSFUs_TmSlotSFU_SfuStatisticss{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			fooType = reflect.TypeOf(huaweiV8R10_trafficmng.Trafficmng_TmSlotSFUs_TmSlotSFU_SfuStatisticss_SfuStatistics{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break

		}
		return resolve
		/* Sensores V8R12 */
	case "huawei-debug":
		switch splited[1] {
		case "debug/cpu-infos/cpu-info":
			fooType := reflect.TypeOf(huaweiV8R12_debug.Debug_CpuInfos_CpuInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		case "debug/memory-infos/memory-info":
			fooType := reflect.TypeOf(huaweiV8R12_debug.Debug_MemoryInfos_MemoryInfo{})
			for i := 0; i < fooType.NumField(); i++ {
				keys := fooType.Field(i)
				resolve[LcFirst(keys.Name)] = keys.Type
			}
			break
		}
		return resolve
	default:
		//fmt.Println("Error Sensor Desconocido", path)
		//fmt.Println(path)

		return resolve
	}
	return resolve
}

/*
  Change the firts character of a string to Lowercase
*/
func LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

/*
  Change the firts character of a string to Uppercase
*/
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func (h *HuaweiRoutersTelemetry) Description() string {
	return "Input plugin for receiving Huawei Router Telemetry data via UDP"
}

func (h *HuaweiRoutersTelemetry) SampleConfig() string {
	return `
  ## UDP Service Port to capture Telemetry
  # service_port = "5600"
  # vrp_version = "V8R10"
  # read_buffer_size = "64KiB"

`
}

func (h *HuaweiRoutersTelemetry) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *HuaweiRoutersTelemetry) Start(acc telegraf.Accumulator) error {
	h.acc = acc

	var err error
	h.decoder, err = internal.NewContentDecoder(h.ContentEncoding)
	if err != nil {
		return err
	}

	pc, err := udpListen(":" + h.ServicePort)
	if err != nil {
		return err
	}

	if h.ReadBufferSize > 0 { // internal.Size
		if srb, ok := pc.(setReadBufferer); ok {
			srb.SetReadBuffer(int(h.ReadBufferSize)) //internal.Size
		} else {
			h.Log.Warnf("Unable to set read buffer on a %s socket", "udp")
		}
	}

	h.Log.Infof("Listening Routers on port %s", pc.LocalAddr())
	h.connection = pc

	h.wg = sync.WaitGroup{}
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.listen()
	}()
	return nil
}

/*
  Creates and add metrics from json mapped data in telegraf metrics SeriesGrouper
  @params:
  - grouper (*metric.SeriesGrouper) - pointer of metric series to append data.
  - tags (map[string]string) json data mapped
  - timestamp (time.Time) -
  - path (string) - sensor path
  - subfield (string) - subkey data.
    vals (string) - subkey content

*/
func CreateMetrics(grouper *metric.SeriesGrouper, tags map[string]string, timestamp time.Time, path string, subfield string, vals string, version string) {
	if subfield == "ifAdminStatus" {
		name := strings.Replace(subfield, "\"", "", -1)
		if vals == "IfAdminStatus_UP" {
			grouper.Add(path, tags, timestamp, string(name), 1)
		} else {
			grouper.Add(path, tags, timestamp, string(name), 0)
		}
	}
	if subfield == "ifOperStatus" {
		name := strings.Replace(subfield, "\"", "", -1)
		if vals == "IfOperStatus_UP" {
			grouper.Add(path, tags, timestamp, string(name), 1)
		} else {
			grouper.Add(path, tags, timestamp, string(name), 0)
		}
	}
	if vals != "" && subfield != "ifName" && subfield != "position" && subfield != "pemIndex" && subfield != "address" && subfield != "i2c" && subfield != "channel" &&
		subfield != "queueType" && subfield != "ifAdminStatus" && subfield != "ifOperStatus" {
		name := strings.Replace(subfield, "\"", "", -1)
		//endPointTypes:=huawei_routers_sensorPath.GetTypeValue(path)
		endPointTypes := GetTypeValue(path, version)
		grouper.Add(path, tags, timestamp, string(name), decodeVal(endPointTypes[name], vals))
	}
}

/*
  Append to the tags the telemetry values for position.
  @params:
  k - Key to evaluate
  v - Content of the Key
  tags - Global tags of the metric
  path - Telemetry path
  @returns
  original tag append the key if its a name Key.

*/
func AppendTags(k string, v string, tags map[string]string, path string, version string) map[string]string {
	resolve := tags
	//endPointTypes:=huawei_routers_sensorPath.GetTypeValue(path)
	endPointTypes := GetTypeValue(path, version)
	if endPointTypes[k] != nil {
		if reflect.TypeOf(decodeVal(endPointTypes[k], v)) == reflect.TypeOf("") {
			if k != "ifAdminStatus" {
				resolve[k] = v
			}
		}
	} else {
		if k == "ifName" || k == "position" || k == "pemIndex" || k == "i2c" || k == "name" || k == "description" {
			resolve[k] = v
		}

	}
	return resolve
}

/*
  Convert the telemetry Data to its type.
  @Params:
  tipo - telemetry path data type
  val - string value
  Returns the converted value
*/
func decodeVal(tipo interface{}, val string) interface{} {
	if tipo == nil {
		return val
	} else {
		value := reflect.New(tipo.(reflect.Type)).Elem().Interface()
		switch value.(type) {
		case uint32:
			resolve, _ := strconv.ParseUint(val, 10, 32)
			return resolve
		case uint64:
			resolve, _ := strconv.ParseUint(val, 10, 64)
			return resolve
		case int32:
			resolve, _ := strconv.ParseInt(val, 10, 32)
			return resolve
		case int64:
			resolve, _ := strconv.ParseInt(val, 10, 64)
			return resolve
		case float64:
			resolve, err := strconv.ParseFloat(val, 64)
			if err != nil {
				name := strings.Replace(val, "\"", "", -1)
				resolve, _ = strconv.ParseFloat(name, 64)
			}
			return resolve
		case bool:
			resolve, _ := strconv.ParseBool(val)
			return resolve
		}
	}
	resolve := val
	return resolve
}

/* Search for a string in a string array.
@Params: a String Array
         x String to Search
@Returns: Returns the index location de x in a or -1 if not Found
*/
func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}

/*
  Search de keys and vals of the data row in telemetry message.
  @params:
  - Message (*TelemetryRowGPB) - data buffer GPB of sensor data
  - sensorType (string) - sensor-path group.
  @returns:
  - keys (string) - Keys of the fields
  - vals (string) - Vals of the fields
*/
func SearchKey(Message *huawei_telemetry.TelemetryRowGPB, path string, version string) ([]string, []string) {
	sensorType := strings.Split(path, ":")[0]
	//sensorMsg := huawei_routers_sensorPath.GetMessageType(sensorType)
	sensorMsg := GetMessageType(sensorType, version)
	err := proto.Unmarshal(Message.Content, sensorMsg)
	if err != nil {
		panic(err)
	}
	primero := reflect.ValueOf(sensorMsg).Interface()

	str := fmt.Sprintf("%v", primero)
	// format string to JsonString with some modifications.
	jsonString := strings.Replace(str, "<>", "0", -1)
	jsonString = strings.Replace(jsonString, "<", "{\"", -1)
	jsonString = strings.Replace(jsonString, ">", "\"}", -1)
	jsonString = strings.Replace(jsonString, " ", ",\"", -1)
	jsonString = strings.Replace(jsonString, ":", "\":", -1)
	jsonString = strings.Replace(jsonString, ",\"\"", "", -1)
	jsonString = strings.Replace(jsonString, "},\"", "}", -1)
	jsonString = strings.Replace(jsonString, ",", " ", -1)
	jsonString = strings.Replace(jsonString, "{", " ", -1)
	jsonString = strings.Replace(jsonString, "}", "", -1)
	jsonString = "\"" + jsonString
	if version == "V8R10" {

		if path == "huawei-ifm:ifm/interfaces/interface/ifDynamicInfo" { // caso particular....
			jsonString = strings.Replace(jsonString, "IfOperStatus_UPifName\"", "IfOperStatus_UP \"ifName\"", -1)
		}
	} else {
		// TODO Check V8R12 structure
		jsonString = strings.Replace(jsonString, " interface\": name\"", " \"interface\": \"name\"", -1)
		jsonString = strings.Replace(jsonString, " \" \"", " \"", -1)
		jsonString = strings.Replace(jsonString, "receive_byte\"", "\"receive_byte\"", -1)
		jsonString = strings.Replace(jsonString, "eth_port_err_sts\"", "\"eth_port_err_sts\"", -1)
		jsonString = strings.Replace(jsonString, "rx_jumbo_octets\"", "\"rx_jumbo_octets\"", -1)
		jsonString = strings.Replace(jsonString, "\":0\" ", "_0\" ", -1)
		jsonString = strings.Replace(jsonString, "\":1\" ", "_1\" ", -1)
		jsonString = strings.Replace(jsonString, "\":2\" ", "_2\" ", -1)
		jsonString = strings.Replace(jsonString, "\":3\" ", "_3\" ", -1)
	}
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}

	// splitting string by space but considering quoted section
	items := strings.FieldsFunc(jsonString, f)

	// create and fill the map
	m := make(map[string]string)
	for _, item := range items {
		x := strings.Split(item, ":")
		m[x[0]] = x[1]
	}
	// get keys and vals of fields
	var keys []string
	var vals []string
	for k, v := range m {
		name := strings.Replace(k, "\"", "", -1) // remove quotes
		keys = append(keys, name)
		vals = append(vals, v)

	}
	// Adaptation to resolve Huawei bad struct Data.
	if path == "huawei-ifm:ifm/interfaces/interface" {
		if Find(keys, "ifAdminStatus") == -1 {
			keys = append(keys, "ifAdminStatus")
			vals = append(vals, "IfAdminStatus_DOWN")
		}
	}
	// Adaptation to resolve Huawei bad struct Data.
	if path == "huawei-ifm:ifm/interfaces/interface/ifDynamicInfo" {
		if Find(keys, "ifOperStatus") == -1 {
			keys = append(keys, "ifOperStatus")
			vals = append(vals, "IfOperStatus_DOWN")
		}
	}

	return keys, vals
}

func udpListen(address string) (net.PacketConn, error) {
	var addr *net.UDPAddr
	var err error
	var ifi *net.Interface
	addr, err = net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	if addr.IP.IsMulticast() {
		return net.ListenMulticastUDP("udp", ifi, addr)
	}
	return net.ListenUDP("udp", addr)
}

func (h *HuaweiRoutersTelemetry) Stop() {
	if h.connection != nil {
		h.connection.Close()
	}
	h.wg.Wait()
}

func init() {
	inputs.Add("huawei_routers_telemetry", func() telegraf.Input { return &HuaweiRoutersTelemetry{} })
}
