package opentsdb

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBuildTagsTelnet(t *testing.T) {
	var tagtests = []struct {
		bpIn    map[string]string
		ptIn    map[string]string
		outTags []string
	}{
		{
			map[string]string{"one": "two"},
			map[string]string{"three": "four"},
			[]string{"one=two", "three=four"},
		},
		{
			map[string]string{"aaa": "bbb"},
			map[string]string{},
			[]string{"aaa=bbb"},
		},
		{
			map[string]string{"one": "two"},
			map[string]string{"aaa": "bbb"},
			[]string{"aaa=bbb", "one=two"},
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
func TestWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &OpenTSDB{
		Host:   testutil.GetLocalHost(),
		Port:   24242,
		Prefix: "prefix.test.",
	}

	// Verify that we can connect to the OpenTSDB instance
	err := o.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to OpenTSDB
	err = o.Write(testutil.MockBatchPoints())
	require.NoError(t, err)

	// Verify postive and negative test cases of writing data
	var bp client.BatchPoints
	bp.Time = time.Now()
	bp.Tags = map[string]string{"testkey": "testvalue"}
	bp.Points = []client.Point{
		{
			Measurement: "justametric.float",
			Fields:      map[string]interface{}{"value": float64(1.0)},
		},
		{
			Measurement: "justametric.int",
			Fields:      map[string]interface{}{"value": int64(123456789)},
		},
		{
			Measurement: "justametric.uint",
			Fields:      map[string]interface{}{"value": uint64(123456789012345)},
		},
		{
			Measurement: "justametric.string",
			Fields:      map[string]interface{}{"value": "Lorem Ipsum"},
		},
		{
			Measurement: "justametric.anotherfloat",
			Fields:      map[string]interface{}{"value": float64(42.0)},
		},
	}
	err = o.Write(bp)
	require.NoError(t, err)

}
