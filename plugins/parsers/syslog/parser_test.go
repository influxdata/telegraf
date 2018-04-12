package syslog

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/go-syslog/rfc5424"
)

func strPtr(s string) *string {
	return &s
}

func TestParser_tags(t *testing.T) {
	tests := []struct {
		name        string
		DefaultTags map[string]string
		msg         *rfc5424.SyslogMessage
		want        map[string]string
	}{
		{
			name: "default tags should be added to the message",
			DefaultTags: map[string]string{
				"host": "localhost",
			},
			msg: &rfc5424.SyslogMessage{
				Priority: 0,
			},
			want: map[string]string{
				"facility": "kernel messages",
				"severity": "emergency",
				"host":     "localhost",
			},
		},
		{
			name: "hostname/appname should be tags",
			msg: &rfc5424.SyslogMessage{
				Priority: 14,
				Hostname: strPtr("scylla.eng.utah.edu"),
				Appname:  strPtr("x11"),
			},
			want: map[string]string{
				"facility": "user-level messages",
				"severity": "informational",
				"hostname": "scylla.eng.utah.edu",
				"appname":  "x11",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Parser{
				DefaultTags: tt.DefaultTags,
			}
			tt.msg.SetPriority(tt.msg.Priority)
			if got := s.tags(tt.msg); !cmp.Equal(tt.want, got) {
				t.Errorf("Parser.tags() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestParser_fields(t *testing.T) {
	tests := []struct {
		name string
		msg  *rfc5424.SyslogMessage
		want map[string]interface{}
	}{
		{
			name: "fields should at least have a version",
			msg:  &rfc5424.SyslogMessage{},
			want: map[string]interface{}{
				"version": 0,
			},
		},
		{
			name: "messages with procid/msgid/message should be fields",
			msg: &rfc5424.SyslogMessage{
				Version: 1,
				ProcID:  strPtr("1"),
				MsgID:   strPtr("1"),
				Message: strPtr("log message here"),
			},
			want: map[string]interface{}{
				"version": 1,
				"procid":  "1",
				"msgid":   "1",
				"message": "log message here",
			},
		},
		{
			name: "messages with structured data should be fields with concatenated names",
			msg: &rfc5424.SyslogMessage{
				Version: 1,
				StructuredData: &map[string]map[string]string{
					"id1": map[string]string{
						"name": "value",
					},
					"id2": map[string]string{
						"name": "value",
					},
				},
			},
			want: map[string]interface{}{
				"version":  1,
				"id1 name": "value",
				"id2 name": "value",
			},
		},
		{
			name: "messages with structured data without params should be bool",
			msg: &rfc5424.SyslogMessage{
				Version: 1,
				StructuredData: &map[string]map[string]string{
					"id1": map[string]string{},
				},
			},
			want: map[string]interface{}{
				"version": 1,
				"id1":     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Parser{}
			if got := s.fields(tt.msg); !cmp.Equal(tt.want, got) {
				t.Errorf("Parser.fields() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}
