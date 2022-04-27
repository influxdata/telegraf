package sflow_a10

import (
	"testing"

	tu "github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestXMLFileNegative(t *testing.T) {
	sflow := SFlow_A10{
		Log: tu.Logger{},
	}
	_, err := sflow.readA10XMLData([]byte(testXMLStringNegativeTag))
	require.Error(t, err)
	_, err = sflow.readA10XMLData([]byte(testXMLStringNoCounters))
	require.Error(t, err)
	_, err = sflow.readA10XMLData([]byte(testXMLStringWrongOffset))
	require.Error(t, err)
	_, err = sflow.readA10XMLData([]byte(testXMLStringWrongDtype))
	require.Error(t, err)
	_, err = sflow.readA10XMLData([]byte(testXMLStringEmptyEnumNameAndFieldName))
	require.Error(t, err)
}

const testXMLStringNegativeTag = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>-1</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_IP_PORT_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>20</ctr:ctrBlkSz>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PROTOCOL</ctr:enumName>
			<ctr:fieldName>Protocol</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_STATE</ctr:enumName>
			<ctr:fieldName>State</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>2</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_EXCEED_BYTE</ctr:enumName>
			<ctr:fieldName>Exceed</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>
`

const testXMLStringNoCounters = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>5</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_IP_PORT_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>20</ctr:ctrBlkSz>
	</ctr:counterBlock>
</ctr:allctrblocks>
`

const testXMLStringWrongOffset = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>5</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_IP_PORT_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>20</ctr:ctrBlkSz>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PROTOCOL</ctr:enumName>
			<ctr:fieldName>Protocol</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_STATE</ctr:enumName>
			<ctr:fieldName>State</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>3</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_EXCEED_BYTE</ctr:enumName>
			<ctr:fieldName>Exceed</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>
`

const testXMLStringWrongDtype = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>5</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_IP_PORT_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>20</ctr:ctrBlkSz>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PROTOCOL</ctr:enumName>
			<ctr:fieldName>Protocol</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_STATE</ctr:enumName>
			<ctr:fieldName>State</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>2</ctr:offset>
			<ctr:dtype>u33</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_EXCEED_BYTE</ctr:enumName>
			<ctr:fieldName>Exceed</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>
`

const testXMLStringEmptyEnumNameAndFieldName = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>5</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_IP_PORT_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>20</ctr:ctrBlkSz>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PROTOCOL</ctr:enumName>
			<ctr:fieldName>Protocol</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_STATE</ctr:enumName>
			<ctr:fieldName>State</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>2</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>
`
