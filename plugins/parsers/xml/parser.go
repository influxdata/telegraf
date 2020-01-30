package xml

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	intExpr     string = "^\\d+$"
	floatExpr   string = "^\\d+\\.+\\d+$"
	ErrNoMetric        = errors.New("no metric in line")
)

type XMLParser struct {
	TagKeys      []string
	CombineNodes bool
	TagNode      bool
	Query        string
	DefaultTags  map[string]string
}

func NewXMLParser(xmlCombineNodes bool,
	xmlTagNode bool,
	xmlQuery string,
	defaultTags map[string]string,
	tagKeys []string) *XMLParser {

	if xmlQuery == "" {
		xmlQuery = "//"
	} else {
		xmlQuery = xmlQuery
	}

    return &XMLParser{
        TagKeys: tagKeys,  
        CombineNodes: xmlCombineNodes,
        TagNode: xmlTagNode,
        Query: xmlQuery,
        DefaultTags: defaultTags,
    }
}

func (p *XMLParser) Parse(b []byte) ([]telegraf.Metric, error) {
	timestamp := time.Now()
	xmlDocument := etree.NewDocument()
	xmlDocument.ReadFromBytes(b)
	metrics := make([]telegraf.Metric, 0)
	xmlTags := make(map[string]string)
	xmlFields := make(map[string]interface{})

	root := xmlDocument.FindElements(p.Query)
	if len := len(root); len > 0 {
		for _, e := range root {

			tags, fields := p.ParseXmlNode(e)
			if p.TagNode == true {
				tags["node_name"] = e.Tag
			}

			if p.CombineNodes == false {
				metric, err := metric.New("xml", tags, fields, timestamp)
				if err != nil {
					return nil, err
				}
				metrics = append(metrics, metric)
			} else {
				xmlTags = mergeTwoTagMaps(xmlTags, tags)
				xmlFields = mergeTwoFieldMaps(xmlFields, fields)
			}

		}
		if p.CombineNodes == true {
			metric, err := metric.New("xml", xmlTags, xmlFields, timestamp)
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, metric)
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
		//toTag := isFieldATagCandidate(node.Tag)
		if p.isFieldATagCandidate(node.Tag) {
			tags[node.Tag] = node.Text()
		} else {
			fields[node.Tag] = identifyFieldType(node.Text())
		}
	}

	attrs := node.Attr
	if len := len(attrs); len > 0 {
		for _, e := range attrs {
			attrText := trimEmptyChars(e.Value)
			if attrText != "" {
				if p.isFieldATagCandidate(e.Key) {
					tags[e.Key] = e.Value
				} else {
					fields[e.Key] = identifyFieldType(e.Value)
				}
			}
		}
	}
	return tags, fields
}

func (p *XMLParser) isFieldATagCandidate(str string) bool {
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

func identifyFieldType(value string) interface{} {
	temp := []byte(value)
	matched, err := regexp.Match(intExpr, temp)
	if matched && err == nil {
		i, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return i
		}
	}
	matched, err = regexp.Match(floatExpr, temp)
	if matched && err == nil {
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return f
		}
	}
	return value
}

func trimEmptyChars(s string) string {
	text := strings.Trim(s, "\n")
	text = strings.Trim(text, "\r")
	text = strings.Trim(text, "\t")
	text = strings.Trim(text, " ")

	return text
}

func (v *XMLParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}
