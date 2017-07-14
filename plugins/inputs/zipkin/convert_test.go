package zipkin

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestLineProtocolConverter_Record(t *testing.T) {
	mockAcc := testutil.Accumulator{}
	type fields struct {
		acc telegraf.Accumulator
	}
	type args struct {
		t Trace
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
				t: Trace{
					Span{
						ID:          "8090652509916334619",
						TraceID:     "2505404965370368069",
						Name:        "Child",
						ParentID:    "22964302721410078",
						Timestamp:   time.Unix(0, 1498688360851331000),
						Duration:    time.Duration(53106) * time.Microsecond,
						Annotations: []Annotation{},
						BinaryAnnotations: []BinaryAnnotation{
							BinaryAnnotation{
								Key:         "lc",
								Value:       "dHJpdmlhbA==",
								Host:        "2130706433:0",
								ServiceName: "trivial",
								Type:        "STRING",
							},
						},
					},
					Span{
						ID:          "103618986556047333",
						TraceID:     "2505404965370368069",
						Name:        "Child",
						ParentID:    "22964302721410078",
						Timestamp:   time.Unix(0, 1498688360904552000),
						Duration:    time.Duration(50410) * time.Microsecond,
						Annotations: []Annotation{},
						BinaryAnnotations: []BinaryAnnotation{
							BinaryAnnotation{
								Key:         "lc",
								Value:       "dHJpdmlhbA==",
								Host:        "2130706433:0",
								ServiceName: "trivial",
								Type:        "STRING",
							},
						},
					},
				},
			},
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":               "8090652509916334619",
						"parent_id":        "22964302721410078",
						"trace_id":         "2505404965370368069",
						"name":             "Child",
						"service_name":     "trivial",
						"annotation_value": "dHJpdmlhbA==",
						"endpoint_host":    "2130706433:0",
						"key":              "lc",
						"type":             "STRING",
					},
					Fields: map[string]interface{}{
						"duration": time.Duration(53106) * time.Microsecond,
					},
					Time: time.Unix(0, 1498688360851331000),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":               "103618986556047333",
						"parent_id":        "22964302721410078",
						"trace_id":         "2505404965370368069",
						"name":             "Child",
						"service_name":     "trivial",
						"annotation_value": "dHJpdmlhbA==",
						"endpoint_host":    "2130706433:0",
						"key":              "lc",
						"type":             "STRING",
					},
					Fields: map[string]interface{}{
						"duration": time.Duration(50410) * time.Microsecond,
					},
					Time: time.Unix(0, 1498688360904552000),
				},
			},
			wantErr: false,
		},

		/*{
			name:    "changeMe2",
			fields:  fields{},
			args:    args{},
			wantErr: false,
		},*/
	}
	for _, tt := range tests {
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LineProtocolConverter.Record() error = \n%#v\n, want \n%#v\n", got, tt.want)
			}
		})
	}
}
