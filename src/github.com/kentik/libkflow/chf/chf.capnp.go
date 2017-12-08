package chf

// AUTO GENERATED - DO NOT EDIT

import (
	math "math"
	strconv "strconv"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Custom struct{ capnp.Struct }
type Custom_value Custom
type Custom_value_Which uint16

const (
	Custom_value_Which_uint32Val  Custom_value_Which = 0
	Custom_value_Which_float32Val Custom_value_Which = 1
	Custom_value_Which_strVal     Custom_value_Which = 2
)

func (w Custom_value_Which) String() string {
	const s = "uint32Valfloat32ValstrVal"
	switch w {
	case Custom_value_Which_uint32Val:
		return s[0:9]
	case Custom_value_Which_float32Val:
		return s[9:19]
	case Custom_value_Which_strVal:
		return s[19:25]

	}
	return "Custom_value_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Custom_TypeID is the unique identifier for the type Custom.
const Custom_TypeID = 0xed5d37861203d027

func NewCustom(s *capnp.Segment) (Custom, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1})
	return Custom{st}, err
}

func NewRootCustom(s *capnp.Segment) (Custom, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1})
	return Custom{st}, err
}

func ReadRootCustom(msg *capnp.Message) (Custom, error) {
	root, err := msg.RootPtr()
	return Custom{root.Struct()}, err
}

func (s Custom) String() string {
	str, _ := text.Marshal(0xed5d37861203d027, s.Struct)
	return str
}

func (s Custom) Id() uint32 {
	return s.Struct.Uint32(0)
}

func (s Custom) SetId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Custom) Value() Custom_value { return Custom_value(s) }

func (s Custom_value) Which() Custom_value_Which {
	return Custom_value_Which(s.Struct.Uint16(8))
}
func (s Custom_value) Uint32Val() uint32 {
	return s.Struct.Uint32(4)
}

func (s Custom_value) SetUint32Val(v uint32) {
	s.Struct.SetUint16(8, 0)
	s.Struct.SetUint32(4, v)
}

func (s Custom_value) Float32Val() float32 {
	return math.Float32frombits(s.Struct.Uint32(4))
}

func (s Custom_value) SetFloat32Val(v float32) {
	s.Struct.SetUint16(8, 1)
	s.Struct.SetUint32(4, math.Float32bits(v))
}

func (s Custom_value) StrVal() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Custom_value) HasStrVal() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Custom_value) StrValBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Custom_value) SetStrVal(v string) error {
	s.Struct.SetUint16(8, 2)
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Custom) IsDimension() bool {
	return s.Struct.Bit(80)
}

func (s Custom) SetIsDimension(v bool) {
	s.Struct.SetBit(80, v)
}

// Custom_List is a list of Custom.
type Custom_List struct{ capnp.List }

// NewCustom creates a new list of Custom.
func NewCustom_List(s *capnp.Segment, sz int32) (Custom_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1}, sz)
	return Custom_List{l}, err
}

func (s Custom_List) At(i int) Custom { return Custom{s.List.Struct(i)} }

func (s Custom_List) Set(i int, v Custom) error { return s.List.SetStruct(i, v.Struct) }

// Custom_Promise is a wrapper for a Custom promised by a client call.
type Custom_Promise struct{ *capnp.Pipeline }

func (p Custom_Promise) Struct() (Custom, error) {
	s, err := p.Pipeline.Struct()
	return Custom{s}, err
}

func (p Custom_Promise) Value() Custom_value_Promise { return Custom_value_Promise{p.Pipeline} }

// Custom_value_Promise is a wrapper for a Custom_value promised by a client call.
type Custom_value_Promise struct{ *capnp.Pipeline }

func (p Custom_value_Promise) Struct() (Custom_value, error) {
	s, err := p.Pipeline.Struct()
	return Custom_value{s}, err
}

type CHF struct{ capnp.Struct }

// CHF_TypeID is the unique identifier for the type CHF.
const CHF_TypeID = 0xa7ab5c68e4bc7b62

func NewCHF(s *capnp.Segment) (CHF, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 224, PointerCount: 14})
	return CHF{st}, err
}

func NewRootCHF(s *capnp.Segment) (CHF, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 224, PointerCount: 14})
	return CHF{st}, err
}

func ReadRootCHF(msg *capnp.Message) (CHF, error) {
	root, err := msg.RootPtr()
	return CHF{root.Struct()}, err
}

func (s CHF) String() string {
	str, _ := text.Marshal(0xa7ab5c68e4bc7b62, s.Struct)
	return str
}

