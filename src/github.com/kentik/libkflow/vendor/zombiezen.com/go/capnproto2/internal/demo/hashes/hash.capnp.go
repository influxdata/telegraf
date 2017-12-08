package hashes

// AUTO GENERATED - DO NOT EDIT

import (
	context "golang.org/x/net/context"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
	server "zombiezen.com/go/capnproto2/server"
)

type HashFactory struct{ Client capnp.Client }

func (c HashFactory) NewSha1(ctx context.Context, params func(HashFactory_newSha1_Params) error, opts ...capnp.CallOption) HashFactory_newSha1_Results_Promise {
	if c.Client == nil {
		return HashFactory_newSha1_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xaead580f97fddabc,
			MethodID:      0,
			InterfaceName: "hash.capnp:HashFactory",
			MethodName:    "newSha1",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(HashFactory_newSha1_Params{Struct: s}) }
	}
	return HashFactory_newSha1_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type HashFactory_Server interface {
	NewSha1(HashFactory_newSha1) error
}

func HashFactory_ServerToClient(s HashFactory_Server) HashFactory {
	c, _ := s.(server.Closer)
	return HashFactory{Client: server.New(HashFactory_Methods(nil, s), c)}
}

func HashFactory_Methods(methods []server.Method, s HashFactory_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xaead580f97fddabc,
			MethodID:      0,
			InterfaceName: "hash.capnp:HashFactory",
			MethodName:    "newSha1",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := HashFactory_newSha1{c, opts, HashFactory_newSha1_Params{Struct: p}, HashFactory_newSha1_Results{Struct: r}}
			return s.NewSha1(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// HashFactory_newSha1 holds the arguments for a server call to HashFactory.newSha1.
type HashFactory_newSha1 struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  HashFactory_newSha1_Params
	Results HashFactory_newSha1_Results
}

type HashFactory_newSha1_Params struct{ capnp.Struct }

// HashFactory_newSha1_Params_TypeID is the unique identifier for the type HashFactory_newSha1_Params.
const HashFactory_newSha1_Params_TypeID = 0x92b20ad1a58ca0ca

func NewHashFactory_newSha1_Params(s *capnp.Segment) (HashFactory_newSha1_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return HashFactory_newSha1_Params{st}, err
}

func NewRootHashFactory_newSha1_Params(s *capnp.Segment) (HashFactory_newSha1_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return HashFactory_newSha1_Params{st}, err
}

func ReadRootHashFactory_newSha1_Params(msg *capnp.Message) (HashFactory_newSha1_Params, error) {
	root, err := msg.RootPtr()
	return HashFactory_newSha1_Params{root.Struct()}, err
}

func (s HashFactory_newSha1_Params) String() string {
	str, _ := text.Marshal(0x92b20ad1a58ca0ca, s.Struct)
	return str
}

// HashFactory_newSha1_Params_List is a list of HashFactory_newSha1_Params.
type HashFactory_newSha1_Params_List struct{ capnp.List }

// NewHashFactory_newSha1_Params creates a new list of HashFactory_newSha1_Params.
func NewHashFactory_newSha1_Params_List(s *capnp.Segment, sz int32) (HashFactory_newSha1_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return HashFactory_newSha1_Params_List{l}, err
}

func (s HashFactory_newSha1_Params_List) At(i int) HashFactory_newSha1_Params {
	return HashFactory_newSha1_Params{s.List.Struct(i)}
}

func (s HashFactory_newSha1_Params_List) Set(i int, v HashFactory_newSha1_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// HashFactory_newSha1_Params_Promise is a wrapper for a HashFactory_newSha1_Params promised by a client call.
type HashFactory_newSha1_Params_Promise struct{ *capnp.Pipeline }

func (p HashFactory_newSha1_Params_Promise) Struct() (HashFactory_newSha1_Params, error) {
	s, err := p.Pipeline.Struct()
	return HashFactory_newSha1_Params{s}, err
}

type HashFactory_newSha1_Results struct{ capnp.Struct }

// HashFactory_newSha1_Results_TypeID is the unique identifier for the type HashFactory_newSha1_Results.
const HashFactory_newSha1_Results_TypeID = 0xea3e50f7663f7bdf

func NewHashFactory_newSha1_Results(s *capnp.Segment) (HashFactory_newSha1_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HashFactory_newSha1_Results{st}, err
}

func NewRootHashFactory_newSha1_Results(s *capnp.Segment) (HashFactory_newSha1_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return HashFactory_newSha1_Results{st}, err
}

func ReadRootHashFactory_newSha1_Results(msg *capnp.Message) (HashFactory_newSha1_Results, error) {
	root, err := msg.RootPtr()
	return HashFactory_newSha1_Results{root.Struct()}, err
}

func (s HashFactory_newSha1_Results) String() string {
	str, _ := text.Marshal(0xea3e50f7663f7bdf, s.Struct)
	return str
}

func (s HashFactory_newSha1_Results) Hash() Hash {
	p, _ := s.Struct.Ptr(0)
	return Hash{Client: p.Interface().Client()}
}

func (s HashFactory_newSha1_Results) HasHash() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s HashFactory_newSha1_Results) SetHash(v Hash) error {
	if v.Client == nil {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// HashFactory_newSha1_Results_List is a list of HashFactory_newSha1_Results.
type HashFactory_newSha1_Results_List struct{ capnp.List }

// NewHashFactory_newSha1_Results creates a new list of HashFactory_newSha1_Results.
func NewHashFactory_newSha1_Results_List(s *capnp.Segment, sz int32) (HashFactory_newSha1_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return HashFactory_newSha1_Results_List{l}, err
}

func (s HashFactory_newSha1_Results_List) At(i int) HashFactory_newSha1_Results {
	return HashFactory_newSha1_Results{s.List.Struct(i)}
}

func (s HashFactory_newSha1_Results_List) Set(i int, v HashFactory_newSha1_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// HashFactory_newSha1_Results_Promise is a wrapper for a HashFactory_newSha1_Results promised by a client call.
type HashFactory_newSha1_Results_Promise struct{ *capnp.Pipeline }

func (p HashFactory_newSha1_Results_Promise) Struct() (HashFactory_newSha1_Results, error) {
	s, err := p.Pipeline.Struct()
	return HashFactory_newSha1_Results{s}, err
}

func (p HashFactory_newSha1_Results_Promise) Hash() Hash {
	return Hash{Client: p.Pipeline.GetPipeline(0).Client()}
}

type Hash struct{ Client capnp.Client }

func (c Hash) Write(ctx context.Context, params func(Hash_write_Params) error, opts ...capnp.CallOption) Hash_write_Results_Promise {
	if c.Client == nil {
		return Hash_write_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xf29f97dd675a9431,
			MethodID:      0,
			InterfaceName: "hash.capnp:Hash",
			MethodName:    "write",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Hash_write_Params{Struct: s}) }
	}
	return Hash_write_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}
func (c Hash) Sum(ctx context.Context, params func(Hash_sum_Params) error, opts ...capnp.CallOption) Hash_sum_Results_Promise {
	if c.Client == nil {
		return Hash_sum_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xf29f97dd675a9431,
			MethodID:      1,
			InterfaceName: "hash.capnp:Hash",
			MethodName:    "sum",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 0}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Hash_sum_Params{Struct: s}) }
	}
	return Hash_sum_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Hash_Server interface {
	Write(Hash_write) error

	Sum(Hash_sum) error
}

