package graphite

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Minimum and maximum supported dates for timestamps.
var (
	MinDate = time.Date(1901, 12, 13, 0, 0, 0, 0, time.UTC)
	MaxDate = time.Date(2038, 1, 19, 0, 0, 0, 0, time.UTC)
)

// Parser encapsulates a Graphite Parser.
type GraphiteParser struct {
	Separator   string
	Templates   []string
	DefaultTags map[string]string

	matcher *matcher
}

func (p *GraphiteParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func NewGraphiteParser(
	separator string,
	templates []string,
	defaultTags map[string]string,
) (*GraphiteParser, error) {
	var err error

	if separator == "" {
		separator = DefaultSeparator
	}
	p := &GraphiteParser{
		Separator: separator,
		Templates: templates,
	}

	if defaultTags != nil {
		p.DefaultTags = defaultTags
	}

	matcher := newMatcher()
	p.matcher = matcher
	defaultTemplate, _ := NewTemplate("measurement*", nil, p.Separator)
	matcher.AddDefaultTemplate(defaultTemplate)

	tmplts := parsedTemplates{}
	for _, pattern := range p.Templates {
		tmplt := parsedTemplate{}
		tmplt.template = pattern
		// Format is [filter] <template> [tag1=value1,tag2=value2]
		parts := strings.Fields(pattern)
		if len(parts) < 1 {
			continue
		} else if len(parts) >= 2 {
			if strings.Contains(parts[1], "=") {
				tmplt.template = parts[0]
				tmplt.tagstring = parts[1]
			} else {
				tmplt.filter = parts[0]
				tmplt.template = parts[1]
				if len(parts) > 2 {
					tmplt.tagstring = parts[2]
				}
			}
		}
		tmplts = append(tmplts, tmplt)
	}

	sort.Sort(tmplts)
	for _, tmplt := range tmplts {
		if err := p.addToMatcher(tmplt); err != nil {
			return nil, err
		}
	}

	if err != nil {
		return p, fmt.Errorf("exec input parser config is error: %s ", err.Error())
	} else {
		return p, nil
	}
}

func (p *GraphiteParser) addToMatcher(tmplt parsedTemplate) error {
	// Parse out the default tags specific to this template
	tags := map[string]string{}
	if tmplt.tagstring != "" {
		for _, kv := range strings.Split(tmplt.tagstring, ",") {
			parts := strings.Split(kv, "=")
			tags[parts[0]] = parts[1]
		}
	}

	tmpl, err := NewTemplate(tmplt.template, tags, p.Separator)
	if err != nil {
		return err
	}
	p.matcher.Add(tmplt.filter, tmpl)
	return nil
}

func (p *GraphiteParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// add newline to end if not exists:
	if len(buf) > 0 && !bytes.HasSuffix(buf, []byte("\n")) {
		buf = append(buf, []byte("\n")...)
	}

	metrics := make([]telegraf.Metric, 0)

	var errStr string
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)
	for {
		// Read up to the next newline.
		buf, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil && err != io.EOF {
			return metrics, err
		}

		// Trim the buffer, even though there should be no padding
		line := strings.TrimSpace(string(buf))
		metric, err := p.ParseLine(line)

		if err == nil {
			metrics = append(metrics, metric)
		} else {
			errStr += err.Error() + "\n"
		}
	}

	if errStr != "" {
		return metrics, fmt.Errorf(strings.TrimSpace(errStr))
	}
	return metrics, nil
}

// Parse performs Graphite parsing of a single line.
func (p *GraphiteParser) ParseLine(line string) (telegraf.Metric, error) {
	// Break into 3 fields (name, value, timestamp).
	fields := strings.Fields(line)
	if len(fields) != 2 && len(fields) != 3 {
		return nil, fmt.Errorf("received %q which doesn't have required fields", line)
	}

	// decode the name and tags
	template := p.matcher.Match(fields[0])
	measurement, tags, field, err := template.Apply(fields[0])
	if err != nil {
		return nil, err
	}

	// Could not extract measurement, use the raw value
	if measurement == "" {
		measurement = fields[0]
	}

	// Parse value.
	v, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf(`field "%s" value: %s`, fields[0], err)
	}

	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil, &UnsupposedValueError{Field: fields[0], Value: v}
	}

	fieldValues := map[string]interface{}{}
	if field != "" {
		fieldValues[field] = v
	} else {
		fieldValues["value"] = v
	}

	// If no 3rd field, use now as timestamp
	timestamp := time.Now().UTC()

	if len(fields) == 3 {
		// Parse timestamp.
		unixTime, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return nil, fmt.Errorf(`field "%s" time: %s`, fields[0], err)
		}

		// -1 is a special value that gets converted to current UTC time
		// See https://github.com/graphite-project/carbon/issues/54
		if unixTime != float64(-1) {
			// Check if we have fractional seconds
			timestamp = time.Unix(int64(unixTime), int64((unixTime-math.Floor(unixTime))*float64(time.Second)))
			if timestamp.Before(MinDate) || timestamp.After(MaxDate) {
				return nil, fmt.Errorf("timestamp out of range")
			}
		}
	}
	// Set the default tags on the point if they are not already set
	for k, v := range p.DefaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	return metric.New(measurement, tags, fieldValues, timestamp)
}

