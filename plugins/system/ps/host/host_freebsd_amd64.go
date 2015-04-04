// +build freebsd
// +build amd64
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs host/types_freebsd.go

package host

const (
	sizeofPtr      = 0x8
	sizeofShort    = 0x2
	sizeofInt      = 0x4
	sizeofLong     = 0x8
	sizeofLongLong = 0x8
)

type (
	_C_short     int16
	_C_int       int32
	_C_long      int64
	_C_long_long int64
)

type Utmp struct {
	Line [8]int8
	Name [16]int8
	Host [16]int8
	Time int32
}
