package sflow_a10

import (
	"bytes"
	"net"
	"testing"

	tu "github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDecodeCounterSample(t *testing.T) {
	dc := NewDecoder()
	dc.CounterBlocks[217] = CounterBlock{
		OffsetHeaders: []HeaderDefinition{
			{},
			{},
			{},
			{},
		},
		Counters: []CounterDefinition{
			{
				Offset:    0,
				Dtype:     "u64",
				EnumName:  "testEnumName",
				FieldName: "testCounter0",
			},
			{
				Offset:    1,
				Dtype:     "u32",
				EnumName:  "testEnumName",
				FieldName: "ignored",
			},
			{
				Offset:    2,
				Dtype:     "u64",
				EnumName:  "testEnumName",
				FieldName: "testCounter1",
			},
			{
				Offset:    3,
				Dtype:     "u64",
				EnumName:  "testEnumName",
				FieldName: "ignored",
			},
			{
				Offset:    4,
				Dtype:     "u32",
				EnumName:  "testEnumName",
				FieldName: "ignored",
			},
			{
				Offset:    5,
				Dtype:     "u64",
				EnumName:  "testEnumName",
				FieldName: "testCounter2",
			},
		},
	}

	octets := bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x02, // sampleType uint32 (counter type is 2)
		0x00, 0x00, 0x03, 0x60, // sample data length

		0x00, 0x00, 0x00, 0x05, // sequenceNumber uint32
		0x00, 0x04, 0x41, 0x18, // sourceID type

		0x00, 0x00, 0x00, 0x01, // counter record count

		0x00, 0x00, 0x00, 0xD9, // counter format - D9 is 217
		0x00, 0x00, 0x00, 0x1, // counter data length

		// headers
		0x00, 0x00, // counterOffset uint16
		0x00, 0x03, // counterNum uint16

		0x00, 0x00, // reserved uint16
		0x00, 0x00, // reserved

		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x11, // counter metric 0 uint64
		0x00, 0x00, 0x00, 0x00, // to skip uint32
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xDC, // counter metric 1 uint64
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, // to skip uint32
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1D, // counter metric 2 uint64
	})

	expected := &Sample{
		SampleType: SampleType(2),
		SampleCounterData: &CounterSample{
			SequenceNumber: uint32(5),
			SourceID:       uint32(278808),
			CounterRecords: []CounterRecord{
				CounterRecord{
					CounterFormat: CounterFormatType(217),
					CounterData: &CounterData{
						CounterFields: map[string]interface{}{
							"testCounter0": uint64(17),
							"testCounter1": uint64(220),
							"testCounter2": uint64(29),
						},
					},
				},
			},
		},
	}

	actual, err := dc.decodeSample(octets, "10.0.1.2")
	require.NoError(t, err)
	require.Equal(t, expected, actual)

}

func TestCounterSampleSimple(t *testing.T) {
	octets := bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x0F, // SequenceNumber uint32
		0x00, 0x00, 0x00, 0x0A, // SourceIDType uint32
		0x00, 0x00, 0x00, 0x00, // CounterRecord count uint32
	})

	dc := NewDecoder()
	actual, err := dc.decodeCounterSample(octets, "10.1.2.3")
	require.NoError(t, err)

	expected := &CounterSample{
		SequenceNumber: 15,
		SourceID:       10,
	}

	require.Equal(t, expected, actual)
}

