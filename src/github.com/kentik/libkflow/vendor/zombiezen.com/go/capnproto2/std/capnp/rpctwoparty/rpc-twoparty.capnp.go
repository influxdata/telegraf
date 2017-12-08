package rpctwoparty

// AUTO GENERATED - DO NOT EDIT

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Side uint16

// Values of Side.
const (
	Side_server Side = 0
	Side_client Side = 1
)

// String returns the enum's constant name.
func (c Side) String() string {
	switch c {
	case Side_server:
		return "server"
	case Side_client:
		return "client"

	default:
		return ""
	}
}

// SideFromString returns the enum value with a name,
// or the zero value if there's no such value.
func SideFromString(c string) Side {
	switch c {
	case "server":
		return Side_server
	case "client":
		return Side_client

	default:
		return 0
	}
}

type Side_List struct{ capnp.List }

func NewSide_List(s *capnp.Segment, sz int32) (Side_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return Side_List{l.List}, err
}

func (l Side_List) At(i int) Side {
	ul := capnp.UInt16List{List: l.List}
	return Side(ul.At(i))
}

func (l Side_List) Set(i int, v Side) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

type VatId struct{ capnp.Struct }

// VatId_TypeID is the unique identifier for the type VatId.
const VatId_TypeID = 0xd20b909fee733a8e

func NewVatId(s *capnp.Segment) (VatId, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return VatId{st}, err
}

func NewRootVatId(s *capnp.Segment) (VatId, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return VatId{st}, err
}

func ReadRootVatId(msg *capnp.Message) (VatId, error) {
	root, err := msg.RootPtr()
	return VatId{root.Struct()}, err
}

func (s VatId) String() string {
	str, _ := text.Marshal(0xd20b909fee733a8e, s.Struct)
	return str
}

func (s VatId) Side() Side {
	return Side(s.Struct.Uint16(0))
}

func (s VatId) SetSide(v Side) {
	s.Struct.SetUint16(0, uint16(v))
}

// VatId_List is a list of VatId.
type VatId_List struct{ capnp.List }

// NewVatId creates a new list of VatId.
func NewVatId_List(s *capnp.Segment, sz int32) (VatId_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return VatId_List{l}, err
}

func (s VatId_List) At(i int) VatId { return VatId{s.List.Struct(i)} }

func (s VatId_List) Set(i int, v VatId) error { return s.List.SetStruct(i, v.Struct) }

// VatId_Promise is a wrapper for a VatId promised by a client call.
type VatId_Promise struct{ *capnp.Pipeline }

func (p VatId_Promise) Struct() (VatId, error) {
	s, err := p.Pipeline.Struct()
	return VatId{s}, err
}

type ProvisionId struct{ capnp.Struct }

// ProvisionId_TypeID is the unique identifier for the type ProvisionId.
const ProvisionId_TypeID = 0xb88d09a9c5f39817

func NewProvisionId(s *capnp.Segment) (ProvisionId, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return ProvisionId{st}, err
}

func NewRootProvisionId(s *capnp.Segment) (ProvisionId, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return ProvisionId{st}, err
}

func ReadRootProvisionId(msg *capnp.Message) (ProvisionId, error) {
	root, err := msg.RootPtr()
	return ProvisionId{root.Struct()}, err
}

func (s ProvisionId) String() string {
	str, _ := text.Marshal(0xb88d09a9c5f39817, s.Struct)
	return str
}

func (s ProvisionId) JoinId() uint32 {
	return s.Struct.Uint32(0)
}

func (s ProvisionId) SetJoinId(v uint32) {
	s.Struct.SetUint32(0, v)
}

// ProvisionId_List is a list of ProvisionId.
type ProvisionId_List struct{ capnp.List }

// NewProvisionId creates a new list of ProvisionId.
func NewProvisionId_List(s *capnp.Segment, sz int32) (ProvisionId_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return ProvisionId_List{l}, err
}

func (s ProvisionId_List) At(i int) ProvisionId { return ProvisionId{s.List.Struct(i)} }

func (s ProvisionId_List) Set(i int, v ProvisionId) error { return s.List.SetStruct(i, v.Struct) }

