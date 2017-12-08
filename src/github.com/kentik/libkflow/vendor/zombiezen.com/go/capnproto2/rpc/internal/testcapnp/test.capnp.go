package testcapnp

// AUTO GENERATED - DO NOT EDIT

import (
	context "golang.org/x/net/context"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
	server "zombiezen.com/go/capnproto2/server"
)

type Handle struct{ Client capnp.Client }

type Handle_Server interface {
}

func Handle_ServerToClient(s Handle_Server) Handle {
	c, _ := s.(server.Closer)
	return Handle{Client: server.New(Handle_Methods(nil, s), c)}
}

func Handle_Methods(methods []server.Method, s Handle_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 0)
	}

	return methods
}

type HandleFactory struct{ Client capnp.Client }

func (c HandleFactory) NewHandle(ctx context.Context, params func(HandleFactory_newHandle_Params) error, opts ...capnp.CallOption) HandleFactory_newHandle_Results_Promise {
	if c.Client == nil {
		return HandleFactory_newHandle_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x8491a7fe75fe0bce,
			MethodID:      0,
			InterfaceName: "test.capnp:HandleFactory",
			MethodName:    "newHandle",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(HandleFactory_newHandle_Params{Struct: s}) }
	}
	return HandleFactory_newHandle_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type HandleFactory_Server interface {
	NewHandle(HandleFactory_newHandle) error
}

func HandleFactory_ServerToClient(s HandleFactory_Server) HandleFactory {
	c, _ := s.(server.Closer)
	return HandleFactory{Client: server.New(HandleFactory_Methods(nil, s), c)}
}

func HandleFactory_Methods(methods []server.Method, s HandleFactory_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x8491a7fe75fe0bce,
			MethodID:      0,
			InterfaceName: "test.capnp:HandleFactory",
			MethodName:    "newHandle",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := HandleFactory_newHandle{c, opts, HandleFactory_newHandle_Params{Struct: p}, HandleFactory_newHandle_Results{Struct: r}}
			return s.NewHandle(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// HandleFactory_newHandle holds the arguments for a server call to HandleFactory.newHandle.
type HandleFactory_newHandle struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  HandleFactory_newHandle_Params
	Results HandleFactory_newHandle_Results
}

type HandleFactory_newHandle_Params struct{ capnp.Struct }

// HandleFactory_newHandle_Params_TypeID is the unique identifier for the type HandleFactory_newHandle_Params.
const HandleFactory_newHandle_Params_TypeID = 0x99821793f0a50b5e

func NewHandleFactory_newHandle_Params(s *capnp.Segment) (HandleFactory_newHandle_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return HandleFactory_newHandle_Params{st}, err
}

func NewRootHandleFactory_newHandle_Params(s *capnp.Segment) (HandleFactory_newHandle_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return HandleFactory_newHandle_Params{st}, err
}

func ReadRootHandleFactory_newHandle_Params(msg *capnp.Message) (HandleFactory_newHandle_Params, error) {
	root, err := msg.RootPtr()
	return HandleFactory_newHandle_Params{root.Struct()}, err
}

func (s HandleFactory_newHandle_Params) String() string {
	str, _ := text.Marshal(0x99821793f0a50b5e, s.Struct)
	return str
}

// HandleFactory_newHandle_Params_List is a list of HandleFactory_newHandle_Params.
type HandleFactory_newHandle_Params_List struct{ capnp.List }

// NewHandleFactory_newHandle_Params creates a new list of HandleFactory_newHandle_Params.
func NewHandleFactory_newHandle_Params_List(s *capnp.Segment, sz int32) (HandleFactory_newHandle_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return HandleFactory_newHandle_Params_List{l}, err
}

func (s HandleFactory_newHandle_Params_List) At(i int) HandleFactory_newHandle_Params {
	return HandleFactory_newHandle_Params{s.List.Struct(i)}
}

func (s HandleFactory_newHandle_Params_List) Set(i int, v HandleFactory_newHandle_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// HandleFactory_newHandle_Params_Promise is a wrapper for a HandleFactory_newHandle_Params promised by a client call.
type HandleFactory_newHandle_Params_Promise struct{ *capnp.Pipeline }

func (p HandleFactory_newHandle_Params_Promise) Struct() (HandleFactory_newHandle_Params, error) {
	s, err := p.Pipeline.Struct()
	return HandleFactory_newHandle_Params{s}, err
}

type HandleFactory_newHandle_Results struct{ capnp.Struct }

// HandleFactory_newHandle_Results_TypeID is the unique identifier for the type HandleFactory_newHandle_Results.
const HandleFactory_newHandle_Results_TypeID = 0xd57b5111c59d048c

func NewHandleFactory_newHandle_Results(s *capnp.Segment) (HandleFactory_newHandle_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HandleFactory_newHandle_Results{st}, err
}

func NewRootHandleFactory_newHandle_Results(s *capnp.Segment) (HandleFactory_newHandle_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HandleFactory_newHandle_Results{st}, err
}

func ReadRootHandleFactory_newHandle_Results(msg *capnp.Message) (HandleFactory_newHandle_Results, error) {
	root, err := msg.RootPtr()
	return HandleFactory_newHandle_Results{root.Struct()}, err
}

func (s HandleFactory_newHandle_Results) String() string {
	str, _ := text.Marshal(0xd57b5111c59d048c, s.Struct)
	return str
}

func (s HandleFactory_newHandle_Results) Handle() Handle {
	p, _ := s.Struct.Ptr(0)
	return Handle{Client: p.Interface().Client()}
}

func (s HandleFactory_newHandle_Results) HasHandle() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HandleFactory_newHandle_Results) SetHandle(v Handle) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// HandleFactory_newHandle_Results_List is a list of HandleFactory_newHandle_Results.
type HandleFactory_newHandle_Results_List struct{ capnp.List }

// NewHandleFactory_newHandle_Results creates a new list of HandleFactory_newHandle_Results.
func NewHandleFactory_newHandle_Results_List(s *capnp.Segment, sz int32) (HandleFactory_newHandle_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HandleFactory_newHandle_Results_List{l}, err
}

func (s HandleFactory_newHandle_Results_List) At(i int) HandleFactory_newHandle_Results {
	return HandleFactory_newHandle_Results{s.List.Struct(i)}
}

func (s HandleFactory_newHandle_Results_List) Set(i int, v HandleFactory_newHandle_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// HandleFactory_newHandle_Results_Promise is a wrapper for a HandleFactory_newHandle_Results promised by a client call.
type HandleFactory_newHandle_Results_Promise struct{ *capnp.Pipeline }

func (p HandleFactory_newHandle_Results_Promise) Struct() (HandleFactory_newHandle_Results, error) {
	s, err := p.Pipeline.Struct()
	return HandleFactory_newHandle_Results{s}, err
}

func (p HandleFactory_newHandle_Results_Promise) Handle() Handle {
	return Handle{Client: p.Pipeline.GetPipeline(0).Client()}
}

type Hanger struct{ Client capnp.Client }

func (c Hanger) Hang(ctx context.Context, params func(Hanger_hang_Params) error, opts ...capnp.CallOption) Hanger_hang_Results_Promise {
	if c.Client == nil {
		return Hanger_hang_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x8ae08044aae8a26e,
			MethodID:      0,
			InterfaceName: "test.capnp:Hanger",
			MethodName:    "hang",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Hanger_hang_Params{Struct: s}) }
	}
	return Hanger_hang_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Hanger_Server interface {
	Hang(Hanger_hang) error
}

func Hanger_ServerToClient(s Hanger_Server) Hanger {
	c, _ := s.(server.Closer)
	return Hanger{Client: server.New(Hanger_Methods(nil, s), c)}
}

func Hanger_Methods(methods []server.Method, s Hanger_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x8ae08044aae8a26e,
			MethodID:      0,
			InterfaceName: "test.capnp:Hanger",
			MethodName:    "hang",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Hanger_hang{c, opts, Hanger_hang_Params{Struct: p}, Hanger_hang_Results{Struct: r}}
			return s.Hang(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 0},
	})

	return methods
}