func (s CHF) TimestampNano() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s CHF) SetTimestampNano(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

func (s CHF) DstAs() uint32 {
	return s.Struct.Uint32(8)
}

func (s CHF) SetDstAs(v uint32) {
	s.Struct.SetUint32(8, v)
}

func (s CHF) DstGeo() uint32 {
	return s.Struct.Uint32(12)
}

func (s CHF) SetDstGeo(v uint32) {
	s.Struct.SetUint32(12, v)
}

func (s CHF) DstMac() uint32 {
	return s.Struct.Uint32(16)
}

func (s CHF) SetDstMac(v uint32) {
	s.Struct.SetUint32(16, v)
}

func (s CHF) HeaderLen() uint32 {
	return s.Struct.Uint32(20)
}

func (s CHF) SetHeaderLen(v uint32) {
	s.Struct.SetUint32(20, v)
}

func (s CHF) InBytes() uint64 {
	return s.Struct.Uint64(24)
}

func (s CHF) SetInBytes(v uint64) {
	s.Struct.SetUint64(24, v)
}

func (s CHF) InPkts() uint64 {
	return s.Struct.Uint64(32)
}

func (s CHF) SetInPkts(v uint64) {
	s.Struct.SetUint64(32, v)
}

func (s CHF) InputPort() uint32 {
	return s.Struct.Uint32(40)
}

func (s CHF) SetInputPort(v uint32) {
	s.Struct.SetUint32(40, v)
}

func (s CHF) IpSize() uint32 {
	return s.Struct.Uint32(44)
}

func (s CHF) SetIpSize(v uint32) {
	s.Struct.SetUint32(44, v)
}

func (s CHF) Ipv4DstAddr() uint32 {
	return s.Struct.Uint32(48)
}

func (s CHF) SetIpv4DstAddr(v uint32) {
	s.Struct.SetUint32(48, v)
}

func (s CHF) Ipv4SrcAddr() uint32 {
	return s.Struct.Uint32(52)
}

func (s CHF) SetIpv4SrcAddr(v uint32) {
	s.Struct.SetUint32(52, v)
}

func (s CHF) L4DstPort() uint32 {
	return s.Struct.Uint32(56)
}

func (s CHF) SetL4DstPort(v uint32) {
	s.Struct.SetUint32(56, v)
}

func (s CHF) L4SrcPort() uint32 {
	return s.Struct.Uint32(60)
}

func (s CHF) SetL4SrcPort(v uint32) {
	s.Struct.SetUint32(60, v)
}

func (s CHF) OutputPort() uint32 {
	return s.Struct.Uint32(64)
}

func (s CHF) SetOutputPort(v uint32) {
	s.Struct.SetUint32(64, v)
}

func (s CHF) Protocol() uint32 {
	return s.Struct.Uint32(68)
}

func (s CHF) SetProtocol(v uint32) {
	s.Struct.SetUint32(68, v)
}

func (s CHF) SampledPacketSize() uint32 {
	return s.Struct.Uint32(72)
}

func (s CHF) SetSampledPacketSize(v uint32) {
	s.Struct.SetUint32(72, v)
}

func (s CHF) SrcAs() uint32 {
	return s.Struct.Uint32(76)
}

func (s CHF) SetSrcAs(v uint32) {
	s.Struct.SetUint32(76, v)
}

func (s CHF) SrcGeo() uint32 {
	return s.Struct.Uint32(80)
}

func (s CHF) SetSrcGeo(v uint32) {
	s.Struct.SetUint32(80, v)
}

func (s CHF) SrcMac() uint32 {
	return s.Struct.Uint32(84)
}

func (s CHF) SetSrcMac(v uint32) {
	s.Struct.SetUint32(84, v)
}

func (s CHF) TcpFlags() uint32 {
	return s.Struct.Uint32(88)
}

func (s CHF) SetTcpFlags(v uint32) {
	s.Struct.SetUint32(88, v)
}

func (s CHF) Tos() uint32 {
	return s.Struct.Uint32(92)
}

func (s CHF) SetTos(v uint32) {
	s.Struct.SetUint32(92, v)
}

func (s CHF) VlanIn() uint32 {
	return s.Struct.Uint32(96)
}

func (s CHF) SetVlanIn(v uint32) {
	s.Struct.SetUint32(96, v)
}

func (s CHF) VlanOut() uint32 {
	return s.Struct.Uint32(100)
}

func (s CHF) SetVlanOut(v uint32) {
	s.Struct.SetUint32(100, v)
}

func (s CHF) Ipv4NextHop() uint32 {
	return s.Struct.Uint32(104)
}

func (s CHF) SetIpv4NextHop(v uint32) {
	s.Struct.SetUint32(104, v)
}

func (s CHF) MplsType() uint32 {
	return s.Struct.Uint32(108)
}

func (s CHF) SetMplsType(v uint32) {
	s.Struct.SetUint32(108, v)
}

func (s CHF) OutBytes() uint64 {
	return s.Struct.Uint64(112)
}

func (s CHF) SetOutBytes(v uint64) {
	s.Struct.SetUint64(112, v)
}

func (s CHF) OutPkts() uint64 {
	return s.Struct.Uint64(120)
}

func (s CHF) SetOutPkts(v uint64) {
	s.Struct.SetUint64(120, v)
}

func (s CHF) TcpRetransmit() uint32 {
	return s.Struct.Uint32(128)
}

func (s CHF) SetTcpRetransmit(v uint32) {
	s.Struct.SetUint32(128, v)
}

func (s CHF) SrcFlowTags() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s CHF) HasSrcFlowTags() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s CHF) SrcFlowTagsBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s CHF) SetSrcFlowTags(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s CHF) DstFlowTags() (string, error) {
	p, err := s.Struct.Ptr(1)
	return p.Text(), err
}

