package decoder

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
)

// DirectiveOp are operations that are performed on values that have been decoded.
// They are expected to be chained together, in a flow programming style, and the
// Decode Directive that they are assigned to then walks back up the linked list to find the root
// operation that will then be performed (passing the value down through various transformations)
type DirectiveOp interface {
	prev() DirectiveOp
	// process method can be executed in two contexts, one to check that the given type
	// of upstream value can be processed (not to process it) and then to actually process
	// the upstream value. The difference in reqwuired behaviour is signalled by the presence
	// of the DecodeContect - if nil. just test, if !nil process
	process(dc *DecodeContext, upstreamValue interface{}) error
}

type baseDOp struct {
	p  DirectiveOp
	do DirectiveOp
	n  DirectiveOp
}

func (op *baseDOp) prev() DirectiveOp {
	return op.p
}

func (op *baseDOp) AsF(name string) DirectiveOp {
	result := &AsFDOp{baseDOp: baseDOp{p: op.do}, name: name}
	result.do = result
	op.n = result
	return result
}

func (op *baseDOp) AsT(name string) DirectiveOp {
	result := &AsTDOp{baseDOp: baseDOp{p: op.do}, name: name}
	result.do = result
	op.n = result
	return result
}

func (op *baseDOp) Set(ptr interface{}) *SetDOp {
	result := &SetDOp{baseDOp: baseDOp{p: op.do}, ptr: ptr}
	result.do = result
	op.n = result
	return result
}

// U32ToU32DOp is a deode operation that can process U32 to U32
type U32ToU32DOp struct {
	baseDOp
	fn func(uint32) uint32
}

func (op *U32ToU32DOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	var out uint32
	switch v := upstreamValue.(type) {
	case *uint32:
		if dc != nil {
			out = op.fn(*v)
		}
	default:
		return fmt.Errorf("cannot process %T", v)
	}

	if dc != nil && op.n != nil {
		return op.n.process(dc, out)
	}
	return nil
}

// ToString answers a U32ToStr decode operation that will transform this output of thie U32ToU32 into a string
func (op *U32ToU32DOp) ToString(fn func(uint32) string) *U32ToStrDOp {
	result := &U32ToStrDOp{baseDOp: baseDOp{p: op}, fn: fn}
	result.do = result
	op.n = result
	return result
}

// AsFDOp is a deode operation that writes fields to metrics
type AsFDOp struct {
	baseDOp
	name string
}

func (op *AsFDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	var m telegraf.Metric
	if dc != nil {
		m = dc.currentMetric()
	}
	switch v := upstreamValue.(type) {
	case *uint64:
		if dc != nil {
			m.AddField(op.name, *v)
		}
	case *uint32:
		if dc != nil {
			m.AddField(op.name, *v)
		}
	case uint32:
		if dc != nil {
			m.AddField(op.name, v)
		}
	case *uint16:
		if dc != nil {
			m.AddField(op.name, *v)
		}
	case uint16:
		if dc != nil {
			m.AddField(op.name, v)
		}
	case *uint8:
		if dc != nil {
			m.AddField(op.name, *v)
		}
	case uint8:
		if dc != nil {
			m.AddField(op.name, v)
		}
	case string:
		if dc != nil {
			m.AddField(op.name, v)
		}
	default:
		return fmt.Errorf("AsF cannot process %T", v)
	}
	return nil
}

// AsTimestampDOp is a deode operation that sets the timestamp on the metric
type AsTimestampDOp struct {
	baseDOp
}

func (op *AsTimestampDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	var m telegraf.Metric
	if dc != nil {
		m = dc.currentMetric()
	}
	switch v := upstreamValue.(type) {
	case *uint32:
		if dc != nil {
			m.SetTime(time.Unix(int64(*v), 0))
			dc.timeHasBeenSet = true
		}
	default:
		return fmt.Errorf("can't process %T", upstreamValue)
	}
	return nil
}

// AsTDOp is a deode operation that writes tags to metrics
type AsTDOp struct {
	baseDOp
	name      string
	skipEmpty bool
}

func (op *AsTDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	var m telegraf.Metric
	if dc != nil {
		m = dc.currentMetric()
	}
	switch v := upstreamValue.(type) {
	case *uint32:
		if dc != nil {
			m.AddTag(op.name, fmt.Sprintf("%d", *v))
		}
	case uint32:
		if dc != nil {
			m.AddTag(op.name, fmt.Sprintf("%d", v))
		}
	case *uint16:
		if dc != nil {
			m.AddTag(op.name, fmt.Sprintf("%d", *v))
		}
	case uint16:
		if dc != nil {
			m.AddTag(op.name, fmt.Sprintf("%d", v))
		}
	case *uint8:
		if dc != nil {
			m.AddTag(op.name, fmt.Sprintf("%d", *v))
		}
	case uint8:
		if dc != nil {
			m.AddTag(op.name, fmt.Sprintf("%d", v))
		}
	case string:
		if dc != nil {
			if !op.skipEmpty || v != "" {
				m.AddTag(op.name, v)
			}
		}
	default:
		return fmt.Errorf("can't process %T", upstreamValue)
	}
	return nil
}

func (op *AsTDOp) prev() DirectiveOp {
	return op.p
}

// BytesToStrDOp is a decode operation that transforms []bytes to strings
type BytesToStrDOp struct {
	baseDOp
	len int
	fn  func([]byte) string
}

