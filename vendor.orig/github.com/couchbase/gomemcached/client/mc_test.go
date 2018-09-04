package memcached

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"testing"

	"github.com/couchbase/gomemcached"
)

func TestConnect(t *testing.T) {
	defer func() { dialFun = net.Dial }()

	dialFun = func(p, dest string) (net.Conn, error) {
		if dest == "broken" {
			return nil, io.ErrNoProgress
		}
		return &net.TCPConn{}, nil
	}

	c, err := Connect("tcp", "broken")
	if err == nil {
		t.Errorf("Expected failure, got %v", c)
	}

	c, err = Connect("tcp", "working")
	if err != nil {
		t.Errorf("Expected a connection, got %v", err)
	}
}

type tracked bool

func (t *tracked) Close() error {
	*t = true
	return nil
}

func (t tracked) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (t tracked) Write([]byte) (int, error) {
	return 0, io.EOF
}

func TestClose(t *testing.T) {
	var tr tracked
	c, err := Wrap(&tr)
	must(err)
	c.Close()

	if !tr {
		t.Errorf("Expected to close, but didn't")
	}
}

func TestHealthy(t *testing.T) {
	var tr tracked
	c, err := Wrap(&tr)
	must(err)
	if !c.IsHealthy() {
		t.Errorf("Expected healthy.  Wasn't.")
	}

	res, err := c.Send(&gomemcached.MCRequest{})
	if err == nil {
		t.Errorf("Expected error transmitting, got %v", res)
	}

	if c.IsHealthy() {
		t.Errorf("Expected unhealthy.  Wasn't.")
	}
}

