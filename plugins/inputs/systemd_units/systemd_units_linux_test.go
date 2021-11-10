package systemd_units

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestSystemdUnits(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		tags   map[string]string
		fields map[string]interface{}
		status int
		err    error
	}{
		{
			name: "example loaded active running",
			line: "example.service                loaded active running example service description",
			tags: map[string]string{"name": "example.service", "load": "loaded", "active": "active", "sub": "running"},
			fields: map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    0,
			},
		},
		{
			name: "example loaded active exited",
			line: "example.service                loaded active exited  example service description",
			tags: map[string]string{"name": "example.service", "load": "loaded", "active": "active", "sub": "exited"},
			fields: map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    4,
			},
		},
		{
			name: "example loaded failed failed",
			line: "example.service                loaded failed failed  example service description",
			tags: map[string]string{"name": "example.service", "load": "loaded", "active": "failed", "sub": "failed"},
			fields: map[string]interface{}{
				"load_code":   0,
				"active_code": 3,
				"sub_code":    12,
			},
		},
		{
			name: "example not-found inactive dead",
			line: "example.service                not-found inactive dead  example service description",
			tags: map[string]string{"name": "example.service", "load": "not-found", "active": "inactive", "sub": "dead"},
			fields: map[string]interface{}{
				"load_code":   2,
				"active_code": 2,
				"sub_code":    1,
			},
		},
		{
			name: "example unknown unknown unknown",
			line: "example.service                unknown unknown unknown  example service description",
			err:  fmt.Errorf("Error parsing field 'load', value not in map: %s", "unknown"),
		},
		{
			name: "example too few fields",
			line: "example.service                loaded fai",
			err:  fmt.Errorf("Error parsing line (expected at least 4 fields): %s", "example.service                loaded fai"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			systemdUnits := &SystemdUnits{
				systemctl: func(timeout config.Duration, unitType string, pattern string) (*bytes.Buffer, error) {
					return bytes.NewBufferString(tt.line), nil
				},
			}
			acc := new(testutil.Accumulator)
			err := acc.GatherError(systemdUnits.Gather)
			if !reflect.DeepEqual(tt.err, err) {
				t.Errorf("%s: expected error '%#v' got '%#v'", tt.name, tt.err, err)
			}
			if len(acc.Metrics) > 0 {
				m := acc.Metrics[0]
				if !reflect.DeepEqual(m.Measurement, measurement) {
					t.Errorf("%s: expected measurement '%#v' got '%#v'\n", tt.name, measurement, m.Measurement)
				}
				if !reflect.DeepEqual(m.Tags, tt.tags) {
					t.Errorf("%s: expected tags\n%#v got\n%#v\n", tt.name, tt.tags, m.Tags)
				}
				if !reflect.DeepEqual(m.Fields, tt.fields) {
					t.Errorf("%s: expected fields\n%#v got\n%#v\n", tt.name, tt.fields, m.Fields)
				}
			}
		})
	}
}
