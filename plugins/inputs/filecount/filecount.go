package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileCount struct {
	Directory   string
	Name        string
	Recursive   bool
	RegularOnly bool
	Size        int64
	MTime       int64
	fileFilters []fileFilterFunc
}

type findFunc func(os.FileInfo)
type fileFilterFunc func(os.FileInfo) bool

func logError(err error) {
	log.Println(err)
}

func rejectNilFilters(filters []fileFilterFunc) []fileFilterFunc {
	filtered := make([]fileFilterFunc, 0, len(filters))
	for _, f := range filters {
		if f != nil {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func readdir(directory string) ([]os.FileInfo, error) {
	f, err := os.Open(directory)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (fc *FileCount) nameFilter() fileFilterFunc {
	if fc.Name == "*" {
		return nil
	}

	return func(f os.FileInfo) bool {
		nameMatch, err := filepath.Match(fc.Name, f.Name())
		if err != nil {
			logError(err)
			return false
		}
		return nameMatch
	}
}

func (fc *FileCount) regularOnlyFilter() fileFilterFunc {
	if !fc.RegularOnly {
		return nil
	}

	return func(f os.FileInfo) bool {
		return f.Mode().IsRegular()
	}
}

func (fc *FileCount) sizeFilter() fileFilterFunc {
	if fc.Size == 0 {
		return nil
	}

	return func(f os.FileInfo) bool {
		if !f.Mode().IsRegular() {
			return false
		}
		if fc.Size < 0 {
			return f.Size() < -fc.Size
		}
		return f.Size() >= fc.Size
	}
}

func (fc *FileCount) mtimeFilter() fileFilterFunc {
	if fc.MTime == 0 {
		return nil
	}

	return func(f os.FileInfo) bool {
		age := time.Duration(absInt(fc.MTime)) * time.Second
		mtime := time.Now().Add(-age)
		if fc.MTime < 0 {
			return f.ModTime().After(mtime)
		}
		return f.ModTime().Before(mtime)
	}
}

func absInt(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func find(directory string, recursive bool, ff findFunc) error {
	files, err := readdir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := filepath.Join(directory, file.Name())

		if recursive && file.IsDir() {
			err = find(path, recursive, ff)
			if err != nil {
				return err
			}
		}

		ff(file)
	}
	return nil
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

func (fc *FileCount) filter(file os.FileInfo) bool {
	if fc.fileFilters == nil {
		fc.initFileFilters()
	}

	for _, fileFilter := range fc.fileFilters {
		if !fileFilter(file) {
			return false
		}
	}

	return true
}

func (fc *FileCount) count() int {
	numFiles := int64(0)
	ff := func(f os.FileInfo) {
		if !fc.filter(f) {
			return
		}
		numFiles++
	}
	err := find(fc.Directory, fc.Recursive, ff)
	if err != nil {
		logError(err)
	}
	return numFiles
}

func main() {
	for _, dir := range os.Args[1:] {
		fc := &FileCount{
			Directory:   dir,
			Name:        "*",
			Recursive:   true,
			RegularOnly: true,
			Size:        0,
			MTime:       0,
			fileFilters: nil,
		}
		fmt.Printf("%v: %v\n", dir, fc.count())
	}
}
