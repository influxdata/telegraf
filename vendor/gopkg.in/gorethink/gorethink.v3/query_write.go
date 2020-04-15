package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// InsertOpts contains the optional arguments for the Insert term
type InsertOpts struct {
	Durability    interface{} `gorethink:"durability,omitempty"`
	ReturnChanges interface{} `gorethink:"return_changes,omitempty"`
	Conflict      interface{} `gorethink:"conflict,omitempty"`
}

func (o InsertOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Insert documents into a table. Accepts a single document or an array
// of documents.
func (t Term) Insert(arg interface{}, optArgs ...InsertOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Insert", p.Term_INSERT, []interface{}{Expr(arg)}, opts)
}

// UpdateOpts contains the optional arguments for the Update term
type UpdateOpts struct {
	Durability    interface{} `gorethink:"durability,omitempty"`
	ReturnChanges interface{} `gorethink:"return_changes,omitempty"`
	NonAtomic     interface{} `gorethink:"non_atomic,omitempty"`
	Conflict      interface{} `gorethink:"conflict,omitempty"`
}

func (o UpdateOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Update JSON documents in a table. Accepts a JSON document, a ReQL expression,
// or a combination of the two. You can pass options like returnChanges that will
// return the old and new values of the row you have modified.
func (t Term) Update(arg interface{}, optArgs ...UpdateOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Update", p.Term_UPDATE, []interface{}{funcWrap(arg)}, opts)
}

// ReplaceOpts contains the optional arguments for the Replace term
type ReplaceOpts struct {
	Durability    interface{} `gorethink:"durability,omitempty"`
	ReturnChanges interface{} `gorethink:"return_changes,omitempty"`
	NonAtomic     interface{} `gorethink:"non_atomic,omitempty"`
}

func (o ReplaceOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Replace documents in a table. Accepts a JSON document or a ReQL expression,
// and replaces the original document with the new one. The new document must
// have the same primary key as the original document.
func (t Term) Replace(arg interface{}, optArgs ...ReplaceOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Replace", p.Term_REPLACE, []interface{}{funcWrap(arg)}, opts)
}

// DeleteOpts contains the optional arguments for the Delete term
type DeleteOpts struct {
	Durability    interface{} `gorethink:"durability,omitempty"`
	ReturnChanges interface{} `gorethink:"return_changes,omitempty"`
}

func (o DeleteOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Delete one or more documents from a table.
func (t Term) Delete(optArgs ...DeleteOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Delete", p.Term_DELETE, []interface{}{}, opts)
}

// Sync ensures that writes on a given table are written to permanent storage.
// Queries that specify soft durability do not give such guarantees, so Sync
// can be used to ensure the state of these queries. A call to Sync does not
// return until all previous writes to the table are persisted.
func (t Term) Sync(args ...interface{}) Term {
	return constructMethodTerm(t, "Sync", p.Term_SYNC, args, map[string]interface{}{})
}
