package containerapp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/stretchr/testify/assert"
)

func TestGatherHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, validJSON)
	}))
	defer ts.Close()

	a := HttpJson{
		Servers: []string{ts.URL},
		Name:    "test",
		Method:  "GET",
		client:  &RealHTTPClient{client: &http.Client{}},
	}

	dapp := NewContainerApp()
	d := internal.Duration{
		Duration: 1 * time.Millisecond,
	}

	hg := &HTTPGather{httpjsonclient: &a, interval: d, server: dapp}
	hg.Run()

	select {
	case msg := <-dapp.metricsCh:

		msg.fields["response_time"] = 1.0
		tags := map[string]string{"server": ts.URL}
		expectedFields["response_time"] = 1.0
		assert.Equal(t, msg.fields, expectedFields, "fields")
		assert.Equal(t, msg.measurement, "test", "measurement")
		assert.Equal(t, msg.tags, tags, "tags")

	}

}

func TestContainerAppClients(t *testing.T) {
	dapp := NewContainerApp()
	dapp.Tags = []string{"db"}
	dapp.TagsMandatory = []string{"db"}
	dapp.TagsPrefix = "tag."
	dapp.HTTPDefaults = map[string]string{
		"interval":              "10s",
		"http_port":             "8080",
		"http_path":             "api..influx-metrics",
		"http_response_timeout": "5s",
		"http_method":           "GET",
		"http_path_delimiter":   "..",
		"custom_tags":           "{\"tct\": \"tct\"}",
	}

	dapp.HTTP = map[string]string{
		"http_port":     "http_port",
		"tag_keys_json": "tag_keys_json",
	}

	conf := &Config{}
	conf.Name = "1"
	conf.IP = "127.0.0.1"
	conf.Values = []map[string]string{{
		"db":            "db",
		"tag.a":         "a",
		"http_port":     "9999",
		"tag_keys_json": "[\"tk\"]",
	}}
	conf.SystemTags = map[string]string{
		"sys": "sys",
	}

	dapp.Add("a", conf)
	assert.Equal(t, len(dapp.clients), 1, "clients count")
	assert.Equal(t, dapp.clients["a"].cfg.Tags, map[string]string{
		"tct":   "tct",
		"sys":   "sys",
		"db":    "db",
		"tag.a": "a",
	}, "clients tags")
	assert.Equal(t, dapp.clients["a"].cfg.Path, "api/influx-metrics", "clients count")
	assert.Equal(t, dapp.clients["a"].cfg.TagKeys, []string{"tk"}, "tag keys")
	dapp.Del("a")
	assert.Equal(t, len(dapp.clients), 0, "clients count")
}

func TestGatherHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)

	}))
	defer ts.Close()

	a := HttpJson{
		Servers: []string{ts.URL},
		Name:    "test",
		Method:  "GET",
		client:  &RealHTTPClient{client: &http.Client{}},
	}

	dapp := NewContainerApp()
	d := internal.Duration{
		Duration: 1 * time.Millisecond,
	}

	hg := &HTTPGather{httpjsonclient: &a, interval: d, server: dapp}
	hg.Run()

	select {
	case err := <-dapp.errCh:
		assert.Contains(t, err.Error(), "has status code 400 (Bad Request)")
	}

}

func TestNewHTTPGather(t *testing.T) {
	dapp := NewContainerApp()
	cfg := &HTTPConfig{
		NameOverride:    "test",
		IP:              "127.0.0.1",
		Port:            8888,
		Path:            "/test/",
		Interval:        "1s",
		ResponseTimeout: "1s",
		Method:          "GET",
		Tags:            map[string]string{"test": "test"},
		TagKeys:         []string{"tk"},
		Parameters:      map[string]string{"tp": "tp"},
		Headers:         map[string]string{"th": "th"},
	}
	hg, _ := NewHTTPGather(dapp, "test", cfg)

	assert.Equal(t, hg.httpjsonclient.Headers, map[string]string{"th": "th"}, "Headers")
	assert.Equal(t, hg.httpjsonclient.Method, "GET", "Method")
	assert.Equal(t, hg.httpjsonclient.Parameters, map[string]string{"tp": "tp"}, "Parameters")
	assert.Equal(t, hg.httpjsonclient.ResponseTimeout, internal.Duration{Duration: time.Second}, "ResponseTimeout")
	assert.Equal(t, hg.httpjsonclient.Servers, []string{"http://127.0.0.1:8888/test/"}, "Servers")
	assert.Equal(t, hg.httpjsonclient.TagKeys, []string{"tk"}, "TagKeys")
}

func TestHttpGatherConf(t *testing.T) {
	settings := map[string]string{
		"http_port":             "E_http_port",
		"name_override":         "E_name_override",
		"http_path":             "E_http_path",
		"interval":              "E_interval",
		"http_response_timeout": "E_http_response_timeout",
		"http_method":           "E_http_method",
		"tag_keys_json":         "E_tag_keys_json",
		"custom_tags":           "E_custom_tags",
		"http_parameters":       "E_http_parameters",
		"http_headers":          "E_http_headers",
	}
	values := map[string]string{
		"E_http_port":             "8888",
		"E_name_override":         "test",
		"E_http_path":             "/test/",
		"E_interval":              "1s",
		"E_http_response_timeout": "1s",
		"E_http_method":           "GET",
		"E_tag_keys_json":         "[\"tk\"]",
		"E_custom_tags":           "{\"tct\": \"tct\"}",
		"E_http_parameters":       "{\"tp\": \"tp\"}",
		"E_http_headers":          "{\"th\": \"th\"}",
	}

	cfg, _ := CreateHTTPGatherConf("c", settings, map[string]string{}, values)
	cfg.IP = "127.0.0.1"
	dapp := NewContainerApp()
	hg, err := NewHTTPGather(dapp, "test", cfg)

	assert.Equal(t, err, nil, "err")

	assert.Equal(t, hg.httpjsonclient.Headers, map[string]string{"th": "th"}, "Headers")
	assert.Equal(t, hg.httpjsonclient.Method, "GET", "Method")
	assert.Equal(t, hg.httpjsonclient.Parameters, map[string]string{"tp": "tp"}, "Parameters")
	assert.Equal(t, hg.httpjsonclient.ResponseTimeout, internal.Duration{Duration: time.Second}, "ResponseTimeout")
	assert.Equal(t, hg.httpjsonclient.Servers, []string{"http://127.0.0.1:8888/test/"}, "Servers")
	assert.Equal(t, hg.httpjsonclient.TagKeys, []string{"tk"}, "TagKeys")
}
