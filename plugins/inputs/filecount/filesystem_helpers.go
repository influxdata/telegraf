package filecount

import (
	"io"
	"os"
)

/*
	The code below is lifted from numerous articles and originates from Andrew Gerrand's 10 things you (probably) don't know about Go.
	It allows for mocking a filesystem; this allows for consistent testing of this code across platforms (directory sizes reported
	differently by different platforms, for example), while preserving the rest of the functionality as-is, without modification.
*/

type fileSystem interface {
	Open(name string) (file, error)
	Stat(name string) (os.FileInfo, error)
}

type file interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}

// osFS implements fileSystem using the local disk
type osFS struct{}

func (osFS) Open(name string) (file, error)        { return os.Open(name) }
func (osFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
