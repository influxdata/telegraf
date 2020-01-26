package xml

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	intExpr   string = "^\\d+$"
	floatExpr string = "^\\d+\\.+\\d+$"
)

type XMLParser struct {
	TagKeys      []string
	CombineNodes bool
	TagNode      bool
	Query        string
	DefaultTags  map[string]string
}

func (p *XMLParser) NewXMLParser(xmlCombineNodes bool, xmlTagNode bool, xmlQuery string, defaultTags map[string]string, tagKeys []string) {
	p.DefaultTags = xmlCombineNodes
	p.TagNode = xmlTagNode
	p.DefaultTags = defaultTags
	p.TagKeys = tagKeys

	if xmlQuery == "" {
		p.Query = "//"
	} else {
		p.Query = xmlQuery
	}
}

func (p *XMLParser) Parse(b []byte) ([]telegraf.Metric, error) {
	xmlDocument := etree.NewDocument()
	xmlDocument.ReadFromBytes(b)
	timestamp := time.Now()
	metrics := make([]telegraf.Metric, 0)
	xmlTags := make(map[string]string)
	xmlFields := make(map[string]interface{})

	root := xmlDocument.FindElements(p.Query)
	if len := len(root); len > 0 {
		for _, e := range root {

			tags, fields := ParseXmlNode(e)
			if p.nodeTag == true {
				tags["node_name"] = e.Tag
			}

			if p.combineNodes == false {
				metric, err := metric.New("xml", tags, fields, timestamp)
				if err != nil {
					return nil, err
				}
				metrics = append(metrics, metric)
			} else {
				xmlTags = append(xmlTags, tags)
				xmlFields = append(xmlFields, fields)
			}

		}
		if p.combineNodes == true {
			metric, err := metric.New("xml", xmlTags, xmlFields, timestamp)
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

func (v *XMLParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(s))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}
	return metrics[0], nil
}

func (p *XMLParser) ParseXmlNode(node *etree.Element) (tags map[string]string, fields map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	nodeText := trimEmptyChars(node.Text())
	if nodeText != "" {
		//toTag := isFieldATagCandidate(node.Tag)
		if isFieldATagCandidate(node.Tag) {
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
				if isFieldATagCandidate(node.Tag) {
					tags[e.Key] = e.Value
				} else {
					fields[e.Key] = identifyFieldType(e.Value)
				}
			}
		}
	}
}

func (p *XMLParser) isFieldATagCandidate(str string) bool {
	for _, a := range p.TagKeys {
		if a == str {
			return true
		}
	}
	return false
}

func identifyFieldType(value string) interface{} {
	matched, err := regexp.Match(intExpr, value)
	if matched {
		i, err := strconv.ParseInt(value, 10, 64)
		return i
	}
	matched, err := regexp.Match(floatExpr, value)
	if matched {
		f, err := strconv.ParseFloat(value, 64)
		return f
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
