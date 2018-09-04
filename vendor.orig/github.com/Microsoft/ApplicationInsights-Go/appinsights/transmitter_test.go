package appinsights

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
)

type testServer struct {
	server *httptest.Server
	notify chan *testRequest

	responseData    []byte
	responseCode    int
	responseHeaders map[string]string
}

type testRequest struct {
	request *http.Request
	body    []byte
}

func (server *testServer) Close() {
	server.server.Close()
	close(server.notify)
}

func (server *testServer) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)

	hdr := writer.Header()
	for k, v := range server.responseHeaders {
		hdr[k] = []string{v}
	}

	writer.WriteHeader(server.responseCode)
	writer.Write(server.responseData)

	server.notify <- &testRequest{
		request: req,
		body:    body,
	}
}

func (server *testServer) waitForRequest(t *testing.T) *testRequest {
	select {
	case req := <-server.notify:
		return req
	case <-time.After(time.Second):
		t.Fatal("Server did not receive request within a second")
		return nil /* not reached */
	}
}

type nullTransmitter struct{}

func (transmitter *nullTransmitter) Transmit(payload []byte, items telemetryBufferItems) (*transmissionResult, error) {
	return &transmissionResult{statusCode: successResponse}, nil
}

func newTestClientServer() (transmitter, *testServer) {
	server := &testServer{}
	server.server = httptest.NewServer(server)
	server.notify = make(chan *testRequest, 1)
	server.responseCode = 200
	server.responseData = make([]byte, 0)
	server.responseHeaders = make(map[string]string)

	client := newTransmitter(fmt.Sprintf("http://%s/v2/track", server.server.Listener.Addr().String()))

	return client, server
}

func TestBasicTransmit(t *testing.T) {
	client, server := newTestClientServer()
	defer server.Close()

	server.responseData = []byte(`{"itemsReceived":3, "itemsAccepted":5, "errors":[]}`)
	server.responseHeaders["Content-type"] = "application/json"
	result, err := client.Transmit([]byte("foobar"), make(telemetryBufferItems, 0))
	req := server.waitForRequest(t)

	if err != nil {
		t.Errorf("err: %s", err.Error())
	}

	if req.request.Method != "POST" {
		t.Error("request.Method")
	}

	cencoding := req.request.Header[http.CanonicalHeaderKey("Content-Encoding")]
	if len(cencoding) != 1 || cencoding[0] != "gzip" {
		t.Errorf("Content-encoding: %q", cencoding)
	}

	// Check for gzip magic number
	if len(req.body) < 2 || req.body[0] != 0x1f || req.body[1] != 0x8b {
		t.Fatal("Missing gzip magic number")
	}

	// Decompress payload
	reader, err := gzip.NewReader(bytes.NewReader(req.body))
	if err != nil {
		t.Fatalf("Couldn't create gzip reader: %s", err.Error())
	}

	body, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatalf("Couldn't read compressed data: %s", err.Error())
	}

	if string(body) != "foobar" {
		t.Error("body")
	}

	ctype := req.request.Header[http.CanonicalHeaderKey("Content-Type")]
	if len(ctype) != 1 || ctype[0] != "application/x-json-stream" {
		t.Errorf("Content-type: %q", ctype)
	}

	if result.statusCode != 200 {
		t.Error("statusCode")
	}

	if result.retryAfter != nil {
		t.Error("retryAfter")
	}

	if result.response == nil {
		t.Fatal("response")
	}

	if result.response.ItemsReceived != 3 {
		t.Error("ItemsReceived")
	}

	if result.response.ItemsAccepted != 5 {
		t.Error("ItemsAccepted")
	}

	if len(result.response.Errors) != 0 {
		t.Error("response.Errors")
	}
}

