package file

import (
	"fmt"
	"io/ioutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Reader struct {
	Files         []string `toml:"files"`
	FromBeginning bool
	parser        parsers.Parser

	filenames []string
}

const sampleConfig = `
  ## Files to parse each interval.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**.log     -> recursively find all .log files in /var/log
  ##   /var/log/*/*.log    -> find all .log files with a parent dir in /var/log
  ##   /var/log/apache.log -> only tail the apache log file
  files = ["/var/log/apache/access.log"]

  ## The dataformat to be read from files 
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = ""
`

// SampleConfig returns the default configuration of the Input
func (r *Reader) SampleConfig() string {
	return sampleConfig
}

func (r *Reader) Description() string {
	return "reload and gather from file[s] on telegraf's interval"
}

func (r *Reader) Gather(acc telegraf.Accumulator) error {
	err := r.refreshFilePaths()
	if err != nil {
		return err
	}
	for _, k := range r.filenames {
		metrics, err := r.readMetric(k)
		if err != nil {
			return err
		}

		for _, m := range metrics {
			acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}
	return nil
}

func (r *Reader) SetParser(p parsers.Parser) {
	r.parser = p
}

func (r *Reader) refreshFilePaths() error {
	var allFiles []string
	for _, file := range r.Files {
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

	r.filenames = allFiles
	return nil
}

func (r *Reader) readMetric(filename string) ([]telegraf.Metric, error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("E! Error file: %v could not be read, %s", filename, err)
	}
	return r.parser.Parse(fileContents)

}

func init() {
	inputs.Add("reader", func() telegraf.Input {
		return &Reader{}
	})
}
