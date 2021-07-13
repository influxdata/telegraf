package addons

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// SendRequest wraps HTTP methods
func SendRequest(httpMethod string, requestURL string,
	headers map[string]string, formValues map[string]string, byteBody []byte) (int, []byte, error) {

	var request *http.Request
	var response *http.Response
	var err error

	urlValues := url.Values{}
	if formValues != nil {
		for key, value := range formValues {
			urlValues.Add(key, value)
		}
		byteBody = []byte(urlValues.Encode())
	}

	var body io.Reader
	if byteBody != nil {
		body = bytes.NewBuffer(byteBody)
	} else {
		body = nil
	}

	request, err = http.NewRequest(httpMethod, requestURL, body)

	if err != nil {
		return -1, nil, err
	}

	request.Header.Set("Connection", "close")
	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}

	response, err = httpClient.Do(request)
	if err != nil {
		return -1, nil, err
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return -1, nil, err
	}
	return response.StatusCode, responseBody, nil
}
