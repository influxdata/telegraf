package http_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/plugins/processors"
	httpplugin "github.com/influxdata/telegraf/plugins/processors/http"
	influxserializer "github.com/influxdata/telegraf/plugins/serializers/influx"
	jsonserializer "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

const (
	metricName = "metricName"
	simpleJSON = `
{
    "a": 1.2
}
`
)

func getMetric() telegraf.Metric {
	return metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
}

func writeBody(t *testing.T, w http.ResponseWriter, body string) {
	if _, err := w.Write([]byte(body)); err != nil {
		t.Error(err)
	}
}

func newPlugin(t *testing.T, serverURL string, opts func(*httpplugin.HTTP)) *httpplugin.HTTP {
	t.Helper()

	serializer := &jsonserializer.Serializer{}
	require.NoError(t, serializer.Init())

	plugin := &httpplugin.HTTP{
		URL:   serverURL,
		Merge: "override",
		Log:   testutil.Logger{},
	}
	if opts != nil {
		opts(plugin)
	}
	plugin.SetSerializer(serializer)
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{MetricName: metricName}
		err := p.Init()
		return p, err
	})
	require.NoError(t, plugin.Init())
	return plugin
}

func TestOverrideMergeJSONResponse(t *testing.T) {
	var requestBody string
	var requestMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		requestBody = string(body)
		requestMethod = r.Method
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Method = http.MethodPost
		p.Merge = "override"
		p.OnError = "keep"
	})
	require.NoError(t, plugin.Init())

	input := metric.New(
		metricName,
		map[string]string{},
		map[string]interface{}{"value": 42.0},
		time.Unix(0, 0),
	)

	results := plugin.Apply(input)
	require.Len(t, results, 1)
	require.Contains(t, requestBody, `"value":42`)
	require.Equal(t, http.MethodPost, requestMethod)

	got := results[0]
	require.Equal(t, metricName, got.Name())
	require.Equal(t, time.Unix(0, 0), got.Time())
	require.InDelta(t, 42.0, got.Fields()["value"], testutil.DefaultDelta)
	require.InDelta(t, 1.2, got.Fields()["a"], testutil.DefaultDelta)
	require.Equal(t, "200", got.Tags()["status_code"])
	require.Equal(t, "success", got.Tags()["result"])
	_, hasError := got.GetField("http_error")
	require.False(t, hasError)
}

func TestDropOriginalReturnsParsedMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.DropOriginal = true
		p.Merge = "none"
	})
	require.NoError(t, plugin.Init())

	input := metric.New(
		metricName,
		map[string]string{"keep": "no"},
		map[string]interface{}{"value": 42.0},
		time.Unix(0, 0),
	)

	results := plugin.Apply(input)
	require.Len(t, results, 1)
	require.Equal(t, metricName, results[0].Name())
	require.InDelta(t, 1.2, results[0].Fields()["a"], testutil.DefaultDelta)
	_, hasTag := results[0].GetTag("keep")
	require.False(t, hasTag)
}

func TestParentMergeKeepsOriginal(t *testing.T) {
	const responseJSON = `[{"a":1.2,"group":"one"},{"a":3.4,"group":"two"}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, responseJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Merge = "parent"
	})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{
			MetricName:   metricName,
			StringFields: []string{"group"},
		}
		err := p.Init()
		return p, err
	})
	require.NoError(t, plugin.Init())

	input := metric.New(
		metricName,
		map[string]string{"source": "starlark"},
		map[string]interface{}{"value": 42.0},
		time.Unix(0, 0),
	)

	results := plugin.Apply(input)
	require.Len(t, results, 3)
	require.Equal(t, input, results[0])
	require.Equal(t, "starlark", results[0].Tags()["source"])
	require.InDelta(t, 42.0, results[0].Fields()["value"], testutil.DefaultDelta)
	_, hasA := results[0].GetField("a")
	require.False(t, hasA)

	require.InDelta(t, 1.2, results[1].Fields()["a"], testutil.DefaultDelta)
	require.Equal(t, "one", results[1].Fields()["group"])
	require.InDelta(t, 42.0, results[1].Fields()["value"], testutil.DefaultDelta)
	require.Equal(t, "starlark", results[1].Tags()["source"])

	require.InDelta(t, 3.4, results[2].Fields()["a"], testutil.DefaultDelta)
	require.Equal(t, "two", results[2].Fields()["group"])
}

func TestParentMergeDropsOriginal(t *testing.T) {
	const responseJSON = `[{"a":1.2,"group":"one"},{"a":3.4,"group":"two"}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, responseJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Merge = "parent"
		p.DropOriginal = true
	})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{
			MetricName:   metricName,
			StringFields: []string{"group"},
		}
		err := p.Init()
		return p, err
	})
	require.NoError(t, plugin.Init())

	input := metric.New(
		metricName,
		map[string]string{"source": "starlark"},
		map[string]interface{}{"value": 42.0},
		time.Unix(0, 0),
	)

	results := plugin.Apply(input)
	require.Len(t, results, 2)
	require.InDelta(t, 1.2, results[0].Fields()["a"], testutil.DefaultDelta)
	require.Equal(t, "one", results[0].Fields()["group"])
	require.InDelta(t, 3.4, results[1].Fields()["a"], testutil.DefaultDelta)
	require.Equal(t, "two", results[1].Fields()["group"])
}