func Hash_ServerToClient(s Hash_Server) Hash {
	c, _ := s.(server.Closer)
	return Hash{Client: server.New(Hash_Methods(nil, s), c)}
}

func Hash_Methods(methods []server.Method, s Hash_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 2)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xf29f97dd675a9431,
			MethodID:      0,
			InterfaceName: "hash.capnp:Hash",
			MethodName:    "write",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Hash_write{c, opts, Hash_write_Params{Struct: p}, Hash_write_Results{Struct: r}}
			return s.Write(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 0},
	})

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xf29f97dd675a9431,
			MethodID:      1,
			InterfaceName: "hash.capnp:Hash",
			MethodName:    "sum",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Hash_sum{c, opts, Hash_sum_Params{Struct: p}, Hash_sum_Results{Struct: r}}
			return s.Sum(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// Hash_write holds the arguments for a server call to Hash.write.
type Hash_write struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Hash_write_Params
	Results Hash_write_Results
}

// Hash_sum holds the arguments for a server call to Hash.sum.
type Hash_sum struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Hash_sum_Params
	Results Hash_sum_Results
}

type Hash_write_Params struct{ capnp.Struct }

// Hash_write_Params_TypeID is the unique identifier for the type Hash_write_Params.
const Hash_write_Params_TypeID = 0xdffe94ae546cdee3

