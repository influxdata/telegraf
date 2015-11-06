package amon

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fakeServerKey    = "123456"
	fakeAmonInstance = "https://demo.amon.cx"
)

func TestUriOverride(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`)
	}))
	defer ts.Close()

	a := &Amon{
		ServerKey:    fakeServerKey,
		AmonInstance: fakeAmonInstance,
	}

	err := a.Connect()
	require.NoError(t, err)
	err = a.Write(testutil.MockBatchPoints().Points())
	require.NoError(t, err)
}

func TestAuthenticatedUrl(t *testing.T) {
	a := &Amon{
		ServerKey:    fakeServerKey,
		AmonInstance: fakeAmonInstance,
	}

	authUrl := a.authenticatedUrl()
	assert.EqualValues(t, fmt.Sprintf("%s/api/system/%s", fakeAmonInstance, fakeServerKey), authUrl)
}

func TestBuildPoint(t *testing.T) {
	tags := make(map[string]string)
	var tagtests = []struct {
		ptIn  *client.Point
		outPt Point
		err   error
	}{
		{
			client.NewPoint(
				"test1",
				tags,
				map[string]interface{}{"value": 0.0},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
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
			Point{
				float64(time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
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
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				10.0,
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
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
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
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
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
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				11234.5,
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
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				11234.5,
			},
			fmt.Errorf("unable to extract value from Fields, undeterminable type"),
		},
	}
	for _, tt := range tagtests {
		pt, err := buildPoint(tt.ptIn)
		if err != nil && tt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", tt.ptIn.Name(), err)
		}
		if tt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned", tt.ptIn.Name(), tt.err.Error())
		}
		if !reflect.DeepEqual(pt, tt.outPt) && tt.err == nil {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n", tt.ptIn.Name(), tt.outPt, pt)
		}
	}
}