// ProvisionId_Promise is a wrapper for a ProvisionId promised by a client call.
type ProvisionId_Promise struct{ *capnp.Pipeline }

func (p ProvisionId_Promise) Struct() (ProvisionId, error) {
	s, err := p.Pipeline.Struct()
	return ProvisionId{s}, err
}

type RecipientId struct{ capnp.Struct }

// RecipientId_TypeID is the unique identifier for the type RecipientId.
const RecipientId_TypeID = 0x89f389b6fd4082c1

func NewRecipientId(s *capnp.Segment) (RecipientId, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return RecipientId{st}, err
}

func NewRootRecipientId(s *capnp.Segment) (RecipientId, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return RecipientId{st}, err
}

func ReadRootRecipientId(msg *capnp.Message) (RecipientId, error) {
	root, err := msg.RootPtr()
	return RecipientId{root.Struct()}, err
}

func (s RecipientId) String() string {
	str, _ := text.Marshal(0x89f389b6fd4082c1, s.Struct)
	return str
}

// RecipientId_List is a list of RecipientId.
type RecipientId_List struct{ capnp.List }

// NewRecipientId creates a new list of RecipientId.
func NewRecipientId_List(s *capnp.Segment, sz int32) (RecipientId_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return RecipientId_List{l}, err
}

func (s RecipientId_List) At(i int) RecipientId { return RecipientId{s.List.Struct(i)} }

func (s RecipientId_List) Set(i int, v RecipientId) error { return s.List.SetStruct(i, v.Struct) }

// RecipientId_Promise is a wrapper for a RecipientId promised by a client call.
type RecipientId_Promise struct{ *capnp.Pipeline }

func (p RecipientId_Promise) Struct() (RecipientId, error) {
	s, err := p.Pipeline.Struct()
	return RecipientId{s}, err
}

type ThirdPartyCapId struct{ capnp.Struct }

// ThirdPartyCapId_TypeID is the unique identifier for the type ThirdPartyCapId.
const ThirdPartyCapId_TypeID = 0xb47f4979672cb59d

func NewThirdPartyCapId(s *capnp.Segment) (ThirdPartyCapId, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return ThirdPartyCapId{st}, err
}

func NewRootThirdPartyCapId(s *capnp.Segment) (ThirdPartyCapId, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return ThirdPartyCapId{st}, err
}

func ReadRootThirdPartyCapId(msg *capnp.Message) (ThirdPartyCapId, error) {
	root, err := msg.RootPtr()
	return ThirdPartyCapId{root.Struct()}, err
}

func (s ThirdPartyCapId) String() string {
	str, _ := text.Marshal(0xb47f4979672cb59d, s.Struct)
	return str
}

// ThirdPartyCapId_List is a list of ThirdPartyCapId.
type ThirdPartyCapId_List struct{ capnp.List }

// NewThirdPartyCapId creates a new list of ThirdPartyCapId.
func NewThirdPartyCapId_List(s *capnp.Segment, sz int32) (ThirdPartyCapId_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return ThirdPartyCapId_List{l}, err
}

func (s ThirdPartyCapId_List) At(i int) ThirdPartyCapId { return ThirdPartyCapId{s.List.Struct(i)} }

func (s ThirdPartyCapId_List) Set(i int, v ThirdPartyCapId) error {
	return s.List.SetStruct(i, v.Struct)
}

// ThirdPartyCapId_Promise is a wrapper for a ThirdPartyCapId promised by a client call.
type ThirdPartyCapId_Promise struct{ *capnp.Pipeline }

func (p ThirdPartyCapId_Promise) Struct() (ThirdPartyCapId, error) {
	s, err := p.Pipeline.Struct()
	return ThirdPartyCapId{s}, err
}

type JoinKeyPart struct{ capnp.Struct }

// JoinKeyPart_TypeID is the unique identifier for the type JoinKeyPart.
const JoinKeyPart_TypeID = 0x95b29059097fca83

func NewJoinKeyPart(s *capnp.Segment) (JoinKeyPart, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return JoinKeyPart{st}, err
}

