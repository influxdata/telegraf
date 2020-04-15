package gorethink

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// A Query represents a query ready to be sent to the database, A Query differs
// from a Term as it contains both a query type and token. These values are used
// by the database to determine if the query is continuing a previous request
// and also allows the driver to identify the response as they can come out of
// order.
type Query struct {
	Type      p.Query_QueryType
	Token     int64
	Term      *Term
	Opts      map[string]interface{}
	builtTerm interface{}
}

func (q *Query) Build() []interface{} {
	res := []interface{}{int(q.Type)}
	if q.Term != nil {
		res = append(res, q.builtTerm)
	}

	if len(q.Opts) > 0 {
		// Clone opts and remove custom gorethink options
		opts := map[string]interface{}{}
		for k, v := range q.Opts {
			switch k {
			case "geometry_format":
			default:
				opts[k] = v
			}
		}

		res = append(res, opts)
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
	name           string
	rawQuery       bool
	rootTerm       bool
	termType       p.Term_TermType
	data           interface{}
	args           []Term
	optArgs        map[string]Term
	lastErr        error
	isMockAnything bool
}

func (t Term) compare(t2 Term, varMap map[int64]int64) bool {
	if t.isMockAnything || t2.isMockAnything {
		return true
	}

	if t.name != t2.name ||
		t.rawQuery != t2.rawQuery ||
		t.rootTerm != t2.rootTerm ||
		t.termType != t2.termType ||
		!reflect.DeepEqual(t.data, t2.data) ||
		len(t.args) != len(t2.args) ||
		len(t.optArgs) != len(t2.optArgs) {
		return false
	}

	for i, v := range t.args {
		if t.termType == p.Term_FUNC && t2.termType == p.Term_FUNC && i == 0 {
			// Functions need to be compared differently as each variable
			// will have a different var ID so first try to create a mapping
			// between the two sets of IDs
			argsArr := t.args[0].args
			argsArr2 := t2.args[0].args

			if len(argsArr) != len(argsArr2) {
				return false
			}

			for j := 0; j < len(argsArr); j++ {
				varMap[argsArr[j].data.(int64)] = argsArr2[j].data.(int64)
			}
		} else if t.termType == p.Term_VAR && t2.termType == p.Term_VAR && i == 0 {
			// When comparing vars use our var map
			v1 := t.args[i].data.(int64)
			v2 := t2.args[i].data.(int64)

			if varMap[v1] != v2 {
				return false
			}
		} else if !v.compare(t2.args[i], varMap) {
			return false
		}
	}

	for k, v := range t.optArgs {
		if _, ok := t2.optArgs[k]; !ok {
			return false
		}

		if !v.compare(t2.optArgs[k], varMap) {
			return false
		}
	}

	return true
}

// build takes the query tree and prepares it to be sent as a JSON
// expression
func (t Term) Build() (interface{}, error) {
	var err error

	if t.lastErr != nil {
		return nil, t.lastErr
	}

	if t.rawQuery {
		return t.data, nil
	}

	switch t.termType {
	case p.Term_DATUM:
		return t.data, nil
	case p.Term_MAKE_OBJ:
		res := map[string]interface{}{}
		for k, v := range t.optArgs {
			res[k], err = v.Build()
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case p.Term_BINARY:
		if len(t.args) == 0 {
			return map[string]interface{}{
				"$reql_type$": "BINARY",
				"data":        t.data,
			}, nil
		}
	}

	args := make([]interface{}, len(t.args))
	optArgs := make(map[string]interface{}, len(t.optArgs))

	for i, v := range t.args {
		arg, err := v.Build()
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}

	for k, v := range t.optArgs {
		optArgs[k], err = v.Build()
		if err != nil {
			return nil, err
		}
	}

	ret := []interface{}{int(t.termType)}

	if len(args) > 0 {
		ret = append(ret, args)
	}
	if len(optArgs) > 0 {
		ret = append(ret, optArgs)
	}

	return ret, nil
}

// String returns a string representation of the query tree
func (t Term) String() string {
	if t.isMockAnything {
		return "r.MockAnything()"
	}

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

	if t.args == nil {
		return "r"
	}

	return fmt.Sprintf("%s.%s(%s)", t.args[0].String(), t.name, strings.Join(allArgsToStringSlice(t.args[1:], t.optArgs), ", "))
}

// OptArgs is an interface used to represent a terms optional arguments. All
// optional argument types have a toMap function, the returned map can be encoded
// and sent as part of the query.
type OptArgs interface {
	toMap() map[string]interface{}
}

func (t Term) OptArgs(args interface{}) Term {
	switch args := args.(type) {
	case OptArgs:
		t.optArgs = convertTermObj(args.toMap())
	case map[string]interface{}:
		t.optArgs = convertTermObj(args)
	}

	return t
}

type QueryExecutor interface {
	IsConnected() bool
	Query(context.Context, Query) (*Cursor, error)
	Exec(context.Context, Query) error

	newQuery(t Term, opts map[string]interface{}) (Query, error)
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
	NewValue interface{} `gorethink:"new_val,omitempty"`
	OldValue interface{} `gorethink:"old_val,omitempty"`
	State    string      `gorethink:"state,omitempty"`
	Error    string      `gorethink:"error,omitempty"`
}

// RunOpts contains the optional arguments for the Run function.
type RunOpts struct {
	DB             interface{} `gorethink:"db,omitempty"`
	Db             interface{} `gorethink:"db,omitempty"` // Deprecated
	Profile        interface{} `gorethink:"profile,omitempty"`
	Durability     interface{} `gorethink:"durability,omitempty"`
	UseOutdated    interface{} `gorethink:"use_outdated,omitempty"` // Deprecated
	ArrayLimit     interface{} `gorethink:"array_limit,omitempty"`
	TimeFormat     interface{} `gorethink:"time_format,omitempty"`
	GroupFormat    interface{} `gorethink:"group_format,omitempty"`
	BinaryFormat   interface{} `gorethink:"binary_format,omitempty"`
	GeometryFormat interface{} `gorethink:"geometry_format,omitempty"`
	ReadMode       interface{} `gorethink:"read_mode,omitempty"`

	MinBatchRows              interface{} `gorethink:"min_batch_rows,omitempty"`
	MaxBatchRows              interface{} `gorethink:"max_batch_rows,omitempty"`
	MaxBatchBytes             interface{} `gorethink:"max_batch_bytes,omitempty"`
	MaxBatchSeconds           interface{} `gorethink:"max_batch_seconds,omitempty"`
	FirstBatchScaledownFactor interface{} `gorethink:"first_batch_scaledown_factor,omitempty"`

	Context context.Context `gorethink:"-"`
}

func (o RunOpts) toMap() map[string]interface{} {
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
func (t Term) Run(s QueryExecutor, optArgs ...RunOpts) (*Cursor, error) {
	opts := map[string]interface{}{}
	var ctx context.Context = nil // if it's nil connection will form context from connection opts
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
		ctx = optArgs[0].Context
	}

	if s == nil || !s.IsConnected() {
		return nil, ErrConnectionClosed
	}

	q, err := s.newQuery(t, opts)
	if err != nil {
		return nil, err
	}

	return s.Query(ctx, q)
}

// RunWrite runs a query using the given connection but unlike Run automatically
// scans the result into a variable of type WriteResponse. This function should be used
// if you are running a write query (such as Insert,  Update, TableCreate, etc...).
//
// If an error occurs when running the write query the first error is returned.
//
//	res, err := r.DB("database").Table("table").Insert(doc).RunWrite(sess)
func (t Term) RunWrite(s QueryExecutor, optArgs ...RunOpts) (WriteResponse, error) {
	var response WriteResponse

	res, err := t.Run(s, optArgs...)
	if err != nil {
		return response, err
	}
	defer res.Close()

	if err = res.One(&response); err != nil {
		return response, err
	}

	if response.Errors > 0 {
		return response, fmt.Errorf("%s", response.FirstError)
	}

	return response, nil
}

// ReadOne is a shortcut method that runs the query on the given connection
// and reads one response from the cursor before closing it.
//
// It returns any errors encountered from running the query or reading the response
func (t Term) ReadOne(dest interface{}, s QueryExecutor, optArgs ...RunOpts) error {
	res, err := t.Run(s, optArgs...)
	if err != nil {
		return err
	}
	return res.One(dest)
}

// ReadAll is a shortcut method that runs the query on the given connection
// and reads all of the responses from the cursor before closing it.
//
// It returns any errors encountered from running the query or reading the responses
func (t Term) ReadAll(dest interface{}, s QueryExecutor, optArgs ...RunOpts) error {
	res, err := t.Run(s, optArgs...)
	if err != nil {
		return err
	}
	return res.All(dest)
}

// ExecOpts contains the optional arguments for the Exec function and  inherits
// its options from RunOpts, the only difference is the addition of the NoReply
// field.
//
// When NoReply is true it causes the driver not to wait to receive the result
// and return immediately.
type ExecOpts struct {
	DB             interface{} `gorethink:"db,omitempty"`
	Db             interface{} `gorethink:"db,omitempty"` // Deprecated
	Profile        interface{} `gorethink:"profile,omitempty"`
	Durability     interface{} `gorethink:"durability,omitempty"`
	UseOutdated    interface{} `gorethink:"use_outdated,omitempty"` // Deprecated
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

	Context context.Context `gorethink:"-"`
}

func (o ExecOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Exec runs the query but does not return the result. Exec will still wait for
// the response to be received unless the NoReply field is true.
//
//	err := r.DB("database").Table("table").Insert(doc).Exec(sess, r.ExecOpts{
//		NoReply: true,
//	})
func (t Term) Exec(s QueryExecutor, optArgs ...ExecOpts) error {
	opts := map[string]interface{}{}
	var ctx context.Context = nil // if it's nil connection will form context from connection opts
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
		ctx = optArgs[0].Context
	}

	if s == nil || !s.IsConnected() {
		return ErrConnectionClosed
	}

	q, err := s.newQuery(t, opts)
	if err != nil {
		return err
	}

	return s.Exec(ctx, q)
}
