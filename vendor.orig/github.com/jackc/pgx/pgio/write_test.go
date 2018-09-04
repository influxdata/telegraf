package pgio

import (
	"reflect"
	"testing"
)

func TestAppendUint16NilBuf(t *testing.T) {
	buf := AppendUint16(nil, 1)
	if !reflect.DeepEqual(buf, []byte{0, 1}) {
		t.Errorf("AppendUint16(nil, 1) => %v, want %v", buf, []byte{0, 1})
	}
}

func TestAppendUint16EmptyBuf(t *testing.T) {
	buf := []byte{}
	buf = AppendUint16(buf, 1)
	if !reflect.DeepEqual(buf, []byte{0, 1}) {
		t.Errorf("AppendUint16(nil, 1) => %v, want %v", buf, []byte{0, 1})
	}
}

func TestAppendUint16BufWithCapacityDoesNotAllocate(t *testing.T) {
	buf := make([]byte, 0, 4)
	AppendUint16(buf, 1)
	buf = buf[0:2]
	if !reflect.DeepEqual(buf, []byte{0, 1}) {
		t.Errorf("AppendUint16(nil, 1) => %v, want %v", buf, []byte{0, 1})
	}
}

func TestAppendUint32NilBuf(t *testing.T) {
	buf := AppendUint32(nil, 1)
	if !reflect.DeepEqual(buf, []byte{0, 0, 0, 1}) {
		t.Errorf("AppendUint32(nil, 1) => %v, want %v", buf, []byte{0, 0, 0, 1})
	}
}

func TestAppendUint32EmptyBuf(t *testing.T) {
	buf := []byte{}
	buf = AppendUint32(buf, 1)
	if !reflect.DeepEqual(buf, []byte{0, 0, 0, 1}) {
		t.Errorf("AppendUint32(nil, 1) => %v, want %v", buf, []byte{0, 0, 0, 1})
	}
}

func TestAppendUint32BufWithCapacityDoesNotAllocate(t *testing.T) {
	buf := make([]byte, 0, 4)
	AppendUint32(buf, 1)
	buf = buf[0:4]
	if !reflect.DeepEqual(buf, []byte{0, 0, 0, 1}) {
		t.Errorf("AppendUint32(nil, 1) => %v, want %v", buf, []byte{0, 0, 0, 1})
	}
}

func TestAppendUint64NilBuf(t *testing.T) {
	buf := AppendUint64(nil, 1)
	if !reflect.DeepEqual(buf, []byte{0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Errorf("AppendUint64(nil, 1) => %v, want %v", buf, []byte{0, 0, 0, 0, 0, 0, 0, 1})
	}
}

func TestAppendUint64EmptyBuf(t *testing.T) {
	buf := []byte{}
	buf = AppendUint64(buf, 1)
	if !reflect.DeepEqual(buf, []byte{0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Errorf("AppendUint64(nil, 1) => %v, want %v", buf, []byte{0, 0, 0, 0, 0, 0, 0, 1})
	}
}

func TestAppendUint64BufWithCapacityDoesNotAllocate(t *testing.T) {
	buf := make([]byte, 0, 8)
	AppendUint64(buf, 1)
	buf = buf[0:8]
	if !reflect.DeepEqual(buf, []byte{0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Errorf("AppendUint64(nil, 1) => %v, want %v", buf, []byte{0, 0, 0, 0, 0, 0, 0, 1})
	}
}
