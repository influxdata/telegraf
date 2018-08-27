package file

import (
	"fmt"
	"io/ioutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type File struct {
	Files  []string `toml:"files"`
	parser parsers.Parser

	filenames []string
}

const sampleConfig = `
  ## Files to parse each interval.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**.log     -> recursively find all .log files in /var/log
  ##   /var/log/*/*.log    -> find all .log files with a parent dir in /var/log
  ##   /var/log/apache.log -> only read the apache log file
  files = ["/var/log/apache/access.log"]

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

// SampleConfig returns the default configuration of the Input
func (f *File) SampleConfig() string {
	return sampleConfig
}

func (f *File) Description() string {
	return "Reload and gather from file[s] on telegraf's interval."
}

func (f *File) Gather(acc telegraf.Accumulator) error {
	err := f.refreshFilePaths()
	if err != nil {
		return err
	}
	for _, k := range f.filenames {
		metrics, err := f.readMetric(k)
		if err != nil {
			return err
		}

		for _, m := range metrics {
			acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}
	return nil
}

func (f *File) SetParser(p parsers.Parser) {
	f.parser = p
}

func (f *File) refreshFilePaths() error {
	var allFiles []string
	for _, file := range f.Files {
		g, err := globpath.Compile(file)
		if err != nil {
			return fmt.Errorf("could not compile glob %v: %v", file, err)
		}
		files := g.Match()
		if len(files) <= 0 {
			return fmt.Errorf("could not find file: %v", file)
		}

		for k := range files {
			allFiles = append(allFiles, k)
		}
	}

	f.filenames = allFiles
	return nil
}

func (f *File) readMetric(filename string) ([]telegraf.Metric, error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("E! Error file: %v could not be read, %s", filename, err)
	}
	return f.parser.Parse(fileContents)

}

func init() {
	inputs.Add("file", func() telegraf.Input {
		return &File{}
	})
}
