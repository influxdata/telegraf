package sflow

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type ethHeaderDecoder struct {
	fieldName    string
	lenFieldName string
}

func (d *ethHeaderDecoder) Decode(r io.Reader, rec Recorder) error {

	lenVal, ok := rec.lookup(d.lenFieldName)
	if !ok {
		return fmt.Errorf("Unable to find ethHeader length field %s", d.lenFieldName)
	}

	lenValUint32, ok := lenVal.(uint32)
	if !ok {
		return fmt.Errorf("can't convert to int %T", lenValUint32)
	}

	nest := rec.nest(d.fieldName, 1)
	rec, ok = nest.next()
	if !ok {
		return fmt.Errorf("unable to nest 1")
	}

	lengthInt := int(lenValUint32)
	data := make([]byte, lengthInt)
	lengthRead, err := r.Read(data)
	if err != nil {
		return err
	}
	if lengthRead != lengthInt {
		return fmt.Errorf("did need read all of header length %d of %d", lengthRead, lengthInt)
	}

	etherType := data[12:14]

	var dataTransport []byte
	var nextHeader byte
	var tos byte
	var ttl byte
	var tcpflags byte
	srcIP := net.IP{}
	dstIP := net.IP{}
	offset := 14

	var srcMac uint64
	var dstMac uint64

	var identification uint16
	var fragOffset uint16

	dstMac = binary.BigEndian.Uint64(append([]byte{0, 0}, data[0:6]...))
	srcMac = binary.BigEndian.Uint64(append([]byte{0, 0}, data[6:12]...))
	rec.record("srcMac", srcMac)
	rec.record("dstMac", dstMac)

	if etherType[0] == 0x81 && etherType[1] == 0x0 { // VLAN 802.1Q
		rec.record("vlanId", uint32(binary.BigEndian.Uint16(data[14:16])))
		offset += 4
		etherType = data[16:18]
	}

	rec.record("etype", uint32(binary.BigEndian.Uint16(etherType[0:2])))

	if etherType[0] == 0x8 && etherType[1] == 0x0 { // IPv4
		rec.record("IPversion", 1) // v4?

		if len(data) >= offset+36 {
			nextHeader = data[offset+9]
			srcIP = data[offset+12 : offset+16]
			dstIP = data[offset+16 : offset+20]
			dataTransport = data[offset+20 : len(data)]
			tos = data[offset+1]
			ttl = data[offset+8]

			identification = binary.BigEndian.Uint16(data[offset+4 : offset+6])
			fragOffset = binary.BigEndian.Uint16(data[offset+6 : offset+8])
		}
	} else if etherType[0] == 0x86 && etherType[1] == 0xdd { // IPv6
		rec.record("IPversion", 2) // v6?
		if len(data) >= offset+40 {
			nextHeader = data[offset+6]
			srcIP = data[offset+8 : offset+24]
			dstIP = data[offset+24 : offset+40]
			dataTransport = data[offset+40 : len(data)]

			tostmp := uint32(binary.BigEndian.Uint16(data[offset : offset+2]))
			tos = uint8(tostmp & 0x0ff0 >> 4)
			ttl = data[offset+7]

			flowLabeltmp := binary.BigEndian.Uint32(data[offset : offset+4])
			rec.record("IPv6FlowLabel", flowLabeltmp&0xFFFFF)
		}
	} else if etherType[0] == 0x8 && etherType[1] == 0x6 { // ARP
	} else {
		return fmt.Errorf("Unknown EtherType: %v", etherType)
	}

	if len(dataTransport) >= 4 && (nextHeader == 17 || nextHeader == 6) {
		//fmt.Println("Recording srcPort and dstPort")
		rec.record("srcPort", uint32(binary.BigEndian.Uint16(dataTransport[0:2])))
		rec.record("dstPort", uint32(binary.BigEndian.Uint16(dataTransport[2:4])))
	} else {
		//fmt.Println("NOT recording srcPort and dstPort ", len(dataTransport), nextHeader)
	}

	//if nextHeader == 6 {
	// get urgent pointer
	//	urgentPointer = dataTransport[18]
	//}

	if len(dataTransport) >= 13 && nextHeader == 6 {
		tcpflags = dataTransport[13]
	}

	// ICMP and ICMPv6
	if len(dataTransport) >= 2 && (nextHeader == 1 || nextHeader == 58) {
		rec.record("IcmpType", uint32(dataTransport[0]))
		rec.record("IcmpCode", uint32(dataTransport[1]))
	}

	rec.record("srcIP", srcIP)
	rec.record("dstIP", dstIP)
	rec.record("proto", uint32(nextHeader))
	rec.record("IPTos", uint32(tos))
	rec.record("IPTTL", uint32(ttl))
	rec.record("TCPFlags", uint32(tcpflags))
	rec.record("fragmentId", uint32(identification))
	rec.record("fragmentOffset", uint32(fragOffset))

	return nil
}

func ethHeader(fieldlName string, lenFieldName string) ItemDecoder {
	return &ethHeaderDecoder{fieldlName, lenFieldName}
}
