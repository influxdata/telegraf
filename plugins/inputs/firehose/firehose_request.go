package firehose

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

type record struct {
	EncodedData string `json:"data"`
}

type requestBody struct {
	RequestID string   `json:"requestId"`
	Timestamp int64    `json:"timestamp"`
	Records   []record `json:"records"`
}

// https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html#requestformat
type request struct {
	req  *http.Request
	body requestBody
	res  response
}

// https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html#responseformat
type response struct {
	body       responseBody
	statusCode int
}

type responseBody struct {
	RequestID    string `json:"requestId"`
	Timestamp    int64  `json:"timestamp"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (r *request) authenticate(expected config.Secret) error {
	if expected.Empty() {
		return nil
	}
	key := r.req.Header.Get("x-amz-firehose-access-key")
	match, err := expected.EqualTo([]byte(key))
	if err != nil {
		return fmt.Errorf("comparing keys failed: %w", err)
	}
	if !match {
		r.res.statusCode = http.StatusUnauthorized
		return fmt.Errorf("unauthorized request %s from %v", r.req.Header.Get("x-amz-firehose-request-id"), r.req.RemoteAddr)
	}
	return nil
}

func (r *request) validate() error {
	requestID := r.req.Header.Get("x-amz-firehose-request-id")
	if requestID == "" {
		r.res.statusCode = http.StatusBadRequest
		return errors.New("x-amz-firehose-request-id header is not set")
	}

	// The maximum body size can be up to a maximum of 64 MiB.
	// https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html
	if r.req.ContentLength > int64(64*1024*1024) {
		r.res.statusCode = http.StatusRequestEntityTooLarge
		return fmt.Errorf("content length too large in request %s", requestID)
	}

	switch r.req.Method {
	case http.MethodPost, http.MethodPut:
		// Do nothing, those methods are allowed
	default:
		r.res.statusCode = http.StatusMethodNotAllowed
		return fmt.Errorf("method %q in request %q is not allowed", r.req.Method, requestID)
	}

	contentType := r.req.Header.Get("content-type")
	if contentType != "application/json" {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("unaccepted content type, %s, in request %s", contentType, requestID)
	}

	encoding := r.req.Header.Get("content-encoding")
	body, err := internal.NewStreamContentDecoder(encoding, r.req.Body)
	if err != nil {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("creating decoder for %q failed: %w", encoding, err)
	}
	defer r.req.Body.Close()

	if err := json.NewDecoder(body).Decode(&r.body); err != nil {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("decode body failed: %w", err)
	}

	if requestID != r.body.RequestID {
		r.res.statusCode = http.StatusBadRequest
		return errors.New("requestId in the body does not match the value of the request header, x-amz-firehose-request-id")
	}

	return nil
}

func (r *request) decodeData() ([][]byte, error) {
	// decode base64-encoded data and return them as a slice of byte slices
	decodedData := make([][]byte, 0)
	for _, record := range r.body.Records {
		data, err := base64.StdEncoding.DecodeString(record.EncodedData)
		if err != nil {
			return nil, err
		}
		decodedData = append(decodedData, data)
	}
	return decodedData, nil
}

func (r *request) sendResponse(res http.ResponseWriter) error {
	var errorMessage string
	if r.res.statusCode != http.StatusOK {
		errorMessage = http.StatusText(r.res.statusCode)
	}
	r.res.body = responseBody{
		RequestID:    r.req.Header.Get("x-amz-firehose-request-id"),
		Timestamp:    time.Now().Unix(),
		ErrorMessage: errorMessage,
	}
	response, err := json.Marshal(r.res.body)
	if err != nil {
		return err
	}
	res.Header().Set("content-type", "application/json")
	res.WriteHeader(r.res.statusCode)
	_, err = res.Write(response)
	return err
}
