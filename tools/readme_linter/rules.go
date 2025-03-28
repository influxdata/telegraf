package main

import (
	"bufio"
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
)

var (
	// Setup regular expression for checking versions and valid choices
	// Matches HTML comments (e.g., <!-- some comment -->) surrounded by optional whitespace
	metaComment = regexp.MustCompile(`(?:\s*<!-- .* -->\s*)`)

	// Matches Telegraf versioning format (e.g., "Telegraf v1.2.3")
	metaVersion = regexp.MustCompile(`^Telegraf v\d+\.\d+\.\d+(?:\s+<!-- .* -->\s*)?$`)

	metaTags = map[plugin][]string{
		pluginInput: {
			"applications",
			"cloud",
			"containers",
			"datastore",
			"hardware",
			"iot",
			"logging",
			"messaging",
			"network",
			"server",
			"system",
			"testing",
			"web",
		},
		pluginOutput: {
			"applications",
			"cloud",
			"containers",
			"datastore",
			"hardware",
			"iot",
			"logging",
			"messaging",
			"network",
			"server",
			"system",
			"testing",
			"web",
		},
		pluginAggregator: {
			"math",
			"sampling",
			"statistics",
			"transformation",
		},
		pluginProcessor: {
			"math",
			"sampling",
			"statistics",
			"transformation",
		},
	}

	metaOSes = []string{
		"all",
		"freebsd",
		"linux",
		"macos",
		"solaris",
		"windows",
	}

	metaOrder = []string{
		"introduction version",
		"deprecation version",
		"removal version",
		"tags",
		"operating systems",
	}
)

