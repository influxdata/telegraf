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
					Span{
						ID:        "22964302721410078",
						TraceID:   "2505404965370368069",
						Name:      "Parent",
						ParentID:  "22964302721410078",
						Timestamp: time.Unix(0, 1498688360851318000),
						Duration:  time.Duration(103680) * time.Microsecond,
						Annotations: []Annotation{
							Annotation{
								Timestamp:   time.Unix(0, 1498688360851325000),
								Value:       "Starting child #0",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
							Annotation{
								Timestamp:   time.Unix(0, 1498688360904545000),
								Value:       "Starting child #1",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
							Annotation{
								Timestamp:   time.Unix(0, 1498688360954992000),
								Value:       "A Log",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
						},
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
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":     "trivial",
						"annotation_value": "Starting child #0",
						"endpoint_host":    "2130706433:0",
						"id":               "22964302721410078",
						"parent_id":        "22964302721410078",
						"trace_id":         "2505404965370368069",
						"name":             "Parent",
					},
					Fields: map[string]interface{}{
						"annotation_timestamp": int64(1498688360),
						"duration":             time.Duration(103680) * time.Microsecond,
					},
					Time: time.Unix(0, 1498688360851318000),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":     "trivial",
						"annotation_value": "Starting child #1",
						"endpoint_host":    "2130706433:0",
						"id":               "22964302721410078",
						"parent_id":        "22964302721410078",
						"trace_id":         "2505404965370368069",
						"name":             "Parent",
					},
					Fields: map[string]interface{}{
						"annotation_timestamp": int64(1498688360),
						"duration":             time.Duration(103680) * time.Microsecond,
					},
					Time: time.Unix(0, 1498688360851318000),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"parent_id":        "22964302721410078",
						"trace_id":         "2505404965370368069",
						"name":             "Parent",
						"service_name":     "trivial",
						"annotation_value": "A Log",
						"endpoint_host":    "2130706433:0",
						"id":               "22964302721410078",
					},
					Fields: map[string]interface{}{
						"annotation_timestamp": int64(1498688360),
						"duration":             time.Duration(103680) * time.Microsecond,
					},
					Time: time.Unix(0, 1498688360851318000),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"trace_id":         "2505404965370368069",
						"service_name":     "trivial",
						"annotation_value": "dHJpdmlhbA==",
						"key":              "lc",
						"type":             "STRING",
						"id":               "22964302721410078",
						"parent_id":        "22964302721410078",
						"name":             "Parent",
						"endpoint_host":    "2130706433:0",
					},
					Fields: map[string]interface{}{
						"duration": time.Duration(103680) * time.Microsecond,
					},
					Time: time.Unix(0, 1498688360851318000),
				},
			},
			wantErr: false,
		},

		// Test data from zipkin cli app:
		//https://github.com/openzipkin/zipkin-go-opentracing/tree/master/examples/cli_with_2_services
		/*{
			name:    "cli",
			fields:  fields{
			acc: &mockAcc,
		},
			args:    args{
			t: Trace{
				Span{
					ID:          "3383422996321511664",
					TraceID:     "243463817635710260",
					Name:        "Concat",
					ParentID:    "4574092882326506380",
					Timestamp:   time.Unix(0, 1499817952283903000),
					Duration:    time.Duration(2888) * time.Microsecond,
					Annotations: []Annotation{
						Annotaitons{
							Timestamp:   time.Unix(0, 1499817952283903000),
							Value:       "cs",
							Host:        "0:0",
							ServiceName: "cli",
						},
				},
					BinaryAnnotations: []BinaryAnnotation{
						BinaryAnnotation{
							Key:         "http.url",
							Value:       "aHR0cDovL2xvY2FsaG9zdDo2MTAwMS9jb25jYXQv",
							Host:        "0:0",
							ServiceName: "cli",
							Type:        "STRING",
						},
					},
		},
		want: []testutil.Metric{
		testutil.Metric{
			Measurement: "zipkin",
			Tags: map[string]string{
				"id":               "3383422996321511664",
				"parent_id":        "4574092882326506380",
				"trace_id":         "8269862291023777619:243463817635710260",
				"name":             "Concat",
				"service_name":     "cli",
				"annotation_value": "cs",
				"endpoint_host":    "0:0",
			},
			Fields: map[string]interface{}{
			"annotation_timestamp": int64(149981795),
				"duration": time.Duration(2888) * time.Microsecond,
			},
			Time: time.Unix(0, 1499817952283903000),
		},
		testutil.Metric{
			Measurement: "zipkin",
			Tags: map[string]string{
			"trace_id":         "2505404965370368069",
			"service_name":     "cli",
			"annotation_value": "aHR0cDovL2xvY2FsaG9zdDo2MTAwMS9jb25jYXQv",
			"key":              "http.url",
			"type":             "STRING",
			"id":               "22964302721410078",
			"parent_id":        "22964302721410078",
			"name":             "Concat",
			"endpoint_host":    "0:0",
			},
			Fields: map[string]interface{}{
				"duration": time.Duration(2888) * time.Microsecond,
			},
			Time: time.Unix(0, 1499817952283903000),
		},
			wantErr: false,
		},*/

		//// Test data from distributed trace repo sample json
		// https://github.com/mattkanwisher/distributedtrace/blob/master/testclient/sample.json
		{
			name: "distributed_trace_sample",
			fields: fields{
				acc: &mockAcc,
			},
			args: args{
				t: Trace{
					Span{
						ID:        "6802735349851856000",
						TraceID:   "0:6802735349851856000",
						Name:      "main.dud",
						ParentID:  "6802735349851856000",
						Timestamp: time.Unix(1, 0),
						Duration:  1,
						Annotations: []Annotation{
							Annotation{
								Timestamp:   time.Unix(0, 1433330263415871000),
								Value:       "cs",
								Host:        "0:9410",
								ServiceName: "go-zipkin-testclient",
							},
						},
						BinaryAnnotations: []BinaryAnnotation{},
					},
				},
			},
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"annotation_value": "cs",
						"endpoint_host":    "0:9410",
						"id":               "6802735349851856000",
						"parent_id":        "6802735349851856000",
						"trace_id":         "0:6802735349851856000",
						"name":             "main.dud",
						"service_name":     "go-zipkin-testclient",
					},
					Fields: map[string]interface{}{
						"annotation_timestamp": int64(1433330263),
						"duration":             time.Duration(1) * time.Nanosecond,
					},
					Time: time.Unix(1, 0),
				},
			},
		},
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
