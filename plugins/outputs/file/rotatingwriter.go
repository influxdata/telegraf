package file

// Rotating things
import (
	"fmt"
	"os"
	"sync"
	"time"
)

// FilePerm defines the permissions that Writer will use for all
// the files it creates.
var FilePerm = os.FileMode(0644)

// Writer implements the io.Writer interface and writes to the
// filename specified.  When current file age exceeds max, it is
// renamed and a new file is created.
type Writer struct {
	filename   string
	current    *os.File
	expireTime time.Time
	max        time.Duration
	sync.Mutex
}

// New creates a new Writer.
func NewRotatingWriter(filename, maxAgeInput string) (*Writer, error) {
	maxAge, err := time.ParseDuration(maxAgeInput)
	if err != nil {
		return nil, err
	}
	l := &Writer{filename: filename, max: maxAge}
	if err := l.openCurrent(); err != nil {
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

func (r *Writer) openCurrent() error {
	var err error
	r.current, err = os.OpenFile(r.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, FilePerm)
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
	rotatedFilename := fmt.Sprintf("%s-%d", r.filename, time.Now().UnixNano()) // UnixNano should be unique enough for this (up until a point)
	if err := os.Rename(r.filename, rotatedFilename); err != nil {
		return err
	}
	return r.openCurrent()
}
