// (c) 2019 Sony Interactive Entertainment Inc.
package jtinative

import (
	"log"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/jtinative/fabric"
	"github.com/influxdata/telegraf/plugins/parsers/jtinative/port"
	"github.com/influxdata/telegraf/plugins/parsers/jtinative/telemetry_top"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var telStream telemetry_top.TelemetryStream = telemetry_top.TelemetryStream{
	SystemId:       proto.String("test"),
	ComponentId:    proto.Uint32(65535),
	SubComponentId: proto.Uint32(1),
	SensorName:     proto.String("test sensor"),
	SequenceNumber: proto.Uint32(1),
	Timestamp:      proto.Uint64(0),
}

func buildJtiProto(telTop *telemetry_top.TelemetryStream, protoMessage interface{}, extensionType *proto.ExtensionDesc) ([]byte, error) {

	jnprNetSensor := &telemetry_top.JuniperNetworksSensors{}
	if err := proto.SetExtension(jnprNetSensor, extensionType, protoMessage); err != nil {
		return nil, err
	}

	entSensor := &telemetry_top.EnterpriseSensors{}
	if err := proto.SetExtension(entSensor, telemetry_top.E_JuniperNetworks, jnprNetSensor); err != nil {
		return nil, err
	}

	telTop.Enterprise = entSensor

	data, err := proto.Marshal(telTop)
	return data, err
}

func TestUnpackTelemetryTop(t *testing.T) {
	telTop := telStream
	data, err := proto.Marshal(&telTop)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	newTelTop := &telemetry_top.TelemetryStream{}
	err = proto.Unmarshal(data, newTelTop)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	assert.Equal(t, telTop.GetSystemId(), newTelTop.GetSystemId())
	assert.Equal(t, telTop.GetSensorName(), newTelTop.GetSensorName())
	assert.Equal(t, telTop.GetComponentId(), newTelTop.GetComponentId())
	assert.Equal(t, telTop.GetSubComponentId(), newTelTop.GetSubComponentId())
	assert.Equal(t, telTop.GetTimestamp(), newTelTop.GetTimestamp())
}

func TestNoExtensions(t *testing.T) {
	parser := JTINativeParser{}

	telegrafMetrics := make([]telegraf.Metric, 0)
	telTop := telStream

	data, err := proto.Marshal(&telTop)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	assert.Equal(t, telegrafMetrics, parsedMetric)
}

func TestMissingTelemetryTopValuesNoExtensions(t *testing.T) {
	parser := JTINativeParser{}
	telegrafMetrics := make([]telegraf.Metric, 0)

	telTop := &telemetry_top.TelemetryStream{
		SystemId:  proto.String("test"),
		Timestamp: proto.Uint64(1),
	}
	data, err := proto.Marshal(telTop)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}
	assert.Equal(t, telegrafMetrics, parsedMetric)
}

func TestPortExtensions(t *testing.T) {
	parser := JTINativeParser{}

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{IfName: proto.String("xe-0/0/0"), IfTransitions: proto.Uint64(3)},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.IfTransitions": uint64(3)},
		time.Unix(0, 0),
	)
	assert.Equalf(t, len(parsedMetric), 1, "Expected one metric to be produced")
	if len(parsedMetric) == 1 {
		testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
	}
}

