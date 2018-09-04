package pgtype_test

import (
	"net"
	"testing"

	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/lib/pq"
)

// Test for renamed types
type _string string
type _bool bool
type _int8 int8
type _int16 int16
type _int16Slice []int16
type _int32Slice []int32
type _int64Slice []int64
type _float32Slice []float32
type _float64Slice []float64
type _byteSlice []byte

func mustParseCIDR(t testing.TB, s string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatal(err)
	}

	return ipnet
}

func mustParseMacaddr(t testing.TB, s string) net.HardwareAddr {
	addr, err := net.ParseMAC(s)
	if err != nil {
		t.Fatal(err)
	}

	return addr
}
