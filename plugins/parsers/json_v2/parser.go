package json_v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/tidwall/gjson"
)

type TimeFunc func() time.Time

type Parser struct {
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
	TimeFunc    func() time.Time

	metricName string

	// For objects
	iterateObjects bool
	ignoredKeys    []string
	includedKeys   []string
	names          map[string]string
	valueTypes     map[string]string
	tagList        []string
}

type Config struct {
	MetricName         string
	UniformCollections []UniformCollection
	ObjectSelections   []ObjectSelection
}

type UniformCollection struct {
	Query     string `toml:"query"`      // REQUIRED
	Name      string `toml:"name"`       // OPTIONAL
	ValueType string `toml:"value_type"` // OPTIONAL
	SetType   string `toml:"set_type"`   // OPTIONAL
}

type ObjectSelection struct {
	Query        string            `toml:"query"`       // REQUIRED
	Names        map[string]string `toml:"names"`       // OPTIONAL
	ValueTypes   map[string]string `toml:"value_types"` // OPTIONAL
	TagList      []string          `toml:"tag_list"`
	IncludedKeys []string          `toml:"included_keys"`
	IgnoredKeys  []string          `toml:"ignored_keys"`
}

// One field is required, set to true if one is added
var FieldExists = false

type MetricNode struct {
	// For unfirom collection
	RootFieldName string
	DesiredType   string // Can be "int", "bool", "string"

	SetType string // Can either be "field" or "tag"
	// TODO: Make this a slice, array's in an object could make this expand
	Metric telegraf.Metric
	gjson.Result
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	// This is used to set a default time in the unit tests
	if p.TimeFunc == nil {
		p.TimeFunc = time.Now
	}

	// Only valid JSON is supported
	if !gjson.Valid(string(input)) {
		return nil, fmt.Errorf("Invalid JSON provided, unable to parse")
	}

	var t []telegraf.Metric

	for _, config := range p.Configs {
		p.metricName = config.MetricName
		// Process all `uniform_collection` configurations
		uniformCollection, err := p.processUniformCollections(config.UniformCollections, input)
		if err != nil {
			return nil, err
		}

		objectMetrics, err := p.processObjectSelections(config.ObjectSelections, input)
		if err != nil {
			return nil, err
		}

		if len(objectMetrics) != 0 && len(uniformCollection) != 0 {
			t = append(t, cartesianProduct(objectMetrics, uniformCollection)...)
		} else if len(objectMetrics) != 0 {
			t = append(t, objectMetrics...)
		} else if len(uniformCollection) != 0 {
			t = append(t, uniformCollection...)
		}
		//TODO: Should object metrics and uniform metrics be merged?!?!?
	}

	if !FieldExists {
		return nil, fmt.Errorf("No field configured for the metrics")
	}

	return t, nil
}

func (p *Parser) processUniformCollections(uniformCollection []UniformCollection, input []byte) ([]telegraf.Metric, error) {
	if len(uniformCollection) == 0 {
		return nil, nil
	}

	// For each uniform_collection configuration, get all the metric data returned from the query
	// Keep the metric data per field separate so all results from each query can be combined
	p.iterateObjects = false
	var metrics [][]telegraf.Metric
	for _, c := range uniformCollection {
		result := gjson.GetBytes(input, c.Query)

		if result.IsObject() {
			p.Log.Debugf("Found object in the uniform collection query: %s, ignoring it please use object_selection to gather metrics from objects", c.Query)
			continue
		}

		var setType string
		if c.SetType == "tag" {
			setType = "tag"
		} else {
			FieldExists = true
			setType = "field"
		}

		fieldName := c.Name
		// Default to the last query word, should be the upper key name
		// TODO: figure out what to do with special characters, probably ok to remove any special characters?
		if fieldName == "" {
			s := strings.Split(c.Query, ".")
			fieldName = s[len(s)-1]
		}

		// Store result into a MetricNode to keep metadata together
		mNode := MetricNode{
			RootFieldName: fieldName,
			DesiredType:   c.ValueType,
			SetType:       setType,
			Metric: metric.New(
				p.metricName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			),
			Result: result,
		}
		// Expand all array's and nested arrays into separate metrics
		nodes, err := p.expandArray(mNode)
		if err != nil {
			return nil, err
		}

		var m []telegraf.Metric
		for _, n := range nodes {
			m = append(m, n.Metric)
		}
		metrics = append(metrics, m)
	}

	for i := 1; i < len(metrics); i++ {
		metrics[i] = cartesianProduct(metrics[i-1], metrics[i])
	}

	return metrics[len(metrics)-1], nil
}

