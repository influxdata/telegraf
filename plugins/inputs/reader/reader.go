package reader

import (
	"io/ioutil"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Reader struct {
	Filepaths     []string `toml:"files"`
	FromBeginning bool
	DataFormat    string `toml:"data_format"`
	ParserConfig  parsers.Config
	Parser        parsers.Parser
	Tags          []string

	Filenames []string

	//for grok parser
	Patterns           []string
	namedPatterns      []string
	CustomPatterns     string
	CustomPatternFiles []string
	TZone              string
}

const sampleConfig = `## Files to parse.
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

## Parse logstash-style "grok" patterns:
##   Telegraf built-in parsing patterns: https://goo.gl/dkay10
[inputs.logparser.grok]
  ## This is a list of patterns to check the given log file(s) for.
  ## Note that adding patterns here increases processing time. The most
  ## efficient configuration is to have one pattern per logparser.
  ## Other common built-in patterns are:
  ##   %{COMMON_LOG_FORMAT}   (plain apache & nginx access logs)
  ##   %{COMBINED_LOG_FORMAT} (access logs + referrer & agent)
  patterns = ["%{COMBINED_LOG_FORMAT}"]

  ## Name of the outputted measurement name.
  measurement = "apache_access_log"

  ## Full path(s) to custom pattern files.
  custom_pattern_files = []

  ## Custom patterns can also be defined here. Put one pattern per line.
  custom_patterns = '''
  '''

  ## Timezone allows you to provide an override for timestamps that
  ## don't already include an offset
  ## e.g. 04/06/2016 12:41:45 data one two 5.43µs
  ##
  ## Default: "" which renders UTC
  ## Options are as follows:
  ##   1. Local             -- interpret based on machine localtime
  ##   2. "Canada/Eastern"  -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  ##   3. UTC               -- or blank/unspecified, will return timestamp in UTC
  timezone = "Canada/Eastern"
`

// SampleConfig returns the default configuration of the Input
func (r *Reader) SampleConfig() string {
	return sampleConfig
}

func (r *Reader) Description() string {
	return "reload and gather from file[s] on telegraf's interval"
}

func (r *Reader) Gather(acc telegraf.Accumulator) error {
	r.refreshFilePaths()
	for _, k := range r.Filenames {
		metrics, err := r.readMetric(k)
		if err != nil {
			return err
		}

		for _, m := range metrics {
			acc.AddFields(m.Name(), m.Fields(), m.Tags())
		}
	}
	return nil
}

func (r *Reader) SetParser(p parsers.Parser) {
	r.Parser = p
}

func (r *Reader) compileParser() {
	if r.DataFormat == "" {
		log.Printf("E! No data_format specified")
		return
	}
	r.ParserConfig = parsers.Config{
		DataFormat: r.DataFormat,
		TagKeys:    r.Tags,

		//grok settings
		Patterns:           r.Patterns,
		NamedPatterns:      r.namedPatterns,
		CustomPatterns:     r.CustomPatterns,
		CustomPatternFiles: r.CustomPatternFiles,
		TimeZone:           r.TZone,
	}
	nParser, err := parsers.NewParser(&r.ParserConfig)
	if err != nil {
		log.Printf("E! Error building parser: %v", err)
	}

	r.Parser = nParser
}

func (r *Reader) refreshFilePaths() {
	var allFiles []string
	for _, filepath := range r.Filepaths {
		g, err := globpath.Compile(filepath)
		if err != nil {
			log.Printf("E! Error Glob %s failed to compile, %s", filepath, err)
			continue
		}
		files := g.Match()

		for k := range files {
			allFiles = append(allFiles, k)
		}
	}

	r.Filenames = allFiles
}

//requires that Parser has been compiled
func (r *Reader) readMetric(filename string) ([]telegraf.Metric, error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("E! File could not be opened: %v", filename)
	}

	return r.Parser.Parse(fileContents)

}
