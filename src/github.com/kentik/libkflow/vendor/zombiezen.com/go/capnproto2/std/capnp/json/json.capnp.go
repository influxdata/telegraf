package json

// AUTO GENERATED - DO NOT EDIT

import (
	math "math"
	strconv "strconv"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type JsonValue struct{ capnp.Struct }
type JsonValue_Which uint16

const (
	JsonValue_Which_null    JsonValue_Which = 0
	JsonValue_Which_boolean JsonValue_Which = 1
	JsonValue_Which_number  JsonValue_Which = 2
	JsonValue_Which_string_ JsonValue_Which = 3
	JsonValue_Which_array   JsonValue_Which = 4
	JsonValue_Which_object  JsonValue_Which = 5
	JsonValue_Which_call    JsonValue_Which = 6
)

func (w JsonValue_Which) String() string {
	const s = "nullbooleannumberstring_arrayobjectcall"
	switch w {
	case JsonValue_Which_null:
		return s[0:4]
	case JsonValue_Which_boolean:
		return s[4:11]
	case JsonValue_Which_number:
		return s[11:17]
	case JsonValue_Which_string_:
		return s[17:24]
	case JsonValue_Which_array:
		return s[24:29]
	case JsonValue_Which_object:
		return s[29:35]
	case JsonValue_Which_call:
		return s[35:39]

	}
	return "JsonValue_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// JsonValue_TypeID is the unique identifier for the type JsonValue.
const JsonValue_TypeID = 0x8825ffaa852cda72

func NewJsonValue(s *capnp.Segment) (JsonValue, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1})
	return JsonValue{st}, err
}

func NewRootJsonValue(s *capnp.Segment) (JsonValue, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1})
	return JsonValue{st}, err
}

func ReadRootJsonValue(msg *capnp.Message) (JsonValue, error) {
	root, err := msg.RootPtr()
	return JsonValue{root.Struct()}, err
}

func (s JsonValue) String() string {
	str, _ := text.Marshal(0x8825ffaa852cda72, s.Struct)
	return str
}

func (s JsonValue) Which() JsonValue_Which {
	return JsonValue_Which(s.Struct.Uint16(0))
}
func (s JsonValue) SetNull() {
	s.Struct.SetUint16(0, 0)

}

func (s JsonValue) Boolean() bool {
	return s.Struct.Bit(16)
}

func (s JsonValue) SetBoolean(v bool) {
	s.Struct.SetUint16(0, 1)
	s.Struct.SetBit(16, v)
}

func (s JsonValue) Number() float64 {
	return math.Float64frombits(s.Struct.Uint64(8))
}

func (s JsonValue) SetNumber(v float64) {
	s.Struct.SetUint16(0, 2)
	s.Struct.SetUint64(8, math.Float64bits(v))
}

func (s JsonValue) String_() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s JsonValue) HasString_() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JsonValue) String_Bytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s JsonValue) SetString_(v string) error {
	s.Struct.SetUint16(0, 3)
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s JsonValue) Array() (JsonValue_List, error) {
	p, err := s.Struct.Ptr(0)
	return JsonValue_List{List: p.List()}, err
}

