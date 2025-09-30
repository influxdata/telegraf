package influxdb_v2_listener

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const (
	testMsg = "cpu_load_short,host=server01 value=12.0 1422568543702900257\n"

	testMsgNoNewline = "cpu_load_short,host=server01 value=12.0 1422568543702900257"

	testMsgs = `cpu_load_short,host=server02 value=12.0 1422568543702900257
cpu_load_short,host=server03 value=12.0 1422568543702900257
cpu_load_short,host=server04 value=12.0 1422568543702900257
cpu_load_short,host=server05 value=12.0 1422568543702900257
cpu_load_short,host=server06 value=12.0 1422568543702900257
`
	testPartial = `cpu,host=a value1=1
cpu,host=b value1=1,value2=+Inf,value3=3
cpu,host=c value1=1`

	badMsg = "blahblahblah: 42\n"

	emptyMsg = ""

	token = "test-token-please-ignore"
)

var (
	pki             = testutil.NewPKI("../../../testutil/pki")
	parserTestCases = []struct {
		parser string
	}{
		{"upstream"},
		{"internal"},
	}
)

func newTestListener() *InfluxDBV2Listener {
	listener := &InfluxDBV2Listener{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		timeFunc:       time.Now,
	}
	return listener
}

func newTestAuthListener() *InfluxDBV2Listener {
	listener := newTestListener()
	listener.Token = config.NewSecret([]byte(token))
	return listener
}

func newRateLimitedTestListener(maxUndeliveredMetrics int) *InfluxDBV2Listener {
	listener := newTestListener()
	listener.MaxUndeliveredMetrics = maxUndeliveredMetrics
	return listener
}

func newTestSecureListener() *InfluxDBV2Listener {
	listener := &InfluxDBV2Listener{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		ServerConfig:   *pki.TLSServerConfig(),
		timeFunc:       time.Now,
	}

	return listener
}

func getSecureClient() *http.Client {
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

func createURL(listener *InfluxDBV2Listener, scheme, path, rawquery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:" + strconv.Itoa(listener.port),
		Path:     path,
		RawQuery: rawquery,
	}
	return u.String()
}

func TestWriteSecureNoClientAuth(t *testing.T) {
	listener := newTestSecureListener()
	listener.TLSAllowedCACerts = nil

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
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
	resp, err := noClientAuthClient.Post(createURL(listener, "https", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsg))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteSecureWithClientAuth(t *testing.T) {
	listener := newTestSecureListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := getSecureClient().Post(createURL(listener, "https", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsg))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteTokenAuth(t *testing.T) {
	listener := newTestAuthListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	client := &http.Client{}

	req, err := http.NewRequest("POST", createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), bytes.NewBufferString(testMsg))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Token "+token)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, http.StatusNoContent, resp.StatusCode)
}

func TestWriteKeepBucket(t *testing.T) {
	testMsgWithDB := "cpu_load_short,host=server01,bucketTag=wrongbucket value=12.0 1422568543702900257\n"

	listener := newTestListener()
	listener.BucketTag = "bucketTag"

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsg))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01", "bucketTag": "mybucket"},
	)

	// post single message to listener with a database tag in it already. It should be clobbered.
	resp, err = http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsgWithDB))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01", "bucketTag": "mybucket"},
	)

	// post multiple message to listener
	resp, err = http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsgs))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(2)
	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag, "bucketTag": "mybucket"},
		)
	}
}

// http listener should add a newline at the end of the buffer if it's not there
func TestWriteNoNewline(t *testing.T) {
	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = tc.parser

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			// post single message to listener
			resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsgNoNewline))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 204, resp.StatusCode)

			acc.Wait(1)
			acc.AssertContainsTaggedFields(t, "cpu_load_short",
				map[string]interface{}{"value": float64(12)},
				map[string]string{"host": "server01"},
			)
		})
	}
}

func TestAllOrNothing(t *testing.T) {
	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = tc.parser

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testPartial))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 400, resp.StatusCode)
		})
	}
}

func TestWriteMaxLineSizeIncrease(t *testing.T) {
	// The term 'master_repl' used here is archaic language from redis
	hugeMetric, err := os.ReadFile("./testdata/huge_metric")
	require.NoError(t, err)

	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := &InfluxDBV2Listener{
				Log:            testutil.Logger{},
				ServiceAddress: "localhost:0",
				timeFunc:       time.Now,
				ParserType:     tc.parser,
			}

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			// Post a gigantic metric to the listener and verify that it writes OK this time:
			resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBuffer(hugeMetric))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 204, resp.StatusCode)
		})
	}
}

