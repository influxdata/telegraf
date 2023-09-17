package opensearch

import (
	"context"
	"fmt"
	"math"
	"testing"
	"text/template"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:                urls,
		IndexName:           `test-{{.Time.Format "2006-01-02"}}`,
		Timeout:             config.Duration(time.Second * 5),
		EnableGzip:          false,
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		Log:                 testutil.Logger{},
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
	// Verify that we can connect to Opensearch
	require.NoError(t, e.Connect())

	// Verify that we can successfully write data to Opensearch
	require.NoError(t, e.Write(testutil.MockMetrics()))
}

func TestConnectAndWriteMetricWithNaNValueEmptyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:                urls,
		IndexName:           `test-{{.Time.Format "2006-01-02"}}`,
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}

	// Verify that we can connect to Opensearch
	require.NoError(t, e.Init())
	require.NoError(t, e.Connect())

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		require.Error(t, e.Write([]telegraf.Metric{m}), "error sending bulk request to Opensearch: "+
			"json: unsupported value: NaN")
	}
}

func TestConnectAndWriteMetricWithNaNValueNoneIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:                urls,
		IndexName:           `test-{{.Time.Format "2006-01-02"}}`,
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		FloatHandling:       "none",
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
	// Verify that we can connect to Opensearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		err = e.Write([]telegraf.Metric{m})
		require.Error(t, err, "error sending bulk request to Opensearch: json: unsupported value: NaN")
	}
}

func TestConnectAndWriteMetricWithNaNValueDropIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:                urls,
		IndexName:           `test-{{.Time.Format "2006-01-02"}}`,
		Timeout:             config.Duration(time.Second * 5),
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: config.Duration(time.Second * 10),
		HealthCheckTimeout:  config.Duration(time.Second * 1),
		FloatHandling:       "drop",
		Log:                 testutil.Logger{},
	}

	metrics := []telegraf.Metric{
		testutil.TestMetric(math.NaN()),
		testutil.TestMetric(math.Inf(1)),
		testutil.TestMetric(math.Inf(-1)),
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
	// Verify that we can connect to Opensearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can fail for metric with unhandled NaN/inf/-inf values
	for _, m := range metrics {
		err = e.Write([]telegraf.Metric{m})
		require.NoError(t, err)
	}
}

func TestConnectAndWriteMetricWithNaNValueReplacementIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		floatHandle      string
		floatReplacement float64
		expectError      bool
	}{
		{
			"none",
			0.0,
			true,
		},
		{
			"drop",
			0.0,
			false,
		},
		{
			"replace",
			0.0,
			false,
		},
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	for _, test := range tests {
		e := &Opensearch{
			URLs:                urls,
			IndexName:           `test-{{.Time.Format "2006-01-02"}}`,
			Timeout:             config.Duration(time.Second * 5),
			ManageTemplate:      true,
			TemplateName:        "telegraf",
			OverwriteTemplate:   false,
			HealthCheckInterval: config.Duration(time.Second * 10),
			HealthCheckTimeout:  config.Duration(time.Second * 1),
			FloatHandling:       test.floatHandle,
			FloatReplacement:    test.floatReplacement,
			Log:                 testutil.Logger{},
		}

		metrics := []telegraf.Metric{
			testutil.TestMetric(math.NaN()),
			testutil.TestMetric(math.Inf(1)),
			testutil.TestMetric(math.Inf(-1)),
		}
		e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
		err := e.Connect()
		require.NoError(t, err)

		for _, m := range metrics {
			err = e.Write([]telegraf.Metric{m})

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}
	}
}

func TestTemplateManagementEmptyTemplateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:              urls,
		IndexName:         `test-{{.Time.Format "2006-01-02"}}`,
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        false,
		ManageTemplate:    true,
		TemplateName:      "",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	err := e.Init()
	require.Error(t, err)
}

func TestTemplateManagementIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:              urls,
		IndexName:         `test-{{.Time.Format "2006-01-02"}}`,
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        false,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)

	err := e.Connect()
	require.NoError(t, err)

	err = e.manageTemplate(ctx)
	require.NoError(t, err)
}

func TestTemplateInvalidIndexPatternIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t, imageVersion1)
	defer container.Terminate()

	urls := []string{
		fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort]),
	}

	e := &Opensearch{
		URLs:              urls,
		IndexName:         `{{.Tag "tag1"}}-{{.Time.Format "2006-01-02"}}`,
		Timeout:           config.Duration(time.Second * 5),
		EnableGzip:        false,
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
		Log:               testutil.Logger{},
	}
	e.indexTmpl, _ = template.New("index").Parse(e.IndexName)
	err := e.Connect()
	require.Error(t, err)
}