func TestFailedTransmit(t *testing.T) {
	client, server := newTestClientServer()
	defer server.Close()

	server.responseCode = errorResponse
	server.responseData = []byte(`{"itemsReceived":3, "itemsAccepted":0, "errors":[{"index": 2, "statusCode": 500, "message": "Hello"}]}`)
	server.responseHeaders["Content-type"] = "application/json"
	result, err := client.Transmit([]byte("foobar"), make(telemetryBufferItems, 0))
	server.waitForRequest(t)

	if err != nil {
		t.Errorf("err: %s", err.Error())
	}

	if result.statusCode != errorResponse {
		t.Error("statusCode")
	}

	if result.retryAfter != nil {
		t.Error("retryAfter")
	}

	if result.response == nil {
		t.Fatal("response")
	}

	if result.response.ItemsReceived != 3 {
		t.Error("ItemsReceived")
	}

	if result.response.ItemsAccepted != 0 {
		t.Error("ItemsAccepted")
	}

	if len(result.response.Errors) != 1 {
		t.Fatal("len(Errors)")
	}

	if result.response.Errors[0].Index != 2 {
		t.Error("Errors[0].index")
	}

	if result.response.Errors[0].StatusCode != errorResponse {
		t.Error("Errors[0].statusCode")
	}

	if result.response.Errors[0].Message != "Hello" {
		t.Error("Errors[0].message")
	}
}

func TestThrottledTransmit(t *testing.T) {
	client, server := newTestClientServer()
	defer server.Close()

	server.responseCode = errorResponse
	server.responseData = make([]byte, 0)
	server.responseHeaders["Content-type"] = "application/json"
	server.responseHeaders["retry-after"] = "Wed, 09 Aug 2017 23:43:57 UTC"
	result, err := client.Transmit([]byte("foobar"), make(telemetryBufferItems, 0))
	server.waitForRequest(t)

	if err != nil {
		t.Errorf("err: %s", err.Error())
	}

	if result.statusCode != errorResponse {
		t.Error("statusCode")
	}

	if result.response != nil {
		t.Fatal("response")
	}

	if result.retryAfter == nil {
		t.Fatal("retryAfter")
	}

	if (*result.retryAfter).Unix() != 1502322237 {
		t.Error("retryAfter.Unix")
	}
}

func TestTransmitDiagnostics(t *testing.T) {
	client, server := newTestClientServer()
	defer server.Close()

	var msgs []string
	notify := make(chan bool, 1)

	NewDiagnosticsMessageListener(func(message string) error {
		if message == "PING" {
			notify <- true
		} else {
			msgs = append(msgs, message)
		}

		return nil
	})

	defer resetDiagnosticsListeners()

	server.responseCode = errorResponse
	server.responseData = []byte(`{"itemsReceived":1, "itemsAccepted":0, "errors":[{"index": 0, "statusCode": 500, "message": "Hello"}]}`)
	server.responseHeaders["Content-type"] = "application/json"
	_, err := client.Transmit([]byte("foobar"), make(telemetryBufferItems, 0))
	server.waitForRequest(t)

	// Wait for diagnostics to catch up.
	diagnosticsWriter.Write("PING")
	<-notify

	if err != nil {
		t.Errorf("err: %s", err.Error())
	}

	// The last line should say "Errors:" and not include the error because the telemetry item wasn't submitted.
	if !strings.Contains(msgs[len(msgs)-1], "Errors:") {
		t.Errorf("Last line should say 'Errors:', with no errors listed.  Instead: %s", msgs[len(msgs)-1])
	}

	// Go again but include telemetry items this time.
	server.responseCode = errorResponse
	server.responseData = []byte(`{"itemsReceived":1, "itemsAccepted":0, "errors":[{"index": 0, "statusCode": 500, "message": "Hello"}]}`)
	server.responseHeaders["Content-type"] = "application/json"
	_, err = client.Transmit([]byte("foobar"), telemetryBuffer(NewTraceTelemetry("World", Warning)))
	server.waitForRequest(t)

	// Wait for diagnostics to catch up.
	diagnosticsWriter.Write("PING")
	<-notify

	if err != nil {
		t.Errorf("err: %s", err.Error())
	}

	if !strings.Contains(msgs[len(msgs)-2], "500 Hello") {
		t.Error("Telemetry error should be prefaced with result code and message")
	}

	if !strings.Contains(msgs[len(msgs)-1], "World") {
		t.Error("Raw telemetry item should be found on last line")
	}

	close(notify)
}

