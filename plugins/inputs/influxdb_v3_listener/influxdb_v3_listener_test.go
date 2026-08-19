package influxdb_v3_listener

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

const (
	testMsg = "cpu_load_short,host=server01 value=12.0 1422568543702900257\n"

	testMsgs = `cpu_load_short,host=server02 value=12.0 1422568543702900257
cpu_load_short,host=server03 value=12.0 1422568543702900257
cpu_load_short,host=server04 value=12.0 1422568543702900257
`

	// The second line is invalid, the first and the last one are valid
	testPartial = `cpu,host=a value=1 1422568543702900257
cpu,host=b value=1,value2=+Inf,value3=3 1422568543702900257
cpu,host=c value=1 1422568543702900257`

	token = "test-token-please-ignore"
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")

	parserTypes = []string{"internal", "upstream"}
)

// response is the fully read answer of the listener
type response struct {
	status int
	header http.Header
	body   string
}

func newTestListener() *InfluxDBV3Listener {
	return &InfluxDBV3Listener{
		ServiceAddress: "localhost:0",
		Statistics:     selfstat.NewCollector(nil),
		Log:            testutil.Logger{},
		timeFunc:       time.Now,
	}
}

func startListener(t *testing.T, listener *InfluxDBV3Listener, acc telegraf.Accumulator) {
	t.Helper()

	require.NoError(t, listener.Init())
	t.Cleanup(listener.Statistics.UnregisterAll)
	require.NoError(t, listener.Start(acc))
	t.Cleanup(listener.Stop)
}

func listenerURL(listener *InfluxDBV3Listener, scheme, path, rawQuery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     listener.listener.Addr().String(),
		Path:     path,
		RawQuery: rawQuery,
	}
	return u.String()
}

func do(t *testing.T, client *http.Client, req *http.Request) response {
	t.Helper()

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{status: resp.StatusCode, header: resp.Header, body: string(body)}
}

func getFrom(t *testing.T, address string) response {
	t.Helper()

	req, err := http.NewRequest("GET", address, nil)
	require.NoError(t, err)

	return do(t, http.DefaultClient, req)
}

func writeTo(t *testing.T, listener *InfluxDBV3Listener, rawQuery, body string) response {
	t.Helper()

	address := listenerURL(listener, "http", "/api/v3/write_lp", rawQuery)
	req, err := http.NewRequest("POST", address, strings.NewReader(body))
	require.NoError(t, err)

	return do(t, http.DefaultClient, req)
}

func TestInitInvalidParserType(t *testing.T) {
	listener := newTestListener()
	listener.ParserType = "unknown"

	require.ErrorContains(t, listener.Init(), `invalid parser type "unknown"`)
}

