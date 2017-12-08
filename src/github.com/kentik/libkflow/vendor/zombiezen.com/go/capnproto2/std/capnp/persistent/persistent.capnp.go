package persistent

// AUTO GENERATED - DO NOT EDIT

import (
	context "golang.org/x/net/context"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
	server "zombiezen.com/go/capnproto2/server"
)

const PersistentAnnotation = uint64(0xf622595091cafb67)

type Persistent struct{ Client capnp.Client }

func (c Persistent) Save(ctx context.Context, params func(Persistent_SaveParams) error, opts ...capnp.CallOption) Persistent_SaveResults_Promise {
	if c.Client == nil {
		return Persistent_SaveResults_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xc8cb212fcd9f5691,
			MethodID:      0,
			InterfaceName: "persistent.capnp:Persistent",
			MethodName:    "save",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Persistent_SaveParams{Struct: s}) }
	}
	return Persistent_SaveResults_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Persistent_Server interface {
	Save(Persistent_save) error
}

func Persistent_ServerToClient(s Persistent_Server) Persistent {
	c, _ := s.(server.Closer)
	return Persistent{Client: server.New(Persistent_Methods(nil, s), c)}
}

func Persistent_Methods(methods []server.Method, s Persistent_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xc8cb212fcd9f5691,
			MethodID:      0,
			InterfaceName: "persistent.capnp:Persistent",
			MethodName:    "save",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Persistent_save{c, opts, Persistent_SaveParams{Struct: p}, Persistent_SaveResults{Struct: r}}
			return s.Save(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// Persistent_save holds the arguments for a server call to Persistent.save.
type Persistent_save struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Persistent_SaveParams
	Results Persistent_SaveResults
}

type Persistent_SaveParams struct{ capnp.Struct }

// Persistent_SaveParams_TypeID is the unique identifier for the type Persistent_SaveParams.
const Persistent_SaveParams_TypeID = 0xf76fba59183073a5

func NewPersistent_SaveParams(s *capnp.Segment) (Persistent_SaveParams, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Persistent_SaveParams{st}, err
}

func NewRootPersistent_SaveParams(s *capnp.Segment) (Persistent_SaveParams, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Persistent_SaveParams{st}, err
}

func ReadRootPersistent_SaveParams(msg *capnp.Message) (Persistent_SaveParams, error) {
	root, err := msg.RootPtr()
	return Persistent_SaveParams{root.Struct()}, err
}

func (s Persistent_SaveParams) String() string {
	str, _ := text.Marshal(0xf76fba59183073a5, s.Struct)
	return str
}

func (s Persistent_SaveParams) SealFor() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Persistent_SaveParams) HasSealFor() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Persistent_SaveParams) SealForPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Persistent_SaveParams) SetSealFor(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s Persistent_SaveParams) SetSealForPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

// Persistent_SaveParams_List is a list of Persistent_SaveParams.
type Persistent_SaveParams_List struct{ capnp.List }

// NewPersistent_SaveParams creates a new list of Persistent_SaveParams.
func NewPersistent_SaveParams_List(s *capnp.Segment, sz int32) (Persistent_SaveParams_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Persistent_SaveParams_List{l}, err
}

func (s Persistent_SaveParams_List) At(i int) Persistent_SaveParams {
	return Persistent_SaveParams{s.List.Struct(i)}
}

func (s Persistent_SaveParams_List) Set(i int, v Persistent_SaveParams) error {
	return s.List.SetStruct(i, v.Struct)
}

// Persistent_SaveParams_Promise is a wrapper for a Persistent_SaveParams promised by a client call.
type Persistent_SaveParams_Promise struct{ *capnp.Pipeline }

func (p Persistent_SaveParams_Promise) Struct() (Persistent_SaveParams, error) {
	s, err := p.Pipeline.Struct()
	return Persistent_SaveParams{s}, err
}

func (p Persistent_SaveParams_Promise) SealFor() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type Persistent_SaveResults struct{ capnp.Struct }

// Persistent_SaveResults_TypeID is the unique identifier for the type Persistent_SaveResults.
const Persistent_SaveResults_TypeID = 0xb76848c18c40efbf

