package kibana

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func defaultTags() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "6.3.2",
		"status":  "green",
	}
}

type transportMock struct {
	statusCode int
	body       string
}

func newTransportMock(statusCode int, body string) http.RoundTripper {
	return &transportMock{
		statusCode: statusCode,
		body:       body,
	}
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func checkKibanaStatusResult(t *testing.T, acc *testutil.Accumulator) {
	tags := defaultTags()
	acc.AssertContainsTaggedFields(t, "kibana", kibanaStatusExpected, tags)
}

func TestGather(t *testing.T) {
	ks := newKibanahWithClient()
	ks.Servers = []string{"http://example.com:5601"}
	ks.client.Transport = newTransportMock(http.StatusOK, kibanaStatusResponse)

	var acc testutil.Accumulator
	if err := acc.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}

	checkKibanaStatusResult(t, &acc)
}

func newKibanahWithClient() *Kibana {
	ks := NewKibana()
	ks.client = &http.Client{}
	return ks
}
