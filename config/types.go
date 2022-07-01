package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/units"
)

// Duration is a time.Duration
type Duration time.Duration

// Size is an int64
type Size int64

// UnmarshalTOML parses the duration from the TOML config file
func (d *Duration) UnmarshalTOML(b []byte) error {
	// convert to string
	durStr := string(b)

	// Value is a TOML number (e.g. 3, 10, 3.5)
	// First try parsing as integer seconds
	sI, err := strconv.ParseInt(durStr, 10, 64)
	if err == nil {
		dur := time.Second * time.Duration(sI)
		*d = Duration(dur)
		return nil
	}
	// Second try parsing as float seconds
	sF, err := strconv.ParseFloat(durStr, 64)
	if err == nil {
		dur := time.Second * time.Duration(sF)
		*d = Duration(dur)
		return nil
	}

	// Finally, try value is a TOML string (e.g. "3s", 3s) or literal (e.g. '3s')
	durStr = strings.ReplaceAll(durStr, "'", "")
	durStr = strings.ReplaceAll(durStr, "\"", "")
	if durStr == "" {
		durStr = "0s"
	}
	// special case: logging interval had a default of 0d, which silently
	// failed, but in order to prevent issues with default configs that had
	// uncommented the option, change it from zero days to zero hours.
	if durStr == "0d" {
		durStr = "0h"
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	return d.UnmarshalTOML(text)
}

func (s *Size) UnmarshalTOML(b []byte) error {
	var err error
	if len(b) == 0 {
		return nil
	}
	str := string(b)
	if b[0] == '"' || b[0] == '\'' {
		str, err = strconv.Unquote(str)
		if err != nil {
			return err
		}
	}

	val, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		*s = Size(val)
		return nil
	}
	val, err = units.ParseStrictBytes(str)
	if err != nil {
		return err
	}
	*s = Size(val)
	return nil
}

func (s *Size) UnmarshalText(text []byte) error {
	return s.UnmarshalTOML(text)
}
