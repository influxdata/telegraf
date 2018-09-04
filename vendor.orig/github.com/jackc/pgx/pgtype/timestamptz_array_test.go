package pgtype_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestTimestamptzArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscodeEqFunc(t, "timestamptz[]", []interface{}{
		&pgtype.TimestamptzArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.TimestamptzArray{
			Elements: []pgtype.Timestamptz{
				{Time: time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.TimestamptzArray{Status: pgtype.Null},
		&pgtype.TimestamptzArray{
			Elements: []pgtype.Timestamptz{
				{Time: time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Time: time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Time: time.Date(2017, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Time: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Status: pgtype.Null},
				{Time: time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.TimestamptzArray{
			Elements: []pgtype.Timestamptz{
				{Time: time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Time: time.Date(2015, 2, 2, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Time: time.Date(2015, 2, 3, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
				{Time: time.Date(2015, 2, 4, 0, 0, 0, 0, time.UTC), Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	}, func(a, b interface{}) bool {
		ata := a.(pgtype.TimestamptzArray)
		bta := b.(pgtype.TimestamptzArray)

		if len(ata.Elements) != len(bta.Elements) || ata.Status != bta.Status {
			return false
		}

		for i := range ata.Elements {
			ae, be := ata.Elements[i], bta.Elements[i]
			if !(ae.Time.Equal(be.Time) && ae.Status == be.Status && ae.InfinityModifier == be.InfinityModifier) {
				return false
			}
		}

		return true
	})
}

func TestTimestamptzArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.TimestamptzArray
	}{
		{
			source: []time.Time{time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC)},
			result: pgtype.TimestamptzArray{
				Elements:   []pgtype.Timestamptz{{Time: time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]time.Time)(nil)),
			result: pgtype.TimestamptzArray{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.TimestamptzArray
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestTimestamptzArrayAssignTo(t *testing.T) {
	var timeSlice []time.Time

	simpleTests := []struct {
		src      pgtype.TimestamptzArray
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.TimestamptzArray{
				Elements:   []pgtype.Timestamptz{{Time: time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &timeSlice,
			expected: []time.Time{time.Date(2015, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			src:      pgtype.TimestamptzArray{Status: pgtype.Null},
			dst:      &timeSlice,
			expected: (([]time.Time)(nil)),
		},
	}

	for i, tt := range simpleTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Interface(); !reflect.DeepEqual(dst, tt.expected) {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	errorTests := []struct {
		src pgtype.TimestamptzArray
		dst interface{}
	}{
		{
			src: pgtype.TimestamptzArray{
				Elements:   []pgtype.Timestamptz{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst: &timeSlice,
		},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}

}