func TestOnErrorKeep(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeBody(t, w, "boom")
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "keep"
	})

	input := getMetric()
	results := plugin.Apply(input)
	require.Len(t, results, 1)
	require.Equal(t, input, results[0])
	require.Equal(t, "500", results[0].Tags()["status_code"])
	require.Equal(t, "response_status_code_mismatch", results[0].Tags()["result"])
	require.Contains(t, results[0].Fields()["http_error"], "received status code 500")
}

func TestOnErrorDrop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "drop"
	})

	results := plugin.Apply(getMetric())
	require.Empty(t, results)
}

func TestRejectGETMethod(t *testing.T) {
	plugin := &httpplugin.HTTP{
		URL:    "http://example.com",
		Method: http.MethodGet,
		Log:    testutil.Logger{},
	}
	plugin.SetSerializer(&jsonserializer.Serializer{})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return &json.Parser{MetricName: metricName}, nil
	})

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "GET is not supported")
}

func TestSuccessStatusCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.SuccessStatusCodes = []int{201}
	})

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.InDelta(t, 1.2, results[0].Fields()["a"], testutil.DefaultDelta)
	require.Equal(t, "201", results[0].Tags()["status_code"])
	require.Equal(t, "success", results[0].Tags()["result"])
}

func TestOverrideOverwritesConflictingFieldsAndTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `{"field":"new","env":"prod"}`)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Merge = "override"
	})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{
			MetricName:   metricName,
			StringFields: []string{"field", "env"},
			TagKeys:      []string{"env"},
		}
		err := p.Init()
		return p, err
	})
	require.NoError(t, plugin.Init())

	input := metric.New(
		metricName,
		map[string]string{"env": "dev"},
		map[string]interface{}{"field": "old"},
		time.Unix(0, 0),
	)

	results := plugin.Apply(input)
	require.Len(t, results, 1)
	require.Equal(t, "new", results[0].Fields()["field"])
	require.Equal(t, "prod", results[0].Tags()["env"])
}

func TestConfigLoad(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	cfg := `
[[processors.http]]
  url = "` + server.URL + `"
  data_format = "json"
  merge = "override"
`
	c := config.NewConfig()
	require.NoError(t, c.LoadConfigData([]byte(cfg), "testdata.toml"))
	require.Len(t, c.Processors, 1)
	require.Len(t, c.AggProcessors, 1)

	for _, rp := range append(c.Processors, c.AggProcessors...) {
		proc := rp.Processor.(processors.HasUnwrap)
		unwrapped, ok := proc.Unwrap().(*httpplugin.HTTP)
		require.True(t, ok)
		require.Equal(t, server.URL, unwrapped.URL)
		require.Equal(t, "override", unwrapped.Merge)
	}
}

func TestConfigRejectsTemplateDataFormat(t *testing.T) {
	cfg := `
[[processors.http]]
  url = "http://example.com"
  data_format = "template"
  template = "x"
`
	c := config.NewConfig()
	err := c.LoadConfigData([]byte(cfg), "testdata.toml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "parser not found")
}

func TestPUTAndPATCHMethods(t *testing.T) {
	methods := []string{http.MethodPut, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var gotMethod string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				writeBody(t, w, simpleJSON)
			}))
			defer server.Close()

			plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
				p.Method = method
			})

			results := plugin.Apply(getMetric())
			require.Len(t, results, 1)
			require.Equal(t, method, gotMethod)
		})
	}
}

func TestCustomHeaders(t *testing.T) {
	const header = "X-Custom"
	const value = "test-value"
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(header)
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	secret := config.NewSecret([]byte(value))
	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Headers = map[string]*config.Secret{header: &secret}
	})

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.Equal(t, value, gotHeader)
}

func TestBearerAuthViaHeaders(t *testing.T) {
	const token = "secret-token"
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	auth := config.NewSecret([]byte("Bearer " + token))
	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Headers = map[string]*config.Secret{"Authorization": &auth}
	})

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.Equal(t, "Bearer "+token, gotAuth)
}

func TestInvalidMerge(t *testing.T) {
	plugin := &httpplugin.HTTP{
		URL:   "http://example.com",
		Merge: "append",
		Log:   testutil.Logger{},
	}
	plugin.SetSerializer(&jsonserializer.Serializer{})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return &json.Parser{MetricName: metricName}, nil
	})
	require.Error(t, plugin.Init())
}