func (s CHF) HasDstFlowTags() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s CHF) DstFlowTagsBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return p.TextBytes(), err
}

func (s CHF) SetDstFlowTags(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(1, t.List.ToPtr())
}

func (s CHF) SampleRate() uint32 {
	return s.Struct.Uint32(132)
}

func (s CHF) SetSampleRate(v uint32) {
	s.Struct.SetUint32(132, v)
}

func (s CHF) DeviceId() uint32 {
	return s.Struct.Uint32(136)
}

func (s CHF) SetDeviceId(v uint32) {
	s.Struct.SetUint32(136, v)
}

func (s CHF) FlowTags() (string, error) {
	p, err := s.Struct.Ptr(2)
	return p.Text(), err
}

func (s CHF) HasFlowTags() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s CHF) FlowTagsBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return p.TextBytes(), err
}

func (s CHF) SetFlowTags(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(2, t.List.ToPtr())
}

func (s CHF) Timestamp() int64 {
	return int64(s.Struct.Uint64(144))
}

func (s CHF) SetTimestamp(v int64) {
	s.Struct.SetUint64(144, uint64(v))
}

func (s CHF) DstBgpAsPath() (string, error) {
	p, err := s.Struct.Ptr(3)
	return p.Text(), err
}

func (s CHF) HasDstBgpAsPath() bool {
	p, err := s.Struct.Ptr(3)
	return p.IsValid() || err != nil
}

func (s CHF) DstBgpAsPathBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(3)
	return p.TextBytes(), err
}

func (s CHF) SetDstBgpAsPath(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(3, t.List.ToPtr())
}

func (s CHF) DstBgpCommunity() (string, error) {
	p, err := s.Struct.Ptr(4)
	return p.Text(), err
}

func (s CHF) HasDstBgpCommunity() bool {
	p, err := s.Struct.Ptr(4)
	return p.IsValid() || err != nil
}

func (s CHF) DstBgpCommunityBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(4)
	return p.TextBytes(), err
}

func (s CHF) SetDstBgpCommunity(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(4, t.List.ToPtr())
}

func (s CHF) SrcBgpAsPath() (string, error) {
	p, err := s.Struct.Ptr(5)
	return p.Text(), err
}

func (s CHF) HasSrcBgpAsPath() bool {
	p, err := s.Struct.Ptr(5)
	return p.IsValid() || err != nil
}

func (s CHF) SrcBgpAsPathBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(5)
	return p.TextBytes(), err
}

func (s CHF) SetSrcBgpAsPath(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(5, t.List.ToPtr())
}

func (s CHF) SrcBgpCommunity() (string, error) {
	p, err := s.Struct.Ptr(6)
	return p.Text(), err
}

func (s CHF) HasSrcBgpCommunity() bool {
	p, err := s.Struct.Ptr(6)
	return p.IsValid() || err != nil
}

func (s CHF) SrcBgpCommunityBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(6)
	return p.TextBytes(), err
}

func (s CHF) SetSrcBgpCommunity(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(6, t.List.ToPtr())
}

func (s CHF) SrcNextHopAs() uint32 {
	return s.Struct.Uint32(140)
}

func (s CHF) SetSrcNextHopAs(v uint32) {
	s.Struct.SetUint32(140, v)
}

func (s CHF) DstNextHopAs() uint32 {
	return s.Struct.Uint32(152)
}

func (s CHF) SetDstNextHopAs(v uint32) {
	s.Struct.SetUint32(152, v)
}

func (s CHF) SrcGeoRegion() uint32 {
	return s.Struct.Uint32(156)
}

func (s CHF) SetSrcGeoRegion(v uint32) {
	s.Struct.SetUint32(156, v)
}

func (s CHF) DstGeoRegion() uint32 {
	return s.Struct.Uint32(160)
}

func (s CHF) SetDstGeoRegion(v uint32) {
	s.Struct.SetUint32(160, v)
}