func NewPersistent_SaveResults(s *capnp.Segment) (Persistent_SaveResults, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Persistent_SaveResults{st}, err
}

func NewRootPersistent_SaveResults(s *capnp.Segment) (Persistent_SaveResults, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Persistent_SaveResults{st}, err
}

func ReadRootPersistent_SaveResults(msg *capnp.Message) (Persistent_SaveResults, error) {
	root, err := msg.RootPtr()
	return Persistent_SaveResults{root.Struct()}, err
}

func (s Persistent_SaveResults) String() string {
	str, _ := text.Marshal(0xb76848c18c40efbf, s.Struct)
	return str
}

func (s Persistent_SaveResults) SturdyRef() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Persistent_SaveResults) HasSturdyRef() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Persistent_SaveResults) SturdyRefPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Persistent_SaveResults) SetSturdyRef(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s Persistent_SaveResults) SetSturdyRefPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

// Persistent_SaveResults_List is a list of Persistent_SaveResults.
type Persistent_SaveResults_List struct{ capnp.List }

// NewPersistent_SaveResults creates a new list of Persistent_SaveResults.
func NewPersistent_SaveResults_List(s *capnp.Segment, sz int32) (Persistent_SaveResults_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Persistent_SaveResults_List{l}, err
}

func (s Persistent_SaveResults_List) At(i int) Persistent_SaveResults {
	return Persistent_SaveResults{s.List.Struct(i)}
}

func (s Persistent_SaveResults_List) Set(i int, v Persistent_SaveResults) error {
	return s.List.SetStruct(i, v.Struct)
}

// Persistent_SaveResults_Promise is a wrapper for a Persistent_SaveResults promised by a client call.
type Persistent_SaveResults_Promise struct{ *capnp.Pipeline }

func (p Persistent_SaveResults_Promise) Struct() (Persistent_SaveResults, error) {
	s, err := p.Pipeline.Struct()
	return Persistent_SaveResults{s}, err
}

func (p Persistent_SaveResults_Promise) SturdyRef() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type RealmGateway struct{ Client capnp.Client }

func (c RealmGateway) Import(ctx context.Context, params func(RealmGateway_import_Params) error, opts ...capnp.CallOption) Persistent_SaveResults_Promise {
	if c.Client == nil {
		return Persistent_SaveResults_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x84ff286cd00a3ed4,
			MethodID:      0,
			InterfaceName: "persistent.capnp:RealmGateway",
			MethodName:    "import",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 2}
		call.ParamsFunc = func(s capnp.Struct) error { return params(RealmGateway_import_Params{Struct: s}) }
	}
	return Persistent_SaveResults_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}
func (c RealmGateway) Export(ctx context.Context, params func(RealmGateway_export_Params) error, opts ...capnp.CallOption) Persistent_SaveResults_Promise {
	if c.Client == nil {
		return Persistent_SaveResults_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x84ff286cd00a3ed4,
			MethodID:      1,
			InterfaceName: "persistent.capnp:RealmGateway",
			MethodName:    "export",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 2}
		call.ParamsFunc = func(s capnp.Struct) error { return params(RealmGateway_export_Params{Struct: s}) }
	}
	return Persistent_SaveResults_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type RealmGateway_Server interface {
	Import(RealmGateway_import) error

	Export(RealmGateway_export) error
}

func RealmGateway_ServerToClient(s RealmGateway_Server) RealmGateway {
	c, _ := s.(server.Closer)
	return RealmGateway{Client: server.New(RealmGateway_Methods(nil, s), c)}
}

func RealmGateway_Methods(methods []server.Method, s RealmGateway_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 2)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x84ff286cd00a3ed4,
			MethodID:      0,
			InterfaceName: "persistent.capnp:RealmGateway",
			MethodName:    "import",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := RealmGateway_import{c, opts, RealmGateway_import_Params{Struct: p}, Persistent_SaveResults{Struct: r}}
			return s.Import(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x84ff286cd00a3ed4,
			MethodID:      1,
			InterfaceName: "persistent.capnp:RealmGateway",
			MethodName:    "export",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := RealmGateway_export{c, opts, RealmGateway_export_Params{Struct: p}, Persistent_SaveResults{Struct: r}}
			return s.Export(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// RealmGateway_import holds the arguments for a server call to RealmGateway.import.
type RealmGateway_import struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  RealmGateway_import_Params
	Results Persistent_SaveResults
}

// RealmGateway_export holds the arguments for a server call to RealmGateway.export.
type RealmGateway_export struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  RealmGateway_export_Params
	Results Persistent_SaveResults
}