// Hanger_hang holds the arguments for a server call to Hanger.hang.
type Hanger_hang struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Hanger_hang_Params
	Results Hanger_hang_Results
}

type Hanger_hang_Params struct{ capnp.Struct }

// Hanger_hang_Params_TypeID is the unique identifier for the type Hanger_hang_Params.
const Hanger_hang_Params_TypeID = 0xb4512d1c0c85f06f

func NewHanger_hang_Params(s *capnp.Segment) (Hanger_hang_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hanger_hang_Params{st}, err
}

func NewRootHanger_hang_Params(s *capnp.Segment) (Hanger_hang_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hanger_hang_Params{st}, err
}

func ReadRootHanger_hang_Params(msg *capnp.Message) (Hanger_hang_Params, error) {
	root, err := msg.RootPtr()
	return Hanger_hang_Params{root.Struct()}, err
}

func (s Hanger_hang_Params) String() string {
	str, _ := text.Marshal(0xb4512d1c0c85f06f, s.Struct)
	return str
}

// Hanger_hang_Params_List is a list of Hanger_hang_Params.
type Hanger_hang_Params_List struct{ capnp.List }

// NewHanger_hang_Params creates a new list of Hanger_hang_Params.
func NewHanger_hang_Params_List(s *capnp.Segment, sz int32) (Hanger_hang_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Hanger_hang_Params_List{l}, err
}

func (s Hanger_hang_Params_List) At(i int) Hanger_hang_Params {
	return Hanger_hang_Params{s.List.Struct(i)}
}