func TestWriteVerySmallMaxBody(t *testing.T) {
	// The term 'master_repl' used here is archaic language from redis
	hugeMetric, err := os.ReadFile("./testdata/huge_metric")
	require.NoError(t, err)

	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := &InfluxDBV2Listener{
				Log:            testutil.Logger{},
				ServiceAddress: "localhost:0",
				MaxBodySize:    config.Size(4096),
				timeFunc:       time.Now,
				ParserType:     tc.parser,
			}

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBuffer(hugeMetric))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 413, resp.StatusCode)
		})
	}
}

func TestWriteLargeLine(t *testing.T) {
	// The term 'master_repl' used here is archaic language from redis
	hugeMetric, err := os.ReadFile("./testdata/huge_metric")
	require.NoError(t, err)
	hugeMetricString := string(hugeMetric)

	listener := &InfluxDBV2Listener{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
		timeFunc: func() time.Time {
			return time.Unix(123456789, 0)
		},
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(hugeMetricString+testMsgs))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	// TODO: with the new parser, long lines aren't a problem.  Do we need to skip them?
	// require.EqualValues(t, 400, resp.StatusCode)

	expected := testutil.MustMetric(
		"super_long_metric",
		map[string]string{"foo": "bar"},
		map[string]interface{}{
			"clients":                     42,
			"connected_followers":         43,
			"evicted_keys":                44,
			"expired_keys":                45,
			"instantaneous_ops_per_sec":   46,
			"keyspace_hitrate":            47.0,
			"keyspace_hits":               48,
			"keyspace_misses":             49,
			"latest_fork_usec":            50,
			"master_repl_offset":          51,
			"mem_fragmentation_ratio":     52.58,
			"pubsub_channels":             53,
			"pubsub_patterns":             54,
			"rdb_changes_since_last_save": 55,
			"repl_backlog_active":         56,
			"repl_backlog_histlen":        57,
			"repl_backlog_size":           58,
			"sync_full":                   59,
			"sync_partial_err":            60,
			"sync_partial_ok":             61,
			"total_commands_processed":    62,
			"total_connections_received":  63,
			"uptime":                      64,
			"used_cpu_sys":                65.07,
			"used_cpu_sys_children":       66.0,
			"used_cpu_user":               67.1,
			"used_cpu_user_children":      68.0,
			"used_memory":                 692048,
			"used_memory_lua":             70792,
			"used_memory_peak":            711128,
			"used_memory_rss":             7298144,
		},
		time.Unix(123456789, 0),
	)

	m, ok := acc.Get("super_long_metric")
	require.True(t, ok)
	testutil.RequireMetricEqual(t, expected, testutil.FromTestMetric(m))

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

// test that writing gzipped data works
func TestWriteGzippedData(t *testing.T) {
	listener := newTestListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	data, err := os.ReadFile("./testdata/testmsgs.gz")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

// writes 25,000 metrics to the listener with 10 different writers
func TestWriteHighTraffic(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Skipping due to hang on darwin")
	}
	listener := newTestListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post many messages to listener
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(innerwg *sync.WaitGroup) {
			defer innerwg.Done()
			for i := 0; i < 500; i++ {
				resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(testMsgs))
				if err != nil {
					return
				}
				if err := resp.Body.Close(); err != nil {
					return
				}
				if resp.StatusCode != 204 {
					return
				}
			}
		}(&wg)
	}

	wg.Wait()
	require.NoError(t, listener.Gather(acc))

	acc.Wait(25000)
	require.Equal(t, int64(25000), int64(acc.NMetrics()))
}

func TestReceive404ForInvalidEndpoint(t *testing.T) {
	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = tc.parser

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			// post single message to listener
			resp, err := http.Post(createURL(listener, "http", "/foobar", ""), "", bytes.NewBufferString(testMsg))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 404, resp.StatusCode)
		})
	}
}

func TestWriteInvalid(t *testing.T) {
	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = tc.parser

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			// post single message to listener
			resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(badMsg))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 400, resp.StatusCode)
		})
	}
}

func TestWriteEmpty(t *testing.T) {
	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = tc.parser

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			// post single message to listener
			resp, err := http.Post(createURL(listener, "http", "/api/v2/write", "bucket=mybucket"), "", bytes.NewBufferString(emptyMsg))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 204, resp.StatusCode)
		})
	}
}

func TestHealth(t *testing.T) {
	listener := newTestListener()
	listener.timeFunc = func() time.Time {
		return time.Unix(42, 123456789)
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post ping to listener
	resp, err := http.Get(createURL(listener, "http", "/api/v2/health", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyBytes), "\"status\":\"pass\"")
	require.Contains(t, string(bodyBytes), "\"message\":\"ready for queries and writes\"")
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
	require.EqualValues(t, 1, listener.healthsServed.Get())

	// check when max undelivered is set, but not reached
	listener.MaxUndeliveredMetrics = 1
	resp, err = http.Get(createURL(listener, "http", "/api/v2/health", ""))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
	require.EqualValues(t, 2, listener.healthsServed.Get())

	// and on the documented base endpoint
	resp, err = http.Get(createURL(listener, "http", "/health", ""))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
	require.EqualValues(t, 3, listener.healthsServed.Get())

	// post ping to listener with too many pending metrics
	listener.totalUndeliveredMetrics.Add(1)
	resp, err = http.Get(createURL(listener, "http", "/api/v2/health", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	bodyBytes, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyBytes), "\"status\":\"fail\"")
	require.Contains(t, string(bodyBytes), "\"message\":\"pending undelivered metrics (1) is above limit\"")
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 503, resp.StatusCode)
	require.EqualValues(t, 4, listener.healthsServed.Get())
}

