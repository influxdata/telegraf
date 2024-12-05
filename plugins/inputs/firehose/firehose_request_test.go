package firehose

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/require"
)

func newTestHTTPRequest(method, body string, headers map[string]string) *http.Request {
	req, err := http.NewRequest(method, "http://localhost:8080/telegraf", bytes.NewReader([]byte(body)))
	if err != nil {
		panic(fmt.Sprintf("failed to create request: %v", err))
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

func TestNewRequest(t *testing.T) {
	body := `{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`
	headers := map[string]string{"x-amz-firehose-request-id": "test-id"}
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost, body, headers))
	require.NoError(t, err)
	require.Equal(t, "test-id", r.body.RequestID)
	require.Equal(t, "dGVzdA==", r.body.Records[0].EncodedData)
}

func TestNewRequestErrors(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		headers map[string]string
		err     string
	}{
		{
			name:    "Missing Request ID Header",
			body:    `{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
			headers: map[string]string{"x-amz-firehose-request-id": ""},
			err:     "x-amz-firehose-request-id header is not set",
		},
		{
			name:    "Body Not JSON",
			body:    "not a json",
			headers: map[string]string{"x-amz-firehose-request-id": "test-id"},
			err:     `decode body for request "test-id" failed`,
		},
		{
			name:    "ID Mismatch",
			body:    `{"requestId":"some-other-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
			headers: map[string]string{"x-amz-firehose-request-id": "test-id"},
			err:     "mismatch between requestID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost, tc.body, tc.headers))
			require.Error(t, err)
			require.ErrorContains(t, err, tc.err)
			require.NotNil(t, r)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		headers map[string]string
		key     config.Secret
	}{
		{
			name:    "No Authentication Required",
			headers: map[string]string{"x-amz-firehose-request-id": "test-id"},
			key:     config.NewSecret([]byte("")),
		},
		{
			name:    "Authentication Required",
			headers: map[string]string{"x-amz-firehose-request-id": "test-id", "x-amz-firehose-access-key": "test-key"},
			key:     config.NewSecret([]byte("test-key")),
		},
	}

	body := `{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost, body, tc.headers))
			require.NoError(t, err)
			err = r.authenticate(tc.key)
			require.NoError(t, err)
		})
	}
}

func TestAuthenticateInvalidKey(t *testing.T) {
	body := `{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`
	headers := map[string]string{"x-amz-firehose-request-id": "test-id", "x-amz-firehose-access-key": "some-other-key"}
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost, body, headers))
	require.NoError(t, err)
	err = r.authenticate(config.NewSecret([]byte("test-key")))
	require.Error(t, err)
	require.ErrorContains(t, err, "unauthorized request")
}

func TestValidate(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id", "content-type": "application/json"},
	))
	require.NoError(t, err)
	err = r.validate()
	require.NoError(t, err)
}

func TestValidateErrors(t *testing.T) {
	testCases := []struct {
		name    string
		method  string
		headers map[string]string
		err     string
	}{
		{
			name:    "Method Not Allowed",
			method:  http.MethodGet,
			headers: map[string]string{"x-amz-firehose-request-id": "test-id", "content-type": "application/json"},
			err:     `method "GET" is not allowed`,
		},
		{
			name:    "Content Not Allowed",
			method:  http.MethodPost,
			headers: map[string]string{"x-amz-firehose-request-id": "test-id", "content-type": "text/html"},
			err:     "content type, text/html, is not allowed",
		},
	}
	body := `{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := newFirehoseRequest(newTestHTTPRequest(tc.method, body, tc.headers))
			require.NoError(t, err)
			err = r.validate()
			require.Error(t, err)
			require.ErrorContains(t, err, tc.err)
		})
	}
}

func TestDecodeData(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id"},
	))
	require.NoError(t, err)
	records, err := r.decodeData()
	require.NoError(t, err)
	require.Equal(t, records[0], []byte("test"))
}

func TestDecodeDataError(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"not a base64 encoded text"}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id"},
	))
	require.NoError(t, err)
	records, err := r.decodeData()
	require.Error(t, err)
	require.Nil(t, records)
}

func TestExtractParameterTags(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id", "x-amz-firehose-common-attributes": `{"commonAttributes":{"env":"test","foo":"bar"}}`},
	))
	require.NoError(t, err)
	paramTags, err := r.extractParameterTags([]string{"env"})
	require.NoError(t, err)
	require.Len(t, paramTags, 1)
	env, ok := paramTags["env"]
	require.True(t, ok)
	require.Equal(t, "test", env)
}

func TestExtractParametersTagsNoHeader(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id"},
	))
	require.NoError(t, err)
	paramTags, err := r.extractParameterTags([]string{"env"})
	require.NoError(t, err)
	require.Empty(t, paramTags)
}

func TestExtractParameterTagsErrors(t *testing.T) {
	testCases := []struct {
		name    string
		headers map[string]string
		err     string
	}{
		{
			name:    "Header Not Json",
			headers: map[string]string{"x-amz-firehose-request-id": "test-id", "x-amz-firehose-common-attributes": "not a json"},
			err:     "decode json data in x-amz-firehose-common-attributes header failed",
		},
		{
			name:    "Key Not Found",
			headers: map[string]string{"x-amz-firehose-request-id": "test-id", "x-amz-firehose-common-attributes": `{"key":"value"}`},
			err:     "commonAttributes key not found",
		},
		{
			name:    "Parse Error",
			headers: map[string]string{"x-amz-firehose-request-id": "test-id", "x-amz-firehose-common-attributes": `{"commonAttributes":"value"}`},
			err:     "parse parameters data failed",
		},
	}
	body := `{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost, body, tc.headers))
			require.NoError(t, err)
			paramTags, err := r.extractParameterTags([]string{"env"})
			require.Error(t, err)
			require.ErrorContains(t, err, tc.err)
			require.Nil(t, paramTags)
		})
	}
}

func TestSendResponseDefault(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id"},
	))
	require.NoError(t, err)
	res := httptest.NewRecorder()
	err = r.sendResponse(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, "test-id", r.res.body.RequestID)
	require.Equal(t, "Internal Server Error", r.res.body.ErrorMessage)
}

func TestSendResponseOk(t *testing.T) {
	r, err := newFirehoseRequest(newTestHTTPRequest(http.MethodPost,
		`{"requestId":"test-id","timestamp":1578090901599,"records":[{"data":"dGVzdA=="}]}`,
		map[string]string{"x-amz-firehose-request-id": "test-id"},
	))
	require.NoError(t, err)
	r.res.statusCode = http.StatusOK
	res := httptest.NewRecorder()
	err = r.sendResponse(res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "test-id", r.res.body.RequestID)
	require.Empty(t, r.res.body.ErrorMessage)
}
