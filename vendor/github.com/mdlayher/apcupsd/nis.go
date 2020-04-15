package apcupsd

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
)

var _ io.ReadWriteCloser = &nisReadWriteCloser{}

// newNISReadWriteCloser wraps an io.ReadWriteCloser.
func newNISReadWriteCloser(rwc io.ReadWriteCloser) *nisReadWriteCloser {
	return &nisReadWriteCloser{
		rwc:  rwc,
		lenb: make([]byte, 2),
	}
}

// An nisReadWriteCloser wraps an io.ReadWriteCloser with one that can encode
// and decode messages using the NIS's protocol.
type nisReadWriteCloser struct {
	mu   sync.Mutex
	rwc  io.ReadWriteCloser
	lenb []byte
}

// Read reads messages from the NIS using its protocol:
//  - 2 bytes: length of next message
//  - N bytes: data
func (rwc *nisReadWriteCloser) Read(b []byte) (int, error) {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()

	// Read two byte length of next data
	if _, err := io.ReadFull(rwc.rwc, rwc.lenb); err != nil {
		return 0, err
	}

	// When no more data returned from server, return io.EOF
	length := binary.BigEndian.Uint16(rwc.lenb)
	if length == 0 {
		return 0, io.EOF
	}

	return io.ReadFull(rwc.rwc, b[:length])
}

var (
	// errBufferTooLarge indicates that nisReadWriteCloser.Write was passed a
	// buffer that is too large to send to the NIS.
	errBufferTooLarge = errors.New("buffer too large; must be size of uint16 or less")
)

// Write writes messages to the NIS using its protocol by prepending each
// message with its 2 byte length.
func (rwc *nisReadWriteCloser) Write(b []byte) (int, error) {
	// Cannot write more than math.MaxUint16 bytes
	if len(b) > math.MaxUint16 {
		return 0, errBufferTooLarge
	}

	rwc.mu.Lock()
	defer rwc.mu.Unlock()

	// Two byte length of data
	binary.BigEndian.PutUint16(rwc.lenb, uint16(len(b)))

	// Send data and indicate the length of the body to caller
	n, err := rwc.rwc.Write(append(rwc.lenb, b...))
	n -= len(rwc.lenb)
	return n, err
}

// Close closes the underlying io.ReadWriteCloser.
func (rwc *nisReadWriteCloser) Close() error {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()

	return rwc.rwc.Close()
}