func TestDecode260(t *testing.T) {
	octets := bytes.NewBuffer([]byte{
		0x01,       // table type uint8
		0x02,       // port type uint 8
		0x3C, 0x76, // port num uint16
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // entry name [64]byte,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // still entry name
		0x3E, 0xFB, // port range end uint16
	})

	dc := NewDecoder()
	expected := &PortDimension{
		TableType:    "DST",
		PortType:     "TCP",
		PortNumber:   15478,
		PortRangeEnd: 16123,
	}
	actual, err := dc.decode260(octets)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestDecode271(t *testing.T) {
	octets := bytes.NewBuffer([]byte{
		0x00,       // startAddressOffset uint8
		0x02,       // address Count uint 8
		0x00, 0x04, // total Address Count uint16
		0x00, 0x00, 0x00, 0x00, // reserved uint32
		0x0A, 0x00, 0x01, 0x03, // ip 0 [4]byte
		0x1E,                   // subnet 0
		0xC0, 0xA8, 0x05, 0x06, // ip 1 [4]byte
		0x0F, // subnet 1
	})

	dc := NewDecoder()
	expected := []IPDimension{
		IPDimension{
			IPAddress:  "10.0.1.3",
			SubnetMask: 30,
		},
		IPDimension{
			IPAddress:  "192.168.5.6",
			SubnetMask: 15,
		},
	}
	actual, err := dc.decode271(octets)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

// TestDecodeEndToEnd tests the entire pipeline
// read the xml file
// read a 260 with sourceID X
// read a 271 with sourceID X
// read some counters with sourceID X
// make sure metrics from sourceID X are emitted
func TestDecodeA10EndToEnd(t *testing.T) {
	sflow := SFlow_A10{
		Log: tu.Logger{},
	}

	const sourceID = 269839
	agent_ip := "10.0.9.1"
	key := createMapKey(sourceID, agent_ip)

	// start by reading the XML file with the metric definitions
	c, err := sflow.readA10XMLData([]byte(testXMLStringEndToEnd))
	require.NoError(t, err)
	expected := map[uint32]CounterBlock{
		217: CounterBlock{
			Tag: 217,
			OffsetHeaders: []HeaderDefinition{
				HeaderDefinition{
					FieldName: "Counter Offset",
				},
				HeaderDefinition{
					FieldName: "Total Counter Num",
				},
				HeaderDefinition{
					FieldName: "Reserved1",
				},
				HeaderDefinition{
					FieldName: "Reserved2",
				},
			},
			Counters: []CounterDefinition{
				{
					Offset:    0,
					Dtype:     "u64",
					EnumName:  "IGNORED",
					FieldName: "ignored",
				},
				{
					Offset:    1,
					Dtype:     "u64",
					EnumName:  "IGNORED",
					FieldName: "ignored",
				},
				{
					Offset:    2,
					Dtype:     "u64",
					EnumName:  "IGNORED",
					FieldName: "ignored",
				},
				{
					Offset:    3,
					Dtype:     "u64",
					EnumName:  "TEST_COUNTER_0",
					FieldName: "test_counter_0",
				},
				{
					Offset:    4,
					Dtype:     "u64",
					EnumName:  "IGNORED",
					FieldName: "ignored",
				},
				{
					Offset:    5,
					Dtype:     "u64",
					EnumName:  "IGNORED",
					FieldName: "ignored",
				},
				{
					Offset:    6,
					Dtype:     "u64",
					EnumName:  "TEST_COUNTER_1",
					FieldName: "test_counter_1",
				},
			},
		},
	}
	require.Equal(t, expected[217].Tag, c[217].Tag)
	for i := 0; i < 4; i++ {
		require.Equal(t, expected[217].OffsetHeaders[i].FieldName, c[217].OffsetHeaders[i].FieldName)
	}
	for i := 0; i < 2; i++ {
		require.Equal(t, expected[217].Counters[i].Offset, c[217].Counters[i].Offset)
		require.Equal(t, expected[217].Counters[i].EnumName, c[217].Counters[i].EnumName)
		require.Equal(t, expected[217].Counters[i].FieldName, c[217].Counters[i].FieldName)
	}

	dc := NewDecoder()
	dc.CounterBlocks = c

	// we've read the XML successfully, so now we'll proceed in reading a 260 sample (contains port information)
	octets := bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x02, // sampleType uint32 (counter type is 2)
		0x00, 0x00, 0x03, 0x60, // sample data length

		0x00, 0x00, 0x00, 0x05, // sequenceNumber uint32
		0x00, 0x04, 0x1e, 0x0f, // sourceID type : 269839

		0x00, 0x00, 0x00, 0x01, // counter record count

		0x00, 0x00, 0x01, 0x04, // counter format - 0x104 is 260
		0x00, 0x00, 0x02, 0x30, // counter data length

		// 260 data follows

		0x01,       // table type uint8
		0x02,       // port type uint 8
		0x3C, 0x76, // port num uint16
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // entry name [64]byte,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // still entry name
		0x3E, 0xFB, // port range end uint16
	})

	_, err = dc.decodeSample(octets, agent_ip)
	require.NoError(t, err)

	portValue, portExists := dc.PortMap.Get(key)
	require.True(t, portExists)
	portDimensions := portValue.(*PortDimension)

	// make sure port information has been added into the map for the correct sourceID
	require.Equal(t, "DST", portDimensions.TableType)
	require.Equal(t, "TCP", portDimensions.PortType)
	require.Equal(t, 15478, portDimensions.PortNumber)
	require.Equal(t, 16123, portDimensions.PortRangeEnd)
	require.Equal(t, 16123, portDimensions.PortRangeEnd)

	// let's proceed in reading a 271 sample (contains IP information)
	octets = bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x02, // sampleType uint32 (counter type is 2)
		0x00, 0x00, 0x03, 0x60, // sample data length

		0x00, 0x00, 0x00, 0x05, // sequenceNumber uint32
		0x00, 0x04, 0x1e, 0x0f, // sourceID type : 269839

		0x00, 0x00, 0x00, 0x01, // counter record count

		0x00, 0x00, 0x01, 0x0F, // counter format - 0x10F is 271
		0x00, 0x00, 0x00, 0x90, // counter data length

		// 271 data follows
		0x00,       // startAddressOffset uint8
		0x02,       // address Count uint 8
		0x00, 0x04, // total Address Count uint16
		0x00, 0x00, 0x00, 0x00, // reserved uint32
		0x0A, 0x00, 0x01, 0x03, // ip 0 [4]byte
		0x1E,                   // subnet 0
		0xC0, 0xA8, 0x05, 0x06, // ip 1 [4]byte
		0x0F, // subnet 1
	})

	_, err = dc.decodeSample(octets, agent_ip)
	require.NoError(t, err)

	expectedIPAddresses := []IPDimension{
		IPDimension{
			IPAddress:  "10.0.1.3",
			SubnetMask: 30,
		},
		IPDimension{
			IPAddress:  "192.168.5.6",
			SubnetMask: 15,
		},
	}

	ipValue, ipExists := dc.IPMap.Get(key)
	portValue, portExists = dc.PortMap.Get(key)
	require.True(t, ipExists)
	require.True(t, portExists)

	ipDimensions := ipValue.([]IPDimension)
	portDimensions = portValue.(*PortDimension)

	require.Equal(t, 2, len(ipDimensions))
	for i := 0; i < 2; i++ {
		require.Equal(t, expectedIPAddresses[0], ipDimensions[0])
	}
	// also make sure port information is still there
	require.Equal(t, "DST", portDimensions.TableType)
	require.Equal(t, "TCP", portDimensions.PortType)
	require.Equal(t, 15478, portDimensions.PortNumber)
	require.Equal(t, 16123, portDimensions.PortRangeEnd)
	require.Equal(t, 16123, portDimensions.PortRangeEnd)

	// now let's read one 217 which contains the actual metrics
	octets = bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x02, // sampleType uint32 (counter type is 2)
		0x00, 0x00, 0x03, 0x60, // sample data length

		0x00, 0x00, 0x00, 0x05, // sequenceNumber uint32
		0x00, 0x04, 0x1e, 0x0f, // sourceID type : 269839

		0x00, 0x00, 0x00, 0x02, // counter record count

		// counter record 0
		0x00, 0x00, 0x00, 0xD9, // counter format - D9 is 217
		0x00, 0x00, 0x00, 0x1, // counter data length

		0x00, 0x00, // counterOffset uint16
		0x00, 0x02, // counterNum uint16

		0x00, 0x00, // reserved uint16
		0x00, 0x00, // reserved

		// counter data
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xDC, // counter metric 0 uint64
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1D, // counter metric 1 uint64

		// counter record 01
		0x00, 0x00, 0x00, 0xD9, // counter format - D9 is 217
		0x00, 0x00, 0x00, 0x1, // counter data length

		0x00, 0x00, // counterOffset uint16
		0x00, 0x02, // counterNum uint16

		0x00, 0x00, // reserved uint16
		0x00, 0x00, // reserved

		// counter data
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE3, // counter metric 0 uint64
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // to skip
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xA1, // counter metric 1 uint64
	})

	s, err := dc.decodeSample(octets, agent_ip)
	require.NoError(t, err)

	require.Equal(t, 2, len(s.SampleCounterData.CounterRecords))
	require.Equal(t, CounterFormatType(217), s.SampleCounterData.CounterRecords[0].CounterFormat)
	require.Equal(t, CounterFormatType(217), s.SampleCounterData.CounterRecords[1].CounterFormat)

	metricsSet0 := map[string]interface{}{
		"test_counter_0": uint64(220),
		"test_counter_1": uint64(29),
	}
	metricsSet1 := map[string]interface{}{
		"test_counter_0": uint64(227),
		"test_counter_1": uint64(161),
	}
	require.Equal(t, metricsSet0, s.SampleCounterData.CounterRecords[0].CounterData.CounterFields)
	require.Equal(t, metricsSet1, s.SampleCounterData.CounterRecords[1].CounterData.CounterFields)

	p := &V5Format{
		AgentAddress: net.IPAddr{
			IP: make([]byte, 4),
		},
		Samples: []Sample{
			*s,
		},
	}
	p.AgentAddress.IP = []byte{10, 0, 9, 1}
	// now let's try to get the actual metrics
	metrics, err := makeMetricsForCounters(p, dc)
	require.NoError(t, err)
	require.Equal(t, 2, len(metrics))
}

