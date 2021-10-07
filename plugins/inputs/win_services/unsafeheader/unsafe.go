//go:build windows
// +build windows

// https://github.com/golang/sys/blob/master/internal/unsafeheader/unsafeheader.go

package unsafeheader

import (
	"unsafe"
)

// Slice is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may change in a later release.
type Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
