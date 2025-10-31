package input

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx/influx_upstream"
	"github.com/influxdata/telegraf/testutil"
)

//go:embed sample.conf
var sampleConfig string

// Plugin struct should be named the same as the plugin
type Plugin struct {
	Files            []string               `toml:"files"`
	DefaultTags      map[string]string      `toml:"default_tag_defs"`
	AdditionalParams map[string]interface{} `toml:"additional_params"`
	Parser           telegraf.Parser        `toml:"-"`
	Log              telegraf.Logger        `toml:"-"`

	// Settings used by test-code
	Path             string `toml:"-"` // start path for relative files
	ExpectedFilename string `toml:"-"` // filename of the expected metrics (default: expected.out)
	UseTypeTag       string `toml:"-"` // if specified use this tag to infer metric type

	// Test-data derived from the files
	Expected              []telegraf.Metric `toml:"-"` // expected metrics
	ExpectedErrors        []string          `toml:"-"` // expected errors
	ShouldIgnoreTimestamp bool              `toml:"-"` // flag indicating if the expected metrics do have timestamps

	// Internal data
	inputFilenames []string
}

func (*Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Init() error {
	// Setup the filenames
	p.inputFilenames = make([]string, 0, len(p.Files))
	for _, fn := range p.Files {
		if !filepath.IsAbs(fn) && p.Path != "" {
			fn = filepath.Join(p.Path, fn)
		}
		p.inputFilenames = append(p.inputFilenames, fn)
	}

	// Setup an influx parser for reading the expected metrics
	expectedParser := &influx_upstream.Parser{}
	if err := expectedParser.Init(); err != nil {
		return err
	}
	expectedParser.SetTimeFunc(func() time.Time { return time.Time{} })

	// Read the expected metrics if any
	expectedFn := "expected.out"
	if p.ExpectedFilename != "" {
		expectedFn = p.ExpectedFilename
	}
	if !filepath.IsAbs(expectedFn) && p.Path != "" {
		expectedFn = filepath.Join(p.Path, expectedFn)
	}
	if _, err := os.Stat(expectedFn); err == nil {
		var err error
		p.Expected, err = testutil.ParseMetricsFromFile(expectedFn, expectedParser)
		if err != nil {
			return err
		}
	}

	// Read the expected errors if any
	expectedErrorFn := "expected.err"
	if !filepath.IsAbs(expectedErrorFn) && p.Path != "" {
		expectedErrorFn = filepath.Join(p.Path, expectedErrorFn)
	}
	if _, err := os.Stat(expectedErrorFn); err == nil {
		var err error
		p.ExpectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFn)
		if err != nil {
			return err
		}
		if len(p.ExpectedErrors) == 0 {
			return errors.New("got empty expected errors file")
		}
	}

	// Fixup the metric type if requested
	if p.UseTypeTag != "" {
		for i, m := range p.Expected {
			typeTag, found := m.GetTag(p.UseTypeTag)
			if !found {
				continue
			}
			var mtype telegraf.ValueType
			switch typeTag {
			case "counter":
				mtype = telegraf.Counter
			case "gauge":
				mtype = telegraf.Gauge
			case "untyped":
				mtype = telegraf.Untyped
			case "summary":
				mtype = telegraf.Summary
			case "histogram":
				mtype = telegraf.Histogram
			default:
				continue
			}
			m.SetType(mtype)
			m.RemoveTag(p.UseTypeTag)
			p.Expected[i] = m
		}
	}

	// Determine if we should check the timestamps indicated by a missing
	// timestamp in the expected input
	for i, m := range p.Expected {
		missingTimestamp := m.Time().IsZero()
		if i == 0 {
			p.ShouldIgnoreTimestamp = missingTimestamp
			continue
		}
		if missingTimestamp != p.ShouldIgnoreTimestamp {
			return errors.New("mixed timestamp and non-timestamp data in expected metrics")
		}
	}

	// Set the parser's default tags, just in case
	if p.Parser != nil {
		p.Parser.SetDefaultTags(p.DefaultTags)
	}

	return nil
}

func (p *Plugin) Gather(acc telegraf.Accumulator) error {
	if p.Parser == nil {
		return errors.New("no parser defined")
	}

	for _, fn := range p.inputFilenames {
		data, err := os.ReadFile(fn)
		if err != nil {
			return err
		}
		metrics, err := p.Parser.Parse(data)
		if err != nil {
			return err
		}
		for _, m := range metrics {
			acc.AddMetric(m)
		}
	}

	return nil
}

func (p *Plugin) SetParser(parser telegraf.Parser) {
	p.Parser = parser
	if len(p.DefaultTags) > 0 {
		p.Parser.SetDefaultTags(p.DefaultTags)
	}
}

// Register the plugin
func init() {
	inputs.Add("test", func() telegraf.Input {
		return &Plugin{}
	})
}
