package burrow

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

type transportMock struct {
	responses map[string]string
}

func newTransportMock() http.RoundTripper {
	responses := make(map[string]string)
	responses["/v2/kafka"] = clusterListResponse
	responses["/v2/kafka/clustername1/topic"] = topicListResponse
	responses["/v2/kafka/clustername1/consumer"] = consumerListResponse
	responses["/v2/kafka/clustername1/topic/topicB"] = topicDetailResponse
	responses["/v2/kafka/clustername1/consumer/group1/lag"] = consumerStatusResponse
	responses["/v2/kafka/clustername1/consumer/group2/lag"] = consumerStatusResponse

	return &transportMock{responses: responses}
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: http.StatusOK,
	}

	body, ok := t.responses[r.URL.RequestURI()]
	if !ok {
		body = ""
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(strings.NewReader(body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {
}

func assertContainsTaggedFieldsWithTS(
	t *testing.T,
	a *testutil.Accumulator,
	measurement string,
	ts time.Time,
	fields map[string]interface{},
	tags map[string]string,
) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}
		if ts != p.Time {
			continue
		}

		if p.Measurement == measurement {
			assert.Equal(t, fields, p.Fields)
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s with tags %v", measurement, tags)
	assert.Fail(t, msg)
}

func TestGather(t *testing.T) {
	b := &Burrow{
		Urls:     []string{"http://localhost"},
		Clusters: []string{"clustername1"},
		Topics:   []string{"topicB"},
	}
	b.client = &http.Client{Transport: newTransportMock()}

	var acc testutil.Accumulator
	if err := b.Gather(&acc); err != nil {
		t.Fatal(err)
	}

	for _, expected := range topicDetailExpected {
		acc.AssertContainsTaggedFields(t, "burrow_topic", expected.Fields, expected.Tags)
	}
	for _, expected := range consumerStatusExpected {
		assertContainsTaggedFieldsWithTS(t, &acc, "burrow_consumer", expected.Time, expected.Fields, expected.Tags)
	}
}
