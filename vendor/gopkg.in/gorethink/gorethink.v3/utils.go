package gorethink

import (
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"

	"gopkg.in/gorethink/gorethink.v3/encoding"
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// Helper functions for constructing terms

// constructRootTerm is an alias for creating a new term.
func constructRootTerm(name string, termType p.Term_TermType, args []interface{}, optArgs map[string]interface{}) Term {
	return Term{
		name:     name,
		rootTerm: true,
		termType: termType,
		args:     convertTermList(args),
		optArgs:  convertTermObj(optArgs),
	}
}

// constructMethodTerm is an alias for creating a new term. Unlike constructRootTerm
// this function adds the previous expression in the tree to the argument list to
// create a method term.
func constructMethodTerm(prevVal Term, name string, termType p.Term_TermType, args []interface{}, optArgs map[string]interface{}) Term {
	args = append([]interface{}{prevVal}, args...)

	return Term{
		name:     name,
		rootTerm: false,
		termType: termType,
		args:     convertTermList(args),
		optArgs:  convertTermObj(optArgs),
	}
}

// Helper functions for creating internal RQL types

func newQuery(t Term, qopts map[string]interface{}, copts *ConnectOpts) (q Query, err error) {
	queryOpts := map[string]interface{}{}
	for k, v := range qopts {
		queryOpts[k], err = Expr(v).Build()
		if err != nil {
			return
		}
	}
	if copts.Database != "" {
		queryOpts["db"], err = DB(copts.Database).Build()
		if err != nil {
			return
		}
	}

	builtTerm, err := t.Build()
	if err != nil {
		return q, err
	}

	// Construct query
	return Query{
		Type:      p.Query_START,
		Term:      &t,
		Opts:      queryOpts,
		builtTerm: builtTerm,
	}, nil
}

// makeArray takes a slice of terms and produces a single MAKE_ARRAY term
func makeArray(args termsList) Term {
	return Term{
		name:     "[...]",
		termType: p.Term_MAKE_ARRAY,
		args:     args,
	}
}

// makeObject takes a map of terms and produces a single MAKE_OBJECT term
func makeObject(args termsObj) Term {
	return Term{
		name:     "{...}",
		termType: p.Term_MAKE_OBJ,
		optArgs:  args,
	}
}

var nextVarID int64

func makeFunc(f interface{}) Term {
	value := reflect.ValueOf(f)
	valueType := value.Type()

	var argNums = make([]interface{}, valueType.NumIn())
	var args = make([]reflect.Value, valueType.NumIn())
	for i := 0; i < valueType.NumIn(); i++ {
		// Get a slice of the VARs to use as the function arguments
		varID := atomic.AddInt64(&nextVarID, 1)
		args[i] = reflect.ValueOf(constructRootTerm("var", p.Term_VAR, []interface{}{varID}, map[string]interface{}{}))
		argNums[i] = varID

		// make sure all input arguments are of type Term
		argValueTypeName := valueType.In(i).String()
		if argValueTypeName != "gorethink.Term" && argValueTypeName != "interface {}" {
			panic("Function argument is not of type Term or interface {}")
		}
	}

	if valueType.NumOut() != 1 {
		panic("Function does not have a single return value")
	}

	body := value.Call(args)[0].Interface()
	argsArr := makeArray(convertTermList(argNums))

	return constructRootTerm("func", p.Term_FUNC, []interface{}{argsArr, body}, map[string]interface{}{})
}

func funcWrap(value interface{}) Term {
	val := Expr(value)

	if implVarScan(val) && val.termType != p.Term_ARGS {
		return makeFunc(func(x Term) Term {
			return val
		})
	}
	return val
}

func funcWrapArgs(args []interface{}) []interface{} {
	for i, arg := range args {
		args[i] = funcWrap(arg)
	}

	return args
}

// implVarScan recursivly checks a value to see if it contains an
// IMPLICIT_VAR term. If it does it returns true
func implVarScan(value Term) bool {
	if value.termType == p.Term_IMPLICIT_VAR {
		return true
	}
	for _, v := range value.args {
		if implVarScan(v) {
			return true
		}
	}

	for _, v := range value.optArgs {
		if implVarScan(v) {
			return true
		}
	}

	return false
}

// Convert an opt args struct to a map.
func optArgsToMap(optArgs OptArgs) map[string]interface{} {
	data, err := encode(optArgs)

	if err == nil && data != nil {
		if m, ok := data.(map[string]interface{}); ok {
			return m
		}
	}

	return map[string]interface{}{}
}

// Convert a list into a slice of terms
func convertTermList(l []interface{}) termsList {
	if len(l) == 0 {
		return nil
	}

	terms := make(termsList, len(l))
	for i, v := range l {
		terms[i] = Expr(v)
	}

	return terms
}

// Convert a map into a map of terms
func convertTermObj(o map[string]interface{}) termsObj {
	if len(o) == 0 {
		return nil
	}

	terms := make(termsObj, len(o))
	for k, v := range o {
		terms[k] = Expr(v)
	}

	return terms
}

// Helper functions for debugging

func allArgsToStringSlice(args termsList, optArgs termsObj) []string {
	allArgs := make([]string, len(args)+len(optArgs))
	i := 0

	for _, v := range args {
		allArgs[i] = v.String()
		i++
	}
	for k, v := range optArgs {
		allArgs[i] = k + "=" + v.String()
		i++
	}

	return allArgs
}

func argsToStringSlice(args termsList) []string {
	allArgs := make([]string, len(args))

	for i, v := range args {
		allArgs[i] = v.String()
	}

	return allArgs
}

func optArgsToStringSlice(optArgs termsObj) []string {
	allArgs := make([]string, len(optArgs))
	i := 0

	for k, v := range optArgs {
		allArgs[i] = k + "=" + v.String()
		i++
	}

	return allArgs
}

func splitAddress(address string) (hostname string, port int) {
	hostname = "localhost"
	port = 28015

	addrParts := strings.Split(address, ":")

	if len(addrParts) >= 1 {
		hostname = addrParts[0]
	}
	if len(addrParts) >= 2 {
		port, _ = strconv.Atoi(addrParts[1])
	}

	return
}

func encode(data interface{}) (interface{}, error) {
	if _, ok := data.(Term); ok {
		return data, nil
	}

	v, err := encoding.Encode(data)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// shouldRetryQuery checks the result of a query and returns true if the query
// should be retried
func shouldRetryQuery(q Query, err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(RQLConnectionError); ok {
		return true
	}

	return err == ErrConnectionClosed
}
