package http_response

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Receives a list with fields that are expected to be absent
func checkAbsentFields(t *testing.T, fields []string, acc *testutil.Accumulator) {
	for _, field := range fields {
		ok := acc.HasField("http_response", field)
		require.False(t, ok)
	}
}

// Receives a list with tags that are expected to be absent
func checkAbsentTags(t *testing.T, tags []string, acc *testutil.Accumulator) {
	for _, tag := range tags {
		ok := acc.HasTag("http_response", tag)
		require.False(t, ok)
	}
}

// Receives a dictionary and with expected fields and their values. If a value is nil, it will only check
// that the field exists, but not its contents
func checkFields(t *testing.T, fields map[string]interface{}, acc *testutil.Accumulator) {
	for key, field := range fields {
		switch v := field.(type) {
		case int:
			value, ok := acc.IntField("http_response", key)
			require.True(t, ok)
			require.Equal(t, field, value)
		case float64:
			value, ok := acc.FloatField("http_response", key)
			require.True(t, ok)
			require.Equal(t, field, value)
		case string:
			value, ok := acc.StringField("http_response", key)
			require.True(t, ok)
			require.Equal(t, field, value)
		case nil:
			ok := acc.HasField("http_response", key)
			require.True(t, ok)
		default:
			t.Log("Unsupported type for field: ", v)
			t.Fail()
		}
	}
}

// Receives a dictionary and with expected tags and their values. If a value is nil, it will only check
// that the tag exists, but not its contents
func checkTags(t *testing.T, tags map[string]interface{}, acc *testutil.Accumulator) {
	for key, tag := range tags {
		switch v := tag.(type) {
		case string:
			ok := acc.HasTag("http_response", key)
			require.True(t, ok)
			require.Equal(t, tag, acc.TagValue("http_response", key))
		case nil:
			ok := acc.HasTag("http_response", key)
			require.True(t, ok)
		default:
			t.Log("Unsupported type for tag: ", v)
			t.Fail()
		}
	}
}

func setUpTestMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/good", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/good", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hit the good page!")
	})
	mux.HandleFunc("/jsonresponse", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "\"service_status\": \"up\", \"healthy\" : \"true\"")
	})
	mux.HandleFunc("/badredirect", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/badredirect", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/mustbepostmethod", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.Error(w, "method wasn't post", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprintf(w, "used post correctly!")
	})
	mux.HandleFunc("/musthaveabody", func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			http.Error(w, "couldn't read request body", http.StatusBadRequest)
			return
		}
		if string(body) == "" {
			http.Error(w, "body was empty", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "sent a body!")
	})
	mux.HandleFunc("/twosecondnap", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Second * 2)
		return
	})
	return mux
}

func checkOutput(t *testing.T, acc *testutil.Accumulator, presentFields map[string]interface{}, presentTags map[string]interface{}, absentFields []string, absentTags []string) {
	if presentFields != nil {
		checkFields(t, presentFields, acc)
	}

	if presentTags != nil {
		checkTags(t, presentTags, acc)
	}

	if absentFields != nil {
		checkAbsentFields(t, absentFields, acc)
	}

	if absentTags != nil {
		checkAbsentTags(t, absentTags, acc)
	}
}

func TestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cHeader := r.Header.Get("Content-Type")
		assert.Equal(t, "Hello", r.Host)
		assert.Equal(t, "application/json", cHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL,
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 2},
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Host":         "Hello",
		},
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code": http.StatusOK,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "success",
	}
	absentFields := []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)
}

func TestFields(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL + "/good",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}

	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code": http.StatusOK,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "success",
	}
	absentFields := []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)
}

func TestRedirects(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL + "/redirect",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code": http.StatusOK,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "success",
	}
	absentFields := []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)

	h = &HTTPResponse{
		Address:         ts.URL + "/badredirect",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	acc = testutil.Accumulator{}
	err = h.Gather(&acc)
	require.NoError(t, err)

	expectedFields = map[string]interface{}{
		"result_type": "connection_failed",
		"result_code": 3,
	}
	expectedTags = map[string]interface{}{
		"server": nil,
		"method": "GET",
		"result": "connection_failed",
	}
	absentFields = []string{"http_response_code", "response_time", "response_string_match"}
	absentTags := []string{"status_code"}
	checkOutput(t, &acc, expectedFields, expectedTags, nil, nil)

	expectedFields = map[string]interface{}{"result_type": "connection_failed"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, absentTags)
}

func TestMethod(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL + "/mustbepostmethod",
		Body:            "{ 'test': 'data'}",
		Method:          "POST",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code": http.StatusOK,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "POST",
		"status_code": "200",
		"result":      "success",
	}
	absentFields := []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)

	h = &HTTPResponse{
		Address:         ts.URL + "/mustbepostmethod",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	acc = testutil.Accumulator{}
	err = h.Gather(&acc)
	require.NoError(t, err)

	expectedFields = map[string]interface{}{
		"http_response_code": http.StatusMethodNotAllowed,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags = map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "405",
		"result":      "success",
	}
	absentFields = []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)

	//check that lowercase methods work correctly
	h = &HTTPResponse{
		Address:         ts.URL + "/mustbepostmethod",
		Body:            "{ 'test': 'data'}",
		Method:          "head",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	acc = testutil.Accumulator{}
	err = h.Gather(&acc)
	require.NoError(t, err)

	expectedFields = map[string]interface{}{
		"http_response_code": http.StatusMethodNotAllowed,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags = map[string]interface{}{
		"server":      nil,
		"method":      "head",
		"status_code": "405",
		"result":      "success",
	}
	absentFields = []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)
}

