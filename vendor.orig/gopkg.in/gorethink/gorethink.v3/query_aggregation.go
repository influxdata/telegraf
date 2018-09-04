package gorethink

import p "gopkg.in/gorethink/gorethink.v3/ql2"

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

func (o DistinctOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Distinct removes duplicate elements from the sequence.
func Distinct(arg interface{}, optArgs ...DistinctOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructRootTerm("Distinct", p.Term_DISTINCT, []interface{}{arg}, opts)
}

// Distinct removes duplicate elements from the sequence.
func (t Term) Distinct(optArgs ...DistinctOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Distinct", p.Term_DISTINCT, []interface{}{}, opts)
}

// GroupOpts contains the optional arguments for the Group term
type GroupOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
	Multi interface{} `gorethink:"multi,omitempty"`
}

func (o GroupOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Group takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
func Group(fieldOrFunctions ...interface{}) Term {
	return constructRootTerm("Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{})
}

// MultiGroup takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
//
// Unlike Group single documents can be assigned to multiple groups, similar
// to the behavior of multi-indexes. When the grouping value is an array, documents
// will be placed in each group that corresponds to the elements of the array. If
// the array is empty the row will be ignored.
func MultiGroup(fieldOrFunctions ...interface{}) Term {
	return constructRootTerm("Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{
		"multi": true,
	})
}

// GroupByIndex takes a stream and partitions it into multiple groups based on the
// fields or functions provided. Commands chained after group will be
// called on each of these grouped sub-streams, producing grouped data.
func GroupByIndex(index interface{}, fieldOrFunctions ...interface{}) Term {
	return constructRootTerm("Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{
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
func MultiGroupByIndex(index interface{}, fieldOrFunctions ...interface{}) Term {
	return constructRootTerm("Group", p.Term_GROUP, funcWrapArgs(fieldOrFunctions), map[string]interface{}{
		"index": index,
		"mutli": true,
	})
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
func Contains(args ...interface{}) Term {
	return constructRootTerm("Contains", p.Term_CONTAINS, funcWrapArgs(args), map[string]interface{}{})
}

// Contains returns whether or not a sequence contains all the specified values,
// or if functions are provided instead, returns whether or not a sequence
// contains values matching all the specified functions.
func (t Term) Contains(args ...interface{}) Term {
	return constructMethodTerm(t, "Contains", p.Term_CONTAINS, funcWrapArgs(args), map[string]interface{}{})
}

// Aggregators
// These standard aggregator objects are to be used in conjunction with Group.

// Count the number of elements in the sequence. With a single argument,
// count the number of elements equal to it. If the argument is a function,
// it is equivalent to calling filter before count.
func Count(args ...interface{}) Term {
	return constructRootTerm("Count", p.Term_COUNT, funcWrapArgs(args), map[string]interface{}{})
}

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
func Sum(args ...interface{}) Term {
	return constructRootTerm("Sum", p.Term_SUM, funcWrapArgs(args), map[string]interface{}{})
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
func Avg(args ...interface{}) Term {
	return constructRootTerm("Avg", p.Term_AVG, funcWrapArgs(args), map[string]interface{}{})
}

// Avg returns the average of all the elements of a sequence. If called with a field
// name, averages all the values of that field in the sequence, skipping elements of
// the sequence that lack that field. If called with a function, calls that function
// on every element of the sequence and averages the results, skipping elements of the
// sequence where that function returns null or a non-existence error.
func (t Term) Avg(args ...interface{}) Term {
	return constructMethodTerm(t, "Avg", p.Term_AVG, funcWrapArgs(args), map[string]interface{}{})
}

// MinOpts contains the optional arguments for the Min term
type MinOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
}

func (o MinOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Min finds the minimum of a sequence. If called with a field name, finds the element
// of that sequence with the smallest value in that field. If called with a function,
// calls that function on every element of the sequence and returns the element
// which produced the smallest value, ignoring any elements where the function
// returns null or produces a non-existence error.
func Min(args ...interface{}) Term {
	return constructRootTerm("Min", p.Term_MIN, funcWrapArgs(args), map[string]interface{}{})
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
func MinIndex(index interface{}, args ...interface{}) Term {
	return constructRootTerm("Min", p.Term_MIN, funcWrapArgs(args), map[string]interface{}{
		"index": index,
	})
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

// MaxOpts contains the optional arguments for the Max term
type MaxOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
}

func (o MaxOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Max finds the maximum of a sequence. If called with a field name, finds the element
// of that sequence with the largest value in that field. If called with a function,
// calls that function on every element of the sequence and returns the element
// which produced the largest value, ignoring any elements where the function
// returns null or produces a non-existence error.
func Max(args ...interface{}) Term {
	return constructRootTerm("Max", p.Term_MAX, funcWrapArgs(args), map[string]interface{}{})
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
func MaxIndex(index interface{}, args ...interface{}) Term {
	return constructRootTerm("Max", p.Term_MAX, funcWrapArgs(args), map[string]interface{}{
		"index": index,
	})
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

// FoldOpts contains the optional arguments for the Fold term
type FoldOpts struct {
	Emit      interface{} `gorethink:"emit,omitempty"`
	FinalEmit interface{} `gorethink:"final_emit,omitempty"`
}

func (o FoldOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Fold applies a function to a sequence in order, maintaining state via an
// accumulator. The Fold command returns either a single value or a new sequence.
//
// In its first form, Fold operates like Reduce, returning a value by applying a
// combining function to each element in a sequence, passing the current element
// and the previous reduction result to the function. However, Fold has the
// following differences from Reduce:
//  - it is guaranteed to proceed through the sequence from first element to last.
//  - it passes an initial base value to the function with the first element in
//    place of the previous reduction result.
//
// In its second form, Fold operates like ConcatMap, returning a new sequence
// rather than a single value. When an emit function is provided, Fold will:
//  - proceed through the sequence in order and take an initial base value, as above.
//  - for each element in the sequence, call both the combining function and a
//    separate emitting function with the current element and previous reduction result.
//  - optionally pass the result of the combining function to the emitting function.
//
// If provided, the emitting function must return a list.
func (t Term) Fold(base, fn interface{}, optArgs ...FoldOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	args := []interface{}{base, funcWrap(fn)}

	return constructMethodTerm(t, "Fold", p.Term_FOLD, args, opts)
}
