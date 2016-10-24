/*
procfilter is an input plugin designed to filter the processes using various methods. (top, exceed, children, ...)
Metrics corresponding to a set of filtered processes can be aggregated to create workloads.
*/
package procfilter

import (
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ProcFilter struct {
	Script             string
	Script_file        string
	Measurement_prefix string
	Tag_prefix         string
	Field_prefix       string
	parser             *Parser
	parseOK            bool
}

func NewProcFilter() *ProcFilter {
	return &ProcFilter{Measurement_prefix: "pf_"}
}

var sampleConfig = `
  ## Set various prefixes used in identifiers.
  # measurement_prefix = "pf_"
  # tag_prefix = ""
  # field_prefix = ""
  ## Describe what you want to measure by writting a script.
  ## (in an external file of directly in this configuration)
  # script_file = ""
  # script = """
  #   joe_filter <- user('joe')
  #   joe = tag(cmd) fields(rss,cpu) <- top(rss,5,joe_filter)
  #   wl_web = fields(cpu,rss,vsz,process_nb) <- pack(user('apache'),group('tomcat',children(cmd('nginx')))
  # """
  ## Syntax is described in the README.md on github.
`

func (_ *ProcFilter) SampleConfig() string {
	return sampleConfig
}

func (_ *ProcFilter) Description() string {
	return "Monitor process cpu and memory usage with filters and aggregation"
}

func init() {
	inputs.Add("procfilter", func() telegraf.Input {
		return NewProcFilter()
	})
}

func (p *ProcFilter) Gather(acc telegraf.Accumulator) error {
	if p.parser == nil {
		if p.Script_file != "" {
			if p.Script != "" {
				logErr(fmt.Sprintf("E! You cannot have non empty script and script_file at the same time"))
				return nil
			}
			s, err := fileContent(p.Script)
			if err != nil {
				logErr(err.Error())
				return nil
			}
			p.Script = s

		}
		// Init and parse the script to build the AST.
		p.parser = NewParser(strings.NewReader(p.Script))
		err := p.parser.Parse()
		if err != nil {
			logErr(err.Error())
			return nil
		}
		p.parseOK = true
	}
	if !p.parseOK {
		// Data stored in the parser may be inconsistent, do not gather.
		return nil
	}

	// Use the ASTs stored in the parser to process all filters then output the measurements.
	parser := p.parser
	if len(parser.measurements) == 0 {
		// No measurement, do nothing!
		return nil
	}
	// Change the current stamp and update all global variables
	newSample()
	for _, m := range parser.measurements {
		err := m.f.Apply()
		if err != nil {
			logErr(err.Error())
			continue
		}
		iStats := m.f.Stats()
		for _, ps := range iStats.pid2Stat {
			tags, err := m.getTags(ps, p.Tag_prefix)
			if err != nil {
				logErr(err.Error())
				continue
			}
			fields, err := m.getFields(ps, p.Field_prefix)
			if err != nil {
				logErr(err.Error())
				continue
			}
			acc.AddFields(p.Measurement_prefix+m.name, fields, tags)
		}
	}
	return nil
}

// Change the stamp for a new sample (thus disabling all previous values from last sample)
// Update the global sets of (P)IDs
func newSample() {
	nStamp := stamp + 1
	if nStamp >= 128 { // We could cycle over 0/1 but that makes debug easier.
		nStamp = 0
	}
	stamp = nStamp
	resetGlobalStatSets()
}
