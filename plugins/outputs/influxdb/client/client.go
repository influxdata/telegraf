package client

import "io"

type Client interface {
	Query(command string) error
	WriteStream(b io.Reader) error
	Close() error
}

type WriteParams struct {
	Database        string
	RetentionPolicy string
	Precision       string
	Consistency     string
}