func (s Hanger_hang_Params_List) Set(i int, v Hanger_hang_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Hanger_hang_Params_Promise is a wrapper for a Hanger_hang_Params promised by a client call.
type Hanger_hang_Params_Promise struct{ *capnp.Pipeline }

func (p Hanger_hang_Params_Promise) Struct() (Hanger_hang_Params, error) {
	s, err := p.Pipeline.Struct()
	return Hanger_hang_Params{s}, err
}

type Hanger_hang_Results struct{ capnp.Struct }

// Hanger_hang_Results_TypeID is the unique identifier for the type Hanger_hang_Results.
const Hanger_hang_Results_TypeID = 0xb9c9455b55ed47b0

func NewHanger_hang_Results(s *capnp.Segment) (Hanger_hang_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hanger_hang_Results{st}, err
}

func NewRootHanger_hang_Results(s *capnp.Segment) (Hanger_hang_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hanger_hang_Results{st}, err
}

func ReadRootHanger_hang_Results(msg *capnp.Message) (Hanger_hang_Results, error) {
	root, err := msg.RootPtr()
	return Hanger_hang_Results{root.Struct()}, err
}

func (s Hanger_hang_Results) String() string {
	str, _ := text.Marshal(0xb9c9455b55ed47b0, s.Struct)
	return str
}

// Hanger_hang_Results_List is a list of Hanger_hang_Results.
type Hanger_hang_Results_List struct{ capnp.List }

// NewHanger_hang_Results creates a new list of Hanger_hang_Results.
func NewHanger_hang_Results_List(s *capnp.Segment, sz int32) (Hanger_hang_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Hanger_hang_Results_List{l}, err
}

func (s Hanger_hang_Results_List) At(i int) Hanger_hang_Results {
	return Hanger_hang_Results{s.List.Struct(i)}
}

func (s Hanger_hang_Results_List) Set(i int, v Hanger_hang_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Hanger_hang_Results_Promise is a wrapper for a Hanger_hang_Results promised by a client call.
type Hanger_hang_Results_Promise struct{ *capnp.Pipeline }

func (p Hanger_hang_Results_Promise) Struct() (Hanger_hang_Results, error) {
	s, err := p.Pipeline.Struct()
	return Hanger_hang_Results{s}, err
}

type CallOrder struct{ Client capnp.Client }

func (c CallOrder) GetCallSequence(ctx context.Context, params func(CallOrder_getCallSequence_Params) error, opts ...capnp.CallOption) CallOrder_getCallSequence_Results_Promise {
	if c.Client == nil {
		return CallOrder_getCallSequence_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x92c5ca8314cdd2a5,
			MethodID:      0,
			InterfaceName: "test.capnp:CallOrder",
			MethodName:    "getCallSequence",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 8, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(CallOrder_getCallSequence_Params{Struct: s}) }
	}
	return CallOrder_getCallSequence_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type CallOrder_Server interface {
	GetCallSequence(CallOrder_getCallSequence) error
}

func CallOrder_ServerToClient(s CallOrder_Server) CallOrder {
	c, _ := s.(server.Closer)
	return CallOrder{Client: server.New(CallOrder_Methods(nil, s), c)}
}

func CallOrder_Methods(methods []server.Method, s CallOrder_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x92c5ca8314cdd2a5,
			MethodID:      0,
			InterfaceName: "test.capnp:CallOrder",
			MethodName:    "getCallSequence",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := CallOrder_getCallSequence{c, opts, CallOrder_getCallSequence_Params{Struct: p}, CallOrder_getCallSequence_Results{Struct: r}}
			return s.GetCallSequence(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 8, PointerCount: 0},
	})

	return methods
}

// CallOrder_getCallSequence holds the arguments for a server call to CallOrder.getCallSequence.
type CallOrder_getCallSequence struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  CallOrder_getCallSequence_Params
	Results CallOrder_getCallSequence_Results
}

type CallOrder_getCallSequence_Params struct{ capnp.Struct }

// CallOrder_getCallSequence_Params_TypeID is the unique identifier for the type CallOrder_getCallSequence_Params.
const CallOrder_getCallSequence_Params_TypeID = 0x993e61d6a54c166f

func NewCallOrder_getCallSequence_Params(s *capnp.Segment) (CallOrder_getCallSequence_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return CallOrder_getCallSequence_Params{st}, err
}

func NewRootCallOrder_getCallSequence_Params(s *capnp.Segment) (CallOrder_getCallSequence_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return CallOrder_getCallSequence_Params{st}, err
}

func ReadRootCallOrder_getCallSequence_Params(msg *capnp.Message) (CallOrder_getCallSequence_Params, error) {
	root, err := msg.RootPtr()
	return CallOrder_getCallSequence_Params{root.Struct()}, err
}

func (s CallOrder_getCallSequence_Params) String() string {
	str, _ := text.Marshal(0x993e61d6a54c166f, s.Struct)
	return str
}

func (s CallOrder_getCallSequence_Params) Expected() uint32 {
	return s.Struct.Uint32(0)
}

func (s CallOrder_getCallSequence_Params) SetExpected(v uint32) {
	s.Struct.SetUint32(0, v)
}

// CallOrder_getCallSequence_Params_List is a list of CallOrder_getCallSequence_Params.
type CallOrder_getCallSequence_Params_List struct{ capnp.List }

// NewCallOrder_getCallSequence_Params creates a new list of CallOrder_getCallSequence_Params.
func NewCallOrder_getCallSequence_Params_List(s *capnp.Segment, sz int32) (CallOrder_getCallSequence_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return CallOrder_getCallSequence_Params_List{l}, err
}

func (s CallOrder_getCallSequence_Params_List) At(i int) CallOrder_getCallSequence_Params {
	return CallOrder_getCallSequence_Params{s.List.Struct(i)}
}

