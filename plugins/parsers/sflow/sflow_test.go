package sflow

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"
)

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func min(x, y int) int {
	if y < x {
		return y
	}
	return x
}

func decodeAndCompare(expectedJSON []byte, packet []byte, t *testing.T) {
	var expected map[string]interface{}
	err := json.Unmarshal(expectedJSON, &expected)
	if err != nil {
		fmt.Println(err)
		t.Error("unable to unmarshal the expected JSON", err)
	}

	packetBytes := make([]byte, hex.DecodedLen(len(packet)))
	_, err = hex.Decode(packetBytes, packet)
	options := NewDefaultV5FormatOptions()
	decoded, err := Decode(V5Format(options), bytes.NewBuffer(packetBytes))
	if err != nil {
		t.Error("unable to decode the packet", err)
	}
	packetAsJSON, err := json.Marshal(decoded)
	if err != nil {
		t.Error("unable to marshal packet object back to JSON", err)
	}
	expectedAsJSON, err := json.Marshal(expected)
	if err != nil {
		t.Error("unable to marshal expected object back to JSON", err)
	}

	if bytes.Compare(packetAsJSON, expectedAsJSON) != 0 {
		var differenceIndex int
		var packetAsJSONSnippet []byte
		var expectedAsJSONSnippet []byte
		for i, b := range packetAsJSON {
			if i >= len(expectedAsJSON) {
				differenceIndex = i
				break
			}
			if expectedAsJSON[i] != b {
				differenceIndex = i
				packetAsJSONSnippet = packetAsJSON[max(differenceIndex-10, 0):min(differenceIndex+10, len(packetAsJSON)-1)]
				expectedAsJSONSnippet = expectedAsJSON[max(differenceIndex-10, 0):min(differenceIndex+10, len(expectedAsJSON)-1)]
				break
			}
		}
		t.Errorf("Actual and expected are not equal at %d, act %s exp %s", differenceIndex, packetAsJSONSnippet, expectedAsJSONSnippet)
	}
}

func Test_sflow_flow_ipv4_sw(t *testing.T) {
	packet := []byte("0000000500000001c0a80102000000100000f3d40bfa047f0000000200000001000000d00001210a000001fe000004000484240000000000000001fe00000200000000020000000100000090000000010000010b0000000400000080000c2936d3d694c691aa97600800450000f9f19040004011b4f5c0a80913c0a8090a00a1ba0500e5641f3081da02010104066d6f746f6770a281cc02047b46462e0201000201003081bd3012060d2b06010201190501010281dc710201003013060d2b06010201190501010281e66802025acc3012060d2b0601020119050101000003e9000000100000000900000000000000090000000000000001000000d00000e3cc000002100000400048eb740000000000000002100000020000000002000000010000009000000001000000970000000400000080000c2936d3d6fcecda44008f81000009080045000081186440003f119098c0a80815c0a8090a9a690202006d23083c33303e4170722031312030393a33333a3031206b6e6f64653120736e6d70645b313039385d3a20436f6e6e656374696f6e2066726f6d205544503a205b3139322e3136382e392e31305d3a34393233362d000003e90000001000000009000000000000000900000000")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "wKgBAg==",
		"samples": [
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 267,
					"header": [
					   {
						  "IPTTL": 64,
						  "IPversion": 4,
						  "dscp": 0,
						  "dstIP": "192.168.9.10",
						  "dstMac": 52231066582,
						  "dstPort": 47621,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 61840,
						  "fragmentOffset": 0,
						  "proto": 17,
						  "srcIP": "192.168.9.19",
						  "srcMac": 163580568311648,
						  "srcPort": 161,
						  "tagOrEType": 2048,
						  "total_length": 249,
						  "udp_length": 229
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 9,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 9
				 }
			  ],
			  "flowRecords.length": 2,
			  "inputFormat": 0,
			  "inputValue": 510,
			  "outputFormat": 0,
			  "outputValue": 512,
			  "sampleData.length": 208,
			  "samplePool": 75768832,
			  "sampleType": 1,
			  "samplingRate": 1024,
			  "sequenceNumber": 73994,
			  "sourceIdType": 0,
			  "sourceIdValue": 510
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 151,
					"header": [
					   {
						  "IPTTL": 63,
						  "IPversion": 4,
						  "dscp": 0,
						  "dstIP": "192.168.9.10",
						  "dstMac": 52231066582,
						  "dstPort": 514,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 6244,
						  "fragmentOffset": 0,
						  "proto": 17,
						  "srcIP": "192.168.8.21",
						  "srcMac": 278094204371087,
						  "srcPort": 39529,
						  "tagOrEType": 33024,
						  "total_length": 129,
						  "udp_length": 109,
						  "vlanID": 9
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 9,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 9
				 }
			  ],
			  "flowRecords.length": 2,
			  "inputFormat": 0,
			  "inputValue": 528,
			  "outputFormat": 0,
			  "outputValue": 512,
			  "sampleData.length": 208,
			  "samplePool": 1223390208,
			  "sampleType": 1,
			  "samplingRate": 16384,
			  "sequenceNumber": 58316,
			  "sourceIdType": 0,
			  "sourceIdValue": 528
		   }
		],
		"samples.length": 2,
		"sequenceNumber": 62420,
		"subAgentId": 16,
		"uptime": 200934527,
		"version": 5
	 }
	`)
	decodeAndCompare(expectedJSON, packet, t)
}

func Test_sflow_expand_flow(t *testing.T) {
	packet := []byte("00000005000000010a00015000000000000f58998ae119780000000300000003000000c4000b62a90000000000100c840000040024fb7e1e0000000000000000001017840000000000100c8400000001000000010000009000000001000005bc0000000400000080001b17000130001201f58d44810023710800450205a6305440007e06ee92ac100016d94d52f505997e701fa1e17aff62574a50100200355f000000ffff00000b004175746f72697a7a6174610400008040ffff000400008040050031303030320500313030302004000000000868a200000000000000000860a200000000000000000003000000c40003cecf000000000010170400004000a168ac1c000000000000000000101784000000000010170400000001000000010000009000000001000005f200000004000000800024e8324338d4ae52aa0b54810020060800450005dc5420400080061397c0a8060cc0a806080050efcfbb25bad9a21c839a501000fff54000008a55f70975a0ff88b05735597ae274bd81fcba17e6e9206b8ea0fb07d05fc27dad06cfe3fdba5d2fc4d057b0add711e596cbe5e9b4bbe8be59cd77537b7a89f7414a628b736d00000003000000c0000c547a0000000000100c04000004005bc3c3b50000000000000000001017840000000000100c0400000001000000010000008c000000010000007e000000040000007a001b17000130001201f58d448100237108004500006824ea4000ff32c326d94d5105501018f02e88d003000001dd39b1d025d1c68689583b2ab21522d5b5a959642243804f6d51e63323091cc04544285433eb3f6b29e1046a6a2fa7806319d62041d8fa4bd25b7cd85b8db54202054a077ac11de84acbe37a550004")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "CgABUA==",
		"samples": [
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 1468,
					"header": [
					   {
						  "IPTTL": 126,
						  "IPversion": 4,
						  "ack_number": 4284634954,
						  "checksum": 13663,
						  "dscp": 0,
						  "dstIP": "217.77.82.245",
						  "dstMac": 116349993264,
						  "dstPort": 32368,
						  "ecn": 2,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 12372,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 530702714,
						  "srcIP": "172.16.0.22",
						  "srcMac": 77342281028,
						  "srcPort": 1433,
						  "tagOrEType": 33024,
						  "tcp_header_length": 64,
						  "tcp_window_size": 512,
						  "total_length": 1446,
						  "urgent_pointer": 0,
						  "vlanID": 881
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1054596,
			  "outputFormat": 0,
			  "outputValue": 1051780,
			  "sampleData.length": 196,
			  "samplePool": 620461598,
			  "sampleType": 3,
			  "samplingRate": 1024,
			  "sequenceNumber": 746153,
			  "sourceIdType": 0,
			  "sourceIdValue": 1051780
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 1522,
					"header": [
					   {
						  "IPTTL": 128,
						  "IPversion": 4,
						  "ack_number": 2719777690,
						  "checksum": 62784,
						  "dscp": 0,
						  "dstIP": "192.168.6.8",
						  "dstMac": 158514430776,
						  "dstPort": 61391,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 21536,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 3139812057,
						  "srcIP": "192.168.6.12",
						  "srcMac": 233845176273748,
						  "srcPort": 80,
						  "tagOrEType": 33024,
						  "tcp_header_length": 64,
						  "tcp_window_size": 255,
						  "total_length": 1500,
						  "urgent_pointer": 0,
						  "vlanID": 6
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1054596,
			  "outputFormat": 0,
			  "outputValue": 1054468,
			  "sampleData.length": 196,
			  "samplePool": 2707991580,
			  "sampleType": 3,
			  "samplingRate": 16384,
			  "sequenceNumber": 249551,
			  "sourceIdType": 0,
			  "sourceIdValue": 1054468
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 140,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 126,
					"header": [
					   {
						  "IPTTL": 255,
						  "IPversion": 4,
						  "WARN": "unimplemented support for protol 50",
						  "dscp": 0,
						  "dstIP": "80.16.24.240",
						  "dstMac": 116349993264,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 9450,
						  "fragmentOffset": 0,
						  "proto": 50,
						  "srcIP": "217.77.81.5",
						  "srcMac": 77342281028,
						  "tagOrEType": 33024,
						  "total_length": 104,
						  "vlanID": 881
					   }
					],
					"header.length": 122,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1054596,
			  "outputFormat": 0,
			  "outputValue": 1051652,
			  "sampleData.length": 192,
			  "samplePool": 1539556277,
			  "sampleType": 3,
			  "samplingRate": 1024,
			  "sequenceNumber": 808058,
			  "sourceIdType": 0,
			  "sourceIdValue": 1051652
		   }
		],
		"samples.length": 3,
		"sequenceNumber": 1005721,
		"subAgentId": 0,
		"uptime": 2330007928,
		"version": 5
	 }	 
	`)
	decodeAndCompare(expectedJSON, packet, t)
}

