package systemd_timings

import (
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestSystemdTiming(t *testing.T) {
	t.Run("systemdTimingsTestAll", func(t *testing.T) {
		systemdTimings := &SystemdTimings{}
		acc := new(testutil.Accumulator)
		err := acc.GatherError(systemdTimings.Gather)
		if err != nil {
			t.Errorf("Error calling Gather: '%#v'", err)
		}
		for _, metric := range acc.Metrics {
			if !reflect.DeepEqual(metric.Measurement, measurement) {
				t.Errorf("expected measurement '%#v' got '%#v'\n",
					measurement, metric.Measurement)
			}

			unitName, isUnit := metric.Tags["UnitName"]
			tsName, isGlobal := metric.Tags["SystemTimestamp"]
			if !isUnit && !isGlobal {
				t.Errorf("no valid metric tags found, expected either "+
					"UnitName or SystemTimestamp, got: %v\n", metric.Tags)
			}

			if isGlobal {
				value, ok := metric.Fields["SystemTimestampValue"].(uint64)
				if ok {
					if value <= 0 {
						t.Errorf("expected positive timestamp for %s, "+
							"got: %d\n", tsName, value)
					}
				} else {
					t.Errorf("failed to convert %s to an integer\n", tsName)
				}
			} else if isUnit {
				value, ok := metric.Fields["RunDuration"].(uint64)
				if ok {
					if value <= 0 {
						t.Errorf("expected positive timestamp for %s, "+
							"got: %d\n", unitName, value)
					}
				} else {
					t.Errorf("failed to convert %s to an integer\n", unitName)
				}
			}
		}
	})
}
