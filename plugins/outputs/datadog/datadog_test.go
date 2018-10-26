package datadog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fakeUrl    = "http://test.datadog.com"
	fakeApiKey = "123456"
)

func NewDatadog(url string) *Datadog {
	return &Datadog{
		URL: url,
	}
}

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
	err = d.Write(testutil.MockMetrics())
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
	err = d.Write(testutil.MockMetrics())
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
		ptIn    []*telegraf.Tag
		outTags []string
	}{
		{
			[]*telegraf.Tag{
				{
					Key:   "one",
					Value: "two",
				},
				{
					Key:   "three",
					Value: "four",
				},
			},
			[]string{"one:two", "three:four"},
		},
		{
			[]*telegraf.Tag{
				{
					Key:   "aaa",
					Value: "bbb",
				},
			},
			[]string{"aaa:bbb"},
		},
		{
			[]*telegraf.Tag{},
			[]string{},
		},
	}
	for _, tt := range tagtests {
		tags := buildTags(tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildPoint(t *testing.T) {
	var tagtests = []struct {
		ptIn  telegraf.Metric
		outPt Point
		err   error
	}{
		{
			testutil.TestMetric(0.0, "test1"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(1.0, "test2"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
		{
			testutil.TestMetric(10, "test3"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				10.0,
			},
			nil,
		},
		{
			testutil.TestMetric(int32(112345), "test4"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
			},
			nil,
		},
		{
			testutil.TestMetric(int64(112345), "test5"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
			},
			nil,
		},
		{
			testutil.TestMetric(float32(11234.5), "test6"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				11234.5,
			},
			nil,
		},
		{
			testutil.TestMetric(bool(true), "test7"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
		{
			testutil.TestMetric(bool(false), "test8"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(int64(0), "test int64"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(uint64(0), "test uint64"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestMetric(true, "test bool"),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
	}
	for _, tt := range tagtests {
		pt, err := buildMetrics(tt.ptIn)
		if err != nil && tt.err == nil {
			t.Errorf("%s: unexpected error, %+v\n", tt.ptIn.Name(), err)
		}
		if tt.err != nil && err == nil {
			t.Errorf("%s: expected an error (%s) but none returned", tt.ptIn.Name(), tt.err.Error())
		}
		if !reflect.DeepEqual(pt["value"], tt.outPt) && tt.err == nil {
			t.Errorf("%s: \nexpected %+v\ngot %+v\n",
				tt.ptIn.Name(), tt.outPt, pt["value"])
		}
	}
}

func TestVerifyValue(t *testing.T) {
	var tagtests = []struct {
		ptIn        telegraf.Metric
		validMetric bool
	}{
		{
			testutil.TestMetric(float32(11234.5), "test1"),
			true,
		},
		{
			testutil.TestMetric("11234.5", "test2"),
			false,
		},
	}
	for _, tt := range tagtests {
		ok := verifyValue(tt.ptIn.Fields()["value"])
		if tt.validMetric != ok {
			t.Errorf("%s: verification failed\n", tt.ptIn.Name())
		}
	}
}
