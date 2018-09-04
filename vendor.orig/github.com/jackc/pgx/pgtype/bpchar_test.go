package pgtype_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestChar3Transcode(t *testing.T) {
	testutil.TestSuccessfulTranscodeEqFunc(t, "char(3)", []interface{}{
		&pgtype.BPChar{String: "a  ", Status: pgtype.Present},
		&pgtype.BPChar{String: " a ", Status: pgtype.Present},
		&pgtype.BPChar{String: "嗨  ", Status: pgtype.Present},
		&pgtype.BPChar{String: "   ", Status: pgtype.Present},
		&pgtype.BPChar{Status: pgtype.Null},
	}, func(aa, bb interface{}) bool {
		a := aa.(pgtype.BPChar)
		b := bb.(pgtype.BPChar)

		return a.Status == b.Status && a.String == b.String
	})
}

func TestBPCharAssignTo(t *testing.T) {
	var (
		str string
		run rune
	)
	simpleTests := []struct {
		src      pgtype.BPChar
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.BPChar{String: "simple", Status: pgtype.Present}, dst: &str, expected: "simple"},
		{src: pgtype.BPChar{String: "嗨", Status: pgtype.Present}, dst: &run, expected: '嗨'},
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

}