func (s CallOrder_getCallSequence_Params_List) Set(i int, v CallOrder_getCallSequence_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// CallOrder_getCallSequence_Params_Promise is a wrapper for a CallOrder_getCallSequence_Params promised by a client call.
type CallOrder_getCallSequence_Params_Promise struct{ *capnp.Pipeline }

func (p CallOrder_getCallSequence_Params_Promise) Struct() (CallOrder_getCallSequence_Params, error) {
	s, err := p.Pipeline.Struct()
	return CallOrder_getCallSequence_Params{s}, err
}

type CallOrder_getCallSequence_Results struct{ capnp.Struct }

// CallOrder_getCallSequence_Results_TypeID is the unique identifier for the type CallOrder_getCallSequence_Results.
const CallOrder_getCallSequence_Results_TypeID = 0x88f809ef7f873e58

func NewCallOrder_getCallSequence_Results(s *capnp.Segment) (CallOrder_getCallSequence_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return CallOrder_getCallSequence_Results{st}, err
}

func NewRootCallOrder_getCallSequence_Results(s *capnp.Segment) (CallOrder_getCallSequence_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return CallOrder_getCallSequence_Results{st}, err
}

func ReadRootCallOrder_getCallSequence_Results(msg *capnp.Message) (CallOrder_getCallSequence_Results, error) {
	root, err := msg.RootPtr()
	return CallOrder_getCallSequence_Results{root.Struct()}, err
}

func (s CallOrder_getCallSequence_Results) String() string {
	str, _ := text.Marshal(0x88f809ef7f873e58, s.Struct)
	return str
}

func (s CallOrder_getCallSequence_Results) N() uint32 {
	return s.Struct.Uint32(0)
}

func (s CallOrder_getCallSequence_Results) SetN(v uint32) {
	s.Struct.SetUint32(0, v)
}

// CallOrder_getCallSequence_Results_List is a list of CallOrder_getCallSequence_Results.
type CallOrder_getCallSequence_Results_List struct{ capnp.List }

// NewCallOrder_getCallSequence_Results creates a new list of CallOrder_getCallSequence_Results.
func NewCallOrder_getCallSequence_Results_List(s *capnp.Segment, sz int32) (CallOrder_getCallSequence_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return CallOrder_getCallSequence_Results_List{l}, err
}

func (s CallOrder_getCallSequence_Results_List) At(i int) CallOrder_getCallSequence_Results {
	return CallOrder_getCallSequence_Results{s.List.Struct(i)}
}

func (s CallOrder_getCallSequence_Results_List) Set(i int, v CallOrder_getCallSequence_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// CallOrder_getCallSequence_Results_Promise is a wrapper for a CallOrder_getCallSequence_Results promised by a client call.
type CallOrder_getCallSequence_Results_Promise struct{ *capnp.Pipeline }

func (p CallOrder_getCallSequence_Results_Promise) Struct() (CallOrder_getCallSequence_Results, error) {
	s, err := p.Pipeline.Struct()
	return CallOrder_getCallSequence_Results{s}, err
}

type Echoer struct{ Client capnp.Client }

func (c Echoer) Echo(ctx context.Context, params func(Echoer_echo_Params) error, opts ...capnp.CallOption) Echoer_echo_Results_Promise {
	if c.Client == nil {
		return Echoer_echo_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x841756c6a41b2a45,
			MethodID:      0,
			InterfaceName: "test.capnp:Echoer",
			MethodName:    "echo",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Echoer_echo_Params{Struct: s}) }
	}
	return Echoer_echo_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}
func (c Echoer) GetCallSequence(ctx context.Context, params func(CallOrder_getCallSequence_Params) error, opts ...capnp.CallOption) CallOrder_getCallSequence_Results_Promise {
	if c.Client == nil {
		return CallOrder_getCallSequence_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x92c5ca8314cdd2a5,
			MethodID:      0,
			InterfaceName: "test.capnp:CallOrder",
			MethodName:    "getCallSequence",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 8, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(CallOrder_getCallSequence_Params{Struct: s}) }
	}
	return CallOrder_getCallSequence_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Echoer_Server interface {
	Echo(Echoer_echo) error

	GetCallSequence(CallOrder_getCallSequence) error
}

func Echoer_ServerToClient(s Echoer_Server) Echoer {
	c, _ := s.(server.Closer)
	return Echoer{Client: server.New(Echoer_Methods(nil, s), c)}
}

func Echoer_Methods(methods []server.Method, s Echoer_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 2)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x841756c6a41b2a45,
			MethodID:      0,
			InterfaceName: "test.capnp:Echoer",
			MethodName:    "echo",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Echoer_echo{c, opts, Echoer_echo_Params{Struct: p}, Echoer_echo_Results{Struct: r}}
			return s.Echo(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x92c5ca8314cdd2a5,
			MethodID:      0,
			InterfaceName: "test.capnp:CallOrder",
			MethodName:    "getCallSequence",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := CallOrder_getCallSequence{c, opts, CallOrder_getCallSequence_Params{Struct: p}, CallOrder_getCallSequence_Results{Struct: r}}
			return s.GetCallSequence(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 8, PointerCount: 0},
	})

	return methods
}

// Echoer_echo holds the arguments for a server call to Echoer.echo.
type Echoer_echo struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Echoer_echo_Params
	Results Echoer_echo_Results
}

type Echoer_echo_Params struct{ capnp.Struct }

// Echoer_echo_Params_TypeID is the unique identifier for the type Echoer_echo_Params.
const Echoer_echo_Params_TypeID = 0xe96a45cad5d1a1d3

func NewEchoer_echo_Params(s *capnp.Segment) (Echoer_echo_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echoer_echo_Params{st}, err
}

func NewRootEchoer_echo_Params(s *capnp.Segment) (Echoer_echo_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echoer_echo_Params{st}, err
}

func ReadRootEchoer_echo_Params(msg *capnp.Message) (Echoer_echo_Params, error) {
	root, err := msg.RootPtr()
	return Echoer_echo_Params{root.Struct()}, err
}

func (s Echoer_echo_Params) String() string {
	str, _ := text.Marshal(0xe96a45cad5d1a1d3, s.Struct)
	return str
}

func (s Echoer_echo_Params) Cap() CallOrder {
	p, _ := s.Struct.Ptr(0)
	return CallOrder{Client: p.Interface().Client()}
}

func (s Echoer_echo_Params) HasCap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Echoer_echo_Params) SetCap(v CallOrder) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// Echoer_echo_Params_List is a list of Echoer_echo_Params.
type Echoer_echo_Params_List struct{ capnp.List }

// NewEchoer_echo_Params creates a new list of Echoer_echo_Params.
func NewEchoer_echo_Params_List(s *capnp.Segment, sz int32) (Echoer_echo_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Echoer_echo_Params_List{l}, err
}

