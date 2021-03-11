package xml

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type Parser struct {
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
}

type Config struct {
	MetricName   string
	MetricQuery  string            `toml:"metric_name"`
	Selection    string            `toml:"metric_selection"`
	Timestamp    string            `toml:"timestamp"`
	TimestampFmt string            `toml:"timestamp_format"`
	Tags         map[string]string `toml:"tags"`
	Fields       map[string]string `toml:"fields"`
	FieldsInt    map[string]string `toml:"fields_int"`

	FieldSelection  string `toml:"field_selection"`
	FieldNameQuery  string `toml:"field_name"`
	FieldValueQuery string `toml:"field_value"`
	FieldNameExpand bool   `toml:"field_name_expansion"`
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	t := time.Now()

	// Parse the XML
	doc, err := xmlquery.Parse(strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}

	// Queries
	metrics := make([]telegraf.Metric, 0)
	for _, config := range p.Configs {
		if len(config.Selection) == 0 {
			config.Selection = "/"
		}
		selectedNodes, err := xmlquery.QueryAll(doc, config.Selection)
		if err != nil {
			return nil, err
		}
		if len(selectedNodes) < 1 || selectedNodes[0] == nil {
			p.debugEmptyQuery("metric selection", doc, config.Selection)
			return nil, fmt.Errorf("cannot parse with empty selection node")
		}
		p.Log.Debugf("Number of selected metric nodes: %d", len(selectedNodes))

		for _, selected := range selectedNodes {
			m, err := p.parseQuery(t, doc, selected, config)
			if err != nil {
				return metrics, err
			}

			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	t := time.Now()

	switch len(p.Configs) {
	case 0:
		return nil, nil
	case 1:
		config := p.Configs[0]

		doc, err := xmlquery.Parse(strings.NewReader(line))
		if err != nil {
			return nil, err
		}

		selected := doc
		if len(config.Selection) > 0 {
			selectedNodes, err := xmlquery.QueryAll(doc, config.Selection)
			if err != nil {
				return nil, err
			}
			if len(selectedNodes) < 1 || selectedNodes[0] == nil {
				p.debugEmptyQuery("metric selection", doc, config.Selection)
				return nil, fmt.Errorf("cannot parse line with empty selection")
			} else if len(selectedNodes) != 1 {
				return nil, fmt.Errorf("cannot parse line with multiple selected nodes (%d)", len(selectedNodes))
			}
			selected = selectedNodes[0]
		}

		return p.parseQuery(t, doc, selected, config)
	}
	return nil, fmt.Errorf("cannot parse line with multiple (%d) configurations", len(p.Configs))
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) parseQuery(starttime time.Time, doc, selected *xmlquery.Node, config Config) (telegraf.Metric, error) {
	var timestamp time.Time
	var metricname string

	// Determine the metric name. If a query was specified, use the result of this query and the default metric name
	// otherwise.
	metricname = config.MetricName
	if len(config.MetricQuery) > 0 {
		v, err := executeQuery(doc, selected, config.MetricQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to query metric name: %v", err)
		}
		metricname = v.(string)
	}

	// By default take the time the parser was invoked and override the value
	// with the queried timestamp if an expresion was specified.
	timestamp = starttime
	if len(config.Timestamp) > 0 {
		v, err := executeQuery(doc, selected, config.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to query timestamp: %v", err)
		}
		switch v.(type) {
		case string:
			// Parse the string with the given format or assume the string to contain
			// a unix timestamp in seconds if no format is given.
			if len(config.TimestampFmt) < 1 || strings.HasPrefix(config.TimestampFmt, "unix") {
				var nanoseconds int64

				t, err := strconv.ParseFloat(v.(string), 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse unix timestamp: %v", err)
				}

				switch config.TimestampFmt {
				case "unix_ns":
					nanoseconds = int64(t)
				case "unix_us":
					nanoseconds = int64(t * 1e3)
				case "unix_ms":
					nanoseconds = int64(t * 1e6)
				default:
					nanoseconds = int64(t * 1e9)
				}
				timestamp = time.Unix(0, nanoseconds)
			} else {
				timestamp, err = time.Parse(config.TimestampFmt, v.(string))
				if err != nil {
					return nil, fmt.Errorf("failed to query timestamp format: %v", err)
				}
			}
		case float64:
			// Assume the value to contain a timestamp in seconds and fractions thereof.
			timestamp = time.Unix(0, int64(v.(float64)*1e9))
		case nil:
			// No timestamp found. Just ignore the time and use "starttime"
		default:
			return nil, fmt.Errorf("unknown format '%T' for timestamp query '%v'", v, config.Timestamp)
		}
	}

	// Query tags and add default ones
	tags := make(map[string]string)
	for name, query := range config.Tags {
		// Execute the query and cast the returned values into strings
		v, err := executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query tag '%s': %v", name, err)
		}
		switch v.(type) {
		case string:
			tags[name] = v.(string)
		case bool:
			tags[name] = strconv.FormatBool(v.(bool))
		case float64:
			tags[name] = strconv.FormatFloat(v.(float64), 'G', -1, 64)
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unknown format '%T' for tag '%s'", v, name)
		}
	}
	for name, v := range p.DefaultTags {
		tags[name] = v
	}

	// Query fields
	fields := make(map[string]interface{})
	for name, query := range config.FieldsInt {
		// Execute the query and cast the returned values into integers
		v, err := executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query field (int) '%s': %v", name, err)
		}
		switch v.(type) {
		case string:
			fields[name], err = strconv.ParseInt(v.(string), 10, 54)
			if err != nil {
				return nil, fmt.Errorf("failed to parse field (int) '%s': %v", name, err)
			}
		case bool:
			fields[name] = int64(0)
			if v.(bool) {
				fields[name] = int64(1)
			}
		case float64:
			fields[name] = int64(v.(float64))
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unknown format '%T' for field (int) '%s'", v, name)
		}
	}

	for name, query := range config.Fields {
		// Execute the query and store the result in fields
		v, err := executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query field '%s': %v", name, err)
		}
		fields[name] = v
	}

	// Handle the field batch definitions if any.
	if len(config.FieldSelection) > 0 {
		fieldnamequery := "name()"
		fieldvaluequery := "."
		if len(config.FieldNameQuery) > 0 {
			fieldnamequery = config.FieldNameQuery
		}
		if len(config.FieldValueQuery) > 0 {
			fieldvaluequery = config.FieldValueQuery
		}

		// Query all fields
		selectedFieldNodes, err := xmlquery.QueryAll(selected, config.FieldSelection)
		if err != nil {
			return nil, err
		}
		p.Log.Debugf("Number of selected field nodes: %d", len(selectedFieldNodes))
		if len(selectedFieldNodes) > 0 && selectedFieldNodes[0] != nil {
			for _, selectedfield := range selectedFieldNodes {
				n, err := executeQuery(doc, selectedfield, fieldnamequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query field name with query '%s': %v", fieldnamequery, err)
				}
				name, ok := n.(string)
				if !ok {
					return nil, fmt.Errorf("failed to query field name with query '%s': result is not a string (%v)", fieldnamequery, n)
				}
				v, err := executeQuery(doc, selectedfield, fieldvaluequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query field value for '%s': %v", name, err)
				}
				path := name
				if config.FieldNameExpand {
					p := getNodePath(selectedfield, selected, "_")
					if len(p) > 0 {
						path = p + "_" + name
					}
				}

				// Check if field name already exists and if so, append an index number.
				if _, ok := fields[path]; ok {
					for i := 1; ; i++ {
						p := path + "_" + strconv.Itoa(i)
						if _, ok := fields[p]; !ok {
							path = p
							break
						}
					}
				}

				fields[path] = v
			}
		} else {
			p.debugEmptyQuery("field selection", selected, config.FieldSelection)
		}
	}

	return metric.New(metricname, tags, fields, timestamp)
}

