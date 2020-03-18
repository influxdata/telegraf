package decoder

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf/metric"
)

// U32 answers a directive for 32bit Unsigned Integers
func U32() ValueDirective {
	return &valueDirective{value: new(uint32)}
}

// U64 answers a directive for 64bit Unsigned Integers
func U64() ValueDirective {
	return &valueDirective{value: new(uint64)}
}

// U8 answers a directive for 8bit Unsigned Integers
func U8() ValueDirective {
	return &valueDirective{value: new(uint8)}
}

// U16 answers a directive for 32bit Unsigned Integers
func U16() ValueDirective {
	return &valueDirective{value: new(uint16)}
}

// U16Value answers a directive that doesn't actually decode itself but reused a value previously decoded of type uint16
func U16Value(value *uint16) ValueDirective {
	return &valueDirective{value: value, noDecode: true}
}

// Bytes answers a value directive that will decode the specified number (len) of bytes from the packet
func Bytes(len int) ValueDirective {
	return &valueDirective{value: make([]byte, len)}
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
	return &valueDirective{reference: r, value: r.value}
}

// Seq ansers a directive that sequentially executes a list of provided directives
func Seq(decoders ...Directive) Directive {
	return &sequenceDirective{decoders: decoders}
}

func SeqOf(decoders []Directive) Directive {
	return &sequenceDirective{decoders: decoders}
}

// OpenMetric answers a directive that opens a new metrics for collecting tags and fields
func OpenMetric(name string) Directive {
	return &openMetric{name: name}
}

// CloseMetric answers a directive that close the current metrics
func CloseMetric() Directive {
	return &closeMetric{}
}

// NewDecodeContext ansewers a new Decode Contect to support the process of decoding
func NewDecodeContext() *DecodeContext {
	m, _ := metric.New("sflow", make(map[string]string), make(map[string]interface{}), time.Now())
	return &DecodeContext{preMetric: m}
}

// U32ToU32 answers a decode operation that transforms a uint32 to a uint32 via the supplied fn
func U32ToU32(fn func(uint32) uint32) *U32ToU32DOp {
	result := &U32ToU32DOp{fn: fn, baseDOp: baseDOp{}}
	result.do = result
	return result
}

// U32ToStr answers a decode operation that transforms a uint32 to a string via the supplied fn
func U32ToStr(fn func(uint32) string) *U32ToStrDOp {
	result := &U32ToStrDOp{baseDOp: baseDOp{}, fn: fn}
	result.do = result
	return result
}

// U16ToStr answers a decode operation that transforms a uint16 to a string via the supplied fn
func U16ToStr(fn func(uint16) string) *U16ToStrDOp {
	result := &U16ToStrDOp{baseDOp: baseDOp{}, fn: fn}
	result.do = result
	return result
}

// U16ToU16 answers a decode operation that transforms a uint16 to a uint16 via the supplied fn
func U16ToU16(fn func(uint16) uint16) *U16ToU16DOp {
	result := &U16ToU16DOp{baseDOp: baseDOp{}, fn: fn}
	result.do = result
	return result
}

// AsF answers a decode operation that will output a field into the open metric with the given name
func AsF(name string) *AsFDOp {
	result := &AsFDOp{baseDOp: baseDOp{}, name: name}
	result.do = result
	return result
}

// AsT answers a decode operation that will output a tag into the open metric with the given name
func AsT(name string) *AsTDOp {
	result := &AsTDOp{name: name, baseDOp: baseDOp{}}
	result.do = result
	return result
}

// AsTimestamp answers a decode operation that will set the tiemstamp on the metric
func AsTimestamp() *AsTimestampDOp {
	result := &AsTimestampDOp{baseDOp: baseDOp{}}
	result.do = result
	return result
}

// BytesToStr answers a decode operation that transforms a []bytes to a string via the supplied fn
func BytesToStr(len int, fn func([]byte) string) *BytesToStrDOp {
	result := &BytesToStrDOp{baseDOp: baseDOp{}, len: len, fn: fn}
	result.do = result
	return result
}

// BytesTo answers a decode operation that transforms a []bytes to a interface{} via the supplied fn
func BytesTo(len int, fn func([]byte) interface{}) *BytesToDOp {
	result := &BytesToDOp{baseDOp: baseDOp{}, len: len, fn: fn}
	result.do = result
	return result
}

// BytesToU32 answers a decode operation that transforms a []bytes to an uint32 via the supplied fn
func BytesToU32(len int, fn func([]byte) uint32) *BytesToU32DOp {
	result := &BytesToU32DOp{baseDOp: baseDOp{}, len: len, fn: fn}
	result.do = result
	return result
}

// MapU32ToStr answers a decode operation that maps an uint32 to a string via the supplied map
func MapU32ToStr(m map[uint32]string) *U32ToStrDOp {
	result := &U32ToStrDOp{fn: func(in uint32) string {
		return m[in]
	}, baseDOp: baseDOp{}}
	result.do = result
	return result
}

// U32Assert answers a decode operation that will assert the uint32 is a particulr value or generate an error
func U32Assert(fn func(v uint32) bool, fmtStr string) *U32AssertDOp {
	result := &U32AssertDOp{baseDOp: baseDOp{}, fn: fn, fmtStr: fmtStr}
	result.do = result
	return result
}

func U16Assert(fn func(v uint16) bool, fmtStr string) *U16AssertDOp {
	result := &U16AssertDOp{baseDOp: baseDOp{}, fn: fn, fmtStr: fmtStr}
	result.do = result
	return result
}

// MapU16ToStr answers a decode operation that maps an uint16 to a string via the supplied map
func MapU16ToStr(m map[uint16]string) *U16ToStrDOp {
	result := &U16ToStrDOp{baseDOp: baseDOp{}, fn: func(in uint16) string {
		return m[in]
	}}
	result.do = result
	return result
}

// Set answers a decode operation that will set the supplied *value to the value passed through the operation
func Set(ptr interface{}) *SetDOp {
	result := &SetDOp{ptr: ptr, baseDOp: baseDOp{}}
	result.do = result
	return result
}

// ErrorDirective answers a decode directive that will generate an error
func ErrorDirective() Directive {
	return &errorDirective{}
}

// ErrorOp answers a decode operation that will generate an error
func ErrorOp(errorOnTestProcess bool) *ErrorDOp {
	result := &ErrorDOp{baseDOp: baseDOp{}, errorOnTestProcess: errorOnTestProcess}
	result.do = result
	return result

}

// Notify answers a decode directive that will notify the supplied function upon execution
func Notify(fn func()) Directive {
	return &notifyDirective{fn}
}

// Nop answer a decode directive that is the null, benign, deocder
func Nop() Directive {
	return Notify(func() {})
}
