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

func U16Value(value *uint16) ValueDirective {
	return &valueDirective{value: value, noDecode: true, location: location(2)}
}

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

// U32Case answers a directive to be used within a Switch clause of a U32 directive
func Case(caseValue interface{}, dd Directive) CaseValueDirective {
	return &caseValueDirective{caseValue: caseValue, isDefault: false, equalsDd: dd}
}

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

func U32ToU32(fn func(uint32) uint32) *U32ToU32DOp {
	result := &U32ToU32DOp{fn: fn, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

func U32ToStr(fn func(uint32) string) *U32ToStrDOp {
	result := &U32ToStrDOp{baseDOp: baseDOp{loc: location(2)}, fn: fn}
	result.do = result
	return result
}

func U16ToStr(fn func(uint16) string) *U16ToStrDOp {
	result := &U16ToStrDOp{baseDOp: baseDOp{loc: location(2)}, fn: fn}
	result.do = result
	return result
}

func U16ToU16(fn func(uint16) uint16) *U16ToU16DOp {
	result := &U16ToU16DOp{baseDOp: baseDOp{loc: location(2)}, fn: fn}
	result.do = result
	return result
}

func AsF(name string) *AsFDOp {
	result := &AsFDOp{baseDOp: baseDOp{loc: location(2)}, name: name}
	result.do = result
	return result
}

func AsT(name string) *AsTDOp {
	result := &AsTDOp{name: name, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

func BytesToStr(len int, fn func([]byte) string) *BytesToStrDOp {
	result := &BytesToStrDOp{baseDOp: baseDOp{loc: location(2)}, len: len, fn: fn}
	result.do = result
	return result
}

func BytesTo(len int, fn func([]byte) interface{}) *BytesToDOp {
	result := &BytesToDOp{baseDOp: baseDOp{loc: location(2)}, len: len, fn: fn}
	result.do = result
	return result
}

func BytesToU32(len int, fn func([]byte) uint32) *BytesToU32DOp {
	result := &BytesToU32DOp{baseDOp: baseDOp{loc: location(2)}, len: len, fn: fn}
	result.do = result
	return result
}

func MapU32ToStr(m map[uint32]string) *U32ToStrDOp {
	result := &U32ToStrDOp{fn: func(in uint32) string {
		return m[in]
	}, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

func U32Assert(fn func(v uint32) bool, fmtStr string) *U32AssertDOp {
	result := &U32AssertDOp{baseDOp: baseDOp{loc: location(2)}, fn: fn, fmtStr: fmtStr}
	result.do = result
	return result
}

func MapU16ToStr(m map[uint16]string) *U16ToStrDOp {
	result := &U16ToStrDOp{baseDOp: baseDOp{loc: location(2)}, fn: func(in uint16) string {
		return m[in]
	}}
	result.do = result
	return result
}

func Set(ptr interface{}) *SetDOp {
	result := &SetDOp{ptr: ptr, baseDOp: baseDOp{loc: location(2)}}
	result.do = result
	return result
}

func ErrorDirective() Directive {
	return &errorDirective{}
}

func ErrorOp(errorOnTestProcess bool) *ErrorDOp {
	result := &ErrorDOp{baseDOp: baseDOp{loc: location(2)}, errorOnTestProcess: errorOnTestProcess}
	result.do = result
	return result

}
