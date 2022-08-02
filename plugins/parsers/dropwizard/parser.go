package dropwizard

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/templating"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

// Parser parses json inputs containing dropwizard metrics,
// either top-level or embedded inside a json field.
// This parser is using gjson for retrieving paths within the json file.
type Parser struct {
	MetricRegistryPath string            `toml:"dropwizard_metric_registry_path"`
	TimePath           string            `toml:"dropwizard_time_path"`
	TimeFormat         string            `toml:"dropwizard_time_format"`
	TagsPath           string            `toml:"dropwizard_tags_path"`
	TagPathsMap        map[string]string `toml:"dropwizard_tag_paths_map"`
	Separator          string            `toml:"separator"`
	Templates          []string          `toml:"templates"`
	DefaultTags        map[string]string `toml:"-"`
	Log                telegraf.Logger   `toml:"-"`

	templateEngine *templating.Engine

	// seriesParser parses line protocol measurement + tags
	seriesParser *influx.Parser
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

func (p *Parser) parseTime(buf []byte) (time.Time, error) {
	if p.TimePath != "" {
		timeFormat := p.TimeFormat
		if timeFormat == "" {
			timeFormat = time.RFC3339
		}
		timeString := gjson.GetBytes(buf, p.TimePath).String()
		if timeString == "" {
			return time.Time{}, fmt.Errorf("time not found in JSON path %s", p.TimePath)
		}
		t, err := time.Parse(timeFormat, timeString)
		if err != nil {
			return time.Time{}, fmt.Errorf("time %s cannot be parsed with format %s, %s", timeString, timeFormat, err)
		}
		return t.UTC(), nil
	}
	return time.Now(), nil
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
	if dwmsTyped, ok := dwms.(map[string]interface{}); ok {
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

func (p *Parser) Init() error {
	parser := &influx.Parser{
		Type: "series",
	}
	err := parser.Init()
	if err != nil {
		return err
	}
	p.seriesParser = parser

	if len(p.Templates) != 0 {
		defaultTemplate, err := templating.NewDefaultTemplateWithPattern("measurement*")
		if err != nil {
			return err
		}

		templateEngine, err := templating.NewEngine(p.Separator, defaultTemplate, p.Templates)
		if err != nil {
			return err
		}
		p.templateEngine = templateEngine
	}

	return nil
}

func init() {
	parsers.Add("dropwizard",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{}
		})
}
