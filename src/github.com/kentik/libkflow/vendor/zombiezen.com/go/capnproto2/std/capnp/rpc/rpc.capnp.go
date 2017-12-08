package rpc

// AUTO GENERATED - DO NOT EDIT

import (
	strconv "strconv"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Message struct{ capnp.Struct }
type Message_Which uint16

const (
	Message_Which_unimplemented  Message_Which = 0
	Message_Which_abort          Message_Which = 1
	Message_Which_bootstrap      Message_Which = 8
	Message_Which_call           Message_Which = 2
	Message_Which_return         Message_Which = 3
	Message_Which_finish         Message_Which = 4
	Message_Which_resolve        Message_Which = 5
	Message_Which_release        Message_Which = 6
	Message_Which_disembargo     Message_Which = 13
	Message_Which_obsoleteSave   Message_Which = 7
	Message_Which_obsoleteDelete Message_Which = 9
	Message_Which_provide        Message_Which = 10
	Message_Which_accept         Message_Which = 11
	Message_Which_join           Message_Which = 12
)

func (w Message_Which) String() string {
	const s = "unimplementedabortbootstrapcallreturnfinishresolvereleasedisembargoobsoleteSaveobsoleteDeleteprovideacceptjoin"
	switch w {
	case Message_Which_unimplemented:
		return s[0:13]
	case Message_Which_abort:
		return s[13:18]
	case Message_Which_bootstrap:
		return s[18:27]
	case Message_Which_call:
		return s[27:31]
	case Message_Which_return:
		return s[31:37]
	case Message_Which_finish:
		return s[37:43]
	case Message_Which_resolve:
		return s[43:50]
	case Message_Which_release:
		return s[50:57]
	case Message_Which_disembargo:
		return s[57:67]
	case Message_Which_obsoleteSave:
		return s[67:79]
	case Message_Which_obsoleteDelete:
		return s[79:93]
	case Message_Which_provide:
		return s[93:100]
	case Message_Which_accept:
		return s[100:106]
	case Message_Which_join:
		return s[106:110]

	}
	return "Message_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Message_TypeID is the unique identifier for the type Message.
const Message_TypeID = 0x91b79f1f808db032

func NewMessage(s *capnp.Segment) (Message, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Message{st}, err
}

func NewRootMessage(s *capnp.Segment) (Message, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Message{st}, err
}

func ReadRootMessage(msg *capnp.Message) (Message, error) {
	root, err := msg.RootPtr()
	return Message{root.Struct()}, err
}

func (s Message) String() string {
	str, _ := text.Marshal(0x91b79f1f808db032, s.Struct)
	return str
}

func (s Message) Which() Message_Which {
	return Message_Which(s.Struct.Uint16(0))
}
func (s Message) Unimplemented() (Message, error) {
	p, err := s.Struct.Ptr(0)
	return Message{Struct: p.Struct()}, err
}

func (s Message) HasUnimplemented() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetUnimplemented(v Message) error {
	s.Struct.SetUint16(0, 0)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewUnimplemented sets the unimplemented field to a newly
// allocated Message struct, preferring placement in s's segment.
func (s Message) NewUnimplemented() (Message, error) {
	s.Struct.SetUint16(0, 0)
	ss, err := NewMessage(s.Struct.Segment())
	if err != nil {
		return Message{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Abort() (Exception, error) {
	p, err := s.Struct.Ptr(0)
	return Exception{Struct: p.Struct()}, err
}

func (s Message) HasAbort() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetAbort(v Exception) error {
	s.Struct.SetUint16(0, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewAbort sets the abort field to a newly
// allocated Exception struct, preferring placement in s's segment.
func (s Message) NewAbort() (Exception, error) {
	s.Struct.SetUint16(0, 1)
	ss, err := NewException(s.Struct.Segment())
	if err != nil {
		return Exception{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Bootstrap() (Bootstrap, error) {
	p, err := s.Struct.Ptr(0)
	return Bootstrap{Struct: p.Struct()}, err
}

func (s Message) HasBootstrap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetBootstrap(v Bootstrap) error {
	s.Struct.SetUint16(0, 8)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBootstrap sets the bootstrap field to a newly
// allocated Bootstrap struct, preferring placement in s's segment.
func (s Message) NewBootstrap() (Bootstrap, error) {
	s.Struct.SetUint16(0, 8)
	ss, err := NewBootstrap(s.Struct.Segment())
	if err != nil {
		return Bootstrap{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Call() (Call, error) {
	p, err := s.Struct.Ptr(0)
	return Call{Struct: p.Struct()}, err
}

func (s Message) HasCall() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetCall(v Call) error {
	s.Struct.SetUint16(0, 2)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewCall sets the call field to a newly
// allocated Call struct, preferring placement in s's segment.
func (s Message) NewCall() (Call, error) {
	s.Struct.SetUint16(0, 2)
	ss, err := NewCall(s.Struct.Segment())
	if err != nil {
		return Call{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Return() (Return, error) {
	p, err := s.Struct.Ptr(0)
	return Return{Struct: p.Struct()}, err
}

func (s Message) HasReturn() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetReturn(v Return) error {
	s.Struct.SetUint16(0, 3)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewReturn sets the return field to a newly
// allocated Return struct, preferring placement in s's segment.
func (s Message) NewReturn() (Return, error) {
	s.Struct.SetUint16(0, 3)
	ss, err := NewReturn(s.Struct.Segment())
	if err != nil {
		return Return{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Finish() (Finish, error) {
	p, err := s.Struct.Ptr(0)
	return Finish{Struct: p.Struct()}, err
}

func (s Message) HasFinish() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetFinish(v Finish) error {
	s.Struct.SetUint16(0, 4)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewFinish sets the finish field to a newly
// allocated Finish struct, preferring placement in s's segment.
func (s Message) NewFinish() (Finish, error) {
	s.Struct.SetUint16(0, 4)
	ss, err := NewFinish(s.Struct.Segment())
	if err != nil {
		return Finish{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Resolve() (Resolve, error) {
	p, err := s.Struct.Ptr(0)
	return Resolve{Struct: p.Struct()}, err
}

func (s Message) HasResolve() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetResolve(v Resolve) error {
	s.Struct.SetUint16(0, 5)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewResolve sets the resolve field to a newly
// allocated Resolve struct, preferring placement in s's segment.
func (s Message) NewResolve() (Resolve, error) {
	s.Struct.SetUint16(0, 5)
	ss, err := NewResolve(s.Struct.Segment())
	if err != nil {
		return Resolve{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Release() (Release, error) {
	p, err := s.Struct.Ptr(0)
	return Release{Struct: p.Struct()}, err
}

func (s Message) HasRelease() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetRelease(v Release) error {
	s.Struct.SetUint16(0, 6)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewRelease sets the release field to a newly
// allocated Release struct, preferring placement in s's segment.
func (s Message) NewRelease() (Release, error) {
	s.Struct.SetUint16(0, 6)
	ss, err := NewRelease(s.Struct.Segment())
	if err != nil {
		return Release{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Disembargo() (Disembargo, error) {
	p, err := s.Struct.Ptr(0)
	return Disembargo{Struct: p.Struct()}, err
}

func (s Message) HasDisembargo() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetDisembargo(v Disembargo) error {
	s.Struct.SetUint16(0, 13)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewDisembargo sets the disembargo field to a newly
// allocated Disembargo struct, preferring placement in s's segment.
func (s Message) NewDisembargo() (Disembargo, error) {
	s.Struct.SetUint16(0, 13)
	ss, err := NewDisembargo(s.Struct.Segment())
	if err != nil {
		return Disembargo{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) ObsoleteSave() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Message) HasObsoleteSave() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) ObsoleteSavePtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Message) SetObsoleteSave(v capnp.Pointer) error {
	s.Struct.SetUint16(0, 7)
	return s.Struct.SetPointer(0, v)
}

func (s Message) SetObsoleteSavePtr(v capnp.Ptr) error {
	s.Struct.SetUint16(0, 7)
	return s.Struct.SetPtr(0, v)
}

func (s Message) ObsoleteDelete() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Message) HasObsoleteDelete() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) ObsoleteDeletePtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Message) SetObsoleteDelete(v capnp.Pointer) error {
	s.Struct.SetUint16(0, 9)
	return s.Struct.SetPointer(0, v)
}

func (s Message) SetObsoleteDeletePtr(v capnp.Ptr) error {
	s.Struct.SetUint16(0, 9)
	return s.Struct.SetPtr(0, v)
}

func (s Message) Provide() (Provide, error) {
	p, err := s.Struct.Ptr(0)
	return Provide{Struct: p.Struct()}, err
}

func (s Message) HasProvide() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetProvide(v Provide) error {
	s.Struct.SetUint16(0, 10)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewProvide sets the provide field to a newly
// allocated Provide struct, preferring placement in s's segment.
func (s Message) NewProvide() (Provide, error) {
	s.Struct.SetUint16(0, 10)
	ss, err := NewProvide(s.Struct.Segment())
	if err != nil {
		return Provide{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Accept() (Accept, error) {
	p, err := s.Struct.Ptr(0)
	return Accept{Struct: p.Struct()}, err
}

func (s Message) HasAccept() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetAccept(v Accept) error {
	s.Struct.SetUint16(0, 11)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewAccept sets the accept field to a newly
// allocated Accept struct, preferring placement in s's segment.
func (s Message) NewAccept() (Accept, error) {
	s.Struct.SetUint16(0, 11)
	ss, err := NewAccept(s.Struct.Segment())
	if err != nil {
		return Accept{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Join() (Join, error) {
	p, err := s.Struct.Ptr(0)
	return Join{Struct: p.Struct()}, err
}

func (s Message) HasJoin() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetJoin(v Join) error {
	s.Struct.SetUint16(0, 12)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewJoin sets the join field to a newly
// allocated Join struct, preferring placement in s's segment.
func (s Message) NewJoin() (Join, error) {
	s.Struct.SetUint16(0, 12)
	ss, err := NewJoin(s.Struct.Segment())
	if err != nil {
		return Join{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Message_List is a list of Message.
type Message_List struct{ capnp.List }

// NewMessage creates a new list of Message.
func NewMessage_List(s *capnp.Segment, sz int32) (Message_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Message_List{l}, err
}

func (s Message_List) At(i int) Message { return Message{s.List.Struct(i)} }

func (s Message_List) Set(i int, v Message) error { return s.List.SetStruct(i, v.Struct) }

// Message_Promise is a wrapper for a Message promised by a client call.
type Message_Promise struct{ *capnp.Pipeline }

func (p Message_Promise) Struct() (Message, error) {
	s, err := p.Pipeline.Struct()
	return Message{s}, err
}

func (p Message_Promise) Unimplemented() Message_Promise {
	return Message_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Abort() Exception_Promise {
	return Exception_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Bootstrap() Bootstrap_Promise {
	return Bootstrap_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Call() Call_Promise {
	return Call_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Return() Return_Promise {
	return Return_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Finish() Finish_Promise {
	return Finish_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Resolve() Resolve_Promise {
	return Resolve_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Release() Release_Promise {
	return Release_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Disembargo() Disembargo_Promise {
	return Disembargo_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) ObsoleteSave() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

func (p Message_Promise) ObsoleteDelete() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

func (p Message_Promise) Provide() Provide_Promise {
	return Provide_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Accept() Accept_Promise {
	return Accept_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Join() Join_Promise {
	return Join_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Bootstrap struct{ capnp.Struct }

// Bootstrap_TypeID is the unique identifier for the type Bootstrap.
const Bootstrap_TypeID = 0xe94ccf8031176ec4

func NewBootstrap(s *capnp.Segment) (Bootstrap, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Bootstrap{st}, err
}

func NewRootBootstrap(s *capnp.Segment) (Bootstrap, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Bootstrap{st}, err
}

func ReadRootBootstrap(msg *capnp.Message) (Bootstrap, error) {
	root, err := msg.RootPtr()
	return Bootstrap{root.Struct()}, err
}

func (s Bootstrap) String() string {
	str, _ := text.Marshal(0xe94ccf8031176ec4, s.Struct)
	return str
}

func (s Bootstrap) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Bootstrap) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Bootstrap) DeprecatedObjectId() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Bootstrap) HasDeprecatedObjectId() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Bootstrap) DeprecatedObjectIdPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Bootstrap) SetDeprecatedObjectId(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s Bootstrap) SetDeprecatedObjectIdPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

// Bootstrap_List is a list of Bootstrap.
type Bootstrap_List struct{ capnp.List }

// NewBootstrap creates a new list of Bootstrap.
func NewBootstrap_List(s *capnp.Segment, sz int32) (Bootstrap_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Bootstrap_List{l}, err
}

func (s Bootstrap_List) At(i int) Bootstrap { return Bootstrap{s.List.Struct(i)} }

func (s Bootstrap_List) Set(i int, v Bootstrap) error { return s.List.SetStruct(i, v.Struct) }

// Bootstrap_Promise is a wrapper for a Bootstrap promised by a client call.
type Bootstrap_Promise struct{ *capnp.Pipeline }

func (p Bootstrap_Promise) Struct() (Bootstrap, error) {
	s, err := p.Pipeline.Struct()
	return Bootstrap{s}, err
}

func (p Bootstrap_Promise) DeprecatedObjectId() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type Call struct{ capnp.Struct }
type Call_sendResultsTo Call
type Call_sendResultsTo_Which uint16

const (
	Call_sendResultsTo_Which_caller     Call_sendResultsTo_Which = 0
	Call_sendResultsTo_Which_yourself   Call_sendResultsTo_Which = 1
	Call_sendResultsTo_Which_thirdParty Call_sendResultsTo_Which = 2
)

func (w Call_sendResultsTo_Which) String() string {
	const s = "calleryourselfthirdParty"
	switch w {
	case Call_sendResultsTo_Which_caller:
		return s[0:6]
	case Call_sendResultsTo_Which_yourself:
		return s[6:14]
	case Call_sendResultsTo_Which_thirdParty:
		return s[14:24]

	}
	return "Call_sendResultsTo_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Call_TypeID is the unique identifier for the type Call.
const Call_TypeID = 0x836a53ce789d4cd4

func NewCall(s *capnp.Segment) (Call, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3})
	return Call{st}, err
}

func NewRootCall(s *capnp.Segment) (Call, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3})
	return Call{st}, err
}

func ReadRootCall(msg *capnp.Message) (Call, error) {
	root, err := msg.RootPtr()
	return Call{root.Struct()}, err
}

func (s Call) String() string {
	str, _ := text.Marshal(0x836a53ce789d4cd4, s.Struct)
	return str
}

func (s Call) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Call) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Call) Target() (MessageTarget, error) {
	p, err := s.Struct.Ptr(0)
	return MessageTarget{Struct: p.Struct()}, err
}

func (s Call) HasTarget() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Call) SetTarget(v MessageTarget) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewTarget sets the target field to a newly
// allocated MessageTarget struct, preferring placement in s's segment.
func (s Call) NewTarget() (MessageTarget, error) {
	ss, err := NewMessageTarget(s.Struct.Segment())
	if err != nil {
		return MessageTarget{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Call) InterfaceId() uint64 {
	return s.Struct.Uint64(8)
}

func (s Call) SetInterfaceId(v uint64) {
	s.Struct.SetUint64(8, v)
}

func (s Call) MethodId() uint16 {
	return s.Struct.Uint16(4)
}

func (s Call) SetMethodId(v uint16) {
	s.Struct.SetUint16(4, v)
}

func (s Call) AllowThirdPartyTailCall() bool {
	return s.Struct.Bit(128)
}

func (s Call) SetAllowThirdPartyTailCall(v bool) {
	s.Struct.SetBit(128, v)
}

func (s Call) Params() (Payload, error) {
	p, err := s.Struct.Ptr(1)
	return Payload{Struct: p.Struct()}, err
}

func (s Call) HasParams() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Call) SetParams(v Payload) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewParams sets the params field to a newly
// allocated Payload struct, preferring placement in s's segment.
func (s Call) NewParams() (Payload, error) {
	ss, err := NewPayload(s.Struct.Segment())
	if err != nil {
		return Payload{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s Call) SendResultsTo() Call_sendResultsTo { return Call_sendResultsTo(s) }

func (s Call_sendResultsTo) Which() Call_sendResultsTo_Which {
	return Call_sendResultsTo_Which(s.Struct.Uint16(6))
}
func (s Call_sendResultsTo) SetCaller() {
	s.Struct.SetUint16(6, 0)

}

func (s Call_sendResultsTo) SetYourself() {
	s.Struct.SetUint16(6, 1)

}

func (s Call_sendResultsTo) ThirdParty() (capnp.Pointer, error) {
	return s.Struct.Pointer(2)
}

func (s Call_sendResultsTo) HasThirdParty() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s Call_sendResultsTo) ThirdPartyPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(2)
}

func (s Call_sendResultsTo) SetThirdParty(v capnp.Pointer) error {
	s.Struct.SetUint16(6, 2)
	return s.Struct.SetPointer(2, v)
}

func (s Call_sendResultsTo) SetThirdPartyPtr(v capnp.Ptr) error {
	s.Struct.SetUint16(6, 2)
	return s.Struct.SetPtr(2, v)
}

// Call_List is a list of Call.
type Call_List struct{ capnp.List }

// NewCall creates a new list of Call.
func NewCall_List(s *capnp.Segment, sz int32) (Call_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 24, PointerCount: 3}, sz)
	return Call_List{l}, err
}

func (s Call_List) At(i int) Call { return Call{s.List.Struct(i)} }

func (s Call_List) Set(i int, v Call) error { return s.List.SetStruct(i, v.Struct) }

// Call_Promise is a wrapper for a Call promised by a client call.
type Call_Promise struct{ *capnp.Pipeline }

func (p Call_Promise) Struct() (Call, error) {
	s, err := p.Pipeline.Struct()
	return Call{s}, err
}

func (p Call_Promise) Target() MessageTarget_Promise {
	return MessageTarget_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Call_Promise) Params() Payload_Promise {
	return Payload_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

func (p Call_Promise) SendResultsTo() Call_sendResultsTo_Promise {
	return Call_sendResultsTo_Promise{p.Pipeline}
}

// Call_sendResultsTo_Promise is a wrapper for a Call_sendResultsTo promised by a client call.
type Call_sendResultsTo_Promise struct{ *capnp.Pipeline }

func (p Call_sendResultsTo_Promise) Struct() (Call_sendResultsTo, error) {
	s, err := p.Pipeline.Struct()
	return Call_sendResultsTo{s}, err
}

func (p Call_sendResultsTo_Promise) ThirdParty() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(2)
}

type Return struct{ capnp.Struct }
type Return_Which uint16

const (
	Return_Which_results               Return_Which = 0
	Return_Which_exception             Return_Which = 1
	Return_Which_canceled              Return_Which = 2
	Return_Which_resultsSentElsewhere  Return_Which = 3
	Return_Which_takeFromOtherQuestion Return_Which = 4
	Return_Which_acceptFromThirdParty  Return_Which = 5
)

func (w Return_Which) String() string {
	const s = "resultsexceptioncanceledresultsSentElsewheretakeFromOtherQuestionacceptFromThirdParty"
	switch w {
	case Return_Which_results:
		return s[0:7]
	case Return_Which_exception:
		return s[7:16]
	case Return_Which_canceled:
		return s[16:24]
	case Return_Which_resultsSentElsewhere:
		return s[24:44]
	case Return_Which_takeFromOtherQuestion:
		return s[44:65]
	case Return_Which_acceptFromThirdParty:
		return s[65:85]

	}
	return "Return_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Return_TypeID is the unique identifier for the type Return.
const Return_TypeID = 0x9e19b28d3db3573a

func NewReturn(s *capnp.Segment) (Return, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1})
	return Return{st}, err
}

func NewRootReturn(s *capnp.Segment) (Return, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1})
	return Return{st}, err
}

func ReadRootReturn(msg *capnp.Message) (Return, error) {
	root, err := msg.RootPtr()
	return Return{root.Struct()}, err
}

func (s Return) String() string {
	str, _ := text.Marshal(0x9e19b28d3db3573a, s.Struct)
	return str
}

func (s Return) Which() Return_Which {
	return Return_Which(s.Struct.Uint16(6))
}
func (s Return) AnswerId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Return) SetAnswerId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Return) ReleaseParamCaps() bool {
	return !s.Struct.Bit(32)
}

func (s Return) SetReleaseParamCaps(v bool) {
	s.Struct.SetBit(32, !v)
}

func (s Return) Results() (Payload, error) {
	p, err := s.Struct.Ptr(0)
	return Payload{Struct: p.Struct()}, err
}

func (s Return) HasResults() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Return) SetResults(v Payload) error {
	s.Struct.SetUint16(6, 0)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewResults sets the results field to a newly
// allocated Payload struct, preferring placement in s's segment.
func (s Return) NewResults() (Payload, error) {
	s.Struct.SetUint16(6, 0)
	ss, err := NewPayload(s.Struct.Segment())
	if err != nil {
		return Payload{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Return) Exception() (Exception, error) {
	p, err := s.Struct.Ptr(0)
	return Exception{Struct: p.Struct()}, err
}

func (s Return) HasException() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Return) SetException(v Exception) error {
	s.Struct.SetUint16(6, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewException sets the exception field to a newly
// allocated Exception struct, preferring placement in s's segment.
func (s Return) NewException() (Exception, error) {
	s.Struct.SetUint16(6, 1)
	ss, err := NewException(s.Struct.Segment())
	if err != nil {
		return Exception{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Return) SetCanceled() {
	s.Struct.SetUint16(6, 2)

}

func (s Return) SetResultsSentElsewhere() {
	s.Struct.SetUint16(6, 3)

}

func (s Return) TakeFromOtherQuestion() uint32 {
	return s.Struct.Uint32(8)
}

func (s Return) SetTakeFromOtherQuestion(v uint32) {
	s.Struct.SetUint16(6, 4)
	s.Struct.SetUint32(8, v)
}

func (s Return) AcceptFromThirdParty() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Return) HasAcceptFromThirdParty() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Return) AcceptFromThirdPartyPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Return) SetAcceptFromThirdParty(v capnp.Pointer) error {
	s.Struct.SetUint16(6, 5)
	return s.Struct.SetPointer(0, v)
}

func (s Return) SetAcceptFromThirdPartyPtr(v capnp.Ptr) error {
	s.Struct.SetUint16(6, 5)
	return s.Struct.SetPtr(0, v)
}

// Return_List is a list of Return.
type Return_List struct{ capnp.List }

// NewReturn creates a new list of Return.
func NewReturn_List(s *capnp.Segment, sz int32) (Return_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 1}, sz)
	return Return_List{l}, err
}

func (s Return_List) At(i int) Return { return Return{s.List.Struct(i)} }

func (s Return_List) Set(i int, v Return) error { return s.List.SetStruct(i, v.Struct) }

// Return_Promise is a wrapper for a Return promised by a client call.
type Return_Promise struct{ *capnp.Pipeline }

func (p Return_Promise) Struct() (Return, error) {
	s, err := p.Pipeline.Struct()
	return Return{s}, err
}

func (p Return_Promise) Results() Payload_Promise {
	return Payload_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Return_Promise) Exception() Exception_Promise {
	return Exception_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Return_Promise) AcceptFromThirdParty() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type Finish struct{ capnp.Struct }

// Finish_TypeID is the unique identifier for the type Finish.
const Finish_TypeID = 0xd37d2eb2c2f80e63

func NewFinish(s *capnp.Segment) (Finish, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Finish{st}, err
}

func NewRootFinish(s *capnp.Segment) (Finish, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Finish{st}, err
}

func ReadRootFinish(msg *capnp.Message) (Finish, error) {
	root, err := msg.RootPtr()
	return Finish{root.Struct()}, err
}

func (s Finish) String() string {
	str, _ := text.Marshal(0xd37d2eb2c2f80e63, s.Struct)
	return str
}

func (s Finish) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Finish) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Finish) ReleaseResultCaps() bool {
	return !s.Struct.Bit(32)
}

func (s Finish) SetReleaseResultCaps(v bool) {
	s.Struct.SetBit(32, !v)
}

// Finish_List is a list of Finish.
type Finish_List struct{ capnp.List }

// NewFinish creates a new list of Finish.
func NewFinish_List(s *capnp.Segment, sz int32) (Finish_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return Finish_List{l}, err
}

func (s Finish_List) At(i int) Finish { return Finish{s.List.Struct(i)} }

func (s Finish_List) Set(i int, v Finish) error { return s.List.SetStruct(i, v.Struct) }

// Finish_Promise is a wrapper for a Finish promised by a client call.
type Finish_Promise struct{ *capnp.Pipeline }

func (p Finish_Promise) Struct() (Finish, error) {
	s, err := p.Pipeline.Struct()
	return Finish{s}, err
}

type Resolve struct{ capnp.Struct }
type Resolve_Which uint16

const (
	Resolve_Which_cap       Resolve_Which = 0
	Resolve_Which_exception Resolve_Which = 1
)

func (w Resolve_Which) String() string {
	const s = "capexception"
	switch w {
	case Resolve_Which_cap:
		return s[0:3]
	case Resolve_Which_exception:
		return s[3:12]

	}
	return "Resolve_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Resolve_TypeID is the unique identifier for the type Resolve.
const Resolve_TypeID = 0xbbc29655fa89086e

func NewResolve(s *capnp.Segment) (Resolve, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Resolve{st}, err
}

func NewRootResolve(s *capnp.Segment) (Resolve, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Resolve{st}, err
}

func ReadRootResolve(msg *capnp.Message) (Resolve, error) {
	root, err := msg.RootPtr()
	return Resolve{root.Struct()}, err
}

func (s Resolve) String() string {
	str, _ := text.Marshal(0xbbc29655fa89086e, s.Struct)
	return str
}

func (s Resolve) Which() Resolve_Which {
	return Resolve_Which(s.Struct.Uint16(4))
}
func (s Resolve) PromiseId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Resolve) SetPromiseId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Resolve) Cap() (CapDescriptor, error) {
	p, err := s.Struct.Ptr(0)
	return CapDescriptor{Struct: p.Struct()}, err
}

func (s Resolve) HasCap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Resolve) SetCap(v CapDescriptor) error {
	s.Struct.SetUint16(4, 0)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewCap sets the cap field to a newly
// allocated CapDescriptor struct, preferring placement in s's segment.
func (s Resolve) NewCap() (CapDescriptor, error) {
	s.Struct.SetUint16(4, 0)
	ss, err := NewCapDescriptor(s.Struct.Segment())
	if err != nil {
		return CapDescriptor{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Resolve) Exception() (Exception, error) {
	p, err := s.Struct.Ptr(0)
	return Exception{Struct: p.Struct()}, err
}

func (s Resolve) HasException() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Resolve) SetException(v Exception) error {
	s.Struct.SetUint16(4, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewException sets the exception field to a newly
// allocated Exception struct, preferring placement in s's segment.
func (s Resolve) NewException() (Exception, error) {
	s.Struct.SetUint16(4, 1)
	ss, err := NewException(s.Struct.Segment())
	if err != nil {
		return Exception{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Resolve_List is a list of Resolve.
type Resolve_List struct{ capnp.List }

// NewResolve creates a new list of Resolve.
func NewResolve_List(s *capnp.Segment, sz int32) (Resolve_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Resolve_List{l}, err
}

func (s Resolve_List) At(i int) Resolve { return Resolve{s.List.Struct(i)} }

func (s Resolve_List) Set(i int, v Resolve) error { return s.List.SetStruct(i, v.Struct) }

// Resolve_Promise is a wrapper for a Resolve promised by a client call.
type Resolve_Promise struct{ *capnp.Pipeline }

func (p Resolve_Promise) Struct() (Resolve, error) {
	s, err := p.Pipeline.Struct()
	return Resolve{s}, err
}

func (p Resolve_Promise) Cap() CapDescriptor_Promise {
	return CapDescriptor_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Resolve_Promise) Exception() Exception_Promise {
	return Exception_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Release struct{ capnp.Struct }

// Release_TypeID is the unique identifier for the type Release.
const Release_TypeID = 0xad1a6c0d7dd07497

func NewRelease(s *capnp.Segment) (Release, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Release{st}, err
}

func NewRootRelease(s *capnp.Segment) (Release, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Release{st}, err
}

func ReadRootRelease(msg *capnp.Message) (Release, error) {
	root, err := msg.RootPtr()
	return Release{root.Struct()}, err
}

func (s Release) String() string {
	str, _ := text.Marshal(0xad1a6c0d7dd07497, s.Struct)
	return str
}

func (s Release) Id() uint32 {
	return s.Struct.Uint32(0)
}

func (s Release) SetId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Release) ReferenceCount() uint32 {
	return s.Struct.Uint32(4)
}

func (s Release) SetReferenceCount(v uint32) {
	s.Struct.SetUint32(4, v)
}

// Release_List is a list of Release.
type Release_List struct{ capnp.List }

// NewRelease creates a new list of Release.
func NewRelease_List(s *capnp.Segment, sz int32) (Release_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return Release_List{l}, err
}

func (s Release_List) At(i int) Release { return Release{s.List.Struct(i)} }

func (s Release_List) Set(i int, v Release) error { return s.List.SetStruct(i, v.Struct) }

// Release_Promise is a wrapper for a Release promised by a client call.
type Release_Promise struct{ *capnp.Pipeline }

func (p Release_Promise) Struct() (Release, error) {
	s, err := p.Pipeline.Struct()
	return Release{s}, err
}

type Disembargo struct{ capnp.Struct }
type Disembargo_context Disembargo
type Disembargo_context_Which uint16

const (
	Disembargo_context_Which_senderLoopback   Disembargo_context_Which = 0
	Disembargo_context_Which_receiverLoopback Disembargo_context_Which = 1
	Disembargo_context_Which_accept           Disembargo_context_Which = 2
	Disembargo_context_Which_provide          Disembargo_context_Which = 3
)

func (w Disembargo_context_Which) String() string {
	const s = "senderLoopbackreceiverLoopbackacceptprovide"
	switch w {
	case Disembargo_context_Which_senderLoopback:
		return s[0:14]
	case Disembargo_context_Which_receiverLoopback:
		return s[14:30]
	case Disembargo_context_Which_accept:
		return s[30:36]
	case Disembargo_context_Which_provide:
		return s[36:43]

	}
	return "Disembargo_context_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Disembargo_TypeID is the unique identifier for the type Disembargo.
const Disembargo_TypeID = 0xf964368b0fbd3711

func NewDisembargo(s *capnp.Segment) (Disembargo, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Disembargo{st}, err
}

func NewRootDisembargo(s *capnp.Segment) (Disembargo, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Disembargo{st}, err
}

func ReadRootDisembargo(msg *capnp.Message) (Disembargo, error) {
	root, err := msg.RootPtr()
	return Disembargo{root.Struct()}, err
}

func (s Disembargo) String() string {
	str, _ := text.Marshal(0xf964368b0fbd3711, s.Struct)
	return str
}

func (s Disembargo) Target() (MessageTarget, error) {
	p, err := s.Struct.Ptr(0)
	return MessageTarget{Struct: p.Struct()}, err
}

func (s Disembargo) HasTarget() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Disembargo) SetTarget(v MessageTarget) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewTarget sets the target field to a newly
// allocated MessageTarget struct, preferring placement in s's segment.
func (s Disembargo) NewTarget() (MessageTarget, error) {
	ss, err := NewMessageTarget(s.Struct.Segment())
	if err != nil {
		return MessageTarget{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Disembargo) Context() Disembargo_context { return Disembargo_context(s) }

func (s Disembargo_context) Which() Disembargo_context_Which {
	return Disembargo_context_Which(s.Struct.Uint16(4))
}
func (s Disembargo_context) SenderLoopback() uint32 {
	return s.Struct.Uint32(0)
}

func (s Disembargo_context) SetSenderLoopback(v uint32) {
	s.Struct.SetUint16(4, 0)
	s.Struct.SetUint32(0, v)
}

func (s Disembargo_context) ReceiverLoopback() uint32 {
	return s.Struct.Uint32(0)
}

func (s Disembargo_context) SetReceiverLoopback(v uint32) {
	s.Struct.SetUint16(4, 1)
	s.Struct.SetUint32(0, v)
}

func (s Disembargo_context) SetAccept() {
	s.Struct.SetUint16(4, 2)

}

func (s Disembargo_context) Provide() uint32 {
	return s.Struct.Uint32(0)
}

func (s Disembargo_context) SetProvide(v uint32) {
	s.Struct.SetUint16(4, 3)
	s.Struct.SetUint32(0, v)
}

// Disembargo_List is a list of Disembargo.
type Disembargo_List struct{ capnp.List }

// NewDisembargo creates a new list of Disembargo.
func NewDisembargo_List(s *capnp.Segment, sz int32) (Disembargo_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Disembargo_List{l}, err
}

func (s Disembargo_List) At(i int) Disembargo { return Disembargo{s.List.Struct(i)} }

func (s Disembargo_List) Set(i int, v Disembargo) error { return s.List.SetStruct(i, v.Struct) }

// Disembargo_Promise is a wrapper for a Disembargo promised by a client call.
type Disembargo_Promise struct{ *capnp.Pipeline }

func (p Disembargo_Promise) Struct() (Disembargo, error) {
	s, err := p.Pipeline.Struct()
	return Disembargo{s}, err
}

func (p Disembargo_Promise) Target() MessageTarget_Promise {
	return MessageTarget_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Disembargo_Promise) Context() Disembargo_context_Promise {
	return Disembargo_context_Promise{p.Pipeline}
}

// Disembargo_context_Promise is a wrapper for a Disembargo_context promised by a client call.
type Disembargo_context_Promise struct{ *capnp.Pipeline }

func (p Disembargo_context_Promise) Struct() (Disembargo_context, error) {
	s, err := p.Pipeline.Struct()
	return Disembargo_context{s}, err
}

type Provide struct{ capnp.Struct }

// Provide_TypeID is the unique identifier for the type Provide.
const Provide_TypeID = 0x9c6a046bfbc1ac5a

func NewProvide(s *capnp.Segment) (Provide, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Provide{st}, err
}

func NewRootProvide(s *capnp.Segment) (Provide, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Provide{st}, err
}

func ReadRootProvide(msg *capnp.Message) (Provide, error) {
	root, err := msg.RootPtr()
	return Provide{root.Struct()}, err
}

func (s Provide) String() string {
	str, _ := text.Marshal(0x9c6a046bfbc1ac5a, s.Struct)
	return str
}

func (s Provide) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Provide) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Provide) Target() (MessageTarget, error) {
	p, err := s.Struct.Ptr(0)
	return MessageTarget{Struct: p.Struct()}, err
}

func (s Provide) HasTarget() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Provide) SetTarget(v MessageTarget) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewTarget sets the target field to a newly
// allocated MessageTarget struct, preferring placement in s's segment.
func (s Provide) NewTarget() (MessageTarget, error) {
	ss, err := NewMessageTarget(s.Struct.Segment())
	if err != nil {
		return MessageTarget{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Provide) Recipient() (capnp.Pointer, error) {
	return s.Struct.Pointer(1)
}

func (s Provide) HasRecipient() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Provide) RecipientPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(1)
}

func (s Provide) SetRecipient(v capnp.Pointer) error {
	return s.Struct.SetPointer(1, v)
}

func (s Provide) SetRecipientPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(1, v)
}

// Provide_List is a list of Provide.
type Provide_List struct{ capnp.List }

// NewProvide creates a new list of Provide.
func NewProvide_List(s *capnp.Segment, sz int32) (Provide_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2}, sz)
	return Provide_List{l}, err
}

func (s Provide_List) At(i int) Provide { return Provide{s.List.Struct(i)} }

func (s Provide_List) Set(i int, v Provide) error { return s.List.SetStruct(i, v.Struct) }

// Provide_Promise is a wrapper for a Provide promised by a client call.
type Provide_Promise struct{ *capnp.Pipeline }

func (p Provide_Promise) Struct() (Provide, error) {
	s, err := p.Pipeline.Struct()
	return Provide{s}, err
}

func (p Provide_Promise) Target() MessageTarget_Promise {
	return MessageTarget_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Provide_Promise) Recipient() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(1)
}

type Accept struct{ capnp.Struct }

// Accept_TypeID is the unique identifier for the type Accept.
const Accept_TypeID = 0xd4c9b56290554016

func NewAccept(s *capnp.Segment) (Accept, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Accept{st}, err
}

func NewRootAccept(s *capnp.Segment) (Accept, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Accept{st}, err
}

func ReadRootAccept(msg *capnp.Message) (Accept, error) {
	root, err := msg.RootPtr()
	return Accept{root.Struct()}, err
}

func (s Accept) String() string {
	str, _ := text.Marshal(0xd4c9b56290554016, s.Struct)
	return str
}

func (s Accept) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Accept) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Accept) Provision() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Accept) HasProvision() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Accept) ProvisionPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Accept) SetProvision(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s Accept) SetProvisionPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

func (s Accept) Embargo() bool {
	return s.Struct.Bit(32)
}

func (s Accept) SetEmbargo(v bool) {
	s.Struct.SetBit(32, v)
}

// Accept_List is a list of Accept.
type Accept_List struct{ capnp.List }

// NewAccept creates a new list of Accept.
func NewAccept_List(s *capnp.Segment, sz int32) (Accept_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Accept_List{l}, err
}

func (s Accept_List) At(i int) Accept { return Accept{s.List.Struct(i)} }

func (s Accept_List) Set(i int, v Accept) error { return s.List.SetStruct(i, v.Struct) }

// Accept_Promise is a wrapper for a Accept promised by a client call.
type Accept_Promise struct{ *capnp.Pipeline }

func (p Accept_Promise) Struct() (Accept, error) {
	s, err := p.Pipeline.Struct()
	return Accept{s}, err
}

func (p Accept_Promise) Provision() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type Join struct{ capnp.Struct }

// Join_TypeID is the unique identifier for the type Join.
const Join_TypeID = 0xfbe1980490e001af

func NewJoin(s *capnp.Segment) (Join, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Join{st}, err
}

func NewRootJoin(s *capnp.Segment) (Join, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Join{st}, err
}

func ReadRootJoin(msg *capnp.Message) (Join, error) {
	root, err := msg.RootPtr()
	return Join{root.Struct()}, err
}

func (s Join) String() string {
	str, _ := text.Marshal(0xfbe1980490e001af, s.Struct)
	return str
}

func (s Join) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s Join) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Join) Target() (MessageTarget, error) {
	p, err := s.Struct.Ptr(0)
	return MessageTarget{Struct: p.Struct()}, err
}

func (s Join) HasTarget() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Join) SetTarget(v MessageTarget) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewTarget sets the target field to a newly
// allocated MessageTarget struct, preferring placement in s's segment.
func (s Join) NewTarget() (MessageTarget, error) {
	ss, err := NewMessageTarget(s.Struct.Segment())
	if err != nil {
		return MessageTarget{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Join) KeyPart() (capnp.Pointer, error) {
	return s.Struct.Pointer(1)
}

func (s Join) HasKeyPart() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Join) KeyPartPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(1)
}

func (s Join) SetKeyPart(v capnp.Pointer) error {
	return s.Struct.SetPointer(1, v)
}

func (s Join) SetKeyPartPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(1, v)
}

// Join_List is a list of Join.
type Join_List struct{ capnp.List }

// NewJoin creates a new list of Join.
func NewJoin_List(s *capnp.Segment, sz int32) (Join_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2}, sz)
	return Join_List{l}, err
}

func (s Join_List) At(i int) Join { return Join{s.List.Struct(i)} }

func (s Join_List) Set(i int, v Join) error { return s.List.SetStruct(i, v.Struct) }

// Join_Promise is a wrapper for a Join promised by a client call.
type Join_Promise struct{ *capnp.Pipeline }

func (p Join_Promise) Struct() (Join, error) {
	s, err := p.Pipeline.Struct()
	return Join{s}, err
}

func (p Join_Promise) Target() MessageTarget_Promise {
	return MessageTarget_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Join_Promise) KeyPart() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(1)
}

type MessageTarget struct{ capnp.Struct }
type MessageTarget_Which uint16

const (
	MessageTarget_Which_importedCap    MessageTarget_Which = 0
	MessageTarget_Which_promisedAnswer MessageTarget_Which = 1
)

func (w MessageTarget_Which) String() string {
	const s = "importedCappromisedAnswer"
	switch w {
	case MessageTarget_Which_importedCap:
		return s[0:11]
	case MessageTarget_Which_promisedAnswer:
		return s[11:25]

	}
	return "MessageTarget_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// MessageTarget_TypeID is the unique identifier for the type MessageTarget.
const MessageTarget_TypeID = 0x95bc14545813fbc1

func NewMessageTarget(s *capnp.Segment) (MessageTarget, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return MessageTarget{st}, err
}

func NewRootMessageTarget(s *capnp.Segment) (MessageTarget, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return MessageTarget{st}, err
}

func ReadRootMessageTarget(msg *capnp.Message) (MessageTarget, error) {
	root, err := msg.RootPtr()
	return MessageTarget{root.Struct()}, err
}

func (s MessageTarget) String() string {
	str, _ := text.Marshal(0x95bc14545813fbc1, s.Struct)
	return str
}

func (s MessageTarget) Which() MessageTarget_Which {
	return MessageTarget_Which(s.Struct.Uint16(4))
}
func (s MessageTarget) ImportedCap() uint32 {
	return s.Struct.Uint32(0)
}

func (s MessageTarget) SetImportedCap(v uint32) {
	s.Struct.SetUint16(4, 0)
	s.Struct.SetUint32(0, v)
}

func (s MessageTarget) PromisedAnswer() (PromisedAnswer, error) {
	p, err := s.Struct.Ptr(0)
	return PromisedAnswer{Struct: p.Struct()}, err
}

func (s MessageTarget) HasPromisedAnswer() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s MessageTarget) SetPromisedAnswer(v PromisedAnswer) error {
	s.Struct.SetUint16(4, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPromisedAnswer sets the promisedAnswer field to a newly
// allocated PromisedAnswer struct, preferring placement in s's segment.
func (s MessageTarget) NewPromisedAnswer() (PromisedAnswer, error) {
	s.Struct.SetUint16(4, 1)
	ss, err := NewPromisedAnswer(s.Struct.Segment())
	if err != nil {
		return PromisedAnswer{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// MessageTarget_List is a list of MessageTarget.
type MessageTarget_List struct{ capnp.List }

// NewMessageTarget creates a new list of MessageTarget.
func NewMessageTarget_List(s *capnp.Segment, sz int32) (MessageTarget_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return MessageTarget_List{l}, err
}

func (s MessageTarget_List) At(i int) MessageTarget { return MessageTarget{s.List.Struct(i)} }

func (s MessageTarget_List) Set(i int, v MessageTarget) error { return s.List.SetStruct(i, v.Struct) }

// MessageTarget_Promise is a wrapper for a MessageTarget promised by a client call.
type MessageTarget_Promise struct{ *capnp.Pipeline }

func (p MessageTarget_Promise) Struct() (MessageTarget, error) {
	s, err := p.Pipeline.Struct()
	return MessageTarget{s}, err
}

func (p MessageTarget_Promise) PromisedAnswer() PromisedAnswer_Promise {
	return PromisedAnswer_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Payload struct{ capnp.Struct }

// Payload_TypeID is the unique identifier for the type Payload.
const Payload_TypeID = 0x9a0e61223d96743b

func NewPayload(s *capnp.Segment) (Payload, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Payload{st}, err
}

func NewRootPayload(s *capnp.Segment) (Payload, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Payload{st}, err
}

func ReadRootPayload(msg *capnp.Message) (Payload, error) {
	root, err := msg.RootPtr()
	return Payload{root.Struct()}, err
}

func (s Payload) String() string {
	str, _ := text.Marshal(0x9a0e61223d96743b, s.Struct)
	return str
}

func (s Payload) Content() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s Payload) HasContent() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Payload) ContentPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s Payload) SetContent(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s Payload) SetContentPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

func (s Payload) CapTable() (CapDescriptor_List, error) {
	p, err := s.Struct.Ptr(1)
	return CapDescriptor_List{List: p.List()}, err
}

func (s Payload) HasCapTable() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Payload) SetCapTable(v CapDescriptor_List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewCapTable sets the capTable field to a newly
// allocated CapDescriptor_List, preferring placement in s's segment.
func (s Payload) NewCapTable(n int32) (CapDescriptor_List, error) {
	l, err := NewCapDescriptor_List(s.Struct.Segment(), n)
	if err != nil {
		return CapDescriptor_List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

// Payload_List is a list of Payload.
type Payload_List struct{ capnp.List }

// NewPayload creates a new list of Payload.
func NewPayload_List(s *capnp.Segment, sz int32) (Payload_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return Payload_List{l}, err
}

func (s Payload_List) At(i int) Payload { return Payload{s.List.Struct(i)} }

func (s Payload_List) Set(i int, v Payload) error { return s.List.SetStruct(i, v.Struct) }

// Payload_Promise is a wrapper for a Payload promised by a client call.
type Payload_Promise struct{ *capnp.Pipeline }

func (p Payload_Promise) Struct() (Payload, error) {
	s, err := p.Pipeline.Struct()
	return Payload{s}, err
}

func (p Payload_Promise) Content() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type CapDescriptor struct{ capnp.Struct }
type CapDescriptor_Which uint16

const (
	CapDescriptor_Which_none             CapDescriptor_Which = 0
	CapDescriptor_Which_senderHosted     CapDescriptor_Which = 1
	CapDescriptor_Which_senderPromise    CapDescriptor_Which = 2
	CapDescriptor_Which_receiverHosted   CapDescriptor_Which = 3
	CapDescriptor_Which_receiverAnswer   CapDescriptor_Which = 4
	CapDescriptor_Which_thirdPartyHosted CapDescriptor_Which = 5
)

func (w CapDescriptor_Which) String() string {
	const s = "nonesenderHostedsenderPromisereceiverHostedreceiverAnswerthirdPartyHosted"
	switch w {
	case CapDescriptor_Which_none:
		return s[0:4]
	case CapDescriptor_Which_senderHosted:
		return s[4:16]
	case CapDescriptor_Which_senderPromise:
		return s[16:29]
	case CapDescriptor_Which_receiverHosted:
		return s[29:43]
	case CapDescriptor_Which_receiverAnswer:
		return s[43:57]
	case CapDescriptor_Which_thirdPartyHosted:
		return s[57:73]

	}
	return "CapDescriptor_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// CapDescriptor_TypeID is the unique identifier for the type CapDescriptor.
const CapDescriptor_TypeID = 0x8523ddc40b86b8b0

func NewCapDescriptor(s *capnp.Segment) (CapDescriptor, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return CapDescriptor{st}, err
}

func NewRootCapDescriptor(s *capnp.Segment) (CapDescriptor, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return CapDescriptor{st}, err
}

func ReadRootCapDescriptor(msg *capnp.Message) (CapDescriptor, error) {
	root, err := msg.RootPtr()
	return CapDescriptor{root.Struct()}, err
}

func (s CapDescriptor) String() string {
	str, _ := text.Marshal(0x8523ddc40b86b8b0, s.Struct)
	return str
}

func (s CapDescriptor) Which() CapDescriptor_Which {
	return CapDescriptor_Which(s.Struct.Uint16(0))
}
func (s CapDescriptor) SetNone() {
	s.Struct.SetUint16(0, 0)

}

func (s CapDescriptor) SenderHosted() uint32 {
	return s.Struct.Uint32(4)
}

func (s CapDescriptor) SetSenderHosted(v uint32) {
	s.Struct.SetUint16(0, 1)
	s.Struct.SetUint32(4, v)
}

func (s CapDescriptor) SenderPromise() uint32 {
	return s.Struct.Uint32(4)
}

func (s CapDescriptor) SetSenderPromise(v uint32) {
	s.Struct.SetUint16(0, 2)
	s.Struct.SetUint32(4, v)
}

func (s CapDescriptor) ReceiverHosted() uint32 {
	return s.Struct.Uint32(4)
}

func (s CapDescriptor) SetReceiverHosted(v uint32) {
	s.Struct.SetUint16(0, 3)
	s.Struct.SetUint32(4, v)
}

func (s CapDescriptor) ReceiverAnswer() (PromisedAnswer, error) {
	p, err := s.Struct.Ptr(0)
	return PromisedAnswer{Struct: p.Struct()}, err
}

func (s CapDescriptor) HasReceiverAnswer() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s CapDescriptor) SetReceiverAnswer(v PromisedAnswer) error {
	s.Struct.SetUint16(0, 4)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewReceiverAnswer sets the receiverAnswer field to a newly
// allocated PromisedAnswer struct, preferring placement in s's segment.
func (s CapDescriptor) NewReceiverAnswer() (PromisedAnswer, error) {
	s.Struct.SetUint16(0, 4)
	ss, err := NewPromisedAnswer(s.Struct.Segment())
	if err != nil {
		return PromisedAnswer{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s CapDescriptor) ThirdPartyHosted() (ThirdPartyCapDescriptor, error) {
	p, err := s.Struct.Ptr(0)
	return ThirdPartyCapDescriptor{Struct: p.Struct()}, err
}

func (s CapDescriptor) HasThirdPartyHosted() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s CapDescriptor) SetThirdPartyHosted(v ThirdPartyCapDescriptor) error {
	s.Struct.SetUint16(0, 5)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewThirdPartyHosted sets the thirdPartyHosted field to a newly
// allocated ThirdPartyCapDescriptor struct, preferring placement in s's segment.
func (s CapDescriptor) NewThirdPartyHosted() (ThirdPartyCapDescriptor, error) {
	s.Struct.SetUint16(0, 5)
	ss, err := NewThirdPartyCapDescriptor(s.Struct.Segment())
	if err != nil {
		return ThirdPartyCapDescriptor{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// CapDescriptor_List is a list of CapDescriptor.
type CapDescriptor_List struct{ capnp.List }

// NewCapDescriptor creates a new list of CapDescriptor.
func NewCapDescriptor_List(s *capnp.Segment, sz int32) (CapDescriptor_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return CapDescriptor_List{l}, err
}

func (s CapDescriptor_List) At(i int) CapDescriptor { return CapDescriptor{s.List.Struct(i)} }

func (s CapDescriptor_List) Set(i int, v CapDescriptor) error { return s.List.SetStruct(i, v.Struct) }

// CapDescriptor_Promise is a wrapper for a CapDescriptor promised by a client call.
type CapDescriptor_Promise struct{ *capnp.Pipeline }

func (p CapDescriptor_Promise) Struct() (CapDescriptor, error) {
	s, err := p.Pipeline.Struct()
	return CapDescriptor{s}, err
}

func (p CapDescriptor_Promise) ReceiverAnswer() PromisedAnswer_Promise {
	return PromisedAnswer_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p CapDescriptor_Promise) ThirdPartyHosted() ThirdPartyCapDescriptor_Promise {
	return ThirdPartyCapDescriptor_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type PromisedAnswer struct{ capnp.Struct }

// PromisedAnswer_TypeID is the unique identifier for the type PromisedAnswer.
const PromisedAnswer_TypeID = 0xd800b1d6cd6f1ca0

func NewPromisedAnswer(s *capnp.Segment) (PromisedAnswer, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return PromisedAnswer{st}, err
}

func NewRootPromisedAnswer(s *capnp.Segment) (PromisedAnswer, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return PromisedAnswer{st}, err
}

func ReadRootPromisedAnswer(msg *capnp.Message) (PromisedAnswer, error) {
	root, err := msg.RootPtr()
	return PromisedAnswer{root.Struct()}, err
}

func (s PromisedAnswer) String() string {
	str, _ := text.Marshal(0xd800b1d6cd6f1ca0, s.Struct)
	return str
}

func (s PromisedAnswer) QuestionId() uint32 {
	return s.Struct.Uint32(0)
}

func (s PromisedAnswer) SetQuestionId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s PromisedAnswer) Transform() (PromisedAnswer_Op_List, error) {
	p, err := s.Struct.Ptr(0)
	return PromisedAnswer_Op_List{List: p.List()}, err
}

func (s PromisedAnswer) HasTransform() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s PromisedAnswer) SetTransform(v PromisedAnswer_Op_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewTransform sets the transform field to a newly
// allocated PromisedAnswer_Op_List, preferring placement in s's segment.
func (s PromisedAnswer) NewTransform(n int32) (PromisedAnswer_Op_List, error) {
	l, err := NewPromisedAnswer_Op_List(s.Struct.Segment(), n)
	if err != nil {
		return PromisedAnswer_Op_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// PromisedAnswer_List is a list of PromisedAnswer.
type PromisedAnswer_List struct{ capnp.List }

// NewPromisedAnswer creates a new list of PromisedAnswer.
func NewPromisedAnswer_List(s *capnp.Segment, sz int32) (PromisedAnswer_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return PromisedAnswer_List{l}, err
}

func (s PromisedAnswer_List) At(i int) PromisedAnswer { return PromisedAnswer{s.List.Struct(i)} }

func (s PromisedAnswer_List) Set(i int, v PromisedAnswer) error { return s.List.SetStruct(i, v.Struct) }

// PromisedAnswer_Promise is a wrapper for a PromisedAnswer promised by a client call.
type PromisedAnswer_Promise struct{ *capnp.Pipeline }

func (p PromisedAnswer_Promise) Struct() (PromisedAnswer, error) {
	s, err := p.Pipeline.Struct()
	return PromisedAnswer{s}, err
}

type PromisedAnswer_Op struct{ capnp.Struct }
type PromisedAnswer_Op_Which uint16

const (
	PromisedAnswer_Op_Which_noop            PromisedAnswer_Op_Which = 0
	PromisedAnswer_Op_Which_getPointerField PromisedAnswer_Op_Which = 1
)

func (w PromisedAnswer_Op_Which) String() string {
	const s = "noopgetPointerField"
	switch w {
	case PromisedAnswer_Op_Which_noop:
		return s[0:4]
	case PromisedAnswer_Op_Which_getPointerField:
		return s[4:19]

	}
	return "PromisedAnswer_Op_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// PromisedAnswer_Op_TypeID is the unique identifier for the type PromisedAnswer_Op.
const PromisedAnswer_Op_TypeID = 0xf316944415569081

func NewPromisedAnswer_Op(s *capnp.Segment) (PromisedAnswer_Op, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return PromisedAnswer_Op{st}, err
}

func NewRootPromisedAnswer_Op(s *capnp.Segment) (PromisedAnswer_Op, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return PromisedAnswer_Op{st}, err
}

func ReadRootPromisedAnswer_Op(msg *capnp.Message) (PromisedAnswer_Op, error) {
	root, err := msg.RootPtr()
	return PromisedAnswer_Op{root.Struct()}, err
}

func (s PromisedAnswer_Op) String() string {
	str, _ := text.Marshal(0xf316944415569081, s.Struct)
	return str
}

func (s PromisedAnswer_Op) Which() PromisedAnswer_Op_Which {
	return PromisedAnswer_Op_Which(s.Struct.Uint16(0))
}
func (s PromisedAnswer_Op) SetNoop() {
	s.Struct.SetUint16(0, 0)

}

func (s PromisedAnswer_Op) GetPointerField() uint16 {
	return s.Struct.Uint16(2)
}

func (s PromisedAnswer_Op) SetGetPointerField(v uint16) {
	s.Struct.SetUint16(0, 1)
	s.Struct.SetUint16(2, v)
}

// PromisedAnswer_Op_List is a list of PromisedAnswer_Op.
type PromisedAnswer_Op_List struct{ capnp.List }

// NewPromisedAnswer_Op creates a new list of PromisedAnswer_Op.
func NewPromisedAnswer_Op_List(s *capnp.Segment, sz int32) (PromisedAnswer_Op_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return PromisedAnswer_Op_List{l}, err
}

func (s PromisedAnswer_Op_List) At(i int) PromisedAnswer_Op {
	return PromisedAnswer_Op{s.List.Struct(i)}
}

func (s PromisedAnswer_Op_List) Set(i int, v PromisedAnswer_Op) error {
	return s.List.SetStruct(i, v.Struct)
}

// PromisedAnswer_Op_Promise is a wrapper for a PromisedAnswer_Op promised by a client call.
type PromisedAnswer_Op_Promise struct{ *capnp.Pipeline }

func (p PromisedAnswer_Op_Promise) Struct() (PromisedAnswer_Op, error) {
	s, err := p.Pipeline.Struct()
	return PromisedAnswer_Op{s}, err
}

type ThirdPartyCapDescriptor struct{ capnp.Struct }

// ThirdPartyCapDescriptor_TypeID is the unique identifier for the type ThirdPartyCapDescriptor.
const ThirdPartyCapDescriptor_TypeID = 0xd37007fde1f0027d

func NewThirdPartyCapDescriptor(s *capnp.Segment) (ThirdPartyCapDescriptor, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return ThirdPartyCapDescriptor{st}, err
}

func NewRootThirdPartyCapDescriptor(s *capnp.Segment) (ThirdPartyCapDescriptor, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return ThirdPartyCapDescriptor{st}, err
}

func ReadRootThirdPartyCapDescriptor(msg *capnp.Message) (ThirdPartyCapDescriptor, error) {
	root, err := msg.RootPtr()
	return ThirdPartyCapDescriptor{root.Struct()}, err
}

func (s ThirdPartyCapDescriptor) String() string {
	str, _ := text.Marshal(0xd37007fde1f0027d, s.Struct)
	return str
}

func (s ThirdPartyCapDescriptor) Id() (capnp.Pointer, error) {
	return s.Struct.Pointer(0)
}

func (s ThirdPartyCapDescriptor) HasId() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s ThirdPartyCapDescriptor) IdPtr() (capnp.Ptr, error) {
	return s.Struct.Ptr(0)
}

func (s ThirdPartyCapDescriptor) SetId(v capnp.Pointer) error {
	return s.Struct.SetPointer(0, v)
}

func (s ThirdPartyCapDescriptor) SetIdPtr(v capnp.Ptr) error {
	return s.Struct.SetPtr(0, v)
}

func (s ThirdPartyCapDescriptor) VineId() uint32 {
	return s.Struct.Uint32(0)
}

func (s ThirdPartyCapDescriptor) SetVineId(v uint32) {
	s.Struct.SetUint32(0, v)
}

// ThirdPartyCapDescriptor_List is a list of ThirdPartyCapDescriptor.
type ThirdPartyCapDescriptor_List struct{ capnp.List }

// NewThirdPartyCapDescriptor creates a new list of ThirdPartyCapDescriptor.
func NewThirdPartyCapDescriptor_List(s *capnp.Segment, sz int32) (ThirdPartyCapDescriptor_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return ThirdPartyCapDescriptor_List{l}, err
}

func (s ThirdPartyCapDescriptor_List) At(i int) ThirdPartyCapDescriptor {
	return ThirdPartyCapDescriptor{s.List.Struct(i)}
}

func (s ThirdPartyCapDescriptor_List) Set(i int, v ThirdPartyCapDescriptor) error {
	return s.List.SetStruct(i, v.Struct)
}

// ThirdPartyCapDescriptor_Promise is a wrapper for a ThirdPartyCapDescriptor promised by a client call.
type ThirdPartyCapDescriptor_Promise struct{ *capnp.Pipeline }

func (p ThirdPartyCapDescriptor_Promise) Struct() (ThirdPartyCapDescriptor, error) {
	s, err := p.Pipeline.Struct()
	return ThirdPartyCapDescriptor{s}, err
}

func (p ThirdPartyCapDescriptor_Promise) Id() *capnp.Pipeline {
	return p.Pipeline.GetPipeline(0)
}

type Exception struct{ capnp.Struct }

// Exception_TypeID is the unique identifier for the type Exception.
const Exception_TypeID = 0xd625b7063acf691a

func NewException(s *capnp.Segment) (Exception, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Exception{st}, err
}

func NewRootException(s *capnp.Segment) (Exception, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Exception{st}, err
}

func ReadRootException(msg *capnp.Message) (Exception, error) {
	root, err := msg.RootPtr()
	return Exception{root.Struct()}, err
}

func (s Exception) String() string {
	str, _ := text.Marshal(0xd625b7063acf691a, s.Struct)
	return str
}

func (s Exception) Reason() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Exception) HasReason() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Exception) ReasonBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Exception) SetReason(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Exception) Type() Exception_Type {
	return Exception_Type(s.Struct.Uint16(4))
}

func (s Exception) SetType(v Exception_Type) {
	s.Struct.SetUint16(4, uint16(v))
}

func (s Exception) ObsoleteIsCallersFault() bool {
	return s.Struct.Bit(0)
}

func (s Exception) SetObsoleteIsCallersFault(v bool) {
	s.Struct.SetBit(0, v)
}

func (s Exception) ObsoleteDurability() uint16 {
	return s.Struct.Uint16(2)
}

func (s Exception) SetObsoleteDurability(v uint16) {
	s.Struct.SetUint16(2, v)
}

// Exception_List is a list of Exception.
type Exception_List struct{ capnp.List }

// NewException creates a new list of Exception.
func NewException_List(s *capnp.Segment, sz int32) (Exception_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Exception_List{l}, err
}

func (s Exception_List) At(i int) Exception { return Exception{s.List.Struct(i)} }

func (s Exception_List) Set(i int, v Exception) error { return s.List.SetStruct(i, v.Struct) }

// Exception_Promise is a wrapper for a Exception promised by a client call.
type Exception_Promise struct{ *capnp.Pipeline }

func (p Exception_Promise) Struct() (Exception, error) {
	s, err := p.Pipeline.Struct()
	return Exception{s}, err
}

type Exception_Type uint16

// Values of Exception_Type.
const (
	Exception_Type_failed        Exception_Type = 0
	Exception_Type_overloaded    Exception_Type = 1
	Exception_Type_disconnected  Exception_Type = 2
	Exception_Type_unimplemented Exception_Type = 3
)

// String returns the enum's constant name.
func (c Exception_Type) String() string {
	switch c {
	case Exception_Type_failed:
		return "failed"
	case Exception_Type_overloaded:
		return "overloaded"
	case Exception_Type_disconnected:
		return "disconnected"
	case Exception_Type_unimplemented:
		return "unimplemented"

	default:
		return ""
	}
}

// Exception_TypeFromString returns the enum value with a name,
// or the zero value if there's no such value.
func Exception_TypeFromString(c string) Exception_Type {
	switch c {
	case "failed":
		return Exception_Type_failed
	case "overloaded":
		return Exception_Type_overloaded
	case "disconnected":
		return Exception_Type_disconnected
	case "unimplemented":
		return Exception_Type_unimplemented

	default:
		return 0
	}
}

type Exception_Type_List struct{ capnp.List }

func NewException_Type_List(s *capnp.Segment, sz int32) (Exception_Type_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return Exception_Type_List{l.List}, err
}

func (l Exception_Type_List) At(i int) Exception_Type {
	ul := capnp.UInt16List{List: l.List}
	return Exception_Type(ul.At(i))
}

func (l Exception_Type_List) Set(i int, v Exception_Type) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

const schema_b312981b2552a250 = "x\xda\x9cX\x7fl\x14\xc7\x15\x9e\xd9=\xdf\x19\xff\xba" +
	"[\xaf\x09\x85\x06\x99\xa6E*\xa8X8M\xdb\xd4-" +
	":~\xd8\x08#[\xf8l\xd3\x10\x1a\xa9]\xdf\x8d\xf1" +
	"\xc2\xf9\xf6\xd8]\x83\x8db\x01mR%\x14T@@" +
	"\x01A\x0b(\x95\xd2\x94\x08B@\xa1mP1\"j" +
	"\x88B\x03\x0aDI\x15\xd4\x10\xa9*T\x8d\x944?" +
	"\x1a\xc0p\xfd\xde\xee\xde\xeeq>\x0b%\x7f \x8e\xf7" +
	"\xbd\x9dy\xef\xcd{\xdf7\xc3\xecG\xc3s\xa5\xc6\xb2" +
	"%\x13\x18K\xa4\xcb\xc2\xb9Km\xfb\x07\xff\xd6\xb5\xf2" +
	"\xe7,Q\xc1\xe5\\\xc7\xa1\xce\xe9_\xdd]\xfb\"+" +
	"\x93#\x8c\xa9[B\xeb\xd4m!\xfc\xfa\xf6\x96\xd0\xaf" +
	"8\xe3\xb9\xa3'\x7fQy\xf6\xca\xd7\x9f$o\x1ex" +
	"\xb7\xf0H\x18\xeeJ\xf8\x8c:9L\xee\x13\xc3\x8f\x90" +
	"\xfb\x83G\xb7l\xa8\xff\xedK\xdb\xc6\xba\xd7\xc0}(" +
	"\xb2]\xdd\x18!\xf7\xe1\xc8$\x19\xee#\xb7\xd4e\xdd" +
	"u/\xef\x1c\xeb.qI=UqF}\xa5\x82\xc2" +
	"\x1a\xa9X\x0b\xef\x1f\xd8\xbb\xe6<\xa0\xd5\xeceJE" +
	"\x81s\x99D\x1e\xd3+\xb7\xab\xb3*\xe9\xd7\x8cJ\xf2" +
	"]~x\xe4\xd6\xaa\xd0\xca}E+\xbb\xce{\xe0|" +
	"\xd0q\xde_y\x04\xceM\x8f\xbc8g\xcb\xb1\xc9\xbf" +
	"!g\xe9\xee$\xb9\xac\xce\xa9\xda\xa4\xb6TQ\xd4\xf3" +
	"\xaa\xfeJI\xfe\xda\xbe0\\\x9d\x9e\xf2|\xd1\xdaT" +
	"6\xb5\xb1f\xbb\xfa\xfd\x1a\xfa\xf5\x9d\x1a\x8ac\xd9\xa9" +
	"\xb6\xf8\xfb\xbb6\x1fcJ\x9d\x94\x9b\xa2\xbf\xd1\x14~" +
	"i\xfa[\x8cqug\xcdk\xeaA\xc7q\x7f\xcd\x0a" +
	"8f\xca\x9f\xbe\xb9t\xd7\x99?\x97.\xc5y,{" +
	"\xd9\xf1\xbeXC\x11\x0fK\x1f^\xbd\x1d\xc9\xbeY\x9c" +
	"\x1e\xa70WGk\xb9\xba1J\xde\xc3Q\x0a\"Y" +
	"\xf3\xf9\x99c\x0d\xc3o\x96\x0a\xf8\x9d\xe8&\xf5\xaa\xe3" +
	"{\xc5\xf1\xbdo\xee\xd2\xad='\xce]*\xb5\xb2\xda" +
	"\x12\xdb\xa4\xb6\xc7\xe8Wk\x8c\xc2h\xbf\xf2c\xf1\x8f" +
	"\xe3=\x97Yb\"\x9c\x95\xef\x9d\x8a\xfe\xf2\xbb\xa9\x1b" +
	"l)\x8f\xf0\x10\xa2\xbe\x1e\xfb72\xfd \xf6/\xb8" +
	"\xfa\xb9\x97Z\xf7\xbcrH\xbd\xacL\xa2 \x148\xff" +
	"\xe5\xc0\xfd\xc6\xf9\xb7^x\xbb\x94\xebH\xedk\xea\xf9" +
	"Zr\xbd\\K\xf1\xee\xf9\xc9\x1f\xa6|v\xf4\xda\xdf" +
	"Y\"\x8av\xf6\x9b{\xa9\x1c\xe12\x0e\xafU\xa5\x10" +
	"\xdaU\x8a\xf6lfR\xe3\x867\xda\xae\x97\x0c\xe1\x03" +
	"\xf5\x90\xfa\xa9J\xbf>Ri\xdd\x8d[\x7f4\xb1y" +
	"\xc7}\x1f\xb3\xc4d\xee\x07\xd4\x1c\x91\xe0\xf0h\xdd\xfb" +
	"\xaa\xa8#W\xad\x8e\\\xfd\xbcK\xc6[\xf7\x9cz\xce" +
	"q~\xc5q>\xc2\xdf\xdb\x1a\xda}\xf5V\xc9\xc6," +
	"\x9b\xb8N\x9d0\xd1\xfdu\x84-\xc8\x99\xd9dCR" +
	"\xcbfX<\xdb\xb4@K\xa7;8O\xdc/\x87\x18" +
	"\x0bq\xc6\x94\x13\xcb1\xd2\xc7e\x9e8-q\xce\xeb" +
	"8\xd9N5\xc1v\x12\xb6\xb3\x12W$\x18\x11\xb02" +
	"\xd2\x03\xe3i\x18_\x87Q\x96\xea\xb8\x0c\xe3\xb9\xc50" +
	"\xbe\x0a\xe3%\x18\xcb\xe0\x89e\x95\x8b\xf4\xf9\xeb0\xbe" +
	"\x8d%\xc3\xbc\xa0\xbc\xcae\x93IJhC\x1d/\xe7" +
	"\\\x199\x03\xbf\xb3\xf0\xbb \xf1\xdc\xea\x01a\xd9\xba" +
	"\x91ark\x8a\x973\x09\x7fx\xdc\xd6\xcc\x15\xc2\xe6" +
	"\xb1`\xc6\x19\xe71\x14@\xcf\xd8\xc2\xec\xd5\x92,\"" +
	"\xe0>\x01\xee\x13`\xed\x17v\x9f\x91jM1\xb8E" +
	"`\x8b`\x89\xacfj\xfd\x16\x96\xf0\x07\xdf[\xc2\x12" +
	"\x99T\xa7\xb0\x06X}\xda\xb6\xba\x8d\x1c*c\xac\xed" +
	"\xee\xd3%3\xd5\xa1\x99\xf6P\xb7\xa6\xa7\xa9\\p\xc7" +
	"R4\xb2\xf9BJT\xc7l\xb3\xb0\x92\xa6\x9e\xb5\x0d" +
	"\x93QE\xbf\"\x87\xaar9\xa7\xa4{f\"\xaf\x1d" +
	"\xc8\xeb\x80\xc4\xa7\xf2;9\xaf\xaa\xfbW\xc2\xbc\x0f\xe6" +
	"ga\x96n\xe7\xbc\xba\xfe\xce\x84\xf9\x19\x98\x8f\xc2," +
	"\x8f\x92\x99*\xfb\xfc:\x98\x0f\xc3|R\xe2\xd5\xa1[" +
	"9\xb7\xb4'\xd6\x05\xa7U]v\x13\xd62:\xafM" +
	"\xc1\xd1D3FF\xb0\xb0\x93\x9e0\x17\x19,j\xd9" +
	"\xc2\xaf\xa8g\xee0Y\xbd\xd1\xaf[\xc2\xb7\x9b\")" +
	"\xf45\xc2d\xf1E\xc6]\x1f\x04\xc0\xbc\x8c\xb5V\x98" +
	"<\x96\xefc\xaf\x8ev\x9f\xeeT\x8c\xdbC\xee\xa7(" +
	"p,\xe0\x16\xcf+_;\x9emj\x17\x96\xa5\xad\xe0" +
	"\x82\xaa\xf6\xb0_5u\x88\xa3\x10]\x83\x18\xb8\xae'" +
	"8\xb2C\xe1\x9c\xba\xa9\x1b\xf9\x83\x00\x1e'\xe0)\x02" +
	"\xe4\xdb9\xa7r\xea\x93\x1c\x85\xee\xda@\xc0f\x02B" +
	"\xa39\xa7v\xea\xd3\x1c\x1d\x88U\x00l%\xa0\xcc+" +
	"\x9f\xba\xc5\x01\x9e\"`\x07\x01a\xaf\x82\xea6>\x1f" +
	"\xc0f\x02v\x13\x10\xb9\x01\x80\xc4i\xa7\x03l%`" +
	"\x1f\x01\x13>\x07\xe0\xd0?\xc7q\xc2\x19\xc03\x04H" +
	"\xff\x03P\x0e\xe0 \xef\x04p\x80\x80\xc3\x04T|\x06" +
	"\x00\xba\xa9\xfe\x9e\xe3\xf0\xba\x9e%\xe08\x01\x95\x9f\x02" +
	"\xa8\x00\xf0\x82\xb3\xc7a\x02N\x12P\xf5\x09\x80J\x00" +
	"'\x9cp\x8f\x12\xf02\x01\xd5\x1f\x03\xa8\x02\xf0G'" +
	"\xf3\xe3\x04\x9c&\xa0\xfc\xbf\x00\xaa\x01\x9c\xe2\x18g8" +
	"\x03x\x15@n \xa3\xf7g\xd3\xa2\x9f\xd5\x8b\x0c\x9d" +
	"j,\x10W\xf7`\xea\xb5\x1e\xc3\xa4\x09+\x90\x15\xb2" +
	"G\x93h}\x98}.t\xcdqS\xd8\x03f\x06\x80" +
	"/w\x1e\xd0\xabgt\xab\x0f\x80\xaf\x13.\xb0\xde\x14" +
	"\x96\x91^#\x80\xf8\xea\xe4#i\xa1Y\x84\xf8b\xe8" +
	"u\x8b\xd1\x83o\x84-X\xb4K\xc3\xa7\xb5\xe8\xc5Z" +
	"\x98{\x0c\xc3\xb6lSc<\x8b\x8f|&.\xfe(" +
	"\xde,\xe8\xef\xfcg\xeb\xb3\xa6\xb1FO\xd1>\xbe\xa0" +
	"{Ak\xc9\xa4\xc8R\xf6\xbe`y\xd9\xaf4tJ" +
	"\xd2\xe7Yo\x8b\x14F\xa6\xbfG3\x99\xbc\xc2\x00\xec" +
	"svQ\x93K\xf9&\x17\xdd\x0e\x819\x04Q\x1e\x10" +
	"\xc4\x0c\xa2\xd2ob^\x1f*\xe8s\xa5\x91f{6" +
	"\xac?\xc4\xc1\xe1\xd8p.\x18\xa6\x08\xa8\xc6\x1fFd" +
	"BS\x9b\x1aw\x18\x0b\xc6\xacC\x1bJ\x1b\x1aOy" +
	"{{t?\x03\xad\x96\xf8\x066\x99\x0d\xc2\xce\xf3\xfd" +
	",b\xf1o\xc1\xb8H\xe2\xeb\x93\x06:%c\xfbE" +
	"\xc7r\xddZOZ\x10\xa9\xd60\xde!c\xa7\xe0F" +
	"\x87}k\x8a\xf6u\xaa\xed\x8ew\x95\xbfo\x0b\xc9L" +
	"3\xb6\xe8\x08d\xa6\x9dtb\x11l\xdd\x052\x93\xc0" +
	"\xf4$\xb0I\xe2\xb1/,\x0a\xa0*=\xab\x8b\x0c\xe3" +
	"A\xf4\x05\x81u:\xad\xcb\x9c\xc3\x98\xe6\x07v\x91r" +
	"\xbf\x80\xfd\xde\xa5\x82LCdP\xa6w\x88P\xdf\x85" +
	"\xf1\x1aMv\xce\xe5\x1b\xe5\x9fT\xbb\xf7`\xfd\x0f\xb1" +
	"\xd0\x1d\x97l\x94\xeb\x14\xf05X?!\x0a\xba\xed\x11" +
	"\xf5G\xb4\xec\x87\xb0\x8e\x12\xff\x8czD}\xe39X" +
	"G1\x9c\xe5\x18\xce\xa9\xe1[9\xc9e\x992~\x0c" +
	"c[Nc[\xe7\xd0\xcfM\x8fe\x14\x8e/`\x03" +
	"0\x8d\xe6Ys\x8e\xddU\xb8\x80\xa1\x9d1\xea\xe0\xa4" +
	"t\xe8\x16\x8b9\x92UF\x9aE\xd37\x00}+\xa1" +
	"\x7fb\x90z_7\x18\xcf\x8c\x1d\x7f\x9cz&\x89u" +
	"\xb1Q$\xe7\xad\xd1\xc5\xd1\x16-iK\xac\x8d\xf6\x09" +
	"\x934\xc6\xd6V\x89\x85hI\xbe\xc4\x86%1 \xea" +
	"\x9d\xd3\xf2#s\xc7k\xa1\xc9\x8d\xfenG%\xa2$" +
	"\xac\xa5\xcf\x86rp\x9b\xa6\xa0Y\xa7\x94j\xd6u^" +
	"\xb3>,qY/\x14\xaa^D\x95I\xb2\xb8X`" +
	"\x0c\xa0\x81} \x98\xca\x16/\xe7LC\xf7PV\xb8" +
	"\xad\x10s\xcevF\x13e\xae|\x0dm\xca%e*" +
	"\xb8\x9d\xcb\xcad(R\xbc\x17\xf7\x00\x91\xca\x19\xd0@" +
	"\xccS\x8a\xc9\xf8\x07x\x00c\x02\x9d\x8d&1\xa3\xc5" +
	",{wb\xc4~c\xa6\xa13\x98\x86j\x9e\xf3\x08" +
	"\xa0\xfd\x81`\x1e\xaa\xa5;\xb9\x12\x03\xe1\x11@+\xe3" +
	"~\xe2\x11\xecT4\x91\xf7>\xde|\x84r\xb6\xa9\xdb" +
	"\xd3o{\xa8\xf0R\xc3\xcd\xf1\x8f\xc2?\x89\xa6\x80\xc6" +
	"\xe8$\xbcs\x8d\xaf\xd13\xa255\xa6\xfe\xa8\xc6B" +
	"G$\x18+Z{y\xb0\x8e?\x82\x8d\xdba|\x08" +
	"\xc6\xb9\xe3\xf0@\xbe\xef;\xb9\xd3\x9eD\x93\x96\xdf\xf7" +
	"\x85\x9b\xces\xba\xd0\xdd\xf4\x1e\x84D\xa5n\x83m\x19" +
	"\x11\xd24\xb7\xfeK\xe7\xdf\x83\x90r\x8e\xbeXn\xa9" +
	"\xf3\x9a\xe3\xc8\x04D\xa2\xc4\xdd\xb1\xd9\x13\x91\x15F\x03" +
	"\x9a(j\x8bA;\x11s\xc4\xc1\x8dB\xa3\x06\xff)" +
	"6L\xe7\xd5\x81\xc2\xd0\x89\x92\xd2\xb0\x0eRs\xdc\xf6" +
	"\xc8g\x80\x8e \x0b\xeb\xe3DI\xa3\x1e\xf9\x0cQ\xc8" +
	"6\xac\x1b\xa4\xfc\x8d\xaf\xcd`q#\xdb\xa3%W\x8d" +
	"\xb9\xd9\xf16\xc3E\x02N\xf1\x84\x91\x85}\xed,q" +
	"\x98\xee0EP\x8dD\x88\x17>R\xf9\xcc(\x8d\x17" +
	"%\xe5\x15[\xa30\x1fC@}\xa8\x86\xe4\xa6)\xfe" +
	"\x04[\x1fl6\xbd'<\xf6_\xbd7\x88\\\xe1\xde" +
	"#c\x98\xee\xd3\x830>!\xd1\x05D\xb3@1U" +
	"\x88\xa7\xaa@\xf4y\xabE\x97ua\xc6\xad\x85\x1a\xda" +
	"\xc1/\xbc\xef\xd0<`j=zZ\x97\xc1@\xde\xe3" +
	" j#L\x1e\x0dB\xc7tD\xef>\xac\x0eOq" +
	"]\xbdE\x1c\x94\xaa\xff\xacS\xf8\x14yIv\x9cV" +
	"\xcewUc\xa7\xa7\xebm\xe35\x10\xae4\x19\xab\x17" +
	"\x0f\x09\xde\x1fH\xac\xbfI\x91\xc4J\xee+\xae!\xff" +
	"~IG\xe9\xf9B\x9d\xedt\x10\xc9L\x0b\x95{\xae" +
	"\xbb\xa3\xdbA\x10\x19\xa5uq@/\xf4\xfe\x90\x1c\x89" +
	"Q\x12\xcb\x83\xfe\x8e'\x9d\x1a\x82\xda\x87\x8c\x01\xd3\x12" +
	"\xe9^\xe2\xff\xfc\x0d\x9f\xc9\xa5\xc9{\xbes-\x8b\x98" +
	"Zv\xfc\xb9\xf6\x8b\xb1\xf7^c\x9d\x12Yt\xa6f" +
	"s\x91Z\xd2\xb3R$m\x02\x8bw\x1ds2\x91\x86" +
	"%\xd9\xe2[\xd6\xcc\x80\xb2\x0a\x9ea\xb3~\x16\xe8\x07" +
	"\x1eKF\x16\xd9\xe2>\xd1\x81\x1b\x1f\xf64\x17\xea\"" +
	"\x9d\xf2\x9f\x8f\x85i\xbas\x1b\xa5\xc1-\xca\xb3\xa9\x90" +
	"\x1by\xc1\xffh(\xb3\xe63i\xdc\x0b\x8b{\xd5\x1a" +
	"\xb4\xefz\xa1/F\x1c_\xf6\xea4?\xa0\xaf/v" +
	"uZ\xbfJ\x0c\x91\x04\xe4\xeb\xfc\xff\x00\x00\x00\xff\xff" +
	"<\xff\x0e\x96"

func init() {
	schemas.Register(schema_b312981b2552a250,
		0x836a53ce789d4cd4,
		0x8523ddc40b86b8b0,
		0x91b79f1f808db032,
		0x95bc14545813fbc1,
		0x9a0e61223d96743b,
		0x9c6a046bfbc1ac5a,
		0x9e19b28d3db3573a,
		0xad1a6c0d7dd07497,
		0xb28c96e23f4cbd58,
		0xbbc29655fa89086e,
		0xd37007fde1f0027d,
		0xd37d2eb2c2f80e63,
		0xd4c9b56290554016,
		0xd562b4df655bdd4d,
		0xd625b7063acf691a,
		0xd800b1d6cd6f1ca0,
		0xdae8b0f61aab5f99,
		0xe94ccf8031176ec4,
		0xf316944415569081,
		0xf964368b0fbd3711,
		0xfbe1980490e001af)
}
