package binaryio

import (
	"bytes"
	"testing"
)

func TestMinReader(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	r := bytes.NewBuffer(b)

	mr := MinReader(r, 10)

	toRead := make([]byte, 5)
	n, err := mr.Read(toRead)
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Error("Expected n to be 5, but was ", n)
	}
	if string(toRead) != string([]byte{1, 2, 3, 4, 5}) {
		t.Error("expected 5 specific bytes to be read")
	}
	err = mr.Close()
	if err != nil {
		t.Error(err)
	}
	n, err = r.Read(toRead) // read from the outer stream
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Error("Expected n to be 5, but was ", n)
	}
	if string(toRead) != string([]byte{11, 12, 13, 14, 15}) {
		t.Error("expected the last 5 bytes to be read")
	}
}
