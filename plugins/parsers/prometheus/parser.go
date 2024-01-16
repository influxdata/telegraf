package prometheus

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Parser struct {
	IgnoreTimestamp bool              `toml:"prometheus_ignore_timestamp"`
	MetricVersion   int               `toml:"prometheus_metric_version"`
	Header          http.Header       `toml:"-"` // set by the prometheus input
	DefaultTags     map[string]string `toml:"-"`
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) Parse(data []byte) ([]telegraf.Metric, error) {
	// Determine the metric transport-type derived from the response header.
	// If no content-type is given, fallback to the text-format.
	mediatype := "text/plain"
	params := map[string]string{}
	if contentType := p.Header.Get("Content-Type"); contentType != "" {
		var err error
		mediatype, params, err = mime.ParseMediaType(contentType)
		if err != nil {
			if !errors.Is(err, mime.ErrInvalidMediaParameter) {
				return nil, fmt.Errorf("detecting media-type failed: %w", err)
			}
		}
	}

	// Make sure we have a finishing newline but no trailing one
	data = bytes.TrimPrefix(data, []byte("\n"))
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, []byte("\n")...)
	}
	buf := bytes.NewBuffer(data)

	// Decode the input data into prometheus metrics
	var metricFamilies map[string]*dto.MetricFamily
	switch mediatype {
	case "application/vnd.google.protobuf":
		encoding := params["encoding"]
		proto := params["proto"]
		if encoding != "delimited" || proto != "io.prometheus.client.MetricFamily" {
			return nil, fmt.Errorf("unable to decode protobuf with encoding %q and proto %q", encoding, proto)
		}
		metricFamilies = make(map[string]*dto.MetricFamily)
		for {
			mf := &dto.MetricFamily{}
			if _, err := pbutil.ReadDelimited(buf, mf); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, fmt.Errorf("reading protocol-buffer format failed: %w", err)
			}
			metricFamilies[mf.GetName()] = mf
		}
	case "text/plain":
		var parser expfmt.TextParser
		var err error
		metricFamilies, err = parser.TextToMetricFamilies(buf)
		if err != nil {
			return nil, fmt.Errorf("reading text format failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported prometheus format %q", mediatype)
	}

	switch p.MetricVersion {
	case 0, 2:
		return p.extractMetricsV2(metricFamilies), nil
	case 1:
		return p.extractMetricsV1(metricFamilies), nil
	}
	return nil, fmt.Errorf("unknown prometheus metric version %d", p.MetricVersion)
}

func init() {
	parsers.Add("prometheus",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{}
		},
	)
}
