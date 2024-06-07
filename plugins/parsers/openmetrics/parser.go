package openmetrics

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
)

func AcceptsContent(header http.Header) bool {
	contentType := header.Get("Content-Type")
	if contentType == "" {
		return false
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	switch mediaType {
	case expfmt.OpenMetricsType:
		return true
	case "application/openmetrics-protobuf":
		return params["version"] == "1.0.0"
	}
	return false
}

type Parser struct {
	IgnoreTimestamp bool              `toml:"openmetrics_ignore_timestamp"`
	MetricVersion   int               `toml:"openmetrics_metric_version"`
	Header          http.Header       `toml:"-"` // set by the input plugin
	DefaultTags     map[string]string `toml:"-"`
	Log             telegraf.Logger   `toml:"-"`
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) Parse(data []byte) ([]telegraf.Metric, error) {
	// Determine the metric transport-type derived from the response header
	contentType := p.Header.Get("Content-Type")
	var mediaType string
	var params map[string]string
	if contentType == "" {
		// Fallback to text type if no content-type is given
		mediaType = expfmt.OpenMetricsType
	} else {
		var err error
		mediaType, params, err = mime.ParseMediaType(contentType)
		if err != nil {
			return nil, fmt.Errorf("unknown media-type in %q", contentType)
		}
	}

	// Parse the raw data into OpenMetrics metrics
	var metricFamilies []*MetricFamily
	switch mediaType {
	case expfmt.OpenMetricsType:
		// Make sure we have a finishing newline but no trailing one
		data = bytes.TrimPrefix(data, []byte("\n"))
		if !bytes.HasSuffix(data, []byte("\n")) {
			data = append(data, []byte("\n")...)
		}

		var err error
		metricFamilies, err = TextToMetricFamilies(data)
		if err != nil {
			return nil, fmt.Errorf("parsing text format failed: %w", err)
		}
	case "application/openmetrics-protobuf":
		if version := params["version"]; version != "1.0.0" {
			return nil, fmt.Errorf("unsupported binary version %q", version)
		}
		var metricSet MetricSet
		if err := proto.Unmarshal(data, &metricSet); err != nil {
			return nil, fmt.Errorf("parsing binary format failed: %w", err)
		}
		metricFamilies = metricSet.GetMetricFamilies()
	}

	// Convert the OpenMetrics metrics into Telegraf metrics
	var metrics []telegraf.Metric
	for _, mf := range metricFamilies {
		switch p.MetricVersion {
		case 0, 2:
			metrics = append(metrics, p.extractMetricsV2(mf)...)
		case 1:
			metrics = append(metrics, p.extractMetricsV1(mf)...)
		default:
			return nil, fmt.Errorf("unknown metric version %d", p.MetricVersion)
		}
	}
	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, errors.New("no metrics in line")
	}

	if len(metrics) > 1 {
		return nil, errors.New("more than one metric in line")
	}

	return metrics[0], nil
}

func getTagsFromLabels(m *Metric, defaultTags map[string]string) map[string]string {
	result := make(map[string]string, len(defaultTags)+len(m.Labels))

	for key, value := range defaultTags {
		result[key] = value
	}

	for _, label := range m.Labels {
		if v := label.GetValue(); v != "" {
			result[label.Name] = v
		}
	}

	return result
}

func init() {
	parsers.Add("openmetrics",
		func(string) telegraf.Parser {
			return &Parser{}
		},
	)
}
