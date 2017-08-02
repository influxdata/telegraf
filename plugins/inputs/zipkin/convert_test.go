package zipkin

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
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
						Timestamp:   time.Unix(0, 1498688360851331000).UTC(),
						Duration:    time.Duration(53106) * time.Microsecond,
						ServiceName: "trivial",
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
						Timestamp:   time.Unix(0, 1498688360904552000).UTC(),
						Duration:    time.Duration(50410) * time.Microsecond,
						ServiceName: "trivial",
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
						ID:          "22964302721410078",
						TraceID:     "2505404965370368069",
						Name:        "Parent",
						ParentID:    "22964302721410078",
						Timestamp:   time.Unix(0, 1498688360851318000).UTC(),
						Duration:    time.Duration(103680) * time.Microsecond,
						ServiceName: "trivial",
						Annotations: []Annotation{
							Annotation{
								Timestamp:   time.Unix(0, 1498688360851325000).UTC(),
								Value:       "Starting child #0",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
							Annotation{
								Timestamp:   time.Unix(0, 1498688360904545000).UTC(),
								Value:       "Starting child #1",
								Host:        "2130706433:0",
								ServiceName: "trivial",
							},
							Annotation{
								Timestamp:   time.Unix(0, 1498688360954992000).UTC(),
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
						"id":           "8090652509916334619",
						"parent_id":    "22964302721410078",
						"trace_id":     "2505404965370368069",
						"service_name": "trivial",
						"name":         "Child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(53106) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851331000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "8090652509916334619",
						"parent_id":      "22964302721410078",
						"trace_id":       "2505404965370368069",
						"name":           "Child",
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
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "103618986556047333",
						"parent_id":    "22964302721410078",
						"trace_id":     "2505404965370368069",
						"service_name": "trivial",
						"name":         "Child",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(50410) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360904552000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":             "103618986556047333",
						"parent_id":      "22964302721410078",
						"trace_id":       "2505404965370368069",
						"name":           "Child",
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
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"id":           "22964302721410078",
						"parent_id":    "22964302721410078",
						"trace_id":     "2505404965370368069",
						"service_name": "trivial",
						"name":         "Parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #0",
						"endpoint_host": "2130706433:0",
						"id":            "22964302721410078",
						"parent_id":     "22964302721410078",
						"trace_id":      "2505404965370368069",
						"name":          "Parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"service_name":  "trivial",
						"annotation":    "Starting child #1",
						"endpoint_host": "2130706433:0",
						"id":            "22964302721410078",
						"parent_id":     "22964302721410078",
						"trace_id":      "2505404965370368069",
						"name":          "Parent",
					},
					Fields: map[string]interface{}{
						"duration_ns": (time.Duration(103680) * time.Microsecond).Nanoseconds(),
					},
					Time: time.Unix(0, 1498688360851318000).UTC(),
				},
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"parent_id":     "22964302721410078",
						"trace_id":      "2505404965370368069",
						"name":          "Parent",
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
				testutil.Metric{
					Measurement: "zipkin",
					Tags: map[string]string{
						"trace_id":       "2505404965370368069",
						"service_name":   "trivial",
						"annotation":     "dHJpdmlhbA==",
						"annotation_key": "lc",
						"id":             "22964302721410078",
						"parent_id":      "22964302721410078",
						"name":           "Parent",
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
				t: Trace{
					Span{
						ID:          "6802735349851856000",
						TraceID:     "0:6802735349851856000",
						Name:        "main.dud",
						ParentID:    "6802735349851856000",
						Timestamp:   time.Unix(1, 0).UTC(),
						Duration:    1,
						ServiceName: "trivial",
						Annotations: []Annotation{
							Annotation{
								Timestamp:   time.Unix(0, 1433330263415871000).UTC(),
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
				testutil.Metric{
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

func Test_microToTime(t *testing.T) {
	type args struct {
		micro int64
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "given zero micro seconds expected unix time zero",
			args: args{
				micro: 0,
			},
			want: time.Unix(0, 0).UTC(),
		},
		{
			name: "given a million micro seconds expected unix time one",
			args: args{
				micro: 1000000,
			},
			want: time.Unix(1, 0).UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := microToTime(tt.args.micro); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("microToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newAnnotation(micro int64) *zipkincore.Annotation {
	return &zipkincore.Annotation{
		Timestamp: micro,
	}
}

func Test_minMax(t *testing.T) {
	type args struct {
		span *zipkincore.Span
	}
	tests := []struct {
		name    string
		args    args
		now     func() time.Time
		wantMin time.Time
		wantMax time.Time
	}{
		{
			name: "Single annotation",
			args: args{
				span: &zipkincore.Span{
					Annotations: []*zipkincore.Annotation{
						newAnnotation(1000000),
					},
				},
			},
			wantMin: time.Unix(1, 0).UTC(),
			wantMax: time.Unix(1, 0).UTC(),
		},
		{
			name: "Three annotations",
			args: args{
				span: &zipkincore.Span{
					Annotations: []*zipkincore.Annotation{
						newAnnotation(1000000),
						newAnnotation(2000000),
						newAnnotation(3000000),
					},
				},
			},
			wantMin: time.Unix(1, 0).UTC(),
			wantMax: time.Unix(3, 0).UTC(),
		},
		{
			name: "Annotations are in the future",
			args: args{
				span: &zipkincore.Span{
					Annotations: []*zipkincore.Annotation{
						newAnnotation(3000000),
					},
				},
			},
			wantMin: time.Unix(2, 0).UTC(),
			wantMax: time.Unix(3, 0).UTC(),
			now: func() time.Time {
				return time.Unix(2, 0).UTC()
			},
		},
		{
			name: "No Annotations",
			args: args{
				span: &zipkincore.Span{
					Annotations: []*zipkincore.Annotation{},
				},
			},
			wantMin: time.Unix(2, 0).UTC(),
			wantMax: time.Unix(2, 0).UTC(),
			now: func() time.Time {
				return time.Unix(2, 0).UTC()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.now != nil {
				now = tt.now
			}
			got, got1 := minMax(tt.args.span)
			if !reflect.DeepEqual(got, tt.wantMin) {
				t.Errorf("minMax() got = %v, want %v", got, tt.wantMin)
			}
			if !reflect.DeepEqual(got1, tt.wantMax) {
				t.Errorf("minMax() got1 = %v, want %v", got1, tt.wantMax)
			}
			now = time.Now
		})
	}
}

func Test_host(t *testing.T) {
	type args struct {
		h *zipkincore.Endpoint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Host Found",
			args: args{
				h: &zipkincore.Endpoint{
					Ipv4: 1234,
					Port: 8888,
				},
			},
			want: "0.0.4.210:8888",
		},
		{
			name: "No Host",
			args: args{
				h: nil,
			},
			want: "0.0.0.0",
		},
		{
			name: "int overflow zipkin uses an int16 type as an unsigned int 16.",
			args: args{
				h: &zipkincore.Endpoint{
					Ipv4: 1234,
					Port: -1,
				},
			},
			want: "0.0.4.210:65535",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := host(tt.args.h); got != tt.want {
				t.Errorf("host() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceName(t *testing.T) {
	type args struct {
		h *zipkincore.Endpoint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Found ServiceName",
			args: args{
				h: &zipkincore.Endpoint{
					ServiceName: "zipkin",
				},
			},
			want: "zipkin",
		},
		{
			name: "No ServiceName",
			args: args{
				h: nil,
			},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serviceName(tt.args.h); got != tt.want {
				t.Errorf("serviceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
