package json_v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/tidwall/gjson"
)

type Parser struct {
	InputJSON   []byte
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
	Timestamp   time.Time

	measurementName string

	iterateObjects bool

	currentSettings JSONObject
	pathResults     []PathResult
}

type PathResult struct {
	result gjson.Result
	tag    bool
	DataSet
}

type Config struct {
	MeasurementName     string `toml:"measurement_name"`      // OPTIONAL
	MeasurementNamePath string `toml:"measurement_name_path"` // OPTIONAL
	TimestampPath       string `toml:"timestamp_path"`        // OPTIONAL
	TimestampFormat     string `toml:"timestamp_format"`      // OPTIONAL, but REQUIRED when timestamp_path is defined
	TimestampTimezone   string `toml:"timestamp_timezone"`    // OPTIONAL, but REQUIRES timestamp_path

	Fields      []DataSet
	Tags        []DataSet
	JSONObjects []JSONObject
}

type DataSet struct {
	Path   string `toml:"path"`   // REQUIRED
	Type   string `toml:"type"`   // OPTIONAL, can't be set for tags they will always be a string
	Rename string `toml:"rename"` // OPTIONAL
}

type JSONObject struct {
	Path               string            `toml:"path"`                 // REQUIRED
	TimestampKey       string            `toml:"timestamp_key"`        // OPTIONAL
	TimestampFormat    string            `toml:"timestamp_format"`     // OPTIONAL, but REQUIRED when timestamp_path is defined
	TimestampTimezone  string            `toml:"timestamp_timezone"`   // OPTIONAL, but REQUIRES timestamp_path
	Renames            map[string]string `toml:"renames"`              // OPTIONAL
	Fields             map[string]string `toml:"fields"`               // OPTIONAL
	Tags               []string          `toml:"tags"`                 // OPTIONAL
	IncludedKeys       []string          `toml:"included_keys"`        // OPTIONAL
	ExcludedKeys       []string          `toml:"excluded_keys"`        // OPTIONAL
	DisablePrependKeys bool              `toml:"disable_prepend_keys"` // OPTIONAL
	FieldPaths         []DataSet         // OPTIONAL
	TagPaths           []DataSet         // OPTIONAL
}

type MetricNode struct {
	ParentIndex       int
	OutputName        string
	SetName           string
	Tag               bool
	DesiredType       string      // Can be "int", "uint", "float", "bool", "string"
	IncludeCollection *PathResult // If set to true, it should be auto included

	Metric telegraf.Metric
	gjson.Result
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	p.InputJSON = input
	// Only valid JSON is supported
	if !gjson.Valid(string(p.InputJSON)) {
		return nil, fmt.Errorf("Invalid JSON provided, unable to parse")
	}

	var metrics []telegraf.Metric

	for _, c := range p.Configs {
		// Measurement name configuration
		p.measurementName = c.MeasurementName
		if c.MeasurementNamePath != "" {
			result := gjson.GetBytes(p.InputJSON, c.MeasurementNamePath)
			if !result.IsArray() && !result.IsObject() {
				p.measurementName = result.String()
			}
		}

		// Timestamp configuration
		p.Timestamp = time.Now()
		if c.TimestampPath != "" {
			result := gjson.GetBytes(p.InputJSON, c.TimestampPath)
			if !result.IsArray() && !result.IsObject() {
				if c.TimestampFormat == "" {
					err := fmt.Errorf("use of 'timestamp_query' requires 'timestamp_format'")
					return nil, err
				}

				var err error
				p.Timestamp, err = internal.ParseTimestamp(c.TimestampFormat, result.Value(), c.TimestampTimezone)
				if err != nil {
					return nil, err
				}
			}
		}

		fields, err := p.processMetric(c.Fields, false)
		if err != nil {
			return nil, err
		}

		tags, err := p.processMetric(c.Tags, true)
		if err != nil {
			return nil, err
		}

		objects, err := p.processObjects(c.JSONObjects)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, cartesianProduct(tags, fields)...)

		if len(objects) != 0 && len(metrics) != 0 {
			metrics = append(metrics, cartesianProduct(objects, metrics)...)
		} else {
			metrics = append(metrics, objects...)
		}
	}

	for k, v := range p.DefaultTags {
		for _, t := range metrics {
			t.AddTag(k, v)
		}
	}

	return metrics, nil
}

