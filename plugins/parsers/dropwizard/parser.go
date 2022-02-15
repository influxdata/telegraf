package dropwizard

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/templating"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

type TimeFunc func() time.Time

// Parser parses json inputs containing dropwizard metrics,
// either top-level or embedded inside a json field.
// This parser is using gjson for retrieving paths within the json file.
type parser struct {
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

	Log telegraf.Logger `toml:"-"`

	separator      string
	templateEngine *templating.Engine

	timeFunc TimeFunc

	// seriesParser parses line protocol measurement + tags
	seriesParser *influx.Parser
}

func NewParser() *parser {
	handler := influx.NewMetricHandler()
	seriesParser := influx.NewSeriesParser(handler)

	parser := &parser{
		timeFunc:     time.Now,
		seriesParser: seriesParser,
	}
	return parser
}

// Parse parses the input bytes to an array of metrics
func (p *parser) Parse(buf []byte) ([]telegraf.Metric, error) {
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

func (p *parser) SetTemplates(separator string, templates []string) error {
	if len(templates) == 0 {
		p.templateEngine = nil
		return nil
	}

	defaultTemplate, err := templating.NewDefaultTemplateWithPattern("measurement*")
	if err != nil {
		return err
	}

	templateEngine, err := templating.NewEngine(separator, defaultTemplate, templates)
	if err != nil {
		return err
	}

	p.separator = separator
	p.templateEngine = templateEngine
	return nil
}

// ParseLine is not supported by the dropwizard format
func (p *parser) ParseLine(line string) (telegraf.Metric, error) {
	return nil, fmt.Errorf("ParseLine not supported: %s, for data format: dropwizard", line)
}

// SetDefaultTags sets the default tags
func (p *parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *parser) readTags(buf []byte) map[string]string {
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
			p.Log.Warnf("Failed to parse tags from JSON path '%s': %s\n", p.TagsPath, err)
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

func (p *parser) parseTime(buf []byte) (time.Time, error) {
	if p.TimePath != "" {
		timeFormat := p.TimeFormat
		if timeFormat == "" {
			timeFormat = time.RFC3339
		}
		timeString := gjson.GetBytes(buf, p.TimePath).String()
		if timeString == "" {
			err := fmt.Errorf("time not found in JSON path %s", p.TimePath)
			return p.timeFunc(), err
		}
		t, err := time.Parse(timeFormat, timeString)
		if err != nil {
			err = fmt.Errorf("time %s cannot be parsed with format %s, %s", timeString, timeFormat, err)
			return p.timeFunc(), err
		}
		return t.UTC(), nil
	}
	return p.timeFunc(), nil
}

func (p *parser) unmarshalMetrics(buf []byte) (map[string]interface{}, error) {
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

func (p *parser) readDWMetrics(metricType string, dwms interface{}, metrics []telegraf.Metric, tm time.Time) []telegraf.Metric {
	if dwmsTyped, ok := dwms.(map[string]interface{}); ok {
		for dwmName, dwmFields := range dwmsTyped {
			measurementName := dwmName
			tags := make(map[string]string)
			fieldPrefix := ""
			if p.templateEngine != nil {
				measurementName, tags, fieldPrefix, _ = p.templateEngine.Apply(dwmName)
				if len(fieldPrefix) > 0 {
					fieldPrefix = fmt.Sprintf("%s%s", fieldPrefix, p.separator)
				}
			}

			parsed, err := p.seriesParser.Parse([]byte(measurementName))
			var m telegraf.Metric
			if err != nil || len(parsed) != 1 {
				m = metric.New(measurementName, map[string]string{}, map[string]interface{}{}, tm)
			} else {
				m = parsed[0]
				m.SetTime(tm)
			}

			m.AddTag("metric_type", metricType)
			for k, v := range tags {
				m.AddTag(k, v)
			}

			if fields, ok := dwmFields.(map[string]interface{}); ok {
				for k, v := range fields {
					switch v := v.(type) {
					case float64, string, bool:
						m.AddField(fieldPrefix+k, v)
					default:
						// ignore
					}
				}
			}

			metrics = append(metrics, m)
		}
	}

	return metrics
}

func (p *parser) SetTimeFunc(f TimeFunc) {
	p.timeFunc = f
}
