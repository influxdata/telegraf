package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// InnerJoin returns the inner product of two sequences (e.g. a table, a filter result)
// filtered by the predicate. The query compares each row of the left sequence
// with each row of the right sequence to find all pairs of rows which satisfy
// the predicate. When the predicate is satisfied, each matched pair of rows
// of both sequences are combined into a result row.
func (t Term) InnerJoin(args ...interface{}) Term {
	return constructMethodTerm(t, "InnerJoin", p.Term_INNER_JOIN, args, map[string]interface{}{})
}

// OuterJoin computes a left outer join by retaining each row in the left table even
// if no match was found in the right table.
func (t Term) OuterJoin(args ...interface{}) Term {
	return constructMethodTerm(t, "OuterJoin", p.Term_OUTER_JOIN, args, map[string]interface{}{})
}

// EqJoinOpts contains the optional arguments for the EqJoin term.
type EqJoinOpts struct {
	Index   interface{} `gorethink:"index,omitempty"`
	Ordered interface{} `gorethink:"ordered,omitempty"`
}

func (o EqJoinOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// EqJoin is an efficient join that looks up elements in the right table by primary key.
//
// Optional arguments: "index" (string - name of the index to use in right table instead of the primary key)
func (t Term) EqJoin(left, right interface{}, optArgs ...EqJoinOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "EqJoin", p.Term_EQ_JOIN, []interface{}{funcWrap(left), right}, opts)
}

// Zip is used to 'zip' up the result of a join by merging the 'right' fields into 'left'
// fields of each member of the sequence.
func (t Term) Zip(args ...interface{}) Term {
	return constructMethodTerm(t, "Zip", p.Term_ZIP, args, map[string]interface{}{})
}
