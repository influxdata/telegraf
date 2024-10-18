package firehose

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf/config"
)

type firehoseRecord struct {
	EncodedData string `json:"data"`
}

type firehoseRequestBody struct {
	RequestID string           `json:"requestId"`
	Timestamp int64            `json:"timestamp"`
	Records   []firehoseRecord `json:"records"`
}

// https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html#requestformat
type firehoseRequest struct {
	req                *http.Request
	body               firehoseRequestBody
	responseStatusCode int
}

func (r *firehoseRequest) authenticate(expectedAccessKey config.Secret) error {
	if expectedAccessKey.Empty() {
		return nil
	}
	accessKey, err := expectedAccessKey.Get()
	if err != nil {
		return fmt.Errorf("getting accesskey failed: %w", err)
	}
	reqAccessKey := r.req.Header.Get("x-amz-firehose-access-key")
	if reqAccessKey != accessKey.String() {
		r.responseStatusCode = http.StatusUnauthorized
		return fmt.Errorf("unauthorized request %s", r.req.Header.Get("x-amz-firehose-request-id"))
	}
	return nil
}

func (r *firehoseRequest) validate() error {
	requestID := r.req.Header.Get("x-amz-firehose-request-id")
	if requestID == "" {
		r.responseStatusCode = http.StatusBadRequest
		return errors.New("x-amz-firehose-request-id header is not set")
	}

	// Check if content length is not over 64 MB.
	if r.req.ContentLength > int64(64*1024*1024) {
		r.responseStatusCode = http.StatusRequestEntityTooLarge
		return fmt.Errorf("content length too large in request %s", requestID)
	}

	// Check if the requested HTTP method is allowed.
	isAcceptedMethod := false
	for _, method := range allowedMethods {
		if r.req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		r.responseStatusCode = http.StatusMethodNotAllowed
		return fmt.Errorf("forbidden method, %s, in request %s", r.req.Method, requestID)
	}

	contentType := r.req.Header.Get("content-type")
	if contentType != "application/json" {
		r.responseStatusCode = http.StatusBadRequest
		return fmt.Errorf("unaccepted content type, %s, in request %s", contentType, requestID)
	}

	contentEncoding := r.req.Header.Get("content-encoding")
	if contentEncoding != "" && contentEncoding != "gzip" {
		r.responseStatusCode = http.StatusBadRequest
		return fmt.Errorf("unaccepted content encoding, %s, in request %s", contentEncoding, requestID)
	}

	err := r.extractBody()
	if err != nil {
		return err
	}

	if requestID != r.body.RequestID {
		r.responseStatusCode = http.StatusBadRequest
		return errors.New("requestId in the body does not match the value of the request header, x-amz-firehose-request-id")
	}

	return nil
}

func (r *firehoseRequest) extractBody() error {
	encoding := r.req.Header.Get("content-encoding")
	switch encoding {
	case "gzip":
		g, err := gzip.NewReader(r.req.Body)
		if err != nil {
			r.responseStatusCode = http.StatusBadRequest
			return fmt.Errorf("unable to decode body - %s", err.Error())
		}
		defer g.Close()
		err = json.NewDecoder(g).Decode(&r.body)
		if err != nil {
			r.responseStatusCode = http.StatusBadRequest
			return err
		}
	default:
		defer r.req.Body.Close()
		err := json.NewDecoder(r.req.Body).Decode(&r.body)
		if err != nil {
			r.responseStatusCode = http.StatusBadRequest
			return err
		}
	}
	return nil
}

func (r *firehoseRequest) decodeData() ([][]byte, bool) {
	// decode base64-encoded data and return them as a slice of byte slices
	decodedData := make([][]byte, 0)
	for _, record := range r.body.Records {
		data, err := base64.StdEncoding.DecodeString(record.EncodedData)
		if err != nil {
			return nil, false
		}
		decodedData = append(decodedData, data)
	}
	return decodedData, true
}

func (r *firehoseRequest) sendResponse(res http.ResponseWriter) error {
	responseBody := struct {
		RequestID    string `json:"requestId"`
		Timestamp    int64  `json:"timestamp"`
		ErrorMessage string `json:"errorMessage,omitempty"`
	}{
		RequestID:    r.req.Header.Get("x-amz-firehose-request-id"),
		Timestamp:    time.Now().Unix(),
		ErrorMessage: statusCodeToMessage[r.responseStatusCode],
	}
	response, err := json.Marshal(responseBody)
	if err != nil {
		return err
	}
	res.Header().Set("content-type", "application/json")
	res.WriteHeader(r.responseStatusCode)
	_, err = res.Write(response)
	return err
}
