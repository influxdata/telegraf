package main

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
)

// The first section is a heading with plugin name and paragraph short
// description
func firstSection(t *T, root ast.Node) error {
	var n ast.Node
	n = root.FirstChild()

	t.assertKind(ast.KindHeading, n)
	t.assertHeadingLevel(1, n)
	t.assertFirstChildRegexp(` Plugin$`, n)

	// Make sure there is some text after the heading
	n = n.NextSibling()
	t.assertKind(ast.KindParagraph, n)
	length := len(n.Text(t.markdown))
	min := 30
	if length < min {
		t.assertNodef(n, "short first section. Please add short description of plugin. length %d, minimum %d", length, min)
	}

	return nil
}

// Somewhere there should be a heading "sample configuration" and a
// toml code block. The toml should match what is in the plugin's go
// code

// Second level headings should include
func requiredSections(t *T, root ast.Node, headings []string) error {
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

// Use this to make a rule that looks for a list of settings. (this is
// a closure of func requiredSection)
func requiredSectionsClose(headings []string) func(*T, ast.Node) error {
	return func(t *T, root ast.Node) error {
		return requiredSections(t, root, headings)
	}
}

func noLongLinesInParagraphs(threshold int) func(*T, ast.Node) error {
	return func(t *T, root ast.Node) error {
		// We're looking for long lines in paragraphs. Find paragraphs
		// first, then which lines are in paragraphs
		paraLines := []int{}
		for n := root.FirstChild(); n != nil; n = n.NextSibling() {
			var p *ast.Paragraph
			var ok bool
			if p, ok = n.(*ast.Paragraph); !ok {
				continue //only looking for paragraphs
			}

			segs := p.Lines()
			for _, seg := range segs.Sliced(0, segs.Len()) {
				line := t.line(seg.Start)
				paraLines = append(paraLines, line)
				// t.printFileLine(line)
				// fmt.Printf("paragraph line\n")
			}
		}

		// Find long lines in the whole file
		longLines := []int{}
		last := 0
		for i, cur := range t.newlineOffsets {
			len := cur - last - 1 // -1 to exclude the newline
			if len > threshold {
				longLines = append(longLines, i)
				// t.printFileLine(i)
				// fmt.Printf("long line\n")
			}
			last = cur
		}

		// Merge both lists
		p := 0
		l := 0
		bads := []int{}
		for p < len(paraLines) && l < len(longLines) {
			long := longLines[l]
			para := paraLines[p]
			switch {
			case long == para:
				bads = append(bads, long)
				p += 1
				l += 1
			case long < para:
				l += 1
			case long > para:
				p += 1
			}
		}

		for _, bad := range bads {
			t.assertLinef(bad, "long line in paragraph")
		}
		return nil
	}
}

func configSection(t *T, root ast.Node) error {
	var config *ast.Heading
	config = nil
	expectedTitle := "Configuration"
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		var h *ast.Heading
		var ok bool
		if h, ok = n.(*ast.Heading); !ok {
			continue
		}

		title := string(h.FirstChild().Text(t.markdown))
		if title == expectedTitle {
			config = h
			continue
		}
	}

	if config == nil {
		t.assertf("missing section '%s'", expectedTitle)
		return nil
	}

	toml := config.NextSibling()
	if toml == nil {
		t.assertNodef(toml, "missing config next sibling")
		return nil
	}

	var b *ast.FencedCodeBlock
	var ok bool
	if b, ok = toml.(*ast.FencedCodeBlock); !ok {
		t.assertNodef(toml, "config next sibling isn't a fenced code block")
		return nil
	}

	if !bytes.Equal(b.Language(t.markdown), []byte("toml")) {
		t.assertNodef(b, "config fenced code block isn't toml language")
		return nil
	}

	return nil
}