func TestWrite(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"cpu_load_short",
			map[string]string{"host": "server01"},
			map[string]any{"value": 12.0},
			time.Unix(0, 1422568543702900257),
		),
	}

	for _, parserType := range parserTypes {
		t.Run(parserType, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = parserType

			var acc testutil.Accumulator
			startListener(t, listener, &acc)

			resp := writeTo(t, listener, "db=mydb", testMsg)
			require.Equal(t, http.StatusNoContent, resp.status)

			acc.Wait(1)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestWriteMultipleMetrics(t *testing.T) {
	for _, parserType := range parserTypes {
		t.Run(parserType, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = parserType

			var acc testutil.Accumulator
			startListener(t, listener, &acc)

			resp := writeTo(t, listener, "db=mydb", testMsgs)
			require.Equal(t, http.StatusNoContent, resp.status)

			acc.Wait(3)
			require.Len(t, acc.GetTelegrafMetrics(), 3)
		})
	}
}

func TestWriteEmptyBody(t *testing.T) {
	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := writeTo(t, listener, "db=mydb", "")
	require.Equal(t, http.StatusNoContent, resp.status)
	require.Empty(t, acc.GetTelegrafMetrics())
}

func TestWriteDatabaseTag(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"cpu_load_short",
			map[string]string{"host": "server01", "database": "mydb"},
			map[string]any{"value": 12.0},
			time.Unix(0, 1422568543702900257),
		),
	}

	listener := newTestListener()
	listener.DatabaseTag = "database"

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := writeTo(t, listener, "db=mydb", testMsg)
	require.Equal(t, http.StatusNoContent, resp.status)

	acc.Wait(1)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestWriteDatabaseTagOverridesLine(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"cpu_load_short",
			map[string]string{"host": "server01", "database": "mydb"},
			map[string]any{"value": 12.0},
			time.Unix(0, 1422568543702900257),
		),
	}

	listener := newTestListener()
	listener.DatabaseTag = "database"

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	msg := "cpu_load_short,host=server01,database=wrongdb value=12.0 1422568543702900257\n"
	resp := writeTo(t, listener, "db=mydb", msg)
	require.Equal(t, http.StatusNoContent, resp.status)

	acc.Wait(1)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestWriteInvalidQueryParameters(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		expected string
	}{
		{
			name:     "no parameters at all",
			rawQuery: "",
			expected: "missing query parameter 'db'",
		},
		{
			name:     "missing database",
			rawQuery: "precision=auto",
			expected: "db name cannot be empty",
		},
		{
			name:     "empty database",
			rawQuery: "db=",
			expected: "db name cannot be empty",
		},
		{
			name:     "unknown precision",
			rawQuery: "db=mydb&precision=fortnight",
			expected: `unknown precision "fortnight"`,
		},
		{
			name:     "non-boolean accept_partial",
			rawQuery: "db=mydb&accept_partial=yes",
			expected: "provided string was not `true` or `false`, got \"yes\"",
		},
		{
			name:     "non-boolean no_sync",
			rawQuery: "db=mydb&no_sync=1",
			expected: "provided string was not `true` or `false`, got \"1\"",
		},
	}

	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := writeTo(t, listener, tt.rawQuery, testMsg)
			require.Equal(t, http.StatusBadRequest, resp.status)
			require.Contains(t, resp.body, tt.expected)
			require.Empty(t, acc.GetTelegrafMetrics())
		})
	}
}

func TestWriteNoSyncAccepted(t *testing.T) {
	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := writeTo(t, listener, "db=mydb&no_sync=true", testMsg)
	require.Equal(t, http.StatusNoContent, resp.status)

	acc.Wait(1)
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}

func TestWritePartial(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
	}{
		{
			name:     "accept_partial defaults to true",
			rawQuery: "db=mydb",
		},
		{
			name:     "accept_partial set to true",
			rawQuery: "db=mydb&accept_partial=true",
		},
	}

	for _, parserType := range parserTypes {
		for _, tt := range tests {
			t.Run(parserType+"/"+tt.name, func(t *testing.T) {
				listener := newTestListener()
				listener.ParserType = parserType

				var acc testutil.Accumulator
				startListener(t, listener, &acc)

				resp := writeTo(t, listener, tt.rawQuery, testPartial)
				require.Equal(t, http.StatusBadRequest, resp.status)
				require.Equal(t, "application/json", resp.header.Get("Content-Type"))

				var body struct {
					Error string      `json:"error"`
					Data  []lineError `json:"data"`
				}
				require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
				require.Equal(t, "partial write of line protocol occurred", body.Error)
				require.Len(t, body.Data, 1)
				require.Equal(t, 2, body.Data[0].LineNumber)
				require.Equal(t, "cpu,host=b value=1,value2=+Inf,value3=3 1422568543702900257", body.Data[0].OriginalLine)
				require.NotEmpty(t, body.Data[0].ErrorMessage)

				// The valid lines around the invalid one are still delivered
				acc.Wait(2)
				metrics := acc.GetTelegrafMetrics()
				require.Len(t, metrics, 2)
				require.Equal(t, "a", metrics[0].Tags()["host"])
				require.Equal(t, "c", metrics[1].Tags()["host"])
			})
		}
	}
}

func TestWritePartialRejected(t *testing.T) {
	for _, parserType := range parserTypes {
		t.Run(parserType, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = parserType

			var acc testutil.Accumulator
			startListener(t, listener, &acc)

			resp := writeTo(t, listener, "db=mydb&accept_partial=false", testPartial)
			require.Equal(t, http.StatusBadRequest, resp.status)

			var body struct {
				Error string    `json:"error"`
				Data  lineError `json:"data"`
			}
			require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
			require.Equal(t, "line protocol parsing error", body.Error)
			require.Equal(t, 2, body.Data.LineNumber)
			require.Equal(t, "cpu,host=b value=1,value2=+Inf,value3=3 1422568543702900257", body.Data.OriginalLine)

			// Nothing at all is delivered, not even the valid first line
			require.Empty(t, acc.GetTelegrafMetrics())
		})
	}
}

