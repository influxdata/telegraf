package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

var (
	// MinVal represents the smallest possible value RethinkDB can store
	MinVal = constructRootTerm("MinVal", p.Term_MINVAL, []interface{}{}, map[string]interface{}{})
	// MaxVal represents the largest possible value RethinkDB can store
	MaxVal = constructRootTerm("MaxVal", p.Term_MAXVAL, []interface{}{}, map[string]interface{}{})
)

// Add sums two numbers or concatenates two arrays.
func (t Term) Add(args ...interface{}) Term {
	return constructMethodTerm(t, "Add", p.Term_ADD, args, map[string]interface{}{})
}

// Add sums two numbers or concatenates two arrays.
func Add(args ...interface{}) Term {
	return constructRootTerm("Add", p.Term_ADD, args, map[string]interface{}{})
}

// Sub subtracts two numbers.
func (t Term) Sub(args ...interface{}) Term {
	return constructMethodTerm(t, "Sub", p.Term_SUB, args, map[string]interface{}{})
}

// Sub subtracts two numbers.
func Sub(args ...interface{}) Term {
	return constructRootTerm("Sub", p.Term_SUB, args, map[string]interface{}{})
}

// Mul multiplies two numbers.
func (t Term) Mul(args ...interface{}) Term {
	return constructMethodTerm(t, "Mul", p.Term_MUL, args, map[string]interface{}{})
}

// Mul multiplies two numbers.
func Mul(args ...interface{}) Term {
	return constructRootTerm("Mul", p.Term_MUL, args, map[string]interface{}{})
}

// Div divides two numbers.
func (t Term) Div(args ...interface{}) Term {
	return constructMethodTerm(t, "Div", p.Term_DIV, args, map[string]interface{}{})
}

// Div divides two numbers.
func Div(args ...interface{}) Term {
	return constructRootTerm("Div", p.Term_DIV, args, map[string]interface{}{})
}

// Mod divides two numbers and returns the remainder.
func (t Term) Mod(args ...interface{}) Term {
	return constructMethodTerm(t, "Mod", p.Term_MOD, args, map[string]interface{}{})
}

// Mod divides two numbers and returns the remainder.
func Mod(args ...interface{}) Term {
	return constructRootTerm("Mod", p.Term_MOD, args, map[string]interface{}{})
}

// And performs a logical and on two values.
func (t Term) And(args ...interface{}) Term {
	return constructMethodTerm(t, "And", p.Term_AND, args, map[string]interface{}{})
}

// And performs a logical and on two values.
func And(args ...interface{}) Term {
	return constructRootTerm("And", p.Term_AND, args, map[string]interface{}{})
}

// Or performs a logical or on two values.
func (t Term) Or(args ...interface{}) Term {
	return constructMethodTerm(t, "Or", p.Term_OR, args, map[string]interface{}{})
}

// Or performs a logical or on two values.
func Or(args ...interface{}) Term {
	return constructRootTerm("Or", p.Term_OR, args, map[string]interface{}{})
}

// Eq returns true if two values are equal.
func (t Term) Eq(args ...interface{}) Term {
	return constructMethodTerm(t, "Eq", p.Term_EQ, args, map[string]interface{}{})
}

// Eq returns true if two values are equal.
func Eq(args ...interface{}) Term {
	return constructRootTerm("Eq", p.Term_EQ, args, map[string]interface{}{})
}

// Ne returns true if two values are not equal.
func (t Term) Ne(args ...interface{}) Term {
	return constructMethodTerm(t, "Ne", p.Term_NE, args, map[string]interface{}{})
}

// Ne returns true if two values are not equal.
func Ne(args ...interface{}) Term {
	return constructRootTerm("Ne", p.Term_NE, args, map[string]interface{}{})
}

// Gt returns true if the first value is greater than the second.
func (t Term) Gt(args ...interface{}) Term {
	return constructMethodTerm(t, "Gt", p.Term_GT, args, map[string]interface{}{})
}

// Gt returns true if the first value is greater than the second.
func Gt(args ...interface{}) Term {
	return constructRootTerm("Gt", p.Term_GT, args, map[string]interface{}{})
}

// Ge returns true if the first value is greater than or equal to the second.
func (t Term) Ge(args ...interface{}) Term {
	return constructMethodTerm(t, "Ge", p.Term_GE, args, map[string]interface{}{})
}