type resultProperties struct {
	isSuccess        bool
	isFailure        bool
	canRetry         bool
	isThrottled      bool
	isPartialSuccess bool
	retryableErrors  bool
}

func checkTransmitResult(t *testing.T, result *transmissionResult, expected *resultProperties) {
	retryAfter := "<nil>"
	if result.retryAfter != nil {
		retryAfter = (*result.retryAfter).String()
	}
	response := "<nil>"
	if result.response != nil {
		response = fmt.Sprintf("%q", *result.response)
	}
	id := fmt.Sprintf("%d, retryAfter:%s, response:%s", result.statusCode, retryAfter, response)

	if result.IsSuccess() != expected.isSuccess {
		t.Errorf("Expected IsSuccess() == %t [%s]", expected.isSuccess, id)
	}

	if result.IsFailure() != expected.isFailure {
		t.Errorf("Expected IsFailure() == %t [%s]", expected.isFailure, id)
	}

	if result.CanRetry() != expected.canRetry {
		t.Errorf("Expected CanRetry() == %t [%s]", expected.canRetry, id)
	}

	if result.IsThrottled() != expected.isThrottled {
		t.Errorf("Expected IsThrottled() == %t [%s]", expected.isThrottled, id)
	}

	if result.IsPartialSuccess() != expected.isPartialSuccess {
		t.Errorf("Expected IsPartialSuccess() == %t [%s]", expected.isPartialSuccess, id)
	}

	// retryableErrors is true if CanRetry() and any error is recoverable
	retryableErrors := false
	if result.CanRetry() && result.response != nil {
		for _, err := range result.response.Errors {
			if err.CanRetry() {
				retryableErrors = true
			}
		}
	}

	if retryableErrors != expected.retryableErrors {
		t.Errorf("Expected any(Errors.CanRetry) == %t [%s]", expected.retryableErrors, id)
	}
}

