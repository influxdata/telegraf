package googlecoreiot

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

const (
	testMsg          = "{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}\n"
	testMsgNoNewline = "{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}"

	testMsgs = `{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}
{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}
{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}
{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}
{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}
`
	badMsg   = "blahblahblah: 42\n"
	emptyMsg = ""

	basicUsername = "test-username-please-ignore"
	basicPassword = "super-secure-password!"
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")
)

func newTestHTTPListener() *HTTPListener {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		TimeFunc:       time.Now,
	}
	return listener
}

func newTestHTTPSListener() *HTTPListener {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		ServerConfig:   *pki.TLSServerConfig(),
		TimeFunc:       time.Now,
	}

	return listener
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

func createURL(listener *HTTPListener, scheme string, path string, rawquery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:" + strconv.Itoa(listener.Port),
		Path:     path,
		RawQuery: rawquery,
	}
	return u.String()
}

func TestWriteHTTPSNoClientAuth(t *testing.T) {
	listener := newTestHTTPSListener()
	listener.TLSAllowedCACerts = nil

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	cas := x509.NewCertPool()
	cas.AppendCertsFromPEM([]byte(pki.ReadServerCert()))
	noClientAuthClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: cas,
			},
		},
	}

	// post single message to listener
	resp, err := noClientAuthClient.Post(createURL(listener, "https", "/write", ""), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteHTTPSWithClientAuth(t *testing.T) {
	listener := newTestHTTPSListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := getHTTPSClient().Post(createURL(listener, "https", "/write", ""), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteHTTP(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "bme_280",
		map[string]interface{}{"temp_c": float64(22.85)},
		map[string]string{"host": "server01"},
	)

	// post multiple message to listener
	resp, err = http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsgs)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(2)
	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "bme_280",
			map[string]interface{}{"temp_c": float64(22.85)},
			map[string]string{"host": hostTag},
		)
	}

	// Post a gigantic metric to the listener and verify that an error is returned:
	resp, err = http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(hugeMetric)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)

	acc.Wait(3)
	acc.AssertContainsTaggedFields(t, "cbme_280",
		map[string]interface{}{"temp_c": float64(22.85)},
		map[string]string{"host": "server01"},
	)
}

// http listener should add a newline at the end of the buffer if it's not there
func TestWriteHTTPNoNewline(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsgNoNewline)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "bme_280",
		map[string]interface{}{"temp_c": float64(22.85)},
		map[string]string{"host": "server01"},
	)
}

func TestWriteHTTPMaxLineSizeIncrease(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxLineSize:    128 * 1000,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// Post a gigantic metric to the listener and verify that it writes OK this time:
	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(hugeMetric)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteHTTPVerySmallMaxBody(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxBodySize:    4096,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(hugeMetric)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 413, resp.StatusCode)
}

func TestWriteHTTPVerySmallMaxLineSize(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxLineSize:    70,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsgs)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "bme_280",
			map[string]interface{}{"temp_c": float64(22.85)},
			map[string]string{"host": hostTag},
		)
	}
}

func TestWriteHTTPLargeLinesSkipped(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxLineSize:    100,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(hugeMetric+testMsgs)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "bme_280",
			map[string]interface{}{"temp_c": float64(22.85)},
			map[string]string{"host": hostTag},
		)
	}
}

// test that writing gzipped data works
func TestWriteHTTPGzippedData(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	data, err := ioutil.ReadFile("./testdata/testmsgs.gz")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", createURL(listener, "http", "/write", ""), bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "bme_280",
			map[string]interface{}{"temp_c": float64(22.85)},
			map[string]string{"host": hostTag},
		)
	}
}

// writes 25,000 metrics to the listener with 10 different writers
func TestWriteHTTPHighTraffic(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping due to hang on darwin")
	}
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post many messages to listener
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(innerwg *sync.WaitGroup) {
			defer innerwg.Done()
			for i := 0; i < 500; i++ {
				resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsgs)))
				require.NoError(t, err)
				resp.Body.Close()
				require.EqualValues(t, 204, resp.StatusCode)
			}
		}(&wg)
	}

	wg.Wait()
	listener.Gather(acc)

	acc.Wait(25000)
	require.Equal(t, int64(25000), int64(acc.NMetrics()))
}

func TestReceive404ForInvalidEndpoint(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/foobar", ""), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 404, resp.StatusCode)
}

func TestWriteHTTPInvalid(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(badMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)
}

func TestWriteHTTPEmpty(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(emptyMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

const hugeMetric = `{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}{\"attributes\": {\"deviceId\": \"myPi\", \"deviceNumId\":\"2808946627307959\", \"deviceRegistryId\":\"my-registry\", \"deviceRegistryLocation\":\"us-central1\", \"projectId\":\"conference-demos\"}, \"data\": \"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuODIsaHVtaWRpdHk9NjMuOTUgMTUzNjkzNDI4OTA1ODM5MDY4NQ==\", \"messageID\": \"193681629562075\", \"message_id\": \"193681629562075\", \"publishTime\": \"2018-09-14T14:11:29.096Z\", \"publish_time\": \"2018-09-14T14:11:29.096Z\"}, \"subscription\": \"projects/conference-demos/subscriptions/my-subscription\"}`
