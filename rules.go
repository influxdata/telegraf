package main

import (
	"github.com/yuin/goldmark/ast"
)

// The first section is a heading with plugin name and paragraph short
// description
func mainHeading(t *T, root ast.Node) error {
	// var children []ast.Node
	// var n ast.Node

	// children = root.GetChildren()
	// t.assert(len(children) > 0)
	// n = children[0]
	var n ast.Node
	n = root.FirstChild()

	t.assertKind(ast.KindHeading, n)
	t.assertFirstChildRegexp(` Plugin$`, n)

	return nil
}

//somewhere there should be a heading "sample configuration" and
//a toml code block. the toml should match what is in the plugin's
//go code