func TestIPv6ByteArrayToString(t *testing.T) {
	testIPv6 := [16]byte{
		0x1F, 0x1F, 0x1F, 0x1F, 0x1F, 0x1F, 0x3F, 0x1F, 0xAA, 0x1F, 0x1F, 0x1F, 0x1F, 0x1F, 0x1F, 0x5F,
	}
	s := fullIPv6(testIPv6)
	var expected = "1f1f:1f1f:1f1f:3f1f:aa1f:1f1f:1f1f:1f5f"
	require.Equal(t, s, expected)
}

const testXMLStringEndToEnd = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:tag>217</ctr:tag>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Counter Offset</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Total Counter Num</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Reserved1</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Reserved2</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>IGNORED</ctr:enumName>
			<ctr:fieldName>ignored</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>IGNORED</ctr:enumName>
			<ctr:fieldName>ignored</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>2</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>IGNORED</ctr:enumName>
			<ctr:fieldName>ignored</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>3</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_0</ctr:enumName>
			<ctr:fieldName>Test Counter 0</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>4</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>IGNORED</ctr:enumName>
			<ctr:fieldName>ignored</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>5</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>IGNORED</ctr:enumName>
			<ctr:fieldName>ignored</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>6</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_1</ctr:enumName>
			<ctr:fieldName>Test Counter 1</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>`
