package xpath

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/jsonquery"
	path "github.com/antchfx/xpath"
	"github.com/srebhan/cborquery"
	"github.com/srebhan/protobufquery"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type dataNode interface{}

type dataDocument interface {
	Parse(buf []byte) (dataNode, error)
	QueryAll(node dataNode, expr string) ([]dataNode, error)
	CreateXPathNavigator(node dataNode) path.NodeNavigator
	GetNodePath(node, relativeTo dataNode, sep string) string
	GetNodeName(node dataNode, sep string, withParent bool) string
	OutputXML(node dataNode) string
}

type Parser struct {
	Format               string            `toml:"-"`
	ProtobufMessageFiles []string          `toml:"xpath_protobuf_files"`
	ProtobufMessageDef   string            `toml:"xpath_protobuf_file" deprecated:"1.32.0;1.40.0;use 'xpath_protobuf_files' instead"`
	ProtobufMessageType  string            `toml:"xpath_protobuf_type"`
	ProtobufImportPaths  []string          `toml:"xpath_protobuf_import_paths"`
	ProtobufSkipBytes    int64             `toml:"xpath_protobuf_skip_bytes"`
	PrintDocument        bool              `toml:"xpath_print_document"`
	AllowEmptySelection  bool              `toml:"xpath_allow_empty_selection"`
	NativeTypes          bool              `toml:"xpath_native_types"`
	Trace                bool              `toml:"xpath_trace" deprecated:"1.35.0;use 'log_level' 'trace' instead"`
	Configs              []Config          `toml:"xpath"`
	DefaultMetricName    string            `toml:"-"`
	DefaultTags          map[string]string `toml:"-"`
	Log                  telegraf.Logger   `toml:"-"`

	// Required for backward compatibility
	ConfigsXML     []Config `toml:"xml" deprecated:"1.23.1;1.35.0;use 'xpath' instead"`
	ConfigsJSON    []Config `toml:"xpath_json" deprecated:"1.23.1;1.35.0;use 'xpath' instead"`
	ConfigsMsgPack []Config `toml:"xpath_msgpack" deprecated:"1.23.1;1.35.0;use 'xpath' instead"`
	ConfigsProto   []Config `toml:"xpath_protobuf" deprecated:"1.23.1;1.35.0;use 'xpath' instead"`

	document dataDocument
}

type Config struct {
	MetricQuery  string            `toml:"metric_name"`
	Selection    string            `toml:"metric_selection"`
	Timestamp    string            `toml:"timestamp"`
	TimestampFmt string            `toml:"timestamp_format"`
	Timezone     string            `toml:"timezone"`
	Tags         map[string]string `toml:"tags"`
	Fields       map[string]string `toml:"fields"`
	FieldsInt    map[string]string `toml:"fields_int"`
	FieldsHex    []string          `toml:"fields_bytes_as_hex"`
	FieldsBase64 []string          `toml:"fields_bytes_as_base64"`

	FieldSelection  string `toml:"field_selection"`
	FieldNameQuery  string `toml:"field_name"`
	FieldValueQuery string `toml:"field_value"`
	FieldNameExpand bool   `toml:"field_name_expansion"`

	TagSelection  string `toml:"tag_selection"`
	TagNameQuery  string `toml:"tag_name"`
	TagValueQuery string `toml:"tag_value"`
	TagNameExpand bool   `toml:"tag_name_expansion"`

	FieldsHexFilter    filter.Filter
	FieldsBase64Filter filter.Filter
	Location           *time.Location
}

