package netflow

import (
	"fmt"

	"github.com/influxdata/telegraf/plugins/parsers/network_flow/decoder"
)

// lenDrivenDecoder function signature for a fn that given a length in bytes, will return an appropriate
// unsigned integer decoder directive or error
type lenDrivenDecoder func(len uint16) decoder.Directive

func uintDecoderByLen(len uint16) decoder.ValueDirective {
	switch len {
	case 1:
		return decoder.U8()
	case 2:
		return decoder.U16()
	case 4:
		return decoder.U32()
	case 8:
		return decoder.U64()
	default:
		panic(fmt.Sprintf("no decoder.U%d available", len))
	}
}

func bytesDecoderByLen(len uint16, assertion uint16) decoder.ValueDirective {
	if len != assertion {
		panic(fmt.Sprintf("decoder.Bytes(%d) not possible - hard requirement Bytes(%d) ", len, assertion))
	}
	return decoder.Bytes(int(assertion))
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml
var fieldDecoderMap map[uint16]lenDrivenDecoder

func init() {
	// https://www.iana.org/assignments/ipfix/ipfix.xhtml
	fieldDecoderMap = map[uint16]lenDrivenDecoder{
		1: uintByLenAsF("octetDeltaCount"),
		2: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsF("packetDeltaCount")) },
		4: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("protocolIdentifier")) },
		5: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("ipClassOfService")) },
		6: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("tcpControlBits")) },
		7: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("sourceTransportPort")) },
		8: func(l uint16) decoder.Directive {
			return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("sourceIPv4Address"))
		},
		9:  func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("sourceIPv4PrefixLength")) },
		10: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("ingressInterface")) },
		11: func(l uint16) decoder.Directive {
			return uintDecoderByLen(l).Do(decoder.AsT("destinationTransportPort"))
		},
		12: func(l uint16) decoder.Directive {
			return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("destinationIPv4Address"))
		},
		13: func(l uint16) decoder.Directive {
			return uintDecoderByLen(l).Do(decoder.AsT("destinationIPv4PrefixLength"))
		},
		14: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("egressInterface")) },
		16: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("bgpSourceAsNumber")) },
		17: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("bgpDestinationAsNumber")) },
		18: func(l uint16) decoder.Directive {
			return bytesDecoderByLen(l, 4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("bgpNextHopIPv4Address"))
		},
		21: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsF("flowEndSysUpTime")) },
		22: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsF("flowStartSysUpTime")) },
		27: func(l uint16) decoder.Directive {
			return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("sourceIPv6Address"))
		},
		28: func(l uint16) decoder.Directive {
			return bytesDecoderByLen(l, 16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("destinationIPv6Address"))
		},
		48: func(l uint16) decoder.Directive {
			return uintDecoderByLen(l).Do(decoder.AsT("samplerId"))
		},
		61: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("flowDirection")) },
		70: func(l uint16) decoder.Directive {
			return bytesDecoderByLen(l, l).Do(decoder.BytesToStr(int(l), bytesToIPStr).AsT("mplsTopLabelStackSection"))
			// not happy with the decoding here though, the bytes are meaning!
			// was causing an issue with messae 1197 of PCAP
			//return bytesDecoderByLen(l, math.MaxUint16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("mplsTopLabelStackSection"))
		},
		89: func(l uint16) decoder.Directive { return uintDecoderByLen(l).Do(decoder.AsT("forwardingStatus")) },
		234: func(l uint16) decoder.Directive {
			return uintDecoderByLen(l).Do(decoder.AsT("ingressVRFID"))
		},
		235: func(l uint16) decoder.Directive {
			return uintDecoderByLen(l).Do(decoder.AsT("egressVRFID"))
		},
	}
	//templateMap = make(map[uint16]*templateDefn)
	obsDomains = make(map[uint32]*obsDomain)
}

func getFieldDecoder(ft uint16, fl uint16) decoder.Directive {
	ldd := fieldDecoderMap[ft]
	if ldd != nil {
		return ldd(fl)
	}
	return decoder.Bytes(int(fl))
}
