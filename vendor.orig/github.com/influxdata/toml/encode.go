package toml

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"time"

	"go/ast"

	"github.com/naoina/go-stringutil"
)

const (
	tagOmitempty = "omitempty"
	tagDocName   = "doc"
	tagSkip      = "-"
)

// Marshal returns the TOML encoding of v.
//
// Struct values encode as TOML. Each exported struct field becomes a field of
// the TOML structure unless
//   - the field's tag is "-", or
//   - the field is empty and its tag specifies the "omitempty" option.
// The "toml" key in the struct field's tag value is the key name, followed by
// an optional comma and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `toml:"-"`
//
//   // Field appears in TOML as key "myName".
//   Field int `toml:"myName"`
//
//   // Field appears in TOML as key "myName" and the field is omitted from the
//   // result of encoding if its value is empty.
//   Field int `toml:"myName,omitempty"`
//
//   // Field appears in TOML as key "field", but the field is skipped if
//   // empty.
//   // Note the leading comma.
//   Field int `toml:",omitempty"`
func Marshal(v interface{}) ([]byte, error) {
	return marshal(nil, "", reflect.ValueOf(v), false, false)
}

// A Encoder writes TOML to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

// Encode writes the TOML of v to the stream.
// See the documentation for Marshal for details about the conversion of Go values to TOML.
func (e *Encoder) Encode(v interface{}) error {
	b, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
}

// Marshaler is the interface implemented by objects that can marshal themselves into valid TOML.
type Marshaler interface {
	MarshalTOML() ([]byte, error)
}

func marshal(buf []byte, prefix string, rv reflect.Value, inArray, arrayTable bool) ([]byte, error) {
	rt := rv.Type()
	for rt.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rv.Type()
	}

	tableBuf := make([]byte, 0)
	valueBuf := make([]byte, 0)

	for i := 0; i < rv.NumField(); i++ {
		ft := rt.Field(i)
		if !ast.IsExported(ft.Name) {
			continue
		}
		colName, rest := extractTag(rt.Field(i).Tag.Get(fieldTagName))
		docStr := rt.Field(i).Tag.Get(tagDocName)

		if colName == tagSkip {
			continue
		}
		if colName == "" {
			colName = stringutil.ToSnakeCase(ft.Name)
		}
		fv := rv.Field(i)
		switch rest {
		case tagOmitempty:
			if fv.Interface() == reflect.Zero(ft.Type).Interface() {
				continue
			}
		}
		var err error
		switch fv.Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice:
			if tableBuf, err = encodeValue(tableBuf, prefix, colName, fv, inArray, arrayTable, docStr); err != nil {
				return nil, err
			}
		default:
			if valueBuf, err = encodeValue(valueBuf, prefix, colName, fv, inArray, arrayTable, docStr); err != nil {
				return nil, err
			}
		}
	}
	return append(append(buf, valueBuf...), tableBuf...), nil
}