// processMetric will iterate over all 'field' or 'tag' configs and create metrics for each
// A field/tag can either be a single value or an array of values, each resulting in its own metric
// For multiple configs, a set of metrics is created from the cartesian product of each separate config
func (p *Parser) processMetric(data []DataSet, tag bool) ([]telegraf.Metric, error) {
	if len(data) == 0 {
		return nil, nil
	}

	p.iterateObjects = false
	var metrics [][]telegraf.Metric

	for _, c := range data {
		if c.Path == "" {
			return nil, fmt.Errorf("GJSON path is required")
		}
		result := gjson.GetBytes(p.InputJSON, c.Path)

		if result.IsObject() {
			p.Log.Debugf("Found object in the path: %s, ignoring it please use 'object' to gather metrics from objects", c.Path)
			continue
		}

		setName := c.Rename
		// Default to the last path word, should be the upper key name
		if setName == "" {
			s := strings.Split(c.Path, ".")
			setName = s[len(s)-1]
		}
		setName = strings.ReplaceAll(setName, " ", "_")

		mNode := MetricNode{
			OutputName:  setName,
			SetName:     setName,
			DesiredType: c.Type,
			Tag:         tag,
			Metric: metric.New(
				p.measurementName,
				map[string]string{},
				map[string]interface{}{},
				p.Timestamp,
			),
			Result: result,
		}

		// Expand all array's and nested arrays into separate metrics
		nodes, err := p.expandArray(mNode)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, nodes)
	}

	for i := 1; i < len(metrics); i++ {
		metrics[i] = cartesianProduct(metrics[i-1], metrics[i])
	}

	return metrics[len(metrics)-1], nil
}

