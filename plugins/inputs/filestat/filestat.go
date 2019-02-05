package filestat

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## Files to gather stats about.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
  ##   "/var/log/apache.log" -> just tail the apache log file
  ##
  ## See https://github.com/gobwas/glob for more examples
  ##
  files = ["/var/log/**.log"]
  ## If true, read the entire file and calculate an md5 checksum.
  md5 = false
`

type FileStat struct {
	Md5   bool
	Files []string

	// maps full file paths to globmatch obj
	globs map[string]*globpath.GlobPath
}

func NewFileStat() *FileStat {
	return &FileStat{
		globs: make(map[string]*globpath.GlobPath),
	}
}

func (_ *FileStat) Description() string {
	return "Read stats about given file(s)"
}

func (_ *FileStat) SampleConfig() string { return sampleConfig }

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
			}

			if fileInfo == nil {
				log.Printf("E! Unable to get info for file [%s], possible permissions issue",
					fileName)
			} else {
				fields["size_bytes"] = fileInfo.Size()
				fields["modification_time"] = fileInfo.ModTime().UnixNano()
			}

			if f.Md5 {
				md5, err := getMd5(fileName)
				if err != nil {
					acc.AddError(err)
				} else {
					fields["md5_sum"] = md5
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