func TestParseFailureOnErrorKeep(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, "not-json")
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "keep"
	})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{MetricName: metricName, Strict: true}
		err := p.Init()
		return p, err
	})
	require.NoError(t, plugin.Init())

	input := getMetric()
	results := plugin.Apply(input)
	require.Len(t, results, 1)
	require.Equal(t, "cpu", results[0].Name())
	require.Equal(t, "200", results[0].Tags()["status_code"])
	require.Equal(t, "body_read_error", results[0].Tags()["result"])
	require.Contains(t, results[0].Fields()["http_error"], "parsing response failed")
}

func TestOnErrorKeepEmptyParserResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "keep"
	})

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.Equal(t, "200", results[0].Tags()["status_code"])
	require.Equal(t, "body_read_error", results[0].Tags()["result"])
	require.Contains(t, results[0].Fields()["http_error"], "parser returned no metrics")
}

func TestOnErrorKeepSerializationFailure(t *testing.T) {
	plugin := &httpplugin.HTTP{
		URL:     "http://example.com",
		OnError: "keep",
		Log:     testutil.Logger{},
	}
	plugin.SetSerializer(&failSerializer{})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return &json.Parser{MetricName: metricName}, nil
	})
	require.NoError(t, plugin.Init())

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.Equal(t, "processing_error", results[0].Tags()["result"])
	require.Contains(t, results[0].Fields()["http_error"], "serializing metric failed")
	_, hasStatus := results[0].GetTag("status_code")
	require.False(t, hasStatus)
}

func TestOnErrorKeepParserInstantiationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "keep"
	})
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return nil, errors.New("parser init failed")
	})
	require.NoError(t, plugin.Init())

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.Equal(t, "200", results[0].Tags()["status_code"])
	require.Equal(t, "processing_error", results[0].Tags()["result"])
	require.Contains(t, results[0].Fields()["http_error"], "instantiating parser failed")
}

type failSerializer struct{}

func (*failSerializer) Init() error { return nil }

func (*failSerializer) Serialize(_ telegraf.Metric) ([]byte, error) {
	return nil, errors.New("serialize failed")
}

func (*failSerializer) SerializeBatch(_ []telegraf.Metric) ([]byte, error) {
	return nil, errors.New("serialize failed")
}

func TestOnErrorKeepTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "keep"
		p.Timeout = config.Duration(50 * time.Millisecond)
	})

	results := plugin.Apply(getMetric())
	require.Len(t, results, 1)
	require.Equal(t, "timeout", results[0].Tags()["result"])
	_, hasStatus := results[0].GetTag("status_code")
	require.False(t, hasStatus)
}

func TestInfluxDataFormatRoundTrip(t *testing.T) {
	var requestBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		requestBody = string(body)
		writeBody(t, w, "metricName value=42 123\n")
	}))
	defer server.Close()

	serializer := &influxserializer.Serializer{}
	require.NoError(t, serializer.Init())

	plugin := &httpplugin.HTTP{
		URL:   server.URL,
		Merge: "override",
		Log:   testutil.Logger{},
	}
	plugin.SetSerializer(serializer)
	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &influx.Parser{}
		err := p.Init()
		return p, err
	})
	require.NoError(t, plugin.Init())

	input := metric.New(metricName, nil, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	results := plugin.Apply(input)
	require.Len(t, results, 1)
	require.Contains(t, requestBody, metricName)
	require.InDelta(t, 42.0, results[0].Fields()["value"], testutil.DefaultDelta)
}

func TestTrackingOnErrorKeep(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "keep"
	})

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, 1)
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input, _ := metric.WithTracking(getMetric(), notify)
	results := plugin.Apply(input)
	require.Len(t, results, 1)
	results[0].Accept()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(delivered) == 1 && delivered[0].Delivered()
	}, time.Second, 10*time.Millisecond)
}

func TestTrackingOnErrorDrop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.OnError = "drop"
	})

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, 1)
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input, _ := metric.WithTracking(getMetric(), notify)
	results := plugin.Apply(input)
	require.Empty(t, results)

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(delivered) == 1
	}, time.Second, 10*time.Millisecond)
}

func TestTrackingOverrideMerge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.Merge = "override"
	})

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, 1)
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input, _ := metric.WithTracking(getMetric(), notify)
	results := plugin.Apply(input)
	require.Len(t, results, 1)
	results[0].Accept()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(delivered) == 1 && delivered[0].Delivered()
	}, time.Second, 10*time.Millisecond)
}

func TestTrackingDropOriginal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, simpleJSON)
	}))
	defer server.Close()

	plugin := newPlugin(t, server.URL, func(p *httpplugin.HTTP) {
		p.DropOriginal = true
		p.Merge = "none"
	})

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, 1)
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input, _ := metric.WithTracking(getMetric(), notify)
	results := plugin.Apply(input)
	require.Len(t, results, 1)
	results[0].Accept()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(delivered) == 1
	}, time.Second, 10*time.Millisecond)
}
