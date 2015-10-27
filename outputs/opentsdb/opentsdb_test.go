package opentsdb

import (
	"reflect"
	"testing"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/testutil"
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
	tags := make(map[string]string)
	bp.AddPoint(client.NewPoint("justametric.float", tags,
		map[string]interface{}{"value": float64(1.0)}))
	bp.AddPoint(client.NewPoint("justametric.int", tags,
		map[string]interface{}{"value": int64(123456789)}))
	bp.AddPoint(client.NewPoint("justametric.uint", tags,
		map[string]interface{}{"value": uint64(123456789012345)}))
	bp.AddPoint(client.NewPoint("justametric.string", tags,
		map[string]interface{}{"value": "Lorem Ipsum"}))
	bp.AddPoint(client.NewPoint("justametric.anotherfloat", tags,
		map[string]interface{}{"value": float64(42.0)}))

	err = o.Write(bp.Points())
	require.NoError(t, err)

}
