package http_out

import (
	"encoding/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func defaultHandler(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")

		var reqBody struct {
			Metrics []Metric
			Data    map[string]string
		}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK"}`))
	}
}

func Server(h func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(h))
}

func TestHttpOutOK(t *testing.T) {
	ts := Server(defaultHandler(t))

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

func TestNotSetupServer(t *testing.T) {
	h := HttpOut{
		Name: "http_out",
	}

	err := h.Write(testutil.MockMetrics())
	assert.Error(t, err)
}

func TestInvalidServer(t *testing.T) {
	ts := Server(defaultHandler(t))
	h := HttpOut{
		Name:   "http_out",
		Server: ts.URL,
	}

	err := h.Write(testutil.MockMetrics())
	assert.NoError(t, err)
}

func TestSetupData(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")

		var reqBody struct {
			Metrics []Metric
			Data    map[string]string
		}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		assert.Equal(t, reqBody.Data["secret"], "12345")
		assert.Equal(t, reqBody.Data["username"], "Username")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK"}`))
	}
	ts := Server(handler)

	data := map[string]string{
		"secret":   "12345",
		"username": "Username",
	}
	h := HttpOut{
		Name:   "http_out",
		Server: ts.URL,
		Data:   data,
	}
	err := h.Write(testutil.MockMetrics())
	assert.NoError(t, err)
}

func TestSetupHeaders(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")

		assert.Equal(t, r.Header.Get("Api-Version"), "v1.0")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK"}`))
	}
	ts := Server(handler)

	headers := map[string]string{
		"Api-Version": "v1.0",
	}
	h := HttpOut{
		Name:    "http_out",
		Server:  ts.URL,
		Headers: headers,
	}
	err := h.Write(testutil.MockMetrics())
	assert.NoError(t, err)
}
