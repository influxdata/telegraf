package xml

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	ErrNoMetric  = errors.New("no metric in line")
	AttrSelector = regexp.MustCompile(`.*\/@(?P<AttrName>.+)$`)
)

type XMLParser struct {
	MetricName  string
	TagKeys     []string
	MergeNodes  bool
	ParseArray  bool
	TagNode     bool
	Query       string
	Tags        []string
	Fields      []string
	DefaultTags map[string]string
}

func NewXMLParser(
	metricName string,
	xmlMergeNodes bool,
	xmlTagNode bool,
	xmlParseArray bool,
	xmlQuery string,
	xmlTags []string,
	xmlFields []string,
	defaultTags map[string]string,
	tagKeys []string,
) *XMLParser {
	if xmlQuery == "" {
		xmlQuery = "//"
	}

	return &XMLParser{
		MetricName:  metricName,
		TagKeys:     tagKeys,
		MergeNodes:  xmlMergeNodes,
		TagNode:     xmlTagNode,
		ParseArray:  xmlParseArray,
		Query:       xmlQuery,
		Tags:        xmlTags,
		Fields:      xmlFields,
		DefaultTags: defaultTags,
	}
}

func (p *XMLParser) Parse(b []byte) ([]telegraf.Metric, error) {
	timestamp := time.Now().UTC()
	xmlDocument := etree.NewDocument()

	err := xmlDocument.ReadFromBytes(b)
	if err != nil {
		return nil, err
	}

	path, err := etree.CompilePath(p.Query)
	if err != nil {
		return nil, err
	}

	root := xmlDocument.FindElementsPath(path)

	//	if len(p.Measurement) > 0 {
	//		name, err := selectSingleValue(&xmlDocument.Element, p.Measurement)
	//		if err != nil {
	//			return nil, err
	//		}
	//		p.MetricName = name
	//	}

	if len(root) > 0 {
		if p.ParseArray == true {
			return p.ParseAsArray(root, timestamp)
		} else {
			return p.ParseAsObject(root, timestamp)
		}
	}

	return make([]telegraf.Metric, 0), nil
}

func (p *XMLParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}
	return metrics[0], nil
}

func (p *XMLParser) ParseAsArray(nodes []*etree.Element, timestamp time.Time) ([]telegraf.Metric, error) {
	results := make([]telegraf.Metric, 0)
	xmlTags := make(map[string]string)
	xmlFields := make(map[string]interface{})

	for _, e := range nodes {
		for _, t := range e.FindElements(".//") {
			tags, fields := p.ParseXmlNode(t)
			xmlTags = mergeTwoTagMaps(xmlTags, tags)
			xmlFields = mergeTwoFieldMaps(xmlFields, fields)
		}

		tags, fields := p.ParseXmlNode(e)
		xmlTags = mergeTwoTagMaps(xmlTags, tags)
		xmlFields = mergeTwoFieldMaps(xmlFields, fields)

		extTags, extFields := p.GetExtraValues(e)
		xmlTags = mergeTwoTagMaps(xmlTags, extTags)
		xmlFields = mergeTwoFieldMaps(xmlFields, extFields)

		if p.TagNode == true {
			xmlTags["xml_node_name"] = e.Tag
		}

		// add default tags
		xmlTags = mergeTwoTagMaps(xmlTags, p.DefaultTags)
		metric, err := metric.New(p.MetricName, xmlTags, xmlFields, timestamp)
		if err != nil {
			return nil, err
		}
		results = append(results, metric)

		xmlTags = make(map[string]string)
		xmlFields = make(map[string]interface{})
	}

	return results, nil
}

