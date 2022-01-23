package librato

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var (
	fakeURL = "http://test.librato.com"
)

func newTestLibrato(testURL string) *Librato {
	l := NewLibrato(testURL)
	l.Log = testutil.Logger{}
	return l
}

func TestUriOverride(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
	defer ts.Close()

	l := newTestLibrato(ts.URL)
	l.APIUser = "telegraf@influxdb.com"
	l.APIToken = "123456"
	err := l.Connect()
	require.NoError(t, err)
	err = l.Write([]telegraf.Metric{newHostMetric(int32(0), "name", "host")})
	require.NoError(t, err)
}

func TestBadStatusCode(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
	defer ts.Close()

	l := newTestLibrato(ts.URL)
	l.APIUser = "telegraf@influxdb.com"
	l.APIToken = "123456"
	err := l.Connect()
	require.NoError(t, err)
	err = l.Write([]telegraf.Metric{newHostMetric(int32(0), "name", "host")})
	if err == nil {
		t.Errorf("error expected but none returned")
	} else {
		require.EqualError(
			t,
			fmt.Errorf("received bad status code, 503\n "), err.Error())
	}
}

func TestBuildGauge(t *testing.T) {
	mtime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()
	var gaugeTests = []struct {
		ptIn     telegraf.Metric
		outGauge *Gauge
		err      error
	}{
		{
			newHostMetric(0.0, "test1", "host1"),
			&Gauge{
				Name:        "test1",
				MeasureTime: mtime,
				Value:       0.0,
				Source:      "host1",
			},
			nil,
		},
		{
			newHostMetric(1.0, "test2", "host2"),
			&Gauge{
				Name:        "test2",
				MeasureTime: mtime,
				Value:       1.0,
				Source:      "host2",
			},
			nil,
		},
		{
			newHostMetric(10, "test3", "host3"),
			&Gauge{
				Name:        "test3",
				MeasureTime: mtime,
				Value:       10.0,
				Source:      "host3",
			},
			nil,
		},
		{
			newHostMetric(int32(112345), "test4", "host4"),
			&Gauge{
				Name:        "test4",
				MeasureTime: mtime,
				Value:       112345.0,
				Source:      "host4",
			},
			nil,
		},
		{
			newHostMetric(int64(112345), "test5", "host5"),
			&Gauge{
				Name:        "test5",
				MeasureTime: mtime,
				Value:       112345.0,
				Source:      "host5",
			},
			nil,
		},
		{
			newHostMetric(float32(11234.5), "test6", "host6"),
			&Gauge{
				Name:        "test6",
				MeasureTime: mtime,
				Value:       11234.5,
				Source:      "host6",
			},
			nil,
		},
		{
			newHostMetric("11234.5", "test7", "host7"),
			nil,
			nil,
		},
	}

	l := newTestLibrato(fakeURL)
	for _, gt := range gaugeTests {
		gauges, err := l.buildGauges(gt.ptIn)
		if err != nil && gt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", gt.ptIn.Name(), err)
		}
		if gt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned",
				gt.ptIn.Name(), gt.err.Error())
		}
		if len(gauges) != 0 && gt.outGauge == nil {
			t.Errorf("%s: unexpected gauge, %+v\n", gt.ptIn.Name(), gt.outGauge)
		}
		if len(gauges) == 0 {
			continue
		}
		if gt.err == nil && !reflect.DeepEqual(gauges[0], gt.outGauge) {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n",
				gt.ptIn.Name(), gt.outGauge, gauges[0])
		}
	}
}

func newHostMetric(value interface{}, name, host string) telegraf.Metric {
	m := metric.New(
		name,
		map[string]string{"host": host},
		map[string]interface{}{"value": value},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	return m
}

func TestBuildGaugeWithSource(t *testing.T) {
	mtime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	pt1 := metric.New(
		"test1",
		map[string]string{"hostname": "192.168.0.1", "tag1": "value1"},
		map[string]interface{}{"value": 0.0},
		mtime,
	)
	pt2 := metric.New(
		"test2",
		map[string]string{"hostnam": "192.168.0.1", "tag1": "value1"},
		map[string]interface{}{"value": 1.0},
		mtime,
	)
	pt3 := metric.New(
		"test3",
		map[string]string{
			"hostname": "192.168.0.1",
			"tag2":     "value2",
			"tag1":     "value1"},
		map[string]interface{}{"value": 1.0},
		mtime,
	)
	pt4 := metric.New(
		"test4",
		map[string]string{
			"hostname": "192.168.0.1",
			"tag2":     "value2",
			"tag1":     "value1"},
		map[string]interface{}{"value": 1.0},
		mtime,
	)
	var gaugeTests = []struct {
		ptIn     telegraf.Metric
		template string
		outGauge *Gauge
		err      error
	}{

		{
			pt1,
			"hostname",
			&Gauge{
				Name:        "test1",
				MeasureTime: mtime.Unix(),
				Value:       0.0,
				Source:      "192_168_0_1",
			},
			nil,
		},
		{
			pt2,
			"hostname",
			&Gauge{
				Name:        "test2",
				MeasureTime: mtime.Unix(),
				Value:       1.0,
			},
			fmt.Errorf("undeterminable Source type from Field, hostname"),
		},
		{
			pt3,
			"tags",
			&Gauge{
				Name:        "test3",
				MeasureTime: mtime.Unix(),
				Value:       1.0,
				Source:      "192_168_0_1.value1.value2",
			},
			nil,
		},
		{
			pt4,
			"hostname.tag2",
			&Gauge{
				Name:        "test4",
				MeasureTime: mtime.Unix(),
				Value:       1.0,
				Source:      "192_168_0_1.value2",
			},
			nil,
		},
	}

	l := newTestLibrato(fakeURL)
	for _, gt := range gaugeTests {
		l.Template = gt.template
		gauges, err := l.buildGauges(gt.ptIn)
		if err != nil && gt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", gt.ptIn.Name(), err)
		}
		if gt.err != nil && err == nil {
			t.Errorf(
				"%s: expected an error (%s) but none returned",
				gt.ptIn.Name(),
				gt.err.Error())
		}
		if len(gauges) == 0 {
			continue
		}
		if gt.err == nil && !reflect.DeepEqual(gauges[0], gt.outGauge) {
			t.Errorf(
				"%s: \nexpected %+v\ngot %+v\n",
				gt.ptIn.Name(),
				gt.outGauge, gauges[0])
		}
	}
}
