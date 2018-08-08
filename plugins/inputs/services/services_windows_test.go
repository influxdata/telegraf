package services

import (
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestServices(t *testing.T) {
	tests := []struct {
		name   string
		svc    win32service
		tags   map[string]string
		fields map[string]interface{}
		err    error
	}{
		{
			name: "Testing startmode=auto state=running status=ok",
			svc: win32service{
				//ExitCode: 0,
				Name:      "ExampleSvc",
				ProcessID: 1,
				StartMode: "Auto",
				State:     "Running",
				Status:    "OK",
			},
			tags:   map[string]string{"name": "ExampleSvc"},
			fields: map[string]interface{}{"status": 0},
		},
		{
			name: "Testing startmode=auto state=stopped status=failed",
			svc: win32service{
				//ExitCode: 0,
				Name:      "ExampleSvc",
				ProcessID: 1,
				StartMode: "Auto",
				State:     "Stopped",
				Status:    "FAILED",
			},
			tags:   map[string]string{"name": "ExampleSvc"},
			fields: map[string]interface{}{"status": 2},
		},
		{
			name: "Testing startmode=foobar state=foobar status=foobar",
			svc: win32service{
				//ExitCode: 0,
				Name:      "ExampleSvc",
				ProcessID: 1,
				StartMode: "foobar",
				State:     "foobar",
				Status:    "foobar",
			},
			tags:   map[string]string{"name": "ExampleSvc"},
			fields: map[string]interface{}{"status": 3},
		},
		{
			name: "Testing startmode=manual state=stopped status=ok",
			svc: win32service{
				//ExitCode: 0,
				Name:      "ExampleSvc",
				ProcessID: 1,
				StartMode: "Manual",
				State:     "Stopped",
				Status:    "OK",
			},
			tags:   map[string]string{"name": "ExampleSvc"},
			fields: map[string]interface{}{"status": 0},
		},
		{
			name: "Testing startmode=disabled state=stopped status=ok",
			svc: win32service{
				//ExitCode: 0,
				Name:      "ExampleSvc",
				ProcessID: 1,
				StartMode: "Disabled",
				State:     "Stopped",
				Status:    "OK",
			},
			tags:   map[string]string{"name": "ExampleSvc"},
			fields: map[string]interface{}{"status": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &Services{
				wmiQuery: func(query string, dst interface{}, connectServerArgs ...interface{}) error {
					dst = []win32service{tt.svc}
					return nil
				},
			}
			acc := new(testutil.Accumulator)
			err := acc.GatherError(services.Gather)
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
