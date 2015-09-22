package datadog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/koksan83/telegraf/testutil"

	"github.com/influxdb/influxdb/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fakeUrl    = "http://test.datadog.com"
	fakeApiKey = "123456"
)

func fakeDatadog() *Datadog {
	d := NewDatadog(fakeUrl)
	d.Apikey = fakeApiKey
	return d
}

func TestUriOverride(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`)
	}))
	defer ts.Close()

	d := NewDatadog(ts.URL)
	d.Apikey = "123456"
	err := d.Connect()
	require.NoError(t, err)
	err = d.Write(testutil.MockBatchPoints())
	require.NoError(t, err)
}

func TestBadStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(`{ 'errors': [
    	'Something bad happened to the server.',
    	'Your query made the server very sad.'
  		]
		}`)
	}))
	defer ts.Close()

	d := NewDatadog(ts.URL)
	d.Apikey = "123456"
	err := d.Connect()
	require.NoError(t, err)
	err = d.Write(testutil.MockBatchPoints())
	if err == nil {
		t.Errorf("error expected but none returned")
	} else {
		require.EqualError(t, fmt.Errorf("received bad status code, 500\n"), err.Error())
	}
}

func TestAuthenticatedUrl(t *testing.T) {
	d := fakeDatadog()

	authUrl := d.authenticatedUrl()
	assert.EqualValues(t, fmt.Sprintf("%s?api_key=%s", fakeUrl, fakeApiKey), authUrl)
}

func TestBuildTags(t *testing.T) {
	var tagtests = []struct {
		bpIn    map[string]string
		ptIn    map[string]string
		outTags []string
	}{
		{
			map[string]string{"one": "two"},
			map[string]string{"three": "four"},
			[]string{"one:two", "three:four"},
		},
		{
			map[string]string{"aaa": "bbb"},
			map[string]string{},
			[]string{"aaa:bbb"},
		},
		{
			map[string]string{},
			map[string]string{},
			[]string{},
		},
	}
	for _, tt := range tagtests {
		tags := buildTags(tt.bpIn, tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildPoint(t *testing.T) {
	var tagtests = []struct {
		bpIn  client.BatchPoints
		ptIn  client.Point
		outPt Point
		err   error
	}{
		{
			client.BatchPoints{
				Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			},
			client.Point{
				Fields: map[string]interface{}{"value": 0.0},
			},
			Point{float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()), 0.0},
			nil,
		},
		{
			client.BatchPoints{},
			client.Point{
				Fields: map[string]interface{}{"value": 1.0},
				Time:   time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC),
			},
			Point{float64(time.Date(2010, time.December, 10, 23, 0, 0, 0, time.UTC).Unix()), 1.0},
			nil,
		},
		{
			client.BatchPoints{
				Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			},
			client.Point{
				Fields: map[string]interface{}{"value": 10},
			},
			Point{float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()), 10.0},
			nil,
		},
		{
			client.BatchPoints{
				Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			},
			client.Point{
				Fields: map[string]interface{}{"value": int32(112345)},
			},
			Point{float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()), 112345.0},
			nil,
		},
		{
			client.BatchPoints{
				Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			},
			client.Point{
				Fields: map[string]interface{}{"value": int64(112345)},
			},
			Point{float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()), 112345.0},
			nil,
		},
		{
			client.BatchPoints{
				Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			},
			client.Point{
				Fields: map[string]interface{}{"value": float32(11234.5)},
			},
			Point{float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()), 11234.5},
			nil,
		},
		{
			client.BatchPoints{
				Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			},
			client.Point{
				Fields: map[string]interface{}{"value": "11234.5"},
			},
			Point{float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()), 11234.5},
			fmt.Errorf("unable to extract value from Fields, undeterminable type"),
		},
	}
	for _, tt := range tagtests {
		pt, err := buildPoint(tt.bpIn, tt.ptIn)
		if err != nil && tt.err == nil {
			t.Errorf("unexpected error, %+v\n", err)
		}
		if tt.err != nil && err == nil {
			t.Errorf("expected an error (%s) but none returned", tt.err.Error())
		}
		if !reflect.DeepEqual(pt, tt.outPt) && tt.err == nil {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outPt, pt)
		}
	}
}
