package kinesis

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func fakeMetrics(t *testing.T, howMany int) []telegraf.Metric {
	metricList := make([]telegraf.Metric, howMany)

	for i := 0; i < howMany; i++ {
		m, err := metric.New(
			"fake_metric",
			map[string]string{
				"index":  fmt.Sprintf("%d", i),
				"static": "static_tag",
			},
			map[string]interface{}{
				"index_measurement": i,
				"nano_seconds":      time.Now().UnixNano(),
			},
			time.Now(),
		)
		if err != nil {
			t.Logf("Failed to make test metrics")
			t.FailNow()
		}
		metricList[i] = m
	}

	return metricList
}

func TestAddMetric(t *testing.T) {
	testMetrics := fakeMetrics(t, 3)

	h := newPutRecordsHandler(influx.NewSerializer())

	for _, m := range testMetrics {
		h.addRawMetric("test", m)
	}

	if len(h.rawMetrics["test"]) != 3 {
		t.Logf("Adding metrics did not end up in the correct bucket.")
		t.Fail()
	}
}

func TestAddPayload(t *testing.T) {
	tests := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
		[]byte("test3"),
	}

	h := newPutRecordsHandler(influx.NewSerializer())
	partkey := "testPartion"
	h.addPayload(partkey, tests...)
	if len(h.payloads[partkey]) != 3 {
		t.Logf("Added 3 slugs but never seen them on the other side. Got: %v", h.payloads[partkey])
		t.Fail()
	}
}

func TestKinesisPackagedMetrics(t *testing.T) {
	// A slug is a records data set that has a maximum size set by maxRecordSizeBytes
	// If we have random keys then we expect there to be many keys and record sets
	// this allows for spreading the load around shards.
	// The test will look for the amount of records and that each generated record is
	// or is lower than the maximum record size.
	tests := []struct {
		name          string
		shards        int64
		nMetrics      int
		staticKey     string
		expectedSlugs int
		encoding      string
	}{
		{
			name:          "micro Random expect 1 slugs",
			shards:        4,
			nMetrics:      2,
			expectedSlugs: 1,
		},
		{
			name:          "large Random expect 4 slugs",
			shards:        4,
			nMetrics:      4041,
			expectedSlugs: 4,
		},
		{
			name:          "vary large Random expect 4 slugs",
			shards:        4,
			nMetrics:      8081,
			expectedSlugs: 4,
		},
		{
			name:          "vary large static expect 1 slugs",
			shards:        4,
			nMetrics:      8081,
			expectedSlugs: 1,
			staticKey:     "static_key",
		},
		{
			name:          "vary large static expect 1 slugs",
			shards:        4,
			nMetrics:      8081 * 2,
			expectedSlugs: 1,
			staticKey:     "static_key",
		},
		{
			name:          "vary large random expect 6 slugs with snappy",
			shards:        2,
			nMetrics:      51200,
			expectedSlugs: 6,
			encoding:      "snappy",
		},
		{
			name:          "vary large random expect 6 slugs with gzip",
			shards:        2,
			nMetrics:      51200,
			expectedSlugs: 6,
			encoding:      "gzip",
		},
	}

	for _, test := range tests {
		h := newPutRecordsHandler(influx.NewSerializer())

		pk := randomPartitionKey
		if test.staticKey != "" {
			pk = test.staticKey
		}

		for _, m := range fakeMetrics(t, test.nMetrics) {
			h.addRawMetric(pk, m)
		}

		if err := h.packageMetrics(test.shards); err != nil {
			t.Logf("%s: Failed to package metrics. Error: %s", test.name, err)
			t.Fail()
		}

		if len(h.payloads) != test.expectedSlugs {
			t.Logf("%s: Expected slug count is wrong.\nWant: %d\nGot: %d", test.name, test.expectedSlugs, len(h.payloads))
			t.Fail()
		}

		for key, slug := range h.payloads {
			for index, recordSet := range slug {
				if len(recordSet) > maxRecordSizeBytes {
					t.Logf("%s: recordSet %d of slug %s is too large. Is: %d, max size is: %d", test.name, index, key, len(recordSet), maxRecordSizeBytes)
				}
			}
		}

		encoder, err := makeEncoder(test.encoding)
		if err != nil {
			t.Logf("Failed to make encoder. You have put something bad into the test")
			t.Fail()
		}
		if err := h.encodePayloadBodies(encoder); err != nil {
			t.Logf("Failed to encoder the data")
			t.Fail()
		}

		// We need to make sure that we don't get panics here.
		h.convertToKinesisPutRequests()

	}
}