func (s CHF) SrcGeoCity() uint32 {
	return s.Struct.Uint32(164)
}

func (s CHF) SetSrcGeoCity(v uint32) {
	s.Struct.SetUint32(164, v)
}

func (s CHF) DstGeoCity() uint32 {
	return s.Struct.Uint32(168)
}

func (s CHF) SetDstGeoCity(v uint32) {
	s.Struct.SetUint32(168, v)
}

func (s CHF) Big() bool {
	return s.Struct.Bit(1376)
}

func (s CHF) SetBig(v bool) {
	s.Struct.SetBit(1376, v)
}

func (s CHF) SampleAdj() bool {
	return s.Struct.Bit(1377)
}

func (s CHF) SetSampleAdj(v bool) {
	s.Struct.SetBit(1377, v)
}

func (s CHF) Ipv4DstNextHop() uint32 {
	return s.Struct.Uint32(176)
}

func (s CHF) SetIpv4DstNextHop(v uint32) {
	s.Struct.SetUint32(176, v)
}

func (s CHF) Ipv4SrcNextHop() uint32 {
	return s.Struct.Uint32(180)
}

func (s CHF) SetIpv4SrcNextHop(v uint32) {
	s.Struct.SetUint32(180, v)
}

func (s CHF) SrcRoutePrefix() uint32 {
	return s.Struct.Uint32(184)
}

func (s CHF) SetSrcRoutePrefix(v uint32) {
	s.Struct.SetUint32(184, v)
}

func (s CHF) DstRoutePrefix() uint32 {
	return s.Struct.Uint32(188)
}

func (s CHF) SetDstRoutePrefix(v uint32) {
	s.Struct.SetUint32(188, v)
}

func (s CHF) SrcRouteLength() uint8 {
	return s.Struct.Uint8(173)
}

func (s CHF) SetSrcRouteLength(v uint8) {
	s.Struct.SetUint8(173, v)
}

func (s CHF) DstRouteLength() uint8 {
	return s.Struct.Uint8(174)
}

func (s CHF) SetDstRouteLength(v uint8) {
	s.Struct.SetUint8(174, v)
}

func (s CHF) SrcSecondAsn() uint32 {
	return s.Struct.Uint32(192)
}

func (s CHF) SetSrcSecondAsn(v uint32) {
	s.Struct.SetUint32(192, v)
}

func (s CHF) DstSecondAsn() uint32 {
	return s.Struct.Uint32(196)
}

func (s CHF) SetDstSecondAsn(v uint32) {
	s.Struct.SetUint32(196, v)
}

func (s CHF) SrcThirdAsn() uint32 {
	return s.Struct.Uint32(200)
}

func (s CHF) SetSrcThirdAsn(v uint32) {
	s.Struct.SetUint32(200, v)
}

func (s CHF) DstThirdAsn() uint32 {
	return s.Struct.Uint32(204)
}

func (s CHF) SetDstThirdAsn(v uint32) {
	s.Struct.SetUint32(204, v)
}

func (s CHF) Ipv6DstAddr() ([]byte, error) {
	p, err := s.Struct.Ptr(7)
	return []byte(p.Data()), err
}

func (s CHF) HasIpv6DstAddr() bool {
	p, err := s.Struct.Ptr(7)
	return p.IsValid() || err != nil
}

func (s CHF) SetIpv6DstAddr(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(7, d.List.ToPtr())
}

func (s CHF) Ipv6SrcAddr() ([]byte, error) {
	p, err := s.Struct.Ptr(8)
	return []byte(p.Data()), err
}

func (s CHF) HasIpv6SrcAddr() bool {
	p, err := s.Struct.Ptr(8)
	return p.IsValid() || err != nil
}

func (s CHF) SetIpv6SrcAddr(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(8, d.List.ToPtr())
}

func (s CHF) SrcEthMac() uint64 {
	return s.Struct.Uint64(208)
}

func (s CHF) SetSrcEthMac(v uint64) {
	s.Struct.SetUint64(208, v)
}

func (s CHF) DstEthMac() uint64 {
	return s.Struct.Uint64(216)
}

func (s CHF) SetDstEthMac(v uint64) {
	s.Struct.SetUint64(216, v)
}

func (s CHF) Custom() (Custom_List, error) {
	p, err := s.Struct.Ptr(9)
	return Custom_List{List: p.List()}, err
}

func (s CHF) HasCustom() bool {
	p, err := s.Struct.Ptr(9)
	return p.IsValid() || err != nil
}

func (s CHF) SetCustom(v Custom_List) error {
	return s.Struct.SetPtr(9, v.List.ToPtr())
}

