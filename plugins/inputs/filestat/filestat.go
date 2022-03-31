package filestat

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type FileStat struct {
	Md5   bool
	Files []string

	Log telegraf.Logger

	// maps full file paths to globmatch obj
	globs map[string]*globpath.GlobPath

	// files that were missing - we only log the first time it's not found.
	missingFiles map[string]bool
	// files that had an error in Stat - we only log the first error.
	filesWithErrors map[string]bool
}

func NewFileStat() *FileStat {
	return &FileStat{
		globs:           make(map[string]*globpath.GlobPath),
		missingFiles:    make(map[string]bool),
		filesWithErrors: make(map[string]bool),
	}
}

func (f *FileStat) Gather(acc telegraf.Accumulator) error {
	var err error

	for _, filepath := range f.Files {
		// Get the compiled glob object for this filepath
		g, ok := f.globs[filepath]
		if !ok {
			if g, err = globpath.Compile(filepath); err != nil {
				acc.AddError(err)
				continue
			}
			f.globs[filepath] = g
		}

		files := g.Match()
		if len(files) == 0 {
			acc.AddFields("filestat",
				map[string]interface{}{
					"exists": int64(0),
				},
				map[string]string{
					"file": filepath,
				})
			continue
		}

		for _, fileName := range files {
			tags := map[string]string{
				"file": fileName,
			}
			fields := map[string]interface{}{
				"exists": int64(1),
			}
			fileInfo, err := os.Stat(fileName)
			if os.IsNotExist(err) {
				fields["exists"] = int64(0)
				acc.AddFields("filestat", fields, tags)
				if !f.missingFiles[fileName] {
					f.Log.Warnf("File %q not found", fileName)
					f.missingFiles[fileName] = true
				}
				continue
			}
			f.missingFiles[fileName] = false

			if fileInfo == nil {
				if !f.filesWithErrors[fileName] {
					f.filesWithErrors[fileName] = true
					f.Log.Errorf("Unable to get info for file %q: %v",
						fileName, err)
				}
			} else {
				f.filesWithErrors[fileName] = false
				fields["size_bytes"] = fileInfo.Size()
				fields["modification_time"] = fileInfo.ModTime().UnixNano()
			}

			if f.Md5 {
				md5Hash, err := getMd5(fileName)
				if err != nil {
					acc.AddError(err)
				} else {
					fields["md5_sum"] = md5Hash
				}
			}

			acc.AddFields("filestat", fields, tags)
		}
	}

	return nil
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

func init() {
	inputs.Add("filestat", func() telegraf.Input {
		return NewFileStat()
	})
}
