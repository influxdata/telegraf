/*
 * These classes is a dynamic implementation of the go binary.ByteOrder interface that lets
 * you specify a byte order consistent with how various modbus libraries express ordering:
 *    [Modbus Poll](https://www.modbustools.com/poll_display_formats.html)
 * Users of this class pass it to a bytes.Reader which should do the rest of the work, including
 * filling in all the types that are not part of binary.ByteOrder
 *
 * Written by Christopher Piggott with the hope it will become standard in a replacement modbus
 * library (some time in the future)
 */

package modbus_gateway

import (
	"strings"
)

var byteOrderCache map[string]*CustomByteOrder = make(map[string]*CustomByteOrder)

func getOrCreateByteOrder(orderSpec string) *CustomByteOrder {
	key := strings.ToUpper(orderSpec)
	if byteOrderCache[key] != nil {
		return byteOrderCache[key]
	} else {
		formatter, _ := CreateCustomByteOrder(key)
		byteOrderCache[key] = formatter
		return formatter
	}
}

func CreateCustomByteOrder(orderSpec string) (*CustomByteOrder, error) {
	orderSpecUC := strings.ToUpper(orderSpec)
	orderSpecLen := len(orderSpecUC)
	orderSpecBytes := []byte(orderSpecUC)

	converter := &CustomByteOrder{
		order: orderSpecUC,
	}

	for i := 0; i < orderSpecLen; i++ {
		converter.positions[i] = int(orderSpecBytes[i] - 'A')
	}

	for i := orderSpecLen; i < 8; i++ {
		position := int(orderSpecBytes[i%orderSpecLen] - 'A')
		block := i / orderSpecLen
		position = position + (block * orderSpecLen)
		converter.positions[i] = position

	}

	return converter, nil
}

type CustomByteOrder struct {
	order     string
	positions [8]int
}

func (o *CustomByteOrder) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	const mod = 2
	return uint16(b[o.positions[1]%mod]) | uint16(b[o.positions[0]%mod])<<8
}

func (o *CustomByteOrder) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	const mod = 2
	b[o.positions[0]%mod] = byte(v >> 8)
	b[o.positions[1]%mod] = byte(v)
}

func (o *CustomByteOrder) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	const mod = 4
	return uint32(b[o.positions[0]%mod])<<24 |
		uint32(b[o.positions[1]%mod])<<16 |
		uint32(b[o.positions[2]%mod])<<8 |
		uint32(b[o.positions[3]%mod])<<0
}

func (o *CustomByteOrder) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	const mod = 4
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v >> 0)
}

func (o *CustomByteOrder) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[o.positions[7]]) |
		uint64(b[o.positions[6]])<<8 |
		uint64(b[o.positions[5]])<<16 |
		uint64(b[o.positions[4]])<<24 |
		uint64(b[o.positions[3]])<<32 |
		uint64(b[o.positions[2]])<<40 |
		uint64(b[o.positions[1]])<<48 |
		uint64(b[o.positions[0]])<<56
}

func (o *CustomByteOrder) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func (o *CustomByteOrder) String() string { return "CustomByteOrder" + o.order }
