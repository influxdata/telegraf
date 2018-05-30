package templating

import (
	"strings"
)

// matcher determines which template should be applied to a given metric
// based on a filter tree.
type matcher struct {
	root            *node
	defaultTemplate *Template
}

// newMatcher creates a new matcher.
func newMatcher(defaultTemplate *Template) *matcher {
	return &matcher{
		root:            &node{},
		defaultTemplate: defaultTemplate,
	}
}

func (m *matcher) addSpec(tmplt templateSpec) error {
	// Parse out the default tags specific to this template
	tags := map[string]string{}
	if tmplt.tagstring != "" {
		for _, kv := range strings.Split(tmplt.tagstring, ",") {
			parts := strings.Split(kv, "=")
			tags[parts[0]] = parts[1]
		}
	}

	tmpl, err := NewTemplate(tmplt.separator, tmplt.template, tags)
	if err != nil {
		return err
	}
	m.add(tmplt.filter, tmpl)
	return nil
}

// add inserts the template in the filter tree based the given filter
func (m *matcher) add(filter string, template *Template) {
	if filter == "" {
		m.defaultTemplate = template
		m.root.separator = template.separator
		return
	}
	m.root.insert(filter, template)
}

// match returns the template that matches the given measurement line.
// If no template matches, the default template is returned.
func (m *matcher) match(line string) *Template {
	tmpl := m.root.search(line)
	if tmpl != nil {
		return tmpl
	}
	return m.defaultTemplate
}