func TestTransmitResults(t *testing.T) {
	retryAfter := time.Unix(1502322237, 0)
	partialNoRetries := &backendResponse{
		ItemsAccepted: 3,
		ItemsReceived: 5,
		Errors: []*itemTransmissionResult{
			&itemTransmissionResult{Index: 2, StatusCode: 400, Message: "Bad 1"},
			&itemTransmissionResult{Index: 4, StatusCode: 400, Message: "Bad 2"},
		},
	}

	partialSomeRetries := &backendResponse{
		ItemsAccepted: 2,
		ItemsReceived: 4,
		Errors: []*itemTransmissionResult{
			&itemTransmissionResult{Index: 2, StatusCode: 400, Message: "Bad 1"},
			&itemTransmissionResult{Index: 4, StatusCode: 408, Message: "OK Later"},
		},
	}

	noneAccepted := &backendResponse{
		ItemsAccepted: 0,
		ItemsReceived: 5,
		Errors: []*itemTransmissionResult{
			&itemTransmissionResult{Index: 0, StatusCode: 500, Message: "Bad 1"},
			&itemTransmissionResult{Index: 1, StatusCode: 500, Message: "Bad 2"},
			&itemTransmissionResult{Index: 2, StatusCode: 500, Message: "Bad 3"},
			&itemTransmissionResult{Index: 3, StatusCode: 500, Message: "Bad 4"},
			&itemTransmissionResult{Index: 4, StatusCode: 500, Message: "Bad 5"},
		},
	}

	allAccepted := &backendResponse{
		ItemsAccepted: 6,
		ItemsReceived: 6,
		Errors:        make([]*itemTransmissionResult, 0),
	}

	checkTransmitResult(t, &transmissionResult{200, nil, allAccepted},
		&resultProperties{isSuccess: true})
	checkTransmitResult(t, &transmissionResult{206, nil, partialSomeRetries},
		&resultProperties{isPartialSuccess: true, canRetry: true, retryableErrors: true})
	checkTransmitResult(t, &transmissionResult{206, nil, partialNoRetries},
		&resultProperties{isPartialSuccess: true, canRetry: true})
	checkTransmitResult(t, &transmissionResult{206, nil, noneAccepted},
		&resultProperties{isPartialSuccess: true, canRetry: true, retryableErrors: true})
	checkTransmitResult(t, &transmissionResult{206, nil, allAccepted},
		&resultProperties{isSuccess: true})
	checkTransmitResult(t, &transmissionResult{400, nil, nil},
		&resultProperties{isFailure: true})
	checkTransmitResult(t, &transmissionResult{408, nil, nil},
		&resultProperties{isFailure: true, canRetry: true})
	checkTransmitResult(t, &transmissionResult{408, &retryAfter, nil},
		&resultProperties{isFailure: true, canRetry: true, isThrottled: true})
	checkTransmitResult(t, &transmissionResult{429, nil, nil},
		&resultProperties{isFailure: true, canRetry: true, isThrottled: true})
	checkTransmitResult(t, &transmissionResult{429, &retryAfter, nil},
		&resultProperties{isFailure: true, canRetry: true, isThrottled: true})
	checkTransmitResult(t, &transmissionResult{500, nil, nil},
		&resultProperties{isFailure: true, canRetry: true})
	checkTransmitResult(t, &transmissionResult{503, nil, nil},
		&resultProperties{isFailure: true, canRetry: true})
	checkTransmitResult(t, &transmissionResult{401, nil, nil},
		&resultProperties{isFailure: true})
	checkTransmitResult(t, &transmissionResult{408, nil, partialSomeRetries},
		&resultProperties{isFailure: true, canRetry: true, retryableErrors: true})
	checkTransmitResult(t, &transmissionResult{500, nil, partialSomeRetries},
		&resultProperties{isFailure: true, canRetry: true, retryableErrors: true})
}

func TestGetRetryItems(t *testing.T) {
	mockClock()
	defer resetClock()

	// Keep a pristine copy.
	originalPayload, originalItems := makePayload()

	res1 := &transmissionResult{
		statusCode: 200,
		response:   &backendResponse{ItemsReceived: 7, ItemsAccepted: 7},
	}

	payload1, items1 := res1.GetRetryItems(makePayload())
	if len(payload1) > 0 || len(items1) > 0 {
		t.Error("GetRetryItems shouldn't return anything")
	}

	res2 := &transmissionResult{statusCode: 408}

	payload2, items2 := res2.GetRetryItems(makePayload())
	if string(originalPayload) != string(payload2) || len(items2) != 7 {
		t.Error("GetRetryItems shouldn't return anything")
	}

	res3 := &transmissionResult{
		statusCode: 206,
		response: &backendResponse{
			ItemsReceived: 7,
			ItemsAccepted: 4,
			Errors: []*itemTransmissionResult{
				&itemTransmissionResult{Index: 1, StatusCode: 200, Message: "OK"},
				&itemTransmissionResult{Index: 3, StatusCode: 400, Message: "Bad"},
				&itemTransmissionResult{Index: 5, StatusCode: 408, Message: "Later"},
				&itemTransmissionResult{Index: 6, StatusCode: 500, Message: "Oops"},
			},
		},
	}

	payload3, items3 := res3.GetRetryItems(makePayload())
	expected3 := telemetryBufferItems{originalItems[5], originalItems[6]}
	if string(payload3) != string(expected3.serialize()) || len(items3) != 2 {
		t.Error("Unexpected result")
	}
}

func makePayload() ([]byte, telemetryBufferItems) {
	buffer := telemetryBuffer()
	for i := 0; i < 7; i++ {
		tr := NewTraceTelemetry(fmt.Sprintf("msg%d", i+1), contracts.SeverityLevel(i%5))
		tr.Tags.Operation().SetId(fmt.Sprintf("op%d", i))
		buffer.add(tr)
	}

	return buffer.serialize(), buffer
}