func cartesianProduct(a, b []telegraf.Metric) []telegraf.Metric {
	p := make([]telegraf.Metric, len(a)*len(b))
	i := 0
	for _, a := range a {
		for _, b := range b {
			m := a.Copy()
			mergeMetric(b, m)
			p[i] = m
			i++
		}
	}

	return p
}

func mergeMetric(a telegraf.Metric, m telegraf.Metric) {
	for _, f := range a.FieldList() {
		m.AddField(f.Key, f.Value)
	}
	for _, t := range a.TagList() {
		m.AddTag(t.Key, t.Value)
	}
}

// expandArray will recursively create a new MetricNode for each element in a JSON array
func (p *Parser) expandArray(result MetricNode) ([]MetricNode, error) {
	var results []MetricNode

	if result.IsObject() {
		if !p.iterateObjects {
			p.Log.Debugf("Found object in the uniform collection query ignoring it please use object_selection to gather metrics from objects")
			return results, nil
		}
		_, err := p.combineObject(result)
		if err != nil {
			return nil, err
		}
	}

	if result.IsArray() {
		var err error
		result.ForEach(func(_, val gjson.Result) bool {
			m := metric.New(
				p.metricName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			)
			// TODO: implement `ignore_objects` config key to ignore this error
			// TODO: combineObjects calls this if it has an array of objects, need to handle this case
			if val.IsObject() {
				if p.iterateObjects {
					n := MetricNode{
						SetType:       result.SetType,
						RootFieldName: result.RootFieldName,
						Metric:        m,
						Result:        val,
					}
					r, err := p.combineObject(n)
					if err != nil {
						return false
					}

					results = append(results, r...)
				} else {
					p.Log.Debugf("Found object in the uniform collection query ignoring it please use object_selection to gather metrics from objects")
				}
				if len(results) != 0 {
					for _, newResult := range results {
						mergeMetric(result.Metric, newResult.Metric)
					}
				}
				return true
			}

			if result.SetType == "field" {
				for _, f := range result.Metric.FieldList() {
					m.AddField(f.Key, f.Value)
				}
			} else {
				for _, f := range result.Metric.TagList() {
					m.AddTag(f.Key, f.Value)
				}
			}
			n := MetricNode{
				SetType:       result.SetType,
				RootFieldName: result.RootFieldName,
				Metric:        m,
				Result:        val,
			}
			r, err := p.expandArray(n)
			if err != nil {
				return false
			}
			results = append(results, r...)
			return true
		})
		if err != nil {
			return nil, err
		}
	} else {
		if result.SetType == "field" && !result.IsObject() {
			v, err := p.convertType(result.Value(), result.DesiredType, result.RootFieldName)
			if err != nil {
				return nil, err
			}
			result.Metric.AddField(result.RootFieldName, v)
		} else if !result.IsObject() {
			v, err := p.convertType(result.Value(), "string", result.RootFieldName)
			if err != nil {
				return nil, err
			}
			result.Metric.AddTag(result.RootFieldName, v.(string))
		}
		results = append(results, result)
	}

	return results, nil
}

