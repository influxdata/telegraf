//go:build linux
// +build linux

package net

import "testing"

func BenchmarkGetConnectionsInet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Connections("inet")
	}
}

func BenchmarkGetConnectionsAll(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Connections("all")
	}
}
