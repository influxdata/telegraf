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

	contentType := r.req.Header.Get("content-type")
	if contentType != "application/json" {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("content type %s is not allowed", contentType)
	}

	encoding := r.req.Header.Get("content-encoding")
	body, err := internal.NewStreamContentDecoder(encoding, r.req.Body)
	if err != nil {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("creating %q decoder failed: %w", encoding, err)
	}
	defer r.req.Body.Close()

	if err := json.NewDecoder(body).Decode(&r.body); err != nil {
		r.res.statusCode = http.StatusBadRequest
		return fmt.Errorf("decode body failed: %w", err)
	}

	if r.body.RequestID != r.req.Header.Get("x-amz-firehose-request-id") {
		r.res.statusCode = http.StatusBadRequest
		return errors.New("requestId in the body does not match x-amz-firehose-request-id request header")
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
