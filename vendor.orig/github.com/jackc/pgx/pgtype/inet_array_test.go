package pgtype_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestInetArrayTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "inet[]", []interface{}{
		&pgtype.InetArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Present,
		},
		&pgtype.InetArray{
			Elements: []pgtype.Inet{
				{IPNet: mustParseCIDR(t, "12.34.56.0/32"), Status: pgtype.Present},
				{Status: pgtype.Null},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.InetArray{Status: pgtype.Null},
		&pgtype.InetArray{
			Elements: []pgtype.Inet{
				{IPNet: mustParseCIDR(t, "127.0.0.1/32"), Status: pgtype.Present},
				{IPNet: mustParseCIDR(t, "12.34.56.0/32"), Status: pgtype.Present},
				{IPNet: mustParseCIDR(t, "192.168.0.1/32"), Status: pgtype.Present},
				{IPNet: mustParseCIDR(t, "2607:f8b0:4009:80b::200e/128"), Status: pgtype.Present},
				{Status: pgtype.Null},
				{IPNet: mustParseCIDR(t, "255.0.0.0/8"), Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{{Length: 3, LowerBound: 1}, {Length: 2, LowerBound: 1}},
			Status:     pgtype.Present,
		},
		&pgtype.InetArray{
			Elements: []pgtype.Inet{
				{IPNet: mustParseCIDR(t, "127.0.0.1/32"), Status: pgtype.Present},
				{IPNet: mustParseCIDR(t, "12.34.56.0/32"), Status: pgtype.Present},
				{IPNet: mustParseCIDR(t, "192.168.0.1/32"), Status: pgtype.Present},
				{IPNet: mustParseCIDR(t, "2607:f8b0:4009:80b::200e/128"), Status: pgtype.Present},
			},
			Dimensions: []pgtype.ArrayDimension{
				{Length: 2, LowerBound: 4},
				{Length: 2, LowerBound: 2},
			},
			Status: pgtype.Present,
		},
	})
}

func TestInetArraySet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.InetArray
	}{
		{
			source: []*net.IPNet{mustParseCIDR(t, "127.0.0.1/32")},
			result: pgtype.InetArray{
				Elements:   []pgtype.Inet{{IPNet: mustParseCIDR(t, "127.0.0.1/32"), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]*net.IPNet)(nil)),
			result: pgtype.InetArray{Status: pgtype.Null},
		},
		{
			source: []net.IP{mustParseCIDR(t, "127.0.0.1/32").IP},
			result: pgtype.InetArray{
				Elements:   []pgtype.Inet{{IPNet: mustParseCIDR(t, "127.0.0.1/32"), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present},
		},
		{
			source: (([]net.IP)(nil)),
			result: pgtype.InetArray{Status: pgtype.Null},
		},
	}

	for i, tt := range successfulTests {
		var r pgtype.InetArray
		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(r, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestInetArrayAssignTo(t *testing.T) {
	var ipnetSlice []*net.IPNet
	var ipSlice []net.IP

	simpleTests := []struct {
		src      pgtype.InetArray
		dst      interface{}
		expected interface{}
	}{
		{
			src: pgtype.InetArray{
				Elements:   []pgtype.Inet{{IPNet: mustParseCIDR(t, "127.0.0.1/32"), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &ipnetSlice,
			expected: []*net.IPNet{mustParseCIDR(t, "127.0.0.1/32")},
		},
		{
			src: pgtype.InetArray{
				Elements:   []pgtype.Inet{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &ipnetSlice,
			expected: []*net.IPNet{nil},
		},
		{
			src: pgtype.InetArray{
				Elements:   []pgtype.Inet{{IPNet: mustParseCIDR(t, "127.0.0.1/32"), Status: pgtype.Present}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &ipSlice,
			expected: []net.IP{mustParseCIDR(t, "127.0.0.1/32").IP},
		},
		{
			src: pgtype.InetArray{
				Elements:   []pgtype.Inet{{Status: pgtype.Null}},
				Dimensions: []pgtype.ArrayDimension{{LowerBound: 1, Length: 1}},
				Status:     pgtype.Present,
			},
			dst:      &ipSlice,
			expected: []net.IP{nil},
		},
		{
			src:      pgtype.InetArray{Status: pgtype.Null},
			dst:      &ipnetSlice,
			expected: (([]*net.IPNet)(nil)),
		},
		{
			src:      pgtype.InetArray{Status: pgtype.Null},
			dst:      &ipSlice,
			expected: (([]net.IP)(nil)),
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
}