func (s JsonValue) HasArray() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JsonValue) SetArray(v JsonValue_List) error {
	s.Struct.SetUint16(0, 4)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewArray sets the array field to a newly
// allocated JsonValue_List, preferring placement in s's segment.
func (s JsonValue) NewArray(n int32) (JsonValue_List, error) {
	s.Struct.SetUint16(0, 4)
	l, err := NewJsonValue_List(s.Struct.Segment(), n)
	if err != nil {
		return JsonValue_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s JsonValue) Object() (JsonValue_Field_List, error) {
	p, err := s.Struct.Ptr(0)
	return JsonValue_Field_List{List: p.List()}, err
}

func (s JsonValue) HasObject() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JsonValue) SetObject(v JsonValue_Field_List) error {
	s.Struct.SetUint16(0, 5)
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewObject sets the object field to a newly
// allocated JsonValue_Field_List, preferring placement in s's segment.
func (s JsonValue) NewObject(n int32) (JsonValue_Field_List, error) {
	s.Struct.SetUint16(0, 5)
	l, err := NewJsonValue_Field_List(s.Struct.Segment(), n)
	if err != nil {
		return JsonValue_Field_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s JsonValue) Call() (JsonValue_Call, error) {
	p, err := s.Struct.Ptr(0)
	return JsonValue_Call{Struct: p.Struct()}, err
}

func (s JsonValue) HasCall() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JsonValue) SetCall(v JsonValue_Call) error {
	s.Struct.SetUint16(0, 6)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewCall sets the call field to a newly
// allocated JsonValue_Call struct, preferring placement in s's segment.
func (s JsonValue) NewCall() (JsonValue_Call, error) {
	s.Struct.SetUint16(0, 6)
	ss, err := NewJsonValue_Call(s.Struct.Segment())
	if err != nil {
		return JsonValue_Call{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// JsonValue_List is a list of JsonValue.
type JsonValue_List struct{ capnp.List }

// NewJsonValue creates a new list of JsonValue.
func NewJsonValue_List(s *capnp.Segment, sz int32) (JsonValue_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1}, sz)
	return JsonValue_List{l}, err
}

func (s JsonValue_List) At(i int) JsonValue { return JsonValue{s.List.Struct(i)} }

func (s JsonValue_List) Set(i int, v JsonValue) error { return s.List.SetStruct(i, v.Struct) }

// JsonValue_Promise is a wrapper for a JsonValue promised by a client call.
type JsonValue_Promise struct{ *capnp.Pipeline }

func (p JsonValue_Promise) Struct() (JsonValue, error) {
	s, err := p.Pipeline.Struct()
	return JsonValue{s}, err
}

func (p JsonValue_Promise) Call() JsonValue_Call_Promise {
	return JsonValue_Call_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type JsonValue_Field struct{ capnp.Struct }

// JsonValue_Field_TypeID is the unique identifier for the type JsonValue_Field.
const JsonValue_Field_TypeID = 0xc27855d853a937cc

func NewJsonValue_Field(s *capnp.Segment) (JsonValue_Field, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return JsonValue_Field{st}, err
}

func NewRootJsonValue_Field(s *capnp.Segment) (JsonValue_Field, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return JsonValue_Field{st}, err
}

func ReadRootJsonValue_Field(msg *capnp.Message) (JsonValue_Field, error) {
	root, err := msg.RootPtr()
	return JsonValue_Field{root.Struct()}, err
}

func (s JsonValue_Field) String() string {
	str, _ := text.Marshal(0xc27855d853a937cc, s.Struct)
	return str
}

func (s JsonValue_Field) Name() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s JsonValue_Field) HasName() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JsonValue_Field) NameBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s JsonValue_Field) SetName(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s JsonValue_Field) Value() (JsonValue, error) {
	p, err := s.Struct.Ptr(1)
	return JsonValue{Struct: p.Struct()}, err
}

func (s JsonValue_Field) HasValue() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s JsonValue_Field) SetValue(v JsonValue) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewValue sets the value field to a newly
// allocated JsonValue struct, preferring placement in s's segment.
func (s JsonValue_Field) NewValue() (JsonValue, error) {
	ss, err := NewJsonValue(s.Struct.Segment())
	if err != nil {
		return JsonValue{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// JsonValue_Field_List is a list of JsonValue_Field.
type JsonValue_Field_List struct{ capnp.List }

// NewJsonValue_Field creates a new list of JsonValue_Field.
func NewJsonValue_Field_List(s *capnp.Segment, sz int32) (JsonValue_Field_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return JsonValue_Field_List{l}, err
}

func (s JsonValue_Field_List) At(i int) JsonValue_Field { return JsonValue_Field{s.List.Struct(i)} }

func (s JsonValue_Field_List) Set(i int, v JsonValue_Field) error {
	return s.List.SetStruct(i, v.Struct)
}

// JsonValue_Field_Promise is a wrapper for a JsonValue_Field promised by a client call.
type JsonValue_Field_Promise struct{ *capnp.Pipeline }

func (p JsonValue_Field_Promise) Struct() (JsonValue_Field, error) {
	s, err := p.Pipeline.Struct()
	return JsonValue_Field{s}, err
}

func (p JsonValue_Field_Promise) Value() JsonValue_Promise {
	return JsonValue_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type JsonValue_Call struct{ capnp.Struct }

// JsonValue_Call_TypeID is the unique identifier for the type JsonValue_Call.
const JsonValue_Call_TypeID = 0x9bbf84153dd4bb60

func NewJsonValue_Call(s *capnp.Segment) (JsonValue_Call, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return JsonValue_Call{st}, err
}

func NewRootJsonValue_Call(s *capnp.Segment) (JsonValue_Call, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return JsonValue_Call{st}, err
}

func ReadRootJsonValue_Call(msg *capnp.Message) (JsonValue_Call, error) {
	root, err := msg.RootPtr()
	return JsonValue_Call{root.Struct()}, err
}

func (s JsonValue_Call) String() string {
	str, _ := text.Marshal(0x9bbf84153dd4bb60, s.Struct)
	return str
}

func (s JsonValue_Call) Function() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s JsonValue_Call) HasFunction() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s JsonValue_Call) FunctionBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s JsonValue_Call) SetFunction(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s JsonValue_Call) Params() (JsonValue_List, error) {
	p, err := s.Struct.Ptr(1)
	return JsonValue_List{List: p.List()}, err
}

func (s JsonValue_Call) HasParams() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s JsonValue_Call) SetParams(v JsonValue_List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewParams sets the params field to a newly
// allocated JsonValue_List, preferring placement in s's segment.
func (s JsonValue_Call) NewParams(n int32) (JsonValue_List, error) {
	l, err := NewJsonValue_List(s.Struct.Segment(), n)
	if err != nil {
		return JsonValue_List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

// JsonValue_Call_List is a list of JsonValue_Call.
type JsonValue_Call_List struct{ capnp.List }

// NewJsonValue_Call creates a new list of JsonValue_Call.
func NewJsonValue_Call_List(s *capnp.Segment, sz int32) (JsonValue_Call_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return JsonValue_Call_List{l}, err
}

func (s JsonValue_Call_List) At(i int) JsonValue_Call { return JsonValue_Call{s.List.Struct(i)} }

func (s JsonValue_Call_List) Set(i int, v JsonValue_Call) error { return s.List.SetStruct(i, v.Struct) }

// JsonValue_Call_Promise is a wrapper for a JsonValue_Call promised by a client call.
type JsonValue_Call_Promise struct{ *capnp.Pipeline }

func (p JsonValue_Call_Promise) Struct() (JsonValue_Call, error) {
	s, err := p.Pipeline.Struct()
	return JsonValue_Call{s}, err
}

const schema_8ef99297a43a5e34 = "x\xdat\x92?hSQ\x18\xc5\xef\xb9\xf7\xbd\xb4\xa5" +
	"\xad\xc93):(]\x14\xb5\xd4\xda\xc6\x82\x10\x90h" +
	"\xfd\x83t\x90>\x83\x8e\xda\x97\xf4))//\xe1%" +
	"1:uQ\xd0AEA\x1c\x9c\x04\x97vrP\xb0" +
	"h\xd1\x16G'q\x90N\x0e.\x82\x8b\x9bU\xeb\xf5" +
	"\\C\x93P\xecp\xe1\xde\xdf\xf7\xbd{\xbe{\xce\x1b" +
	"=\x89cr\xccn(!\xdc\xfdvLG\xab\xc37" +
	"\x17\xf4\xde[\xc2\xed\x85\xd4\xe3\x173O\x1f=X\xbb" +
	"+N\xa1\xabK\x88\xe4\x13,$\xe7\xb1O\x88\xc3K" +
	"\xb8\x07\x01=\xfd\xea\xe3\xd1\x81\x1bo\x1e\x0bg\x00\xed" +
	"omi\x9ao\xab\x0f\xc9\x87\xca\xec\xee\xab\x06{\xdf" +
	"\x1f\x99\xcf}:\x7fm\xe5\x7f\xbd\xb0V\x93\xfd\x96\xd9" +
	"\xf5X\x0dqV\xcfV\xcb\xe1H\xc1\xab \xacd&" +
	"\xb9\xbf\x10\xf7\x82\xba\xefv\xa3\xf3\x9a\x9et\x87\xbe=" +
	"4x\xba\xe8\x073\xf1\x13^\x10\xb8\xbb\x94\xd5\xa7\xb5" +
	"\x05!\x9c\x17C|\xda3\x05\xf7\xb5\xc4n\xfc\xd1\x89" +
	"\x14\x0c^\x9c ~N\xfc\x96X\xaek\xa4 \x89\x97" +
	"2\xc4/\x89\xdfI\xf4\xab\xdf:\x05:\xe3,g\x9c" +
	"\xe5A\xf73\xf17b\xeb\x17\xb1E\xfc5\xcd\xe6/" +
	"\x0a\xe7@j\xff$\xb5I\xd7\xcd\x15?\x14r)\x83" +
	"ck\xc41>\xcc\x01\x07\xc9\xf5\x81\x85\x9d,\xc4\xc3" +
	"z\x10\x88\xd8\\\xbe\\\x0e|/\xe4L\x92\x0b\xd9\xb0" +
	"^\xca\xfb\x11zy\xec\xe5\xb1Z\x8b\x8a\xe1\x15\xd7b" +
	"\x18\xdf\xef\x1c\xda\xb1}zqE\xb8\x96\xc4\xf1\x04\xd0" +
	"G1L\xcc5[.\x09A \xb90\xe8E\x91w" +
	"\x1d\xdb\x04\xa6\x14\x90h{-``\xb6\x9c\x9f\xf5\x0b" +
	"\xb5v\xbd\xe5h\xb3\x1e/\xd0A\xe2\x96\xb7\xc4\x09\xe6" +
	"\xb7\x91\x89\xdc\xc8\xc4D2B\xbb\x11L\x01n\xb7\xa2" +
	"%\xff\x1c?0i~&\x9a5.\xe1\x00M\xbf\xc7" +
	"\x8c+\xc3\x84g$\xf4\xe5zX\xa8\x15\xcb\xa1h\x0f" +
	"\x9d\xadx\x91W\xaan9\xf5\x16\xf2\xcc\\\x053\x9b" +
	"\xf4M\xe2{(5\xda\xa1\x7f0\xdd\x1e*\x1ez%" +
	"\xbf\xe5\xd6Us\xd1&A\xbe\xf7o\x00\x00\x00\xff\xff" +
	"\x9f\xdb\xc7\x93"

func init() {
	schemas.Register(schema_8ef99297a43a5e34,
		0x8825ffaa852cda72,
		0x9bbf84153dd4bb60,
		0xc27855d853a937cc)
}
