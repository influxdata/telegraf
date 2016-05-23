package http_response

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"

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
	fields, err := h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusOK, fields["http_response_code"])
	}
	assert.NotNil(t, fields["response_time"])
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
	fields, err := h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusOK, fields["http_response_code"])
	}
	assert.NotNil(t, fields["response_time"])
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
	fields, err := h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusOK, fields["http_response_code"])
	}

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
	fields, err = h.HTTPGather()
	require.Error(t, err)
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
	fields, err := h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusOK, fields["http_response_code"])
	}

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
	fields, err = h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusMethodNotAllowed, fields["http_response_code"])
	}

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
	fields, err = h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusMethodNotAllowed, fields["http_response_code"])
	}
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
	fields, err := h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusOK, fields["http_response_code"])
	}

	h = &HTTPResponse{
		Address:         ts.URL + "/musthaveabody",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 20},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	fields, err = h.HTTPGather()
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
	if assert.NotNil(t, fields["http_response_code"]) {
		assert.Equal(t, http.StatusBadRequest, fields["http_response_code"])
	}
}

func TestTimeout(t *testing.T) {
	mux := setUpTestMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	h := &HTTPResponse{
		Address:         ts.URL + "/twosecondnap",
		Body:            "{ 'test': 'data'}",
		Method:          "GET",
		ResponseTimeout: internal.Duration{Duration: time.Second * 1},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		FollowRedirects: true,
	}
	_, err := h.HTTPGather()
	require.Error(t, err)
}