func (s Echoer_echo_Params_List) At(i int) Echoer_echo_Params {
	return Echoer_echo_Params{s.List.Struct(i)}
}

func (s Echoer_echo_Params_List) Set(i int, v Echoer_echo_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Echoer_echo_Params_Promise is a wrapper for a Echoer_echo_Params promised by a client call.
type Echoer_echo_Params_Promise struct{ *capnp.Pipeline }

func (p Echoer_echo_Params_Promise) Struct() (Echoer_echo_Params, error) {
	s, err := p.Pipeline.Struct()
	return Echoer_echo_Params{s}, err
}

func (p Echoer_echo_Params_Promise) Cap() CallOrder {
	return CallOrder{Client: p.Pipeline.GetPipeline(0).Client()}
}

type Echoer_echo_Results struct{ capnp.Struct }

// Echoer_echo_Results_TypeID is the unique identifier for the type Echoer_echo_Results.
const Echoer_echo_Results_TypeID = 0x8b45b4847bd839c8

func NewEchoer_echo_Results(s *capnp.Segment) (Echoer_echo_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echoer_echo_Results{st}, err
}

func NewRootEchoer_echo_Results(s *capnp.Segment) (Echoer_echo_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Echoer_echo_Results{st}, err
}

func ReadRootEchoer_echo_Results(msg *capnp.Message) (Echoer_echo_Results, error) {
	root, err := msg.RootPtr()
	return Echoer_echo_Results{root.Struct()}, err
}

func (s Echoer_echo_Results) String() string {
	str, _ := text.Marshal(0x8b45b4847bd839c8, s.Struct)
	return str
}

func (s Echoer_echo_Results) Cap() CallOrder {
	p, _ := s.Struct.Ptr(0)
	return CallOrder{Client: p.Interface().Client()}
}

func (s Echoer_echo_Results) HasCap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Echoer_echo_Results) SetCap(v CallOrder) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// Echoer_echo_Results_List is a list of Echoer_echo_Results.
type Echoer_echo_Results_List struct{ capnp.List }

// NewEchoer_echo_Results creates a new list of Echoer_echo_Results.
func NewEchoer_echo_Results_List(s *capnp.Segment, sz int32) (Echoer_echo_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Echoer_echo_Results_List{l}, err
}

func (s Echoer_echo_Results_List) At(i int) Echoer_echo_Results {
	return Echoer_echo_Results{s.List.Struct(i)}
}

func (s Echoer_echo_Results_List) Set(i int, v Echoer_echo_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Echoer_echo_Results_Promise is a wrapper for a Echoer_echo_Results promised by a client call.
type Echoer_echo_Results_Promise struct{ *capnp.Pipeline }

func (p Echoer_echo_Results_Promise) Struct() (Echoer_echo_Results, error) {
	s, err := p.Pipeline.Struct()
	return Echoer_echo_Results{s}, err
}

func (p Echoer_echo_Results_Promise) Cap() CallOrder {
	return CallOrder{Client: p.Pipeline.GetPipeline(0).Client()}
}

type PingPong struct{ Client capnp.Client }

func (c PingPong) EchoNum(ctx context.Context, params func(PingPong_echoNum_Params) error, opts ...capnp.CallOption) PingPong_echoNum_Results_Promise {
	if c.Client == nil {
		return PingPong_echoNum_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xf004c474c2f8ee7a,
			MethodID:      0,
			InterfaceName: "test.capnp:PingPong",
			MethodName:    "echoNum",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 8, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(PingPong_echoNum_Params{Struct: s}) }
	}
	return PingPong_echoNum_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type PingPong_Server interface {
	EchoNum(PingPong_echoNum) error
}

func PingPong_ServerToClient(s PingPong_Server) PingPong {
	c, _ := s.(server.Closer)
	return PingPong{Client: server.New(PingPong_Methods(nil, s), c)}
}

func PingPong_Methods(methods []server.Method, s PingPong_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xf004c474c2f8ee7a,
			MethodID:      0,
			InterfaceName: "test.capnp:PingPong",
			MethodName:    "echoNum",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := PingPong_echoNum{c, opts, PingPong_echoNum_Params{Struct: p}, PingPong_echoNum_Results{Struct: r}}
			return s.EchoNum(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 8, PointerCount: 0},
	})

	return methods
}

// PingPong_echoNum holds the arguments for a server call to PingPong.echoNum.
type PingPong_echoNum struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  PingPong_echoNum_Params
	Results PingPong_echoNum_Results
}

type PingPong_echoNum_Params struct{ capnp.Struct }

// PingPong_echoNum_Params_TypeID is the unique identifier for the type PingPong_echoNum_Params.
const PingPong_echoNum_Params_TypeID = 0xd797e0a99edf0921

func NewPingPong_echoNum_Params(s *capnp.Segment) (PingPong_echoNum_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return PingPong_echoNum_Params{st}, err
}

func NewRootPingPong_echoNum_Params(s *capnp.Segment) (PingPong_echoNum_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return PingPong_echoNum_Params{st}, err
}

func ReadRootPingPong_echoNum_Params(msg *capnp.Message) (PingPong_echoNum_Params, error) {
	root, err := msg.RootPtr()
	return PingPong_echoNum_Params{root.Struct()}, err
}

func (s PingPong_echoNum_Params) String() string {
	str, _ := text.Marshal(0xd797e0a99edf0921, s.Struct)
	return str
}

func (s PingPong_echoNum_Params) N() int32 {
	return int32(s.Struct.Uint32(0))
}