func TestPing(t *testing.T) {
	listener := newTestListener()
	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Get(createURL(listener, "http", "/ping", ""))
	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestReady(t *testing.T) {
	listener := newTestListener()
	listener.timeFunc = func() time.Time {
		return time.Unix(42, 123456789)
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post ping to listener
	resp, err := http.Get(createURL(listener, "http", "/api/v2/ready", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyBytes), "\"status\":\"ready\"")
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
	require.EqualValues(t, 1, listener.readysServed.Get())

	// and on the documented base endpoint
	resp, err = http.Get(createURL(listener, "http", "/ready", ""))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 200, resp.StatusCode)
	require.EqualValues(t, 2, listener.readysServed.Get())
}

func TestWriteWithPrecision(t *testing.T) {
	for _, tc := range parserTestCases {
		t.Run("parser "+tc.parser, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = tc.parser

			acc := &testutil.Accumulator{}
			require.NoError(t, listener.Init())
			require.NoError(t, listener.Start(acc))
			defer listener.Stop()

			msg := "xyzzy value=42 1422568543\n"
			resp, err := http.Post(
				createURL(listener, "http", "/api/v2/write", "bucket=mybucket&precision=s"), "", bytes.NewBufferString(msg))
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.EqualValues(t, 204, resp.StatusCode)

			acc.Wait(1)
			// When timestamp is provided, the precision parameter is
			// overloaded to specify the timestamp's unit
			require.Equal(t, time.Unix(0, 1422568543000000000), acc.Metrics[0].Time)
		})
	}
}

func TestWriteWithPrecisionNoTimestamp(t *testing.T) {
	listener := newTestListener()
	listener.timeFunc = func() time.Time {
		return time.Unix(42, 123456789)
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	msg := "xyzzy value=42\n"
	resp, err := http.Post(
		createURL(listener, "http", "/api/v2/write", "bucket=mybucket&precision=s"), "", bytes.NewBufferString(msg))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	require.Len(t, acc.Metrics, 1)
	// When timestamp is omitted, the precision parameter actually
	// specifies the precision.  The timestamp is set to the greatest
	// integer unit less than the provided timestamp (floor).
	require.Equal(t, time.Unix(42, 0), acc.Metrics[0].Time)
}

func TestRateLimitedConnectionDropsSecondRequest(t *testing.T) {
	listener := newRateLimitedTestListener(1)
	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	msg := "xyzzy value=42\n"
	postURL := createURL(listener, "http", "/api/v2/write", "bucket=mybucket&precision=s")
	resp, err := http.Post(postURL, "", bytes.NewBufferString(msg)) // #nosec G107 -- url has to be dynamic due to dynamic port number
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	resp, err = http.Post(postURL, "", bytes.NewBufferString(msg)) // #nosec G107 -- url has to be dynamic due to dynamic port number
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 429, resp.StatusCode)
}

func TestRateLimitedConnectionAcceptsNewRequestOnDelivery(t *testing.T) {
	listener := newRateLimitedTestListener(1)
	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	msg := "xyzzy value=42\n"
	postURL := createURL(listener, "http", "/api/v2/write", "bucket=mybucket&precision=s")
	resp, err := http.Post(postURL, "", bytes.NewBufferString(msg)) // #nosec G107 -- url has to be dynamic due to dynamic port number
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)

	ms := acc.GetTelegrafMetrics()
	for _, m := range ms {
		m.Accept()
	}

	resp, err = http.Post(postURL, "", bytes.NewBufferString(msg)) // #nosec G107 -- url has to be dynamic due to dynamic port number
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestRateLimitedConnectionRejectsBatchesLargerThanMaxUndeliveredMetrics(t *testing.T) {
	listener := newRateLimitedTestListener(1)
	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Init())
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	msg := "xyzzy value=42\nxyzzy value=43"
	postURL := createURL(listener, "http", "/api/v2/write", "bucket=mybucket&precision=s")
	resp, err := http.Post(postURL, "", bytes.NewBufferString(msg)) // #nosec G107 -- url has to be dynamic due to dynamic port number
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.EqualValues(t, 413, resp.StatusCode)
}

// The term 'master_repl' used here is archaic language from redis
