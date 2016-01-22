package opentsdb

import (
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBuildTagsTelnet(t *testing.T) {
	var tagtests = []struct {
		ptIn    map[string]string
		outTags []string
	}{
		{
			map[string]string{"one": "two", "three": "four"},
			[]string{"one=two", "three=four"},
		},
		{
			map[string]string{"aaa": "bbb"},
			[]string{"aaa=bbb"},
		},
		{
			map[string]string{"one": "two", "aaa": "bbb"},
			[]string{"aaa=bbb", "one=two"},
		},
		{
			map[string]string{},
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

func TestWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &OpenTSDB{
		Host:   testutil.GetLocalHost(),
		Port:   4242,
		Prefix: "prefix.test.",
	}

	// Verify that we can connect to the OpenTSDB instance
	err := o.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to OpenTSDB
	err = o.Write(testutil.MockBatchPoints().Points())
	require.NoError(t, err)

	// Verify postive and negative test cases of writing data
	bp := testutil.MockBatchPoints()
	bp.AddPoint(testutil.TestPoint(float64(1.0), "justametric.float"))
	bp.AddPoint(testutil.TestPoint(int64(123456789), "justametric.int"))
	bp.AddPoint(testutil.TestPoint(uint64(123456789012345), "justametric.uint"))
	bp.AddPoint(testutil.TestPoint("Lorem Ipsum", "justametric.string"))
	bp.AddPoint(testutil.TestPoint(float64(42.0), "justametric.anotherfloat"))

	err = o.Write(bp.Points())
	require.NoError(t, err)

}
