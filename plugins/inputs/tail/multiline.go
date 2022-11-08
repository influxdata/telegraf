package tail

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
)

// Indicates relation to the multiline event: previous or next
type MultilineMatchWhichLine int

type Multiline struct {
	config        *MultilineConfig
	enabled       bool
	patternRegexp *regexp.Regexp
	quote         byte
	inQuote       bool
}

type MultilineConfig struct {
	Pattern        string                  `toml:"pattern"`
	MatchWhichLine MultilineMatchWhichLine `toml:"match_which_line"`
	InvertMatch    bool                    `toml:"invert_match"`
	Quotation      string                  `toml:"quotation"`
	Timeout        *config.Duration        `toml:"timeout"`
}

const (
	// Previous => Append current line to previous line
	Previous MultilineMatchWhichLine = iota
	// Next => Next line will be appended to current line
	Next
)

func (m *MultilineConfig) NewMultiline() (*Multiline, error) {
	var r *regexp.Regexp

	if m.Pattern != "" {
		var err error
		if r, err = regexp.Compile(m.Pattern); err != nil {
			return nil, err
		}
	}

	var quote byte
	switch m.Quotation {
	case "", "ignore":
		m.Quotation = "ignore"
	case "single-quotes":
		quote = '\''
	case "double-quotes":
		quote = '"'
	case "backticks":
		quote = '`'
	default:
		return nil, errors.New("invalid 'quotation' setting")
	}

	enabled := m.Pattern != "" || quote != 0
	if m.Timeout == nil || time.Duration(*m.Timeout).Nanoseconds() == int64(0) {
		d := config.Duration(5 * time.Second)
		m.Timeout = &d
	}

	return &Multiline{
		config:        m,
		enabled:       enabled,
		patternRegexp: r,
		quote:         quote,
	}, nil
}

func (m *Multiline) IsEnabled() bool {
	return m.enabled
}

func (m *Multiline) ProcessLine(text string, buffer *bytes.Buffer) string {
	if m.matchQuotation(text) {
		// Ignore the returned error as we cannot do anything about it anyway
		_, _ = buffer.WriteString(text + "\n")
		return ""
	}
	if m.matchString(text) {
		// Ignore the returned error as we cannot do anything about it anyway
		_, _ = buffer.WriteString(text)
		return ""
	}

	if m.config.MatchWhichLine == Previous {
		previousText := buffer.String()
		buffer.Reset()
		if _, err := buffer.WriteString(text); err != nil {
			return ""
		}
		text = previousText
	} else {
		// Next
		if buffer.Len() > 0 {
			if _, err := buffer.WriteString(text); err != nil {
				return ""
			}
			text = buffer.String()
			buffer.Reset()
		}
	}

	return text
}

func (m *Multiline) Flush(buffer *bytes.Buffer) string {
	if buffer.Len() == 0 {
		return ""
	}
	text := buffer.String()
	buffer.Reset()
	return text
}

func (m *Multiline) matchQuotation(text string) bool {
	if m.config.Quotation == "ignore" {
		return false
	}
	escaped := false
	count := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' {
			escaped = !escaped
			continue
		}

		if text[i] == m.quote && !escaped {
			count++
		}
		escaped = false
	}
	even := count%2 == 0
	m.inQuote = (m.inQuote && even) || (!m.inQuote && !even)
	return m.inQuote
}

func (m *Multiline) matchString(text string) bool {
	if m.patternRegexp != nil {
		return m.patternRegexp.MatchString(text) != m.config.InvertMatch
	}
	return false
}

func (w MultilineMatchWhichLine) String() string {
	switch w {
	case Previous:
		return "previous"
	case Next:
		return "next"
	}
	return ""
}

// UnmarshalTOML implements ability to unmarshal MultilineMatchWhichLine from TOML files.
func (w *MultilineMatchWhichLine) UnmarshalTOML(data []byte) (err error) {
	return w.UnmarshalText(data)
}

// UnmarshalText implements encoding.TextUnmarshaler
func (w *MultilineMatchWhichLine) UnmarshalText(data []byte) (err error) {
	s := string(data)
	switch strings.ToUpper(s) {
	case `PREVIOUS`, `"PREVIOUS"`, `'PREVIOUS'`:
		*w = Previous
		return nil

	case `NEXT`, `"NEXT"`, `'NEXT'`:
		*w = Next
		return nil
	}
	*w = -1
	return fmt.Errorf("unknown multiline MatchWhichLine")
}

// MarshalText implements encoding.TextMarshaler
func (w MultilineMatchWhichLine) MarshalText() ([]byte, error) {
	s := w.String()
	if s != "" {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("unknown multiline MatchWhichLine")
}
