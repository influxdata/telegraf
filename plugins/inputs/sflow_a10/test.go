package sflow_a10

import (
	"testing"

	tu "github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// TestXMLUnMarshalSimple tests for a simple unmarshaling
func TestXMLUnMarshalSimple(t *testing.T) {
	sflow := SFlow_A10{
		Log: tu.Logger{},
	}

	c, err := sflow.readA10XMLData([]byte(testXMLStringSimple))
	require.NoError(t, err)
	require.Equal(t, 1, len(c))

	require.Equal(t, 4, len(c[217].OffsetHeaders))
	require.Equal(t, 2, len(c[217].Counters))

	require.Equal(t, "Counter Offset", c[217].OffsetHeaders[0].FieldName)
	require.Equal(t, "Total Counter Num", c[217].OffsetHeaders[1].FieldName)

	require.Equal(t, "test_counter_0", c[217].Counters[0].FieldName)
	require.Equal(t, "test_counter_1", c[217].Counters[1].FieldName)

	require.Equal(t, 0, c[217].Counters[0].Offset)
	require.Equal(t, 1, c[217].Counters[1].Offset)

}

// TestXMLUnMarshalSameTagReturnsError makes sure that if we have the same tag for two different counter blocks we will get an error
func TestXMLUnMarshalSameTagReturnsError(t *testing.T) {
	sflow := SFlow_A10{
		Log: tu.Logger{},
	}

	_, err := sflow.readA10XMLData([]byte(testXMLStringSameTag))
	require.Error(t, err)
}

// TestXMLUnMarshalWrongOrderReturnedInCorrectOrder checks that an XML with wrong order in the offset headers is returned in the correct order
func TestXMLUnMarshalWrongOrderReturnedInCorrectOrder(t *testing.T) {
	sflow := SFlow_A10{
		Log: tu.Logger{},
	}

	c, err := sflow.readA10XMLData([]byte(testXMLStringOffsetWrongOrder))
	require.NoError(t, err)
	require.Equal(t, 1, len(c))
	require.Equal(t, 2, len(c[217].Counters))
	// check that counters have been ordered by offset in ascending order
	require.Equal(t, 7, c[217].Counters[0].Offset)
	require.Equal(t, 15, c[217].Counters[1].Offset)
}

const testXMLStringSimple = `
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
			<ctr:enumName>TEST_COUNTER_0</ctr:enumName>
			<ctr:fieldName>Test Counter 0</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_1</ctr:enumName>
			<ctr:fieldName>Test Counter 1</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>`

const testXMLStringSameTag = `
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
			<ctr:enumName>TEST_COUNTER_0</ctr:enumName>
			<ctr:fieldName>Test Counter 0</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_1</ctr:enumName>
			<ctr:fieldName>Test Counter 1</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
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
			<ctr:enumName>TEST_COUNTER_3</ctr:enumName>
			<ctr:fieldName>Test Counter 3</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_4</ctr:enumName>
			<ctr:fieldName>Test Counter 4</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>`

const testXMLStringOffsetWrongOrder = `
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
			<ctr:offset>15</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_0</ctr:enumName>
			<ctr:fieldName>Test Counter 0</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>7</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>TEST_COUNTER_1</ctr:enumName>
			<ctr:fieldName>Test Counter 1</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>`