// The first section is a heading with plugin name and paragraph short
// description
func firstSection(t *T, root ast.Node) error {
	var n ast.Node
	n = root.FirstChild()

	// Ignore HTML comments such as linter ignore sections
	for {
		if n == nil {
			break
		}
		if _, ok := n.(*ast.HTMLBlock); !ok {
			break
		}
		n = n.NextSibling()
	}

	t.assertKind(ast.KindHeading, n)
	t.assertHeadingLevel(1, n)
	t.assertFirstChildRegexp(` Plugin$`, n)

	// Make sure there is some text after the heading
	n = n.NextSibling()
	t.assertKind(ast.KindParagraph, n)
	length := len(n.(*ast.Paragraph).Lines().Value(t.markdown))
	if length < 30 {
		t.assertNodef(n, "short first section. Please add short description of plugin. length %d, minimum 30", length)
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

		child := h.FirstChild()
		if child == nil {
			continue
		}
		//nolint:staticcheck // need to use this since we aren't sure the type
		title := strings.TrimSpace(string(child.Text(t.markdown)))
		if headingsSet.has(title) && h.Level != expectedLevel {
			t.assertNodef(n, "has required section %q but wrong heading level. Expected level %d, found %d",
				title, expectedLevel, h.Level)
		}

		titleCounts[title]++
	}

	headingsSet.forEach(func(title string) {
		if _, exists := titleCounts[title]; !exists {
			t.assertf("missing required section %q", title)
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
		paraLines := make([]int, 0)
		for n := root.FirstChild(); n != nil; n = n.NextSibling() {
			var p *ast.Paragraph
			var ok bool
			if p, ok = n.(*ast.Paragraph); !ok {
				continue // only looking for paragraphs
			}

			segs := p.Lines()
			for _, seg := range segs.Sliced(0, segs.Len()) {
				line := t.line(seg.Start)
				paraLines = append(paraLines, line)
			}
		}

		// Find long lines in the whole file
		longLines := make([]int, 0, len(t.newlineOffsets))
		last := 0
		for i, cur := range t.newlineOffsets {
			length := cur - last - 1 // -1 to exclude the newline
			if length > threshold {
				longLines = append(longLines, i)
			}
			last = cur
		}

		// Merge both lists
		p := 0
		l := 0
		bads := make([]int, 0, max(len(paraLines), len(longLines)))
		for p < len(paraLines) && l < len(longLines) {
			long := longLines[l]
			para := paraLines[p]
			switch {
			case long == para:
				bads = append(bads, long)
				p++
				l++
			case long < para:
				l++
			case long > para:
				p++
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

		//nolint:staticcheck // need to use this since we aren't sure the type
		title := string(h.FirstChild().Text(t.markdown))
		if title == expectedTitle {
			config = h
			continue
		}
	}

	if config == nil {
		t.assertf("missing required section %q", expectedTitle)
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

// Links from one markdown file to another in the repo should be relative
func relativeTelegrafLinks(t *T, root ast.Node) error {
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		if _, ok := n.(*ast.Paragraph); !ok {
			continue
		}

		for n2 := n.FirstChild(); n2 != nil; n2 = n2.NextSibling() {
			var l *ast.Link
			var ok bool
			if l, ok = n2.(*ast.Link); !ok {
				continue
			}
			link := string(l.Destination)
			if strings.HasPrefix(link, "https://github.com/influxdata/telegraf/blob") {
				t.assertNodef(n, "in-repo link must be relative: %s", link)
			}
		}
	}
	return nil
}

// Each plugin should have metadata for documentation generation
func metadata(t *T, root ast.Node) error {
	const icons string = "⭐🚩🔥🏷️💻"

	n := root.FirstChild()
	if n == nil {
		t.assertf("no metadata section found")
		return nil
	}

	// Advance to the first heading which should be the plugin header
	for n != nil {
		if _, ok := n.(*ast.Heading); ok {
			t.assertHeadingLevel(1, n)
			break
		}
		n = n.NextSibling()
	}

	// Get the description text and check for metadata
	positions := make([]string, 0, 5)
	for n != nil {
		n = n.NextSibling()

		// The next heading will end the initial section
		if _, ok := n.(*ast.Heading); ok {
			break
		}

		// Ignore everything that is not text
		para, ok := n.(*ast.Paragraph)
		if !ok {
			continue
		}

		// Metadata should be separate paragraph with the items ordered.
		var inMetadata bool
		var counter int
		scanner := bufio.NewScanner(bytes.NewBuffer(para.Lines().Value(t.markdown)))
		for scanner.Scan() {
			txt := scanner.Text()
			if counter == 0 {
				inMetadata = strings.ContainsAny(txt, icons)
			}
			counter++

			// If we are not in a metadata section, we need to make sure we don't
			// see any metadata in this text.
			if !inMetadata {
				if strings.ContainsAny(txt, icons) {
					t.assertNodeLineOffsetf(n, counter-1, "metadata found in section not surrounded by empty lines")
					return nil
				}
				continue
			}

			icon, remainder, found := strings.Cut(txt, " ")
			if !found || !strings.Contains(icons, icon) {
				t.assertNodeLineOffsetf(n, counter-1, "metadata line must start with a valid icon and a space")
				continue
			}
			if strings.ContainsAny(remainder, icons) {
				t.assertNodeLineOffsetf(n, counter-1, "each metadata entry must be on a separate line")
				continue
			}

			// We are in a metadata section, so test for the correct structure
			switch icon {
			case "⭐":
				if !metaVersion.MatchString(remainder) {
					t.assertNodeLineOffsetf(n, counter-1, "invalid introduction version format; has to be 'Telegraf vX.Y.Z'")
				}
				positions = append(positions, "introduction version")
			case "🚩":
				if !metaVersion.MatchString(remainder) {
					t.assertNodeLineOffsetf(n, counter-1, "invalid deprecation version format; has to be 'Telegraf vX.Y.Z'")
				}
				positions = append(positions, "deprecation version")
			case "🔥":
				if !metaVersion.MatchString(remainder) {
					t.assertNodeLineOffsetf(n, counter-1, "invalid removal version format; has to be 'Telegraf vX.Y.Z'")
				}
				positions = append(positions, "removal version")
			case "🏷️":
				validTags, found := metaTags[t.pluginType]
				if !found {
					t.assertNodeLineOffsetf(n, counter-1, "no tags expected for plugin type")
					continue
				}

				for _, tag := range strings.Split(remainder, ",") {
					tag = metaComment.ReplaceAllString(tag, "")
					if !slices.Contains(validTags, strings.TrimSpace(tag)) {
						t.assertNodeLineOffsetf(n, counter-1, "unknown tag %q", tag)
					}
				}
				positions = append(positions, "tags")
			case "💻":
				for _, os := range strings.Split(remainder, ",") {
					os = metaComment.ReplaceAllString(os, "")
					if !slices.Contains(metaOSes, strings.TrimSpace(os)) {
						t.assertNodeLineOffsetf(n, counter-1, "unknown operating system %q", os)
					}
				}
				positions = append(positions, "operating systems")
			default:
				t.assertNodeLineOffsetf(n, counter-1, "invalid metadata icon")
				continue
			}
		}
	}

	if len(positions) == 0 {
		t.assertf("metadata is missing")
		return nil
	}

	// Check for duplicate entries
	seen := make(map[string]bool)
	for _, p := range positions {
		if seen[p] {
			t.assertNodef(n, "duplicate metadata entry for %q", p)
			return nil
		}
		seen[p] = true
	}

	// Remove the optional entries from the checklist
	validOrder := append(make([]string, 0, len(metaOrder)), metaOrder...)
	if !slices.Contains(positions, "deprecation version") && !slices.Contains(positions, "removal version") {
		idx := slices.Index(validOrder, "deprecation version")
		validOrder = slices.Delete(validOrder, idx, idx+1)
		idx = slices.Index(validOrder, "removal version")
		validOrder = slices.Delete(validOrder, idx, idx+1)
	}
	if _, found := metaTags[t.pluginType]; !found {
		idx := slices.Index(metaOrder, "tags")
		metaOrder = slices.Delete(metaOrder, idx, idx+1)
	}

	// Check the order of the metadata entries and required entries
	if len(validOrder) != len(positions) {
		for _, v := range validOrder {
			if !slices.Contains(positions, v) {
				t.assertNodef(n, "metadata entry for %q is missing", v)
			}
		}
		return nil
	}

	for i, v := range validOrder {
		if v != positions[i] {
			if i == 0 {
				t.assertNodef(n, "%q has to be the first entry", v)
			} else {
				t.assertNodef(n, "%q has to follow %q", v, validOrder[i-1])
			}
			return nil
		}
	}

	return nil
}

// To do: Check markdown files that aren't plugin readme files for paragraphs
// with long lines

// To do: Check the toml inside the configuration section for syntax errors
