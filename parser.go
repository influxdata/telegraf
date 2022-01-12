package telegraf

// Parser is an interface defining functions that a parser plugin must satisfy.
type Parser interface {
	// Parse takes a byte buffer separated by newlines
	// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
	// and parses it into telegraf metrics
	//
	// Must be thread-safe.
	Parse(buf []byte) ([]Metric, error)

	// ParseLine takes a single string metric
	// ie, "cpu.usage.idle 90"
	// and parses it into a telegraf metric.
	//
	// Must be thread-safe.
	ParseLine(line string) (Metric, error)

	// SetDefaultTags tells the parser to add all of the given tags
	// to each parsed metric.
	// NOTE: do _not_ modify the map after you've passed it here!!
	SetDefaultTags(tags map[string]string)
}

// HeaderParser is an optional interface for parsers that require
// parsing of header-information. This is relevant for input plugins
// using the ParseLine function.
type HeaderParser interface {
	// Return the number of lines to skip before start parsing
	// This might be necessary to drop some garbage at the beginning
	// of the data returning malformed data or invalid metrics.
	GetSkipLineCount() int

	// Return the number of lines required read the complete header
	// This might be necessary to read header information at the
	// beginning of the data.
	GetHeaderLineCount() int

	// Read the header information
	// It might be necessary to call this multiple times for the
	// header information to complete.
	ParseHeaderLine(line string) error
}

type ParserFunc func() (Parser, error)

// StatefulParser is an optional interface for parsers
// requiring special handling to generate a new instance.
// By default the same instance is returned.
type StatefulParser interface {
	// Return a new instance of the parser avoiding side-effect
	// due to multiple calls to the parser within the same plugin.
	NewInstance() (Parser, error)
}

// ParserInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserInput interface {
	// SetParser sets the parser function for the interface
	SetParser(parser Parser)
}

// ParserFuncInput is an interface for input plugins that are able to parse
// arbitrary data formats.
type ParserFuncInput interface {
	// GetParser returns a new parser.
	SetParserFunc(fn ParserFunc)
}
