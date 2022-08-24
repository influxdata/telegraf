package sflow_a10

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/sflow/binaryio"
	"github.com/pkg/errors"

	hm "github.com/cornelk/hashmap"
)

type PacketDecoder struct {
	onPacket         func(p *V5Format)
	Log              telegraf.Logger
	CounterBlocks    map[uint32]CounterBlock
	IgnoreZeroValues bool

	IPMap   *hm.HashMap
	PortMap *hm.HashMap
}

func NewDecoder() *PacketDecoder {
	return &PacketDecoder{
		IPMap:            &hm.HashMap{},
		PortMap:          &hm.HashMap{},
		CounterBlocks:    make(map[uint32]CounterBlock),
		IgnoreZeroValues: true,
	}
}

func (d *PacketDecoder) debug(args ...interface{}) {
	if d.Log != nil {
		d.Log.Debug(args...)
	}
}

func (d *PacketDecoder) OnPacket(f func(p *V5Format)) {
	d.onPacket = f
}

func (d *PacketDecoder) Decode(r io.Reader) error {
	var err error
	var packet *V5Format
	for err == nil {
		packet, err = d.DecodeOnePacket(r)
		if err != nil {
			break
		}
		d.onPacket(packet)
	}
	if err != nil && errors.Cause(err) == io.EOF {
		return nil
	}
	return err
}

type AddressType uint32

const (
	AddressTypeUnknown AddressType = 0
	AddressTypeIPV4    AddressType = 1
	AddressTypeIPV6    AddressType = 2
)

func (d *PacketDecoder) DecodeOnePacket(r io.Reader) (*V5Format, error) {
	p := &V5Format{}
	err := read(r, &p.Version, "version")
	if err != nil {
		return nil, err
	}
	if p.Version != 5 {
		return nil, fmt.Errorf("version %d not supported, only version 5", p.Version)
	}
	var addressIPType AddressType
	if err = read(r, &addressIPType, "address ip type"); err != nil {
		return nil, err
	}
	switch addressIPType {
	case AddressTypeUnknown:
		p.AgentAddress.IP = make([]byte, 0)
	case AddressTypeIPV4:
		p.AgentAddress.IP = make([]byte, 4)
	case AddressTypeIPV6:
		p.AgentAddress.IP = make([]byte, 16)
	default:
		return nil, fmt.Errorf("unknown address IP type %d", addressIPType)
	}
	if err = read(r, &p.AgentAddress.IP, "Agent Address IP"); err != nil {
		return nil, err
	}
	if err = read(r, &p.SubAgentID, "SubAgentID"); err != nil {
		return nil, err
	}
	if err = read(r, &p.SequenceNumber, "SequenceNumber"); err != nil {
		return nil, err
	}
	if err = read(r, &p.Uptime, "Uptime"); err != nil {
		return nil, err
	}

	p.Samples, err = d.decodeSamples(r, p.AgentAddress.String())
	return p, err
}

func (d *PacketDecoder) decodeSamples(r io.Reader, agentAddress string) ([]Sample, error) {
	result := []Sample{}

	var numOfSamples uint32
	if err := read(r, &numOfSamples, "sample count"); err != nil {
		return nil, err
	}

	for i := 0; i < int(numOfSamples); i++ {
		sam, err := d.decodeSample(r, agentAddress)
		if err != nil {
			return result, err
		}
		result = append(result, *sam)
	}

	return result, nil
}

func (d *PacketDecoder) decodeSample(r io.Reader, agentAddress string) (*Sample, error) {
	var err error
	sam := &Sample{}
	if err := read(r, &sam.SampleType, "SampleType"); err != nil {
		return sam, err
	}
	sampleDataLen := uint32(0)
	if err := read(r, &sampleDataLen, "Sample data length"); err != nil {
		return sam, err
	}

	mr := binaryio.MinReader(r, int64(sampleDataLen))
	defer mr.Close()

	// we're only interested in counter samples
	switch sam.SampleType {
	case SampleTypeCounter:
		sam.SampleCounterData, err = d.decodeCounterSample(mr, agentAddress)
	default:
		d.debug("Unknown sample type: ", sam.SampleType)
	}
	return sam, err
}

