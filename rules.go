package main

import (
	"github.com/yuin/goldmark/ast"
)

// The first section is a heading with plugin name and paragraph short
// description
func mainHeading(t *T, root ast.Node) error {
	var n ast.Node
	n = root.FirstChild()

	t.assertKind(ast.KindHeading, n)
	t.assertHeadingLevel(1, n)
	t.assertFirstChildRegexp(` Plugin$`, n)

	return nil
}

// Somewhere there should be a heading "sample configuration" and a
// toml code block. The toml should match what is in the plugin's go
// code

// Second level headings should include
func requiredHeadings(t *T, root ast.Node, headings []string) error {
	headingsSet := newSet(headings)

	expectedLevel := 2

	titleCounts := make(map[string]int)

	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		var h *ast.Heading
		var ok bool
		if h, ok = n.(*ast.Heading); !ok {
			continue
		}

		title := string(h.FirstChild().Text(t.markdown))
		if headingsSet.has(title) && h.Level != expectedLevel {
			t.assertNodef(n, "has required section '%s' but wrong heading level. Expected level %d, found %d",
				title, expectedLevel, h.Level)
		}

		titleCounts[title] += 1
	}

	headingsSet.forEach(func(title string) {
		if _, exists := titleCounts[title]; !exists {
			t.assertf("missing required section '%s'", title)
		}
	})

	return nil
}

func requiredHeadingsClose(headings []string) func(*T, ast.Node) error {
	return func(t *T, root ast.Node) error {
		return requiredHeadings(t, root, headings)
	}
}