// NewCustom sets the custom field to a newly
// allocated Custom_List, preferring placement in s's segment.
func (s CHF) NewCustom(n int32) (Custom_List, error) {
	l, err := NewCustom_List(s.Struct.Segment(), n)
	if err != nil {
		return Custom_List{}, err
	}
	err = s.Struct.SetPtr(9, l.List.ToPtr())
	return l, err
}

func (s CHF) Ipv6SrcNextHop() ([]byte, error) {
	p, err := s.Struct.Ptr(10)
	return []byte(p.Data()), err
}

func (s CHF) HasIpv6SrcNextHop() bool {
	p, err := s.Struct.Ptr(10)
	return p.IsValid() || err != nil
}

func (s CHF) SetIpv6SrcNextHop(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(10, d.List.ToPtr())
}

func (s CHF) Ipv6DstNextHop() ([]byte, error) {
	p, err := s.Struct.Ptr(11)
	return []byte(p.Data()), err
}

func (s CHF) HasIpv6DstNextHop() bool {
	p, err := s.Struct.Ptr(11)
	return p.IsValid() || err != nil
}

func (s CHF) SetIpv6DstNextHop(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(11, d.List.ToPtr())
}

func (s CHF) Ipv6SrcRoutePrefix() ([]byte, error) {
	p, err := s.Struct.Ptr(12)
	return []byte(p.Data()), err
}

func (s CHF) HasIpv6SrcRoutePrefix() bool {
	p, err := s.Struct.Ptr(12)
	return p.IsValid() || err != nil
}

func (s CHF) SetIpv6SrcRoutePrefix(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(12, d.List.ToPtr())
}

func (s CHF) Ipv6DstRoutePrefix() ([]byte, error) {
	p, err := s.Struct.Ptr(13)
	return []byte(p.Data()), err
}

func (s CHF) HasIpv6DstRoutePrefix() bool {
	p, err := s.Struct.Ptr(13)
	return p.IsValid() || err != nil
}

func (s CHF) SetIpv6DstRoutePrefix(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(13, d.List.ToPtr())
}

// CHF_List is a list of CHF.
type CHF_List struct{ capnp.List }

// NewCHF creates a new list of CHF.
func NewCHF_List(s *capnp.Segment, sz int32) (CHF_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 224, PointerCount: 14}, sz)
	return CHF_List{l}, err
}

func (s CHF_List) At(i int) CHF { return CHF{s.List.Struct(i)} }

func (s CHF_List) Set(i int, v CHF) error { return s.List.SetStruct(i, v.Struct) }

// CHF_Promise is a wrapper for a CHF promised by a client call.
type CHF_Promise struct{ *capnp.Pipeline }

func (p CHF_Promise) Struct() (CHF, error) {
	s, err := p.Pipeline.Struct()
	return CHF{s}, err
}

type PackedCHF struct{ capnp.Struct }

// PackedCHF_TypeID is the unique identifier for the type PackedCHF.
const PackedCHF_TypeID = 0xb158a6a28e2d29c2

func NewPackedCHF(s *capnp.Segment) (PackedCHF, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PackedCHF{st}, err
}

func NewRootPackedCHF(s *capnp.Segment) (PackedCHF, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PackedCHF{st}, err
}

func ReadRootPackedCHF(msg *capnp.Message) (PackedCHF, error) {
	root, err := msg.RootPtr()
	return PackedCHF{root.Struct()}, err
}

func (s PackedCHF) String() string {
	str, _ := text.Marshal(0xb158a6a28e2d29c2, s.Struct)
	return str
}

func (s PackedCHF) Msgs() (CHF_List, error) {
	p, err := s.Struct.Ptr(0)
	return CHF_List{List: p.List()}, err
}