func NewHash_write_Params(s *capnp.Segment) (Hash_write_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Hash_write_Params{st}, err
}

func NewRootHash_write_Params(s *capnp.Segment) (Hash_write_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Hash_write_Params{st}, err
}

func ReadRootHash_write_Params(msg *capnp.Message) (Hash_write_Params, error) {
	root, err := msg.RootPtr()
	return Hash_write_Params{root.Struct()}, err
}

func (s Hash_write_Params) String() string {
	str, _ := text.Marshal(0xdffe94ae546cdee3, s.Struct)
	return str
}

func (s Hash_write_Params) Data() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Hash_write_Params) HasData() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Hash_write_Params) SetData(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, d.List.ToPtr())
}

// Hash_write_Params_List is a list of Hash_write_Params.
type Hash_write_Params_List struct{ capnp.List }

// NewHash_write_Params creates a new list of Hash_write_Params.
func NewHash_write_Params_List(s *capnp.Segment, sz int32) (Hash_write_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Hash_write_Params_List{l}, err
}

func (s Hash_write_Params_List) At(i int) Hash_write_Params {
	return Hash_write_Params{s.List.Struct(i)}
}

func (s Hash_write_Params_List) Set(i int, v Hash_write_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Hash_write_Params_Promise is a wrapper for a Hash_write_Params promised by a client call.
type Hash_write_Params_Promise struct{ *capnp.Pipeline }

func (p Hash_write_Params_Promise) Struct() (Hash_write_Params, error) {
	s, err := p.Pipeline.Struct()
	return Hash_write_Params{s}, err
}

type Hash_write_Results struct{ capnp.Struct }

// Hash_write_Results_TypeID is the unique identifier for the type Hash_write_Results.
const Hash_write_Results_TypeID = 0x80ac741ec7fb8f65

func NewHash_write_Results(s *capnp.Segment) (Hash_write_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hash_write_Results{st}, err
}

func NewRootHash_write_Results(s *capnp.Segment) (Hash_write_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hash_write_Results{st}, err
}

func ReadRootHash_write_Results(msg *capnp.Message) (Hash_write_Results, error) {
	root, err := msg.RootPtr()
	return Hash_write_Results{root.Struct()}, err
}

func (s Hash_write_Results) String() string {
	str, _ := text.Marshal(0x80ac741ec7fb8f65, s.Struct)
	return str
}

// Hash_write_Results_List is a list of Hash_write_Results.
type Hash_write_Results_List struct{ capnp.List }

// NewHash_write_Results creates a new list of Hash_write_Results.
func NewHash_write_Results_List(s *capnp.Segment, sz int32) (Hash_write_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Hash_write_Results_List{l}, err
}

func (s Hash_write_Results_List) At(i int) Hash_write_Results {
	return Hash_write_Results{s.List.Struct(i)}
}

func (s Hash_write_Results_List) Set(i int, v Hash_write_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Hash_write_Results_Promise is a wrapper for a Hash_write_Results promised by a client call.
type Hash_write_Results_Promise struct{ *capnp.Pipeline }

func (p Hash_write_Results_Promise) Struct() (Hash_write_Results, error) {
	s, err := p.Pipeline.Struct()
	return Hash_write_Results{s}, err
}

type Hash_sum_Params struct{ capnp.Struct }

// Hash_sum_Params_TypeID is the unique identifier for the type Hash_sum_Params.
const Hash_sum_Params_TypeID = 0xe74bb2d0190cf89c

func NewHash_sum_Params(s *capnp.Segment) (Hash_sum_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hash_sum_Params{st}, err
}

func NewRootHash_sum_Params(s *capnp.Segment) (Hash_sum_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Hash_sum_Params{st}, err
}

func ReadRootHash_sum_Params(msg *capnp.Message) (Hash_sum_Params, error) {
	root, err := msg.RootPtr()
	return Hash_sum_Params{root.Struct()}, err
}

func (s Hash_sum_Params) String() string {
	str, _ := text.Marshal(0xe74bb2d0190cf89c, s.Struct)
	return str
}

// Hash_sum_Params_List is a list of Hash_sum_Params.
type Hash_sum_Params_List struct{ capnp.List }

// NewHash_sum_Params creates a new list of Hash_sum_Params.
func NewHash_sum_Params_List(s *capnp.Segment, sz int32) (Hash_sum_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Hash_sum_Params_List{l}, err
}

func (s Hash_sum_Params_List) At(i int) Hash_sum_Params { return Hash_sum_Params{s.List.Struct(i)} }

func (s Hash_sum_Params_List) Set(i int, v Hash_sum_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Hash_sum_Params_Promise is a wrapper for a Hash_sum_Params promised by a client call.
type Hash_sum_Params_Promise struct{ *capnp.Pipeline }

func (p Hash_sum_Params_Promise) Struct() (Hash_sum_Params, error) {
	s, err := p.Pipeline.Struct()
	return Hash_sum_Params{s}, err
}

type Hash_sum_Results struct{ capnp.Struct }

// Hash_sum_Results_TypeID is the unique identifier for the type Hash_sum_Results.
const Hash_sum_Results_TypeID = 0xd093963b95a4e107

func NewHash_sum_Results(s *capnp.Segment) (Hash_sum_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Hash_sum_Results{st}, err
}

func NewRootHash_sum_Results(s *capnp.Segment) (Hash_sum_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Hash_sum_Results{st}, err
}

func ReadRootHash_sum_Results(msg *capnp.Message) (Hash_sum_Results, error) {
	root, err := msg.RootPtr()
	return Hash_sum_Results{root.Struct()}, err
}

func (s Hash_sum_Results) String() string {
	str, _ := text.Marshal(0xd093963b95a4e107, s.Struct)
	return str
}

func (s Hash_sum_Results) Hash() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Hash_sum_Results) HasHash() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Hash_sum_Results) SetHash(v []byte) error {
	d, err := capnp.NewData(s.Struct.Segment(), []byte(v))
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, d.List.ToPtr())
}

