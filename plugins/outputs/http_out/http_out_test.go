package http_out

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpOut(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "OK"}`))

			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			if err != nil {
				panic(err)
			}
			fmt.Printf("reqBody = %+v\n", reqBody)
		}),
	)

	c := serializers.Config{DataFormat: "json"}
	s, _ := serializers.NewSerializer(&c)
	data := map[string]string{
		"data1": "data1",
		"data2": "data2",
	}
	h := HttpOut{
		Name:       "http_out",
		Server:     ts.URL,
		Data:       data,
		serializer: s,
	}

	h.Connect()

	h.Write(testutil.MockMetrics())
}