func TestTransmitReq(t *testing.T) {
	b := bytes.NewBuffer([]byte{})
	buf := bufio.NewWriter(b)

	req := gomemcached.MCRequest{
		Opcode:  gomemcached.SET,
		Cas:     938424885,
		Opaque:  7242,
		VBucket: 824,
		Extras:  []byte{},
		Key:     []byte("somekey"),
		Body:    []byte("somevalue"),
	}

	// Verify nil transmit is OK
	_, err := transmitRequest(nil, &req)
	if err != errNoConn {
		t.Errorf("Expected errNoConn with no conn, got %v", err)
	}

	_, err = transmitRequest(buf, &req)
	if err != nil {
		t.Fatalf("Error transmitting request: %v", err)
	}

	buf.Flush()

	expected := []byte{
		gomemcached.REQ_MAGIC, byte(gomemcached.SET),
		0x0, 0x7, // length of key
		0x0,       // extra length
		0x0,       // reserved
		0x3, 0x38, // vbucket
		0x0, 0x0, 0x0, 0x10, // Length of value
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e'}

	if len(b.Bytes()) != req.Size() {
		t.Fatalf("Expected %v bytes, got %v", req.Size(),
			len(b.Bytes()))
	}

	if !reflect.DeepEqual(b.Bytes(), expected) {
		t.Fatalf("Expected:\n%#v\n  -- got -- \n%#v",
			expected, b.Bytes())
	}
}

func TestTransmitReqWithExtMeta(t *testing.T) {
	// test data for extended metadata
	ExtMetaStr := "extmeta"

	b := bytes.NewBuffer([]byte{})
	buf := bufio.NewWriter(b)

	req := gomemcached.MCRequest{
		Opcode:  gomemcached.SET,
		Cas:     938424885,
		Opaque:  7242,
		VBucket: 824,
		Key:     []byte("somekey"),
		Body:    []byte("somevalue"),
		ExtMeta: []byte(ExtMetaStr),
	}

	// add length of extended metadata to the corresponding bytes in Extras
	req.Extras = make([]byte, 30)
	binary.BigEndian.PutUint16(req.Extras[28:30], uint16(len(ExtMetaStr)))

	// Verify nil transmit is OK
	_, err := transmitRequest(nil, &req)
	if err != errNoConn {
		t.Errorf("Expected errNoConn with no conn, got %v", err)
	}

	_, err = transmitRequest(buf, &req)
	if err != nil {
		t.Fatalf("Error transmitting request: %v", err)
	}

	buf.Flush()

	expected := []byte{
		gomemcached.REQ_MAGIC, byte(gomemcached.SET),
		0x0, 0x7, // length of key
		0x1e,      // extra length = 30 = 0x1e
		0x0,       // reserved
		0x3, 0x38, // vbucket
		0x0, 0x0, 0x0, 0x35, // Length of value = 7(key) + 9(value) + 30(extras) + 7(extmeta) = 53 = 0x35
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, //
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // --> Extras (with 28:30 carrying ExtMetaLen)
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7, //
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e',
		'e', 'x', 't', 'm', 'e', 't', 'a'}

	if len(b.Bytes()) != req.Size() {
		t.Fatalf("Expected %v bytes, got %v", req.Size(),
			len(b.Bytes()))
	}

	if !reflect.DeepEqual(b.Bytes(), expected) {
		t.Fatalf("Expected:\n%#v\n  -- got -- \n%#v",
			expected, b.Bytes())
	}
}

func BenchmarkTransmitReq(b *testing.B) {
	bout := bytes.NewBuffer([]byte{})

	req := gomemcached.MCRequest{
		Opcode:  gomemcached.SET,
		Cas:     938424885,
		Opaque:  7242,
		VBucket: 824,
		Extras:  []byte{},
		Key:     []byte("somekey"),
		Body:    []byte("somevalue"),
	}

	b.SetBytes(int64(req.Size()))

	for i := 0; i < b.N; i++ {
		bout.Reset()
		buf := bufio.NewWriterSize(bout, req.Size()*2)
		_, err := transmitRequest(buf, &req)
		if err != nil {
			b.Fatalf("Error transmitting request: %v", err)
		}
	}
}

func BenchmarkTransmitReqLarge(b *testing.B) {
	bout := bytes.NewBuffer([]byte{})

	req := gomemcached.MCRequest{
		Opcode:  gomemcached.SET,
		Cas:     938424885,
		Opaque:  7242,
		VBucket: 824,
		Extras:  []byte{},
		Key:     []byte("somekey"),
		Body:    make([]byte, 24*1024),
	}

	b.SetBytes(int64(req.Size()))

	for i := 0; i < b.N; i++ {
		bout.Reset()
		buf := bufio.NewWriterSize(bout, req.Size()*2)
		_, err := transmitRequest(buf, &req)
		if err != nil {
			b.Fatalf("Error transmitting request: %v", err)
		}
	}
}

func BenchmarkTransmitReqNull(b *testing.B) {
	req := gomemcached.MCRequest{
		Opcode:  gomemcached.SET,
		Cas:     938424885,
		Opaque:  7242,
		VBucket: 824,
		Extras:  []byte{},
		Key:     []byte("somekey"),
		Body:    []byte("somevalue"),
	}

	b.SetBytes(int64(req.Size()))

	for i := 0; i < b.N; i++ {
		_, err := transmitRequest(ioutil.Discard, &req)
		if err != nil {
			b.Fatalf("Error transmitting request: %v", err)
		}
	}
}

/*
       |0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|
       +---------------+---------------+---------------+---------------+
      0| 0x81          | 0x00          | 0x00          | 0x00          |
       +---------------+---------------+---------------+---------------+
      4| 0x04          | 0x00          | 0x00          | 0x00          |
       +---------------+---------------+---------------+---------------+
      8| 0x00          | 0x00          | 0x00          | 0x09          |
       +---------------+---------------+---------------+---------------+
     12| 0x00          | 0x00          | 0x00          | 0x00          |
       +---------------+---------------+---------------+---------------+
     16| 0x00          | 0x00          | 0x00          | 0x00          |
       +---------------+---------------+---------------+---------------+
     20| 0x00          | 0x00          | 0x00          | 0x01          |
       +---------------+---------------+---------------+---------------+
     24| 0xde          | 0xad          | 0xbe          | 0xef          |
       +---------------+---------------+---------------+---------------+
     28| 0x57 ('W')    | 0x6f ('o')    | 0x72 ('r')    | 0x6c ('l')    |
       +---------------+---------------+---------------+---------------+
     32| 0x64 ('d')    |
       +---------------+

   Field        (offset) (value)
   Magic        (0)    : 0x81
   Opcode       (1)    : 0x00
   Key length   (2,3)  : 0x0000
   Extra length (4)    : 0x04
   Data type    (5)    : 0x00
   Status       (6,7)  : 0x0000
   Total body   (8-11) : 0x00000009
   Opaque       (12-15): 0x00000000
   CAS          (16-23): 0x0000000000000001
   Extras              :
     Flags      (24-27): 0xdeadbeef
   Key                 : None
   Value        (28-32): The textual string "World"

*/

func TestDecodeSpecSample(t *testing.T) {
	data := []byte{
		0x81, 0x00, 0x00, 0x00, // 0
		0x04, 0x00, 0x00, 0x00, // 4
		0x00, 0x00, 0x00, 0x09, // 8
		0x00, 0x00, 0x00, 0x00, // 12
		0x00, 0x00, 0x00, 0x00, // 16
		0x00, 0x00, 0x00, 0x01, // 20
		0xde, 0xad, 0xbe, 0xef, // 24
		0x57, 0x6f, 0x72, 0x6c, // 28
		0x64, // 32
	}

	buf := make([]byte, gomemcached.HDR_LEN)
	res, _, err := getResponse(bytes.NewReader(data), buf)
	if err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}

	expected := &gomemcached.MCResponse{
		Opcode: gomemcached.GET,
		Status: 0,
		Opaque: 0,
		Cas:    1,
		Extras: []byte{0xde, 0xad, 0xbe, 0xef},
		Key:    []byte{},
		Body:   []byte("World"),
		Fatal:  false,
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected\n%#v -- got --\n%#v", expected, res)
	}

}