func TestWritePartialLineNumbers(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedLines []int
		expectedText  []string
	}{
		{
			name:          "first line invalid",
			body:          "invalid line\ncpu,host=a value=1 1\n",
			expectedLines: []int{1},
			expectedText:  []string{"invalid line"},
		},
		{
			name:          "last line invalid without trailing newline",
			body:          "cpu,host=a value=1 1\ninvalid line",
			expectedLines: []int{2},
			expectedText:  []string{"invalid line"},
		},
		{
			name:          "carriage returns are trimmed",
			body:          "cpu,host=a value=1 1\r\ninvalid line\r\n",
			expectedLines: []int{2},
			expectedText:  []string{"invalid line"},
		},
		{
			name:          "multiple invalid lines",
			body:          "invalid one\ncpu,host=a value=1 1\ninvalid two\n",
			expectedLines: []int{1, 3},
			expectedText:  []string{"invalid one", "invalid two"},
		},
		{
			name:          "newline inside a quoted field does not shift the count",
			body:          "cpu,host=a msg=\"first\nsecond\" 1\ninvalid line\n",
			expectedLines: []int{3},
			expectedText:  []string{"invalid line"},
		},
	}

	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := writeTo(t, listener, "db=mydb", tt.body)
			require.Equal(t, http.StatusBadRequest, resp.status)

			var body struct {
				Data []lineError `json:"data"`
			}
			require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

			lines := make([]int, 0, len(body.Data))
			text := make([]string, 0, len(body.Data))
			for _, e := range body.Data {
				lines = append(lines, e.LineNumber)
				text = append(text, e.OriginalLine)
			}
			require.Equal(t, tt.expectedLines, lines)
			require.Equal(t, tt.expectedText, text)
		})
	}
}

func TestWritePrecision(t *testing.T) {
	tests := []struct {
		name      string
		precision string
		timestamp string
		expected  time.Time
	}{
		{"second", "second", "1422568543", time.Unix(1422568543, 0)},
		{"second abbreviated", "s", "1422568543", time.Unix(1422568543, 0)},
		{"millisecond", "millisecond", "1422568543702", time.Unix(0, 1422568543702000000)},
		{"millisecond abbreviated", "ms", "1422568543702", time.Unix(0, 1422568543702000000)},
		{"microsecond", "microsecond", "1422568543702900", time.Unix(0, 1422568543702900000)},
		{"microsecond abbreviated", "us", "1422568543702900", time.Unix(0, 1422568543702900000)},
		{"microsecond single letter", "u", "1422568543702900", time.Unix(0, 1422568543702900000)},
		{"nanosecond", "nanosecond", "1422568543702900257", time.Unix(0, 1422568543702900257)},
		{"nanosecond abbreviated", "ns", "1422568543702900257", time.Unix(0, 1422568543702900257)},
		{"nanosecond single letter", "n", "1422568543702900257", time.Unix(0, 1422568543702900257)},
		{"auto detects seconds", "auto", "1422568543", time.Unix(1422568543, 0)},
		{"auto detects milliseconds", "auto", "1422568543702", time.Unix(0, 1422568543702000000)},
		{"auto detects microseconds", "auto", "1422568543702900", time.Unix(0, 1422568543702900000)},
		{"auto detects nanoseconds", "auto", "1422568543702900257", time.Unix(0, 1422568543702900257)},
		{"auto handles negative timestamps", "auto", "-1422568543", time.Unix(-1422568543, 0)},
		{"auto is the default", "", "1422568543", time.Unix(1422568543, 0)},
	}

	for _, parserType := range parserTypes {
		for _, tt := range tests {
			t.Run(parserType+"/"+tt.name, func(t *testing.T) {
				listener := newTestListener()
				listener.ParserType = parserType

				var acc testutil.Accumulator
				startListener(t, listener, &acc)

				rawQuery := "db=mydb"
				if tt.precision != "" {
					rawQuery += "&precision=" + tt.precision
				}
				resp := writeTo(t, listener, rawQuery, "cpu,host=a value=1 "+tt.timestamp+"\n")
				require.Equal(t, http.StatusNoContent, resp.status)

				acc.Wait(1)
				metrics := acc.GetTelegrafMetrics()
				require.Len(t, metrics, 1)
				require.Equal(t, tt.expected.UTC(), metrics[0].Time().UTC())
			})
		}
	}
}

