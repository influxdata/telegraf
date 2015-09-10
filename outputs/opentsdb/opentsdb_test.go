package opentsdb

import (
	"reflect"
	"testing"
)

var (
	fakeHost = "metrics.example.com"
	fakePort = 4242
)

func fakeOpenTSDB() *OpenTSDB {
	var o OpenTSDB
	o.Host = fakeHost
	o.Port = fakePort
	return &o
}

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
