package templating

import (
	"fmt"
	"strings"
)

// Template represents a pattern and tags to map a metric string to a influxdb Point
type Template struct {
	separator         string
	parts             []string
	defaultTags       map[string]string
	greedyField       bool
	greedyMeasurement bool
}

// apply extracts the template fields from the given line and returns the measurement
// name, tags and field name
func (t *Template) Apply(line string, joiner string) (string, map[string]string, string, error) {
	fields := strings.Split(line, t.separator)
	var (
		measurement []string
		tags        = make(map[string][]string)
		field       []string
	)

	// Set any default tags
	for k, v := range t.defaultTags {
		tags[k] = append(tags[k], v)
	}

	// See if an invalid combination has been specified in the template:
	for _, tag := range t.parts {
		if tag == "measurement*" {
			t.greedyMeasurement = true
		} else if tag == "field*" {
			t.greedyField = true
		}
	}
	if t.greedyField && t.greedyMeasurement {
		return "", nil, "",
			fmt.Errorf("either 'field*' or 'measurement*' can be used in each "+
				"template (but not both together): %q",
				strings.Join(t.parts, joiner))
	}

	for i, tag := range t.parts {
		if i >= len(fields) {
			continue
		}
		if tag == "" {
			continue
		}

		switch tag {
		case "measurement":
			measurement = append(measurement, fields[i])
		case "field":
			field = append(field, fields[i])
		case "field*":
			field = append(field, fields[i:]...)
			break
		case "measurement*":
			measurement = append(measurement, fields[i:]...)
			break
		default:
			tags[tag] = append(tags[tag], fields[i])
		}
	}

	// Convert to map of strings.
	outtags := make(map[string]string)
	for k, values := range tags {
		outtags[k] = strings.Join(values, joiner)
	}

	return strings.Join(measurement, joiner), outtags, strings.Join(field, joiner), nil
}

func NewDefaultTemplateWithPattern(pattern string) (*Template, error) {
	return NewTemplate(DefaultSeparator, pattern, nil)
}

// NewTemplate returns a new template ensuring it has a measurement
// specified.
func NewTemplate(separator string, pattern string, defaultTags map[string]string) (*Template, error) {
	parts := strings.Split(pattern, separator)
	hasMeasurement := false
	template := &Template{
		separator:   separator,
		parts:       parts,
		defaultTags: defaultTags,
	}

	for _, part := range parts {
		if strings.HasPrefix(part, "measurement") {
			hasMeasurement = true
		}
		if part == "measurement*" {
			template.greedyMeasurement = true
		} else if part == "field*" {
			template.greedyField = true
		}
	}

	if !hasMeasurement {
		return nil, fmt.Errorf("no measurement specified for template. %q", pattern)
	}

	return template, nil
}

// templateSpec is a template string split in its constituent parts
type templateSpec struct {
	separator string
	filter    string
	template  string
	tagstring string
}

// templateSpecs is simply an array of template specs implementing the sorting interface
type templateSpecs []templateSpec

// Less reports whether the element with
// index j should sort before the element with index k.
func (e templateSpecs) Less(j, k int) bool {
	if len(e[j].filter) == 0 && len(e[k].filter) == 0 {
		jlength := len(strings.Split(e[j].template, e[j].separator))
		klength := len(strings.Split(e[k].template, e[k].separator))
		return jlength < klength
	}
	if len(e[j].filter) == 0 {
		return true
	}
	if len(e[k].filter) == 0 {
		return false
	}

	jlength := len(strings.Split(e[j].template, e[j].separator))
	klength := len(strings.Split(e[k].template, e[k].separator))
	return jlength < klength
}

// Swap swaps the elements with indexes i and j.
func (e templateSpecs) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// Len is the number of elements in the collection.
func (e templateSpecs) Len() int { return len(e) }
