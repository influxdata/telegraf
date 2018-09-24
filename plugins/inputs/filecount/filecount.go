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

  ## Also compute total size of matched elements. Defaults to false.
  count_size = false

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

  ## Output stats for every subdirectory. Defaults to false.
  recursive_print = false

  ## Only output directories whose sub elements weighs more than this
  ## size in bytes. Defaults to 0.
  recursive_print_size = 0
`

type FileCount struct {
	Directory          string
	CountSize          bool
	Name               string
	Recursive          bool
	RegularOnly        bool
	Size               int64
	MTime              internal.Duration `toml:"mtime"`
	RecursivePrint     bool
	RecursivePrintSize int64
	fileFilters        []fileFilterFunc
}

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

func (fc *FileCount) initFileFilters() {
	filters := []fileFilterFunc{
		fc.nameFilter(),
		fc.regularOnlyFilter(),
		fc.sizeFilter(),
		fc.mtimeFilter(),
	}
	fc.fileFilters = rejectNilFilters(filters)
}

func (fc *FileCount) count(acc telegraf.Accumulator, basedir string) (int64, int64) {
	numFiles, totalSize, nf, ts := int64(0), int64(0), int64(0), int64(0)

	directory, err := os.Open(basedir)
	if err != nil {
		acc.AddError(err)
		return numFiles, totalSize
	}
	files, err := directory.Readdir(0)
	directory.Close()
	if err != nil {
		acc.AddError(err)
		return numFiles, totalSize
	}
	for _, file := range files {
		if fc.Recursive && file.IsDir() {
			nf, ts = fc.count(acc, basedir + string(os.PathSeparator) + file.Name())
			numFiles += nf
			totalSize += ts
		}
		matches, err := fc.filter(file)
		if err != nil {
			acc.AddError(err)
		}
		if matches {
			numFiles++
			totalSize += file.Size()
		}
	}

	if fc.RecursivePrint || basedir == fc.Directory {
		if totalSize >= fc.RecursivePrintSize || basedir == fc.Directory {
			gauge := map[string]interface{}{
				"count": numFiles,
			}
			if fc.CountSize {
				gauge["size"] = totalSize
			}
			acc.AddGauge("filecount", gauge,
				map[string]string{
					"directory": basedir,
				})
		}
	}

	return numFiles, totalSize

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

	fc.count(acc, fc.Directory)

	return nil
}

func NewFileCount() *FileCount {
	return &FileCount{
		Directory:          "",
		CountSize:          false,
		Name:               "*",
		Recursive:          true,
		RegularOnly:        true,
		Size:               0,
		MTime:              internal.Duration{Duration: 0},
		RecursivePrint:     false,
		RecursivePrintSize: 0,
		fileFilters:        nil,
	}
}

func init() {
	inputs.Add("filecount", func() telegraf.Input {
		return NewFileCount()
	})
}
