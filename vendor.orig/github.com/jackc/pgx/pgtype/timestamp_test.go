package pgtype_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestTimestampTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscodeEqFunc(t, "timestamp", []interface{}{
		&pgtype.Timestamp{Time: time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(1905, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Time: time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
		&pgtype.Timestamp{Status: pgtype.Null},
		&pgtype.Timestamp{Status: pgtype.Present, InfinityModifier: pgtype.Infinity},
		&pgtype.Timestamp{Status: pgtype.Present, InfinityModifier: -pgtype.Infinity},
	}, func(a, b interface{}) bool {
		at := a.(pgtype.Timestamp)
		bt := b.(pgtype.Timestamp)

		return at.Time.Equal(bt.Time) && at.Status == bt.Status && at.InfinityModifier == bt.InfinityModifier
	})
}

func TestTimestampSet(t *testing.T) {
	type _time time.Time

	successfulTests := []struct {
		source interface{}
		result pgtype.Timestamp
	}{
		{source: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), result: pgtype.Timestamp{Time: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
		{source: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), result: pgtype.Timestamp{Time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
		{source: time.Date(1999, 12, 31, 12, 59, 59, 0, time.UTC), result: pgtype.Timestamp{Time: time.Date(1999, 12, 31, 12, 59, 59, 0, time.UTC), Status: pgtype.Present}},
		{source: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), result: pgtype.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
		{source: time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC), result: pgtype.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC), Status: pgtype.Present}},
		{source: time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC), result: pgtype.Timestamp{Time: time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
		{source: time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local), result: pgtype.Timestamp{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
		{source: _time(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)), result: pgtype.Timestamp{Time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
	}

	for i, tt := range successfulTests {
		var r pgtype.Timestamp
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if r != tt.result {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestTimestampAssignTo(t *testing.T) {
	var tim time.Time
	var ptim *time.Time

	simpleTests := []struct {
		src      pgtype.Timestamp
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Timestamp{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}, dst: &tim, expected: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)},
		{src: pgtype.Timestamp{Time: time.Time{}, Status: pgtype.Null}, dst: &ptim, expected: ((*time.Time)(nil))},
	}

	for i, tt := range simpleTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Interface(); dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	pointerAllocTests := []struct {
		src      pgtype.Timestamp
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Timestamp{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local), Status: pgtype.Present}, dst: &ptim, expected: time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local)},
	}

	for i, tt := range pointerAllocTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Elem().Interface(); dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	errorTests := []struct {
		src pgtype.Timestamp
		dst interface{}
	}{
		{src: pgtype.Timestamp{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local), InfinityModifier: pgtype.Infinity, Status: pgtype.Present}, dst: &tim},
		{src: pgtype.Timestamp{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local), InfinityModifier: pgtype.NegativeInfinity, Status: pgtype.Present}, dst: &tim},
		{src: pgtype.Timestamp{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local), Status: pgtype.Null}, dst: &tim},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}
}
