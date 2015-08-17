package gorethink

import (
	p "github.com/dancannon/gorethink/ql2"
)

// Aggregation
// These commands are used to compute smaller values from large sequences.

// Reduce produces a single value from a sequence through repeated application
// of a reduction function
//
// It takes one argument of type `func (r.Term, r.Term) interface{}`, for
// example this query sums all elements in an array:
//
//     r.Expr([]int{1,3,6}).Reduce(func (left, right r.Term) interface{} {
//         return left.Add(right)
//     })
func (t Term) Reduce(args ...interface{}) Term {
	return constructMethodTerm(t, "Reduce", p.Term_REDUCE, funcWrapArgs(args), map[string]interface{}{})
}

// DistinctOpts contains the optional arguments for the Distinct term
type DistinctOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
}

func (o *DistinctOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Distinct removes duplicate elements from the sequence.
func (t Term) Distinct(optArgs ...DistinctOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Distinct", p.Term_DISTINCT, []interface{}{}, opts)
}

// Group takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
func (t Term) Group(fieldOrFunctions ...interface{}) Term {
	return constructMethodTerm(t, "Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{})
}

// MultiGroup takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
//
// Unlike Group single documents can be assigned to multiple groups, similar
// to the behavior of multi-indexes. When the grouping value is an array, documents
// will be placed in each group that corresponds to the elements of the array. If
// the array is empty the row will be ignored.
func (t Term) MultiGroup(fieldOrFunctions ...interface{}) Term {
	return constructMethodTerm(t, "Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{
		"multi": true,
	})
}

// GroupByIndex takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
func (t Term) GroupByIndex(index interface{}, fieldOrFunctions ...interface{}) Term {
	return constructMethodTerm(t, "Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{
		"index": index,
	})
}

// MultiGroupByIndex takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
//
// Unlike Group single documents can be assigned to multiple groups, similar
// to the behavior of multi-indexes. When the grouping value is an array, documents
// will be placed in each group that corresponds to the elements of the array. If
// the array is empty the row will be ignored.
func (t Term) MultiGroupByIndex(index interface{}, fieldOrFunctions ...interface{}) Term {
	return constructMethodTerm(t, "Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{
		"index": index,
		"mutli": true,
	})
}

// Ungroup takes a grouped stream or grouped data and turns it into an array of
// objects representing the groups. Any commands chained after Ungroup will
// operate on this array, rather than operating on each group individually.
// This is useful if you want to e.g. order the groups by the value of their
// reduction.
func (t Term) Ungroup(args ...interface{}) Term {
	return constructMethodTerm(t, "Ungroup", p.Term_UNGROUP, args, map[string]interface{}{})
}

// Contains returns whether or not a sequence contains all the specified values,
// or if functions are provided instead, returns whether or not a sequence
// contains values matching all the specified functions.
func (t Term) Contains(args ...interface{}) Term {
	return constructMethodTerm(t, "Contains", p.Term_CONTAINS, args, map[string]interface{}{})
}

// Aggregators
// These standard aggregator objects are to be used in conjunction with Group.

// Count the number of elements in the sequence. With a single argument,
// count the number of elements equal to it. If the argument is a function,
// it is equivalent to calling filter before count.
func (t Term) Count(args ...interface{}) Term {
	return constructMethodTerm(t, "Count", p.Term_COUNT, funcWrapArgs(args), map[string]interface{}{})
}

// Sum returns the sum of all the elements of a sequence. If called with a field
// name, sums all the values of that field in the sequence, skipping elements of
// the sequence that lack that field. If called with a function, calls that
// function on every element of the sequence and sums the results, skipping
// elements of the sequence where that function returns null or a non-existence
// error.
func (t Term) Sum(args ...interface{}) Term {
	return constructMethodTerm(t, "Sum", p.Term_SUM, funcWrapArgs(args), map[string]interface{}{})
}

// Avg returns the average of all the elements of a sequence. If called with a field
// name, averages all the values of that field in the sequence, skipping elements of
// the sequence that lack that field. If called with a function, calls that function
// on every element of the sequence and averages the results, skipping elements of the
// sequence where that function returns null or a non-existence error.
func (t Term) Avg(args ...interface{}) Term {
	return constructMethodTerm(t, "Avg", p.Term_AVG, funcWrapArgs(args), map[string]interface{}{})
}

// Min finds the minimum of a sequence. If called with a field name, finds the element
// of that sequence with the smallest value in that field. If called with a function,
// calls that function on every element of the sequence and returns the element
// which produced the smallest value, ignoring any elements where the function
// returns null or produces a non-existence error.
func (t Term) Min(args ...interface{}) Term {
	return constructMethodTerm(t, "Min", p.Term_MIN, funcWrapArgs(args), map[string]interface{}{})
}

// MinIndex finds the minimum of a sequence. If called with a field name, finds the element
// of that sequence with the smallest value in that field. If called with a function,
// calls that function on every element of the sequence and returns the element
// which produced the smallest value, ignoring any elements where the function
// returns null or produces a non-existence error.
func (t Term) MinIndex(index interface{}, args ...interface{}) Term {
	return constructMethodTerm(t, "Min", p.Term_MIN, funcWrapArgs(args), map[string]interface{}{
		"index": index,
	})
}

// Max finds the maximum of a sequence. If called with a field name, finds the element
// of that sequence with the largest value in that field. If called with a function,
// calls that function on every element of the sequence and returns the element
// which produced the largest value, ignoring any elements where the function
// returns null or produces a non-existence error.
func (t Term) Max(args ...interface{}) Term {
	return constructMethodTerm(t, "Max", p.Term_MAX, funcWrapArgs(args), map[string]interface{}{})
}

// MaxIndex finds the maximum of a sequence. If called with a field name, finds the element
// of that sequence with the largest value in that field. If called with a function,
// calls that function on every element of the sequence and returns the element
// which produced the largest value, ignoring any elements where the function
// returns null or produces a non-existence error.
func (t Term) MaxIndex(index interface{}, args ...interface{}) Term {
	return constructMethodTerm(t, "Max", p.Term_MAX, funcWrapArgs(args), map[string]interface{}{
		"index": index,
	})
}
