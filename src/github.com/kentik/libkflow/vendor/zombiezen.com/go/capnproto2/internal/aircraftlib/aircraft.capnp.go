package aircraftlib

// AUTO GENERATED - DO NOT EDIT

import (
	context "golang.org/x/net/context"
	math "math"
	strconv "strconv"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
	server "zombiezen.com/go/capnproto2/server"
)

// Constants defined in aircraft.capnp.
const (
	ConstEnum = Airport_jfk
)

// Constants defined in aircraft.capnp.
var (
	ConstDate = Zdate{Struct: capnp.MustUnmarshalRootPtr(x_832bcc6686a26d56[0:24]).Struct()}
	ConstList = Zdate_List{List: capnp.MustUnmarshalRootPtr(x_832bcc6686a26d56[24:64]).List()}
)

type Zdate struct{ capnp.Struct }

// Zdate_TypeID is the unique identifier for the type Zdate.
const Zdate_TypeID = 0xde50aebbad57549d

func NewZdate(s *capnp.Segment) (Zdate, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Zdate{st}, err
}

func NewRootZdate(s *capnp.Segment) (Zdate, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Zdate{st}, err
}

func ReadRootZdate(msg *capnp.Message) (Zdate, error) {
	root, err := msg.RootPtr()
	return Zdate{root.Struct()}, err
}

func (s Zdate) String() string {
	str, _ := text.Marshal(0xde50aebbad57549d, s.Struct)
	return str
}

func (s Zdate) Year() int16 {
	return int16(s.Struct.Uint16(0))
}

func (s Zdate) SetYear(v int16) {
	s.Struct.SetUint16(0, uint16(v))
}

func (s Zdate) Month() uint8 {
	return s.Struct.Uint8(2)
}

func (s Zdate) SetMonth(v uint8) {
	s.Struct.SetUint8(2, v)
}

func (s Zdate) Day() uint8 {
	return s.Struct.Uint8(3)
}

func (s Zdate) SetDay(v uint8) {
	s.Struct.SetUint8(3, v)
}

// Zdate_List is a list of Zdate.
type Zdate_List struct{ capnp.List }

// NewZdate creates a new list of Zdate.
func NewZdate_List(s *capnp.Segment, sz int32) (Zdate_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return Zdate_List{l}, err
}

func (s Zdate_List) At(i int) Zdate { return Zdate{s.List.Struct(i)} }

func (s Zdate_List) Set(i int, v Zdate) error { return s.List.SetStruct(i, v.Struct) }

// Zdate_Promise is a wrapper for a Zdate promised by a client call.
type Zdate_Promise struct{ *capnp.Pipeline }

func (p Zdate_Promise) Struct() (Zdate, error) {
	s, err := p.Pipeline.Struct()
	return Zdate{s}, err
}

type Zdata struct{ capnp.Struct }

// Zdata_TypeID is the unique identifier for the type Zdata.
const Zdata_TypeID = 0xc7da65f9a2f20ba2

func NewZdata(s *capnp.Segment) (Zdata, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Zdata{st}, err
}

func NewRootZdata(s *capnp.Segment) (Zdata, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Zdata{st}, err
}

func ReadRootZdata(msg *capnp.Message) (Zdata, error) {
	root, err := msg.RootPtr()
	return Zdata{root.Struct()}, err
}

func (s Zdata) String() string {
	str, _ := text.Marshal(0xc7da65f9a2f20ba2, s.Struct)
	return str
}

func (s Zdata) Data() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Zdata) HasData() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Zdata) SetData(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, d.List.ToPtr())
}

// Zdata_List is a list of Zdata.
type Zdata_List struct{ capnp.List }

// NewZdata creates a new list of Zdata.
func NewZdata_List(s *capnp.Segment, sz int32) (Zdata_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Zdata_List{l}, err
}

func (s Zdata_List) At(i int) Zdata { return Zdata{s.List.Struct(i)} }

func (s Zdata_List) Set(i int, v Zdata) error { return s.List.SetStruct(i, v.Struct) }

// Zdata_Promise is a wrapper for a Zdata promised by a client call.
type Zdata_Promise struct{ *capnp.Pipeline }

func (p Zdata_Promise) Struct() (Zdata, error) {
	s, err := p.Pipeline.Struct()
	return Zdata{s}, err
}

type Airport uint16

// Values of Airport.
const (
	Airport_none Airport = 0
	Airport_jfk  Airport = 1
	Airport_lax  Airport = 2
	Airport_sfo  Airport = 3
	Airport_luv  Airport = 4
	Airport_dfw  Airport = 5
	Airport_test Airport = 6
)

// String returns the enum's constant name.
func (c Airport) String() string {
	switch c {
	case Airport_none:
		return "none"
	case Airport_jfk:
		return "jfk"
	case Airport_lax:
		return "lax"
	case Airport_sfo:
		return "sfo"
	case Airport_luv:
		return "luv"
	case Airport_dfw:
		return "dfw"
	case Airport_test:
		return "test"

	default:
		return ""
	}
}

// AirportFromString returns the enum value with a name,
// or the zero value if there's no such value.
func AirportFromString(c string) Airport {
	switch c {
	case "none":
		return Airport_none
	case "jfk":
		return Airport_jfk
	case "lax":
		return Airport_lax
	case "sfo":
		return Airport_sfo
	case "luv":
		return Airport_luv
	case "dfw":
		return Airport_dfw
	case "test":
		return Airport_test

	default:
		return 0
	}
}

type Airport_List struct{ capnp.List }

func NewAirport_List(s *capnp.Segment, sz int32) (Airport_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return Airport_List{l.List}, err
}

func (l Airport_List) At(i int) Airport {
	ul := capnp.UInt16List{List: l.List}
	return Airport(ul.At(i))
}

func (l Airport_List) Set(i int, v Airport) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

type PlaneBase struct{ capnp.Struct }

// PlaneBase_TypeID is the unique identifier for the type PlaneBase.
const PlaneBase_TypeID = 0xd8bccf6e60a73791

func NewPlaneBase(s *capnp.Segment) (PlaneBase, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 2})
	return PlaneBase{st}, err
}

func NewRootPlaneBase(s *capnp.Segment) (PlaneBase, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 2})
	return PlaneBase{st}, err
}

func ReadRootPlaneBase(msg *capnp.Message) (PlaneBase, error) {
	root, err := msg.RootPtr()
	return PlaneBase{root.Struct()}, err
}

func (s PlaneBase) String() string {
	str, _ := text.Marshal(0xd8bccf6e60a73791, s.Struct)
	return str
}

func (s PlaneBase) Name() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s PlaneBase) HasName() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s PlaneBase) NameBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s PlaneBase) SetName(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s PlaneBase) Homes() (Airport_List, error) {
	p, err := s.Struct.Ptr(1)
	return Airport_List{List: p.List()}, err
}

