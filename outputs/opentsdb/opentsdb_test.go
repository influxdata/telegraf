package opentsdb

import (
	"reflect"
	"testing"

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
}
