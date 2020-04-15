package internal

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	client      = &http.Client{Timeout: time.Second * 10}
	reportError = errors.New("error: invalid Format or points")
)

const (
	contentType     = "Content-Type"
	contentEncoding = "Content-Encoding"
	authzHeader     = "Authorization"
	bearer          = "Bearer "
	gzipFormat      = "gzip"
	octetStream     = "application/octet-stream"
	reportEndpoint  = "/report"
	formatKey       = "f"
)

// The implementation of a Reporter that reports points directly to a Wavefront server.
type directReporter struct {
	serverURL string
	token     string
}

func NewDirectReporter(server string, token string) Reporter {
	return &directReporter{serverURL: server, token: token}
}

func (reporter directReporter) Report(format string, pointLines string) (*http.Response, error) {
	if format == "" || pointLines == "" {
		return nil, reportError
	}

	// compress
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(pointLines))
	if err != nil {
		zw.Close()
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}

	apiURL := reporter.serverURL + reportEndpoint
	req, err := http.NewRequest("POST", apiURL, &buf)
	req.Header.Set(contentType, octetStream)
	req.Header.Set(contentEncoding, gzipFormat)
	req.Header.Set(authzHeader, bearer+reporter.token)
	if err != nil {
		return &http.Response{}, err
	}

	q := req.URL.Query()
	q.Add(formatKey, format)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	return resp, nil
}

func (reporter directReporter) Server() string {
	return reporter.serverURL
}
