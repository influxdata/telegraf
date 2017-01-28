package client

import "io"

type Client interface {
	Query(command string) error

	Write(b []byte) (int, error)
	WriteWithParams(b []byte, params WriteParams) (int, error)

	WriteStream(b io.Reader, contentLength int) (int, error)
	WriteStreamWithParams(b io.Reader, contentLength int, params WriteParams) (int, error)

	Close() error
}

type WriteParams struct {
	Database        string
	RetentionPolicy string
	Precision       string
	Consistency     string
}
