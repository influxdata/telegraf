//go:build windows

package win_eventlog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestSetStateBookmark(t *testing.T) {
	tests := []struct {
		name         string
		state        persistedState
		expectedFlag evtSubscribeFlag
		expectedErr  string
	}{
		{
			name:         "no state",
			state:        persistedState{},
			expectedFlag: evtSubscribeToFutureEvents,
		},
		{
			// What earlier Telegraf versions persist for a run that received no event
			name:         "bookmark without position",
			state:        persistedState{Bookmark: "<BookmarkList>\r\n</BookmarkList>"},
			expectedFlag: evtSubscribeToFutureEvents,
		},
		{
			name:         "empty bookmark list",
			state:        persistedState{Bookmark: "<BookmarkList/>"},
			expectedFlag: evtSubscribeToFutureEvents,
		},
		{
			name:         "bookmark without position at the channel start",
			state:        persistedState{Bookmark: "<BookmarkList>\r\n</BookmarkList>", AtChannelStart: true},
			expectedFlag: evtSubscribeStartAtOldestRecord,
		},
		{
			name: "bookmark with position",
			state: persistedState{
				Bookmark: "<BookmarkList>\r\n  <Bookmark Channel='Application' RecordId='1' IsCurrent='true'/>\r\n</BookmarkList>",
			},
			expectedFlag: evtSubscribeStartAfterBookmark,
		},
		{
			name:        "malformed bookmark",
			state:       persistedState{Bookmark: "<BookmarkList>"},
			expectedErr: "unmarshalling bookmark failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &WinEventLog{EventlogName: "Application", Log: testutil.Logger{}}
			require.NoError(t, plugin.Init())
			require.Equal(t, evtSubscribeToFutureEvents, plugin.subscriptionFlag)

			if tt.expectedErr != "" {
				require.ErrorContains(t, plugin.SetState(tt.state), tt.expectedErr)
				return
			}
			require.NoError(t, plugin.SetState(tt.state))
			require.Equal(t, tt.expectedFlag, plugin.subscriptionFlag)
		})
	}
}

func TestSetStateBookmarkFromBeginning(t *testing.T) {
	plugin := &WinEventLog{EventlogName: "Application", FromBeginning: true, Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())
	require.Equal(t, evtSubscribeStartAtOldestRecord, plugin.subscriptionFlag)

	require.NoError(t, plugin.SetState(persistedState{Bookmark: "<BookmarkList>\r\n</BookmarkList>"}))
	require.Equal(t, evtSubscribeStartAtOldestRecord, plugin.subscriptionFlag)
}

func TestSetStateInvalidType(t *testing.T) {
	plugin := &WinEventLog{EventlogName: "Application", Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	require.ErrorContains(t, plugin.SetState("<BookmarkList/>"), "invalid type string for state")
}

func TestStateUnmarshal(t *testing.T) {
	tests := []struct {
		name       string
		serialized string
		expected   persistedState
	}{
		{
			// What earlier Telegraf versions write
			name:       "plain bookmark",
			serialized: `"<BookmarkList/>"`,
			expected:   persistedState{Bookmark: "<BookmarkList/>"},
		},
		{
			name:       "bookmark and marker",
			serialized: `{"bookmark":"<BookmarkList/>","at_channel_start":true}`,
			expected:   persistedState{Bookmark: "<BookmarkList/>", AtChannelStart: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var state persistedState
			require.NoError(t, json.Unmarshal([]byte(tt.serialized), &state))
			require.Equal(t, tt.expected, state)
		})
	}
}

func TestStartAnchorsBookmark(t *testing.T) {
	plugin := &WinEventLog{EventlogName: "Application", Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Equal(t, evtSubscribeStartAfterBookmark, plugin.subscriptionFlag)

	state, ok := plugin.GetState().(persistedState)
	require.True(t, ok)
	require.Contains(t, state.Bookmark, "<Bookmark ")
	require.False(t, state.AtChannelStart)
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

	state, ok := plugin.GetState().(persistedState)
	require.True(t, ok)
	require.NotContains(t, state.Bookmark, "<Bookmark ")
	require.True(t, state.AtChannelStart)

	// Restoring that state continues where the quiet run left, i.e. at the beginning of the channel
	restarted := &WinEventLog{EventlogName: "Application", Query: plugin.Query, Log: testutil.Logger{}}
	require.NoError(t, restarted.Init())
	require.NoError(t, restarted.SetState(state))
	require.Equal(t, evtSubscribeStartAtOldestRecord, restarted.subscriptionFlag)
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
