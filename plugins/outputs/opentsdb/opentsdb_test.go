package opentsdb

import (
	"reflect"
	"testing"
	// "github.com/influxdata/telegraf/testutil"
	// "github.com/stretchr/testify/require"
)

func TestCleanTags(t *testing.T) {
	var tagtests = []struct {
		ptIn    map[string]string
		outTags TagSet
	}{
		{
			map[string]string{"one": "two", "three": "four"},
			TagSet{"one": "two", "three": "four"},
		},
		{
			map[string]string{"aaa": "bbb"},
			TagSet{"aaa": "bbb"},
		},
		{
			map[string]string{"Sp%ci@l Chars": "g$t repl#ced"},
			TagSet{"Sp-ci-l_Chars": "g-t_repl-ced"},
		},
		{
			map[string]string{},
			TagSet{},
		},
	}
	for _, tt := range tagtests {
		tags := cleanTags(tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

func TestBuildTagsTelnet(t *testing.T) {
	var tagtests = []struct {
		ptIn    TagSet
		outTags string
	}{
		{
			TagSet{"one": "two", "three": "four"},
			"one=two three=four",
		},
		{
			TagSet{"aaa": "bbb"},
			"aaa=bbb",
		},
		{
			TagSet{"one": "two", "aaa": "bbb"},
			"aaa=bbb one=two",
		},
		{
			TagSet{},
			"",
		},
	}
	for _, tt := range tagtests {
		tags := tt.ptIn.ToLineFormat()
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected %+v\ngot %+v\n", tt.outTags, tags)
		}
	}
}

// func TestWrite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	o := &OpenTSDB{
// 		Host:   testutil.GetLocalHost(),
// 		Port:   4242,
// 		Prefix: "prefix.test.",
// 	}

// 	// Verify that we can connect to the OpenTSDB instance
// 	err := o.Connect()
// 	require.NoError(t, err)

// 	// Verify that we can successfully write data to OpenTSDB
// 	err = o.Write(testutil.MockMetrics())
// 	require.NoError(t, err)

// 	// Verify postive and negative test cases of writing data
// 	metrics := testutil.MockMetrics()
// 	metrics = append(metrics, testutil.TestMetric(float64(1.0),
// 		"justametric.float"))
// 	metrics = append(metrics, testutil.TestMetric(int64(123456789),
// 		"justametric.int"))
// 	metrics = append(metrics, testutil.TestMetric(uint64(123456789012345),
// 		"justametric.uint"))
// 	metrics = append(metrics, testutil.TestMetric("Lorem Ipsum",
// 		"justametric.string"))
// 	metrics = append(metrics, testutil.TestMetric(float64(42.0),
// 		"justametric.anotherfloat"))
// 	metrics = append(metrics, testutil.TestMetric(float64(42.0),
// 		"metric w/ specialchars"))

// 	err = o.Write(metrics)
// 	require.NoError(t, err)
// }
