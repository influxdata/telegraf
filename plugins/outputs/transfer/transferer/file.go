package transferer

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type FileTransferer struct {
	BaseTransferer
}

func NewFileTransferer() *FileTransferer {
	return &FileTransferer{}
}

func (f *FileTransferer) Rename(from *url.URL, to string) error {
	return nil
}

func (f *FileTransferer) Send(source string, dest *url.URL) error {
	// open files r and w
	var r *os.File
	var err error

	// Check for windows paths. Both file:///c:/path/to/file.txt and
	// file://c:/path/to/file.txt will work.
	if found, _ := regexp.MatchString("\\/[A-Z]\\:\\/", dest.Path); found {
		// Remove the first character from the path
		dest.Path = dest.Path[1:]
	} else if len(dest.Host) == 2 && dest.Host[1] == ':' {
		dest.Path = dest.Host + dest.Path
		dest.Host = ""
	}

	for i := 0; i < 10; i++ {
		r, err = os.Open(source)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}
	if err != nil {
		return err
	}
	defer r.Close()

	dir := filepath.Dir(dest.Path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0664)
	}

	w, err := os.Create(dest.Path)
	if err != nil {
		return err
	}
	defer w.Close()

	// do the actual work
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	err = w.Sync()
	if err != nil {
		return err
	}

	return nil
}
