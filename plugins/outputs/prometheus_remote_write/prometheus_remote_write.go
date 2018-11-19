package prometheus_remote_write

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client"
)

func init() {
	outputs.Add("prometheus_remote_write", func() telegraf.Output {
		return &PrometheusRemoteWrite{}
	})
}

type PrometheusRemoteWrite struct {
	URL string
}

var sampleConfig = `
  ## URL to send Prometheus remote write requests to.
  url = "http://localhost/push"
`

func (p *PrometheusRemoteWrite) Connect() error {
	return nil
}

func (p *PrometheusRemoteWrite) Close() error {
	return nil
}

func (p *PrometheusRemoteWrite) Description() string {
	return "Configuration for the Prometheus remote write client to spawn"
}

func (p *PrometheusRemoteWrite) SampleConfig() string {
	return sampleConfig
}

func (p *PrometheusRemoteWrite) Write(metrics []telegraf.Metric) error {
	var req WriteRequest

	for _, metric := range metrics {
		tags := metric.TagList()
		commonLabels := make([]*Label, 0, len(tags))
		for _, tag := range tags {
			commonLabels = append(commonLabels, &Label{
				Name:  prometheus_client.Sanitize(tag.Key),
				Value: tag.Value,
			})
		}

	fields:
		for _, field := range metric.FieldList() {
			labels := make([]*Label, len(commonLabels), len(commonLabels)+1)
			copy(labels, commonLabels)
			labels = append(labels, &Label{
				Name:  "__name__",
				Value: metric.Name() + "_" + field.Key,
			})
			sort.Sort(byName(labels))

			// Ignore histograms and summaries.
			switch metric.Type() {
			case telegraf.Histogram, telegraf.Summary:
				continue fields
			}

			// Ignore string and bool fields.
			var value float64
			switch fv := field.Value.(type) {
			case int64:
				value = float64(fv)
			case uint64:
				value = float64(fv)
			case float64:
				value = fv
			default:
				continue fields
			}

			req.Timeseries = append(req.Timeseries, &TimeSeries{
				Labels: labels,
				Samples: []*Sample{{
					Timestamp: metric.Time().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond)),
					Value:     value,
				}},
			})
		}
	}

	buf, err := proto.Marshal(&req)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, buf)
	resp, err := http.Post(p.URL, "application/x-protobuf", bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("server returned HTTP status %s (%d)", resp.Status, resp.StatusCode)
	}
	return nil
}

type byName []*Label

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
