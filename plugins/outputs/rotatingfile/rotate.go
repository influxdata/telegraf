package rotatingfile

// Rotating things
import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

// RootPerm defines the permissions that Writer will use if it
// needs to create the root directory.
var RootPerm = os.FileMode(0755)

// FilePerm defines the permissions that Writer will use for all
// the files it creates.
var FilePerm = os.FileMode(0644)

// Writer implements the io.Writer interface and writes to the
// "current" file in the root directory.  When current file age
// exceeds max, it is renamed and a new file is created.
type Writer struct {
	root       string
	prefix     string
	current    *os.File
	expireTime time.Time
	max        time.Duration
	sync.Mutex
}

// New creates a new Writer.  The files will be created in the
// root directory.  root will be created if necessary.  The
// filenames will start with prefix.
func NewRotatingWriter(root, prefix string, maxAgeInput string) (*Writer, error) {
	maxAge, err := time.ParseDuration(maxAgeInput)
	if err != nil {
		return nil, err
	}
	l := &Writer{root: root, prefix: prefix, max: maxAge}
	if err := l.setup(); err != nil {
		return nil, err
	}
	return l, nil
}

// Write writes p to the current file, then checks to see if
// rotation is necessary.
func (r *Writer) Write(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	n, err = r.current.Write(p)
	if err != nil {
		return n, err
	}
	if time.Now().After(r.expireTime) {
		if err := r.rotate(); err != nil {
			return n, err
		}
	}
	return n, nil
}

// Close closes the current file.  Writer is unusable after this
// is called.
func (r *Writer) Close() error {
	r.Lock()
	defer r.Unlock()

	// Rotate before closing
	if err := r.rotate(); err != nil {
		return err
	}

	if err := r.current.Close(); err != nil {
		return err
	}
	r.current = nil
	return nil
}

// setup creates the root directory if necessary, then opens the
// current file.
func (r *Writer) setup() error {
	fi, err := os.Stat(r.root)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(r.root, RootPerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !fi.IsDir() {
		return errors.New("root must be a directory")
	}

	// root exists, and it is a directory

	return r.openCurrent()
}

func (r *Writer) openCurrent() error {
	cp := path.Join(r.root, fmt.Sprintf("%s-current", r.prefix)) // It should be safe to use Sprintf here since path.Join() uses path.Clean() on the path afterwards
	var err error
	r.current, err = os.OpenFile(cp, os.O_RDWR|os.O_CREATE|os.O_APPEND, FilePerm)
	r.expireTime = time.Now().Add(r.max)
	if err != nil {
		return err
	}
	return nil
}

func (r *Writer) rotate() error {
	if err := r.current.Close(); err != nil {
		return err
	}
	filename := fmt.Sprintf("%s-%d", r.prefix, time.Now().UnixNano()) // UnixNano should be unique enough for this (up until a point)
	if err := os.Rename(path.Join(r.root, fmt.Sprintf("%s-current", r.prefix)), path.Join(r.root, filename)); err != nil {
		return err
	}
	return r.openCurrent()
}
