package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/template"
	"github.com/compose-spec/compose-go/utils"
)

type trimmer struct {
	input  *bytes.Reader
	output bytes.Buffer
}

func removeComments(buf []byte) ([]byte, error) {
	t := &trimmer{
		input:  bytes.NewReader(buf),
		output: bytes.Buffer{},
	}
	err := t.process()
	return t.output.Bytes(), err
}

func (t *trimmer) process() error {
	for {
		// Read the next byte until EOF
		c, err := t.input.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		// Switch states if we need to
		switch c {
		case '\\':
			//nolint:errcheck // next byte is known
			t.input.UnreadByte()
			err = t.escape()
		case '\'':
			//nolint:errcheck // next byte is known
			t.input.UnreadByte()
			if t.hasNQuotes(c, 3) {
				err = t.tripleSingleQuote()
			} else {
				err = t.singleQuote()
			}
		case '"':
			//nolint:errcheck // next byte is known
			t.input.UnreadByte()
			if t.hasNQuotes(c, 3) {
				err = t.tripleDoubleQuote()
			} else {
				err = t.doubleQuote()
			}
		case '#':
			err = t.comment()
		default:
			t.output.WriteByte(c)
			continue
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}
	return nil
}

func (t *trimmer) hasNQuotes(ref byte, limit int64) bool {
	var count int64
	// Look ahead check if the next characters are what we expect
	for count = 0; count < limit; count++ {
		c, err := t.input.ReadByte()
		if err != nil || c != ref {
			break
		}
	}
	// We also need to unread the non-matching character
	offset := -count
	if count < limit {
		offset--
	}
	//nolint:errcheck // Unread the already matched characters
	t.input.Seek(offset, io.SeekCurrent)
	return count >= limit
}

func (t *trimmer) readWriteByte() (byte, error) {
	c, err := t.input.ReadByte()
	if err != nil {
		return 0, err
	}
	return c, t.output.WriteByte(c)
}

func (t *trimmer) escape() error {
	//nolint:errcheck // Consume the known starting backslash and quote
	t.readWriteByte()

	// Read the next character which is the escaped one and exit
	_, err := t.readWriteByte()
	return err
}

func (t *trimmer) singleQuote() error {
	//nolint:errcheck // Consume the known starting quote
	t.readWriteByte()

	// Read bytes until EOF, line end or another single quote
	for {
		if c, err := t.readWriteByte(); err != nil || c == '\'' || c == '\n' {
			return err
		}
	}
}

func (t *trimmer) tripleSingleQuote() error {
	for i := 0; i < 3; i++ {
		//nolint:errcheck // Consume the known starting quotes
		t.readWriteByte()
	}

	// Read bytes until EOF or another set of triple single quotes
	for {
		c, err := t.readWriteByte()
		if err != nil {
			return err
		}

		if c == '\'' && t.hasNQuotes('\'', 2) {
			//nolint:errcheck // Consume the two additional ending quotes
			t.readWriteByte()
			//nolint:errcheck // Consume the two additional ending quotes
			t.readWriteByte()
			return nil
		}
	}
}

func (t *trimmer) doubleQuote() error {
	//nolint:errcheck // Consume the known starting quote
	t.readWriteByte()

	// Read bytes until EOF, line end or another double quote
	for {
		c, err := t.input.ReadByte()
		if err != nil {
			return err
		}
		switch c {
		case '\\':
			//nolint:errcheck // Consume the found escaped character
			t.input.UnreadByte()
			if err := t.escape(); err != nil {
				return err
			}
			continue
		case '"', '\n':
			// Found terminator
			return t.output.WriteByte(c)
		}
		t.output.WriteByte(c)
	}
}

func (t *trimmer) tripleDoubleQuote() error {
	for i := 0; i < 3; i++ {
		//nolint:errcheck // Consume the known starting quotes
		t.readWriteByte()
	}

	// Read bytes until EOF or another set of triple double quotes
	for {
		c, err := t.input.ReadByte()
		if err != nil {
			return err
		}
		switch c {
		case '\\':
			//nolint:errcheck // Consume the found escape character
			t.input.UnreadByte()
			if err := t.escape(); err != nil {
				return err
			}
			continue
		case '"':
			t.output.WriteByte(c)
			if t.hasNQuotes('"', 2) {
				//nolint:errcheck // Consume the two additional ending quotes
				t.readWriteByte()
				//nolint:errcheck // Consume the two additional ending quotes
				t.readWriteByte()
				return nil
			}
			continue
		}
		t.output.WriteByte(c)
	}
}

func (t *trimmer) comment() error {
	// Read bytes until EOF or a line break
	for {
		c, err := t.input.ReadByte()
		if err != nil {
			return err
		}
		if c == '\n' {
			return t.output.WriteByte(c)
		}
	}
}

func substituteEnvironment(contents []byte, oldReplacementBehavior bool) ([]byte, error) {
	options := []template.Option{
		template.WithReplacementFunction(func(s string, m template.Mapping, cfg *template.Config) (string, error) {
			result, applied, err := template.DefaultReplacementAppliedFunc(s, m, cfg)
			if err == nil && !applied {
				// Keep undeclared environment-variable patterns to reproduce
				// pre-v1.27 behavior
				return s, nil
			}
			if err != nil && strings.HasPrefix(err.Error(), "Invalid template:") {
				// Keep invalid template patterns to ignore regexp substitutions
				// like ${1}
				return s, nil
			}
			return result, err
		}),
		template.WithoutLogging,
	}
	if oldReplacementBehavior {
		options = append(options, template.WithPattern(oldVarRe))
	}

	envMap := utils.GetAsEqualsMap(os.Environ())
	retVal, err := template.SubstituteWithOptions(string(contents), func(k string) (string, bool) {
		if v, ok := envMap[k]; ok {
			return v, ok
		}
		return "", false
	}, options...)
	return []byte(retVal), err
}
