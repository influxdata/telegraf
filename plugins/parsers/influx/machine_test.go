package influx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/stretchr/testify/require"
)

type TestingHandler struct {
	results []Result
}

func (h *TestingHandler) SetMeasurement(name []byte) error {
	mname := Result{
		Name:  Measurement,
		Value: name,
	}
	h.results = append(h.results, mname)
	return nil
}

func (h *TestingHandler) AddTag(key []byte, value []byte) error {
	tagkey := Result{
		Name:  TagKey,
		Value: key,
	}
	tagvalue := Result{
		Name:  TagValue,
		Value: value,
	}
	h.results = append(h.results, tagkey, tagvalue)
	return nil
}

func (h *TestingHandler) AddInt(key []byte, value []byte) error {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldInt,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
	return nil
}

func (h *TestingHandler) AddUint(key []byte, value []byte) error {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldUint,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
	return nil
}

func (h *TestingHandler) AddFloat(key []byte, value []byte) error {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldFloat,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
	return nil
}

func (h *TestingHandler) AddString(key []byte, value []byte) error {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldString,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
	return nil
}

func (h *TestingHandler) AddBool(key []byte, value []byte) error {
	fieldkey := Result{
		Name:  FieldKey,
		Value: key,
	}
	fieldvalue := Result{
		Name:  FieldBool,
		Value: value,
	}
	h.results = append(h.results, fieldkey, fieldvalue)
	return nil
}

func (h *TestingHandler) SetTimestamp(tm []byte) error {
	timestamp := Result{
		Name:  Timestamp,
		Value: tm,
	}
	h.results = append(h.results, timestamp)
	return nil
}

func (h *TestingHandler) Result(err error) {
	var res Result
	if err == nil {
		res = Result{
			Name: Success,
		}
	} else {
		res = Result{
			Name: Error,
			err:  err,
		}
	}
	h.results = append(h.results, res)
}

func (h *TestingHandler) Results() []Result {
	return h.results
}

type BenchmarkingHandler struct {
}

func (h *BenchmarkingHandler) SetMeasurement(name []byte) error {
	return nil
}

func (h *BenchmarkingHandler) AddTag(key []byte, value []byte) error {
	return nil
}

func (h *BenchmarkingHandler) AddInt(key []byte, value []byte) error {
	return nil
}

func (h *BenchmarkingHandler) AddUint(key []byte, value []byte) error {
	return nil
}

func (h *BenchmarkingHandler) AddFloat(key []byte, value []byte) error {
	return nil
}

func (h *BenchmarkingHandler) AddString(key []byte, value []byte) error {
	return nil
}

func (h *BenchmarkingHandler) AddBool(key []byte, value []byte) error {
	return nil
}

