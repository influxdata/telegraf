package xml

import (
	"errors"
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
	Query       string
	DefaultTags map[string]string
}

func NewXMLParser(
	metricName string,
	xmlMergeNodes bool,
	xmlTagNode bool,
	xmlParseArray bool,
	xmlQuery string,
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
		DefaultTags: defaultTags,
	}
}

func (p *XMLParser) Parse(b []byte) ([]telegraf.Metric, error) {
	measurementName := p.MetricName
	timestamp := time.Now().UTC()
	xmlDocument := etree.NewDocument()
	xmlDocument.ReadFromBytes(b)
	metrics := make([]telegraf.Metric, 0)
	xmlTags := make(map[string]string)
	xmlFields := make(map[string]interface{})

	path, err := etree.CompilePath(p.Query)
	if err != nil {
		return nil, err
	}

	root := xmlDocument.FindElementsPath(path)

	if len := len(root); len > 0 {
		if p.ParseArray == true {
			for _, e := range root {
				for _, t := range e.FindElements(".//") {
					tags, fields := p.ParseXmlNode(t)
					xmlTags = mergeTwoTagMaps(xmlTags, tags)
					xmlFields = mergeTwoFieldMaps(xmlFields, fields)
				}

				if p.TagNode == true {
					xmlTags["xml_node_name"] = e.Tag
				}

				metric, err := metric.New(measurementName, xmlTags, xmlFields, timestamp)
				if err != nil {
					return nil, err
				}
				metrics = append(metrics, metric)

				xmlTags = make(map[string]string)
				xmlFields = make(map[string]interface{})
			}
		} else {
			for _, e := range root {
				tags, fields := p.ParseXmlNode(e)

				if p.TagNode == true {
					tags["xml_node_name"] = e.Tag
				}

				if p.MergeNodes == true {
					xmlTags = mergeTwoTagMaps(xmlTags, tags)
					xmlFields = mergeTwoFieldMaps(xmlFields, fields)
				} else {
					metric, err := metric.New(measurementName, tags, fields, timestamp)
					if err != nil {
						return nil, err
					}
					metrics = append(metrics, metric)
				}
			}

			if p.MergeNodes == true {
				metric, err := metric.New(measurementName, xmlTags, xmlFields, timestamp)
				if err != nil {
					return nil, err
				}
				metrics = append(metrics, metric)
			}
		}
	}
	return metrics, nil
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
	if len := len(attrs); len > 0 {
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