func TestBody(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL + "/musthaveabody",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code": http.StatusOK,
		"result_type":        "success",
		"result_code":        0,
		"response_time":      nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "success",
	}
	absentFields := []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)

	h = &HTTPResponse{
		Address:         ts.URL + "/musthaveabody",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	acc = testutil.Accumulator{}
	err = h.Gather(&acc)
	require.NoError(t, err)

	expectedFields = map[string]interface{}{
		"http_response_code": http.StatusBadRequest,
		"result_type":        "success",
		"result_code":        0,
	}
	expectedTags = map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "400",
		"result":      "success",
	}
	absentFields = []string{"response_string_match"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, nil)
}

func TestStringMatch(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:             ts.URL + "/good",
		Body:                "{ 'test': 'data'}",
		Method:              "GET",
		ResponseStringMatch: "hit the good page",
		ResponseTimeout:     internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code":    http.StatusOK,
		"response_string_match": 1,
		"result_type":           "success",
		"result_code":           0,
		"response_time":         nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "success",
	}
	checkOutput(t, &acc, expectedFields, expectedTags, nil, nil)
}

func TestStringMatchJson(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:             ts.URL + "/jsonresponse",
		Body:                "{ 'test': 'data'}",
		Method:              "GET",
		ResponseStringMatch: "\"service_status\": \"up\"",
		ResponseTimeout:     internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code":    http.StatusOK,
		"response_string_match": 1,
		"result_type":           "success",
		"result_code":           0,
		"response_time":         nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "success",
	}
	checkOutput(t, &acc, expectedFields, expectedTags, nil, nil)
}

func TestStringMatchFail(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:             ts.URL + "/good",
		Body:                "{ 'test': 'data'}",
		Method:              "GET",
		ResponseStringMatch: "hit the bad page",
		ResponseTimeout:     internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}

	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code":    http.StatusOK,
		"response_string_match": 0,
		"result_type":           "response_string_mismatch",
		"result_code":           1,
		"response_time":         nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "200",
		"result":      "response_string_mismatch",
	}
	checkOutput(t, &acc, expectedFields, expectedTags, nil, nil)
}

func TestTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test with sleep in short mode.")
	}

	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL + "/twosecondnap",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"result_type": "timeout",
		"result_code": 4,
	}
	expectedTags := map[string]interface{}{
		"server": nil,
		"method": "GET",
		"result": "timeout",
	}
	absentFields := []string{"http_response_code", "response_time", "response_string_match"}
	absentTags := []string{"status_code"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, absentTags)
}

func TestPluginErrors(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Bad regex test. Should return an error and return nothing
	h := &HTTPResponse{
		Address:             ts.URL + "/good",
		Body:                "{ 'test': 'data'}",
		Method:              "GET",
		ResponseStringMatch: "bad regex:[[",
		ResponseTimeout:     internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}

	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.Error(t, err)

	absentFields := []string{"http_response_code", "response_time", "response_string_match", "result_type", "result_code"}
	absentTags := []string{"status_code", "result", "server", "method"}
	checkOutput(t, &acc, nil, nil, absentFields, absentTags)

	// Attempt to read empty body test
	h = &HTTPResponse{
		Address:             ts.URL + "/redirect",
		Body:                "",
		Method:              "GET",
		ResponseStringMatch: ".*",
		ResponseTimeout:     internal.Duration{Duration: time.Second * 20},
		FollowRedirects:     false,
	}

	acc = testutil.Accumulator{}
	err = h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"http_response_code":    http.StatusMovedPermanently,
		"response_string_match": 0,
		"result_type":           "body_read_error",
		"result_code":           2,
		"response_time":         nil,
	}
	expectedTags := map[string]interface{}{
		"server":      nil,
		"method":      "GET",
		"status_code": "301",
		"result":      "body_read_error",
	}
	checkOutput(t, &acc, expectedFields, expectedTags, nil, nil)
}

func TestNetworkErrors(t *testing.T) {
	// DNS error
	h := &HTTPResponse{
		Address:         "https://nonexistent.nonexistent", // Any non-resolvable URL works here
		Body:            "",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		FollowRedirects: false,
	}

	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"result_type": "dns_error",
		"result_code": 5,
	}
	expectedTags := map[string]interface{}{
		"server": nil,
		"method": "GET",
		"result": "dns_error",
	}
	absentFields := []string{"http_response_code", "response_time", "response_string_match"}
	absentTags := []string{"status_code"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, absentTags)

	// Connecton failed
	h = &HTTPResponse{
		Address:         "https:/nonexistent.nonexistent", // Any non-routable IP works here
		Body:            "",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 5},
		FollowRedirects: false,
	}

	acc = testutil.Accumulator{}
	err = h.Gather(&acc)
	require.NoError(t, err)

	expectedFields = map[string]interface{}{
		"result_type": "connection_failed",
		"result_code": 3,
	}
	expectedTags = map[string]interface{}{
		"server": nil,
		"method": "GET",
		"result": "connection_failed",
	}
	absentFields = []string{"http_response_code", "response_time", "response_string_match"}
	absentTags = []string{"status_code"}
	checkOutput(t, &acc, expectedFields, expectedTags, absentFields, absentTags)
}
