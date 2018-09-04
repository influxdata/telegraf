//
// This is a thin wrapper around sync/atomic to help with alignment issues.
// This is for 64-bit OS and hence is a no-op effectively.
//

// +build !386

package platform

import "unsafe"
import orig "sync/atomic"

type AlignedInt64 int64
type AlignedUint64 uint64

func NewAlignedInt64(v int64) AlignedInt64 {
	return AlignedInt64(v)
}

func NewAlignedUint64(v uint64) AlignedUint64 {
	return AlignedUint64(v)
}

func SwapInt32(addr *int32, new int32) int32 {
	return orig.SwapInt32(addr, new)
}

func SwapInt64(addr *AlignedInt64, new int64) int64 {
	return orig.SwapInt64((*int64)(addr), new)
}

func SwapUint32(addr *uint32, new uint32) uint32 {
	return orig.SwapUint32(addr, new)
}

func SwapUint64(addr *AlignedUint64, new uint64) uint64 {
	return orig.SwapUint64((*uint64)(addr), new)
}

func SwapUintptr(addr *uintptr, new uintptr) uintptr {
	return orig.SwapUintptr(addr, new)
}

func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	return orig.SwapPointer(addr, new)
}

func CompareAndSwapInt32(addr *int32, old, new int32) bool {
	return orig.CompareAndSwapInt32(addr, old, new)
}

func CompareAndSwapInt64(addr *AlignedInt64, old, new int64) bool {
	return orig.CompareAndSwapInt64((*int64)(addr), old, new)
}

func CompareAndSwapUint32(addr *uint32, old, new uint32) bool {
	return orig.CompareAndSwapUint32(addr, old, new)
}

func CompareAndSwapUint64(addr *AlignedUint64, old, new uint64) bool {
	return orig.CompareAndSwapUint64((*uint64)(addr), old, new)
}

func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) bool {
	return orig.CompareAndSwapUintptr(addr, old, new)
}

func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) bool {
	return orig.CompareAndSwapPointer(addr, old, new)
}

func AddInt32(addr *int32, delta int32) int32 {
	return orig.AddInt32(addr, delta)
}

func AddUint32(addr *uint32, delta uint32) uint32 {
	return orig.AddUint32(addr, delta)
}

func AddInt64(addr *AlignedInt64, delta int64) int64 {
	return orig.AddInt64((*int64)(addr), delta)
}

func AddUint64(addr *AlignedUint64, delta uint64) uint64 {
	return orig.AddUint64((*uint64)(addr), delta)
}

func AddUintptr(addr *uintptr, delta uintptr) uintptr {
	return orig.AddUintptr(addr, delta)
}

func LoadInt32(addr *int32) int32 {
	return orig.LoadInt32(addr)
}

func LoadInt64(addr *AlignedInt64) int64 {
	return orig.LoadInt64((*int64)(addr))
}

func LoadUint32(addr *uint32) uint32 {
	return orig.LoadUint32(addr)
}

func LoadUint64(addr *AlignedUint64) uint64 {
	return orig.LoadUint64((*uint64)(addr))
}

func LoadUintptr(addr *uintptr) uintptr {
	return orig.LoadUintptr(addr)
}

func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer {
	return orig.LoadPointer(addr)
}

func StoreInt32(addr *int32, val int32) {
	orig.StoreInt32(addr, val)
}

func StoreInt64(addr *AlignedInt64, val int64) {
	orig.StoreInt64((*int64)(addr), val)
}

func StoreUint32(addr *uint32, val uint32) {
	orig.StoreUint32(addr, val)
}

func StoreUint64(addr *AlignedUint64, val uint64) {
	orig.StoreUint64((*uint64)(addr), val)
}

func StoreUintptr(addr *uintptr, val uintptr) {
	orig.StoreUintptr(addr, val)
}

func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) {
	orig.StorePointer(addr, val)
}
