//go:build windows
// +build windows

//revive:disable-next-line:var-naming
// Package win_eventlog Input plugin to collect Windows Event Log messages
package win_eventlog

import (
	"testing"
)

func TestWinEventLog_shouldExcludeEmptyField(t *testing.T) {
	type args struct {
		field      string
		fieldType  string
		fieldValue interface{}
	}
	tests := []struct {
		name       string
		w          *WinEventLog
		args       args
		wantShould bool
	}{
		{
			name:       "Not in list",
			args:       args{field: "qq", fieldType: "string", fieldValue: ""},
			wantShould: false,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:       "Empty string",
			args:       args{field: "test", fieldType: "string", fieldValue: ""},
			wantShould: true,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:       "Non-empty string",
			args:       args{field: "test", fieldType: "string", fieldValue: "qq"},
			wantShould: false,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:       "Zero int",
			args:       args{field: "test", fieldType: "int", fieldValue: int(0)},
			wantShould: true,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:       "Non-zero int",
			args:       args{field: "test", fieldType: "int", fieldValue: int(-1)},
			wantShould: false,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:       "Zero uint32",
			args:       args{field: "test", fieldType: "uint32", fieldValue: uint32(0)},
			wantShould: true,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:       "Non-zero uint32",
			args:       args{field: "test", fieldType: "uint32", fieldValue: uint32(0xc0fefeed)},
			wantShould: false,
			w:          &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotShould := tt.w.shouldExcludeEmptyField(tt.args.field, tt.args.fieldType, tt.args.fieldValue); gotShould != tt.wantShould {
				t.Errorf("WinEventLog.shouldExcludeEmptyField() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestWinEventLog_shouldProcessField(t *testing.T) {
	tags := []string{"Source", "Level*"}
	fields := []string{"EventID", "Message*"}
	excluded := []string{"Message*"}
	type args struct {
		field string
	}
	tests := []struct {
		name       string
		w          *WinEventLog
		args       args
		wantShould bool
		wantList   string
	}{
		{
			name:       "Not in tags",
			args:       args{field: "test"},
			wantShould: false,
			wantList:   "excluded",
			w:          &WinEventLog{EventTags: tags, EventFields: fields, ExcludeFields: excluded},
		},
		{
			name:       "In Tags",
			args:       args{field: "LevelText"},
			wantShould: true,
			wantList:   "tags",
			w:          &WinEventLog{EventTags: tags, EventFields: fields, ExcludeFields: excluded},
		},
		{
			name:       "Not in Fields",
			args:       args{field: "EventId"},
			wantShould: false,
			wantList:   "excluded",
			w:          &WinEventLog{EventTags: tags, EventFields: fields, ExcludeFields: excluded},
		},
		{
			name:       "In Fields",
			args:       args{field: "EventID"},
			wantShould: true,
			wantList:   "fields",
			w:          &WinEventLog{EventTags: tags, EventFields: fields, ExcludeFields: excluded},
		},
		{
			name:       "In Fields and Excluded",
			args:       args{field: "Messages"},
			wantShould: false,
			wantList:   "excluded",
			w:          &WinEventLog{EventTags: tags, EventFields: fields, ExcludeFields: excluded},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShould, gotList := tt.w.shouldProcessField(tt.args.field)
			if gotShould != tt.wantShould {
				t.Errorf("WinEventLog.shouldProcessField() gotShould = %v, want %v", gotShould, tt.wantShould)
			}
			if gotList != tt.wantList {
				t.Errorf("WinEventLog.shouldProcessField() gotList = %v, want %v", gotList, tt.wantList)
			}
		})
	}
}
