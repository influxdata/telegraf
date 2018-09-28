package jsonV1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec"
)

func TestJSON_Decode(t *testing.T) {
	addr := func(i int64) *int64 { return &i }
	tests := []struct {
		name    string
		octets  []byte
		want    []codec.Span
		wantErr bool
	}{
		{
			name: "bad json is error",
			octets: []byte(`
			[
				{
			]`),
			wantErr: true,
		},
		{
			name: "Decodes simple trace",
			octets: []byte(`
			[
				{
				  "traceId": "6b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c"
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "6b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
				},
			},
		},
		{
			name: "Decodes two spans",
			octets: []byte(`
			[
				{
				  "traceId": "6b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c"
				},
				{
					"traceId": "6b221d5bc9e6496c",
					"name": "get-traces",
					"id": "c6946e9cb5d122b6",
					"parentId": "6b221d5bc9e6496c",
					"duration": 10000
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "6b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
				},
				&span{
					TraceID:  "6b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "c6946e9cb5d122b6",
					ParentID: "6b221d5bc9e6496c",
					Dur:      addr(10000),
				},
			},
		},
		{
			name: "Decodes trace with timestamp",
			octets: []byte(`
			[
				{
				  "traceId": "6b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "timestamp": 1503031538791000
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "6b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					Time:     addr(1503031538791000),
				},
			},
		},
		{
			name: "Decodes simple trace with high and low trace id",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c"
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
				},
			},
		},
		{
			name: "Error when trace id is null",
			octets: []byte(`
			[
				{
				  "traceId": null,
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c"
				}
			]`),
			wantErr: true,
		},
		{
			name: "ignore null parentId",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "parentId": null
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
				},
			},
		},
		{
			name: "ignore null timestamp",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "timestamp": null
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
				},
			},
		},
		{
			name: "ignore null duration",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "duration": null
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
				},
			},
		},
		{
			name: "ignore null annotation endpoint",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "annotations": [
						{
							"timestamp": 1461750491274000,
							"value": "cs",
							"endpoint": null
						}
					]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					Anno: []annotation{
						{
							Time: 1461750491274000,
							Val:  "cs",
						},
					},
				},
			},
		},
		{
			name: "ignore null binary annotation endpoint",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "lc",
							"value": "JDBCSpanStore",
							"endpoint": null
						}
				  ]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					BAnno: []binaryAnnotation{
						{
							K: "lc",
							V: json.RawMessage(`"JDBCSpanStore"`),
						},
					},
				},
			},
		},
		{
			name: "Error when binary annotation has no key",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"value": "JDBCSpanStore",
							"endpoint": null
						}
				  ]
				}
			]`),
			wantErr: true,
		},
		{
			name: "Error when binary annotation has no value",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "lc",
							"endpoint": null
						}
				  ]
				}
			]`),
			wantErr: true,
		},
		{
			name: "binary annotation with endpoint",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "lc",
							"value": "JDBCSpanStore",
							"endpoint": {
								"serviceName": "service",
								"port": 65535
							}
						}
					]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					BAnno: []binaryAnnotation{
						{
							K: "lc",
							V: json.RawMessage(`"JDBCSpanStore"`),
							Endpoint: &endpoint{
								ServiceName: "service",
								Port:        65535,
							},
						},
					},
				},
			},
		},
		{
			name: "binary annotation with double value",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "num",
							"value": 1.23456789,
							"type": "DOUBLE"
						}
					]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					BAnno: []binaryAnnotation{
						{
							K:    "num",
							V:    json.RawMessage{0x31, 0x2e, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39},
							Type: "DOUBLE",
						},
					},
				},
			},
		},
		{
			name: "binary annotation with integer value",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "num",
							"value": 1,
							"type": "I16"
						}
					]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					BAnno: []binaryAnnotation{
						{
							K:    "num",
							V:    json.RawMessage{0x31},
							Type: "I16",
						},
					},
				},
			},
		},
		{
			name: "binary annotation with bool value",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "num",
							"value": true,
							"type": "BOOL"
						}
					]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					BAnno: []binaryAnnotation{
						{
							K:    "num",
							V:    json.RawMessage(`true`),
							Type: "BOOL",
						},
					},
				},
			},
		},
		{
			name: "binary annotation with bytes value",
			octets: []byte(`
			[
				{
				  "traceId": "48485a3953bb61246b221d5bc9e6496c",
				  "name": "get-traces",
				  "id": "6b221d5bc9e6496c",
				  "binaryAnnotations": [
						{
							"key": "num",
							"value": "1",
							"type": "BYTES"
						}
					]
				}
			]`),
			want: []codec.Span{
				&span{
					TraceID:  "48485a3953bb61246b221d5bc9e6496c",
					SpanName: "get-traces",
					ID:       "6b221d5bc9e6496c",
					BAnno: []binaryAnnotation{
						{
							K:    "num",
							V:    json.RawMessage(`"1"`),
							Type: "BYTES",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{}
			got, err := j.Decode(tt.octets)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("JSON.Decode() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func Test_span_Trace(t *testing.T) {
	tests := []struct {
		name    string
		TraceID string
		want    string
		wantErr bool
	}{
		{
			name:    "Trace IDs cannot be null",
			TraceID: "",
			wantErr: true,
		},
		{
			name:    "converts hex string correctly",
			TraceID: "deadbeef",
			want:    "deadbeef",
		},
		{
			name:    "converts high and low trace id correctly",
			TraceID: "48485a3953bb61246b221d5bc9e6496c",
			want:    "48485a3953bb61246b221d5bc9e6496c",
		},
		{
			name:    "errors when string isn't hex",
			TraceID: "oxdeadbeef",
			wantErr: true,
		},
		{
			name:    "errors when id is too long",
			TraceID: "1234567890abcdef1234567890abcdef1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &span{
				TraceID: tt.TraceID,
			}
			got, err := s.Trace()
			if (err != nil) != tt.wantErr {
				t.Errorf("span.Trace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("span.Trace() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func Test_span_SpanID(t *testing.T) {
	tests := []struct {
		name    string
		ID      string
		want    string
		wantErr bool
	}{
		{
			name:    "Span IDs cannot be null",
			ID:      "",
			wantErr: true,
		},
		{
			name: "validates known id correctly",
			ID:   "b26412d1ac16767d",
			want: "b26412d1ac16767d",
		},
		{
			name: "validates hex string correctly",
			ID:   "deadbeef",
			want: "deadbeef",
		},
		{
			name:    "errors when string isn't hex",
			ID:      "oxdeadbeef",
			wantErr: true,
		},
		{
			name:    "errors when id is too long",
			ID:      "1234567890abcdef1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &span{
				ID: tt.ID,
			}
			got, err := s.SpanID()
			if (err != nil) != tt.wantErr {
				t.Errorf("span.SpanID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("span.SpanID() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func Test_span_Parent(t *testing.T) {
	tests := []struct {
		name     string
		ParentID string
		want     string
		wantErr  bool
	}{
		{
			name:     "when there is no parent return empty string",
			ParentID: "",
			want:     "",
		},
		{
			name:     "validates hex string correctly",
			ParentID: "deadbeef",
			want:     "deadbeef",
		},
		{
			name:     "errors when string isn't hex",
			ParentID: "oxdeadbeef",
			wantErr:  true,
		},
		{
			name:     "errors when parent id is too long",
			ParentID: "1234567890abcdef1",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &span{
				ParentID: tt.ParentID,
			}
			got, err := s.Parent()
			if (err != nil) != tt.wantErr {
				t.Errorf("span.Parent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("span.Parent() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func Test_span_Timestamp(t *testing.T) {
	tests := []struct {
		name string
		Time *int64
		want time.Time
	}{
		{
			name: "converts to microseconds",
			Time: func(i int64) *int64 { return &i }(3000000),
			want: time.Unix(3, 0).UTC(),
		},
		{
			name: "nil time should be zero time",
			Time: nil,
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &span{
				Time: tt.Time,
			}
			if got := s.Timestamp(); !cmp.Equal(tt.want, got) {
				t.Errorf("span.Timestamp() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func Test_span_Duration(t *testing.T) {
	tests := []struct {
		name string
		dur  *int64
		want time.Duration
	}{
		{
			name: "converts from 3 microseconds",
			dur:  func(i int64) *int64 { return &i }(3000000),
			want: 3 * time.Second,
		},
		{
			name: "nil time should be zero duration",
			dur:  nil,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &span{
				Dur: tt.dur,
			}
			if got := s.Duration(); got != tt.want {
				t.Errorf("span.Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_annotation(t *testing.T) {
	type fields struct {
		Endpoint *endpoint
		Time     int64
		Val      string
	}
	tests := []struct {
		name     string
		fields   fields
		tm       time.Time
		val      string
		endpoint *endpoint
	}{
		{
			name: "returns all fields",
			fields: fields{
				Time: 3000000,
				Val:  "myvalue",
				Endpoint: &endpoint{
					ServiceName: "myservice",
					Ipv4:        "127.0.0.1",
					Port:        443,
				},
			},
			tm:  time.Unix(3, 0).UTC(),
			val: "myvalue",
			endpoint: &endpoint{
				ServiceName: "myservice",
				Ipv4:        "127.0.0.1",
				Port:        443,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			an := annotation(tt.fields)
			a := &an
			if got := a.Timestamp(); got != tt.tm {
				t.Errorf("annotation.Timestamp() = %v, want %v", got, tt.tm)
			}
			if got := a.Value(); got != tt.val {
				t.Errorf("annotation.Value() = %v, want %v", got, tt.val)
			}
			if got := a.Host(); !cmp.Equal(tt.endpoint, got) {
				t.Errorf("annotation.Endpoint() = %v, want %v", got, tt.endpoint)
			}
		})
	}
}

func Test_binaryAnnotation(t *testing.T) {
	type fields struct {
		K        string
		V        json.RawMessage
		Type     string
		Endpoint *endpoint
	}
	tests := []struct {
		name     string
		fields   fields
		key      string
		value    string
		endpoint *endpoint
	}{
		{
			name: "returns all fields",
			fields: fields{
				K: "key",
				V: json.RawMessage(`"value"`),
				Endpoint: &endpoint{
					ServiceName: "myservice",
					Ipv4:        "127.0.0.1",
					Port:        443,
				},
			},
			key:   "key",
			value: "value",
			endpoint: &endpoint{
				ServiceName: "myservice",
				Ipv4:        "127.0.0.1",
				Port:        443,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bin := binaryAnnotation(tt.fields)
			b := &bin
			if got := b.Key(); got != tt.key {
				t.Errorf("binaryAnnotation.Key() = %v, want %v", got, tt.key)
			}
			if got := b.Value(); got != tt.value {
				t.Errorf("binaryAnnotation.Value() = %v, want %v", got, tt.value)
			}
			if got := b.Host(); !cmp.Equal(tt.endpoint, got) {
				t.Errorf("binaryAnnotation.Endpoint() = %v, want %v", got, tt.endpoint)
			}
		})
	}
}

func Test_endpoint_Host(t *testing.T) {
	type fields struct {
		Ipv4 string
		Port int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "with port",
			fields: fields{
				Ipv4: "127.0.0.1",
				Port: 443,
			},
			want: "127.0.0.1:443",
		},
		{
			name: "no port",
			fields: fields{
				Ipv4: "127.0.0.1",
			},
			want: "127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &endpoint{
				Ipv4: tt.fields.Ipv4,
				Port: tt.fields.Port,
			}
			if got := e.Host(); got != tt.want {
				t.Errorf("endpoint.Host() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_endpoint_Name(t *testing.T) {
	tests := []struct {
		name        string
		ServiceName string
		want        string
	}{
		{
			name:        "has service name",
			ServiceName: "myservicename",
			want:        "myservicename",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &endpoint{
				ServiceName: tt.ServiceName,
			}
			if got := e.Name(); got != tt.want {
				t.Errorf("endpoint.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTraceIDFromString(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{
			name: "Convert hex string id",
			s:    "6b221d5bc9e6496c",
			want: "6b221d5bc9e6496c",
		},
		{
			name:    "error : id too long",
			s:       "1234567890abcdef1234567890abcdef1",
			wantErr: true,
		},
		{
			name:    "error : not parsable",
			s:       "howdyhowdyhowdy",
			wantErr: true,
		},
		{
			name: "Convert hex string with high/low",
			s:    "48485a3953bb61246b221d5bc9e6496c",
			want: "48485a3953bb61246b221d5bc9e6496c",
		},
		{
			name:    "errors in high",
			s:       "ERR85a3953bb61246b221d5bc9e6496c",
			wantErr: true,
		},
		{
			name:    "errors in low",
			s:       "48485a3953bb61246b221d5bc9e64ERR",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TraceIDFromString(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("TraceIDFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TraceIDFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIDFromString(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{
			name: "validates hex string id",
			s:    "6b221d5bc9e6496c",
			want: "6b221d5bc9e6496c",
		},
		{
			name:    "error : id too long",
			s:       "1234567890abcdef1",
			wantErr: true,
		},
		{
			name:    "error : not parsable",
			s:       "howdyhowdyhowdy",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IDFromString(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IDFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
