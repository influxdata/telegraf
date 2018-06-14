package kube_state

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockHandler struct {
	unauthorized bool
	noContent    bool
	// responseMap is the path to repsonse interface
	// we will ouput the serialized response in json when serving http
	// example '/computer/api/json': *gojenkins.
	responseMap map[string]interface{}
}

func (h mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.unauthorized {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if h.noContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	o, ok := h.responseMap[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(o)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func TestCreateGetRequest(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		token    string
		want     *http.Request
		hasError bool
	}{
		{
			name:     "create request bad url",
			url:      "// /foo?bad url",
			hasError: true,
		},
		{
			name: "request without token",
			url:  "http://goodurl",
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme: "http",
					Host:   "goodurl",
				},
				Header: map[string][]string{
					"Accept": []string{"application/json"},
				},
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Host:       "goodurl",
			},
			hasError: false,
		},
		{
			name:  "request with token",
			url:   "http://goodurl",
			token: "tok",
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme: "http",
					Host:   "goodurl",
				},
				Header: map[string][]string{
					"Accept": []string{"application/json"},
					"Authorization": []string{
						"Bearer tok",
					},
				},
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Host:       "goodurl",
			},
			hasError: false,
		},
	}
	for _, v := range tests {
		req, err := createGetRequest(v.url, v.token)
		if err == nil && v.hasError {
			t.Fatalf("%s failed, should have error", v.name)
		} else if err != nil && !v.hasError {
			t.Fatalf("%s failed, err: %v", v.name, err)
		}
		assert.Equal(t, v.want, req, v.name+" req")
	}
}

func TestGets(t *testing.T) {
	cli := &client{
		httpClient: &http.Client{Transport: &http.Transport{}},
		semaphore:  make(chan struct{}, 1),
	}
	mock := &mockHandler{
		responseMap: map[string]interface{}{
			"/": &metav1.APIResourceList{
				GroupVersion: "v1",
			},
		},
	}
	badMock :=
		&mockHandler{
			responseMap: map[string]interface{}{
				"/": &metav1.APIResourceList{},
			},
		}
	tests := []struct {
		name     string
		handler  *mockHandler
		reqFunc  string
		wantObj  interface{}
		hasError bool
	}{
		{
			name:    "api source",
			handler: mock,
			wantObj: &metav1.APIResourceList{
				GroupVersion: "v1",
			},
		},
		{
			name:     "bad api source",
			handler:  badMock,
			hasError: true,
		},
	}
	for _, v := range tests {
		ts := httptest.NewServer(v.handler)
		defer ts.Close()

		cli.baseURL = ts.URL
		result, err := cli.getAPIResourceList(context.Background())
		if err == nil && v.hasError {
			t.Fatalf("%s failed, should have error", v.name)
		} else if err != nil && !v.hasError {
			t.Fatalf("%s failed, err: %v", v.name, err)
		}

		if v.wantObj != nil {
			assert.Equal(t, v.wantObj, result, v.name+" json decode")
		}
	}
}

func TestDoGet(t *testing.T) {
	type mockObj struct {
		Fld1 string
		Fld2 int
	}
	type mockObj2 struct {
		Fld1 int
		Fld2 string
	}

	tests := []struct {
		name          string
		cli           *client
		handler       *mockHandler
		wantBearToken string
		reqObj        *mockObj
		wantObj       *mockObj
		hasError      bool
		expectedErr   error
	}{
		{
			name: "bad request",
			cli: &client{
				baseURL: "// /foo?bad url",
			},
			hasError: true,
		},
		{
			name: "unauthorized",
			cli: &client{
				httpClient:  &http.Client{Transport: &http.Transport{}},
				semaphore:   make(chan struct{}, 1),
				bearerToken: "tok",
			},
			handler: &mockHandler{
				unauthorized: true,
			},
			expectedErr: APIError{
				URL:        "/",
				StatusCode: http.StatusUnauthorized,
				Title:      fmt.Sprintf("%d %s", http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)),
			},
			hasError: true,
		},
		{
			name: "no content",
			cli: &client{
				httpClient: &http.Client{Transport: &http.Transport{}},
				semaphore:  make(chan struct{}, 1),
			},
			handler: &mockHandler{
				noContent: true,
			},
		},
		{
			name: "good json decoder",
			cli: &client{
				httpClient: &http.Client{Transport: &http.Transport{}},
				semaphore:  make(chan struct{}, 1),
			},
			reqObj: new(mockObj),
			wantObj: &mockObj{
				Fld1: "str1",
				Fld2: 2,
			},
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/": &mockObj{
						Fld1: "str1",
						Fld2: 2,
					},
				},
			},
			hasError: false,
		},
		{
			name: "bad json decoder",
			cli: &client{
				httpClient: &http.Client{Transport: &http.Transport{}},
				semaphore:  make(chan struct{}, 1),
			},
			reqObj: new(mockObj),
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/": &mockObj2{
						Fld1: 1,
						Fld2: "str1",
					},
				},
			},
			hasError: true,
		},
	}
	for _, v := range tests {
		ts := httptest.NewServer(v.handler)
		defer ts.Close()
		if v.cli.baseURL == "" {
			v.cli.baseURL = ts.URL
		}
		err := v.cli.doGet(context.Background(), "/", v.reqObj)
		if err == nil && v.hasError {
			t.Fatalf("%s failed, should have error", v.name)
		} else if err != nil && !v.hasError {
			t.Fatalf("%s failed, err: %v", v.name, err)
		}

		if v.hasError {
			if apiErr, ok := v.expectedErr.(APIError); ok {
				assert.Equal(t, apiErr, err, v.name)
			}
		}
		assert.Equal(t, v.wantBearToken, v.cli.bearerToken, "%s failed breaer token")
		if v.wantObj != nil {
			assert.Equal(t, v.wantObj, v.reqObj, v.name+" json decode")
		}
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name   string
		err    APIError
		result string
	}{
		{
			name: "with description",
			err: APIError{
				URL:         "a url",
				StatusCode:  http.StatusBadGateway,
				Title:       "title1",
				Description: "desp1",
			},
			result: "[a url] title1: desp1",
		},
		{
			name: "without description",
			err: APIError{
				URL:        "a url",
				StatusCode: http.StatusBadGateway,
				Title:      "title1",
			},
			result: "[a url] title1",
		},
	}
	for _, v := range tests {
		assert.Equal(t, v.result, v.err.Error(), v.name)
	}
}