type RealmGateway_import_Params struct{ capnp.Struct }

// RealmGateway_import_Params_TypeID is the unique identifier for the type RealmGateway_import_Params.
const RealmGateway_import_Params_TypeID = 0xf0c2cc1d3909574d

func NewRealmGateway_import_Params(s *capnp.Segment) (RealmGateway_import_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return RealmGateway_import_Params{st}, err
}

func NewRootRealmGateway_import_Params(s *capnp.Segment) (RealmGateway_import_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return RealmGateway_import_Params{st}, err
}

func ReadRootRealmGateway_import_Params(msg *capnp.Message) (RealmGateway_import_Params, error) {
	root, err := msg.RootPtr()
	return RealmGateway_import_Params{root.Struct()}, err
}

func (s RealmGateway_import_Params) String() string {
	str, _ := text.Marshal(0xf0c2cc1d3909574d, s.Struct)
	return str
}

func (s RealmGateway_import_Params) Cap() Persistent {
	p, _ := s.Struct.Ptr(0)
	return Persistent{Client: p.Interface().Client()}
}

func (s RealmGateway_import_Params) HasCap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s RealmGateway_import_Params) SetCap(v Persistent) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

func (s RealmGateway_import_Params) Params() (Persistent_SaveParams, error) {
	p, err := s.Struct.Ptr(1)
	return Persistent_SaveParams{Struct: p.Struct()}, err
}

