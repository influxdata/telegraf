package parsers

import (
	"github.com/influxdata/telegraf"
)

// Creator is the function to create a new parser
type Creator func(defaultMetricName string) telegraf.Parser

// Parsers contains the registry of all known parsers (following the new style)
var Parsers = map[string]Creator{}

// Add adds a parser to the registry. Usually this function is called in the plugin's init function
func Add(name string, creator Creator) {
	Parsers[name] = creator
}

type ParserFunc func() (Parser, error)

// ParserInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserInput interface {
	// SetParser sets the parser function for the interface
	SetParser(parser Parser)
}

// ParserFuncInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserFuncInput interface {
	// SetParserFunc returns a new parser.
	SetParserFunc(fn ParserFunc)
}

// Parser is an interface defining functions that a parser plugin must satisfy.
type Parser interface {
	// Parse takes a byte buffer separated by newlines
	// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
	// and parses it into telegraf metrics
	//
	// Must be thread-safe.
	Parse(buf []byte) ([]telegraf.Metric, error)

	// ParseLine takes a single string metric
	// ie, "cpu.usage.idle 90"
	// and parses it into a telegraf metric.
	//
	// Must be thread-safe.
	// This function is only called by plugins that expect line based protocols
	// Doesn't need to be implemented by non-linebased parsers (e.g. json, xml)
	ParseLine(line string) (telegraf.Metric, error)

	// SetDefaultTags tells the parser to add all of the given tags
	// to each parsed metric.
	// NOTE: do _not_ modify the map after you've passed it here!!
	SetDefaultTags(tags map[string]string)
}