// ApplyTemplate extracts the template fields from the given line and
// returns the measurement name and tags.
func (p *GraphiteParser) ApplyTemplate(line string) (string, map[string]string, string, error) {
	// Break line into fields (name, value, timestamp), only name is used
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", make(map[string]string), "", nil
	}
	// decode the name and tags
	template := p.matcher.Match(fields[0])
	name, tags, field, err := template.Apply(fields[0])

	// Set the default tags on the point if they are not already set
	for k, v := range p.DefaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	return name, tags, field, err
}

// template represents a pattern and tags to map a graphite metric string to a influxdb Point
type template struct {
	tags              []string
	defaultTags       map[string]string
	greedyField       bool
	greedyMeasurement bool
	separator         string
}

// NewTemplate returns a new template ensuring it has a measurement
// specified.
func NewTemplate(pattern string, defaultTags map[string]string, separator string) (*template, error) {
	tags := strings.Split(pattern, ".")
	hasMeasurement := false
	template := &template{tags: tags, defaultTags: defaultTags, separator: separator}

	for _, tag := range tags {
		if strings.HasPrefix(tag, "measurement") {
			hasMeasurement = true
		}
		if tag == "measurement*" {
			template.greedyMeasurement = true
		} else if tag == "field*" {
			template.greedyField = true
		}
	}

	if !hasMeasurement {
		return nil, fmt.Errorf("no measurement specified for template. %q", pattern)
	}

	return template, nil
}

// Apply extracts the template fields from the given line and returns the measurement
// name and tags
func (t *template) Apply(line string) (string, map[string]string, string, error) {
	fields := strings.Split(line, ".")
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
	for _, tag := range t.tags {
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
				strings.Join(t.tags, t.separator))
	}

	for i, tag := range t.tags {
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
		outtags[k] = strings.Join(values, t.separator)
	}

	return strings.Join(measurement, t.separator), outtags, strings.Join(field, t.separator), nil
}

// matcher determines which template should be applied to a given metric
// based on a filter tree.
type matcher struct {
	root            *node
	defaultTemplate *template
}

func newMatcher() *matcher {
	return &matcher{
		root: &node{},
	}
}

// Add inserts the template in the filter tree based the given filter
func (m *matcher) Add(filter string, template *template) {
	if filter == "" {
		m.AddDefaultTemplate(template)
		return
	}
	m.root.Insert(filter, template)
}

func (m *matcher) AddDefaultTemplate(template *template) {
	m.defaultTemplate = template
}

// Match returns the template that matches the given graphite line
func (m *matcher) Match(line string) *template {
	tmpl := m.root.Search(line)
	if tmpl != nil {
		return tmpl
	}

	return m.defaultTemplate
}

// node is an item in a sorted k-ary tree.  Each child is sorted by its value.
// The special value of "*", is always last.
type node struct {
	value    string
	children nodes
	template *template
}

func (n *node) insert(values []string, template *template) {
	// Add the end, set the template
	if len(values) == 0 {
		n.template = template
		return
	}

	// See if the the current element already exists in the tree. If so, insert the
	// into that sub-tree
	for _, v := range n.children {
		if v.value == values[0] {
			v.insert(values[1:], template)
			return
		}
	}

	// New element, add it to the tree and sort the children
	newNode := &node{value: values[0]}
	n.children = append(n.children, newNode)
	sort.Sort(&n.children)

	// Now insert the rest of the tree into the new element
	newNode.insert(values[1:], template)
}

// Insert inserts the given string template into the tree.  The filter string is separated
// on "." and each part is used as the path in the tree.
func (n *node) Insert(filter string, template *template) {
	n.insert(strings.Split(filter, "."), template)
}

func (n *node) search(lineParts []string) *template {
	// Nothing to search
	if len(lineParts) == 0 || len(n.children) == 0 {
		return n.template
	}

	// If last element is a wildcard, don't include in this search since it's sorted
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
		return n.children[i].search(lineParts[1:])
	}
	// Not an exact match, see if we have a wildcard child to search
	if n.children[len(n.children)-1].value == "*" {
		return n.children[len(n.children)-1].search(lineParts[1:])
	}
	return n.template
}

func (n *node) Search(line string) *template {
	return n.search(strings.Split(line, "."))
}

type nodes []*node

// Less returns a boolean indicating whether the filter at position j
// is less than the filter at position k.  Filters are order by string
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

func (n *nodes) Swap(i, j int) { (*n)[i], (*n)[j] = (*n)[j], (*n)[i] }
func (n *nodes) Len() int      { return len(*n) }

type parsedTemplate struct {
	template  string
	filter    string
	tagstring string
}
type parsedTemplates []parsedTemplate

func (e parsedTemplates) Less(j, k int) bool {
	if len(e[j].filter) == 0 && len(e[k].filter) == 0 {
		nj := len(strings.Split(e[j].template, "."))
		nk := len(strings.Split(e[k].template, "."))
		return nj < nk
	}
	if len(e[j].filter) == 0 {
		return true
	}
	if len(e[k].filter) == 0 {
		return false
	}

	nj := len(strings.Split(e[j].template, "."))
	nk := len(strings.Split(e[k].template, "."))
	return nj < nk
}
func (e parsedTemplates) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e parsedTemplates) Len() int      { return len(e) }
