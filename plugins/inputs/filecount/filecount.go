package filecount

import (
	"os"
	"path/filepath"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## Directory to gather stats about.
  directory = "/var/cache/apt/archives"

  ## Only count files that match the name pattern. Defaults to "*".
  name = "*.deb"

  ## Count files in subdirectories. Defaults to true.
  recursive = false

  ## Only count regular files. Defaults to true.
  regular_only = true

  ## Only count files that are at least this size in bytes. If size is
  ## a negative number, only count files that are smaller than the
  ## absolute value of size. Defaults to 0.
  size = 0

  ## Only count files that have not been touched for at least this
  ## duration. If mtime is negative, only count files that have been
  ## touched in this duration. Defaults to "0s".
  mtime = "0s"
`

type FileCount struct {
	Directory   string
	Name        string
	Recursive   bool
	RegularOnly bool
	Size        int64
	MTime       internal.Duration `toml:"mtime"`
	fileFilters []fileFilterFunc
}

type countFunc func(os.FileInfo)
type fileFilterFunc func(os.FileInfo) (bool, error)

func (_ *FileCount) Description() string {
	return "Count files in a directory"
}

func (_ *FileCount) SampleConfig() string { return sampleConfig }

func rejectNilFilters(filters []fileFilterFunc) []fileFilterFunc {
	filtered := make([]fileFilterFunc, 0, len(filters))
	for _, f := range filters {
		if f != nil {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func (fc *FileCount) nameFilter() fileFilterFunc {
	if fc.Name == "*" {
		return nil
	}

	return func(f os.FileInfo) (bool, error) {
		match, err := filepath.Match(fc.Name, f.Name())
		if err != nil {
			return false, err
		}
		return match, nil
	}
}

func (fc *FileCount) regularOnlyFilter() fileFilterFunc {
	if !fc.RegularOnly {
		return nil
	}

	return func(f os.FileInfo) (bool, error) {
		return f.Mode().IsRegular(), nil
	}
}

func (fc *FileCount) sizeFilter() fileFilterFunc {
	if fc.Size == 0 {
		return nil
	}

	return func(f os.FileInfo) (bool, error) {
		if !f.Mode().IsRegular() {
			return false, nil
		}
		if fc.Size < 0 {
			return f.Size() < -fc.Size, nil
		}
		return f.Size() >= fc.Size, nil
	}
}

func (fc *FileCount) mtimeFilter() fileFilterFunc {
	if fc.MTime.Duration == 0 {
		return nil
	}

	return func(f os.FileInfo) (bool, error) {
		age := absDuration(fc.MTime.Duration)
		mtime := time.Now().Add(-age)
		if fc.MTime.Duration < 0 {
			return f.ModTime().After(mtime), nil
		}
		return f.ModTime().Before(mtime), nil
	}
}

func absDuration(x time.Duration) time.Duration {
	if x < 0 {
		return -x
	}
	return x
}

func count(basedir string, recursive bool, countFn countFunc) error {
	walkFn := func(path string, file os.FileInfo, err error) error {
		if path == basedir {
			return nil
		}
		countFn(file)
		if !recursive && file.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}
	return filepath.Walk(basedir, walkFn)
}

func (fc *FileCount) initFileFilters() {
	filters := []fileFilterFunc{
		fc.nameFilter(),
		fc.regularOnlyFilter(),
		fc.sizeFilter(),
		fc.mtimeFilter(),
	}
	fc.fileFilters = rejectNilFilters(filters)
}

func (fc *FileCount) filter(file os.FileInfo) (bool, error) {
	if fc.fileFilters == nil {
		fc.initFileFilters()
	}

	for _, fileFilter := range fc.fileFilters {
		match, err := fileFilter(file)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}

	return true, nil
}

func (fc *FileCount) Gather(acc telegraf.Accumulator) error {
	numFiles := int64(0)
	countFn := func(f os.FileInfo) {
		match, err := fc.filter(f)
		if err != nil {
			acc.AddError(err)
			return
		}
		if !match {
			return
		}
		numFiles++
	}
	err := count(fc.Directory, fc.Recursive, countFn)
	if err != nil {
		acc.AddError(err)
	}

	acc.AddFields("filecount",
		map[string]interface{}{
			"count": numFiles,
		},
		map[string]string{
			"directory": fc.Directory,
		})

	return nil
}

func NewFileCount() *FileCount {
	return &FileCount{
		Directory:   "",
		Name:        "*",
		Recursive:   true,
		RegularOnly: true,
		Size:        0,
		MTime:       internal.Duration{Duration: 0},
		fileFilters: nil,
	}
}

func init() {
	inputs.Add("filecount", func() telegraf.Input {
		return NewFileCount()
	})
}
