//go:build windows

package win_eventlog

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestStartAnchorsBookmark(t *testing.T) {
	plugin := &WinEventLog{EventlogName: "Application", Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Equal(t, evtSubscribeStartAfterBookmark, plugin.subscriptionFlag)

	state, ok := plugin.GetState().(string)
	require.True(t, ok)
	require.Contains(t, state, "<Bookmark ")
}

func TestStartWithoutMatchingEvent(t *testing.T) {
	plugin := &WinEventLog{
		EventlogName: "Application",
		Query:        "*[System[(EventID=999999)]]",
		Log:          testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Equal(t, evtSubscribeToFutureEvents, plugin.subscriptionFlag)

	// The bookmark holds no position, which is what makes the next run continue at the channel start
	state, ok := plugin.GetState().(string)
	require.True(t, ok)
	require.NotContains(t, state, "<Bookmark ")

	restarted := &WinEventLog{EventlogName: "Application", Query: plugin.Query, Log: testutil.Logger{}}
	require.NoError(t, restarted.Init())
	require.NoError(t, restarted.SetState(state))
	require.Equal(t, evtSubscribeStartAfterBookmark, restarted.subscriptionFlag)

	// Subscribing after a positionless bookmark must work, as this is how
	// the next run continues at the channel start
	require.NoError(t, restarted.Start(&acc))
	restarted.Stop()
}

func TestWinEventLog_shouldExcludeEmptyField(t *testing.T) {
	type args struct {
		field      string
		fieldType  string
		fieldValue interface{}
	}
	tests := []struct {
		name     string
		w        *WinEventLog
		args     args
		expected bool
	}{
		{
			name:     "Not in list",
			args:     args{field: "qq", fieldType: "string", fieldValue: ""},
			expected: false,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:     "Empty string",
			args:     args{field: "test", fieldType: "string", fieldValue: ""},
			expected: true,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:     "Non-empty string",
			args:     args{field: "test", fieldType: "string", fieldValue: "qq"},
			expected: false,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:     "Zero int",
			args:     args{field: "test", fieldType: "int", fieldValue: int(0)},
			expected: true,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:     "Non-zero int",
			args:     args{field: "test", fieldType: "int", fieldValue: int(-1)},
			expected: false,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:     "Zero uint32",
			args:     args{field: "test", fieldType: "uint32", fieldValue: uint32(0)},
			expected: true,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
		{
			name:     "Non-zero uint32",
			args:     args{field: "test", fieldType: "uint32", fieldValue: uint32(0xc0fefeed)},
			expected: false,
			w:        &WinEventLog{ExcludeEmpty: []string{"te*"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.w.Init())
			actual := tt.w.shouldExcludeEmptyField(tt.args.field, tt.args.fieldType, tt.args.fieldValue)
			require.Equal(t, tt.expected, actual)
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
			require.NoError(t, tt.w.Init())
			should, list := tt.w.shouldProcessField(tt.args.field)
			require.Equal(t, tt.wantShould, should)
			require.Equal(t, tt.wantList, list)
		})
	}
}