func TestWritePrecisionAutoWithoutTimestamp(t *testing.T) {
	now := time.Unix(0, 1422568543702900257)

	for _, parserType := range parserTypes {
		t.Run(parserType, func(t *testing.T) {
			listener := newTestListener()
			listener.ParserType = parserType
			listener.timeFunc = func() time.Time { return now }

			var acc testutil.Accumulator
			startListener(t, listener, &acc)

			resp := writeTo(t, listener, "db=mydb&precision=auto", "cpu,host=a value=1\n")
			require.Equal(t, http.StatusNoContent, resp.status)

			acc.Wait(1)
			metrics := acc.GetTelegrafMetrics()
			require.Len(t, metrics, 1)
			require.Equal(t, now.UTC(), metrics[0].Time().UTC())
		})
	}
}

func TestWriteGzippedBody(t *testing.T) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write([]byte(testMsg))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	req, err := http.NewRequest("POST", listenerURL(listener, "http", "/api/v3/write_lp", "db=mydb"), &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	resp := do(t, http.DefaultClient, req)
	require.Equal(t, http.StatusNoContent, resp.status)

	acc.Wait(1)
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}

func TestWriteBodyTooLarge(t *testing.T) {
	listener := newTestListener()
	listener.MaxBodySize = config.Size(len(testMsg) - 1)

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := writeTo(t, listener, "db=mydb", testMsg)
	require.Equal(t, http.StatusRequestEntityTooLarge, resp.status)
	require.Equal(t, "http: request body too large", resp.header.Get("X-Influxdb-Error"))
	require.Empty(t, acc.GetTelegrafMetrics())
}

func TestWriteBodyTooLargeWithoutContentLength(t *testing.T) {
	listener := newTestListener()
	listener.MaxBodySize = config.Size(len(testMsg) - 1)

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	// Sending the body from a reader of unknown length makes the client use
	// chunked transfer encoding, so the size is only known while reading it
	address := listenerURL(listener, "http", "/api/v3/write_lp", "db=mydb")
	req, err := http.NewRequest("POST", address, io.NopCloser(strings.NewReader(testMsg)))
	require.NoError(t, err)

	resp := do(t, http.DefaultClient, req)
	require.Equal(t, http.StatusRequestEntityTooLarge, resp.status)
	require.Empty(t, acc.GetTelegrafMetrics())
}

func TestWriteUndeliveredMetricsRateLimit(t *testing.T) {
	listener := newTestListener()
	listener.MaxUndeliveredMetrics = 2

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	// A batch larger than the limit can never be delivered
	resp := writeTo(t, listener, "db=mydb", testMsgs)
	require.Equal(t, http.StatusRequestEntityTooLarge, resp.status)

	// Filling up the limit works, exceeding what is left is rate limited
	resp = writeTo(t, listener, "db=mydb", testMsg+testMsg)
	require.Equal(t, http.StatusNoContent, resp.status)

	resp = writeTo(t, listener, "db=mydb", testMsg)
	require.Equal(t, http.StatusTooManyRequests, resp.status)
}

func TestWriteAuth(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected int
	}{
		{"bearer scheme", "Bearer " + token, http.StatusNoContent},
		{"token scheme", "Token " + token, http.StatusNoContent},
		{"basic scheme", "Basic " + base64.StdEncoding.EncodeToString([]byte("user:"+token)), http.StatusNoContent},
		{"wrong token", "Bearer nope", http.StatusUnauthorized},
		{"wrong scheme", "Digest " + token, http.StatusUnauthorized},
		{"token without scheme", token, http.StatusUnauthorized},
		{"no authorization at all", "", http.StatusUnauthorized},
	}

	listener := newTestListener()
	listener.Token = config.NewSecret([]byte(token))

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := listenerURL(listener, "http", "/api/v3/write_lp", "db=mydb")
			req, err := http.NewRequest("POST", address, strings.NewReader(testMsg))
			require.NoError(t, err)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			require.Equal(t, tt.expected, do(t, http.DefaultClient, req).status)
		})
	}
}

