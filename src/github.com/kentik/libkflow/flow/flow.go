package flow

// #include "../kflow.h"
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/kentik/libkflow/chf"
)

// Ckflow is an alias for C.kflow used because tests run by
// `go test` cannot reference the "C" package.
type Ckflow C.kflow

const MAX_CUSTOM_STR_LEN = 384

type Flow struct {
	TimestampNano      int64
	DstAs              uint32
	DstGeo             uint32
	DstMac             uint32
	HeaderLen          uint32
	InBytes            uint64
	InPkts             uint64
	InputPort          uint32
	IpSize             uint32
	Ipv4DstAddr        uint32
	Ipv4SrcAddr        uint32
	L4DstPort          uint32
	L4SrcPort          uint32
	OutputPort         uint32
	Protocol           uint32
	SampledPacketSize  uint32
	SrcAs              uint32
	SrcGeo             uint32
	SrcMac             uint32
	TcpFlags           uint32
	Tos                uint32
	VlanIn             uint32
	VlanOut            uint32
	Ipv4NextHop        uint32
	MplsType           uint32
	OutBytes           uint64
	OutPkts            uint64
	TcpRetransmit      uint32
	SrcFlowTags        string
	DstFlowTags        string
	SampleRate         uint32
	DeviceId           uint32
	FlowTags           string
	Timestamp          int64
	DstBgpAsPath       string
	DstBgpCommunity    string
	SrcBgpAsPath       string
	SrcBgpCommunity    string
	SrcNextHopAs       uint32
	DstNextHopAs       uint32
	SrcGeoRegion       uint32
	DstGeoRegion       uint32
	SrcGeoCity         uint32
	DstGeoCity         uint32
	Big                bool
	SampleAdj          bool
	Ipv4DstNextHop     uint32
	Ipv4SrcNextHop     uint32
	SrcRoutePrefix     uint32
	DstRoutePrefix     uint32
	SrcRouteLength     uint8
	DstRouteLength     uint8
	SrcSecondAsn       uint32
	DstSecondAsn       uint32
	SrcThirdAsn        uint32
	DstThirdAsn        uint32
	Ipv6DstAddr        []byte
	Ipv6SrcAddr        []byte
	SrcEthMac          uint64
	DstEthMac          uint64
	Ipv6SrcNextHop     []byte
	Ipv6DstNextHop     []byte
	Ipv6SrcRoutePrefix []byte
	Ipv6DstRoutePrefix []byte
	Customs            []Custom
}

type Type int

const (
	Str Type = iota
	U32
	F32
)

type Custom struct {
	ID   uint32
	Type Type
	Str  string
	U32  uint32
	F32  float32
}

