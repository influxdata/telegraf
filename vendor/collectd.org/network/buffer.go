package network // import "collectd.org/network"

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
	"time"

	"collectd.org/api"
	"collectd.org/cdtime"
)

// ErrNotEnoughSpace is returned when adding a ValueList would exeed the buffer
// size.
var ErrNotEnoughSpace = errors.New("not enough space")

// ErrUnknownType is returned when attempting to write values of an unknown type
var ErrUnknownType = errors.New("unknown type")

// Buffer contains the binary representation of multiple ValueLists and state
// optimally write the next ValueList.
type Buffer struct {
	lock               *sync.Mutex
	buffer             *bytes.Buffer
	output             io.Writer
	state              api.ValueList
	size               int
	username, password string
	securityLevel      SecurityLevel
}

// NewBuffer initializes a new Buffer. If "size" is 0, DefaultBufferSize will
// be used.
func NewBuffer(size int) *Buffer {
	if size <= 0 {
		size = DefaultBufferSize
	}

	return &Buffer{
		lock:   new(sync.Mutex),
		buffer: new(bytes.Buffer),
		size:   size,
	}
}

// Sign enables cryptographic signing of data.
func (b *Buffer) Sign(username, password string) {
	b.username = username
	b.password = password
	b.securityLevel = Sign
}

// Encrypt enables encryption of data.
func (b *Buffer) Encrypt(username, password string) {
	b.username = username
	b.password = password
	b.securityLevel = Encrypt
}

// Available returns the number of bytes still available in the buffer.
func (b *Buffer) Available() int {
	var overhead int
	switch b.securityLevel {
	case Sign:
		overhead = 36 + len(b.username)
	case Encrypt:
		overhead = 42 + len(b.username)
	}

	unavail := overhead + b.buffer.Len()
	if b.size < unavail {
		return 0
	}
	return b.size - unavail
}

// Bytes returns the content of the buffer as a byte slice.
// If signing or encrypting are enabled, the content will be signed / encrypted
// prior to being returned.
// This method resets the buffer.
func (b *Buffer) Bytes() ([]byte, error) {
	tmp := make([]byte, b.size)

	n, err := b.Read(tmp)
	if err != nil {
		return nil, err
	}

	return tmp[:n], nil
}

// Read reads the buffer into "out". If signing or encryption is enabled, data
// will be signed / encrypted before writing it to "out". Returns
// ErrNotEnoughSpace if the provided buffer is too small to hold the entire
// packet data.
func (b *Buffer) Read(out []byte) (int, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	switch b.securityLevel {
	case Sign:
		return b.readSigned(out)
	case Encrypt:
		return b.readEncrypted(out)
	}

	if len(out) < b.buffer.Len() {
		return 0, ErrNotEnoughSpace
	}

	n := copy(out, b.buffer.Bytes())

	b.reset()
	return n, nil
}

func (b *Buffer) readSigned(out []byte) (int, error) {
	if len(out) < 36+len(b.username)+b.buffer.Len() {
		return 0, ErrNotEnoughSpace
	}

	signed := signSHA256(b.buffer.Bytes(), b.username, b.password)

	b.reset()
	return copy(out, signed), nil
}

func (b *Buffer) readEncrypted(out []byte) (int, error) {
	if len(out) < 42+len(b.username)+b.buffer.Len() {
		return 0, ErrNotEnoughSpace
	}

	ciphertext, err := encryptAES256(b.buffer.Bytes(), b.username, b.password)
	if err != nil {
		return 0, err
	}

	b.reset()
	return copy(out, ciphertext), nil
}

// WriteTo writes the buffer contents to "w". It implements the io.WriteTo
// interface.
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	tmp := make([]byte, b.size)

	n, err := b.Read(tmp)
	if err != nil {
		return 0, err
	}

	n, err = w.Write(tmp[:n])
	return int64(n), err
}

// Write adds a ValueList to the buffer. Returns ErrNotEnoughSpace if not
// enough space in the buffer is available to add this value list. In that
// case, call Read() to empty the buffer and try again.
func (b *Buffer) Write(_ context.Context, vl *api.ValueList) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	// remember the original buffer size so we can truncate all potentially
	// written data in case of an error.
	l := b.buffer.Len()

	if err := b.writeValueList(vl); err != nil {
		if l != 0 {
			b.buffer.Truncate(l)
		}
		return err
	}
	return nil
}

func (b *Buffer) writeValueList(vl *api.ValueList) error {
	if err := b.writeIdentifier(vl.Identifier); err != nil {
		return err
	}

	if err := b.writeTime(vl.Time); err != nil {
		return err
	}

	if err := b.writeInterval(vl.Interval); err != nil {
		return err
	}

	if err := b.writeValues(vl.Values); err != nil {
		return err
	}

	return nil
}