func (d *PacketDecoder) decodeCounterSample(r io.Reader, agentAddress string) (*CounterSample, error) {
	s := &CounterSample{}
	if err := read(r, &s.SequenceNumber, "SequenceNumber"); err != nil { // sflow_version_5.txt line: 1661
		return s, err
	}
	if err := read(r, &s.SourceID, "SourceID"); err != nil { // sflow_version_5.txt line: 1671
		return s, err
	}

	// as per A10 documentation we should use the least significant 28 bits for "entry mapping info"
	s.SourceID = s.SourceID & 0xFFFFFFF

	var err error
	s.CounterRecords, err = d.decodeCounterRecords(r, s.SourceID, agentAddress)
	return s, err
}

func (d *PacketDecoder) decodeCounterRecords(r io.Reader, sourceID uint32, agentAddress string) ([]CounterRecord, error) {
	var count uint32
	if err := read(r, &count, "CounterRecord count"); err != nil {
		return nil, err
	}
	var recs []CounterRecord
	for i := uint32(0); i < count; i++ {
		cr := &CounterRecord{}
		if err := read(r, &cr.CounterFormat, "CounterFormat"); err != nil {
			return recs, err
		}
		var counterDataLen uint32
		if err := read(r, &counterDataLen, "Counter data length"); err != nil {
			return recs, err
		}

		mr := binaryio.MinReader(r, int64(counterDataLen))

		tagCF := cr.CounterFormat & 0xFFF // the least significant 12 bits, sflow_version_5.txt line: 1410
		tag := uint32(tagCF)

		if tag == 260 { // hex 104 - contains port information
			portDimensions, err := d.decode260(r)
			if err != nil {
				return recs, err
			}

			key := createMapKey(sourceID, agentAddress)
			_, ok := d.PortMap.Get(key)

			if !ok {
				d.PortMap.Set(key, portDimensions)
			}

			continue
		} else if tag == 271 { // hex 10F - contains IPv4 information
			ipDimensions, err := d.decode271(r)
			if err != nil {
				return recs, err
			}

			key := createMapKey(sourceID, agentAddress)
			_, ok := d.IPMap.Get(key)
			if !ok {
				d.IPMap.Set(key, ipDimensions)
			}

			continue
		} else if tag == 272 { // hex 110 - contains IPv6 information
			ipDimensions, err := d.decode272(r)
			if err != nil {
				return recs, err
			}

			key := createMapKey(sourceID, agentAddress)
			_, ok := d.IPMap.Get(key)
			if !ok {
				d.IPMap.Set(key, ipDimensions)
			}

			continue
		}

		// we're checking if the tag we got exists on our counter record definitions (that were loaded from the A10 xml file)
		if _, exists := d.CounterBlocks[tag]; !exists {
			d.debug(fmt.Sprintf("  tag %x for sourceID %x NOT found on xml file list. Ignoring counter record", tag, sourceID))
			continue
		}

		// as per A10, each packet of either counter block 293 or 294 is one sample of 293 or 294
		// plus, we are not getting any IP and PORT information
		cr.NeedsIPAndPort = tag != 293 && tag != 294

		cr.IsEthernetCounters = tag == 294

		d.debug(fmt.Sprintf("  tag %x for sourceID %x needs ip and and port: %t", tag, sourceID, cr.NeedsIPAndPort))
		d.debug(fmt.Sprintf("  tag %x for sourceID %x found on xml file list. Gonna decode counter record", tag, sourceID))

		err := d.decodeCounterRecord(mr, cr, tag, sourceID)
		if err != nil {
			return recs, err
		}

		recs = append(recs, *cr)
		mr.Close()
	}

	return recs, nil
}