func getNodePath(node, relativeTo *xmlquery.Node, sep string) string {
	names := make([]string, 0)

	// Climb up the tree and collect the node names
	n := node.Parent
	for n != nil && n != relativeTo {
		names = append(names, n.Data)
		n = n.Parent
	}

	if len(names) < 1 {
		return ""
	}

	// Construct the nodes
	path := ""
	for _, name := range names {
		path = name + sep + path
	}

	return path[:len(path)-1]
}

func executeQuery(doc, selected *xmlquery.Node, query string) (r interface{}, err error) {
	// Check if the query is relative or absolute and set the root for the query
	root := selected
	if strings.HasPrefix(query, "/") {
		root = doc
	}

	// Compile the query
	expr, err := xpath.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile query '%s': %v", query, err)
	}

	// Evaluate the compiled expression and handle returned node-iterators
	// separately. Those iterators will be returned for queries directly
	// referencing a node (value or attribute).
	n := expr.Evaluate(xmlquery.CreateXPathNavigator(root))
	if iter, ok := n.(*xpath.NodeIterator); ok {
		// We got an iterator, so take the first match and get the referenced
		// property. This will always be a string.
		if iter.MoveNext() {
			r = iter.Current().Value()
		}
	} else {
		r = n
	}

	return r, nil
}

func splitLastPathElement(query string) []string {
	// This is a rudimentary xpath-parser that splits the path
	// into the last path element and the remaining path-part.
	// The last path element is then further splitted into
	// parts such as attributes or selectors. Each returned
	// element is a full path!

	// Nothing left
	if query == "" || query == "/" || query == "//" || query == "." {
		return []string{}
	}

	seperatorIdx := strings.LastIndex(query, "/")
	if seperatorIdx < 0 {
		query = "./" + query
		seperatorIdx = 1
	}

	// For double slash we want to split at the first slash
	if seperatorIdx > 0 && query[seperatorIdx-1] == byte('/') {
		seperatorIdx--
	}

	base := query[:seperatorIdx]
	if base == "" {
		base = "/"
	}

	elements := make([]string, 1)
	elements[0] = base

	offset := seperatorIdx
	if i := strings.Index(query[offset:], "::"); i >= 0 {
		// Check for axis operator
		offset += i
		elements = append(elements, query[:offset]+"::*")
	}

	if i := strings.Index(query[offset:], "["); i >= 0 {
		// Check for predicates
		offset += i
		elements = append(elements, query[:offset])
	} else if i := strings.Index(query[offset:], "@"); i >= 0 {
		// Check for attributes
		offset += i
		elements = append(elements, query[:offset])
	}

	return elements
}

func (p *Parser) debugEmptyQuery(operation string, root *xmlquery.Node, initialquery string) {
	if p.Log == nil {
		return
	}

	query := initialquery

	// We already know that the
	p.Log.Debugf("got 0 nodes for query %q in %s", query, operation)
	for {
		parts := splitLastPathElement(query)
		if len(parts) < 1 {
			return
		}
		for i := len(parts) - 1; i >= 0; i-- {
			q := parts[i]
			nodes, err := xmlquery.QueryAll(root, q)
			if err != nil {
				p.Log.Debugf("executing query %q in %s failed: %v", q, operation, err)
				return
			}
			p.Log.Debugf("got %d nodes for query %q in %s", len(nodes), q, operation)
			if len(nodes) > 0 && nodes[0] != nil {
				return
			}
			query = parts[0]
		}
	}
}