func (p *XMLParser) ParseAsObject(nodes []*etree.Element, timestamp time.Time) ([]telegraf.Metric, error) {
	results := make([]telegraf.Metric, 0)
	xmlTags := make(map[string]string)
	xmlFields := make(map[string]interface{})

	for _, e := range nodes {
		tags, fields := p.ParseXmlNode(e)

		extTags, extFields := p.GetExtraValues(e)
		tags = mergeTwoTagMaps(tags, extTags)
		fields = mergeTwoFieldMaps(fields, extFields)

		if p.TagNode == true {
			tags["xml_node_name"] = e.Tag
		}

		if p.MergeNodes == true {
			xmlTags = mergeTwoTagMaps(xmlTags, tags)
			xmlFields = mergeTwoFieldMaps(xmlFields, fields)
		} else {
			// add default tags
			tags = mergeTwoTagMaps(tags, p.DefaultTags)
			metric, err := metric.New(p.MetricName, tags, fields, timestamp)
			if err != nil {
				return nil, err
			}
			results = append(results, metric)
		}
	}

	if p.MergeNodes == true {
		// add default tags
		xmlTags = mergeTwoTagMaps(xmlTags, p.DefaultTags)
		metric, err := metric.New(p.MetricName, xmlTags, xmlFields, timestamp)
		if err != nil {
			return nil, err
		}
		results = append(results, metric)
	}

	return results, nil
}

func (p *XMLParser) ParseXmlNode(node *etree.Element) (tags map[string]string, fields map[string]interface{}) {
	tags = make(map[string]string)
	fields = make(map[string]interface{})

	nodeText := trimEmptyChars(node.Text())
	if nodeText != "" {
		if p.isTag(node.Tag) {
			tags[node.Tag] = node.Text()
		} else {
			fields[node.Tag] = convertField(node.Text())
		}
	}

	attrs := node.Attr
	if len(attrs) > 0 {
		for _, e := range attrs {
			attrText := trimEmptyChars(e.Value)
			if attrText != "" {
				if p.isTag(e.Key) {
					tags[e.Key] = e.Value
				} else {
					fields[e.Key] = convertField(e.Value)
				}
			}
		}
	}
	return tags, fields
}

func (p *XMLParser) GetExtraValues(doc *etree.Element) (tags map[string]string, fields map[string]interface{}) {
	tags = make(map[string]string)
	fields = make(map[string]interface{})

	if len(p.Tags) > 0 {
		for _, q := range p.Tags {
			k, v := selectSingleValue(doc, q)
			if (k != "") && (v != "") {
				tags[k] = v
			}
		}
	}

	if len(p.Fields) > 0 {
		for _, q := range p.Fields {
			k, v := selectSingleValue(doc, q)
			if (k != "") && (v != "") {
				fields[k] = convertField(v)
			}
		}
	}

	return tags, fields
}

func selectSingleValue(doc *etree.Element, query string) (key string, value string) {
	if AttrSelector.MatchString(query) {
		attrName := AttrSelector.FindStringSubmatch(query)[1]
		nodePath := strings.TrimSuffix(query, fmt.Sprintf("/@%v", attrName))

		node := selectSingleNode(doc, nodePath)
		if node == nil {
			return "", ""
		}

		attr := node.SelectAttrValue(attrName, "")
		return attrName, trimEmptyChars(attr)
	} else {
		node := selectSingleNode(doc, query)
		if node == nil {
			return "", ""
		}

		return node.Tag, trimEmptyChars(node.Text())
	}
}

func selectSingleNode(doc *etree.Element, query string) *etree.Element {
	path, err := etree.CompilePath(query)
	if err != nil {
		log.Printf("E! [parsers.xml] An error occurred while executing %q query: %v", query, err.Error())
		return nil
	}

	node := doc.FindElementPath(path)
	return node
}

func (p *XMLParser) isTag(str string) bool {
	for _, a := range p.TagKeys {
		if a == str {
			return true
		}
	}
	return false
}

func mergeTwoFieldMaps(parent map[string]interface{}, child map[string]interface{}) map[string]interface{} {
	for key, value := range child {
		parent[key] = value
	}
	return parent
}

func mergeTwoTagMaps(parent map[string]string, child map[string]string) map[string]string {
	for key, value := range child {
		parent[key] = value
	}
	return parent
}

func convertField(value string) interface{} {
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	} else if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	} else if b, err := strconv.ParseBool(value); err == nil {
		return b
	} else {
		return value
	}
}

func trimEmptyChars(s string) string {
	text := strings.Trim(s, "\n\r\t ")
	return text
}

func (v *XMLParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}
