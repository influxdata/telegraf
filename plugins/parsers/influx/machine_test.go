package influx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestingHandler struct {
	results []Result
}

func (h *TestingHandler) SetMeasurement(name []byte) {
	mname := Result{
		Name:  Measurement,
		Value: name,
	}
	h.results = append(h.results, mname)
}

func (h *TestingHandler) AddTag(key []byte, value []byte) {
	tagkey := Result{
		Name:  TagKey,
		Value: key,
	}
	tagvalue := Result{
		Name:  TagValue,
		Value: value,
	}
	h.results = append(h.results, tagkey, tagvalue)
}

func (h *TestingHandler) AddInt(key []byte, value []byte) {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldInt,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
}

func (h *TestingHandler) AddUint(key []byte, value []byte) {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldUint,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
}

func (h *TestingHandler) AddFloat(key []byte, value []byte) {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldFloat,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
}

func (h *TestingHandler) AddString(key []byte, value []byte) {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldString,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
}

func (h *TestingHandler) AddBool(key []byte, value []byte) {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldBool,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
}

func (h *TestingHandler) SetTimestamp(tm []byte) {
	timestamp := Result{
		Name:  Timestamp,
		Value: tm,
	}
	h.results = append(h.results, timestamp)
}

func (h *TestingHandler) Reset() {
}

func (h *TestingHandler) Results() []Result {
	return h.results
}

func (h *TestingHandler) AddError(err error) {
	e := Result{
		err: err,
	}
	h.results = append(h.results, e)
}

type BenchmarkingHandler struct {
}

func (h *BenchmarkingHandler) SetMeasurement(name []byte) {
}

func (h *BenchmarkingHandler) AddTag(key []byte, value []byte) {
}

func (h *BenchmarkingHandler) AddInt(key []byte, value []byte) {
}

func (h *BenchmarkingHandler) AddUint(key []byte, value []byte) {
}

func (h *BenchmarkingHandler) AddFloat(key []byte, value []byte) {
}

func (h *BenchmarkingHandler) AddString(key []byte, value []byte) {
}

func (h *BenchmarkingHandler) AddBool(key []byte, value []byte) {
}

func (h *BenchmarkingHandler) SetTimestamp(tm []byte) {
}

func (h *BenchmarkingHandler) Reset() {
}

type TokenType int

const (
	NoMatch TokenType = iota
	Measurement
	TagKey
	TagValue
	FieldKey
	FieldString
	FieldInt
	FieldUint
	FieldFloat
	FieldBool
	Timestamp
	EOL
	EOF
	Punc
	WhiteSpace
)

func (t TokenType) String() string {
	switch t {
	case NoMatch:
		return "NoMatch"
	case Measurement:
		return "Measurement"
	case TagKey:
		return "TagKey"
	case TagValue:
		return "TagValue"
	case FieldKey:
		return "FieldKey"
	case FieldInt:
		return "FieldInt"
	case FieldUint:
		return "FieldUint"
	case FieldFloat:
		return "FieldFloat"
	case FieldString:
		return "FieldString"
	case FieldBool:
		return "FieldBool"
	case Timestamp:
		return "Timestamp"
	case EOL:
		return "EOL"
	case EOF:
		return "EOF"
	case Punc:
		return "Punc"
	case WhiteSpace:
		return "WhiteSpace"
	default:
		panic("Unknown TokenType")
	}
}

type Token struct {
	Name  TokenType
	Value []byte
}

func (t Token) String() string {
	return fmt.Sprintf("(%s %q)", t.Name, t.Value)
}

type Result struct {
	Name  TokenType
	Value []byte
	err   error
}

func (r Result) String() string {
	return fmt.Sprintf("(%s, %q, %v)", r.Name, r.Value, r.err)
}

