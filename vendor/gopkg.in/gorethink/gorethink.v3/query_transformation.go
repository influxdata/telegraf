package gorethink

import p "gopkg.in/gorethink/gorethink.v3/ql2"

// Map transform each element of the sequence by applying the given mapping
// function. It takes two arguments, a sequence and a function of type
// `func (r.Term) interface{}`.
//
// For example this query doubles each element in an array:
//
//     r.Map([]int{1,3,6}, func (row r.Term) interface{} {
//         return row.Mul(2)
//     })
func Map(args ...interface{}) Term {
	if len(args) > 0 {
		args = append(args[:len(args)-1], funcWrap(args[len(args)-1]))
	}

	return constructRootTerm("Map", p.Term_MAP, args, map[string]interface{}{})
}

// Map transforms each element of the sequence by applying the given mapping
// function. It takes one argument of type `func (r.Term) interface{}`.
//
// For example this query doubles each element in an array:
//
//     r.Expr([]int{1,3,6}).Map(func (row r.Term) interface{} {
//         return row.Mul(2)
//     })
func (t Term) Map(args ...interface{}) Term {
	if len(args) > 0 {
		args = append(args[:len(args)-1], funcWrap(args[len(args)-1]))
	}

	return constructMethodTerm(t, "Map", p.Term_MAP, args, map[string]interface{}{})
}

// WithFields takes a sequence of objects and a list of fields. If any objects in the
// sequence don't have all of the specified fields, they're dropped from the
// sequence. The remaining objects have the specified fields plucked out.
// (This is identical to `HasFields` followed by `Pluck` on a sequence.)
func (t Term) WithFields(args ...interface{}) Term {
	return constructMethodTerm(t, "WithFields", p.Term_WITH_FIELDS, args, map[string]interface{}{})
}

// ConcatMap concatenates one or more elements into a single sequence using a
// mapping function. ConcatMap works in a similar fashion to Map, applying the
// given function to each element in a sequence, but it will always return a
// single sequence.
func (t Term) ConcatMap(args ...interface{}) Term {
	return constructMethodTerm(t, "ConcatMap", p.Term_CONCAT_MAP, funcWrapArgs(args), map[string]interface{}{})
}

// OrderByOpts contains the optional arguments for the OrderBy term
type OrderByOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
}

func (o OrderByOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// OrderBy sorts the sequence by document values of the given key(s). To specify
// the ordering, wrap the attribute with either r.Asc or r.Desc (defaults to
// ascending).
//
// Sorting without an index requires the server to hold the sequence in memory,
// and is limited to 100,000 documents (or the setting of the ArrayLimit option
// for run). Sorting with an index can be done on arbitrarily large tables, or
// after a between command using the same index.
func (t Term) OrderBy(args ...interface{}) Term {
	var opts = map[string]interface{}{}

	// Look for options map
	if len(args) > 0 {
		if possibleOpts, ok := args[len(args)-1].(OrderByOpts); ok {
			opts = possibleOpts.toMap()
			args = args[:len(args)-1]
		}
	}

	for k, arg := range args {
		if t, ok := arg.(Term); !(ok && (t.termType == p.Term_DESC || t.termType == p.Term_ASC)) {
			args[k] = funcWrap(arg)
		}
	}

	return constructMethodTerm(t, "OrderBy", p.Term_ORDER_BY, args, opts)
}

// Desc is used by the OrderBy term to specify the ordering to be descending.
func Desc(args ...interface{}) Term {
	return constructRootTerm("Desc", p.Term_DESC, funcWrapArgs(args), map[string]interface{}{})
}

// Asc is used by the OrderBy term to specify that the ordering be ascending (the
// default).
func Asc(args ...interface{}) Term {
	return constructRootTerm("Asc", p.Term_ASC, funcWrapArgs(args), map[string]interface{}{})
}

// Skip skips a number of elements from the head of the sequence.
func (t Term) Skip(args ...interface{}) Term {
	return constructMethodTerm(t, "Skip", p.Term_SKIP, args, map[string]interface{}{})
}

// Limit ends the sequence after the given number of elements.
func (t Term) Limit(args ...interface{}) Term {
	return constructMethodTerm(t, "Limit", p.Term_LIMIT, args, map[string]interface{}{})
}

// SliceOpts contains the optional arguments for the Slice term
type SliceOpts struct {
	LeftBound  interface{} `gorethink:"left_bound,omitempty"`
	RightBound interface{} `gorethink:"right_bound,omitempty"`
}

func (o SliceOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Slice trims the sequence to within the bounds provided.
func (t Term) Slice(args ...interface{}) Term {
	var opts = map[string]interface{}{}

	// Look for options map
	if len(args) > 0 {
		if possibleOpts, ok := args[len(args)-1].(SliceOpts); ok {
			opts = possibleOpts.toMap()
			args = args[:len(args)-1]
		}
	}

	return constructMethodTerm(t, "Slice", p.Term_SLICE, args, opts)
}

// AtIndex gets a single field from an object or the nth element from a sequence.
func (t Term) AtIndex(args ...interface{}) Term {
	return constructMethodTerm(t, "AtIndex", p.Term_BRACKET, args, map[string]interface{}{})
}

// Nth gets the nth element from a sequence.
func (t Term) Nth(args ...interface{}) Term {
	return constructMethodTerm(t, "Nth", p.Term_NTH, args, map[string]interface{}{})
}

// OffsetsOf gets the indexes of an element in a sequence. If the argument is a
// predicate, get the indexes of all elements matching it.
func (t Term) OffsetsOf(args ...interface{}) Term {
	return constructMethodTerm(t, "OffsetsOf", p.Term_OFFSETS_OF, funcWrapArgs(args), map[string]interface{}{})
}

// IsEmpty tests if a sequence is empty.
func (t Term) IsEmpty(args ...interface{}) Term {
	return constructMethodTerm(t, "IsEmpty", p.Term_IS_EMPTY, args, map[string]interface{}{})
}

// UnionOpts contains the optional arguments for the Slice term
type UnionOpts struct {
	Interleave interface{} `gorethink:"interleave,omitempty"`
}

func (o UnionOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Union concatenates two sequences.
func Union(args ...interface{}) Term {
	return constructRootTerm("Union", p.Term_UNION, args, map[string]interface{}{})
}

// Union concatenates two sequences.
func (t Term) Union(args ...interface{}) Term {
	return constructMethodTerm(t, "Union", p.Term_UNION, args, map[string]interface{}{})
}

// UnionWithOpts like Union concatenates two sequences however allows for optional
// arguments to be passed.
func UnionWithOpts(optArgs UnionOpts, args ...interface{}) Term {
	return constructRootTerm("Union", p.Term_UNION, args, optArgs.toMap())
}

// UnionWithOpts like Union concatenates two sequences however allows for optional
// arguments to be passed.
func (t Term) UnionWithOpts(optArgs UnionOpts, args ...interface{}) Term {
	return constructMethodTerm(t, "Union", p.Term_UNION, args, optArgs.toMap())
}

// Sample selects a given number of elements from a sequence with uniform random
// distribution. Selection is done without replacement.
func (t Term) Sample(args ...interface{}) Term {
	return constructMethodTerm(t, "Sample", p.Term_SAMPLE, args, map[string]interface{}{})
}
