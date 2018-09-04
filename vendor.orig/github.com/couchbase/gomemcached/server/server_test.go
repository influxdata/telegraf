package memcached

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/couchbase/gomemcached"
)

func TestTransmitRes(t *testing.T) {
	b := &bytes.Buffer{}
	buf := bufio.NewWriter(b)

	res := gomemcached.MCResponse{
		Opcode: gomemcached.SET,
		Cas:    938424885,
		Opaque: 7242,
		Status: 0x338,
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	_, err := transmitResponse(buf, &res)
	if err != nil {
		t.Fatalf("Error transmitting request: %v", err)
	}

	buf.Flush()

	expected := []byte{
		gomemcached.RES_MAGIC, byte(gomemcached.SET),
		0x0, 0x7, // length of key
		0x0,       // extra length
		0x0,       // reserved
		0x3, 0x38, // Status
		0x0, 0x0, 0x0, 0x10, // Length of value
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e'}

	if len(b.Bytes()) != res.Size() {
		t.Fatalf("Expected %v bytes, got %v", res.Size(),
			len(b.Bytes()))
	}

	if !reflect.DeepEqual(b.Bytes(), expected) {
		t.Fatalf("Expected:\n%#v\n  -- got -- \n%#v",
			expected, b.Bytes())
	}
}

type closedrwc struct{}

func (c closedrwc) Close() error {
	return nil
}

func (c closedrwc) Write([]byte) (int, error) {
	return 0, io.EOF
}

func (c closedrwc) Read([]byte) (int, error) {
	return 0, io.EOF
}

func TestHandleIO(t *testing.T) {
	c := closedrwc{}
	err := HandleIO(c, nil)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func BenchmarkTransmitRes(b *testing.B) {
	bout := &bytes.Buffer{}

	res := gomemcached.MCResponse{
		Opcode: gomemcached.SET,
		Cas:    938424885,
		Opaque: 7242,
		Status: 824,
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	b.SetBytes(int64(res.Size()))

	for i := 0; i < b.N; i++ {
		bout.Reset()
		buf := bufio.NewWriterSize(bout, res.Size()*2)
		_, err := transmitResponse(buf, &res)
		if err != nil {
			b.Fatalf("Error transmitting request: %v", err)
		}
	}
}

func BenchmarkTransmitResLarge(b *testing.B) {
	bout := &bytes.Buffer{}

	res := gomemcached.MCResponse{
		Opcode: gomemcached.SET,
		Cas:    938424885,
		Opaque: 7242,
		Status: 824,
		Key:    []byte("somekey"),
		Body:   make([]byte, 24*1024),
	}

	b.SetBytes(int64(res.Size()))

	for i := 0; i < b.N; i++ {
		bout.Reset()
		buf := bufio.NewWriterSize(bout, res.Size()*2)
		_, err := transmitResponse(buf, &res)
		if err != nil {
			b.Fatalf("Error transmitting request: %v", err)
		}
	}
}

func BenchmarkTransmitResNull(b *testing.B) {
	res := gomemcached.MCResponse{
		Opcode: gomemcached.SET,
		Cas:    938424885,
		Opaque: 7242,
		Status: 824,
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	b.SetBytes(int64(res.Size()))

	for i := 0; i < b.N; i++ {
		_, err := transmitResponse(ioutil.Discard, &res)
		if err != nil {
			b.Fatalf("Error transmitting request: %v", err)
		}
	}
}

func TestMust(t *testing.T) {
	must(nil)
	errored := false
	func() {
		defer func() { _, errored = recover().(error) }()
		must(&gomemcached.MCResponse{})
	}()
}

func TestFuncHandler(t *testing.T) {
	ran := false
	h := FuncHandler(func(io.Writer, *gomemcached.MCRequest) *gomemcached.MCResponse {
		ran = true
		return nil
	})
	h.HandleMessage(nil, nil)
	if !ran {
		t.Fatalf("Didn't run our custom function")
	}
}

func BenchmarkReceive(b *testing.B) {
	res := gomemcached.MCResponse{
		Opcode: gomemcached.SET,
		Cas:    938424885,
		Opaque: 7242,
		Status: 824,
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	datum := res.Bytes()
	datum[0] = gomemcached.REQ_MAGIC
	b.SetBytes(int64(len(datum)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadPacket(bytes.NewReader(datum))
		if err != nil {
			b.Fatalf("Failed to read: %v", err)
		}
	}
}
