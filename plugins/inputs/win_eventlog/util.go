// +build windows

package win_eventlog

import (
	"bytes"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

func DecodeUTF16(b []byte) ([]byte, error) {

	if len(b)%2 != 0 {
		return nil, fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.Bytes(), nil
}