func (s PingPong_echoNum_Params) SetN(v int32) {
	s.Struct.SetUint32(0, uint32(v))
}

// PingPong_echoNum_Params_List is a list of PingPong_echoNum_Params.
type PingPong_echoNum_Params_List struct{ capnp.List }

// NewPingPong_echoNum_Params creates a new list of PingPong_echoNum_Params.
func NewPingPong_echoNum_Params_List(s *capnp.Segment, sz int32) (PingPong_echoNum_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return PingPong_echoNum_Params_List{l}, err
}

func (s PingPong_echoNum_Params_List) At(i int) PingPong_echoNum_Params {
	return PingPong_echoNum_Params{s.List.Struct(i)}
}

func (s PingPong_echoNum_Params_List) Set(i int, v PingPong_echoNum_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// PingPong_echoNum_Params_Promise is a wrapper for a PingPong_echoNum_Params promised by a client call.
type PingPong_echoNum_Params_Promise struct{ *capnp.Pipeline }

func (p PingPong_echoNum_Params_Promise) Struct() (PingPong_echoNum_Params, error) {
	s, err := p.Pipeline.Struct()
	return PingPong_echoNum_Params{s}, err
}

type PingPong_echoNum_Results struct{ capnp.Struct }

// PingPong_echoNum_Results_TypeID is the unique identifier for the type PingPong_echoNum_Results.
const PingPong_echoNum_Results_TypeID = 0x85ddfd96db252600

func NewPingPong_echoNum_Results(s *capnp.Segment) (PingPong_echoNum_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return PingPong_echoNum_Results{st}, err
}

func NewRootPingPong_echoNum_Results(s *capnp.Segment) (PingPong_echoNum_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return PingPong_echoNum_Results{st}, err
}

func ReadRootPingPong_echoNum_Results(msg *capnp.Message) (PingPong_echoNum_Results, error) {
	root, err := msg.RootPtr()
	return PingPong_echoNum_Results{root.Struct()}, err
}

func (s PingPong_echoNum_Results) String() string {
	str, _ := text.Marshal(0x85ddfd96db252600, s.Struct)
	return str
}

func (s PingPong_echoNum_Results) N() int32 {
	return int32(s.Struct.Uint32(0))
}

func (s PingPong_echoNum_Results) SetN(v int32) {
	s.Struct.SetUint32(0, uint32(v))
}

// PingPong_echoNum_Results_List is a list of PingPong_echoNum_Results.
type PingPong_echoNum_Results_List struct{ capnp.List }

// NewPingPong_echoNum_Results creates a new list of PingPong_echoNum_Results.
func NewPingPong_echoNum_Results_List(s *capnp.Segment, sz int32) (PingPong_echoNum_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return PingPong_echoNum_Results_List{l}, err
}

func (s PingPong_echoNum_Results_List) At(i int) PingPong_echoNum_Results {
	return PingPong_echoNum_Results{s.List.Struct(i)}
}

func (s PingPong_echoNum_Results_List) Set(i int, v PingPong_echoNum_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// PingPong_echoNum_Results_Promise is a wrapper for a PingPong_echoNum_Results promised by a client call.
type PingPong_echoNum_Results_Promise struct{ *capnp.Pipeline }

func (p PingPong_echoNum_Results_Promise) Struct() (PingPong_echoNum_Results, error) {
	s, err := p.Pipeline.Struct()
	return PingPong_echoNum_Results{s}, err
}

type Adder struct{ Client capnp.Client }

func (c Adder) Add(ctx context.Context, params func(Adder_add_Params) error, opts ...capnp.CallOption) Adder_add_Results_Promise {
	if c.Client == nil {
		return Adder_add_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x8f9cac550b1bf41f,
			MethodID:      0,
			InterfaceName: "test.capnp:Adder",
			MethodName:    "add",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 8, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Adder_add_Params{Struct: s}) }
	}
	return Adder_add_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Adder_Server interface {
	Add(Adder_add) error
}

func Adder_ServerToClient(s Adder_Server) Adder {
	c, _ := s.(server.Closer)
	return Adder{Client: server.New(Adder_Methods(nil, s), c)}
}

func Adder_Methods(methods []server.Method, s Adder_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x8f9cac550b1bf41f,
			MethodID:      0,
			InterfaceName: "test.capnp:Adder",
			MethodName:    "add",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Adder_add{c, opts, Adder_add_Params{Struct: p}, Adder_add_Results{Struct: r}}
			return s.Add(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 8, PointerCount: 0},
	})

	return methods
}

// Adder_add holds the arguments for a server call to Adder.add.
type Adder_add struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Adder_add_Params
	Results Adder_add_Results
}

type Adder_add_Params struct{ capnp.Struct }

// Adder_add_Params_TypeID is the unique identifier for the type Adder_add_Params.
const Adder_add_Params_TypeID = 0x9ed99eb5024ed6ef

func NewAdder_add_Params(s *capnp.Segment) (Adder_add_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Adder_add_Params{st}, err
}

func NewRootAdder_add_Params(s *capnp.Segment) (Adder_add_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Adder_add_Params{st}, err
}

func ReadRootAdder_add_Params(msg *capnp.Message) (Adder_add_Params, error) {
	root, err := msg.RootPtr()
	return Adder_add_Params{root.Struct()}, err
}

func (s Adder_add_Params) String() string {
	str, _ := text.Marshal(0x9ed99eb5024ed6ef, s.Struct)
	return str
}

func (s Adder_add_Params) A() int32 {
	return int32(s.Struct.Uint32(0))
}

func (s Adder_add_Params) SetA(v int32) {
	s.Struct.SetUint32(0, uint32(v))
}

