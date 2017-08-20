package thrift

import (
	"testing"

	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
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
