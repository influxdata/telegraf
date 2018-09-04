package gomemcached

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestEncodingResponse(t *testing.T) {
	req := MCResponse{
		Opcode: SET,
		Status: 1582,
		Opaque: 7242,
		Cas:    938424885,
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	got := req.Bytes()

	expected := []byte{
		RES_MAGIC, byte(SET),
		0x0, 0x7, // length of key
		0x0,       // extra length
		0x0,       // reserved
		0x6, 0x2e, // status
		0x0, 0x0, 0x0, 0x10, // Length of value
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e'}

	if len(got) != req.Size() {
		t.Fatalf("Expected %v bytes, got %v", got,
			len(got))
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Expected:\n%#v\n  -- got -- \n%#v",
			expected, got)
	}

	exp := `{MCResponse status=0x62e keylen=7, extralen=0, bodylen=9}`
	if req.String() != exp {
		t.Errorf("Expected string=%q, got %q", exp, req.String())
	}

	exp = `MCResponse status=0x62e, opcode=SET, opaque=7242, msg: somevalue`
	if req.Error() != exp {
		t.Errorf("Expected string=%q, got %q", exp, req.Error())
	}
}

func TestEncodingResponseWithExtras(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 1582,
		Opaque: 7242,
		Cas:    938424885,
		Extras: []byte{1, 2, 3, 4},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	buf := &bytes.Buffer{}
	res.Transmit(buf)
	got := buf.Bytes()

	expected := []byte{
		RES_MAGIC, byte(SET),
		0x0, 0x7, // length of key
		0x4,       // extra length
		0x0,       // reserved
		0x6, 0x2e, // status
		0x0, 0x0, 0x0, 0x14, // Length of remainder
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		1, 2, 3, 4, // extras
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e'}

	if len(got) != res.Size() {
		t.Fatalf("Expected %v bytes, got %v", got,
			len(got))
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Expected:\n%#v\n  -- got -- \n%#v",
			expected, got)
	}
}

func TestEncodingResponseWithLargeBody(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 1582,
		Opaque: 7242,
		Cas:    938424885,
		Extras: []byte{1, 2, 3, 4},
		Key:    []byte("somekey"),
		Body:   make([]byte, 256),
	}

	buf := &bytes.Buffer{}
	res.Transmit(buf)
	got := buf.Bytes()

	expected := append([]byte{
		RES_MAGIC, byte(SET),
		0x0, 0x7, // length of key
		0x4,       // extra length
		0x0,       // reserved
		0x6, 0x2e, // status
		0x0, 0x0, 0x1, 0xb, // Length of remainder
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		1, 2, 3, 4, // extras
		's', 'o', 'm', 'e', 'k', 'e', 'y',
	}, make([]byte, 256)...)

	if len(got) != res.Size() {
		t.Fatalf("Expected %v bytes, got %v", got,
			len(got))
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Expected:\n%#v\n  -- got -- \n%#v",
			expected, got)
	}
}

