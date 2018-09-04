package pgtype_test

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestMacaddrTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "macaddr", []interface{}{
		&pgtype.Macaddr{Addr: mustParseMacaddr(t, "01:23:45:67:89:ab"), Status: pgtype.Present},
		&pgtype.Macaddr{Status: pgtype.Null},
	})
}

func TestMacaddrSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Macaddr
	}{
		{
			source: mustParseMacaddr(t, "01:23:45:67:89:ab"),
			result: pgtype.Macaddr{Addr: mustParseMacaddr(t, "01:23:45:67:89:ab"), Status: pgtype.Present},
		},
		{
			source: "01:23:45:67:89:ab",
			result: pgtype.Macaddr{Addr: mustParseMacaddr(t, "01:23:45:67:89:ab"), Status: pgtype.Present},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.Macaddr
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestMacaddrAssignTo(t *testing.T) {
	{
		src := pgtype.Macaddr{Addr: mustParseMacaddr(t, "01:23:45:67:89:ab"), Status: pgtype.Present}
		var dst net.HardwareAddr
		expected := mustParseMacaddr(t, "01:23:45:67:89:ab")

		err := src.AssignTo(&dst)
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare([]byte(dst), []byte(expected)) != 0 {
			t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
		}
	}

	{
		src := pgtype.Macaddr{Addr: mustParseMacaddr(t, "01:23:45:67:89:ab"), Status: pgtype.Present}
		var dst string
		expected := "01:23:45:67:89:ab"

		err := src.AssignTo(&dst)
		if err != nil {
			t.Error(err)
		}

		if dst != expected {
			t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
		}
	}
}
