package arc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// prometheusV2Like mimics inputs.prometheus metric_version=2: one measurement,
// one field per metric, ~590 distinct field names, 13 tags.
func prometheusV2Like(n int) []telegraf.Metric {
	ms := make([]telegraf.Metric, 0, n)
	ts := time.Unix(1700000000, 0)
	for i := range n {
		tags := map[string]string{"cycle_id": "0", "series_id": strconv.Itoa(i % 10), "url": "http://10.0.0.1:9001/metrics"}
		for k := range 10 {
			tags["label_key_kkkkk_"+strconv.Itoa(k)] = "label_val_vvvvv_" + strconv.Itoa(k)
		}
		ms = append(ms, metric.New("prometheus", tags,
			map[string]any{"avalanche_metric_mmmmm_" + strconv.Itoa(i%590): float64(i)}, ts))
	}
	return ms
}

// denseLike mimics inputs.cpu: one measurement, every metric carries the same 10 fields, 3 tags.
func denseLike(n int) []telegraf.Metric {
	ms := make([]telegraf.Metric, 0, n)
	ts := time.Unix(1700000000, 0)
	for i := range n {
		fields := make(map[string]any, 10)
		for k := range 10 {
			fields["usage_"+strconv.Itoa(k)] = float64(i + k)
		}
		ms = append(ms, metric.New("cpu",
			map[string]string{"host": "h" + strconv.Itoa(i%50), "cpu": "cpu" + strconv.Itoa(i%8), "env": "prod"}, fields, ts))
	}
	return ms
}

// manyMeasurementsLike mimics inputs.prometheus metric_version=1: ~514 measurements, few fields each.
func manyMeasurementsLike(n int) []telegraf.Metric {
	ms := make([]telegraf.Metric, 0, n)
	ts := time.Unix(1700000000, 0)
	for i := range n {
		tags := map[string]string{"cycle_id": "0", "series_id": strconv.Itoa(i % 10)}
		for k := range 10 {
			tags["label_key_kkkkk_"+strconv.Itoa(k)] = "label_val_vvvvv_" + strconv.Itoa(k)
		}
		ms = append(ms, metric.New("avalanche_metric_"+strconv.Itoa(i%514), tags, map[string]any{"counter": float64(i)}, ts))
	}
	return ms
}

func benchWrite(b *testing.B, metrics []telegraf.Metric, encoding string) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	plugin := &Arc{URL: ts.URL, Timeout: config.Duration(5 * time.Second), ContentEncoding: encoding, Log: testutil.Logger{}}
	if err := plugin.Init(); err != nil {
		b.Fatal(err)
	}
	if err := plugin.Connect(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := plugin.Write(metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	for _, bc := range []struct {
		name string
		gen  func(int) []telegraf.Metric
	}{{"sparse_v2", prometheusV2Like}, {"dense", denseLike}, {"many_measurements_v1", manyMeasurementsLike}} {
		for _, enc := range []string{"identity", "gzip"} {
			b.Run(bc.name+"/"+enc, func(b *testing.B) { benchWrite(b, bc.gen(5000), enc) })
		}
	}
}