func (p *Parser) processObjectSelections(objectSelections []ObjectSelection, input []byte) ([]telegraf.Metric, error) {
	p.iterateObjects = true
	var t []telegraf.Metric
	for _, c := range objectSelections {
		// TODO: Verify this through tag_list somehow
		FieldExists = true

		p.tagList = c.TagList
		p.valueTypes = c.ValueTypes
		p.names = c.Names
		p.ignoredKeys = c.IgnoredKeys
		p.includedKeys = c.IncludedKeys
		result := gjson.GetBytes(input, c.Query)

		if result.Type == gjson.Null {
			return nil, fmt.Errorf("Query returned null")
		}

		// TODO: Figoure out how to handle root fieldname, will be blank
		// Default to the last query word, should be the upper key name
		// TODO: figure out what to do with special characters, probably ok to remove any special characters?
		s := strings.Split(c.Query, ".")
		fieldName := s[len(s)-1]

		rootObject := MetricNode{
			RootFieldName: fieldName,
			SetType:       "field", // TODO: Somehow set this to field
			Metric: metric.New(
				p.metricName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			),
			Result: result,
		}
		metrics, err := p.expandArray(rootObject)
		if err != nil {
			return nil, err
		}
		for _, m := range metrics {
			t = append(t, m.Metric)
		}
	}

	return t, nil
}

func (p *Parser) combineObject(result MetricNode) ([]MetricNode, error) {
	var results []MetricNode
	if result.IsArray() || result.IsObject() {
		var err error
		result.ForEach(func(key, val gjson.Result) bool {
			if p.isIgnored(key.String()) || !p.isIncluded(key.String()) {
				return true
			}

			arrayNode := MetricNode{
				RootFieldName: key.String(),
				Metric:        result.Metric,
				Result:        val,
				SetType:       result.SetType,
			}

			if val.IsObject() {
				_, err := p.combineObject(arrayNode)
				if err != nil {
					return false
				}
			} else {
				if !val.IsArray() {
					for k, t := range p.valueTypes {
						if key.String() == k {
							arrayNode.DesiredType = t
							break
						}
					}

					for _, t := range p.tagList {
						if key.String() == t {
							arrayNode.SetType = t
							break
						}
					}
				}
				results, err = p.expandArray(arrayNode)
				if err != nil {
					return false
				}
			}

			return true
		})

		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (p *Parser) isIncluded(key string) bool {
	if len(p.includedKeys) == 0 {
		return true
	}
	for _, i := range p.includedKeys {
		if i == key {
			return true
		}
	}
	return false
}

func (p *Parser) isIgnored(key string) bool {
	for _, i := range p.ignoredKeys {
		if i == key {
			return true
		}
	}
	return false
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	return nil, fmt.Errorf("ParseLine is designed for parsing influx line protocol, therefore not implemented for parsing JSON")
}

func (p *Parser) SetDefaultTags(tags map[string]string) {

}

func (p *Parser) SetTimeFunc(fn TimeFunc) {
	p.TimeFunc = fn
}

// convertType will convert the value parsed from the input JSON to the specified type in the config
func (p *Parser) convertType(input interface{}, desiredType string, name string) (interface{}, error) {
	switch inputType := input.(type) {
	case string:
		if desiredType != "string" {
			switch desiredType {
			case "int":
				r, err := strconv.Atoi(inputType)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type int: %v", name, err)
				}
				return r, nil
			case "float":
				r, err := strconv.ParseFloat(inputType, 64)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type float: %v", name, err)
				}
				return r, nil
			case "bool":
				r, err := strconv.ParseBool(inputType)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type bool: %v", name, err)
				}
				return r, nil
			}
		}
	case bool:
		if desiredType != "bool" {
			switch desiredType {
			case "string":
				return strconv.FormatBool(inputType), nil
			case "int":
				if inputType {
					return int64(1), nil
				}

				return int64(0), nil
			}
		}
	case float64:
		if desiredType != "float" {
			switch desiredType {
			case "string":
				return fmt.Sprint(inputType), nil
			case "int":
				return int64(inputType), nil
			case "bool":
				if inputType == 0 {
					return false, nil
				} else if inputType == 1 {
					return true, nil
				} else {
					return nil, fmt.Errorf("Unable to convert field '%s' to type bool", name)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unknown format '%T' for field  '%s'", inputType, name)
	}

	return input, nil
}