func (h *BenchmarkingHandler) SetTimestamp(tm []byte) error {
	return nil
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
	Success
	Error
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
	case Success:
		return "Success"
	case Error:
		return "Error"
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
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "newline",
		input: []byte("cpu value=42\n"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "minimal with timestamp",
		input: []byte("cpu value=42 1516241192000000000"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name:  Timestamp,
				Value: []byte("1516241192000000000"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "measurement escape non-special",
		input: []byte(`c\pu value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte(`c\pu`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "measurement escaped trailing backslash",
		input: []byte(`cpu\\ value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte(`cpu\\`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "single char measurement",
		input: []byte("c value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("c"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "escape backslash in measurement",
		input: []byte(`cp\\u value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte(`cp\\u`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "measurement escape space",
		input: []byte(`cpu\ abc value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte(`cpu\ abc`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "scientific float",
		input: []byte("cpu value=42e0"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42e0"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "scientific float negative mantissa",
		input: []byte("cpu value=-42e0"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("-42e0"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "scientific float negative exponent",
		input: []byte("cpu value=42e-1"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42e-1"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "scientific float big e",
		input: []byte("cpu value=42E0"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42E0"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "scientific float missing exponent",
		input: []byte("cpu value=42E"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "float with decimal",
		input: []byte("cpu value=42.2"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42.2"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "negative float",
		input: []byte("cpu value=-42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("-42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "float without integer digits",
		input: []byte("cpu value=.42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte(".42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "float without integer digits negative",
		input: []byte("cpu value=-.42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("-.42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "float with multiple leading 0",
		input: []byte("cpu value=00.42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("00.42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "invalid float with only dot",
		input: []byte("cpu value=."),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "multiple fields",
		input: []byte("cpu x=42,y=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("x"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name:  FieldKey,
				Value: []byte("y"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "integer field",
		input: []byte("cpu value=42i"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldInt,
				Value: []byte("42i"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "negative integer field",
		input: []byte("cpu value=-42i"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldInt,
				Value: []byte("-42i"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "zero integer field",
		input: []byte("cpu value=0i"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldInt,
				Value: []byte("0i"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "negative zero integer field",
		input: []byte("cpu value=-0i"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldInt,
				Value: []byte("-0i"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "integer field overflow okay",
		input: []byte("cpu value=9223372036854775808i"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldInt,
				Value: []byte("9223372036854775808i"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "invalid field",
		input: []byte("cpu value=howdy"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "string field",
		input: []byte("cpu value=\"42\""),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldString,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "newline in string field",
		input: []byte("cpu value=\"4\n2\""),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldString,
				Value: []byte("4\n2"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "bool field",
		input: []byte("cpu value=true"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldBool,
				Value: []byte("true"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag",
		input: []byte("cpu,host=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte("host"),
			},
			{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag key escape space",
		input: []byte("cpu,h\\ ost=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte(`h\ ost`),
			},
			{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag key escape comma",
		input: []byte("cpu,h\\,ost=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte(`h\,ost`),
			},
			{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag key escape equal",
		input: []byte("cpu,h\\=ost=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte(`h\=ost`),
			},
			{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "multiple tags",
		input: []byte("cpu,host=localhost,cpu=cpu0 value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte("host"),
			},
			{
				Name:  TagValue,
				Value: []byte("localhost"),
			},
			{
				Name:  TagKey,
				Value: []byte("cpu"),
			},
			{
				Name:  TagValue,
				Value: []byte("cpu0"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag value escape space",
		input: []byte(`cpu,host=two\ words value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte("host"),
			},
			{
				Name:  TagValue,
				Value: []byte(`two\ words`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag value double escape space",
		input: []byte(`cpu,host=two\\ words value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte("host"),
			},
			{
				Name:  TagValue,
				Value: []byte(`two\\ words`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag value triple escape space",
		input: []byte(`cpu,host=two\\\ words value=42`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte("host"),
			},
			{
				Name:  TagValue,
				Value: []byte(`two\\\ words`),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "tag invalid missing separator",
		input: []byte("cpu,xyzzy value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid missing value",
		input: []byte("cpu,xyzzy= value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid unescaped space",
		input: []byte("cpu,h ost=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid unescaped comma",
		input: []byte("cpu,h,ost=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "tag invalid unescaped equals",
		input: []byte("cpu,h=ost=localhost value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "timestamp negative",
		input: []byte("cpu value=42 -1"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name:  Timestamp,
				Value: []byte("-1"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "timestamp zero",
		input: []byte("cpu value=42 0"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name:  Timestamp,
				Value: []byte("0"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "multiline",
		input: []byte("cpu value=42\n\n\n\ncpu value=43"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("43"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "error recovery",
		input: []byte("cpu value=howdy,value2=42\ncpu\ncpu value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "line whitespace",
		input: []byte("   cpu   value=42  1516241192000000000  \n\n cpu value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name:  Timestamp,
				Value: []byte("1516241192000000000"),
			},
			{
				Name: Success,
			},
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "leading newline",
		input: []byte("\ncpu value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "invalid missing field value",
		input: []byte("cpu value="),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid eof field key",
		input: []byte("cpu value"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid measurement only",
		input: []byte("cpu"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "invalid measurement char",
		input: []byte(","),
		results: []Result{
			{
				Name: Error,
				err:  influx.ErrNameParse,
			},
		},
	},
	{
		name:  "invalid missing tag",
		input: []byte("cpu, value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrTagParse,
			},
		},
	},
	{
		name:  "invalid missing field",
		input: []byte("cpu,x=y "),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  TagKey,
				Value: []byte("x"),
			},
			{
				Name:  TagValue,
				Value: []byte("y"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid too many fields",
		input: []byte("cpu value=42 value=43"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Error,
				err:  influx.ErrTimestampParse,
			},
		},
	},
	{
		name:  "invalid timestamp too long",
		input: []byte("cpu value=42 12345678901234567890"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Error,
				err:  influx.ErrTimestampParse,
			},
		},
	},
	{
		name:  "invalid open string field",
		input: []byte(`cpu value="42 12345678901234567890`),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid field value",
		input: []byte("cpu value=howdy"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name: Error,
				err:  influx.ErrFieldParse,
			},
		},
	},
	{
		name:  "invalid quoted timestamp",
		input: []byte("cpu value=42 \"12345678901234567890\""),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Error,
				err:  influx.ErrTimestampParse,
			},
		},
	},
	{
		name:    "comment only",
		input:   []byte("# blah blah"),
		results: []Result(nil),
	},
	{
		name:  "commented line",
		input: []byte("# blah blah\ncpu value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "middle comment",
		input: []byte("cpu value=42\n# blah blah\ncpu value=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "end with comment",
		input: []byte("cpu value=42\n# blah blah"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "end with comment and whitespace",
		input: []byte("cpu value=42\n# blah blah\n\n  "),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("value"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
	{
		name:  "unicode",
		input: []byte("cpu ☺=42"),
		results: []Result{
			{
				Name:  Measurement,
				Value: []byte("cpu"),
			},
			{
				Name:  FieldKey,
				Value: []byte("☺"),
			},
			{
				Name:  FieldFloat,
				Value: []byte("42"),
			},
			{
				Name: Success,
			},
		},
	},
}

func TestMachine(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TestingHandler{}
			fsm := influx.NewMachine(handler)
			fsm.SetData(tt.input)

			for i := 0; i < 20; i++ {
				err := fsm.Next()
				if err != nil && err == influx.EOF {
					break
				}
				handler.Result(err)
			}

			results := handler.Results()
			require.Equal(t, tt.results, results)
		})
	}
}

func TestMachinePosition(t *testing.T) {
	var tests = []struct {
		name   string
		input  []byte
		lineno int
		column int
	}{
		{
			name:   "empty string",
			input:  []byte(""),
			lineno: 1,
			column: 1,
		},
		{
			name:   "minimal",
			input:  []byte("cpu value=42"),
			lineno: 1,
			column: 13,
		},
		{
			name:   "one newline",
			input:  []byte("cpu value=42\ncpu value=42"),
			lineno: 2,
			column: 13,
		},
		{
			name:   "several newlines",
			input:  []byte("cpu value=42\n\n\n"),
			lineno: 4,
			column: 1,
		},
		{
			name:   "error on second line",
			input:  []byte("cpu value=42\ncpu value=invalid"),
			lineno: 2,
			column: 11,
		},
		{
			name:   "error after comment line",
			input:  []byte("cpu value=42\n# comment\ncpu value=invalid"),
			lineno: 3,
			column: 11,
		},
		{
			name:   "dos line endings",
			input:  []byte("cpu value=42\r\ncpu value=invalid"),
			lineno: 2,
			column: 11,
		},
		{
			name:   "mac line endings not supported",
			input:  []byte("cpu value=42\rcpu value=invalid"),
			lineno: 1,
			column: 14,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TestingHandler{}
			fsm := influx.NewMachine(handler)
			fsm.SetData(tt.input)

			// Parse until an error or eof
			for i := 0; i < 20; i++ {
				err := fsm.Next()
				if err != nil {
					break
				}
			}

			require.Equal(t, tt.lineno, fsm.LineNumber(), "lineno")
			require.Equal(t, tt.column, fsm.Column(), "column")
		})
	}
}

func BenchmarkMachine(b *testing.B) {
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			handler := &BenchmarkingHandler{}
			fsm := influx.NewMachine(handler)

			for n := 0; n < b.N; n++ {
				fsm.SetData(tt.input)

				for {
					err := fsm.Next()
					if err != nil {
						break
					}
				}
			}
		})
	}
}

func TestMachineProcstat(t *testing.T) {
	input := []byte("procstat,exe=bash,process_name=bash voluntary_context_switches=42i,memory_rss=5103616i,rlimit_memory_data_hard=2147483647i,cpu_time_user=0.02,rlimit_file_locks_soft=2147483647i,pid=29417i,cpu_time_nice=0,rlimit_memory_locked_soft=65536i,read_count=259i,rlimit_memory_vms_hard=2147483647i,memory_swap=0i,rlimit_num_fds_soft=1024i,rlimit_nice_priority_hard=0i,cpu_time_soft_irq=0,cpu_time=0i,rlimit_memory_locked_hard=65536i,realtime_priority=0i,signals_pending=0i,nice_priority=20i,cpu_time_idle=0,memory_stack=139264i,memory_locked=0i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,cpu_time_guest=0,cpu_time_guest_nice=0,rlimit_memory_data_soft=2147483647i,read_bytes=0i,rlimit_cpu_time_soft=2147483647i,involuntary_context_switches=2i,write_bytes=106496i,cpu_time_system=0,cpu_time_irq=0,cpu_usage=0,memory_vms=21659648i,memory_data=1576960i,rlimit_memory_stack_hard=2147483647i,num_threads=1i,cpu_time_stolen=0,rlimit_memory_rss_soft=2147483647i,rlimit_realtime_priority_soft=0i,num_fds=4i,write_count=35i,rlimit_signals_pending_soft=78994i,cpu_time_steal=0,rlimit_num_fds_hard=4096i,rlimit_file_locks_hard=2147483647i,rlimit_cpu_time_hard=2147483647i,rlimit_signals_pending_hard=78994i,rlimit_nice_priority_soft=0i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_realtime_priority_hard=0i 1517620624000000000")
	handler := &TestingHandler{}
	fsm := influx.NewMachine(handler)
	fsm.SetData(input)
	for {
		err := fsm.Next()
		if err != nil {
			break
		}
	}
}

func BenchmarkMachineProcstat(b *testing.B) {
	input := []byte("procstat,exe=bash,process_name=bash voluntary_context_switches=42i,memory_rss=5103616i,rlimit_memory_data_hard=2147483647i,cpu_time_user=0.02,rlimit_file_locks_soft=2147483647i,pid=29417i,cpu_time_nice=0,rlimit_memory_locked_soft=65536i,read_count=259i,rlimit_memory_vms_hard=2147483647i,memory_swap=0i,rlimit_num_fds_soft=1024i,rlimit_nice_priority_hard=0i,cpu_time_soft_irq=0,cpu_time=0i,rlimit_memory_locked_hard=65536i,realtime_priority=0i,signals_pending=0i,nice_priority=20i,cpu_time_idle=0,memory_stack=139264i,memory_locked=0i,rlimit_memory_stack_soft=8388608i,cpu_time_iowait=0,cpu_time_guest=0,cpu_time_guest_nice=0,rlimit_memory_data_soft=2147483647i,read_bytes=0i,rlimit_cpu_time_soft=2147483647i,involuntary_context_switches=2i,write_bytes=106496i,cpu_time_system=0,cpu_time_irq=0,cpu_usage=0,memory_vms=21659648i,memory_data=1576960i,rlimit_memory_stack_hard=2147483647i,num_threads=1i,cpu_time_stolen=0,rlimit_memory_rss_soft=2147483647i,rlimit_realtime_priority_soft=0i,num_fds=4i,write_count=35i,rlimit_signals_pending_soft=78994i,cpu_time_steal=0,rlimit_num_fds_hard=4096i,rlimit_file_locks_hard=2147483647i,rlimit_cpu_time_hard=2147483647i,rlimit_signals_pending_hard=78994i,rlimit_nice_priority_soft=0i,rlimit_memory_rss_hard=2147483647i,rlimit_memory_vms_soft=2147483647i,rlimit_realtime_priority_hard=0i 1517620624000000000")
	handler := &BenchmarkingHandler{}
	fsm := influx.NewMachine(handler)
	for n := 0; n < b.N; n++ {
		fsm.SetData(input)
		for {
			err := fsm.Next()
			if err != nil {
				break
			}
		}
	}
}

func TestSeriesMachine(t *testing.T) {
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
			name:  "no tags",
			input: []byte("cpu"),
			results: []Result{
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name: Success,
				},
			},
		},
		{
			name:  "tags",
			input: []byte("cpu,a=x,b=y"),
			results: []Result{
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name:  TagKey,
					Value: []byte("a"),
				},
				{
					Name:  TagValue,
					Value: []byte("x"),
				},
				{
					Name:  TagKey,
					Value: []byte("b"),
				},
				{
					Name:  TagValue,
					Value: []byte("y"),
				},
				{
					Name: Success,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TestingHandler{}
			fsm := influx.NewSeriesMachine(handler)
			fsm.SetData(tt.input)

			for {
				err := fsm.Next()
				if err != nil {
					break
				}
				handler.Result(err)
			}

			results := handler.Results()
			require.Equal(t, tt.results, results)
		})
	}
}

type MockHandler struct {
	SetMeasurementF func(name []byte) error
	AddTagF         func(key []byte, value []byte) error
	AddIntF         func(key []byte, value []byte) error
	AddUintF        func(key []byte, value []byte) error
	AddFloatF       func(key []byte, value []byte) error
	AddStringF      func(key []byte, value []byte) error
	AddBoolF        func(key []byte, value []byte) error
	SetTimestampF   func(tm []byte) error

	TestingHandler
}

func (h *MockHandler) SetMeasurement(name []byte) error {
	h.TestingHandler.SetMeasurement(name)
	return h.SetMeasurementF(name)
}

func (h *MockHandler) AddTag(name, value []byte) error {
	return h.AddTagF(name, value)
}

func (h *MockHandler) AddInt(name, value []byte) error {
	err := h.AddIntF(name, value)
	if err != nil {
		return err
	}
	h.TestingHandler.AddInt(name, value)
	return nil
}

func (h *MockHandler) AddUint(name, value []byte) error {
	err := h.AddUintF(name, value)
	if err != nil {
		return err
	}
	h.TestingHandler.AddUint(name, value)
	return nil
}

func (h *MockHandler) AddFloat(name, value []byte) error {
	return h.AddFloatF(name, value)
}

func (h *MockHandler) AddString(name, value []byte) error {
	return h.AddStringF(name, value)
}

func (h *MockHandler) AddBool(name, value []byte) error {
	return h.AddBoolF(name, value)
}

func (h *MockHandler) SetTimestamp(tm []byte) error {
	return h.SetTimestampF(tm)
}

func TestHandlerErrorRecovery(t *testing.T) {
	var tests = []struct {
		name    string
		input   []byte
		handler *MockHandler
		results []Result
	}{
		{
			name:  "integer",
			input: []byte("cpu value=43i\ncpu value=42i"),
			handler: &MockHandler{
				SetMeasurementF: func(name []byte) error {
					return nil
				},
				AddIntF: func(name, value []byte) error {
					if string(value) != "42i" {
						return errors.New("handler error")
					}
					return nil
				},
			},
			results: []Result{
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name: Error,
					err:  errors.New("handler error"),
				},
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name:  FieldKey,
					Value: []byte("value"),
				},
				{
					Name:  FieldInt,
					Value: []byte("42i"),
				},
				{
					Name: Success,
				},
			},
		},
		{
			name:  "integer with timestamp",
			input: []byte("cpu value=43i 1516241192000000000\ncpu value=42i"),
			handler: &MockHandler{
				SetMeasurementF: func(name []byte) error {
					return nil
				},
				AddIntF: func(name, value []byte) error {
					if string(value) != "42i" {
						return errors.New("handler error")
					}
					return nil
				},
			},
			results: []Result{
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name: Error,
					err:  errors.New("handler error"),
				},
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name:  FieldKey,
					Value: []byte("value"),
				},
				{
					Name:  FieldInt,
					Value: []byte("42i"),
				},
				{
					Name: Success,
				},
			},
		},
		{
			name:  "unsigned",
			input: []byte("cpu value=43u\ncpu value=42u"),
			handler: &MockHandler{
				SetMeasurementF: func(name []byte) error {
					return nil
				},
				AddUintF: func(name, value []byte) error {
					if string(value) != "42u" {
						return errors.New("handler error")
					}
					return nil
				},
			},
			results: []Result{
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name: Error,
					err:  errors.New("handler error"),
				},
				{
					Name:  Measurement,
					Value: []byte("cpu"),
				},
				{
					Name:  FieldKey,
					Value: []byte("value"),
				},
				{
					Name:  FieldUint,
					Value: []byte("42u"),
				},
				{
					Name: Success,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsm := influx.NewMachine(tt.handler)
			fsm.SetData(tt.input)

			for i := 0; i < 20; i++ {
				err := fsm.Next()
				if err != nil && err == influx.EOF {
					break
				}
				tt.handler.Result(err)
			}

			results := tt.handler.Results()
			require.Equal(t, tt.results, results)
		})
	}
}
