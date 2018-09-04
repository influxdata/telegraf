package pgtype

import (
	"bytes"
	"testing"
)

func TestParseUntypedTextRange(t *testing.T) {
	tests := []struct {
		src    string
		result UntypedTextRange
		err    error
	}{
		{
			src:    `[1,2)`,
			result: UntypedTextRange{Lower: "1", Upper: "2", LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `[1,2]`,
			result: UntypedTextRange{Lower: "1", Upper: "2", LowerType: Inclusive, UpperType: Inclusive},
			err:    nil,
		},
		{
			src:    `(1,3)`,
			result: UntypedTextRange{Lower: "1", Upper: "3", LowerType: Exclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    ` [1,2) `,
			result: UntypedTextRange{Lower: "1", Upper: "2", LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `[ foo , bar )`,
			result: UntypedTextRange{Lower: " foo ", Upper: " bar ", LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `["foo","bar")`,
			result: UntypedTextRange{Lower: "foo", Upper: "bar", LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `["f""oo","b""ar")`,
			result: UntypedTextRange{Lower: `f"oo`, Upper: `b"ar`, LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `["f""oo","b""ar")`,
			result: UntypedTextRange{Lower: `f"oo`, Upper: `b"ar`, LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `["","bar")`,
			result: UntypedTextRange{Lower: ``, Upper: `bar`, LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `[f\"oo\,,b\\ar\))`,
			result: UntypedTextRange{Lower: `f"oo,`, Upper: `b\ar)`, LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    `empty`,
			result: UntypedTextRange{Lower: "", Upper: "", LowerType: Empty, UpperType: Empty},
			err:    nil,
		},
	}

	for i, tt := range tests {
		r, err := ParseUntypedTextRange(tt.src)
		if err != tt.err {
			t.Errorf("%d. `%v`: expected err %v, got %v", i, tt.src, tt.err, err)
			continue
		}

		if r.LowerType != tt.result.LowerType {
			t.Errorf("%d. `%v`: expected result lower type %v, got %v", i, tt.src, string(tt.result.LowerType), string(r.LowerType))
		}

		if r.UpperType != tt.result.UpperType {
			t.Errorf("%d. `%v`: expected result upper type %v, got %v", i, tt.src, string(tt.result.UpperType), string(r.UpperType))
		}

		if r.Lower != tt.result.Lower {
			t.Errorf("%d. `%v`: expected result lower %v, got %v", i, tt.src, tt.result.Lower, r.Lower)
		}

		if r.Upper != tt.result.Upper {
			t.Errorf("%d. `%v`: expected result upper %v, got %v", i, tt.src, tt.result.Upper, r.Upper)
		}
	}
}

func TestParseUntypedBinaryRange(t *testing.T) {
	tests := []struct {
		src    []byte
		result UntypedBinaryRange
		err    error
	}{
		{
			src:    []byte{0, 0, 0, 0, 2, 0, 4, 0, 0, 0, 2, 0, 5},
			result: UntypedBinaryRange{Lower: []byte{0, 4}, Upper: []byte{0, 5}, LowerType: Exclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    []byte{1},
			result: UntypedBinaryRange{Lower: nil, Upper: nil, LowerType: Empty, UpperType: Empty},
			err:    nil,
		},
		{
			src:    []byte{2, 0, 0, 0, 2, 0, 4, 0, 0, 0, 2, 0, 5},
			result: UntypedBinaryRange{Lower: []byte{0, 4}, Upper: []byte{0, 5}, LowerType: Inclusive, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    []byte{4, 0, 0, 0, 2, 0, 4, 0, 0, 0, 2, 0, 5},
			result: UntypedBinaryRange{Lower: []byte{0, 4}, Upper: []byte{0, 5}, LowerType: Exclusive, UpperType: Inclusive},
			err:    nil,
		},
		{
			src:    []byte{6, 0, 0, 0, 2, 0, 4, 0, 0, 0, 2, 0, 5},
			result: UntypedBinaryRange{Lower: []byte{0, 4}, Upper: []byte{0, 5}, LowerType: Inclusive, UpperType: Inclusive},
			err:    nil,
		},
		{
			src:    []byte{8, 0, 0, 0, 2, 0, 5},
			result: UntypedBinaryRange{Lower: nil, Upper: []byte{0, 5}, LowerType: Unbounded, UpperType: Exclusive},
			err:    nil,
		},
		{
			src:    []byte{12, 0, 0, 0, 2, 0, 5},
			result: UntypedBinaryRange{Lower: nil, Upper: []byte{0, 5}, LowerType: Unbounded, UpperType: Inclusive},
			err:    nil,
		},
		{
			src:    []byte{16, 0, 0, 0, 2, 0, 4},
			result: UntypedBinaryRange{Lower: []byte{0, 4}, Upper: nil, LowerType: Exclusive, UpperType: Unbounded},
			err:    nil,
		},
		{
			src:    []byte{18, 0, 0, 0, 2, 0, 4},
			result: UntypedBinaryRange{Lower: []byte{0, 4}, Upper: nil, LowerType: Inclusive, UpperType: Unbounded},
			err:    nil,
		},
		{
			src:    []byte{24},
			result: UntypedBinaryRange{Lower: nil, Upper: nil, LowerType: Unbounded, UpperType: Unbounded},
			err:    nil,
		},
	}

	for i, tt := range tests {
		r, err := ParseUntypedBinaryRange(tt.src)
		if err != tt.err {
			t.Errorf("%d. `%v`: expected err %v, got %v", i, tt.src, tt.err, err)
			continue
		}

		if r.LowerType != tt.result.LowerType {
			t.Errorf("%d. `%v`: expected result lower type %v, got %v", i, tt.src, string(tt.result.LowerType), string(r.LowerType))
		}

		if r.UpperType != tt.result.UpperType {
			t.Errorf("%d. `%v`: expected result upper type %v, got %v", i, tt.src, string(tt.result.UpperType), string(r.UpperType))
		}

		if bytes.Compare(r.Lower, tt.result.Lower) != 0 {
			t.Errorf("%d. `%v`: expected result lower %v, got %v", i, tt.src, tt.result.Lower, r.Lower)
		}

		if bytes.Compare(r.Upper, tt.result.Upper) != 0 {
			t.Errorf("%d. `%v`: expected result upper %v, got %v", i, tt.src, tt.result.Upper, r.Upper)
		}
	}
}
