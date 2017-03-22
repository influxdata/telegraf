// +build linux

package journalparser

import (
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/logparser/grok"
)

const sampleConfig = `
[[inputs.journalparser]]
  ## Match filters to apply to journal entries. Only entries which match will be processed.
  matches = ["_COMM=httpd"]
  ## Read journal from beginning.
  from_beginning = false

  ## Parse logstash-style "grok" patterns:
  ##   Telegraf built-in parsing patterns: https://goo.gl/dkay10
  [inputs.journalparser.grok]
    ## Name of the outputted measurement name.
    measurement = "apache_access_log"
    ## Full path(s) to custom pattern files.
    custom_pattern_files = []
    ## Custom patterns can also be defined here. Put one pattern per line.
    custom_patterns = '''
    '''

    [inputs.journalparser.grok.patterns]
      ## This is a list of patterns to check the given log file(s) for.
      ## The parameter name is the journal field to apply the pattern on,
      ## typically "MESSAGE".
      ## Note that adding patterns here increases processing time. The most
      ## efficient configuration is to have one pattern per logparser.
      ## Other common built-in patterns are:
      ##   %{COMMON_LOG_FORMAT}   (plain apache & nginx access logs)
      ##   %{COMBINED_LOG_FORMAT} (access logs + referrer & agent)
      MESSAGE = ["%{COMBINED_LOG_FORMAT}"]
`
const description = ""

type JournalGrokParser struct {
	Patterns     map[string][]string
	fieldParsers map[string]*grok.Parser
	grok.Parser
}

type JournalParser struct {
	JournalPath string
	Matches     []string

	GrokParser JournalGrokParser `toml:"grok"`

	journalClient *journalClient
	sync.Mutex
}

func (jp *JournalParser) SampleConfig() string {
	return sampleConfig
}

func (jp *JournalParser) Description() string {
	return description
}

func (jp *JournalParser) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (jp *JournalParser) Start(acc telegraf.Accumulator) error {
	if len(jp.GrokParser.Patterns) == 0 {
		return fmt.Errorf("no parser fields configured")
	}

	if jp.GrokParser.Measurement == "" {
		jp.GrokParser.Measurement = "journalparser"
	}

	jp.GrokParser.fieldParsers = map[string]*grok.Parser{}
	for field, patterns := range jp.GrokParser.Patterns {
		gp := jp.GrokParser.Parser
		gp.Patterns = patterns
		if err := gp.Compile(); err != nil {
			return fmt.Errorf("error in configuration for %s: %s", field, err)
		}
		jp.GrokParser.fieldParsers[field] = &gp
	}

	jc, err := GetJournalStreamer(jp.JournalPath).NewClient(jp.Matches)
	if err != nil {
		return err
	}
	jp.journalClient = jc
	go jp.readJournal(jc.jeChan, acc)

	return nil
}

func (jp *JournalParser) Stop() error {
	GetJournalStreamer(jp.JournalPath).RemoveClient(jp.journalClient)
	return nil
}

func (jp *JournalParser) readJournal(jeChan <-chan *journalEntry, acc telegraf.Accumulator) {
	defer jp.Stop()

ENTRY:
	for je := range jeChan {
		var measurement string
		tags := map[string]string{}
		fields := map[string]interface{}{}
		timestamp := je.time

		for fk, gp := range jp.GrokParser.fieldParsers {
			fv, ok := je.fields[fk]
			if !ok {
				continue ENTRY
			}

			gpMeasurement, gpTags, gpFields, gpTimestamp, err := gp.ParseLine(string(fv))
			if err != nil {
				acc.AddError(fmt.Errorf("error parsing %q: %s", fv, err))
				continue ENTRY
			}
			measurement = gpMeasurement
			for k, v := range gpTags {
				tags[k] = v
			}
			for k, v := range gpFields {
				fields[k] = v
			}
			if !gpTimestamp.IsZero() {
				timestamp = gpTimestamp
			}
		}

		acc.AddFields(measurement, fields, tags, timestamp)
	}
}

func init() {
	inputs.Add("journalparser", func() telegraf.Input {
		return &JournalParser{}
	})
}