func New(cflow *Ckflow) Flow {
	return Flow{
		TimestampNano:      int64(cflow.timestampNano),
		DstAs:              uint32(cflow.dstAs),
		DstGeo:             uint32(cflow.dstGeo),
		DstMac:             uint32(cflow.dstMac),
		HeaderLen:          uint32(cflow.headerLen),
		InBytes:            uint64(cflow.inBytes),
		InPkts:             uint64(cflow.inPkts),
		InputPort:          uint32(cflow.inputPort),
		IpSize:             uint32(cflow.ipSize),
		Ipv4DstAddr:        uint32(cflow.ipv4DstAddr),
		Ipv4SrcAddr:        uint32(cflow.ipv4SrcAddr),
		L4DstPort:          uint32(cflow.l4DstPort),
		L4SrcPort:          uint32(cflow.l4SrcPort),
		OutputPort:         uint32(cflow.outputPort),
		Protocol:           uint32(cflow.protocol),
		SampledPacketSize:  uint32(cflow.sampledPacketSize),
		SrcAs:              uint32(cflow.srcAs),
		SrcGeo:             uint32(cflow.srcGeo),
		SrcMac:             uint32(cflow.srcMac),
		TcpFlags:           uint32(cflow.tcpFlags),
		Tos:                uint32(cflow.tos),
		VlanIn:             uint32(cflow.vlanIn),
		VlanOut:            uint32(cflow.vlanOut),
		Ipv4NextHop:        uint32(cflow.ipv4NextHop),
		MplsType:           uint32(cflow.mplsType),
		OutBytes:           uint64(cflow.outBytes),
		OutPkts:            uint64(cflow.outPkts),
		TcpRetransmit:      uint32(cflow.tcpRetransmit),
		SrcFlowTags:        C.GoString(cflow.srcFlowTags),
		DstFlowTags:        C.GoString(cflow.dstFlowTags),
		SampleRate:         uint32(cflow.sampleRate),
		DeviceId:           uint32(cflow.deviceId),
		FlowTags:           C.GoString(cflow.flowTags),
		Timestamp:          int64(cflow.timestamp),
		DstBgpAsPath:       C.GoString(cflow.dstBgpAsPath),
		DstBgpCommunity:    C.GoString(cflow.dstBgpCommunity),
		SrcBgpAsPath:       C.GoString(cflow.srcBgpAsPath),
		SrcBgpCommunity:    C.GoString(cflow.srcBgpCommunity),
		SrcNextHopAs:       uint32(cflow.srcNextHopAs),
		DstNextHopAs:       uint32(cflow.dstNextHopAs),
		SrcGeoRegion:       uint32(cflow.srcGeoRegion),
		DstGeoRegion:       uint32(cflow.dstGeoRegion),
		SrcGeoCity:         uint32(cflow.srcGeoCity),
		DstGeoCity:         uint32(cflow.dstGeoCity),
		Big:                cflow.big == 1,
		SampleAdj:          cflow.sampleAdj == 1,
		Ipv4DstNextHop:     uint32(cflow.ipv4DstNextHop),
		Ipv4SrcNextHop:     uint32(cflow.ipv4SrcNextHop),
		SrcRoutePrefix:     uint32(cflow.srcRoutePrefix),
		DstRoutePrefix:     uint32(cflow.dstRoutePrefix),
		SrcRouteLength:     uint8(cflow.srcRouteLength),
		DstRouteLength:     uint8(cflow.dstRouteLength),
		SrcSecondAsn:       uint32(cflow.srcSecondAsn),
		DstSecondAsn:       uint32(cflow.dstSecondAsn),
		SrcThirdAsn:        uint32(cflow.srcThirdAsn),
		DstThirdAsn:        uint32(cflow.dstThirdAsn),
		Ipv6DstAddr:        bts(cflow.ipv6DstAddr, 16),
		Ipv6SrcAddr:        bts(cflow.ipv6SrcAddr, 16),
		SrcEthMac:          uint64(cflow.srcEthMac),
		DstEthMac:          uint64(cflow.dstEthMac),
		Ipv6SrcNextHop:     bts(cflow.ipv6SrcNextHop, 16),
		Ipv6DstNextHop:     bts(cflow.ipv6DstNextHop, 16),
		Ipv6SrcRoutePrefix: bts(cflow.ipv6SrcRoutePrefix, 16),
		Ipv6DstRoutePrefix: bts(cflow.ipv6DstRoutePrefix, 16),
		Customs:            newCustoms(cflow),
	}
}

