package client

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient_Write(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			// test form values:
			if r.FormValue("db") != "test" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"wrong db name"}`)
			}
			if r.FormValue("rp") != "policy" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"wrong rp name"}`)
			}
			if r.FormValue("precision") != "ns" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"wrong precision"}`)
			}
			if r.FormValue("consistency") != "all" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"wrong consistency"}`)
			}
			// test that user agent is set properly
			if r.UserAgent() != "test-agent" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"wrong agent name"}`)
			}
			// test basic auth params
			user, pass, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"basic auth not set"}`)
			}
			if user != "test-user" || pass != "test-password" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"basic auth incorrect"}`)
			}

			// test that user-specified http header is set properly
			if r.Header.Get("X-Test-Header") != "Test-Value" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"results":[{}],"error":"wrong http header value"}`)
			}

			// Validate Content-Length Header
			if r.ContentLength != 13 {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				msg := fmt.Sprintf(`{"results":[{}],"error":"Content-Length: expected [13], got [%d]"}`, r.ContentLength)
				fmt.Fprintln(w, msg)
			}

			// Validate the request body:
			buf := make([]byte, 100)
			n, _ := r.Body.Read(buf)
			expected := "cpu value=99"
			got := string(buf[0 : n-1])
			if expected != got {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				msg := fmt.Sprintf(`{"results":[{}],"error":"expected [%s], got [%s]"}`, expected, got)
				fmt.Fprintln(w, msg)
			}

			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("Content-Type", "application/json")
		case "/query":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}]}`)
		}
	}))
	defer ts.Close()

	config := HTTPConfig{
		URL:       ts.URL,
		UserAgent: "test-agent",
		Username:  "test-user",
		Password:  "test-password",
		HTTPHeaders: HTTPHeaders{
			"X-Test-Header": "Test-Value",
		},
	}
	wp := WriteParams{
		Database:        "test",
		RetentionPolicy: "policy",
		Precision:       "ns",
		Consistency:     "all",
	}
	client, err := NewHTTP(config, wp)
	defer client.Close()
	assert.NoError(t, err)

	err = client.WriteStream(bytes.NewReader([]byte("cpu value=99\n")))
	assert.NoError(t, err)
}

func TestHTTPClient_Write_Errors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			w.WriteHeader(http.StatusTeapot)
		case "/query":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}]}`)
		}
	}))
	defer ts.Close()

	config := HTTPConfig{
		URL: ts.URL,
	}
	defaultWP := WriteParams{
		Database: "test",
	}
	client, err := NewHTTP(config, defaultWP)
	defer client.Close()
	assert.NoError(t, err)

	lp := []byte("cpu value=99\n")
	err = client.WriteStream(bytes.NewReader(lp))
	assert.Error(t, err)
}

func TestNewHTTPErrors(t *testing.T) {
	// No URL:
	config := HTTPConfig{}
	defaultWP := WriteParams{
		Database: "test",
	}
	client, err := NewHTTP(config, defaultWP)
	assert.Error(t, err)
	assert.Nil(t, client)

	// No Database:
	config = HTTPConfig{
		URL: "http://localhost:8086",
	}
	defaultWP = WriteParams{}
	client, err = NewHTTP(config, defaultWP)
	assert.Nil(t, client)
	assert.Error(t, err)

	// Invalid URL:
	config = HTTPConfig{
		URL: "http://192.168.0.%31:8080/",
	}
	defaultWP = WriteParams{
		Database: "test",
	}
	client, err = NewHTTP(config, defaultWP)
	assert.Nil(t, client)
	assert.Error(t, err)

	// Invalid URL scheme:
	config = HTTPConfig{
		URL: "mailto://localhost:8086",
	}
	defaultWP = WriteParams{
		Database: "test",
	}
	client, err = NewHTTP(config, defaultWP)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestHTTPClient_Query(t *testing.T) {
	command := "CREATE DATABASE test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			w.WriteHeader(http.StatusNoContent)
		case "/query":
			// validate the create database command is correct
			got := r.FormValue("q")
			if got != command {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
				msg := fmt.Sprintf(`{"results":[{}],"error":"got %s, expected %s"}`, got, command)
				fmt.Fprintln(w, msg)
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}]}`)
		}
	}))
	defer ts.Close()

	config := HTTPConfig{
		URL: ts.URL,
	}
	defaultWP := WriteParams{
		Database: "test",
	}
	client, err := NewHTTP(config, defaultWP)
	defer client.Close()
	assert.NoError(t, err)
	err = client.Query(command)
	assert.NoError(t, err)
}

func TestHTTPClient_Query_ResponseError(t *testing.T) {
	command := "CREATE DATABASE test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			w.WriteHeader(http.StatusNoContent)
		case "/query":
			w.WriteHeader(http.StatusTeapot)
			w.Header().Set("Content-Type", "application/json")
			msg := fmt.Sprintf(`{"results":[{}],"error":"couldnt create database"}`)
			fmt.Fprintln(w, msg)
		}
	}))
	defer ts.Close()

	config := HTTPConfig{
		URL: ts.URL,
	}
	defaultWP := WriteParams{
		Database: "test",
	}
	client, err := NewHTTP(config, defaultWP)
	defer client.Close()
	assert.NoError(t, err)
	err = client.Query(command)
	assert.Error(t, err)
}

func TestHTTPClient_Query_JSONDecodeError(t *testing.T) {
	command := "CREATE DATABASE test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			w.WriteHeader(http.StatusNoContent)
		case "/query":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			// write JSON missing a ']'
			msg := fmt.Sprintf(`{"results":[{}}`)
			fmt.Fprintln(w, msg)
		}
	}))
	defer ts.Close()

	config := HTTPConfig{
		URL: ts.URL,
	}
	defaultWP := WriteParams{
		Database: "test",
	}
	client, err := NewHTTP(config, defaultWP)
	defer client.Close()
	assert.NoError(t, err)
	err = client.Query(command)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}

func TestGzipCompression(t *testing.T) {
	influxLine := "cpu value=99\n"

	// Compress the payload using GZIP.
	payload := bytes.NewReader([]byte(influxLine))
	compressed, err := compressWithGzip(payload)
	assert.Nil(t, err)

	// Decompress the compressed payload and make sure
	// that its original value has not changed.
	gr, err := gzip.NewReader(compressed)
	assert.Nil(t, err)
	gr.Close()

	var uncompressed bytes.Buffer
	_, err = uncompressed.ReadFrom(gr)
	assert.Nil(t, err)

	assert.Equal(t, []byte(influxLine), uncompressed.Bytes())
}

func TestHTTPClient_PathPrefix(t *testing.T) {
	prefix := "/some/random/prefix"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case prefix + "/write":
			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("Content-Type", "application/json")
		case prefix + "/query":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			msg := fmt.Sprintf("Path not found: %s", r.URL.Path)
			fmt.Fprintln(w, msg)
		}
	}))
	defer ts.Close()

	config := HTTPConfig{
		URL: ts.URL + prefix,
	}
	wp := WriteParams{
		Database: "test",
	}
	client, err := NewHTTP(config, wp)
	defer client.Close()
	assert.NoError(t, err)
	err = client.Query("CREATE DATABASE test")
	assert.NoError(t, err)
	err = client.WriteStream(bytes.NewReader([]byte("cpu value=99\n")))
	assert.NoError(t, err)
}
