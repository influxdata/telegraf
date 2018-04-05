package dropwizard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/templating"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/tidwall/gjson"
)

var fieldEscaper = strings.NewReplacer("\\", "\\\\", "\"", "\\\"")
var keyEscaper = strings.NewReplacer(" ", "\\ ", ",", "\\,", "=", "\\=")

// Parser parses json inputs containing dropwizard metrics,
// either top-level or embedded inside a json field.
// This parser is using gjon for retrieving paths within the json file.
type Parser struct {

	// an optional json path containing the metric registry object
	// if left empty, the whole json object is parsed as a metric registry
	MetricRegistryPath string

	// an optional json path containing the default time of the metrics
	// if left empty, or if cannot be parsed the current processing time is used as the time of the metrics
	TimePath string

	// time format to use for parsing the time field
	// defaults to time.RFC3339
	TimeFormat string

	// an optional json path pointing to a json object with tag key/value pairs
	// takes precedence over TagPathsMap
	TagsPath string

	// an optional map containing tag names as keys and json paths to retrieve the tag values from as values
	// used if TagsPath is empty or doesn't return any tags
	TagPathsMap map[string]string

	// an optional map of default tags to use for metrics
	DefaultTags map[string]string

	// templating configuration
	Separator string
	Templates []string

	templateEngine *templating.Engine
}

// Parse parses the input bytes to an array of metrics
func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {

	metrics := make([]telegraf.Metric, 0)

	metricTime, err := p.parseTime(buf)
	if err != nil {
		return nil, err
	}
	dwr, err := p.unmarshalMetrics(buf)
	if err != nil {
		return nil, err
	}

	metrics = p.readDWMetrics("counter", dwr["counters"], metrics, metricTime)
	metrics = p.readDWMetrics("meter", dwr["meters"], metrics, metricTime)
	metrics = p.readDWMetrics("gauge", dwr["gauges"], metrics, metricTime)
	metrics = p.readDWMetrics("histogram", dwr["histograms"], metrics, metricTime)
	metrics = p.readDWMetrics("timer", dwr["timers"], metrics, metricTime)

	jsonTags := p.readTags(buf)

	// fill json tags first
	if len(jsonTags) > 0 {
		for _, m := range metrics {
			for k, v := range jsonTags {
				// only set the tag if it doesn't already exist:
				if !m.HasTag(k) {
					m.AddTag(k, v)
				}
			}
		}
	}
	// fill default tags last
	if len(p.DefaultTags) > 0 {
		for _, m := range metrics {
			for k, v := range p.DefaultTags {
				// only set the default tag if it doesn't already exist:
				if !m.HasTag(k) {
					m.AddTag(k, v)
				}
			}
		}
	}

	return metrics, nil
}

// InitTemplating initializes the templating support
func (p *Parser) InitTemplating() error {
	if len(p.Templates) > 0 {
		defaultTemplate, _ := templating.NewDefaultTemplateWithPattern("measurement*")
		templateEngine, err := templating.NewEngine(p.Separator, defaultTemplate, p.Templates)
		p.templateEngine = templateEngine
		return err
	}
	return nil
}

// ParseLine is not supported by the dropwizard format
func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	return nil, fmt.Errorf("ParseLine not supported: %s, for data format: dropwizard", line)
}

// SetDefaultTags sets the default tags
func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) readTags(buf []byte) map[string]string {

	if p.TagsPath != "" {
		var tagsBytes []byte
		tagsResult := gjson.GetBytes(buf, p.TagsPath)
		if tagsResult.Index > 0 {
			tagsBytes = buf[tagsResult.Index : tagsResult.Index+len(tagsResult.Raw)]
		} else {
			tagsBytes = []byte(tagsResult.Raw)
		}
		var tags map[string]string
		err := json.Unmarshal(tagsBytes, &tags)
		if err != nil {
			log.Printf("W! failed to parse tags from JSON path '%s': %s\n", p.TagsPath, err)
		} else if len(tags) > 0 {
			return tags
		}
	}

	tags := make(map[string]string)
	for tagKey, jsonPath := range p.TagPathsMap {
		tags[tagKey] = gjson.GetBytes(buf, jsonPath).String()
	}
	return tags
}

