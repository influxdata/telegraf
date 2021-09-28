package thrift

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec/thrift/gen-go/zipkincore"
)

func Test_endpointHost(t *testing.T) {
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
			e := endpoint{tt.args.h}
			if got := e.Host(); got != tt.want {
				t.Errorf("host() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_endpointName(t *testing.T) {
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
			e := endpoint{tt.args.h}
			if got := e.Name(); got != tt.want {
				t.Errorf("serviceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalThrift(t *testing.T) {
	addr := func(i int64) *int64 { return &i }
	tests := []struct {
		name     string
		filename string
		want     []*zipkincore.Span
		wantErr  bool
	}{
		{
			name:     "threespans",
			filename: "../../testdata/threespans.dat",
			want: []*zipkincore.Span{
				{
					TraceID:     2505404965370368069,
					Name:        "Child",
					ID:          8090652509916334619,
					ParentID:    addr(22964302721410078),
					Timestamp:   addr(1498688360851331),
					Duration:    addr(53106),
					Annotations: []*zipkincore.Annotation{},
					BinaryAnnotations: []*zipkincore.BinaryAnnotation{
						{
							Key:            "lc",
							AnnotationType: zipkincore.AnnotationType_STRING,
							Value:          []byte("trivial"),
							Host: &zipkincore.Endpoint{
								Ipv4:        2130706433,
								ServiceName: "trivial",
							},
						},
					},
				},
				{
					TraceID:     2505404965370368069,
					Name:        "Child",
					ID:          103618986556047333,
					ParentID:    addr(22964302721410078),
					Timestamp:   addr(1498688360904552),
					Duration:    addr(50410),
					Annotations: []*zipkincore.Annotation{},
					BinaryAnnotations: []*zipkincore.BinaryAnnotation{
						{
							Key:            "lc",
							AnnotationType: zipkincore.AnnotationType_STRING,
							Value:          []byte("trivial"),
							Host: &zipkincore.Endpoint{
								Ipv4:        2130706433,
								ServiceName: "trivial",
							},
						},
					},
				},
				{
					TraceID:   2505404965370368069,
					Name:      "Parent",
					ID:        22964302721410078,
					Timestamp: addr(1498688360851318),
					Duration:  addr(103680),
					Annotations: []*zipkincore.Annotation{
						{
							Timestamp: 1498688360851325,
							Value:     "Starting child #0",
							Host: &zipkincore.Endpoint{
								Ipv4:        2130706433,
								ServiceName: "trivial",
							},
						},
						{
							Timestamp: 1498688360904545,
							Value:     "Starting child #1",
							Host: &zipkincore.Endpoint{
								Ipv4:        2130706433,
								ServiceName: "trivial",
							},
						},
						{
							Timestamp: 1498688360954992,
							Value:     "A Log",
							Host: &zipkincore.Endpoint{
								Ipv4:        2130706433,
								ServiceName: "trivial",
							},
						},
					},
					BinaryAnnotations: []*zipkincore.BinaryAnnotation{
						{
							Key:            "lc",
							AnnotationType: zipkincore.AnnotationType_STRING,
							Value:          []byte("trivial"),
							Host: &zipkincore.Endpoint{
								Ipv4:        2130706433,
								ServiceName: "trivial",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("Could not find file %s\n", tt.filename)
			}

			got, err := UnmarshalThrift(dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalThrift() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("UnmarshalThrift() got(-)/want(+): %s", cmp.Diff(tt.want, got))
			}
		})
	}
}
