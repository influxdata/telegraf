package config

import "bytes"

func unquoteTomlString(b []byte) []byte {
	if len(b) >= 6 {
		if bytes.HasPrefix(b, []byte(`'''`)) && bytes.HasSuffix(b, []byte(`'''`)) {
			return b[3 : len(b)-3]
		}
		if bytes.HasPrefix(b, []byte(`"""`)) && bytes.HasSuffix(b, []byte(`"""`)) {
			return b[3 : len(b)-3]
		}
	}
	if len(b) >= 2 {
		if bytes.HasPrefix(b, []byte(`'`)) && bytes.HasSuffix(b, []byte(`'`)) {
			return b[1 : len(b)-1]
		}
		if bytes.HasPrefix(b, []byte(`"`)) && bytes.HasSuffix(b, []byte(`"`)) {
			return b[1 : len(b)-1]
		}
	}
	return b
}