func (p *Parser) Init() error {
	switch p.Format {
	case "", "xml":
		p.document = &xmlDocument{}

		// Required for backward compatibility
		if len(p.ConfigsXML) > 0 {
			p.Configs = append(p.Configs, p.ConfigsXML...)
			config.PrintOptionDeprecationNotice("parsers.xpath", "xml", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "1.35.0",
				Notice:    "use 'xpath' instead",
			})
		}
	case "xpath_cbor":
		p.document = &cborDocument{}
	case "xpath_json":
		p.document = &jsonDocument{}

		// Required for backward compatibility
		if len(p.ConfigsJSON) > 0 {
			p.Configs = append(p.Configs, p.ConfigsJSON...)
			config.PrintOptionDeprecationNotice("parsers.xpath", "xpath_json", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "1.35.0",
				Notice:    "use 'xpath' instead",
			})
		}
	case "xpath_msgpack":
		p.document = &msgpackDocument{}

		// Required for backward compatibility
		if len(p.ConfigsMsgPack) > 0 {
			p.Configs = append(p.Configs, p.ConfigsMsgPack...)
			config.PrintOptionDeprecationNotice("parsers.xpath", "xpath_msgpack", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "1.35.0",
				Notice:    "use 'xpath' instead",
			})
		}
	case "xpath_protobuf":
		if p.ProtobufMessageDef != "" && !slices.Contains(p.ProtobufMessageFiles, p.ProtobufMessageDef) {
			p.ProtobufMessageFiles = append(p.ProtobufMessageFiles, p.ProtobufMessageDef)
		}
		pbdoc := protobufDocument{
			MessageFiles: p.ProtobufMessageFiles,
			MessageType:  p.ProtobufMessageType,
			ImportPaths:  p.ProtobufImportPaths,
			SkipBytes:    p.ProtobufSkipBytes,
			Log:          p.Log,
		}
		if err := pbdoc.Init(); err != nil {
			return err
		}
		p.document = &pbdoc

		// Required for backward compatibility
		if len(p.ConfigsProto) > 0 {
			p.Configs = append(p.Configs, p.ConfigsProto...)
			config.PrintOptionDeprecationNotice("parsers.xpath", "xpath_proto", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "1.35.0",
				Notice:    "use 'xpath' instead",
			})
		}
	default:
		return fmt.Errorf("unknown data-format %q for xpath parser", p.Format)
	}

	// Make sure we do have a metric name
	if p.DefaultMetricName == "" {
		return errors.New("missing default metric name")
	}

	// Update the configs with default values
	for i, cfg := range p.Configs {
		if cfg.Selection == "" {
			cfg.Selection = "/"
		}
		if cfg.TimestampFmt == "" {
			cfg.TimestampFmt = "unix"
		}
		if cfg.Timezone == "" {
			cfg.Location = time.UTC
		} else {
			loc, err := time.LoadLocation(cfg.Timezone)
			if err != nil {
				return fmt.Errorf("invalid location in config %d: %w", i+1, err)
			}
			cfg.Location = loc
		}
		f, err := filter.Compile(cfg.FieldsHex)
		if err != nil {
			return fmt.Errorf("creating hex-fields filter failed: %w", err)
		}
		cfg.FieldsHexFilter = f

		bf, err := filter.Compile(cfg.FieldsBase64)
		if err != nil {
			return fmt.Errorf("creating base64-fields filter failed: %w", err)
		}
		cfg.FieldsBase64Filter = bf

		p.Configs[i] = cfg
	}

	return nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	t := time.Now()

	// Parse the XML
	doc, err := p.document.Parse(buf)
	if err != nil {
		return nil, err
	}
	if p.PrintDocument {
		p.Log.Debugf("XML document equivalent: %q", p.document.OutputXML(doc))
	}

	// Queries
	metrics := make([]telegraf.Metric, 0)
	p.Log.Debugf("Number of configs: %d", len(p.Configs))
	for _, cfg := range p.Configs {
		selectedNodes, err := p.document.QueryAll(doc, cfg.Selection)
		if err != nil {
			return nil, err
		}
		if (len(selectedNodes) < 1 || selectedNodes[0] == nil) && !p.AllowEmptySelection {
			p.debugEmptyQuery("metric selection", doc, cfg.Selection)
			return metrics, errors.New("cannot parse with empty selection node")
		}
		p.Log.Debugf("Number of selected metric nodes: %d", len(selectedNodes))

		for _, selected := range selectedNodes {
			m, err := p.parseQuery(t, doc, selected, cfg)
			if err != nil {
				return metrics, err
			}

			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	switch len(metrics) {
	case 0:
		return nil, nil
	case 1:
		return metrics[0], nil
	default:
		return metrics[0], fmt.Errorf("cannot parse line with multiple (%d) metrics", len(metrics))
	}
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) parseQuery(starttime time.Time, doc, selected dataNode, cfg Config) (telegraf.Metric, error) {
	var timestamp time.Time
	var metricname string

	// Determine the metric name. If a query was specified, use the result of this query and the default metric name
	// otherwise.
	metricname = p.DefaultMetricName
	if len(cfg.MetricQuery) > 0 {
		v, err := p.executeQuery(doc, selected, cfg.MetricQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to query metric name: %w", err)
		}
		var ok bool
		if metricname, ok = v.(string); !ok {
			if v == nil {
				p.Log.Infof("Hint: Empty metric-name-node. If you wanted to set a constant please use `metric_name = \"'name'\"`.")
			}
			return nil, fmt.Errorf("failed to query metric name: query result is of type %T not 'string'", v)
		}
	}

	// By default take the time the parser was invoked and override the value
	// with the queried timestamp if an expression was specified.
	timestamp = starttime
	if len(cfg.Timestamp) > 0 {
		v, err := p.executeQuery(doc, selected, cfg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to query timestamp: %w", err)
		}
		if v != nil {
			timestamp, err = internal.ParseTimestamp(cfg.TimestampFmt, v, cfg.Location)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
		}
	}

	// Query tags and add default ones
	tags := make(map[string]string)

	// Handle the tag batch definitions if any.
	if len(cfg.TagSelection) > 0 {
		tagnamequery := "name()"
		tagvaluequery := "."
		if len(cfg.TagNameQuery) > 0 {
			tagnamequery = cfg.TagNameQuery
		}
		if len(cfg.TagValueQuery) > 0 {
			tagvaluequery = cfg.TagValueQuery
		}

		// Query all tags
		selectedTagNodes, err := p.document.QueryAll(selected, cfg.TagSelection)
		if err != nil {
			return nil, err
		}
		p.Log.Debugf("Number of selected tag nodes: %d", len(selectedTagNodes))
		if len(selectedTagNodes) > 0 && selectedTagNodes[0] != nil {
			for _, selectedtag := range selectedTagNodes {
				n, err := p.executeQuery(doc, selectedtag, tagnamequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query tag name with query %q: %w", tagnamequery, err)
				}
				name, ok := n.(string)
				if !ok {
					return nil, fmt.Errorf("failed to query tag name with query %q: result is not a string (%v)", tagnamequery, n)
				}
				name = p.constructFieldName(selected, selectedtag, name, cfg.TagNameExpand)

				v, err := p.executeQuery(doc, selectedtag, tagvaluequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query tag value for %q: %w", name, err)
				}

				// Check if field name already exists and if so, append an index number.
				if _, ok := tags[name]; ok {
					for i := 1; ; i++ {
						p := name + "_" + strconv.Itoa(i)
						if _, ok := tags[p]; !ok {
							name = p
							break
						}
					}
				}

				// Convert the tag to be a string
				s, err := internal.ToString(v)
				if err != nil {
					return nil, fmt.Errorf("failed to query tag value for %q: result is not a string (%v)", name, v)
				}
				tags[name] = s
			}
		} else {
			p.debugEmptyQuery("tag selection", selected, cfg.TagSelection)
		}
	}

	// Handle explicitly defined tags
	for name, query := range cfg.Tags {
		// Execute the query and cast the returned values into strings
		v, err := p.executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query tag %q: %w", name, err)
		}
		switch v := v.(type) {
		case string:
			tags[name] = v
		case bool:
			tags[name] = strconv.FormatBool(v)
		case float64:
			tags[name] = strconv.FormatFloat(v, 'G', -1, 64)
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unknown format '%T' for tag %q", v, name)
		}
	}

	// Add default tags
	for name, v := range p.DefaultTags {
		tags[name] = v
	}

	// Query fields
	fields := make(map[string]interface{})

	// Handle the field batch definitions if any.
	if len(cfg.FieldSelection) > 0 {
		fieldnamequery := "name()"
		fieldvaluequery := "."
		if len(cfg.FieldNameQuery) > 0 {
			fieldnamequery = cfg.FieldNameQuery
		}
		if len(cfg.FieldValueQuery) > 0 {
			fieldvaluequery = cfg.FieldValueQuery
		}

		// Query all fields
		selectedFieldNodes, err := p.document.QueryAll(selected, cfg.FieldSelection)
		if err != nil {
			return nil, err
		}
		p.Log.Debugf("Number of selected field nodes: %d", len(selectedFieldNodes))
		if len(selectedFieldNodes) > 0 && selectedFieldNodes[0] != nil {
			for _, selectedfield := range selectedFieldNodes {
				n, err := p.executeQuery(doc, selectedfield, fieldnamequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query field name with query %q: %w", fieldnamequery, err)
				}
				name, ok := n.(string)
				if !ok {
					return nil, fmt.Errorf("failed to query field name with query %q: result is not a string (%v)", fieldnamequery, n)
				}
				name = p.constructFieldName(selected, selectedfield, name, cfg.FieldNameExpand)

				v, err := p.executeQuery(doc, selectedfield, fieldvaluequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query field value for %q: %w", name, err)
				}

				// Check if field name already exists and if so, append an index number.
				if _, ok := fields[name]; ok {
					for i := 1; ; i++ {
						p := name + "_" + strconv.Itoa(i)
						if _, ok := fields[p]; !ok {
							name = p
							break
						}
					}
				}

				// Handle complex types which would be dropped otherwise for
				// native type handling
				if v != nil {
					switch reflect.TypeOf(v).Kind() {
					case reflect.Array, reflect.Slice, reflect.Map:
						if b, ok := v.([]byte); ok {
							if cfg.FieldsHexFilter != nil && cfg.FieldsHexFilter.Match(name) {
								v = hex.EncodeToString(b)
							}
							if cfg.FieldsBase64Filter != nil && cfg.FieldsBase64Filter.Match(name) {
								v = base64.StdEncoding.EncodeToString(b)
							}
						} else {
							v = fmt.Sprintf("%v", v)
						}
					}
				}

				fields[name] = v
			}
		} else {
			p.debugEmptyQuery("field selection", selected, cfg.FieldSelection)
		}
	}

	// Handle explicitly defined fields
	for name, query := range cfg.FieldsInt {
		// Execute the query and cast the returned values into integers
		v, err := p.executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query field (int) %q: %w", name, err)
		}
		switch v := v.(type) {
		case string:
			fields[name], err = strconv.ParseInt(v, 10, 54)
			if err != nil {
				return nil, fmt.Errorf("failed to parse field (int) %q: %w", name, err)
			}
		case bool:
			fields[name] = int64(0)
			if v {
				fields[name] = int64(1)
			}
		case float64:
			fields[name] = int64(v)
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unknown format '%T' for field (int) %q", v, name)
		}
	}

	for name, query := range cfg.Fields {
		// Execute the query and store the result in fields
		v, err := p.executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query field %q: %w", name, err)
		}

		// Handle complex types which would be dropped otherwise for
		// native type handling
		if v != nil {
			switch reflect.TypeOf(v).Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if b, ok := v.([]byte); ok {
					if cfg.FieldsHexFilter != nil && cfg.FieldsHexFilter.Match(name) {
						v = hex.EncodeToString(b)
					}
					if cfg.FieldsBase64Filter != nil && cfg.FieldsBase64Filter.Match(name) {
						v = base64.StdEncoding.EncodeToString(b)
					}
				} else {
					v = fmt.Sprintf("%v", v)
				}
			}
		}

		fields[name] = v
	}

	return metric.New(metricname, tags, fields, timestamp), nil
}

