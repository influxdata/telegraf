package rotate

// Rotating things
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FilePerm defines the permissions that Writer will use for all
// the files it creates.
const (
	FilePerm   = os.FileMode(0644)
	DateFormat = "2006-01-02"
)

// FileWriter implements the io.Writer interface and writes to the
// filename specified.
// Will rotate at the specified interval and/or when the current file size exceeds maxSizeInBytes
// At rotation time, current file is renamed and a new file is created.
// If the number of archives exceeds maxArchives, older files are deleted.
type FileWriter struct {
	filename                 string
	filenameRotationTemplate string
	current                  *os.File
	interval                 time.Duration
	maxSizeInBytes           int64
	maxArchives              int
	expireTime               time.Time
	bytesWritten             int64
	sync.Mutex
}

// NewFileWriter creates a new file writer.
func NewFileWriter(filename string, interval time.Duration, maxSizeInBytes int64, maxArchives int) (io.WriteCloser, error) {
	if interval == 0 && maxSizeInBytes <= 0 {
		// No rotation needed so a basic io.Writer will do the trick
		return openFile(filename)
	}

	w := &FileWriter{
		filename:                 filename,
		interval:                 interval,
		maxSizeInBytes:           maxSizeInBytes,
		maxArchives:              maxArchives,
		filenameRotationTemplate: getFilenameRotationTemplate(filename),
	}

	if err := w.openCurrent(); err != nil {
		return nil, err
	}

	return w, nil
}

func openFile(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, FilePerm)
}

func getFilenameRotationTemplate(filename string) string {
	// Extract the file extension
	fileExt := filepath.Ext(filename)
	// Remove the file extension from the filename (if any)
	stem := strings.TrimSuffix(filename, fileExt)
	return stem + ".%s-%s" + fileExt
}

// Write writes p to the current file, then checks to see if
// rotation is necessary.
func (w *FileWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	if n, err = w.current.Write(p); err != nil {
		return 0, err
	}
	w.bytesWritten += int64(n)

	if err := w.rotateIfNeeded(); err != nil {
		return 0, err
	}

	return n, nil
}

// Close closes the current file.  Writer is unusable after this
// is called.
func (w *FileWriter) Close() (err error) {
	w.Lock()
	defer w.Unlock()

	// Rotate before closing
	if err := w.rotateIfNeeded(); err != nil {
		return err
	}

	// Close the file if we did not rotate
	if err := w.current.Close(); err != nil {
		return err
	}

	w.current = nil
	return nil
}

func (w *FileWriter) openCurrent() (err error) {
	// In case ModTime() fails, we use time.Now()
	w.expireTime = time.Now().Add(w.interval)
	w.bytesWritten = 0
	w.current, err = openFile(w.filename)

	if err != nil {
		return err
	}

	// Goal here is to rotate old pre-existing files.
	// For that we use fileInfo.ModTime, instead of time.Now().
	// Example: telegraf is restarted every 23 hours and
	// the rotation interval is set to 24 hours.
	// With time.now() as a reference we'd never rotate the file.
	if fileInfo, err := w.current.Stat(); err == nil {
		w.expireTime = fileInfo.ModTime().Add(w.interval)
		w.bytesWritten = fileInfo.Size()
	}

	return w.rotateIfNeeded()
}

func (w *FileWriter) rotateIfNeeded() error {
	if (w.interval > 0 && time.Now().After(w.expireTime)) ||
		(w.maxSizeInBytes > 0 && w.bytesWritten >= w.maxSizeInBytes) {
		if err := w.rotate(); err != nil {
			//Ignore rotation errors and keep the log open
			fmt.Printf("unable to rotate the file %q, %s", w.filename, err.Error())
		}
		return w.openCurrent()
	}
	return nil
}

func (w *FileWriter) rotate() (err error) {
	if err := w.current.Close(); err != nil {
		return err
	}

	// Use year-month-date for readability, unix time to make the file name unique with second precision
	now := time.Now()
	rotatedFilename := fmt.Sprintf(w.filenameRotationTemplate, now.Format(DateFormat), strconv.FormatInt(now.Unix(), 10))
	if err := os.Rename(w.filename, rotatedFilename); err != nil {
		return err
	}

	return w.purgeArchivesIfNeeded()
}

func (w *FileWriter) purgeArchivesIfNeeded() (err error) {
	if w.maxArchives == -1 {
		//Skip archiving
		return nil
	}

	var matches []string
	if matches, err = filepath.Glob(fmt.Sprintf(w.filenameRotationTemplate, "*", "*")); err != nil {
		return err
	}

	//if there are more archives than the configured maximum, then purge older files
	if len(matches) > w.maxArchives {
		//sort files alphanumerically to delete older files first
		sort.Strings(matches)
		for _, filename := range matches[:len(matches)-w.maxArchives] {
			if err := os.Remove(filename); err != nil {
				return err
			}
		}
	}
	return nil
}