func encodeValue(buf []byte, prefix, name string, fv reflect.Value, inArray, arrayTable bool, doc string) ([]byte, error) {
	switch t := fv.Interface().(type) {
	case Marshaler:
		b, err := t.MarshalTOML()
		if err != nil {
			return nil, err
		}
		return appendNewline(appendDocInline(append(appendKey(buf, name, inArray, arrayTable), b...), doc), inArray, arrayTable), nil
	case time.Time:
		return appendNewline(appendDocInline(encodeTime(appendKey(buf, name, inArray, arrayTable), t), doc), inArray, arrayTable), nil
	}
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return appendNewline(appendDocInline(encodeInt(appendKey(buf, name, inArray, arrayTable), fv.Int()), doc), inArray, arrayTable), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return appendNewline(appendDocInline(encodeUint(appendKey(buf, name, inArray, arrayTable), fv.Uint()), doc), inArray, arrayTable), nil
	case reflect.Float32, reflect.Float64:
		return appendNewline(appendDocInline(encodeFloat(appendKey(buf, name, inArray, arrayTable), fv.Float()), doc), inArray, arrayTable), nil
	case reflect.Bool:
		return appendNewline(appendDocInline(encodeBool(appendKey(buf, name, inArray, arrayTable), fv.Bool()), doc), inArray, arrayTable), nil
	case reflect.String:
		return appendNewline(appendDocInline(encodeString(appendKey(buf, name, inArray, arrayTable), fv.String()), doc), inArray, arrayTable), nil
	case reflect.Slice, reflect.Array:
		ft := fv.Type().Elem()
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			name := tableName(prefix, name)
			var err error
			for i := 0; i < fv.Len(); i++ {
				if buf, err = marshal(append(append(append(buf, '[', '['), name...), ']', ']', '\n'), name, fv.Index(i), false, true); err != nil {
					return nil, err
				}
			}
			return buf, nil
		}
		buf = append(appendKey(buf, name, inArray, arrayTable), '[')
		var err error
		for i := 0; i < fv.Len(); i++ {
			if i != 0 {
				buf = append(buf, ',')
			}
			if buf, err = encodeValue(buf, prefix, name, fv.Index(i), true, false, doc); err != nil {
				return nil, err
			}
		}
		return appendNewline(appendDocInline(append(buf, ']'), doc), inArray, arrayTable), nil
	case reflect.Struct:
		name := tableName(prefix, name)
		if doc != "" {
			buf = appendNewline(appendDoc(buf, doc), false, false)
		}
		return marshal(append(append(append(buf, '['), name...), ']', '\n'), name, fv, inArray, arrayTable)
	case reflect.Interface:
		var err error
		if buf, err = encodeInterface(appendKey(buf, name, inArray, arrayTable), fv.Interface()); err != nil {
			return nil, err
		}
		return appendNewline(buf, inArray, arrayTable), nil
	case reflect.Ptr:
		newElem := fv.Elem()
		if newElem.IsValid() {
			return encodeValue(buf, prefix, name, newElem, inArray, arrayTable, doc)
		} else {
			return encodeValue(buf, prefix, name, reflect.New(fv.Type().Elem()), inArray, arrayTable, doc)
		}
	case reflect.Map:
		name := tableName(prefix, name)
		buf := append(append(append(buf, '['), name...), ']', '\n')

		keys := fv.MapKeys()
		sortedKeys := make([]string, 0, len(keys))
		for _, key := range keys {
			var keyStr string
			switch key.Interface().(type) {
			case fmt.Stringer:
				keyStr = key.String()
			case string:
				keyStr = key.Interface().(string)
			}
			sortedKeys = append(sortedKeys, keyStr)
		}
		sort.Strings(sortedKeys)

		var err error
		for _, key := range sortedKeys {
			buf, err = encodeValue(buf, name, key, fv.MapIndex(reflect.ValueOf(key)), inArray, arrayTable, doc)
			if err != nil {
				return nil, err
			}
		}
		return buf, nil
	}
	return nil, fmt.Errorf("toml: marshal: unsupported type %v", fv.Kind())
}

func appendDocInline(buf []byte, doc string) []byte {
	if doc != "" {
		return append(append(append(append(buf, ' '), '#'), ' '), doc...)
	} else {
		return buf
	}
}

func appendDoc(buf []byte, doc string) []byte {
	return appendDocInline(buf, doc)[1:]
}

func appendKey(buf []byte, key string, inArray, arrayTable bool) []byte {
	if !inArray {
		return append(append(buf, key...), '=')
	}
	return buf
}

func appendNewline(buf []byte, inArray, arrayTable bool) []byte {
	if !inArray {
		return append(buf, '\n')
	}
	return buf
}

func encodeInterface(buf []byte, v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case int:
		return encodeInt(buf, int64(v)), nil
	case int8:
		return encodeInt(buf, int64(v)), nil
	case int16:
		return encodeInt(buf, int64(v)), nil
	case int32:
		return encodeInt(buf, int64(v)), nil
	case int64:
		return encodeInt(buf, v), nil
	case uint:
		return encodeUint(buf, uint64(v)), nil
	case uint8:
		return encodeUint(buf, uint64(v)), nil
	case uint16:
		return encodeUint(buf, uint64(v)), nil
	case uint32:
		return encodeUint(buf, uint64(v)), nil
	case uint64:
		return encodeUint(buf, v), nil
	case float32:
		return encodeFloat(buf, float64(v)), nil
	case float64:
		return encodeFloat(buf, v), nil
	case bool:
		return encodeBool(buf, v), nil
	case string:
		return encodeString(buf, v), nil
	}
	return nil, fmt.Errorf("toml: marshal: unable to detect a type of value `%v'", v)
}

func encodeInt(buf []byte, i int64) []byte {
	return strconv.AppendInt(buf, i, 10)
}

func encodeUint(buf []byte, u uint64) []byte {
	return strconv.AppendUint(buf, u, 10)
}

func encodeFloat(buf []byte, f float64) []byte {
	return strconv.AppendFloat(buf, f, 'e', -1, 64)
}

func encodeBool(buf []byte, b bool) []byte {
	return strconv.AppendBool(buf, b)
}

func encodeString(buf []byte, s string) []byte {
	return strconv.AppendQuote(buf, s)
}

func encodeTime(buf []byte, t time.Time) []byte {
	return append(buf, t.Format(time.RFC3339Nano)...)
}