func Test_sflow_flow_ipv4_sw_rt(t *testing.T) {
	packet := []byte("000000050000000189dd4f010000000000003d4f21151ad40000000600000001000000bc354b97090000020c000013b175792bea000000000000028f0000020c0000000300000001000000640000000100000058000000040000005408b2587a57624c16fc0b61a5080045000046c3e440003a1118a0052aada7569e5ab367a6e35b0032d7bbf1f2fb2eb2490a97f87abc31e135834be367000002590000ffffffffffffffff02add830d51e0aec14cf000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e32a000000160000000b00000001000000a88b8ffb57000002a2000013b12e344fd800000000000002a20000028f0000000300000001000000500000000100000042000000040000003e4c16fc0b6202c03e0fdecafe080045000030108000007d11fe45575185a718693996f0570e8c001c20614ad602003fd6d4afa6a6d18207324000271169b00000000003e90000001000000000000000000000000000000000000003ea000000100000000189dd4f210000000f0000001800000001000000e8354b970a0000020c000013b175793f9b000000000000028f0000020c00000003000000010000009000000001000001a500000004000000800231466d0b2c4c16fc0b61a5080045000193198f40003a114b75052aae1f5f94c778678ef24d017f50ea7622287c30799e1f7d45932d01ca92c46d930000927c0000ffffffffffffffff02ad0eea6498953d1c7ebb6dbdf0525c80e1a9a62bacfea92f69b7336c2f2f60eba0593509e14eef167eb37449f05ad70b8241c1a46d000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000160000001000000001000000e8354b970b0000020c000013b17579534c000000000000028f0000020c00000003000000010000009000000001000000b500000004000000800231466d0b2c4c16fc0b61a50800450000a327c240003606fd67b93c706a021ff365045fe8a0976d624df8207083501800edb31b0000485454502f312e3120323030204f4b0d0a5365727665723a2050726f746f636f6c20485454500d0a436f6e74656e742d4c656e6774683a20313430340d0a436f6e6e656374696f6e3a20000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000170000001000000001000000e8354b970c0000020c000013b1757966fd000000000000028f0000020c000000030000000100000090000000010000018e00000004000000800231466d0b2c4c16fc0b61a508004500017c7d2c40003a116963052abd8d021c940e67e7e0d501682342dbe7936bd47ef487dee5591ec1b24d83622e000072250000ffffffffffffffff02ad0039d8ba86a90017071d76b177de4d8c4e23bcaaaf4d795f77b032f959e0fb70234d4c28922d4e08dd3330c66e34bff51cc8ade5000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000160000001000000001000000e80d6146ac000002a1000013b17880b49d00000000000002a10000028f00000003000000010000009000000001000005ee00000004000000804c16fc0b6201d8b122766a2c0800450005dc04574000770623a11fcd80a218691d4cf2fe01bbd4f47482065fd63a5010fabd7987000052a20002c8c43ea91ca1eaa115663f5218a37fbb409dfbbedff54731ef41199b35535905ac2366a05a803146ced544abf45597f3714327d59f99e30c899c39fc5a4b67d12087bf8db2bc000003e90000001000000000000000000000000000000000000003ea000000100000000189dd4f210000001000000018")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "id1PAQ==",
		"samples": [
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 100,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 88,
					"header": [
					   {
						  "IPTTL": 58,
						  "IPversion": 4,
						  "dscp": 0,
						  "dstIP": "86.158.90.179",
						  "dstMac": 9562081613666,
						  "dstPort": 58203,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 50148,
						  "fragmentOffset": 0,
						  "proto": 17,
						  "srcIP": "5.42.173.167",
						  "srcMac": 83661601595813,
						  "srcPort": 26534,
						  "tagOrEType": 2048,
						  "total_length": 70,
						  "udp_length": 50
					   }
					],
					"header.length": 84,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 0,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 0
				 },
				 {
					"dstMaskLen": 11,
					"flowData.length": 16,
					"flowFormat": "extendedRouterFlowData",
					"nextHop.address": "w0LjKg==",
					"nextHop.addressType": "IPV4",
					"srcMaskLen": 22
				 }
			  ],
			  "flowRecords.length": 3,
			  "inputFormat": 0,
			  "inputValue": 655,
			  "outputFormat": 0,
			  "outputValue": 524,
			  "sampleData.length": 188,
			  "samplePool": 1970875370,
			  "sampleType": 1,
			  "samplingRate": 5041,
			  "sequenceNumber": 894146313,
			  "sourceIdType": 0,
			  "sourceIdValue": 524
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 80,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 66,
					"header": [
					   {
						  "IPTTL": 125,
						  "IPversion": 4,
						  "dscp": 0,
						  "dstIP": "24.105.57.150",
						  "dstMac": 83661601595906,
						  "dstPort": 3724,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 0,
						  "fragmentId": 4224,
						  "fragmentOffset": 0,
						  "proto": 17,
						  "srcIP": "87.81.133.167",
						  "srcMac": 211372786764542,
						  "srcPort": 61527,
						  "tagOrEType": 2048,
						  "total_length": 48,
						  "udp_length": 28
					   }
					],
					"header.length": 62,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 0,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 0
				 },
				 {
					"dstMaskLen": 24,
					"flowData.length": 16,
					"flowFormat": "extendedRouterFlowData",
					"nextHop.address": "id1PIQ==",
					"nextHop.addressType": "IPV4",
					"srcMaskLen": 15
				 }
			  ],
			  "flowRecords.length": 3,
			  "inputFormat": 0,
			  "inputValue": 674,
			  "outputFormat": 0,
			  "outputValue": 655,
			  "sampleData.length": 168,
			  "samplePool": 775180248,
			  "sampleType": 1,
			  "samplingRate": 5041,
			  "sequenceNumber": 2341469015,
			  "sourceIdType": 0,
			  "sourceIdValue": 674
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 421,
					"header": [
					   {
						  "IPTTL": 58,
						  "IPversion": 4,
						  "dscp": 0,
						  "dstIP": "95.148.199.120",
						  "dstMac": 2410658204460,
						  "dstPort": 62029,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 6543,
						  "fragmentOffset": 0,
						  "proto": 17,
						  "srcIP": "5.42.174.31",
						  "srcMac": 83661601595813,
						  "srcPort": 26510,
						  "tagOrEType": 2048,
						  "total_length": 403,
						  "udp_length": 383
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 0,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 0
				 },
				 {
					"dstMaskLen": 16,
					"flowData.length": 16,
					"flowFormat": "extendedRouterFlowData",
					"nextHop.address": "w0Lh/Q==",
					"nextHop.addressType": "IPV4",
					"srcMaskLen": 22
				 }
			  ],
			  "flowRecords.length": 3,
			  "inputFormat": 0,
			  "inputValue": 655,
			  "outputFormat": 0,
			  "outputValue": 524,
			  "sampleData.length": 232,
			  "samplePool": 1970880411,
			  "sampleType": 1,
			  "samplingRate": 5041,
			  "sequenceNumber": 894146314,
			  "sourceIdType": 0,
			  "sourceIdValue": 524
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 181,
					"header": [
					   {
						  "IPTTL": 54,
						  "IPversion": 4,
						  "ack_number": 4162875523,
						  "checksum": 45851,
						  "dscp": 0,
						  "dstIP": "2.31.243.101",
						  "dstMac": 2410658204460,
						  "dstPort": 59552,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 10178,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 2540528205,
						  "srcIP": "185.60.112.106",
						  "srcMac": 83661601595813,
						  "srcPort": 1119,
						  "tagOrEType": 2048,
						  "tcp_header_length": 64,
						  "tcp_window_size": 237,
						  "total_length": 163,
						  "urgent_pointer": 0
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 0,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 0
				 },
				 {
					"dstMaskLen": 16,
					"flowData.length": 16,
					"flowFormat": "extendedRouterFlowData",
					"nextHop.address": "w0Lh/Q==",
					"nextHop.addressType": "IPV4",
					"srcMaskLen": 23
				 }
			  ],
			  "flowRecords.length": 3,
			  "inputFormat": 0,
			  "inputValue": 655,
			  "outputFormat": 0,
			  "outputValue": 524,
			  "sampleData.length": 232,
			  "samplePool": 1970885452,
			  "sampleType": 1,
			  "samplingRate": 5041,
			  "sequenceNumber": 894146315,
			  "sourceIdType": 0,
			  "sourceIdValue": 524
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 398,
					"header": [
					   {
						  "IPTTL": 58,
						  "IPversion": 4,
						  "dscp": 0,
						  "dstIP": "2.28.148.14",
						  "dstMac": 2410658204460,
						  "dstPort": 57557,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 32044,
						  "fragmentOffset": 0,
						  "proto": 17,
						  "srcIP": "5.42.189.141",
						  "srcMac": 83661601595813,
						  "srcPort": 26599,
						  "tagOrEType": 2048,
						  "total_length": 380,
						  "udp_length": 360
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 0,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 0
				 },
				 {
					"dstMaskLen": 16,
					"flowData.length": 16,
					"flowFormat": "extendedRouterFlowData",
					"nextHop.address": "w0Lh/Q==",
					"nextHop.addressType": "IPV4",
					"srcMaskLen": 22
				 }
			  ],
			  "flowRecords.length": 3,
			  "inputFormat": 0,
			  "inputValue": 655,
			  "outputFormat": 0,
			  "outputValue": 524,
			  "sampleData.length": 232,
			  "samplePool": 1970890493,
			  "sampleType": 1,
			  "samplingRate": 5041,
			  "sequenceNumber": 894146316,
			  "sourceIdType": 0,
			  "sourceIdValue": 524
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 1518,
					"header": [
					   {
						  "IPTTL": 119,
						  "IPversion": 4,
						  "ack_number": 106944058,
						  "checksum": 31111,
						  "dscp": 0,
						  "dstIP": "24.105.29.76",
						  "dstMac": 83661601595905,
						  "dstPort": 443,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 1111,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 3572790402,
						  "srcIP": "31.205.128.162",
						  "srcMac": 238255298996780,
						  "srcPort": 62206,
						  "tagOrEType": 2048,
						  "tcp_header_length": 64,
						  "tcp_window_size": 64189,
						  "total_length": 1500,
						  "urgent_pointer": 0
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 0,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 0
				 },
				 {
					"dstMaskLen": 24,
					"flowData.length": 16,
					"flowFormat": "extendedRouterFlowData",
					"nextHop.address": "id1PIQ==",
					"nextHop.addressType": "IPV4",
					"srcMaskLen": 16
				 }
			  ],
			  "flowRecords.length": 3,
			  "inputFormat": 0,
			  "inputValue": 673,
			  "outputFormat": 0,
			  "outputValue": 655,
			  "sampleData.length": 232,
			  "samplePool": 2021700765,
			  "sampleType": 1,
			  "samplingRate": 5041,
			  "sequenceNumber": 224478892,
			  "sourceIdType": 0,
			  "sourceIdValue": 673
		   }
		],
		"samples.length": 6,
		"sequenceNumber": 15695,
		"subAgentId": 0,
		"uptime": 555031252,
		"version": 5
	 }
	`)
	decodeAndCompare(expectedJSON, packet, t)
}

func Test_sflow_flow_ipv6_sw(t *testing.T) {
	packet := []byte("00000005000000010ae0648100000002000093d824ac82340000000100000001000000d000019f94000001010000100019f94000000000000000010100000000000000020000000100000090000000010000058c00000008000000800008e3fffc10d4f4be04612486dd60000000054e113a2607f8b0400200140000000000000008262000edc000e804a25e30c581af36fa01bbfa6f054e249810b584bcbf12926c2e29a779c26c72db483e8191524fe2288bfdaceaf9d2e724d04305706efcfdef70db86873bbacf29698affe4e7d6faa21d302f9b4b023291a05a000003e90000001000000001000000000000000100000000")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "CuBkgQ==",
		"samples": [
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 1420,
					"header": [
					   {
						  "IPv6FlowLabel": 0,
						  "IPversion": 0,
						  "dstIP": "2620:ed:c000:e804:a25e:30c5:81af:36fa",
						  "dstMac": 38184942608,
						  "etype": 34525,
						  "hopLimit": 58,
						  "nextHeader": 17,
						  "paylloadLength": 1358,
						  "srcIP": "2607:f8b0:4002:14::8",
						  "srcMac": 234147625066788,
						  "tagOrEType": 34525
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 8
				 },
				 {
					"dstPriority": 0,
					"dstVlan": 1,
					"flowData.length": 16,
					"flowFormat": "extendedSwitchFlowData",
					"srcPriority": 0,
					"srcVlan": 1
				 }
			  ],
			  "flowRecords.length": 2,
			  "inputFormat": 0,
			  "inputValue": 257,
			  "outputFormat": 0,
			  "outputValue": 0,
			  "sampleData.length": 208,
			  "samplePool": 435765248,
			  "sampleType": 1,
			  "samplingRate": 4096,
			  "sequenceNumber": 106388,
			  "sourceIdType": 0,
			  "sourceIdValue": 257
		   }
		],
		"samples.length": 1,
		"sequenceNumber": 37848,
		"subAgentId": 2,
		"uptime": 615285300,
		"version": 5
	 }	 
	`)
	decodeAndCompare(expectedJSON, packet, t)
}