func (b *Buffer) writeIdentifier(id api.Identifier) error {
	if id.Host != b.state.Host {
		if err := b.writeString(typeHost, id.Host); err != nil {
			return err
		}
		b.state.Host = id.Host
	}
	if id.Plugin != b.state.Plugin {
		if err := b.writeString(typePlugin, id.Plugin); err != nil {
			return err
		}
		b.state.Plugin = id.Plugin
	}
	if id.PluginInstance != b.state.PluginInstance {
		if err := b.writeString(typePluginInstance, id.PluginInstance); err != nil {
			return err
		}
		b.state.PluginInstance = id.PluginInstance
	}
	if id.Type != b.state.Type {
		if err := b.writeString(typeType, id.Type); err != nil {
			return err
		}
		b.state.Type = id.Type
	}
	if id.TypeInstance != b.state.TypeInstance {
		if err := b.writeString(typeTypeInstance, id.TypeInstance); err != nil {
			return err
		}
		b.state.TypeInstance = id.TypeInstance
	}

	return nil
}

func (b *Buffer) writeTime(t time.Time) error {
	if b.state.Time == t {
		return nil
	}
	b.state.Time = t

	return b.writeInt(typeTimeHR, uint64(cdtime.New(t)))
}

func (b *Buffer) writeInterval(d time.Duration) error {
	if b.state.Interval == d {
		return nil
	}
	b.state.Interval = d

	return b.writeInt(typeIntervalHR, uint64(cdtime.NewDuration(d)))
}

func (b *Buffer) writeValues(values []api.Value) error {
	size := 6 + 9*len(values)
	if size > b.Available() {
		return ErrNotEnoughSpace
	}

	binary.Write(b.buffer, binary.BigEndian, uint16(typeValues))
	binary.Write(b.buffer, binary.BigEndian, uint16(size))
	binary.Write(b.buffer, binary.BigEndian, uint16(len(values)))

	for _, v := range values {
		switch v.(type) {
		case api.Gauge:
			binary.Write(b.buffer, binary.BigEndian, uint8(dsTypeGauge))
		case api.Derive:
			binary.Write(b.buffer, binary.BigEndian, uint8(dsTypeDerive))
		case api.Counter:
			binary.Write(b.buffer, binary.BigEndian, uint8(dsTypeCounter))
		default:
			return ErrUnknownType
		}
	}

	for _, v := range values {
		switch v := v.(type) {
		case api.Gauge:
			if math.IsNaN(float64(v)) {
				b.buffer.Write([]byte{0, 0, 0, 0, 0, 0, 0xf8, 0x7f})
			} else {
				// sic: floats are encoded in little endian.
				binary.Write(b.buffer, binary.LittleEndian, float64(v))
			}
		case api.Derive:
			binary.Write(b.buffer, binary.BigEndian, int64(v))
		case api.Counter:
			binary.Write(b.buffer, binary.BigEndian, uint64(v))
		default:
			return ErrUnknownType
		}
	}

	return nil
}

func (b *Buffer) writeString(typ uint16, s string) error {
	encoded := bytes.NewBufferString(s)
	encoded.Write([]byte{0})

	// Because s is a Unicode string, encoded.Len() may be larger than
	// len(s).
	size := 4 + encoded.Len()
	if size > b.Available() {
		return ErrNotEnoughSpace
	}

	binary.Write(b.buffer, binary.BigEndian, typ)
	binary.Write(b.buffer, binary.BigEndian, uint16(size))
	b.buffer.Write(encoded.Bytes())

	return nil
}

func (b *Buffer) writeInt(typ uint16, n uint64) error {
	size := 12
	if size > b.Available() {
		return ErrNotEnoughSpace
	}

	binary.Write(b.buffer, binary.BigEndian, typ)
	binary.Write(b.buffer, binary.BigEndian, uint16(size))
	binary.Write(b.buffer, binary.BigEndian, n)

	return nil
}

func (b *Buffer) reset() {
	b.buffer.Reset()
	b.state = api.ValueList{}
}

/*
func (b *Buffer) flush() error {
	if b.buffer.Len() == 0 {
		return nil
	}

	buf := make([]byte, b.buffer.Len())
	if _, err := b.buffer.Read(buf); err != nil {
		return err
	}

	if b.username != "" && b.password != "" {
		if b.encrypt {
			var err error
			if buf, err = encryptAES256(buf, b.username, b.password); err != nil {
				return err
			}
		} else {
			buf = signSHA256(buf, b.username, b.password)
		}
	}

	if _, err := b.output.Write(buf); err != nil {
		return err
	}

	// zero state
	b.state = api.ValueList{}
	return nil
}
*/
