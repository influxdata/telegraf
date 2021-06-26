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
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
	Timestamp   time.Time

	measurementName string

	iterateObjects   bool
	currentSettings  JSONObject
	fieldPathResults []PathResult
	tagPathResults   []PathResult
}

type PathResult struct {
	result gjson.Result
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
	ParentIndex int
	OutputName  string
	SetName     string
	Tag         bool
	DesiredType string // Can be "int", "uint", "float", "bool", "string"

	Metric telegraf.Metric
	gjson.Result
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	// Only valid JSON is supported
	if !gjson.Valid(string(input)) {
		return nil, fmt.Errorf("Invalid JSON provided, unable to parse")
	}

	var metrics []telegraf.Metric

	for _, c := range p.Configs {
		// Measurement name configuration
		p.measurementName = c.MeasurementName
		if c.MeasurementNamePath != "" {
			result := gjson.GetBytes(input, c.MeasurementNamePath)
			if !result.IsArray() && !result.IsObject() {
				p.measurementName = result.String()
			}
		}

		// Timestamp configuration
		p.Timestamp = time.Now()
		if c.TimestampPath != "" {
			result := gjson.GetBytes(input, c.TimestampPath)
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

		fields, err := p.processMetric(c.Fields, input, false)
		if err != nil {
			return nil, err
		}

		tags, err := p.processMetric(c.Tags, input, true)
		if err != nil {
			return nil, err
		}

		objects, err := p.processObjects(c.JSONObjects, input)
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
func (p *Parser) processMetric(data []DataSet, input []byte, tag bool) ([]telegraf.Metric, error) {
	if len(data) == 0 {
		return nil, nil
	}

	p.iterateObjects = false
	var metrics [][]telegraf.Metric

	for _, c := range data {
		if c.Path == "" {
			return nil, fmt.Errorf("GJSON path is required")
		}
		result := gjson.GetBytes(input, c.Path)

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
				p.Timestamp,
			)

			if val.IsObject() {
				if p.iterateObjects {
					n := MetricNode{
						ParentIndex: result.ParentIndex + val.Index,
						SetName:     result.SetName,
						Metric:      m,
						Result:      val,
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
			n := MetricNode{
				ParentIndex: result.ParentIndex + val.Index,
				Tag:         result.Tag,
				DesiredType: result.DesiredType,
				OutputName:  result.OutputName,
				SetName:     result.SetName,
				Metric:      m,
				Result:      val,
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
				fieldDataSet := existsInPathResults(result.ParentIndex, p.fieldPathResults)
				tagDataSet := existsInPathResults(result.ParentIndex, p.tagPathResults)
				if fieldDataSet == nil && tagDataSet == nil {
					return results, nil
				}
				outputName := result.OutputName
				if fieldDataSet == nil && tagDataSet != nil {
					result.Tag = true
					if tagDataSet.Rename != "" {
						outputName = tagDataSet.Rename
					}
				}

				if result.Tag {
					result.DesiredType = "string"
				}
				v, err := p.convertType(result.Result, result.DesiredType, result.SetName)
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

func existsInPathResults(index int, indexList []PathResult) *DataSet {
	for _, f := range indexList {
		if f.result.Index == index {
			return &f.DataSet
		}
	}
	return nil
}

// processObjects will iterate over all 'object' configs and create metrics for each
func (p *Parser) processObjects(objects []JSONObject, input []byte) ([]telegraf.Metric, error) {
	p.iterateObjects = true
	var t []telegraf.Metric
	for _, c := range objects {
		p.currentSettings = c

		if c.Path == "" {
			return nil, fmt.Errorf("GJSON path is required")
		}
		result := gjson.GetBytes(input, c.Path)

		// hastag doesn't return index! idea: replace all hastags with index, and find lenght of array
		scopedJSON := []byte(result.Raw)
		for _, f := range c.FieldPaths {
			var r PathResult
			r.result = gjson.GetBytes(scopedJSON, f.Path)
			r.DataSet = f
			p.fieldPathResults = append(p.fieldPathResults, r)
		}

		for _, f := range c.TagPaths {
			var r PathResult
			r.result = gjson.GetBytes(scopedJSON, f.Path)
			r.DataSet = f
			p.tagPathResults = append(p.tagPathResults, r)
		}

		// resultTest := gjson.GetBytes(input, "root.station")
		// subResult := gjson.GetBytes(input, "root.station.0.etd.0.estimate.0.minutes")

		// fmt.Println("key", subResult.Str, "val", subResult.Raw, "val index", subResult.Index)
		// fmt.Println(string(input[subResult.Index : subResult.Index+len(subResult.Raw)]))

		// // fmt.Println(string(input[10 : 10+len(result.Raw)]))
		// resultTest.ForEach(func(key, val gjson.Result) bool {
		// 	// fmt.Println("key", key.Str, "val", val.Raw, "val index", val.Index)
		// 	// fmt.Println(string(input[val.Index : val.Index+len(val.Raw)]))
		// 	// if val.IsArray() {
		// 	// 	val.ForEach(func(subkey, subval gjson.Result) bool {
		// 	// 		fmt.Println("key", subkey.Str, "val", subval.Raw, "val index", subval.Index)
		// 	// 		newIndex := subval.Index + val.Index
		// 	// 		fmt.Println(newIndex)
		// 	// 		fmt.Println(string(input[newIndex : newIndex+len(subval.Raw)]))
		// 	if val.IsObject() {
		// 		val.ForEach(func(subsubkey, subsubval gjson.Result) bool {
		// 			fmt.Println("key", subsubkey.Str, "val", subsubval.Raw, "val index", subsubval.Index)
		// 			newnewIndex := subsubval.Index + (val.Index + resultTest.Index)
		// 			fmt.Println(newnewIndex)
		// 			fmt.Println(string(input[newnewIndex : newnewIndex+len(subsubval.Raw)]))
		// 			return true
		// 		})
		// 	}
		// 	// 		return true
		// 	// 	})
		// 	// }
		// 	return true
		// })

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

			arrayNode := MetricNode{
				ParentIndex: result.ParentIndex + val.Index,
				DesiredType: result.DesiredType,
				Tag:         result.Tag,
				OutputName:  outputName,
				SetName:     setName,
				Metric:      result.Metric,
				Result:      val,
			}

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
