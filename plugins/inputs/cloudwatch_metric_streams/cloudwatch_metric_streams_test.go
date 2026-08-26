package cloudwatch_metric_streams

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestInvalidListenerConfig(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "address_without_port",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.Error(t, plugin.Start(&acc))

	// Stop is called when any ServiceInput fails to start; it must succeed regardless of state
	plugin.Stop()
}

func TestWriteHTTPSNoClientAuth(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		ServerConfig:   *pki.TLSServerConfig(),
		Paths:          []string{"/write"},
	}
	plugin.TLSAllowedCACerts = nil
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "https://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	cas := x509.NewCertPool()
	cas.AppendCertsFromPEM([]byte(pki.ReadServerCert()))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: cas,
			},
		},
	}

	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	resp, err := client.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPSWithClientAuth(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		ServerConfig:   *pki.TLSServerConfig(),
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "https://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	tlsConfig, err := pki.TLSClientConfig().TLSConfig()
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	resp, err := client.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPSuccessfulAuth(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		AccessKey:      "super-secure-password!",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	client := &http.Client{}

	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", addr, bytes.NewBuffer(records))
	require.NoError(t, err)
	req.Header.Set("X-Amz-Firehose-Access-Key", "super-secure-password!")
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPFailedAuth(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		AccessKey:      "super-secure-password!",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	client := &http.Client{}

	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", addr, bytes.NewBuffer(records))
	require.NoError(t, err)
	req.Header.Set("X-Amz-Firehose-Access-Key", "super-insecure-password!")
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWriteHTTP(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPMultipleRecords(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post multiple records to the metric stream listener
	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPExactMaxBodySize(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
		MaxBodySize:    config.Size(616),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPVerySmallMaxBody(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
		MaxBodySize:    config.Size(512),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post single message to the metric stream listener
	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestReceive404ForInvalidEndpoint(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/foobar"

	// post single message to the metric stream listener
	records, err := os.ReadFile("testdata/record.json")
	require.NoError(t, err)

	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 404, resp.StatusCode)
}

func TestWriteHTTPInvalid(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post a badly formatted message to the metric stream listener
	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBufferString("blahblahblah: 42\n"))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWriteHTTPEmpty(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	// post empty message to the metric stream listener
	//nolint:gosec // We must construct the address from the server due to dynamic port assignment
	resp, err := http.Post(addr, "", bytes.NewBufferString(""))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
}

func TestComposeMetrics(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// compose a data object for writing
	data := data{
		MetricStreamName: "cloudwatch-metric-stream",
		AccountID:        "546734499701",
		Region:           "us-west-2",
		Namespace:        "AWS/EC2",
		MetricName:       "CPUUtilization",
		Dimensions:       map[string]string{"AutoScalingGroupName": "test-autoscaling-group"},
		Timestamp:        1651679400000,
		Value:            map[string]float64{"max": 0.4366666666666666, "min": 0.3683333333333333, "sum": 1.9399999999999997, "count": 5.0},
		Unit:             "Percent",
	}

	// Compose the metrics from data
	plugin.composeMetrics(data)

	expected := []telegraf.Metric{
		metric.New(
			"aws_ec2_cpuutilization",
			map[string]string{
				"AutoScalingGroupName": "test-autoscaling-group",
				"accountId":            "546734499701",
				"region":               "us-west-2",
			},
			map[string]interface{}{
				"max":   0.4366666666666666,
				"min":   0.3683333333333333,
				"sum":   1.9399999999999997,
				"count": 5.0,
			},
			time.Unix(1651679400, 0),
		),
	}

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 1*time.Second, 100*time.Millisecond)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestComposeAPICompatibleMetrics(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:              testutil.Logger{},
		ServiceAddress:   "localhost:8080",
		Paths:            []string{"/write"},
		MaxBodySize:      config.Size(524288000),
		APICompatability: true,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// compose a data object for writing
	data := data{
		MetricStreamName: "cloudwatch-metric-stream",
		AccountID:        "546734499701",
		Region:           "us-west-2",
		Namespace:        "AWS/EC2",
		MetricName:       "CPUUtilization",
		Dimensions:       map[string]string{"AutoScalingGroupName": "test-autoscaling-group"},
		Timestamp:        1651679400000,
		Value:            map[string]float64{"max": 0.4366666666666666, "min": 0.3683333333333333, "sum": 1.9399999999999997, "count": 5.0},
		Unit:             "Percent",
	}

	// Compose the metrics from data
	plugin.composeMetrics(data)

	expected := []telegraf.Metric{
		metric.New(
			"aws_ec2_cpuutilization",
			map[string]string{
				"AutoScalingGroupName": "test-autoscaling-group",
				"accountId":            "546734499701",
				"region":               "us-west-2",
			},
			map[string]interface{}{
				"maximum":     0.4366666666666666,
				"minimum":     0.3683333333333333,
				"sum":         1.9399999999999997,
				"samplecount": 5.0,
			},
			time.Unix(1651679400, 0),
		),
	}

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 1*time.Second, 100*time.Millisecond)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

// post GZIP encoded data to the metric stream listener
func TestWriteHTTPGzippedData(t *testing.T) {
	plugin := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		Paths:          []string{"/write"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String() + "/write"

	data, err := os.ReadFile("./testdata/records.gz")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", addr, bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}