// Ge returns true if the first value is greater than or equal to the second.
func Ge(args ...interface{}) Term {
	return constructRootTerm("Ge", p.Term_GE, args, map[string]interface{}{})
}

// Lt returns true if the first value is less than the second.
func (t Term) Lt(args ...interface{}) Term {
	return constructMethodTerm(t, "Lt", p.Term_LT, args, map[string]interface{}{})
}

// Lt returns true if the first value is less than the second.
func Lt(args ...interface{}) Term {
	return constructRootTerm("Lt", p.Term_LT, args, map[string]interface{}{})
}

// Le returns true if the first value is less than or equal to the second.
func (t Term) Le(args ...interface{}) Term {
	return constructMethodTerm(t, "Le", p.Term_LE, args, map[string]interface{}{})
}

// Le returns true if the first value is less than or equal to the second.
func Le(args ...interface{}) Term {
	return constructRootTerm("Le", p.Term_LE, args, map[string]interface{}{})
}

// Not performs a logical not on a value.
func (t Term) Not(args ...interface{}) Term {
	return constructMethodTerm(t, "Not", p.Term_NOT, args, map[string]interface{}{})
}

// Not performs a logical not on a value.
func Not(args ...interface{}) Term {
	return constructRootTerm("Not", p.Term_NOT, args, map[string]interface{}{})
}

// RandomOpts contains the optional arguments for the Random term.
type RandomOpts struct {
	Float interface{} `gorethink:"float,omitempty"`
}

func (o RandomOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Random generates a random number between given (or implied) bounds. Random
// takes zero, one or two arguments.
//
// With zero arguments, the result will be a floating-point number in the range
// [0,1).
//
// With one argument x, the result will be in the range [0,x), and will be an
// integer unless the Float option is set to true. Specifying a floating point
// number without the Float option will raise an error.
//
// With two arguments x and y, the result will be in the range [x,y), and will
// be an integer unless the Float option is set to true. If x and y are equal an
// error will occur, unless the floating-point option has been specified, in
// which case x will be returned. Specifying a floating point number without the
// float option will raise an error.
//
// Note: Any integer responses can be be coerced to floating-points, when
// unmarshaling to a Go floating-point type. The last argument given will always
// be the ‘open’ side of the range,  but when generating a floating-point
// number, the ‘open’ side may be less than the ‘closed’ side.
func Random(args ...interface{}) Term {
	var opts = map[string]interface{}{}

	// Look for options map
	if len(args) > 0 {
		if possibleOpts, ok := args[len(args)-1].(RandomOpts); ok {
			opts = possibleOpts.toMap()
			args = args[:len(args)-1]
		}
	}

	return constructRootTerm("Random", p.Term_RANDOM, args, opts)
}

// Round causes the input number to be rounded the given value to the nearest whole integer.
func (t Term) Round(args ...interface{}) Term {
	return constructMethodTerm(t, "Round", p.Term_ROUND, args, map[string]interface{}{})
}

// Round causes the input number to be rounded the given value to the nearest whole integer.
func Round(args ...interface{}) Term {
	return constructRootTerm("Round", p.Term_ROUND, args, map[string]interface{}{})
}

// Ceil rounds the given value up, returning the smallest integer value greater
// than or equal to the given value (the value’s ceiling).
func (t Term) Ceil(args ...interface{}) Term {
	return constructMethodTerm(t, "Ceil", p.Term_CEIL, args, map[string]interface{}{})
}

// Ceil rounds the given value up, returning the smallest integer value greater
// than or equal to the given value (the value’s ceiling).
func Ceil(args ...interface{}) Term {
	return constructRootTerm("Ceil", p.Term_CEIL, args, map[string]interface{}{})
}

// Floor rounds the given value down, returning the largest integer value less
// than or equal to the given value (the value’s floor).
func (t Term) Floor(args ...interface{}) Term {
	return constructMethodTerm(t, "Floor", p.Term_FLOOR, args, map[string]interface{}{})
}

// Floor rounds the given value down, returning the largest integer value less
// than or equal to the given value (the value’s floor).
func Floor(args ...interface{}) Term {
	return constructRootTerm("Floor", p.Term_FLOOR, args, map[string]interface{}{})
}