func Test_sflow_expand_flow_counter(t *testing.T) {
	packet := []byte("00000005000000010a00015000000000000f58898ae0fa380000000700000004000000ec00006ece0000000000101784000000030000000200000034000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000058001017840000000600000002540be400000000010000000300007b8ebd37b97e61ff94860803e8e908ffb2b500000000000000000000000000018e7c31ee7ba4195f041874579ff021ba936300000000000000000000000100000007000000380011223344550003f8b15645e7e7d6960000002fe2fc02fc01edbf580000000000000000000000000000000001dcb9cf000000000000000000000004000000ec00006ece0000000000100184000000030000000200000034000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000058001001840000000600000002540be400000000010000000300000841131d1fd9f850bfb103617cb401e6598900000000000000000000000000000bec1902e5da9212e3e96d7996e922513250000000000000000000000001000000070000003800112233445500005c260acbddb3000100000003e2fc02fc01ee414f0000000000000000000000000000000001dccdd30000000000000000000000030000008400004606000000000010030400004000ad9dc19b0000000000000000001017840000000000100304000000010000000100000050000000010000004400000004000000400012815116c4001517cf426d8100200608004500002895da40008006d74bc0a8060ac0a8064f04ef04aab1797122cf7eaf4f5010ffff7727000000000000000000000003000000b0001bd698000000000010148400000400700b180f000000000000000000101504000000000010148400000001000000010000007c000000010000006f000000040000006b001b17000131f0f755b9afc081000439080045000059045340005206920c1f0d4703d94d52e201bbf14977d1e9f15498af36801800417f1100000101080afdf3c70400e043871503010020ff268cfe2e2fd5fffe1d3d704a91d57b895f174c4b4428c66679d80a307294303f00000003000000c40003ceca000000000010170400004000a166aa7a000000000000000000101784000000000010170400000001000000010000009000000001000005f200000004000000800024e8369e2bd4ae52aa0b54810020060800450005dc4c71400080061b45c0a8060cc0a806090050f855692a7a94a1154ae1801001046b6a00000101080a6869a48d151016d046a84a7aa1c6743fa05179f7ecbd4e567150cb6f2077ff89480ae730637d26d2237c08548806f672c7476eb1b5a447b42cb9ce405994d152fa3e000000030000008c001bd699000000000010148400000400700b180f0000000000000000001015040000000000101484000000010000000100000058000000010000004a0000000400000046001b17000131f0f755b9afc0810004390800450000340ce040003a06bea5c1ce8793d94d528f00504c3b08b18f275b83d5df8010054586ad00000101050a5b83d5de5b83d5df11d800000003000000c400004e07000000000010028400004000c7ec97f2000000000000000000100784000000000010028400000001000000010000009000000001000005f2000000040000008000005e0001ff005056800dd18100000a0800450005dc5a42400040066ef70a000ac8c0a8967201bbe17c81597908caf8a05f5010010328610000f172263da0ba5d6223c079b8238bc841256bf17c4ffb08ad11c4fbff6f87ae1624a6b057b8baa9342114e5f5b46179083020cb560c4e9eadcec6dfd83e102ddbc27024803eb5")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "CgABUA==",
		"samples": [
		   {
			  "counters": [
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 },
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 150975157,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 134473961,
					"ifInOctets": 135852990118270,
					"ifInUcastPkts": 1644139654,
					"ifInUnknownProtos": 0,
					"ifIndex": 1054596,
					"ifOutBroadcastPkts": 565875555,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 1951899632,
					"ifOutOctets": 438139041512356,
					"ifOutUcastPkts": 425657368,
					"ifPromiscuousMode": 1,
					"ifSpeed": 10000000000,
					"ifStatus": 3,
					"ifType": 6
				 },
				 {
					"WARN": "unhandled counterFormat 7",
					"counterData.length": 56,
					"counterFormat": 7
				 }
			  ],
			  "counters.length": 3,
			  "sampleData.length": 236,
			  "sampleType": 4,
			  "sequenceNumber": 28366,
			  "sourceIdType": 0,
			  "sourceIdValue": 1054596
		   },
		   {
			  "counters": [
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 },
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 31873417,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 56720564,
					"ifInOctets": 9075586572249,
					"ifInUcastPkts": 4166041521,
					"ifInUnknownProtos": 0,
					"ifIndex": 1048964,
					"ifOutBroadcastPkts": 575746640,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 1836685033,
					"ifOutOctets": 13108659807706,
					"ifOutUcastPkts": 2450711529,
					"ifPromiscuousMode": 1,
					"ifSpeed": 10000000000,
					"ifStatus": 3,
					"ifType": 6
				 },
				 {
					"WARN": "unhandled counterFormat 7",
					"counterData.length": 56,
					"counterFormat": 7
				 }
			  ],
			  "counters.length": 3,
			  "sampleData.length": 236,
			  "sampleType": 4,
			  "sequenceNumber": 28366,
			  "sourceIdType": 0,
			  "sourceIdValue": 1048964
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 80,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 68,
					"header": [
					   {
						  "IPTTL": 128,
						  "IPversion": 4,
						  "ack_number": 3481186127,
						  "checksum": 30503,
						  "dscp": 0,
						  "dstIP": "192.168.6.79",
						  "dstMac": 79478986436,
						  "dstPort": 1194,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 38362,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 2977526050,
						  "srcIP": "192.168.6.10",
						  "srcMac": 90593772141,
						  "srcPort": 1263,
						  "tagOrEType": 33024,
						  "tcp_header_length": 64,
						  "tcp_window_size": 65535,
						  "total_length": 40,
						  "urgent_pointer": 0,
						  "vlanID": 6
					   }
					],
					"header.length": 64,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1054596,
			  "outputFormat": 0,
			  "outputValue": 1049348,
			  "sampleData.length": 132,
			  "samplePool": 2912797083,
			  "sampleType": 3,
			  "samplingRate": 16384,
			  "sequenceNumber": 17926,
			  "sourceIdType": 0,
			  "sourceIdValue": 1049348
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 124,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 111,
					"header": [
					   {
						  "IPTTL": 82,
						  "IPversion": 4,
						  "ack_number": 1419292470,
						  "checksum": 32529,
						  "dscp": 0,
						  "dstIP": "217.77.82.226",
						  "dstMac": 116349993265,
						  "dstPort": 61769,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 1107,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 2010245617,
						  "srcIP": "31.13.71.3",
						  "srcMac": 264945085820864,
						  "srcPort": 443,
						  "tagOrEType": 33024,
						  "tcp_header_length": 0,
						  "tcp_window_size": 65,
						  "total_length": 89,
						  "urgent_pointer": 0,
						  "vlanID": 1081
					   }
					],
					"header.length": 107,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1053956,
			  "outputFormat": 0,
			  "outputValue": 1053828,
			  "sampleData.length": 176,
			  "samplePool": 1879775247,
			  "sampleType": 3,
			  "samplingRate": 1024,
			  "sequenceNumber": 1824408,
			  "sourceIdType": 0,
			  "sourceIdValue": 1053828
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 1522,
					"header": [
					   {
						  "IPTTL": 128,
						  "IPversion": 4,
						  "ack_number": 2702527201,
						  "checksum": 27498,
						  "dscp": 0,
						  "dstIP": "192.168.6.9",
						  "dstMac": 158514716203,
						  "dstPort": 63573,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 19569,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 1764391572,
						  "srcIP": "192.168.6.12",
						  "srcMac": 233845176273748,
						  "srcPort": 80,
						  "tagOrEType": 33024,
						  "tcp_header_length": 0,
						  "tcp_window_size": 260,
						  "total_length": 1500,
						  "urgent_pointer": 0,
						  "vlanID": 6
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1054596,
			  "outputFormat": 0,
			  "outputValue": 1054468,
			  "sampleData.length": 196,
			  "samplePool": 2707860090,
			  "sampleType": 3,
			  "samplingRate": 16384,
			  "sequenceNumber": 249546,
			  "sourceIdType": 0,
			  "sourceIdValue": 1054468
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 88,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 74,
					"header": [
					   {
						  "IPTTL": 58,
						  "IPversion": 4,
						  "ack_number": 1535366623,
						  "checksum": 34477,
						  "dscp": 0,
						  "dstIP": "217.77.82.143",
						  "dstMac": 116349993265,
						  "dstPort": 19515,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 3296,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 145854247,
						  "srcIP": "193.206.135.147",
						  "srcMac": 264945085820864,
						  "srcPort": 80,
						  "tagOrEType": 33024,
						  "tcp_header_length": 0,
						  "tcp_window_size": 1349,
						  "total_length": 52,
						  "urgent_pointer": 0,
						  "vlanID": 1081
					   }
					],
					"header.length": 70,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1053956,
			  "outputFormat": 0,
			  "outputValue": 1053828,
			  "sampleData.length": 140,
			  "samplePool": 1879775247,
			  "sampleType": 3,
			  "samplingRate": 1024,
			  "sequenceNumber": 1824409,
			  "sourceIdType": 0,
			  "sourceIdValue": 1053828
		   },
		   {
			  "drops": 0,
			  "flowRecords": [
				 {
					"flowData.length": 144,
					"flowFormat": "rawPacketHeaderFlowData",
					"frameLength": 1522,
					"header": [
					   {
						  "IPTTL": 64,
						  "IPversion": 4,
						  "ack_number": 3405291615,
						  "checksum": 10337,
						  "dscp": 0,
						  "dstIP": "192.168.150.114",
						  "dstMac": 1577058815,
						  "dstPort": 57724,
						  "ecn": 0,
						  "etype": 2048,
						  "flags": 2,
						  "fragmentId": 23106,
						  "fragmentOffset": 0,
						  "proto": 6,
						  "sequence": 2170124552,
						  "srcIP": "10.0.10.200",
						  "srcMac": 345048616401,
						  "srcPort": 443,
						  "tagOrEType": 33024,
						  "tcp_header_length": 64,
						  "tcp_window_size": 259,
						  "total_length": 1500,
						  "urgent_pointer": 0,
						  "vlanID": 10
					   }
					],
					"header.length": 128,
					"protocol": "ETHERNET-ISO88023",
					"stripped": 4
				 }
			  ],
			  "flowRecords.length": 1,
			  "inputFormat": 0,
			  "inputValue": 1050500,
			  "outputFormat": 0,
			  "outputValue": 1049220,
			  "sampleData.length": 196,
			  "samplePool": 3354171378,
			  "sampleType": 3,
			  "samplingRate": 16384,
			  "sequenceNumber": 19975,
			  "sourceIdType": 0,
			  "sourceIdValue": 1049220
		   }
		],
		"samples.length": 7,
		"sequenceNumber": 1005705,
		"subAgentId": 0,
		"uptime": 2329999928,
		"version": 5
	 } 
	`)
	decodeAndCompare(expectedJSON, packet, t)
}

func Test_sflow_expand_counter(t *testing.T) {
	packet := []byte("00000005000000010a000150000000000006d14d8ae0fe200000000200000004000000ac00006d15000000004b00ca000000000200000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000584b00ca0000000001000000000000000000000001000000010000308ae33bb950eb92a8a3004d0bb406899571000000000000000000000000000012f7ed9c9db8c24ed90604eaf0bd04636edb00000000000000000000000100000004000000ac00006d15000000004b0054000000000200000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000584b00540000000001000000003b9aca000000000100000003000067ba8e64fd23fa65f26d0215ec4a0021086600000000000000000000000000002002c3b21045c2378ad3001fb2f300061872000000000000000000000001")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "CgABUA==",
		"samples": [
		   {
			  "counters": [
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 },
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 109679985,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 5049268,
					"ifInOctets": 53373075962192,
					"ifInUcastPkts": 3952257187,
					"ifInUnknownProtos": 0,
					"ifIndex": 1258342912,
					"ifOutBroadcastPkts": 73625307,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 82505917,
					"ifOutOctets": 20856052686264,
					"ifOutUcastPkts": 3259947270,
					"ifPromiscuousMode": 1,
					"ifSpeed": 0,
					"ifStatus": 1,
					"ifType": 1
				 }
			  ],
			  "counters.length": 2,
			  "sampleData.length": 172,
			  "sampleType": 4,
			  "sequenceNumber": 27925,
			  "sourceIdType": 0,
			  "sourceIdValue": 1258342912
		   },
		   {
			  "counters": [
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 },
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 2164838,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 34991178,
					"ifInOctets": 114050950561059,
					"ifInUcastPkts": 4200985197,
					"ifInUnknownProtos": 0,
					"ifIndex": 1258312704,
					"ifOutBroadcastPkts": 399474,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 2077427,
					"ifOutOctets": 35196245250117,
					"ifOutUcastPkts": 3258419923,
					"ifPromiscuousMode": 1,
					"ifSpeed": 1000000000,
					"ifStatus": 3,
					"ifType": 1
				 }
			  ],
			  "counters.length": 2,
			  "sampleData.length": 172,
			  "sampleType": 4,
			  "sequenceNumber": 27925,
			  "sourceIdType": 0,
			  "sourceIdValue": 1258312704
		   }
		],
		"samples.length": 2,
		"sequenceNumber": 446797,
		"subAgentId": 0,
		"uptime": 2330000928,
		"version": 5
	 }	 
	`)
	decodeAndCompare(expectedJSON, packet, t)
}

