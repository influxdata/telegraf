package filecount

import (
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
  ##   deprecated in 1.9; use the directories option
  directory = "/var/cache/apt/archives"

  ## Directories to gather stats about.
  ## This accept standard unit glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**    -> recursively find all directories in /var/log and count files in each directories
  ##   /var/log/*/*   -> find all directories with a parent dir in /var/log and count files in each directories
  ##   /var/log       -> count all files in /var/log and all of its subdirectories
  directories = ["/var/cache/apt/archives"]

  ## Only count files that match the name pattern. Defaults to "*".
  name = "*.deb"

  ## Count files in subdirectories. Defaults to true.
  recursive = false

  ## Only count regular files. Defaults to true.
  regular_only = true

  ## Only count files that are at least this size. If size is
  ## a negative number, only count files that are smaller than the
  ## absolute value of size. Acceptable units are B, KiB, MiB, KB, ...
  ## Without quotes and units, interpreted as size in bytes.
  size = "0B"

  ## Only count files that have not been touched for at least this
  ## duration. If mtime is negative, only count files that have been
  ## touched in this duration. Defaults to "0s".
  mtime = "0s"
`

type FileCount struct {
	Directory   string // deprecated in 1.9
	Directories []string
	Name        string
	Recursive   bool
	RegularOnly bool
	Size        internal.Size
	MTime       internal.Duration `toml:"mtime"`
	fileFilters []fileFilterFunc
	globPaths   []globpath.GlobPath
}

func (_ *FileCount) Description() string {
	return "Count files in a directory"
}

func (_ *FileCount) SampleConfig() string { return sampleConfig }

type fileFilterFunc func(os.FileInfo) (bool, error)

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
	if fc.Size.Size == 0 {
		return nil
	}

	return func(f os.FileInfo) (bool, error) {
		if !f.Mode().IsRegular() {
			return false, nil
		}
		if fc.Size.Size < 0 {
			return f.Size() < -fc.Size.Size, nil
		}
		return f.Size() >= fc.Size.Size, nil
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

func (fc *FileCount) count(acc telegraf.Accumulator, basedir string, glob globpath.GlobPath) (int64, int64) {
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
		if file.IsDir() && (fc.Recursive || glob.HasSuperMeta) {
			nf, ts = fc.count(acc, basedir+string(os.PathSeparator)+file.Name(), glob)
			if fc.Recursive {
				numFiles += nf
				totalSize += ts
			}
		}
		match, err := fc.filter(file)
		if err != nil {
			acc.AddError(err)
		}
		if match {
			numFiles++
			totalSize += file.Size()
		}
	}

	if glob.MatchString(basedir) {
		gauge := map[string]interface{}{
			"count":      numFiles,
			"size_bytes": totalSize,
		}
		acc.AddGauge("filecount", gauge,
			map[string]string{
				"directory": basedir,
			})
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
	if fc.globPaths == nil {
		fc.globPaths = fc.getGlobPaths(acc)
	}

	for _, glob := range fc.globPaths {
		for _, dir := range onlyDirectories(glob.GetRoots()) {
			fc.count(acc, dir, glob)
		}
	}

	return nil
}

func onlyDirectories(directories []string) []string {
	out := make([]string, 0)
	for _, path := range directories {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			out = append(out, path)
		}
	}
	return out
}

func (fc *FileCount) getDirs() []string {
	dirs := make([]string, len(fc.Directories))
	for i, dir := range fc.Directories {
		dirs[i] = dir
	}

	if fc.Directory != "" {
		dirs = append(dirs, fc.Directory)
	}

	return dirs
}

func (fc *FileCount) getGlobPaths(acc telegraf.Accumulator) []globpath.GlobPath {
	globPaths := []globpath.GlobPath{}
	for _, directory := range fc.getDirs() {
		glob, err := globpath.Compile(directory)
		if err != nil {
			acc.AddError(err)
		} else {
			globPaths = append(globPaths, *glob)
		}
	}

	return globPaths
}

func NewFileCount() *FileCount {
	return &FileCount{
		Directory:   "",
		Directories: []string{},
		Name:        "*",
		Recursive:   true,
		RegularOnly: true,
		Size:        internal.Size{Size: 0},
		MTime:       internal.Duration{Duration: 0},
		fileFilters: nil,
	}
}

func init() {
	inputs.Add("filecount", func() telegraf.Input {
		return NewFileCount()
	})
}