func BenchmarkEncodingResponse(b *testing.B) {
	req := MCResponse{
		Opcode: SET,
		Status: 1582,
		Opaque: 7242,
		Cas:    938424885,
		Extras: []byte{},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	b.SetBytes(int64(req.Size()))

	for i := 0; i < b.N; i++ {
		req.Bytes()
	}
}

func BenchmarkEncodingResponseLarge(b *testing.B) {
	req := MCResponse{
		Opcode: SET,
		Status: 1582,
		Opaque: 7242,
		Cas:    938424885,
		Extras: []byte{},
		Key:    []byte("somekey"),
		Body:   make([]byte, 24*1024),
	}

	b.SetBytes(int64(req.Size()))

	for i := 0; i < b.N; i++ {
		req.Bytes()
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		e  error
		is bool
	}{
		{nil, false},
		{errors.New("something"), false},
		{&MCResponse{}, false},
		{&MCResponse{Status: KEY_ENOENT}, true},
	}

	for i, x := range tests {
		if IsNotFound(x.e) != x.is {
			t.Errorf("Expected %v for %#v (%v)", x.is, x.e, i)
		}
	}
}

func TestIsFatal(t *testing.T) {
	tests := []struct {
		e  error
		is bool
	}{
		{nil, false},
		{errors.New("something"), true},
		{&MCResponse{}, true},
		{&MCResponse{Status: KEY_ENOENT}, false},
		{&MCResponse{Status: EINVAL}, true},
		{&MCResponse{Status: TMPFAIL}, false},
	}

	for i, x := range tests {
		if IsFatal(x.e) != x.is {
			t.Errorf("Expected %v for %#v (%v)", x.is, x.e, i)
		}
	}
}

func TestResponseTransmit(t *testing.T) {
	res := MCResponse{Key: []byte("thekey")}
	_, err := res.Transmit(ioutil.Discard)
	if err != nil {
		t.Errorf("Error sending small response: %v", err)
	}

	res.Body = make([]byte, 256)
	_, err = res.Transmit(ioutil.Discard)
	if err != nil {
		t.Errorf("Error sending large response thing: %v", err)
	}

}

func TestReceiveResponse(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 74,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := res.Bytes()

	res2 := MCResponse{}
	_, err := res2.Receive(bytes.NewReader(data), nil)
	if err != nil {
		t.Fatalf("Error receiving: %v", err)
	}

	if !reflect.DeepEqual(res, res2) {
		t.Fatalf("Expected %#v == %#v", res, res2)
	}
}

func TestReceiveResponseBadMagic(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 74,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := res.Bytes()
	data[0] = 0x13

	res2 := MCResponse{}
	_, err := res2.Receive(bytes.NewReader(data), nil)
	if err == nil {
		t.Fatalf("Expected error, got: %#v", res2)
	}
}

func TestReceiveResponseShortHeader(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 74,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := res.Bytes()
	data[0] = 0x13

	res2 := MCResponse{}
	_, err := res2.Receive(bytes.NewReader(data[:13]), nil)
	if err == nil {
		t.Fatalf("Expected error, got: %#v", res2)
	}
}

func TestReceiveResponseShortBody(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 74,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := res.Bytes()
	data[0] = 0x13

	res2 := MCResponse{}
	_, err := res2.Receive(bytes.NewReader(data[:len(data)-3]), nil)
	if err == nil {
		t.Fatalf("Expected error, got: %#v", res2)
	}
}

func TestReceiveResponseWithBuffer(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 74,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := res.Bytes()

	res2 := MCResponse{}
	buf := make([]byte, HDR_LEN)
	_, err := res2.Receive(bytes.NewReader(data), buf)
	if err != nil {
		t.Fatalf("Error receiving: %v", err)
	}

	if !reflect.DeepEqual(res, res2) {
		t.Fatalf("Expected %#v == %#v", res, res2)
	}
}

func TestReceiveResponseNoContent(t *testing.T) {
	res := MCResponse{
		Opcode: SET,
		Status: 74,
		Opaque: 7242,
		Extras: []byte{},
		Key:    []byte{},
		Body:   []byte{},
	}

	data := res.Bytes()

	res2 := MCResponse{}
	_, err := res2.Receive(bytes.NewReader(data), nil)
	if err != nil {
		t.Fatalf("Error receiving: %v", err)
	}

	// Can't use reflect here because []byte{} != nil, though they
	// look the same.
	if fmt.Sprintf("%#v", res) != fmt.Sprintf("%#v", res2) {
		t.Fatalf("Expected %#v == %#v", res, res2)
	}
}

func BenchmarkReceiveResponse(b *testing.B) {
	req := MCResponse{
		Opcode: SET,
		Status: 183,
		Cas:    0,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := req.Bytes()
	rdr := bytes.NewReader(data)

	b.SetBytes(int64(len(data)))

	b.ResetTimer()
	buf := make([]byte, HDR_LEN)
	for i := 0; i < b.N; i++ {
		res2 := MCResponse{}
		rdr.Seek(0, 0)
		res2.Receive(rdr, buf)
	}
}

func BenchmarkReceiveResponseNoBuf(b *testing.B) {
	req := MCResponse{
		Opcode: SET,
		Status: 183,
		Cas:    0,
		Opaque: 7242,
		Extras: []byte{1},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
	}

	data := req.Bytes()
	rdr := bytes.NewReader(data)

	b.SetBytes(int64(len(data)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res2 := MCResponse{}
		rdr.Seek(0, 0)
		res2.Receive(rdr, nil)
	}
}
