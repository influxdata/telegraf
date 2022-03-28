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

//somewhere there should be a heading "sample configuration" and
//a toml code block. the toml should match what is in the plugin's
//go code

var titles *set

func init() {
	titles = newSet()
	titles.add("Example Output")
	titles.add("Sample Config")
	titles.add("Metric Format")
}

// Second level headings should include
func requiredSections(t *T, root ast.Node) error {

	expectedLevel := 2

	titleCounts := make(map[string]int)

	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		var h *ast.Heading
		var ok bool
		if h, ok = n.(*ast.Heading); !ok {
			continue
		}

		title := string(h.FirstChild().Text(t.markdown))
		if titles.has(title) && h.Level != expectedLevel {
			t.assertNodef(n, "has required section '%s' but wrong heading level. Expected level %d, found %d",
				title, expectedLevel, h.Level)
		}

		titleCounts[title] += 1
	}

	titles.forEach(func(title string) bool {
		if _, exists := titleCounts[title]; !exists {
			t.assertf("missing required section '%s'", title)
		}
		return true
	})

	return nil
}

type set struct {
	m map[string]struct{}
}

func (s *set) empty() bool {
	return len(s.m) == 0
}

func (s *set) add(key string) {
	s.m[key] = struct{}{}
}

func (s *set) has(key string) bool {
	var ok bool
	_, ok = s.m[key]
	return ok
}

func (s *set) forEach(f func(string) bool) {
	for key := range s.m {
		if !f(key) {
			return
		}
	}
}

func newSet() *set {
	s := &set{
		m: make(map[string]struct{}),
	}
	return s
}
