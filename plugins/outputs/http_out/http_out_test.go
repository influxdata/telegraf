package http_out

import (
	"encoding/json"
	"github.com/influxdata/telegraf/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpOutOK(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			var reqBody struct {
				Metrics []Metric
				Data    map[string]string
			}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			if err != nil {
				panic(err)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "OK"}`))
		}),
	)

	data := map[string]string{
		"data1": "data1",
		"data2": "data2",
	}
	h := HttpOut{
		Name:   "http_out",
		Server: ts.URL,
		Data:   data,
	}

	h.Write(testutil.MockMetrics())
}