// Hash_sum_Results_List is a list of Hash_sum_Results.
type Hash_sum_Results_List struct{ capnp.List }

// NewHash_sum_Results creates a new list of Hash_sum_Results.
func NewHash_sum_Results_List(s *capnp.Segment, sz int32) (Hash_sum_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Hash_sum_Results_List{l}, err
}

func (s Hash_sum_Results_List) At(i int) Hash_sum_Results { return Hash_sum_Results{s.List.Struct(i)} }

func (s Hash_sum_Results_List) Set(i int, v Hash_sum_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Hash_sum_Results_Promise is a wrapper for a Hash_sum_Results promised by a client call.
type Hash_sum_Results_Promise struct{ *capnp.Pipeline }

func (p Hash_sum_Results_Promise) Struct() (Hash_sum_Results, error) {
	s, err := p.Pipeline.Struct()
	return Hash_sum_Results{s}, err
}

const schema_db8274f9144abc7e = "x\xda\x12\xf8\xed\xc0d\xc8\xaa\xce\xc2\xc0\x10h\xc0\xca" +
	"\xf6?\xb5\xff\xf7q\xb9\x925\x0d\x0c\x82\x02\x8c\x0c\x0c" +
	",\xec\x0c\x0c\xc2?\x19?1\xb0\xfc?\xb5\xa0g\xe9" +
	"E\xaeM\x93\x18\x04\xc5\xa1\x12\xc6w\x19\x8d\x18\x812" +
	"{n\xfd\x9d\xce\x1f\xb1v\x1d\x83 7\xf3\xff\xba=" +
	"^\"?K\x9an300\x0a\x1fe\xdc%|\x96" +
	"\x11d\xc4IFw\xe1\x8f \xd6\x7f\xf6\x87K\xa6Z" +
	"O\x9b|\x01b>+X\xf6&\xe3#\xa0\xea\xbb\x8c" +
	"\xf6\x0c\x8c\xff\x1f\xdf\xcb\x09Y7\xe5\xdf}d\xf9\xbf" +
	"\x8c\xaf\x80\xf2\x8cL \xf99?x$/l\xf2~" +
	"\x8e\xe4>E\xa6[@W\xdc\xaf\xb6O\xfb\x1e`\xf7" +
	"\x0a\xe2>\xb0FcV&+F\xa0N^\xb0N\xc3" +
	")Q\xe9w\xa7\xcf\xff\x84\xe1L]\xa6&aC&" +
	"\x90I\xbaL\xed\xc2\xb5@\x96\xce\xff\x8c\xc4\xe2\x0c\xbd" +
	"\xe4\xc4\x02\xa6\xbc\x02+\x0f\x10\xbb\xbc(\xb3$U%" +
	"(U\xbe\xb84\xa7\xa4\x18.\xcf\x0c\x95wKL." +
	"\xc9/\xaa\xd4\xcbK-\x0f\xceH4T\x09\x90O," +
	"J\xccE\xa8c\x84\xa9\xb3\x87(\x0c`d\x0cda" +
	"f\x05\x06\x08,\\\x19a\x1e\x10\x14tb`\x12d" +
	"e\xaf\x87\x9a\xe5\xc0\x08T\x8c\xe9\xa0\xe2\xd2\\\xa0s" +
	"\x8aK\xd9\x81\xce\x01\x1a\x05\x8c?\x16\xa0\xbf\x05y\xb5" +
	"\x80\x11\xc9\xc1\xcc\x18(\xc2\xc4\xc8\x0f\xd2\xc4\xc8\xcb\xc0" +
	"\x04\xc4\x8c\xb8\xbc\x14\x90\xc8\x0fr).#R\x12K" +
	"\x12q\x1b\x01rD\x00\xd0\xa7\xcc\xb9\x84\x83$\xc8>" +
	"\x15\x1cvx\x1d+\x88\x88&\x06FFA$;a" +
	"!\xc8\x00\x0a:\x0ep\xd0\xc1\xd2\x0a#,\xd1\x0a\x1a" +
	"\x1a\x01\x83N\x95\x9d\x11\x91N\x18a\x09NPR\x09" +
	"(\xc7\xcb.\x0f\xf6\xb6\x03#;\xd0\xed\xe0\xa0\x05\x04" +
	"\x00\x00\xff\xff<.\xe3\xa6"

func init() {
	schemas.Register(schema_db8274f9144abc7e,
		0x80ac741ec7fb8f65,
		0x92b20ad1a58ca0ca,
		0xaead580f97fddabc,
		0xd093963b95a4e107,
		0xdffe94ae546cdee3,
		0xe74bb2d0190cf89c,
		0xea3e50f7663f7bdf,
		0xf29f97dd675a9431)
}