func TestPortNestedMessage(t *testing.T) {
	parser := JTINativeParser{}

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName: proto.String("xe-0/0/0"),
				EgressStats: &port.InterfaceStats{
					If_1SecOctets: proto.Uint64(100000),
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressStats.If_1SecOctets": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestPortRepeatedMessage(t *testing.T) {
	parser := JTINativeParser{}

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName: proto.String("xe-0/0/0"),
				EgressQueueInfo: []*port.QueueStats{
					{Bytes: proto.Uint64(100000)},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestPortRepeatedMessageQueue(t *testing.T) {
	parser := JTINativeParser{}

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName: proto.String("xe-0/0/0"),
				EgressQueueInfo: []*port.QueueStats{
					{
						QueueNumber: proto.Uint32(1),
						Bytes:       proto.Uint64(100000),
					},
					{
						QueueNumber: proto.Uint32(2),
						Bytes:       proto.Uint64(200000),
					},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric1 := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"QueueNumber":  "1",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric1, parsedMetric[0])

	testMetric2 := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"QueueNumber":  "2",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(200000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric2, parsedMetric[1])
}

func TestPortOptionsOverideType(t *testing.T) {
	parser := JTINativeParser{}

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName: proto.String("xe-0/0/0"),
				EgressQueueInfo: []*port.QueueStats{
					{
						QueueNumber: proto.Uint32(1),
						Bytes:       proto.Uint64(100000),
					},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"QueueNumber":  "1",
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestPortStrAsTagTrue(t *testing.T) {
	parser := JTINativeParser{}
	parser.JTIStrAsTag = true

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName:       proto.String("xe-0/0/0"),
				ParentAeName: proto.String("ae0"),
				EgressQueueInfo: []*port.QueueStats{
					{
						QueueNumber: proto.Uint32(1),
						Bytes:       proto.Uint64(100000),
					},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"QueueNumber":  "1",
			"IfName":       "xe-0/0/0",
			"ParentAeName": "ae0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestPortStrAsTagFalse(t *testing.T) {
	parser := JTINativeParser{}
	parser.JTIStrAsTag = false

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName:       proto.String("xe-0/0/0"),
				ParentAeName: proto.String("ae0"),
				EgressQueueInfo: []*port.QueueStats{
					{
						QueueNumber: proto.Uint32(1),
						Bytes:       proto.Uint64(100000),
					},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}
	firstTestMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"QueueNumber":  "1",
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, firstTestMetric, parsedMetric[0])

	secondTestMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.ParentAeName": "ae0"},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, secondTestMetric, parsedMetric[1])
}

func TestFabricEnum(t *testing.T) {
	parser := JTINativeParser{}

	pbFabric := &fabric.FabricMessage{
		Location: fabric.FabricMessage_Switch_Fabric.Enum(),
		Edges: []*fabric.EdgeStats{
			{
				SourceType:      fabric.EdgeStats_Linecard.Enum(),
				DestinationType: fabric.EdgeStats_Switch_Fabric.Enum(),
				ClassStats: []*fabric.ClassStats{
					{
						Priority: proto.String("1"),
						TransmitCounts: &fabric.Counters{
							Bytes: proto.Uint64(1000),
						},
					},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbFabric, fabric.E_FabricMessageExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"Location":        "Switch_Fabric",
			"DestinationType": "Switch_Fabric",
			"SourceType":      "Linecard",
			"Priority":        "1",
			"component":       "65535",
			"device":          "test",
			"sensor":          "test sensor",
			"subcomponent":    "1",
		},
		map[string]interface{}{"Edges.ClassStats.TransmitCounts.Bytes": uint64(1000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestModifyMeasurementName(t *testing.T) {
	parser := JTINativeParser{
		JTINativeMeasurementOverride: []map[string]string{
			{"*": "new_measurement"},
		},
	}
	parser.BuildOverrides()

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{IfName: proto.String("xe-0/0/0"), IfTransitions: proto.Uint64(3)},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"new_measurement",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.IfTransitions": uint64(3)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestModifyTagName(t *testing.T) {
	parser := JTINativeParser{
		JTINativeTagOverride: []map[string]string{
			{"*.IfName": "interface"},
		},
	}
	parser.BuildOverrides()

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{IfName: proto.String("xe-0/0/0"), IfTransitions: proto.Uint64(3)},
		},
	}

	newTelTop := telStream
	newTelTop.SystemId = proto.String("new system id")
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"interface":    "xe-0/0/0",
			"component":    "65535",
			"device":       "new system id",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.IfTransitions": uint64(3)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestModifyDefualtTags(t *testing.T) {
	parser := JTINativeParser{
		DefaultTags: map[string]string{
			"dtag1": "foo",
			"dtag2": "bar",
		},
	}
	parser.BuildOverrides()

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{IfName: proto.String("xe-0/0/0"), IfTransitions: proto.Uint64(3)},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"IfName":       "xe-0/0/0",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
			"dtag1":        "foo",
			"dtag2":        "bar",
		},
		map[string]interface{}{"InterfaceStats.IfTransitions": uint64(3)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestConvertFields(t *testing.T) {
	parser := JTINativeParser{
		JTINativeConvertField: []string{
			"InterfaceStats.IfName",
		},
	}
	parser.BuildOverrides()

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{IfName: proto.String("xe-0/0/0")},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.IfName": "xe-0/0/0"},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}

func TestConvertTags(t *testing.T) {
	parser := JTINativeParser{
		JTINativeConvertTag: []string{
			"InterfaceStats.ParentAeName",
			"InterfaceStats.SnmpIfIndex",
		},
	}

	pbPort := &port.Port{
		InterfaceStats: []*port.InterfaceInfos{
			{
				IfName:       proto.String("xe-0/0/0"),
				ParentAeName: proto.String("ae0"),
				SnmpIfIndex:  proto.Uint32(42),
				EgressQueueInfo: []*port.QueueStats{
					{
						QueueNumber: proto.Uint32(1),
						Bytes:       proto.Uint64(100000),
					},
				},
			},
		},
	}

	newTelTop := telStream
	data, err := buildJtiProto(&newTelTop, pbPort, port.E_JnprInterfaceExt)
	if err != nil {
		log.Fatal("Build error: ", err)
	}

	parsedMetric, err := parser.Parse(data)
	if err != nil {
		log.Fatal("parser error: ", err)
	}

	testMetric := testutil.MustMetric(
		"test sensor",
		map[string]string{
			"QueueNumber":  "1",
			"IfName":       "xe-0/0/0",
			"ParentAeName": "ae0",
			"SnmpIfIndex":  "42",
			"component":    "65535",
			"device":       "test",
			"sensor":       "test sensor",
			"subcomponent": "1",
		},
		map[string]interface{}{"InterfaceStats.EgressQueueInfo.Bytes": uint64(100000)},
		time.Unix(0, 0),
	)
	testutil.RequireMetricEqual(t, testMetric, parsedMetric[0])
}
