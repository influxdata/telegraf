package raft

import (
	"net"
	"testing"
)

func TestTCPTransport_BadAddr(t *testing.T) {
	_, err := NewTCPTransport("0.0.0.0:0", nil, 1, 0, nil)
	if err != errNotAdvertisable {
		t.Fatalf("err: %v", err)
	}
}

func TestTCPTransport_WithAdvertise(t *testing.T) {
	addr := &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 12345}
	trans, err := NewTCPTransport("0.0.0.0:0", addr, 1, 0, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if trans.LocalAddr() != "127.0.0.1:12345" {
		t.Fatalf("bad: %v", trans.LocalAddr())
	}
}