func TestWriteWithoutTokenConfigured(t *testing.T) {
	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := writeTo(t, listener, "db=mydb", testMsg)
	require.Equal(t, http.StatusNoContent, resp.status)
}

func TestWriteSecureWithClientAuth(t *testing.T) {
	listener := newTestListener()
	listener.ServerConfig = *pki.TLSServerConfig()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	tlsConfig, err := pki.TLSClientConfig().TLSConfig()
	require.NoError(t, err)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	address := listenerURL(listener, "https", "/api/v3/write_lp", "db=mydb")
	req, err := http.NewRequest("POST", address, strings.NewReader(testMsg))
	require.NoError(t, err)

	require.Equal(t, http.StatusNoContent, do(t, client, req).status)

	acc.Wait(1)
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}

func TestWriteSecureNoClientAuth(t *testing.T) {
	listener := newTestListener()
	listener.ServerConfig = *pki.TLSServerConfig()
	listener.TLSAllowedCACerts = nil

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	cas := x509.NewCertPool()
	require.True(t, cas.AppendCertsFromPEM([]byte(pki.ReadServerCert())))
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: cas}}}

	address := listenerURL(listener, "https", "/api/v3/write_lp", "db=mydb")
	req, err := http.NewRequest("POST", address, strings.NewReader(testMsg))
	require.NoError(t, err)

	require.Equal(t, http.StatusNoContent, do(t, client, req).status)
}

func TestHealth(t *testing.T) {
	for _, path := range []string{"/health", "/api/v1/health"} {
		t.Run(path, func(t *testing.T) {
			listener := newTestListener()

			var acc testutil.Accumulator
			startListener(t, listener, &acc)

			resp := getFrom(t, listenerURL(listener, "http", path, ""))
			require.Equal(t, http.StatusOK, resp.status)
			require.Equal(t, "OK", resp.body)
		})
	}
}

func TestHealthUndeliveredMetricsAboveLimit(t *testing.T) {
	listener := newTestListener()
	listener.MaxUndeliveredMetrics = 1

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	require.Equal(t, http.StatusNoContent, writeTo(t, listener, "db=mydb", testMsg).status)

	resp := getFrom(t, listenerURL(listener, "http", "/health", ""))
	require.Equal(t, http.StatusServiceUnavailable, resp.status)
	require.Contains(t, resp.body, "pending undelivered metrics (1) is above limit")
}

func TestPing(t *testing.T) {
	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := getFrom(t, listenerURL(listener, "http", "/ping", ""))
	require.Equal(t, http.StatusOK, resp.status)
	require.Equal(t, "telegraf", resp.header.Get("X-Influxdb-Build"))

	var body map[string]string
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	require.Equal(t, "telegraf", body["product_name"])
	require.NotEmpty(t, body["process_id"])
}

func TestUnknownPath(t *testing.T) {
	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := getFrom(t, listenerURL(listener, "http", "/api/v2/write", "bucket=mybucket"))
	require.Equal(t, http.StatusNotFound, resp.status)
}

func TestUnknownPathRequiresAuth(t *testing.T) {
	listener := newTestListener()
	listener.Token = config.NewSecret([]byte(token))

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	resp := getFrom(t, listenerURL(listener, "http", "/api/v2/write", "bucket=mybucket"))
	require.Equal(t, http.StatusUnauthorized, resp.status)
}

func TestWriteConcurrentRequests(t *testing.T) {
	const requests = 50

	listener := newTestListener()

	var acc testutil.Accumulator
	startListener(t, listener, &acc)

	type result struct {
		status int
		err    error
	}

	address := listenerURL(listener, "http", "/api/v3/write_lp", "db=mydb")
	results := make(chan result, requests)
	for i := range requests {
		go func() {
			body := "cpu,host=server" + strconv.Itoa(i) + " value=12.0 1422568543702900257\n"
			req, err := http.NewRequest("POST", address, strings.NewReader(body))
			if err != nil {
				results <- result{err: err}
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- result{err: err}
				return
			}
			defer resp.Body.Close()

			results <- result{status: resp.StatusCode}
		}()
	}

	for range requests {
		r := <-results
		require.NoError(t, r.err)
		require.Equal(t, http.StatusNoContent, r.status)
	}

	acc.Wait(requests)
	require.Len(t, acc.GetTelegrafMetrics(), requests)
}
