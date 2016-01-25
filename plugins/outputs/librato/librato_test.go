package librato

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/stretchr/testify/require"
)

var (
	fakeUrl   = "http://test.librato.com"
	fakeUser  = "telegraf@influxdb.com"
	fakeToken = "123456"
)

func fakeLibrato() *Librato {
	l := NewLibrato(fakeUrl)
	l.ApiUser = fakeUser
	l.ApiToken = fakeToken
	return l
}

func TestUriOverride(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	l := NewLibrato(ts.URL)
	l.ApiUser = "telegraf@influxdb.com"
	l.ApiToken = "123456"
	err := l.Connect()
	require.NoError(t, err)
	err = l.Write(testutil.MockBatchPoints().Points())
	require.NoError(t, err)
}

func TestBadStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(`{
      "errors": {
        "system": [
          "The API is currently down for maintenance. It'll be back shortly."
        ]
      }
    }`)
	}))
	defer ts.Close()

	l := NewLibrato(ts.URL)
	l.ApiUser = "telegraf@influxdb.com"
	l.ApiToken = "123456"
	err := l.Connect()
	require.NoError(t, err)
	err = l.Write(testutil.MockBatchPoints().Points())
	if err == nil {
		t.Errorf("error expected but none returned")
	} else {
		require.EqualError(t, fmt.Errorf("received bad status code, 503\n"), err.Error())
	}
}

func TestBuildGauge(t *testing.T) {
	var gaugeTests = []struct {
		ptIn     *client.Point
		outGauge *Gauge
		err      error
	}{
		{
			testutil.TestPoint(0.0, "test1"),
			&Gauge{
				Name:        "test1",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       0.0,
			},
			nil,
		},
		{
			testutil.TestPoint(1.0, "test2"),
			&Gauge{
				Name:        "test2",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       1.0,
			},
			nil,
		},
		{
			testutil.TestPoint(10, "test3"),
			&Gauge{
				Name:        "test3",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       10.0,
			},
			nil,
		},
		{
			testutil.TestPoint(int32(112345), "test4"),
			&Gauge{
				Name:        "test4",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       112345.0,
			},
			nil,
		},
		{
			testutil.TestPoint(int64(112345), "test5"),
			&Gauge{
				Name:        "test5",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       112345.0,
			},
			nil,
		},
		{
			testutil.TestPoint(float32(11234.5), "test6"),
			&Gauge{
				Name:        "test6",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       11234.5,
			},
			nil,
		},
		{
			testutil.TestPoint("11234.5", "test7"),
			&Gauge{
				Name:        "test7",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       11234.5,
			},
			fmt.Errorf("unable to extract value from Fields, undeterminable type"),
		},
	}

	l := NewLibrato(fakeUrl)
	for _, gt := range gaugeTests {
		gauges, err := l.buildGauges(gt.ptIn)
		if err != nil && gt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", gt.ptIn.Name(), err)
		}
		if gt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned",
				gt.ptIn.Name(), gt.err.Error())
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

func TestBuildGaugeWithSource(t *testing.T) {
	pt1, _ := client.NewPoint(
		"test1",
		map[string]string{"hostname": "192.168.0.1"},
		map[string]interface{}{"value": 0.0},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	pt2, _ := client.NewPoint(
		"test2",
		map[string]string{"hostnam": "192.168.0.1"},
		map[string]interface{}{"value": 1.0},
		time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC),
	)
	var gaugeTests = []struct {
		ptIn     *client.Point
		outGauge *Gauge
		err      error
	}{

		{
			pt1,
			&Gauge{
				Name:        "test1",
				MeasureTime: time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       0.0,
				Source:      "192.168.0.1",
			},
			nil,
		},
		{
			pt2,
			&Gauge{
				Name:        "test2",
				MeasureTime: time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       1.0,
			},
			fmt.Errorf("undeterminable Source type from Field, hostname"),
		},
	}

	l := NewLibrato(fakeUrl)
	l.SourceTag = "hostname"
	for _, gt := range gaugeTests {
		gauges, err := l.buildGauges(gt.ptIn)
		if err != nil && gt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", gt.ptIn.Name(), err)
		}
		if gt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned", gt.ptIn.Name(), gt.err.Error())
		}
		if len(gauges) == 0 {
			continue
		}
		if gt.err == nil && !reflect.DeepEqual(gauges[0], gt.outGauge) {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n", gt.ptIn.Name(), gt.outGauge, gauges[0])
		}
	}
}
