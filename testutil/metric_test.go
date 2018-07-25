package testutil

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestMustEqual(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name string
		got  telegraf.Metric
		want Metric
	}{
		{
			name: "telegraf and testutil metrics should be equal",
			got: func() telegraf.Metric {
				m, _ := metric.New(
					"test",
					map[string]string{
						"t1": "v1",
						"t2": "v2",
					},
					map[string]interface{}{
						"f1": 1,
						"f2": 3.14,
						"f3": "v3",
					},
					time.Unix(0, 0),
				)
				return m
			}(),
			want: Metric{
				Measurement: "test",
				Tags: map[string]string{
					"t1": "v1",
					"t2": "v2",
				},
				Fields: map[string]interface{}{
					"f1": int64(1),
					"f2": 3.14,
					"f3": "v3",
				},
				Time: time.Unix(0, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MustEqual(t, tt.got, tt.want)
		})
	}
}