func NewRootJoinKeyPart(s *capnp.Segment) (JoinKeyPart, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return JoinKeyPart{st}, err
}

func ReadRootJoinKeyPart(msg *capnp.Message) (JoinKeyPart, error) {
	root, err := msg.RootPtr()
	return JoinKeyPart{root.Struct()}, err
}

func (s JoinKeyPart) String() string {
	str, _ := text.Marshal(0x95b29059097fca83, s.Struct)
	return str
}

func (s JoinKeyPart) JoinId() uint32 {
	return s.Struct.Uint32(0)
}

func (s JoinKeyPart) SetJoinId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s JoinKeyPart) PartCount() uint16 {
	return s.Struct.Uint16(4)
}

func (s JoinKeyPart) SetPartCount(v uint16) {
	s.Struct.SetUint16(4, v)
}

func (s JoinKeyPart) PartNum() uint16 {
	return s.Struct.Uint16(6)
}

func (s JoinKeyPart) SetPartNum(v uint16) {
	s.Struct.SetUint16(6, v)
}

// JoinKeyPart_List is a list of JoinKeyPart.
type JoinKeyPart_List struct{ capnp.List }

// NewJoinKeyPart creates a new list of JoinKeyPart.
func NewJoinKeyPart_List(s *capnp.Segment, sz int32) (JoinKeyPart_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return JoinKeyPart_List{l}, err
}

func (s JoinKeyPart_List) At(i int) JoinKeyPart { return JoinKeyPart{s.List.Struct(i)} }

func (s JoinKeyPart_List) Set(i int, v JoinKeyPart) error { return s.List.SetStruct(i, v.Struct) }

// JoinKeyPart_Promise is a wrapper for a JoinKeyPart promised by a client call.
type JoinKeyPart_Promise struct{ *capnp.Pipeline }

func (p JoinKeyPart_Promise) Struct() (JoinKeyPart, error) {
	s, err := p.Pipeline.Struct()
	return JoinKeyPart{s}, err
}

type JoinResult struct{ capnp.Struct }

// JoinResult_TypeID is the unique identifier for the type JoinResult.
const JoinResult_TypeID = 0x9d263a3630b7ebee

func NewJoinResult(s *capnp.Segment) (JoinResult, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return JoinResult{st}, err
}

func NewRootJoinResult(s *capnp.Segment) (JoinResult, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return JoinResult{st}, err
}

func ReadRootJoinResult(msg *capnp.Message) (JoinResult, error) {
	root, err := msg.RootPtr()
	return JoinResult{root.Struct()}, err
}

func (s JoinResult) String() string {
	str, _ := text.Marshal(0x9d263a3630b7ebee, s.Struct)
	return str
}

func (s JoinResult) JoinId() uint32 {
	return s.Struct.Uint32(0)
}

func (s JoinResult) SetJoinId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s JoinResult) Succeeded() bool {
	return s.Struct.Bit(32)
}

func (s JoinResult) SetSucceeded(v bool) {
	s.Struct.SetBit(32, v)
}

func (s JoinResult) Cap() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s JoinResult) HasCap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JoinResult) CapPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s JoinResult) SetCap(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s JoinResult) SetCapPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

// JoinResult_List is a list of JoinResult.
type JoinResult_List struct{ capnp.List }

// NewJoinResult creates a new list of JoinResult.
func NewJoinResult_List(s *capnp.Segment, sz int32) (JoinResult_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return JoinResult_List{l}, err
}

func (s JoinResult_List) At(i int) JoinResult { return JoinResult{s.List.Struct(i)} }

func (s JoinResult_List) Set(i int, v JoinResult) error { return s.List.SetStruct(i, v.Struct) }

// JoinResult_Promise is a wrapper for a JoinResult promised by a client call.
type JoinResult_Promise struct{ *capnp.Pipeline }

func (p JoinResult_Promise) Struct() (JoinResult, error) {
	s, err := p.Pipeline.Struct()
	return JoinResult{s}, err
}

func (p JoinResult_Promise) Cap() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