func (f *Flow) FillCHF(kflow chf.CHF, list chf.Custom_List) {
	kflow.SetTimestampNano(f.TimestampNano)
	kflow.SetDstAs(f.DstAs)
	kflow.SetDstGeo(f.DstGeo)
	kflow.SetDstMac(f.DstMac)
	kflow.SetHeaderLen(f.HeaderLen)
	kflow.SetInBytes(f.InBytes)
	kflow.SetInPkts(f.InPkts)
	kflow.SetInputPort(f.InputPort)
	kflow.SetIpSize(f.IpSize)
	kflow.SetIpv4DstAddr(f.Ipv4DstAddr)
	kflow.SetIpv4SrcAddr(f.Ipv4SrcAddr)
	kflow.SetL4DstPort(f.L4DstPort)
	kflow.SetL4SrcPort(f.L4SrcPort)
	kflow.SetOutputPort(f.OutputPort)
	kflow.SetProtocol(f.Protocol)
	kflow.SetSampledPacketSize(f.SampledPacketSize)
	kflow.SetSrcAs(f.SrcAs)
	kflow.SetSrcGeo(f.SrcGeo)
	kflow.SetSrcMac(f.SrcMac)
	kflow.SetTcpFlags(f.TcpFlags)
	kflow.SetTos(f.Tos)
	kflow.SetVlanIn(f.VlanIn)
	kflow.SetVlanOut(f.VlanOut)
	kflow.SetIpv4NextHop(f.Ipv4NextHop)
	kflow.SetMplsType(f.MplsType)
	kflow.SetOutBytes(f.OutBytes)
	kflow.SetOutPkts(f.OutPkts)
	kflow.SetTcpRetransmit(f.TcpRetransmit)
	kflow.SetSrcFlowTags(f.SrcFlowTags)
	kflow.SetDstFlowTags(f.DstFlowTags)
	kflow.SetSampleRate(f.SampleRate)
	kflow.SetDeviceId(f.DeviceId)
	kflow.SetFlowTags(f.FlowTags)
	kflow.SetTimestamp(f.Timestamp)
	kflow.SetDstBgpAsPath(f.DstBgpAsPath)
	kflow.SetDstBgpCommunity(f.DstBgpCommunity)
	kflow.SetSrcBgpAsPath(f.SrcBgpAsPath)
	kflow.SetSrcBgpCommunity(f.SrcBgpCommunity)
	kflow.SetSrcNextHopAs(f.SrcNextHopAs)
	kflow.SetDstNextHopAs(f.DstNextHopAs)
	kflow.SetSrcGeoRegion(f.SrcGeoRegion)
	kflow.SetDstGeoRegion(f.DstGeoRegion)
	kflow.SetSrcGeoCity(f.SrcGeoCity)
	kflow.SetDstGeoCity(f.DstGeoCity)
	kflow.SetBig(f.Big)
	kflow.SetSampleAdj(f.SampleAdj)
	kflow.SetIpv4DstNextHop(f.Ipv4DstNextHop)
	kflow.SetIpv4SrcNextHop(f.Ipv4SrcNextHop)
	kflow.SetSrcRoutePrefix(f.SrcRoutePrefix)
	kflow.SetDstRoutePrefix(f.DstRoutePrefix)
	kflow.SetSrcRouteLength(f.SrcRouteLength)
	kflow.SetDstRouteLength(f.DstRouteLength)
	kflow.SetSrcSecondAsn(f.SrcSecondAsn)
	kflow.SetDstSecondAsn(f.DstSecondAsn)
	kflow.SetSrcThirdAsn(f.SrcThirdAsn)
	kflow.SetDstThirdAsn(f.DstThirdAsn)
	kflow.SetIpv6DstAddr(f.Ipv6DstAddr)
	kflow.SetIpv6SrcAddr(f.Ipv6SrcAddr)
	kflow.SetSrcEthMac(f.SrcEthMac)
	kflow.SetDstEthMac(f.DstEthMac)
	kflow.SetIpv6SrcNextHop(f.Ipv6SrcNextHop)
	kflow.SetIpv6DstNextHop(f.Ipv6DstNextHop)
	kflow.SetIpv6SrcRoutePrefix(f.Ipv6SrcRoutePrefix)
	kflow.SetIpv6DstRoutePrefix(f.Ipv6DstRoutePrefix)

	for i, c := range f.Customs {
		kc := list.At(i)
		kc.SetId(uint32(c.ID))

		switch c.Type {
		case Str:
			kc.Value().SetStrVal(c.Str)
		case U32:
			kc.Value().SetUint32Val(c.U32)
		case F32:
			kc.Value().SetFloat32Val(c.F32)
		}
	}

	kflow.SetCustom(list)
}

func newCustoms(cflow *Ckflow) []Custom {
	cslice := *(*[]C.kflowCustom)(unsafe.Pointer(&reflect.SliceHeader{
		Data: (uintptr)(unsafe.Pointer(cflow.customs)),
		Len:  int(cflow.numCustoms),
		Cap:  int(cflow.numCustoms),
	}))

	customs := make([]Custom, len(cslice))

	for i, ccc := range cslice {
		custom := &customs[i]
		custom.ID = uint32(ccc.id)

		p := unsafe.Pointer(&ccc.value[0])
		switch ccc.vtype {
		case C.KFLOWCUSTOMSTR:
			custom.Type = Str
			custom.Str = trunc(C.GoString(*(**C.char)(p)))
		case C.KFLOWCUSTOMU32:
			custom.Type = U32
			custom.U32 = uint32(*(*C.uint32_t)(p))
		case C.KFLOWCUSTOMF32:
			custom.Type = F32
			custom.F32 = float32(*(*C.float)(p))
		}
	}

	return customs
}

func bts(p *C.uint8_t, len C.int) []byte {
	if p == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(p), len)
}

func trunc(s string) string {
	if n := MAX_CUSTOM_STR_LEN; len(s) > n {
		return s[:n]
	}
	return s
}
