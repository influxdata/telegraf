package binaryio

import "io"

// MinimumReader is the implementation for MinReader.
type MinimumReader struct {
	R                      io.Reader
	MinNumberOfBytesToRead int64 // Min number of bytes we need to read from the reader
}

// MinReader reads from R but ensures there is at least N bytes read from the reader.
// The reader should call Close() when they are done reading.
// Closing the MinReader will read and discard any unread bytes up to MinNumberOfBytesToRead.
// CLosing the MinReader does NOT close the underlying reader.
// The underlying implementation is a MinimumReader, which implements ReaderCloser.
func MinReader(r io.Reader, minNumberOfBytesToRead int64) *MinimumReader {
	return &MinimumReader{
		R:                      r,
		MinNumberOfBytesToRead: minNumberOfBytesToRead,
	}
}

func (r *MinimumReader) Read(p []byte) (n int, err error) {
	n, err = r.R.Read(p)
	r.MinNumberOfBytesToRead -= int64(n)
	return n, err
}

// Close does not close the underlying reader, only the MinimumReader
func (r *MinimumReader) Close() error {
	if r.MinNumberOfBytesToRead > 0 {
		b := make([]byte, r.MinNumberOfBytesToRead)
		_, err := r.R.Read(b)
		return err
	}
	return nil
}