const schema_a184c7885cdaf2a1 = "x\xda|\x92\xcfk\x13A\x14\xc7\xdf\x9bIL\x82\x96" +
	"\xb0\x9d\x88\xe8E\x11\xf4 *-\x8aB@6\x18\x04" +
	"W\x8bdc\x15\x0b\xf6\x10v\x07]iw\xb7\xbbI" +
	"e\x0f\xb2\xa0\x82\xf6\xa0\xf4\"x\xb0\x96\x1e=\x89\xe2" +
	"\xaf\x82\x1e\x14D\xe8\xd1\x83\x07\xff\x05\xa1\x87\xf6V\x90" +
	"\xf5\xcd\x82\x8d$\xbb\x1e\x86\xbc|\xf9\xbc\xef|\xdf\x9b" +
	"\x1d;\x83\x0d6^\x8c\x0b\x00\xe6\xc9\xe2\x8e\xe4\xf3\xdd" +
	"\xc6\xefw\x0b\x9b\x0b\xa0\x09LV6~^\x7f\xf8\xed" +
	"\xfe\x0a\x14J\x00\xe2\x10\xdb\x12\xe3\xac\x04<\xb9\xb7\x16" +
	"W\xa6\x16_?\x01S\xe0 5B\xd4^\xa6\xaa\xdd" +
	"\xec%`\xb2\xfe\xeb\xfd\xd8\xa9\xfa\xe1\xa5\x01\xb6\x88\x0a" +
	"y\xc56\xc4\xa7\x14^M\xe1gs\xab\x0f>>\xff" +
	"\xb1L\xb7\xb3>\x0b(\xae\xf0/b\x9a+p\x8a\x9f" +
	"&p\xe9\xed\xd1\x1b\x91\x11\xbf\x19\x8eyb\x9a\xefC" +
	"1\xcbU\xce=O7\xbf\xbe\xa8<\xfa\x90\x95\xd3\xe0" +
	"[\xe4\xaa*\x93\xeb\xe4\xf8\xb8\x1e\xae//\xee\xfc\x9e" +
	"\xc5\xce\xf15q'e#b\x9bI\xe0[\xc7\xba\xb7" +
	"=\x9fu\x82nt\xdc\xea\xf8\xae_oK\xcb\xf1u" +
	"G\xba]\xc3n\x91C\x16s\xc1s\xdc\x8b\xba\x8cZ" +
	"\xa4\x11c\xee\xe2\xb4\xf6\x02\x02h\xe7\xea\xb4\xff\x06G" +
	"s\x82\xa1\x86\xac\x86J4\xda$\x9e'q\x92D\xc6" +
	"k\xc8H4\xcf\x928A\xe25\x86\xfa-24l" +
	",\x03\xa3\x83\x89O\xbeM\xaf\xe7\x02vi\xbd\x8c\x0e" +
	"\xc6J\xbb\xd4\x9b\xfd\xfb?7W{\xbf\x0c{3\xff" +
	"\x8du`8\x96\x8a\x9a\xa6:\x98\x9f*\xecY\x96\x94" +
	"\xb6\x04\xb4\xa9\x9fz\x00Kt/\x8eR=\x9a\x93\xe8" +
	"\xb2C\xbc\xcaRN\xed5\xca\x81\xa8U\xe8G\x0fe" +
	"0/\x03\xdd\x9aQ\xbb\xden\xe6\xff4O\xdet\x02" +
	"[\xed8jv|\x9e\xff\x1c\xad\xc0\x9bw\xf4\xd0\xf1" +
	"\xdc\x941\x0b\xdbs\x8f\xa8\xb9\xcb4Nmx\x9c," +
	"\xa7\xab\x1dzu\x80\x01\x93#}\x93jH\x03a\xb5" +
	"\xff\x81\xd38U\xc0?\x01\x00\x00\xff\xff+!\xf3\xa9"

func init() {
	schemas.Register(schema_a184c7885cdaf2a1,
		0x89f389b6fd4082c1,
		0x95b29059097fca83,
		0x9d263a3630b7ebee,
		0x9fd69ebc87b9719c,
		0xb47f4979672cb59d,
		0xb88d09a9c5f39817,
		0xd20b909fee733a8e)
}