func (p *Parser) executeQuery(doc, selected dataNode, query string) (r interface{}, err error) {
	// Check if the query is relative or absolute and set the root for the query
	root := selected
	if strings.HasPrefix(query, "/") {
		root = doc
	}

	// Compile the query
	expr, err := path.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile query %q: %w", query, err)
	}

	// Evaluate the compiled expression and handle returned node-iterators
	// separately. Those iterators will be returned for queries directly
	// referencing a node (value or attribute).
	n := expr.Evaluate(p.document.CreateXPathNavigator(root))
	iter, ok := n.(*path.NodeIterator)
	if !ok {
		return n, nil
	}
	// We got an iterator, so take the first match and get the referenced
	// property. This will always be a string.
	if iter.MoveNext() {
		current := iter.Current()
		// If the dataformat supports native types and if support is
		// enabled, we should return the native type of the data
		if p.NativeTypes {
			switch nn := current.(type) {
			case *cborquery.NodeNavigator:
				return nn.GetValue(), nil
			case *jsonquery.NodeNavigator:
				return nn.GetValue(), nil
			case *protobufquery.NodeNavigator:
				return nn.GetValue(), nil
			}
		}

		return iter.Current().Value(), nil
	}

	return nil, nil
}

func splitLastPathElement(query string) []string {
	// This is a rudimentary xpath-parser that splits the path
	// into the last path element and the remaining path-part.
	// The last path element is then further split into
	// parts such as attributes or selectors. Each returned
	// element is a full path!

	// Nothing left
	if query == "" || query == "/" || query == "//" || query == "." {
		return []string{}
	}

	separatorIdx := strings.LastIndex(query, "/")
	if separatorIdx < 0 {
		query = "./" + query
		separatorIdx = 1
	}

	// For double slash we want to split at the first slash
	if separatorIdx > 0 && query[separatorIdx-1] == byte('/') {
		separatorIdx--
	}

	base := query[:separatorIdx]
	if base == "" {
		base = "/"
	}

	elements := make([]string, 0, 3)
	elements = append(elements, base)

	offset := separatorIdx
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

func (p *Parser) constructFieldName(root, node dataNode, name string, expand bool) string {
	var expansion string

	// In case the name is empty we should determine the current node's name.
	// This involves array index expansion in case the parent of the node is
	// and array. If we expanded here, we should skip our parent as this is
	// already encoded in the name
	if name == "" {
		name = p.document.GetNodeName(node, "_", !expand)
	}

	// If name expansion is requested, construct a path between the current
	// node and the root node of the selection. Concatenate the elements with
	// an underscore.
	if expand {
		expansion = p.document.GetNodePath(node, root, "_")
	}

	if len(expansion) > 0 {
		name = expansion + "_" + name
	}
	return name
}

func (p *Parser) debugEmptyQuery(operation string, root dataNode, initialquery string) {
	if p.Log == nil || !(p.Log.Level().Includes(telegraf.Trace) || p.Trace) { // for backward compatibility
		return
	}

	query := initialquery

	// We already know that the
	p.Log.Tracef("got 0 nodes for query %q in %s", query, operation)
	for {
		parts := splitLastPathElement(query)
		if len(parts) < 1 {
			return
		}
		for i := len(parts) - 1; i >= 0; i-- {
			q := parts[i]
			nodes, err := p.document.QueryAll(root, q)
			if err != nil {
				p.Log.Tracef("executing query %q in %s failed: %v", q, operation, err)
				return
			}
			p.Log.Tracef("got %d nodes for query %q in %s", len(nodes), q, operation)
			if len(nodes) > 0 && nodes[0] != nil {
				return
			}
			query = parts[0]
		}
	}
}

func init() {
	// Register all variants
	parsers.Add("xml",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xml",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_cbor",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_cbor",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_json",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_json",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_msgpack",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_msgpack",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_protobuf",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_protobuf",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
}
