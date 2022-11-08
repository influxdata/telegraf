package cloudwatch_metric_streams

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const (
	badMsg       = "blahblahblah: 42\n"
	emptyMsg     = ""
	accessKey    = "super-secure-password!"
	badAccessKey = "super-insecure-password!"
	maxBodySize  = 524288000
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")
)

func newTestCloudWatchMetricStreams() *CloudWatchMetricStreams {
	metricStream := &CloudWatchMetricStreams{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:8080",
		Paths:          []string{"/write"},
		MaxBodySize:    config.Size(maxBodySize),
		close:          make(chan struct{}),
	}
	return metricStream
}

func newTestMetricStreamAuth() *CloudWatchMetricStreams {
	metricStream := newTestCloudWatchMetricStreams()
	metricStream.AccessKey = accessKey
	return metricStream
}

func newTestMetricStreamHTTPS() *CloudWatchMetricStreams {
	metricStream := newTestCloudWatchMetricStreams()
	metricStream.ServerConfig = *pki.TLSServerConfig()

	return metricStream
}

func newTestCompatibleCloudWatchMetricStreams() *CloudWatchMetricStreams {
	metricStream := newTestCloudWatchMetricStreams()
	metricStream.APICompatability = true
	return metricStream
}

func getHTTPSClient() *http.Client {
	tlsConfig, err := pki.TLSClientConfig().TLSConfig()
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

func createURL(scheme string, path string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:8080",
		Path:     path,
		RawQuery: "",
	}
	return u.String()
}

func readJSON(t *testing.T, jsonFilePath string) []byte {
	data, err := os.ReadFile(jsonFilePath)
	require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)

	return data
}

func TestInvalidListenerConfig(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()
	metricStream.ServiceAddress = "address_without_port"

	acc := &testutil.Accumulator{}
	require.Error(t, metricStream.Start(acc))

	// Stop is called when any ServiceInput fails to start; it must succeed regardless of state
	metricStream.Stop()
}

func TestWriteHTTPSNoClientAuth(t *testing.T) {
	metricStream := newTestMetricStreamHTTPS()
	metricStream.TLSAllowedCACerts = nil

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	cas := x509.NewCertPool()
	cas.AppendCertsFromPEM([]byte(pki.ReadServerCert()))
	noClientAuthClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: cas,
			},
		},
	}

	// post single message to the metric stream listener
	record := readJSON(t, "testdata/record.json")
	resp, err := noClientAuthClient.Post(createURL("https", "/write"), "", bytes.NewBuffer(record))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
}

func TestWriteHTTPSWithClientAuth(t *testing.T) {
	metricStream := newTestMetricStreamHTTPS()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post single message to the metric stream listener
	record := readJSON(t, "testdata/record.json")
	resp, err := getHTTPSClient().Post(createURL("https", "/write"), "", bytes.NewBuffer(record))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
}

func TestWriteHTTPSuccessfulAuth(t *testing.T) {
	metricStream := newTestMetricStreamAuth()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	client := &http.Client{}

	record := readJSON(t, "testdata/record.json")
	req, err := http.NewRequest("POST", createURL("http", "/write"), bytes.NewBuffer(record))
	require.NoError(t, err)
	req.Header.Set("X-Amz-Firehose-Access-Key", accessKey)

	// post single message to the metric stream listener
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusOK, resp.StatusCode)
}

func TestWriteHTTPFailedAuth(t *testing.T) {
	metricStream := newTestMetricStreamAuth()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	client := &http.Client{}

	record := readJSON(t, "testdata/record.json")
	req, err := http.NewRequest("POST", createURL("http", "/write"), bytes.NewBuffer(record))
	require.NoError(t, err)
	req.Header.Set("X-Amz-Firehose-Access-Key", badAccessKey)

	// post single message to the metric stream listener
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWriteHTTP(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post single message to the metric stream listener
	record := readJSON(t, "testdata/record.json")
	resp, err := http.Post(createURL("http", "/write"), "", bytes.NewBuffer(record))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
}

func TestWriteHTTPMultipleRecords(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post multiple records to the metric stream listener
	records := readJSON(t, "testdata/records.json")
	resp, err := http.Post(createURL("http", "/write"), "", bytes.NewBuffer(records))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
}

func TestWriteHTTPExactMaxBodySize(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()
	record := readJSON(t, "testdata/record.json")
	metricStream.MaxBodySize = config.Size(len(record))

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post single message to the metric stream listener
	resp, err := http.Post(createURL("http", "/write"), "", bytes.NewBuffer(record))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
}

func TestWriteHTTPVerySmallMaxBody(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()
	metricStream.MaxBodySize = config.Size(512)

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post single message to the metric stream listener
	record := readJSON(t, "testdata/record.json")
	resp, err := http.Post(createURL("http", "/write"), "", bytes.NewBuffer(record))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 413, resp.StatusCode)
}

func TestReceive404ForInvalidEndpoint(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post single message to the metric stream listener
	record := readJSON(t, "testdata/record.json")
	resp, err := http.Post(createURL("http", "/foobar"), "", bytes.NewBuffer(record))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 404, resp.StatusCode)
}

func TestWriteHTTPInvalid(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post a badly formatted message to the metric stream listener
	resp, err := http.Post(createURL("http", "/write"), "", bytes.NewBuffer([]byte(badMsg)))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 400, resp.StatusCode)
}

func TestWriteHTTPEmpty(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// post empty message to the metric stream listener
	resp, err := http.Post(createURL("http", "/write"), "", bytes.NewBuffer([]byte(emptyMsg)))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 400, resp.StatusCode)
}

func TestComposeMetrics(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// compose a Data object for writing
	data := Data{
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
	metricStream.composeMetrics(data)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "aws_ec2_cpuutilization",
		map[string]interface{}{"max": 0.4366666666666666, "min": 0.3683333333333333, "sum": 1.9399999999999997, "count": 5.0},
		map[string]string{"AutoScalingGroupName": "test-autoscaling-group", "accountId": "546734499701", "region": "us-west-2"},
	)
}

func TestComposeAPICompatibleMetrics(t *testing.T) {
	metricStream := newTestCompatibleCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	// compose a Data object for writing
	data := Data{
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
	metricStream.composeMetrics(data)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "aws_ec2_cpuutilization",
		map[string]interface{}{"maximum": 0.4366666666666666, "minimum": 0.3683333333333333, "sum": 1.9399999999999997, "samplecount": 5.0},
		map[string]string{"AutoScalingGroupName": "test-autoscaling-group", "accountId": "546734499701", "region": "us-west-2"},
	)
}

// post GZIP encoded data to the metric stream listener
func TestWriteHTTPGzippedData(t *testing.T) {
	metricStream := newTestCloudWatchMetricStreams()

	acc := &testutil.Accumulator{}
	require.NoError(t, metricStream.Init())
	require.NoError(t, metricStream.Start(acc))
	defer metricStream.Stop()

	data, err := os.ReadFile("./testdata/records.gz")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", createURL("http", "/write"), bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
}
