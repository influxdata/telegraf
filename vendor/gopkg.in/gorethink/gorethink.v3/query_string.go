package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// Match matches against a regular expression. If no match is found, returns
// null. If there is a match then an object with the following fields is
// returned:
//   str: The matched string
//   start: The matched string’s start
//   end: The matched string’s end
//   groups: The capture groups defined with parentheses
//
// Accepts RE2 syntax (https://code.google.com/p/re2/wiki/Syntax). You can
// enable case-insensitive matching by prefixing the regular expression with
// (?i). See the linked RE2 documentation for more flags.
//
// The match command does not support backreferences.
func (t Term) Match(args ...interface{}) Term {
	return constructMethodTerm(t, "Match", p.Term_MATCH, args, map[string]interface{}{})
}

// Split splits a string into substrings. Splits on whitespace when called with no arguments.
// When called with a separator, splits on that separator. When called with a separator
// and a maximum number of splits, splits on that separator at most max_splits times.
// (Can be called with null as the separator if you want to split on whitespace while still
// specifying max_splits.)
//
// Mimics the behavior of Python's string.split in edge cases, except for splitting on the
// empty string, which instead produces an array of single-character strings.
func (t Term) Split(args ...interface{}) Term {
	return constructMethodTerm(t, "Split", p.Term_SPLIT, funcWrapArgs(args), map[string]interface{}{})
}

// Upcase upper-cases a string.
func (t Term) Upcase(args ...interface{}) Term {
	return constructMethodTerm(t, "Upcase", p.Term_UPCASE, args, map[string]interface{}{})
}

// Downcase lower-cases a string.
func (t Term) Downcase(args ...interface{}) Term {
	return constructMethodTerm(t, "Downcase", p.Term_DOWNCASE, args, map[string]interface{}{})
}
