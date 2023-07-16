package opensearch

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

const servicePort = "9200"
const imageVersion1 = "1.1.0"
const imageVersion2 = "2.8.0"

func launchTestContainer(t *testing.T, imageVersion string) *testutil.Container {
	container := testutil.Container{
		Image:        "opensearchproject/opensearch:" + imageVersion,
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"discovery.type":              "single-node",
			"DISABLE_INSTALL_DEMO_CONFIG": "true",
			"DISABLE_SECURITY_PLUGIN":     "true",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("Init AD version hash ring successfully"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")

	return &container
}

func TestGetReplacementKeys(t *testing.T) {
	e := &Opensearch{
		DefaultTagValue: "none",
		Log:             testutil.Logger{},
	}

	tests := []struct {
		IndexName         string
		ExpectedIndexName string
		ExpectedTagKeys   []string
	}{
		{
			"indexname",
			"indexname",
			[]string{},
		}, {
			`indexname-{{.Time.Format "2006-01"}}`,
			`indexname-{{.Time.Format "2006-01"}}`,
			[]string{},
		}, {
			`indexname-{{.Time.Format "2006-01-02"}}`,
			`indexname-{{.Time.Format "2006-01-02"}}`,
			[]string{},
		}, {
			`indexname-{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
			`indexname-%s-{{.Time.Format "2006-01-02"}}`,
			[]string{"tag1"},
		}, {
			`indexname-{{.Tag "tag1"}}-{{.Tag "tag2"}}-{{.Time.Format "2006-01-02"}}`,
			`indexname-%s-%s-{{.Time.Format "2006-01-02"}}`,
			[]string{"tag1", "tag2"},
		}, {
			`indexname-{{.Tag "tag1"}}-{{.Tag "tag2"}}-{{.Tag "tag3"}}-{{.Time.Format "2006-01-02"}}`,
			`indexname-%s-%s-%s-{{.Time.Format "2006-01-02"}}`,
			[]string{"tag1", "tag2", "tag3"},
		},
		{
			`indexname-{{.Tag "tag1"}}-{{.Tag "tag2"}}-{{.Time.Format "2006-01-02"}}-{{.Tag "tag3"}}`,
			`indexname-%s-%s-{{.Time.Format "2006-01-02"}}-%s`,
			[]string{"tag1", "tag2", "tag3"},
		},
	}
	for _, test := range tests {
		indexName, tagKeys := e.GetReplacementKeys(test.IndexName, ".Tag", "%s")
		if indexName != test.ExpectedIndexName {
			t.Errorf("Expected indexname %s, got %s\n", test.ExpectedIndexName, indexName)
		}
		if !reflect.DeepEqual(tagKeys, test.ExpectedTagKeys) {
			t.Errorf("Expected tagKeys %s, got %s\n", test.ExpectedTagKeys, tagKeys)
		}
	}
}

func TestGetIndexName(t *testing.T) {
	e := &Opensearch{
		DefaultTagValue: "none",
		Log:             testutil.Logger{},
	}

	tests := []struct {
		EventTime time.Time
		Tags      map[string]string
		TagKeys   []string
		IndexName string
		Expected  string
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"indexname",
			"indexname",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			`indexname-{{.Time.Format "2006"}}`,
			"indexname-2014",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			`indexname-{{.Time.Format "2006-01"}}`,
			"indexname-2014-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			`indexname-{{.Time.Format "2006-01-02"}}`,
			"indexname-2014-12-01",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{"tag1"},
			`indexname-%s-{{.Time.Format "2006-01"}}`,
			"indexname-value1-2014-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{"tag1", "tag2"},
			`indexname-%s-%s-{{.Time.Format "2006-01"}}`,
			"indexname-value1-value2-2014-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{"tag1", "tag2", "tag3"},
			`indexname-%s-%s-%s-{{.Time.Format "2006-01"}}`,
			"indexname-value1-value2-none-2014-12",
		},
	}
	for _, test := range tests {
		indexName := e.GetIndexName(test.IndexName, test.EventTime, test.TagKeys, test.Tags)
		if indexName != test.Expected {
			t.Errorf("Expected indexname %s, got %s\n", test.Expected, indexName)
		}
	}
}

func TestGetPipelineName(t *testing.T) {
	e := &Opensearch{
		UsePipeline:     "{{es-pipeline}}",
		DefaultPipeline: "myDefaultPipeline",
		Log:             testutil.Logger{},
	}
	e.pipelineName, e.pipelineTagKeys = e.GetReplacementKeys(e.UsePipeline, "", "%s")

	tests := []struct {
		EventTime       time.Time
		Tags            map[string]string
		PipelineTagKeys []string
		Expected        string
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"myDefaultPipeline",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"myDefaultPipeline",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			[]string{},
			"myOtherPipeline",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"pipeline2",
		},
	}
	for _, test := range tests {
		pipelineName := e.getPipelineName(e.pipelineName, e.pipelineTagKeys, test.Tags)
		require.Equal(t, test.Expected, pipelineName)
	}

	// Setup testing for testing no pipeline set. All the tests in this case should return "".
	e = &Opensearch{
		Log: testutil.Logger{},
	}
	e.pipelineName, e.pipelineTagKeys = e.GetReplacementKeys(e.UsePipeline, "", "%s")

	for _, test := range tests {
		pipelineName := e.getPipelineName(e.pipelineName, e.pipelineTagKeys, test.Tags)
		require.Equal(t, "", pipelineName)
	}
}

func TestPipelineConfigs(t *testing.T) {
	tests := []struct {
		EventTime       time.Time
		Tags            map[string]string
		PipelineTagKeys []string
		Expected        string
		Elastic         *Opensearch
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"",
			&Opensearch{
				Log: testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "tag2": "value2"},
			[]string{},
			"",
			&Opensearch{
				DefaultPipeline: "myDefaultPipeline",
				Log:             testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "myOtherPipeline"},
			[]string{},
			"myDefaultPipeline",
			&Opensearch{
				UsePipeline: "myDefaultPipeline",
				Log:         testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"",
			&Opensearch{
				DefaultPipeline: "myDefaultPipeline",
				Log:             testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"pipeline2",
			&Opensearch{
				UsePipeline: "{{es-pipeline}}",
				Log:         testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1", "es-pipeline": "pipeline2"},
			[]string{},
			"value1-pipeline2",
			&Opensearch{
				UsePipeline: "{{tag1}}-{{es-pipeline}}",
				Log:         testutil.Logger{},
			},
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			map[string]string{"tag1": "value1"},
			[]string{},
			"",
			&Opensearch{
				UsePipeline: "{{es-pipeline}}",
				Log:         testutil.Logger{},
			},
		},
	}

	for _, test := range tests {
		e := test.Elastic
		e.pipelineName, e.pipelineTagKeys = e.GetReplacementKeys(e.UsePipeline, "", "%s")
		pipelineName := e.getPipelineName(e.pipelineName, e.pipelineTagKeys, test.Tags)
		require.Equal(t, test.Expected, pipelineName)
	}
}

func TestRequestHeaderWhenGzipIsEnabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			require.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
			require.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
			_, err := w.Write([]byte("{}"))
			require.NoError(t, err)
			return
		default:
			_, err := w.Write([]byte(`{"version": {"number": "7.8"}}`))
			require.NoError(t, err)
			return
		}
	}))
	defer ts.Close()

	urls := []string{"http://" + ts.Listener.Addr().String()}

	e := &Opensearch{
		URLs:           urls,
		IndexName:      `{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:        config.Duration(time.Second * 5),
		EnableGzip:     true,
		ManageTemplate: false,
		Log:            testutil.Logger{},
	}

	e.IndexName, e.tagKeys = e.GetReplacementKeys(e.IndexName, ".Tag", "%s")
	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestRequestHeaderWhenGzipIsDisabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			require.NotEqual(t, "gzip", r.Header.Get("Content-Encoding"))
			_, err := w.Write([]byte("{}"))
			require.NoError(t, err)
			return
		default:
			_, err := w.Write([]byte(`{"version": {"number": "7.8"}}`))
			require.NoError(t, err)
			return
		}
	}))
	defer ts.Close()

	urls := []string{"http://" + ts.Listener.Addr().String()}

	e := &Opensearch{
		URLs:           urls,
		IndexName:      `{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:        config.Duration(time.Second * 5),
		EnableGzip:     false,
		ManageTemplate: false,
		Log:            testutil.Logger{},
	}

	e.IndexName, e.tagKeys = e.GetReplacementKeys(e.IndexName, ".Tag", "%s")
	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestAuthorizationHeaderWhenBearerTokenIsPresent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_bulk":
			require.Equal(t, "Bearer 0123456789abcdef", r.Header.Get("Authorization"))
			_, err := w.Write([]byte("{}"))
			require.NoError(t, err)
			return
		default:
			_, err := w.Write([]byte(`{"version": {"number": "7.8"}}`))
			require.NoError(t, err)
			return
		}
	}))
	defer ts.Close()

	urls := []string{"http://" + ts.Listener.Addr().String()}

	e := &Opensearch{
		URLs:            urls,
		IndexName:       `{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:         config.Duration(time.Second * 5),
		EnableGzip:      false,
		ManageTemplate:  false,
		Log:             testutil.Logger{},
		AuthBearerToken: config.NewSecret([]byte("0123456789abcdef")),
	}

	e.IndexName, e.tagKeys = e.GetReplacementKeys(e.IndexName, ".Tag", "%s")
	err := e.Connect()
	require.NoError(t, err)

	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
