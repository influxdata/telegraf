package zipkin

import "time"

type UnitTest struct {
	expected    []TestData
	measurement string
	datafile    string
	waitPoints  int
}

type TestData struct {
	expectedTags   map[string]string
	expectedValues map[string]interface{}
}

var tests = []UnitTest{
	UnitTest{
		measurement: "zipkin",
		datafile:    "testdata/threespans.dat",
		expected: []TestData{
			{
				expectedTags: map[string]string{
					"id":               "8090652509916334619",
					"parent_id":        "22964302721410078",
					"trace_id":         "2505404965370368069",
					"name":             "Child",
					"service_name":     "trivial",
					"annotation_value": "trivial",
					"endpoint_host":    "2130706433:0",
					"key":              "lc",
					"type":             "STRING",
				},
				expectedValues: map[string]interface{}{
					"duration": time.Duration(53106) * time.Microsecond,
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "103618986556047333",
					"parent_id":        "22964302721410078",
					"trace_id":         "2505404965370368069",
					"name":             "Child",
					"service_name":     "trivial",
					"annotation_value": "trivial",
					"endpoint_host":    "2130706433:0",
					"key":              "lc",
					"type":             "STRING",
				},
				expectedValues: map[string]interface{}{
					"duration": time.Duration(50410) * time.Microsecond,
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "Starting child #0",
					"endpoint_host":    "2130706433:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1498688360851325),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "Starting child #1",
					"endpoint_host":    "2130706433:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1498688360904545),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "A Log",
					"endpoint_host":    "2130706433:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1498688360954992),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "trivial",
					"endpoint_host":    "2130706433:0",
					"key":              "lc",
					"type":             "STRING",
				},
				expectedValues: map[string]interface{}{
					"duration": time.Duration(103680) * time.Microsecond,
					"time":     time.Unix(1498688360, 851318*int64(time.Microsecond)),
				},
			},
		},
	},

	// Test data from the cli app
	UnitTest{
		measurement: "zipkin",
		datafile:    "testdata/file.dat",
		expected: []TestData{
			{
				expectedTags: map[string]string{
					"id":               "3383422996321511664",
					"parent_id":        "4574092882326506380",
					"trace_id":         "8269862291023777619243463817635710260",
					"name":             "Concat",
					"service_name":     "cli",
					"annotation_value": "cs",
					"endpoint_host":    "0:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1499817952283903),
				},
			},
		},
	},
}