var tests = []struct {
	name    string
	input   []byte
	results []Result
	err     error
}{
	{
		name:    "empty string",
		input:   []byte(""),
		results: nil,
	},
	{
		name:  "minimal",
		input: []byte("cpu value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "newline",
		input: []byte("cpu value=42\n"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "minimal with timestamp",
		input: []byte("cpu value=42 1516241192000000000"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				Name:  Timestamp,
				Value: []byte("1516241192000000000"),
			},
		},
	},
	{
		name:  "measurement escape non-special",
		input: []byte(`c\pu value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte(`c\pu`),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "measurement escaped trailing backslash",
		input: []byte(`cpu\\ value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte(`cpu\\`),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "single char measurement",
		input: []byte("c value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("c"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "escape backslash in measurement",
		input: []byte(`cp\\u value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte(`cp\\u`),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "measurement escape space",
		input: []byte(`cpu\ abc value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte(`cpu\ abc`),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "scientific float",
		input: []byte("cpu value=42e0"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42e0"),
			},
		},
	},
	{
		name:  "scientific float negative mantissa",
		input: []byte("cpu value=-42e0"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("-42e0"),
			},
		},
	},
	{
		name:  "scientific float negative exponent",
		input: []byte("cpu value=42e-1"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42e-1"),
			},
		},
	},
	{
		name:  "scientific float big e",
		input: []byte("cpu value=42E0"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42E0"),
			},
		},
	},
	{
		name:  "scientific float missing exponent",
		input: []byte("cpu value=42E"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "float with decimal",
		input: []byte("cpu value=42.2"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42.2"),
			},
		},
	},
	{
		name:  "negative float",
		input: []byte("cpu value=-42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("-42"),
			},
		},
	},
	{
		name:  "float without integer digits",
		input: []byte("cpu value=.42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte(".42"),
			},
		},
	},
	{
		name:  "float without integer digits negative",
		input: []byte("cpu value=-.42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("-.42"),
			},
		},
	},
	{
		name:  "float with multiple leading 0",
		input: []byte("cpu value=00.42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("00.42"),
			},
		},
	},
	{
		name:  "invalid float with only dot",
		input: []byte("cpu value=."),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "multiple fields",
		input: []byte("cpu x=42,y=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("x"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("y"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "integer field",
		input: []byte("cpu value=42i"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldInt,
				Value: []byte("42i"),
			},
		},
	},
	{
		name:  "negative integer field",
		input: []byte("cpu value=-42i"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldInt,
				Value: []byte("-42i"),
			},
		},
	},
	{
		name:  "zero integer field",
		input: []byte("cpu value=0i"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldInt,
				Value: []byte("0i"),
			},
		},
	},
	{
		name:  "negative zero integer field",
		input: []byte("cpu value=-0i"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldInt,
				Value: []byte("-0i"),
			},
		},
	},
	{
		name:  "invalid field",
		input: []byte("cpu value=howdy"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "string field",
		input: []byte(`cpu value="42"`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldString,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "bool field",
		input: []byte(`cpu value=true`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldBool,
				Value: []byte("true"),
			},
		},
	},
	{
		name:  "tag",
		input: []byte(`cpu,host=localhost value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagKey,
				Value: []byte("host"),
			},
			Result{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "tag key escape space",
		input: []byte(`cpu,h\ ost=localhost value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagKey,
				Value: []byte(`h\ ost`),
			},
			Result{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "tag key escape comma",
		input: []byte(`cpu,h\,ost=localhost value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagKey,
				Value: []byte(`h\,ost`),
			},
			Result{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "tag key escape equal",
		input: []byte(`cpu,h\=ost=localhost value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagKey,
				Value: []byte(`h\=ost`),
			},
			Result{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "multiple tags",
		input: []byte(`cpu,host=localhost,cpu=cpu0 value=42`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagKey,
				Value: []byte("host"),
			},
			Result{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			Result{
				Name:  TagKey,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagValue,
				Value: []byte("cpu0"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "tag invalid missing separator",
		input: []byte("cpu,xyzzy value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid missing value",
		input: []byte("cpu,xyzzy= value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid unescaped space",
		input: []byte("cpu,h ost=localhost value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid unescaped comma",
		input: []byte("cpu,h,ost=localhost value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid unescaped equals",
		input: []byte("cpu,h=ost=localhost value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrTagParse,
			},
		},
	},
	{
		name:  "timestamp negative",
		input: []byte("cpu value=42 -1"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				Name:  Timestamp,
				Value: []byte("-1"),
			},
		},
	},
	{
		name:  "timestamp zero",
		input: []byte("cpu value=42 0"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				Name:  Timestamp,
				Value: []byte("0"),
			},
		},
	},
	{
		name:  "multiline",
		input: []byte("cpu value=42\n\n\ncpu value=43\n"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("43"),
			},
		},
	},
	{
		name:  "error recovery",
		input: []byte("cpu value=howdy\ncpu\ncpu value=42\n"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
			Result{
				err: ErrFieldParse,
			},
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "line whitespace",
		input: []byte("   cpu   value=42  1516241192000000000  \n\n cpu value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				Name:  Timestamp,
				Value: []byte("1516241192000000000"),
			},
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "leading newline",
		input: []byte("\ncpu value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "invalid missing field value",
		input: []byte("cpu value="),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid eof field key",
		input: []byte("cpu value"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid measurement only",
		input: []byte("cpu"),
		results: []Result{
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid measurement only eol",
		input: []byte("cpu\n"),
		results: []Result{
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid missing tag",
		input: []byte("cpu, value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrTagParse,
			},
		},
	},
	{
		name:  "invalid missing field",
		input: []byte("cpu,x=y "),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  TagKey,
				Value: []byte("x"),
			},
			Result{
				Name:  TagValue,
				Value: []byte("y"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid too many fields",
		input: []byte("cpu value=42 value=43"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				err: ErrTimestampParse,
			},
		},
	},
	{
		name:  "invalid timestamp too long",
		input: []byte("cpu value=42 12345678901234567890"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				err: ErrTimestampParse,
			},
		},
	},
	{
		name:  "invalid open string field",
		input: []byte(`cpu value="42 12345678901234567890`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid newline in string field",
		input: []byte("cpu value=\"4\n2\""),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid field value",
		input: []byte(`cpu value=howdy`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				err: ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid quoted timestamp",
		input: []byte(`cpu value=42 "12345678901234567890"`),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			Result{
				err: ErrTimestampParse,
			},
		},
	},
	{
		name:  "commented line",
		input: []byte("# blah blah\ncpu value=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "end with comment",
		input: []byte("cpu value=42\n# blah blah"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "end with comment and whitespace",
		input: []byte("cpu value=42\n# blah blah\n\n  "),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
	{
		name:  "unicode",
		input: []byte("cpu ☺=42"),
		results: []Result{
			Result{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			Result{
				Name:  FieldKey,
				Value: []byte("☺"),
			},
			Result{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
		},
	},
}

func TestMachine(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TestingHandler{}
			fsm := NewMachine(handler)
			fsm.SetData(tt.input)

			count := 0
			for fsm.ParseLine() {
				if fsm.Err() != nil {
					handler.AddError(fsm.Err())
				}
				count++
				if count > 20 {
					break
				}
			}

			if fsm.Err() != nil {
				handler.AddError(fsm.Err())
			}

			results := handler.Results()
			require.Equal(t, tt.results, results)
		})
	}
}

func BenchmarkMachine(b *testing.B) {
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			handler := &BenchmarkingHandler{}
			fsm := NewMachine(handler)

			for n := 0; n < b.N; n++ {
				fsm.SetData(tt.input)

				for fsm.ParseLine() {
				}
			}
		})
	}
}

func TestMachineProcstat(t *testing.T) {
	input := []byte("procstat,exe=bash,process_name=bash voluntary_context_switches=42i,memory_rss=5103616i,rlimit_memory_data_hard=2147483647i,cpu_time_user=0.02,rlimit_file_locks_soft=2147483647i,pid=29417i,cpu_time_nice=0,rlimit_memory_locked_soft=65536i,read_count=259i,rlimit_memory_vms_hard=2147483647i,memory_swap=0i,rlimit_num_fds_soft=1024i,rlimit_nice_priority_hard=0i,cpu_time_soft_irq=0,cpu_time=0i,rlimit_memory_locked_hard=65536i,realtime_priority=0i,signals_pending=0i,nice_priority=20i,cpu_time_idle=0,memory_stack=139264i,memory_locked=0i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,cpu_time_guest=0,cpu_time_guest_nice=0,rlimit_memory_data_soft=2147483647i,read_bytes=0i,rlimit_cpu_time_soft=2147483647i,involuntary_context_switches=2i,write_bytes=106496i,cpu_time_system=0,cpu_time_irq=0,cpu_usage=0,memory_vms=21659648i,memory_data=1576960i,rlimit_memory_stack_hard=2147483647i,num_threads=1i,cpu_time_stolen=0,rlimit_memory_rss_soft=2147483647i,rlimit_realtime_priority_soft=0i,num_fds=4i,write_count=35i,rlimit_signals_pending_soft=78994i,cpu_time_steal=0,rlimit_num_fds_hard=4096i,rlimit_file_locks_hard=2147483647i,rlimit_cpu_time_hard=2147483647i,rlimit_signals_pending_hard=78994i,rlimit_nice_priority_soft=0i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_realtime_priority_hard=0i 1517620624000000000")
	handler := &TestingHandler{}
	fsm := NewMachine(handler)
	fsm.SetData(input)
	for fsm.ParseLine() {
	}
}

func BenchmarkMachineProcstat(b *testing.B) {
	input := []byte("procstat,exe=bash,process_name=bash voluntary_context_switches=42i,memory_rss=5103616i,rlimit_memory_data_hard=2147483647i,cpu_time_user=0.02,rlimit_file_locks_soft=2147483647i,pid=29417i,cpu_time_nice=0,rlimit_memory_locked_soft=65536i,read_count=259i,rlimit_memory_vms_hard=2147483647i,memory_swap=0i,rlimit_num_fds_soft=1024i,rlimit_nice_priority_hard=0i,cpu_time_soft_irq=0,cpu_time=0i,rlimit_memory_locked_hard=65536i,realtime_priority=0i,signals_pending=0i,nice_priority=20i,cpu_time_idle=0,memory_stack=139264i,memory_locked=0i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,cpu_time_guest=0,cpu_time_guest_nice=0,rlimit_memory_data_soft=2147483647i,read_bytes=0i,rlimit_cpu_time_soft=2147483647i,involuntary_context_switches=2i,write_bytes=106496i,cpu_time_system=0,cpu_time_irq=0,cpu_usage=0,memory_vms=21659648i,memory_data=1576960i,rlimit_memory_stack_hard=2147483647i,num_threads=1i,cpu_time_stolen=0,rlimit_memory_rss_soft=2147483647i,rlimit_realtime_priority_soft=0i,num_fds=4i,write_count=35i,rlimit_signals_pending_soft=78994i,cpu_time_steal=0,rlimit_num_fds_hard=4096i,rlimit_file_locks_hard=2147483647i,rlimit_cpu_time_hard=2147483647i,rlimit_signals_pending_hard=78994i,rlimit_nice_priority_soft=0i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_realtime_priority_hard=0i 1517620624000000000")
	handler := &BenchmarkingHandler{}
	fsm := NewMachine(handler)
	for n := 0; n < b.N; n++ {
		fsm.SetData(input)
		for fsm.ParseLine() {
		}
	}
}
