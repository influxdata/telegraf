package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// Row returns the currently visited document. Note that Row does not work within
// subqueries to access nested documents; you should use anonymous functions to
// access those documents instead. Also note that unlike in other drivers to
// access a rows fields you should call Field. For example:
//   r.row("fieldname") should instead be r.Row.Field("fieldname")
var Row = constructRootTerm("Doc", p.Term_IMPLICIT_VAR, []interface{}{}, map[string]interface{}{})

// Literal replaces an object in a field instead of merging it with an existing
// object in a merge or update operation.
func Literal(args ...interface{}) Term {
	return constructRootTerm("Literal", p.Term_LITERAL, args, map[string]interface{}{})
}

// Field gets a single field from an object. If called on a sequence, gets that field
// from every object in the sequence, skipping objects that lack it.
func (t Term) Field(args ...interface{}) Term {
	return constructMethodTerm(t, "Field", p.Term_GET_FIELD, args, map[string]interface{}{})
}

// HasFields tests if an object has all of the specified fields. An object has a field if
// it has the specified key and that key maps to a non-null value. For instance,
//  the object `{'a':1,'b':2,'c':null}` has the fields `a` and `b`.
func (t Term) HasFields(args ...interface{}) Term {
	return constructMethodTerm(t, "HasFields", p.Term_HAS_FIELDS, args, map[string]interface{}{})
}

// Pluck plucks out one or more attributes from either an object or a sequence of
// objects (projection).
func (t Term) Pluck(args ...interface{}) Term {
	return constructMethodTerm(t, "Pluck", p.Term_PLUCK, args, map[string]interface{}{})
}

// Without is the opposite of pluck; takes an object or a sequence of objects, and returns
// them with the specified paths removed.
func (t Term) Without(args ...interface{}) Term {
	return constructMethodTerm(t, "Without", p.Term_WITHOUT, args, map[string]interface{}{})
}

// Merge merges two objects together to construct a new object with properties from both.
// Gives preference to attributes from other when there is a conflict.
func (t Term) Merge(args ...interface{}) Term {
	return constructMethodTerm(t, "Merge", p.Term_MERGE, funcWrapArgs(args), map[string]interface{}{})
}

// Append appends a value to an array.
func (t Term) Append(args ...interface{}) Term {
	return constructMethodTerm(t, "Append", p.Term_APPEND, args, map[string]interface{}{})
}

// Prepend prepends a value to an array.
func (t Term) Prepend(args ...interface{}) Term {
	return constructMethodTerm(t, "Prepend", p.Term_PREPEND, args, map[string]interface{}{})
}

// Difference removes the elements of one array from another array.
func (t Term) Difference(args ...interface{}) Term {
	return constructMethodTerm(t, "Difference", p.Term_DIFFERENCE, args, map[string]interface{}{})
}

// SetInsert adds a value to an array and return it as a set (an array with distinct values).
func (t Term) SetInsert(args ...interface{}) Term {
	return constructMethodTerm(t, "SetInsert", p.Term_SET_INSERT, args, map[string]interface{}{})
}

// SetUnion adds several values to an array and return it as a set (an array with
// distinct values).
func (t Term) SetUnion(args ...interface{}) Term {
	return constructMethodTerm(t, "SetUnion", p.Term_SET_UNION, args, map[string]interface{}{})
}

// SetIntersection calculates the intersection of two arrays returning values that
// occur in both of them as a set (an array with distinct values).
func (t Term) SetIntersection(args ...interface{}) Term {
	return constructMethodTerm(t, "SetIntersection", p.Term_SET_INTERSECTION, args, map[string]interface{}{})
}

// SetDifference removes the elements of one array from another and return them as a set (an
// array with distinct values).
func (t Term) SetDifference(args ...interface{}) Term {
	return constructMethodTerm(t, "SetDifference", p.Term_SET_DIFFERENCE, args, map[string]interface{}{})
}

// InsertAt inserts a value in to an array at a given index. Returns the modified array.
func (t Term) InsertAt(args ...interface{}) Term {
	return constructMethodTerm(t, "InsertAt", p.Term_INSERT_AT, args, map[string]interface{}{})
}

// SpliceAt inserts several values in to an array at a given index. Returns the modified array.
func (t Term) SpliceAt(args ...interface{}) Term {
	return constructMethodTerm(t, "SpliceAt", p.Term_SPLICE_AT, args, map[string]interface{}{})
}

// DeleteAt removes an element from an array at a given index. Returns the modified array.
func (t Term) DeleteAt(args ...interface{}) Term {
	return constructMethodTerm(t, "DeleteAt", p.Term_DELETE_AT, args, map[string]interface{}{})
}

// ChangeAt changes a value in an array at a given index. Returns the modified array.
func (t Term) ChangeAt(args ...interface{}) Term {
	return constructMethodTerm(t, "ChangeAt", p.Term_CHANGE_AT, args, map[string]interface{}{})
}

// Keys returns an array containing all of the object's keys.
func (t Term) Keys(args ...interface{}) Term {
	return constructMethodTerm(t, "Keys", p.Term_KEYS, args, map[string]interface{}{})
}

func (t Term) Values(args ...interface{}) Term {
	return constructMethodTerm(t, "Values", p.Term_VALUES, args, map[string]interface{}{})
}

// Object creates an object from a list of key-value pairs, where the keys must be strings.
func Object(args ...interface{}) Term {
	return constructRootTerm("Object", p.Term_OBJECT, args, map[string]interface{}{})
}
