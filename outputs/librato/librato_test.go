package librato

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdb/telegraf/testutil"

	"github.com/influxdb/influxdb/client/v2"
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
	tags := make(map[string]string)
	var gaugeTests = []struct {
		ptIn     *client.Point
		outGauge *Gauge
		err      error
	}{
		{
			client.NewPoint(
				"test1",
				tags,
				map[string]interface{}{"value": 0.0},
				time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test1",
				MeasureTime: time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       0.0,
			},
			nil,
		},
		{
			client.NewPoint(
				"test2",
				tags,
				map[string]interface{}{"value": 1.0},
				time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test2",
				MeasureTime: time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       1.0,
			},
			nil,
		},
		{
			client.NewPoint(
				"test3",
				tags,
				map[string]interface{}{"value": 10},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test3",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       10.0,
			},
			nil,
		},
		{
			client.NewPoint(
				"test4",
				tags,
				map[string]interface{}{"value": int32(112345)},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test4",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       112345.0,
			},
			nil,
		},
		{
			client.NewPoint(
				"test5",
				tags,
				map[string]interface{}{"value": int64(112345)},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test5",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       112345.0,
			},
			nil,
		},
		{
			client.NewPoint(
				"test6",
				tags,
				map[string]interface{}{"value": float32(11234.5)},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test6",
				MeasureTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       11234.5,
			},
			nil,
		},
		{
			client.NewPoint(
				"test7",
				tags,
				map[string]interface{}{"value": "11234.5"},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
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
		gauge, err := l.buildGauge(gt.ptIn)
		if err != nil && gt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", gt.ptIn.Name(), err)
		}
		if gt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned", gt.ptIn.Name(), gt.err.Error())
		}
		if !reflect.DeepEqual(gauge, gt.outGauge) && gt.err == nil {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n", gt.ptIn.Name(), gt.outGauge, gauge)
		}
	}
}

func TestBuildGaugeWithSource(t *testing.T) {
	var gaugeTests = []struct {
		ptIn     *client.Point
		outGauge *Gauge
		err      error
	}{
		{
			client.NewPoint(
				"test1",
				map[string]string{"hostname": "192.168.0.1"},
				map[string]interface{}{"value": 0.0},
				time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			&Gauge{
				Name:        "test1",
				MeasureTime: time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC).Unix(),
				Value:       0.0,
				Source:      "192.168.0.1",
			},
			nil,
		},
		{
			client.NewPoint(
				"test2",
				map[string]string{"hostnam": "192.168.0.1"},
				map[string]interface{}{"value": 1.0},
				time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC),
			),
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
		gauge, err := l.buildGauge(gt.ptIn)
		if err != nil && gt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", gt.ptIn.Name(), err)
		}
		if gt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned", gt.ptIn.Name(), gt.err.Error())
		}
		if !reflect.DeepEqual(gauge, gt.outGauge) && gt.err == nil {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n", gt.ptIn.Name(), gt.outGauge, gauge)
		}
	}
}
