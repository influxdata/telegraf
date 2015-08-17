package gorethink

import (
	"fmt"
	"strconv"
	"strings"

	p "github.com/dancannon/gorethink/ql2"
)

// A Query represents a query ready to be sent to the database, A Query differs
// from a Term as it contains both a query type and token. These values are used
// by the database to determine if the query is continuing a previous request
// and also allows the driver to identify the response as they can come out of
// order.
type Query struct {
	Type  p.Query_QueryType
	Token int64
	Term  *Term
	Opts  map[string]interface{}
}

func (q *Query) build() []interface{} {
	res := []interface{}{int(q.Type)}
	if q.Term != nil {
		res = append(res, q.Term.build())
	}

	if len(q.Opts) > 0 {
		res = append(res, q.Opts)
	}

	return res
}

type termsList []Term
type termsObj map[string]Term

// A Term represents a query that is being built. Terms consist of a an array of
// "sub-terms" and a term type. When a Term is a sub-term the first element of
// the terms data is its parent Term.
//
// When built the term becomes a JSON array, for more information on the format
// see http://rethinkdb.com/docs/writing-drivers/.
type Term struct {
	name     string
	rootTerm bool
	termType p.Term_TermType
	data     interface{}
	args     []Term
	optArgs  map[string]Term
}

// build takes the query tree and prepares it to be sent as a JSON
// expression
func (t Term) build() interface{} {
	switch t.termType {
	case p.Term_DATUM:
		return t.data
	case p.Term_MAKE_OBJ:
		res := map[string]interface{}{}
		for k, v := range t.optArgs {
			res[k] = v.build()
		}
		return res
	case p.Term_BINARY:
		if len(t.args) == 0 {
			return map[string]interface{}{
				"$reql_type$": "BINARY",
				"data":        t.data,
			}
		}
	}

	args := []interface{}{}
	optArgs := map[string]interface{}{}

	for _, v := range t.args {
		args = append(args, v.build())
	}

	for k, v := range t.optArgs {
		optArgs[k] = v.build()
	}

	ret := []interface{}{int(t.termType)}

	if len(args) > 0 {
		ret = append(ret, args)
	}
	if len(optArgs) > 0 {
		ret = append(ret, optArgs)
	}

	return ret
}