func (s PackedCHF) HasMsgs() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s PackedCHF) SetMsgs(v CHF_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMsgs sets the msgs field to a newly
// allocated CHF_List, preferring placement in s's segment.
func (s PackedCHF) NewMsgs(n int32) (CHF_List, error) {
	l, err := NewCHF_List(s.Struct.Segment(), n)
	if err != nil {
		return CHF_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// PackedCHF_List is a list of PackedCHF.
type PackedCHF_List struct{ capnp.List }

// NewPackedCHF creates a new list of PackedCHF.
func NewPackedCHF_List(s *capnp.Segment, sz int32) (PackedCHF_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return PackedCHF_List{l}, err
}

func (s PackedCHF_List) At(i int) PackedCHF { return PackedCHF{s.List.Struct(i)} }

func (s PackedCHF_List) Set(i int, v PackedCHF) error { return s.List.SetStruct(i, v.Struct) }

// PackedCHF_Promise is a wrapper for a PackedCHF promised by a client call.
type PackedCHF_Promise struct{ *capnp.Pipeline }

func (p PackedCHF_Promise) Struct() (PackedCHF, error) {
	s, err := p.Pipeline.Struct()
	return PackedCHF{s}, err
}

const schema_c75f49ee0059f55d = "x\xdat\xd7]l\x1cW\xd9\x07\xf0\xf3\xec\xae=\x1b" +
	"\xc7_\xeb=I\x9c\xc4\xa9\x9d\xc4I\xec\xd4v\xfc\x95" +
	"/\xd7\xe9\xda\x89\x93\xd7~e\x87]\xdb\x89R\x94\x88" +
	"Lv'\xf6\xb4\xfb\xc5\xee8\xc4%\xa8\x05%(E" +
	"\x05\x05T\xa4\x80z\x11\xaa\x0a\x8a\x08\x94JA\x02\x14" +
	"$\x84@\xe5\xa2H\xb9(R/@\xe4\xa2HTJ" +
	"\x81\x0b**AX\xfe\xffc\xef\x97k\"m\xe4\xf9" +
	"\xcd\x99\xe7\x9cy\xcey\xce\xcc\xf4\xff)0\xe6\x1b\xa8" +
	"y\xdb\xa7T\xac\xa3\xa6\xb6p\xe9\xf3\xf7\xdf_<\xff" +
	"\x83\xef\xa9X\x9d\xb4\xfd\xf3\xc2G\xcf\xfcu\xea3o" +
	"\xd74ZJ\x85\xa7$\x17\x9e\x11\xfc54%\x91F" +
	"%\x85_u\xf7~\xed\xb5\xef\x9e{K\x85\xea\xa4\xd4" +
	"\x94\x0d\xc23\xa1\xd7\xc2gB\xfc+\x16\x8a\xa0\xe5\xbe" +
	"\x07\xfe\x96/\x1f\xbe\xf0!\xa3\xfa\xaa\x9b~1\xf4\x95" +
	"\xf0K\xa6\xe9\x8d\xd0\x9bJ\xfe\xf6\xf8\x83\x1b7\xce\xde" +
	"\xf9W,$\xbe\xf2eg\xc4\x12\xbf\x04\xc25-?" +
	"W\x82\xff\xdfT\xbd\x85\xf8\xe2\xe5\xbe\xb8\x9dM\xab\xf6" +
	"\xec\xc8\x89\xc9SQ\x91\xb9\x80\xf8\x03J\x05\x04\xc1\x1a" +
	"\xac\x9cRs\xf5\x96_\xe6Z-\x9f \x9a\x16\xfa&" +
	"k\x10\xdeLo\xa3\xfb\xfcZp\xef\xe1\xad\xd6\x08\\" +
	"\xd3;\xe8\xfe\x80\x16?|\x87\xf1Vz'=P\xa3" +
	"\x05}\x84wZ\xb3\xf0\x0ez\x0f\xbd\x06qj\xe0\xdd" +
	"\xd6qx'\xbd\x9f^\x8b8\xb5\xf0^\x13\xa7\x8b>" +
	"L\xb7\xea\xb4\xb9\xfd\x01\x13\xa7\x9f>J\x0fn\xd4\x12" +
	"\x84\x1f5\xed\x87\xe9c\xf4\x0d\xf5Z6\xc0\x8fY\x97" +
	"\xe0\xa3\xf4Iz]\x83\x96:\xf8I\xe3\x13\xf4(}" +
	"c\xa3\x96\x8d\x9c\x09\x13\x7f\x9a~\x8e^\xdf\xa4\xa5\x1e" +
	"~\xc6\xf8<\xfd\"\xbd\xa1YK\x03\xfc\x82\xf5i\xf8" +
	"y\xfa\"\xbd1\xa4\xa5\x11\xeeX\xff\x0fO\xd0\xb3\xf4" +
	"\xa6\x16-M\xf0\x94\xf5\x0dx\x96~\x8d\xde\x1c\xd6\xd2" +
	"\x0c_6y\xf6\xe8/\xd2CZK\x08\xfe\x05s_" +
	"W\xe9\xd7\xe9-\x9b\xb4\xb4p\x19\x18\xbfF\xbfI\x0f" +
	"o\xd6\x12\xe6\xa20\xfd^\xa7\xdf\xa2\xeb-Z4\xfc" +
	"\xab\xd6.\xf8M\xfa+\xf4M\xadZ6\xc1\xbfn\xe2" +
	"\xbcL\xbfM\xdf\xbcU\xcbf\xf87\xcd\xbc\xdc\xa2\xbf" +
	"J\xdf\xb2M\xcb\x16\xf8\xb7L\xden\xd3_\xa7\xb7n" +
	"\xd7\xd2\x0a\xff\x8e\xe9\xf7\x0e\xfd.}+\xf2\xb9\x15\xfe" +
	"}\xe3o\xd0\xef\xd1\xb7!\x9f\xdb\xe0o\x99\xf8w\xe9" +
	"?\xa5o\xef\xd0\xb2\x1d\xfe\x13\xb3\x0e\xef\xd1\x7f\x09\x97" +
	"6-m\xe0_\x98n\xef\x93\x7f\xcb\xe6;D\xcb\x0e" +
	"\xf8o\x8c\xff\x9a\xfe\x80\xfe\xc4N-O\xc0\x7fg\xa6" +
	"\xe5\x1d\xfa{\xf4\xf6]Z\xda\xe1\xbf7\xc3y\x97\xfe" +
	"\x90\xde\x81e\xde\x01\xff\xa3\xf1?\xd0\xffB\xdf\x89\xe9" +
	"\xda\x09\xff\xb3\x99\xf6\xf7\xe9\x7f\xa7\xef\xc2\xb2E\"\xc3" +
	"\x1fZ\xcf\xc2\x1f\xd1?\xa6\xef\xc6\xb2\xdd\x0d\xff\xc8\xfa" +
	"\x12\xfc\x1f\xf0\xd9 \xb8\x13\xab\xbf\x13\xfc\xd84\xff7" +
	"\x9b\x07\xe9{j\xb5\xec\x81\xd7\x04\xd9<\x10\x847\xd3" +
	"\xf7\xee\xd6\xb2\x97\xd5\x18d\xfbzz+}\xdf^-" +
	"\xfbX\x8d\xc65\xbd\x83\xde\xb5OK\x17\xab\xcex\x1b" +
	"\xbd\x8b\xde\xdd\xa5\xa5\x1b\xbe\xc7x'\xbd\x9f\xbe\xbf[" +
	"\xcb~VW\x90\xe9\xe9\xa1\x1f\xa1?\xb9_\xcb\x93\xf0" +
	"\x83\xc6\x87\xe9c\xf0\x81\x9e\x8b\xb8\x81\x1e\x96Q\x90\xcb" +
	"\xe7\x08OL\xf0D\xaf\x8d\x13\xbd81\x1ed\x82\xc6" +
	"xb\x9a\x91\xfaz\xb4\xf4q\xf3\x0b>\x0f\x9f\xa4\xcf" +
	"\xd3\x0f\xf4j9\xc0\xfd\xcdx\x94~\x9e\xde\xdf\xa7\xa5" +
	"\x1f\xfe\x8c\xf1s\xf4\x04}\xe0\x80\x96\x01\xb8m\xfc\"" +
	"=I\x1f\xfc\xa1\x16TJ\xd85\xbeH\xf7\xe8C?" +
	"\xd22\x04\xff\xac\xf1,\xfd\x1a}\xb8_\xcb0\xeb\xcb" +
	"d\xe2*\xfd:\xfd\xe0\x80\x96\x83\xac#\xe3/\xd2_" +
	"\xa6\x1f\x1a\xd4r\x08\xfeR\x90\x0b\xeb&\xfd\x15\xfa\xe1" +
	"!-\x87Y/\xc6o\xd1_\xa5\x1f\xb1\xb4\x1ca]" +
	"\x18\xbfM\x7f\x9d~4\xa8\xe5(\xeb\xc2\xf8\x1d\xfa]" +
	"\xfa\x08\xeah\x84ua\xf2\xf6\x06\xfd\x1e\xfd)\xd4\xd1" +
	"S\xac\x0b\xe3?\xa6\xdf\xa7\x8fn\xd02\x0a\xffY\x90" +
	"uz\x8f\xfe\x80~\x0c\xfb\xe11.ts\xbf\xef\xd0" +
	"\xdf\xa3?\x8d\xfd\xf0i.t\xe3\xef\xd2\x1f\xd2#\xd8" +
	"\x0f#\\\xe8\xc1o\xc3\x1f\xd2\x1f\xd1\xc7\xb0\x1f\x8e\xc1" +
	"?0\xfe\x88\xfe1\xbc\xe0\xb9)'\xef\xd9)<-" +
	"N\xdb\xe9\x0c\xf6j\x1f~\xd2\x9e\xc8{\xe3y\xec\xb8" +
	">\xfc$\x82\xa3\xffs2\x95\x873v\xbcxXX" +
	"t\xec\x84\x93\x9bv\x94\xa4\x8b\xf6\x82\x9b>\xbe\xec9" +
	"y\xec\xcd>\xfc$\xe2\xa6\xa3\xcfy\xa5\xc3\x82\x9b\xce" +
	".y\xd1LN\x89W\x8a\xeaf\xe7\xdc\xe7\x9dRT" +
	"7{ex\x02\xa3PV\"\x91\xab\xd2\xb9\\|\x8d" +
	"&\xd9\xb2:\x1c\x0c\xed\xd6Xf\xc93\xfd*\x7f\xae" +
	"\x8c\xd9\\\xc6\xcb\xc43I\xa5T\xc9\xf2v*\x9bt" +
	"\x12Q\xb1\xe3\xcf9\x1e\xc6%\xa5\x81\xb5\xe7\xd1}9" +
	"58\xaaL\x0d\x0e+S\xe3\xc5\xb3\xa7\x92\xf6B\xbe" +
	"\"\xb6\xe5e\xcaW_I\xda\xe9\xa9r\xd6x\xf8\xa9" +
	"%\xaf\xeanO;W=eMf\xb2%\xc5\xd0\xf2" +
	"\xf3\xcbY\xa7r\xc0\xb83\x93p\xdaj\x92_\x80U" +
	"%\x1d\x83\x99u\xbc\x9c\xad\xda\xd3\xf9\x94[\xee\x05c" +
	">\x95\xcc|n^Y\x18)\x9e\x82>\xfc\xa4\x80I" +
	"^GW\x123k+\xbfW\x9e\xaa\x84s\xc5\x8d;" +
	"S\x89\xca\x01]\xe6\xb5\xabw^\xbc\xb8\xb4\xd8$[" +
	"\\h\xec\xe6\xf8Bv<\xaf\x9a\xa2\xb6\xb7X\xd9;" +
	"\xf8DFR\xa9\xa5\xb4\xeb-\x97\x83`\xb4\xeb]\xb0" +
	"\xc2\xff\xe3\x02\xe6pR5e\xb2\xe5\x89c\x0f\xeb\xf1" +
	"\xca|\xce:\xaai\xc1\xcd\xa4+[\xaf\xc7+\xadO" +
	"\xb8H\xc7\xf2\x9a\xb6\xd5h]r\x17\xf0\xa2\x85']" +
	")\x89\xe3\x09%\xcf\x96lu\xb9\x9fV\x11\x0e\xaab" +
	"\xb6WW\xfc'O\xa0\xefYL\xb1\xa3\"\xd1\x9cs" +
	"\xd9\xbdZ\xd9\xff\xfa'\xcaWL;\xe9\x05$\xaf\x16" +
	"'j\xab\xaeXs\x02W\xcc9\xf1LZ5%\xc6" +
	"\xf3U\xd9X\x8f\xd1z~\xd1\xcd%\x94\xb5\xa6\xf1:" +
	"\x8a\xdb:T.\xef\x06h\xc3\xaa\x96\xcb\xbb\xa8\x88{" +
	"\xd2[\x9c\xb1\x95\xc4K\x8b\x19Q\xd7Z$\xbe\x94\xf7" +
	"2)\xbc\x97I\xd4/\xd2\\~OVB,F/" +
	"\xa7\xb2\xb2\xdb\xaa\xe4\xaf\x19\xcf\xac0;\xcc\xa6\x1f\xe9" +
	"\\s\xd5\xfa'\x8b\xaf\xe1\x92\x1d\x89r\x13IX+" +
	"/\xe3\xb1@\xf1]<\xd4\x80Gt\x0c;q\xac\xd3" +
	"'M\xa9<\xaa\xac4\xf2\xd2\xe7\xc6\xea\xc8+\xa2\x9d" +
	"07\xa9\x14c\xd5\x97b\x9d\xc4\xcbVl\x0c\xb1\xa6" +
	"\xb1\x9c\xf0\xaf\xf8\xbd\x10\x9a\x1aT\xbe\x90/j\xde\xce" +
	"C\x03xH\xc5\xfa\xd1j\xd4'~7Q\xda\xd2\xae" +
	"\xd8\xc9%\xa7\xe0\xe6'P\xa0\xe9\xbc\xb2\xb8\xc4\x8b\x0b" +
	"\xf3\x13]G\xfaLsv\xdeV(\x88\xf9~\x08\x9d" +
	"\xc4\xe3,6\x81\xc8Q\x9f\xec\x90\xff\x90\xf1\xf9\x10\x9a" +
	"\xc1{Fl\x1a|\x0eY\xf1=.\x98\x8f\x87\xd0\x19" +
	"<\xe3b\xb8\xd1\xd8y<\x81\x96\xdc\xb474x\x16" +
	"\x13\x99\xac\xdc<l\xa2\xf2\xdbI\xbc\xc6\xfb\xf0\xc3\xd6" +
	"\xea\xe5\xce\xe2p\xb5\xac\xff\x1b\x00\x00\xff\xff\xc3\xc6\xa1" +
	"\x15"

func init() {
	schemas.Register(schema_c75f49ee0059f55d,
		0xa7ab5c68e4bc7b62,
		0xb158a6a28e2d29c2,
		0xed5d37861203d027,
		0xfba056008585e9fd)
}
