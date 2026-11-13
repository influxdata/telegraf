package prometheus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
)

// benchmarkExposition generates a text exposition with counters, gauges and
// histograms carrying ten labels each, similar to a typical exporter.
func benchmarkExposition(families, series int) string {
	var labels strings.Builder
	for i := range 10 {
		fmt.Fprintf(&labels, `,label_%d="value_%d"`, i, i)
	}

	var sb strings.Builder
	for f := range families {
		switch f % 3 {
		case 0:
			fmt.Fprintf(&sb, "# TYPE counter_%d_total counter\n", f)
			for s := range series {
				fmt.Fprintf(&sb, "counter_%d_total{series=\"%d\"%s} %d\n", f, s, labels.String(), f*s)
			}
		case 1:
			fmt.Fprintf(&sb, "# TYPE gauge_%d gauge\n", f)
			for s := range series {
				fmt.Fprintf(&sb, "gauge_%d{series=\"%d\"%s} %d.5\n", f, s, labels.String(), f*s)
			}
		case 2:
			fmt.Fprintf(&sb, "# TYPE histogram_%d histogram\n", f)
			for s := range series {
				for _, le := range []string{"0.1", "1", "10", "+Inf"} {
					fmt.Fprintf(&sb, "histogram_%d_bucket{series=\"%d\",le=\"%s\"%s} %d\n", f, s, le, labels.String(), s)
				}
				fmt.Fprintf(&sb, "histogram_%d_sum{series=\"%d\"%s} %d\n", f, s, labels.String(), s)
				fmt.Fprintf(&sb, "histogram_%d_count{series=\"%d\"%s} %d\n", f, s, labels.String(), s)
			}
		}
	}
	return sb.String()
}

func BenchmarkGather(b *testing.B) {
	body := []byte(benchmarkExposition(300, 10))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if _, err := w.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	for _, version := range []int{1, 2} {
		b.Run(fmt.Sprintf("v%d", version), func(b *testing.B) {
			p := &Prometheus{
				URLs:          []string{ts.URL},
				URLTag:        "url",
				MetricVersion: version,
				Log:           testutil.Logger{},
			}
			if err := p.Init(); err != nil {
				b.Fatal(err)
			}

			ch := make(chan telegraf.Metric, 1024)
			done := make(chan struct{})
			go func() {
				for range ch {
				}
				close(done)
			}()
			acc := agent.NewAccumulator(models.NewRunningInput(p, &models.InputConfig{Name: "prometheus"}), ch)

			b.ReportAllocs()
			for b.Loop() {
				if err := p.Gather(acc); err != nil {
					b.Fatal(err)
				}
			}
			close(ch)
			<-done
		})
	}
}
