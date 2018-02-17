package templating

import (
	"sort"
	"strings"
)

const (
	// DefaultSeparator is the default separation character to use when separating template parts.
	DefaultSeparator = "."
)

// Engine uses a Matcher to retrieve the appropriate template and applies the template
// to the input string
type Engine struct {
	joiner  string
	matcher *matcher
}

// Apply extracts the template fields from the given line and returns the measurement
// name, tags and field name
func (e *Engine) Apply(line string) (string, map[string]string, string, error) {
	return e.matcher.match(line).Apply(line, e.joiner)
}

// NewEngine creates a new templating engine
func NewEngine(joiner string, defaultTemplate *Template, templates []string) (*Engine, error) {
	engine := Engine{
		joiner:  joiner,
		matcher: newMatcher(defaultTemplate),
	}
	templateSpecs := parseTemplateSpecs(templates)

	for _, templateSpec := range templateSpecs {
		if err := engine.matcher.addSpec(templateSpec); err != nil {
			return nil, err
		}
	}

	return &engine, nil
}

func parseTemplateSpecs(templates []string) templateSpecs {
	tmplts := templateSpecs{}
	for _, pattern := range templates {
		tmplt := templateSpec{
			separator: DefaultSeparator,
		}

		// Format is [separator] [filter] <template> [tag1=value1,tag2=value2]
		parts := strings.Fields(pattern)
		partsLength := len(parts)
		if partsLength < 1 {
			// ignore
			continue
		}
		if partsLength == 1 {
			tmplt.template = pattern
		} else if partsLength == 4 {
			tmplt.separator = parts[0]
			tmplt.filter = parts[1]
			tmplt.template = parts[2]
			tmplt.tagstring = parts[3]
		} else {
			hasTagstring := strings.Contains(parts[partsLength-1], "=")
			if hasTagstring {
				tmplt.tagstring = parts[partsLength-1]
				tmplt.template = parts[partsLength-2]
				if partsLength == 3 {
					tmplt.filter = parts[0]
				}
			} else {
				tmplt.template = parts[partsLength-1]
				if partsLength == 2 {
					tmplt.filter = parts[0]
				} else { // length == 3
					tmplt.separator = parts[0]
					tmplt.filter = parts[1]
				}
			}
		}
		tmplts = append(tmplts, tmplt)
	}
	sort.Sort(tmplts)
	return tmplts
}
