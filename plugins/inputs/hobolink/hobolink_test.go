package hobolink

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
)

func TestHOBOlink_ParseJSONResults(t *testing.T) {
	data := helperLoadBytes(t, "hobolink.json")

	c := NewTestClient(func(req *http.Request) *http.Response {

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
			Header:     make(http.Header),
		}
	})

	h := NewHOBOlink()
	h.client = c

	obs, err := h.parseJSON()
	if err != nil {
		t.Fatal(err)
	}

	if obs.Message != "OK: Found: 674 results." {
		t.Fatal("unexpected observation message", "got", obs.Message, "exp", "OK")
	}

	if len(obs.ObservationList) != 674 {
		t.Fatal("number of observations did not match", "got", len(obs.ObservationList), "exp", "674")
	}
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path to test file
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return bytes
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