func (op *BytesToStrDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case []byte:
		if len(v) == op.len {
			if dc != nil {
				out := op.fn(v)
				if op.n != nil {
					return op.n.process(dc, out)
				}
			}
		} else {
			return fmt.Errorf("cannot process len(%d) as requrire %d", len(v), op.len)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// U32AssertDOp is a decode operation that asserts a particular uint32 value
type U32AssertDOp struct {
	baseDOp
	fn     func(uint32) bool
	fmtStr string
}

func (op *U32AssertDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case *uint32:
		if dc != nil && !op.fn(*v) {
			return fmt.Errorf(op.fmtStr, *v)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// U16AssertDOp is a decode operation that asserts a particular uint32 value
type U16AssertDOp struct {
	baseDOp
	fn     func(uint16) bool
	fmtStr string
}

func (op *U16AssertDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case *uint16:
		if dc != nil && !op.fn(*v) {
			return fmt.Errorf(op.fmtStr, *v)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// U32ToStrDOp is a decod eoperation that transforms a uint32 to a string
type U32ToStrDOp struct {
	baseDOp
	fn func(uint32) string
}

func (op *U32ToStrDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case uint32:
		if dc != nil && op.n != nil {
			op.n.process(dc, (op.fn(v)))
		}
	case *uint32:
		if dc != nil && op.n != nil {
			return op.n.process(dc, (op.fn(*v)))
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// BreakIf answers a BreakIf operation that will break the current decode operation chain, without an error, if the value processed
// is the supplied value
func (op *U32ToStrDOp) BreakIf(value string) *BreakIfDOp {
	result := &BreakIfDOp{baseDOp: baseDOp{p: op}, value: value}
	result.do = result
	op.n = result
	return result
}

// U16ToStrDOp is a decode operation that transforms a uint16 to a string
type U16ToStrDOp struct {
	baseDOp
	fn func(uint16) string
}

func (op *U16ToStrDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case *uint16:
		if dc != nil {
			return op.n.process(dc, (op.fn(*v)))
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// BreakIfDOp is a decode operation that will break the current outer iteration
type BreakIfDOp struct {
	baseDOp
	value string
}

func (op *BreakIfDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case string:
		if dc != nil {
			if v != op.value {
				op.n.process(dc, v)
			}
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// U16ToU16DOp is a decode operation that transfirms one uint16 to another uint16
type U16ToU16DOp struct {
	baseDOp
	fn func(uint16) uint16
}

func (op *U16ToU16DOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	var out uint16
	var err error
	switch v := upstreamValue.(type) {
	case *uint16:
		if dc != nil {
			out = op.fn(*v)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	if err != nil {
		return err
	}
	if op.n != nil && dc != nil {
		return op.n.process(dc, out)
	}
	return nil
}

// BytesToU32DOp is a decode operation that transforms a []byte to a uint32
type BytesToU32DOp struct {
	baseDOp
	len int
	fn  func([]byte) uint32
}

func (op *BytesToU32DOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case []byte:
		if len(v) == op.len {
			out := op.fn(v)
			if op.n != nil {
				return op.n.process(dc, out)
			}
		} else {
			return fmt.Errorf("cannot process %T as len(%d) != %d", upstreamValue, v, op.len)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// SetDOp is a decode operation that will Set a pointer to a value to be the value processed
type SetDOp struct {
	baseDOp
	ptr interface{}
}

func (op *SetDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case *uint32:
		ptr, ok := op.ptr.(*uint32)
		if ok {
			if dc != nil {
				*ptr = *v
			}
		} else {
			return fmt.Errorf("cannot process as ptr %T and not *uint32", op.ptr)
		}
	case uint32:
		ptr, ok := op.ptr.(*uint32)
		if ok {
			if dc != nil {
				*ptr = v
			}
		} else {
			return fmt.Errorf("cannot process as ptr %T and not *uint32", op.ptr)
		}
	case *uint16:
		ptr, ok := op.ptr.(*uint16)
		if ok {
			if dc != nil {
				*ptr = *v
			}
		} else {
			return fmt.Errorf("cannot process as ptr %T and not *uint16", op.ptr)
		}
	case uint16:
		ptr, ok := op.ptr.(*uint16)
		if ok {
			if dc != nil {
				*ptr = v
			}
		} else {
			return fmt.Errorf("cannot process as ptr %T and not *uint16", op.ptr)
		}
	case string:
		ptr, ok := op.ptr.(*string)
		if ok {
			if dc != nil {
				*ptr = v
			}
		} else {
			return fmt.Errorf("cannot process as ptr %T and not *string", op.ptr)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	if op.n != nil && dc != nil {
		return op.n.process(dc, upstreamValue)
	}
	return nil
}

// BytesToDOp is a decode operation that will transform []byte to interface{} according to a suppied function
type BytesToDOp struct {
	baseDOp
	len int
	fn  func([]byte) interface{}
}

func (op *BytesToDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	switch v := upstreamValue.(type) {
	case []byte:
		if len(v) == op.len {
			if dc != nil {
				out := op.fn(v)
				return op.n.process(dc, out)
			}
		} else {
			return fmt.Errorf("cannot process as len:%d required %d", len(v), op.len)
		}
	default:
		return fmt.Errorf("cannot process %T", upstreamValue)
	}
	return nil
}

// ErrorDOp is a decode operation that will generate an error
type ErrorDOp struct {
	baseDOp
	errorOnTestProcess bool
}

func (op *ErrorDOp) process(dc *DecodeContext, upstreamValue interface{}) error {
	if dc == nil && !op.errorOnTestProcess {
		return nil
	}
	return fmt.Errorf("Error Op")
}