func (s Adder_add_Params) B() int32 {
	return int32(s.Struct.Uint32(4))
}

func (s Adder_add_Params) SetB(v int32) {
	s.Struct.SetUint32(4, uint32(v))
}

// Adder_add_Params_List is a list of Adder_add_Params.
type Adder_add_Params_List struct{ capnp.List }

// NewAdder_add_Params creates a new list of Adder_add_Params.
func NewAdder_add_Params_List(s *capnp.Segment, sz int32) (Adder_add_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return Adder_add_Params_List{l}, err
}

func (s Adder_add_Params_List) At(i int) Adder_add_Params { return Adder_add_Params{s.List.Struct(i)} }

func (s Adder_add_Params_List) Set(i int, v Adder_add_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Adder_add_Params_Promise is a wrapper for a Adder_add_Params promised by a client call.
type Adder_add_Params_Promise struct{ *capnp.Pipeline }

func (p Adder_add_Params_Promise) Struct() (Adder_add_Params, error) {
	s, err := p.Pipeline.Struct()
	return Adder_add_Params{s}, err
}

type Adder_add_Results struct{ capnp.Struct }

// Adder_add_Results_TypeID is the unique identifier for the type Adder_add_Results.
const Adder_add_Results_TypeID = 0xa74428796527f253

func NewAdder_add_Results(s *capnp.Segment) (Adder_add_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Adder_add_Results{st}, err
}

func NewRootAdder_add_Results(s *capnp.Segment) (Adder_add_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Adder_add_Results{st}, err
}

func ReadRootAdder_add_Results(msg *capnp.Message) (Adder_add_Results, error) {
	root, err := msg.RootPtr()
	return Adder_add_Results{root.Struct()}, err
}

func (s Adder_add_Results) String() string {
	str, _ := text.Marshal(0xa74428796527f253, s.Struct)
	return str
}

func (s Adder_add_Results) Result() int32 {
	return int32(s.Struct.Uint32(0))
}

func (s Adder_add_Results) SetResult(v int32) {
	s.Struct.SetUint32(0, uint32(v))
}

// Adder_add_Results_List is a list of Adder_add_Results.
type Adder_add_Results_List struct{ capnp.List }

// NewAdder_add_Results creates a new list of Adder_add_Results.
func NewAdder_add_Results_List(s *capnp.Segment, sz int32) (Adder_add_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return Adder_add_Results_List{l}, err
}

func (s Adder_add_Results_List) At(i int) Adder_add_Results {
	return Adder_add_Results{s.List.Struct(i)}
}

func (s Adder_add_Results_List) Set(i int, v Adder_add_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Adder_add_Results_Promise is a wrapper for a Adder_add_Results promised by a client call.
type Adder_add_Results_Promise struct{ *capnp.Pipeline }

func (p Adder_add_Results_Promise) Struct() (Adder_add_Results, error) {
	s, err := p.Pipeline.Struct()
	return Adder_add_Results{s}, err
}

const schema_ef12a34b9807e19c = "x\xda\x9cU_h\x1c\xd5\x17>g\xee\xbd\xbfIH" +
	"\x7f\x84\x9b\x1b\x8d\xb5Bi\x88\xda\x16\xb3Tj\x1f\x1a" +
	"\xb0\x89\xda\xed\x8a\xd5\xba\x93\x12\xf1?Lw.\x9b\x96" +
	"\xdd\xd9uv\x97\x1a\xfbP[[\x94\xaaH#\xfe\x09" +
	"\xb4\x81TCi\xf1\xc1\x87\xfa\x90W1b\x03\x15\xa3" +
	"\xa4Z$\xd8\x0aE+\xd4\xc4\xa8P\x0af\xe4\xde\xec" +
	"\xdd\x9d\xddM\x8b\xfa\xb0\xb03\xe7\x9b\xef|\xe7;\xdf" +
	"\x9d\xd9\xb0\xcd\xea\xb3\xeees6\x80\xf3,\xfb_8" +
	"\xfd\xd5\xf6\x9f~\x9fu\x0f\x00o!\xe1\xb1K\xf6\xfb" +
	"\xdb?h\x9b\x03@1K\x86\xc5eb\x03\x88K\xc4" +
	"V?\x800\xbe~\xd5\x87\x9f?\xd1q\xa8\x01|\x96" +
	"\x0c\x8bi\x0d>G\x12\xe2O\xd2\x01\x10~\xd9\xb2X" +
	"Z<y\xb4\x11|\x99L\x89\xdf4\xf8*I\x88\x95" +
	"\xd4\x06X\xbc\xeb\xce\xef\xdf\xfdk\xf6\xb0\xd3\x8e\x08\xa0" +
	"\xeelD\xda\x89\x80\xa2\x99\xf6\x02\x86Onyu\xff" +
	"\\\xf3\xb5\xd7\xc0\xb9\xa5\x02XG\xf7(\xc0&\x0d\xf0" +
	"O\xfc|z\xeb\xcb\x17\x8f44\x1b\xa0\xc3\xe29\x85" +
	"\x17O\xd1\x848\xa0\xfe\x85_l\xfen\xdf\xa13\xf1" +
	"\xd7\x81\xb7!\x00CU\xddM\xaf\x03\x8a\xac&[\xfd" +
	"\xc7\xaa\x96\x81\x8f\x8e\xbd\xd5@\xf6&=\"\xde\xd1d" +
	"GiBLh\xb2\xf1\xaf\xcf\xb5\xbf259\xdc\x00" +
	"\x1e\xa3\xa7\xc5)\x0d\x1e\xa7\x091\xad\xc1\xb9[\x1f\x1d" +
	"?\xefn\x19\x89\xce1Aw\xa99&u\xeb\xe7[" +
	"\xc6\xe7\xdf\xee88\x02|\xa5\xa9_\xa1\xfd\x084\x9c" +
	";\xbf\xc3\xfad\xf4\xc2(8\xdc<*f\xe8\x8f\x80" +
	"\xe2\x02\xdd\x0b\x18\xee\\\xb8[\x0e\xad\xddz2Z\xdf" +
	"\xc4~\x01\x14\x9b\x99b\xce\xcd\x1f^qG\xb7sf" +
	"i\xe8%G\xd8\x02\xd0\xf0\xe3\xc4\xd5\x81g\xe2g'" +
	"\"\x85\xfb\xd9u\xa0\xe1\x1b\xf4\xf8$w\xf6\xcd,i" +
	"\xd16m\\\xc3\x9eVb\xbb5\xe5\x9a\xe6\x1fFO" +
	"]|\xef[\x88\xac\xed1v\xbb\x02\x0ch\xc07c" +
	"\xd33S\xf1=W\xa2F\x97\xd8\x02\xa0\x18\xd2\xf5\x97" +
	"~\xbd\xf6i\xf13:\xdf\xe0\xdd\x08;!\xc6\x98\x82" +
	"\x1fg\x091\xc9l\xb8',\xcaB1\x96r\xf3\xe8" +
	"\xe7{\x1ev}/\x832\x89\x98$,\x89XS\x8c" +
	"\xa7\x06s\x12\x83$\xa2C\x09\x03\xa8\xc8@\xb3x\xce" +
	"\xd7\x83\xc5\x99\xdd*S\x83\xb9>t(bu\x8f\x00" +
	"\x156\xcb\xb4\x92\xdb\xdcT1\x17\x0c\x01TY\xcd\xaa" +
	"\xd0\xf8\xc4y?X\xbc\xd9\x0e}\xb9W?\x05(\xfb" +
	"0\xaa\x8e\xf8\xf9\x9e\xe4n?\x9d\xcc\xf9\xe9\x98\xea\xbd" +
	"\xa3\x94\xed\xea\x97\x85\x92\x9d)\x16\x1cJ(\x00E\x00" +
	"\xfe\xff6\x00\xa7\x89\xa0\xd3n!\xfaH\xc1B\x0aU" +
	"\x1a\xea\xe7{\x1er3\x99\xc7\x03O\x06\xb1\xb4,\xaa" +
	"\x8b\x9d\xf2\x85\x92\xf4SR\xf1\xb5\x96n\xca\xd7\x04\x16" +
	"6\x01\xd6;\x9a\xae1\xcd\xe4\x05M>*\xa6\x0d\xba" +
	"~\xbav.\xcb\xb8\x1e\xe8\xa9\xba\xfa{e\xa1^B" +
	"gU\x82\xadz\xf2\x88\xe3\x88\xbcN\xce\x03\x9e'\x83" +
	"\xa8\xdb&\xfehr\xcey\xa7\x96c\xbb\x9eW\xab\x06" +
	"\x8d=\xad\xca\x9f\xc8D\xe5\xc3\x87\xe6m\xc2\xf9\xc1\xa5" +
	"\x85\x19\x0b\xb1\xec!\xd4\x12\xde\xdc\xef\xa4\x1b\xd8n\xb6" +
	"f\xd6G\x00\x9c\x15\x04\x9d\xdb,\x0c\xe5\x8by\x99*" +
	"J\x0f\x00\x1a\x8c\xa7\xf5\xf9\x8a\x99\xe8hZ7\x8b\x85" +
	"\x1a\x8f\xb5+1\xd7\xf3*M\x9b*M\xd7\xa9\x1dw" +
	"\x11t6X\xc8\x11\xf5\x89\xe4\xdd\xea\xe6Z\x82\xce}" +
	"\x16\xa2k\x82\x84\xbb\x1a\"U\xcb\xbe\\\x84z\xaa\xfb" +
	"\xeb\x0d\xf4z\x97%\xd19\x0ab*#]\xc9\xd5j" +
	"\x86\xc2\x0d\xeb&&\xff\xc8\x0fuL2\xe4\xc6\x9a\x06" +
	"5\x0cy\xf5\x8bV\x17\xabe\x0f\x9fr\x99d\xff\xc5" +
	"\xd9\xab\x8fzy\xc6\xff\x9et-\xc9\xce\xf9\xe9jP" +
	"\xcd{\x15\xa1\xfcY\xe4\xfcA\x9d\xf5\xfde\xd9:\x9e" +
	"\x7f\x07\x00\x00\xff\xff<\xf6/\x9f"

func init() {
	schemas.Register(schema_ef12a34b9807e19c,
		0x8161ddf3e74bd0d1,
		0x841756c6a41b2a45,
		0x8491a7fe75fe0bce,
		0x85ddfd96db252600,
		0x88f809ef7f873e58,
		0x8ae08044aae8a26e,
		0x8b45b4847bd839c8,
		0x8f9cac550b1bf41f,
		0x92c5ca8314cdd2a5,
		0x993e61d6a54c166f,
		0x99821793f0a50b5e,
		0x9ed99eb5024ed6ef,
		0xa74428796527f253,
		0xb4512d1c0c85f06f,
		0xb9c9455b55ed47b0,
		0xd57b5111c59d048c,
		0xd797e0a99edf0921,
		0xe96a45cad5d1a1d3,
		0xf004c474c2f8ee7a)
}
