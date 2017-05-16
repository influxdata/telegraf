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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)

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
	require.Error(t, err)

	value, ok = acc.IntField("http_response", "http_response_code")
	require.False(t, ok)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)

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

	value, ok = acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, value)

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

	value, ok = acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, value)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)

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

	value, ok = acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusBadRequest, value)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)
	value, ok = acc.IntField("http_response", "response_string_match")
	require.True(t, ok)
	require.Equal(t, 1, value)
	_, ok = acc.FloatField("http_response", "response_time")
	require.True(t, ok)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)
	value, ok = acc.IntField("http_response", "response_string_match")
	require.True(t, ok)
	require.Equal(t, 1, value)
	_, ok = acc.FloatField("http_response", "response_time")
	require.True(t, ok)
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

	value, ok := acc.IntField("http_response", "http_response_code")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, value)
	value, ok = acc.IntField("http_response", "response_string_match")
	require.True(t, ok)
	require.Equal(t, 0, value)
	_, ok = acc.FloatField("http_response", "response_time")
	require.True(t, ok)
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
		ResponseTimeout: internal.Duration{Duration: time.Millisecond},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	var acc testutil.Accumulator
	err := h.Gather(&acc)
	require.NoError(t, err)

	ok := acc.HasIntField("http_response", "http_response_code")
	require.False(t, ok)
}
