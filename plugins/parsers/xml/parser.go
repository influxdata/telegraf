package xml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	ErrNoMetric = errors.New("no metric in line")
)

type XMLParser struct {
	MetricName  string
	TagKeys     []string
	MergeNodes  bool
	ParseArray  bool
	TagNode     bool
	DetectType  bool
	Query       string
	AttrPrefix  string
	DefaultTags map[string]string
}

func NewXMLParser(
	metricName string,
	xmlMergeNodes bool,
	xmlTagNode bool,
	xmlParseArray bool,
	xmlDetectType bool,
	xmlQuery string,
	xmlAttrPrefix string,
	defaultTags map[string]string,
	tagKeys []string,
) *XMLParser {
	if xmlQuery == "" {
		xmlQuery = "//"
	}

	if xmlAttrPrefix == "" {
		xmlAttrPrefix = "@"
	}

	return &XMLParser{
		MetricName:  metricName,
		TagKeys:     tagKeys,
		MergeNodes:  xmlMergeNodes,
		TagNode:     xmlTagNode,
		ParseArray:  xmlParseArray,
		DetectType:  xmlDetectType,
		Query:       xmlQuery,
		AttrPrefix:  xmlAttrPrefix,
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
			tags, fields := p.ParseXmlNode(t, e)
			xmlTags = mergeTwoTagMaps(xmlTags, tags)
			xmlFields = mergeTwoFieldMaps(xmlFields, fields)
		}

		tags, fields := p.ParseXmlNode(e, e)
		xmlTags = mergeTwoTagMaps(xmlTags, tags)
		xmlFields = mergeTwoFieldMaps(xmlFields, fields)

		if p.TagNode == true {
			xmlTags["xml_node_name"] = e.Tag
		}

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
		tags, fields := p.ParseXmlNode(e, nodes[0].Parent())

		if p.TagNode == true {
			tags["xml_node_name"] = e.Tag
		}

		if p.MergeNodes {
			xmlTags = mergeTwoTagMaps(xmlTags, tags)
			xmlFields = mergeTwoFieldMaps(xmlFields, fields)
		} else {
			metric, err := metric.New(p.MetricName, tags, fields, timestamp)
			if err != nil {
				return nil, err
			}
			results = append(results, metric)
		}
	}

	if p.MergeNodes {
		metric, err := metric.New(p.MetricName, xmlTags, xmlFields, timestamp)
		if err != nil {
			return nil, err
		}
		results = append(results, metric)
	}

	return results, nil
}

func (p *XMLParser) ParseXmlNode(node *etree.Element, parent *etree.Element) (tags map[string]string, fields map[string]interface{}) {
	tags = make(map[string]string)
	fields = make(map[string]interface{})

	nodeText := trimEmptyChars(node.Text())
	if nodeText != "" {
		path := getRelativePath(node, parent)
		if p.isTag(path) {
			tags[path] = node.Text()
		} else {
			fields[path] = p.convertField(node.Text())
		}
	}

	attrs := node.Attr
	if len(attrs) > 0 {
		for _, e := range attrs {
			attrText := trimEmptyChars(e.Value)
			if attrText != "" {
				path := fmt.Sprintf("%v%v%v", getRelativePath(node, parent), p.AttrPrefix, e.Key)
				if p.isTag(path) {
					tags[path] = e.Value
				} else {
					fields[path] = p.convertField(e.Value)
				}
			}
		}
	}
	// add default tags
	tags = mergeTwoTagMaps(tags, p.DefaultTags)
	return tags, fields
}

func (p *XMLParser) isTag(str string) bool {
	for _, a := range p.TagKeys {
		if a == str {
			return true
		}
	}
	return false
}

func getPath(node *etree.Element) (path string) {
	npath := ""

	for seg := node; seg != nil; seg = seg.Parent() {
		if seg.Tag != "" {
			index := ""
			z := seg.Parent().FindElements(seg.Tag)
			if len(z) > 1 {
				for i, x := range z {
					if x.Index() == seg.Index() {
						index = fmt.Sprintf("[%v]", strconv.Itoa(i))
					}
				}
			}
			npath = fmt.Sprintf("%v%v/%v", seg.Tag, index, npath)
		}
	}

	npath = fmt.Sprintf("/%v", npath)
	npath = strings.Trim(npath, "/")
	npath = strings.ReplaceAll(npath, "]/", "_")
	npath = strings.ReplaceAll(npath, "[", "_")
	npath = strings.ReplaceAll(npath, "]", "")
	npath = strings.ReplaceAll(npath, "/", "_")

	return npath
}

func getRelativePath(node *etree.Element, parent *etree.Element) (path string) {
	ppath := getPath(parent)
	npath := getPath(node)
	return strings.TrimPrefix(strings.TrimPrefix(npath, ppath), "_")
}

func (p *XMLParser) convertField(value string) interface{} {
	if p.DetectType {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		} else if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		} else if b, err := strconv.ParseBool(value); err == nil {
			return b
		} else {
			return value
		}
	} else {
		return value
	}
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

func trimEmptyChars(s string) string {
	text := strings.Trim(s, "\n\r\t ")
	return text
}

func (v *XMLParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}
