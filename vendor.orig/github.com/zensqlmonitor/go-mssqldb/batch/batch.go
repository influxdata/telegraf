// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// bach splits a single script containing multiple batches separated by
// a keyword into multiple scripts.
package batch

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Split the provided SQL into multiple sql scripts based on a given
// separator, often "GO". It also allows escaping newlines with a
// backslash.
func Split(sql, separator string) []string {
	if len(separator) == 0 || len(sql) < len(separator) {
		return []string{sql}
	}
	l := &lexer{
		Sql: sql,
		Sep: separator,
		At:  0,
	}
	state := stateWhitespace
	for state != nil {
		state = state(l)
	}
	l.AddCurrent(1)
	return l.Batch
}

const debugPrintStateName = false

func printStateName(name string, l *lexer) {
	if debugPrintStateName {
		fmt.Printf("state %s At=%d\n", name, l.At)
	}
}

func hasPrefixFold(s, sep string) bool {
	if len(s) < len(sep) {
		return false
	}
	return strings.EqualFold(s[:len(sep)], sep)
}

type lexer struct {
	Sql   string
	Sep   string
	At    int
	Start int

	Skip []int

	Batch []string
}

func (l *lexer) Add(b string) {
	if len(b) == 0 {
		return
	}
	l.Batch = append(l.Batch, b)
}

func (l *lexer) Next() bool {
	l.At++
	return l.At < len(l.Sql)
}

func (l *lexer) AddCurrent(count int64) bool {
	if count < 0 {
		count = 0
	}
	if l.At >= len(l.Sql) {
		l.At = len(l.Sql)
	}
	text := l.Sql[l.Start:l.At]
	if len(l.Skip) > 0 {
		buf := &bytes.Buffer{}
		nextSkipIndex := 0
		nextSkip := l.Skip[nextSkipIndex]
		for i, r := range text {
			if i == nextSkip {
				nextSkipIndex++
				if nextSkipIndex < len(l.Skip) {
					nextSkip = l.Skip[nextSkipIndex]
				}
				continue
			}
			buf.WriteRune(r)
		}
		text = buf.String()
		l.Skip = nil
	}
	// Limit the number of counts for sanity.
	if count > 1000 {
		count = 1000
	}
	for i := int64(0); i < count; i++ {
		l.Add(text)
	}
	l.At += len(l.Sep)
	l.Start = l.At
	return (l.At < len(l.Sql))
}

type stateFn func(*lexer) stateFn

const (
	lineComment  = "--"
	leftComment  = "/*"
	rightComment = "*/"
)

func stateSep(l *lexer) stateFn {
	printStateName("sep", l)
	if l.At+len(l.Sep) >= len(l.Sql) {
		return nil
	}
	s := l.Sql[l.At+len(l.Sep):]

	parseNumberStart := -1
loop:
	for i, r := range s {
		switch {
		case r == '\n', r == '\r':
			l.AddCurrent(1)
			return stateWhitespace
		case unicode.IsSpace(r):
		case unicode.IsNumber(r):
			parseNumberStart = i
			break loop
		}
	}
	if parseNumberStart < 0 {
		return nil
	}

	parseNumberCount := 0
numLoop:
	for i, r := range s[parseNumberStart:] {
		switch {
		case unicode.IsNumber(r):
			parseNumberCount = i
		default:
			break numLoop
		}
	}
	parseNumberEnd := parseNumberStart + parseNumberCount + 1

	count, err := strconv.ParseInt(s[parseNumberStart:parseNumberEnd], 10, 64)
	if err != nil {
		return stateText
	}
	for _, r := range s[parseNumberEnd:] {
		switch {
		case r == '\n', r == '\r':
			l.AddCurrent(count)
			l.At += parseNumberEnd
			l.Start = l.At
			return stateWhitespace
		case unicode.IsSpace(r):
		default:
			return stateText
		}
	}

	return nil
}

func stateText(l *lexer) stateFn {
	printStateName("text", l)
	for {
		ch := l.Sql[l.At]

		switch {
		case strings.HasPrefix(l.Sql[l.At:], lineComment):
			l.At += len(lineComment)
			return stateLineComment
		case strings.HasPrefix(l.Sql[l.At:], leftComment):
			l.At += len(leftComment)
			return stateMultiComment
		case ch == '\'':
			l.At += 1
			return stateString
		case ch == '\r', ch == '\n':
			l.At += 1
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}

func stateWhitespace(l *lexer) stateFn {
	printStateName("whitespace", l)
	if l.At >= len(l.Sql) {
		return nil
	}
	ch := l.Sql[l.At]

	switch {
	case unicode.IsSpace(rune(ch)):
		l.At += 1
		return stateWhitespace
	case hasPrefixFold(l.Sql[l.At:], l.Sep):
		return stateSep
	default:
		return stateText
	}
}

func stateLineComment(l *lexer) stateFn {
	printStateName("line-comment", l)
	for {
		if l.At >= len(l.Sql) {
			return nil
		}
		ch := l.Sql[l.At]

		switch {
		case ch == '\r', ch == '\n':
			l.At += 1
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}

func stateMultiComment(l *lexer) stateFn {
	printStateName("multi-line-comment", l)
	for {
		switch {
		case strings.HasPrefix(l.Sql[l.At:], rightComment):
			l.At += len(leftComment)
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}

func stateString(l *lexer) stateFn {
	printStateName("string", l)
	for {
		if l.At >= len(l.Sql) {
			return nil
		}
		ch := l.Sql[l.At]
		chNext := rune(-1)
		if l.At+1 < len(l.Sql) {
			chNext = rune(l.Sql[l.At+1])
		}

		switch {
		case ch == '\\' && (chNext == '\r' || chNext == '\n'):
			next := 2
			l.Skip = append(l.Skip, l.At, l.At+1)
			if chNext == '\r' && l.At+2 < len(l.Sql) && l.Sql[l.At+2] == '\n' {
				l.Skip = append(l.Skip, l.At+2)
				next = 3
			}
			l.At += next
		case ch == '\'' && chNext == '\'':
			l.At += 2
		case ch == '\'' && chNext != '\'':
			l.At += 1
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}
