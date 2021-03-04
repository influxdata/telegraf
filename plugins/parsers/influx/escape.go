package influx

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

const (
	escapes            = " ,="
	nameEscapes        = " ,"
	stringFieldEscapes = `\"`
)

var (
	unescaper = strings.NewReplacer(
		`\,`, `,`,
		`\"`, `"`, // ???
		`\ `, ` `,
		`\=`, `=`,
	)

	nameUnescaper = strings.NewReplacer(
		`\,`, `,`,
		`\ `, ` `,
	)

	stringFieldUnescaper = strings.NewReplacer(
		`\"`, `"`,
		`\\`, `\`,
	)
)

func unescape(b []byte) string {
	if bytes.ContainsAny(b, escapes) {
		return unescaper.Replace(unsafeBytesToString(b))
	}
	return string(b)
}

func nameUnescape(b []byte) string {
	if bytes.ContainsAny(b, nameEscapes) {
		return nameUnescaper.Replace(unsafeBytesToString(b))
	}
	return string(b)
}

func stringFieldUnescape(b []byte) string {
	if bytes.ContainsAny(b, stringFieldEscapes) {
		return stringFieldUnescaper.Replace(unsafeBytesToString(b))
	}
	return string(b)
}

// parseIntBytes is a zero-alloc wrapper around strconv.ParseInt.
func parseIntBytes(b []byte, base int, bitSize int) (i int64, err error) {
	s := unsafeBytesToString(b)
	return strconv.ParseInt(s, base, bitSize)
}

// parseUintBytes is a zero-alloc wrapper around strconv.ParseUint.
func parseUintBytes(b []byte, base int, bitSize int) (i uint64, err error) {
	s := unsafeBytesToString(b)
	return strconv.ParseUint(s, base, bitSize)
}

// parseFloatBytes is a zero-alloc wrapper around strconv.ParseFloat.
func parseFloatBytes(b []byte, bitSize int) (float64, error) {
	s := unsafeBytesToString(b)
	return strconv.ParseFloat(s, bitSize)
}

// parseBoolBytes is a zero-alloc wrapper around strconv.ParseBool.
func parseBoolBytes(b []byte) (bool, error) {
	return strconv.ParseBool(unsafeBytesToString(b))
}

// unsafeBytesToString converts a []byte to a string without a heap allocation.
//
// It is unsafe, and is intended to prepare input to short-lived functions
// that require strings.
func unsafeBytesToString(in []byte) string {
	src := *(*reflect.SliceHeader)(unsafe.Pointer(&in))
	dst := reflect.StringHeader{
		Data: src.Data,
		Len:  src.Len,
	}
	s := *(*string)(unsafe.Pointer(&dst))
	return s
}