func TestNilReader(t *testing.T) {
	res, _, err := getResponse(nil, nil)
	if err != errNoConn {
		t.Fatalf("Expected error reading from nil, got %#v", res)
	}
}

func TestDecode(t *testing.T) {
	data := []byte{
		gomemcached.RES_MAGIC, byte(gomemcached.SET),
		0x0, 0x7, // length of key
		0x0,       // extra length
		0x0,       // reserved
		0x6, 0x2e, // status
		0x0, 0x0, 0x0, 0x10, // Length of value
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e'}

	buf := make([]byte, gomemcached.HDR_LEN)
	res, _, err := getResponse(bytes.NewReader(data), buf)
	res, err = UnwrapMemcachedError(res, err)
	if err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}

	expected := &gomemcached.MCResponse{
		Opcode: gomemcached.SET,
		Status: 1582,
		Opaque: 7242,
		Cas:    938424885,
		Extras: []byte{},
		Key:    []byte("somekey"),
		Body:   []byte("somevalue"),
		Fatal:  false,
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Expected\n%#v -- got --\n%#v", expected, res)
	}
}

func BenchmarkDecodeResponse(b *testing.B) {
	data := []byte{
		gomemcached.RES_MAGIC, byte(gomemcached.SET),
		0x0, 0x7, // length of key
		0x0,       // extra length
		0x0,       // reserved
		0x6, 0x2e, // status
		0x0, 0x0, 0x0, 0x10, // Length of value
		0x0, 0x0, 0x1c, 0x4a, // opaque
		0x0, 0x0, 0x0, 0x0, 0x37, 0xef, 0x3a, 0x35, // CAS
		's', 'o', 'm', 'e', 'k', 'e', 'y',
		's', 'o', 'm', 'e', 'v', 'a', 'l', 'u', 'e'}
	buf := make([]byte, gomemcached.HDR_LEN)
	b.SetBytes(int64(len(buf)))

	for i := 0; i < b.N; i++ {
		getResponse(bytes.NewReader(data), buf)
	}
}

func TestUnwrap(t *testing.T) {
	res := &gomemcached.MCResponse{}

	_, e := UnwrapMemcachedError(res, res)
	if e != nil {
		t.Errorf("Expected error to be nilled, got %v", e)
	}

	_, e = UnwrapMemcachedError(res, errNoConn)
	if e != errNoConn {
		t.Errorf("Expected error to come through, got %v", e)
	}
}

func panics(f func() interface{}) (got interface{}, panicked bool) {
	defer func() { panicked = recover() != nil }()
	return f(), false
}

func TestCasOpError(t *testing.T) {
	known := map[CasOp]string{
		CASStore:  "CAS store",
		CASQuit:   "CAS quit",
		CASDelete: "CAS delete",
	}

	for i := 0; i < 0x100; i++ {
		c := CasOp(i)
		if s, ok := known[c]; ok {
			if s != c.Error() {
				t.Errorf("Error on %T(%#v), got %v, expected %v", c, c, c.Error(), s)
			}
		} else {
			if got, panicked := panics(func() interface{} { return c.Error() }); !panicked {
				t.Errorf("Expected panic for %T(%#v), got %v", c, c, got)
			}
		}
	}
}