func cartesianProduct(a, b []telegraf.Metric) []telegraf.Metric {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
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
func (p *Parser) expandArray(result MetricNode) ([]telegraf.Metric, error) {
	var results []telegraf.Metric

	if result.IsObject() {
		if !p.iterateObjects {
			p.Log.Debugf("Found object in query ignoring it please use 'object' to gather metrics from objects")
			return results, nil
		}
		if result.IncludeCollection == nil && (len(p.currentSettings.FieldPaths) > 0 || len(p.currentSettings.TagPaths) > 0) {
			result.IncludeCollection = p.existsInpathResults(result.Index, result.Raw)
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
		if result.IncludeCollection == nil && (len(p.currentSettings.FieldPaths) > 0 || len(p.currentSettings.TagPaths) > 0) {
			result.IncludeCollection = p.existsInpathResults(result.Index, result.Raw)
		}
		result.ForEach(func(_, val gjson.Result) bool {
			m := metric.New(
				p.measurementName,
				map[string]string{},
				map[string]interface{}{},
				p.Timestamp,
			)
			if val.IsObject() {
				if p.iterateObjects {
					n := result
					n.ParentIndex += val.Index
					n.Metric = m
					n.Result = val
					if n.IncludeCollection == nil && (len(p.currentSettings.FieldPaths) > 0 || len(p.currentSettings.TagPaths) > 0) {
						n.IncludeCollection = p.existsInpathResults(n.Index, n.Raw)
					}
					r, err := p.combineObject(n)
					if err != nil {
						return false
					}

					results = append(results, r...)
				} else {
					p.Log.Debugf("Found object in query ignoring it please use 'object' to gather metrics from objects")
				}
				if len(results) != 0 {
					for _, newResult := range results {
						mergeMetric(result.Metric, newResult)
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
			n := result
			n.ParentIndex += val.Index
			n.Metric = m
			n.Result = val
			if n.IncludeCollection == nil && (len(p.currentSettings.FieldPaths) > 0 || len(p.currentSettings.TagPaths) > 0) {
				n.IncludeCollection = p.existsInpathResults(n.Index, n.Raw)
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
		if result.SetName == p.currentSettings.TimestampKey {
			if p.currentSettings.TimestampFormat == "" {
				err := fmt.Errorf("use of 'timestamp_query' requires 'timestamp_format'")
				return nil, err
			}
			timestamp, err := internal.ParseTimestamp(p.currentSettings.TimestampFormat, result.Value(), p.currentSettings.TimestampTimezone)
			if err != nil {
				return nil, err
			}
			result.Metric.SetTime(timestamp)
		} else {
			switch result.Value().(type) {
			case nil: // Ignore JSON values that are set as null
			default:
				outputName := result.OutputName
				desiredType := result.DesiredType

				if len(p.currentSettings.FieldPaths) > 0 || len(p.currentSettings.TagPaths) > 0 {
					var pathResult *PathResult
					if result.IncludeCollection != nil {
						pathResult = result.IncludeCollection
					} else {
						pathResult = p.existsInpathResults(result.ParentIndex, result.Raw)
					}
					if pathResult == nil {
						return results, nil
					}
					if pathResult.tag {
						result.Tag = true
					}

					if !pathResult.tag {
						desiredType = pathResult.Type
					}

					if pathResult.Rename != "" {
						outputName = pathResult.Rename
					}
				}

				if result.Tag {
					desiredType = "string"
				}
				v, err := p.convertType(result.Result, desiredType, result.SetName)
				if err != nil {
					return nil, err
				}
				if result.Tag {
					result.Metric.AddTag(outputName, v.(string))
				} else {
					result.Metric.AddField(outputName, v)
				}
			}
		}

		results = append(results, result.Metric)
	}

	return results, nil
}

func (p *Parser) existsInpathResults(index int, raw string) *PathResult {
	for _, f := range p.pathResults {
		if f.result.Index == 0 {
			for _, i := range f.result.Indexes {
				if i == index {
					return &f
				}
			}
		} else if f.result.Index == index {
			return &f
		}
	}
	return nil
}

// processObjects will iterate over all 'object' configs and create metrics for each
func (p *Parser) processObjects(objects []JSONObject) ([]telegraf.Metric, error) {
	p.iterateObjects = true
	var t []telegraf.Metric
	for _, c := range objects {
		p.currentSettings = c

		if c.Path == "" {
			return nil, fmt.Errorf("GJSON path is required")
		}
		result := gjson.GetBytes(p.InputJSON, c.Path)

		// hastag doesn't return index! idea: replace all hastags with index, and find lenght of array
		scopedJSON := []byte(result.Raw)
		for _, f := range c.FieldPaths {
			var r PathResult
			r.result = gjson.GetBytes(scopedJSON, f.Path)
			r.DataSet = f
			p.pathResults = append(p.pathResults, r)
		}

		for _, f := range c.TagPaths {
			var r PathResult
			r.result = gjson.GetBytes(scopedJSON, f.Path)
			r.DataSet = f
			r.tag = true
			p.pathResults = append(p.pathResults, r)
		}

		if result.Type == gjson.Null {
			return nil, fmt.Errorf("GJSON Path returned null")
		}

		rootObject := MetricNode{
			ParentIndex: 0,
			Metric: metric.New(
				p.measurementName,
				map[string]string{},
				map[string]interface{}{},
				p.Timestamp,
			),
			Result: result,
		}
		metrics, err := p.expandArray(rootObject)
		if err != nil {
			return nil, err
		}
		t = append(t, metrics...)
	}

	return t, nil
}

// combineObject will add all fields/tags to a single metric
// If the object has multiple array's as elements it won't comine those, they will remain separate metrics
func (p *Parser) combineObject(result MetricNode) ([]telegraf.Metric, error) {
	var results []telegraf.Metric
	if result.IsArray() || result.IsObject() {
		var err error
		result.ForEach(func(key, val gjson.Result) bool {
			// Determine if field/tag set name is configured
			var setName string
			if result.SetName != "" {
				setName = result.SetName + "_" + strings.ReplaceAll(key.String(), " ", "_")
			} else {
				setName = strings.ReplaceAll(key.String(), " ", "_")
			}

			if p.isExcluded(setName) || !p.isIncluded(setName, val) {
				return true
			}

			var outputName string
			if p.currentSettings.DisablePrependKeys {
				outputName = strings.ReplaceAll(key.String(), " ", "_")
			} else {
				outputName = setName
			}
			for k, n := range p.currentSettings.Renames {
				if k == setName {
					outputName = n
					break
				}
			}

			arrayNode := result
			arrayNode.ParentIndex += val.Index
			arrayNode.OutputName = outputName
			arrayNode.SetName = setName
			arrayNode.Result = val

			for k, t := range p.currentSettings.Fields {
				if setName == k {
					arrayNode.DesiredType = t
					break
				}
			}

			tag := false
			for _, t := range p.currentSettings.Tags {
				if setName == t {
					tag = true
					break
				}
			}

			arrayNode.Tag = tag

			if val.IsObject() {
				results, err = p.combineObject(arrayNode)
				if err != nil {
					return false
				}
			} else {
				r, err := p.expandArray(arrayNode)
				if err != nil {
					return false
				}
				results = cartesianProduct(results, r)
			}

			return true
		})

		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (p *Parser) isIncluded(key string, val gjson.Result) bool {
	if len(p.currentSettings.IncludedKeys) == 0 {
		return true
	}
	// automatically adds tags to included_keys so it does NOT have to be repeated in the config
	allKeys := append(p.currentSettings.IncludedKeys, p.currentSettings.Tags...)
	for _, i := range allKeys {
		if i == key {
			return true
		}
		if val.IsArray() || val.IsObject() {
			// Check if the included key is a sub element
			if strings.HasPrefix(i, key) {
				return true
			}
		}
	}
	return false
}

func (p *Parser) isExcluded(key string) bool {
	for _, i := range p.currentSettings.ExcludedKeys {
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
func (p *Parser) convertType(input gjson.Result, desiredType string, name string) (interface{}, error) {
	switch inputType := input.Value().(type) {
	case string:
		if desiredType != "string" {
			switch desiredType {
			case "uint":
				r, err := strconv.ParseUint(inputType, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type uint: %v", name, err)
				}
				return r, nil
			case "int":
				r, err := strconv.ParseInt(inputType, 10, 64)
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
		switch desiredType {
		case "string":
			return strconv.FormatBool(inputType), nil
		case "int":
			if inputType {
				return int64(1), nil
			}

			return int64(0), nil
		case "uint":
			if inputType {
				return uint64(1), nil
			}

			return uint64(0), nil
		}
	case float64:
		if desiredType != "float" {
			switch desiredType {
			case "string":
				return fmt.Sprint(inputType), nil
			case "int":
				return input.Int(), nil
			case "uint":
				return input.Uint(), nil
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

	return input.Value(), nil
}
