package templating

import (
	"sort"
	"strings"
)

// node is an item in a sorted k-ary tree of filter parts. Each child is sorted by its part value.
// The special value of "*", is always sorted last.
type node struct {
	separator string
	value     string
	children  nodes
	template  *Template
}

// insert inserts the given string template into the tree.  The filter string is separated
// on the template separator and each part is used as the path in the tree.
func (n *node) insert(filter string, template *Template) {
	n.separator = template.separator
	n.recursiveInsert(strings.Split(filter, n.separator), template)
}

// recursiveInsert does the actual recursive insertion
func (n *node) recursiveInsert(values []string, template *Template) {
	// Add the end, set the template
	if len(values) == 0 {
		n.template = template
		return
	}

	// See if the the current element already exists in the tree. If so, insert the
	// into that sub-tree
	for _, v := range n.children {
		if v.value == values[0] {
			v.recursiveInsert(values[1:], template)
			return
		}
	}

	// New element, add it to the tree and sort the children
	newNode := &node{value: values[0]}
	n.children = append(n.children, newNode)
	sort.Sort(&n.children)

	// Now insert the rest of the tree into the new element
	newNode.recursiveInsert(values[1:], template)
}

// search searches for a template matching the input string
func (n *node) search(line string) *Template {
	separator := n.separator
	return n.recursiveSearch(strings.Split(line, separator))
}

// recursiveSearch performs the actual recursive search
func (n *node) recursiveSearch(lineParts []string) *Template {
	// Nothing to search
	if len(lineParts) == 0 || len(n.children) == 0 {
		return n.template
	}

	// If last element is a wildcard, don't include it in this search since it's sorted
	// to the end but lexicographically it would not always be and sort.Search assumes
	// the slice is sorted.
	length := len(n.children)
	if n.children[length-1].value == "*" {
		length--
	}

	// Find the index of child with an exact match
	i := sort.Search(length, func(i int) bool {
		return n.children[i].value >= lineParts[0]
	})

	// Found an exact match, so search that child sub-tree
	if i < len(n.children) && n.children[i].value == lineParts[0] {
		return n.children[i].recursiveSearch(lineParts[1:])
	}
	// Not an exact match, see if we have a wildcard child to search
	if n.children[len(n.children)-1].value == "*" {
		return n.children[len(n.children)-1].recursiveSearch(lineParts[1:])
	}
	return n.template
}

// nodes is simply an array of nodes implementing the sorting interface.
type nodes []*node

// Less returns a boolean indicating whether the filter at position j
// is less than the filter at position k. Filters are order by string
// comparison of each component parts.  A wildcard value "*" is never
// less than a non-wildcard value.
//
// For example, the filters:
//             "*.*"
//             "servers.*"
//             "servers.localhost"
//             "*.localhost"
//
// Would be sorted as:
//             "servers.localhost"
//             "servers.*"
//             "*.localhost"
//             "*.*"
func (n *nodes) Less(j, k int) bool {
	if (*n)[j].value == "*" && (*n)[k].value != "*" {
		return false
	}

	if (*n)[j].value != "*" && (*n)[k].value == "*" {
		return true
	}

	return (*n)[j].value < (*n)[k].value
}

// Swap swaps two elements of the array
func (n *nodes) Swap(i, j int) { (*n)[i], (*n)[j] = (*n)[j], (*n)[i] }

// Len returns the length of the array
func (n *nodes) Len() int { return len(*n) }
