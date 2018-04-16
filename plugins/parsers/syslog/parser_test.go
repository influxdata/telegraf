package syslog

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/go-syslog/rfc5424"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		name string
		opts []ParserOpt
		want string
	}{
		{
			name: "name should be syslog if default options",
			want: "syslog",
		},
		{
			name: "name should use the option",
			opts: []ParserOpt{
				WithName("name1"),
			},
			want: "name1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.opts...)
			if got := p.Name; got != tt.want {
				t.Errorf("NewParser() = %v, want %v", got, tt.want)
			}
		})
	}
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
				Priority: uintPtr(0),
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
				Priority: uintPtr(14),
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
			tt.msg.SetPriority(*tt.msg.Priority)
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
				"id1":     true,
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

func TestParser_tm(t *testing.T) {
	tests := []struct {
		name string
		now  func() time.Time
		msg  *rfc5424.SyslogMessage
		want time.Time
	}{
		{
			name: "should return now if time not in message",
			now: func() time.Time {
				return time.Time{}
			},
			msg:  &rfc5424.SyslogMessage{},
			want: time.Time{},
		},
		{
			name: "should return message time if set",
			now: func() time.Time {
				return time.Time{}
			},
			msg: &rfc5424.SyslogMessage{
				Timestamp: timePtr(time.Date(1997, 4, 1, 7, 30, 0, 0, time.UTC)),
			},
			want: time.Date(1997, 4, 1, 7, 30, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Parser{
				now: tt.now,
			}
			if got := s.tm(tt.msg); !cmp.Equal(tt.want, got) {
				t.Errorf("Parser.tm() = got(-)/want(+) %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestParser_ParseLine(t *testing.T) {
	type fields struct {
		DefaultTags map[string]string
		Name        string
		p           *rfc5424.Parser
		now         func() time.Time
	}
	type metric struct {
		tm     time.Time
		name   string
		tags   map[string]string
		fields map[string]interface{}
	}
	tests := []struct {
		name    string
		line    string
		fields  fields
		want    metric
		wantErr bool
	}{
		{
			name: "should parse syslog and generate telegraf Metric",
			line: "<78>1 2016-01-15T00:04:01+00:00 host1 CROND 10391 - [sdid] some_message",
			fields: fields{
				DefaultTags: map[string]string{
					"t1": "v1",
				},
				Name: "measurement1",
				p:    rfc5424.NewParser(),
				now:  func() time.Time { return time.Time{} },
			},
			want: metric{
				tm:   timeParse(time.RFC3339Nano, "2016-01-15T00:04:01+00:00"),
				name: "measurement1",
				tags: map[string]string{
					"facility": "clock daemon",
					"severity": "informational",
					"hostname": "host1",
					"appname":  "CROND",
					"t1":       "v1",
				},
				fields: map[string]interface{}{
					"version": int64(1),
					"sdid":    true,
					"procid":  "10391",
					"message": "some_message",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Parser{
				DefaultTags: tt.fields.DefaultTags,
				Name:        tt.fields.Name,
				p:           tt.fields.p,
				now:         tt.fields.now,
			}
			mt, err := s.ParseLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ParseLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got, want := mt.Name(), tt.want.name; got != want {
				t.Errorf("Parser.ParseLine().name = %v, want %v", got, want)
			}
			if got, want := mt.Time().String(), tt.want.tm.String(); got != want {
				t.Errorf("Parser.ParseLine().tm =%v, want %v", got, want)
			}
			if got, want := mt.Tags(), tt.want.tags; !cmp.Equal(want, got) {
				t.Errorf("Parser.ParseLine().tags = got(-)/want(+) %s", cmp.Diff(want, got))
			}
			if got, want := mt.Fields(), tt.want.fields; !cmp.Equal(want, got) {
				t.Errorf("Parser.ParseLine().fields = got(-)/want(+) %s", cmp.Diff(want, got))
			}
		})
	}
}

func uintPtr(i uint8) *uint8 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func timePtr(tm time.Time) *time.Time {
	return &tm
}

func timeParse(layout, value string) time.Time {
	t, _ := time.Parse(layout, value)
	return t
}
