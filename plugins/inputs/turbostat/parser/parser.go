//go:build linux && amd64

package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/influxdata/telegraf"
)

var abbreviations = map[string]string{
	"%":    "percent",
	"+":    "plus",
	"-":    "minus",
	"watt": "power_watt",
	"mhz":  "frequency_mhz",
	"tmp":  "temperature_celsius",
	"thr":  "throttle",
	"avg":  "average",
	"cor":  "core",
	"bzy":  "busy",
	"pkg":  "package",
	"sys":  "system",
	"unc":  "uncore",
	"u":    "uncore",
	"a":    "actual",
	"j":    "energy_joule",
}

var keepSymbols = map[rune]struct{}{
	'%': {},
	'+': {},
	'-': {},
}

var knownTokens = []string{
	"Watt",
	"MHz",
	"Tmp",
	"Thr",
	"GFX",
	"SAM",
}

var tagNames = map[string]struct{}{
	"package": {},
	"node":    {},
	"die":     {},
	"core":    {},
	"cpu":     {},
	"apic":    {},
	"x2apic":  {},
}

type column struct {
	name      string
	isTag     bool
	isIgnored bool
}

type tagMap = map[string]string
type fieldMap = map[string]any

// ProcessStream reads metrics from a stream (such as Turbostat stdout) and
// adds them to an accumulator. If an error is encountered, the function
// returns it and stops further processing.
func ProcessStream(r io.Reader, acc telegraf.Accumulator) error {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return scanner.Err()
	}
	header := scanner.Text()
	columns := make([]column, 0)
	for _, s := range strings.Fields(header) {
		columns = append(columns, createColumn(s))
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == header {
			continue
		}
		values := strings.Fields(line)
		tags, fields, err := processValues(columns, values)
		if err != nil {
			return err
		}
		acc.AddFields("turbostat", fields, tags)
	}
	return scanner.Err()
}

// Creates tags and fields from columns and values.
func processValues(columns []column, values []string) (tagMap, fieldMap, error) {
	if len(values) > len(columns) {
		return nil, nil, fmt.Errorf("too many values: %d columns, %d values", len(columns), len(values))
	}
	tags := make(tagMap, 0)
	fields := make(fieldMap, 0)
	for i := range values {
		if columns[i].isIgnored {
			continue
		}
		if columns[i].isTag {
			if !isValidTagValue(values[i]) {
				return nil, nil, fmt.Errorf("invalid tag: %s", values[i])
			}
			tags[columns[i].name] = values[i]
		} else {
			v, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				return nil, nil, err
			}
			fields[columns[i].name] = v
		}
	}
	if len(fields) == 0 {
		return nil, nil, errors.New("no value for any field")
	}
	return tags, fields, nil
}

// Creates a column struct from a Turbostat column name.
func createColumn(s string) column {
	c := column{}
	// Split the Turbostat column name into tokens.
	tokens := make([]string, 0)
	for _, element := range splitSymbols(s) {
		tokens = append(tokens, splitKnownTokens(element)...)
	}
	for i, token := range tokens {
		token = strings.ToLower(token)
		// If a token is an abbreviation, replace it.
		if replacement, ok := abbreviations[token]; ok {
			token = replacement
		}
		tokens[i] = token
	}
	// Build the camel case column name.
	c.name = strings.Join(tokens, "_")
	// Ignore the timestamp column.
	if c.name == "time_of_day_seconds" {
		c.isIgnored = true
	}
	// If the name matches a tag, mark the column as such.
	if _, ok := tagNames[c.name]; ok {
		c.isTag = true
	}
	return c
}

// Splits a string into tokens using non-alphanumeric characters as delimiters.
// The delimiters are discarded unless they belong to a set of symbols to keep.
func splitSymbols(s string) []string {
	tokens := make([]string, 0)
	i := 0
	start := i
	for i, c := range s {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) {
			if start < i {
				tokens = append(tokens, s[start:i])
			}
			if _, ok := keepSymbols[c]; ok {
				tokens = append(tokens, string(c))
			}
			start = i + 1
		}
	}
	if start < len(s) {
		tokens = append(tokens, s[start:])
	}
	return tokens
}

// Splits an alphanumeric string into tokens, using a list
// of known tokens to determine boundaries.
func splitKnownTokens(s string) []string {
	tokens := make([]string, 0)
	i := 0
	start := i
	for i < len(s) {
		match := false
		for _, hint := range knownTokens {
			if strings.HasPrefix(s[i:], hint) {
				match = true
				if start < i {
					tokens = append(tokens, s[start:i])
				}
				tokens = append(tokens, hint)
				i += len(hint)
				start = i
				break
			}
		}
		if !match {
			i++
		}
	}
	if start < i {
		tokens = append(tokens, s[start:i])
	}
	return tokens
}

// Returns whether a string represents a tag value or not.
// Turbostat only uses integers and "-".
func isValidTagValue(s string) bool {
	if s == "-" {
		return true
	}
	for _, c := range s {
		if !unicode.IsNumber(c) {
			return false
		}
	}
	return true
}
