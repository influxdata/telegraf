package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Duration just wraps time.Duration
type Duration struct {
	Duration time.Duration
}

// UnmarshalTOML parses the duration from the TOML config file
func (d *Duration) UnmarshalTOML(b []byte) error {
	dur, err := time.ParseDuration(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}

	d.Duration = dur

	return nil
}

var NotImplementedError = errors.New("not implemented yet")

type JSONFlattener struct {
	Fields map[string]interface{}
}

// FlattenJSON flattens nested maps/interfaces into a fields map
func (f *JSONFlattener) FlattenJSON(
	fieldname string,
	v interface{},
) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}
	fieldname = strings.Trim(fieldname, "_")
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			err := f.FlattenJSON(fieldname+"_"+k+"_", v)
			if err != nil {
				return err
			}
		}
	case float64:
		f.Fields[fieldname] = t
	case bool, string, []interface{}, nil:
		// ignored types
		return nil
	default:
		return fmt.Errorf("JSON Flattener: got unexpected type %T with value %v (%s)",
			t, t, fieldname)
	}
	return nil
}

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

// Glob will test a string pattern, potentially containing globs, against a
// subject string. The result is a simple true/false, determining whether or
// not the glob pattern matched the subject text.
//
// Adapted from https://github.com/ryanuber/go-glob/blob/master/glob.go
// thanks Ryan Uber!
func Glob(pattern, measurement string) bool {
	// Empty pattern can only match empty subject
	if pattern == "" {
		return measurement == pattern
	}

	// If the pattern _is_ a glob, it matches everything
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, "*")

	if len(parts) == 1 {
		// No globs in pattern, so test for match
		return pattern == measurement
	}

	leadingGlob := strings.HasPrefix(pattern, "*")
	trailingGlob := strings.HasSuffix(pattern, "*")
	end := len(parts) - 1

	for i, part := range parts {
		switch i {
		case 0:
			if leadingGlob {
				continue
			}
			if !strings.HasPrefix(measurement, part) {
				return false
			}
		case end:
			if len(measurement) > 0 {
				return trailingGlob || strings.HasSuffix(measurement, part)
			}
		default:
			if !strings.Contains(measurement, part) {
				return false
			}
		}

		// Trim evaluated text from measurement as we loop over the pattern.
		idx := strings.Index(measurement, part) + len(part)
		measurement = measurement[idx:]
	}

	// All parts of the pattern matched
	return true
}