func (p *Parser) parseTime(buf []byte) (time.Time, error) {

	if p.TimePath != "" {
		timeFormat := p.TimeFormat
		if timeFormat == "" {
			timeFormat = time.RFC3339
		}
		timeString := gjson.GetBytes(buf, p.TimePath).String()
		if timeString == "" {
			err := fmt.Errorf("time not found in JSON path %s", p.TimePath)
			return time.Now().UTC(), err
		}
		t, err := time.Parse(timeFormat, timeString)
		if err != nil {
			err = fmt.Errorf("time %s cannot be parsed with format %s, %s", timeString, timeFormat, err)
			return time.Now().UTC(), err
		}
		return t.UTC(), nil
	}
	return time.Now().UTC(), nil
}

func (p *Parser) unmarshalMetrics(buf []byte) (map[string]interface{}, error) {

	var registryBytes []byte
	if p.MetricRegistryPath != "" {
		regResult := gjson.GetBytes(buf, p.MetricRegistryPath)
		if regResult.Index > 0 {
			registryBytes = buf[regResult.Index : regResult.Index+len(regResult.Raw)]
		} else {
			registryBytes = []byte(regResult.Raw)
		}
		if len(registryBytes) == 0 {
			err := fmt.Errorf("metric registry not found in JSON path %s", p.MetricRegistryPath)
			return nil, err
		}
	} else {
		registryBytes = buf
	}
	var jsonOut map[string]interface{}
	err := json.Unmarshal(registryBytes, &jsonOut)
	if err != nil {
		err = fmt.Errorf("unable to parse dropwizard metric registry from JSON document, %s", err)
		return nil, err
	}
	return jsonOut, nil
}

func (p *Parser) readDWMetrics(metricType string, dwms interface{}, metrics []telegraf.Metric, tm time.Time) []telegraf.Metric {

	switch dwmsTyped := dwms.(type) {
	case map[string]interface{}:
		var metricsBuffer bytes.Buffer
		for dwmName, dwmFields := range dwmsTyped {
			measurementName := dwmName
			tags := make(map[string]string)
			fieldPrefix := ""
			if p.templateEngine != nil {
				measurementName, tags, fieldPrefix, _ = p.templateEngine.Apply(dwmName)
				if len(fieldPrefix) > 0 {
					fieldPrefix = fmt.Sprintf("%s%s", fieldPrefix, p.Separator)
				}
			}
			tags["metric_type"] = metricType

			measurementWithTags := measurementName
			for tagName, tagValue := range tags {
				tagKeyValue := fmt.Sprintf("%s=%s", keyEscaper.Replace(tagName), keyEscaper.Replace(tagValue))
				measurementWithTags = fmt.Sprintf("%s,%s", measurementWithTags, tagKeyValue)
			}

			fields := make([]string, 0)
			switch t := dwmFields.(type) {
			case map[string]interface{}: // json object
				for fieldName, fieldValue := range t {
					key := keyEscaper.Replace(fieldPrefix + fieldName)
					switch v := fieldValue.(type) {
					case float64:
						fields = append(fields, fmt.Sprintf("%s=%f", key, v))
					case string:
						fields = append(fields, fmt.Sprintf("%s=\"%s\"", key, fieldEscaper.Replace(v)))
					case bool:
						fields = append(fields, fmt.Sprintf("%s=%t", key, v))
					default: // ignore
					}
				}
			default: // ignore
			}

			metricsBuffer.WriteString(fmt.Sprintf("%s,metric_type=%s ", measurementWithTags, metricType))
			metricsBuffer.WriteString(strings.Join(fields, ","))
			metricsBuffer.WriteString("\n")
		}

		handler := influx.NewMetricHandler()
		handler.SetTimeFunc(func() time.Time { return tm })
		parser := influx.NewParser(handler)
		newMetrics, err := parser.Parse(metricsBuffer.Bytes())
		if err != nil {
			log.Printf("W! failed to create metric of type '%s': %s\n", metricType, err)
		}

		return append(metrics, newMetrics...)
	default:
		return metrics
	}

}

func arraymap(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
