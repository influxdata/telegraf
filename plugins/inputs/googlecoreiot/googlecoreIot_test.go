package googlecoreiot

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

const (
	testMsg          = `{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}\n`
	testMsgNoNewline = `{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`

	testMsgs = `{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}
	{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}
	{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}
	{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}
	{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}
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
	acc.AssertContainsTaggedFields(t, "testingGoogle",
		map[string]interface{}{"temp_c": float64(23.95), "humidity": float64(62.83)},
		map[string]string{"projectId": "conference-demos", "deviceRegistryId": "my-registry", "sensor": "bme_280", "subscription": "projects/conference-demos/subscriptions/my-subscription", "deviceId": "myPi", "deviceNumId": "2808946627307959", "deviceRegistryLocation": "us-central1", "message_id_2": "204004313210337", "subFolder": "", "message_id": "204004313210337"},
	)

	// post multiple message to listener
	resp, err = http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsgs)))
	fmt.Println(acc.Metrics[0])
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
	acc.Wait(2)
	x := 0
	for x < len(testMsgs) {
		acc.AssertContainsTaggedFields(t, "testingGoogle",
			map[string]interface{}{"temp_c": float64(23.95), "humidity": float64(62.83)},
			map[string]string{"projectId": "conference-demos", "deviceRegistryId": "my-registry", "sensor": "bme_280", "subscription": "projects/conference-demos/subscriptions/my-subscription", "deviceId": "myPi", "deviceNumId": "2808946627307959", "deviceRegistryLocation": "us-central1", "message_id_2": "204004313210337", "subFolder": "", "message_id": "204004313210337"},
		)
		x++
	}

}

// http listener should add a newline at the end of the buffer if it's not there
func TestWriteHTTPNoNewline(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsgNoNewline)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "testingGoogle",
		map[string]interface{}{"temp_c": float64(23.95), "humidity": float64(62.83)},
		map[string]string{"projectId": "conference-demos", "deviceRegistryId": "my-registry", "sensor": "bme_280", "subscription": "projects/conference-demos/subscriptions/my-subscription", "deviceId": "myPi", "deviceNumId": "2808946627307959", "deviceRegistryLocation": "us-central1", "message_id_2": "204004313210337", "subFolder": "", "message_id": "204004313210337"},
	)
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

}

// writes 25,000 metrics to the listener with 10 different writers
func TestWriteHTTPHighTraffic(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping due to hang on darwin")
	}
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post many messages to listener
	for i := 0; i < 10; i++ {
		for i := 0; i < 2500; i++ {
			resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsg)))
			require.NoError(t, err)
			resp.Body.Close()
			require.EqualValues(t, 204, resp.StatusCode)
		}

	}
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

	// post empty message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(emptyMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)
}
