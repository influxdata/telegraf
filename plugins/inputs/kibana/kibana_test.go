package kibana

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func defaultTags6_3() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "6.3.2",
		"status":  "green",
	}
}

func defaultTags6_5() map[string]string {
	return map[string]string{
		"name":    "my-kibana",
		"source":  "example.com:5601",
		"version": "6.5.4",
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

func checkKibanaStatusResult(version string, t *testing.T, acc *testutil.Accumulator) {
	if version == "6.3.2" {
		tags := defaultTags6_3()
		acc.AssertContainsTaggedFields(t, "kibana", kibanaStatusExpected6_3, tags)
	} else {
		tags := defaultTags6_5()
		acc.AssertContainsTaggedFields(t, "kibana", kibanaStatusExpected6_5, tags)
	}
}

func TestGather(t *testing.T) {
	ks := newKibanahWithClient()
	ks.Servers = []string{"http://example.com:5601"}
	// Unit test for Kibana version < 6.4
	ks.client.Transport = newTransportMock(http.StatusOK, kibanaStatusResponse6_3)
	var acc1 testutil.Accumulator
	if err := acc1.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaultTags6_3()["version"], t, &acc1)

	//Unit test for Kibana version >= 6.4
	ks.client.Transport = newTransportMock(http.StatusOK, kibanaStatusResponse6_5)
	var acc2 testutil.Accumulator
	if err := acc2.GatherError(ks.Gather); err != nil {
		t.Fatal(err)
	}
	checkKibanaStatusResult(defaultTags6_5()["version"], t, &acc2)
}

func newKibanahWithClient() *Kibana {
	ks := NewKibana()
	ks.client = &http.Client{}
	return ks
}
