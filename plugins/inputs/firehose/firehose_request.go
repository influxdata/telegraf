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

func newFirehoseRequest(req *http.Request) (*request, error) {
	r := &request{req: req}
	requestID := r.req.Header.Get("x-amz-firehose-request-id")
	if requestID == "" {
		r.res.statusCode = http.StatusBadRequest
		return r, errors.New("x-amz-firehose-request-id header is not set")
	}
	// Set a default response status code
	r.res.statusCode = http.StatusInternalServerError

	encoding := r.req.Header.Get("content-encoding")
	body, err := internal.NewStreamContentDecoder(encoding, r.req.Body)
	if err != nil {
		r.res.statusCode = http.StatusBadRequest
		return r, fmt.Errorf("creating %q decoder for request %q failed: %w", encoding, requestID, err)
	}
	defer r.req.Body.Close()

	if err := json.NewDecoder(body).Decode(&r.body); err != nil {
		r.res.statusCode = http.StatusBadRequest
		return r, fmt.Errorf("decode body for request %q failed: %w", requestID, err)
	}

	if requestID != r.body.RequestID {
		r.res.statusCode = http.StatusBadRequest
		return r, fmt.Errorf("mismatch between requestID in the request header (%q) and the request body (%s)", requestID, r.body.RequestID)
	}

	return r, nil
}

func (r *request) authenticate(expected config.Secret) error {
	// We completely switch off authentication if no 'access_key' was provided in the config, it's intended!
	if expected.Empty() {
		return nil
	}
	key := r.req.Header.Get("x-amz-firehose-access-key")
	match, err := expected.EqualTo([]byte(key))
	if err != nil {
		r.res.statusCode = http.StatusInternalServerError
		return fmt.Errorf("comparing keys failed: %w", err)
	}
	if !match {
		r.res.statusCode = http.StatusUnauthorized
		return fmt.Errorf("unauthorized request from %v", r.req.RemoteAddr)
	}
	return nil
}

func (r *request) validate() error {
	// The maximum body size can be up to a maximum of 64 MiB.
	// https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html
	if r.req.ContentLength > int64(64*1024*1024) {
		r.res.statusCode = http.StatusRequestEntityTooLarge
		return errors.New("content length is too large")
	}

	switch r.req.Method {
	case http.MethodPost, http.MethodPut:
		// Do nothing, those methods are allowed
	default:
		r.res.statusCode = http.StatusMethodNotAllowed
		return fmt.Errorf("method %q is not allowed", r.req.Method)
	}

	if r.req.Header.Get("content-type") != "application/json" {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("content type, %s, is not allowed", r.req.Header.Get("content-type"))
	}

	return nil
}

func (r *request) decodeData() ([][]byte, error) {
	// decode base64-encoded data and return them as a slice of byte slices
	decodedData := make([][]byte, 0)
	for _, record := range r.body.Records {
		data, err := base64.StdEncoding.DecodeString(record.EncodedData)
		if err != nil {
			r.res.statusCode = http.StatusBadRequest
			return nil, err
		}
		decodedData = append(decodedData, data)
	}
	return decodedData, nil
}

func (r *request) extractParameterTags(parameterTags []string) (map[string]string, error) {
	paramTags := make(map[string]string)
	attributesHeader := r.req.Header.Get("x-amz-firehose-common-attributes")
	if len(parameterTags) == 0 || len(attributesHeader) == 0 {
		return paramTags, nil
	}
	var parameters map[string]interface{}
	if err := json.Unmarshal([]byte(attributesHeader), &parameters); err != nil {
		r.res.statusCode = http.StatusBadRequest
		return nil, fmt.Errorf("decode json data in x-amz-firehose-common-attributes header failed: %w", err)
	}
	paramsRaw, ok := parameters["commonAttributes"]
	if !ok {
		r.res.statusCode = http.StatusBadRequest
		return nil, errors.New("commonAttributes key not found in json data in x-amz-firehose-common-attributes header")
	}
	parameters, ok = paramsRaw.(map[string]interface{})
	if !ok {
		r.res.statusCode = http.StatusBadRequest
		return nil, errors.New("parse parameters data failed")
	}
	for _, param := range parameterTags {
		if value, ok := parameters[param]; ok {
			paramTags[param] = value.(string)
		}
	}
	return paramTags, nil
}

func (r *request) processRequest(key config.Secret, tags []string) ([][]byte, map[string]string, error) {
	if err := r.authenticate(key); err != nil {
		return nil, nil, fmt.Errorf("authentication for request %q failed: %w", r.body.RequestID, err)
	}

	if err := r.validate(); err != nil {
		return nil, nil, fmt.Errorf("validation for request %q failed: %w", r.body.RequestID, err)
	}

	records, err := r.decodeData()
	if err != nil {
		return nil, nil, fmt.Errorf("decode base64 data from request %q failed: %w", r.body.RequestID, err)
	}

	paramTags, err := r.extractParameterTags(tags)
	if err != nil {
		return nil, nil, fmt.Errorf("extracting parameter tags for request %q failed: %w", r.body.RequestID, err)
	}

	return records, paramTags, nil
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
	if err != nil {
		return fmt.Errorf("writing response to request %s failed: %w", r.res.body.RequestID, err)
	}
	return nil
}
