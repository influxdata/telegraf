package gorethink

import (
	"encoding/base64"
	"encoding/json"

	"reflect"

	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// Expr converts any value to an expression and is also used by many other terms
// such as Insert and Update. This function can convert the following basic Go
// types (bool, int, uint, string, float) and even pointers, maps and structs.
//
// When evaluating structs they are encoded into a map before being sent to the
// server. Each exported field is added to the map unless
//
//  - the field's tag is "-", or
//  - the field is empty and its tag specifies the "omitempty" option.
//
// Each fields default name in the map is the field name but can be specified
// in the struct field's tag value. The "gorethink" key in the struct field's
// tag value is the key name, followed by an optional comma and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `gorethink:"-"`
//   // Field appears as key "myName".
//   Field int `gorethink:"myName"`
//   // Field appears as key "myName" and
//   // the field is omitted from the object if its value is empty,
//   // as defined above.
//   Field int `gorethink:"myName,omitempty"`
//   // Field appears as key "Field" (the default), but
//   // the field is skipped if empty.
//   // Note the leading comma.
//   Field int `gorethink:",omitempty"`
func Expr(val interface{}) Term {
	if val == nil {
		return Term{
			termType: p.Term_DATUM,
			data:     nil,
		}
	}

	switch val := val.(type) {
	case Term:
		return val
	case []interface{}:
		vals := make([]Term, len(val))
		for i, v := range val {
			vals[i] = Expr(v)
		}

		return makeArray(vals)
	case map[string]interface{}:
		vals := make(map[string]Term, len(val))
		for k, v := range val {
			vals[k] = Expr(v)
		}

		return makeObject(vals)
	case
		bool,
		int,
		int8,
		int16,
		int32,
		int64,
		uint,
		uint8,
		uint16,
		uint32,
		uint64,
		float32,
		float64,
		uintptr,
		string,
		*bool,
		*int,
		*int8,
		*int16,
		*int32,
		*int64,
		*uint,
		*uint8,
		*uint16,
		*uint32,
		*uint64,
		*float32,
		*float64,
		*uintptr,
		*string:
		return Term{
			termType: p.Term_DATUM,
			data:     val,
		}
	default:
		// Use reflection to check for other types
		valType := reflect.TypeOf(val)
		valValue := reflect.ValueOf(val)

		switch valType.Kind() {
		case reflect.Func:
			return makeFunc(val)
		case reflect.Struct, reflect.Map, reflect.Ptr:
			data, err := encode(val)

			if err != nil || data == nil {
				return Term{
					termType: p.Term_DATUM,
					data:     nil,
					lastErr:  err,
				}
			}

			return Expr(data)

		case reflect.Slice, reflect.Array:
			// Check if slice is a byte slice
			if valType.Elem().Kind() == reflect.Uint8 {
				data, err := encode(val)

				if err != nil || data == nil {
					return Term{
						termType: p.Term_DATUM,
						data:     nil,
						lastErr:  err,
					}
				}

				return Expr(data)
			}

			vals := make([]Term, valValue.Len())
			for i := 0; i < valValue.Len(); i++ {
				vals[i] = Expr(valValue.Index(i).Interface())
			}

			return makeArray(vals)
		default:
			data, err := encode(val)

			if err != nil || data == nil {
				return Term{
					termType: p.Term_DATUM,
					data:     nil,
					lastErr:  err,
				}
			}

			return Term{
				termType: p.Term_DATUM,
				data:     data,
			}
		}
	}
}

// JSOpts contains the optional arguments for the JS term
type JSOpts struct {
	Timeout interface{} `gorethink:"timeout,omitempty"`
}

func (o JSOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// JS creates a JavaScript expression which is evaluated by the database when
// running the query.
func JS(jssrc interface{}, optArgs ...JSOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructRootTerm("Js", p.Term_JAVASCRIPT, []interface{}{jssrc}, opts)
}

// HTTPOpts contains the optional arguments for the HTTP term
type HTTPOpts struct {
	// General Options
	Timeout      interface{} `gorethink:"timeout,omitempty"`
	Reattempts   interface{} `gorethink:"reattempts,omitempty"`
	Redirects    interface{} `gorethink:"redirect,omitempty"`
	Verify       interface{} `gorethink:"verify,omitempty"`
	ResultFormat interface{} `gorethink:"result_format,omitempty"`

	// Request Options
	Method interface{} `gorethink:"method,omitempty"`
	Auth   interface{} `gorethink:"auth,omitempty"`
	Params interface{} `gorethink:"params,omitempty"`
	Header interface{} `gorethink:"header,omitempty"`
	Data   interface{} `gorethink:"data,omitempty"`

	// Pagination
	Page      interface{} `gorethink:"page,omitempty"`
	PageLimit interface{} `gorethink:"page_limit,omitempty"`
}

func (o HTTPOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// HTTP retrieves data from the specified URL over HTTP. The return type depends
// on the resultFormat option, which checks the Content-Type of the response by
// default.
func HTTP(url interface{}, optArgs ...HTTPOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructRootTerm("Http", p.Term_HTTP, []interface{}{url}, opts)
}

// JSON parses a JSON string on the server.
func JSON(args ...interface{}) Term {
	return constructRootTerm("Json", p.Term_JSON, args, map[string]interface{}{})
}

// Error throws a runtime error. If called with no arguments inside the second argument
// to `default`, re-throw the current error.
func Error(args ...interface{}) Term {
	return constructRootTerm("Error", p.Term_ERROR, args, map[string]interface{}{})
}

// Args is a special term usd to splice an array of arguments into another term.
// This is useful when you want to call a varadic term such as GetAll with a set
// of arguments provided at runtime.
func Args(args ...interface{}) Term {
	return constructRootTerm("Args", p.Term_ARGS, args, map[string]interface{}{})
}

// Binary encapsulates binary data within a query.
//
// The type of data binary accepts depends on the client language. In Go, it
// expects either a byte array/slice or a bytes.Buffer.
//
// Only a limited subset of ReQL commands may be chained after binary:
//  - coerceTo can coerce binary objects to string types
//  - count will return the number of bytes in the object
//  - slice will treat bytes like array indexes (i.e., slice(10,20) will return bytes 10–19)
//  - typeOf returns PTYPE<BINARY>
//  - info will return information on a binary object.
func Binary(data interface{}) Term {
	var b []byte

	switch data := data.(type) {
	case Term:
		return constructRootTerm("Binary", p.Term_BINARY, []interface{}{data}, map[string]interface{}{})
	case []byte:
		b = data
	default:
		typ := reflect.TypeOf(data)
		if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
			return Binary(reflect.ValueOf(data).Bytes())
		} else if typ.Kind() == reflect.Array && typ.Elem().Kind() == reflect.Uint8 {
			v := reflect.ValueOf(data)
			b = make([]byte, v.Len())
			for i := 0; i < v.Len(); i++ {
				b[i] = v.Index(i).Interface().(byte)
			}
			return Binary(b)
		}
		panic("Unsupported binary type")
	}

	return binaryTerm(base64.StdEncoding.EncodeToString(b))
}

func binaryTerm(data string) Term {
	t := constructRootTerm("Binary", p.Term_BINARY, []interface{}{}, map[string]interface{}{})
	t.data = data

	return t
}

// Do evaluates the expr in the context of one or more value bindings. The type of
// the result is the type of the value returned from expr.
func (t Term) Do(args ...interface{}) Term {
	newArgs := []interface{}{}
	newArgs = append(newArgs, funcWrap(args[len(args)-1]))
	newArgs = append(newArgs, t)
	newArgs = append(newArgs, args[:len(args)-1]...)

	return constructRootTerm("Do", p.Term_FUNCALL, newArgs, map[string]interface{}{})
}

// Do evaluates the expr in the context of one or more value bindings. The type of
// the result is the type of the value returned from expr.
func Do(args ...interface{}) Term {
	newArgs := []interface{}{}
	newArgs = append(newArgs, funcWrap(args[len(args)-1]))
	newArgs = append(newArgs, args[:len(args)-1]...)

	return constructRootTerm("Do", p.Term_FUNCALL, newArgs, map[string]interface{}{})
}

// Branch evaluates one of two control paths based on the value of an expression.
// branch is effectively an if renamed due to language constraints.
//
// The type of the result is determined by the type of the branch that gets executed.
func Branch(args ...interface{}) Term {
	return constructRootTerm("Branch", p.Term_BRANCH, args, map[string]interface{}{})
}

// Branch evaluates one of two control paths based on the value of an expression.
// branch is effectively an if renamed due to language constraints.
//
// The type of the result is determined by the type of the branch that gets executed.
func (t Term) Branch(args ...interface{}) Term {
	return constructMethodTerm(t, "Branch", p.Term_BRANCH, args, map[string]interface{}{})
}

// ForEach loops over a sequence, evaluating the given write query for each element.
//
// It takes one argument of type `func (r.Term) interface{}`, for
// example clones a table:
//
//     r.Table("table").ForEach(func (row r.Term) interface{} {
//         return r.Table("new_table").Insert(row)
//     })
func (t Term) ForEach(args ...interface{}) Term {
	return constructMethodTerm(t, "Foreach", p.Term_FOR_EACH, funcWrapArgs(args), map[string]interface{}{})
}

// Range generates a stream of sequential integers in a specified range. It
// accepts 0, 1, or 2 arguments, all of which should be numbers.
func Range(args ...interface{}) Term {
	return constructRootTerm("Range", p.Term_RANGE, args, map[string]interface{}{})
}

// Default handles non-existence errors. Tries to evaluate and return its first argument.
// If an error related to the absence of a value is thrown in the process, or if
// its first argument returns null, returns its second argument. (Alternatively,
// the second argument may be a function which will be called with either the
// text of the non-existence error or null.)
func (t Term) Default(args ...interface{}) Term {
	return constructMethodTerm(t, "Default", p.Term_DEFAULT, args, map[string]interface{}{})
}

// CoerceTo converts a value of one type into another.
//
// You can convert: a selection, sequence, or object into an ARRAY, an array of
// pairs into an OBJECT, and any DATUM into a STRING.
func (t Term) CoerceTo(args ...interface{}) Term {
	return constructMethodTerm(t, "CoerceTo", p.Term_COERCE_TO, args, map[string]interface{}{})
}

// TypeOf gets the type of a value.
func TypeOf(args ...interface{}) Term {
	return constructRootTerm("TypeOf", p.Term_TYPE_OF, args, map[string]interface{}{})
}

// TypeOf gets the type of a value.
func (t Term) TypeOf(args ...interface{}) Term {
	return constructMethodTerm(t, "TypeOf", p.Term_TYPE_OF, args, map[string]interface{}{})
}

// ToJSON converts a ReQL value or object to a JSON string.
func (t Term) ToJSON() Term {
	return constructMethodTerm(t, "ToJSON", p.Term_TO_JSON_STRING, []interface{}{}, map[string]interface{}{})
}

// Info gets information about a RQL value.
func (t Term) Info(args ...interface{}) Term {
	return constructMethodTerm(t, "Info", p.Term_INFO, args, map[string]interface{}{})
}

// UUID returns a UUID (universally unique identifier), a string that can be used
// as a unique ID. If a string is passed to uuid as an argument, the UUID will be
// deterministic, derived from the string’s SHA-1 hash.
func UUID(args ...interface{}) Term {
	return constructRootTerm("UUID", p.Term_UUID, args, map[string]interface{}{})
}

// RawQuery creates a new query from a JSON string, this bypasses any encoding
// done by GoRethink. The query should not contain the query type or any options
// as this should be handled using the normal driver API.
//
// THis query will only work if this is the only term in the query.
func RawQuery(q []byte) Term {
	data := json.RawMessage(q)
	return Term{
		name:     "RawQuery",
		rootTerm: true,
		rawQuery: true,
		data:     &data,
		args: []Term{
			Term{
				termType: p.Term_DATUM,
				data:     string(q),
			},
		},
	}
}
