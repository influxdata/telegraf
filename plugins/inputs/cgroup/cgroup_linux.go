// +build linux

package cgroup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/influxdata/telegraf"
)

const metricName = "cgroup"

func (g *CGroup) Gather(acc telegraf.Accumulator) error {
	list := make(chan pathInfo)
	go g.generateDirs(list)

	for dir := range list {
		if dir.err != nil {
			acc.AddError(dir.err)
			continue
		}
		if err := g.gatherDir(dir.path, acc); err != nil {
			acc.AddError(err)
		}
	}

	return nil
}

func (g *CGroup) gatherDir(dir string, acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})

	list := make(chan pathInfo)
	go g.generateFiles(dir, list)

	for file := range list {
		if file.err != nil {
			return file.err
		}

		raw, err := ioutil.ReadFile(file.path)
		if err != nil {
			return err
		}
		if len(raw) == 0 {
			continue
		}

		fd := fileData{data: raw, path: file.path}
		if err := fd.parse(fields); err != nil {
			return err
		}
	}

	tags := map[string]string{"path": dir}

	acc.AddFields(metricName, fields, tags)

	return nil
}

// ======================================================================

type pathInfo struct {
	path string
	err  error
}

func isDir(path string) (bool, error) {
	result, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return result.IsDir(), nil
}

func (g *CGroup) generateDirs(list chan<- pathInfo) {
	defer close(list)
	for _, dir := range g.Paths {
		// getting all dirs that match the pattern 'dir'
		items, err := filepath.Glob(dir)
		if err != nil {
			list <- pathInfo{err: err}
			return
		}

		for _, item := range items {
			ok, err := isDir(item)
			if err != nil {
				list <- pathInfo{err: err}
				return
			}
			// supply only dirs
			if ok {
				list <- pathInfo{path: item}
			}
		}
	}
}

func (g *CGroup) generateFiles(dir string, list chan<- pathInfo) {
	defer close(list)
	for _, file := range g.Files {
		// getting all file paths that match the pattern 'dir + file'
		// path.Base make sure that file variable does not contains part of path
		items, err := filepath.Glob(path.Join(dir, path.Base(file)))
		if err != nil {
			list <- pathInfo{err: err}
			return
		}

		for _, item := range items {
			ok, err := isDir(item)
			if err != nil {
				list <- pathInfo{err: err}
				return
			}
			// supply only files not dirs
			if !ok {
				list <- pathInfo{path: item}
			}
		}
	}
}

// ======================================================================

type fileData struct {
	data []byte
	path string
}

func (fd *fileData) format() (*fileFormat, error) {
	for _, ff := range fileFormats {
		ok, err := ff.match(fd.data)
		if err != nil {
			return nil, err
		}
		if ok {
			return &ff, nil
		}
	}

	return nil, fmt.Errorf("%v: unknown file format", fd.path)
}

func (fd *fileData) parse(fields map[string]interface{}) error {
	format, err := fd.format()
	if err != nil {
		return err
	}

	format.parser(filepath.Base(fd.path), fields, fd.data)
	return nil
}

// ======================================================================

type fileFormat struct {
	name    string
	pattern string
	parser  func(measurement string, fields map[string]interface{}, b []byte)
}

const keyPattern = "[[:alpha:]_]+"
const valuePattern = "[\\d-]+"

var fileFormats = [...]fileFormat{
	// 	VAL\n
	{
		name:    "Single value",
		pattern: "^" + valuePattern + "\n$",
		parser: func(measurement string, fields map[string]interface{}, b []byte) {
			re := regexp.MustCompile("^(" + valuePattern + ")\n$")
			matches := re.FindAllStringSubmatch(string(b), -1)
			fields[measurement] = numberOrString(matches[0][1])
		},
	},
	// 	VAL0\n
	// 	VAL1\n
	// 	...
	{
		name:    "New line separated values",
		pattern: "^(" + valuePattern + "\n){2,}$",
		parser: func(measurement string, fields map[string]interface{}, b []byte) {
			re := regexp.MustCompile("(" + valuePattern + ")\n")
			matches := re.FindAllStringSubmatch(string(b), -1)
			for i, v := range matches {
				fields[measurement+"."+strconv.Itoa(i)] = numberOrString(v[1])
			}
		},
	},
	// 	VAL0 VAL1 ...\n
	{
		name:    "Space separated values",
		pattern: "^(" + valuePattern + " )+\n$",
		parser: func(measurement string, fields map[string]interface{}, b []byte) {
			re := regexp.MustCompile("(" + valuePattern + ") ")
			matches := re.FindAllStringSubmatch(string(b), -1)
			for i, v := range matches {
				fields[measurement+"."+strconv.Itoa(i)] = numberOrString(v[1])
			}
		},
	},
	// 	KEY0 VAL0\n
	// 	KEY1 VAL1\n
	// 	...
	{
		name:    "New line separated key-space-value's",
		pattern: "^(" + keyPattern + " " + valuePattern + "\n)+$",
		parser: func(measurement string, fields map[string]interface{}, b []byte) {
			re := regexp.MustCompile("(" + keyPattern + ") (" + valuePattern + ")\n")
			matches := re.FindAllStringSubmatch(string(b), -1)
			for _, v := range matches {
				fields[measurement+"."+v[1]] = numberOrString(v[2])
			}
		},
	},
}

func numberOrString(s string) interface{} {
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return i
	}

	return s
}

func (f fileFormat) match(b []byte) (bool, error) {
	ok, err := regexp.Match(f.pattern, b)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return false, nil
}
