package tail

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal"
)

// Indicates relation to the multiline event: previous or next
type MultilineMatchWhichLine int

type Multiline struct {
	config        *MultilineConfig
	enabled       bool
	patternRegexp *regexp.Regexp
}

type MultilineConfig struct {
	Pattern        string
	MatchWhichLine MultilineMatchWhichLine `toml:"match_which_line"`
	InvertMatch    bool
	Timeout        *internal.Duration
}

const (
	// Previous => Append current line to previous line
	Previous MultilineMatchWhichLine = iota
	// Next => Next line will be appended to current line
	Next
)

func (m *MultilineConfig) NewMultiline() (*Multiline, error) {
	enabled := false
	var r *regexp.Regexp
	var err error

	if m.Pattern != "" {
		enabled = true
		if r, err = regexp.Compile(m.Pattern); err != nil {
			return nil, err
		}
		if m.Timeout == nil || m.Timeout.Duration.Nanoseconds() == int64(0) {
			m.Timeout = &internal.Duration{Duration: 5 * time.Second}
		}
	}

	return &Multiline{
		config:        m,
		enabled:       enabled,
		patternRegexp: r}, nil
}

func (m *Multiline) IsEnabled() bool {
	return m.enabled
}

func (m *Multiline) ProcessLine(text string, buffer *bytes.Buffer) string {
	if m.matchString(text) {
		buffer.WriteString(text)
		return ""
	}

	if m.config.MatchWhichLine == Previous {
		previousText := buffer.String()
		buffer.Reset()
		buffer.WriteString(text)
		text = previousText
	} else {
		// Next
		if buffer.Len() > 0 {
			buffer.WriteString(text)
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

func (m *Multiline) matchString(text string) bool {
	return m.patternRegexp.MatchString(text) != m.config.InvertMatch
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
		return

	case `NEXT`, `"NEXT"`, `'NEXT'`:
		*w = Next
		return
	}
	*w = -1
	return fmt.Errorf("E! [inputs.tail] unknown multiline MatchWhichLine")
}

// MarshalText implements encoding.TextMarshaler
func (w MultilineMatchWhichLine) MarshalText() ([]byte, error) {
	s := w.String()
	if s != "" {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("E! [inputs.tail] unknown multiline MatchWhichLine")
}