func (s RealmGateway_import_Params) HasParams() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s RealmGateway_import_Params) SetParams(v Persistent_SaveParams) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewParams sets the params field to a newly
// allocated Persistent_SaveParams struct, preferring placement in s's segment.
func (s RealmGateway_import_Params) NewParams() (Persistent_SaveParams, error) {
	ss, err := NewPersistent_SaveParams(s.Struct.Segment())
	if err != nil {
		return Persistent_SaveParams{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// RealmGateway_import_Params_List is a list of RealmGateway_import_Params.
type RealmGateway_import_Params_List struct{ capnp.List }

// NewRealmGateway_import_Params creates a new list of RealmGateway_import_Params.
func NewRealmGateway_import_Params_List(s *capnp.Segment, sz int32) (RealmGateway_import_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return RealmGateway_import_Params_List{l}, err
}

func (s RealmGateway_import_Params_List) At(i int) RealmGateway_import_Params {
	return RealmGateway_import_Params{s.List.Struct(i)}
}

func (s RealmGateway_import_Params_List) Set(i int, v RealmGateway_import_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// RealmGateway_import_Params_Promise is a wrapper for a RealmGateway_import_Params promised by a client call.
type RealmGateway_import_Params_Promise struct{ *capnp.Pipeline }

func (p RealmGateway_import_Params_Promise) Struct() (RealmGateway_import_Params, error) {
	s, err := p.Pipeline.Struct()
	return RealmGateway_import_Params{s}, err
}

func (p RealmGateway_import_Params_Promise) Cap() Persistent {
	return Persistent{Client: p.Pipeline.GetPipeline(0).Client()}
}

func (p RealmGateway_import_Params_Promise) Params() Persistent_SaveParams_Promise {
	return Persistent_SaveParams_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type RealmGateway_export_Params struct{ capnp.Struct }

// RealmGateway_export_Params_TypeID is the unique identifier for the type RealmGateway_export_Params.
const RealmGateway_export_Params_TypeID = 0xecafa18b482da3aa

func NewRealmGateway_export_Params(s *capnp.Segment) (RealmGateway_export_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return RealmGateway_export_Params{st}, err
}

func NewRootRealmGateway_export_Params(s *capnp.Segment) (RealmGateway_export_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return RealmGateway_export_Params{st}, err
}

func ReadRootRealmGateway_export_Params(msg *capnp.Message) (RealmGateway_export_Params, error) {
	root, err := msg.RootPtr()
	return RealmGateway_export_Params{root.Struct()}, err
}

func (s RealmGateway_export_Params) String() string {
	str, _ := text.Marshal(0xecafa18b482da3aa, s.Struct)
	return str
}

func (s RealmGateway_export_Params) Cap() Persistent {
	p, _ := s.Struct.Ptr(0)
	return Persistent{Client: p.Interface().Client()}
}

func (s RealmGateway_export_Params) HasCap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s RealmGateway_export_Params) SetCap(v Persistent) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

func (s RealmGateway_export_Params) Params() (Persistent_SaveParams, error) {
	p, err := s.Struct.Ptr(1)
	return Persistent_SaveParams{Struct: p.Struct()}, err
}

func (s RealmGateway_export_Params) HasParams() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s RealmGateway_export_Params) SetParams(v Persistent_SaveParams) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewParams sets the params field to a newly
// allocated Persistent_SaveParams struct, preferring placement in s's segment.
func (s RealmGateway_export_Params) NewParams() (Persistent_SaveParams, error) {
	ss, err := NewPersistent_SaveParams(s.Struct.Segment())
	if err != nil {
		return Persistent_SaveParams{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// RealmGateway_export_Params_List is a list of RealmGateway_export_Params.
type RealmGateway_export_Params_List struct{ capnp.List }

// NewRealmGateway_export_Params creates a new list of RealmGateway_export_Params.
func NewRealmGateway_export_Params_List(s *capnp.Segment, sz int32) (RealmGateway_export_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return RealmGateway_export_Params_List{l}, err
}

func (s RealmGateway_export_Params_List) At(i int) RealmGateway_export_Params {
	return RealmGateway_export_Params{s.List.Struct(i)}
}

func (s RealmGateway_export_Params_List) Set(i int, v RealmGateway_export_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// RealmGateway_export_Params_Promise is a wrapper for a RealmGateway_export_Params promised by a client call.
type RealmGateway_export_Params_Promise struct{ *capnp.Pipeline }

func (p RealmGateway_export_Params_Promise) Struct() (RealmGateway_export_Params, error) {
	s, err := p.Pipeline.Struct()
	return RealmGateway_export_Params{s}, err
}

func (p RealmGateway_export_Params_Promise) Cap() Persistent {
	return Persistent{Client: p.Pipeline.GetPipeline(0).Client()}
}

func (p RealmGateway_export_Params_Promise) Params() Persistent_SaveParams_Promise {
	return Persistent_SaveParams_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

const schema_b8630836983feed7 = "x\xda\xbcSMH\x14o\x18\x7f\x9e\xf9\xd8\xd9\xfd\xa3" +
	"\xb8\xef\x8e\xa2\x7f\xd06E1#\xbf\x0a\"%\xda]" +
	"\xc1\xd4C\xb5\xb3Ba\x104kc\x09\xbb\xb3\xc3\xcc" +
	"\x98z\x88\x0a\x82\x88\xf2\xe0\xad[Dt\x11\xa2c\xd9" +
	"%\x92\x82\xe8\x03\x0a\x0a\xef\x9e\xa3\xacC\x1d:L\xcf" +
	"\xbb\x9f\x93\xabH\x18\x1d\xde\x81y\xde\xe7y\x9f\xe7\xf7" +
	"\xf1\xf4\xadc\\\xe8\x97W%\x00\xed\x90\x1c\xf0>\x1c" +
	"\xf9\xef]f\x8fw\x0d\x18\x13\xbd\xd5\xcf\xb1\xdb\x07\x83" +
	"\x93\x8f\x01 \x8c\xaa&|S\xcf\x08\x0a\x80:!\\" +
	"WuQ\xa1\xd3\xe9=\xfd\x12\xbf\xf5l\xf4\xc2#`" +
	"\xcd\xe8-\x9e\xbc\xf3\xb6\xb7\xf5\xf5K\x90Q\x09\xe3\x81" +
	"K\xe2\x10\xaa\x0b\"/\xb9!\xc6\xc0w_\xfd\xfa\x03" +
	"qM]\x16;)\xf5\xa38\xa2vH\x0a\x9dFo" +
	"\xe9^\xf7\xe8\xcd\xbb\x0f?\x01\xdb\x85\x00\xb2\xc0_\xed" +
	"\x90\xd2\x08\xa8\xf6K\xb3\xf4\xe2\xb1S\xa1\x81\x967+" +
	"\xeb\xfe\x84\x169\x9f\xd0%\xf3\x84\xf3?_-&'" +
	"\xda\xbe\xc3{&\xefF_ST\x1b\x02kjk@" +
	"\xa1\x13\x1d\x1f\x0d\x88T\xe3\xddw\xfa\x9a&\x9e\xe4~" +
	"l\x86f 0\x88\xeaX\x80\xa3\x19\x0e\xc4\xe0\xb8g" +
	"\x19\xb63\xed\xb8\x86`\xba=\x93\xbaeZ\x83)C" +
	"\xcfdG\xf4\xa8k\xcc\xea\xf3ID-(\xca\x00\xe5" +
	"!\xb1D\x16\xeb\x1f\x04H\xec\xc3\xc4adW\x15\xc4" +
	"2\xceJ\xc6\x0c\xcf\xb00q\x05\xd9\x0b%6\x9d\xb5" +
	"r\xb6\xcb0\xaaI\x02VDB\xc2\\\x0a\x96\xa7\xe5" +
	"\xa1N-\x88\x98\xef\xcf\xbfa\xfaF\xd0WG\xe8#" +
	"2\x0a\xfe@\x1cc\xc6\xdc\x8e\x9b\xc8\xd5]\xc4\xdf\xbb" +
	"\x10+\x890\xb2\xda4ci\xd6`\xb3\xffmo\xcc" +
	"t\x0d\xdb\xd43\xa0\xa4\x8c)ox\xce\xffW\xbe\x8b" +
	"\x9e\x985\x0d\xbbr[\xfc/i \x955H\x16#" +
	"\x14\x18\xd7/\x1a)\xc3\x99\xc9\xb8\x0ep5$\x91l" +
	".q8\xb5)\xf2{\x8d\x88Z\x13\x81r\xdc\x19\xfb" +
	"\xdc|\xca\x00\x9c\xca\xd3\xe4\x03\x89\x112E\xb5\xce\xc5" +
	"\x1e\x0a5\xe1\x1c\xf8l\x13:\xed\xdb\x88P\xda\xe3#" +
	"$u[\x071\xebx\xa5y@\xa1\x89h\x1a\xee\x8d" +
	"RiEy\xb6\x97\x94\xaf\xc1D3\xb2n\xa5\xce\xa1" +
	"\x9a*\xf67\x95\x84\x07\x0b\xfc\x06\x91\xc9)\x16\xda\xef" +
	"\x8dW\x90m\xc9W\xd1\xb3y\xcb\xf6\x14,\xd0N\x13" +
	"+z\xd6!eK\x84u\xb5\x11a\xedD\xd8\x9c\x80" +
	"\x0c\xb1>?\x01\xf7\xa8fQ\xf0\xb9\x80\x0a\xbd\x86\xcc" +
	"\xcf]\x1c\xff\x925\x91\x01\xc6,\"1\xeb`\xb8\xc2" +
	"\xf5\x9ft\xd8\xce\x97\x18\xf6\xe9\xbc\x057\x85\x1d\xfc\xd7" +
	"\xdcl;\xf9\xce\xc9\xd9\x8e\xfe\xf0\xa6;`Uv " +
	"\xbf[T\xf4u\xa1\xb71rvy\x05x\xdbD\x13" +
	"b\x0d\xef\xb8\xe4\x95V\x12M7a\x9a9W\xafs" +
	"\xa7s&H\xe5W\xc5\xad\xb67\x96\xcc#\xdb\xb0\xbc" +
	"CD-\xcd\xaf\xd5\x0bx\xd9!\x89\x8e\xe6\xec\x02Q" +
	"\x1b6\xf7W\x00\x00\x00\xff\xff\x1ar\xf0\xc2"

func init() {
	schemas.Register(schema_b8630836983feed7,
		0x84ff286cd00a3ed4,
		0xb76848c18c40efbf,
		0xc8cb212fcd9f5691,
		0xecafa18b482da3aa,
		0xf0c2cc1d3909574d,
		0xf622595091cafb67,
		0xf76fba59183073a5)
}
