package utils

// This is split out from the 'postgresql' package as its depended upon by both the 'postgresql' and
// 'postgresql/template' packages.

import (
	"sort"
	"strings"
)

// ColumnRole specifies the role of a column in a metric.
// It helps map the columns to the DB.
type ColumnRole int

const (
	TimeColType ColumnRole = iota + 1
	TagsIDColType
	TagColType
	FieldColType
)

type Column struct {
	Name string
	// the data type of each column should have in the db. used when checking
	// if the schema matches or it needs updates
	Type string
	// the role each column has, helps properly map the metric to the db
	Role ColumnRole
}

// ColumnList implements sort.Interface.
// Columns are sorted first into groups of time,tag_id,tags,fields, and then alphabetically within
// each group.
type ColumnList []Column

func (cl ColumnList) Len() int {
	return len(cl)
}

func (cl ColumnList) Less(i, j int) bool {
	if cl[i].Role != cl[j].Role {
		return cl[i].Role < cl[j].Role
	}
	return strings.ToLower(cl[i].Name) < strings.ToLower(cl[j].Name)
}

func (cl ColumnList) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}

func (cl ColumnList) Sort() {
	sort.Sort(cl)
}
