package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Replay struct {
	Files      []string `toml:"files"`
	FileTag    string   `toml:"file_tag"`
	Iterations int      `toml:"iterations"`
	parser     parsers.Parser

	filenames []string
}

const sampleConfig = `
  ## Files to parse each interval.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/data/**.csv     -> recursively find all csv files in /var/data
  ##   /var/data/*/*.csv    -> find all .csv files with a parent dir in /var/data
  ##   /var/data/replay.csv -> only replay "replay.csv"
  files = ["/var/data/**.csv"]

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "csv"

  ## CSV Specific configuration options
  csv_header_row_count = 1
  csv_timestamp_column = "time"
  csv_timestamp_format = "unix_ns"
  csv_measurement_column = "name"
  csv_trim_space = true
  
  ## How many times to iterate through the file. -1 to continually replay the 
  ## file over and over again
  iterations = -1

  ## Might be useful if the csv file has more data than you need
  # fielddrop = [ "column1", "column2" ]

  ## Name a tag containing the name of the file the data was parsed from.  Leave empty
  ## to disable.
  # file_tag = ""
`

// SampleConfig returns the default configuration of the Input
func (r *Replay) SampleConfig() string {
	return sampleConfig
}

func (r *Replay) Description() string {
	return "Reload and gather from file[s] on telegraf's interval."
}

func (r *Replay) Init() error {
	return nil
}

func (r *Replay) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (r *Replay) Start(acc telegraf.Accumulator) error {
	err := r.refreshFilePaths()
	if err != nil {
		return err
	}
	for _, k := range r.filenames {
		metrics, err := r.readMetrics(k)
		if err != nil {
			return err
		}

		go r.processMetrics(metrics, acc)
	}

	return nil
}

func (r *Replay) Stop() {

}

func (r *Replay) SetParser(p parsers.Parser) {
	r.parser = p
}

func (r *Replay) refreshFilePaths() error {
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
		allFiles = append(allFiles, files...)
	}

	r.filenames = allFiles
	return nil
}

func (r *Replay) readMetrics(filename string) ([]telegraf.Metric, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("E! Error file: %v could not be read, %s", filename, err)
	}

	metrics, err := r.parser.Parse(fileContents)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *Replay) processMetrics(metrics []telegraf.Metric, acc telegraf.Accumulator) {
	for i := 0; i != r.Iterations; i++ {
		prevTime := metrics[0].Time()
		for _, metric := range metrics {
			currTime := metric.Time()
			delay := currTime.Sub(prevTime)
			time.Sleep(delay)
			prevTime = currTime
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags())
		}
	}
}

func init() {
	inputs.Add("replay", func() telegraf.Input {
		return &Replay{}
	})
}
