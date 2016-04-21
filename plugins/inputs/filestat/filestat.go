package filestat

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sepStr = fmt.Sprintf("%v", string(os.PathSeparator))

const sampleConfig = `
  ## Files to gather stats about.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". See https://github.com/gobwas/glob.
  ["/etc/telegraf/telegraf.conf", "/var/log/**.log"]
  ## If true, read the entire file and calculate an md5 checksum.
  md5 = false
`

type FileStat struct {
	Md5   bool
	Files []string

	// maps full file paths to glob obj
	globs map[string]glob.Glob
	// maps full file paths to their root dir
	roots map[string]string
}

func NewFileStat() *FileStat {
	return &FileStat{
		globs: make(map[string]glob.Glob),
		roots: make(map[string]string),
	}
}

func (_ *FileStat) Description() string {
	return "Read stats about given file(s)"
}

func (_ *FileStat) SampleConfig() string { return sampleConfig }

func (f *FileStat) Gather(acc telegraf.Accumulator) error {
	var errS string
	var err error

	for _, filepath := range f.Files {
		// Get the compiled glob object for this filepath
		g, ok := f.globs[filepath]
		if !ok {
			if g, err = glob.Compile(filepath, os.PathSeparator); err != nil {
				errS += err.Error() + " "
				continue
			}
			f.globs[filepath] = g
		}
		// Get the root directory for this filepath
		root, ok := f.roots[filepath]
		if !ok {
			root = findRootDir(filepath)
			f.roots[filepath] = root
		}

		var matches []string
		// Do not walk file tree if we don't have to.
		if !hasMeta(filepath) {
			matches = []string{filepath}
		} else {
			matches = walkFilePath(f.roots[filepath], f.globs[filepath])
		}
		for _, file := range matches {
			tags := map[string]string{
				"file": file,
			}
			fields := map[string]interface{}{
				"exists": int64(0),
			}
			// Get file stats
			fileInfo, err := os.Stat(file)
			if os.IsNotExist(err) {
				// file doesn't exist, so move on to the next
				acc.AddFields("filestat", fields, tags)
				continue
			}
			if err != nil {
				errS += err.Error() + " "
				continue
			}

			// file exists and no errors encountered
			fields["exists"] = int64(1)
			fields["size_bytes"] = fileInfo.Size()

			if f.Md5 {
				md5, err := getMd5(file)
				if err != nil {
					errS += err.Error() + " "
				} else {
					fields["md5_sum"] = md5
				}
			}

			acc.AddFields("filestat", fields, tags)
		}
	}

	if errS != "" {
		return fmt.Errorf(errS)
	}
	return nil
}

// walk the filepath from the given root and return a list of files that match
// the given glob.
func walkFilePath(root string, g glob.Glob) []string {
	matchedFiles := []string{}
	walkfn := func(path string, _ os.FileInfo, _ error) error {
		if g.Match(path) {
			matchedFiles = append(matchedFiles, path)
		}
		return nil
	}
	filepath.Walk(root, walkfn)
	return matchedFiles
}

// Read given file and calculate an md5 hash.
func getMd5(file string) (string, error) {
	of, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer of.Close()

	hash := md5.New()
	_, err = io.Copy(hash, of)
	if err != nil {
		// fatal error
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// find the root dir of the given path (could include globs).
// ie:
//   /var/log/telegraf.conf -> /var/log/
//   /home/** ->               /home/
//   /home/*/** ->             /home/
//   /lib/share/*/*/**.txt ->  /lib/share/
func findRootDir(path string) string {
	pathItems := strings.Split(path, sepStr)
	outpath := sepStr
	for i, item := range pathItems {
		if i == len(pathItems)-1 {
			break
		}
		if item == "" {
			continue
		}
		if hasMeta(item) {
			break
		}
		outpath += item + sepStr
	}
	return outpath
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(path string) bool {
	return strings.IndexAny(path, "*?[") >= 0
}

func init() {
	inputs.Add("filestat", func() telegraf.Input {
		return NewFileStat()
	})
}
