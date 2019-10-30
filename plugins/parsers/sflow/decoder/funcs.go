package decoder

import (
	"fmt"
	"runtime"
	"time"

	"github.com/influxdata/telegraf/metric"
)

func location(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s:%d", file, line)
}

// U32 answers a directive for 32bit Unsigned Integers
func U32() ValueDirective {
	return &valueDirective{value: new(uint32), location: location(2), resetFn: func(in interface{}) {
		ui32Ptr, ok := in.(*uint32)
		if !ok {
			// Can't be tested
			panic("must be an *uint32")
		}
		(*ui32Ptr) = 0
	}}
}

// U16 answers a directive for 32bit Unsigned Integers
func U16() ValueDirective {
	return &valueDirective{value: new(uint16), location: location(2)}
}

// U16Value answers a directive that doesn't actually decode itself but reused a value previously decoded of type uint16
func U16Value(value *uint16) ValueDirective {
	return &valueDirective{value: value, noDecode: true, location: location(2)}
}

// Bytes answers a value directive that will decode the specified number (len) of bytes from the packet
func Bytes(len int) ValueDirective {
	return &valueDirective{value: make([]byte, len), location: location(2), resetFn: func(in interface{}) {
		b, ok := in.([]byte)
		if !ok {
			// Can't be tested
			panic("must be an []bytes")
		}
		for i := range b {
			b[i] = 0x0
		}
	}}
}

// Case answers a directive to be used within a Switch clause of a U32 directive
func Case(caseValue interface{}, dd Directive) CaseValueDirective {
	return &caseValueDirective{caseValue: caseValue, isDefault: false, equalsDd: dd}
}

// DefaultCase answers a case decoder directive that can be used as the default, catch all, of a Switch
func DefaultCase(dd Directive) CaseValueDirective {
	return &caseValueDirective{caseValue: nil, isDefault: true, equalsDd: dd}
}

// Ref answers a decoder that reuses, through referal, an existing U32 directive
func Ref(target interface{}) ValueDirective {
	if target == nil {
		panic("Ref given a nil reference")
	}
	r, ok := target.(*valueDirective)
	if !ok {
		panic(fmt.Sprintf("Ref not given a ValueDirective reference but a %T", target))
	}
	return &valueDirective{reference: r, value: r.value, location: location(2)}
}

// Seq ansers a directive that sequentially executes a list of provided directives
func Seq(decoders ...Directive) Directive {
	return &sequenceDirective{decoders: decoders, location: location(2)}
}

// OpenMetric answers a directive that opens a new metrics for collecting tags and fields
func OpenMetric() Directive {
	return &openMetric{location: location(2)}
}

// CloseMetric answers a directive that close the current metrics
func CloseMetric() Directive {
	return &closeMetric{location: location(2)}
}

// NewDecodeContext ansewers a new Decode Contect to support the process of decoding
func NewDecodeContext(trace bool) *DecodeContext {
	m, _ := metric.New("sflow", make(map[string]string), make(map[string]interface{}), time.Now())
	return &DecodeContext{preMetric: m, trace: trace}
}

// U32ToU32 answers a decode operation that transforms a uint32 to a uint32 via the supplied fn
func U32ToU32(fn func(uint32) uint32) *U32ToU32DOp {
	result := &U32ToU32DOp{fn: fn, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

// U32ToStr answers a decode operation that transforms a uint32 to a string via the supplied fn
func U32ToStr(fn func(uint32) string) *U32ToStrDOp {
	result := &U32ToStrDOp{baseDOp: baseDOp{loc: location(2)}, fn: fn}
	result.do = result
	return result
}

// U16ToStr answers a decode operation that transforms a uint16 to a string via the supplied fn
func U16ToStr(fn func(uint16) string) *U16ToStrDOp {
	result := &U16ToStrDOp{baseDOp: baseDOp{loc: location(2)}, fn: fn}
	result.do = result
	return result
}

// U16ToU16 answers a decode operation that transforms a uint16 to a uint16 via the supplied fn
func U16ToU16(fn func(uint16) uint16) *U16ToU16DOp {
	result := &U16ToU16DOp{baseDOp: baseDOp{loc: location(2)}, fn: fn}
	result.do = result
	return result
}

// AsF answers a decode operation that will output a field into the open metric with the given name
func AsF(name string) *AsFDOp {
	result := &AsFDOp{baseDOp: baseDOp{loc: location(2)}, name: name}
	result.do = result
	return result
}

// AsT answers a decode operation that will output a tag into the open metric with the given name
func AsT(name string) *AsTDOp {
	result := &AsTDOp{name: name, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

// BytesToStr answers a decode operation that transforms a []bytes to a string via the supplied fn
func BytesToStr(len int, fn func([]byte) string) *BytesToStrDOp {
	result := &BytesToStrDOp{baseDOp: baseDOp{loc: location(2)}, len: len, fn: fn}
	result.do = result
	return result
}

// BytesTo answers a decode operation that transforms a []bytes to a interface{} via the supplied fn
func BytesTo(len int, fn func([]byte) interface{}) *BytesToDOp {
	result := &BytesToDOp{baseDOp: baseDOp{loc: location(2)}, len: len, fn: fn}
	result.do = result
	return result
}

// BytesToU32 answers a decode operation that transforms a []bytes to an uint32 via the supplied fn
func BytesToU32(len int, fn func([]byte) uint32) *BytesToU32DOp {
	result := &BytesToU32DOp{baseDOp: baseDOp{loc: location(2)}, len: len, fn: fn}
	result.do = result
	return result
}

// MapU32ToStr answers a decode operation that maps an uint32 to a string via the supplied map
func MapU32ToStr(m map[uint32]string) *U32ToStrDOp {
	result := &U32ToStrDOp{fn: func(in uint32) string {
		return m[in]
	}, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

// U32Assert answers a decode operation that will assert the uint32 is a particulr value or generate an error
func U32Assert(fn func(v uint32) bool, fmtStr string) *U32AssertDOp {
	result := &U32AssertDOp{baseDOp: baseDOp{loc: location(2)}, fn: fn, fmtStr: fmtStr}
	result.do = result
	return result
}

// MapU16ToStr answers a decode operation that maps an uint16 to a string via the supplied map
func MapU16ToStr(m map[uint16]string) *U16ToStrDOp {
	result := &U16ToStrDOp{baseDOp: baseDOp{loc: location(2)}, fn: func(in uint16) string {
		return m[in]
	}}
	result.do = result
	return result
}

// Set answers a decode operation that will set the supplied *value to the value passed through the operation
func Set(ptr interface{}) *SetDOp {
	result := &SetDOp{ptr: ptr, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

// ErrorDirective answers a decode directive that will generate an error
func ErrorDirective() Directive {
	return &errorDirective{}
}

// ErrorOp answers a decode operation that will generate an error
func ErrorOp(errorOnTestProcess bool) *ErrorDOp {
	result := &ErrorDOp{baseDOp: baseDOp{loc: location(2)}, errorOnTestProcess: errorOnTestProcess}
	result.do = result
	return result

}
