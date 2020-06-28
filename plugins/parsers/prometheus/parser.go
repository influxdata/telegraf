package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"

	v1 "github.com/influxdata/telegraf/plugins/parsers/prometheus/v1"
	v2 "github.com/influxdata/telegraf/plugins/parsers/prometheus/v2"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Parser struct {
	MetricVersion int
	DefaultTags   map[string]string
	Protobuf      bool
}

func IsProtobuf(header http.Header) (bool, error) {
	mediatype, params, error := mime.ParseMediaType(header.Get("Content-Type"))

	if error != nil {
		return false, error
	}

	return mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily", nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var parser expfmt.TextParser
	var err error
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	// Prepare output
	metricFamilies := make(map[string]*dto.MetricFamily)

	if p.Protobuf {
		for {
			mf := &dto.MetricFamily{}
			if _, ierr := pbutil.ReadDelimited(reader, mf); ierr != nil {
				if ierr == io.EOF {
					break
				}
				return nil, fmt.Errorf("reading metric family protocol buffer failed: %s", ierr)
			}
			metricFamilies[mf.GetName()] = mf
		}
	} else {
		metricFamilies, err = parser.TextToMetricFamilies(reader)
		if err != nil {
			return nil, fmt.Errorf("reading text format failed: %s", err)
		}
	}

	if p.MetricVersion == 2 {
		return v2.Parse(metricFamilies, p.DefaultTags, time.Now())
	} else {
		return v1.Parse(metricFamilies, p.DefaultTags, time.Now())
	}
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("No metrics in line")
	}

	if len(metrics) > 1 {
		return nil, fmt.Errorf("More than one metric in line")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
