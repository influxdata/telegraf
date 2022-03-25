package main

import (
	"github.com/yuin/goldmark/ast"
)

// The first section is a heading with plugin name and paragraph short
// description
func mainHeading(t *T, root ast.Node) error {
	// var children []ast.Node
	// var n ast.Node

	// var validTitle = regexp.MustCompile(`Plugin$`)

	// children = root.GetChildren()
	// t.assert(len(children) > 0)
	// n = children[0]
	var n ast.Node
	n = root.FirstChild()

	// t.assert(getNodeType(n) == "Heading")
	t.assertKind(ast.KindHeading, n)
	t.assertKind(ast.KindAutoLink, n)

	// children = n.GetChildren()
	// t.assert(len(children) > 0)
	// n = children[0]
	// t.assert(getNodeType(n) == "Text")
	// title := getContent(n)
	// t.assert(validTitle.MatchString(title))

	return nil
}

//somewhere there should be a heading "sample configuration" and
//a toml code block. the toml should match what is in the plugin's
//go code
