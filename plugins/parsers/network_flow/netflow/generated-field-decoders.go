package netflow

// GENERATED do not edit

import (
	"encoding/hex"
	"fmt"

	"github.com/influxdata/telegraf/plugins/parsers/network_flow/decoder"
)

func bytesToMACStr(b []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}

func bytesToHexStr(buf []byte) string {
	return hex.EncodeToString(buf)
}

func getFieldDecoder(elementID uint16, l uint16) decoder.Directive {
	switch elementID {
	// TAGS
	case 4:
		return uintDecoderByLen(l).Do(decoder.AsT("protocolIdentifier"))
	case 5:
		return uintDecoderByLen(l).Do(decoder.AsT("ipClassOfService"))
	case 6:
		return uintDecoderByLen(l).Do(decoder.AsT("tcpControlBits"))
	case 7:
		return uintDecoderByLen(l).Do(decoder.AsT("sourceTransportPort"))
	case 8:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("sourceIPv4Address"))
	case 9:
		return uintDecoderByLen(l).Do(decoder.AsT("sourceIPv4PrefixLength"))
	case 10:
		return uintDecoderByLen(l).Do(decoder.AsT("ingressInterface"))
	case 11:
		return uintDecoderByLen(l).Do(decoder.AsT("destinationTransportPort"))
	case 12:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("destinationIPv4Address"))
	case 13:
		return uintDecoderByLen(l).Do(decoder.AsT("destinationIPv4PrefixLength"))
	case 14:
		return uintDecoderByLen(l).Do(decoder.AsT("egressInterface"))
	case 16:
		return uintDecoderByLen(l).Do(decoder.AsT("bgpSourceAsNumber"))
	case 17:
		return uintDecoderByLen(l).Do(decoder.AsT("bgpDestinationAsNumber"))
	case 18:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("bgpNextHopIPv4Address"))
	case 27:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("sourceIPv6Address"))
	case 28:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("destinationIPv6Address"))
	case 48:
		return uintDecoderByLen(l).Do(decoder.AsT("samplerId"))
	case 61:
		return uintDecoderByLen(l).Do(decoder.AsT("flowDirection"))
	case 70:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsT("mplsTopLabelStackSection"))
	case 89:
		return uintDecoderByLen(l).Do(decoder.AsT("forwardingStatus"))
	case 234:
		return uintDecoderByLen(l).Do(decoder.AsT("ingressVRFID"))
	case 235:
		return uintDecoderByLen(l).Do(decoder.AsT("egressVRFID"))

	// FIELDS
	case 1:
		return uintDecoderByLen(l).Do(decoder.AsF("octetDeltaCount"))
	case 2:
		return uintDecoderByLen(l).Do(decoder.AsF("packetDeltaCount"))
	case 3:
		return uintDecoderByLen(l).Do(decoder.AsF("deltaFlowCount"))
	case 15:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("ipNextHopIPv4Address"))
	case 19:
		return uintDecoderByLen(l).Do(decoder.AsF("postMCastPacketDeltaCount"))
	case 20:
		return uintDecoderByLen(l).Do(decoder.AsF("postMCastOctetDeltaCount"))
	case 21:
		return uintDecoderByLen(l).Do(decoder.AsF("flowEndSysUpTime"))
	case 22:
		return uintDecoderByLen(l).Do(decoder.AsF("flowStartSysUpTime"))
	case 23:
		return uintDecoderByLen(l).Do(decoder.AsF("postOctetDeltaCount"))
	case 24:
		return uintDecoderByLen(l).Do(decoder.AsF("postPacketDeltaCount"))
	case 25:
		return uintDecoderByLen(l).Do(decoder.AsF("minimumIpTotalLength"))
	case 26:
		return uintDecoderByLen(l).Do(decoder.AsF("maximumIpTotalLength"))
	case 29:
		return uintDecoderByLen(l).Do(decoder.AsF("sourceIPv6PrefixLength"))
	case 30:
		return uintDecoderByLen(l).Do(decoder.AsF("destinationIPv6PrefixLength"))
	case 31:
		return uintDecoderByLen(l).Do(decoder.AsF("flowLabelIPv6"))
	case 32:
		return uintDecoderByLen(l).Do(decoder.AsF("icmpTypeCodeIPv4"))
	case 33:
		return uintDecoderByLen(l).Do(decoder.AsF("igmpType"))
	case 34:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingInterval"))
	case 35:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingAlgorithm"))
	case 36:
		return uintDecoderByLen(l).Do(decoder.AsF("flowActiveTimeout"))
	case 37:
		return uintDecoderByLen(l).Do(decoder.AsF("flowIdleTimeout"))
	case 38:
		return uintDecoderByLen(l).Do(decoder.AsF("engineType"))
	case 39:
		return uintDecoderByLen(l).Do(decoder.AsF("engineId"))
	case 40:
		return uintDecoderByLen(l).Do(decoder.AsF("exportedOctetTotalCount"))
	case 41:
		return uintDecoderByLen(l).Do(decoder.AsF("exportedMessageTotalCount"))
	case 42:
		return uintDecoderByLen(l).Do(decoder.AsF("exportedFlowRecordTotalCount"))
	case 43:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("ipv4RouterSc"))
	case 44:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("sourceIPv4Prefix"))
	case 45:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("destinationIPv4Prefix"))
	case 46:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsTopLabelType"))
	case 47:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("mplsTopLabelIPv4Address"))
	case 49:
		return uintDecoderByLen(l).Do(decoder.AsF("samplerMode"))
	case 50:
		return uintDecoderByLen(l).Do(decoder.AsF("samplerRandomInterval"))
	case 51:
		return uintDecoderByLen(l).Do(decoder.AsF("classId"))
	case 52:
		return uintDecoderByLen(l).Do(decoder.AsF("minimumTTL"))
	case 53:
		return uintDecoderByLen(l).Do(decoder.AsF("maximumTTL"))
	case 54:
		return uintDecoderByLen(l).Do(decoder.AsF("fragmentIdentification"))
	case 55:
		return uintDecoderByLen(l).Do(decoder.AsF("postIpClassOfService"))
	case 56:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("sourceMacAddress"))
	case 57:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("postDestinationMacAddress"))
	case 58:
		return uintDecoderByLen(l).Do(decoder.AsF("vlanId"))
	case 59:
		return uintDecoderByLen(l).Do(decoder.AsF("postVlanId"))
	case 60:
		return uintDecoderByLen(l).Do(decoder.AsF("ipVersion"))
	case 62:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("ipNextHopIPv6Address"))
	case 63:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("bgpNextHopIPv6Address"))
	case 64:
		return uintDecoderByLen(l).Do(decoder.AsF("ipv6ExtensionHeaders"))
	case 71:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection2"))
	case 72:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection3"))
	case 73:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection4"))
	case 74:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection5"))
	case 75:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection6"))
	case 76:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection7"))
	case 77:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection8"))
	case 78:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection9"))
	case 79:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection10"))
	case 80:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("destinationMacAddress"))
	case 81:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("postSourceMacAddress"))
	case 85:
		return uintDecoderByLen(l).Do(decoder.AsF("octetTotalCount"))
	case 86:
		return uintDecoderByLen(l).Do(decoder.AsF("packetTotalCount"))
	case 87:
		return uintDecoderByLen(l).Do(decoder.AsF("flagsAndSamplerId"))
	case 88:
		return uintDecoderByLen(l).Do(decoder.AsF("fragmentOffset"))
	case 90:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsVpnRouteDistinguisher"))
	case 91:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsTopLabelPrefixLength"))
	case 92:
		return uintDecoderByLen(l).Do(decoder.AsF("srcTrafficIndex"))
	case 93:
		return uintDecoderByLen(l).Do(decoder.AsF("dstTrafficIndex"))
	case 95:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("applicationId"))
	case 98:
		return uintDecoderByLen(l).Do(decoder.AsF("postIpDiffServCodePoint"))
	case 99:
		return uintDecoderByLen(l).Do(decoder.AsF("multicastReplicationFactor"))
	case 101:
		return uintDecoderByLen(l).Do(decoder.AsF("classificationEngineId"))
	case 102:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2packetSectionOffset"))
	case 103:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2packetSectionSize"))
	case 104:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("layer2packetSectionData"))
	case 128:
		return uintDecoderByLen(l).Do(decoder.AsF("bgpNextAdjacentAsNumber"))
	case 129:
		return uintDecoderByLen(l).Do(decoder.AsF("bgpPrevAdjacentAsNumber"))
	case 130:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("exporterIPv4Address"))
	case 131:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("exporterIPv6Address"))
	case 132:
		return uintDecoderByLen(l).Do(decoder.AsF("droppedOctetDeltaCount"))
	case 133:
		return uintDecoderByLen(l).Do(decoder.AsF("droppedPacketDeltaCount"))
	case 134:
		return uintDecoderByLen(l).Do(decoder.AsF("droppedOctetTotalCount"))
	case 135:
		return uintDecoderByLen(l).Do(decoder.AsF("droppedPacketTotalCount"))
	case 136:
		return uintDecoderByLen(l).Do(decoder.AsF("flowEndReason"))
	case 137:
		return uintDecoderByLen(l).Do(decoder.AsF("commonPropertiesId"))
	case 138:
		return uintDecoderByLen(l).Do(decoder.AsF("observationPointId"))
	case 139:
		return uintDecoderByLen(l).Do(decoder.AsF("icmpTypeCodeIPv6"))
	case 140:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("mplsTopLabelIPv6Address"))
	case 141:
		return uintDecoderByLen(l).Do(decoder.AsF("lineCardId"))
	case 142:
		return uintDecoderByLen(l).Do(decoder.AsF("portId"))
	case 143:
		return uintDecoderByLen(l).Do(decoder.AsF("meteringProcessId"))
	case 144:
		return uintDecoderByLen(l).Do(decoder.AsF("exportingProcessId"))
	case 145:
		return uintDecoderByLen(l).Do(decoder.AsF("templateId"))
	case 146:
		return uintDecoderByLen(l).Do(decoder.AsF("wlanChannelId"))
	case 148:
		return uintDecoderByLen(l).Do(decoder.AsF("flowId"))
	case 149:
		return uintDecoderByLen(l).Do(decoder.AsF("observationDomainId"))
	case 158:
		return uintDecoderByLen(l).Do(decoder.AsF("flowStartDeltaMicroseconds"))
	case 159:
		return uintDecoderByLen(l).Do(decoder.AsF("flowEndDeltaMicroseconds"))
	case 161:
		return uintDecoderByLen(l).Do(decoder.AsF("flowDurationMilliseconds"))
	case 162:
		return uintDecoderByLen(l).Do(decoder.AsF("flowDurationMicroseconds"))
	case 163:
		return uintDecoderByLen(l).Do(decoder.AsF("observedFlowTotalCount"))
	case 164:
		return uintDecoderByLen(l).Do(decoder.AsF("ignoredPacketTotalCount"))
	case 165:
		return uintDecoderByLen(l).Do(decoder.AsF("ignoredOctetTotalCount"))
	case 166:
		return uintDecoderByLen(l).Do(decoder.AsF("notSentFlowTotalCount"))
	case 167:
		return uintDecoderByLen(l).Do(decoder.AsF("notSentPacketTotalCount"))
	case 168:
		return uintDecoderByLen(l).Do(decoder.AsF("notSentOctetTotalCount"))
	case 169:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("destinationIPv6Prefix"))
	case 170:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("sourceIPv6Prefix"))
	case 171:
		return uintDecoderByLen(l).Do(decoder.AsF("postOctetTotalCount"))
	case 172:
		return uintDecoderByLen(l).Do(decoder.AsF("postPacketTotalCount"))
	case 173:
		return uintDecoderByLen(l).Do(decoder.AsF("flowKeyIndicator"))
	case 174:
		return uintDecoderByLen(l).Do(decoder.AsF("postMCastPacketTotalCount"))
	case 175:
		return uintDecoderByLen(l).Do(decoder.AsF("postMCastOctetTotalCount"))
	case 176:
		return uintDecoderByLen(l).Do(decoder.AsF("icmpTypeIPv4"))
	case 177:
		return uintDecoderByLen(l).Do(decoder.AsF("icmpCodeIPv4"))
	case 178:
		return uintDecoderByLen(l).Do(decoder.AsF("icmpTypeIPv6"))
	case 179:
		return uintDecoderByLen(l).Do(decoder.AsF("icmpCodeIPv6"))
	case 180:
		return uintDecoderByLen(l).Do(decoder.AsF("udpSourcePort"))
	case 181:
		return uintDecoderByLen(l).Do(decoder.AsF("udpDestinationPort"))
	case 182:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpSourcePort"))
	case 183:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpDestinationPort"))
	case 184:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpSequenceNumber"))
	case 185:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpAcknowledgementNumber"))
	case 186:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpWindowSize"))
	case 187:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpUrgentPointer"))
	case 188:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpHeaderLength"))
	case 189:
		return uintDecoderByLen(l).Do(decoder.AsF("ipHeaderLength"))
	case 190:
		return uintDecoderByLen(l).Do(decoder.AsF("totalLengthIPv4"))
	case 191:
		return uintDecoderByLen(l).Do(decoder.AsF("payloadLengthIPv6"))
	case 192:
		return uintDecoderByLen(l).Do(decoder.AsF("ipTTL"))
	case 193:
		return uintDecoderByLen(l).Do(decoder.AsF("nextHeaderIPv6"))
	case 194:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsPayloadLength"))
	case 195:
		return uintDecoderByLen(l).Do(decoder.AsF("ipDiffServCodePoint"))
	case 196:
		return uintDecoderByLen(l).Do(decoder.AsF("ipPrecedence"))
	case 197:
		return uintDecoderByLen(l).Do(decoder.AsF("fragmentFlags"))
	case 198:
		return uintDecoderByLen(l).Do(decoder.AsF("octetDeltaSumOfSquares"))
	case 199:
		return uintDecoderByLen(l).Do(decoder.AsF("octetTotalSumOfSquares"))
	case 200:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsTopLabelTTL"))
	case 201:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsLabelStackLength"))
	case 202:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsLabelStackDepth"))
	case 203:
		return uintDecoderByLen(l).Do(decoder.AsF("mplsTopLabelExp"))
	case 204:
		return uintDecoderByLen(l).Do(decoder.AsF("ipPayloadLength"))
	case 205:
		return uintDecoderByLen(l).Do(decoder.AsF("udpMessageLength"))
	case 206:
		return uintDecoderByLen(l).Do(decoder.AsF("isMulticast"))
	case 207:
		return uintDecoderByLen(l).Do(decoder.AsF("ipv4IHL"))
	case 208:
		return uintDecoderByLen(l).Do(decoder.AsF("ipv4Options"))
	case 209:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpOptions"))
	case 210:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("paddingOctets"))
	case 211:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("collectorIPv4Address"))
	case 212:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("collectorIPv6Address"))
	case 213:
		return uintDecoderByLen(l).Do(decoder.AsF("exportInterface"))
	case 214:
		return uintDecoderByLen(l).Do(decoder.AsF("exportProtocolVersion"))
	case 215:
		return uintDecoderByLen(l).Do(decoder.AsF("exportTransportProtocol"))
	case 216:
		return uintDecoderByLen(l).Do(decoder.AsF("collectorTransportPort"))
	case 217:
		return uintDecoderByLen(l).Do(decoder.AsF("exporterTransportPort"))
	case 218:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpSynTotalCount"))
	case 219:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpFinTotalCount"))
	case 220:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpRstTotalCount"))
	case 221:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpPshTotalCount"))
	case 222:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpAckTotalCount"))
	case 223:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpUrgTotalCount"))
	case 224:
		return uintDecoderByLen(l).Do(decoder.AsF("ipTotalLength"))
	case 225:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("postNATSourceIPv4Address"))
	case 226:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("postNATDestinationIPv4Address"))
	case 227:
		return uintDecoderByLen(l).Do(decoder.AsF("postNAPTSourceTransportPort"))
	case 228:
		return uintDecoderByLen(l).Do(decoder.AsF("postNAPTDestinationTransportPort"))
	case 229:
		return uintDecoderByLen(l).Do(decoder.AsF("natOriginatingAddressRealm"))
	case 230:
		return uintDecoderByLen(l).Do(decoder.AsF("natEvent"))
	case 231:
		return uintDecoderByLen(l).Do(decoder.AsF("initiatorOctets"))
	case 232:
		return uintDecoderByLen(l).Do(decoder.AsF("responderOctets"))
	case 233:
		return uintDecoderByLen(l).Do(decoder.AsF("firewallEvent"))
	case 237:
		return uintDecoderByLen(l).Do(decoder.AsF("postMplsTopLabelExp"))
	case 238:
		return uintDecoderByLen(l).Do(decoder.AsF("tcpWindowScale"))
	case 239:
		return uintDecoderByLen(l).Do(decoder.AsF("biflowDirection"))
	case 240:
		return uintDecoderByLen(l).Do(decoder.AsF("ethernetHeaderLength"))
	case 241:
		return uintDecoderByLen(l).Do(decoder.AsF("ethernetPayloadLength"))
	case 242:
		return uintDecoderByLen(l).Do(decoder.AsF("ethernetTotalLength"))
	case 243:
		return uintDecoderByLen(l).Do(decoder.AsF("dot1qVlanId"))
	case 244:
		return uintDecoderByLen(l).Do(decoder.AsF("dot1qPriority"))
	case 245:
		return uintDecoderByLen(l).Do(decoder.AsF("dot1qCustomerVlanId"))
	case 246:
		return uintDecoderByLen(l).Do(decoder.AsF("dot1qCustomerPriority"))
	case 248:
		return uintDecoderByLen(l).Do(decoder.AsF("metroEvcType"))
	case 249:
		return uintDecoderByLen(l).Do(decoder.AsF("pseudoWireId"))
	case 250:
		return uintDecoderByLen(l).Do(decoder.AsF("pseudoWireType"))
	case 251:
		return uintDecoderByLen(l).Do(decoder.AsF("pseudoWireControlWord"))
	case 252:
		return uintDecoderByLen(l).Do(decoder.AsF("ingressPhysicalInterface"))
	case 253:
		return uintDecoderByLen(l).Do(decoder.AsF("egressPhysicalInterface"))
	case 254:
		return uintDecoderByLen(l).Do(decoder.AsF("postDot1qVlanId"))
	case 255:
		return uintDecoderByLen(l).Do(decoder.AsF("postDot1qCustomerVlanId"))
	case 256:
		return uintDecoderByLen(l).Do(decoder.AsF("ethernetType"))
	case 257:
		return uintDecoderByLen(l).Do(decoder.AsF("postIpPrecedence"))
	case 259:
		return uintDecoderByLen(l).Do(decoder.AsF("exportSctpStreamId"))
	case 262:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("messageMD5Checksum"))
	case 263:
		return uintDecoderByLen(l).Do(decoder.AsF("messageScope"))
	case 266:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("opaqueOctets"))
	case 267:
		return uintDecoderByLen(l).Do(decoder.AsF("sessionScope"))
	case 274:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("collectorCertificate"))
	case 275:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("exporterCertificate"))
	case 277:
		return uintDecoderByLen(l).Do(decoder.AsF("observationPointType"))
	case 278:
		return uintDecoderByLen(l).Do(decoder.AsF("newConnectionDeltaCount"))
	case 279:
		return uintDecoderByLen(l).Do(decoder.AsF("connectionSumDurationSeconds"))
	case 280:
		return uintDecoderByLen(l).Do(decoder.AsF("connectionTransactionId"))
	case 281:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("postNATSourceIPv6Address"))
	case 282:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("postNATDestinationIPv6Address"))
	case 283:
		return uintDecoderByLen(l).Do(decoder.AsF("natPoolId"))
	case 285:
		return uintDecoderByLen(l).Do(decoder.AsF("anonymizationFlags"))
	case 286:
		return uintDecoderByLen(l).Do(decoder.AsF("anonymizationTechnique"))
	case 287:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementIndex"))
	case 294:
		return uintDecoderByLen(l).Do(decoder.AsF("bgpValidityState"))
	case 295:
		return uintDecoderByLen(l).Do(decoder.AsF("IPSecSPI"))
	case 296:
		return uintDecoderByLen(l).Do(decoder.AsF("greKey"))
	case 297:
		return uintDecoderByLen(l).Do(decoder.AsF("natType"))
	case 298:
		return uintDecoderByLen(l).Do(decoder.AsF("initiatorPackets"))
	case 299:
		return uintDecoderByLen(l).Do(decoder.AsF("responderPackets"))
	case 301:
		return uintDecoderByLen(l).Do(decoder.AsF("selectionSequenceId"))
	case 302:
		return uintDecoderByLen(l).Do(decoder.AsF("selectorId"))
	case 303:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementId"))
	case 304:
		return uintDecoderByLen(l).Do(decoder.AsF("selectorAlgorithm"))
	case 305:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingPacketInterval"))
	case 306:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingPacketSpace"))
	case 307:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingTimeInterval"))
	case 308:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingTimeSpace"))
	case 309:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingSize"))
	case 310:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingPopulation"))
	case 312:
		return uintDecoderByLen(l).Do(decoder.AsF("dataLinkFrameSize"))
	case 313:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("ipHeaderPacketSection"))
	case 314:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("ipPayloadPacketSection"))
	case 315:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("dataLinkFrameSection"))
	case 316:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsLabelStackSection"))
	case 317:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mplsPayloadPacketSection"))
	case 318:
		return uintDecoderByLen(l).Do(decoder.AsF("selectorIdTotalPktsObserved"))
	case 319:
		return uintDecoderByLen(l).Do(decoder.AsF("selectorIdTotalPktsSelected"))
	case 326:
		return uintDecoderByLen(l).Do(decoder.AsF("digestHashValue"))
	case 327:
		return uintDecoderByLen(l).Do(decoder.AsF("hashIPPayloadOffset"))
	case 328:
		return uintDecoderByLen(l).Do(decoder.AsF("hashIPPayloadSize"))
	case 329:
		return uintDecoderByLen(l).Do(decoder.AsF("hashOutputRangeMin"))
	case 330:
		return uintDecoderByLen(l).Do(decoder.AsF("hashOutputRangeMax"))
	case 331:
		return uintDecoderByLen(l).Do(decoder.AsF("hashSelectedRangeMin"))
	case 332:
		return uintDecoderByLen(l).Do(decoder.AsF("hashSelectedRangeMax"))
	case 334:
		return uintDecoderByLen(l).Do(decoder.AsF("hashInitialiserValue"))
	case 339:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementDataType"))
	case 342:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementRangeBegin"))
	case 343:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementRangeEnd"))
	case 344:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementSemantics"))
	case 345:
		return uintDecoderByLen(l).Do(decoder.AsF("informationElementUnits"))
	case 346:
		return uintDecoderByLen(l).Do(decoder.AsF("privateEnterpriseNumber"))
	case 347:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("virtualStationInterfaceId"))
	case 349:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("virtualStationUUID"))
	case 351:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2SegmentId"))
	case 352:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2OctetDeltaCount"))
	case 353:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2OctetTotalCount"))
	case 354:
		return uintDecoderByLen(l).Do(decoder.AsF("ingressUnicastPacketTotalCount"))
	case 355:
		return uintDecoderByLen(l).Do(decoder.AsF("ingressMulticastPacketTotalCount"))
	case 356:
		return uintDecoderByLen(l).Do(decoder.AsF("ingressBroadcastPacketTotalCount"))
	case 357:
		return uintDecoderByLen(l).Do(decoder.AsF("egressUnicastPacketTotalCount"))
	case 358:
		return uintDecoderByLen(l).Do(decoder.AsF("egressBroadcastPacketTotalCount"))
	case 361:
		return uintDecoderByLen(l).Do(decoder.AsF("portRangeStart"))
	case 362:
		return uintDecoderByLen(l).Do(decoder.AsF("portRangeEnd"))
	case 363:
		return uintDecoderByLen(l).Do(decoder.AsF("portRangeStepSize"))
	case 364:
		return uintDecoderByLen(l).Do(decoder.AsF("portRangeNumPorts"))
	case 365:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("staMacAddress"))
	case 366:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("staIPv4Address"))
	case 367:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("wtpMacAddress"))
	case 368:
		return uintDecoderByLen(l).Do(decoder.AsF("ingressInterfaceType"))
	case 369:
		return uintDecoderByLen(l).Do(decoder.AsF("egressInterfaceType"))
	case 370:
		return uintDecoderByLen(l).Do(decoder.AsF("rtpSequenceNumber"))
	case 375:
		return uintDecoderByLen(l).Do(decoder.AsF("originalFlowsPresent"))
	case 376:
		return uintDecoderByLen(l).Do(decoder.AsF("originalFlowsInitiated"))
	case 377:
		return uintDecoderByLen(l).Do(decoder.AsF("originalFlowsCompleted"))
	case 378:
		return uintDecoderByLen(l).Do(decoder.AsF("distinctCountOfSourceIPAddress"))
	case 379:
		return uintDecoderByLen(l).Do(decoder.AsF("distinctCountOfDestinationIPAddress"))
	case 380:
		return uintDecoderByLen(l).Do(decoder.AsF("distinctCountOfSourceIPv4Address"))
	case 381:
		return uintDecoderByLen(l).Do(decoder.AsF("distinctCountOfDestinationIPv4Address"))
	case 382:
		return uintDecoderByLen(l).Do(decoder.AsF("distinctCountOfSourceIPv6Address"))
	case 383:
		return uintDecoderByLen(l).Do(decoder.AsF("distinctCountOfDestinationIPv6Address"))
	case 384:
		return uintDecoderByLen(l).Do(decoder.AsF("valueDistributionMethod"))
	case 385:
		return uintDecoderByLen(l).Do(decoder.AsF("rfc3550JitterMilliseconds"))
	case 386:
		return uintDecoderByLen(l).Do(decoder.AsF("rfc3550JitterMicroseconds"))
	case 387:
		return uintDecoderByLen(l).Do(decoder.AsF("rfc3550JitterNanoseconds"))
	case 390:
		return uintDecoderByLen(l).Do(decoder.AsF("flowSelectorAlgorithm"))
	case 391:
		return uintDecoderByLen(l).Do(decoder.AsF("flowSelectedOctetDeltaCount"))
	case 392:
		return uintDecoderByLen(l).Do(decoder.AsF("flowSelectedPacketDeltaCount"))
	case 393:
		return uintDecoderByLen(l).Do(decoder.AsF("flowSelectedFlowDeltaCount"))
	case 394:
		return uintDecoderByLen(l).Do(decoder.AsF("selectorIDTotalFlowsObserved"))
	case 395:
		return uintDecoderByLen(l).Do(decoder.AsF("selectorIDTotalFlowsSelected"))
	case 396:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingFlowInterval"))
	case 397:
		return uintDecoderByLen(l).Do(decoder.AsF("samplingFlowSpacing"))
	case 398:
		return uintDecoderByLen(l).Do(decoder.AsF("flowSamplingTimeInterval"))
	case 399:
		return uintDecoderByLen(l).Do(decoder.AsF("flowSamplingTimeSpacing"))
	case 400:
		return uintDecoderByLen(l).Do(decoder.AsF("hashFlowDomain"))
	case 401:
		return uintDecoderByLen(l).Do(decoder.AsF("transportOctetDeltaCount"))
	case 402:
		return uintDecoderByLen(l).Do(decoder.AsF("transportPacketDeltaCount"))
	case 403:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("originalExporterIPv4Address"))
	case 404:
		return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsF("originalExporterIPv6Address"))
	case 405:
		return uintDecoderByLen(l).Do(decoder.AsF("originalObservationDomainId"))
	case 406:
		return uintDecoderByLen(l).Do(decoder.AsF("intermediateProcessId"))
	case 407:
		return uintDecoderByLen(l).Do(decoder.AsF("ignoredDataRecordTotalCount"))
	case 408:
		return uintDecoderByLen(l).Do(decoder.AsF("dataLinkFrameType"))
	case 409:
		return uintDecoderByLen(l).Do(decoder.AsF("sectionOffset"))
	case 410:
		return uintDecoderByLen(l).Do(decoder.AsF("sectionExportedOctets"))
	case 411:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("dot1qServiceInstanceTag"))
	case 412:
		return uintDecoderByLen(l).Do(decoder.AsF("dot1qServiceInstanceId"))
	case 413:
		return uintDecoderByLen(l).Do(decoder.AsF("dot1qServiceInstancePriority"))
	case 414:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("dot1qCustomerSourceMacAddress"))
	case 415:
		return bytesDecoderByLen(l, 6).Do(decoder.BytesToStr(6, bytesToMACStr).AsF("dot1qCustomerDestinationMacAddress"))
	case 417:
		return uintDecoderByLen(l).Do(decoder.AsF("postLayer2OctetDeltaCount"))
	case 418:
		return uintDecoderByLen(l).Do(decoder.AsF("postMCastLayer2OctetDeltaCount"))
	case 420:
		return uintDecoderByLen(l).Do(decoder.AsF("postLayer2OctetTotalCount"))
	case 421:
		return uintDecoderByLen(l).Do(decoder.AsF("postMCastLayer2OctetTotalCount"))
	case 422:
		return uintDecoderByLen(l).Do(decoder.AsF("minimumLayer2TotalLength"))
	case 423:
		return uintDecoderByLen(l).Do(decoder.AsF("maximumLayer2TotalLength"))
	case 424:
		return uintDecoderByLen(l).Do(decoder.AsF("droppedLayer2OctetDeltaCount"))
	case 425:
		return uintDecoderByLen(l).Do(decoder.AsF("droppedLayer2OctetTotalCount"))
	case 426:
		return uintDecoderByLen(l).Do(decoder.AsF("ignoredLayer2OctetTotalCount"))
	case 427:
		return uintDecoderByLen(l).Do(decoder.AsF("notSentLayer2OctetTotalCount"))
	case 428:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2OctetDeltaSumOfSquares"))
	case 429:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2OctetTotalSumOfSquares"))
	case 430:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2FrameDeltaCount"))
	case 431:
		return uintDecoderByLen(l).Do(decoder.AsF("layer2FrameTotalCount"))
	case 432:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("pseudoWireDestinationIPv4Address"))
	case 433:
		return uintDecoderByLen(l).Do(decoder.AsF("ignoredLayer2FrameTotalCount"))
	case 435:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mibObjectValueOctetString"))
	case 436:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mibObjectValueOID"))
	case 437:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mibObjectValueBits"))
	case 438:
		return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsF("mibObjectValueIPAddress"))
	case 439:
		return uintDecoderByLen(l).Do(decoder.AsF("mibObjectValueCounter"))
	case 440:
		return uintDecoderByLen(l).Do(decoder.AsF("mibObjectValueGauge"))
	case 441:
		return uintDecoderByLen(l).Do(decoder.AsF("mibObjectValueTimeTicks"))
	case 442:
		return uintDecoderByLen(l).Do(decoder.AsF("mibObjectValueUnsigned"))
	case 445:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mibObjectIdentifier"))
	case 446:
		return uintDecoderByLen(l).Do(decoder.AsF("mibSubIdentifier"))
	case 447:
		return uintDecoderByLen(l).Do(decoder.AsF("mibIndexIndicator"))
	case 448:
		return uintDecoderByLen(l).Do(decoder.AsF("mibCaptureTimeSemantics"))
	case 449:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("mibContextEngineID"))
	case 457:
		return uintDecoderByLen(l).Do(decoder.AsF("httpStatusCode"))
	case 458:
		return uintDecoderByLen(l).Do(decoder.AsF("sourceTransportPortsLimit"))
	case 463:
		return uintDecoderByLen(l).Do(decoder.AsF("natInstanceID"))
	case 464:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("internalAddressRealm"))
	case 465:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("externalAddressRealm"))
	case 466:
		return uintDecoderByLen(l).Do(decoder.AsF("natQuotaExceededEvent"))
	case 467:
		return uintDecoderByLen(l).Do(decoder.AsF("natThresholdEvent"))
	case 471:
		return uintDecoderByLen(l).Do(decoder.AsF("maxSessionEntries"))
	case 472:
		return uintDecoderByLen(l).Do(decoder.AsF("maxBIBEntries"))
	case 473:
		return uintDecoderByLen(l).Do(decoder.AsF("maxEntriesPerUser"))
	case 474:
		return uintDecoderByLen(l).Do(decoder.AsF("maxSubscribers"))
	case 475:
		return uintDecoderByLen(l).Do(decoder.AsF("maxFragmentsPendingReassembly"))
	case 476:
		return uintDecoderByLen(l).Do(decoder.AsF("addressPoolHighThreshold"))
	case 477:
		return uintDecoderByLen(l).Do(decoder.AsF("addressPoolLowThreshold"))
	case 478:
		return uintDecoderByLen(l).Do(decoder.AsF("addressPortMappingHighThreshold"))
	case 479:
		return uintDecoderByLen(l).Do(decoder.AsF("addressPortMappingLowThreshold"))
	case 480:
		return uintDecoderByLen(l).Do(decoder.AsF("addressPortMappingPerUserHighThreshold"))
	case 481:
		return uintDecoderByLen(l).Do(decoder.AsF("globalAddressMappingHighThreshold"))
	case 482:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("vpnIdentifier"))
	case 483:
		return uintDecoderByLen(l).Do(decoder.AsF("bgpCommunity"))
	case 486:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("bgpExtendedCommunity"))
	case 489:
		return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToHexStr).AsF("bgpLargeCommunity"))

	default:
		// UNHANDLED AT THE MOMENT
		// ID TYPE NAME
		// 82 string interfaceName
		// 83 string interfaceDescription
		// 84 string samplerName
		// 94 string applicationDescription
		// 96 string applicationName
		// 97  Assigned for NetFlow v9 compatibility
		// 100 string className
		// 147 string wlanSSID
		// 150 dateTimeSeconds flowStartSeconds
		// 151 dateTimeSeconds flowEndSeconds
		// 152 dateTimeMilliseconds flowStartMilliseconds
		// 153 dateTimeMilliseconds flowEndMilliseconds
		// 154 dateTimeMicroseconds flowStartMicroseconds
		// 155 dateTimeMicroseconds flowEndMicroseconds
		// 156 dateTimeNanoseconds flowStartNanoseconds
		// 157 dateTimeNanoseconds flowEndNanoseconds
		// 160 dateTimeMilliseconds systemInitTimeMilliseconds
		// 236 string VRFname
		// 247 string metroEvcId
		// 258 dateTimeMilliseconds collectionTimeMilliseconds
		// 260 dateTimeSeconds maxExportSeconds
		// 261 dateTimeSeconds maxFlowEndSeconds
		// 264 dateTimeSeconds minExportSeconds
		// 265 dateTimeSeconds minFlowStartSeconds
		// 268 dateTimeMicroseconds maxFlowEndMicroseconds
		// 269 dateTimeMilliseconds maxFlowEndMilliseconds
		// 270 dateTimeNanoseconds maxFlowEndNanoseconds
		// 271 dateTimeMicroseconds minFlowStartMicroseconds
		// 272 dateTimeMilliseconds minFlowStartMilliseconds
		// 273 dateTimeNanoseconds minFlowStartNanoseconds
		// 276 boolean dataRecordsReliability
		// 284 string natPoolName
		// 288 string p2pTechnology
		// 289 string tunnelTechnology
		// 290 string encryptedTechnology
		// 291 basicList basicList
		// 292 subTemplateList subTemplateList
		// 293 subTemplateMultiList subTemplateMultiList
		// 300 string observationDomainName
		// 311 float64 samplingProbability
		// 320 float64 absoluteError
		// 321 float64 relativeError
		// 322 dateTimeSeconds observationTimeSeconds
		// 323 dateTimeMilliseconds observationTimeMilliseconds
		// 324 dateTimeMicroseconds observationTimeMicroseconds
		// 325 dateTimeNanoseconds observationTimeNanoseconds
		// 333 boolean hashDigestOutput
		// 335 string selectorName
		// 336 float64 upperCILimit
		// 337 float64 lowerCILimit
		// 338 float64 confidenceLevel
		// 340 string informationElementDescription
		// 341 string informationElementName
		// 348 string virtualStationInterfaceName
		// 350 string virtualStationName
		// 359 dateTimeMilliseconds monitoringIntervalStartMilliSeconds
		// 360 dateTimeMilliseconds monitoringIntervalEndMilliSeconds
		// 371 string userName
		// 372 string applicationCategoryName
		// 373 string applicationSubCategoryName
		// 374 string applicationGroupName
		// 388 boolean dot1qDEI
		// 389 boolean dot1qCustomerDEI
		// 434 signed32 mibObjectValueInteger
		// 443 subTemplateList mibObjectValueTable
		// 444 subTemplateList mibObjectValueRow
		// 450 string mibContextName
		// 451 string mibObjectName
		// 452 string mibObjectDescription
		// 453 string mibObjectSyntax
		// 454 string mibModuleName
		// 455 string mobileIMSI
		// 456 string mobileMSISDN
		// 459 string httpRequestMethod
		// 460 string httpRequestHost
		// 461 string httpRequestTarget
		// 462 string httpMessageVersion
		// 468 string httpUserAgent
		// 469 string httpContentType
		// 470 string httpReasonPhrase
		// 484 basicList bgpSourceCommunityList
		// 485 basicList bgpDestinationCommunityList
		// 487 basicList bgpSourceExtendedCommunityList
		// 488 basicList bgpDestinationExtendedCommunityList
		// 490 basicList bgpSourceLargeCommunityList
		// 491 basicList bgpDestinationLargeCommunityList
		return bytesDecoderByLen(l, l)
	}
}
