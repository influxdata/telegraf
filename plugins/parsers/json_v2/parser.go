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

	measurementName string

	// For objects
	iterateObjects bool
	ignoredKeys    []string
	includedKeys   []string
	names          map[string]string
	valueTypes     map[string]string
	tagList        []string
}

type Config struct {
	DefaultMeasurementName string
	MeasurementNameQuery   string
	UniformCollections     []UniformCollection
	ObjectSelections       []ObjectSelection
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

// A Metric requires at least one field, returns an error if false so that the user isn't surprised when no metrics are returned
var FieldExists = false

type MetricNode struct {
	SetName     string
	DesiredType string // Can be "int", "bool", "string"

	Metric telegraf.Metric
	gjson.Result
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	p.TimeFunc = time.Now

	// Only valid JSON is supported
	if !gjson.Valid(string(input)) {
		return nil, fmt.Errorf("Invalid JSON provided, unable to parse")
	}

	var metrics []telegraf.Metric

	for _, c := range p.Configs {
		p.measurementName = c.DefaultMeasurementName
		if c.MeasurementNameQuery != "" {
			result := gjson.GetBytes(input, c.MeasurementNameQuery)
			if !result.IsArray() && !result.IsObject() {
				p.measurementName = result.String()
			}
		}

		uniformCollection, err := p.processUniformCollections(c.UniformCollections, input)
		if err != nil {
			return nil, err
		}

		objectMetrics, err := p.processObjectSelections(c.ObjectSelections, input)
		if err != nil {
			return nil, err
		}

		if len(objectMetrics) != 0 && len(uniformCollection) != 0 {
			metrics = append(metrics, cartesianProduct(objectMetrics, uniformCollection)...)
		} else if len(objectMetrics) != 0 {
			metrics = append(metrics, objectMetrics...)
		} else if len(uniformCollection) != 0 {
			metrics = append(metrics, uniformCollection...)
		}
	}

	if !FieldExists {
		return nil, fmt.Errorf("No field configured for the metrics")
	}

	for k, v := range p.DefaultTags {
		for _, t := range metrics {
			t.AddTag(k, v)
		}
	}

	return metrics, nil
}

// processUniformCollections will iterate over all 'uniform_collection' configs and create metrics for each
// A uniform collection can either be a single value or an array of values, each resulting in its own metric
// For multiple configs, a set of metrics is created from the cartesian product of each separate config
func (p *Parser) processUniformCollections(uniformCollection []UniformCollection, input []byte) ([]telegraf.Metric, error) {
	if len(uniformCollection) == 0 {
		return nil, nil
	}

	p.iterateObjects = false
	var metrics [][]telegraf.Metric

	for _, config := range uniformCollection {
		result := gjson.GetBytes(input, config.Query)

		if result.IsObject() {
			p.Log.Debugf("Found object in the uniform collection query: %s, ignoring it please use object_selection to gather metrics from objects", config.Query)
			continue
		}

		if config.SetType != "" && config.SetType != "tag" && config.SetType != "field" {
			p.Log.Debugf("set_type was defined as %v, it can only be configured to 'tag' or 'field'", config.SetType)
		}

		setType := "tag"
		if config.SetType != "tag" {
			FieldExists = true
			setType = "field"
		}

		setName := config.Name
		// Default to the last query word, should be the upper key name
		if setName == "" {
			s := strings.Split(config.Query, ".")
			setName = s[len(s)-1]
		}

		mNode := MetricNode{
			SetName:     setName,
			DesiredType: config.ValueType,
			Metric: metric.New(
				p.measurementName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			),
			Result: result,
		}

		// Expand all array's and nested arrays into separate metrics
		nodes, err := p.expandArray(mNode, setType)
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

// expandArray will recursively create a new MetricNode for each element in a JSON array or single value
func (p *Parser) expandArray(result MetricNode, setType string) ([]MetricNode, error) {
	var results []MetricNode

	if result.IsObject() {
		if !p.iterateObjects {
			p.Log.Debugf("Found object in the uniform collection query ignoring it please use object_selection to gather metrics from objects")
			return results, nil
		}
		r, err := p.combineObject(result)
		if err != nil {
			return nil, err
		}
		results = append(results, r...)
		return results, nil
	}

	if result.IsArray() {
		var err error
		result.ForEach(func(_, val gjson.Result) bool {
			m := metric.New(
				p.measurementName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			)

			if val.IsObject() {
				if p.iterateObjects {
					n := MetricNode{
						SetName: result.SetName,
						Metric:  m,
						Result:  val,
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

			for _, f := range result.Metric.FieldList() {
				m.AddField(f.Key, f.Value)
			}
			for _, f := range result.Metric.TagList() {
				m.AddTag(f.Key, f.Value)
			}
			n := MetricNode{
				SetName: result.SetName,
				Metric:  m,
				Result:  val,
			}
			r, err := p.expandArray(n, setType)
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
		if setType == "field" && !result.IsObject() {
			v, err := p.convertType(result.Value(), result.DesiredType, result.SetName)
			if err != nil {
				return nil, err
			}
			result.Metric.AddField(result.SetName, v)
		} else if !result.IsObject() {
			v, err := p.convertType(result.Value(), "string", result.SetName)
			if err != nil {
				return nil, err
			}
			result.Metric.AddTag(result.SetName, v.(string))
		}
		results = append(results, result)
	}

	return results, nil
}

// processObjectSelections will iterate over all 'object_selection' configs and create metrics for each
func (p *Parser) processObjectSelections(objectSelections []ObjectSelection, input []byte) ([]telegraf.Metric, error) {
	p.iterateObjects = true
	var t []telegraf.Metric
	for _, c := range objectSelections {
		p.tagList = c.TagList
		p.valueTypes = c.ValueTypes
		p.names = c.Names
		p.ignoredKeys = c.IgnoredKeys
		p.includedKeys = c.IncludedKeys

		result := gjson.GetBytes(input, c.Query)

		if result.Type == gjson.Null {
			return nil, fmt.Errorf("Query returned null")
		}

		// Default to the last query word, should be the upper key name
		s := strings.Split(c.Query, ".")
		fieldName := s[len(s)-1]

		rootObject := MetricNode{
			SetName: fieldName,
			Metric: metric.New(
				p.measurementName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			),
			Result: result,
		}
		metrics, err := p.expandArray(rootObject, "field")
		if err != nil {
			return nil, err
		}
		for _, m := range metrics {
			t = append(t, m.Metric)
		}
	}

	return t, nil
}

// combineObject will add all fields/tags to a single metric
// If the object has multiple array's as elements it won't comine those, they will remain separate metrics
func (p *Parser) combineObject(result MetricNode) ([]MetricNode, error) {
	var results []MetricNode
	if result.IsArray() || result.IsObject() {
		var err error
		var prevArray bool
		result.ForEach(func(key, val gjson.Result) bool {
			if p.isIgnored(key.String()) || !p.isIncluded(key.String()) {
				return true
			}

			// Determine if field/tag set name is configured
			setName := key.String()
			for k, n := range p.names {
				if k == key.String() {
					setName = n
					break
				}
			}

			arrayNode := MetricNode{
				SetName: setName,
				Metric:  result.Metric,
				Result:  val,
			}

			if val.IsObject() {
				prevArray = false
				_, err := p.combineObject(arrayNode)
				if err != nil {
					return false
				}
			} else {
				setType := "field"

				if !val.IsArray() {
					for k, t := range p.valueTypes {
						if key.String() == k {
							arrayNode.DesiredType = t
							break
						}
					}

					for _, t := range p.tagList {
						if key.String() == t {
							setType = "tag"
							break
						}
					}
				}

				if setType == "field" {
					FieldExists = true
				}

				r, err := p.expandArray(arrayNode, setType)
				if err != nil {
					return false
				}
				if prevArray {
					if !arrayNode.IsArray() {
						// If another non-array element was found, merge it into all previous gathered metrics
						if len(results) != 0 {
							for _, newResult := range results {
								mergeMetric(result.Metric, newResult.Metric)
							}
						}
					} else {
						// Multiple array's won't be merged but kept separate, add additional metrics gathered from an array
						results = append(results, r...)
					}
				} else {
					// Continue using the same metric if its an object
					results = r
				}

				if val.IsArray() {
					prevArray = true
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
	p.DefaultTags = tags
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