// String returns a string representation of the query tree
func (t Term) String() string {
	switch t.termType {
	case p.Term_MAKE_ARRAY:
		return fmt.Sprintf("[%s]", strings.Join(argsToStringSlice(t.args), ", "))
	case p.Term_MAKE_OBJ:
		return fmt.Sprintf("{%s}", strings.Join(optArgsToStringSlice(t.optArgs), ", "))
	case p.Term_FUNC:
		// Get string representation of each argument
		args := []string{}
		for _, v := range t.args[0].args {
			args = append(args, fmt.Sprintf("var_%d", v.data))
		}

		return fmt.Sprintf("func(%s r.Term) r.Term { return %s }",
			strings.Join(args, ", "),
			t.args[1].String(),
		)
	case p.Term_VAR:
		return fmt.Sprintf("var_%s", t.args[0])
	case p.Term_IMPLICIT_VAR:
		return "r.Row"
	case p.Term_DATUM:
		switch v := t.data.(type) {
		case string:
			return strconv.Quote(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	case p.Term_BINARY:
		if len(t.args) == 0 {
			return fmt.Sprintf("r.binary(<data>)")
		}
	}

	if t.rootTerm {
		return fmt.Sprintf("r.%s(%s)", t.name, strings.Join(allArgsToStringSlice(t.args, t.optArgs), ", "))
	}
	return fmt.Sprintf("%s.%s(%s)", t.args[0].String(), t.name, strings.Join(allArgsToStringSlice(t.args[1:], t.optArgs), ", "))
}

// OptArgs is an interface used to represent a terms optional arguments. All
// optional argument types have a toMap function, the returned map can be encoded
// and sent as part of the query.
type OptArgs interface {
	toMap() map[string]interface{}
}

// WriteResponse is a helper type used when dealing with the response of a
// write query. It is also returned by the RunWrite function.
type WriteResponse struct {
	Errors        int              `gorethink:"errors"`
	Inserted      int              `gorethink:"inserted"`
	Updated       int              `gorethink:"updated"`
	Unchanged     int              `gorethink:"unchanged"`
	Replaced      int              `gorethink:"replaced"`
	Renamed       int              `gorethink:"renamed"`
	Skipped       int              `gorethink:"skipped"`
	Deleted       int              `gorethink:"deleted"`
	Created       int              `gorethink:"created"`
	DBsCreated    int              `gorethink:"dbs_created"`
	TablesCreated int              `gorethink:"tables_created"`
	Dropped       int              `gorethink:"dropped"`
	DBsDropped    int              `gorethink:"dbs_dropped"`
	TablesDropped int              `gorethink:"tables_dropped"`
	GeneratedKeys []string         `gorethink:"generated_keys"`
	FirstError    string           `gorethink:"first_error"` // populated if Errors > 0
	ConfigChanges []ChangeResponse `gorethink:"config_changes"`
	Changes       []ChangeResponse
}

// ChangeResponse is a helper type used when dealing with changefeeds. The type
// contains both the value before the query and the new value.
type ChangeResponse struct {
	NewValue interface{} `gorethink:"new_val"`
	OldValue interface{} `gorethink:"old_val"`
}

// RunOpts contains the optional arguments for the Run function.
type RunOpts struct {
	Db             interface{} `gorethink:"db,omitempty"`
	Profile        interface{} `gorethink:"profile,omitempty"`
	UseOutdated    interface{} `gorethink:"use_outdated,omitempty"`
	ArrayLimit     interface{} `gorethink:"array_limit,omitempty"`
	TimeFormat     interface{} `gorethink:"time_format,omitempty"`
	GroupFormat    interface{} `gorethink:"group_format,omitempty"`
	BinaryFormat   interface{} `gorethink:"binary_format,omitempty"`
	GeometryFormat interface{} `gorethink:"geometry_format,omitempty"`

	MinBatchRows              interface{} `gorethink:"min_batch_rows,omitempty"`
	MaxBatchRows              interface{} `gorethink:"max_batch_rows,omitempty"`
	MaxBatchBytes             interface{} `gorethink:"max_batch_bytes,omitempty"`
	MaxBatchSeconds           interface{} `gorethink:"max_batch_seconds,omitempty"`
	FirstBatchScaledownFactor interface{} `gorethink:"first_batch_scaledown_factor,omitempty"`
}

func (o *RunOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Run runs a query using the given connection.
//
//	rows, err := query.Run(sess)
//	if err != nil {
//		// error
//	}
//
//  var doc MyDocumentType
//	for rows.Next(&doc) {
//      // Do something with document
//	}
func (t Term) Run(s *Session, optArgs ...RunOpts) (*Cursor, error) {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return s.Query(s.newQuery(t, opts))
}

// RunWrite runs a query using the given connection but unlike Run automatically
// scans the result into a variable of type WriteResponss. This function should be used
// if you are running a write query (such as Insert,  Update, TableCreate, etc...).
//
// If an error occurs when running the write query the first error is returned.
//
//	res, err := r.Db("database").Table("table").Insert(doc).RunWrite(sess)
func (t Term) RunWrite(s *Session, optArgs ...RunOpts) (WriteResponse, error) {
	var response WriteResponse

	res, err := t.Run(s, optArgs...)
	if err != nil {
		return response, err
	}

	if err = res.One(&response); err != nil {
		return response, err
	}

	if err = res.Close(); err != nil {
		return response, err
	}

	if response.Errors > 0 {
		return response, fmt.Errorf(response.FirstError)
	}

	return response, nil
}

// ExecOpts contains the optional arguments for the Exec function and  inherits
// its options from RunOpts, the only difference is the addition of the NoReply
// field.
//
// When NoReply is true it causes the driver not to wait to receive the result
// and return immediately.
type ExecOpts struct {
	Db             interface{} `gorethink:"db,omitempty"`
	Profile        interface{} `gorethink:"profile,omitempty"`
	UseOutdated    interface{} `gorethink:"use_outdated,omitempty"`
	ArrayLimit     interface{} `gorethink:"array_limit,omitempty"`
	TimeFormat     interface{} `gorethink:"time_format,omitempty"`
	GroupFormat    interface{} `gorethink:"group_format,omitempty"`
	BinaryFormat   interface{} `gorethink:"binary_format,omitempty"`
	GeometryFormat interface{} `gorethink:"geometry_format,omitempty"`

	MinBatchRows              interface{} `gorethink:"min_batch_rows,omitempty"`
	MaxBatchRows              interface{} `gorethink:"max_batch_rows,omitempty"`
	MaxBatchBytes             interface{} `gorethink:"max_batch_bytes,omitempty"`
	MaxBatchSeconds           interface{} `gorethink:"max_batch_seconds,omitempty"`
	FirstBatchScaledownFactor interface{} `gorethink:"first_batch_scaledown_factor,omitempty"`

	NoReply interface{} `gorethink:"noreply,omitempty"`
}

func (o *ExecOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Exec runs the query but does not return the result. Exec will still wait for
// the response to be received unless the NoReply field is true.
//
//	err := r.Db("database").Table("table").Insert(doc).Exec(sess, r.ExecOpts{
//		NoReply: true,
//	})
func (t Term) Exec(s *Session, optArgs ...ExecOpts) error {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return s.Exec(s.newQuery(t, opts))
}
