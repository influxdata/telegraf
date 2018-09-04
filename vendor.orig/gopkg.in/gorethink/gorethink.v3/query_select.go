package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// DB references a database.
func DB(args ...interface{}) Term {
	return constructRootTerm("DB", p.Term_DB, args, map[string]interface{}{})
}

// TableOpts contains the optional arguments for the Table term
type TableOpts struct {
	ReadMode         interface{} `gorethink:"read_mode,omitempty"`
	UseOutdated      interface{} `gorethink:"use_outdated,omitempty"` // Deprecated
	IdentifierFormat interface{} `gorethink:"identifier_format,omitempty"`
}

func (o TableOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Table selects all documents in a table. This command can be chained with
// other commands to do further processing on the data.
//
// There are two optional arguments.
//   - useOutdated: if true, this allows potentially out-of-date data to be
//     returned, with potentially faster reads. It also allows you to perform reads
//     from a secondary replica if a primary has failed. Default false.
//   - identifierFormat: possible values are name and uuid, with a default of name.
//     If set to uuid, then system tables will refer to servers, databases and tables
//     by UUID rather than name. (This only has an effect when used with system tables.)
func Table(name interface{}, optArgs ...TableOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructRootTerm("Table", p.Term_TABLE, []interface{}{name}, opts)
}

// Table selects all documents in a table. This command can be chained with
// other commands to do further processing on the data.
//
// There are two optional arguments.
//   - useOutdated: if true, this allows potentially out-of-date data to be
//     returned, with potentially faster reads. It also allows you to perform reads
//     from a secondary replica if a primary has failed. Default false.
//   - identifierFormat: possible values are name and uuid, with a default of name.
//     If set to uuid, then system tables will refer to servers, databases and tables
//     by UUID rather than name. (This only has an effect when used with system tables.)
func (t Term) Table(name interface{}, optArgs ...TableOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Table", p.Term_TABLE, []interface{}{name}, opts)
}

// Get gets a document by primary key. If nothing was found, RethinkDB will return a nil value.
func (t Term) Get(args ...interface{}) Term {
	return constructMethodTerm(t, "Get", p.Term_GET, args, map[string]interface{}{})
}

// GetAllOpts contains the optional arguments for the GetAll term
type GetAllOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
}

func (o GetAllOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// GetAll gets all documents where the given value matches the value of the primary
// index. Multiple values can be passed this function if you want to select multiple
// documents. If the documents you are fetching have composite keys then each
// argument should be a slice. For more information see the examples.
func (t Term) GetAll(keys ...interface{}) Term {
	return constructMethodTerm(t, "GetAll", p.Term_GET_ALL, keys, map[string]interface{}{})
}

// GetAllByIndex gets all documents where the given value matches the value of
// the requested index.
func (t Term) GetAllByIndex(index interface{}, keys ...interface{}) Term {
	return constructMethodTerm(t, "GetAll", p.Term_GET_ALL, keys, map[string]interface{}{"index": index})
}

// BetweenOpts contains the optional arguments for the Between term
type BetweenOpts struct {
	Index      interface{} `gorethink:"index,omitempty"`
	LeftBound  interface{} `gorethink:"left_bound,omitempty"`
	RightBound interface{} `gorethink:"right_bound,omitempty"`
}

func (o BetweenOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Between gets all documents between two keys. Accepts three optional arguments:
// index, leftBound, and rightBound. If index is set to the name of a secondary
// index, between will return all documents where that index’s value is in the
// specified range (it uses the primary key by default). leftBound or rightBound
// may be set to open or closed to indicate whether or not to include that endpoint
// of the range (by default, leftBound is closed and rightBound is open).
//
// You may also use the special constants r.minval and r.maxval for boundaries,
// which represent “less than any index key” and “more than any index key”
// respectively. For instance, if you use r.minval as the lower key, then between
// will return all documents whose primary keys (or indexes) are less than the
// specified upper key.
func (t Term) Between(lowerKey, upperKey interface{}, optArgs ...BetweenOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Between", p.Term_BETWEEN, []interface{}{lowerKey, upperKey}, opts)
}

// FilterOpts contains the optional arguments for the Filter term
type FilterOpts struct {
	Default interface{} `gorethink:"default,omitempty"`
}

func (o FilterOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Filter gets all the documents for which the given predicate is true.
//
// Filter can be called on a sequence, selection, or a field containing an array
// of elements. The return type is the same as the type on which the function was
// called on. The body of every filter is wrapped in an implicit `.default(false)`,
// and the default value can be changed by passing the optional argument `default`.
// Setting this optional argument to `r.error()` will cause any non-existence
// errors to abort the filter.
func (t Term) Filter(f interface{}, optArgs ...FilterOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Filter", p.Term_FILTER, []interface{}{funcWrap(f)}, opts)
}
