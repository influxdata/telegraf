package zipkin

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/trace"
	"github.com/influxdata/telegraf/testutil"
)

func TestLineProtocolConverter_Record(t *testing.T) {
	mockAcc := testutil.Accumulator{}
	type fields struct {
		acc telegraf.Accumulator
	}
	type args struct {
		t trace.Trace
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    []testutil.Metric
	}{
		{
			name: "threespan",
			fields: fields{
				acc: &mockAcc,
			},
			args: args{
				t: trace.Trace{
					{
						ID:          "8090652509916334619",
						TraceID:     "2505404965370368069",
						Name:        "Child",
						ParentID:    "22964302721410078",
						Timestamp:   time.Unix(0, 1498688360851331000).UTC(),
						Duration:    time.Duration(53106) * time.Microsecond,
						ServiceName: "trivial",
						Annotations: []trace.Annotation{},
						BinaryAnnotations: []trace.BinaryAnnotation{
							{
								Key:         "lc",
								Value:       "dHJpdmlhbA==",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
						},
					},
					{
						ID:          "103618986556047333",
						TraceID:     "2505404965370368069",
						Name:        "Child",
						ParentID:    "22964302721410078",
						Timestamp:   time.Unix(0, 1498688360904552000).UTC(),
						Duration:    time.Duration(50410) * time.Microsecond,
						ServiceName: "trivial",
						Annotations: []trace.Annotation{},
						BinaryAnnotations: []trace.BinaryAnnotation{
							{
								Key:         "lc",
								Value:       "dHJpdmlhbA==",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
						},
					},
					{
						ID:          "22964302721410078",
						TraceID:     "2505404965370368069",
						Name:        "Parent",
						ParentID:    "22964302721410078",
						Timestamp:   time.Unix(0, 1498688360851318000).UTC(),
						Duration:    time.Duration(103680) * time.Microsecond,
						ServiceName: "trivial",
						Annotations: []trace.Annotation{
							{
								Timestamp:   time.Unix(0, 1498688360851325000).UTC(),
								Value:       "Starting child #0",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
							{
								Timestamp:   time.Unix(0, 1498688360904545000).UTC(),
								Value:       "Starting child #1",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
							{
								Timestamp:   time.Unix(0, 1498688360954992000).UTC(),
								Value:       "A Log",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
						},
						BinaryAnnotations: []trace.BinaryAnnotation{
							{
								Key:         "lc",
								Value:       "dHJpdmlhbA==",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
						},
					},
				},
			},
			want: []testutil.Metric{
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "8090652509916334619",
						"parent_id":    "22964302721410078",
						"trace_id":     "2505404965370368069",
						"service_name": "trivial",
						"name":         "child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "8090652509916334619",
						"parent_id":      "22964302721410078",
						"trace_id":       "2505404965370368069",
						"name":           "child",
						"service_name":   "trivial",
						"annotation":     "dHJpdmlhbA==",
						"endpoint_host":  "2130706433:0",
						"annotation_key": "lc",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "103618986556047333",
						"parent_id":    "22964302721410078",
						"trace_id":     "2505404965370368069",
						"service_name": "trivial",
						"name":         "child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "103618986556047333",
						"parent_id":      "22964302721410078",
						"trace_id":       "2505404965370368069",
						"name":           "child",
						"service_name":   "trivial",
						"annotation":     "dHJpdmlhbA==",
						"endpoint_host":  "2130706433:0",
						"annotation_key": "lc",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "22964302721410078",
						"parent_id":    "22964302721410078",
						"trace_id":     "2505404965370368069",
						"service_name": "trivial",
						"name":         "parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #0",
						"endpoint_host": "2130706433:0",
						"id":            "22964302721410078",
						"parent_id":     "22964302721410078",
						"trace_id":      "2505404965370368069",
						"name":          "parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #1",
						"endpoint_host": "2130706433:0",
						"id":            "22964302721410078",
						"parent_id":     "22964302721410078",
						"trace_id":      "2505404965370368069",
						"name":          "parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"parent_id":     "22964302721410078",
						"trace_id":      "2505404965370368069",
						"name":          "parent",
						"service_name":  "trivial",
						"annotation":    "A Log",
						"endpoint_host": "2130706433:0",
						"id":            "22964302721410078",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"trace_id":       "2505404965370368069",
						"service_name":   "trivial",
						"annotation":     "dHJpdmlhbA==",
						"annotation_key": "lc",
						"id":             "22964302721410078",
						"parent_id":      "22964302721410078",
						"name":           "parent",
						"endpoint_host":  "2130706433:0",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
			},
			wantErr: false,
		},

		//// Test data from distributed trace repo sample json
		// https://github.com/mattkanwisher/distributedtrace/blob/master/testclient/sample.json
		{
			name: "distributed_trace_sample",
			fields: fields{
				acc: &mockAcc,
			},
			args: args{
				t: trace.Trace{
					{
						ID:          "6802735349851856000",
						TraceID:     "0:6802735349851856000",
						Name:        "main.dud",
						ParentID:    "6802735349851856000",
						Timestamp:   time.Unix(1, 0).UTC(),
						Duration:    1,
						ServiceName: "trivial",
						Annotations: []trace.Annotation{
							{
								Timestamp:   time.Unix(0, 1433330263415871000).UTC(),
								Value:       "cs",
								Host:        "0:9410",
								ServiceName: "go-zipkin-testclient",
							},
						},
						BinaryAnnotations: []trace.BinaryAnnotation{},
					},
				},
			},
			want: []testutil.Metric{
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "6802735349851856000",
						"parent_id":    "6802735349851856000",
						"trace_id":     "0:6802735349851856000",
						"name":         "main.dud",
						"service_name": "trivial",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Nanosecond).Nanoseconds(),
					},
					Time: time.Unix(1, 0).UTC(),
				},
				{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation":    "cs",
						"endpoint_host": "0:9410",
						"id":            "6802735349851856000",
						"parent_id":     "6802735349851856000",
						"trace_id":      "0:6802735349851856000",
						"name":          "main.dud",
						"service_name":  "go-zipkin-testclient",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(1) * time.Nanosecond).Nanoseconds(),
					},
					Time: time.Unix(1, 0).UTC(),
				},
			},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAcc.ClearMetrics()
			l := &LineProtocolConverter{
				acc: tt.fields.acc,
			}
			if err := l.Record(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("LineProtocolConverter.Record() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := []testutil.Metric{}
			for _, metric := range mockAcc.Metrics {
				got = append(got, *metric)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("LineProtocolConverter.Record()/%s/%d error = %s ", tt.name, i, cmp.Diff(got, tt.want))
			}
		})
	}
}
