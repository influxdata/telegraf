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

	h := newPutRecordsHandler()

	for _, m := range testMetrics {
		h.addMetric("test", m)
	}

	if len(h.rawMetrics["test"]) != 3 {
		t.Logf("Adding metrics did not end up in the correct bucket.")
		t.Fail()
	}
}

func TestAddSlugs(t *testing.T) {
	tests := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
		[]byte("test3"),
	}

	h := newPutRecordsHandler()
	partkey := "testPartion"
	h.addSlugs(partkey, tests...)
	if len(h.slugs[partkey]) != 3 {
		t.Logf("Added 3 slugs but never seen them on the other side. Got: %v", h.slugs[partkey])
		t.Fail()
	}
}

func TestKinesisPackagedMetrics(t *testing.T) {
	tests := []struct {
		name          string
		shards        int64
		nMetrics      int
		staticKey     string
		expectedSlugs int
		snappy        bool
		gzip          bool
	}{
		{
			name:          "micro Random expect 2 slugs",
			shards:        4,
			nMetrics:      2,
			expectedSlugs: 2,
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
			name:          "vary large random expect 2 slugs",
			shards:        2,
			nMetrics:      51200,
			expectedSlugs: 1,
			snappy:        true,
			staticKey:     "static_key",
		},
		{
			name:          "vary large random expect 2 slugs",
			shards:        2,
			nMetrics:      51200,
			expectedSlugs: 1,
			gzip:          true,
			staticKey:     "static_key",
		},
	}

	for _, test := range tests {
		h := newPutRecordsHandler()
		h.setSerializer(influx.NewSerializer())

		pk := randomPartitionKey
		if test.staticKey != "" {
			pk = test.staticKey
		}

		for _, m := range fakeMetrics(t, test.nMetrics) {
			h.addMetric(pk, m)
		}

		if err := h.packageMetrics(test.shards); err != nil {
			t.Logf("%s: Failed to package metrics. Error: %s", test.name, err)
			t.Fail()
		}

		if len(h.slugs) != test.expectedSlugs {
			t.Logf("%s: Expected slug count is wrong.\nWant: %d\nGot: %d", test.name, test.expectedSlugs, len(h.slugs))
			t.Fail()
		}

		if test.snappy {
			// Snappy doesn't error, just testing for panic :(
			h.snappyCompressSlugs()
		}

		if test.gzip {
			if err := h.gzipCompressSlugs(); err != nil {
				t.Logf("%s: Error when gzip compressing slug. Error: %s", test.name, err)
				t.FailNow()
			}
		}

		// We need to make sure that we don't get panics here.
		h.convertToKinesisPutRequests()

	}
}