func (d *PacketDecoder) decodeCounterRecord(r io.Reader, cr *CounterRecord, tag uint32, sourceID uint32) error {
	counterBlock := d.CounterBlocks[tag]

	// reading the header values, we assume they are always 4 uint16
	var counterOffset uint16
	var totalCounterNum uint16
	if len(counterBlock.OffsetHeaders) > 0 {
		if err := read(r, &counterOffset, "counterOffset"); err != nil {
			d.debug("    error reading counterOffset variable")
			return err
		}
		d.debug(fmt.Sprintf("    header counterOffset has value hex %#v dec %d", counterOffset, counterOffset))

		if err := read(r, &totalCounterNum, "totalCounterNum"); err != nil {
			d.debug("    error reading totalCounterNum variable")
			return err
		}
		d.debug(fmt.Sprintf("    header totalCounterNum has value hex %#v dec %d", totalCounterNum, totalCounterNum))

		// read the two reserved uint16 values
		for i := 0; i < 2; i++ {
			var headerTemp uint16
			if err := read(r, &headerTemp, "HEADERTEMP"); err != nil {
				d.debug(fmt.Sprintf("    error reading header variable: %s", err))
				return err
			}
		}
	}

	cr.CounterData = &CounterData{
		CounterFields: make(map[string]interface{}),
	}

	//d.debug(fmt.Sprintf("offsetInfo %d, len %d, total %d, sourceID %d", counterOffset, len(counterBlock.Counters), totalCounterNum, sourceID))

	// reading all counters
	for i := int(counterOffset); i < len(counterBlock.Counters); i++ {
		counter := counterBlock.Counters[i]

		var counterValue uint64
		var err error
		if counterValue, err = readBinary(r, counter.Dtype, counter.FieldName); err != nil {
			d.debug(fmt.Sprintf("    error reading counter variable %s with error %s", counter.FieldName, err))
			continue
		}

		if counterValue != uint64(0) || (counterValue == uint64(0) && !d.IgnoreZeroValues) {
			//d.debug(fmt.Sprintf("    getting non-zero counter %s with value hex %x %#v %T for sourceID %x", counter.FieldName, counterValue, counterValue, counterValue, sourceID))
			cr.CounterData.CounterFields[counter.FieldName] = counterValue
		}
	}
	return nil
}

func readBinary(r io.Reader, bitsNum string, name string) (uint64, error) {
	var temp8 uint8
	var temp16 uint16
	var temp32 uint32
	var temp64 uint64
	var err error
	if bitsNum == "u8" {
		err = read(r, &temp8, name)
		if err != nil {
			return 0, err
		}
		return uint64(temp8), nil
	} else if bitsNum == "u16" {
		err = read(r, &temp16, name)
		if err != nil {
			return 0, err
		}
		return uint64(temp16), nil
	} else if bitsNum == "u32" {
		err = read(r, &temp32, name)
		if err != nil {
			return 0, err
		}
		return uint64(temp32), nil
	} else if bitsNum == "u64" {
		err = read(r, &temp64, name)
		if err != nil {
			return 0, err
		}
		return temp64, nil
	} else if bitsNum == "string" {
		// TODO: skipping for the time being
		return 0, nil
	} else {
		return 0, fmt.Errorf("uknown bitsNum")
	}
}

func (d *PacketDecoder) decode260(r io.Reader) (*PortDimension, error) {
	var tableType uint8
	var portType uint8
	var portNum uint16
	var entryName [64]byte
	var portRangeEnd uint16

	if err := read(r, &tableType, "TableType"); err != nil {
		return nil, err
	}
	if err := read(r, &portType, "PortType"); err != nil {
		return nil, err
	}
	if err := read(r, &portNum, "portNum"); err != nil {
		return nil, err
	}
	if err := read(r, &entryName, "entryName"); err != nil {
		return nil, err
	}
	if err := read(r, &portRangeEnd, "PortRangeEnd"); err != nil {
		return nil, err
	}

	d.debug(fmt.Sprintf("Read 260 with TableType %d, PortType %d, portNum %d, PortRangeEnd %d", tableType, portType, portNum, portRangeEnd))

	portDimensions := &PortDimension{
		TableType:    tableTypeIntToString(tableType),
		PortType:     portTypeIntToString(portType),
		PortNumber:   int(portNum),
		PortRangeEnd: int(portRangeEnd),
	}

	return portDimensions, nil
}

