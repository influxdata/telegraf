package http_listener

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

// newListener is the minimal HTTPListener construction to serve writes.
func newListener() *HTTPListener {
	listener := &HTTPListener{
		TimeFunc:  time.Now,
		acc:       &testutil.NopAccumulator{},
		BytesRecv: selfstat.Register("http_listener", "bytes_received", map[string]string{}),
		handler:   influx.NewMetricHandler(),
		pool:      NewPool(200, DEFAULT_MAX_LINE_SIZE),
		MaxLineSize: internal.Size{
			Size: DEFAULT_MAX_LINE_SIZE,
		},
		MaxBodySize: internal.Size{
			Size: DEFAULT_MAX_BODY_SIZE,
		},
	}
	listener.parser = influx.NewParser(listener.handler)
	return listener
}

func BenchmarkHTTPListener_serveWrite(b *testing.B) {
	res := httptest.NewRecorder()
	addr := "http://localhost/write?db=mydb"

	benchmarks := []struct {
		name  string
		lines string
	}{
		{
			name:  "single line, tag, and field",
			lines: lines(1, 1, 1),
		},
		{
			name:  "single line, 10 tags and fields",
			lines: lines(1, 10, 10),
		},
		{
			name:  "single line, 100 tags and fields",
			lines: lines(1, 100, 100),
		},
		{
			name:  "1k lines, single tag and field",
			lines: lines(1000, 1, 1),
		},
		{
			name:  "1k lines, 10 tags and fields",
			lines: lines(1000, 10, 10),
		},
		{
			name:  "10k lines, 10 tags and fields",
			lines: lines(10000, 10, 10),
		},
		{
			name:  "100k lines, 10 tags and fields",
			lines: lines(100000, 10, 10),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			listener := newListener()

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				req, err := http.NewRequest("POST", addr, strings.NewReader(bm.lines))
				if err != nil {
					b.Error(err)
				}
				listener.serveWrite(res, req)
				if res.Code != http.StatusNoContent {
					b.Errorf("unexpected status %d", res.Code)
				}
			}
		})
	}
}

func lines(lines, numTags, numFields int) string {
	lp := make([]string, lines)
	for i := 0; i < lines; i++ {
		tags := make([]string, numTags)
		for j := 0; j < numTags; j++ {
			tags[j] = fmt.Sprintf("t%d=v%d", j, j)
		}

		fields := make([]string, numFields)
		for k := 0; k < numFields; k++ {
			fields[k] = fmt.Sprintf("f%d=%d", k, k)
		}

		lp[i] = fmt.Sprintf("m%d,%s %s",
			i,
			strings.Join(tags, ","),
			strings.Join(fields, ","),
		)
	}

	return strings.Join(lp, "\n")
}