func Test_sflow_counter_genif_ether(t *testing.T) {
	packet := []byte("0000000500000001c0a80102000000100000f3e70bfb3f590000000400000002000000a800000005000001fc000000020000000100000058000001fc00000006000000003b9aca000000000100000003000000035cfc18b203042a08000000120000004900000000000000000000000000000000c818b33e018afb7d00176fa30021698f00000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001fa000000020000000100000058000001fa00000006000000003b9aca00000000010000000300000132e5eee21da6c2e42d000003fa0000001500000000000000000000000000000100abe764d694ed34b100176bca0021697f00000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001f8000000020000000100000058000001f800000006000000003b9aca000000000100000003000001302c8b23eab41128d2000003e5000000120000000000000000000000000000019abd2b695de4797c3400176c3a0021699100000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001f6000000020000000100000058000001f600000006000000003b9aca0000000001000000030000011dbd163e689cba2cc7000003e5000000520000000000000000000000000000010348cead1888a1e1ae00176c4f00216b89000000000000000000000000000000020000003400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	expectedJSON := []byte(`
	{
		"addressType": "IPV4",
		"agentAddress": "wKgBAg==",
		"samples": [
		   {
			  "counters": [
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 73,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 18,
					"ifInOctets": 14444927154,
					"ifInUcastPkts": 50604552,
					"ifInUnknownProtos": 0,
					"ifIndex": 508,
					"ifOutBroadcastPkts": 2189711,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 1535907,
					"ifOutOctets": 3357061950,
					"ifOutUcastPkts": 25885565,
					"ifPromiscuousMode": 0,
					"ifSpeed": 1000000000,
					"ifStatus": 3,
					"ifType": 6
				 },
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 }
			  ],
			  "counters.length": 2,
			  "sampleData.length": 168,
			  "sampleType": 2,
			  "sequenceNumber": 5,
			  "sourceIdType": 0,
			  "sourceIdValue": 508
		   },
		   {
			  "counters": [
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 21,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 1018,
					"ifInOctets": 1318117630493,
					"ifInUcastPkts": 2797790253,
					"ifInUnknownProtos": 0,
					"ifIndex": 506,
					"ifOutBroadcastPkts": 2189695,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 1534922,
					"ifOutOctets": 1102395696342,
					"ifOutUcastPkts": 2498573489,
					"ifPromiscuousMode": 0,
					"ifSpeed": 1000000000,
					"ifStatus": 3,
					"ifType": 6
				 },
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 }
			  ],
			  "counters.length": 2,
			  "sampleData.length": 168,
			  "sampleType": 2,
			  "sequenceNumber": 5,
			  "sourceIdType": 0,
			  "sourceIdValue": 506
		   },
		   {
			  "counters": [
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 18,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 997,
					"ifInOctets": 1306417374186,
					"ifInUcastPkts": 3021023442,
					"ifInUnknownProtos": 0,
					"ifIndex": 504,
					"ifOutBroadcastPkts": 2189713,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 1535034,
					"ifOutOctets": 1764110330205,
					"ifOutUcastPkts": 3833166900,
					"ifPromiscuousMode": 0,
					"ifSpeed": 1000000000,
					"ifStatus": 3,
					"ifType": 6
				 },
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 }
			  ],
			  "counters.length": 2,
			  "sampleData.length": 168,
			  "sampleType": 2,
			  "sequenceNumber": 5,
			  "sourceIdType": 0,
			  "sourceIdValue": 504
		   },
		   {
			  "counters": [
				 {
					"counterData.length": 88,
					"counterFormat": 1,
					"ifDirection": 1,
					"ifInBroadcastPkts": 82,
					"ifInDiscards": 0,
					"ifInErrors": 0,
					"ifInMulticastPkts": 997,
					"ifInOctets": 1227238030952,
					"ifInUcastPkts": 2629446855,
					"ifInUnknownProtos": 0,
					"ifIndex": 502,
					"ifOutBroadcastPkts": 2190217,
					"ifOutDiscards": 0,
					"ifOutErrors": 0,
					"ifOutMulticastPkts": 1535055,
					"ifOutOctets": 1113618033944,
					"ifOutUcastPkts": 2292310446,
					"ifPromiscuousMode": 0,
					"ifSpeed": 1000000000,
					"ifStatus": 3,
					"ifType": 6
				 },
				 {
					"counterData.length": 52,
					"counterFormat": 2,
					"dot3StatsAlignmentErrors": 0,
					"dot3StatsCarrierSenseErrors": 0,
					"dot3StatsDeferredTransmissions": 0,
					"dot3StatsExcessiveCollisions": 0,
					"dot3StatsFCSErrors": 0,
					"dot3StatsFrameTooLongs": 0,
					"dot3StatsInternalMacReceiveErrors": 0,
					"dot3StatsInternalMacTransmitErrors": 0,
					"dot3StatsLateCollisions": 0,
					"dot3StatsMultipleCollisionFrames": 0,
					"dot3StatsSQETestErrors": 0,
					"dot3StatsSingleCollisionFrames": 0,
					"dot3StatsSymbolErrors": 0
				 }
			  ],
			  "counters.length": 2,
			  "sampleData.length": 168,
			  "sampleType": 2,
			  "sequenceNumber": 5,
			  "sourceIdType": 0,
			  "sourceIdValue": 502
		   }
		],
		"samples.length": 4,
		"sequenceNumber": 62439,
		"subAgentId": 16,
		"uptime": 201015129,
		"version": 5
	 }
		 `)
	decodeAndCompare(expectedJSON, packet, t)
}

// Test_stochasicPacketGeneration randomly (deterministically) modifies a bunch of packets and attempts to
// decode them. It has a short form that runs in < 10s and a longer one that runs for quite a while
// It is looking for panics or general error v decoded count different from expected (previously run examples)
func Test_stochasicPacketGeneration(t *testing.T) {

	iterations := 1000000

	expectedDecodedCount := 1000000
	expectedErrCount := 0
	if testing.Short() {
		iterations = 100000
		expectedDecodedCount = 100000
		expectedErrCount = 0
	}

	r := rand.New(rand.NewSource(0))
	testSelectPacket := func() []byte {
		var src []byte
		switch r.Intn(6) {
		case 0:
			src = []byte("0000000500000001c0a80102000000100000f3d40bfa047f0000000200000001000000d00001210a000001fe000004000484240000000000000001fe00000200000000020000000100000090000000010000010b0000000400000080000c2936d3d694c691aa97600800450000f9f19040004011b4f5c0a80913c0a8090a00a1ba0500e5641f3081da02010104066d6f746f6770a281cc02047b46462e0201000201003081bd3012060d2b06010201190501010281dc710201003013060d2b06010201190501010281e66802025acc3012060d2b0601020119050101000003e9000000100000000900000000000000090000000000000001000000d00000e3cc000002100000400048eb740000000000000002100000020000000002000000010000009000000001000000970000000400000080000c2936d3d6fcecda44008f81000009080045000081186440003f119098c0a80815c0a8090a9a690202006d23083c33303e4170722031312030393a33333a3031206b6e6f64653120736e6d70645b313039385d3a20436f6e6e656374696f6e2066726f6d205544503a205b3139322e3136382e392e31305d3a34393233362d000003e90000001000000009000000000000000900000000")
		case 1:
			src = []byte("00000005000000010a00015000000000000f58998ae119780000000300000003000000c4000b62a90000000000100c840000040024fb7e1e0000000000000000001017840000000000100c8400000001000000010000009000000001000005bc0000000400000080001b17000130001201f58d44810023710800450205a6305440007e06ee92ac100016d94d52f505997e701fa1e17aff62574a50100200355f000000ffff00000b004175746f72697a7a6174610400008040ffff000400008040050031303030320500313030302004000000000868a200000000000000000860a200000000000000000003000000c40003cecf000000000010170400004000a168ac1c000000000000000000101784000000000010170400000001000000010000009000000001000005f200000004000000800024e8324338d4ae52aa0b54810020060800450005dc5420400080061397c0a8060cc0a806080050efcfbb25bad9a21c839a501000fff54000008a55f70975a0ff88b05735597ae274bd81fcba17e6e9206b8ea0fb07d05fc27dad06cfe3fdba5d2fc4d057b0add711e596cbe5e9b4bbe8be59cd77537b7a89f7414a628b736d00000003000000c0000c547a0000000000100c04000004005bc3c3b50000000000000000001017840000000000100c0400000001000000010000008c000000010000007e000000040000007a001b17000130001201f58d448100237108004500006824ea4000ff32c326d94d5105501018f02e88d003000001dd39b1d025d1c68689583b2ab21522d5b5a959642243804f6d51e63323091cc04544285433eb3f6b29e1046a6a2fa7806319d62041d8fa4bd25b7cd85b8db54202054a077ac11de84acbe37a550004")
		case 2:
			src = []byte("000000050000000189dd4f010000000000003d4f21151ad40000000600000001000000bc354b97090000020c000013b175792bea000000000000028f0000020c0000000300000001000000640000000100000058000000040000005408b2587a57624c16fc0b61a5080045000046c3e440003a1118a0052aada7569e5ab367a6e35b0032d7bbf1f2fb2eb2490a97f87abc31e135834be367000002590000ffffffffffffffff02add830d51e0aec14cf000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e32a000000160000000b00000001000000a88b8ffb57000002a2000013b12e344fd800000000000002a20000028f0000000300000001000000500000000100000042000000040000003e4c16fc0b6202c03e0fdecafe080045000030108000007d11fe45575185a718693996f0570e8c001c20614ad602003fd6d4afa6a6d18207324000271169b00000000003e90000001000000000000000000000000000000000000003ea000000100000000189dd4f210000000f0000001800000001000000e8354b970a0000020c000013b175793f9b000000000000028f0000020c00000003000000010000009000000001000001a500000004000000800231466d0b2c4c16fc0b61a5080045000193198f40003a114b75052aae1f5f94c778678ef24d017f50ea7622287c30799e1f7d45932d01ca92c46d930000927c0000ffffffffffffffff02ad0eea6498953d1c7ebb6dbdf0525c80e1a9a62bacfea92f69b7336c2f2f60eba0593509e14eef167eb37449f05ad70b8241c1a46d000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000160000001000000001000000e8354b970b0000020c000013b17579534c000000000000028f0000020c00000003000000010000009000000001000000b500000004000000800231466d0b2c4c16fc0b61a50800450000a327c240003606fd67b93c706a021ff365045fe8a0976d624df8207083501800edb31b0000485454502f312e3120323030204f4b0d0a5365727665723a2050726f746f636f6c20485454500d0a436f6e74656e742d4c656e6774683a20313430340d0a436f6e6e656374696f6e3a20000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000170000001000000001000000e8354b970c0000020c000013b1757966fd000000000000028f0000020c000000030000000100000090000000010000018e00000004000000800231466d0b2c4c16fc0b61a508004500017c7d2c40003a116963052abd8d021c940e67e7e0d501682342dbe7936bd47ef487dee5591ec1b24d83622e000072250000ffffffffffffffff02ad0039d8ba86a90017071d76b177de4d8c4e23bcaaaf4d795f77b032f959e0fb70234d4c28922d4e08dd3330c66e34bff51cc8ade5000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000160000001000000001000000e80d6146ac000002a1000013b17880b49d00000000000002a10000028f00000003000000010000009000000001000005ee00000004000000804c16fc0b6201d8b122766a2c0800450005dc04574000770623a11fcd80a218691d4cf2fe01bbd4f47482065fd63a5010fabd7987000052a20002c8c43ea91ca1eaa115663f5218a37fbb409dfbbedff54731ef41199b35535905ac2366a05a803146ced544abf45597f3714327d59f99e30c899c39fc5a4b67d12087bf8db2bc000003e90000001000000000000000000000000000000000000003ea000000100000000189dd4f210000001000000018")
		case 3:
			src = []byte("00000005000000010ae0648100000002000093d824ac82340000000100000001000000d000019f94000001010000100019f94000000000000000010100000000000000020000000100000090000000010000058c00000008000000800008e3fffc10d4f4be04612486dd60000000054e113a2607f8b0400200140000000000000008262000edc000e804a25e30c581af36fa01bbfa6f054e249810b584bcbf12926c2e29a779c26c72db483e8191524fe2288bfdaceaf9d2e724d04305706efcfdef70db86873bbacf29698affe4e7d6faa21d302f9b4b023291a05a000003e90000001000000001000000000000000100000000")
		case 4:
			src = []byte("00000005000000010a00015000000000000f58898ae0fa380000000700000004000000ec00006ece0000000000101784000000030000000200000034000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000058001017840000000600000002540be400000000010000000300007b8ebd37b97e61ff94860803e8e908ffb2b500000000000000000000000000018e7c31ee7ba4195f041874579ff021ba936300000000000000000000000100000007000000380011223344550003f8b15645e7e7d6960000002fe2fc02fc01edbf580000000000000000000000000000000001dcb9cf000000000000000000000004000000ec00006ece0000000000100184000000030000000200000034000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000058001001840000000600000002540be400000000010000000300000841131d1fd9f850bfb103617cb401e6598900000000000000000000000000000bec1902e5da9212e3e96d7996e922513250000000000000000000000001000000070000003800112233445500005c260acbddb3000100000003e2fc02fc01ee414f0000000000000000000000000000000001dccdd30000000000000000000000030000008400004606000000000010030400004000ad9dc19b0000000000000000001017840000000000100304000000010000000100000050000000010000004400000004000000400012815116c4001517cf426d8100200608004500002895da40008006d74bc0a8060ac0a8064f04ef04aab1797122cf7eaf4f5010ffff7727000000000000000000000003000000b0001bd698000000000010148400000400700b180f000000000000000000101504000000000010148400000001000000010000007c000000010000006f000000040000006b001b17000131f0f755b9afc081000439080045000059045340005206920c1f0d4703d94d52e201bbf14977d1e9f15498af36801800417f1100000101080afdf3c70400e043871503010020ff268cfe2e2fd5fffe1d3d704a91d57b895f174c4b4428c66679d80a307294303f00000003000000c40003ceca000000000010170400004000a166aa7a000000000000000000101784000000000010170400000001000000010000009000000001000005f200000004000000800024e8369e2bd4ae52aa0b54810020060800450005dc4c71400080061b45c0a8060cc0a806090050f855692a7a94a1154ae1801001046b6a00000101080a6869a48d151016d046a84a7aa1c6743fa05179f7ecbd4e567150cb6f2077ff89480ae730637d26d2237c08548806f672c7476eb1b5a447b42cb9ce405994d152fa3e000000030000008c001bd699000000000010148400000400700b180f0000000000000000001015040000000000101484000000010000000100000058000000010000004a0000000400000046001b17000131f0f755b9afc0810004390800450000340ce040003a06bea5c1ce8793d94d528f00504c3b08b18f275b83d5df8010054586ad00000101050a5b83d5de5b83d5df11d800000003000000c400004e07000000000010028400004000c7ec97f2000000000000000000100784000000000010028400000001000000010000009000000001000005f2000000040000008000005e0001ff005056800dd18100000a0800450005dc5a42400040066ef70a000ac8c0a8967201bbe17c81597908caf8a05f5010010328610000f172263da0ba5d6223c079b8238bc841256bf17c4ffb08ad11c4fbff6f87ae1624a6b057b8baa9342114e5f5b46179083020cb560c4e9eadcec6dfd83e102ddbc27024803eb5")
		case 5:
			src = []byte("00000005000000010a000150000000000006d14d8ae0fe200000000200000004000000ac00006d15000000004b00ca000000000200000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000584b00ca0000000001000000000000000000000001000000010000308ae33bb950eb92a8a3004d0bb406899571000000000000000000000000000012f7ed9c9db8c24ed90604eaf0bd04636edb00000000000000000000000100000004000000ac00006d15000000004b0054000000000200000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000584b00540000000001000000003b9aca000000000100000003000067ba8e64fd23fa65f26d0215ec4a0021086600000000000000000000000000002002c3b21045c2378ad3001fb2f300061872000000000000000000000001")
		case 6:
			src = []byte("0000000500000001c0a80102000000100000f3e70bfb3f590000000400000002000000a800000005000001fc000000020000000100000058000001fc00000006000000003b9aca000000000100000003000000035cfc18b203042a08000000120000004900000000000000000000000000000000c818b33e018afb7d00176fa30021698f00000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001fa000000020000000100000058000001fa00000006000000003b9aca00000000010000000300000132e5eee21da6c2e42d000003fa0000001500000000000000000000000000000100abe764d694ed34b100176bca0021697f00000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001f8000000020000000100000058000001f800000006000000003b9aca000000000100000003000001302c8b23eab41128d2000003e5000000120000000000000000000000000000019abd2b695de4797c3400176c3a0021699100000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001f6000000020000000100000058000001f600000006000000003b9aca0000000001000000030000011dbd163e689cba2cc7000003e5000000520000000000000000000000000000010348cead1888a1e1ae00176c4f00216b89000000000000000000000000000000020000003400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		}
		result := make([]byte, len(src))
		copy(result[:], src)
		return result
	}

	packet := testSelectPacket()
	packetBytes := make([]byte, hex.DecodedLen(len(packet)))
	_, err := hex.Decode(packetBytes, packet)
	if err != nil {
		log.Panicln(err)
	}
	errCount := 0
	decodedCount := 0
	for i := 0; i < iterations; i++ {
		if r.Intn(20) <= 15 {
			packet = testSelectPacket()
			packetBytes = make([]byte, hex.DecodedLen(len(packet)))
			_, err := hex.Decode(packetBytes, packet)
			if err != nil {
				log.Panicln(err)
			}
		} else {
			byteToMessWith := r.Intn(len(packetBytes) - 1)
			packetBytes[byteToMessWith] = uint8(r.Intn(255))
		}
		packetBytesHNexDecoded := make([]byte, hex.DecodedLen(len(packet)))
		_, err = hex.Decode(packetBytesHNexDecoded, packet)
		options := NewDefaultV5FormatOptions()
		decoded, err := Decode(V5Format(options), bytes.NewBuffer(packetBytesHNexDecoded))
		if err != nil {
			errCount++
		}
		if decoded != nil {
			decodedCount++
		}
	}

	if errCount != expectedErrCount {
		t.Errorf("errCount %d != expected %d", errCount, expectedErrCount)
	}
	if decodedCount != expectedDecodedCount {
		t.Errorf("decodedCount %d != expected %d", decodedCount, expectedDecodedCount)
	}

}
