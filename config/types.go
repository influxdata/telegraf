package config

import (
	"bytes"
	"strconv"
	"time"

	"github.com/alecthomas/units"
)

// Duration is a time.Duration
type Duration time.Duration

// Size is an int64
type Size int64

// UnmarshalTOML parses the duration from the TOML config file
func (d *Duration) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	// see if we can directly convert it
	dur, err := time.ParseDuration(string(b))
	if err == nil {
		*d = Duration(dur)
		return nil
	}

	// Parse string duration, ie, "1s"
	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		dur, err := time.ParseDuration(uq)
		if err == nil {
			*d = Duration(dur)
			return nil
		}
	}

	// First try parsing as integer seconds
	sI, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		dur := time.Second * time.Duration(sI)
		*d = Duration(dur)
		return nil
	}
	// Second try parsing as float seconds
	sF, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		dur := time.Second * time.Duration(sF)
		*d = Duration(dur)
		return nil
	}

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
