package parsers

import "errors"

// EOF is similar to io.EOF but is a separate type to make sure we
// have checked the parsers using it to have the same meaning (i.e.
// it needs more data to complete parsing) and a way to detect partial
// data.
var EOF = errors.New("not enough data")
