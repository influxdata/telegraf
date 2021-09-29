package websocket

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"

	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// testSerializer serializes to a number of metrics to simplify tests here.
type testSerializer struct{}

func newTestSerializer() *testSerializer {
	return &testSerializer{}
}

func (t testSerializer) Serialize(_ telegraf.Metric) ([]byte, error) {
	return []byte("1"), nil
}

func (t testSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	return []byte(strconv.Itoa(len(metrics))), nil
}

type testServer struct {
	*httptest.Server
	t                *testing.T
	messages         chan []byte
	upgradeDelay     time.Duration
	expectTextFrames bool
}

func newTestServer(t *testing.T, messages chan []byte, tls bool) *testServer {
	s := &testServer{}
	s.t = t
	if tls {
		s.Server = httptest.NewTLSServer(s)
	} else {
		s.Server = httptest.NewServer(s)
	}
	s.URL = makeWsProto(s.Server.URL)
	s.messages = messages
	return s
}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

const (
	testHeaderName  = "X-Telegraf-Test"
	testHeaderValue = "1"
)

var testUpgrader = ws.Upgrader{}

func (s *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(testHeaderName) != testHeaderValue {
		s.t.Fatalf("expected test header found in request, got: %#v", r.Header)
	}
	if s.upgradeDelay > 0 {
		// Emulate long handshake.
		select {
		case <-r.Context().Done():
			return
		case <-time.After(s.upgradeDelay):
		}
	}
	conn, err := testUpgrader.Upgrade(w, r, http.Header{})
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if s.expectTextFrames && messageType != ws.TextMessage {
			s.t.Fatalf("unexpected frame type: %d", messageType)
		}
		select {
		case s.messages <- data:
		case <-time.After(5 * time.Second):
			s.t.Fatal("timeout writing to messages channel, make sure there are readers")
		}
	}
}

func initWebSocket(s *testServer) *WebSocket {
	w := newWebSocket()
	w.Log = testutil.Logger{}
	w.URL = s.URL
	w.Headers = map[string]string{testHeaderName: testHeaderValue}
	w.SetSerializer(newTestSerializer())
	return w
}

func connect(t *testing.T, w *WebSocket) {
	err := w.Connect()
	require.NoError(t, err)
}

func TestWebSocket_NoURL(t *testing.T) {
	w := newWebSocket()
	err := w.Init()
	require.ErrorIs(t, err, errInvalidURL)
}

func TestWebSocket_Connect_Timeout(t *testing.T) {
	s := newTestServer(t, nil, false)
	s.upgradeDelay = time.Second
	defer s.Close()
	w := initWebSocket(s)
	w.ConnectTimeout = config.Duration(10 * time.Millisecond)
	err := w.Connect()
	require.Error(t, err)
}

func TestWebSocket_Connect_OK(t *testing.T) {
	s := newTestServer(t, nil, false)
	defer s.Close()
	w := initWebSocket(s)
	connect(t, w)
}

func TestWebSocket_ConnectTLS_OK(t *testing.T) {
	s := newTestServer(t, nil, true)
	defer s.Close()
	w := initWebSocket(s)
	w.ClientConfig.InsecureSkipVerify = true
	connect(t, w)
}

func TestWebSocket_Write_OK(t *testing.T) {
	messages := make(chan []byte, 1)

	s := newTestServer(t, messages, false)
	defer s.Close()

	w := initWebSocket(s)
	connect(t, w)

	var metrics []telegraf.Metric
	metrics = append(metrics, testutil.TestMetric(0.4, "test"))
	metrics = append(metrics, testutil.TestMetric(0.5, "test"))
	err := w.Write(metrics)
	require.NoError(t, err)

	select {
	case data := <-messages:
		require.Equal(t, []byte("2"), data)
	case <-time.After(time.Second):
		t.Fatal("timeout receiving data")
	}
}

func TestWebSocket_Write_Error(t *testing.T) {
	s := newTestServer(t, nil, false)
	defer s.Close()

	w := initWebSocket(s)
	connect(t, w)

	require.NoError(t, w.conn.Close())

	metrics := []telegraf.Metric{testutil.TestMetric(0.4, "test")}
	err := w.Write(metrics)
	require.Error(t, err)
	require.Nil(t, w.conn)
}

func TestWebSocket_Write_Reconnect(t *testing.T) {
	messages := make(chan []byte, 1)
	s := newTestServer(t, messages, false)
	s.expectTextFrames = true // Also use text frames in this test.
	defer s.Close()

	w := initWebSocket(s)
	w.UseTextFrames = true
	connect(t, w)

	metrics := []telegraf.Metric{testutil.TestMetric(0.4, "test")}

	require.NoError(t, w.conn.Close())

	err := w.Write(metrics)
	require.Error(t, err)
	require.Nil(t, w.conn)

	err = w.Write(metrics)
	require.NoError(t, err)

	select {
	case data := <-messages:
		require.Equal(t, []byte("1"), data)
	case <-time.After(time.Second):
		t.Fatal("timeout receiving data")
	}
}

func TestWebSocket_Close(t *testing.T) {
	s := newTestServer(t, nil, false)
	defer s.Close()

	w := initWebSocket(s)
	connect(t, w)
	require.NoError(t, w.Close())
	// Check no error on second close.
	require.NoError(t, w.Close())
}