func (d *PacketDecoder) decode271(r io.Reader) ([]IPDimension, error) {
	var startAddressOffset uint8
	var addressCount uint8
	var totalAddressCount uint16
	var reserved uint32

	if err := read(r, &startAddressOffset, "startAddressOffset"); err != nil {
		return nil, err
	}
	if err := read(r, &addressCount, "addressCount"); err != nil {
		return nil, err
	}
	if err := read(r, &totalAddressCount, "totalAddressCount"); err != nil {
		return nil, err
	}
	if err := read(r, &reserved, "reserved"); err != nil {
		return nil, err
	}

	d.debug(fmt.Sprintf("Read 271 with startAddressOffset %x, addressCount %x, totalAddressCount %x", startAddressOffset, addressCount, totalAddressCount))

	ipDimensions := make([]IPDimension, addressCount)
	for i := 0; i < int(addressCount); i++ {
		var ip32 [4]byte
		var subnetMask uint8
		if err := read(r, &ip32, "ip32"); err != nil {
			continue
		}
		if err := read(r, &subnetMask, "SubnetMask"); err != nil {
			continue
		}
		d.debug(fmt.Sprintf("Read 271 ip with ip32 %v and SubnetMask %d", ip32, subnetMask))

		ipDimensions[i] = IPDimension{
			IPAddress:  fmt.Sprintf("%d.%d.%d.%d", int(ip32[0]), int(ip32[1]), int(ip32[2]), int(ip32[3])),
			SubnetMask: subnetMask,
		}
	}
	return ipDimensions, nil
}

func (d *PacketDecoder) decode272(r io.Reader) ([]IPDimension, error) {
	var startAddressOffset uint8
	var addressCount uint8
	var totalAddressCount uint16
	var reserved uint32

	if err := read(r, &startAddressOffset, "startAddressOffset"); err != nil {
		return nil, err
	}
	if err := read(r, &addressCount, "addressCount"); err != nil {
		return nil, err
	}
	if err := read(r, &totalAddressCount, "totalAddressCount"); err != nil {
		return nil, err
	}
	if err := read(r, &reserved, "reserved"); err != nil {
		return nil, err
	}
	d.debug(fmt.Sprintf("Read 272 with startAddressOffset %d, addressCount %d, totalAddressCount %d", startAddressOffset, addressCount, totalAddressCount))

	ipDimensions := make([]IPDimension, addressCount)
	for i := 0; i < int(addressCount); i++ {
		var ip64 [16]byte
		var subnetMask uint8
		if err := read(r, &ip64, "ip64"); err != nil {
			continue
		}
		if err := read(r, &subnetMask, "SubnetMask"); err != nil {
			continue
		}
		d.debug(fmt.Sprintf("Read 272 ip with ip64 %#v and SubnetMask %d", ip64, subnetMask))

		ipDimensions[i] = IPDimension{
			IPAddress:  fullIPv6(ip64),
			SubnetMask: subnetMask,
		}
	}

	return ipDimensions, nil
}

func fullIPv6(ip64 [16]byte) string {
	i := 0
	length := 16
	var s string
	for {
		if i == length {
			break
		}

		if i%2 == 0 || i == length-1 {
			s = fmt.Sprintf("%s%x", s, ip64[i])
		} else {
			s = fmt.Sprintf("%s%x:", s, ip64[i])
		}
		i++
	}
	return s
}

func read(r io.Reader, data interface{}, name string) error {
	err := binary.Read(r, binary.BigEndian, data)
	return errors.Wrapf(err, "failed to read %s", name)
}

func createMapKey(sourceID uint32, addr string) string {
	return fmt.Sprintf("%s_%x", addr, sourceID)
}
