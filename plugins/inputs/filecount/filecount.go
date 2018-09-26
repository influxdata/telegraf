package filecount

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/globpath"
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

func (fc *FileCount) count(acc telegraf.Accumulator, basedir string, recursive bool) {
	numFiles := int64(0)
	walkFn := func(path string, file os.FileInfo, err error) error {
		if path == basedir {
			return nil
		}
		match, err := fc.filter(file)
		if err != nil {
			acc.AddError(err)
			return nil
		}
		if match {
			numFiles++
		}
		if !recursive && file.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	err := filepath.Walk(basedir, walkFn)
	if err != nil {
		acc.AddError(err)
		return
	}

	acc.AddFields("filecount",
		map[string]interface{}{
			"count": numFiles,
		},
		map[string]string{
			"directory": basedir,
		},
	)
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
	dirs, err := getTargetDirs(fc.Directory)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		fc.count(acc, dir, fc.Recursive)
	}

	return nil
}

func getTargetDirs(directory string) ([]string, error) {
	g, err := globpath.Compile(directory)
	if err != nil {
		return nil, fmt.Errorf("could not compile glob %v: %v", directory, err)
	}

	filtered := []string{}
	for path, file := range g.Match() {
		if file.IsDir() == true {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
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
