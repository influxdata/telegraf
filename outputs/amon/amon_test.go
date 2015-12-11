package amon

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/influxdb/telegraf/testutil"

	"github.com/influxdb/influxdb/client/v2"
)

func TestBuildPoint(t *testing.T) {
	var tagtests = []struct {
		ptIn  *client.Point
		outPt Point
		err   error
	}{
		{
			testutil.TestPoint(float64(0.0)),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				0.0,
			},
			nil,
		},
		{
			testutil.TestPoint(float64(1.0)),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				1.0,
			},
			nil,
		},
		{
			testutil.TestPoint(int(10)),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				10.0,
			},
			nil,
		},
		{
			testutil.TestPoint(int32(112345)),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
			},
			nil,
		},
		{
			testutil.TestPoint(int64(112345)),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				112345.0,
			},
			nil,
		},
		{
			testutil.TestPoint(float32(11234.5)),
			Point{
				float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
				11234.5,
			},
			nil,
		},
		{
			testutil.TestPoint("11234.5"),
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