func (s PlaneBase) HasHomes() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s PlaneBase) SetHomes(v Airport_List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewHomes sets the homes field to a newly
// allocated Airport_List, preferring placement in s's segment.
func (s PlaneBase) NewHomes(n int32) (Airport_List, error) {
	l, err := NewAirport_List(s.Struct.Segment(), n)
	if err != nil {
		return Airport_List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

func (s PlaneBase) Rating() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s PlaneBase) SetRating(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

func (s PlaneBase) CanFly() bool {
	return s.Struct.Bit(64)
}

func (s PlaneBase) SetCanFly(v bool) {
	s.Struct.SetBit(64, v)
}

func (s PlaneBase) Capacity() int64 {
	return int64(s.Struct.Uint64(16))
}

func (s PlaneBase) SetCapacity(v int64) {
	s.Struct.SetUint64(16, uint64(v))
}

func (s PlaneBase) MaxSpeed() float64 {
	return math.Float64frombits(s.Struct.Uint64(24))
}

func (s PlaneBase) SetMaxSpeed(v float64) {
	s.Struct.SetUint64(24, math.Float64bits(v))
}

// PlaneBase_List is a list of PlaneBase.
type PlaneBase_List struct{ capnp.List }

// NewPlaneBase creates a new list of PlaneBase.
func NewPlaneBase_List(s *capnp.Segment, sz int32) (PlaneBase_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 32, PointerCount: 2}, sz)
	return PlaneBase_List{l}, err
}

func (s PlaneBase_List) At(i int) PlaneBase { return PlaneBase{s.List.Struct(i)} }

func (s PlaneBase_List) Set(i int, v PlaneBase) error { return s.List.SetStruct(i, v.Struct) }

// PlaneBase_Promise is a wrapper for a PlaneBase promised by a client call.
type PlaneBase_Promise struct{ *capnp.Pipeline }

func (p PlaneBase_Promise) Struct() (PlaneBase, error) {
	s, err := p.Pipeline.Struct()
	return PlaneBase{s}, err
}

type B737 struct{ capnp.Struct }

// B737_TypeID is the unique identifier for the type B737.
const B737_TypeID = 0xccb3b2e3603826e0

func NewB737(s *capnp.Segment) (B737, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return B737{st}, err
}

func NewRootB737(s *capnp.Segment) (B737, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return B737{st}, err
}

func ReadRootB737(msg *capnp.Message) (B737, error) {
	root, err := msg.RootPtr()
	return B737{root.Struct()}, err
}

func (s B737) String() string {
	str, _ := text.Marshal(0xccb3b2e3603826e0, s.Struct)
	return str
}

func (s B737) Base() (PlaneBase, error) {
	p, err := s.Struct.Ptr(0)
	return PlaneBase{Struct: p.Struct()}, err
}

func (s B737) HasBase() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s B737) SetBase(v PlaneBase) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBase sets the base field to a newly
// allocated PlaneBase struct, preferring placement in s's segment.
func (s B737) NewBase() (PlaneBase, error) {
	ss, err := NewPlaneBase(s.Struct.Segment())
	if err != nil {
		return PlaneBase{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// B737_List is a list of B737.
type B737_List struct{ capnp.List }

// NewB737 creates a new list of B737.
func NewB737_List(s *capnp.Segment, sz int32) (B737_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return B737_List{l}, err
}

func (s B737_List) At(i int) B737 { return B737{s.List.Struct(i)} }

func (s B737_List) Set(i int, v B737) error { return s.List.SetStruct(i, v.Struct) }

// B737_Promise is a wrapper for a B737 promised by a client call.
type B737_Promise struct{ *capnp.Pipeline }

func (p B737_Promise) Struct() (B737, error) {
	s, err := p.Pipeline.Struct()
	return B737{s}, err
}

func (p B737_Promise) Base() PlaneBase_Promise {
	return PlaneBase_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type A320 struct{ capnp.Struct }

// A320_TypeID is the unique identifier for the type A320.
const A320_TypeID = 0xd98c608877d9cb8d

func NewA320(s *capnp.Segment) (A320, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return A320{st}, err
}

func NewRootA320(s *capnp.Segment) (A320, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return A320{st}, err
}

func ReadRootA320(msg *capnp.Message) (A320, error) {
	root, err := msg.RootPtr()
	return A320{root.Struct()}, err
}

func (s A320) String() string {
	str, _ := text.Marshal(0xd98c608877d9cb8d, s.Struct)
	return str
}

func (s A320) Base() (PlaneBase, error) {
	p, err := s.Struct.Ptr(0)
	return PlaneBase{Struct: p.Struct()}, err
}

func (s A320) HasBase() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s A320) SetBase(v PlaneBase) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBase sets the base field to a newly
// allocated PlaneBase struct, preferring placement in s's segment.
func (s A320) NewBase() (PlaneBase, error) {
	ss, err := NewPlaneBase(s.Struct.Segment())
	if err != nil {
		return PlaneBase{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// A320_List is a list of A320.
type A320_List struct{ capnp.List }

// NewA320 creates a new list of A320.
func NewA320_List(s *capnp.Segment, sz int32) (A320_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return A320_List{l}, err
}

func (s A320_List) At(i int) A320 { return A320{s.List.Struct(i)} }

func (s A320_List) Set(i int, v A320) error { return s.List.SetStruct(i, v.Struct) }

// A320_Promise is a wrapper for a A320 promised by a client call.
type A320_Promise struct{ *capnp.Pipeline }

func (p A320_Promise) Struct() (A320, error) {
	s, err := p.Pipeline.Struct()
	return A320{s}, err
}

func (p A320_Promise) Base() PlaneBase_Promise {
	return PlaneBase_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type F16 struct{ capnp.Struct }

// F16_TypeID is the unique identifier for the type F16.
const F16_TypeID = 0xe1c9eac512335361

func NewF16(s *capnp.Segment) (F16, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return F16{st}, err
}

func NewRootF16(s *capnp.Segment) (F16, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return F16{st}, err
}

func ReadRootF16(msg *capnp.Message) (F16, error) {
	root, err := msg.RootPtr()
	return F16{root.Struct()}, err
}

func (s F16) String() string {
	str, _ := text.Marshal(0xe1c9eac512335361, s.Struct)
	return str
}

func (s F16) Base() (PlaneBase, error) {
	p, err := s.Struct.Ptr(0)
	return PlaneBase{Struct: p.Struct()}, err
}

func (s F16) HasBase() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s F16) SetBase(v PlaneBase) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBase sets the base field to a newly
// allocated PlaneBase struct, preferring placement in s's segment.
func (s F16) NewBase() (PlaneBase, error) {
	ss, err := NewPlaneBase(s.Struct.Segment())
	if err != nil {
		return PlaneBase{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// F16_List is a list of F16.
type F16_List struct{ capnp.List }

// NewF16 creates a new list of F16.
func NewF16_List(s *capnp.Segment, sz int32) (F16_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return F16_List{l}, err
}

func (s F16_List) At(i int) F16 { return F16{s.List.Struct(i)} }

func (s F16_List) Set(i int, v F16) error { return s.List.SetStruct(i, v.Struct) }

// F16_Promise is a wrapper for a F16 promised by a client call.
type F16_Promise struct{ *capnp.Pipeline }

func (p F16_Promise) Struct() (F16, error) {
	s, err := p.Pipeline.Struct()
	return F16{s}, err
}

func (p F16_Promise) Base() PlaneBase_Promise {
	return PlaneBase_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Regression struct{ capnp.Struct }

// Regression_TypeID is the unique identifier for the type Regression.
const Regression_TypeID = 0xb1f0385d845e367f

func NewRegression(s *capnp.Segment) (Regression, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3})
	return Regression{st}, err
}

func NewRootRegression(s *capnp.Segment) (Regression, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3})
	return Regression{st}, err
}

func ReadRootRegression(msg *capnp.Message) (Regression, error) {
	root, err := msg.RootPtr()
	return Regression{root.Struct()}, err
}

func (s Regression) String() string {
	str, _ := text.Marshal(0xb1f0385d845e367f, s.Struct)
	return str
}

func (s Regression) Base() (PlaneBase, error) {
	p, err := s.Struct.Ptr(0)
	return PlaneBase{Struct: p.Struct()}, err
}

func (s Regression) HasBase() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Regression) SetBase(v PlaneBase) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBase sets the base field to a newly
// allocated PlaneBase struct, preferring placement in s's segment.
func (s Regression) NewBase() (PlaneBase, error) {
	ss, err := NewPlaneBase(s.Struct.Segment())
	if err != nil {
		return PlaneBase{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Regression) B0() float64 {
	return math.Float64frombits(s.Struct.Uint64(0))
}

func (s Regression) SetB0(v float64) {
	s.Struct.SetUint64(0, math.Float64bits(v))
}

func (s Regression) Beta() (capnp.Float64List, error) {
	p, err := s.Struct.Ptr(1)
	return capnp.Float64List{List: p.List()}, err
}

func (s Regression) HasBeta() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Regression) SetBeta(v capnp.Float64List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewBeta sets the beta field to a newly
// allocated capnp.Float64List, preferring placement in s's segment.
func (s Regression) NewBeta(n int32) (capnp.Float64List, error) {
	l, err := capnp.NewFloat64List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Float64List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

func (s Regression) Planes() (Aircraft_List, error) {
	p, err := s.Struct.Ptr(2)
	return Aircraft_List{List: p.List()}, err
}

func (s Regression) HasPlanes() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s Regression) SetPlanes(v Aircraft_List) error {
	return s.Struct.SetPtr(2, v.List.ToPtr())
}

// NewPlanes sets the planes field to a newly
// allocated Aircraft_List, preferring placement in s's segment.
func (s Regression) NewPlanes(n int32) (Aircraft_List, error) {
	l, err := NewAircraft_List(s.Struct.Segment(), n)
	if err != nil {
		return Aircraft_List{}, err
	}
	err = s.Struct.SetPtr(2, l.List.ToPtr())
	return l, err
}

func (s Regression) Ymu() float64 {
	return math.Float64frombits(s.Struct.Uint64(8))
}

func (s Regression) SetYmu(v float64) {
	s.Struct.SetUint64(8, math.Float64bits(v))
}

func (s Regression) Ysd() float64 {
	return math.Float64frombits(s.Struct.Uint64(16))
}

func (s Regression) SetYsd(v float64) {
	s.Struct.SetUint64(16, math.Float64bits(v))
}

// Regression_List is a list of Regression.
type Regression_List struct{ capnp.List }

// NewRegression creates a new list of Regression.
func NewRegression_List(s *capnp.Segment, sz int32) (Regression_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3}, sz)
	return Regression_List{l}, err
}

func (s Regression_List) At(i int) Regression { return Regression{s.List.Struct(i)} }

func (s Regression_List) Set(i int, v Regression) error { return s.List.SetStruct(i, v.Struct) }

// Regression_Promise is a wrapper for a Regression promised by a client call.
type Regression_Promise struct{ *capnp.Pipeline }

func (p Regression_Promise) Struct() (Regression, error) {
	s, err := p.Pipeline.Struct()
	return Regression{s}, err
}

func (p Regression_Promise) Base() PlaneBase_Promise {
	return PlaneBase_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Aircraft struct{ capnp.Struct }
type Aircraft_Which uint16

const (
	Aircraft_Which_void Aircraft_Which = 0
	Aircraft_Which_b737 Aircraft_Which = 1
	Aircraft_Which_a320 Aircraft_Which = 2
	Aircraft_Which_f16  Aircraft_Which = 3
)

func (w Aircraft_Which) String() string {
	const s = "voidb737a320f16"
	switch w {
	case Aircraft_Which_void:
		return s[0:4]
	case Aircraft_Which_b737:
		return s[4:8]
	case Aircraft_Which_a320:
		return s[8:12]
	case Aircraft_Which_f16:
		return s[12:15]

	}
	return "Aircraft_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Aircraft_TypeID is the unique identifier for the type Aircraft.
const Aircraft_TypeID = 0xe54e10aede55c7b1

func NewAircraft(s *capnp.Segment) (Aircraft, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Aircraft{st}, err
}

func NewRootAircraft(s *capnp.Segment) (Aircraft, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Aircraft{st}, err
}

func ReadRootAircraft(msg *capnp.Message) (Aircraft, error) {
	root, err := msg.RootPtr()
	return Aircraft{root.Struct()}, err
}

func (s Aircraft) String() string {
	str, _ := text.Marshal(0xe54e10aede55c7b1, s.Struct)
	return str
}

func (s Aircraft) Which() Aircraft_Which {
	return Aircraft_Which(s.Struct.Uint16(0))
}
func (s Aircraft) SetVoid() {
	s.Struct.SetUint16(0, 0)

}

func (s Aircraft) B737() (B737, error) {
	p, err := s.Struct.Ptr(0)
	return B737{Struct: p.Struct()}, err
}

func (s Aircraft) HasB737() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Aircraft) SetB737(v B737) error {
	s.Struct.SetUint16(0, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewB737 sets the b737 field to a newly
// allocated B737 struct, preferring placement in s's segment.
func (s Aircraft) NewB737() (B737, error) {
	s.Struct.SetUint16(0, 1)
	ss, err := NewB737(s.Struct.Segment())
	if err != nil {
		return B737{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Aircraft) A320() (A320, error) {
	p, err := s.Struct.Ptr(0)
	return A320{Struct: p.Struct()}, err
}

func (s Aircraft) HasA320() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Aircraft) SetA320(v A320) error {
	s.Struct.SetUint16(0, 2)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewA320 sets the a320 field to a newly
// allocated A320 struct, preferring placement in s's segment.
func (s Aircraft) NewA320() (A320, error) {
	s.Struct.SetUint16(0, 2)
	ss, err := NewA320(s.Struct.Segment())
	if err != nil {
		return A320{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Aircraft) F16() (F16, error) {
	p, err := s.Struct.Ptr(0)
	return F16{Struct: p.Struct()}, err
}

func (s Aircraft) HasF16() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Aircraft) SetF16(v F16) error {
	s.Struct.SetUint16(0, 3)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewF16 sets the f16 field to a newly
// allocated F16 struct, preferring placement in s's segment.
func (s Aircraft) NewF16() (F16, error) {
	s.Struct.SetUint16(0, 3)
	ss, err := NewF16(s.Struct.Segment())
	if err != nil {
		return F16{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Aircraft_List is a list of Aircraft.
type Aircraft_List struct{ capnp.List }

// NewAircraft creates a new list of Aircraft.
func NewAircraft_List(s *capnp.Segment, sz int32) (Aircraft_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Aircraft_List{l}, err
}

func (s Aircraft_List) At(i int) Aircraft { return Aircraft{s.List.Struct(i)} }

func (s Aircraft_List) Set(i int, v Aircraft) error { return s.List.SetStruct(i, v.Struct) }

// Aircraft_Promise is a wrapper for a Aircraft promised by a client call.
type Aircraft_Promise struct{ *capnp.Pipeline }

func (p Aircraft_Promise) Struct() (Aircraft, error) {
	s, err := p.Pipeline.Struct()
	return Aircraft{s}, err
}

func (p Aircraft_Promise) B737() B737_Promise {
	return B737_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Aircraft_Promise) A320() A320_Promise {
	return A320_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Aircraft_Promise) F16() F16_Promise {
	return F16_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Z struct{ capnp.Struct }
type Z_grp Z
type Z_Which uint16

const (
	Z_Which_void        Z_Which = 0
	Z_Which_zz          Z_Which = 1
	Z_Which_f64         Z_Which = 2
	Z_Which_f32         Z_Which = 3
	Z_Which_i64         Z_Which = 4
	Z_Which_i32         Z_Which = 5
	Z_Which_i16         Z_Which = 6
	Z_Which_i8          Z_Which = 7
	Z_Which_u64         Z_Which = 8
	Z_Which_u32         Z_Which = 9
	Z_Which_u16         Z_Which = 10
	Z_Which_u8          Z_Which = 11
	Z_Which_bool        Z_Which = 12
	Z_Which_text        Z_Which = 13
	Z_Which_blob        Z_Which = 14
	Z_Which_f64vec      Z_Which = 15
	Z_Which_f32vec      Z_Which = 16
	Z_Which_i64vec      Z_Which = 17
	Z_Which_i32vec      Z_Which = 18
	Z_Which_i16vec      Z_Which = 19
	Z_Which_i8vec       Z_Which = 20
	Z_Which_u64vec      Z_Which = 21
	Z_Which_u32vec      Z_Which = 22
	Z_Which_u16vec      Z_Which = 23
	Z_Which_u8vec       Z_Which = 24
	Z_Which_boolvec     Z_Which = 39
	Z_Which_datavec     Z_Which = 40
	Z_Which_textvec     Z_Which = 41
	Z_Which_zvec        Z_Which = 25
	Z_Which_zvecvec     Z_Which = 26
	Z_Which_zdate       Z_Which = 27
	Z_Which_zdata       Z_Which = 28
	Z_Which_aircraftvec Z_Which = 29
	Z_Which_aircraft    Z_Which = 30
	Z_Which_regression  Z_Which = 31
	Z_Which_planebase   Z_Which = 32
	Z_Which_airport     Z_Which = 33
	Z_Which_b737        Z_Which = 34
	Z_Which_a320        Z_Which = 35
	Z_Which_f16         Z_Which = 36
	Z_Which_zdatevec    Z_Which = 37
	Z_Which_zdatavec    Z_Which = 38
	Z_Which_grp         Z_Which = 42
)

func (w Z_Which) String() string {
	const s = "voidzzf64f32i64i32i16i8u64u32u16u8booltextblobf64vecf32veci64veci32veci16veci8vecu64vecu32vecu16vecu8vecboolvecdatavectextveczveczvecveczdatezdataaircraftvecaircraftregressionplanebaseairportb737a320f16zdateveczdatavecgrp"
	switch w {
	case Z_Which_void:
		return s[0:4]
	case Z_Which_zz:
		return s[4:6]
	case Z_Which_f64:
		return s[6:9]
	case Z_Which_f32:
		return s[9:12]
	case Z_Which_i64:
		return s[12:15]
	case Z_Which_i32:
		return s[15:18]
	case Z_Which_i16:
		return s[18:21]
	case Z_Which_i8:
		return s[21:23]
	case Z_Which_u64:
		return s[23:26]
	case Z_Which_u32:
		return s[26:29]
	case Z_Which_u16:
		return s[29:32]
	case Z_Which_u8:
		return s[32:34]
	case Z_Which_bool:
		return s[34:38]
	case Z_Which_text:
		return s[38:42]
	case Z_Which_blob:
		return s[42:46]
	case Z_Which_f64vec:
		return s[46:52]
	case Z_Which_f32vec:
		return s[52:58]
	case Z_Which_i64vec:
		return s[58:64]
	case Z_Which_i32vec:
		return s[64:70]
	case Z_Which_i16vec:
		return s[70:76]
	case Z_Which_i8vec:
		return s[76:81]
	case Z_Which_u64vec:
		return s[81:87]
	case Z_Which_u32vec:
		return s[87:93]
	case Z_Which_u16vec:
		return s[93:99]
	case Z_Which_u8vec:
		return s[99:104]
	case Z_Which_boolvec:
		return s[104:111]
	case Z_Which_datavec:
		return s[111:118]
	case Z_Which_textvec:
		return s[118:125]
	case Z_Which_zvec:
		return s[125:129]
	case Z_Which_zvecvec:
		return s[129:136]
	case Z_Which_zdate:
		return s[136:141]
	case Z_Which_zdata:
		return s[141:146]
	case Z_Which_aircraftvec:
		return s[146:157]
	case Z_Which_aircraft:
		return s[157:165]
	case Z_Which_regression:
		return s[165:175]
	case Z_Which_planebase:
		return s[175:184]
	case Z_Which_airport:
		return s[184:191]
	case Z_Which_b737:
		return s[191:195]
	case Z_Which_a320:
		return s[195:199]
	case Z_Which_f16:
		return s[199:202]
	case Z_Which_zdatevec:
		return s[202:210]
	case Z_Which_zdatavec:
		return s[210:218]
	case Z_Which_grp:
		return s[218:221]

	}
	return "Z_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Z_TypeID is the unique identifier for the type Z.
const Z_TypeID = 0xea26e9973bd6a0d9

func NewZ(s *capnp.Segment) (Z, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 1})
	return Z{st}, err
}

func NewRootZ(s *capnp.Segment) (Z, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 1})
	return Z{st}, err
}

func ReadRootZ(msg *capnp.Message) (Z, error) {
	root, err := msg.RootPtr()
	return Z{root.Struct()}, err
}

func (s Z) String() string {
	str, _ := text.Marshal(0xea26e9973bd6a0d9, s.Struct)
	return str
}

func (s Z) Which() Z_Which {
	return Z_Which(s.Struct.Uint16(0))
}
func (s Z) SetVoid() {
	s.Struct.SetUint16(0, 0)

}

func (s Z) Zz() (Z, error) {
	p, err := s.Struct.Ptr(0)
	return Z{Struct: p.Struct()}, err
}

func (s Z) HasZz() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZz(v Z) error {
	s.Struct.SetUint16(0, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewZz sets the zz field to a newly
// allocated Z struct, preferring placement in s's segment.
func (s Z) NewZz() (Z, error) {
	s.Struct.SetUint16(0, 1)
	ss, err := NewZ(s.Struct.Segment())
	if err != nil {
		return Z{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) F64() float64 {
	return math.Float64frombits(s.Struct.Uint64(8))
}

func (s Z) SetF64(v float64) {
	s.Struct.SetUint16(0, 2)
	s.Struct.SetUint64(8, math.Float64bits(v))
}

func (s Z) F32() float32 {
	return math.Float32frombits(s.Struct.Uint32(8))
}

func (s Z) SetF32(v float32) {
	s.Struct.SetUint16(0, 3)
	s.Struct.SetUint32(8, math.Float32bits(v))
}

func (s Z) I64() int64 {
	return int64(s.Struct.Uint64(8))
}

func (s Z) SetI64(v int64) {
	s.Struct.SetUint16(0, 4)
	s.Struct.SetUint64(8, uint64(v))
}

func (s Z) I32() int32 {
	return int32(s.Struct.Uint32(8))
}

func (s Z) SetI32(v int32) {
	s.Struct.SetUint16(0, 5)
	s.Struct.SetUint32(8, uint32(v))
}

func (s Z) I16() int16 {
	return int16(s.Struct.Uint16(8))
}

func (s Z) SetI16(v int16) {
	s.Struct.SetUint16(0, 6)
	s.Struct.SetUint16(8, uint16(v))
}

func (s Z) I8() int8 {
	return int8(s.Struct.Uint8(8))
}

func (s Z) SetI8(v int8) {
	s.Struct.SetUint16(0, 7)
	s.Struct.SetUint8(8, uint8(v))
}

func (s Z) U64() uint64 {
	return s.Struct.Uint64(8)
}

func (s Z) SetU64(v uint64) {
	s.Struct.SetUint16(0, 8)
	s.Struct.SetUint64(8, v)
}

func (s Z) U32() uint32 {
	return s.Struct.Uint32(8)
}

func (s Z) SetU32(v uint32) {
	s.Struct.SetUint16(0, 9)
	s.Struct.SetUint32(8, v)
}

func (s Z) U16() uint16 {
	return s.Struct.Uint16(8)
}

func (s Z) SetU16(v uint16) {
	s.Struct.SetUint16(0, 10)
	s.Struct.SetUint16(8, v)
}

func (s Z) U8() uint8 {
	return s.Struct.Uint8(8)
}

func (s Z) SetU8(v uint8) {
	s.Struct.SetUint16(0, 11)
	s.Struct.SetUint8(8, v)
}

func (s Z) Bool() bool {
	return s.Struct.Bit(64)
}

func (s Z) SetBool(v bool) {
	s.Struct.SetUint16(0, 12)
	s.Struct.SetBit(64, v)
}

func (s Z) Text() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Z) HasText() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) TextBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Z) SetText(v string) error {
	s.Struct.SetUint16(0, 13)
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Z) Blob() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Z) HasBlob() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetBlob(v []byte) error {
	s.Struct.SetUint16(0, 14)
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, d.List.ToPtr())
}

func (s Z) F64vec() (capnp.Float64List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.Float64List{List: p.List()}, err
}

func (s Z) HasF64vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetF64vec(v capnp.Float64List) error {
	s.Struct.SetUint16(0, 15)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewF64vec sets the f64vec field to a newly
// allocated capnp.Float64List, preferring placement in s's segment.
func (s Z) NewF64vec(n int32) (capnp.Float64List, error) {
	s.Struct.SetUint16(0, 15)
	l, err := capnp.NewFloat64List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Float64List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) F32vec() (capnp.Float32List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.Float32List{List: p.List()}, err
}

func (s Z) HasF32vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetF32vec(v capnp.Float32List) error {
	s.Struct.SetUint16(0, 16)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewF32vec sets the f32vec field to a newly
// allocated capnp.Float32List, preferring placement in s's segment.
func (s Z) NewF32vec(n int32) (capnp.Float32List, error) {
	s.Struct.SetUint16(0, 16)
	l, err := capnp.NewFloat32List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Float32List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) I64vec() (capnp.Int64List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.Int64List{List: p.List()}, err
}

func (s Z) HasI64vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetI64vec(v capnp.Int64List) error {
	s.Struct.SetUint16(0, 17)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewI64vec sets the i64vec field to a newly
// allocated capnp.Int64List, preferring placement in s's segment.
func (s Z) NewI64vec(n int32) (capnp.Int64List, error) {
	s.Struct.SetUint16(0, 17)
	l, err := capnp.NewInt64List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Int64List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) I32vec() (capnp.Int32List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.Int32List{List: p.List()}, err
}

func (s Z) HasI32vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetI32vec(v capnp.Int32List) error {
	s.Struct.SetUint16(0, 18)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewI32vec sets the i32vec field to a newly
// allocated capnp.Int32List, preferring placement in s's segment.
func (s Z) NewI32vec(n int32) (capnp.Int32List, error) {
	s.Struct.SetUint16(0, 18)
	l, err := capnp.NewInt32List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Int32List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) I16vec() (capnp.Int16List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.Int16List{List: p.List()}, err
}

func (s Z) HasI16vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetI16vec(v capnp.Int16List) error {
	s.Struct.SetUint16(0, 19)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewI16vec sets the i16vec field to a newly
// allocated capnp.Int16List, preferring placement in s's segment.
func (s Z) NewI16vec(n int32) (capnp.Int16List, error) {
	s.Struct.SetUint16(0, 19)
	l, err := capnp.NewInt16List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Int16List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) I8vec() (capnp.Int8List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.Int8List{List: p.List()}, err
}

func (s Z) HasI8vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetI8vec(v capnp.Int8List) error {
	s.Struct.SetUint16(0, 20)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewI8vec sets the i8vec field to a newly
// allocated capnp.Int8List, preferring placement in s's segment.
func (s Z) NewI8vec(n int32) (capnp.Int8List, error) {
	s.Struct.SetUint16(0, 20)
	l, err := capnp.NewInt8List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Int8List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) U64vec() (capnp.UInt64List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.UInt64List{List: p.List()}, err
}

func (s Z) HasU64vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetU64vec(v capnp.UInt64List) error {
	s.Struct.SetUint16(0, 21)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewU64vec sets the u64vec field to a newly
// allocated capnp.UInt64List, preferring placement in s's segment.
func (s Z) NewU64vec(n int32) (capnp.UInt64List, error) {
	s.Struct.SetUint16(0, 21)
	l, err := capnp.NewUInt64List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.UInt64List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) U32vec() (capnp.UInt32List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.UInt32List{List: p.List()}, err
}

func (s Z) HasU32vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetU32vec(v capnp.UInt32List) error {
	s.Struct.SetUint16(0, 22)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewU32vec sets the u32vec field to a newly
// allocated capnp.UInt32List, preferring placement in s's segment.
func (s Z) NewU32vec(n int32) (capnp.UInt32List, error) {
	s.Struct.SetUint16(0, 22)
	l, err := capnp.NewUInt32List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.UInt32List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) U16vec() (capnp.UInt16List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.UInt16List{List: p.List()}, err
}

func (s Z) HasU16vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetU16vec(v capnp.UInt16List) error {
	s.Struct.SetUint16(0, 23)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewU16vec sets the u16vec field to a newly
// allocated capnp.UInt16List, preferring placement in s's segment.
func (s Z) NewU16vec(n int32) (capnp.UInt16List, error) {
	s.Struct.SetUint16(0, 23)
	l, err := capnp.NewUInt16List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.UInt16List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) U8vec() (capnp.UInt8List, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.UInt8List{List: p.List()}, err
}

func (s Z) HasU8vec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetU8vec(v capnp.UInt8List) error {
	s.Struct.SetUint16(0, 24)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewU8vec sets the u8vec field to a newly
// allocated capnp.UInt8List, preferring placement in s's segment.
func (s Z) NewU8vec(n int32) (capnp.UInt8List, error) {
	s.Struct.SetUint16(0, 24)
	l, err := capnp.NewUInt8List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.UInt8List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Boolvec() (capnp.BitList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.BitList{List: p.List()}, err
}

func (s Z) HasBoolvec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetBoolvec(v capnp.BitList) error {
	s.Struct.SetUint16(0, 39)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewBoolvec sets the boolvec field to a newly
// allocated capnp.BitList, preferring placement in s's segment.
func (s Z) NewBoolvec(n int32) (capnp.BitList, error) {
	s.Struct.SetUint16(0, 39)
	l, err := capnp.NewBitList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.BitList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Datavec() (capnp.DataList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.DataList{List: p.List()}, err
}

func (s Z) HasDatavec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetDatavec(v capnp.DataList) error {
	s.Struct.SetUint16(0, 40)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewDatavec sets the datavec field to a newly
// allocated capnp.DataList, preferring placement in s's segment.
func (s Z) NewDatavec(n int32) (capnp.DataList, error) {
	s.Struct.SetUint16(0, 40)
	l, err := capnp.NewDataList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.DataList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Textvec() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.TextList{List: p.List()}, err
}

func (s Z) HasTextvec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetTextvec(v capnp.TextList) error {
	s.Struct.SetUint16(0, 41)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewTextvec sets the textvec field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s Z) NewTextvec(n int32) (capnp.TextList, error) {
	s.Struct.SetUint16(0, 41)
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Zvec() (Z_List, error) {
	p, err := s.Struct.Ptr(0)
	return Z_List{List: p.List()}, err
}

func (s Z) HasZvec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZvec(v Z_List) error {
	s.Struct.SetUint16(0, 25)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewZvec sets the zvec field to a newly
// allocated Z_List, preferring placement in s's segment.
func (s Z) NewZvec(n int32) (Z_List, error) {
	s.Struct.SetUint16(0, 25)
	l, err := NewZ_List(s.Struct.Segment(), n)
	if err != nil {
		return Z_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Zvecvec() (capnp.PointerList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.PointerList{List: p.List()}, err
}

func (s Z) HasZvecvec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZvecvec(v capnp.PointerList) error {
	s.Struct.SetUint16(0, 26)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewZvecvec sets the zvecvec field to a newly
// allocated capnp.PointerList, preferring placement in s's segment.
func (s Z) NewZvecvec(n int32) (capnp.PointerList, error) {
	s.Struct.SetUint16(0, 26)
	l, err := capnp.NewPointerList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.PointerList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Zdate() (Zdate, error) {
	p, err := s.Struct.Ptr(0)
	return Zdate{Struct: p.Struct()}, err
}

func (s Z) HasZdate() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZdate(v Zdate) error {
	s.Struct.SetUint16(0, 27)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewZdate sets the zdate field to a newly
// allocated Zdate struct, preferring placement in s's segment.
func (s Z) NewZdate() (Zdate, error) {
	s.Struct.SetUint16(0, 27)
	ss, err := NewZdate(s.Struct.Segment())
	if err != nil {
		return Zdate{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) Zdata() (Zdata, error) {
	p, err := s.Struct.Ptr(0)
	return Zdata{Struct: p.Struct()}, err
}

func (s Z) HasZdata() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZdata(v Zdata) error {
	s.Struct.SetUint16(0, 28)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewZdata sets the zdata field to a newly
// allocated Zdata struct, preferring placement in s's segment.
func (s Z) NewZdata() (Zdata, error) {
	s.Struct.SetUint16(0, 28)
	ss, err := NewZdata(s.Struct.Segment())
	if err != nil {
		return Zdata{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) Aircraftvec() (Aircraft_List, error) {
	p, err := s.Struct.Ptr(0)
	return Aircraft_List{List: p.List()}, err
}

func (s Z) HasAircraftvec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetAircraftvec(v Aircraft_List) error {
	s.Struct.SetUint16(0, 29)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewAircraftvec sets the aircraftvec field to a newly
// allocated Aircraft_List, preferring placement in s's segment.
func (s Z) NewAircraftvec(n int32) (Aircraft_List, error) {
	s.Struct.SetUint16(0, 29)
	l, err := NewAircraft_List(s.Struct.Segment(), n)
	if err != nil {
		return Aircraft_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Aircraft() (Aircraft, error) {
	p, err := s.Struct.Ptr(0)
	return Aircraft{Struct: p.Struct()}, err
}

func (s Z) HasAircraft() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetAircraft(v Aircraft) error {
	s.Struct.SetUint16(0, 30)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewAircraft sets the aircraft field to a newly
// allocated Aircraft struct, preferring placement in s's segment.
func (s Z) NewAircraft() (Aircraft, error) {
	s.Struct.SetUint16(0, 30)
	ss, err := NewAircraft(s.Struct.Segment())
	if err != nil {
		return Aircraft{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) Regression() (Regression, error) {
	p, err := s.Struct.Ptr(0)
	return Regression{Struct: p.Struct()}, err
}

func (s Z) HasRegression() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetRegression(v Regression) error {
	s.Struct.SetUint16(0, 31)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewRegression sets the regression field to a newly
// allocated Regression struct, preferring placement in s's segment.
func (s Z) NewRegression() (Regression, error) {
	s.Struct.SetUint16(0, 31)
	ss, err := NewRegression(s.Struct.Segment())
	if err != nil {
		return Regression{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) Planebase() (PlaneBase, error) {
	p, err := s.Struct.Ptr(0)
	return PlaneBase{Struct: p.Struct()}, err
}

func (s Z) HasPlanebase() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetPlanebase(v PlaneBase) error {
	s.Struct.SetUint16(0, 32)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPlanebase sets the planebase field to a newly
// allocated PlaneBase struct, preferring placement in s's segment.
func (s Z) NewPlanebase() (PlaneBase, error) {
	s.Struct.SetUint16(0, 32)
	ss, err := NewPlaneBase(s.Struct.Segment())
	if err != nil {
		return PlaneBase{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) Airport() Airport {
	return Airport(s.Struct.Uint16(8))
}

func (s Z) SetAirport(v Airport) {
	s.Struct.SetUint16(0, 33)
	s.Struct.SetUint16(8, uint16(v))
}

func (s Z) B737() (B737, error) {
	p, err := s.Struct.Ptr(0)
	return B737{Struct: p.Struct()}, err
}

func (s Z) HasB737() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetB737(v B737) error {
	s.Struct.SetUint16(0, 34)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewB737 sets the b737 field to a newly
// allocated B737 struct, preferring placement in s's segment.
func (s Z) NewB737() (B737, error) {
	s.Struct.SetUint16(0, 34)
	ss, err := NewB737(s.Struct.Segment())
	if err != nil {
		return B737{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) A320() (A320, error) {
	p, err := s.Struct.Ptr(0)
	return A320{Struct: p.Struct()}, err
}

func (s Z) HasA320() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetA320(v A320) error {
	s.Struct.SetUint16(0, 35)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewA320 sets the a320 field to a newly
// allocated A320 struct, preferring placement in s's segment.
func (s Z) NewA320() (A320, error) {
	s.Struct.SetUint16(0, 35)
	ss, err := NewA320(s.Struct.Segment())
	if err != nil {
		return A320{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) F16() (F16, error) {
	p, err := s.Struct.Ptr(0)
	return F16{Struct: p.Struct()}, err
}

func (s Z) HasF16() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetF16(v F16) error {
	s.Struct.SetUint16(0, 36)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewF16 sets the f16 field to a newly
// allocated F16 struct, preferring placement in s's segment.
func (s Z) NewF16() (F16, error) {
	s.Struct.SetUint16(0, 36)
	ss, err := NewF16(s.Struct.Segment())
	if err != nil {
		return F16{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Z) Zdatevec() (Zdate_List, error) {
	p, err := s.Struct.Ptr(0)
	return Zdate_List{List: p.List()}, err
}

func (s Z) HasZdatevec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZdatevec(v Zdate_List) error {
	s.Struct.SetUint16(0, 37)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewZdatevec sets the zdatevec field to a newly
// allocated Zdate_List, preferring placement in s's segment.
func (s Z) NewZdatevec(n int32) (Zdate_List, error) {
	s.Struct.SetUint16(0, 37)
	l, err := NewZdate_List(s.Struct.Segment(), n)
	if err != nil {
		return Zdate_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Zdatavec() (Zdata_List, error) {
	p, err := s.Struct.Ptr(0)
	return Zdata_List{List: p.List()}, err
}

func (s Z) HasZdatavec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Z) SetZdatavec(v Zdata_List) error {
	s.Struct.SetUint16(0, 38)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewZdatavec sets the zdatavec field to a newly
// allocated Zdata_List, preferring placement in s's segment.
func (s Z) NewZdatavec(n int32) (Zdata_List, error) {
	s.Struct.SetUint16(0, 38)
	l, err := NewZdata_List(s.Struct.Segment(), n)
	if err != nil {
		return Zdata_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Z) Grp() Z_grp { return Z_grp(s) }

func (s Z) SetGrp() {
	s.Struct.SetUint16(0, 42)
}

func (s Z_grp) First() uint64 {
	return s.Struct.Uint64(8)
}

func (s Z_grp) SetFirst(v uint64) {
	s.Struct.SetUint64(8, v)
}

func (s Z_grp) Second() uint64 {
	return s.Struct.Uint64(16)
}

func (s Z_grp) SetSecond(v uint64) {
	s.Struct.SetUint64(16, v)
}

// Z_List is a list of Z.
type Z_List struct{ capnp.List }

// NewZ creates a new list of Z.
func NewZ_List(s *capnp.Segment, sz int32) (Z_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 24, PointerCount: 1}, sz)
	return Z_List{l}, err
}

func (s Z_List) At(i int) Z { return Z{s.List.Struct(i)} }

func (s Z_List) Set(i int, v Z) error { return s.List.SetStruct(i, v.Struct) }

// Z_Promise is a wrapper for a Z promised by a client call.
type Z_Promise struct{ *capnp.Pipeline }

func (p Z_Promise) Struct() (Z, error) {
	s, err := p.Pipeline.Struct()
	return Z{s}, err
}

func (p Z_Promise) Zz() Z_Promise {
	return Z_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) Zdate() Zdate_Promise {
	return Zdate_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) Zdata() Zdata_Promise {
	return Zdata_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) Aircraft() Aircraft_Promise {
	return Aircraft_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) Regression() Regression_Promise {
	return Regression_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) Planebase() PlaneBase_Promise {
	return PlaneBase_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) B737() B737_Promise {
	return B737_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) A320() A320_Promise {
	return A320_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) F16() F16_Promise {
	return F16_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Z_Promise) Grp() Z_grp_Promise { return Z_grp_Promise{p.Pipeline} }

// Z_grp_Promise is a wrapper for a Z_grp promised by a client call.
type Z_grp_Promise struct{ *capnp.Pipeline }

func (p Z_grp_Promise) Struct() (Z_grp, error) {
	s, err := p.Pipeline.Struct()
	return Z_grp{s}, err
}

type Counter struct{ capnp.Struct }

// Counter_TypeID is the unique identifier for the type Counter.
const Counter_TypeID = 0x8748bc095e10cb5d

func NewCounter(s *capnp.Segment) (Counter, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Counter{st}, err
}

func NewRootCounter(s *capnp.Segment) (Counter, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Counter{st}, err
}

func ReadRootCounter(msg *capnp.Message) (Counter, error) {
	root, err := msg.RootPtr()
	return Counter{root.Struct()}, err
}

func (s Counter) String() string {
	str, _ := text.Marshal(0x8748bc095e10cb5d, s.Struct)
	return str
}

func (s Counter) Size() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s Counter) SetSize(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

func (s Counter) Words() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Counter) HasWords() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Counter) WordsBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Counter) SetWords(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Counter) Wordlist() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(1)
	return capnp.TextList{List: p.List()}, err
}

func (s Counter) HasWordlist() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Counter) SetWordlist(v capnp.TextList) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewWordlist sets the wordlist field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s Counter) NewWordlist(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

// Counter_List is a list of Counter.
type Counter_List struct{ capnp.List }

// NewCounter creates a new list of Counter.
func NewCounter_List(s *capnp.Segment, sz int32) (Counter_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2}, sz)
	return Counter_List{l}, err
}

func (s Counter_List) At(i int) Counter { return Counter{s.List.Struct(i)} }

func (s Counter_List) Set(i int, v Counter) error { return s.List.SetStruct(i, v.Struct) }

// Counter_Promise is a wrapper for a Counter promised by a client call.
type Counter_Promise struct{ *capnp.Pipeline }

func (p Counter_Promise) Struct() (Counter, error) {
	s, err := p.Pipeline.Struct()
	return Counter{s}, err
}

type Bag struct{ capnp.Struct }

// Bag_TypeID is the unique identifier for the type Bag.
const Bag_TypeID = 0xd636fba4f188dabe

func NewBag(s *capnp.Segment) (Bag, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Bag{st}, err
}

func NewRootBag(s *capnp.Segment) (Bag, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Bag{st}, err
}

func ReadRootBag(msg *capnp.Message) (Bag, error) {
	root, err := msg.RootPtr()
	return Bag{root.Struct()}, err
}

func (s Bag) String() string {
	str, _ := text.Marshal(0xd636fba4f188dabe, s.Struct)
	return str
}

func (s Bag) Counter() (Counter, error) {
	p, err := s.Struct.Ptr(0)
	return Counter{Struct: p.Struct()}, err
}

func (s Bag) HasCounter() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Bag) SetCounter(v Counter) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewCounter sets the counter field to a newly
// allocated Counter struct, preferring placement in s's segment.
func (s Bag) NewCounter() (Counter, error) {
	ss, err := NewCounter(s.Struct.Segment())
	if err != nil {
		return Counter{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Bag_List is a list of Bag.
type Bag_List struct{ capnp.List }

// NewBag creates a new list of Bag.
func NewBag_List(s *capnp.Segment, sz int32) (Bag_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Bag_List{l}, err
}

func (s Bag_List) At(i int) Bag { return Bag{s.List.Struct(i)} }

func (s Bag_List) Set(i int, v Bag) error { return s.List.SetStruct(i, v.Struct) }

// Bag_Promise is a wrapper for a Bag promised by a client call.
type Bag_Promise struct{ *capnp.Pipeline }

func (p Bag_Promise) Struct() (Bag, error) {
	s, err := p.Pipeline.Struct()
	return Bag{s}, err
}

func (p Bag_Promise) Counter() Counter_Promise {
	return Counter_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Zserver struct{ capnp.Struct }

// Zserver_TypeID is the unique identifier for the type Zserver.
const Zserver_TypeID = 0xcc4411e60ba9c498

func NewZserver(s *capnp.Segment) (Zserver, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Zserver{st}, err
}

func NewRootZserver(s *capnp.Segment) (Zserver, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Zserver{st}, err
}

func ReadRootZserver(msg *capnp.Message) (Zserver, error) {
	root, err := msg.RootPtr()
	return Zserver{root.Struct()}, err
}

func (s Zserver) String() string {
	str, _ := text.Marshal(0xcc4411e60ba9c498, s.Struct)
	return str
}

func (s Zserver) Waitingjobs() (Zjob_List, error) {
	p, err := s.Struct.Ptr(0)
	return Zjob_List{List: p.List()}, err
}

func (s Zserver) HasWaitingjobs() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Zserver) SetWaitingjobs(v Zjob_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewWaitingjobs sets the waitingjobs field to a newly
// allocated Zjob_List, preferring placement in s's segment.
func (s Zserver) NewWaitingjobs(n int32) (Zjob_List, error) {
	l, err := NewZjob_List(s.Struct.Segment(), n)
	if err != nil {
		return Zjob_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// Zserver_List is a list of Zserver.
type Zserver_List struct{ capnp.List }

// NewZserver creates a new list of Zserver.
func NewZserver_List(s *capnp.Segment, sz int32) (Zserver_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Zserver_List{l}, err
}

func (s Zserver_List) At(i int) Zserver { return Zserver{s.List.Struct(i)} }

func (s Zserver_List) Set(i int, v Zserver) error { return s.List.SetStruct(i, v.Struct) }

// Zserver_Promise is a wrapper for a Zserver promised by a client call.
type Zserver_Promise struct{ *capnp.Pipeline }

func (p Zserver_Promise) Struct() (Zserver, error) {
	s, err := p.Pipeline.Struct()
	return Zserver{s}, err
}

type Zjob struct{ capnp.Struct }

// Zjob_TypeID is the unique identifier for the type Zjob.
const Zjob_TypeID = 0xddd1416669fb7613

func NewZjob(s *capnp.Segment) (Zjob, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Zjob{st}, err
}

func NewRootZjob(s *capnp.Segment) (Zjob, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Zjob{st}, err
}

func ReadRootZjob(msg *capnp.Message) (Zjob, error) {
	root, err := msg.RootPtr()
	return Zjob{root.Struct()}, err
}

func (s Zjob) String() string {
	str, _ := text.Marshal(0xddd1416669fb7613, s.Struct)
	return str
}

func (s Zjob) Cmd() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Zjob) HasCmd() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Zjob) CmdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Zjob) SetCmd(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Zjob) Args() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(1)
	return capnp.TextList{List: p.List()}, err
}

func (s Zjob) HasArgs() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Zjob) SetArgs(v capnp.TextList) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewArgs sets the args field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s Zjob) NewArgs(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

// Zjob_List is a list of Zjob.
type Zjob_List struct{ capnp.List }

// NewZjob creates a new list of Zjob.
func NewZjob_List(s *capnp.Segment, sz int32) (Zjob_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return Zjob_List{l}, err
}

func (s Zjob_List) At(i int) Zjob { return Zjob{s.List.Struct(i)} }

func (s Zjob_List) Set(i int, v Zjob) error { return s.List.SetStruct(i, v.Struct) }

// Zjob_Promise is a wrapper for a Zjob promised by a client call.
type Zjob_Promise struct{ *capnp.Pipeline }

func (p Zjob_Promise) Struct() (Zjob, error) {
	s, err := p.Pipeline.Struct()
	return Zjob{s}, err
}

type VerEmpty struct{ capnp.Struct }

// VerEmpty_TypeID is the unique identifier for the type VerEmpty.
const VerEmpty_TypeID = 0x93c99951eacc72ff

func NewVerEmpty(s *capnp.Segment) (VerEmpty, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return VerEmpty{st}, err
}

func NewRootVerEmpty(s *capnp.Segment) (VerEmpty, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return VerEmpty{st}, err
}

func ReadRootVerEmpty(msg *capnp.Message) (VerEmpty, error) {
	root, err := msg.RootPtr()
	return VerEmpty{root.Struct()}, err
}

func (s VerEmpty) String() string {
	str, _ := text.Marshal(0x93c99951eacc72ff, s.Struct)
	return str
}

// VerEmpty_List is a list of VerEmpty.
type VerEmpty_List struct{ capnp.List }

// NewVerEmpty creates a new list of VerEmpty.
func NewVerEmpty_List(s *capnp.Segment, sz int32) (VerEmpty_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return VerEmpty_List{l}, err
}

func (s VerEmpty_List) At(i int) VerEmpty { return VerEmpty{s.List.Struct(i)} }

func (s VerEmpty_List) Set(i int, v VerEmpty) error { return s.List.SetStruct(i, v.Struct) }

// VerEmpty_Promise is a wrapper for a VerEmpty promised by a client call.
type VerEmpty_Promise struct{ *capnp.Pipeline }

func (p VerEmpty_Promise) Struct() (VerEmpty, error) {
	s, err := p.Pipeline.Struct()
	return VerEmpty{s}, err
}

type VerOneData struct{ capnp.Struct }

// VerOneData_TypeID is the unique identifier for the type VerOneData.
const VerOneData_TypeID = 0xfca3742893be4cde

func NewVerOneData(s *capnp.Segment) (VerOneData, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return VerOneData{st}, err
}

func NewRootVerOneData(s *capnp.Segment) (VerOneData, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return VerOneData{st}, err
}

func ReadRootVerOneData(msg *capnp.Message) (VerOneData, error) {
	root, err := msg.RootPtr()
	return VerOneData{root.Struct()}, err
}

func (s VerOneData) String() string {
	str, _ := text.Marshal(0xfca3742893be4cde, s.Struct)
	return str
}

func (s VerOneData) Val() int16 {
	return int16(s.Struct.Uint16(0))
}

func (s VerOneData) SetVal(v int16) {
	s.Struct.SetUint16(0, uint16(v))
}

// VerOneData_List is a list of VerOneData.
type VerOneData_List struct{ capnp.List }

// NewVerOneData creates a new list of VerOneData.
func NewVerOneData_List(s *capnp.Segment, sz int32) (VerOneData_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return VerOneData_List{l}, err
}

func (s VerOneData_List) At(i int) VerOneData { return VerOneData{s.List.Struct(i)} }

func (s VerOneData_List) Set(i int, v VerOneData) error { return s.List.SetStruct(i, v.Struct) }

// VerOneData_Promise is a wrapper for a VerOneData promised by a client call.
type VerOneData_Promise struct{ *capnp.Pipeline }

func (p VerOneData_Promise) Struct() (VerOneData, error) {
	s, err := p.Pipeline.Struct()
	return VerOneData{s}, err
}

type VerTwoData struct{ capnp.Struct }

// VerTwoData_TypeID is the unique identifier for the type VerTwoData.
const VerTwoData_TypeID = 0xf705dc45c94766fd

func NewVerTwoData(s *capnp.Segment) (VerTwoData, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0})
	return VerTwoData{st}, err
}

func NewRootVerTwoData(s *capnp.Segment) (VerTwoData, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0})
	return VerTwoData{st}, err
}

func ReadRootVerTwoData(msg *capnp.Message) (VerTwoData, error) {
	root, err := msg.RootPtr()
	return VerTwoData{root.Struct()}, err
}

func (s VerTwoData) String() string {
	str, _ := text.Marshal(0xf705dc45c94766fd, s.Struct)
	return str
}

func (s VerTwoData) Val() int16 {
	return int16(s.Struct.Uint16(0))
}

func (s VerTwoData) SetVal(v int16) {
	s.Struct.SetUint16(0, uint16(v))
}

func (s VerTwoData) Duo() int64 {
	return int64(s.Struct.Uint64(8))
}

func (s VerTwoData) SetDuo(v int64) {
	s.Struct.SetUint64(8, uint64(v))
}

// VerTwoData_List is a list of VerTwoData.
type VerTwoData_List struct{ capnp.List }

// NewVerTwoData creates a new list of VerTwoData.
func NewVerTwoData_List(s *capnp.Segment, sz int32) (VerTwoData_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0}, sz)
	return VerTwoData_List{l}, err
}

func (s VerTwoData_List) At(i int) VerTwoData { return VerTwoData{s.List.Struct(i)} }

func (s VerTwoData_List) Set(i int, v VerTwoData) error { return s.List.SetStruct(i, v.Struct) }

// VerTwoData_Promise is a wrapper for a VerTwoData promised by a client call.
type VerTwoData_Promise struct{ *capnp.Pipeline }

func (p VerTwoData_Promise) Struct() (VerTwoData, error) {
	s, err := p.Pipeline.Struct()
	return VerTwoData{s}, err
}

type VerOnePtr struct{ capnp.Struct }

// VerOnePtr_TypeID is the unique identifier for the type VerOnePtr.
const VerOnePtr_TypeID = 0x94bf7df83408218d

func NewVerOnePtr(s *capnp.Segment) (VerOnePtr, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return VerOnePtr{st}, err
}

func NewRootVerOnePtr(s *capnp.Segment) (VerOnePtr, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return VerOnePtr{st}, err
}

func ReadRootVerOnePtr(msg *capnp.Message) (VerOnePtr, error) {
	root, err := msg.RootPtr()
	return VerOnePtr{root.Struct()}, err
}

func (s VerOnePtr) String() string {
	str, _ := text.Marshal(0x94bf7df83408218d, s.Struct)
	return str
}

func (s VerOnePtr) Ptr() (VerOneData, error) {
	p, err := s.Struct.Ptr(0)
	return VerOneData{Struct: p.Struct()}, err
}

func (s VerOnePtr) HasPtr() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VerOnePtr) SetPtr(v VerOneData) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPtr sets the ptr field to a newly
// allocated VerOneData struct, preferring placement in s's segment.
func (s VerOnePtr) NewPtr() (VerOneData, error) {
	ss, err := NewVerOneData(s.Struct.Segment())
	if err != nil {
		return VerOneData{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// VerOnePtr_List is a list of VerOnePtr.
type VerOnePtr_List struct{ capnp.List }

// NewVerOnePtr creates a new list of VerOnePtr.
func NewVerOnePtr_List(s *capnp.Segment, sz int32) (VerOnePtr_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return VerOnePtr_List{l}, err
}

func (s VerOnePtr_List) At(i int) VerOnePtr { return VerOnePtr{s.List.Struct(i)} }

func (s VerOnePtr_List) Set(i int, v VerOnePtr) error { return s.List.SetStruct(i, v.Struct) }

// VerOnePtr_Promise is a wrapper for a VerOnePtr promised by a client call.
type VerOnePtr_Promise struct{ *capnp.Pipeline }

func (p VerOnePtr_Promise) Struct() (VerOnePtr, error) {
	s, err := p.Pipeline.Struct()
	return VerOnePtr{s}, err
}

func (p VerOnePtr_Promise) Ptr() VerOneData_Promise {
	return VerOneData_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type VerTwoPtr struct{ capnp.Struct }

// VerTwoPtr_TypeID is the unique identifier for the type VerTwoPtr.
const VerTwoPtr_TypeID = 0xc95babe3bd394d2d

func NewVerTwoPtr(s *capnp.Segment) (VerTwoPtr, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return VerTwoPtr{st}, err
}

func NewRootVerTwoPtr(s *capnp.Segment) (VerTwoPtr, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return VerTwoPtr{st}, err
}

func ReadRootVerTwoPtr(msg *capnp.Message) (VerTwoPtr, error) {
	root, err := msg.RootPtr()
	return VerTwoPtr{root.Struct()}, err
}

func (s VerTwoPtr) String() string {
	str, _ := text.Marshal(0xc95babe3bd394d2d, s.Struct)
	return str
}

func (s VerTwoPtr) Ptr1() (VerOneData, error) {
	p, err := s.Struct.Ptr(0)
	return VerOneData{Struct: p.Struct()}, err
}

func (s VerTwoPtr) HasPtr1() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VerTwoPtr) SetPtr1(v VerOneData) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPtr1 sets the ptr1 field to a newly
// allocated VerOneData struct, preferring placement in s's segment.
func (s VerTwoPtr) NewPtr1() (VerOneData, error) {
	ss, err := NewVerOneData(s.Struct.Segment())
	if err != nil {
		return VerOneData{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s VerTwoPtr) Ptr2() (VerOneData, error) {
	p, err := s.Struct.Ptr(1)
	return VerOneData{Struct: p.Struct()}, err
}

func (s VerTwoPtr) HasPtr2() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s VerTwoPtr) SetPtr2(v VerOneData) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewPtr2 sets the ptr2 field to a newly
// allocated VerOneData struct, preferring placement in s's segment.
func (s VerTwoPtr) NewPtr2() (VerOneData, error) {
	ss, err := NewVerOneData(s.Struct.Segment())
	if err != nil {
		return VerOneData{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// VerTwoPtr_List is a list of VerTwoPtr.
type VerTwoPtr_List struct{ capnp.List }

// NewVerTwoPtr creates a new list of VerTwoPtr.
func NewVerTwoPtr_List(s *capnp.Segment, sz int32) (VerTwoPtr_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return VerTwoPtr_List{l}, err
}

func (s VerTwoPtr_List) At(i int) VerTwoPtr { return VerTwoPtr{s.List.Struct(i)} }

func (s VerTwoPtr_List) Set(i int, v VerTwoPtr) error { return s.List.SetStruct(i, v.Struct) }

// VerTwoPtr_Promise is a wrapper for a VerTwoPtr promised by a client call.
type VerTwoPtr_Promise struct{ *capnp.Pipeline }

func (p VerTwoPtr_Promise) Struct() (VerTwoPtr, error) {
	s, err := p.Pipeline.Struct()
	return VerTwoPtr{s}, err
}

func (p VerTwoPtr_Promise) Ptr1() VerOneData_Promise {
	return VerOneData_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p VerTwoPtr_Promise) Ptr2() VerOneData_Promise {
	return VerOneData_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type VerTwoDataTwoPtr struct{ capnp.Struct }

// VerTwoDataTwoPtr_TypeID is the unique identifier for the type VerTwoDataTwoPtr.
const VerTwoDataTwoPtr_TypeID = 0xb61ee2ecff34ca73

func NewVerTwoDataTwoPtr(s *capnp.Segment) (VerTwoDataTwoPtr, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 2})
	return VerTwoDataTwoPtr{st}, err
}

func NewRootVerTwoDataTwoPtr(s *capnp.Segment) (VerTwoDataTwoPtr, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 2})
	return VerTwoDataTwoPtr{st}, err
}

func ReadRootVerTwoDataTwoPtr(msg *capnp.Message) (VerTwoDataTwoPtr, error) {
	root, err := msg.RootPtr()
	return VerTwoDataTwoPtr{root.Struct()}, err
}

func (s VerTwoDataTwoPtr) String() string {
	str, _ := text.Marshal(0xb61ee2ecff34ca73, s.Struct)
	return str
}

func (s VerTwoDataTwoPtr) Val() int16 {
	return int16(s.Struct.Uint16(0))
}

func (s VerTwoDataTwoPtr) SetVal(v int16) {
	s.Struct.SetUint16(0, uint16(v))
}

func (s VerTwoDataTwoPtr) Duo() int64 {
	return int64(s.Struct.Uint64(8))
}

func (s VerTwoDataTwoPtr) SetDuo(v int64) {
	s.Struct.SetUint64(8, uint64(v))
}

func (s VerTwoDataTwoPtr) Ptr1() (VerOneData, error) {
	p, err := s.Struct.Ptr(0)
	return VerOneData{Struct: p.Struct()}, err
}

func (s VerTwoDataTwoPtr) HasPtr1() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VerTwoDataTwoPtr) SetPtr1(v VerOneData) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPtr1 sets the ptr1 field to a newly
// allocated VerOneData struct, preferring placement in s's segment.
func (s VerTwoDataTwoPtr) NewPtr1() (VerOneData, error) {
	ss, err := NewVerOneData(s.Struct.Segment())
	if err != nil {
		return VerOneData{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s VerTwoDataTwoPtr) Ptr2() (VerOneData, error) {
	p, err := s.Struct.Ptr(1)
	return VerOneData{Struct: p.Struct()}, err
}

func (s VerTwoDataTwoPtr) HasPtr2() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s VerTwoDataTwoPtr) SetPtr2(v VerOneData) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewPtr2 sets the ptr2 field to a newly
// allocated VerOneData struct, preferring placement in s's segment.
func (s VerTwoDataTwoPtr) NewPtr2() (VerOneData, error) {
	ss, err := NewVerOneData(s.Struct.Segment())
	if err != nil {
		return VerOneData{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// VerTwoDataTwoPtr_List is a list of VerTwoDataTwoPtr.
type VerTwoDataTwoPtr_List struct{ capnp.List }

// NewVerTwoDataTwoPtr creates a new list of VerTwoDataTwoPtr.
func NewVerTwoDataTwoPtr_List(s *capnp.Segment, sz int32) (VerTwoDataTwoPtr_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 2}, sz)
	return VerTwoDataTwoPtr_List{l}, err
}

func (s VerTwoDataTwoPtr_List) At(i int) VerTwoDataTwoPtr { return VerTwoDataTwoPtr{s.List.Struct(i)} }

func (s VerTwoDataTwoPtr_List) Set(i int, v VerTwoDataTwoPtr) error {
	return s.List.SetStruct(i, v.Struct)
}

// VerTwoDataTwoPtr_Promise is a wrapper for a VerTwoDataTwoPtr promised by a client call.
type VerTwoDataTwoPtr_Promise struct{ *capnp.Pipeline }

func (p VerTwoDataTwoPtr_Promise) Struct() (VerTwoDataTwoPtr, error) {
	s, err := p.Pipeline.Struct()
	return VerTwoDataTwoPtr{s}, err
}

func (p VerTwoDataTwoPtr_Promise) Ptr1() VerOneData_Promise {
	return VerOneData_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p VerTwoDataTwoPtr_Promise) Ptr2() VerOneData_Promise {
	return VerOneData_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type HoldsVerEmptyList struct{ capnp.Struct }

// HoldsVerEmptyList_TypeID is the unique identifier for the type HoldsVerEmptyList.
const HoldsVerEmptyList_TypeID = 0xde9ed43cfaa83093

func NewHoldsVerEmptyList(s *capnp.Segment) (HoldsVerEmptyList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerEmptyList{st}, err
}

func NewRootHoldsVerEmptyList(s *capnp.Segment) (HoldsVerEmptyList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerEmptyList{st}, err
}

func ReadRootHoldsVerEmptyList(msg *capnp.Message) (HoldsVerEmptyList, error) {
	root, err := msg.RootPtr()
	return HoldsVerEmptyList{root.Struct()}, err
}

func (s HoldsVerEmptyList) String() string {
	str, _ := text.Marshal(0xde9ed43cfaa83093, s.Struct)
	return str
}

func (s HoldsVerEmptyList) Mylist() (VerEmpty_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerEmpty_List{List: p.List()}, err
}

func (s HoldsVerEmptyList) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerEmptyList) SetMylist(v VerEmpty_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerEmpty_List, preferring placement in s's segment.
func (s HoldsVerEmptyList) NewMylist(n int32) (VerEmpty_List, error) {
	l, err := NewVerEmpty_List(s.Struct.Segment(), n)
	if err != nil {
		return VerEmpty_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerEmptyList_List is a list of HoldsVerEmptyList.
type HoldsVerEmptyList_List struct{ capnp.List }

// NewHoldsVerEmptyList creates a new list of HoldsVerEmptyList.
func NewHoldsVerEmptyList_List(s *capnp.Segment, sz int32) (HoldsVerEmptyList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerEmptyList_List{l}, err
}

func (s HoldsVerEmptyList_List) At(i int) HoldsVerEmptyList {
	return HoldsVerEmptyList{s.List.Struct(i)}
}

func (s HoldsVerEmptyList_List) Set(i int, v HoldsVerEmptyList) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerEmptyList_Promise is a wrapper for a HoldsVerEmptyList promised by a client call.
type HoldsVerEmptyList_Promise struct{ *capnp.Pipeline }

func (p HoldsVerEmptyList_Promise) Struct() (HoldsVerEmptyList, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerEmptyList{s}, err
}

type HoldsVerOneDataList struct{ capnp.Struct }

// HoldsVerOneDataList_TypeID is the unique identifier for the type HoldsVerOneDataList.
const HoldsVerOneDataList_TypeID = 0xabd055422a4d7df1

func NewHoldsVerOneDataList(s *capnp.Segment) (HoldsVerOneDataList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerOneDataList{st}, err
}

func NewRootHoldsVerOneDataList(s *capnp.Segment) (HoldsVerOneDataList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerOneDataList{st}, err
}

func ReadRootHoldsVerOneDataList(msg *capnp.Message) (HoldsVerOneDataList, error) {
	root, err := msg.RootPtr()
	return HoldsVerOneDataList{root.Struct()}, err
}

func (s HoldsVerOneDataList) String() string {
	str, _ := text.Marshal(0xabd055422a4d7df1, s.Struct)
	return str
}

func (s HoldsVerOneDataList) Mylist() (VerOneData_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerOneData_List{List: p.List()}, err
}

func (s HoldsVerOneDataList) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerOneDataList) SetMylist(v VerOneData_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerOneData_List, preferring placement in s's segment.
func (s HoldsVerOneDataList) NewMylist(n int32) (VerOneData_List, error) {
	l, err := NewVerOneData_List(s.Struct.Segment(), n)
	if err != nil {
		return VerOneData_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerOneDataList_List is a list of HoldsVerOneDataList.
type HoldsVerOneDataList_List struct{ capnp.List }

// NewHoldsVerOneDataList creates a new list of HoldsVerOneDataList.
func NewHoldsVerOneDataList_List(s *capnp.Segment, sz int32) (HoldsVerOneDataList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerOneDataList_List{l}, err
}

func (s HoldsVerOneDataList_List) At(i int) HoldsVerOneDataList {
	return HoldsVerOneDataList{s.List.Struct(i)}
}

func (s HoldsVerOneDataList_List) Set(i int, v HoldsVerOneDataList) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerOneDataList_Promise is a wrapper for a HoldsVerOneDataList promised by a client call.
type HoldsVerOneDataList_Promise struct{ *capnp.Pipeline }

func (p HoldsVerOneDataList_Promise) Struct() (HoldsVerOneDataList, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerOneDataList{s}, err
}

type HoldsVerTwoDataList struct{ capnp.Struct }

// HoldsVerTwoDataList_TypeID is the unique identifier for the type HoldsVerTwoDataList.
const HoldsVerTwoDataList_TypeID = 0xcbdc765fd5dff7ba

func NewHoldsVerTwoDataList(s *capnp.Segment) (HoldsVerTwoDataList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoDataList{st}, err
}

func NewRootHoldsVerTwoDataList(s *capnp.Segment) (HoldsVerTwoDataList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoDataList{st}, err
}

func ReadRootHoldsVerTwoDataList(msg *capnp.Message) (HoldsVerTwoDataList, error) {
	root, err := msg.RootPtr()
	return HoldsVerTwoDataList{root.Struct()}, err
}

func (s HoldsVerTwoDataList) String() string {
	str, _ := text.Marshal(0xcbdc765fd5dff7ba, s.Struct)
	return str
}

func (s HoldsVerTwoDataList) Mylist() (VerTwoData_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoData_List{List: p.List()}, err
}

func (s HoldsVerTwoDataList) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerTwoDataList) SetMylist(v VerTwoData_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerTwoData_List, preferring placement in s's segment.
func (s HoldsVerTwoDataList) NewMylist(n int32) (VerTwoData_List, error) {
	l, err := NewVerTwoData_List(s.Struct.Segment(), n)
	if err != nil {
		return VerTwoData_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerTwoDataList_List is a list of HoldsVerTwoDataList.
type HoldsVerTwoDataList_List struct{ capnp.List }

// NewHoldsVerTwoDataList creates a new list of HoldsVerTwoDataList.
func NewHoldsVerTwoDataList_List(s *capnp.Segment, sz int32) (HoldsVerTwoDataList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerTwoDataList_List{l}, err
}

func (s HoldsVerTwoDataList_List) At(i int) HoldsVerTwoDataList {
	return HoldsVerTwoDataList{s.List.Struct(i)}
}

func (s HoldsVerTwoDataList_List) Set(i int, v HoldsVerTwoDataList) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerTwoDataList_Promise is a wrapper for a HoldsVerTwoDataList promised by a client call.
type HoldsVerTwoDataList_Promise struct{ *capnp.Pipeline }

func (p HoldsVerTwoDataList_Promise) Struct() (HoldsVerTwoDataList, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerTwoDataList{s}, err
}

type HoldsVerOnePtrList struct{ capnp.Struct }

// HoldsVerOnePtrList_TypeID is the unique identifier for the type HoldsVerOnePtrList.
const HoldsVerOnePtrList_TypeID = 0xe508a29c83a059f8

func NewHoldsVerOnePtrList(s *capnp.Segment) (HoldsVerOnePtrList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerOnePtrList{st}, err
}

func NewRootHoldsVerOnePtrList(s *capnp.Segment) (HoldsVerOnePtrList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerOnePtrList{st}, err
}

func ReadRootHoldsVerOnePtrList(msg *capnp.Message) (HoldsVerOnePtrList, error) {
	root, err := msg.RootPtr()
	return HoldsVerOnePtrList{root.Struct()}, err
}

func (s HoldsVerOnePtrList) String() string {
	str, _ := text.Marshal(0xe508a29c83a059f8, s.Struct)
	return str
}

func (s HoldsVerOnePtrList) Mylist() (VerOnePtr_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerOnePtr_List{List: p.List()}, err
}

func (s HoldsVerOnePtrList) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerOnePtrList) SetMylist(v VerOnePtr_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerOnePtr_List, preferring placement in s's segment.
func (s HoldsVerOnePtrList) NewMylist(n int32) (VerOnePtr_List, error) {
	l, err := NewVerOnePtr_List(s.Struct.Segment(), n)
	if err != nil {
		return VerOnePtr_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerOnePtrList_List is a list of HoldsVerOnePtrList.
type HoldsVerOnePtrList_List struct{ capnp.List }

// NewHoldsVerOnePtrList creates a new list of HoldsVerOnePtrList.
func NewHoldsVerOnePtrList_List(s *capnp.Segment, sz int32) (HoldsVerOnePtrList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerOnePtrList_List{l}, err
}

func (s HoldsVerOnePtrList_List) At(i int) HoldsVerOnePtrList {
	return HoldsVerOnePtrList{s.List.Struct(i)}
}

func (s HoldsVerOnePtrList_List) Set(i int, v HoldsVerOnePtrList) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerOnePtrList_Promise is a wrapper for a HoldsVerOnePtrList promised by a client call.
type HoldsVerOnePtrList_Promise struct{ *capnp.Pipeline }

func (p HoldsVerOnePtrList_Promise) Struct() (HoldsVerOnePtrList, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerOnePtrList{s}, err
}

type HoldsVerTwoPtrList struct{ capnp.Struct }

// HoldsVerTwoPtrList_TypeID is the unique identifier for the type HoldsVerTwoPtrList.
const HoldsVerTwoPtrList_TypeID = 0xcf9beaca1cc180c8

func NewHoldsVerTwoPtrList(s *capnp.Segment) (HoldsVerTwoPtrList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoPtrList{st}, err
}

func NewRootHoldsVerTwoPtrList(s *capnp.Segment) (HoldsVerTwoPtrList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoPtrList{st}, err
}

func ReadRootHoldsVerTwoPtrList(msg *capnp.Message) (HoldsVerTwoPtrList, error) {
	root, err := msg.RootPtr()
	return HoldsVerTwoPtrList{root.Struct()}, err
}

func (s HoldsVerTwoPtrList) String() string {
	str, _ := text.Marshal(0xcf9beaca1cc180c8, s.Struct)
	return str
}

func (s HoldsVerTwoPtrList) Mylist() (VerTwoPtr_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoPtr_List{List: p.List()}, err
}

func (s HoldsVerTwoPtrList) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerTwoPtrList) SetMylist(v VerTwoPtr_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerTwoPtr_List, preferring placement in s's segment.
func (s HoldsVerTwoPtrList) NewMylist(n int32) (VerTwoPtr_List, error) {
	l, err := NewVerTwoPtr_List(s.Struct.Segment(), n)
	if err != nil {
		return VerTwoPtr_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerTwoPtrList_List is a list of HoldsVerTwoPtrList.
type HoldsVerTwoPtrList_List struct{ capnp.List }

// NewHoldsVerTwoPtrList creates a new list of HoldsVerTwoPtrList.
func NewHoldsVerTwoPtrList_List(s *capnp.Segment, sz int32) (HoldsVerTwoPtrList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerTwoPtrList_List{l}, err
}

func (s HoldsVerTwoPtrList_List) At(i int) HoldsVerTwoPtrList {
	return HoldsVerTwoPtrList{s.List.Struct(i)}
}

func (s HoldsVerTwoPtrList_List) Set(i int, v HoldsVerTwoPtrList) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerTwoPtrList_Promise is a wrapper for a HoldsVerTwoPtrList promised by a client call.
type HoldsVerTwoPtrList_Promise struct{ *capnp.Pipeline }

func (p HoldsVerTwoPtrList_Promise) Struct() (HoldsVerTwoPtrList, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerTwoPtrList{s}, err
}

type HoldsVerTwoTwoList struct{ capnp.Struct }

// HoldsVerTwoTwoList_TypeID is the unique identifier for the type HoldsVerTwoTwoList.
const HoldsVerTwoTwoList_TypeID = 0x95befe3f14606e6b

func NewHoldsVerTwoTwoList(s *capnp.Segment) (HoldsVerTwoTwoList, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoTwoList{st}, err
}

func NewRootHoldsVerTwoTwoList(s *capnp.Segment) (HoldsVerTwoTwoList, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoTwoList{st}, err
}

func ReadRootHoldsVerTwoTwoList(msg *capnp.Message) (HoldsVerTwoTwoList, error) {
	root, err := msg.RootPtr()
	return HoldsVerTwoTwoList{root.Struct()}, err
}

func (s HoldsVerTwoTwoList) String() string {
	str, _ := text.Marshal(0x95befe3f14606e6b, s.Struct)
	return str
}

func (s HoldsVerTwoTwoList) Mylist() (VerTwoDataTwoPtr_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoDataTwoPtr_List{List: p.List()}, err
}

func (s HoldsVerTwoTwoList) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerTwoTwoList) SetMylist(v VerTwoDataTwoPtr_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerTwoDataTwoPtr_List, preferring placement in s's segment.
func (s HoldsVerTwoTwoList) NewMylist(n int32) (VerTwoDataTwoPtr_List, error) {
	l, err := NewVerTwoDataTwoPtr_List(s.Struct.Segment(), n)
	if err != nil {
		return VerTwoDataTwoPtr_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerTwoTwoList_List is a list of HoldsVerTwoTwoList.
type HoldsVerTwoTwoList_List struct{ capnp.List }

// NewHoldsVerTwoTwoList creates a new list of HoldsVerTwoTwoList.
func NewHoldsVerTwoTwoList_List(s *capnp.Segment, sz int32) (HoldsVerTwoTwoList_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerTwoTwoList_List{l}, err
}

func (s HoldsVerTwoTwoList_List) At(i int) HoldsVerTwoTwoList {
	return HoldsVerTwoTwoList{s.List.Struct(i)}
}

func (s HoldsVerTwoTwoList_List) Set(i int, v HoldsVerTwoTwoList) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerTwoTwoList_Promise is a wrapper for a HoldsVerTwoTwoList promised by a client call.
type HoldsVerTwoTwoList_Promise struct{ *capnp.Pipeline }

func (p HoldsVerTwoTwoList_Promise) Struct() (HoldsVerTwoTwoList, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerTwoTwoList{s}, err
}

type HoldsVerTwoTwoPlus struct{ capnp.Struct }

// HoldsVerTwoTwoPlus_TypeID is the unique identifier for the type HoldsVerTwoTwoPlus.
const HoldsVerTwoTwoPlus_TypeID = 0x87c33f2330feb3d8

func NewHoldsVerTwoTwoPlus(s *capnp.Segment) (HoldsVerTwoTwoPlus, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoTwoPlus{st}, err
}

func NewRootHoldsVerTwoTwoPlus(s *capnp.Segment) (HoldsVerTwoTwoPlus, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HoldsVerTwoTwoPlus{st}, err
}

func ReadRootHoldsVerTwoTwoPlus(msg *capnp.Message) (HoldsVerTwoTwoPlus, error) {
	root, err := msg.RootPtr()
	return HoldsVerTwoTwoPlus{root.Struct()}, err
}

func (s HoldsVerTwoTwoPlus) String() string {
	str, _ := text.Marshal(0x87c33f2330feb3d8, s.Struct)
	return str
}

func (s HoldsVerTwoTwoPlus) Mylist() (VerTwoTwoPlus_List, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoTwoPlus_List{List: p.List()}, err
}

func (s HoldsVerTwoTwoPlus) HasMylist() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsVerTwoTwoPlus) SetMylist(v VerTwoTwoPlus_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMylist sets the mylist field to a newly
// allocated VerTwoTwoPlus_List, preferring placement in s's segment.
func (s HoldsVerTwoTwoPlus) NewMylist(n int32) (VerTwoTwoPlus_List, error) {
	l, err := NewVerTwoTwoPlus_List(s.Struct.Segment(), n)
	if err != nil {
		return VerTwoTwoPlus_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// HoldsVerTwoTwoPlus_List is a list of HoldsVerTwoTwoPlus.
type HoldsVerTwoTwoPlus_List struct{ capnp.List }

// NewHoldsVerTwoTwoPlus creates a new list of HoldsVerTwoTwoPlus.
func NewHoldsVerTwoTwoPlus_List(s *capnp.Segment, sz int32) (HoldsVerTwoTwoPlus_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HoldsVerTwoTwoPlus_List{l}, err
}

func (s HoldsVerTwoTwoPlus_List) At(i int) HoldsVerTwoTwoPlus {
	return HoldsVerTwoTwoPlus{s.List.Struct(i)}
}

func (s HoldsVerTwoTwoPlus_List) Set(i int, v HoldsVerTwoTwoPlus) error {
	return s.List.SetStruct(i, v.Struct)
}

// HoldsVerTwoTwoPlus_Promise is a wrapper for a HoldsVerTwoTwoPlus promised by a client call.
type HoldsVerTwoTwoPlus_Promise struct{ *capnp.Pipeline }

func (p HoldsVerTwoTwoPlus_Promise) Struct() (HoldsVerTwoTwoPlus, error) {
	s, err := p.Pipeline.Struct()
	return HoldsVerTwoTwoPlus{s}, err
}

type VerTwoTwoPlus struct{ capnp.Struct }

// VerTwoTwoPlus_TypeID is the unique identifier for the type VerTwoTwoPlus.
const VerTwoTwoPlus_TypeID = 0xce44aee2d9e25049

func NewVerTwoTwoPlus(s *capnp.Segment) (VerTwoTwoPlus, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3})
	return VerTwoTwoPlus{st}, err
}

func NewRootVerTwoTwoPlus(s *capnp.Segment) (VerTwoTwoPlus, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3})
	return VerTwoTwoPlus{st}, err
}

func ReadRootVerTwoTwoPlus(msg *capnp.Message) (VerTwoTwoPlus, error) {
	root, err := msg.RootPtr()
	return VerTwoTwoPlus{root.Struct()}, err
}

func (s VerTwoTwoPlus) String() string {
	str, _ := text.Marshal(0xce44aee2d9e25049, s.Struct)
	return str
}

func (s VerTwoTwoPlus) Val() int16 {
	return int16(s.Struct.Uint16(0))
}

func (s VerTwoTwoPlus) SetVal(v int16) {
	s.Struct.SetUint16(0, uint16(v))
}

func (s VerTwoTwoPlus) Duo() int64 {
	return int64(s.Struct.Uint64(8))
}

func (s VerTwoTwoPlus) SetDuo(v int64) {
	s.Struct.SetUint64(8, uint64(v))
}

func (s VerTwoTwoPlus) Ptr1() (VerTwoDataTwoPtr, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoDataTwoPtr{Struct: p.Struct()}, err
}

func (s VerTwoTwoPlus) HasPtr1() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VerTwoTwoPlus) SetPtr1(v VerTwoDataTwoPtr) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPtr1 sets the ptr1 field to a newly
// allocated VerTwoDataTwoPtr struct, preferring placement in s's segment.
func (s VerTwoTwoPlus) NewPtr1() (VerTwoDataTwoPtr, error) {
	ss, err := NewVerTwoDataTwoPtr(s.Struct.Segment())
	if err != nil {
		return VerTwoDataTwoPtr{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s VerTwoTwoPlus) Ptr2() (VerTwoDataTwoPtr, error) {
	p, err := s.Struct.Ptr(1)
	return VerTwoDataTwoPtr{Struct: p.Struct()}, err
}

func (s VerTwoTwoPlus) HasPtr2() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s VerTwoTwoPlus) SetPtr2(v VerTwoDataTwoPtr) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewPtr2 sets the ptr2 field to a newly
// allocated VerTwoDataTwoPtr struct, preferring placement in s's segment.
func (s VerTwoTwoPlus) NewPtr2() (VerTwoDataTwoPtr, error) {
	ss, err := NewVerTwoDataTwoPtr(s.Struct.Segment())
	if err != nil {
		return VerTwoDataTwoPtr{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s VerTwoTwoPlus) Tre() int64 {
	return int64(s.Struct.Uint64(16))
}

func (s VerTwoTwoPlus) SetTre(v int64) {
	s.Struct.SetUint64(16, uint64(v))
}

func (s VerTwoTwoPlus) Lst3() (capnp.Int64List, error) {
	p, err := s.Struct.Ptr(2)
	return capnp.Int64List{List: p.List()}, err
}

func (s VerTwoTwoPlus) HasLst3() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s VerTwoTwoPlus) SetLst3(v capnp.Int64List) error {
	return s.Struct.SetPtr(2, v.List.ToPtr())
}

// NewLst3 sets the lst3 field to a newly
// allocated capnp.Int64List, preferring placement in s's segment.
func (s VerTwoTwoPlus) NewLst3(n int32) (capnp.Int64List, error) {
	l, err := capnp.NewInt64List(s.Struct.Segment(), n)
	if err != nil {
		return capnp.Int64List{}, err
	}
	err = s.Struct.SetPtr(2, l.List.ToPtr())
	return l, err
}

// VerTwoTwoPlus_List is a list of VerTwoTwoPlus.
type VerTwoTwoPlus_List struct{ capnp.List }

// NewVerTwoTwoPlus creates a new list of VerTwoTwoPlus.
func NewVerTwoTwoPlus_List(s *capnp.Segment, sz int32) (VerTwoTwoPlus_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3}, sz)
	return VerTwoTwoPlus_List{l}, err
}

func (s VerTwoTwoPlus_List) At(i int) VerTwoTwoPlus { return VerTwoTwoPlus{s.List.Struct(i)} }

func (s VerTwoTwoPlus_List) Set(i int, v VerTwoTwoPlus) error { return s.List.SetStruct(i, v.Struct) }

// VerTwoTwoPlus_Promise is a wrapper for a VerTwoTwoPlus promised by a client call.
type VerTwoTwoPlus_Promise struct{ *capnp.Pipeline }

func (p VerTwoTwoPlus_Promise) Struct() (VerTwoTwoPlus, error) {
	s, err := p.Pipeline.Struct()
	return VerTwoTwoPlus{s}, err
}

func (p VerTwoTwoPlus_Promise) Ptr1() VerTwoDataTwoPtr_Promise {
	return VerTwoDataTwoPtr_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p VerTwoTwoPlus_Promise) Ptr2() VerTwoDataTwoPtr_Promise {
	return VerTwoDataTwoPtr_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type HoldsText struct{ capnp.Struct }

// HoldsText_TypeID is the unique identifier for the type HoldsText.
const HoldsText_TypeID = 0xe5817f849ff906dc

func NewHoldsText(s *capnp.Segment) (HoldsText, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return HoldsText{st}, err
}

func NewRootHoldsText(s *capnp.Segment) (HoldsText, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return HoldsText{st}, err
}

func ReadRootHoldsText(msg *capnp.Message) (HoldsText, error) {
	root, err := msg.RootPtr()
	return HoldsText{root.Struct()}, err
}

func (s HoldsText) String() string {
	str, _ := text.Marshal(0xe5817f849ff906dc, s.Struct)
	return str
}

func (s HoldsText) Txt() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s HoldsText) HasTxt() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HoldsText) TxtBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s HoldsText) SetTxt(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s HoldsText) Lst() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(1)
	return capnp.TextList{List: p.List()}, err
}

func (s HoldsText) HasLst() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s HoldsText) SetLst(v capnp.TextList) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewLst sets the lst field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s HoldsText) NewLst(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

func (s HoldsText) Lstlst() (capnp.PointerList, error) {
	p, err := s.Struct.Ptr(2)
	return capnp.PointerList{List: p.List()}, err
}

func (s HoldsText) HasLstlst() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s HoldsText) SetLstlst(v capnp.PointerList) error {
	return s.Struct.SetPtr(2, v.List.ToPtr())
}

// NewLstlst sets the lstlst field to a newly
// allocated capnp.PointerList, preferring placement in s's segment.
func (s HoldsText) NewLstlst(n int32) (capnp.PointerList, error) {
	l, err := capnp.NewPointerList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.PointerList{}, err
	}
	err = s.Struct.SetPtr(2, l.List.ToPtr())
	return l, err
}

// HoldsText_List is a list of HoldsText.
type HoldsText_List struct{ capnp.List }

// NewHoldsText creates a new list of HoldsText.
func NewHoldsText_List(s *capnp.Segment, sz int32) (HoldsText_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return HoldsText_List{l}, err
}

func (s HoldsText_List) At(i int) HoldsText { return HoldsText{s.List.Struct(i)} }

func (s HoldsText_List) Set(i int, v HoldsText) error { return s.List.SetStruct(i, v.Struct) }

// HoldsText_Promise is a wrapper for a HoldsText promised by a client call.
type HoldsText_Promise struct{ *capnp.Pipeline }

func (p HoldsText_Promise) Struct() (HoldsText, error) {
	s, err := p.Pipeline.Struct()
	return HoldsText{s}, err
}

type WrapEmpty struct{ capnp.Struct }

// WrapEmpty_TypeID is the unique identifier for the type WrapEmpty.
const WrapEmpty_TypeID = 0x9ab599979b02ac59

func NewWrapEmpty(s *capnp.Segment) (WrapEmpty, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return WrapEmpty{st}, err
}

func NewRootWrapEmpty(s *capnp.Segment) (WrapEmpty, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return WrapEmpty{st}, err
}

func ReadRootWrapEmpty(msg *capnp.Message) (WrapEmpty, error) {
	root, err := msg.RootPtr()
	return WrapEmpty{root.Struct()}, err
}

func (s WrapEmpty) String() string {
	str, _ := text.Marshal(0x9ab599979b02ac59, s.Struct)
	return str
}

func (s WrapEmpty) MightNotBeReallyEmpty() (VerEmpty, error) {
	p, err := s.Struct.Ptr(0)
	return VerEmpty{Struct: p.Struct()}, err
}

func (s WrapEmpty) HasMightNotBeReallyEmpty() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s WrapEmpty) SetMightNotBeReallyEmpty(v VerEmpty) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewMightNotBeReallyEmpty sets the mightNotBeReallyEmpty field to a newly
// allocated VerEmpty struct, preferring placement in s's segment.
func (s WrapEmpty) NewMightNotBeReallyEmpty() (VerEmpty, error) {
	ss, err := NewVerEmpty(s.Struct.Segment())
	if err != nil {
		return VerEmpty{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// WrapEmpty_List is a list of WrapEmpty.
type WrapEmpty_List struct{ capnp.List }

// NewWrapEmpty creates a new list of WrapEmpty.
func NewWrapEmpty_List(s *capnp.Segment, sz int32) (WrapEmpty_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return WrapEmpty_List{l}, err
}

func (s WrapEmpty_List) At(i int) WrapEmpty { return WrapEmpty{s.List.Struct(i)} }

func (s WrapEmpty_List) Set(i int, v WrapEmpty) error { return s.List.SetStruct(i, v.Struct) }

// WrapEmpty_Promise is a wrapper for a WrapEmpty promised by a client call.
type WrapEmpty_Promise struct{ *capnp.Pipeline }

func (p WrapEmpty_Promise) Struct() (WrapEmpty, error) {
	s, err := p.Pipeline.Struct()
	return WrapEmpty{s}, err
}

func (p WrapEmpty_Promise) MightNotBeReallyEmpty() VerEmpty_Promise {
	return VerEmpty_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Wrap2x2 struct{ capnp.Struct }

// Wrap2x2_TypeID is the unique identifier for the type Wrap2x2.
const Wrap2x2_TypeID = 0xe1a2d1d51107bead

func NewWrap2x2(s *capnp.Segment) (Wrap2x2, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Wrap2x2{st}, err
}

func NewRootWrap2x2(s *capnp.Segment) (Wrap2x2, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Wrap2x2{st}, err
}

func ReadRootWrap2x2(msg *capnp.Message) (Wrap2x2, error) {
	root, err := msg.RootPtr()
	return Wrap2x2{root.Struct()}, err
}

func (s Wrap2x2) String() string {
	str, _ := text.Marshal(0xe1a2d1d51107bead, s.Struct)
	return str
}

func (s Wrap2x2) MightNotBeReallyEmpty() (VerTwoDataTwoPtr, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoDataTwoPtr{Struct: p.Struct()}, err
}

func (s Wrap2x2) HasMightNotBeReallyEmpty() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Wrap2x2) SetMightNotBeReallyEmpty(v VerTwoDataTwoPtr) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewMightNotBeReallyEmpty sets the mightNotBeReallyEmpty field to a newly
// allocated VerTwoDataTwoPtr struct, preferring placement in s's segment.
func (s Wrap2x2) NewMightNotBeReallyEmpty() (VerTwoDataTwoPtr, error) {
	ss, err := NewVerTwoDataTwoPtr(s.Struct.Segment())
	if err != nil {
		return VerTwoDataTwoPtr{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Wrap2x2_List is a list of Wrap2x2.
type Wrap2x2_List struct{ capnp.List }

// NewWrap2x2 creates a new list of Wrap2x2.
func NewWrap2x2_List(s *capnp.Segment, sz int32) (Wrap2x2_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Wrap2x2_List{l}, err
}

func (s Wrap2x2_List) At(i int) Wrap2x2 { return Wrap2x2{s.List.Struct(i)} }

func (s Wrap2x2_List) Set(i int, v Wrap2x2) error { return s.List.SetStruct(i, v.Struct) }

// Wrap2x2_Promise is a wrapper for a Wrap2x2 promised by a client call.
type Wrap2x2_Promise struct{ *capnp.Pipeline }

func (p Wrap2x2_Promise) Struct() (Wrap2x2, error) {
	s, err := p.Pipeline.Struct()
	return Wrap2x2{s}, err
}

func (p Wrap2x2_Promise) MightNotBeReallyEmpty() VerTwoDataTwoPtr_Promise {
	return VerTwoDataTwoPtr_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Wrap2x2plus struct{ capnp.Struct }

// Wrap2x2plus_TypeID is the unique identifier for the type Wrap2x2plus.
const Wrap2x2plus_TypeID = 0xe684eb3aef1a6859

func NewWrap2x2plus(s *capnp.Segment) (Wrap2x2plus, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Wrap2x2plus{st}, err
}

func NewRootWrap2x2plus(s *capnp.Segment) (Wrap2x2plus, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Wrap2x2plus{st}, err
}

func ReadRootWrap2x2plus(msg *capnp.Message) (Wrap2x2plus, error) {
	root, err := msg.RootPtr()
	return Wrap2x2plus{root.Struct()}, err
}

func (s Wrap2x2plus) String() string {
	str, _ := text.Marshal(0xe684eb3aef1a6859, s.Struct)
	return str
}

func (s Wrap2x2plus) MightNotBeReallyEmpty() (VerTwoTwoPlus, error) {
	p, err := s.Struct.Ptr(0)
	return VerTwoTwoPlus{Struct: p.Struct()}, err
}

func (s Wrap2x2plus) HasMightNotBeReallyEmpty() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Wrap2x2plus) SetMightNotBeReallyEmpty(v VerTwoTwoPlus) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewMightNotBeReallyEmpty sets the mightNotBeReallyEmpty field to a newly
// allocated VerTwoTwoPlus struct, preferring placement in s's segment.
func (s Wrap2x2plus) NewMightNotBeReallyEmpty() (VerTwoTwoPlus, error) {
	ss, err := NewVerTwoTwoPlus(s.Struct.Segment())
	if err != nil {
		return VerTwoTwoPlus{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Wrap2x2plus_List is a list of Wrap2x2plus.
type Wrap2x2plus_List struct{ capnp.List }

// NewWrap2x2plus creates a new list of Wrap2x2plus.
func NewWrap2x2plus_List(s *capnp.Segment, sz int32) (Wrap2x2plus_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Wrap2x2plus_List{l}, err
}

func (s Wrap2x2plus_List) At(i int) Wrap2x2plus { return Wrap2x2plus{s.List.Struct(i)} }

func (s Wrap2x2plus_List) Set(i int, v Wrap2x2plus) error { return s.List.SetStruct(i, v.Struct) }

// Wrap2x2plus_Promise is a wrapper for a Wrap2x2plus promised by a client call.
type Wrap2x2plus_Promise struct{ *capnp.Pipeline }

func (p Wrap2x2plus_Promise) Struct() (Wrap2x2plus, error) {
	s, err := p.Pipeline.Struct()
	return Wrap2x2plus{s}, err
}

func (p Wrap2x2plus_Promise) MightNotBeReallyEmpty() VerTwoTwoPlus_Promise {
	return VerTwoTwoPlus_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type VoidUnion struct{ capnp.Struct }
type VoidUnion_Which uint16

const (
	VoidUnion_Which_a VoidUnion_Which = 0
	VoidUnion_Which_b VoidUnion_Which = 1
)

func (w VoidUnion_Which) String() string {
	const s = "ab"
	switch w {
	case VoidUnion_Which_a:
		return s[0:1]
	case VoidUnion_Which_b:
		return s[1:2]

	}
	return "VoidUnion_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// VoidUnion_TypeID is the unique identifier for the type VoidUnion.
const VoidUnion_TypeID = 0x8821cdb23640783a

func NewVoidUnion(s *capnp.Segment) (VoidUnion, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return VoidUnion{st}, err
}

func NewRootVoidUnion(s *capnp.Segment) (VoidUnion, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return VoidUnion{st}, err
}

func ReadRootVoidUnion(msg *capnp.Message) (VoidUnion, error) {
	root, err := msg.RootPtr()
	return VoidUnion{root.Struct()}, err
}

func (s VoidUnion) String() string {
	str, _ := text.Marshal(0x8821cdb23640783a, s.Struct)
	return str
}

func (s VoidUnion) Which() VoidUnion_Which {
	return VoidUnion_Which(s.Struct.Uint16(0))
}
func (s VoidUnion) SetA() {
	s.Struct.SetUint16(0, 0)

}

func (s VoidUnion) SetB() {
	s.Struct.SetUint16(0, 1)

}

// VoidUnion_List is a list of VoidUnion.
type VoidUnion_List struct{ capnp.List }

// NewVoidUnion creates a new list of VoidUnion.
func NewVoidUnion_List(s *capnp.Segment, sz int32) (VoidUnion_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return VoidUnion_List{l}, err
}

func (s VoidUnion_List) At(i int) VoidUnion { return VoidUnion{s.List.Struct(i)} }

func (s VoidUnion_List) Set(i int, v VoidUnion) error { return s.List.SetStruct(i, v.Struct) }

// VoidUnion_Promise is a wrapper for a VoidUnion promised by a client call.
type VoidUnion_Promise struct{ *capnp.Pipeline }

func (p VoidUnion_Promise) Struct() (VoidUnion, error) {
	s, err := p.Pipeline.Struct()
	return VoidUnion{s}, err
}

type Nester1Capn struct{ capnp.Struct }

// Nester1Capn_TypeID is the unique identifier for the type Nester1Capn.
const Nester1Capn_TypeID = 0xf14fad09425d081c

func NewNester1Capn(s *capnp.Segment) (Nester1Capn, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Nester1Capn{st}, err
}

func NewRootNester1Capn(s *capnp.Segment) (Nester1Capn, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Nester1Capn{st}, err
}

func ReadRootNester1Capn(msg *capnp.Message) (Nester1Capn, error) {
	root, err := msg.RootPtr()
	return Nester1Capn{root.Struct()}, err
}

func (s Nester1Capn) String() string {
	str, _ := text.Marshal(0xf14fad09425d081c, s.Struct)
	return str
}

func (s Nester1Capn) Strs() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.TextList{List: p.List()}, err
}

func (s Nester1Capn) HasStrs() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Nester1Capn) SetStrs(v capnp.TextList) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewStrs sets the strs field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s Nester1Capn) NewStrs(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// Nester1Capn_List is a list of Nester1Capn.
type Nester1Capn_List struct{ capnp.List }

// NewNester1Capn creates a new list of Nester1Capn.
func NewNester1Capn_List(s *capnp.Segment, sz int32) (Nester1Capn_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Nester1Capn_List{l}, err
}

func (s Nester1Capn_List) At(i int) Nester1Capn { return Nester1Capn{s.List.Struct(i)} }

func (s Nester1Capn_List) Set(i int, v Nester1Capn) error { return s.List.SetStruct(i, v.Struct) }

// Nester1Capn_Promise is a wrapper for a Nester1Capn promised by a client call.
type Nester1Capn_Promise struct{ *capnp.Pipeline }

func (p Nester1Capn_Promise) Struct() (Nester1Capn, error) {
	s, err := p.Pipeline.Struct()
	return Nester1Capn{s}, err
}

type RWTestCapn struct{ capnp.Struct }

// RWTestCapn_TypeID is the unique identifier for the type RWTestCapn.
const RWTestCapn_TypeID = 0xf7ff4414476c186a

func NewRWTestCapn(s *capnp.Segment) (RWTestCapn, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return RWTestCapn{st}, err
}

func NewRootRWTestCapn(s *capnp.Segment) (RWTestCapn, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return RWTestCapn{st}, err
}

func ReadRootRWTestCapn(msg *capnp.Message) (RWTestCapn, error) {
	root, err := msg.RootPtr()
	return RWTestCapn{root.Struct()}, err
}

func (s RWTestCapn) String() string {
	str, _ := text.Marshal(0xf7ff4414476c186a, s.Struct)
	return str
}

func (s RWTestCapn) NestMatrix() (capnp.PointerList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.PointerList{List: p.List()}, err
}

func (s RWTestCapn) HasNestMatrix() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s RWTestCapn) SetNestMatrix(v capnp.PointerList) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewNestMatrix sets the nestMatrix field to a newly
// allocated capnp.PointerList, preferring placement in s's segment.
func (s RWTestCapn) NewNestMatrix(n int32) (capnp.PointerList, error) {
	l, err := capnp.NewPointerList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.PointerList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// RWTestCapn_List is a list of RWTestCapn.
type RWTestCapn_List struct{ capnp.List }

// NewRWTestCapn creates a new list of RWTestCapn.
func NewRWTestCapn_List(s *capnp.Segment, sz int32) (RWTestCapn_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return RWTestCapn_List{l}, err
}

func (s RWTestCapn_List) At(i int) RWTestCapn { return RWTestCapn{s.List.Struct(i)} }

func (s RWTestCapn_List) Set(i int, v RWTestCapn) error { return s.List.SetStruct(i, v.Struct) }

// RWTestCapn_Promise is a wrapper for a RWTestCapn promised by a client call.
type RWTestCapn_Promise struct{ *capnp.Pipeline }

func (p RWTestCapn_Promise) Struct() (RWTestCapn, error) {
	s, err := p.Pipeline.Struct()
	return RWTestCapn{s}, err
}

type ListStructCapn struct{ capnp.Struct }

// ListStructCapn_TypeID is the unique identifier for the type ListStructCapn.
const ListStructCapn_TypeID = 0xb1ac056ed7647011

func NewListStructCapn(s *capnp.Segment) (ListStructCapn, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ListStructCapn{st}, err
}

func NewRootListStructCapn(s *capnp.Segment) (ListStructCapn, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return ListStructCapn{st}, err
}

func ReadRootListStructCapn(msg *capnp.Message) (ListStructCapn, error) {
	root, err := msg.RootPtr()
	return ListStructCapn{root.Struct()}, err
}

func (s ListStructCapn) String() string {
	str, _ := text.Marshal(0xb1ac056ed7647011, s.Struct)
	return str
}

func (s ListStructCapn) Vec() (Nester1Capn_List, error) {
	p, err := s.Struct.Ptr(0)
	return Nester1Capn_List{List: p.List()}, err
}

func (s ListStructCapn) HasVec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s ListStructCapn) SetVec(v Nester1Capn_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewVec sets the vec field to a newly
// allocated Nester1Capn_List, preferring placement in s's segment.
func (s ListStructCapn) NewVec(n int32) (Nester1Capn_List, error) {
	l, err := NewNester1Capn_List(s.Struct.Segment(), n)
	if err != nil {
		return Nester1Capn_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// ListStructCapn_List is a list of ListStructCapn.
type ListStructCapn_List struct{ capnp.List }

// NewListStructCapn creates a new list of ListStructCapn.
func NewListStructCapn_List(s *capnp.Segment, sz int32) (ListStructCapn_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return ListStructCapn_List{l}, err
}

func (s ListStructCapn_List) At(i int) ListStructCapn { return ListStructCapn{s.List.Struct(i)} }

func (s ListStructCapn_List) Set(i int, v ListStructCapn) error { return s.List.SetStruct(i, v.Struct) }

// ListStructCapn_Promise is a wrapper for a ListStructCapn promised by a client call.
type ListStructCapn_Promise struct{ *capnp.Pipeline }

func (p ListStructCapn_Promise) Struct() (ListStructCapn, error) {
	s, err := p.Pipeline.Struct()
	return ListStructCapn{s}, err
}

type Echo struct{ Client capnp.Client }

func (c Echo) Echo(ctx context.Context, params func(Echo_echo_Params) error, opts ...capnp.CallOption) Echo_echo_Results_Promise {
	if c.Client == nil {
		return Echo_echo_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x8e5322c1e9282534,
			MethodID:      0,
			InterfaceName: "aircraft.capnp:Echo",
			MethodName:    "echo",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Echo_echo_Params{Struct: s}) }
	}
	return Echo_echo_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Echo_Server interface {
	Echo(Echo_echo) error
}

func Echo_ServerToClient(s Echo_Server) Echo {
	c, _ := s.(server.Closer)
	return Echo{Client: server.New(Echo_Methods(nil, s), c)}
}

func Echo_Methods(methods []server.Method, s Echo_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x8e5322c1e9282534,
			MethodID:      0,
			InterfaceName: "aircraft.capnp:Echo",
			MethodName:    "echo",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Echo_echo{c, opts, Echo_echo_Params{Struct: p}, Echo_echo_Results{Struct: r}}
			return s.Echo(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// Echo_echo holds the arguments for a server call to Echo.echo.
type Echo_echo struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Echo_echo_Params
	Results Echo_echo_Results
}

type Echo_echo_Params struct{ capnp.Struct }

// Echo_echo_Params_TypeID is the unique identifier for the type Echo_echo_Params.
const Echo_echo_Params_TypeID = 0x8a165fb4d71bf3a2

func NewEcho_echo_Params(s *capnp.Segment) (Echo_echo_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echo_echo_Params{st}, err
}

func NewRootEcho_echo_Params(s *capnp.Segment) (Echo_echo_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echo_echo_Params{st}, err
}

func ReadRootEcho_echo_Params(msg *capnp.Message) (Echo_echo_Params, error) {
	root, err := msg.RootPtr()
	return Echo_echo_Params{root.Struct()}, err
}

func (s Echo_echo_Params) String() string {
	str, _ := text.Marshal(0x8a165fb4d71bf3a2, s.Struct)
	return str
}

func (s Echo_echo_Params) In() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Echo_echo_Params) HasIn() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Echo_echo_Params) InBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Echo_echo_Params) SetIn(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

// Echo_echo_Params_List is a list of Echo_echo_Params.
type Echo_echo_Params_List struct{ capnp.List }

// NewEcho_echo_Params creates a new list of Echo_echo_Params.
func NewEcho_echo_Params_List(s *capnp.Segment, sz int32) (Echo_echo_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Echo_echo_Params_List{l}, err
}

func (s Echo_echo_Params_List) At(i int) Echo_echo_Params { return Echo_echo_Params{s.List.Struct(i)} }

func (s Echo_echo_Params_List) Set(i int, v Echo_echo_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Echo_echo_Params_Promise is a wrapper for a Echo_echo_Params promised by a client call.
type Echo_echo_Params_Promise struct{ *capnp.Pipeline }

func (p Echo_echo_Params_Promise) Struct() (Echo_echo_Params, error) {
	s, err := p.Pipeline.Struct()
	return Echo_echo_Params{s}, err
}

type Echo_echo_Results struct{ capnp.Struct }

// Echo_echo_Results_TypeID is the unique identifier for the type Echo_echo_Results.
const Echo_echo_Results_TypeID = 0x9b37d729b9dd7b9d

func NewEcho_echo_Results(s *capnp.Segment) (Echo_echo_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echo_echo_Results{st}, err
}

func NewRootEcho_echo_Results(s *capnp.Segment) (Echo_echo_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echo_echo_Results{st}, err
}

func ReadRootEcho_echo_Results(msg *capnp.Message) (Echo_echo_Results, error) {
	root, err := msg.RootPtr()
	return Echo_echo_Results{root.Struct()}, err
}

func (s Echo_echo_Results) String() string {
	str, _ := text.Marshal(0x9b37d729b9dd7b9d, s.Struct)
	return str
}

func (s Echo_echo_Results) Out() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Echo_echo_Results) HasOut() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Echo_echo_Results) OutBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Echo_echo_Results) SetOut(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

// Echo_echo_Results_List is a list of Echo_echo_Results.
type Echo_echo_Results_List struct{ capnp.List }

// NewEcho_echo_Results creates a new list of Echo_echo_Results.
func NewEcho_echo_Results_List(s *capnp.Segment, sz int32) (Echo_echo_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Echo_echo_Results_List{l}, err
}

func (s Echo_echo_Results_List) At(i int) Echo_echo_Results {
	return Echo_echo_Results{s.List.Struct(i)}
}

func (s Echo_echo_Results_List) Set(i int, v Echo_echo_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Echo_echo_Results_Promise is a wrapper for a Echo_echo_Results promised by a client call.
type Echo_echo_Results_Promise struct{ *capnp.Pipeline }

func (p Echo_echo_Results_Promise) Struct() (Echo_echo_Results, error) {
	s, err := p.Pipeline.Struct()
	return Echo_echo_Results{s}, err
}

type Hoth struct{ capnp.Struct }

// Hoth_TypeID is the unique identifier for the type Hoth.
const Hoth_TypeID = 0xad87da456fb0ebb9

func NewHoth(s *capnp.Segment) (Hoth, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Hoth{st}, err
}

func NewRootHoth(s *capnp.Segment) (Hoth, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Hoth{st}, err
}

func ReadRootHoth(msg *capnp.Message) (Hoth, error) {
	root, err := msg.RootPtr()
	return Hoth{root.Struct()}, err
}

func (s Hoth) String() string {
	str, _ := text.Marshal(0xad87da456fb0ebb9, s.Struct)
	return str
}

func (s Hoth) Base() (EchoBase, error) {
	p, err := s.Struct.Ptr(0)
	return EchoBase{Struct: p.Struct()}, err
}

func (s Hoth) HasBase() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Hoth) SetBase(v EchoBase) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBase sets the base field to a newly
// allocated EchoBase struct, preferring placement in s's segment.
func (s Hoth) NewBase() (EchoBase, error) {
	ss, err := NewEchoBase(s.Struct.Segment())
	if err != nil {
		return EchoBase{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Hoth_List is a list of Hoth.
type Hoth_List struct{ capnp.List }

// NewHoth creates a new list of Hoth.
func NewHoth_List(s *capnp.Segment, sz int32) (Hoth_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Hoth_List{l}, err
}

func (s Hoth_List) At(i int) Hoth { return Hoth{s.List.Struct(i)} }

func (s Hoth_List) Set(i int, v Hoth) error { return s.List.SetStruct(i, v.Struct) }

// Hoth_Promise is a wrapper for a Hoth promised by a client call.
type Hoth_Promise struct{ *capnp.Pipeline }

func (p Hoth_Promise) Struct() (Hoth, error) {
	s, err := p.Pipeline.Struct()
	return Hoth{s}, err
}

func (p Hoth_Promise) Base() EchoBase_Promise {
	return EchoBase_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type EchoBase struct{ capnp.Struct }

// EchoBase_TypeID is the unique identifier for the type EchoBase.
const EchoBase_TypeID = 0xa8bf13fef2674866

func NewEchoBase(s *capnp.Segment) (EchoBase, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return EchoBase{st}, err
}

func NewRootEchoBase(s *capnp.Segment) (EchoBase, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return EchoBase{st}, err
}

func ReadRootEchoBase(msg *capnp.Message) (EchoBase, error) {
	root, err := msg.RootPtr()
	return EchoBase{root.Struct()}, err
}

func (s EchoBase) String() string {
	str, _ := text.Marshal(0xa8bf13fef2674866, s.Struct)
	return str
}

func (s EchoBase) Echo() Echo {
	p, _ := s.Struct.Ptr(0)
	return Echo{Client: p.Interface().Client()}
}

func (s EchoBase) HasEcho() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s EchoBase) SetEcho(v Echo) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// EchoBase_List is a list of EchoBase.
type EchoBase_List struct{ capnp.List }

// NewEchoBase creates a new list of EchoBase.
func NewEchoBase_List(s *capnp.Segment, sz int32) (EchoBase_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return EchoBase_List{l}, err
}

func (s EchoBase_List) At(i int) EchoBase { return EchoBase{s.List.Struct(i)} }

func (s EchoBase_List) Set(i int, v EchoBase) error { return s.List.SetStruct(i, v.Struct) }

// EchoBase_Promise is a wrapper for a EchoBase promised by a client call.
type EchoBase_Promise struct{ *capnp.Pipeline }

func (p EchoBase_Promise) Struct() (EchoBase, error) {
	s, err := p.Pipeline.Struct()
	return EchoBase{s}, err
}

func (p EchoBase_Promise) Echo() Echo {
	return Echo{Client: p.Pipeline.GetPipeline(0).Client()}
}

type StackingRoot struct{ capnp.Struct }

// StackingRoot_TypeID is the unique identifier for the type StackingRoot.
const StackingRoot_TypeID = 0x8fae7b41c61fc890

func NewStackingRoot(s *capnp.Segment) (StackingRoot, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return StackingRoot{st}, err
}

func NewRootStackingRoot(s *capnp.Segment) (StackingRoot, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return StackingRoot{st}, err
}

func ReadRootStackingRoot(msg *capnp.Message) (StackingRoot, error) {
	root, err := msg.RootPtr()
	return StackingRoot{root.Struct()}, err
}

func (s StackingRoot) String() string {
	str, _ := text.Marshal(0x8fae7b41c61fc890, s.Struct)
	return str
}

func (s StackingRoot) A() (StackingA, error) {
	p, err := s.Struct.Ptr(1)
	return StackingA{Struct: p.Struct()}, err
}

func (s StackingRoot) HasA() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s StackingRoot) SetA(v StackingA) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewA sets the a field to a newly
// allocated StackingA struct, preferring placement in s's segment.
func (s StackingRoot) NewA() (StackingA, error) {
	ss, err := NewStackingA(s.Struct.Segment())
	if err != nil {
		return StackingA{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s StackingRoot) AWithDefault() (StackingA, error) {
	p, err := s.Struct.Ptr(0)
	if err != nil {
		return StackingA{}, err
	}
	ss, err := p.StructDefault(x_832bcc6686a26d56[64:96])
	return StackingA{Struct: ss}, err
}

func (s StackingRoot) HasAWithDefault() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s StackingRoot) SetAWithDefault(v StackingA) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewAWithDefault sets the aWithDefault field to a newly
// allocated StackingA struct, preferring placement in s's segment.
func (s StackingRoot) NewAWithDefault() (StackingA, error) {
	ss, err := NewStackingA(s.Struct.Segment())
	if err != nil {
		return StackingA{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// StackingRoot_List is a list of StackingRoot.
type StackingRoot_List struct{ capnp.List }

// NewStackingRoot creates a new list of StackingRoot.
func NewStackingRoot_List(s *capnp.Segment, sz int32) (StackingRoot_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return StackingRoot_List{l}, err
}

func (s StackingRoot_List) At(i int) StackingRoot { return StackingRoot{s.List.Struct(i)} }

func (s StackingRoot_List) Set(i int, v StackingRoot) error { return s.List.SetStruct(i, v.Struct) }

// StackingRoot_Promise is a wrapper for a StackingRoot promised by a client call.
type StackingRoot_Promise struct{ *capnp.Pipeline }

func (p StackingRoot_Promise) Struct() (StackingRoot, error) {
	s, err := p.Pipeline.Struct()
	return StackingRoot{s}, err
}

func (p StackingRoot_Promise) A() StackingA_Promise {
	return StackingA_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

func (p StackingRoot_Promise) AWithDefault() StackingA_Promise {
	return StackingA_Promise{Pipeline: p.Pipeline.GetPipelineDefault(0, x_832bcc6686a26d56[96:128])}
}

type StackingA struct{ capnp.Struct }

// StackingA_TypeID is the unique identifier for the type StackingA.
const StackingA_TypeID = 0x9d3032ff86043b75

func NewStackingA(s *capnp.Segment) (StackingA, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return StackingA{st}, err
}

func NewRootStackingA(s *capnp.Segment) (StackingA, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return StackingA{st}, err
}

func ReadRootStackingA(msg *capnp.Message) (StackingA, error) {
	root, err := msg.RootPtr()
	return StackingA{root.Struct()}, err
}

func (s StackingA) String() string {
	str, _ := text.Marshal(0x9d3032ff86043b75, s.Struct)
	return str
}

func (s StackingA) Num() int32 {
	return int32(s.Struct.Uint32(0))
}

func (s StackingA) SetNum(v int32) {
	s.Struct.SetUint32(0, uint32(v))
}

func (s StackingA) B() (StackingB, error) {
	p, err := s.Struct.Ptr(0)
	return StackingB{Struct: p.Struct()}, err
}

func (s StackingA) HasB() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s StackingA) SetB(v StackingB) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewB sets the b field to a newly
// allocated StackingB struct, preferring placement in s's segment.
func (s StackingA) NewB() (StackingB, error) {
	ss, err := NewStackingB(s.Struct.Segment())
	if err != nil {
		return StackingB{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// StackingA_List is a list of StackingA.
type StackingA_List struct{ capnp.List }

// NewStackingA creates a new list of StackingA.
func NewStackingA_List(s *capnp.Segment, sz int32) (StackingA_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return StackingA_List{l}, err
}

func (s StackingA_List) At(i int) StackingA { return StackingA{s.List.Struct(i)} }

func (s StackingA_List) Set(i int, v StackingA) error { return s.List.SetStruct(i, v.Struct) }

// StackingA_Promise is a wrapper for a StackingA promised by a client call.
type StackingA_Promise struct{ *capnp.Pipeline }

func (p StackingA_Promise) Struct() (StackingA, error) {
	s, err := p.Pipeline.Struct()
	return StackingA{s}, err
}

func (p StackingA_Promise) B() StackingB_Promise {
	return StackingB_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type StackingB struct{ capnp.Struct }

// StackingB_TypeID is the unique identifier for the type StackingB.
const StackingB_TypeID = 0x85257b30d6edf8c5

func NewStackingB(s *capnp.Segment) (StackingB, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return StackingB{st}, err
}

func NewRootStackingB(s *capnp.Segment) (StackingB, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return StackingB{st}, err
}

func ReadRootStackingB(msg *capnp.Message) (StackingB, error) {
	root, err := msg.RootPtr()
	return StackingB{root.Struct()}, err
}

func (s StackingB) String() string {
	str, _ := text.Marshal(0x85257b30d6edf8c5, s.Struct)
	return str
}

func (s StackingB) Num() int32 {
	return int32(s.Struct.Uint32(0))
}

func (s StackingB) SetNum(v int32) {
	s.Struct.SetUint32(0, uint32(v))
}

// StackingB_List is a list of StackingB.
type StackingB_List struct{ capnp.List }

// NewStackingB creates a new list of StackingB.
func NewStackingB_List(s *capnp.Segment, sz int32) (StackingB_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return StackingB_List{l}, err
}

func (s StackingB_List) At(i int) StackingB { return StackingB{s.List.Struct(i)} }

func (s StackingB_List) Set(i int, v StackingB) error { return s.List.SetStruct(i, v.Struct) }

// StackingB_Promise is a wrapper for a StackingB promised by a client call.
type StackingB_Promise struct{ *capnp.Pipeline }

func (p StackingB_Promise) Struct() (StackingB, error) {
	s, err := p.Pipeline.Struct()
	return StackingB{s}, err
}

type CallSequence struct{ Client capnp.Client }

func (c CallSequence) GetNumber(ctx context.Context, params func(CallSequence_getNumber_Params) error, opts ...capnp.CallOption) CallSequence_getNumber_Results_Promise {
	if c.Client == nil {
		return CallSequence_getNumber_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xabaedf5f7817c820,
			MethodID:      0,
			InterfaceName: "aircraft.capnp:CallSequence",
			MethodName:    "getNumber",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(CallSequence_getNumber_Params{Struct: s}) }
	}
	return CallSequence_getNumber_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type CallSequence_Server interface {
	GetNumber(CallSequence_getNumber) error
}

func CallSequence_ServerToClient(s CallSequence_Server) CallSequence {
	c, _ := s.(server.Closer)
	return CallSequence{Client: server.New(CallSequence_Methods(nil, s), c)}
}

func CallSequence_Methods(methods []server.Method, s CallSequence_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xabaedf5f7817c820,
			MethodID:      0,
			InterfaceName: "aircraft.capnp:CallSequence",
			MethodName:    "getNumber",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := CallSequence_getNumber{c, opts, CallSequence_getNumber_Params{Struct: p}, CallSequence_getNumber_Results{Struct: r}}
			return s.GetNumber(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 8, PointerCount: 0},
	})

	return methods
}

// CallSequence_getNumber holds the arguments for a server call to CallSequence.getNumber.
type CallSequence_getNumber struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  CallSequence_getNumber_Params
	Results CallSequence_getNumber_Results
}

type CallSequence_getNumber_Params struct{ capnp.Struct }

// CallSequence_getNumber_Params_TypeID is the unique identifier for the type CallSequence_getNumber_Params.
const CallSequence_getNumber_Params_TypeID = 0xf58782f48a121998

func NewCallSequence_getNumber_Params(s *capnp.Segment) (CallSequence_getNumber_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return CallSequence_getNumber_Params{st}, err
}

func NewRootCallSequence_getNumber_Params(s *capnp.Segment) (CallSequence_getNumber_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return CallSequence_getNumber_Params{st}, err
}

func ReadRootCallSequence_getNumber_Params(msg *capnp.Message) (CallSequence_getNumber_Params, error) {
	root, err := msg.RootPtr()
	return CallSequence_getNumber_Params{root.Struct()}, err
}

func (s CallSequence_getNumber_Params) String() string {
	str, _ := text.Marshal(0xf58782f48a121998, s.Struct)
	return str
}

// CallSequence_getNumber_Params_List is a list of CallSequence_getNumber_Params.
type CallSequence_getNumber_Params_List struct{ capnp.List }

// NewCallSequence_getNumber_Params creates a new list of CallSequence_getNumber_Params.
func NewCallSequence_getNumber_Params_List(s *capnp.Segment, sz int32) (CallSequence_getNumber_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return CallSequence_getNumber_Params_List{l}, err
}

func (s CallSequence_getNumber_Params_List) At(i int) CallSequence_getNumber_Params {
	return CallSequence_getNumber_Params{s.List.Struct(i)}
}

func (s CallSequence_getNumber_Params_List) Set(i int, v CallSequence_getNumber_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// CallSequence_getNumber_Params_Promise is a wrapper for a CallSequence_getNumber_Params promised by a client call.
type CallSequence_getNumber_Params_Promise struct{ *capnp.Pipeline }

func (p CallSequence_getNumber_Params_Promise) Struct() (CallSequence_getNumber_Params, error) {
	s, err := p.Pipeline.Struct()
	return CallSequence_getNumber_Params{s}, err
}

type CallSequence_getNumber_Results struct{ capnp.Struct }

// CallSequence_getNumber_Results_TypeID is the unique identifier for the type CallSequence_getNumber_Results.
const CallSequence_getNumber_Results_TypeID = 0xa465f9502fd11e97

func NewCallSequence_getNumber_Results(s *capnp.Segment) (CallSequence_getNumber_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return CallSequence_getNumber_Results{st}, err
}

func NewRootCallSequence_getNumber_Results(s *capnp.Segment) (CallSequence_getNumber_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return CallSequence_getNumber_Results{st}, err
}

func ReadRootCallSequence_getNumber_Results(msg *capnp.Message) (CallSequence_getNumber_Results, error) {
	root, err := msg.RootPtr()
	return CallSequence_getNumber_Results{root.Struct()}, err
}

func (s CallSequence_getNumber_Results) String() string {
	str, _ := text.Marshal(0xa465f9502fd11e97, s.Struct)
	return str
}

func (s CallSequence_getNumber_Results) N() uint32 {
	return s.Struct.Uint32(0)
}

func (s CallSequence_getNumber_Results) SetN(v uint32) {
	s.Struct.SetUint32(0, v)
}

// CallSequence_getNumber_Results_List is a list of CallSequence_getNumber_Results.
type CallSequence_getNumber_Results_List struct{ capnp.List }

// NewCallSequence_getNumber_Results creates a new list of CallSequence_getNumber_Results.
func NewCallSequence_getNumber_Results_List(s *capnp.Segment, sz int32) (CallSequence_getNumber_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return CallSequence_getNumber_Results_List{l}, err
}

func (s CallSequence_getNumber_Results_List) At(i int) CallSequence_getNumber_Results {
	return CallSequence_getNumber_Results{s.List.Struct(i)}
}

func (s CallSequence_getNumber_Results_List) Set(i int, v CallSequence_getNumber_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// CallSequence_getNumber_Results_Promise is a wrapper for a CallSequence_getNumber_Results promised by a client call.
type CallSequence_getNumber_Results_Promise struct{ *capnp.Pipeline }

func (p CallSequence_getNumber_Results_Promise) Struct() (CallSequence_getNumber_Results, error) {
	s, err := p.Pipeline.Struct()
	return CallSequence_getNumber_Results{s}, err
}

type Defaults struct{ capnp.Struct }

// Defaults_TypeID is the unique identifier for the type Defaults.
const Defaults_TypeID = 0x97e38948c61f878d

func NewDefaults(s *capnp.Segment) (Defaults, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 2})
	return Defaults{st}, err
}

func NewRootDefaults(s *capnp.Segment) (Defaults, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 2})
	return Defaults{st}, err
}

func ReadRootDefaults(msg *capnp.Message) (Defaults, error) {
	root, err := msg.RootPtr()
	return Defaults{root.Struct()}, err
}

func (s Defaults) String() string {
	str, _ := text.Marshal(0x97e38948c61f878d, s.Struct)
	return str
}

func (s Defaults) Text() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextDefault("foo"), err
}

func (s Defaults) HasText() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Defaults) TextBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytesDefault("foo"), err
}

func (s Defaults) SetText(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Defaults) Data() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return []byte(p.DataDefault([]byte{0x62, 0x61, 0x72})), err
}

func (s Defaults) HasData() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Defaults) SetData(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(1, d.List.ToPtr())
}

func (s Defaults) Float() float32 {
	return math.Float32frombits(s.Struct.Uint32(0) ^ 0x4048f5c3)
}

func (s Defaults) SetFloat(v float32) {
	s.Struct.SetUint32(0, math.Float32bits(v)^0x4048f5c3)
}

func (s Defaults) Int() int32 {
	return int32(s.Struct.Uint32(4) ^ 4294967173)
}

func (s Defaults) SetInt(v int32) {
	s.Struct.SetUint32(4, uint32(v)^4294967173)
}

func (s Defaults) Uint() uint32 {
	return s.Struct.Uint32(8) ^ 42
}

func (s Defaults) SetUint(v uint32) {
	s.Struct.SetUint32(8, v^42)
}

// Defaults_List is a list of Defaults.
type Defaults_List struct{ capnp.List }

// NewDefaults creates a new list of Defaults.
func NewDefaults_List(s *capnp.Segment, sz int32) (Defaults_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 2}, sz)
	return Defaults_List{l}, err
}

func (s Defaults_List) At(i int) Defaults { return Defaults{s.List.Struct(i)} }

func (s Defaults_List) Set(i int, v Defaults) error { return s.List.SetStruct(i, v.Struct) }

// Defaults_Promise is a wrapper for a Defaults promised by a client call.
type Defaults_Promise struct{ *capnp.Pipeline }

func (p Defaults_Promise) Struct() (Defaults, error) {
	s, err := p.Pipeline.Struct()
	return Defaults{s}, err
}

type BenchmarkA struct{ capnp.Struct }

// BenchmarkA_TypeID is the unique identifier for the type BenchmarkA.
const BenchmarkA_TypeID = 0xde2a1a960863c11c

func NewBenchmarkA(s *capnp.Segment) (BenchmarkA, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 2})
	return BenchmarkA{st}, err
}

func NewRootBenchmarkA(s *capnp.Segment) (BenchmarkA, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 2})
	return BenchmarkA{st}, err
}

func ReadRootBenchmarkA(msg *capnp.Message) (BenchmarkA, error) {
	root, err := msg.RootPtr()
	return BenchmarkA{root.Struct()}, err
}

func (s BenchmarkA) String() string {
	str, _ := text.Marshal(0xde2a1a960863c11c, s.Struct)
	return str
}

func (s BenchmarkA) Name() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s BenchmarkA) HasName() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s BenchmarkA) NameBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s BenchmarkA) SetName(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s BenchmarkA) BirthDay() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s BenchmarkA) SetBirthDay(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

func (s BenchmarkA) Phone() (string, error) {
	p, err := s.Struct.Ptr(1)
	return p.Text(), err
}

func (s BenchmarkA) HasPhone() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s BenchmarkA) PhoneBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return p.TextBytes(), err
}

func (s BenchmarkA) SetPhone(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(1, t.List.ToPtr())
}

func (s BenchmarkA) Siblings() int32 {
	return int32(s.Struct.Uint32(8))
}

func (s BenchmarkA) SetSiblings(v int32) {
	s.Struct.SetUint32(8, uint32(v))
}

func (s BenchmarkA) Spouse() bool {
	return s.Struct.Bit(96)
}

func (s BenchmarkA) SetSpouse(v bool) {
	s.Struct.SetBit(96, v)
}

func (s BenchmarkA) Money() float64 {
	return math.Float64frombits(s.Struct.Uint64(16))
}

func (s BenchmarkA) SetMoney(v float64) {
	s.Struct.SetUint64(16, math.Float64bits(v))
}

// BenchmarkA_List is a list of BenchmarkA.
type BenchmarkA_List struct{ capnp.List }

// NewBenchmarkA creates a new list of BenchmarkA.
func NewBenchmarkA_List(s *capnp.Segment, sz int32) (BenchmarkA_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 24, PointerCount: 2}, sz)
	return BenchmarkA_List{l}, err
}

func (s BenchmarkA_List) At(i int) BenchmarkA { return BenchmarkA{s.List.Struct(i)} }

func (s BenchmarkA_List) Set(i int, v BenchmarkA) error { return s.List.SetStruct(i, v.Struct) }

// BenchmarkA_Promise is a wrapper for a BenchmarkA promised by a client call.
type BenchmarkA_Promise struct{ *capnp.Pipeline }

func (p BenchmarkA_Promise) Struct() (BenchmarkA, error) {
	s, err := p.Pipeline.Struct()
	return BenchmarkA{s}, err
}

const schema_832bcc6686a26d56 = "x\xda\xacZ\x0btU\xd5\x99>\xfb\xbeN\x92\x9b\x9b" +
	"\x9b\x93s\x8dya \x92\x00\xb1\xd8<\x18\xa4t\\" +
	"y\x90Tt\x81M\x08\xc8\xe0H\xe5$9I.\xde" +
	"\x17\xe7\x9e\x0b\x04\xeb\xa2\x9dBQ\x97Le)\x15\x1f" +
	"\x99\xd2\x0c\xcc\x80\x12\x06,\xcc\x00\x03\x14,X\x12a" +
	"*.P`\x00\x05E\x0b\xea\xd4\xd82BU\xee\xfc" +
	"\xff>\xcf{\xef\xb9\xa0\xacY\x8b\x98\x93\xff\xfb\xf7\xbf" +
	"\xff\xfd\xbf\xf6\xbf\xf7\xb6jJv\xbd\xad\xda\xf9+\x9e" +
	"aZ?p\xba\xe2\x07\xaf|\xfav\xd5\xa3\xe5\xcb\x99" +
	"V/!\xf1\x07\x82\xfd?\xef:|\xc7\xcf\x18\x07\xcb" +
	"0\xfc\xf6\xcc!~\x7f&~\xed\xc9\xaccH|\xee" +
	"\x1b\xb9?\xca\xdc=uE\x12\xaf\xd3\x86,g2w" +
	"\xf1\x17(\xf3\xb9\xcc\x7f\x03\xe6\x13\xbf\xb9Vu{\xdd" +
	"\xefV0\x9c\xd7\xccK\x80\xa3veV\x1e\xe1\xfb\xb2" +
	"\x90\xf9\xf9,\x94<yq\xfd\xc4W\x8f\x8cz<I" +
	"r\x13k\x03\x96\xc1\xac!\xfe8e>\x9a\xb5\x08\x98" +
	"\xfb\xff\\\xfc\xce\xb6\x87\xf3\x9fd8\x1faT\x89\xe3" +
	"\xdd6\xc2\x10\xbe\xda\x8d\xd2&\x94\x8f\xbd\xb8\xbf\xac\xed" +
	"\x1faj\xbb!\x0c\xe0Vw??\xc7\x8d\x92f\xb9" +
	"\xef\xe1\x1f\xc3\xaf\xf8\xd3\x87J_oxt\xf3/\x92" +
	"\xf4\xa4k\x12\xdd\xe7\xf9\x05\x94?\xe8\xc6\x99\xe3\xd2\xe1" +
	"K\xad\xcf\x0f>\x93\xc8Kmu\xdc\xfd\x1a\x7f\x06X" +
	"\xed\xf1\xecc\xbf<\x90\xf7J\xd5\xb3\xc0\xe4H\x98}" +
	"\xa7{\x88?\x88\xd2\xda\xf6\xb9\xed\xa4\xed4\xaa\xcc\xc4" +
	"W\x8e\xca\x98p\xe5\xb1\xdf>ka'\xfe$\x8c\xb8" +
	"@\xe7?G\x17\xf6Hh\x9e\xaf\xee\xda\xde\xd5V6" +
	"%\xd9`S.\x1b\x99=\xd9\xc8\xbcrE\xe9\xebS" +
	"\x9fx\xff9\xb4\xa9-yewg\xbf\xc67#s" +
	"mCv)X.>g\x93\xed\xc5\xe7\x9e\xdf\xfe\x82" +
	"\x95\x1as<C\xbc\xe8\xc1/\xc1\x83\x92\xfb\x1e=\xb3" +
	"s\xdc;w\xbdhv\xc0\x13\x9e,t\xc0*\xcap" +
	"h\xd6y\xe7\xae1\xbfx1\xc5\x04[A\xd2\x1e\x94" +
	"\xd4\xb6\xc3\x03&8\xe0\xa1&\x88}\xdf\xf1\xf3xM" +
	"U_r\\\xd1\xc9\xb7\xc3\x90\xfdt\xf2=\x1e\xf4\xc1" +
	"s\xb7\x1d\xfdn\xcbUq\x1d\xd3ZB`0\x1a\xbf" +
	"\xb60G\xc2\xc9\xcbsp\xf2\xae\xa9\xdd\x9f_\xe3\x7f" +
	"\xbb\xc1j%\xcd9\xaf\xf1\xd3s\xf0\xeb^\xca;\xf2" +
	"\xd0\xad\x8b\x1f~w\xf3+)\x91\xe2\xcf9\xcf\xc7(" +
	"\xe3\x82\x9c{\xf8>\xfc\x8a\x0f?6\xbd\xb2q\xd6\x9b" +
	"\xafXY\x7fyN\x11\xe1W\xd3\x01\xab\xa8\xe4\x9d\x1f" +
	"o\x097\x9fZ1`\xa5\xc5\xfe\x9c~~\x90\xf2\x1e" +
	"\xa4\xbc\\\xa4\xf3\x9d\x90s\xd3V+\xde\x0b9\x9f\xf3" +
	"\xc3\x94\xf7S\xca\xbbt\xe2\x8f\x96\xcd\x9d\xf4\xd9V\xb4" +
	"\x95Ie\xa7\x1dYFx\xdf\xe2\xc7yQ\x9fr\xef" +
	"l\xf4jthB\xfc\x93\xf3\xb7\xfd\xbbU\x0c\xd4\xae" +
	"\xca\xb5A\x16\xe6\xd2,\xcc\xfd\x08\xb8\xd7\xfd\xcb\xae\xf2" +
	"\xd7\x83w\xfc\x07\xd3\xca\x81\xec\x93k\xdf\xfe\xfes\x17" +
	"+.1\xb7\x10\x96`\x0ap/\x83a\x82\x1cMA" +
	"\xf7\xe7\xfdW\xc5S\xbf\xb7\xd2x\x90{\x99?\xca\xe1" +
	"\xd7\x11\x0e5\x1e?\xfd{{\xde\x7f\xe5\xef\x07\xad\x12" +
	"\xecSn\x88\xbfJy/S\xb9\xbb\xbex\xf7\xf8\xc3" +
	"\x0bO\xbfae\xe2\xd6<0\xb1\x90\x87\xccs\xf3P" +
	"\xf0\x9a\x03\x1b\xdd\x1frM\x87\xad\x94X\x99\xb7\x8b_" +
	"MyWQ\xde\xf7*&\xcd{\xff\xd5\xdfX\xf2\x1e" +
	"\xcc\xeb\xe7\x8fP\xdeA\xca{o\xcb\xf9\x93\xe777" +
	"\xfd\x97\xa5\x89/\xe6]\xe2/#s\xedp\x1e5\xf1" +
	"\xa1\x9f\xec/\x19\xba\xf4\xe2\x1f\xacT\x9e\xee\x83\x9c\x9c" +
	"\xeb\xa39\xe4C\xd1{O=>\xbc\xee\xcb\x89o[" +
	"\xa9\xf1\x84\xef\x05~\x15\xe5]IyW\xdd\xf5\xaf\xf3" +
	"B\x7f\xd8}\x02\xd5p$\x1bn\xc07\xc4\xefD\xe6" +
	"\xda\xed>\xaa\xc6\xca7N.z|\xdeS'\xad$" +
	"\x17\xe6\xf7\xf3\xa3\xf2i\x80\xe4\xa3d~\xe1\x97\xfe\xae" +
	"\x86\xa3g\xac<r7\xf06S\xde\x86|\xf4H\xc9" +
	"\xfe\x8e\x8c_\x16U\x9eM6\x86\xa2E\xfe[\xfc\xce" +
	"|\xaaE>\xd5\xa2o\xe6\xec\x81\xff\xdc\xdcr\xd6j" +
	"7\xc9,x\x99\xe7\x0ah}*\xc0\x0d\xe2\x99\xaa\x0d" +
	"\x7f\xfd\xdbc\xfft\xd6\xcap\xeb\x0b\xb2\x08\xbf\x9d2" +
	"o-@\x95\x07\xf6\xb2\xdc\xf1\xa3\xfd\xe7\xac\x96w\xae" +
	"`\x17\x7f\x91\xf2^\xa0\xbcB[m\xde\xc1K\x83\x96" +
	"\xbc\x9e\xc2\x17\xf8[\x0a\xf1\x8b+D\xde+s\xd6\xfe" +
	"\xec\xa5\xfe\x8c\x0bVJT\x17\x82\xf7\x1a(\xf3\xdd\x94" +
	"y\xeb\xefg\x9d\xdd\x9c{\xff\x85\xa4\xd55\x13\xd6\x81" +
	"\xf9Q\xf8\x1a\x1f\xa3\xdc\x0b\x0a1\x9bF}\xf7J\xf1" +
	"W\xcb\xe7\xa2h[Bmi.\xda\xc5O/\xa2E" +
	"\xa8\x08\x0dq\xdau\xf5W\xcb\x96\xfe4Y\x07\x1an" +
	"}EC\xfcF\xca\xbb\x9e\xf2\xce\xe9)\xfa\xd3\xe4\x8f" +
	"\x97}h\xb5\xb6`\xf1)\xbe\xb7\x18\xbfb\xc5t\xb7" +
	"8\xbew\xdd@\xd1\x82\x8fR\xaap_1\x08E\xc6" +
	"\xb6u\xc5P\x85\xb7\x14\xd3*\xac\xe7{\xa2\xa7au" +
	"w\xe0\xf4\xc5O\xf2\x038\xa6vcq8\x13\x03#" +
	"cnc\xe6\xc0\x0f\x87\xad49Rq\x8a?YA" +
	"\xb7\xc5\x0a\x9a\xa9\x85yO\xfe\xe5\x1fV\\f\xb8\x12" +
	"\xadd;\xc7\xcc'\x8c#\xfeu\xd7=\x83\xcd\xa7\x9d" +
	"_$\x15(\x1a0\xc3\x15o\xf1_S)W+0" +
	"\x14\xe7\x17\x04\xee\xf15\xc5\xbf\xb0\x9aq\xd6\x98\xb7x" +
	"a\x0c\xad\x0dcp\xc6\xb3\xd3\xf6>3V\xfe\xe7\xaf" +
	"\xac\x02\xf1\xd7\xc0;@y7\x02\xef\xf8\xb8\xe0\x97:" +
	"$\xa1K\xb6\xdd\xd9!DB\x91\xc9m\xb2\xd0\xf1\x88" +
	"?\xd4\xdd\xc80-\x84\xb4:\xec\xe0]\x07\xe8\xcdy" +
	"\xca\xa0i\xca\xb0\x93V\x9f\x8d\xb0\xa1X\x908\x18\x1b" +
	"\xfc\x10]\x02Q%L\xa9\x0b\xc7B\xb2(\xe1\xf0l" +
	"}xs%\x0c\xaf\x87\xe1\xd3l\x84\x10\x1f\xee^\xdc" +
	"\xbd5@k\x02Z\x8b\x8dp6 B\xc3\xc3M\xbf" +
	"\x0f\x88\xd3\x80\xd8c#\xde\xa8\x7f\x89H\x9c0\x91\x93" +
	"!\xa5\x8b\xc2Rg\x94d\xc3_\xd90-\xfe\x15\xf0" +
	"Ge\x06\x9c\x9a\xc3\x90\x16;\xa1P\x8eI#\xbb\xaa" +
	"\xd1\xd4p\xa03\xfa\x80(\xcd\\\x14\x86\x7f-\x81\x18" +
	"\x89&\xadm\xb2\xba\xb6\xd16R\x17\xecE\xb1\x9a\xcc" +
	"\\\xa3\"2$A\xbaf\xb1\x07\xc2\xfe\xceY!\x7f" +
	"8\xa4X,\xc3\xee\xc8\x8e\xc7\xa9\xd8qy v4" +
	"\x88\xad\xb2\x11\x0f\xb9\x16WV=\x1e\xa9c\x81:\x01" +
	",!0.\xd2\xce\xb8RTn\xee\xe8\x09\xdf)\xc2" +
	"\x7fF\xb7\x08\x92\x10\x8c2fm\x8b\x0cO\xd8\xfd!" +
	"\xdd\"\xc9\x8ehfa\xbc\xb2P'\x84\xb8\xd69\x12" +
	"\xad\x83\xe1\xb8J\xc6\xc69Y/\xceSO\x803m" +
	"4\xcc`\xc3aY]\x1d!\xf0\x8fp\xe3\xe6\x1b\xcb" +
	"\xc8U\x1d\xfa=\\\xda\x04\xa0\xd5\xdb@\xd6l\xbf\xdc" +
	"\xd3$v1^!\x16\x90\xc1\x90z\xa7\x03\x86\xcc\xa5" +
	"\xc6\"\x10\x14DH\x81,l,J\xcd\xc1\x88\xdc\xcb" +
	"X)\xd9\x11\x0eE\xe5i4\x14P\xa6\xee8\xbd\x1e" +
	"S\xc71\x1c\xb9\x95\xcb \xdew\xd9\x8cb\xfcO\x89" +
	"\xd5$?\x0c\x89-\xb2t\xdd\xd0\x8f\xc8\x12\xc8\xd6S" +
	",Ia\xeb\x90\x03\xe5\x88\xfc-BN\xef\\\xd2\x84" +
	"\x1cX\x15m\x1a\xa5z\xfa\x14\x99\xe0\x92\xc70\xc7\x16" +
	"\x83\xcce\x90O4\xc9\x80\xb8\x1c\x89?\x01\xe2S\x10" +
	"n6\xc81\xa0\xad\xc4\xc4{\x1ch\xcf\x02\xa3\x1d\x18" +
	"aVn\x15\xae\xf2) \xae\x01\xa2\x038A&\xb7" +
	"\x1aG?\x0d\xc4\x97 \x1beq\xb1\xacF\x1bX\xb3" +
	"\x8c\xed\x0a\x87\xbd\x9d\x82,\x10\x0f\xd0<H+b\xdb" +
	"\x05\xa9\xb4+\x10\x16d\x92\xc5\xd8\x86\xb3~wyj" +
	"=CX\x7fH\xc6z1\xecX\x1e\x8f\xc7\x19\xe2\x8d" +
	"!!\x03\xa2/\xa3\xd2by\xb3%!\xa2\xb8;\xd9" +
	"\x11\xaf\x826\xb9\xa0M\x09\xc4W\xd0\xdf\xdd#\xdf\x1f" +
	"\x96I\xa38C\x14\x02\x81\xdeR:\x06\xcc\xa7\x9fT" +
	"\xd28\xc7H\xae\x19b\x94\xda\x91I\xe7\xedpLN" +
	"\xc9\xaf\x84\xb8k\x86R\xa8\xc4\x9d\xd7\xd8\xf4\x18\xe2\xf4" +
	"\x12&}:5\xe8\xa5B\x9ds\\\x99Q)\xb4\xea" +
	"h\xae\x13\xe6\x82K\xdaa\x89\xfa\xc95i\x89\x0e\xad" +
	"\x08\x83A\xda\xc4\x0511\xd4!\xde\xd9-\xca\xf7\xc7" +
	"\x82\xed\xa2\x04\xeb-\xa5\x0b6/7\xcfX.\x09\xa1" +
	"S\xe0'Uu\xb4Y\xa3\x10\x15\x93=Ri\x8c\xa6" +
	"u\x84p\xc6\x09\x14t\xe3,D\xe9\xba\xb1\xa0\x9cQ" +
	"\x9f\xb4}\x92h\x87\x1c\x8e\x9b\x01\x11\x92\xc9\xc65\xfd" +
	"\x19\"%\x96\xa9\xe4l\x83\xf4m\x82x\x9c\xe6\xb7G" +
	"\xbfM\xba\x99\x939\xc7\xa2\x90Ne\xc3r\xcfu\xd6" +
	"\xdd\x0ev\x011\xfa\xd9+M\x11\xc3\x12\xd5&K\xb1" +
	"\x8eRy\x0a\x10\xd2\xd4\x18P\x90](v\x18\xda\xe9" +
	"\xbdF\x9ab0C\xec\x96\xc4h\xd4\x1f&Td\x81" +
	".\xf2y\xd4\xf1Y\x10\xb9\xd6\x08\xaa>\xdc;\xd6\x00" +
	"m\x9di\xcb\xfd52\xbe\x04\xc4\x1dX\x0e \xf3\xed" +
	"@\xdc\x8e\x16\xdb\x02\xc4\xc3X\x0e\x80\x13\xc4r\x83\xa8" +
	"\xe5\x01 \xbe\x09D'p\x82\xe3\xb8#H<\x04\xc4" +
	"c\x86-\xf4\xfe]\xb1\x85\xbd\xbd\x8a\xb8!\xb2\xdc\x90" +
	"\xfb\xed\"T\x0cuune\xc7\xae\x8b\x04\x84\x90\x18" +
	"5\xd6\xac7\x9b\xca\x9a\xd9\xde`L\x1b\xcf\xf6F;" +
	"\xb5\xef\x94@P*.\xc6\x00n\xf4P\xc9\xd1\"\xb9" +
	"\xbaE\x04\xd4\xf4!\xa5\xb7\xe04\x93\x88H\x9c\x07\xc4" +
	"\x80Z \x81\xe6G\x8bt\x02-\xa2\x15H \x06\x91" +
	"\xd8\x03D\x19]$\x04\x80\x08\x1b0h\xd4\x19\x0bk" +
	"M\x8a\x17\xb6\x87\xea\xd4\xfd\x01\xc95\xd7\xd96\xb4H" +
	"{\xd0{g\xb7\x14\xc1\xc2\x80[j%V\x86\x1a\xa3" +
	"2p\xa8\xde\x1dX\x1a&\x1b\xa5\xa1\xb4\xcb/A0" +
	"g\xc2\xfc\xd0\x99\xd6EE(K\x9d\xda\x9f\x16\x13`" +
	"\xc1\xbeN,\x9b\xea\xb9\xe56\xac\xda5\xb9~U\x9a" +
	"\xb5\xd4\x0aX\xa5\xa1\xe5M\xda\xc5b;\xbd\x99\x04\xd7" +
	"\x1b\xed4\x09\xfe \x98MZ\xa8\xb4\xac&\x91\xed " +
	"2\x1bD\x8e\x85\xddf\x91\xe0\x97\xa1r\xcfg\xd8p" +
	"\xbb)P\xf5\xd3d\x1a\xc9\x8d\xec]\xb5w\xdd\xb8t" +
	"$\xa5K\x1a\xc3\xc3?/t\xaf\xd1\xa44/3\xd2" +
	"\\\xb7}_\x99\x91\xe7ZP\xebi\xbe\xc1\x14\xd4\xeb" +
	"\x91\xb8\x16\x88\x9b\xb4]\x1f\x88\x1bq\xf4: n1" +
	"\xa5\xf9\x00rn\x00\xe2\x81\x1b\x87\xbf\xb9\x811\xb99" +
	"\x89\xcc\xca\x92\xde\xdd{\x03Q\xb9V\xb3\xab\xf3\xc6\x8d" +
	"<\x04\xe1\xb7\xed\xaa\xf4\xbb\x98t\xbe\xb2\x0b\xddI\xe2" +
	"\x1a\x0dW-\xedP\x8e5 H\xbf\x0aN\xe3\xad\x16" +
	"\xacft\xb3d\xd2Wd\xc3U5\xaa\xab\xb6\x19\xae" +
	"\xda\x8a\xeb\xd8\xa4U\xe4\xfa\xa4\x8a\xbc\xdb\xe4\xaa\x9dx" +
	"\\\xda\xa1x\x85s\xda\x15W\xedG\xe2>\xa5v{" +
	"CBP\xd4z\x98\xd2\x9ep\xd0\xa8\xb3\x09\x1d\x0b\xad" +
	"\xc3\x92\x80A\xae\xf9\xa4\xaeC\x08\xfd \xd0\x0bZ\x82" +
	"f\xb0NX\x9d\xd0\xe1\xc7\xb6\x8c\xd1X\xe2Aaq" +
	"[D\x14;\x91\x96\\\x955\xc36\xb0\xb55U7" +
	"\x9f\x04z\x96\xb2\xf3\xc3\xed\xe9\xfb\xa6\xd4\xba\xd3\x04\x81" +
	"\xda\x11\xec\xd4\x16\xef\x15\xa4\xeeh\xba\xb3\xa2\xe6\xbbF" +
	"\xe8\x98z\x82\x82\xf4\x08i\xf8\x06\xbb\xe9}\xa6\x8c\xd2" +
	"v\xd3\xf55\xa6\x8c\xd2v\xd3\x8d\xf7\xa9\xc9\xb3\x0d}" +
	"7O\xf1]\x82\x97\xb54\xdb^cx9\xc1w\xf1" +
	"v\xbf\x04\x07*\xc1l\xfe\xd2HO8dpD\xfd" +
	"\xed\x01\xf0_\x149\xd4n\xb1.\x1a\x09\xc7\xc0\xbc\xaa" +
	"\x0fK\x83\xc0\xdf\x9b\xd6St{\x10\xd3\x9f\xdc\xe9\x0e" +
	"\x94zt\xb7\xabGw\xf4\xc6T \xce\x04\xd5{E" +
	"A\xd2J\x04\xce*\xf7\x10\x17\xfc\xe5\xc2\x82!\xf4j" +
	"\xdfiS\x9c\xf6\xf2\xfa\xa1\xee\x9b\xa6\xb8\xb9\xf3\xb7J" +
	"\xf1\xd9\x10\xe1\x91\x9a\xc557u\xacH\xaa^)\xc2" +
	"\x7f`\xaf\x9ex\xf3Qn\xd1\xcb\xdeD\x8d\xd3\x1fT" +
	"\xd24\x8b\x0d\xea\xdfZc\xa4]U\x08\x95Fgd" +
	"\\U\x88\x95Fk\xe4\xb1}\x1dOm\x8e<\xf6\xaf" +
	"\xe2jwTftG\xde\x85a\x7f'\xe3\xf2\xb6\xc3" +
	"\xe6\x07J\xe9\xf7\xcf\xean @9@]\xb5[[" +
	"u7\xe8\xaa\x9e\x08T\xfd\x023\x8d\x95\x1b\xea\xfcR" +
	"$,Q\xa3\x94(aWIO\x1a\xcd\xa0\x00\xb1q" +
	"w\xe3/;\xf77\xf8\xcb\xc1\x8d\xc7_N\xae\x1c\x7f" +
	"\xb9\xb8\x11\xc0\xe9\x0dA\x06\xb0\xf3\xbb\x1ea\x03\xc2b" +
	"6\xda\x15f\x03\xb1\x85lg\xd7\"8\xdeF\xe5\x14" +
	"\x83Qw\xcc\x84\x83\xaf\x12\x87\xa6\xb4(3\xa7\x85v" +
	"\xa3U\xa6\xa6\xc5<L\x0b\xb5\x98\xcfE\x87\xfd\x9dr" +
	",ge\xfd\x08M\xd8\x80\xe19\xb5\x1c\xd5\x01\xc9D" +
	"\xbdA\xad\x9a\xad\x04s$`\x8fEo*\xa2\xcdW" +
	"[\xb9\xe9N\xb8\xd0p)\x1bY\xd2\x95J.\xc3\xe4" +
	"*\xd7)\xa9eD1\xd6&=\xbe\xf8\xf5\x0e0}" +
	"\xdbZ\x87\x9d\xb4mr\x98B\x8c\xdf\xe8(\xc2\xabX" +
	"\x04\xb6\x000\x02\xa2L\xa9\xa3\xfc\x80\x03\x8c\xd9\xb6\x01" +
	"\x91m\x88@\xa4)\xc5\x94\xdfJ\x91M\x88\xec@\xc4" +
	"\xf1e\\9\x9f\xf0\xdb)\xb2\x05\x91\xdd\x888\xff\x1a" +
	"W\xca*\xbf\x93\"\xdb\x10\xd9\x87\x88\xebj\xdc\xe1\x83" +
	"\x0a\xc4\xf0{(\xb2\x03\x91\x03\x88\xb0W\xe2\x19>\xe5" +
	"\xa9\x8a\xea\xb6\x1b\x91C\x88d|\x81\xf3d\xe0\xab\x09" +
	"\x1d\xb3\x0f\x91\xc3\x88d\xfe/\xce\x93\x89\xaf(\x149" +
	"\x80\xc8\x9b\x88d]\xc6y\xb2\xf0^\x98\"\x87\x109" +
	"\x86\x88\xfb/8\x8f\x1b_v\xe9<\x87\x119\x81H" +
	"\xf6\x9f\xe3\xb0\xe5g\xe3\x0d25\xdb\x9b\x88\x9cF\xb3" +
	"y>\x07\xb3y\xf0q\x94\x02\xc7\x10x\x0f\x81\x9ca" +
	"\x00r\xf0%\x9a\x02'\x10\xf8\x00\x01\xefg\x00x\xf1" +
	"\xa1\xc0\x01a\x08R\x00\xb8\x82@\xee\x9f\x00\xc8\xc5\xc7" +
	"'\x0a|\x86@\xb6\x13\x00\xee\x7f\x00\xe0\xf0\xcd\xc2\x89" +
	"\x80\xc3\x09\xc0H\x04\xf2>\x05 \x0f\x1fR(P\x80" +
	"\xc0\x04\x04\xf8O\x00\xe0\x01\xa8\xa6\xc0w\x10\x98\x8a\x80" +
	"\xefc\x00|\xf8\xf4\xe8\x84\xfd\xa2\xad\x1e\x81\x87\x10\xb8" +
	"\xe5\x12\x00\xb7\xe0\xcb\x10\x1d1\x13\x81\x08\x02\xf9\x17\x01" +
	"\xc8\xc7\xfb|\x0a\xf4 \xb0\x0c\x81[\xff\x08\xc0\xad\x00" +
	"\xfc\x94\x02?F`\x0d\x02\x05\x1f\x01P\x00\xc0j:" +
	"\xc7\xd3\x08lB\xa0\xe4C\x00\x0a1\xc4\x9ch\x92u" +
	"\x08\xecC`\xc4\x05\x00\x8a\xd0\xf3\xceF\xf4<\x02\x1f" +
	" p\xdb\x07\x00\x14\xa3\xad\xa8\xa8\xd3\x08\xfc\x11\x81\xd2" +
	"\xf7\x01(\xc17\x16\x0a\xbc\x87\xc0'\x08\x8c<\x0f\xc0" +
	"\x08|\x19s\xc2a\x01\x98\x01p\xb8\x00\x18u\x0e\x80" +
	"\xdb\x00 .\xd8\xf9g\xb8\xd0\xb8H/{\x0f\xe8\xa5" +
	"h\\\x17$I[\x06\x02>\x04n\x7f\x17\x80\x91\xf8" +
	"4\xe3\x9a\x01@.\x02%\x00\x8c\x18}\x16\x03h\x14" +
	"\xbed\xb9P]\x1f\"#qH\xf9\x19\x18R\x86\xfe" +
	"p\xe1\x02\x0b\x10\x18\x8d@\xc5i\x00n\x07`\x14\x05" +
	"J\x10\x18\x8b\xc0\x98\xff\x06`4\x00\xe5.\x8c\xc5\x91" +
	"\x08|\x07\x81\xb1\xa7\x00(\x07`\x1c\xaa\x0b\xcc\x004" +
	"!0\xee$\x00\x15\xf8 F\x81z\x04\x1eB\xa0\xf0" +
	"\x04\x00c\xd0\x83T\xab\x99\x08D\x10(z\x07\x80\xb1" +
	"\xe8A\x0a\xf4 \xb0\x0c\x81\xe2\xb7\x01\x18\x87\x1e\xa4\xc0" +
	"\x8f\x11X\x83@\xe5\xf181\xbd\xbf\xf2\xabA9\x9b" +
	"\xba\x85\xd8\x97,\x81\xb2\xa3\xbf\xceh;\xc5\xc4\x09\xfa" +
	"u@Wm\x0d^5\xc2\x0fa\xfd@W;#\xd6" +
	"\x0ft\xb5\x07b\xfd\xb0\xb3\xa8\xad\x88\xdd?\x09\xaa\x8b" +
	"\x0d~\x08\x1b\x03v\xf5\xa0\xcc\xc6\x80]\xbd\xfbbc" +
	"\xc0\xce\xc27\x0b\xec\xb1IZ\xab\xe2m\x0f\x87\x03Z" +
	"\x1fe\xbe\x0b\x05$\x10n\xd7\x0e\xcdu\xa0\x9c\xe9\xfe" +
	"F\xbb\xe1\x005M\xd4,\x95\xeaO\xe0uj\xd4\x04" +
	"^\x87F\xad\x9eh\xa2\xda\x15j\xa9\x7f\x92\x89hS" +
	"Yc\x09b35j\x82\xd8\x0c\x8d\x9a \x96U\xc5" +
	"\xc6\xccb]\x0a\xd1\xbb$\xe1^\xca\xec\x14 .E" +
	"\xd4\xc4\x90\x8e\xaft\x09\xb6\x98){\x89B\xc7\xb7\x00" +
	"\xfdy=iSb\x12\xaf\xc5\x92\xae\x88\x0c6 $" +
	"\xa2(DR/\xc9\x18{8\x04\xb0\xfe?\x12\xa80" +
	"\xbd\x7f\x82\x06\x8d!\x16-\xdaRA\xe98\x92\xceK" +
	"^t\xfc\xffC\x83C\xed\x01K3=d%\xbf]" +
	"(LB2\x93\xd9R\xe8\x01\x0cP\x93\x8d\x88\xe2\xb6" +
	"\xa5\xeaP\x8d\xecQ\xc9\x18\xc1&\xb2\xdad\xb0\xddR" +
	"$\xa5\x0f\xb8\x1f:#Q\xaa\x9e\"\xd8S\xae-+" +
	"\x8df\xd4\x1b\x95\xa5\xb4\xe7\xab\x1b\\L\xb7\x08^|" +
	"\xe6Js\xf1\x01-\x08\x11\xbe\xc9\xc9\xaf\xccte\x9e" +
	"\xe6\x8a\"\xf5\xfet\xf6LX\xde\x14!B\x92\xd7\xf6" +
	"\xa0z\xff3\x09\x9a\xa8\x10\xf0L\x17d\x89\xb1\xfb\x17" +
	"\xa7\x84\xf8\x8d\xaeh\xf5\xabi\"\\\xe7e\xc9\xa4\xf0" +
	"\xff\x05\x00\x00\xff\xffYgL\xc7"

func init() {
	schemas.Register(schema_832bcc6686a26d56,
		0x85257b30d6edf8c5,
		0x8748bc095e10cb5d,
		0x87c33f2330feb3d8,
		0x8821cdb23640783a,
		0x8a165fb4d71bf3a2,
		0x8e5322c1e9282534,
		0x8fae7b41c61fc890,
		0x93c99951eacc72ff,
		0x9430ab12c496d40c,
		0x94bf7df83408218d,
		0x95befe3f14606e6b,
		0x97e38948c61f878d,
		0x9ab599979b02ac59,
		0x9b37d729b9dd7b9d,
		0x9b8f27ba05e255c8,
		0x9d3032ff86043b75,
		0xa465f9502fd11e97,
		0xa8bf13fef2674866,
		0xabaedf5f7817c820,
		0xabd055422a4d7df1,
		0xad87da456fb0ebb9,
		0xb1ac056ed7647011,
		0xb1f0385d845e367f,
		0xb61ee2ecff34ca73,
		0xb72b6dc625baa6a4,
		0xc7da65f9a2f20ba2,
		0xc95babe3bd394d2d,
		0xcbdc765fd5dff7ba,
		0xcc4411e60ba9c498,
		0xccb3b2e3603826e0,
		0xce44aee2d9e25049,
		0xcf9beaca1cc180c8,
		0xd636fba4f188dabe,
		0xd8bccf6e60a73791,
		0xd98c608877d9cb8d,
		0xddd1416669fb7613,
		0xde2a1a960863c11c,
		0xde50aebbad57549d,
		0xde9ed43cfaa83093,
		0xe1a2d1d51107bead,
		0xe1c9eac512335361,
		0xe508a29c83a059f8,
		0xe54e10aede55c7b1,
		0xe55d85fc1bf82f21,
		0xe5817f849ff906dc,
		0xe684eb3aef1a6859,
		0xe7711aada4bed56b,
		0xea26e9973bd6a0d9,
		0xf14fad09425d081c,
		0xf58782f48a121998,
		0xf705dc45c94766fd,
		0xf7ff4414476c186a,
		0xfca3742893be4cde)
}

var x_832bcc6686a26d56 = []byte{
	0, 0, 0, 0, 2, 0, 0, 0,
	0, 0, 0, 0, 1, 0, 0, 0,
	223, 7, 8, 27, 0, 0, 0, 0,
	0, 0, 0, 0, 4, 0, 0, 0,
	1, 0, 0, 0, 23, 0, 0, 0,
	8, 0, 0, 0, 1, 0, 0, 0,
	223, 7, 8, 27, 0, 0, 0, 0,
	223, 7, 8, 28, 0, 0, 0, 0,
	0, 0, 0, 0, 3, 0, 0, 0,
	0, 0, 0, 0, 1, 0, 1, 0,
	42, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 3, 0, 0, 0,
	0, 0, 0, 0, 1, 0, 1, 0,
	42, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}
