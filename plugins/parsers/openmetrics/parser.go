package openmetrics

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	IgnoreTimestamp bool              `toml:"openmetrics_ignore_timestamp"`
	MetricVersion   int               `toml:"openmetrics_metric_version"`
	Header          http.Header       `toml:"-"` // set by the openmetrics input
	DefaultTags     map[string]string `toml:"-"`
	Log             telegraf.Logger   `toml:"-"`
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) Parse(data []byte) ([]telegraf.Metric, error) {
	// Make sure we have a finishing newline but no trailing one
	data = bytes.TrimPrefix(data, []byte("\n"))
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, []byte("\n")...)
	}

	// // Determine the metric transport-type derived from the response header and
	// // create a matching decoder.
	// format := expfmt.ResponseFormat(p.Header)
	// if format == expfmt.FmtUnknown {
	// 	p.Log.Warnf("Unknown format %q... Trying to continue...", p.Header.Get("Content-Type"))
	// }
	// decoder := expfmt.NewDecoder(buf, format)

	metricFamilies, err := TextToMetricFamilies(data)
	if err != nil {
		return nil, err
	}

	// // Decode the input data into prometheus metrics
	var metrics []telegraf.Metric
	// for {
	// 	var mf dto.MetricFamily
	// 	if err := decoder.Decode(&mf); err != nil {
	// 		if errors.Is(err, io.EOF) {
	// 			break
	// 		}
	// 		return nil, fmt.Errorf("decoding response failed: %w", err)
	// 	}

	for _, mf := range metricFamilies {
		switch p.MetricVersion {
		case 0, 2:
			metrics = append(metrics, p.extractMetricsV2(mf)...)
		case 1:
			metrics = append(metrics, p.extractMetricsV1(mf)...)
		default:
			return nil, fmt.Errorf("unknown prometheus metric version %d", p.MetricVersion)
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
		return nil, fmt.Errorf("no metrics in line")
	}

	if len(metrics) > 1 {
		return nil, fmt.Errorf("more than one metric in line")
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
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{}
		},
	)
}
